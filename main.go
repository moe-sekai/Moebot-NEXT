package main

import (
	"embed"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/commands"
	"moebot-next/internal/config"
	"moebot-next/internal/database"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/ranking"
	"moebot-next/internal/renderer"
	"moebot-next/internal/sekai"
	"moebot-next/internal/web"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed web/dist/*
var webUI embed.FS

func main() {
	cfgPath := os.Getenv("MOEBOT_CONFIG")
	if cfgPath == "" {
		cfgPath = "config.yml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	setupLogger(cfg.Log)

	if err := ensureRuntimeDirs(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to create runtime directories")
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	store := masterdata.NewStore()
	loader := masterdata.NewLoader(cfg.Masterdata, store)
	if err := loader.LoadAll(); err != nil {
		log.Warn().Err(err).Msg("Initial masterdata load failed; bot will still start with empty data")
	}
	if cfg.Masterdata.RefreshInterval > 0 {
		loader.StartPeriodicRefresh(time.Duration(cfg.Masterdata.RefreshInterval) * time.Second)
		defer loader.StopPeriodicRefresh()
	}

	rendererClient := renderer.New(cfg.Renderer)
	if err := rendererClient.StartProcess("renderer", cfg.Renderer.Port); err != nil {
		log.Warn().Err(err).Msg("Renderer process failed to start; commands will fallback to text")
	} else {
		defer rendererClient.StopProcess()
	}

	sekaiClient := sekai.NewClient(cfg.SekaiAPI)
	rankingClient := ranking.NewClient(ranking.Config{
		BaseURL: cfg.RankingAPI.BaseURL,
		Region:  cfg.RankingAPI.Region,
		Timeout: cfg.RankingAPI.Timeout,
	})

	bot.RegisterMiddleware(db)
	commands.RegisterAll(&commands.Deps{
		Store:    store,
		DB:       db,
		Renderer: rendererClient,
		Sekai:    sekaiClient,
		Ranking:  rankingClient,
	})

	webServer := web.New(cfg, db, store, rendererClient)
	webServer.SetupStaticFiles(webUI)
	go func() {
		if err := webServer.Start(); err != nil {
			log.Error().Err(err).Msg("Web server stopped")
		}
	}()
	defer webServer.Shutdown()

	b := bot.New(cfg.Bot)
	go b.Run()

	waitForSignal()
	log.Info().Msg("Moebot NEXT shutting down")
}

func setupLogger(cfg config.LogConfig) {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	if cfg.Format != "json" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime})
	}
}

func ensureRuntimeDirs(cfg *config.Config) error {
	dirs := []string{
		filepath.Dir(cfg.Database.Path),
		cfg.Masterdata.LocalPath,
		cfg.Renderer.Cache.Path,
		cfg.Assets.StickerPath,
	}
	for _, dir := range dirs {
		if dir == "" || dir == "." {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func waitForSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}

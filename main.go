package main

import (
	"context"
	"embed"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"moebot-next/internal/assets"
	"moebot-next/internal/b30"
	"moebot-next/internal/bot"
	"moebot-next/internal/filter"
	"moebot-next/internal/models"

	"moebot-next/internal/commandparser"
	"moebot-next/internal/commands"
	"moebot-next/internal/config"
	"moebot-next/internal/database"
	"moebot-next/internal/logbuffer"
	"moebot-next/internal/renderer"
	"moebot-next/internal/servers"
	"moebot-next/internal/web"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed web/dist/*
var webUI embed.FS

func main() {
	cfgPath := os.Getenv("MOEBOT_CONFIG")
	if cfgPath == "" {
		cfgPath = "data/config.yml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	logBuffer := setupLogger(cfg.Log)

	if err := ensureRuntimeDirs(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to create runtime directories")
	}
	if _, err := assets.Configure(cfg.Assets, cfg.Server.Region); err != nil {
		log.Warn().Err(err).Msg("Asset CDN config is invalid; using built-in default")
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	serverManager := servers.NewManager(cfg)
	serverManager.LoadEnabled()
	serverManager.StartPeriodicRefresh()
	defer serverManager.StopPeriodicRefresh()

	b30Client := b30.NewClient(cfg.B30)
	rendererClient := renderer.New(cfg.Renderer)
	if err := rendererClient.StartProcess("renderer", cfg.Renderer.Port); err != nil {
		log.Warn().Err(err).Msg("Renderer process failed to start; commands will fallback to text")
	} else {
		defer rendererClient.StopProcess()
	}

	commandDefinitions := commandparser.Definitions(cfg.Bot.CommandAliases)
	bot.RegisterMiddleware(db)
	commands.RegisterAll(&commands.Deps{
		DB:          db,
		Renderer:    rendererClient,
		Servers:     serverManager,
		B30:         b30Client,
		Definitions: commandDefinitions,
	})

	if err := seedBuiltinFilterApp(db, cfg.Bot.Driver); err != nil {
		log.Warn().Err(err).Msg("Failed to seed builtin filter app")
	}
	filterManager := filter.New(db)
	if err := filterManager.Start(context.Background()); err != nil {
		log.Warn().Err(err).Msg("Filter gateway failed to start")
	}
	defer filterManager.Stop()

	webServer := web.New(cfg, db, serverManager.Default().Store, rendererClient, cfgPath, serverManager.Default().Loader)
	webServer.Servers = serverManager
	webServer.Logs = logBuffer
	webServer.Filter = filterManager
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

func setupLogger(cfg config.LogConfig) *logbuffer.Buffer {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	buf := logbuffer.New(cfg.Buffer)

	var stdoutWriter zerolog.LevelWriter
	if cfg.Format == "json" {
		stdoutWriter = zerolog.MultiLevelWriter(os.Stdout)
	} else {
		stdoutWriter = zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime})
	}

	multi := zerolog.MultiLevelWriter(stdoutWriter, buf)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	return buf
}

func ensureRuntimeDirs(cfg *config.Config) error {
	dirs := []string{
		filepath.Dir(cfg.Database.Path),
		cfg.Masterdata.LocalPath,
		cfg.Renderer.Cache.Path,
		cfg.Assets.StickerPath,
	}
	for _, region := range config.RegionKeys() {
		profile := config.ResolveGameServerProfile(cfg, region)
		dirs = append(dirs, profile.Masterdata.LocalPath, profile.Assets.StickerPath)
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

// seedBuiltinFilterApp ensures a "moebot-builtin" downstream app exists,
// connecting the filter gateway to Moebot's own ZeroBot reverse-WS endpoint.
// This realises the "Moebot is the default built-in plugin" behaviour.
func seedBuiltinFilterApp(db *database.DB, drv config.DriverConfig) error {
	const builtinName = "moebot-builtin"
	if _, err := db.GetFilterAppByName(builtinName); err == nil {
		return nil // already present, do not overwrite user edits
	}
	listen := drv.Listen
	if listen == "" {
		listen = "127.0.0.1:6700"
	}
	if !strings.Contains(listen, ":") {
		listen = "127.0.0.1:" + listen
	}
	host, port, _ := strings.Cut(listen, ":")
	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	uri := "ws://" + host + ":" + port
	app := &models.FilterApp{
		Name:                builtinName,
		URI:                 uri,
		AccessToken:         drv.Token,
		Enabled:             true,
		Builtin:             true,
		SortOrder:           0,
		UserIDRules:         filter.EncodeIDRule(filter.IDRule{Mode: filter.ModeOn}),
		GroupIDRules:        filter.EncodeIDRule(filter.IDRule{Mode: filter.ModeOn}),
		MessageRules:        filter.EncodeMessageRule(filter.MessageRule{Mode: filter.ModeOn}),
		PrivateMessageRules: filter.EncodeMessageRule(filter.MessageRule{Mode: filter.ModeDefault}),
		GroupMessageRules:   filter.EncodeMessageRule(filter.MessageRule{Mode: filter.ModeDefault}),
	}
	return db.CreateFilterApp(app)
}

func waitForSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}

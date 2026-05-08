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

	"moebot-next/internal/bot"
	"moebot-next/internal/filter"
	"moebot-next/internal/models"
	"moebot-next/internal/plugin"

	"moebot-next/internal/config"
	"moebot-next/internal/database"
	"moebot-next/internal/logbuffer"
	"moebot-next/internal/renderer"
	"moebot-next/internal/web"

	// 内置官方插件：通过空导入触发各插件的 init() 注册到 plugin.Registry。
	// 第三方/市场插件按相同方式追加：
	//
	//   _ "github.com/FloatTech/ZeroBot-Plugin/plugin/example"
	//
	// （二期接入 zbputils/control 桥接后，上述导入即可零改动生效）
	_ "moebot-next/internal/plugins/moesekai"

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
	for {
		restart := runOnce(cfgPath)
		if !restart {
			return
		}
		log.Info().Msg("Moebot NEXT restarting (in-process)…")
		// 短暂 sleep 让操作系统回收监听端口，避免 bind error。
		time.Sleep(500 * time.Millisecond)
	}
}

// runOnce 执行一次完整的启动 / 等待信号 / 关闭流程。
// 返回 true 表示因 plugin.Registry 触发"启用插件"而需要进程内重启；
// 返回 false 表示收到 SIGINT/SIGTERM，应退出。
func runOnce(cfgPath string) bool {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	logBuffer := setupLogger(cfg.Log)

	if err := ensureCoreRuntimeDirs(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to create runtime directories")
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	rendererClient := renderer.New(cfg.Renderer)
	if err := rendererClient.StartProcess("renderer", cfg.Renderer.Port); err != nil {
		log.Warn().Err(err).Msg("Renderer process failed to start; commands will fallback to text")
	} else {
		defer rendererClient.StopProcess()
	}

	bot.RegisterMiddleware(db)

	if err := seedBuiltinFilterApp(db, cfg.Bot.Driver); err != nil {
		log.Warn().Err(err).Msg("Failed to seed builtin filter app")
	}
	filterManager := filter.New(db)
	if err := filterManager.Start(context.Background()); err != nil {
		log.Warn().Err(err).Msg("Filter gateway failed to start")
	}
	defer filterManager.Stop()

	webServer := web.New(cfg, db, nil, rendererClient, cfgPath, nil)
	webServer.Logs = logBuffer
	webServer.Filter = filterManager
	webServer.SetupStaticFiles(webUI)

	pluginDataDir := pluginsDataDir(cfg)
	if err := os.MkdirAll(pluginDataDir, 0o755); err != nil {
		log.Warn().Err(err).Str("dir", pluginDataDir).Msg("Failed to create plugins data dir")
	}
	registry := plugin.NewRegistry(db.DB, pluginDataDir)
	if err := registry.SeedDefaults(cfg.Plugins.Enabled); err != nil {
		log.Warn().Err(err).Msg("Failed to seed plugin enable defaults")
	}
	registry.InitEnabled(plugin.Context{
		Ctx:            context.Background(),
		DB:             db,
		Renderer:       rendererClient,
		Filter:         filterManager,
		Web:            webServer,
		CoreConfig:     cfg,
		CoreConfigPath: cfgPath,
	})
	defer registry.Shutdown()

	go func() {
		if err := webServer.Start(); err != nil {
			log.Error().Err(err).Msg("Web server stopped")
		}
	}()
	defer webServer.Shutdown()

	b := bot.New(cfg.Bot)
	go b.Run()

	restart := waitForSignalOrRestart(registry.RestartChan())
	if restart {
		log.Info().Msg("Plugin enabled → in-process restart requested")
	} else {
		log.Info().Msg("Moebot NEXT shutting down")
	}
	return restart
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

// ensureCoreRuntimeDirs 仅创建核心目录；插件相关目录由各插件在 Init 内自建。
func ensureCoreRuntimeDirs(cfg *config.Config) error {
	dirs := []string{
		filepath.Dir(cfg.Database.Path),
		cfg.Renderer.Cache.Path,
		pluginsDataDir(cfg),
	}
	for _, dir := range dirs {
		if dir == "" || dir == "." {
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func pluginsDataDir(cfg *config.Config) string {
	if cfg.Plugins.DataDir != "" {
		return cfg.Plugins.DataDir
	}
	return "./data/plugins"
}

// seedBuiltinFilterApp 与原行为一致：确保 "moebot-builtin" 下游 app 存在。
func seedBuiltinFilterApp(db *database.DB, drv config.DriverConfig) error {
	const builtinName = "moebot-builtin"
	if _, err := db.GetFilterAppByName(builtinName); err == nil {
		return nil
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

// waitForSignalOrRestart 阻塞直到收到 SIGINT/SIGTERM（返回 false）或
// plugin.Registry 通过 RequestRestart 触发的重启请求（返回 true）。
func waitForSignalOrRestart(restartCh <-chan struct{}) bool {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)
	select {
	case <-ch:
		return false
	case <-restartCh:
		return true
	}
}

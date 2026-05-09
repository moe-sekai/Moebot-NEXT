package main

import (
	"context"
	"embed"
	"os"
	"os/exec"
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
	_ "moebot-next/internal/plugins/autochat"
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
	restart := runOnce(cfgPath)
	if !restart {
		return
	}
	// 进程内重启不可行：ZeroBot 的 WSServer 没有 Shutdown 接口，无法释放
	// ws-reverse 监听端口；而且其命令 handler 是全局注册、无法注销。
	// 因此用重新 exec 自身的方式触发一次真正干净的进程重启。
	log.Info().Msg("Moebot NEXT restarting (re-exec)…")
	if err := reExecSelf(); err != nil {
		log.Fatal().Err(err).Msg("Failed to re-exec for restart; exiting")
	}
}

// reExecSelf 重新启动当前二进制，并在子进程成功拉起后退出父进程。
// 不使用 syscall.Exec（Windows 不支持），改为启动子进程 + os.Exit(0)。
func reExecSelf() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	// 给子进程留出一小段时间初始化（日志更清晰，也避免父进程过早退出时
	// 子进程被 TTY 关闭信号影响）。
	time.Sleep(200 * time.Millisecond)
	os.Exit(0)
	return nil
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

	// 必须在所有 /api/* 路由（核心 + 插件）注册完成之后再挂 SPA fallback，
	// 否则 SetupStaticFiles 的 NotFoundFile 会先于插件路由匹配，把 JSON 接
	// 口请求都返回成 index.html，体现为前端 /api/masterdata/summary 等被
	// 重定向到主页。
	webServer.SetupStaticFiles(webUI)

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

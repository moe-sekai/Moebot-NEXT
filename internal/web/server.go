package web

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"moebot-next/internal/backup"
	"moebot-next/internal/config"
	"moebot-next/internal/database"
	"moebot-next/internal/filter"
	"moebot-next/internal/logbuffer"
	"moebot-next/internal/plugins/moesekai/b30"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/servers"
	"moebot-next/internal/renderer"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"
)

// Server holds the Fiber app and dependencies.
type Server struct {
	App             *fiber.App
	Config          *config.Config
	ConfigPath      string
	DB              *database.DB
	Store           *masterdata.Store
	Loader          *masterdata.Loader
	Servers         *servers.Manager
	Renderer        *renderer.Client
	B30             *b30.Client
	Logs            *logbuffer.Buffer
	Filter          *filter.Manager
	Backup          *backup.Service
	BackupScheduler *backup.Scheduler
	startedAt       time.Time
}

// New creates a new web server.
func New(cfg *config.Config, db *database.DB, store *masterdata.Store, rendererClient *renderer.Client, configPath string, loader *masterdata.Loader) *Server {
	app := fiber.New(fiber.Config{
		AppName:               "Moebot NEXT",
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		},
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	server := &Server{
		App:        app,
		Config:     cfg,
		ConfigPath: configPath,
		DB:         db,
		Store:      store,
		Loader:     loader,
		Renderer:   rendererClient,
		B30:        b30.NewClient(cfg.B30),
		Backup:     backup.New(cfg, db),
		startedAt:  time.Now(),
	}

	// Register API routes
	server.registerRoutes()

	return server
}

// registerRoutes sets up all API endpoints.
func (s *Server) registerRoutes() {
	// 在 App 层注册鉴权中间件，按 path 白名单放行；这样插件后续通过
	// `app.Group("/api")` 注册的路由也会被同一个中间件保护。Fiber v2 的
	// 子 Group `Use` 不会影响其他独立 Group，因此必须挂在 App 上。
	s.App.Use("/api", s.authMiddleware)

	api := s.App.Group("/api")

	// --- Public auth & setup endpoints (无需登录) ---
	api.Get("/auth/status", s.handleAuthStatus)
	api.Post("/auth/login", s.handleLogin)
	api.Post("/setup", s.handleSetup)
	api.Get("/deployer", s.handleDeployer)
	// /api/health 同样在白名单中（authMiddleware 内判断），保留外部探活能力。
	api.Get("/health", s.handleHealth)

	// 已登录账号相关
	api.Get("/auth/me", s.handleAuthMe)
	api.Post("/auth/change-password", s.handleChangePassword)

	api.Get("/status", s.handleStatus)
	// /api/masterdata/summary is registered by the moesekai plugin.
	api.Get("/renderer/health", s.handleRendererHealth)
	api.Get("/renderer/fonts", s.handleRendererFonts)
	api.Get("/renderer/previews", s.handleRendererPreviews)
	api.Get("/renderer/previews/:id/image", s.handleRendererPreviewImage)
	api.Get("/renderer/cache/render", s.handleRenderCacheStats)
	api.Delete("/renderer/cache/render", s.handleRenderCacheClear)
	api.Put("/renderer/cache/render/config", s.handleRenderCacheConfig)
	api.Get("/renderer/budget", s.handleRendererBudgetStats)
	api.Put("/renderer/budget", s.handleRendererBudgetUpdate)
	api.Get("/commands/recent", s.handleRecentCommands)
	// /commands/definitions, /commands/parse{,/image}, /commands/aliases*,
	// /renderer/cache/card-thumbnails* are registered by the moesekai plugin
	// (internal/plugins/moesekai/webroutes/) at plugin Init time.
	api.Get("/config/public", s.handlePublicConfig)
	api.Put("/config/public", s.handleUpdatePublicConfig)
	// /api/config/sekai/test-system and /api/masterdata/reload are registered
	// by the moesekai plugin.

	// Dashboard
	api.Get("/dashboard", s.handleDashboard)

	// Groups
	api.Get("/groups", s.handleListGroups)
	api.Put("/groups/:id", s.handleUpdateGroup)
	api.Delete("/groups/:id", s.handleDeleteGroup)
	api.Get("/groups/:id/commands", s.handleGroupRecentCommands)

	// Users
	api.Get("/users", s.handleListUsers)
	api.Delete("/users/:id", s.handleDeleteUser)

	// Stats
	api.Get("/stats/commands", s.handleCommandStats)

	// Logs
	api.Get("/logs", s.handleListLogs)

	// Backup / restore
	api.Get("/backup/config", s.handleBackupConfig)
	api.Put("/backup/config", s.handleUpdateBackupConfig)
	api.Post("/backup/test", s.handleBackupTest)
	api.Get("/backup/objects", s.handleListBackups)
	api.Post("/backup", s.handleCreateBackup)
	api.Post("/backup/restore", s.handleRestoreBackup)
	api.Delete("/backup", s.handleDeleteBackup)

	// Filter (OneBot gateway)
	api.Get("/filter/status", s.handleFilterStatus)
	api.Get("/filter/gateway", s.handleGetFilterGateway)
	api.Put("/filter/gateway", s.handleUpdateFilterGateway)
	api.Get("/filter/apps", s.handleListFilterApps)
	api.Post("/filter/apps", s.handleCreateFilterApp)
	api.Put("/filter/apps/:id", s.handleUpdateFilterApp)
	api.Delete("/filter/apps/:id", s.handleDeleteFilterApp)
	api.Get("/filter/templates", s.handleListFilterTemplates)
	api.Post("/filter/templates", s.handleCreateFilterTemplate)
	api.Get("/filter/templates/:id", s.handleGetFilterTemplate)
	api.Put("/filter/templates/:id", s.handleUpdateFilterTemplate)
	api.Delete("/filter/templates/:id", s.handleDeleteFilterTemplate)
	api.Get("/filter/events", s.handleFilterEvents)
	api.Get("/filter/events/recent", s.handleFilterRecentEvents)
	api.Post("/filter/test-regex", s.handleFilterTestRegex)
	api.Get("/filter/export-yaml", s.handleFilterExportYAML)
	api.Post("/filter/import-yaml", s.handleFilterImportYAML)

	// /api/search/* is registered by the moesekai plugin (see
	// internal/plugins/moesekai/webroutes/search.go).

	// Plugins
	api.Get("/plugins", s.handleListPlugins)
	api.Get("/plugins/market", s.handleListMarketPlugins)
	api.Post("/plugins/:name/enable", s.handleSetPluginEnabled(true))
	api.Post("/plugins/:name/disable", s.handleSetPluginEnabled(false))
	api.Get("/plugins/:name/config", s.handleGetPluginConfig)
	api.Put("/plugins/:name/config", s.handleUpdatePluginConfig)
	api.Get("/plugins/:name/settings", s.handleGetPluginSettings)
	api.Put("/plugins/:name/settings", s.handleUpdatePluginSettings)

}

// SetupStaticFiles configures serving the embedded Vue SPA.
func (s *Server) SetupStaticFiles(webFS embed.FS) {
	s.App.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(webFS),
		PathPrefix:   "web/dist",
		Index:        "index.html",
		NotFoundFile: "web/dist/index.html",
	}))
}

// Start begins listening on the configured host:port.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.Config.Web.Host, s.Config.Web.Port)
	log.Info().Str("addr", addr).Msg("Web server starting")
	return s.App.Listen(addr)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() error {
	return s.App.Shutdown()
}

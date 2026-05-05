package web

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/database"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/renderer"
	"moebot-next/internal/servers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"
)

// Server holds the Fiber app and dependencies.
type Server struct {
	App        *fiber.App
	Config     *config.Config
	ConfigPath string
	DB         *database.DB
	Store      *masterdata.Store
	Loader     *masterdata.Loader
	Servers    *servers.Manager
	Renderer   *renderer.Client
	startedAt  time.Time
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
		startedAt:  time.Now(),
	}

	// Register API routes
	server.registerRoutes()

	return server
}

// registerRoutes sets up all API endpoints.
func (s *Server) registerRoutes() {
	api := s.App.Group("/api")

	// Health and runtime status
	api.Get("/health", s.handleHealth)
	api.Get("/status", s.handleStatus)
	api.Get("/masterdata/summary", s.handleMasterdataSummary)
	api.Get("/renderer/health", s.handleRendererHealth)
	api.Get("/renderer/cache/card-thumbnails", s.handleRendererCardThumbnailCacheStatus)
	api.Post("/renderer/cache/card-thumbnails/preload", s.handleRendererCardThumbnailPreload)
	api.Get("/renderer/previews", s.handleRendererPreviews)
	api.Get("/renderer/previews/:id/image", s.handleRendererPreviewImage)
	api.Get("/commands/recent", s.handleRecentCommands)
	api.Get("/commands/definitions", s.handleCommandDefinitions)
	api.Get("/commands/parse", s.handleParseCommand)
	api.Get("/commands/parse/image", s.handleParseCommandImage)
	api.Get("/commands/aliases", s.handleGetCommandAliases)
	api.Put("/commands/aliases", s.handleUpdateCommandAliases)
	api.Post("/commands/aliases/reset", s.handleResetCommandAliases)
	api.Get("/commands/aliases/export", s.handleExportCommandAliases)
	api.Post("/commands/aliases/import", s.handleImportCommandAliases)
	api.Get("/config/public", s.handlePublicConfig)
	api.Put("/config/public", s.handleUpdatePublicConfig)
	api.Post("/config/sekai/test-system", s.handleTestSekaiSystem)
	api.Post("/masterdata/reload", s.handleReloadMasterdata)

	// Dashboard
	api.Get("/dashboard", s.handleDashboard)

	// Groups
	api.Get("/groups", s.handleListGroups)
	api.Put("/groups/:id", s.handleUpdateGroup)

	// Users
	api.Get("/users", s.handleListUsers)
	api.Delete("/users/:id", s.handleDeleteUser)

	// Stats
	api.Get("/stats/commands", s.handleCommandStats)

	// Masterdata search
	api.Get("/search/cards", s.handleSearchCards)
	api.Get("/search/musics", s.handleSearchMusics)
	api.Get("/search/events", s.handleSearchEvents)
	api.Get("/search/gachas", s.handleSearchGachas)
	api.Get("/search/virtual-lives", s.handleSearchVirtualLives)

	// TODO: auth middleware, settings, renderer preview, WebSocket
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

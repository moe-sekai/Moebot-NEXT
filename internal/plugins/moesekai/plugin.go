package moesekai

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"moebot-next/internal/config"
	"moebot-next/internal/database"
	"moebot-next/internal/plugin"
	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/plugins/moesekai/b30"
	"moebot-next/internal/plugins/moesekai/commandparser"
	"moebot-next/internal/plugins/moesekai/commands"
	"moebot-next/internal/plugins/moesekai/servers"
	"moebot-next/internal/plugins/moesekai/webroutes"
	"moebot-next/internal/renderer"
	"moebot-next/internal/web"

	"github.com/rs/zerolog/log"
)

const PluginName = "moesekai"

// pluginImpl implements plugin.Plugin for the official MoeSekai plugin.
type pluginImpl struct{}

// Manifest returns metadata exposed to the WebUI plugin list.
func (p *pluginImpl) Manifest() plugin.Manifest {
	return plugin.Manifest{
		Name:          PluginName,
		Title:         "MoeSekai (Project Sekai)",
		Version:       "0.1.0",
		Author:        "Moebot Team",
		Category:      plugin.CategoryOfficial,
		Description:   "Project Sekai business: card / music / gacha / suite / deck recommend / ranking / B30 etc.",
		Homepage:      "https://github.com/moe-sekai/Moebot-NEXT",
		SettingsRoute: "/plugins/moesekai",
		Tags:          []string{"pjsk", "official"},
	}
}

// Init runs the PJSK boot sequence when the plugin is enabled:
//  1. Read the plugin sub-config and merge into the core config.
//  2. Ensure PJSK runtime directories.
//  3. Configure assets + servers + B30 + register all chat commands.
//  4. Attach servers/store/loader/B30 onto the web.Server.
func (p *pluginImpl) Init(ctx *plugin.Context) error {
	cfg, ok := ctx.CoreConfig.(*config.Config)
	if !ok || cfg == nil {
		return errors.New("moesekai: missing core config in plugin context")
	}
	db, _ := ctx.DB.(*database.DB)
	rendererClient, _ := ctx.Renderer.(*renderer.Client)
	webServer, _ := ctx.Web.(*web.Server)

	// 1) Load and merge sub-config (missing file falls back to core defaults).
	var sub Config
	if err := plugin.ReadYAMLInto(ctx.PluginConfigPath, &sub); err != nil {
		log.Warn().Err(err).Str("path", ctx.PluginConfigPath).Msg("moesekai: failed to read plugin config, using defaults")
	}
	sub.applyTo(cfg)

	// 2) Ensure PJSK runtime directories exist.
	if err := ensureMoesekaiDirs(cfg); err != nil {
		log.Warn().Err(err).Msg("moesekai: failed to create runtime directories")
	}

	// 3) Asset CDN.
	if _, err := assets.Configure(cfg.Assets, cfg.Server.Region); err != nil {
		log.Warn().Err(err).Msg("moesekai: asset CDN config invalid; using built-in default")
	}

	// 4) Per-server runtimes.
	serverManager := servers.NewManager(cfg)
	serverManager.LoadEnabled()
	serverManager.StartPeriodicRefresh()
	ctx.OnShutdown(serverManager.StopPeriodicRefresh)

	// 5) B30 client.
	b30Client := b30.NewClient(cfg.B30)

	// 6) Command parser definitions (with custom aliases).
	defs := commandparser.Definitions(cfg.Bot.CommandAliases)

	// 7) Register chat handlers (must run before zero.RunAndBlock).
	if db != nil && rendererClient != nil {
		commands.RegisterAll(&commands.Deps{
			DB:          db,
			Renderer:    rendererClient,
			Servers:     serverManager,
			B30:         b30Client,
			Definitions: defs,
		})
	} else {
		log.Warn().Msg("moesekai: db/renderer not ready, commands not registered")
	}

	// 8) Attach PJSK resources back onto the shared web server (still used by
	// the dashboard / search / status handlers in internal/web/handlers.go).
	// 9) Register PJSK-owned web routes via the moesekai webroutes package so
	// they no longer pollute internal/web.
	if webServer != nil {
		webServer.Servers = serverManager
		defaultRuntime := serverManager.Default()
		if defaultRuntime != nil {
			webServer.Store = defaultRuntime.Store
			webServer.Loader = defaultRuntime.Loader
		}
		webServer.B30 = b30Client

		deps := webroutes.Deps{
			Config:     cfg,
			ConfigPath: ctx.CoreConfigPath,
			Renderer:   rendererClient,
			Servers:    serverManager,
			Store:      webServer.Store,
			B30:        b30Client,
			SaveConfig: func() error { return config.Save(cfg, ctx.CoreConfigPath) },
		}
		api := webServer.App.Group("/api")
		webroutes.RegisterCommandParser(api, deps)
		webroutes.RegisterRendererCache(api, deps)
	}
	return nil
}

// ensureMoesekaiDirs replicates the legacy main.go behaviour for PJSK dirs.
func ensureMoesekaiDirs(cfg *config.Config) error {
	dirs := []string{
		cfg.Masterdata.LocalPath,
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
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	return nil
}

// SaveConfigFile persists the moesekai sub-config back to disk. Used by the
// WebUI write path.
func SaveConfigFile(path string, sub *Config) error {
	if path == "" {
		return errors.New("moesekai: empty config path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return plugin.WriteYAMLFrom(path, sub)
}

func init() {
	plugin.Register(&pluginImpl{})
}

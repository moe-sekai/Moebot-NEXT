package moesekai

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"moebot-next/internal/config"
	"moebot-next/internal/database"
	"moebot-next/internal/filter"
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
	zerobot "github.com/wdvxdr1123/ZeroBot"
)

const PluginName = "moesekai"

// pluginImpl implements plugin.Plugin (and plugin.Configurable) for the
// official MoeSekai plugin. Init 后会缓存 ctx/cfg 引用，供 schema 读写使用。
type pluginImpl struct {
	mu         sync.RWMutex
	cfg        *config.Config // shared core config (live)
	configPath string         // data/plugins/moesekai.yml

	// sharedDeps 被 Fiber 路由 handler 以指针方式捕获；每次 Init 时
	// 原地更新其字段，禁用时清空关键字段。routesOnce 确保 Fiber 路由只
	// 注册一次（Fiber 不支持路由注销，重复 Register 会叠加 handler）。
	sharedDeps *webroutes.Deps
	routesOnce sync.Once
}

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
		Settings: []plugin.SettingField{
			{
				Key:         "region",
				Label:       "默认游戏服 (region)",
				Type:        "select",
				Group:       "基础",
				Description: "默认查询使用的游戏服，影响 masterdata / sekai_api 等。",
				Options: []plugin.SettingChoice{
					{Label: "日服 (JP)", Value: config.RegionJP},
					{Label: "国服 (CN)", Value: config.RegionCN},
					{Label: "台服 (TW)", Value: config.RegionTW},
					{Label: "韩服 (KR)", Value: config.RegionKR},
					{Label: "国际服 (EN)", Value: config.RegionEN},
				},
			},
		},
	}
}

// GetSettings 实现 plugin.Configurable，按 Manifest.Settings 中声明的 Key
// 返回当前生效值（优先取内存里的 *config.Config，回退插件 yaml）。
func (p *pluginImpl) GetSettings() (map[string]any, error) {
	p.mu.RLock()
	cfg := p.cfg
	p.mu.RUnlock()
	out := map[string]any{}
	if cfg != nil {
		out["region"] = cfg.Server.Region
	}
	return out, nil
}

// UpdateSettings 实现 plugin.Configurable，按 key 写回内存配置 + 插件 yaml。
// 未识别的 key 会被忽略。
func (p *pluginImpl) UpdateSettings(values map[string]any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cfg == nil || p.configPath == "" {
		return errors.New("moesekai: plugin not initialized")
	}
	// 读现有 yaml 作为基线，避免覆盖未管理字段。
	var sub Config
	_ = plugin.ReadYAMLInto(p.configPath, &sub)

	if v, ok := values["region"]; ok {
		region, ok := v.(string)
		if !ok || region == "" {
			return fmt.Errorf("moesekai: invalid region value %v", v)
		}
		p.cfg.Server.Region = region
		sub.Region = region
	}
	config.NormalizeConfig(p.cfg)
	return SaveConfigFile(p.configPath, &sub)
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
	filterMgr, _ := ctx.Filter.(*filter.Manager)

	// 在 Filter 网关中 seed 本插件对应的 internal app；让控制台「Filter」
	// 页面能为 PJSK 业务独立分配模板/规则（与 autochat 解耦）。
	if db != nil {
		if err := filter.SeedInternalApp(db, PluginName, "MoeSekai"); err != nil {
			log.Warn().Err(err).Msg("moesekai: 创建 internal filter app 失败")
		} else if filterMgr != nil && filterMgr.IsRunning() {
			_ = filterMgr.Reload(ctx.Ctx)
		}
	}

	// 1) Load and merge sub-config (missing file falls back to core defaults).
	var sub Config
	if err := plugin.ReadYAMLInto(ctx.PluginConfigPath, &sub); err != nil {
		log.Warn().Err(err).Str("path", ctx.PluginConfigPath).Msg("moesekai: failed to read plugin config, using defaults")
	}
	sub.applyTo(cfg)

	// 缓存引用以支持 plugin.Configurable 的 GetSettings/UpdateSettings。
	p.mu.Lock()
	p.cfg = cfg
	p.configPath = ctx.PluginConfigPath
	p.mu.Unlock()

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
	// 插件禁用时同步清理 webServer 上的 moesekai 资源指针，避免 Dashboard
	// 误报"已加载"。重新启用时 Init 会重新赋值。
	if webServer, ok := ctx.Web.(*web.Server); ok && webServer != nil {
		ctx.OnShutdown(func() {
			webServer.Servers = nil
			webServer.Store = nil
			webServer.Loader = nil
			webServer.B30 = nil
		})
	}

	// 5) B30 client.
	b30Client := b30.NewClient(cfg.B30)

	// 6) Command parser definitions (with custom aliases).
	defs := commandparser.Definitions(cfg.Bot.CommandAliases)

	// 7) Register chat handlers (must run before zero.RunAndBlock).
	if db != nil && rendererClient != nil {
		// plugin:moesekai 这一层的统一前置过滤。挂在 Engine.UsePreHandler 上，
		// 一处覆盖所有命令；返回 false 时整个事件被本插件忽略。
		// 仅过滤 message 类事件，notice/request/meta 等放行。
		filterAppName := filter.InternalAppName(PluginName)
		var preHandler commands.PreHandler
		if filterMgr != nil {
			preHandler = func(zctx *zerobot.Ctx) bool {
				ev := zctx.Event
				if ev == nil || ev.PostType != "message" {
					return true
				}
				isPrivate := ev.MessageType == "private" || ev.DetailType == "private"
				return filterMgr.AllowMessage(filterAppName, ev.GroupID, ev.UserID, isPrivate, ev.RawMessage)
			}
		}
		commands.RegisterAll(&commands.Deps{
			DB:          db,
			Renderer:    rendererClient,
			Servers:     serverManager,
			B30:         b30Client,
			Definitions: defs,
			PreHandler:  preHandler,
		})
		// 插件禁用时清理已注册的 ZeroBot matcher，避免重启后重复触发。
		ctx.OnShutdown(commands.ResetEngine)
	} else {
		log.Warn().Msg("moesekai: db/renderer not ready, commands not registered")
	}

	// 8) Attach PJSK resources back onto the shared web server (still used by
	// the dashboard / search / status handlers in internal/web/handlers.go).
	// 9) Register PJSK-owned web routes via the moesekai webroutes package so
	// they no longer pollute internal/web. Routes are registered ONLY ONCE
	// per process — Fiber 不支持注销路由，重复 Register 会叠加 handler。
	// 重新启用时通过更新 sharedDeps 字段让原 handler 看到最新依赖。
	if webServer != nil {
		webServer.Servers = serverManager
		defaultRuntime := serverManager.Default()
		if defaultRuntime != nil {
			webServer.Store = defaultRuntime.Store
			webServer.Loader = defaultRuntime.Loader
		}
		webServer.B30 = b30Client

		p.mu.Lock()
		if p.sharedDeps == nil {
			p.sharedDeps = &webroutes.Deps{}
		}
		p.sharedDeps.Config = cfg
		p.sharedDeps.ConfigPath = ctx.CoreConfigPath
		p.sharedDeps.Renderer = rendererClient
		p.sharedDeps.Servers = serverManager
		p.sharedDeps.Store = webServer.Store
		p.sharedDeps.Loader = webServer.Loader
		p.sharedDeps.B30 = b30Client
		p.sharedDeps.SaveConfig = func() error { return config.Save(cfg, ctx.CoreConfigPath) }
		shared := p.sharedDeps
		p.mu.Unlock()

		p.routesOnce.Do(func() {
			api := webServer.App.Group("/api")
			webroutes.RegisterCommandParser(api, shared)
			webroutes.RegisterRendererCache(api, shared)
			webroutes.RegisterSearch(api, shared)
			webroutes.RegisterSekaiTest(api, shared)
			webroutes.RegisterMasterdata(api, shared)
		})

		// 禁用时把 sharedDeps 中可观测字段清零，让已注册的 handler 返回
		// "service unavailable" 而不是命中陈旧 store/Servers。
		ctx.OnShutdown(func() {
			p.mu.Lock()
			if p.sharedDeps != nil {
				p.sharedDeps.Servers = nil
				p.sharedDeps.Store = nil
				p.sharedDeps.Loader = nil
				p.sharedDeps.B30 = nil
			}
			p.mu.Unlock()
		})
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

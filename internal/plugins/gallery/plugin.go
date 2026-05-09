package gallery

import (
	"errors"
	"os"
	"sync"

	"moebot-next/internal/database"
	"moebot-next/internal/plugin"
	"moebot-next/internal/web"

	"github.com/rs/zerolog/log"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const PluginName = "gallery"

type pluginImpl struct {
	mu         sync.RWMutex
	routesOnce sync.Once
	configPath string
	mgr        *GalleryManager
	engine     *zero.Engine
}

func (p *pluginImpl) Manifest() plugin.Manifest {
	return plugin.Manifest{
		Name:          PluginName,
		Title:         "Gallery (图片画廊)",
		Version:       "0.1.0",
		Author:        "Moebot Team",
		Category:      plugin.CategoryOfficial,
		Description:   "图片画廊管理插件，支持上传/查看/去重/撤销等功能。",
		Homepage:      "https://github.com/moe-sekai/Moebot-NEXT",
		SettingsRoute: "/plugins/gallery",
		Tags:          []string{"gallery", "official"},
		Settings: []plugin.SettingField{
			{Key: "size_limit_mb", Label: "图片大小限制(MB)", Type: "int", Default: 1, Group: "基本"},
			{Key: "pick_limit", Label: "单次查看上限", Type: "int", Default: 5, Group: "基本"},
			{Key: "hash1_difference_threshold", Label: "哈希1差异阈值", Type: "int", Default: 5, Group: "去重",
				Description: "感知哈希汉明距离阈值，越大越宽松"},
			{Key: "hash2_difference_threshold", Label: "哈希2差异阈值", Type: "int", Default: 1000, Group: "去重",
				Description: "灰度像素MAE阈值，越大越宽松"},
			{Key: "revert_expired_hours", Label: "撤销过期(小时)", Type: "int", Default: 24, Group: "基本"},
			{Key: "data_dir", Label: "数据目录", Type: "string", Default: "data/gallery", Group: "基本"},
		},
	}
}

func (p *pluginImpl) GetSettings() (map[string]any, error) {
	c := getConfig()
	if c == nil {
		return map[string]any{}, nil
	}
	return map[string]any{
		"size_limit_mb":              c.SizeLimitMB,
		"pick_limit":                 c.PickLimit,
		"hash1_difference_threshold": c.Hash1DifferenceThreshold,
		"hash2_difference_threshold": c.Hash2DifferenceThreshold,
		"revert_expired_hours":       c.RevertExpiredHours,
		"data_dir":                   c.DataDir,
	}, nil
}

func (p *pluginImpl) UpdateSettings(values map[string]any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	c := getConfig()
	if c == nil {
		return errors.New("gallery: config not loaded")
	}
	if v, ok := values["size_limit_mb"]; ok {
		if n, ok := toInt(v); ok {
			c.SizeLimitMB = n
		}
	}
	if v, ok := values["pick_limit"]; ok {
		if n, ok := toInt(v); ok {
			c.PickLimit = n
		}
	}
	if v, ok := values["hash1_difference_threshold"]; ok {
		if n, ok := toInt(v); ok {
			c.Hash1DifferenceThreshold = n
		}
	}
	if v, ok := values["hash2_difference_threshold"]; ok {
		if n, ok := toInt(v); ok {
			c.Hash2DifferenceThreshold = n
		}
	}
	if v, ok := values["revert_expired_hours"]; ok {
		if n, ok := toInt(v); ok {
			c.RevertExpiredHours = n
		}
	}
	if v, ok := values["data_dir"]; ok {
		if s, ok := v.(string); ok && s != "" {
			c.DataDir = s
		}
	}
	if p.mgr != nil {
		p.mgr.cfg = c
	}
	return plugin.WriteYAMLFrom(p.configPath, c)
}

func (p *pluginImpl) Init(ctx *plugin.Context) error {
	db, _ := ctx.DB.(*database.DB)
	if db == nil || db.DB == nil {
		return errors.New("gallery: database not available")
	}

	// AutoMigrate 画廊表
	if err := db.DB.AutoMigrate(&GalleryInfo{}, &GalleryPic{}, &GalleryUploadRecord{}); err != nil {
		return err
	}

	// 读取配置
	var c Config
	if err := plugin.ReadYAMLInto(ctx.PluginConfigPath, &c); err != nil {
		log.Warn().Err(err).Msg("[gallery] 读取配置失败，使用默认值")
	}
	applyDefaults(&c)
	if _, err := os.Stat(ctx.PluginConfigPath); os.IsNotExist(err) {
		_ = plugin.WriteYAMLFrom(ctx.PluginConfigPath, &c)
	}
	setConfig(&c)
	p.configPath = ctx.PluginConfigPath

	// 确保数据目录
	if err := os.MkdirAll(c.DataDir, 0o755); err != nil {
		log.Warn().Err(err).Str("dir", c.DataDir).Msg("[gallery] 创建数据目录失败")
	}

	// 创建 Manager
	p.mgr = NewGalleryManager(db.DB, &c)

	// 注册 ZeroBot 命令
	p.engine = p.registerHandlers()

	// 注册 Web 路由
	if webServer, ok := ctx.Web.(*web.Server); ok && webServer != nil {
		p.routesOnce.Do(func() {
			p.registerWebRoutes(webServer.App.Group("/api"))
		})
	}

	// 关闭钩子
	ctx.OnShutdown(func() {
		if p.engine != nil {
			p.engine.Delete()
			p.engine = nil
		}
		log.Info().Msg("[gallery] 已停止")
	})

	log.Info().Str("data_dir", c.DataDir).Msg("[gallery] 已启动")
	return nil
}

func toInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	case string:
		return 0, false
	}
	return 0, false
}

func init() {
	plugin.Register(&pluginImpl{})
}

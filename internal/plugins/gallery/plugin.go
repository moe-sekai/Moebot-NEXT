package gallery

import (
	"errors"
	"os"
	"sync"

	"moebot-next/internal/database"
	"moebot-next/internal/filter"
	"moebot-next/internal/plugin"
	"moebot-next/internal/renderer"
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
	filterMgr  *filter.Manager  // 用于查询本插件的 internal FilterApp 规则
	rendererCl *renderer.Client // 调用 satori 渲染服务（/看所有 拼图）
}

// filterAppName 返回本插件在 filter 网关中的 internal app 名字。
func (p *pluginImpl) filterAppName() string { return filter.InternalAppName(PluginName) }

// allowedByFilter 查询 filter 网关：当前消息是否被本插件的 internal app 放行。
// 当 filter 未启用 / 该 app 未 seed 时返回 true（不阻塞）。
func (p *pluginImpl) allowedByFilter(groupID, userID int64, isPrivate bool, raw string) bool {
	if p.filterMgr == nil {
		return true
	}
	return p.filterMgr.AllowMessage(p.filterAppName(), groupID, userID, isPrivate, raw)
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
	filterMgr, _ := ctx.Filter.(*filter.Manager)

	// 在 Filter 网关中 seed 本插件对应的 internal app；让控制台「Filter」
	// 页面能够独立分配模板/规则。已存在时不覆盖用户配置。
	if err := filter.SeedInternalApp(db, PluginName, "Gallery (图片画廊)"); err != nil {
		log.Warn().Err(err).Msg("[gallery] 创建 internal filter app 失败")
	} else if filterMgr != nil && filterMgr.IsRunning() {
		_ = filterMgr.Reload(ctx.Ctx)
	}

	// 旧表迁移：早期版本 GORM 默认将 PID 字段映射成 p_id 列，导致后续手写
	// SQL（Where/Order 用 "pid"）全部报 "no such column: pid"。
	// 在 AutoMigrate 之前先把 p_id 列改名为 pid（仅当 pid 列不存在时执行）。
	mig := db.DB.Migrator()
	if mig.HasTable(&GalleryPic{}) && mig.HasColumn(&GalleryPic{}, "p_id") && !mig.HasColumn(&GalleryPic{}, "pid") {
		if err := mig.RenameColumn(&GalleryPic{}, "p_id", "pid"); err != nil {
			log.Warn().Err(err).Msg("[gallery] 迁移 gallery_pics.p_id -> pid 失败")
		} else {
			log.Info().Msg("[gallery] 已将旧表 gallery_pics.p_id 重命名为 pid")
		}
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

	// 保存 filter manager 引用，用于消息处理时查询规则
	rendererClient, _ := ctx.Renderer.(*renderer.Client)
	p.mu.Lock()
	p.filterMgr = filterMgr
	p.rendererCl = rendererClient
	p.mu.Unlock()

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

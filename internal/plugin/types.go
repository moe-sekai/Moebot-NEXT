// Package plugin 提供 Moebot NEXT 的轻量插件框架。
//
// 设计目标：
//   - 让官方插件（如 moesekai）以"业务+数据"整体打包注册
//   - 给 WebUI 暴露统一的 manifest / 启用开关 / 设置入口
//   - 后续可桥接 FloatTech/zbputils/control 让上游 ZeroBot-Plugin 仓库
//     里的插件零改动 import 即可注册（二期目标）
package plugin

import (
	"context"
)

// Category 区分插件来源，用于 WebUI 区分官方 / 第三方 / 市场。
type Category string

const (
	CategoryOfficial Category = "official" // 官方内置
	CategoryMarket   Category = "market"   // 插件市场（FloatTech 上游等）
	CategoryThird    Category = "third"    // 第三方/本地开发
)

// SettingField 描述一项可在 WebUI 表单中呈现的设置项。
// 字段会被 JSON 化给前端，前端按 type 渲染。
type SettingField struct {
	Key         string         `json:"key"`
	Label       string         `json:"label"`
	Type        string         `json:"type"`        // string / int / bool / select / textarea
	Default     any            `json:"default,omitempty"`
	Description string         `json:"description,omitempty"`
	Group       string         `json:"group,omitempty"`
	Options     []SettingChoice `json:"options,omitempty"`
}

// SettingChoice 为 select 类型字段提供选项。
type SettingChoice struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Manifest 描述一个插件的元信息，用于 WebUI 展示。
type Manifest struct {
	Name        string         `json:"name"`               // 唯一标识（kebab-case 或 snake_case）
	Title       string         `json:"title"`              // 显示名
	Version     string         `json:"version"`            // 语义化版本
	Author      string         `json:"author,omitempty"`
	Category    Category       `json:"category"`
	Description string         `json:"description,omitempty"`
	Repo        string         `json:"repo,omitempty"`     // 仓库链接，用于"插件市场"
	Homepage    string         `json:"homepage,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	// SettingsRoute 指定 WebUI 中跳转到该插件设置页的路径，例如 "/plugins/moesekai"。
	// 留空表示插件没有专属设置页。
	SettingsRoute string         `json:"settings_route,omitempty"`
	Settings      []SettingField `json:"settings,omitempty"`
}

// Plugin 是所有插件需实现的最小接口。
type Plugin interface {
	// Manifest 返回插件元信息。多次调用应返回稳定的副本。
	Manifest() Manifest
	// Init 在插件启用时调用一次。返回 error 表示加载失败，框架会日志告警，
	// 但不会中断其它插件加载。Init 内部应：
	//   - 读取自己的子配置文件 (Context.PluginConfigPath)
	//   - 注册 ZeroBot 处理器 / Web 路由 / 周期任务等
	//   - 用 Context.OnShutdown 注册关闭钩子
	Init(ctx *Context) error
}

// Context 是插件初始化时拿到的依赖容器。
//
// 字段使用 any 包裹避免在 internal/plugin 包里反向引用业务子包，从而不
// 引入循环依赖。具体类型由插件自行 type assert（约定见 README）。
type Context struct {
	Ctx context.Context

	// 全局核心依赖
	DB         any // *database.DB
	Renderer   any // *renderer.Client
	Filter     any // *filter.Manager
	Web        any // *web.Server
	Bot        any // *bot.Bot

	// 插件相关路径
	PluginName       string // 插件名（用于日志）
	PluginConfigPath string // data/plugins/<name>.yml 的绝对路径
	PluginDataDir    string // data/plugins/ 的绝对路径
	CoreConfigPath   string // data/config.yml

	// CoreConfig 是核心配置的不透明指针 (*config.Config)，插件可读写。
	CoreConfig any

	shutdown []func()
}

// OnShutdown 注册一个在 moebot 关闭时触发的钩子。
// 钩子按注册逆序执行。
func (c *Context) OnShutdown(fn func()) {
	if fn == nil {
		return
	}
	c.shutdown = append(c.shutdown, fn)
}

// RunShutdownHooks 由插件管理器在主进程退出时调用。
func (c *Context) RunShutdownHooks() {
	for i := len(c.shutdown) - 1; i >= 0; i-- {
		func() {
			defer func() { _ = recover() }()
			c.shutdown[i]()
		}()
	}
	c.shutdown = nil
}

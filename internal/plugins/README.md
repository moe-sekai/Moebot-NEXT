# Moebot NEXT 插件子系统

本目录托管 Moebot NEXT 的官方内置插件，第三方/市场插件可放在外部仓库
并以 `import _` 方式接入。

## 整体架构

- `internal/plugin/`：插件框架核心
  - `Manifest`：元信息（name / version / category / settings_route / ...）
  - `Plugin` 接口：`Manifest()` + `Init(ctx)`
  - `Registry`：注册表 + 启用状态（持久化在 SQLite `plugin_states` 表）
  - `Context`：传给 `Init` 的依赖容器（DB / Renderer / Filter / Web /
    CoreConfig / PluginConfigPath）
  - `ReadYAMLInto` / `WriteYAMLFrom`：插件子配置的 YAML 工具
- `internal/plugins/<name>/`：单个插件的实现包
  - 通过 `init()` 内调用 `plugin.Register(...)` 完成自注册
  - 在 `Init(ctx)` 中读取 `<data_dir>/<name>.yml`、注册 ZeroBot 处理器、
    挂资源到 `web.Server`、登记关闭钩子等

启动流程（`main.go`）：

1. 加载核心 `config.yml`（精简后只剩 bot/web/database/renderer/log/plugins）
2. 启动 db / renderer / filter / bot / web 等核心子系统
3. 构造 `plugin.Registry`，根据 `plugins.enabled` 种子化默认启用列表
4. `Registry.InitEnabled` 调用所有启用插件的 `Init`
5. `bot.Run` 进入消息循环

## 插件如何启用 / 禁用

- 控制台「插件」页提供启用开关（`POST /api/plugins/:name/enable|disable`）。
- 启用状态写入 SQLite `plugin_states` 表 → 重启 moebot 后才会真正加/卸载。
- 控制台「插件市场」页列出官方与上游 ZeroBot-Plugin 仓库的精选插件，
  当前为编译期选入模型：要"安装"市场插件，需在 `main.go` 加上对应
  `_ "github.com/.../plugin/xxx"` 导入并重新构建。

## 写一个官方插件（参考 moesekai）

```go
package myplugin

import (
    "moebot-next/internal/plugin"
)

type myPlugin struct{}

func (p *myPlugin) Manifest() plugin.Manifest {
    return plugin.Manifest{
        Name:          "myplugin",
        Title:         "示例插件",
        Version:       "0.1.0",
        Category:      plugin.CategoryOfficial,
        SettingsRoute: "/plugins/myplugin",
    }
}

func (p *myPlugin) Init(ctx *plugin.Context) error {
    var cfg MyConfig
    if err := plugin.ReadYAMLInto(ctx.PluginConfigPath, &cfg); err != nil {
        return err
    }
    // ... 注册 ZeroBot 处理器、挂 web 路由、StartTickers ...
    ctx.OnShutdown(func() { /* cleanup */ })
    return nil
}

func init() { plugin.Register(&myPlugin{}) }
```

然后在 `main.go` 里 `_ "moebot-next/internal/plugins/myplugin"` 即可被注册。

## 接入 FloatTech ZeroBot-Plugin

桥接代码位于 `@d:/Python_Project/Moebot-NEXT-Go/internal/plugin/control_bridge.go`，
以 `floatcontrol` build tag 隔离。启用步骤：

```bash
go get github.com/FloatTech/zbputils/control@latest
go build -tags floatcontrol ./...
```

启用后，所有通过 `control.Register(...)` 注册的上游插件都会自动出现在
Moebot 的「插件」页面，category 标记为 `market`，启用/禁用走 Moebot 的
`plugin_states` 表。在 `main.go` 里添加 `import _ "github.com/.../plugin/<name>"`
即可将上游包编译进二进制。

默认（不带 tag）的二进制不依赖 `zbputils`，对 `wdvxdr1123/ZeroBot v1.8` 的
版本固定不会引入冲突。

## 插件注册 Web 路由

PJSK 业务相关的 HTTP 路由统一放在
`@d:/Python_Project/Moebot-NEXT-Go/internal/plugins/moesekai/webroutes/`
（package `webroutes`），通过插件 Init 时拿到 `*web.Server.App`，构造
`webroutes.Deps`，挂到 `/api` 上：

```go
api := webServer.App.Group("/api")
deps := webroutes.Deps{Config: cfg, Renderer: rendererClient, ...}
webroutes.RegisterCommandParser(api, deps)
webroutes.RegisterRendererCache(api, deps)
```

已迁出（核心 `internal/web` 不再持有这些路由代码）：

- `/api/commands/definitions`、`/api/commands/parse{,/image}`、
  `/api/commands/aliases*`
- `/api/renderer/cache/card-thumbnails*`
- `/api/search/{cards,musics,events,gachas,virtual-lives}`
- `/api/config/sekai/test-system`
- `/api/masterdata/{summary,reload}`

仍留在 `internal/web/handlers.go` 的路由（含设计动机，**不应**机械搬迁）：

- `/api/dashboard`、`/api/status`、`/api/renderer/health`、
  `/api/renderer/previews*` —— 核心运维 / 渲染服务关注点，不属于 PJSK 业务
- `/api/config/public`（GET + PUT）—— **混合**核心字段（`server.region`、
  `bot.*`、`renderer.*`）与 PJSK 字段（`masterdata` / `assets` /
  `sekai_api` / `suite_api` / `ranking_api` / `b30` / `game_servers`）；
  整体搬迁会让 moesekai 路由越权修改核心配置。后续应改为"插件向核心注册
  配置 schema 片段 + apply 钩子"，让核心 `publicConfig` 自动聚合每个插件
  贡献的字段（设计待定）

`internal/web.Server` 仍持有 `Servers / Store / Loader / B30` 字段，是
为了让上述未迁的 status / publicConfig handlers 能直接读取 PJSK 状态。
当 publicConfig 完成 schema 注入式重构后，这些字段可一并撤掉。

## 兼容旧 config.yml

为了平滑迁移，moesekai 插件支持两种配置来源：

1. `data/plugins/moesekai.yml`（推荐，新结构）
2. 历史 `config.yml` 中的 `masterdata / sekai_api / suite_api / ranking_api /
   b30 / assets / game_servers / bot.command_aliases` 字段（兼容路径）

两者并存时，插件子配置优先（在 `Init` 中通过 `applyTo` 覆写核心 `*config.Config`）。

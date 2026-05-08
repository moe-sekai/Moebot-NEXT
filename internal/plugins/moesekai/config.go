// Package moesekai 是 Moebot NEXT 内置的 Project Sekai 业务插件。
//
// 本包不再持有 PJSK 子模块的实现（这些仍存在于 internal/{commands,b30,...}
// 等历史目录下，二期会做物理搬迁），仅作为"启用插件 → 初始化 PJSK 业务栈"
// 的统一编排入口。
//
// 关键职责：
//  1. 读取 data/plugins/moesekai.yml，把 PJSK 配置注入到核心 *config.Config
//     的对应字段（masterdata/sekai_api/suite_api/ranking_api/b30/assets/
//     game_servers/command_aliases/server.region），以兼容下游模块。
//  2. 执行原 main.go 中的 PJSK 启动序列：assets.Configure → servers.Manager
//     → b30 → commands.RegisterAll → 把 servers/store/loader 挂到 web。
//  3. 把这些资源的关闭钩子注册到 plugin.Context 上。
package moesekai

import (
	"moebot-next/internal/config"
)

// Config 是 moesekai 插件的子配置文件 (data/plugins/moesekai.yml) 的根结构。
//
// 字段直接复用 internal/config 里的现成类型，方便插入到 *config.Config。
// 若用户的子配置文件中省略某段，则保留 *config.Config 默认值不变。
type Config struct {
	Region         string                             `yaml:"region"`          // 默认游戏服 region
	CommandAliases map[string][]string                `yaml:"command_aliases"`
	Masterdata     *config.MasterdataConfig           `yaml:"masterdata,omitempty"`
	SekaiAPI       *config.SekaiAPIConfig             `yaml:"sekai_api,omitempty"`
	SuiteAPI       *config.SuiteAPIConfig             `yaml:"suite_api,omitempty"`
	RankingAPI     *config.RankingAPIConfig           `yaml:"ranking_api,omitempty"`
	B30            *config.B30Config                  `yaml:"b30,omitempty"`
	Assets         *config.AssetsConfig               `yaml:"assets,omitempty"`
	GameServers    map[string]config.GameServerConfig `yaml:"game_servers,omitempty"`
}

// applyTo 把已加载的 moesekai 子配置覆盖回核心 *config.Config 的相应字段，
// 使原有 servers/commands 等下游代码无需修改即可读取插件配置。
func (m *Config) applyTo(cfg *config.Config) {
	if m == nil || cfg == nil {
		return
	}
	if m.Region != "" {
		cfg.Server.Region = m.Region
	}
	if len(m.CommandAliases) > 0 {
		cfg.Bot.CommandAliases = m.CommandAliases
	}
	if m.Masterdata != nil {
		cfg.Masterdata = *m.Masterdata
	}
	if m.SekaiAPI != nil {
		cfg.SekaiAPI = *m.SekaiAPI
	}
	if m.SuiteAPI != nil {
		cfg.SuiteAPI = *m.SuiteAPI
	}
	if m.RankingAPI != nil {
		cfg.RankingAPI = *m.RankingAPI
	}
	if m.B30 != nil {
		cfg.B30 = *m.B30
	}
	if m.Assets != nil {
		cfg.Assets = *m.Assets
	}
	if len(m.GameServers) > 0 {
		cfg.GameServers = m.GameServers
	}
	config.NormalizeConfig(cfg)
}

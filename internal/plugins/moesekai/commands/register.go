package commands

import (
	"moebot-next/internal/database"
	"moebot-next/internal/plugins/moesekai/b30"
	"moebot-next/internal/plugins/moesekai/commandparser"
	"moebot-next/internal/plugins/moesekai/servers"
	"moebot-next/internal/renderer"

	"github.com/rs/zerolog/log"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// PreHandler 在 RegisterAll 创建新 Engine 后挂作 UsePreHandler；返回 false
// 时该消息事件会被本插件忽略。通常由父包传入：调用 filter.Manager.AllowMessage
// 实现 plugin:moesekai 这一层的白/黑名单过滤。
//
// 设计上把 filter 依赖通过函数指针注入，让 commands 包不直接 import filter。
type PreHandler = zero.Rule

// Engine 是 moesekai 命令所有 OnXxx 注册的目标 ZeroBot 引擎；
// 调用 RegisterAll 时会重置为新的 engine，从而支持插件 disable→enable
// 的运行期重置（旧 engine 通过 ResetEngine() 调用 Delete 回收）。
var Engine *zero.Engine

// ResetEngine 释放当前 Engine 上的所有命令注册（plugin shutdown 钩子调用）。
func ResetEngine() {
	if Engine != nil {
		Engine.Delete()
		Engine = nil
	}
}

// Deps holds shared dependencies for all commands.
type Deps struct {
	DB          *database.DB
	Renderer    *renderer.Client
	Servers     *servers.Manager
	B30         *b30.Client
	Definitions []commandparser.Definition
	// PreHandler 若非 nil，会挂到新 Engine 的 UsePreHandler 上，作为
	// plugin:moesekai 这一层的统一前置过滤（白/黑名单）。
	PreHandler PreHandler
}

// RegisterAll registers all bot commands.
func RegisterAll(deps *Deps) {
	log.Info().Msg("Registering bot commands...")
	if len(deps.Definitions) == 0 {
		deps.Definitions = commandparser.BaseDefinitions()
	}
	// 每次注册都使用全新 Engine，避免和上一次禁用残留的 matcher 重复触发。
	if Engine != nil {
		Engine.Delete()
	}
	Engine = zero.New()
	if deps.PreHandler != nil {
		Engine.UsePreHandler(deps.PreHandler)
	}

	RegisterHelp(deps)
	RegisterCard(deps)
	RegisterMusic(deps)
	RegisterEvent(deps)
	RegisterGacha(deps)
	RegisterVirtualLive(deps)
	RegisterProfile(deps)
	RegisterSuite(deps)
	RegisterBond(deps)
	RegisterMusicOverview(deps)
	RegisterBest30(deps)
	RegisterChallengeInfo(deps)
	RegisterEventRecord(deps)
	RegisterLeaderCount(deps)
	RegisterCharacterRankMission(deps)
	RegisterAnvo(deps)
	RegisterSuiteCardBox(deps)
	RegisterDeckRecommend(deps)
	RegisterRanking(deps)
	RegisterBirthday()
	RegisterDefaultRegion(deps)
	RegisterSkillCalc(deps)
	RegisterUpdate(deps)
	// RegisterSticker(deps)  // Phase 2
	// RegisterGuess(deps)    // Phase 3

	log.Info().Msg("All commands registered")
}

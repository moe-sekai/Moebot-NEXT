package commands

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// RegisterHelp registers the /帮助 command.
func RegisterHelp() {
	zero.OnCommand("帮助").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text(helpText()))
	})
}

func helpText() string {
	return `🎵 Moebot NEXT — PJSK 查询机器人

📋 可用指令:
  /查卡 [关键词]      — 搜索卡牌信息
  /查曲 [关键词]      — 搜索曲目信息
  /查歌 [关键词]      — /查曲 的别名
  /查谱 [关键词]      — 查询谱面等级与 notes
  /查活动 [关键词]    — 搜索活动信息
  /查卡池 [关键词]    — 搜索卡池信息
  /查扭蛋 [关键词]    — /查卡池 的别名
  /绑定 [游戏ID]      — 绑定 PJSK 账号
  /解绑               — 解除账号绑定
  /个人信息           — 查看个人数据
  /生日               — 查看近期角色生日
  /帮助               — 显示本帮助信息

💡 提示: 搜索支持模糊匹配，可以用角色名、日文名、简称等
🌐 管理面板: http://localhost:8080`
}

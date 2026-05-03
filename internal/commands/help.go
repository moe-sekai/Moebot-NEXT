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
	return `🎵 Moebot NEXT — PJSK 多服务器查询机器人

📋 可用指令:
  /查卡 [关键词]      — 搜索卡牌信息
  /查曲 [关键词]      — 搜索曲目信息
  /查歌 [关键词]      — /查曲 的别名
  /查谱 [关键词]      — 查询谱面等级与 notes
  /查活动 [关键词]    — 搜索活动信息
  /查卡池 [关键词]    — 搜索卡池信息
  /查扭蛋 [关键词]    — /查卡池 的别名
  /绑定 [游戏ID]      — 绑定日服 PJSK 账号
  /cn绑定 [游戏ID]    — 绑定国服账号（tw/kr/en 同理）
  /解绑               — 解除当前默认/绑定服务器账号
  /个人信息           — 查看对应服务器个人数据
  /榜线 [名次]        — 查看活动榜线
  /查房 [名次]        — 查看活跃/时速数据
  /生日               — 查看近期角色生日
  /帮助               — 显示本帮助信息

🌐 服务器前缀: jp / cn / tw / kr / en
💡 例: /cn查卡 初音、/en查曲 tell your world、/kr榜线 1000
💡 无前缀查询会优先使用你的绑定服务器，未绑定则默认日服。
🌐 管理面板: http://localhost:8080`
}

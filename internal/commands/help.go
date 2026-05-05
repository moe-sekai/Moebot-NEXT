package commands

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// RegisterHelp registers the /帮助 command.
func RegisterHelp(deps *Deps) {
	for _, cmd := range parserCommands(deps, "帮助") {
		zero.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(helpText()))
		})
	}
}

func helpText() string {
	return `🎵 Moebot NEXT — PJSK 多服务器查询机器人

📋 可用指令:
  /查卡 [ID/筛选]     — 纯 ID 查详情，其它角色/属性/限定条件查列表
  /查曲 [ID/关键词]   — 搜索曲目信息，支持曲名别名库与候选列表
  /查歌 [关键词]      — /查曲 的别名（song/songinfo/music/musicinfo 同理）
  /查谱 [ID/关键词]   — 查询谱面并直出预览 PNG（谱面/谱面预览/chart 同理）
  /查活动 [关键词]    — 搜索活动信息
  /查卡池 [关键词]    — 搜索卡池信息
  /查扭蛋 [关键词]    — /查卡池 的别名
  /绑定 [游戏ID]      — 绑定日服 PJSK 账号
  /cn绑定 [游戏ID]    — 绑定国服账号（tw/kr/en 同理）
  /解绑               — 解除当前默认/绑定服务器账号
  /个人信息           — 查看对应服务器个人数据
  /抓包状态           — 查看 Suite 抓包更新时间与来源
  /卡牌一览 [条件]    — 查看 Suite 持有卡牌一览，支持 box/id/mr/sl/time
  /羁绊 /冲榜记录     — 查看 Suite 羁绊等级与活动冲榜记录
  /sk [名次/UID]      — 查询指定榜线，支持 1k/1w/范围/绑定账号
  /sk线 /skl /榜线    — 查看整体活动榜线
  /skp /sk预测        — 查看预测/最终线（仅 cn/jp）
  /cf /查房           — 查看活跃/时速数据
  /csb /查水表        — 查看单玩家小时周回和停车区间
  /生日               — 查看近期角色生日
  /帮助               — 显示本帮助信息

🌐 服务器前缀: jp / cn / tw / kr / en
💡 例: /cn查卡 1204、/jp查卡 mnr 4 蓝 限定、/en查曲 tell your world、/krsk 1k
💡 WL 例: /wlsk线、/cnwlsk 1 100、/enwlcf、/wlcsb 1 100
💡 无前缀查询会优先使用你的绑定服务器，未绑定则默认日服。
🌐 管理面板: http://localhost:8080`
}

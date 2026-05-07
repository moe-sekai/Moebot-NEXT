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

� 账号绑定
  /绑定 [游戏ID]        — 绑定日服账号（cn/tw/kr/en 前缀切换区服）
                          例: /绑定 123456789012345678、/cn绑定 123456789012345678
  /解绑                 — 解除当前/指定区服绑定（例: /cn解绑）
  /个人信息 /profile    — 查看绑定账号资料图（例: /cn个人信息）

🔍 查询 / 榜线
  /查卡 [ID/筛选]       — 例: /查卡 1204、/查卡 mnr 4 蓝 限定、/查卡 event123
  /查曲 /查歌 [关键词]  — 例: /查曲 Tell Your World、/查歌 千本樱
  /查谱 [关键词 难度]   — 难度: ex/ma/mas/apd/ap/hd/nm/ez（例: /查谱 千本樱 mas）
  /查活动 [关键词]      — 例: /查活动、/查活动 +1、/event 周年
  /查卡池 /查扭蛋       — 例: /查卡池 700、/查扭蛋 当前、/gacha fes
  /查演唱会 /vlive      — 例: /查演唱会 当前、/vlive 1
  /榜线 /sk线 /skl      — 整体榜线，可加 wl 前缀（例: /cnskl、/wlsk线）
  /sk [名次/范围/UID]   — 例: /sk 100、/sk 1k、/sk 1-10、/cnwlsk 1 100
  /skp /sk预测 [活动ID] — 仅 cn/jp（例: /skp、/sk预测 165）
  /cf /查房             — 时速 / 活跃（例: /cncf 1k、/wlcf 1）
  /csb /查水表          — 单玩家小时周回与停车区间（例: /cncsb 1k、/wlcsb 1 100）

📡 Suite 数据 (需绑定 + Suite 公开)
  /抓包状态 /suite      — Suite 更新时间与来源
  /隐藏抓包 /展示抓包   — 控制自己 Suite 数据是否公开
  /羁绊 /牵绊           — 角色羁绊等级 TOP（例: /cn羁绊）
  /打歌进度 /歌曲奖励   — Suite 打歌统计（例: /打歌进度、/歌曲奖励）
  /b30 /best30          — 社区定数 Best30（例: /cnb30）
  /挑战信息             — 挑战 Live 进度
  /活动记录 /冲榜记录   — 活动 PT 与 WL 章节记录
  /队长次数             — 各角色队长游玩次数
  /cr任务 [角色]        — 例: /cr任务 miku、/cr任务 miku all 队长次数
  /anvo [角色]          — Another Vocal 持有（例: /anvo miku、/cnanvo mnr）
  /卡牌一览 [条件]      — 例: /卡牌一览、/卡牌一览 四星 限定、/卡牌一览 box

🃏 组卡推荐 (需绑定 Suite)
  /组卡 [活动/歌曲/难度/多人/单人/auto/综合力/实效]
                          例: /组卡 多人、/组卡 event123 master、/cn组卡 综合力
  /最强组卡 /长草组卡   — 例: /最强组卡、/长草组卡 实效 5套
  /挑战组卡 [角色]      — 例: /挑战组卡 miku、/挑战组卡 一歌 all
  /加成组卡 /控分组卡   — 例: /加成组卡 300、/控分组卡 event123 250 260 270
  /烤森组卡 [活动ID]    — 例: /烤森组卡、/烤森组卡 event180

🎲 其它
  /抽卡模拟 /roll       — 娱乐向抽卡结果图
  /生日                 — 近期角色生日
  /帮助 /help           — 显示本帮助

🌐 区服前缀: jp / cn / tw / kr / en；WL 加在区服与命令之间
💡 例: /cn查卡 1204、/krsk 1k、/cnwlcf、/wlcsb 1 100
💡 无前缀按你的绑定服务器，未绑定则默认日服
🌐 管理面板: http://localhost:8080`
}

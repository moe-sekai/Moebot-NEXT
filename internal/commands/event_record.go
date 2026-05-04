package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

const eventRecordDefaultLimit = 10

type eventRecordProfile struct {
	suite.BaseProfile
	UserGamedata    suite.UserGamedata     `json:"userGamedata"`
	UserEvents      []userEventRecord      `json:"userEvents"`
	UserWorldBlooms []userWorldBloomRecord `json:"userWorldBlooms"`
}

type userEventRecord struct {
	EventID    int `json:"eventId"`
	EventPoint int `json:"eventPoint"`
	Rank       int `json:"rank"`
}

type userWorldBloomRecord struct {
	EventID                 int `json:"eventId"`
	GameCharacterID         int `json:"gameCharacterId"`
	EventPoint              int `json:"eventPoint"`
	WorldBloomChapterPoint  int `json:"worldBloomChapterPoint"`
	WorldBloomChapterRank   int `json:"worldBloomChapterRank"`
	WorldBloomChapterNumber int `json:"chapterNo"`
}

func eventRecordFields() []string {
	return suite.Fields(suite.FieldUserEvents, suite.FieldUserWorldBlooms)
}

func RegisterEventRecord(deps *Deps) {
	for _, cmd := range parserCommands(deps, "活动记录") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, user := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || !runtime.Enabled {
				ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
				return
			}
			if forcedRegion != "" {
				var err error
				user, err = deps.DB.GetUserByPlatformRegion("onebot", userIDFromCtx(ctx), runtime.Region)
				if err != nil && err != gorm.ErrRecordNotFound {
					ctx.SendChain(message.Text("数据库错误，请稍后重试"))
					return
				}
			}
			if user == nil || user.GameID == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("你还没有绑定%s游戏账号~\n使用 /%s绑定 [游戏ID] 来绑定", runtime.Label, runtime.Region)))
				return
			}
			if runtime.Suite == nil || !runtime.Suite.Enabled() {
				ctx.SendChain(message.Text(fmt.Sprintf("暂不支持查询%s的抓包数据", runtime.Label)))
				return
			}
			setting := suiteSettingOrDefault(deps, userIDFromCtx(ctx), runtime.Region, runtime.Profile.SuiteAPI.DefaultMode)
			if setting.Hidden {
				ctx.SendChain(message.Text(fmt.Sprintf("你已隐藏%s抓包信息，发送 /%s展示抓包 可重新展示", runtime.Label, runtime.Region)))
				return
			}
			var profile eventRecordProfile
			if err := runtime.Suite.GetUserData(user.GameID, setting.Mode, eventRecordFields(), &profile); err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取你的%sSuite抓包数据失败，发送 /抓包 获取帮助\n%s", runtime.Label, err.Error())))
				return
			}
			ctx.SendChain(message.Text(formatEventRecordText(runtime.Region, profile, runtime.Store, eventRecordDefaultLimit)))
			bot.RecordCommandRegion(deps.DB, "活动记录", runtime.Region, ctx, start)
		})
	}
}

func formatEventRecordText(region string, profile eventRecordProfile, store *masterdata.Store, limit int) string {
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
	events := append([]userEventRecord(nil), profile.UserEvents...)
	sort.SliceStable(events, func(i, j int) bool { return events[i].EventPoint > events[j].EventPoint })
	blooms := append([]userWorldBloomRecord(nil), profile.UserWorldBlooms...)
	sort.SliceStable(blooms, func(i, j int) bool { return worldBloomPoint(blooms[i]) > worldBloomPoint(blooms[j]) })
	if limit <= 0 {
		limit = max(len(events), len(blooms))
	}
	lines := []string{
		fmt.Sprintf("%s 活动记录", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("更新时间: %s", suiteUpdateText(profile.UploadTime)),
		fmt.Sprintf("数据来源: %s", suiteSourceText(profile.BaseProfile)),
	}
	if len(events) == 0 && len(blooms) == 0 {
		lines = append(lines, "暂无活动记录")
		return strings.Join(lines, "\n")
	}
	if len(events) > 0 {
		lines = append(lines, "---", "活动PT")
		for i := 0; i < min(limit, len(events)); i++ {
			event := events[i]
			lines = append(lines, fmt.Sprintf("%d. %s: %dpt", i+1, eventName(store, event.EventID), event.EventPoint))
		}
	}
	if len(blooms) > 0 {
		lines = append(lines, "---", "WL章节")
		for i := 0; i < min(limit, len(blooms)); i++ {
			bloom := blooms[i]
			lines = append(lines, fmt.Sprintf("%d. %s %s: %dpt", i+1, eventName(store, bloom.EventID), characterDisplayName(bloom.GameCharacterID), worldBloomPoint(bloom)))
		}
	}
	return strings.Join(lines, "\n")
}

func eventName(store *masterdata.Store, eventID int) string {
	if store != nil {
		if event := store.GetEvent(eventID); event != nil && strings.TrimSpace(event.Name) != "" {
			return fmt.Sprintf("#%d %s", eventID, event.Name)
		}
	}
	return fmt.Sprintf("活动 #%d", eventID)
}

func worldBloomPoint(record userWorldBloomRecord) int {
	if record.WorldBloomChapterPoint > 0 {
		return record.WorldBloomChapterPoint
	}
	return record.EventPoint
}

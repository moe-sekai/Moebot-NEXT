package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
)

const eventRecordDefaultLimit = 10

type eventRecordProfile struct {
	suite.BaseProfile
	UserGamedata    suite.UserGamedata     `json:"userGamedata"`
	UserDecks       []suite.UserDeck       `json:"userDecks"`
	UserCards       []suite.UserCard       `json:"userCards"`
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
			runtime, inferredUser, ok := requireRuntime(deps, ctx, forcedRegion)
			if !ok {
				return
			}
			user, ok := requireBoundUser(deps, ctx, runtime, forcedRegion, inferredUser)
			if !ok {
				return
			}
			if !requireSuite(ctx, runtime, "活动记录") {
				return
			}
			if _, ok := requireSuiteVisible(deps, ctx, runtime); !ok {
				return
			}
			var profile eventRecordProfile
			if !fetchSuiteUserData(ctx, runtime, user.GameID, "活动记录", eventRecordFields(), &profile) {
				return
			}
			payload := buildSuitePanel(runtime, suitePanelTitle(runtime, "活动记录"), "", profile)
			payload.Subtitle = suitePanelSubtitle(profile.BaseProfile)
			sections, stats := rowsFromEventRecord(profile, runtime.Store, runtime.Assets, eventRecordDefaultLimit)
			payload.Stats = append(suiteBasicStats(profile.commonSuiteProfile()), stats...)
			payload.Sections = sections
			sendSuitePanelOrText(ctx, deps, payload, formatEventRecordText(runtime.Region, profile, runtime.Store, eventRecordDefaultLimit))
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

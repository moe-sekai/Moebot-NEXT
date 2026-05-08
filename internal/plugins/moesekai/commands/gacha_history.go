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
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

const gachaHistoryDefaultLimit = 10

type gachaHistoryProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	UserGachas   []userGachaRecord  `json:"userGachas"`
}

type userGachaRecord struct {
	GachaID int `json:"gachaId"`
	Count   int `json:"count"`
}

func gachaHistoryFields() []string {
	return []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserGachas}
}

func RegisterGachaHistory(deps *Deps) {
	for _, cmd := range parserCommands(deps, "抽卡记录") {
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
			if !requireSuite(ctx, runtime, "抽卡记录") {
				return
			}
			if _, ok := requireSuiteVisible(deps, ctx, runtime); !ok {
				return
			}

			var profile gachaHistoryProfile
			if !fetchSuiteUserData(ctx, runtime, user.GameID, "抽卡记录", gachaHistoryFields(), &profile) {
				return
			}
			common := suiteCommandProfile{BaseProfile: profile.BaseProfile, UserGamedata: profile.UserGamedata}
			payload := buildSuitePanel(runtime, suitePanelTitle(runtime, "抽卡记录"), "", common)
			payload.Subtitle = suitePanelSubtitle(profile.BaseProfile)
			rows, stats := rowsFromGachaHistory(profile, runtime.Store, gachaHistoryDefaultLimit)
			payload.Stats = append(suiteBasicStats(common), stats...)
			payload.Sections = []renderpayloads.SuiteSectionPayload{{Title: "卡池抽卡记录", Rows: rows}}
			sendSuitePanelOrText(ctx, deps, payload, formatGachaHistoryText(runtime.Region, profile, runtime.Store, gachaHistoryDefaultLimit))
			bot.RecordCommandRegion(deps.DB, "抽卡记录", runtime.Region, ctx, start)
		})
	}
}

func formatGachaHistoryText(region string, profile gachaHistoryProfile, store *masterdata.Store, limit int) string {
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
	source := suiteSourceText(profile.BaseProfile)
	updateText := suiteUpdateText(profile.UploadTime)

	records := make([]userGachaRecord, 0, len(profile.UserGachas))
	total := 0
	for _, record := range profile.UserGachas {
		if record.Count <= 0 {
			continue
		}
		records = append(records, record)
		total += record.Count
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Count == records[j].Count {
			return records[i].GachaID > records[j].GachaID
		}
		return records[i].Count > records[j].Count
	})
	if limit <= 0 || limit > len(records) {
		limit = len(records)
	}

	lines := []string{
		fmt.Sprintf("%s 抽卡记录", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("总抽数: %d", total),
		fmt.Sprintf("更新时间: %s", updateText),
		fmt.Sprintf("数据来源: %s", source),
	}
	if len(records) == 0 {
		lines = append(lines, "暂无抽卡记录")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "---")
	for i := 0; i < limit; i++ {
		record := records[i]
		lines = append(lines, fmt.Sprintf("%d. %s: %d抽", i+1, gachaHistoryName(store, record.GachaID), record.Count))
	}
	return strings.Join(lines, "\n")
}

func gachaHistoryName(store *masterdata.Store, gachaID int) string {
	if store != nil {
		if gacha := store.GetGacha(gachaID); gacha != nil && strings.TrimSpace(gacha.Name) != "" {
			return fmt.Sprintf("#%d %s", gachaID, gacha.Name)
		}
	}
	return fmt.Sprintf("未知卡池 #%d", gachaID)
}

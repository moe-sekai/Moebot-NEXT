package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/renderer"
	"moebot-next/internal/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
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

			setting := suiteSettingOrDefault(deps, userIDFromCtx(ctx), runtime.Region)
			if setting.Hidden {
				ctx.SendChain(message.Text(fmt.Sprintf("你已隐藏%s抓包信息，发送 /%s展示抓包 可重新展示", runtime.Label, runtime.Region)))
				return
			}

			var profile gachaHistoryProfile
			if err := runtime.Suite.GetUserData(user.GameID, "", gachaHistoryFields(), &profile); err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取你的%s Haruki Suite 公开数据失败\n%s", runtime.Label, err.Error())))
				return
			}
			common := suiteCommandProfile{BaseProfile: profile.BaseProfile, UserGamedata: profile.UserGamedata}
			payload := buildSuitePanel(runtime, suitePanelTitle(runtime, "抽卡记录"), "", common)
			payload.Subtitle = suitePanelSubtitle(profile.BaseProfile)
			rows, stats := rowsFromGachaHistory(profile, runtime.Store, gachaHistoryDefaultLimit)
			payload.Stats = append(suiteBasicStats(common), stats...)
			payload.Sections = []renderer.SuiteSectionPayload{{Title: "卡池抽卡记录", Rows: rows}}
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

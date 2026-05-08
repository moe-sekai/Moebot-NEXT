package commands

import (
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/models"
	"moebot-next/internal/plugins/moesekai/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

func RegisterSuite(deps *Deps) {
	registerSuiteStatusCommands(deps)
	registerSuiteVisibilityCommands(deps)
}

func registerSuiteStatusCommands(deps *Deps) {
	for _, cmd := range parserCommands(deps, "抓包状态") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		Engine.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, inferredUser, ok := requireRuntime(deps, ctx, forcedRegion)
			if !ok {
				return
			}
			user, ok := requireBoundUser(deps, ctx, runtime, forcedRegion, inferredUser)
			if !ok {
				return
			}
			if !requireSuite(ctx, runtime, "抓包状态") {
				return
			}
			if _, ok := requireSuiteVisible(deps, ctx, runtime); !ok {
				return
			}
			var profile suiteCommandProfile
			if !fetchSuiteUserData(ctx, runtime, user.GameID, "抓包状态", suite.Fields(), &profile) {
				return
			}
			status := suite.Status{
				UserID:      profile.UserGamedata.UserID.String(),
				Name:        profile.UserGamedata.Name,
				Source:      profile.Source,
				LocalSource: profile.LocalSource,
				UploadTime:  profile.UploadTime,
			}
			payload := buildSuitePanel(runtime, suitePanelTitle(runtime, "Suite数据"), "", profile)
			payload.Subtitle = suitePanelSubtitle(profile.BaseProfile)
			payload.Stats = append(suiteBasicStats(profile), renderpayloads.SuiteStatPayload{Label: "数据来源", Value: suiteSourceText(profile.BaseProfile)})
			payload.Sections = []renderpayloads.SuiteSectionPayload{{Title: "Suite 状态", Rows: []renderpayloads.SuiteSectionRowPayload{
				{Label: "玩家", Value: payload.Profile.Name},
				{Label: "用户ID", Value: payload.Profile.UserID},
				{Label: "更新时间", Value: suiteUpdateText(profile.UploadTime)},
				{Label: "数据来源", Value: suiteSourceText(profile.BaseProfile)},
				{Label: "接口", Value: "Haruki 公开 API"},
			}}}
			sendSuitePanelOrText(ctx, deps, payload, formatSuiteStatusText(runtime.Region, status))
			bot.RecordCommandRegion(deps.DB, "抓包状态", runtime.Region, ctx, start)
		})
	}
}

func registerSuiteVisibilityCommands(deps *Deps) {
	for _, action := range []struct {
		Command string
		Hidden  bool
		Text    string
	}{
		{Command: "隐藏抓包", Hidden: true, Text: "已隐藏%s抓包信息"},
		{Command: "展示抓包", Hidden: false, Text: "已展示%s抓包信息"},
	} {
		action := action
		for _, cmd := range parserCommands(deps, action.Command) {
			forcedRegion := cmd.Region
			Engine.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
				runtime, _, ok := requireRuntime(deps, ctx, forcedRegion)
				if !ok {
					return
				}
				setting := suiteSettingOrDefault(deps, userIDFromCtx(ctx), runtime.Region)
				setting.Hidden = action.Hidden
				if err := deps.DB.UpsertSuiteSetting(setting); err != nil {
					ctx.SendChain(message.Text("保存抓包隐私设置失败，请稍后重试"))
					return
				}
				ctx.SendChain(message.Text(fmt.Sprintf(action.Text, runtime.Label)))
			})
		}
	}
}

func suiteSettingOrDefault(deps *Deps, platformID string, region string) *models.SuiteSetting {
	setting, err := deps.DB.GetSuiteSetting("onebot", platformID, region)
	if err == nil && setting != nil {
		setting.Mode = config.SuiteModeHaruki
		return setting
	}
	return &models.SuiteSetting{
		Platform:     "onebot",
		PlatformID:   platformID,
		ServerRegion: region,
		Mode:         config.SuiteModeHaruki,
		Hidden:       err != gorm.ErrRecordNotFound,
	}
}

func formatSuiteStatusText(region string, status suite.Status, masks ...func(string) string) string {
	uid := status.UserID
	if uid == "" {
		uid = "未知"
	}
	for _, mask := range masks {
		if mask != nil {
			uid = mask(uid)
		}
	}
	source := status.Source
	if source == "" {
		source = suite.PublicSource
	}
	if status.LocalSource != "" {
		source += "(" + status.LocalSource + ")"
	}
	updateText := "未知"
	if status.UploadTime > 0 {
		updateText = time.UnixMilli(status.UploadTime).Format("2006-01-02 15:04:05")
	}
	name := status.Name
	if name == "" {
		name = "未知玩家"
	}
	return strings.Join([]string{
		fmt.Sprintf("%s Suite数据", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("用户ID: %s", uid),
		fmt.Sprintf("更新时间: %s", updateText),
		fmt.Sprintf("数据来源: %s", source),
		"接口: Haruki 公开 API",
	}, "\n")
}

func hideUIDExceptLast(keep int) func(string) string {
	return func(uid string) string {
		runes := []rune(uid)
		if keep <= 0 || len(runes) <= keep {
			return strings.Repeat("*", len(runes))
		}
		return strings.Repeat("*", len(runes)-keep) + string(runes[len(runes)-keep:])
	}
}

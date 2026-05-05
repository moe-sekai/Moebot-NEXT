package commands

import (
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/models"
	"moebot-next/internal/renderer"
	"moebot-next/internal/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

func RegisterSuite(deps *Deps) {
	registerSuiteStatusCommands(deps)
	registerSuiteVisibilityCommands(deps)
}

func registerSuiteStatusCommands(deps *Deps) {
	for _, cmd := range parserCommands(deps, "抓包状态") {
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
				ctx.SendChain(message.Text(fmt.Sprintf("暂不支持查询%s的 Haruki Suite 公开数据", runtime.Label)))
				return
			}

			setting := suiteSettingOrDefault(deps, userIDFromCtx(ctx), runtime.Region)
			if setting.Hidden {
				ctx.SendChain(message.Text(fmt.Sprintf("你已隐藏%s Suite 信息，发送 /%s展示抓包 可重新展示", runtime.Label, runtime.Region)))
				return
			}
			var profile suiteCommandProfile
			if err := runtime.Suite.GetUserData(user.GameID, "", suite.Fields(), &profile); err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取你的%s Haruki Suite 公开数据失败\n%s", runtime.Label, err.Error())))
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
			payload.Stats = append(suiteBasicStats(profile), renderer.SuiteStatPayload{Label: "数据来源", Value: suiteSourceText(profile.BaseProfile)})
			payload.Sections = []renderer.SuiteSectionPayload{{Title: "Suite 状态", Rows: []renderer.SuiteSectionRowPayload{
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
			zero.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
				runtime, _ := runtimeForCommand(deps, ctx, forcedRegion)
				if runtime == nil || !runtime.Enabled {
					ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
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

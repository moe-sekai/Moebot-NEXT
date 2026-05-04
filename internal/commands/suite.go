package commands

import (
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/models"
	"moebot-next/internal/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

func RegisterSuite(deps *Deps) {
	registerSuiteStatusCommands(deps)
	registerSuiteModeCommands(deps)
	registerSuiteVisibilityCommands(deps)
}

func registerSuiteStatusCommands(deps *Deps) {
	for _, cmd := range regionalCommands("抓包状态") {
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
			status, err := runtime.Suite.GetStatus(user.GameID, setting.Mode)
			if err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取你的%sSuite抓包数据失败，发送 /抓包 获取帮助\n%s", runtime.Label, err.Error())))
				return
			}
			ctx.SendChain(message.Text(formatSuiteStatusText(runtime.Region, setting.Mode, status)))
			bot.RecordCommandRegion(deps.DB, "抓包状态", runtime.Region, ctx, start)
		})
	}
}

func registerSuiteModeCommands(deps *Deps) {
	for _, cmd := range regionalCommands("抓包模式") {
		forcedRegion := cmd.Region
		zero.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			runtime, _ := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || !runtime.Enabled {
				ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
				return
			}
			args := strings.TrimSpace(commandArgs(ctx))
			current := suiteSettingOrDefault(deps, userIDFromCtx(ctx), runtime.Region, runtime.Profile.SuiteAPI.DefaultMode)
			if args == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("你的%s抓包数据获取模式: %s\n可用模式: latest / local / haruki / moesekai", runtime.Label, current.Mode)))
				return
			}
			mode := config.NormalizeSuiteMode(args)
			if !config.IsValidSuiteMode(mode) {
				ctx.SendChain(message.Text("错误的抓包数据获取模式，可用模式: latest / local / haruki / moesekai"))
				return
			}
			current.Mode = mode
			if err := deps.DB.UpsertSuiteSetting(current); err != nil {
				ctx.SendChain(message.Text("保存抓包模式失败，请稍后重试"))
				return
			}
			ctx.SendChain(message.Text(fmt.Sprintf("切换%s抓包数据获取模式: %s", runtime.Label, mode)))
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
		for _, cmd := range regionalCommands(action.Command) {
			forcedRegion := cmd.Region
			zero.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
				runtime, _ := runtimeForCommand(deps, ctx, forcedRegion)
				if runtime == nil || !runtime.Enabled {
					ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
					return
				}
				setting := suiteSettingOrDefault(deps, userIDFromCtx(ctx), runtime.Region, runtime.Profile.SuiteAPI.DefaultMode)
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

func suiteSettingOrDefault(deps *Deps, platformID string, region string, defaultMode string) *models.SuiteSetting {
	setting, err := deps.DB.GetSuiteSetting("onebot", platformID, region)
	if err == nil && setting != nil {
		if setting.Mode == "" {
			setting.Mode = config.NormalizeSuiteMode(defaultMode)
		}
		return setting
	}
	return &models.SuiteSetting{
		Platform:     "onebot",
		PlatformID:   platformID,
		ServerRegion: region,
		Mode:         config.NormalizeSuiteMode(defaultMode),
	}
}

func formatSuiteStatusText(region string, mode string, status suite.Status, masks ...func(string) string) string {
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
	if status.LocalSource != "" {
		source += "(" + status.LocalSource + ")"
	}
	if source == "" {
		source = "未知"
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
		fmt.Sprintf("获取模式: %s", config.NormalizeSuiteMode(mode)),
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

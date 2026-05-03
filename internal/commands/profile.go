package commands

import (
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/models"
	"moebot-next/internal/renderer"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

// RegisterProfile registers the /绑定, /cn绑定, /解绑, /个人信息 commands.
func RegisterProfile(deps *Deps) {
	registerBindCommands(deps)
	registerUnbindCommands(deps)
	registerProfileInfoCommands(deps)
}

func registerBindCommands(deps *Deps) {
	for _, cmd := range regionalCommands("绑定") {
		commandName := cmd.Name
		bindRegion := normalizeRequestedRegion(cmd.Region)
		if bindRegion == "" {
			bindRegion = config.RegionJP
		}
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			gameID := commandArgs(ctx)
			if gameID == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("请输入你的 %s PJSK 游戏 ID~\n例: /%s 123456789012345678", regionLabel(bindRegion), commandName)))
				return
			}

			userID := userIDFromCtx(ctx)
			user, err := deps.DB.GetUserByPlatformRegion("onebot", userID, bindRegion)
			if err != nil && err != gorm.ErrRecordNotFound {
				ctx.SendChain(message.Text("数据库错误，请稍后重试"))
				return
			}
			if user == nil {
				user = &models.User{Platform: "onebot", PlatformID: userID, ServerRegion: bindRegion}
			}
			user.GameID = gameID
			user.ServerRegion = bindRegion
			if err := deps.DB.UpsertUser(user); err != nil {
				ctx.SendChain(message.Text("绑定失败，请稍后重试"))
				return
			}

			ctx.SendChain(message.Text(fmt.Sprintf("✅ %s绑定成功！\n游戏 ID: %s", regionLabel(bindRegion), gameID)))
			bot.RecordCommandRegion(deps.DB, "绑定", bindRegion, ctx, start)
		})
	}
}

func registerUnbindCommands(deps *Deps) {
	for _, cmd := range regionalCommands("解绑") {
		forcedRegion := cmd.Region
		zero.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			region := normalizeRequestedRegion(forcedRegion)
			if region == "" {
				runtime, user := runtimeForCommand(deps, ctx, "")
				if user != nil && user.ServerRegion != "" {
					region = user.ServerRegion
				} else if runtime != nil {
					region = runtime.Region
				}
			}
			if region == "" {
				region = config.RegionJP
			}

			user, err := deps.DB.GetUserByPlatformRegion("onebot", userIDFromCtx(ctx), region)
			if err != nil || user.GameID == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("你还没有绑定过%s游戏账号", regionLabel(region))))
				return
			}
			user.GameID = ""
			if err := deps.DB.UpsertUser(user); err != nil {
				ctx.SendChain(message.Text("解绑失败，请稍后重试"))
				return
			}

			ctx.SendChain(message.Text(fmt.Sprintf("✅ 已解除%s账号绑定", regionLabel(region))))
			bot.RecordCommandRegion(deps.DB, "解绑", region, ctx, start)
		})
	}
}

func registerProfileInfoCommands(deps *Deps) {
	for _, cmd := range regionalCommands("个人信息") {
		forcedRegion := cmd.Region
		zero.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, inferredUser := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || !runtime.Enabled {
				ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
				return
			}

			user := inferredUser
			var err error
			if forcedRegion != "" {
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

			if runtime.Sekai != nil && runtime.Sekai.Enabled() {
				profile, err := runtime.Sekai.GetProfile(user.GameID)
				if err == nil {
					payload := renderer.BuildProfileCardPayloadWithAssets(runtime.Store, *profile, runtime.Assets)
					if deps.Renderer != nil && deps.Renderer.Health() {
						png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "profile_card", Data: payload})
						if err == nil {
							ctx.SendChain(message.ImageBytes(png))
							bot.RecordCommandRegion(deps.DB, "个人信息", runtime.Region, ctx, start)
							return
						}
					}
					ctx.SendChain(message.Text(formatProfileText(payload)))
					bot.RecordCommandRegion(deps.DB, "个人信息", runtime.Region, ctx, start)
					return
				}
			}

			ctx.SendChain(message.Text(fmt.Sprintf("👤 个人信息\n服务器: %s (%s)\n游戏 ID: %s\n绑定时间: %s",
				runtime.Label, strings.ToUpper(runtime.Region), user.GameID, user.CreatedAt.Format("2006-01-02 15:04"))))
			bot.RecordCommandRegion(deps.DB, "个人信息", runtime.Region, ctx, start)
		})
	}
}

func formatProfileText(profile renderer.ProfileCardPayload) string {
	lines := []string{
		"👤 个人资料",
		fmt.Sprintf("昵称：%s", profile.Name),
		fmt.Sprintf("Rank：%d", profile.Rank),
		fmt.Sprintf("用户 ID：%s", profile.UserID),
	}
	if profile.TotalPower > 0 {
		lines = append(lines, fmt.Sprintf("总综合力：%s", formatNumber(profile.TotalPower)))
	}
	if profile.Signature != "" {
		lines = append(lines, fmt.Sprintf("签名：%s", profile.Signature))
	}
	return strings.Join(lines, "\n")
}

func formatNumber(value int) string {
	raw := fmt.Sprintf("%d", value)
	if len(raw) <= 3 {
		return raw
	}
	parts := make([]string, 0, len(raw)/3+1)
	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}
	parts = append([]string{raw}, parts...)
	return strings.Join(parts, ",")
}

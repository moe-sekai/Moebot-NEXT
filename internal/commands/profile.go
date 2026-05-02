package commands

import (
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/models"
	"moebot-next/internal/renderer"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

// RegisterProfile registers the /绑定, /解绑, /个人信息 commands.
func RegisterProfile(deps *Deps) {
	// /绑定 [游戏ID]
	zero.OnCommand("绑定").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		start := time.Now()
		gameID := strings.TrimSpace(fmt.Sprintf("%v", ctx.State["args"]))

		if gameID == "" {
			ctx.SendChain(message.Text("请输入你的 PJSK 游戏 ID~\n例: /绑定 123456789"))
			return
		}

		userID := fmt.Sprintf("%d", ctx.Event.UserID)

		user, err := deps.DB.GetUserByPlatform("onebot", userID)
		if err != nil && err != gorm.ErrRecordNotFound {
			ctx.SendChain(message.Text("数据库错误，请稍后重试"))
			return
		}

		if user == nil {
			user = &models.User{
				Platform:   "onebot",
				PlatformID: userID,
			}
		}

		user.GameID = gameID
		if err := deps.DB.UpsertUser(user); err != nil {
			ctx.SendChain(message.Text("绑定失败，请稍后重试"))
			return
		}

		ctx.SendChain(message.Text(fmt.Sprintf("✅ 绑定成功！\n游戏 ID: %s", gameID)))
		bot.RecordCommand(deps.DB, "绑定", ctx, start)
	})

	// /解绑
	zero.OnCommand("解绑").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		start := time.Now()
		userID := fmt.Sprintf("%d", ctx.Event.UserID)

		user, err := deps.DB.GetUserByPlatform("onebot", userID)
		if err != nil {
			ctx.SendChain(message.Text("你还没有绑定过游戏账号"))
			return
		}

		user.GameID = ""
		if err := deps.DB.UpsertUser(user); err != nil {
			ctx.SendChain(message.Text("解绑失败，请稍后重试"))
			return
		}

		ctx.SendChain(message.Text("✅ 已解除账号绑定"))
		bot.RecordCommand(deps.DB, "解绑", ctx, start)
	})

	// /个人信息
	zero.OnCommand("个人信息").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		start := time.Now()
		userID := fmt.Sprintf("%d", ctx.Event.UserID)

		user, err := deps.DB.GetUserByPlatform("onebot", userID)
		if err != nil {
			ctx.SendChain(message.Text("你还没有绑定游戏账号~\n使用 /绑定 [游戏ID] 来绑定"))
			return
		}

		if user.GameID == "" {
			ctx.SendChain(message.Text("你还没有绑定游戏账号~\n使用 /绑定 [游戏ID] 来绑定"))
			return
		}

		if deps.Sekai != nil && deps.Sekai.Enabled() {
			profile, err := deps.Sekai.GetProfile(user.GameID)
			if err == nil {
				payload := renderer.BuildProfileCardPayloadWithStore(deps.Store, *profile)
				if deps.Renderer != nil && deps.Renderer.Health() {
					png, err := deps.Renderer.Render(renderer.RenderRequest{
						Template: "profile_card",
						Data:     payload,
					})
					if err == nil {
						ctx.SendChain(message.ImageBytes(png))
						bot.RecordCommand(deps.DB, "个人信息", ctx, start)
						return
					}
				}
				ctx.SendChain(message.Text(formatProfileText(payload)))
				bot.RecordCommand(deps.DB, "个人信息", ctx, start)
				return
			}
		}

		ctx.SendChain(message.Text(fmt.Sprintf("👤 个人信息\n游戏 ID: %s\n绑定时间: %s",
			user.GameID, user.CreatedAt.Format("2006-01-02 15:04"))))
		bot.RecordCommand(deps.DB, "个人信息", ctx, start)
	})
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

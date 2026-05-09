package commands

import (
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/models"
	"moebot-next/internal/renderer"

	"moebot-next/internal/plugins/moesekai/renderpayloads"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

// isPlausibleGameID checks that the input looks like a PJSK numeric long ID.
// Strips spaces and ensures only ASCII digits remain. Real long IDs are 16-19
// digits, but we accept anything 6+ digits to remain forgiving.
func isPlausibleGameID(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) >= 6
}

// summarizeBindError trims SEKAI API error messages so end users get a short
// reason instead of a stack of wrapped %w errors.
func summarizeBindError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "404"):
		return "游戏服务器未找到该 ID"
	case strings.Contains(msg, "403"):
		return "SEKAI API 拒绝访问（请检查 token / 请求头）"
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline"):
		return "请求超时，请稍后重试"
	case strings.Contains(msg, "disabled"):
		return "SEKAI API 未启用"
	}
	if i := strings.LastIndex(msg, ": "); i >= 0 && i+2 < len(msg) {
		return msg[i+2:]
	}
	return msg
}

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
		Engine.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			gameID, ok := requireArgument(ctx, commandArgs(ctx), fmt.Sprintf("%s PJSK 游戏 ID（例：/%s 123456789012345678）", regionLabel(bindRegion), commandName))
			if !ok {
				return
			}
			if !isPlausibleGameID(gameID) {
				ctx.SendChain(message.Text("游戏 ID 格式不正确：应为纯数字，请检查后重试"))
				return
			}

			// 若该区服已配置 sekai-api，则额外请求一次 profile 以校验 ID 合法性。
			var verifiedName string
			if runtime := runtimeForRegion(deps, bindRegion); runtime != nil && runtime.Sekai != nil && runtime.Sekai.Enabled() {
				profile, err := runtime.Sekai.GetProfile(gameID)
				if err != nil {
					ctx.SendChain(message.Text(fmt.Sprintf(
						"❌ %s ID %s 校验失败：%s\n请确认 ID 正确（不是 6/8 位短 ID，应为长 ID）后重试",
						regionLabel(bindRegion), gameID, summarizeBindError(err),
					)))
					return
				}
				if profile != nil {
					verifiedName = strings.TrimSpace(profile.Name)
				}
			}

			userID := userIDFromCtx(ctx)
			user, err := deps.DB.GetUserByPlatformRegion("onebot", userID, bindRegion)
			if err != nil && err != gorm.ErrRecordNotFound {
				ctx.SendChain(message.Text("数据库繁忙，请稍后重试"))
				return
			}
			if user == nil {
				user = &models.User{Platform: "onebot", PlatformID: userID, ServerRegion: bindRegion}
			}
			user.GameID = gameID
			user.ServerRegion = bindRegion
			if verifiedName != "" {
				user.Nickname = verifiedName
			}
			if err := deps.DB.UpsertUser(user); err != nil {
				ctx.SendChain(message.Text("绑定失败，请稍后重试"))
				return
			}

			lines := []string{fmt.Sprintf("✅ %s绑定成功！", regionLabel(bindRegion)), fmt.Sprintf("游戏 ID: %s", gameID)}
			if verifiedName != "" {
				lines = append(lines, fmt.Sprintf("游戏昵称: %s", verifiedName))
				lines = append(lines, "（已通过 SEKAI API 校验）")
			}
			ctx.SendChain(message.Text(strings.Join(lines, "\n")))
			bot.RecordCommandRegion(deps.DB, "绑定", bindRegion, ctx, start)
		})
	}
}

func registerUnbindCommands(deps *Deps) {
	for _, cmd := range regionalCommands("解绑") {
		forcedRegion := cmd.Region
		Engine.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
			if err != nil || user == nil || user.GameID == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("你还没有绑定过 %s 游戏账号", regionLabel(region))))
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
	for _, cmd := range parserCommands(deps, "个人信息") {
		forcedRegion := cmd.Region
		recordCommand := cmd.Primary
		if recordCommand == "" {
			recordCommand = "个人信息"
		}
		Engine.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, inferredUser, ok := requireRuntime(deps, ctx, forcedRegion)
			if !ok {
				return
			}
			user, ok := requireBoundUser(deps, ctx, runtime, forcedRegion, inferredUser)
			if !ok {
				return
			}

			sekaiErr := error(nil)
			if runtime.Sekai != nil && runtime.Sekai.Enabled() {
				profile, err := runtime.Sekai.GetProfile(user.GameID)
				if err == nil {
					payload := renderpayloads.BuildProfileCardPayloadWithAssets(runtime.Store, *profile, runtime.Assets)
					if deps.Renderer != nil && deps.Renderer.Health() {
						png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "profile_card", Data: payload})
						if err == nil {
							ctx.SendChain(message.ImageBytes(png))
							bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
							return
						}
					}
					ctx.SendChain(message.Text(formatProfileText(payload)))
					bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
					return
				}
				sekaiErr = err
			}

			lines := []string{
				"👤 个人信息",
				fmt.Sprintf("服务器: %s (%s)", runtime.Label, strings.ToUpper(runtime.Region)),
				fmt.Sprintf("游戏 ID: %s", user.GameID),
				fmt.Sprintf("绑定时间: %s", user.CreatedAt.Format("2006-01-02 15:04")),
			}
			if sekaiErr != nil {
				lines = append(lines, "", "⚠️ 访问 SEKAI-API 失败，请联系管理员检查配置")
			}
			ctx.SendChain(message.Text(strings.Join(lines, "\n")))
			bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
		})
	}
}

func formatProfileText(profile renderpayloads.ProfileCardPayload) string {
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

package commands

import (
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/renderer"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// RegisterCard registers the /查卡 command.
func RegisterCard(deps *Deps) {
	zero.OnCommand("查卡").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		start := time.Now()
		keyword := strings.TrimSpace(fmt.Sprintf("%v", ctx.State["args"]))

		if keyword == "" {
			ctx.SendChain(message.Text("请输入要搜索的卡牌关键词~\n例: /查卡 初音未来"))
			return
		}

		// Search cards from masterdata
		results := deps.Store.SearchCards(keyword)
		if len(results) == 0 {
			ctx.SendChain(message.Text(fmt.Sprintf("没有找到与「%s」匹配的卡牌", keyword)))
			return
		}

		// Take the best match and adapt it to renderer props.
		card := results[0]
		payload := renderer.BuildCardDetailPayload(deps.Store, card)

		// Try to render an image via the renderer service.
		if deps.Renderer != nil && deps.Renderer.Health() {
			png, err := deps.Renderer.Render(renderer.RenderRequest{
				Template: "card_detail",
				Data:     payload,
			})
			if err == nil {
				ctx.SendChain(message.ImageBytes(png))
				bot.RecordCommand(deps.DB, "查卡", ctx, start)
				return
			}
			// Fallback to text if rendering fails
		}

		// Text fallback.
		text := formatCardText(payload)
		ctx.SendChain(message.Text(text))
		bot.RecordCommand(deps.DB, "查卡", ctx, start)
	})
}

// formatCardText formats a card's info as plain text.
func formatCardText(card renderer.CardDetailPayload) string {
	lines := []string{
		fmt.Sprintf("卡牌：%s", card.Prefix),
		fmt.Sprintf("角色：%s", card.CharacterName),
		fmt.Sprintf("稀有度：%s", card.CardRarityType),
		fmt.Sprintf("属性：%s", card.Attr),
		fmt.Sprintf("ID：%d", card.ID),
	}
	if card.Power > 0 {
		lines = append(lines, fmt.Sprintf("综合力：%d", card.Power))
	}
	if card.SkillName != "" {
		lines = append(lines, fmt.Sprintf("技能：%s", card.SkillName))
	}
	if card.GachaPhrase != "" {
		lines = append(lines, fmt.Sprintf("招募台词：%s", card.GachaPhrase))
	}
	return strings.Join(lines, "\n")
}

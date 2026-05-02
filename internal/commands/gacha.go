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

// RegisterGacha registers gacha query commands.
func RegisterGacha(deps *Deps) {
	registerGachaCommand(deps, "查卡池")
	registerGachaCommand(deps, "查扭蛋")
}

func registerGachaCommand(deps *Deps, command string) {
	zero.OnCommand(command).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		start := time.Now()
		keyword := strings.TrimSpace(fmt.Sprintf("%v", ctx.State["args"]))

		if keyword == "" {
			ctx.SendChain(message.Text(fmt.Sprintf("请输入要搜索的扭蛋关键词~\n例: /%s 限定", command)))
			return
		}

		results := deps.Store.SearchGachas(keyword)
		if len(results) == 0 {
			ctx.SendChain(message.Text(fmt.Sprintf("没有找到与「%s」匹配的扭蛋", keyword)))
			return
		}

		gacha := results[0]
		payload := renderer.BuildGachaInfoPayload(deps.Store, gacha)

		if deps.Renderer != nil && deps.Renderer.Health() {
			png, err := deps.Renderer.Render(renderer.RenderRequest{
				Template: "gacha_info",
				Data:     payload,
			})
			if err == nil {
				ctx.SendChain(message.ImageBytes(png))
				bot.RecordCommand(deps.DB, command, ctx, start)
				return
			}
		}

		ctx.SendChain(message.Text(formatGachaText(payload)))
		bot.RecordCommand(deps.DB, command, ctx, start)
	})
}

func formatGachaText(gacha renderer.GachaInfoPayload) string {
	lines := []string{
		fmt.Sprintf("卡池：%s", gacha.Name),
		fmt.Sprintf("类型：%s", gachaTypeLabel(gacha.GachaType)),
		fmt.Sprintf("ID：%d", gacha.ID),
	}
	if gacha.StartAt > 0 {
		lines = append(lines, fmt.Sprintf("开始：%s", formatMillis(gacha.StartAt)))
	}
	if gacha.EndAt > 0 {
		lines = append(lines, fmt.Sprintf("结束：%s", formatMillis(gacha.EndAt)))
	}
	if len(gacha.PickupCards) > 0 {
		cards := make([]string, 0, len(gacha.PickupCards))
		for _, card := range gacha.PickupCards {
			cards = append(cards, fmt.Sprintf("%s #%d", card.CharacterName, card.ID))
		}
		lines = append(lines, "Pickup："+strings.Join(cards, "、"))
	}
	if len(gacha.Rates) > 0 {
		rates := make([]string, 0, len(gacha.Rates))
		for _, rate := range gacha.Rates {
			rates = append(rates, fmt.Sprintf("%s %.2f%%", rate.CardRarityType, rate.Rate))
		}
		lines = append(lines, "概率："+strings.Join(rates, "，"))
	}
	return strings.Join(lines, "\n")
}

func gachaTypeLabel(value string) string {
	switch value {
	case "ceil":
		return "天井扭蛋"
	case "normal":
		return "普通扭蛋"
	case "limited":
		return "限定扭蛋"
	case "birthday":
		return "生日扭蛋"
	case "colorful_festival":
		return "Colorful Festival"
	default:
		return value
	}
}

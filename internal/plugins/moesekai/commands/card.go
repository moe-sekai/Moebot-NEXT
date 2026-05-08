package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/plugins/moesekai/cardquery"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/renderer"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

// RegisterCard registers the /查卡 command.
func RegisterCard(deps *Deps) {
	for _, cmd := range parserCommands(deps, "查卡") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		recordCommand := cmd.Primary
		if recordCommand == "" {
			recordCommand = "查卡"
		}
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			keyword := commandArgs(ctx)
			runtime, _ := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || runtime.Store == nil || !runtime.Enabled {
				ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
				return
			}

			if keyword == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("请输入要搜索的卡牌关键词~\n例: /%s 初音未来", commandName)))
				return
			}

			result := cardquery.Resolve(runtime.Store, keyword)
			if result.Message != "" {
				ctx.SendChain(message.Text(result.Message))
				return
			}

			if result.Mode == cardquery.ModeList {
				payload := renderpayloads.BuildCardListPayloadWithAssets("卡牌查询", cardListSubtitle(keyword, result.Query), result.Cards, runtime.Store, runtime.Assets, result.Page, result.TotalPages, result.Total)
				if deps.Renderer != nil && deps.Renderer.Health() {
					png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "card_list", Data: payload})
					if err == nil {
						ctx.SendChain(message.ImageBytes(png))
						bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
						return
					}
				}
				ctx.SendChain(message.Text(formatCardListText(payload)))
				bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
				return
			}

			card := result.Cards[0]
			payload := renderpayloads.BuildCardDetailPayloadWithAssets(runtime.Store, card, runtime.Assets)
			if deps.Renderer != nil && deps.Renderer.Health() {
				png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "card_detail", Data: payload})
				if err == nil {
					ctx.SendChain(message.ImageBytes(png))
					bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
					return
				}
			}

			ctx.SendChain(message.Text(formatCardText(payload)))
			bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
		})
	}
}

func cardsForCharacter(store *masterdata.Store, characterID int) []masterdata.CardInfo {
	cards := make([]masterdata.CardInfo, 0)
	for _, card := range store.AllCards() {
		if card.CharacterID == characterID {
			cards = append(cards, card)
		}
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].ReleaseAt != cards[j].ReleaseAt {
			return cards[i].ReleaseAt < cards[j].ReleaseAt
		}
		return cards[i].ID < cards[j].ID
	})
	return cards
}

func cardListSubtitle(raw string, query cardquery.Query) string {
	parts := make([]string, 0)
	if query.Keyword != "" {
		parts = append(parts, "关键词："+query.Keyword)
	} else if strings.TrimSpace(raw) != "" {
		parts = append(parts, "条件："+strings.TrimSpace(raw))
	}
	if query.Page > 1 {
		parts = append(parts, fmt.Sprintf("第 %d 页", query.Page))
	}
	if len(parts) == 0 {
		return "列表查询"
	}
	return strings.Join(parts, " · ")
}

// formatCardText formats a card's info as plain text.
func formatCardListText(payload renderpayloads.CardListPayload) string {
	lines := []string{fmt.Sprintf("%s（第 %d/%d 页，共 %d 张）", payload.Title, payload.Page, payload.TotalPages, payload.Total)}
	for _, card := range payload.Cards {
		lines = append(lines, fmt.Sprintf("#%d %s · %s · %s", card.ID, card.Prefix, card.CharacterName, card.CardRarityType))
	}
	return strings.Join(lines, "\n")
}

func formatCardText(card renderpayloads.CardDetailPayload) string {
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
	if len(card.Events) > 0 {
		events := make([]string, 0, len(card.Events))
		for _, event := range card.Events {
			events = append(events, fmt.Sprintf("#%d %s", event.ID, event.Name))
		}
		lines = append(lines, "关联活动："+strings.Join(events, "、"))
	}
	return strings.Join(lines, "\n")
}

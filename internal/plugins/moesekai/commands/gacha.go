package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/renderer"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

// RegisterGacha registers gacha query commands.
func RegisterGacha(deps *Deps) {
	registerGachaCommand(deps, "查卡池")
}

func registerGachaCommand(deps *Deps, command string) {
	for _, cmd := range parserCommands(deps, command) {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		recordCommand := cmd.Primary
		if recordCommand == "" {
			recordCommand = command
		}
		Engine.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			keyword := commandArgs(ctx)
			runtime, _ := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || runtime.Store == nil || !runtime.Enabled {
				ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
				return
			}

			gachas, listMode, page, totalPages, total := searchGachasForCommand(runtime.Store, keyword)
			if len(gachas) == 0 {
				if keyword == "" {
					ctx.SendChain(message.Text("当前没有可查询的卡池"))
				} else {
					ctx.SendChain(message.Text(fmt.Sprintf("没有找到与「%s」匹配的扭蛋", keyword)))
				}
				return
			}

			if listMode || len(gachas) > 1 {
				payload := renderpayloads.BuildGachaListPayloadWithAssets("卡池查询", gachaSubtitleForCommand(keyword), gachas, runtime.Store, runtime.Assets, page, totalPages, total)
				if deps.Renderer != nil {
					png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "gacha_list", Data: payload})
					if err == nil {
						ctx.SendChain(message.ImageBytes(png))
						bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
						return
					}
				}
				ctx.SendChain(message.Text(formatGachaListText(payload)))
				bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
				return
			}

			gacha := gachas[0]
			payload := renderpayloads.BuildGachaInfoPayloadWithAssets(runtime.Store, gacha, runtime.Assets)
			if deps.Renderer != nil {
				png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "gacha_info", Data: payload})
				if err == nil {
					ctx.SendChain(message.ImageBytes(png))
					bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
					return
				}
			}

			ctx.SendChain(message.Text(formatGachaText(payload)))
			bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
		})
	}
}

func searchGachasForCommand(store *masterdata.Store, keyword string) ([]masterdata.GachaInfo, bool, int, int, int) {
	all, listMode := collectGachasForCommand(store, keyword)
	if len(all) == 0 {
		return nil, listMode, 1, 1, 0
	}
	if !listMode {
		return all[:1], false, 1, 1, len(all)
	}
	options := parseSearchOptions(keyword)
	paged, page, totalPages := paginate(all, options.Page, listPageSize)
	return paged, true, page, totalPages, len(all)
}

func collectGachasForCommand(store *masterdata.Store, keyword string) ([]masterdata.GachaInfo, bool) {
	if store == nil {
		return nil, false
	}
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return currentOrRecentGachas(store.AllGachas()), true
	}
	if eventID, ok := parseEventToken(keyword); ok {
		if gacha := gachaByEventID(store, eventID); gacha != nil {
			return []masterdata.GachaInfo{*gacha}, false
		}
		return nil, false
	}
	if idx, ok := parseRelativeIndex(keyword); ok && idx < 0 {
		gachas := sortedReleasedGachas(store.AllGachas())
		pos := len(gachas) + idx
		if pos >= 0 && pos < len(gachas) {
			return []masterdata.GachaInfo{gachas[pos]}, false
		}
		return nil, false
	}
	if id, err := strconv.Atoi(keyword); err == nil && id > 0 {
		if gacha := store.GetGacha(id); gacha != nil {
			return []masterdata.GachaInfo{*gacha}, false
		}
	}

	options := parseSearchOptions(keyword)
	listMode := options.Year > 0 || options.Leak || options.Current || options.CardID > 0 || options.GachaType != "" || options.Rerelease || options.Recall
	base := []masterdata.GachaInfo{}
	if options.Keyword != "" {
		base = store.SearchGachas(options.Keyword)
	} else {
		base = store.AllGachas()
		listMode = true
	}
	filtered := make([]masterdata.GachaInfo, 0, len(base))
	for _, gacha := range base {
		if options.Year > 0 && !sameYear(gacha.StartAt, options.Year) {
			continue
		}
		if options.Leak && !isFuture(gacha.StartAt) {
			continue
		}
		if !options.Leak && gacha.StartAt > time.Now().UnixMilli() && listMode {
			continue
		}
		if options.Current && !isNowBetween(gacha.StartAt, gacha.EndAt) {
			continue
		}
		if options.CardID > 0 && !gachaHasCard(gacha, options.CardID) {
			continue
		}
		if options.GachaType != "" && !gachaMatchesType(gacha, options.GachaType) {
			continue
		}
		if options.Rerelease && !strings.Contains(normalizeQuery(gacha.Name), "复刻") && !strings.Contains(strings.ToLower(gacha.GachaType), "rerelease") {
			continue
		}
		if options.Recall && !strings.Contains(normalizeQuery(gacha.Name), "回响") && !strings.Contains(strings.ToLower(gacha.GachaType), "recall") {
			continue
		}
		filtered = append(filtered, gacha)
	}
	if listMode {
		sort.SliceStable(filtered, func(i, j int) bool {
			if filtered[i].StartAt != filtered[j].StartAt {
				return filtered[i].StartAt < filtered[j].StartAt
			}
			return filtered[i].ID < filtered[j].ID
		})
	}
	return filtered, listMode
}

func currentOrRecentGachas(gachas []masterdata.GachaInfo) []masterdata.GachaInfo {
	current := make([]masterdata.GachaInfo, 0)
	for _, gacha := range gachas {
		if isNowBetween(gacha.StartAt, gacha.EndAt) {
			current = append(current, gacha)
		}
	}
	if len(current) > 0 {
		sort.SliceStable(current, func(i, j int) bool { return current[i].StartAt < current[j].StartAt })
		return current
	}
	released := sortedReleasedGachas(gachas)
	if len(released) > listPageSize {
		return released[len(released)-listPageSize:]
	}
	return released
}

func sortedReleasedGachas(gachas []masterdata.GachaInfo) []masterdata.GachaInfo {
	now := time.Now().UnixMilli()
	out := make([]masterdata.GachaInfo, 0, len(gachas))
	for _, gacha := range gachas {
		if gacha.StartAt <= now {
			out = append(out, gacha)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].StartAt != out[j].StartAt {
			return out[i].StartAt < out[j].StartAt
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func gachaByEventID(store *masterdata.Store, eventID int) *masterdata.GachaInfo {
	cardIDs := map[int]struct{}{}
	for _, link := range store.GetEventCards(eventID) {
		cardIDs[link.CardID] = struct{}{}
	}
	if len(cardIDs) == 0 {
		return nil
	}
	for _, gacha := range store.AllGachas() {
		for _, pickup := range gacha.GachaPickups {
			if _, ok := cardIDs[pickup.CardID]; ok {
				return &gacha
			}
		}
	}
	return nil
}

func gachaHasCard(gacha masterdata.GachaInfo, cardID int) bool {
	for _, pickup := range gacha.GachaPickups {
		if pickup.CardID == cardID {
			return true
		}
	}
	return false
}

func gachaMatchesType(gacha masterdata.GachaInfo, target string) bool {
	value := strings.ToLower(gacha.GachaType)
	switch target {
	case "festival":
		return strings.Contains(value, "festival") || strings.Contains(value, "fes")
	case "limited":
		return strings.Contains(value, "limited")
	case "birthday":
		return strings.Contains(value, "birthday")
	case "normal":
		return strings.Contains(value, "normal") || strings.Contains(value, "permanent") || value == "ceil"
	default:
		return strings.Contains(value, target)
	}
}

func gachaSubtitleForCommand(keyword string) string {
	if strings.TrimSpace(keyword) == "" {
		return "当前/最近卡池"
	}
	return "关键词：" + keyword
}

func formatGachaListText(payload renderpayloads.GachaListPayload) string {
	lines := []string{fmt.Sprintf("%s（第 %d/%d 页，共 %d 个）", payload.Title, payload.Page, payload.TotalPages, payload.Total)}
	for _, gacha := range payload.Gachas {
		lines = append(lines, fmt.Sprintf("#%d %s · %s", gacha.ID, gacha.Name, gachaTypeLabel(gacha.GachaType)))
	}
	return strings.Join(lines, "\n")
}

func formatGachaText(gacha renderpayloads.GachaInfoPayload) string {
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

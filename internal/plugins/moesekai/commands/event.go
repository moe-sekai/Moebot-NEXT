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

// RegisterEvent registers the /查活动 command.
func RegisterEvent(deps *Deps) {
	for _, cmd := range parserCommands(deps, "查活动") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		recordCommand := cmd.Primary
		if recordCommand == "" {
			recordCommand = "查活动"
		}
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			keyword := commandArgs(ctx)
			runtime, _ := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || runtime.Store == nil || !runtime.Enabled {
				ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
				return
			}

			events, listMode, page, totalPages, total := searchEventsForCommand(runtime.Store, keyword)
			if len(events) == 0 {
				if keyword == "" {
					ctx.SendChain(message.Text("当前没有可查询的活动"))
				} else {
					ctx.SendChain(message.Text(fmt.Sprintf("没有找到与「%s」匹配的活动", keyword)))
				}
				return
			}

			if listMode || len(events) > 1 {
				payload := renderpayloads.BuildEventListPayloadWithAssets("活动查询", eventSubtitle(keyword), events, runtime.Store, runtime.Assets, page, totalPages, total)
				if deps.Renderer != nil && deps.Renderer.Health() {
					png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "event_list", Data: payload})
					if err == nil {
						ctx.SendChain(message.ImageBytes(png))
						bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
						return
					}
				}
				ctx.SendChain(message.Text(formatEventListText(payload)))
				bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
				return
			}

			event := events[0]
			payload := renderpayloads.BuildEventInfoPayloadWithAssets(runtime.Store, event, runtime.Assets)
			if deps.Renderer != nil && deps.Renderer.Health() {
				png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "event_info", Data: payload})
				if err == nil {
					ctx.SendChain(message.ImageBytes(png))
					bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
					return
				}
			}

			ctx.SendChain(message.Text(formatEventText(payload)))
			bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
		})
	}
}

func searchEventsForCommand(store *masterdata.Store, keyword string) ([]masterdata.EventInfo, bool, int, int, int) {
	all, listMode := collectEventsForCommand(store, keyword)
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

func collectEventsForCommand(store *masterdata.Store, keyword string) ([]masterdata.EventInfo, bool) {
	if store == nil {
		return nil, false
	}
	keyword = strings.TrimSpace(keyword)
	if keyword == "" || normalizeQuery(keyword) == "当前" || normalizeQuery(keyword) == "现在" {
		if ev := currentOrNextEvent(store.AllEvents()); ev != nil {
			return []masterdata.EventInfo{*ev}, false
		}
		return nil, false
	}
	if offset, ok := parseRelativeIndex(keyword); ok {
		if ev := eventByOffset(store.AllEvents(), offset); ev != nil {
			return []masterdata.EventInfo{*ev}, false
		}
		return nil, false
	}
	if id, err := strconv.Atoi(keyword); err == nil && id > 0 {
		if event := store.GetEvent(id); event != nil {
			return []masterdata.EventInfo{*event}, false
		}
	}
	options := parseSearchOptions(keyword)
	listMode := options.Year > 0 || options.Leak || options.Unit != "" || options.Attr != "" || options.Current
	base := []masterdata.EventInfo{}
	if options.Keyword != "" {
		base = store.SearchEvents(options.Keyword)
	} else {
		base = store.AllEvents()
		listMode = true
	}
	filtered := make([]masterdata.EventInfo, 0, len(base))
	for _, event := range base {
		if options.Year > 0 && !sameYear(event.StartAt, options.Year) {
			continue
		}
		if options.Leak && !isFuture(event.StartAt) {
			continue
		}
		if options.Current && !isNowBetween(event.StartAt, event.ClosedAt) {
			continue
		}
		if options.Unit != "" && event.Unit != options.Unit {
			continue
		}
		if options.Attr != "" && !eventHasAttr(store, event.ID, options.Attr) {
			continue
		}
		filtered = append(filtered, event)
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

func currentOrNextEvent(events []masterdata.EventInfo) *masterdata.EventInfo {
	now := time.Now().UnixMilli()
	sort.SliceStable(events, func(i, j int) bool { return events[i].StartAt < events[j].StartAt })
	for _, event := range events {
		if event.StartAt <= now && now <= event.ClosedAt {
			return &event
		}
	}
	for _, event := range events {
		if event.StartAt > now {
			return &event
		}
	}
	if len(events) == 0 {
		return nil
	}
	return &events[len(events)-1]
}

func eventByOffset(events []masterdata.EventInfo, offset int) *masterdata.EventInfo {
	cur := currentOrNextEvent(events)
	if cur == nil {
		return nil
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].StartAt < events[j].StartAt })
	idx := 0
	for i, event := range events {
		if event.ID == cur.ID {
			idx = i
			break
		}
	}
	if offset < 0 {
		idx += offset + 1
	} else {
		idx += offset
	}
	if idx < 0 || idx >= len(events) {
		return nil
	}
	return &events[idx]
}

func eventHasAttr(store *masterdata.Store, eventID int, attr string) bool {
	for _, bonus := range store.GetEventDeckBonuses(eventID) {
		if bonus.CardAttr == attr {
			return true
		}
	}
	return false
}

func eventSubtitle(keyword string) string {
	if strings.TrimSpace(keyword) == "" {
		return "当前/近期活动"
	}
	return "关键词：" + keyword
}

func formatEventListText(payload renderpayloads.EventListPayload) string {
	lines := []string{fmt.Sprintf("%s（第 %d/%d 页，共 %d 个）", payload.Title, payload.Page, payload.TotalPages, payload.Total)}
	for _, event := range payload.Events {
		lines = append(lines, fmt.Sprintf("#%d %s · %s", event.ID, event.Name, event.EventType))
	}
	return strings.Join(lines, "\n")
}

func formatEventText(event renderpayloads.EventInfoPayload) string {
	lines := []string{
		fmt.Sprintf("活动：%s", event.Name),
		fmt.Sprintf("类型：%s", event.EventType),
		fmt.Sprintf("ID：%d", event.ID),
	}
	if event.Unit != "" && event.Unit != "none" {
		lines = append(lines, fmt.Sprintf("团组：%s", event.Unit))
	}
	if event.StartAt > 0 {
		lines = append(lines, fmt.Sprintf("开始：%s", formatMillis(event.StartAt)))
	}
	if event.AggregateAt > 0 {
		lines = append(lines, fmt.Sprintf("结算：%s", formatMillis(event.AggregateAt)))
	}
	if event.ClosedAt > 0 {
		lines = append(lines, fmt.Sprintf("关闭：%s", formatMillis(event.ClosedAt)))
	}
	if event.BonusAttr != "" {
		lines = append(lines, fmt.Sprintf("加成属性：%s", event.BonusAttr))
	}
	if len(event.BonusCharacters) > 0 {
		lines = append(lines, "加成角色："+strings.Join(event.BonusCharacters, "、"))
	}
	if len(event.BonusCards) > 0 {
		cards := make([]string, 0, len(event.BonusCards))
		for _, card := range event.BonusCards {
			cards = append(cards, fmt.Sprintf("#%d %s · %s", card.ID, card.CharacterName, card.Prefix))
		}
		lines = append(lines, "加成卡："+strings.Join(cards, "、"))
	}
	if len(event.PickupCards) > 0 {
		cards := make([]string, 0, len(event.PickupCards))
		for _, card := range event.PickupCards {
			cards = append(cards, fmt.Sprintf("#%d %s · %s", card.ID, card.CharacterName, card.Prefix))
		}
		lines = append(lines, "Pickup卡："+strings.Join(cards, "、"))
	}
	return strings.Join(lines, "\n")
}

func formatMillis(ms int64) string {
	if ms <= 0 {
		return "-"
	}
	return time.Unix(0, ms*int64(time.Millisecond)).Format("2006-01-02 15:04")
}

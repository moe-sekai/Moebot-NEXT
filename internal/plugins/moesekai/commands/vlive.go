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

// RegisterVirtualLive registers virtual live / concert query commands.
func RegisterVirtualLive(deps *Deps) {
	for _, cmd := range parserCommands(deps, "查演唱会") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		recordCommand := cmd.Primary
		if recordCommand == "" {
			recordCommand = "查演唱会"
		}
		Engine.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			keyword := commandArgs(ctx)
			runtime, _ := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || runtime.Store == nil || !runtime.Enabled {
				ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
				return
			}

			lives, page, totalPages, total := searchVirtualLivesForCommand(runtime.Store, keyword)
			if len(lives) == 0 {
				ctx.SendChain(message.Text("当前没有近期虚拟 Live"))
				return
			}
			payload := renderpayloads.BuildVirtualLiveListPayloadWithAssets("虚拟 Live", virtualLiveSubtitle(keyword), lives, runtime.Store, runtime.Assets, page, totalPages, total)
			if deps.Renderer != nil {
				png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "virtual_live_list", Data: payload})
				if err == nil {
					ctx.SendChain(message.ImageBytes(png))
					bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
					return
				}
			}

			ctx.SendChain(message.Text(formatVirtualLiveListText(payload)))
			bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
		})
	}
}

func searchVirtualLivesForCommand(store *masterdata.Store, keyword string) ([]masterdata.VirtualLive, int, int, int) {
	all := collectVirtualLivesForCommand(store, keyword)
	if len(all) == 0 {
		return nil, 1, 1, 0
	}
	options := parseSearchOptions(keyword)
	paged, page, totalPages := paginate(all, options.Page, listPageSize)
	return paged, page, totalPages, len(all)
}

func collectVirtualLivesForCommand(store *masterdata.Store, keyword string) []masterdata.VirtualLive {
	if store == nil {
		return nil
	}
	keyword = strings.TrimSpace(keyword)
	if id, err := strconv.Atoi(keyword); err == nil && id > 0 {
		if live := store.GetVirtualLive(id); live != nil {
			return []masterdata.VirtualLive{*live}
		}
	}
	options := parseSearchOptions(keyword)
	lives := store.AllVirtualLives()
	filtered := make([]masterdata.VirtualLive, 0, len(lives))
	for _, live := range lives {
		if len(live.VirtualLiveSchedules) == 0 {
			continue
		}
		if options.Year > 0 && !sameYear(virtualLiveStart(live), options.Year) {
			continue
		}
		if options.Current && !virtualLiveIsCurrent(live) {
			continue
		}
		if options.Leak && !isFuture(virtualLiveStart(live)) {
			continue
		}
		if options.Keyword != "" && bestLocalScore(options.Keyword, live.Name, live.AssetbundleName, live.VirtualLiveType) == 0 {
			continue
		}
		if options.Keyword == "" && !options.Leak && !options.Current && !isRecentVirtualLive(live) {
			continue
		}
		filtered = append(filtered, live)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return virtualLiveStart(filtered[i]) < virtualLiveStart(filtered[j])
	})
	return filtered
}

func isRecentVirtualLive(live masterdata.VirtualLive) bool {
	now := time.Now().UnixMilli()
	start := virtualLiveStart(live)
	end := virtualLiveEnd(live)
	return end > now && start-now < int64(7*24*time.Hour/time.Millisecond) && end-start < int64(30*24*time.Hour/time.Millisecond)
}

func virtualLiveIsCurrent(live masterdata.VirtualLive) bool {
	now := time.Now().UnixMilli()
	for _, schedule := range live.VirtualLiveSchedules {
		if schedule.StartAt <= now && now <= schedule.EndAt {
			return true
		}
	}
	return false
}

func virtualLiveStart(live masterdata.VirtualLive) int64 {
	if len(live.VirtualLiveSchedules) == 0 {
		return live.StartAt
	}
	start := live.VirtualLiveSchedules[0].StartAt
	for _, schedule := range live.VirtualLiveSchedules[1:] {
		if schedule.StartAt < start {
			start = schedule.StartAt
		}
	}
	return start
}

func virtualLiveEnd(live masterdata.VirtualLive) int64 {
	if len(live.VirtualLiveSchedules) == 0 {
		return live.EndAt
	}
	end := live.VirtualLiveSchedules[0].EndAt
	for _, schedule := range live.VirtualLiveSchedules[1:] {
		if schedule.EndAt > end {
			end = schedule.EndAt
		}
	}
	return end
}

func virtualLiveSubtitle(keyword string) string {
	if strings.TrimSpace(keyword) == "" {
		return "未来 7 天内的近期虚拟 Live"
	}
	return "关键词：" + keyword
}

func formatVirtualLiveListText(payload renderpayloads.VirtualLiveListPayload) string {
	lines := []string{fmt.Sprintf("%s（第 %d/%d 页，共 %d 个）", payload.Title, payload.Page, payload.TotalPages, payload.Total)}
	for _, live := range payload.VirtualLives {
		lines = append(lines, fmt.Sprintf("#%d %s · %s - %s", live.ID, live.Name, formatMillis(live.StartAt), formatMillis(live.EndAt)))
	}
	return strings.Join(lines, "\n")
}

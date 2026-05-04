package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/ranking"
	"moebot-next/internal/renderer"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func RegisterRanking(deps *Deps) {
	for _, cmd := range parserCommands(deps, "榜线") {
		forcedRegion := cmd.Region
		recordCommand := cmd.Primary
		if recordCommand == "" {
			recordCommand = "榜线"
		}
		zero.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, _ := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || runtime.Ranking == nil || !runtime.Enabled {
				ctx.SendChain(message.Text("榜线服务未配置"))
				return
			}
			board, err := runtime.Ranking.GetLatest()
			if err != nil {
				ctx.SendChain(message.Text("榜线获取失败，请稍后重试"))
				return
			}
			rank := parseRankArg(ctx.State["args"])
			filtered := filterRankings(board.Rankings, rank, defaultBorderRanks())
			view := *board
			view.Rankings = filtered
			payload := renderer.BuildRankingListPayloadWithAssets("活动榜线", view, runtime.Assets)
			if sendRankingImage(ctx, deps.Renderer, payload) {
				bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
				return
			}
			ctx.SendChain(message.Text(formatRankingText(payload)))
			bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
		})
	}

	for _, cmd := range regionalCommands("查房") {
		forcedRegion := cmd.Region
		zero.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, _ := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || runtime.Ranking == nil || !runtime.Enabled {
				ctx.SendChain(message.Text("查房服务未配置"))
				return
			}
			board, err := runtime.Ranking.GetChurn()
			if err != nil {
				ctx.SendChain(message.Text("查房获取失败，请稍后重试"))
				return
			}
			rank := parseRankArg(ctx.State["args"])
			filtered := filterRankings(board.Rankings, rank, defaultBorderRanks())
			view := *board
			view.Rankings = filtered
			payload := renderer.BuildChurnRankingListPayloadWithAssets(view, runtime.Assets)
			if sendRankingImageWithTemplate(ctx, deps.Renderer, "churn_ranking_list", payload) {
				bot.RecordCommandRegion(deps.DB, "查房", runtime.Region, ctx, start)
				return
			}
			ctx.SendChain(message.Text(formatRankingText(payload)))
			bot.RecordCommandRegion(deps.DB, "查房", runtime.Region, ctx, start)
		})
	}
}

func parseRankArg(raw any) int {
	value := strings.TrimSpace(fmt.Sprintf("%v", raw))
	if value == "" {
		return 0
	}
	rank, _ := strconv.Atoi(value)
	return rank
}

func defaultBorderRanks() []int {
	return []int{1, 10, 50, 100, 500, 1000, 2000, 5000, 10000}
}

func filterRankings(entries []ranking.RankingEntry, rank int, fallbackRanks []int) []ranking.RankingEntry {
	if len(entries) == 0 {
		return nil
	}
	if rank > 0 {
		return []ranking.RankingEntry{nearestRanking(entries, rank)}
	}
	wanted := map[int]struct{}{}
	for _, r := range fallbackRanks {
		wanted[r] = struct{}{}
	}
	result := make([]ranking.RankingEntry, 0, len(fallbackRanks))
	seen := map[int]struct{}{}
	for _, entry := range entries {
		if _, ok := wanted[entry.Rank]; ok {
			result = append(result, entry)
			seen[entry.Rank] = struct{}{}
		}
	}
	if len(result) > 0 {
		return result
	}
	limit := len(entries)
	if limit > 10 {
		limit = 10
	}
	return entries[:limit]
}

func nearestRanking(entries []ranking.RankingEntry, rank int) ranking.RankingEntry {
	best := entries[0]
	bestDiff := absInt(entries[0].Rank - rank)
	for _, entry := range entries[1:] {
		diff := absInt(entry.Rank - rank)
		if diff < bestDiff {
			best = entry
			bestDiff = diff
		}
	}
	return best
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func sendRankingImage(ctx *zero.Ctx, client *renderer.Client, payload renderer.RankingListPayload) bool {
	return sendRankingImageWithTemplate(ctx, client, "ranking_list", payload)
}

func sendRankingImageWithTemplate(ctx *zero.Ctx, client *renderer.Client, template string, payload renderer.RankingListPayload) bool {
	if client == nil || !client.Health() {
		return false
	}
	png, err := client.Render(renderer.RenderRequest{Template: template, Data: payload})
	if err != nil {
		return false
	}
	ctx.SendChain(message.ImageBytes(png))
	return true
}

func formatRankingText(payload renderer.RankingListPayload) string {
	lines := []string{payload.Title}
	for _, entry := range payload.Rankings {
		line := fmt.Sprintf("#%d %s %sP", entry.Rank, entry.Name, formatInt64(entry.Score))
		if entry.ScoreDelta != 0 {
			line += fmt.Sprintf(" Δ%s", formatSignedInt64(entry.ScoreDelta))
		}
		if entry.Signature != "" {
			line += " · " + entry.Signature
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func formatInt64(value int64) string {
	return formatNumber(int(value))
}

func formatSignedInt64(value int64) string {
	if value > 0 {
		return "+" + formatInt64(value)
	}
	return formatInt64(value)
}

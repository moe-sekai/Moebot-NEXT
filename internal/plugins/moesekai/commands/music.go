package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/bot"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/musicsearch"
	"moebot-next/internal/renderer"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

// RegisterMusic registers music-related query commands.
func RegisterMusic(deps *Deps) {
	registerMusicDetailCommand(deps, "查曲")
	registerChartCommand(deps)
}

func registerMusicDetailCommand(deps *Deps, command string) {
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

			if keyword == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("请输入要搜索的曲目关键词~\n例: /%s 千本樱", commandName)))
				return
			}

			if strings.EqualFold(keyword, "leak") || keyword == "剧透" || keyword == "未来" {
				musics := filterLeakMusics(runtime.Store.AllMusics())
				if len(musics) == 0 {
					ctx.SendChain(message.Text("当前没有未发布曲目"))
					return
				}
				sendMusicList(ctx, deps, runtime.Store, runtime.Assets, recordCommand, runtime.Region, start, "未发布曲目", "leak", musics, 1, 1, len(musics))
				return
			}
			if ids, ok := parseMultipleMusicIDs(keyword); ok {
				musics := collectMusicsByIDs(runtime.Store, ids)
				if len(musics) == 0 {
					ctx.SendChain(message.Text("没有找到指定 ID 的曲目"))
					return
				}
				sendMusicList(ctx, deps, runtime.Store, runtime.Assets, recordCommand, runtime.Region, start, "曲目列表", keyword, musics, 1, 1, len(musics))
				return
			}

			result := musicsearch.Search(runtime.Store, runtime.MusicAliases, keyword, musicsearch.Options{Limit: musicsearch.DefaultListLimit})
			if result.Mode == musicsearch.ModeList {
				sendMusicList(ctx, deps, runtime.Store, runtime.Assets, recordCommand, runtime.Region, start, musicListTitleForQuery(keyword, "曲目候选列表"), keyword, result.Musics, result.Page, result.TotalPages, result.Total)
				return
			}
			if result.Music == nil {
				ctx.SendChain(message.Text(result.Message))
				return
			}
			payload := renderpayloads.BuildMusicDetailPayloadWithAssets(runtime.Store, *result.Music, runtime.Assets)
			if deps.Renderer != nil && deps.Renderer.Health() {
				png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "music_detail", Data: payload})
				if err == nil {
					if result.Message != "" {
						ctx.SendChain(message.ImageBytes(png), message.Text("\n"+result.Message))
					} else {
						ctx.SendChain(message.ImageBytes(png))
					}
					bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
					return
				}
			}

			text := formatMusicText(payload)
			if result.Message != "" {
				text += "\n" + result.Message
			}
			ctx.SendChain(message.Text(text))
			bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
		})
	}
}

func registerChartCommand(deps *Deps) {
	for _, cmd := range parserCommands(deps, "查谱") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		recordCommand := cmd.Primary
		if recordCommand == "" {
			recordCommand = "查谱"
		}
		Engine.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			keyword := commandArgs(ctx)
			runtime, _ := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || runtime.Store == nil || !runtime.Enabled {
				ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
				return
			}

			if keyword == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("请输入要搜索的谱面关键词~\n例: /%s 千本樱", commandName)))
				return
			}

			query, options := parseMusicQuery(keyword)
			if query == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("请输入要搜索的谱面关键词~\n例: /%s master 千本樱", commandName)))
				return
			}
			result := musicsearch.Search(runtime.Store, runtime.MusicAliases, query, musicsearch.Options{Difficulty: options.Difficulty, Limit: musicsearch.DefaultListLimit})
			if result.Mode == musicsearch.ModeList {
				sendMusicList(ctx, deps, runtime.Store, runtime.Assets, recordCommand, runtime.Region, start, musicListTitleForQuery(query, "谱面候选列表"), query, result.Musics, result.Page, result.TotalPages, result.Total)
				return
			}
			if result.Music == nil {
				ctx.SendChain(message.Text(result.Message))
				return
			}

			payload := renderpayloads.BuildMusicDetailPayloadWithAssets(runtime.Store, *result.Music, runtime.Assets)
			payload = selectedDifficultyPayload(payload, options.Difficulty, runtime.Profile.Assets.ChartSourceURL)
			if deps.Renderer != nil && deps.Renderer.Health() {
				if payload.ChartURL != "" {
					if png, err := deps.Renderer.RenderChartURL(payload.ChartURL); err == nil {
						ctx.SendChain(message.ImageBytes(png))
						bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
						return
					}
				}
				png, err := deps.Renderer.Render(buildChartRenderRequest(payload))
				if err == nil {
					if result.Message != "" {
						ctx.SendChain(message.ImageBytes(png), message.Text("\n"+result.Message))
					} else {
						ctx.SendChain(message.ImageBytes(png))
					}
					bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
					return
				}
			}

			text := formatChartText(payload)
			if result.Message != "" {
				text += "\n" + result.Message
			}
			ctx.SendChain(message.Text(text))
			bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
		})
	}
}

func sendMusicList(ctx *zero.Ctx, deps *Deps, store *masterdata.Store, resolver *assets.Resolver, recordCommand string, region string, start time.Time, title string, subtitle string, musics []masterdata.MusicInfo, page int, totalPages int, total int) {
	assetResolver := resolver
	if total <= 0 {
		total = len(musics)
	}
	if page <= 0 {
		page = 1
	}
	if totalPages <= 0 {
		totalPages = musicsearch.TotalPages(total, musicsearch.DefaultListLimit)
	}
	shown := musicsearch.LimitMusics(musics, musicsearch.DefaultListLimit)
	payload := renderpayloads.BuildMusicListPayloadWithAssets(title, subtitle, shown, store, assetResolver, page, totalPages, total)
	if deps.Renderer != nil && deps.Renderer.Health() {
		png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "music_list", Data: payload})
		if err == nil {
			ctx.SendChain(message.ImageBytes(png))
			bot.RecordCommandRegion(deps.DB, recordCommand, region, ctx, start)
			return
		}
	}
	ctx.SendChain(message.Text(formatMusicListText(payload)))
	bot.RecordCommandRegion(deps.DB, recordCommand, region, ctx, start)
}

func musicListTitleForQuery(query string, fallback string) string {
	if strings.HasPrefix(musicsearch.Normalize(query), "event") {
		return "活动关联曲目"
	}
	return fallback
}

func filterLeakMusics(musics []masterdata.MusicInfo) []masterdata.MusicInfo {
	now := time.Now().UnixMilli()
	out := make([]masterdata.MusicInfo, 0)
	for _, music := range musics {
		if music.PublishedAt > now {
			out = append(out, music)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PublishedAt != out[j].PublishedAt {
			return out[i].PublishedAt < out[j].PublishedAt
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func parseMultipleMusicIDs(keyword string) ([]int, bool) {
	fields := strings.Fields(keyword)
	if len(fields) <= 1 {
		return nil, false
	}
	ids := make([]int, 0, len(fields))
	for _, field := range fields {
		id, err := strconv.Atoi(field)
		if err != nil || id <= 0 {
			return nil, false
		}
		ids = append(ids, id)
	}
	return ids, true
}

func collectMusicsByIDs(store *masterdata.Store, ids []int) []masterdata.MusicInfo {
	out := make([]masterdata.MusicInfo, 0, len(ids))
	for _, id := range ids {
		if music := store.GetMusic(id); music != nil {
			out = append(out, *music)
		}
	}
	return out
}

func buildChartRenderRequest(payload renderpayloads.MusicDetailPayload) renderer.RenderRequest {
	return renderer.RenderRequest{
		Template: "chart_detail",
		Data:     payload,
	}
}

func formatMusicListText(payload renderpayloads.MusicListPayload) string {
	lines := []string{fmt.Sprintf("%s（共 %d 首）", payload.Title, payload.Total)}
	for _, music := range payload.Musics {
		lines = append(lines, fmt.Sprintf("#%d %s", music.ID, music.Title))
	}
	return strings.Join(lines, "\n")
}

func formatMusicText(music renderpayloads.MusicDetailPayload) string {
	lines := []string{
		fmt.Sprintf("曲目：%s", music.Title),
		fmt.Sprintf("ID：%d", music.ID),
	}
	if music.Pronunciation != "" {
		lines = append(lines, fmt.Sprintf("读音：%s", music.Pronunciation))
	}
	if music.Lyricist != "" {
		lines = append(lines, fmt.Sprintf("作词：%s", music.Lyricist))
	}
	if music.Composer != "" {
		lines = append(lines, fmt.Sprintf("作曲：%s", music.Composer))
	}
	if music.Arranger != "" {
		lines = append(lines, fmt.Sprintf("编曲：%s", music.Arranger))
	}
	if len(music.Difficulties) > 0 {
		diffs := make([]string, 0, len(music.Difficulties))
		for _, d := range music.Difficulties {
			diffs = append(diffs, fmt.Sprintf("%s Lv.%d/%d notes", d.MusicDifficulty, d.PlayLevel, d.TotalNoteCount))
		}
		lines = append(lines, "难度："+strings.Join(diffs, "，"))
	}
	return strings.Join(lines, "\n")
}

func formatChartText(music renderpayloads.MusicDetailPayload) string {
	lines := []string{
		fmt.Sprintf("谱面：%s", music.Title),
		fmt.Sprintf("ID：%d", music.ID),
	}
	if len(music.Categories) > 0 {
		lines = append(lines, "分类："+strings.Join(music.Categories, " / "))
	}
	if len(music.Difficulties) == 0 {
		lines = append(lines, "暂无谱面难度数据")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "难度：")
	for _, d := range music.Difficulties {
		lines = append(lines, fmt.Sprintf("%s：Lv.%d · %d notes", strings.ToUpper(d.MusicDifficulty), d.PlayLevel, d.TotalNoteCount))
	}
	return strings.Join(lines, "\n")
}

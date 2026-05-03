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

// RegisterMusic registers music-related query commands.
func RegisterMusic(deps *Deps) {
	registerMusicDetailCommand(deps, "查曲")
	registerMusicDetailCommand(deps, "查歌")
	registerChartCommand(deps)
}

func registerMusicDetailCommand(deps *Deps, command string) {
	for _, cmd := range regionalCommands(command) {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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

			results := runtime.Store.SearchMusics(keyword)
			if len(results) == 0 {
				ctx.SendChain(message.Text(fmt.Sprintf("没有找到与「%s」匹配的曲目", keyword)))
				return
			}

			music := results[0]
			payload := renderer.BuildMusicDetailPayloadWithAssets(runtime.Store, music, runtime.Assets)

			if deps.Renderer != nil && deps.Renderer.Health() {
				png, err := deps.Renderer.Render(renderer.RenderRequest{
					Template: "music_detail",
					Data:     payload,
				})
				if err == nil {
					ctx.SendChain(message.ImageBytes(png))
					bot.RecordCommandRegion(deps.DB, command, runtime.Region, ctx, start)
					return
				}
			}

			ctx.SendChain(message.Text(formatMusicText(payload)))
			bot.RecordCommandRegion(deps.DB, command, runtime.Region, ctx, start)
		})
	}
}

func registerChartCommand(deps *Deps) {
	for _, cmd := range regionalCommands("查谱") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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

			results := runtime.Store.SearchMusics(keyword)
			if len(results) == 0 {
				ctx.SendChain(message.Text(fmt.Sprintf("没有找到与「%s」匹配的谱面", keyword)))
				return
			}

			payload := renderer.BuildMusicDetailPayloadWithAssets(runtime.Store, results[0], runtime.Assets)
			if deps.Renderer != nil && deps.Renderer.Health() {
				png, err := deps.Renderer.Render(buildChartRenderRequest(payload))
				if err == nil {
					ctx.SendChain(message.ImageBytes(png))
					bot.RecordCommandRegion(deps.DB, "查谱", runtime.Region, ctx, start)
					return
				}
			}

			ctx.SendChain(message.Text(formatChartText(payload)))
			bot.RecordCommandRegion(deps.DB, "查谱", runtime.Region, ctx, start)
		})
	}
}

func buildChartRenderRequest(payload renderer.MusicDetailPayload) renderer.RenderRequest {
	return renderer.RenderRequest{
		Template: "chart_detail",
		Data:     payload,
	}
}

func formatMusicText(music renderer.MusicDetailPayload) string {
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

func formatChartText(music renderer.MusicDetailPayload) string {
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

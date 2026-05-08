package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/plugins/moesekai/b30"
	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/renderer"
	"moebot-next/internal/plugins/moesekai/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

type best30Profile struct {
	suite.BaseProfile
	UserGamedata     suite.UserGamedata    `json:"userGamedata"`
	UserDecks        []suite.UserDeck      `json:"userDecks"`
	UserCards        []suite.UserCard      `json:"userCards"`
	UserMusicResults []b30.UserMusicResult `json:"userMusicResults"`
	UserMusics       []b30.LegacyUserMusic `json:"userMusics"`
}

func best30Fields() []string {
	return suite.Fields(suite.FieldUserMusicResults, suite.FieldUserMusics)
}

func RegisterBest30(deps *Deps) {
	for _, cmd := range parserCommands(deps, "best30") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, inferredUser, ok := requireRuntime(deps, ctx, forcedRegion)
			if !ok {
				return
			}
			user, ok := requireBoundUser(deps, ctx, runtime, forcedRegion, inferredUser)
			if !ok {
				return
			}
			if !requireSuite(ctx, runtime, "Best30") {
				return
			}
			if _, ok := requireSuiteVisible(deps, ctx, runtime); !ok {
				return
			}

			var profile best30Profile
			if !fetchSuiteUserData(ctx, runtime, user.GameID, "Best30", best30Fields(), &profile) {
				return
			}
			result, constantsSource, err := buildBest30Result(deps, profile, renderpayloads.Best30MusicMetaResolver(runtime.Store, runtime.Assets))
			if err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取 Best30 社区定数失败\n%s", err.Error())))
				return
			}
			if len(result.Entries) == 0 {
				ctx.SendChain(message.Text(formatBest30EmptyText(runtime.Region, profile, result)))
				return
			}
			payload := renderpayloads.BuildBest30Payload(best30Title(runtime.Region), runtime.Region, profile.BaseProfile, profile.UserGamedata, result, runtime.Store, runtime.Assets, constantsSource)
			sendBest30OrText(ctx, deps, payload, formatBest30Text(runtime.Region, profile, result))
			bot.RecordCommandRegion(deps.DB, "Best30", runtime.Region, ctx, start)
		})
	}
}

func buildBest30Result(deps *Deps, profile best30Profile, resolver b30.MetaResolver) (b30.Result, string, error) {
	client := best30Client(deps)
	table, err := client.Get(context.Background())
	if err != nil {
		return b30.Result{}, client.URL(), err
	}
	results := b30.MergeLegacyResults(profile.UserMusicResults, profile.UserMusics)
	return b30.Calculate(results, table, resolver), client.URL(), nil
}

func best30Client(deps *Deps) *b30.Client {
	if deps != nil && deps.B30 != nil {
		return deps.B30
	}
	return b30.NewClient(config.B30Config{ConstantsURL: config.DefaultB30ConstantsURL, Timeout: 10, RefreshInterval: 21600})
}

func best30Title(region string) string {
	region = config.NormalizeRegion(region)
	if region == "" {
		region = config.RegionJP
	}
	return fmt.Sprintf("%s Best30", strings.ToUpper(region))
}

func sendBest30OrText(ctx *zero.Ctx, deps *Deps, payload renderpayloads.Best30Payload, fallback string) {
	if deps != nil && deps.Renderer != nil && deps.Renderer.Health() {
		png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "best30", Data: payload, Precision: 1.2})
		if err == nil {
			ctx.SendChain(message.ImageBytes(png))
			return
		}
	}
	ctx.SendChain(message.Text(fallback))
}

func formatBest30EmptyText(region string, profile best30Profile, result b30.Result) string {
	lines := []string{
		fmt.Sprintf("%s Best30", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", best30PlayerName(profile)),
		fmt.Sprintf("更新时间: %s", best30UpdateText(profile.UploadTime)),
		"暂无可计入 Best30 的 AP/FC 谱面。",
	}
	if result.MissingConstantsCount > 0 {
		lines = append(lines, fmt.Sprintf("有 %d 个 AP/FC 谱面缺少社区定数。", result.MissingConstantsCount))
	}
	return strings.Join(lines, "\n")
}

func formatBest30Text(region string, profile best30Profile, result b30.Result) string {
	lines := []string{
		fmt.Sprintf("%s Best30: %.2f", strings.ToUpper(config.NormalizeRegion(region)), result.Average),
		fmt.Sprintf("玩家: %s", best30PlayerName(profile)),
		fmt.Sprintf("更新时间: %s", best30UpdateText(profile.UploadTime)),
		fmt.Sprintf("计入: %d首 | AP %d | FC %d", len(result.Entries), result.APCount, result.FCCount),
	}
	if result.MissingConstantsCount > 0 {
		lines = append(lines, fmt.Sprintf("缺少定数: %d", result.MissingConstantsCount))
	}
	lines = append(lines, "---")
	for _, entry := range result.Entries {
		lines = append(lines, fmt.Sprintf("#%02d %.1f %s %s %.1f · %s", entry.Rank, entry.UserRating, strings.ToUpper(entry.Difficulty), entry.PlayResult, entry.Constant, entry.Title))
	}
	lines = append(lines, "---", "公式: AP=定数；FC=定数-1(≥33) / 定数-1.5(<33)")
	return strings.Join(lines, "\n")
}

func best30PlayerName(profile best30Profile) string {
	name := strings.TrimSpace(profile.UserGamedata.Name)
	if name == "" {
		return "未知玩家"
	}
	return name
}

func best30UpdateText(uploadTime int64) string {
	if uploadTime > 0 && uploadTime < 100000000000 {
		uploadTime *= 1000
	}
	if uploadTime <= 0 {
		return "未知"
	}
	return time.UnixMilli(uploadTime).Format("2006-01-02 15:04:05")
}

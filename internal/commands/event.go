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

			if keyword == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("请输入要搜索的活动关键词~\n例: /%s 周年", commandName)))
				return
			}

			results := runtime.Store.SearchEvents(keyword)
			if len(results) == 0 {
				ctx.SendChain(message.Text(fmt.Sprintf("没有找到与「%s」匹配的活动", keyword)))
				return
			}

			event := results[0]
			payload := renderer.BuildEventInfoPayloadWithAssets(runtime.Store, event, runtime.Assets)

			if deps.Renderer != nil && deps.Renderer.Health() {
				png, err := deps.Renderer.Render(renderer.RenderRequest{
					Template: "event_info",
					Data:     payload,
				})
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

func formatEventText(event renderer.EventInfoPayload) string {
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
	return strings.Join(lines, "\n")
}

func formatMillis(ms int64) string {
	if ms <= 0 {
		return "-"
	}
	return time.Unix(0, ms*int64(time.Millisecond)).Format("2006-01-02 15:04")
}

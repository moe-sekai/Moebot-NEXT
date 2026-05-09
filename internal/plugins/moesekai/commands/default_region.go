package commands

import (
	"fmt"
	"strings"

	"moebot-next/internal/config"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// defaultRegionCommands enumerates the command names handled by /pjsk服务器.
// They share a single handler. Aliases are intentionally typed without the
// leading slash to match ZeroBot's OnCommand contract.
var defaultRegionCommands = []string{
	"pjsk服务器", "pjsk区服", "pjsk默认服务器", "pjsk默认区服",
}

// RegisterDefaultRegion registers the /pjsk服务器 command which lets a user
// query or set their default game region. Once set, region-prefix-less
// moesekai commands will use this region by default.
func RegisterDefaultRegion(deps *Deps) {
	for _, name := range defaultRegionCommands {
		Engine.OnCommand(name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			handleDefaultRegion(deps, ctx)
		})
	}
}

func handleDefaultRegion(deps *Deps, ctx *zero.Ctx) {
	platformID := userIDFromCtx(ctx)
	args := strings.TrimSpace(commandArgs(ctx))
	available := availableRegionsHint()
	if args == "" {
		current, _ := deps.DB.GetUserDefaultRegion("onebot", platformID)
		if current == "" {
			ctx.SendChain(message.Text(fmt.Sprintf(
				"你尚未设置默认区服，当前按规则回退到日服 (JP)。\n可选: %s\n用法: /pjsk服务器 <区服>",
				available,
			)))
			return
		}
		ctx.SendChain(message.Text(fmt.Sprintf(
			"当前默认区服: %s (%s)\n可选: %s\n用法: /pjsk服务器 <区服>",
			config.RegionLabel(current), strings.ToUpper(current), available,
		)))
		return
	}

	region := config.NormalizeRegion(args)
	if region == "" || !config.IsValidRegion(region) {
		ctx.SendChain(message.Text(fmt.Sprintf(
			"无效的区服 %q。可选: %s",
			args, available,
		)))
		return
	}
	if err := deps.DB.SetUserDefaultRegion("onebot", platformID, region); err != nil {
		ctx.SendChain(message.Text("设置默认区服失败，请稍后重试"))
		return
	}
	ctx.SendChain(message.Text(fmt.Sprintf(
		"已将默认区服设为: %s (%s)\n之后无前缀的指令会自动使用此区服。",
		config.RegionLabel(region), strings.ToUpper(region),
	)))
}

func availableRegionsHint() string {
	keys := config.RegionKeys()
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s(%s)", key, config.RegionLabel(key)))
	}
	return strings.Join(parts, " / ")
}

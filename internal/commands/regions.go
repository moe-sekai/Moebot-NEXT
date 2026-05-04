package commands

import (
	"fmt"
	"strings"

	"moebot-next/internal/commandparser"
	"moebot-next/internal/config"
	"moebot-next/internal/models"
	"moebot-next/internal/servers"

	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/gorm"
)

type regionalCommand struct {
	Name        string
	Region      string
	Base        string
	Primary     string
	MatchSource string
	WorldLink   bool
}

var regionalPrefixes = []string{config.RegionJP, config.RegionCN, config.RegionTW, config.RegionKR, config.RegionEN}

func regionalCommands(command string) []regionalCommand {
	commands := []regionalCommand{{Name: command, Base: command, Primary: command}}
	for _, region := range regionalPrefixes {
		commands = append(commands, regionalCommand{Name: region + command, Region: region, Base: command, Primary: command})
	}
	return commands
}

func parserCommands(deps *Deps, primary string) []regionalCommand {
	if deps == nil || len(deps.Definitions) == 0 {
		return regionalCommands(primary)
	}
	for _, def := range deps.Definitions {
		if def.PrimaryCommand != primary {
			continue
		}
		botCommands := commandparser.BotCommandsFor(def)
		out := make([]regionalCommand, 0, len(botCommands))
		for _, cmd := range botCommands {
			out = append(out, regionalCommand{
				Name:        cmd.Name,
				Region:      cmd.Region,
				Base:        cmd.Base,
				Primary:     cmd.Primary,
				MatchSource: cmd.MatchSource,
			})
		}
		return out
	}
	return regionalCommands(primary)
}

func parseRegionalCommandName(command string) (base string, region string) {
	for _, prefix := range regionalPrefixes {
		if strings.HasPrefix(command, prefix) && len(command) > len(prefix) {
			return strings.TrimPrefix(command, prefix), prefix
		}
	}
	return command, ""
}

func regionLabel(region string) string {
	return config.RegionLabel(region)
}

func normalizeRequestedRegion(region string) string {
	region = config.NormalizeRegion(region)
	if region == "" || !config.IsValidRegion(region) {
		return ""
	}
	return region
}

func userIDFromCtx(ctx *zero.Ctx) string {
	return fmt.Sprintf("%d", ctx.Event.UserID)
}

func runtimeForCommand(deps *Deps, ctx *zero.Ctx, forcedRegion string) (*servers.Runtime, *models.User) {
	region := normalizeRequestedRegion(forcedRegion)
	if region != "" {
		return deps.Servers.Get(region), nil
	}
	user, err := deps.DB.GetUserByPlatform("onebot", userIDFromCtx(ctx))
	if err != nil && err != gorm.ErrRecordNotFound {
		return deps.Servers.Default(), nil
	}
	if user != nil && user.ServerRegion == "" {
		user.ServerRegion = config.RegionJP
	}
	return deps.Servers.ForUser(user), user
}

func runtimeForRegion(deps *Deps, region string) *servers.Runtime {
	region = normalizeRequestedRegion(region)
	if region == "" {
		region = config.RegionJP
	}
	return deps.Servers.Get(region)
}

func commandArgs(ctx *zero.Ctx) string {
	return strings.TrimSpace(fmt.Sprintf("%v", ctx.State["args"]))
}

func runtimeUnavailableText(runtime *servers.Runtime) string {
	if runtime == nil {
		return "服务器配置不可用"
	}
	if !runtime.Enabled {
		return fmt.Sprintf("%s 暂未启用，请在管理面板 /settings 开启", runtime.Label)
	}
	return "服务器配置不可用"
}

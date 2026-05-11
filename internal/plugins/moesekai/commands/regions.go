package commands

import (
	"fmt"
	"sort"
	"strings"

	"moebot-next/internal/config"
	"moebot-next/internal/models"
	"moebot-next/internal/plugins/moesekai/commandparser"
	"moebot-next/internal/plugins/moesekai/servers"

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
		return sortRegionalCommandsByLength(regionalCommands(primary))
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
		return sortRegionalCommandsByLength(out)
	}
	return sortRegionalCommandsByLength(regionalCommands(primary))
}

// sortRegionalCommandsByLength sorts commands by name length descending so that
// longer aliases get registered to ZeroBot before shorter prefixes. ZeroBot's
// OnCommand uses HasPrefix matching, so without this the alias "谱面" would
// shadow "谱面预览" and parse `/谱面预览 123` as command="谱面" args="预览 123".
func sortRegionalCommandsByLength(cmds []regionalCommand) []regionalCommand {
	sort.SliceStable(cmds, func(i, j int) bool {
		if len(cmds[i].Name) != len(cmds[j].Name) {
			return len(cmds[i].Name) > len(cmds[j].Name)
		}
		return cmds[i].Name < cmds[j].Name
	})
	return cmds
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
		return deps.Servers.GetExact(region), nil
	}
	platformID := userIDFromCtx(ctx)
	// 优先使用用户通过 /pjsk服务器 设置的默认区服。
	if def, _ := deps.DB.GetUserDefaultRegion("onebot", platformID); def != "" {
		runtime := deps.Servers.GetExact(def)
		user, err := deps.DB.GetUserByPlatformRegion("onebot", platformID, def)
		if err != nil && err != gorm.ErrRecordNotFound {
			return runtime, nil
		}
		if runtime != nil {
			return runtime, user
		}
	}
	user, err := deps.DB.GetUserByPlatform("onebot", platformID)
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

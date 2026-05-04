package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

const leaderCountDefaultLimit = 26

type leaderCountProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata         `json:"userGamedata"`
	Missions     []characterMissionV2       `json:"userCharacterMissionV2s"`
	Statuses     []characterMissionV2Status `json:"userCharacterMissionV2Statuses"`
}

type characterMissionV2 struct {
	CharacterID          int    `json:"characterId"`
	CharacterMissionType string `json:"characterMissionType"`
	Progress             int    `json:"progress"`
}

type characterMissionV2Status struct {
	CharacterID      int    `json:"characterId"`
	ParameterGroupID int    `json:"parameterGroupId"`
	Seq              int    `json:"seq"`
	MissionStatus    string `json:"missionStatus"`
}

type leaderCountRow struct {
	CharacterID int
	PlayLive    int
	PlayLiveEx  int
}

func leaderCountFields() []string {
	return suite.Fields(suite.FieldUserCharacterMissionV2s, suite.FieldUserCharacterMissionV2Statuses)
}

func RegisterLeaderCount(deps *Deps) {
	for _, cmd := range parserCommands(deps, "队长次数") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, user := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || !runtime.Enabled {
				ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
				return
			}
			if forcedRegion != "" {
				var err error
				user, err = deps.DB.GetUserByPlatformRegion("onebot", userIDFromCtx(ctx), runtime.Region)
				if err != nil && err != gorm.ErrRecordNotFound {
					ctx.SendChain(message.Text("数据库错误，请稍后重试"))
					return
				}
			}
			if user == nil || user.GameID == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("你还没有绑定%s游戏账号~\n使用 /%s绑定 [游戏ID] 来绑定", runtime.Label, runtime.Region)))
				return
			}
			if runtime.Suite == nil || !runtime.Suite.Enabled() {
				ctx.SendChain(message.Text(fmt.Sprintf("暂不支持查询%s的抓包数据", runtime.Label)))
				return
			}
			setting := suiteSettingOrDefault(deps, userIDFromCtx(ctx), runtime.Region, runtime.Profile.SuiteAPI.DefaultMode)
			if setting.Hidden {
				ctx.SendChain(message.Text(fmt.Sprintf("你已隐藏%s抓包信息，发送 /%s展示抓包 可重新展示", runtime.Label, runtime.Region)))
				return
			}
			var profile leaderCountProfile
			if err := runtime.Suite.GetUserData(user.GameID, setting.Mode, leaderCountFields(), &profile); err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取你的%sSuite抓包数据失败，发送 /抓包 获取帮助\n%s", runtime.Label, err.Error())))
				return
			}
			ctx.SendChain(message.Text(formatLeaderCountText(runtime.Region, profile, leaderCountDefaultLimit)))
			bot.RecordCommandRegion(deps.DB, "队长次数", runtime.Region, ctx, start)
		})
	}
}

func formatLeaderCountText(region string, profile leaderCountProfile, limit int) string {
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
	rowsByCharacter := map[int]*leaderCountRow{}
	for _, mission := range profile.Missions {
		if mission.CharacterID <= 0 {
			continue
		}
		row := rowsByCharacter[mission.CharacterID]
		if row == nil {
			row = &leaderCountRow{CharacterID: mission.CharacterID}
			rowsByCharacter[mission.CharacterID] = row
		}
		switch mission.CharacterMissionType {
		case "play_live":
			row.PlayLive = max(row.PlayLive, mission.Progress)
		case "play_live_ex":
			row.PlayLiveEx = max(row.PlayLiveEx, mission.Progress)
		}
	}
	rows := make([]leaderCountRow, 0, len(rowsByCharacter))
	for _, row := range rowsByCharacter {
		if row.PlayLive == 0 && row.PlayLiveEx == 0 {
			continue
		}
		rows = append(rows, *row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].PlayLive == rows[j].PlayLive {
			return rows[i].PlayLiveEx > rows[j].PlayLiveEx
		}
		return rows[i].PlayLive > rows[j].PlayLive
	})
	if limit <= 0 || limit > len(rows) {
		limit = len(rows)
	}
	lines := []string{
		fmt.Sprintf("%s 队长次数", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("更新时间: %s", suiteUpdateText(profile.UploadTime)),
		fmt.Sprintf("数据来源: %s", suiteSourceText(profile.BaseProfile)),
	}
	if len(rows) == 0 {
		lines = append(lines, "暂无队长次数数据")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "---")
	for i := 0; i < limit; i++ {
		row := rows[i]
		lines = append(lines, fmt.Sprintf("%d. %s: %d | EX %d", i+1, characterDisplayName(row.CharacterID), row.PlayLive, row.PlayLiveEx))
	}
	return strings.Join(lines, "\n")
}

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

const challengeDefaultLimit = 26

type challengeProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	Results      []challengeResult  `json:"userChallengeLiveSoloResults"`
	Stages       []challengeStage   `json:"userChallengeLiveSoloStages"`
	Rewards      []challengeReward  `json:"userChallengeLiveSoloHighScoreRewards"`
}

type challengeResult struct {
	CharacterID int `json:"characterId"`
	HighScore   int `json:"highScore"`
}

type challengeStage struct {
	CharacterID int `json:"characterId"`
	Rank        int `json:"rank"`
}

type challengeReward struct {
	CharacterID int `json:"characterId"`
	RewardID    int `json:"challengeLiveHighScoreRewardId"`
}

type challengeSummaryRow struct {
	CharacterID int
	HighScore   int
	Rank        int
	RewardCount int
}

func challengeFields() []string {
	return suite.Fields(suite.FieldUserChallengeLiveSoloResults, suite.FieldUserChallengeLiveSoloStages, suite.FieldUserChallengeLiveSoloHighScoreRewards)
}

func RegisterChallengeInfo(deps *Deps) {
	for _, cmd := range parserCommands(deps, "挑战信息") {
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
			var profile challengeProfile
			if err := runtime.Suite.GetUserData(user.GameID, setting.Mode, challengeFields(), &profile); err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取你的%sSuite抓包数据失败，发送 /抓包 获取帮助\n%s", runtime.Label, err.Error())))
				return
			}
			ctx.SendChain(message.Text(formatChallengeText(runtime.Region, profile, challengeDefaultLimit)))
			bot.RecordCommandRegion(deps.DB, "挑战信息", runtime.Region, ctx, start)
		})
	}
}

func formatChallengeText(region string, profile challengeProfile, limit int) string {
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
	rowsByCharacter := map[int]*challengeSummaryRow{}
	for _, result := range profile.Results {
		row := challengeRow(rowsByCharacter, result.CharacterID)
		row.HighScore = max(row.HighScore, result.HighScore)
	}
	for _, stage := range profile.Stages {
		row := challengeRow(rowsByCharacter, stage.CharacterID)
		row.Rank = max(row.Rank, stage.Rank)
	}
	for _, reward := range profile.Rewards {
		row := challengeRow(rowsByCharacter, reward.CharacterID)
		row.RewardCount++
	}
	rows := make([]challengeSummaryRow, 0, len(rowsByCharacter))
	for _, row := range rowsByCharacter {
		if row.HighScore == 0 && row.Rank == 0 && row.RewardCount == 0 {
			continue
		}
		rows = append(rows, *row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].HighScore == rows[j].HighScore {
			return rows[i].Rank > rows[j].Rank
		}
		return rows[i].HighScore > rows[j].HighScore
	})
	if limit <= 0 || limit > len(rows) {
		limit = len(rows)
	}
	lines := []string{
		fmt.Sprintf("%s 挑战信息", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("更新时间: %s", suiteUpdateText(profile.UploadTime)),
		fmt.Sprintf("数据来源: %s", suiteSourceText(profile.BaseProfile)),
	}
	if len(rows) == 0 {
		lines = append(lines, "暂无挑战数据")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "---")
	for i := 0; i < limit; i++ {
		row := rows[i]
		lines = append(lines, fmt.Sprintf("%d. %s: %d | Lv.%d | 奖励 %d", i+1, characterDisplayName(row.CharacterID), row.HighScore, row.Rank, row.RewardCount))
	}
	return strings.Join(lines, "\n")
}

func challengeRow(rows map[int]*challengeSummaryRow, characterID int) *challengeSummaryRow {
	row := rows[characterID]
	if row == nil {
		row = &challengeSummaryRow{CharacterID: characterID}
		rows[characterID] = row
	}
	return row
}

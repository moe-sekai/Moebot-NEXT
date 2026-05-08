package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

const (
	challengeDefaultLimit         = 26
	challengeLiveHighScorePurpose = "challenge_live_high_score"
)

type challengeProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	UserDecks    []suite.UserDeck   `json:"userDecks"`
	UserCards    []suite.UserCard   `json:"userCards"`
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
	CharacterID                        int `json:"characterId"`
	GameCharacterID                    int `json:"gameCharacterId"`
	RewardID                           int `json:"challengeLiveHighScoreRewardId"`
	ChallengeLiveSoloHighScoreRewardID int `json:"challengeLiveSoloHighScoreRewardId"`
	RewardIDAlias                      int `json:"rewardId"`
}

type challengeSummaryRow struct {
	CharacterID    int
	HighScore      int
	Rank           int
	RewardCount    int
	RemainJewel    int
	RemainFragment int
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
			runtime, inferredUser, ok := requireRuntime(deps, ctx, forcedRegion)
			if !ok {
				return
			}
			user, ok := requireBoundUser(deps, ctx, runtime, forcedRegion, inferredUser)
			if !ok {
				return
			}
			if !requireSuite(ctx, runtime, "挑战信息") {
				return
			}
			if _, ok := requireSuiteVisible(deps, ctx, runtime); !ok {
				return
			}
			var profile challengeProfile
			if !fetchSuiteUserData(ctx, runtime, user.GameID, "挑战信息", challengeFields(), &profile) {
				return
			}
			payload := buildSuitePanel(runtime, suitePanelTitle(runtime, "挑战信息"), "", profile)
			payload.Subtitle = suitePanelSubtitle(profile.BaseProfile)
			rows, stats, sectionExtra := rowsFromChallenge(profile, runtime.Store, challengeDefaultLimit)
			payload.Stats = append(suiteBasicStats(profile.commonSuiteProfile()), stats...)
			payload.Sections = []renderpayloads.SuiteSectionPayload{{Title: "每日挑战 Live", Kind: "challenge_info", Note: "参考 lunabot：按角色统计挑战等级、最高分，以及未领取高分奖励中的水晶/碎片。", Rows: rows, Extra: sectionExtra}}
			sendSuitePanelOrText(ctx, deps, payload, formatChallengeTextWithStore(runtime.Region, profile, runtime.Store, challengeDefaultLimit))
			bot.RecordCommandRegion(deps.DB, "挑战信息", runtime.Region, ctx, start)
		})
	}
}

func formatChallengeText(region string, profile challengeProfile, limit int) string {
	return formatChallengeTextWithStore(region, profile, nil, limit)
}

func formatChallengeTextWithStore(region string, profile challengeProfile, store *masterdata.Store, limit int) string {
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
	rowsByCharacter := challengeRows(profile, store)
	completed := challengeCompletedRewardIDs(profile.Rewards, store)
	totalJewel, totalFragment, totalRemainRewards := 0, 0, 0
	rankCounts := map[int]int{}
	rows := make([]challengeSummaryRow, 0, len(rowsByCharacter))
	for cid := 1; cid <= 26; cid++ {
		row := rowsByCharacter[cid]
		if row == nil {
			row = &challengeSummaryRow{CharacterID: cid}
		}
		row.RemainJewel, row.RemainFragment = challengeRemainRewards(store, cid, completed[cid])
		totalJewel += row.RemainJewel
		totalFragment += row.RemainFragment
		totalRemainRewards += challengeRemainRewardCount(store, cid, completed[cid])
		rankCounts[row.Rank]++
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
	if store != nil {
		lines = append(lines,
			fmt.Sprintf("剩余总量: 水晶 %d | 碎片 %d | 奖励档 %d", totalJewel, totalFragment, totalRemainRewards),
			fmt.Sprintf("挑战等级分布: %s", formatChallengeRankDistributionText(rankCounts)),
		)
	}
	if len(rows) == 0 {
		lines = append(lines, "暂无挑战数据")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "---")
	for i := 0; i < limit; i++ {
		row := rows[i]
		parts := []string{fmt.Sprintf("%d. %s: %d | Lv.%d | 奖励 %d", i+1, characterDisplayName(row.CharacterID), row.HighScore, row.Rank, row.RewardCount)}
		if store != nil {
			remainRewards := challengeRemainRewardCount(store, row.CharacterID, completed[row.CharacterID])
			parts = append(parts, fmt.Sprintf("剩余档 %d | 水晶 %d | 碎片 %d", remainRewards, row.RemainJewel, row.RemainFragment))
		}
		lines = append(lines, strings.Join(parts, " | "))
	}
	return strings.Join(lines, "\n")
}

func formatChallengeRankDistributionText(counts map[int]int) string {
	if len(counts) == 0 {
		return "-"
	}
	levels := make([]int, 0, len(counts))
	for level := range counts {
		levels = append(levels, level)
	}
	sort.SliceStable(levels, func(i, j int) bool { return levels[i] > levels[j] })
	parts := make([]string, 0, len(levels))
	for _, level := range levels {
		parts = append(parts, fmt.Sprintf("%s×%d", challengeRankLabel(level), counts[level]))
	}
	return strings.Join(parts, " / ")
}

func challengeRow(rows map[int]*challengeSummaryRow, characterID int) *challengeSummaryRow {
	row := rows[characterID]
	if row == nil {
		row = &challengeSummaryRow{CharacterID: characterID}
		rows[characterID] = row
	}
	return row
}

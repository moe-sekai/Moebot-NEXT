package commands

import (
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
)

const musicOverviewDefaultLimit = 10
const musicRewardDefaultLimit = musicOverviewDefaultLimit

type musicOverviewProfile struct {
	suite.BaseProfile
	UserGamedata     suite.UserGamedata `json:"userGamedata"`
	UserDecks        []suite.UserDeck   `json:"userDecks"`
	UserCards        []suite.UserCard   `json:"userCards"`
	UserMusicResults []userMusicResult  `json:"userMusicResults"`
	Achievements     []musicAchievement `json:"userMusicAchievements"`
}

type musicProgressProfile = musicOverviewProfile
type musicRewardProfile = musicOverviewProfile

type userMusicResult struct {
	MusicID             int    `json:"musicId"`
	MusicDifficulty     string `json:"musicDifficulty"`
	MusicDifficultyType string `json:"musicDifficultyType"`
	PlayResult          string `json:"playResult"`
	FullComboFlg        bool   `json:"fullComboFlg"`
	FullPerfectFlg      bool   `json:"fullPerfectFlg"`
}

type musicProgressCount struct {
	Total      int
	Played     int
	Clear      int
	FullCombo  int
	AllPerfect int
}

type musicAchievement struct {
	MusicID            int `json:"musicId"`
	MusicAchievementID int `json:"musicAchievementId"`
}

type musicRewardRow struct {
	MusicID int
	Count   int
}

func musicOverviewFields() []string {
	return suite.Fields(suite.FieldUserMusicResults, suite.FieldUserMusicAchievements)
}

func musicProgressFields() []string { return musicOverviewFields() }
func musicRewardFields() []string   { return musicOverviewFields() }

func RegisterMusicOverview(deps *Deps) {
	for _, cmd := range parserCommands(deps, "打歌进度") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		baseCommand := cmd.Base
		Engine.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, inferredUser, ok := requireRuntime(deps, ctx, forcedRegion)
			if !ok {
				return
			}
			user, ok := requireBoundUser(deps, ctx, runtime, forcedRegion, inferredUser)
			if !ok {
				return
			}
			if !requireSuite(ctx, runtime, "打歌进度") {
				return
			}
			if _, ok := requireSuiteVisible(deps, ctx, runtime); !ok {
				return
			}
			var profile musicOverviewProfile
			if !fetchSuiteUserData(ctx, runtime, user.GameID, "打歌进度", musicOverviewFields(), &profile) {
				return
			}
			payload := buildSuitePanel(runtime, suitePanelTitle(runtime, "打歌进度 / 歌曲奖励"), "", profile)
			payload.Subtitle = suitePanelSubtitle(profile.BaseProfile)
			sections, stats := sectionsFromMusicOverview(profile, runtime.Store, musicOverviewDefaultLimit)
			payload.Stats = append(suiteBasicStats(profile.commonSuiteProfile()), stats...)
			payload.Sections = sections
			sendSuitePanelOrText(ctx, deps, payload, formatMusicOverviewText(runtime.Region, profile, runtime.Store, musicOverviewDefaultLimit))
			bot.RecordCommandRegion(deps.DB, musicOverviewRecordName(baseCommand), runtime.Region, ctx, start)
		})
	}
}

func musicOverviewRecordName(base string) string {
	switch strings.TrimSpace(base) {
	case "歌曲奖励", "打歌奖励", "歌曲挖矿", "打歌挖矿":
		return "歌曲奖励"
	default:
		return "打歌进度"
	}
}

func formatMusicOverviewText(region string, profile musicOverviewProfile, store *masterdata.Store, limit int) string {
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
	counts := musicProgressCounts(profile)
	summary := musicRewardSummaryFromProfile(profile, store, limit)
	lines := []string{
		fmt.Sprintf("%s 打歌进度 / 歌曲奖励", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("更新时间: %s", suiteUpdateText(profile.UploadTime)),
		fmt.Sprintf("数据来源: %s", suiteSourceText(profile.BaseProfile)),
	}
	if len(counts) == 0 && len(profile.Achievements) == 0 && summary.ValidMusicCount == 0 {
		lines = append(lines, "暂无打歌数据 / 歌曲奖励数据")
		return strings.Join(lines, "\n")
	}
	if len(counts) > 0 {
		lines = append(lines, "---", "打歌进度")
		for _, diff := range musicDifficultyOrder() {
			count := counts[diff]
			if count == nil {
				continue
			}
			lines = append(lines, fmt.Sprintf("%s: 游玩 %d | Clear %d | FC %d | AP %d", strings.ToUpper(diff), count.Played, count.Clear, count.FullCombo, count.AllPerfect))
		}
	}
	lines = append(lines, "---", "歌曲奖励")
	lines = append(lines, fmt.Sprintf("剩余总量: 水晶 %d | 碎片 %d", summary.TotalJewelRemain, summary.TotalShardRemain))
	lines = append(lines, fmt.Sprintf("S评级剩余: %d水晶（%d/%d首未达成）", summary.RankJewelRemain, summary.RankRemainCount, summary.ValidMusicCount))
	lines = append(lines, fmt.Sprintf("连击奖励剩余: %s", formatRewardTotal(summary.ComboRows)))
	lines = append(lines, fmt.Sprintf("已达成奖励数: %d | 涉及歌曲数: %d", summary.AchievementTotal, summary.AchievedMusicCount))
	if len(summary.TopRows) > 0 {
		lines = append(lines, "---", "已达成奖励 TOP")
		for i, row := range summary.TopRows {
			if i >= clampLimit(limit, len(summary.TopRows)) {
				break
			}
			lines = append(lines, fmt.Sprintf("%d. %s: %s", i+1, row.Label, row.Value))
		}
	}
	return strings.Join(lines, "\n")
}

func formatMusicProgressText(region string, profile musicProgressProfile) string {
	return formatMusicOverviewText(region, profile, nil, musicOverviewDefaultLimit)
}

func formatMusicRewardText(region string, profile musicRewardProfile, limit int) string {
	return formatMusicOverviewText(region, profile, nil, limit)
}

func musicResultDifficulty(result userMusicResult) string {
	if result.MusicDifficultyType != "" {
		return strings.ToLower(result.MusicDifficultyType)
	}
	return strings.ToLower(result.MusicDifficulty)
}

func musicResultCleared(result userMusicResult) bool {
	return result.PlayResult != "" && result.PlayResult != "not_clear"
}

func musicResultFullCombo(result userMusicResult) bool {
	return result.FullComboFlg || result.FullPerfectFlg || result.PlayResult == "full_combo" || result.PlayResult == "all_perfect"
}

func musicResultAllPerfect(result userMusicResult) bool {
	return result.FullPerfectFlg || result.PlayResult == "all_perfect"
}

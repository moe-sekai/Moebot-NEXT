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

const musicRewardDefaultLimit = 10

type musicRewardProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	Achievements []musicAchievement `json:"userMusicAchievements"`
}

type musicAchievement struct {
	MusicID            int `json:"musicId"`
	MusicAchievementID int `json:"musicAchievementId"`
}

type musicRewardRow struct {
	MusicID int
	Count   int
}

func musicRewardFields() []string {
	return []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserMusicAchievements}
}

func RegisterMusicReward(deps *Deps) {
	for _, cmd := range parserCommands(deps, "歌曲奖励") {
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
			var profile musicRewardProfile
			if err := runtime.Suite.GetUserData(user.GameID, setting.Mode, musicRewardFields(), &profile); err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取你的%sSuite抓包数据失败，发送 /抓包 获取帮助\n%s", runtime.Label, err.Error())))
				return
			}
			ctx.SendChain(message.Text(formatMusicRewardText(runtime.Region, profile, musicRewardDefaultLimit)))
			bot.RecordCommandRegion(deps.DB, "歌曲奖励", runtime.Region, ctx, start)
		})
	}
}

func formatMusicRewardText(region string, profile musicRewardProfile, limit int) string {
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
	counts := map[int]int{}
	for _, achievement := range profile.Achievements {
		if achievement.MusicID <= 0 {
			continue
		}
		counts[achievement.MusicID]++
	}
	rows := make([]musicRewardRow, 0, len(counts))
	total := 0
	for mid, count := range counts {
		rows = append(rows, musicRewardRow{MusicID: mid, Count: count})
		total += count
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Count == rows[j].Count {
			return rows[i].MusicID < rows[j].MusicID
		}
		return rows[i].Count > rows[j].Count
	})
	if limit <= 0 || limit > len(rows) {
		limit = len(rows)
	}
	lines := []string{
		fmt.Sprintf("%s 歌曲奖励", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("已达成奖励数: %d", total),
		fmt.Sprintf("涉及歌曲数: %d", len(rows)),
		fmt.Sprintf("更新时间: %s", suiteUpdateText(profile.UploadTime)),
		fmt.Sprintf("数据来源: %s", suiteSourceText(profile.BaseProfile)),
	}
	if len(rows) == 0 {
		lines = append(lines, "暂无歌曲奖励数据")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "---")
	for i := 0; i < limit; i++ {
		row := rows[i]
		lines = append(lines, fmt.Sprintf("%d. 歌曲 #%d: %d", i+1, row.MusicID, row.Count))
	}
	return strings.Join(lines, "\n")
}

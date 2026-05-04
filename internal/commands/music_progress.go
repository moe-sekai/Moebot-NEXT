package commands

import (
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

type musicProgressProfile struct {
	suite.BaseProfile
	UserGamedata     suite.UserGamedata `json:"userGamedata"`
	UserMusicResults []userMusicResult  `json:"userMusicResults"`
}

type userMusicResult struct {
	MusicID             int    `json:"musicId"`
	MusicDifficulty     string `json:"musicDifficulty"`
	MusicDifficultyType string `json:"musicDifficultyType"`
	PlayResult          string `json:"playResult"`
	FullComboFlg        bool   `json:"fullComboFlg"`
	FullPerfectFlg      bool   `json:"fullPerfectFlg"`
}

type musicProgressCount struct {
	Played     int
	Clear      int
	FullCombo  int
	AllPerfect int
}

func musicProgressFields() []string {
	return []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserMusicResults}
}

func RegisterMusicProgress(deps *Deps) {
	for _, cmd := range parserCommands(deps, "打歌进度") {
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
			var profile musicProgressProfile
			if err := runtime.Suite.GetUserData(user.GameID, setting.Mode, musicProgressFields(), &profile); err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取你的%sSuite抓包数据失败，发送 /抓包 获取帮助\n%s", runtime.Label, err.Error())))
				return
			}
			ctx.SendChain(message.Text(formatMusicProgressText(runtime.Region, profile)))
			bot.RecordCommandRegion(deps.DB, "打歌进度", runtime.Region, ctx, start)
		})
	}
}

func formatMusicProgressText(region string, profile musicProgressProfile) string {
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
	source := suiteSourceText(profile.BaseProfile)
	updateText := suiteUpdateText(profile.UploadTime)
	counts := map[string]*musicProgressCount{}
	for _, result := range profile.UserMusicResults {
		diff := musicResultDifficulty(result)
		if diff == "" {
			continue
		}
		count := counts[diff]
		if count == nil {
			count = &musicProgressCount{}
			counts[diff] = count
		}
		count.Played++
		if musicResultCleared(result) {
			count.Clear++
		}
		if musicResultFullCombo(result) {
			count.FullCombo++
		}
		if musicResultAllPerfect(result) {
			count.AllPerfect++
		}
	}
	lines := []string{
		fmt.Sprintf("%s 打歌进度", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("更新时间: %s", updateText),
		fmt.Sprintf("数据来源: %s", source),
	}
	if len(counts) == 0 {
		lines = append(lines, "暂无打歌数据")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "---")
	for _, diff := range []string{"easy", "normal", "hard", "expert", "master", "append"} {
		count := counts[diff]
		if count == nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: 游玩 %d | Clear %d | FC %d | AP %d", strings.ToUpper(diff), count.Played, count.Clear, count.FullCombo, count.AllPerfect))
	}
	return strings.Join(lines, "\n")
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

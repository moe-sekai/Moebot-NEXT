package commands

import (
	"strings"
	"testing"

	"moebot-next/internal/commandparser"
	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/suite"
)

func TestSuiteStatusTextShowsUploadSourceAndInterface(t *testing.T) {
	text := formatSuiteStatusText(config.RegionCN, suite.Status{
		UserID:     "123456789012345678",
		Name:       "测试玩家",
		Source:     "moesekai",
		UploadTime: 1700000000000,
	})
	for _, want := range []string{"CN", "123456789012345678", "测试玩家", "moesekai", "Suite数据", "更新时间", "Haruki 公开 API"} {
		if !strings.Contains(text, want) {
			t.Fatalf("status text missing %q in:\n%s", want, text)
		}
	}
}

func TestParserCommandsIncludesSuiteStatusAliases(t *testing.T) {
	cmds := parserCommands(&Deps{Definitions: commandparser.BaseDefinitions()}, "抓包状态")
	seen := map[string]bool{}
	for _, cmd := range cmds {
		seen[cmd.Name] = true
	}
	for _, want := range []string{"抓包状态", "抓包数据", "抓包信息", "suite", "cn抓包状态", "cn抓包数据", "cn抓包信息", "cnsuite"} {
		if !seen[want] {
			t.Fatalf("parserCommands missing %q in %#v", want, cmds)
		}
	}
}

func TestParserCommandsIncludesSuiteShowAlias(t *testing.T) {
	cmds := parserCommands(&Deps{Definitions: commandparser.BaseDefinitions()}, "展示抓包")
	seen := map[string]bool{}
	for _, cmd := range cmds {
		seen[cmd.Name] = true
	}
	for _, want := range []string{"展示抓包", "显示抓包", "cn展示抓包", "cn显示抓包"} {
		if !seen[want] {
			t.Fatalf("parserCommands missing %q in %#v", want, cmds)
		}
	}
}

func TestSuiteStatusTextCanHideUID(t *testing.T) {
	text := formatSuiteStatusText(config.RegionJP, suite.Status{
		UserID:     "123456789012345678",
		Name:       "测试玩家",
		Source:     "local",
		UploadTime: 1700000000000,
	}, hideUIDExceptLast(6))
	if strings.Contains(text, "123456789012345678") {
		t.Fatalf("uid should be hidden in:\n%s", text)
	}
	if !strings.Contains(text, "************345678") {
		t.Fatalf("hidden suffix missing in:\n%s", text)
	}
}

func TestGachaHistoryFieldsUsesOnlyRequiredFields(t *testing.T) {
	fields := gachaHistoryFields()
	want := []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserGachas}
	if len(fields) != len(want) {
		t.Fatalf("fields len = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("fields[%d] = %q, want %q; all fields: %#v", i, fields[i], want[i], fields)
		}
	}
}

func TestFormatGachaHistoryTextSortsAndNamesGachas(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{Gachas: []masterdata.GachaInfo{
		{ID: 100, Name: "测试卡池A"},
		{ID: 200, Name: "测试卡池B"},
	}})
	profile := gachaHistoryProfile{
		BaseProfile: suite.BaseProfile{UploadTime: 1700000000000, Source: "moesekai"},
		UserGamedata: suite.UserGamedata{
			Name: "测试玩家",
		},
		UserGachas: []userGachaRecord{
			{GachaID: 100, Count: 20},
			{GachaID: 200, Count: 50},
			{GachaID: 999, Count: 30},
		},
	}

	text := formatGachaHistoryText(config.RegionCN, profile, store, 10)
	for _, want := range []string{"CN 抽卡记录", "测试玩家", "总抽数: 100", "测试卡池B", "50抽", "未知卡池 #999", "30抽", "测试卡池A", "20抽", "更新时间", "数据来源: moesekai"} {
		if !strings.Contains(text, want) {
			t.Fatalf("gacha history text missing %q in:\n%s", want, text)
		}
	}
	if strings.Index(text, "测试卡池B") > strings.Index(text, "未知卡池 #999") || strings.Index(text, "未知卡池 #999") > strings.Index(text, "测试卡池A") {
		t.Fatalf("records should be sorted by count desc:\n%s", text)
	}
}

func TestFormatCardBoxTextAppliesSortOptions(t *testing.T) {
	cards := []masterdata.CardInfo{
		{ID: 3, CharacterID: 2, Prefix: "低专精", ReleaseAt: 3000},
		{ID: 1, CharacterID: 1, Prefix: "高专精", ReleaseAt: 1000},
		{ID: 2, CharacterID: 1, Prefix: "中专精", ReleaseAt: 2000},
	}
	profile := cardBoxProfile{
		BaseProfile:  suite.BaseProfile{UploadTime: 1700000000000, Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
	}
	owned := map[int]suite.UserCard{
		1: {CardID: 1, Level: 50, MasterRank: 5, SkillLevel: 1, CreatedAt: 1000},
		2: {CardID: 2, Level: 40, MasterRank: 2, SkillLevel: 4, CreatedAt: 3000},
		3: {CardID: 3, Level: 30, MasterRank: 1, SkillLevel: 2, CreatedAt: 2000},
	}

	text := formatCardBoxText(config.RegionCN, profile, cards, owned, cardBoxQueryOptions{SortBy: "mr"})
	if strings.Index(text, "#1") > strings.Index(text, "#2") || strings.Index(text, "#2") > strings.Index(text, "#3") {
		t.Fatalf("mr sort should order by master rank desc, got:\n%s", text)
	}

	text = formatCardBoxText(config.RegionCN, profile, cards, owned, cardBoxQueryOptions{SortBy: "time"})
	if strings.Index(text, "#2") > strings.Index(text, "#3") || strings.Index(text, "#3") > strings.Index(text, "#1") {
		t.Fatalf("time sort should order by created time desc, got:\n%s", text)
	}

	text = formatCardBoxText(config.RegionCN, profile, cards, owned, cardBoxQueryOptions{SortBy: "sl"})
	if strings.Index(text, "#2") > strings.Index(text, "#3") || strings.Index(text, "#3") > strings.Index(text, "#1") {
		t.Fatalf("sl sort should order by skill level desc, got:\n%s", text)
	}
}

func TestParseCardBoxOptionsIDSortsByID(t *testing.T) {
	options := parseCardBoxOptions("id")
	if !options.ShowID || options.SortBy != "id" {
		t.Fatalf("id option = %#v, want ShowID and id sort", options)
	}
}

func TestFormatGachaHistoryTextHandlesEmptyRecords(t *testing.T) {
	text := formatGachaHistoryText(config.RegionJP, gachaHistoryProfile{
		BaseProfile:  suite.BaseProfile{Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
	}, masterdata.NewStore(), 10)
	if !strings.Contains(text, "暂无抽卡记录") {
		t.Fatalf("empty history should be explained, got:\n%s", text)
	}
}

func TestBondFieldsUsesUserBonds(t *testing.T) {
	fields := bondFields()
	want := suite.Fields(suite.FieldUserBonds)
	if len(fields) != len(want) {
		t.Fatalf("fields len = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("fields[%d] = %q, want %q; all fields: %#v", i, fields[i], want[i], fields)
		}
	}
}

func TestFormatBondTextSortsTopBonds(t *testing.T) {
	profile := bondProfile{
		BaseProfile:  suite.BaseProfile{UploadTime: 1700000000000, Source: "moesekai"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
		UserBonds: []userBond{
			{CharacterID1: 1, CharacterID2: 21, Rank: 12, Exp: 100},
			{BondsGroupID: 20222, Rank: 20, Exp: 10},
			{BondsGroupID: 30323, Rank: 20, Exp: 50},
		},
	}

	text := formatBondText(config.RegionCN, profile, 2)
	for _, want := range []string{"CN 羁绊", "测试玩家", "穗波 × 镜音连", "Lv.20", "EXP 50", "咲希 × 镜音铃", "数据来源: moesekai"} {
		if !strings.Contains(text, want) {
			t.Fatalf("bond text missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "一歌 × 初音未来") {
		t.Fatalf("limit should hide third bond:\n%s", text)
	}
	if strings.Index(text, "穗波 × 镜音连") > strings.Index(text, "咲希 × 镜音铃") {
		t.Fatalf("bonds should be sorted by rank and exp desc:\n%s", text)
	}
}

func TestFormatBondTextHandlesEmptyBonds(t *testing.T) {
	text := formatBondText(config.RegionJP, bondProfile{
		BaseProfile:  suite.BaseProfile{Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
	}, 10)
	if !strings.Contains(text, "暂无羁绊数据") {
		t.Fatalf("empty bonds should be explained, got:\n%s", text)
	}
}

func TestMusicProgressFieldsUsesMusicResults(t *testing.T) {
	fields := musicProgressFields()
	want := []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserMusicResults}
	if len(fields) != len(want) {
		t.Fatalf("fields len = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("fields[%d] = %q, want %q; all fields: %#v", i, fields[i], want[i], fields)
		}
	}
}

func TestFormatMusicProgressTextSummarizesByDifficulty(t *testing.T) {
	profile := musicProgressProfile{
		BaseProfile:  suite.BaseProfile{UploadTime: 1700000000000, Source: "local", LocalSource: "manual"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
		UserMusicResults: []userMusicResult{
			{MusicID: 1, MusicDifficultyType: "expert", PlayResult: "clear"},
			{MusicID: 2, MusicDifficultyType: "expert", PlayResult: "full_combo", FullComboFlg: true},
			{MusicID: 3, MusicDifficulty: "master", PlayResult: "all_perfect", FullPerfectFlg: true},
			{MusicID: 4, MusicDifficultyType: "master", PlayResult: "not_clear"},
		},
	}

	text := formatMusicProgressText(config.RegionCN, profile)
	for _, want := range []string{"CN 打歌进度", "测试玩家", "EXPERT", "游玩 2", "Clear 2", "FC 1", "AP 0", "MASTER", "游玩 2", "Clear 1", "FC 1", "AP 1", "数据来源: local(manual)"} {
		if !strings.Contains(text, want) {
			t.Fatalf("music progress missing %q in:\n%s", want, text)
		}
	}
}

func TestFormatMusicProgressTextHandlesEmptyResults(t *testing.T) {
	text := formatMusicProgressText(config.RegionJP, musicProgressProfile{
		BaseProfile:  suite.BaseProfile{Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
	})
	if !strings.Contains(text, "暂无打歌数据") {
		t.Fatalf("empty music progress should be explained, got:\n%s", text)
	}
}

func TestMaterialFieldsUsesMaterials(t *testing.T) {
	fields := materialFields()
	want := []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserMaterials}
	if len(fields) != len(want) {
		t.Fatalf("fields len = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("fields[%d] = %q, want %q; all fields: %#v", i, fields[i], want[i], fields)
		}
	}
}

func TestFormatMaterialTextSortsMaterialsAndShowsCoin(t *testing.T) {
	profile := materialProfile{
		BaseProfile:  suite.BaseProfile{UploadTime: 1700000000000, Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家", Coin: 123456},
		UserMaterials: []userMaterial{
			{MaterialID: 1, Quantity: 100},
			{MaterialID: 2, Quantity: 300},
			{MaterialID: 3, Quantity: 200},
		},
	}
	text := formatMaterialText(config.RegionCN, profile, 2)
	for _, want := range []string{"CN 材料信息", "测试玩家", "金币: 123456", "材料 #2: 300", "材料 #3: 200", "数据来源: local"} {
		if !strings.Contains(text, want) {
			t.Fatalf("material text missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "材料 #1") {
		t.Fatalf("limit should hide third material:\n%s", text)
	}
}

func TestFormatMaterialTextHandlesEmptyMaterials(t *testing.T) {
	text := formatMaterialText(config.RegionJP, materialProfile{
		BaseProfile:  suite.BaseProfile{Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
	}, 10)
	if !strings.Contains(text, "暂无材料数据") {
		t.Fatalf("empty materials should be explained, got:\n%s", text)
	}
}

func TestChallengeFieldsUsesChallengeData(t *testing.T) {
	fields := challengeFields()
	want := suite.Fields(suite.FieldUserChallengeLiveSoloResults, suite.FieldUserChallengeLiveSoloStages, suite.FieldUserChallengeLiveSoloHighScoreRewards)
	if len(fields) != len(want) {
		t.Fatalf("fields len = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("fields[%d] = %q, want %q; all fields: %#v", i, fields[i], want[i], fields)
		}
	}
}

func TestFormatChallengeTextSummarizesCharacters(t *testing.T) {
	profile := challengeProfile{
		BaseProfile:  suite.BaseProfile{UploadTime: 1700000000000, Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
		Results: []challengeResult{
			{CharacterID: 1, HighScore: 1000000},
			{CharacterID: 20, HighScore: 2500000},
		},
		Stages: []challengeStage{
			{CharacterID: 1, Rank: 2},
			{CharacterID: 1, Rank: 3},
			{CharacterID: 20, Rank: 7},
		},
		Rewards: []challengeReward{
			{CharacterID: 1, RewardID: 1},
			{CharacterID: 20, RewardID: 2},
			{CharacterID: 20, RewardID: 3},
		},
	}
	text := formatChallengeText(config.RegionCN, profile, 2)
	for _, want := range []string{"CN 挑战信息", "测试玩家", "瑞希", "2500000", "Lv.7", "奖励 2", "一歌", "1000000", "Lv.3", "奖励 1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("challenge text missing %q in:\n%s", want, text)
		}
	}
	if strings.Index(text, "瑞希") > strings.Index(text, "一歌") {
		t.Fatalf("challenge rows should sort by high score desc:\n%s", text)
	}
}

func TestFormatChallengeTextHandlesEmptyData(t *testing.T) {
	text := formatChallengeText(config.RegionJP, challengeProfile{
		BaseProfile:  suite.BaseProfile{Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
	}, 10)
	if !strings.Contains(text, "暂无挑战数据") {
		t.Fatalf("empty challenge should be explained, got:\n%s", text)
	}
}

func TestEventRecordFieldsUsesEventData(t *testing.T) {
	fields := eventRecordFields()
	want := suite.Fields(suite.FieldUserEvents, suite.FieldUserWorldBlooms)
	if len(fields) != len(want) {
		t.Fatalf("fields len = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("fields[%d] = %q, want %q; all fields: %#v", i, fields[i], want[i], fields)
		}
	}
}

func TestFormatEventRecordTextSortsEventsAndWorldBlooms(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{Events: []masterdata.EventInfo{{ID: 100, Name: "活动A"}, {ID: 200, Name: "活动B"}}})
	profile := eventRecordProfile{
		BaseProfile:  suite.BaseProfile{UploadTime: 1700000000000, Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
		UserEvents: []userEventRecord{
			{EventID: 100, EventPoint: 1000},
			{EventID: 200, EventPoint: 3000},
		},
		UserWorldBlooms: []userWorldBloomRecord{
			{EventID: 300, GameCharacterID: 20, WorldBloomChapterPoint: 5000},
		},
	}
	text := formatEventRecordText(config.RegionCN, profile, store, 5)
	for _, want := range []string{"CN 活动记录", "测试玩家", "活动B", "3000pt", "活动A", "1000pt", "WL章节", "瑞希", "5000pt"} {
		if !strings.Contains(text, want) {
			t.Fatalf("event record missing %q in:\n%s", want, text)
		}
	}
	if strings.Index(text, "活动B") > strings.Index(text, "活动A") {
		t.Fatalf("events should sort by point desc:\n%s", text)
	}
}

func TestFormatEventRecordTextHandlesEmptyData(t *testing.T) {
	text := formatEventRecordText(config.RegionJP, eventRecordProfile{
		BaseProfile:  suite.BaseProfile{Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
	}, masterdata.NewStore(), 10)
	if !strings.Contains(text, "暂无活动记录") {
		t.Fatalf("empty event record should be explained, got:\n%s", text)
	}
}

func TestLeaderCountFieldsUsesCharacterMissions(t *testing.T) {
	fields := leaderCountFields()
	want := suite.Fields(suite.FieldUserCharacterMissionV2s, suite.FieldUserCharacterMissionV2Statuses)
	if len(fields) != len(want) {
		t.Fatalf("fields len = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("fields[%d] = %q, want %q; all fields: %#v", i, fields[i], want[i], fields)
		}
	}
}

func TestFormatLeaderCountTextSortsPlayLive(t *testing.T) {
	profile := leaderCountProfile{
		BaseProfile:  suite.BaseProfile{UploadTime: 1700000000000, Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
		Missions: []characterMissionV2{
			{CharacterID: 1, CharacterMissionType: "play_live", Progress: 100},
			{CharacterID: 20, CharacterMissionType: "play_live", Progress: 300},
			{CharacterID: 20, CharacterMissionType: "play_live_ex", Progress: 50},
		},
	}
	text := formatLeaderCountText(config.RegionCN, profile, 2)
	for _, want := range []string{"CN 队长次数", "瑞希", "300", "EX 50", "一歌", "100"} {
		if !strings.Contains(text, want) {
			t.Fatalf("leader count missing %q in:\n%s", want, text)
		}
	}
	if strings.Index(text, "瑞希") > strings.Index(text, "一歌") {
		t.Fatalf("leader counts should sort by play count desc:\n%s", text)
	}
}

func TestFormatLeaderCountTextHandlesEmptyData(t *testing.T) {
	text := formatLeaderCountText(config.RegionJP, leaderCountProfile{
		BaseProfile:  suite.BaseProfile{Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
	}, 10)
	if !strings.Contains(text, "暂无队长次数数据") {
		t.Fatalf("empty leader count should be explained, got:\n%s", text)
	}
}

func TestMusicRewardFieldsUsesAchievements(t *testing.T) {
	fields := musicRewardFields()
	want := []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserMusicAchievements}
	if len(fields) != len(want) {
		t.Fatalf("fields len = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("fields[%d] = %q, want %q; all fields: %#v", i, fields[i], want[i], fields)
		}
	}
}

func TestFormatMusicRewardTextSummarizesAchievements(t *testing.T) {
	profile := musicRewardProfile{
		BaseProfile:  suite.BaseProfile{UploadTime: 1700000000000, Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
		Achievements: []musicAchievement{
			{MusicID: 1, MusicAchievementID: 1},
			{MusicID: 1, MusicAchievementID: 2},
			{MusicID: 2, MusicAchievementID: 3},
		},
	}
	text := formatMusicRewardText(config.RegionCN, profile, 5)
	for _, want := range []string{"CN 歌曲奖励", "测试玩家", "已达成奖励数: 3", "涉及歌曲数: 2", "歌曲 #1: 2", "歌曲 #2: 1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("music reward missing %q in:\n%s", want, text)
		}
	}
}

func TestFormatMusicRewardTextHandlesEmptyData(t *testing.T) {
	text := formatMusicRewardText(config.RegionJP, musicRewardProfile{
		BaseProfile:  suite.BaseProfile{Source: "local"},
		UserGamedata: suite.UserGamedata{Name: "测试玩家"},
	}, 10)
	if !strings.Contains(text, "暂无歌曲奖励数据") {
		t.Fatalf("empty music reward should be explained, got:\n%s", text)
	}
}

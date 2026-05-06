package commandparser

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/b30"
	"moebot-next/internal/cardquery"
	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/renderer"
	"moebot-next/internal/servers"
	"moebot-next/internal/suite"
)

const (
	debugDefaultLimit            = 10
	debugLeaderCountProgressMax  = 50000
	debugChallengeProgressMax    = 3000000
	debugChallengeBoxPurpose     = "challenge_live_high_score"
	debugMusicRewardRankRewardID = 4
)

type suiteDebugMusicAchievementReward struct {
	Coin  int
	Jewel int
	Shard int
}

var suiteDebugMusicRankRewards = map[int]suiteDebugMusicAchievementReward{
	1: {Jewel: 10},
	2: {Jewel: 20},
	3: {Jewel: 30},
	4: {Jewel: 50},
}

var suiteDebugMusicComboRewards = map[string]map[int]suiteDebugMusicAchievementReward{
	"easy": {
		5: {Coin: 500},
		6: {Coin: 1000},
		7: {Coin: 2000},
		8: {Coin: 5000},
	},
	"normal": {
		9:  {Coin: 1000},
		10: {Coin: 2000},
		11: {Coin: 4000},
		12: {Coin: 10000},
	},
	"hard": {
		13: {Coin: 1500},
		14: {Coin: 3000},
		15: {Coin: 6000},
		16: {Jewel: 50},
	},
	"expert": {
		17: {Coin: 2000},
		18: {Coin: 4000},
		19: {Jewel: 20},
		20: {Jewel: 50},
	},
	"master": {
		21: {Coin: 3000},
		22: {Coin: 6000},
		23: {Jewel: 20},
		24: {Jewel: 50},
	},
	"append": {
		25: {Coin: 3000},
		26: {Coin: 6000},
		27: {Shard: 5},
		28: {Shard: 10},
	},
}

type debugBindingResult struct {
	Region   string
	Results  []EntityResult
	Selected *EntityResult
	Message  string
	Warnings []string
	Used     bool
}

func (s *Service) buildDebugBindingPayload(def Definition, parsedRegion string, argument string, binding DebugBinding) (debugBindingResult, bool) {
	if !def.RequiresBinding {
		return debugBindingResult{}, false
	}

	region := config.NormalizeRegion(binding.Region)
	if region == "" {
		region = config.NormalizeRegion(parsedRegion)
	}
	if region == "" {
		region = s.defaultRegion()
	}
	if !config.IsValidRegion(region) {
		region = s.defaultRegion()
	}

	gameID := strings.TrimSpace(binding.GameID)
	if gameID == "" {
		return debugBindingResult{
			Region:   region,
			Message:  "该功能需要账号绑定上下文；请输入临时区服与游戏 UID 后可调试真实预览。",
			Warnings: []string{"未填写临时游戏 UID，将使用静态预览兜底。"},
		}, true
	}

	switch def.BindingKind {
	case "profile":
		return s.buildProfileDebugPayload(def, region, gameID), true
	case "suite":
		return s.buildSuiteDebugPayload(def, region, gameID, argument), true
	default:
		return debugBindingResult{
			Region:   region,
			Message:  "该绑定类功能暂未接入调试数据构建，将使用静态预览兜底。",
			Warnings: []string{"未知绑定调试类型。"},
		}, true
	}
}

func (s *Service) buildProfileDebugPayload(def Definition, region string, gameID string) debugBindingResult {
	runtime := s.runtimeForRegion(region)
	if runtime == nil || !runtime.Enabled {
		return debugBindingResult{Region: region, Message: runtimeUnavailableMessage(runtime), Warnings: []string{"服务器不可用，将使用静态预览兜底。"}}
	}
	if runtime.Sekai == nil || !runtime.Sekai.Enabled() {
		return debugBindingResult{Region: runtime.Region, Message: fmt.Sprintf("%s Sekai API 未配置，无法拉取个人信息。", runtime.Label), Warnings: []string{"Sekai API 不可用，将使用静态预览兜底。"}}
	}
	profile, err := runtime.Sekai.GetProfile(gameID)
	if err != nil {
		return debugBindingResult{Region: runtime.Region, Message: "个人信息获取失败：" + err.Error(), Warnings: []string{"真实资料获取失败，将使用静态预览兜底。"}}
	}
	payload := renderer.BuildProfileCardPayloadWithAssets(runtime.Store, *profile, runtime.Assets)
	selected := EntityResult{ID: 0, Title: firstNonEmptyDebug(profile.Name, "玩家资料"), Subtitle: fmt.Sprintf("%s · UID %s", runtime.Label, profile.UserID), Type: "profile_card", Payload: payload}
	return debugBindingResult{
		Region:   runtime.Region,
		Results:  []EntityResult{{ID: 0, Title: selected.Title, Subtitle: selected.Subtitle, Type: selected.Type}},
		Selected: &selected,
		Message:  fmt.Sprintf("已使用临时 %s UID %s 拉取个人信息。", runtime.Label, gameID),
		Used:     true,
	}
}

func (s *Service) buildSuiteDebugPayload(def Definition, region string, gameID string, argument string) debugBindingResult {
	runtime := s.runtimeForRegion(region)
	if runtime == nil || !runtime.Enabled {
		return debugBindingResult{Region: region, Message: runtimeUnavailableMessage(runtime), Warnings: []string{"服务器不可用，将使用静态预览兜底。"}}
	}
	if runtime.Suite == nil || !runtime.Suite.Enabled() {
		return debugBindingResult{Region: runtime.Region, Message: fmt.Sprintf("%s Haruki 公开 API 未配置，无法拉取 Suite 数据。", runtime.Label), Warnings: []string{"Haruki 公开 API 不可用，将使用静态预览兜底。"}}
	}

	payload, selected, rows, err := s.buildSuiteDebugPayloadForDefinition(def, runtime, gameID, argument)
	if err != nil {
		return debugBindingResult{Region: runtime.Region, Message: "Suite 调试数据获取失败：" + err.Error(), Warnings: []string{"真实 Suite 公开数据获取失败，将使用静态预览兜底。"}}
	}
	if selected == nil {
		selected = &EntityResult{ID: 0, Title: def.Name, Subtitle: fmt.Sprintf("%s · UID %s · Haruki 公开 API", runtime.Label, gameID), Type: def.Template}
	}
	selected.Payload = payload
	if len(rows) == 0 {
		rows = []EntityResult{{ID: selected.ID, Title: selected.Title, Subtitle: selected.Subtitle, Type: selected.Type}}
	}
	return debugBindingResult{
		Region:   runtime.Region,
		Results:  rows,
		Selected: selected,
		Message:  fmt.Sprintf("已使用临时 %s UID %s 通过 Haruki 公开 API 拉取 Suite 调试数据。", runtime.Label, gameID),
		Used:     true,
	}
}

func (s *Service) buildSuiteDebugPayloadForDefinition(def Definition, runtime *servers.Runtime, gameID string, argument string) (any, *EntityResult, []EntityResult, error) {
	definitionID := def.ID
	if definitionID == "music-reward" {
		definitionID = "music-progress"
	}
	switch definitionID {
	case "suite-status":
		var profile suiteDebugCommonProfile
		if err := runtime.Suite.GetUserData(gameID, "", suite.Fields(), &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "Suite数据", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), renderer.SuiteStatPayload{Label: "数据来源", Value: suiteDebugSourceText(profile.BaseProfile)})
		payload.Sections = []renderer.SuiteSectionPayload{{Title: "Suite 状态", Rows: []renderer.SuiteSectionRowPayload{
			{Label: "玩家", Value: payload.Profile.Name},
			{Label: "用户ID", Value: payload.Profile.UserID},
			{Label: "更新时间", Value: suiteDebugUpdateText(profile.UploadTime)},
			{Label: "数据来源", Value: suiteDebugSourceText(profile.BaseProfile)},
			{Label: "接口", Value: "Haruki 公开 API"},
		}}}
		return payload, suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_panel"), suiteDebugRowsFromSections(payload.Sections), nil
	case "bond-list":
		var profile suiteDebugBondProfile
		if err := runtime.Suite.GetUserData(gameID, "", suite.Fields(suite.FieldUserBonds), &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "羁绊查询", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		sectionRows, stats := suiteDebugRowsFromBonds(profile, debugDefaultLimit)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), stats...)
		payload.Sections = []renderer.SuiteSectionPayload{{Title: "羁绊 TOP", Kind: "bond_list", Note: "角色头像来自本地 assets/characters。", Rows: sectionRows}}
		return payload, suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_panel"), suiteDebugRowsFromSections(payload.Sections), nil
	case "music-progress":
		var profile suiteDebugMusicProgressProfile
		if err := runtime.Suite.GetUserData(gameID, "", []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserMusicResults, suite.FieldUserMusicAchievements}, &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "打歌进度 / 歌曲奖励", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		sections, stats := suiteDebugSectionsFromMusicOverview(profile, runtime.Store, debugDefaultLimit)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), stats...)
		payload.Sections = sections
		return payload, suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_panel"), suiteDebugRowsFromSections(payload.Sections), nil
	case "best30":
		var profile suiteDebugBest30Profile
		if err := runtime.Suite.GetUserData(gameID, "", suite.Fields(suite.FieldUserMusicResults, suite.FieldUserMusics), &profile); err != nil {
			return nil, nil, nil, err
		}
		client := s.best30Client()
		table, err := client.Get(context.Background())
		if err != nil {
			return nil, nil, nil, err
		}
		results := b30.MergeLegacyResults(profile.UserMusicResults, profile.UserMusics)
		result := b30.Calculate(results, table, renderer.Best30MusicMetaResolver(runtime.Store, runtime.Assets))
		payload := renderer.BuildBest30Payload(suiteDebugPanelTitle(runtime, "Best30"), runtime.Region, profile.BaseProfile, profile.UserGamedata, result, runtime.Store, runtime.Assets, client.URL())
		selected := suiteDebugSelected(def, runtime, profile.UserGamedata, "best30")
		rows := suiteDebugRowsFromBest30(result)
		if len(rows) == 0 {
			rows = []EntityResult{{ID: 0, Title: "Best30", Subtitle: "暂无可计入 AP/FC 谱面", Type: "best30"}}
		}
		return payload, selected, rows, nil
	case "challenge-info":
		var profile suiteDebugChallengeProfile
		if err := runtime.Suite.GetUserData(gameID, "", suite.Fields(suite.FieldUserChallengeLiveSoloResults, suite.FieldUserChallengeLiveSoloStages, suite.FieldUserChallengeLiveSoloHighScoreRewards), &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "挑战信息", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		sectionRows, stats, sectionExtra := suiteDebugRowsFromChallenge(profile, runtime.Store, 26)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), stats...)
		payload.Sections = []renderer.SuiteSectionPayload{{Title: "每日挑战 Live", Kind: "challenge_info", Note: "参考 lunabot：按角色统计挑战等级、最高分，以及未领取高分奖励中的水晶/碎片。", Rows: sectionRows, Extra: sectionExtra}}
		return payload, suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_panel"), suiteDebugRowsFromSections(payload.Sections), nil
	case "event-record":
		var profile suiteDebugEventRecordProfile
		if err := runtime.Suite.GetUserData(gameID, "", suite.Fields(suite.FieldUserEvents, suite.FieldUserWorldBlooms), &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "活动记录", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		sections, stats := suiteDebugRowsFromEventRecord(profile, runtime.Store, runtime.Assets, debugDefaultLimit)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), stats...)
		payload.Sections = sections
		return payload, suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_panel"), suiteDebugRowsFromSections(payload.Sections), nil
	case "leader-count":
		var profile suiteDebugLeaderCountProfile
		if err := runtime.Suite.GetUserData(gameID, "", suite.Fields(suite.FieldUserCharacterMissionV2s, suite.FieldUserCharacterMissionV2Statuses), &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "队长次数", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		sectionRows, stats, sectionExtra := suiteDebugRowsFromLeaderCount(profile, runtime.Store, 26)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), stats...)
		payload.Sections = []renderer.SuiteSectionPayload{{Title: "角色队长次数", Kind: "leader_count", Note: "参考 lunabot：普通档位读取 parameterGroupId=1；EX 等级/次数读取 parameterGroupId=101 并累计已完成轮次。", Rows: sectionRows, Extra: sectionExtra}}
		return payload, suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_panel"), suiteDebugRowsFromSections(payload.Sections), nil
	case "suite-card-box":
		var profile suiteDebugCardBoxProfile
		if err := runtime.Suite.GetUserData(gameID, "", suite.Fields(), &profile); err != nil {
			return nil, nil, nil, err
		}
		options := suiteDebugParseCardBoxOptions(argument)
		cards, msg := suiteDebugCardBoxCards(runtime.Store, options)
		if msg != "" {
			return nil, nil, nil, fmt.Errorf("%s", msg)
		}
		owned := renderer.SuiteUserCardMap(profile.UserCards)
		payload := renderer.BuildSuiteCardBoxPayload(
			suiteDebugPanelTitle(runtime, "卡牌一览"),
			suiteDebugCardBoxSubtitle(options, len(cards), len(owned)),
			runtime.Region,
			"",
			profile.BaseProfile,
			profile.UserGamedata,
			cards,
			owned,
			suiteDebugCardBoxDeckSet(profile),
			runtime.Store,
			runtime.Assets,
			renderer.SuiteCardBoxOptions{ShowID: options.ShowID, OwnedOnly: options.OwnedOnly, UseBeforeTraining: options.UseBeforeTraining, ShowCreatedAt: options.ShowCreatedAt, SortBy: options.SortBy},
		)
		selected := suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_card_box")
		rows := []EntityResult{{ID: payload.OwnedTotal, Title: "卡牌一览", Subtitle: payload.Subtitle, Type: "suite_card_box"}}
		return payload, selected, rows, nil
	default:
		return nil, nil, nil, fmt.Errorf("%s 暂未支持临时绑定调试", def.Name)
	}
}

type suiteDebugCommonProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	UserDecks    []suite.UserDeck   `json:"userDecks"`
	UserCards    []suite.UserCard   `json:"userCards"`
}

type suiteDebugGachaHistoryProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata      `json:"userGamedata"`
	UserGachas   []suiteDebugGachaRecord `json:"userGachas"`
}

type suiteDebugGachaRecord struct {
	GachaID int `json:"gachaId"`
	Count   int `json:"count"`
}

type suiteDebugBondProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	UserDecks    []suite.UserDeck   `json:"userDecks"`
	UserCards    []suite.UserCard   `json:"userCards"`
	UserBonds    []suiteDebugBond   `json:"userBonds"`
}

type suiteDebugBond struct {
	BondsGroupID     int `json:"bondsGroupId"`
	CharacterID1     int `json:"characterId1"`
	CharacterID2     int `json:"characterId2"`
	GameCharacterID1 int `json:"gameCharacterId1"`
	GameCharacterID2 int `json:"gameCharacterId2"`
	Rank             int `json:"rank"`
	Exp              int `json:"exp"`
}

type suiteDebugMusicProgressProfile struct {
	suite.BaseProfile
	UserGamedata     suite.UserGamedata           `json:"userGamedata"`
	UserDecks        []suite.UserDeck             `json:"userDecks"`
	UserCards        []suite.UserCard             `json:"userCards"`
	UserMusicResults []suiteDebugMusicResult      `json:"userMusicResults"`
	Achievements     []suiteDebugMusicAchievement `json:"userMusicAchievements"`
}

type suiteDebugBest30Profile struct {
	suite.BaseProfile
	UserGamedata     suite.UserGamedata    `json:"userGamedata"`
	UserDecks        []suite.UserDeck      `json:"userDecks"`
	UserCards        []suite.UserCard      `json:"userCards"`
	UserMusicResults []b30.UserMusicResult `json:"userMusicResults"`
	UserMusics       []b30.LegacyUserMusic `json:"userMusics"`
}

type suiteDebugMusicResult struct {
	MusicID             int    `json:"musicId"`
	MusicDifficulty     string `json:"musicDifficulty"`
	MusicDifficultyType string `json:"musicDifficultyType"`
	PlayResult          string `json:"playResult"`
	FullComboFlg        bool   `json:"fullComboFlg"`
	FullPerfectFlg      bool   `json:"fullPerfectFlg"`
}

type suiteDebugMusicProgressCount struct {
	Total      int
	Played     int
	Clear      int
	FullCombo  int
	AllPerfect int
}

type suiteDebugMaterialProfile struct {
	suite.BaseProfile
	UserGamedata  suite.UserGamedata   `json:"userGamedata"`
	UserDecks     []suite.UserDeck     `json:"userDecks"`
	UserCards     []suite.UserCard     `json:"userCards"`
	UserMaterials []suiteDebugMaterial `json:"userMaterials"`
}

type suiteDebugMaterial struct {
	MaterialID int   `json:"materialId"`
	Quantity   int64 `json:"quantity"`
}

type suiteDebugChallengeProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata          `json:"userGamedata"`
	UserDecks    []suite.UserDeck            `json:"userDecks"`
	UserCards    []suite.UserCard            `json:"userCards"`
	Results      []suiteDebugChallengeResult `json:"userChallengeLiveSoloResults"`
	Stages       []suiteDebugChallengeStage  `json:"userChallengeLiveSoloStages"`
	Rewards      []suiteDebugChallengeReward `json:"userChallengeLiveSoloHighScoreRewards"`
}

type suiteDebugChallengeResult struct {
	CharacterID int `json:"characterId"`
	HighScore   int `json:"highScore"`
}

type suiteDebugChallengeStage struct {
	CharacterID int `json:"characterId"`
	Rank        int `json:"rank"`
}

type suiteDebugChallengeReward struct {
	CharacterID                        int `json:"characterId"`
	GameCharacterID                    int `json:"gameCharacterId"`
	RewardID                           int `json:"challengeLiveHighScoreRewardId"`
	ChallengeLiveSoloHighScoreRewardID int `json:"challengeLiveSoloHighScoreRewardId"`
	RewardIDAlias                      int `json:"rewardId"`
}

type suiteDebugChallengeRow struct {
	CharacterID    int
	HighScore      int
	Rank           int
	RewardCount    int
	RemainJewel    int
	RemainFragment int
}

type suiteDebugEventRecordProfile struct {
	suite.BaseProfile
	UserGamedata    suite.UserGamedata           `json:"userGamedata"`
	UserDecks       []suite.UserDeck             `json:"userDecks"`
	UserCards       []suite.UserCard             `json:"userCards"`
	UserEvents      []suiteDebugEventRecord      `json:"userEvents"`
	UserWorldBlooms []suiteDebugWorldBloomRecord `json:"userWorldBlooms"`
}

type suiteDebugEventRecord struct {
	EventID    int `json:"eventId"`
	EventPoint int `json:"eventPoint"`
	Rank       int `json:"rank"`
}

type suiteDebugWorldBloomRecord struct {
	EventID                 int `json:"eventId"`
	GameCharacterID         int `json:"gameCharacterId"`
	EventPoint              int `json:"eventPoint"`
	WorldBloomChapterPoint  int `json:"worldBloomChapterPoint"`
	WorldBloomChapterRank   int `json:"worldBloomChapterRank"`
	WorldBloomChapterNumber int `json:"chapterNo"`
}

type suiteDebugLeaderCountProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata                 `json:"userGamedata"`
	UserDecks    []suite.UserDeck                   `json:"userDecks"`
	UserCards    []suite.UserCard                   `json:"userCards"`
	Missions     []suiteDebugCharacterMission       `json:"userCharacterMissionV2s"`
	Statuses     []suiteDebugCharacterMissionStatus `json:"userCharacterMissionV2Statuses"`
}

type suiteDebugCharacterMission struct {
	CharacterID          int    `json:"characterId"`
	CharacterMissionType string `json:"characterMissionType"`
	Progress             int    `json:"progress"`
}

type suiteDebugCharacterMissionStatus struct {
	CharacterID      int    `json:"characterId"`
	ParameterGroupID int    `json:"parameterGroupId"`
	Seq              int    `json:"seq"`
	MissionStatus    string `json:"missionStatus"`
}

type suiteDebugLeaderCountRow struct {
	CharacterID int
	PlayLive    int
	PlayLiveEx  int
}

type suiteDebugMusicRewardProfile = suiteDebugMusicProgressProfile

type suiteDebugMusicAchievement struct {
	MusicID            int `json:"musicId"`
	MusicAchievementID int `json:"musicAchievementId"`
}

type suiteDebugMusicRewardRow struct {
	MusicID int
	Count   int
}

type suiteDebugCardBoxProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	UserDecks    []suite.UserDeck   `json:"userDecks"`
	UserCards    []suite.UserCard   `json:"userCards"`
}

type suiteDebugCardBoxOptions struct {
	ShowID            bool
	OwnedOnly         bool
	UseBeforeTraining bool
	ShowCreatedAt     bool
	SortBy            string
	FilterText        string
}

func (p suiteDebugCommonProfile) commonSuiteProfile() renderer.SuiteCommonProfile {
	return renderer.SuiteCommonProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p suiteDebugGachaHistoryProfile) commonSuiteProfile() renderer.SuiteCommonProfile {
	return renderer.SuiteCommonProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata}
}

func (p suiteDebugBondProfile) commonSuiteProfile() renderer.SuiteCommonProfile {
	return renderer.SuiteCommonProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p suiteDebugMusicProgressProfile) commonSuiteProfile() renderer.SuiteCommonProfile {
	return renderer.SuiteCommonProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p suiteDebugBest30Profile) commonSuiteProfile() renderer.SuiteCommonProfile {
	return renderer.SuiteCommonProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p suiteDebugMaterialProfile) commonSuiteProfile() renderer.SuiteCommonProfile {
	return renderer.SuiteCommonProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p suiteDebugChallengeProfile) commonSuiteProfile() renderer.SuiteCommonProfile {
	return renderer.SuiteCommonProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p suiteDebugEventRecordProfile) commonSuiteProfile() renderer.SuiteCommonProfile {
	return renderer.SuiteCommonProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p suiteDebugLeaderCountProfile) commonSuiteProfile() renderer.SuiteCommonProfile {
	return renderer.SuiteCommonProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p suiteDebugCardBoxProfile) commonSuiteProfile() renderer.SuiteCommonProfile {
	return renderer.SuiteCommonProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func newSuiteDebugPanel(runtime *servers.Runtime, title string, profile renderer.SuiteCommonProfile) renderer.SuitePanelPayload {
	payload := renderer.SuitePanelPayload{
		Title:       suiteDebugPanelTitle(runtime, title),
		Profile:     renderer.BuildSuiteProfilePayload(runtime.Region, "", profile.BaseProfile, profile.UserGamedata),
		AssetSource: "",
	}
	if runtime.Assets != nil {
		payload.AssetSource = runtime.Assets.RendererAssetSource()
	}
	payload.DeckCards = renderer.BuildSuiteDeckCards(profile.UserDecks, profile.UserCards, profile.UserGamedata.Deck, runtime.Store, runtime.Assets)
	return payload
}

func suiteDebugPanelTitle(runtime *servers.Runtime, name string) string {
	region := ""
	if runtime != nil {
		region = runtime.Region
	}
	return fmt.Sprintf("%s %s", strings.ToUpper(config.NormalizeRegion(region)), name)
}

func suiteDebugPanelSubtitle(profile suite.BaseProfile) string {
	return fmt.Sprintf("更新时间: %s · 数据来源: %s", suiteDebugUpdateText(profile.UploadTime), suiteDebugSourceText(profile))
}

func suiteDebugSourceText(profile suite.BaseProfile) string {
	source := profile.Source
	if source == "" {
		source = suite.PublicSource
	}
	if profile.LocalSource != "" {
		source += "(" + profile.LocalSource + ")"
	}
	return source
}

func suiteDebugUpdateText(uploadTime int64) string {
	uploadTime = normalizeSuiteDebugMillis(uploadTime)
	if uploadTime <= 0 {
		return "未知"
	}
	return time.UnixMilli(uploadTime).Format("2006-01-02 15:04:05")
}

func normalizeSuiteDebugMillis(value int64) int64 {
	if value > 0 && value < 100000000000 {
		return value * 1000
	}
	return value
}

func suiteDebugBasicStats(profile renderer.SuiteCommonProfile) []renderer.SuiteStatPayload {
	stats := make([]renderer.SuiteStatPayload, 0, 4)
	if profile.UserGamedata.Rank > 0 {
		stats = append(stats, renderer.SuiteStatPayload{Label: "Rank", Value: formatDebugInt(profile.UserGamedata.Rank)})
	}
	if profile.UserGamedata.Coin > 0 {
		stats = append(stats, renderer.SuiteStatPayload{Label: "金币", Value: formatDebugInt64(profile.UserGamedata.Coin)})
	}
	if len(profile.UserCards) > 0 {
		stats = append(stats, renderer.SuiteStatPayload{Label: "持有卡牌", Value: formatDebugInt(len(profile.UserCards))})
	}
	return stats
}

func suiteDebugSelected(def Definition, runtime *servers.Runtime, game suite.UserGamedata, resultType string) *EntityResult {
	name := strings.TrimSpace(game.Name)
	if name == "" {
		name = "未知玩家"
	}
	uid := game.UserID.String()
	if uid == "" {
		uid = "临时 UID"
	}
	return &EntityResult{ID: 0, Title: name, Subtitle: fmt.Sprintf("%s · UID %s · Haruki 公开 API", runtime.Label, uid), Type: resultType}
}

func (s *Service) best30Client() *b30.Client {
	if s != nil && s.B30 != nil {
		return s.B30
	}
	return b30.DefaultClient()
}

func suiteDebugRowsFromBest30(result b30.Result) []EntityResult {
	rows := make([]EntityResult, 0, minDebug(len(result.Entries), 12))
	for _, entry := range result.Entries {
		if len(rows) >= 12 {
			break
		}
		rows = append(rows, EntityResult{
			ID:       entry.Rank,
			Title:    fmt.Sprintf("#%02d %.1f %s", entry.Rank, entry.UserRating, entry.Title),
			Subtitle: fmt.Sprintf("%s %s · 定数 %.1f · %s", strings.ToUpper(entry.Difficulty), entry.PlayResult, entry.Constant, suiteDebugMusicName(nil, entry.MusicID)),
			Type:     "best30",
		})
	}
	return rows
}

func suiteDebugRowsFromSections(sections []renderer.SuiteSectionPayload) []EntityResult {
	rows := make([]EntityResult, 0)
	for _, section := range sections {
		for i, row := range section.Rows {
			id := row.Rank
			if id == 0 {
				id = i + 1
			}
			rows = append(rows, EntityResult{ID: id, Title: row.Label, Subtitle: strings.TrimSpace(strings.Join(nonEmpty(row.Value, row.Meta), " · ")), Type: section.Title})
			if len(rows) >= 12 {
				return rows
			}
		}
	}
	return rows
}

func suiteDebugRowsFromGachaHistory(profile suiteDebugGachaHistoryProfile, store *masterdata.Store, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	records := make([]suiteDebugGachaRecord, 0, len(profile.UserGachas))
	total := 0
	for _, record := range profile.UserGachas {
		if record.Count <= 0 {
			continue
		}
		records = append(records, record)
		total += record.Count
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Count == records[j].Count {
			return records[i].GachaID > records[j].GachaID
		}
		return records[i].Count > records[j].Count
	})
	limit = clampDebugLimit(limit, len(records))
	rows := make([]renderer.SuiteSectionRowPayload, 0, limit)
	for i := 0; i < limit; i++ {
		record := records[i]
		rows = append(rows, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: suiteDebugGachaName(store, record.GachaID), Value: fmt.Sprintf("%d抽", record.Count)})
	}
	return rows, []renderer.SuiteStatPayload{{Label: "总抽数", Value: formatDebugInt(total)}, {Label: "卡池数", Value: formatDebugInt(len(records))}}
}

func suiteDebugRowsFromBonds(profile suiteDebugBondProfile, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	bonds := make([]suiteDebugBond, 0, len(profile.UserBonds))
	for _, bond := range profile.UserBonds {
		cid1, cid2 := suiteDebugBondCharacterIDs(bond)
		if cid1 <= 0 || cid2 <= 0 {
			continue
		}
		bonds = append(bonds, bond)
	}
	sort.SliceStable(bonds, func(i, j int) bool {
		if bonds[i].Rank == bonds[j].Rank {
			return bonds[i].Exp > bonds[j].Exp
		}
		return bonds[i].Rank > bonds[j].Rank
	})
	limit = clampDebugLimit(limit, len(bonds))
	rows := make([]renderer.SuiteSectionRowPayload, 0, limit)
	maxRank := 0
	for i, bond := range bonds {
		if bond.Rank > maxRank {
			maxRank = bond.Rank
		}
		if i >= limit {
			continue
		}
		cid1, cid2 := suiteDebugBondCharacterIDs(bond)
		name1, name2 := suiteDebugCharacterName(cid1), suiteDebugCharacterName(cid2)
		rows = append(rows, renderer.SuiteSectionRowPayload{
			Rank:  i + 1,
			Label: name1 + " × " + name2,
			Value: fmt.Sprintf("Lv.%d", bond.Rank),
			Meta:  fmt.Sprintf("EXP %d", bond.Exp),
			Extra: map[string]interface{}{
				"characterId1":   cid1,
				"characterId2":   cid2,
				"characterName1": name1,
				"characterName2": name2,
				"rankLevel":      bond.Rank,
				"exp":            bond.Exp,
			},
		})
	}
	return rows, []renderer.SuiteStatPayload{{Label: "羁绊组数", Value: formatDebugInt(len(bonds))}, {Label: "最高羁绊", Value: fmt.Sprintf("Lv.%d", maxRank)}}
}

func suiteDebugRowsFromMusicProgress(profile suiteDebugMusicProgressProfile) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	summary := suiteDebugMusicProgressSummaryFromProfile(profile, nil)
	return summary.Rows, summary.Stats
}

func suiteDebugMusicProgressSummaryFromProfile(profile suiteDebugMusicProgressProfile, store *masterdata.Store) suiteDebugMusicProgressSummary {
	counts := suiteDebugMusicProgressCounts(profile)
	levelCounts := suiteDebugMusicProgressLevelCounts(profile, store)
	rows := make([]renderer.SuiteSectionRowPayload, 0, len(counts))
	totalPlayed, totalClear, totalFC, totalAP := 0, 0, 0, 0
	for _, diff := range suiteDebugMusicDifficultyOrder() {
		count := counts[diff]
		if count == nil {
			continue
		}
		totalPlayed += count.Played
		totalClear += count.Clear
		totalFC += count.FullCombo
		totalAP += count.AllPerfect
		rows = append(rows, renderer.SuiteSectionRowPayload{Label: strings.ToUpper(diff), Value: fmt.Sprintf("Clear %d / FC %d / AP %d", count.Clear, count.FullCombo, count.AllPerfect), Meta: fmt.Sprintf("游玩 %d", count.Played), Color: suiteDebugDifficultyColor(diff), Extra: map[string]interface{}{"diff": diff, "played": count.Played, "clear": count.Clear, "fc": count.FullCombo, "ap": count.AllPerfect}})
	}
	levelRows := suiteDebugMusicProgressLevelRows(levelCounts)
	return suiteDebugMusicProgressSummary{Rows: rows, LevelRows: levelRows, Stats: []renderer.SuiteStatPayload{{Label: "游玩", Value: formatDebugInt(totalPlayed)}, {Label: "Clear", Value: formatDebugInt(totalClear)}, {Label: "FC", Value: formatDebugInt(totalFC)}, {Label: "AP", Value: formatDebugInt(totalAP)}}, TotalSongs: totalPlayed, TotalClear: totalClear, TotalFC: totalFC, TotalAP: totalAP}
}

func suiteDebugRowsFromMaterials(profile suiteDebugMaterialProfile, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	materials := make([]suiteDebugMaterial, 0, len(profile.UserMaterials))
	for _, material := range profile.UserMaterials {
		if material.Quantity <= 0 {
			continue
		}
		materials = append(materials, material)
	}
	sort.SliceStable(materials, func(i, j int) bool {
		if materials[i].Quantity == materials[j].Quantity {
			return materials[i].MaterialID < materials[j].MaterialID
		}
		return materials[i].Quantity > materials[j].Quantity
	})
	limit = clampDebugLimit(limit, len(materials))
	rows := make([]renderer.SuiteSectionRowPayload, 0, limit)
	for i := 0; i < limit; i++ {
		material := materials[i]
		rows = append(rows, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: fmt.Sprintf("材料 #%d", material.MaterialID), Value: formatDebugInt64(material.Quantity)})
	}
	return rows, []renderer.SuiteStatPayload{{Label: "金币", Value: formatDebugInt64(profile.UserGamedata.Coin)}, {Label: "材料种类", Value: formatDebugInt(len(materials))}}
}

type suiteDebugMusicRewardSummary struct {
	RankJewelRemain    int
	RankRemainCount    int
	ValidMusicCount    int
	AchievementTotal   int
	AchievedMusicCount int
	ComboJewelRemain   int
	ComboShardRemain   int
	TotalJewelRemain   int
	TotalShardRemain   int
	ComboRows          []renderer.SuiteSectionRowPayload
	TopRows            []renderer.SuiteSectionRowPayload
}

type suiteDebugMusicProgressSummary struct {
	Rows       []renderer.SuiteSectionRowPayload
	LevelRows  []renderer.SuiteSectionRowPayload
	Stats      []renderer.SuiteStatPayload
	TotalSongs int
	TotalClear int
	TotalFC    int
	TotalAP    int
}

func suiteDebugRowsFromChallenge(profile suiteDebugChallengeProfile, store *masterdata.Store, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload, map[string]interface{}) {
	rowsByCharacter := suiteDebugChallengeRows(profile, store)
	completed := suiteDebugChallengeCompletedRewardIDs(profile.Rewards, store)
	for cid, row := range rowsByCharacter {
		row.RemainJewel, row.RemainFragment = suiteDebugChallengeRemainRewards(store, cid, completed[cid])
	}
	ordered := make([]suiteDebugChallengeRow, 0, 26)
	activeCount := 0
	maxScore := 0
	for cid := 1; cid <= 26; cid++ {
		row := rowsByCharacter[cid]
		if row == nil {
			row = &suiteDebugChallengeRow{CharacterID: cid}
			row.RemainJewel, row.RemainFragment = suiteDebugChallengeRemainRewards(store, cid, completed[cid])
		}
		if row.HighScore > 0 || row.Rank > 0 || row.RewardCount > 0 {
			activeCount++
		}
		if row.HighScore > maxScore {
			maxScore = row.HighScore
		}
		ordered = append(ordered, *row)
	}
	displayMax := maxDebug(maxDebug(maxScore, suiteDebugChallengeMasterMaxScore(store)), debugChallengeProgressMax)
	out := make([]renderer.SuiteSectionRowPayload, 0, len(ordered))
	totalRemainJewel, totalRemainFragment, totalRemainRewards := 0, 0, 0
	rankCounts := map[int]int{}
	for _, row := range ordered {
		rewardRemain := suiteDebugChallengeRemainRewardCount(store, row.CharacterID, completed[row.CharacterID])
		totalRemainJewel += row.RemainJewel
		totalRemainFragment += row.RemainFragment
		totalRemainRewards += rewardRemain
		rankCounts[row.Rank]++
		value := "-"
		if row.HighScore > 0 {
			value = formatDebugInt(row.HighScore)
		}
		rankText := "-"
		if row.Rank > 0 {
			rankText = fmt.Sprintf("Lv.%d", row.Rank)
		}
		out = append(out, renderer.SuiteSectionRowPayload{
			ID:            row.CharacterID,
			Rank:          row.CharacterID,
			Label:         suiteDebugCharacterName(row.CharacterID),
			Value:         value,
			Meta:          fmt.Sprintf("%s · 水晶 %d · 碎片 %d · 剩余档 %d", rankText, row.RemainJewel, row.RemainFragment, rewardRemain),
			CharacterID:   row.CharacterID,
			Progress:      float64(row.HighScore),
			ProgressMax:   float64(displayMax),
			ProgressLabel: fmt.Sprintf("%s / %s", formatDebugInt(row.HighScore), formatDebugInt(displayMax)),
			Extra: map[string]interface{}{
				"rankLevel":      row.Rank,
				"rewardCount":    row.RewardCount,
				"rewardRemain":   rewardRemain,
				"remainJewel":    row.RemainJewel,
				"remainFragment": row.RemainFragment,
				"highScore":      row.HighScore,
				"jewel":          row.RemainJewel,
				"shard":          row.RemainFragment,
			},
		})
	}
	extra := map[string]interface{}{"totalRemainJewel": totalRemainJewel, "totalRemainFragment": totalRemainFragment, "totalRemainRewards": totalRemainRewards, "totalRemainRewardSlots": totalRemainRewards, "rankDistribution": suiteDebugChallengeRankDistribution(rankCounts)}
	stats := []renderer.SuiteStatPayload{{Label: "角色数", Value: formatDebugInt(activeCount)}, {Label: "最高分", Value: formatDebugInt(maxScore)}, {Label: "剩余水晶", Value: formatDebugInt(totalRemainJewel)}, {Label: "剩余碎片", Value: formatDebugInt(totalRemainFragment)}, {Label: "剩余奖励档", Value: formatDebugInt(totalRemainRewards)}}
	return out, stats, extra
}

func suiteDebugChallengeRows(profile suiteDebugChallengeProfile, store *masterdata.Store) map[int]*suiteDebugChallengeRow {
	rowsByCharacter := map[int]*suiteDebugChallengeRow{}
	rewardByID := suiteDebugChallengeRewardMasterByID(store)
	for _, result := range profile.Results {
		if result.CharacterID <= 0 {
			continue
		}
		row := suiteDebugChallengeRowFor(rowsByCharacter, result.CharacterID)
		row.HighScore = maxDebug(row.HighScore, result.HighScore)
	}
	for _, stage := range profile.Stages {
		if stage.CharacterID <= 0 {
			continue
		}
		row := suiteDebugChallengeRowFor(rowsByCharacter, stage.CharacterID)
		row.Rank = maxDebug(row.Rank, stage.Rank)
	}
	for _, reward := range profile.Rewards {
		cid := suiteDebugChallengeRewardCharacterID(reward)
		if cid <= 0 {
			if masterReward, ok := rewardByID[suiteDebugChallengeRewardID(reward)]; ok {
				cid = masterReward.CharacterID
			}
		}
		if cid <= 0 {
			continue
		}
		row := suiteDebugChallengeRowFor(rowsByCharacter, cid)
		row.RewardCount++
	}
	return rowsByCharacter
}

func suiteDebugChallengeCompletedRewardIDs(rewards []suiteDebugChallengeReward, store *masterdata.Store) map[int]map[int]struct{} {
	out := map[int]map[int]struct{}{}
	rewardByID := suiteDebugChallengeRewardMasterByID(store)
	for _, reward := range rewards {
		rid := suiteDebugChallengeRewardID(reward)
		if rid <= 0 {
			continue
		}
		cid := suiteDebugChallengeRewardCharacterID(reward)
		if cid <= 0 {
			if masterReward, ok := rewardByID[rid]; ok {
				cid = masterReward.CharacterID
			}
		}
		if cid <= 0 {
			continue
		}
		ids := out[cid]
		if ids == nil {
			ids = map[int]struct{}{}
			out[cid] = ids
		}
		ids[rid] = struct{}{}
	}
	return out
}

func suiteDebugChallengeRewardMasterByID(store *masterdata.Store) map[int]masterdata.ChallengeLiveHighScoreReward {
	out := map[int]masterdata.ChallengeLiveHighScoreReward{}
	if store == nil {
		return out
	}
	for _, reward := range store.AllChallengeLiveHighScoreRewards() {
		out[reward.ID] = reward
	}
	return out
}

func suiteDebugChallengeRewardID(reward suiteDebugChallengeReward) int {
	if reward.RewardID > 0 {
		return reward.RewardID
	}
	if reward.ChallengeLiveSoloHighScoreRewardID > 0 {
		return reward.ChallengeLiveSoloHighScoreRewardID
	}
	return reward.RewardIDAlias
}

func suiteDebugChallengeRewardCharacterID(reward suiteDebugChallengeReward) int {
	if reward.CharacterID > 0 {
		return reward.CharacterID
	}
	return reward.GameCharacterID
}

func suiteDebugChallengeRemainRewards(store *masterdata.Store, characterID int, completed map[int]struct{}) (int, int) {
	if store == nil || characterID <= 0 {
		return 0, 0
	}
	jewel, fragment := 0, 0
	for _, reward := range store.GetChallengeLiveHighScoreRewards(characterID) {
		if _, ok := completed[reward.ID]; ok {
			continue
		}
		amount := suiteDebugCollectChallengeResourceBox(store, reward.ResourceBoxID)
		jewel += amount.Jewel
		fragment += amount.Fragment
	}
	return jewel, fragment
}

type suiteDebugChallengeRewardAmount struct {
	Jewel    int
	Fragment int
}

func suiteDebugCollectChallengeResourceBox(store *masterdata.Store, rootBoxID int) suiteDebugChallengeRewardAmount {
	return suiteDebugCollectChallengeResourceBoxWithVisited(store, rootBoxID, map[int]struct{}{})
}

func suiteDebugCollectChallengeResourceBoxWithVisited(store *masterdata.Store, boxID int, visited map[int]struct{}) suiteDebugChallengeRewardAmount {
	if store == nil || boxID <= 0 {
		return suiteDebugChallengeRewardAmount{}
	}
	if _, ok := visited[boxID]; ok {
		return suiteDebugChallengeRewardAmount{}
	}
	visited[boxID] = struct{}{}
	details := suiteDebugChallengeResourceBoxDetails(store, debugChallengeBoxPurpose, boxID)
	amount := suiteDebugChallengeRewardAmount{}
	for _, detail := range details {
		quantity := detail.ResourceQuantity
		if quantity <= 0 {
			continue
		}
		resourceType := strings.ToLower(detail.ResourceType)
		switch {
		case strings.Contains(resourceType, "jewel"):
			amount.Jewel += quantity
		case resourceType == "material" && detail.ResourceID == 15:
			amount.Fragment += quantity
		case strings.Contains(resourceType, "box"):
			nested := suiteDebugCollectChallengeResourceBoxWithVisited(store, detail.ResourceID, visited)
			amount.Jewel += nested.Jewel
			amount.Fragment += nested.Fragment
		}
	}
	return amount
}

func suiteDebugChallengeResourceBoxDetails(store *masterdata.Store, purpose string, boxID int) []masterdata.ResourceBoxDetail {
	if box := store.GetResourceBox(purpose, boxID); box != nil && len(box.Details) > 0 {
		return box.Details
	}
	return store.GetResourceBoxDetails(purpose, boxID)
}

func suiteDebugChallengeRemainRewardCount(store *masterdata.Store, characterID int, completed map[int]struct{}) int {
	if store == nil || characterID <= 0 {
		return 0
	}
	count := 0
	for _, reward := range store.GetChallengeLiveHighScoreRewards(characterID) {
		if _, ok := completed[reward.ID]; ok {
			continue
		}
		count++
	}
	return count
}

func suiteDebugChallengeRankDistribution(counts map[int]int) []map[string]interface{} {
	levels := make([]int, 0, len(counts))
	for level := range counts {
		levels = append(levels, level)
	}
	sort.SliceStable(levels, func(i, j int) bool { return levels[i] > levels[j] })
	out := make([]map[string]interface{}, 0, len(levels))
	for _, level := range levels {
		out = append(out, map[string]interface{}{"level": level, "count": counts[level], "label": suiteDebugChallengeRankLabel(level)})
	}
	return out
}

func suiteDebugChallengeRankLabel(level int) string {
	if level <= 0 {
		return "Lv.0"
	}
	return fmt.Sprintf("Lv.%d", level)
}

func suiteDebugChallengeMasterMaxScore(store *masterdata.Store) int {
	if store == nil {
		return 0
	}
	maxScore := 0
	for _, reward := range store.AllChallengeLiveHighScoreRewards() {
		if reward.HighScore > maxScore {
			maxScore = reward.HighScore
		}
	}
	return maxScore
}

func suiteDebugRowsFromEventRecord(profile suiteDebugEventRecordProfile, store *masterdata.Store, resolver interface{ GetEventBannerURL(string) string }, limit int) ([]renderer.SuiteSectionPayload, []renderer.SuiteStatPayload) {
	events := append([]suiteDebugEventRecord(nil), profile.UserEvents...)
	suiteDebugSortEventRecords(events)
	blooms := append([]suiteDebugWorldBloomRecord(nil), profile.UserWorldBlooms...)
	suiteDebugSortWorldBloomRecords(blooms)
	if limit <= 0 {
		limit = maxDebug(len(events), len(blooms))
	}
	sections := make([]renderer.SuiteSectionPayload, 0, 2)
	if len(events) > 0 {
		rows := make([]renderer.SuiteSectionRowPayload, 0, minDebug(limit, len(events)))
		for i := 0; i < minDebug(limit, len(events)); i++ {
			event := events[i]
			rows = append(rows, suiteDebugEventRecordRow(store, resolver, event.EventID, event.EventPoint, event.Rank, 0, 0, i+1))
		}
		sections = append(sections, renderer.SuiteSectionPayload{Title: "活动PT", Kind: "event_record", Note: "每次抓包仅包含最近活动记录；上传时增量更新，未上传过的记录可能缺失。", Rows: rows})
	}
	if len(blooms) > 0 {
		rows := make([]renderer.SuiteSectionRowPayload, 0, minDebug(limit, len(blooms)))
		for i := 0; i < minDebug(limit, len(blooms)); i++ {
			bloom := blooms[i]
			rows = append(rows, suiteDebugEventRecordRow(store, resolver, bloom.EventID, suiteDebugWorldBloomPoint(bloom), bloom.WorldBloomChapterRank, bloom.GameCharacterID, bloom.WorldBloomChapterNumber, i+1))
		}
		sections = append(sections, renderer.SuiteSectionPayload{Title: "WL章节", Kind: "event_record_wl", Note: "WL 章节记录按章节 PT 排序，角色头像来自本地 assets/characters。", Rows: rows})
	}
	return sections, []renderer.SuiteStatPayload{{Label: "活动记录", Value: formatDebugInt(len(events))}, {Label: "WL记录", Value: formatDebugInt(len(blooms))}}
}

func suiteDebugRowsFromLeaderCount(profile suiteDebugLeaderCountProfile, store *masterdata.Store, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload, map[string]interface{}) {
	rowsByCharacter := map[int]*suiteDebugLeaderCountRow{}
	for _, mission := range profile.Missions {
		if mission.CharacterID <= 0 {
			continue
		}
		row := rowsByCharacter[mission.CharacterID]
		if row == nil {
			row = &suiteDebugLeaderCountRow{CharacterID: mission.CharacterID}
			rowsByCharacter[mission.CharacterID] = row
		}
		switch mission.CharacterMissionType {
		case "play_live":
			row.PlayLive = maxDebug(row.PlayLive, mission.Progress)
		case "play_live_ex":
			row.PlayLiveEx = maxDebug(row.PlayLiveEx, mission.Progress)
		}
	}
	exLevels := suiteDebugLeaderExLevels(profile.Statuses)
	exTotals := suiteDebugLeaderExTotals(rowsByCharacter, exLevels, store)
	progressMax := suiteDebugLeaderProgressMax(store)
	groups := suiteDebugLeaderNormalGroups(store)
	maxLevel := len(groups)
	out := make([]renderer.SuiteSectionRowPayload, 0, 26)
	total, totalRemain, totalMissionLevel, totalMissionRemain, totalEx := 0, 0, 0, 0, 0
	activeCount := 0
	for cid := 1; cid <= 26; cid++ {
		row := rowsByCharacter[cid]
		if row == nil {
			row = &suiteDebugLeaderCountRow{CharacterID: cid}
		}
		total += row.PlayLive
		exTotal := exTotals[cid]
		totalEx += exTotal
		missionLevel := suiteDebugLeaderMissionLevel(groups, row.PlayLive)
		missionRemain := maxDebug(maxLevel-missionLevel, 0)
		nextNeed := suiteDebugLeaderNextNeed(groups, row.PlayLive)
		playLiveRemain := maxDebug(progressMax-row.PlayLive, 0)
		totalRemain += playLiveRemain
		totalMissionLevel += missionLevel
		totalMissionRemain += missionRemain
		if row.PlayLive > 0 || exTotal > 0 || exLevels[cid] > 0 {
			activeCount++
		}
		value := "-"
		if row.PlayLive > 0 {
			value = formatDebugInt(row.PlayLive)
		}
		out = append(out, renderer.SuiteSectionRowPayload{
			ID:            cid,
			Rank:          cid,
			Label:         suiteDebugCharacterName(cid),
			Value:         value,
			Meta:          fmt.Sprintf("剩余 %s · 档位 %d/%d · EX等级 x%d · EX次数 %s", formatDebugInt(playLiveRemain), missionLevel, maxLevel, exLevels[cid], suiteDebugDashInt(exTotal)),
			CharacterID:   cid,
			Progress:      float64(row.PlayLive),
			ProgressMax:   float64(progressMax),
			ProgressLabel: fmt.Sprintf("%s / %s", formatDebugInt(row.PlayLive), formatDebugInt(progressMax)),
			Extra: map[string]interface{}{
				"playLive":           row.PlayLive,
				"playLiveRemain":     playLiveRemain,
				"playLiveEx":         exTotal,
				"playLiveExRaw":      row.PlayLiveEx,
				"exLevel":            exLevels[cid],
				"missionLevel":       missionLevel,
				"missionLevelMax":    maxLevel,
				"missionLevelRemain": missionRemain,
				"nextNeed":           nextNeed,
				"progressRate":       suiteDebugProgressRate(float64(row.PlayLive), float64(progressMax)),
			},
		})
	}
	totalMissionMax := maxLevel * 26
	extra := map[string]interface{}{"totalPlayLive": total, "totalRemain": totalRemain, "totalMissionLevel": totalMissionLevel, "totalMissionMax": totalMissionMax, "totalMissionRemain": totalMissionRemain, "totalEx": totalEx, "progressMax": progressMax}
	stats := []renderer.SuiteStatPayload{{Label: "总队长次数", Value: formatDebugInt(total)}, {Label: "剩余总次数", Value: formatDebugInt(totalRemain)}, {Label: "普通档位", Value: fmt.Sprintf("%d/%d", totalMissionLevel, totalMissionMax)}, {Label: "剩余档位", Value: formatDebugInt(totalMissionRemain)}, {Label: "EX总次数", Value: formatDebugInt(totalEx)}, {Label: "角色数", Value: formatDebugInt(activeCount)}}
	return out, stats, extra
}

func suiteDebugRowsFromMusicReward(profile suiteDebugMusicRewardProfile, store *masterdata.Store, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	summary := suiteDebugMusicRewardSummaryFromProfile(profile, store, limit)
	return summary.TopRows, []renderer.SuiteStatPayload{{Label: "S评级剩余", Value: formatDebugInt(summary.RankJewelRemain)}, {Label: "连击剩余", Value: suiteDebugFormatRewardTotal(summary.ComboRows)}, {Label: "涉及歌曲", Value: formatDebugInt(summary.AchievedMusicCount)}}
}

func suiteDebugSectionsFromMusicReward(profile suiteDebugMusicRewardProfile, store *masterdata.Store, limit int) ([]renderer.SuiteSectionPayload, []renderer.SuiteStatPayload) {
	summary := suiteDebugMusicRewardSummaryFromProfile(profile, store, limit)
	sections := suiteDebugMusicRewardSections(summary)
	return sections, []renderer.SuiteStatPayload{{Label: "S评级剩余", Value: formatDebugInt(summary.RankJewelRemain)}, {Label: "剩余连击奖励", Value: suiteDebugFormatRewardTotal(summary.ComboRows)}, {Label: "有效歌曲", Value: formatDebugInt(summary.ValidMusicCount)}}
}

func suiteDebugSectionsFromMusicOverview(profile suiteDebugMusicProgressProfile, store *masterdata.Store, limit int) ([]renderer.SuiteSectionPayload, []renderer.SuiteStatPayload) {
	progress := suiteDebugMusicProgressSummaryFromProfile(profile, store)
	reward := suiteDebugMusicRewardSummaryFromProfile(profile, store, limit)
	sections := []renderer.SuiteSectionPayload{{Title: "打歌进度", Kind: "music_progress_summary", Note: "按歌曲+难度去重后统计；同一谱面取最佳 Clear/FC/AP 状态。", Rows: progress.Rows, Extra: map[string]interface{}{"totalPlayed": progress.TotalSongs, "totalClear": progress.TotalClear, "totalFC": progress.TotalFC, "totalAP": progress.TotalAP}}}
	if len(progress.LevelRows) > 0 {
		sections = append(sections, renderer.SuiteSectionPayload{Title: "等级数量", Kind: "music_progress_level", Rows: progress.LevelRows})
	}
	sections = append(sections, suiteDebugMusicRewardSections(reward)...)
	stats := append([]renderer.SuiteStatPayload{}, progress.Stats...)
	stats = append(stats, renderer.SuiteStatPayload{Label: "S评级剩余", Value: formatDebugInt(reward.RankJewelRemain)}, renderer.SuiteStatPayload{Label: "连击剩余", Value: suiteDebugFormatRewardTotal(reward.ComboRows)})
	return sections, stats
}

func suiteDebugMusicRewardSections(summary suiteDebugMusicRewardSummary) []renderer.SuiteSectionPayload {
	sections := []renderer.SuiteSectionPayload{{
		Title: "歌曲评级奖励(S)",
		Kind:  "music_reward_summary",
		Note:  "参考 lunabot：统计尚未获得的 S 评级水晶奖励；连击奖励按谱面等级汇总剩余值。",
		Rows: []renderer.SuiteSectionRowPayload{
			{Label: "S评级剩余水晶", Value: formatDebugInt(summary.RankJewelRemain), Meta: fmt.Sprintf("%d首未达成 / 共%d首", summary.RankRemainCount, summary.ValidMusicCount), Extra: map[string]interface{}{"rewardType": "jewel", "amount": summary.RankJewelRemain, "remainCount": summary.RankRemainCount, "validMusicCount": summary.ValidMusicCount}},
			{Label: "已达成奖励", Value: formatDebugInt(summary.AchievementTotal), Meta: fmt.Sprintf("涉及%d首歌曲", summary.AchievedMusicCount), Extra: map[string]interface{}{"achievementTotal": summary.AchievementTotal, "achievedMusicCount": summary.AchievedMusicCount}},
		},
		Extra: suiteDebugMusicRewardExtra(summary),
	}}
	if len(summary.ComboRows) > 0 {
		sections = append(sections, renderer.SuiteSectionPayload{Title: "连击奖励剩余", Kind: "music_reward_combo", Rows: summary.ComboRows, Extra: suiteDebugMusicRewardComboExtra(summary)})
	}
	if len(summary.TopRows) > 0 {
		sections = append(sections, renderer.SuiteSectionPayload{Title: "已达成奖励 TOP", Kind: "music_reward_achieved", Rows: summary.TopRows})
	}
	return sections
}

func suiteDebugMusicRewardExtra(summary suiteDebugMusicRewardSummary) map[string]interface{} {
	return map[string]interface{}{
		"rankJewelRemain":    summary.RankJewelRemain,
		"rankRemainCount":    summary.RankRemainCount,
		"validMusicCount":    summary.ValidMusicCount,
		"achievementTotal":   summary.AchievementTotal,
		"achievedMusicCount": summary.AchievedMusicCount,
		"comboJewelRemain":   summary.ComboJewelRemain,
		"comboShardRemain":   summary.ComboShardRemain,
		"totalJewelRemain":   summary.TotalJewelRemain,
		"totalShardRemain":   summary.TotalShardRemain,
	}
}

func suiteDebugMusicRewardComboExtra(summary suiteDebugMusicRewardSummary) map[string]interface{} {
	extra := suiteDebugMusicRewardExtra(summary)
	extra["total"] = suiteDebugFormatRewardTotal(summary.ComboRows)
	return extra
}

func suiteDebugEventRecordRow(store *masterdata.Store, resolver interface{ GetEventBannerURL(string) string }, eventID int, point int, rank int, characterID int, chapterNo int, order int) renderer.SuiteSectionRowPayload {
	label := suiteDebugEventName(store, eventID)
	var startAt, endAt int64
	bannerURL := ""
	dateText := ""
	if store != nil {
		if event := store.GetEvent(eventID); event != nil {
			label = event.Name
			if strings.TrimSpace(label) == "" {
				label = fmt.Sprintf("活动 #%d", eventID)
			}
			startAt = event.StartAt
			endAt = event.AggregateAt
			if endAt <= 0 {
				endAt = event.ClosedAt
			}
			dateText = suiteDebugEventDateRange(startAt, endAt)
			if resolver != nil && event.AssetbundleName != "" {
				bannerURL = resolver.GetEventBannerURL(event.AssetbundleName)
			}
		}
	}
	metaParts := []string{}
	if rank > 0 {
		metaParts = append(metaParts, formatDebugRank(rank))
	}
	if characterID > 0 {
		metaParts = append(metaParts, suiteDebugCharacterName(characterID))
	}
	if chapterNo > 0 {
		metaParts = append(metaParts, fmt.Sprintf("第%d章", chapterNo))
	}
	return renderer.SuiteSectionRowPayload{
		ID:          eventID,
		Rank:        order,
		Label:       label,
		Value:       fmt.Sprintf("%dpt", point),
		Meta:        strings.Join(metaParts, " · "),
		EventID:     eventID,
		CharacterID: characterID,
		BannerURL:   bannerURL,
		DateText:    dateText,
		StartAt:     startAt,
		EndAt:       endAt,
		Extra: map[string]interface{}{
			"point":     point,
			"rank":      rank,
			"chapterNo": chapterNo,
		},
	}
}

func suiteDebugSortEventRecords(events []suiteDebugEventRecord) {
	hasRank := false
	for _, event := range events {
		if event.Rank > 0 {
			hasRank = true
			break
		}
	}
	sort.SliceStable(events, func(i, j int) bool {
		if hasRank {
			ir, jr := suiteDebugNormalizedRank(events[i].Rank), suiteDebugNormalizedRank(events[j].Rank)
			if ir != jr {
				return ir < jr
			}
		}
		return events[i].EventPoint > events[j].EventPoint
	})
}

func suiteDebugSortWorldBloomRecords(blooms []suiteDebugWorldBloomRecord) {
	hasRank := false
	for _, bloom := range blooms {
		if bloom.WorldBloomChapterRank > 0 {
			hasRank = true
			break
		}
	}
	sort.SliceStable(blooms, func(i, j int) bool {
		if hasRank {
			ir, jr := suiteDebugNormalizedRank(blooms[i].WorldBloomChapterRank), suiteDebugNormalizedRank(blooms[j].WorldBloomChapterRank)
			if ir != jr {
				return ir < jr
			}
		}
		return suiteDebugWorldBloomPoint(blooms[i]) > suiteDebugWorldBloomPoint(blooms[j])
	})
}

func suiteDebugLeaderExLevels(statuses []suiteDebugCharacterMissionStatus) map[int]int {
	out := map[int]int{}
	for _, status := range statuses {
		if status.CharacterID <= 0 || status.ParameterGroupID != 101 {
			continue
		}
		out[status.CharacterID] = maxDebug(out[status.CharacterID], status.Seq)
	}
	return out
}

func suiteDebugLeaderExTotals(rows map[int]*suiteDebugLeaderCountRow, exLevels map[int]int, store *masterdata.Store) map[int]int {
	out := map[int]int{}
	for cid := 1; cid <= 26; cid++ {
		progressRaw := 0
		if row := rows[cid]; row != nil {
			progressRaw = row.PlayLiveEx
		}
		clearedTotal := suiteDebugLeaderExClearedTotal(store, exLevels[cid])
		currentTotal := progressRaw
		if progressRaw < clearedTotal {
			currentTotal = clearedTotal + progressRaw
		} else if progressRaw == 0 {
			currentTotal = clearedTotal
		}
		out[cid] = currentTotal
	}
	return out
}

func suiteDebugLeaderExClearedTotal(store *masterdata.Store, seq int) int {
	if store == nil || seq <= 0 {
		return 0
	}
	groups := store.GetCharacterMissionV2ParameterGroups(101)
	if len(groups) == 0 {
		return 0
	}
	total := 0
	for round := 1; round <= seq; round++ {
		total += suiteDebugLeaderRequirementForRound(groups, round)
	}
	return total
}

func suiteDebugLeaderRequirementForRound(groups []masterdata.CharacterMissionV2ParameterGroup, round int) int {
	req := 0
	for _, group := range groups {
		if group.Seq > round {
			break
		}
		req = group.Requirement
	}
	return req
}

func suiteDebugLeaderProgressMax(store *masterdata.Store) int {
	if store == nil {
		return debugLeaderCountProgressMax
	}
	maxReq := 0
	for _, group := range store.GetCharacterMissionV2ParameterGroups(1) {
		if group.Requirement > maxReq {
			maxReq = group.Requirement
		}
	}
	if maxReq <= 0 {
		return debugLeaderCountProgressMax
	}
	return maxReq
}

func suiteDebugLeaderNormalGroups(store *masterdata.Store) []masterdata.CharacterMissionV2ParameterGroup {
	if store == nil {
		return nil
	}
	return store.GetCharacterMissionV2ParameterGroups(1)
}

func suiteDebugLeaderMissionLevel(groups []masterdata.CharacterMissionV2ParameterGroup, playLive int) int {
	level := 0
	for _, group := range groups {
		if group.Requirement <= playLive {
			level++
		}
	}
	return level
}

func suiteDebugLeaderNextNeed(groups []masterdata.CharacterMissionV2ParameterGroup, playLive int) int {
	for _, group := range groups {
		if group.Requirement > playLive {
			return group.Requirement - playLive
		}
	}
	return 0
}

func suiteDebugMusicRewardSummaryFromProfile(profile suiteDebugMusicRewardProfile, store *masterdata.Store, limit int) suiteDebugMusicRewardSummary {
	achievements := suiteDebugMusicAchievementsByMusic(profile.Achievements)
	validMusics := suiteDebugValidMusicRewardMusics(store, achievements)
	summary := suiteDebugMusicRewardSummary{ValidMusicCount: len(validMusics), AchievedMusicCount: len(achievements)}
	for _, ids := range achievements {
		summary.AchievementTotal += len(ids)
	}
	for _, music := range validMusics {
		ids := achievements[music.ID]
		if !ids[debugMusicRewardRankRewardID] {
			summary.RankJewelRemain += suiteDebugMusicRankRewards[debugMusicRewardRankRewardID].Jewel
			summary.RankRemainCount++
		}
	}
	summary.ComboRows = suiteDebugMusicRewardComboRows(store, validMusics, achievements)
	summary.ComboJewelRemain, summary.ComboShardRemain = suiteDebugRewardTotals(summary.ComboRows)
	summary.TotalJewelRemain = summary.RankJewelRemain + summary.ComboJewelRemain
	summary.TotalShardRemain = summary.ComboShardRemain
	summary.TopRows = suiteDebugMusicRewardTopRows(store, achievements, limit)
	return summary
}

func suiteDebugMusicAchievementsByMusic(achievements []suiteDebugMusicAchievement) map[int]map[int]bool {
	out := map[int]map[int]bool{}
	for _, achievement := range achievements {
		if achievement.MusicID <= 0 || achievement.MusicAchievementID <= 0 {
			continue
		}
		ids := out[achievement.MusicID]
		if ids == nil {
			ids = map[int]bool{}
			out[achievement.MusicID] = ids
		}
		ids[achievement.MusicAchievementID] = true
	}
	return out
}

func suiteDebugValidMusicRewardMusics(store *masterdata.Store, achievements map[int]map[int]bool) []masterdata.MusicInfo {
	if store != nil && store.IsLoaded() {
		musics := store.AllMusics()
		now := time.Now().UnixMilli()
		out := make([]masterdata.MusicInfo, 0, len(musics))
		for _, music := range musics {
			if music.ID <= 0 {
				continue
			}
			if music.PublishedAt > 0 && music.PublishedAt > now {
				continue
			}
			out = append(out, music)
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
		return out
	}
	ids := make([]int, 0, len(achievements))
	for mid := range achievements {
		ids = append(ids, mid)
	}
	sort.Ints(ids)
	out := make([]masterdata.MusicInfo, 0, len(ids))
	for _, mid := range ids {
		out = append(out, masterdata.MusicInfo{ID: mid, Title: fmt.Sprintf("歌曲 #%d", mid)})
	}
	return out
}

func suiteDebugMusicRewardComboRows(store *masterdata.Store, musics []masterdata.MusicInfo, achievements map[int]map[int]bool) []renderer.SuiteSectionRowPayload {
	type rewardBucket struct {
		Diff       string
		Level      int
		Amount     int
		RewardType string
		Count      int
	}
	buckets := map[string]*rewardBucket{}
	for _, music := range musics {
		ids := achievements[music.ID]
		for _, diff := range []string{"hard", "expert", "master", "append"} {
			level := suiteDebugMusicDifficultyLevel(store, music.ID, diff)
			if level <= 0 {
				continue
			}
			amount := 0
			rewardType := "jewel"
			for achievementID, reward := range suiteDebugMusicComboRewards[diff] {
				if ids[achievementID] {
					continue
				}
				if diff == "append" {
					amount += reward.Shard
					rewardType = "shard"
				} else {
					amount += reward.Jewel
				}
			}
			if amount <= 0 {
				continue
			}
			key := fmt.Sprintf("%s:%d", diff, level)
			bucket := buckets[key]
			if bucket == nil {
				bucket = &rewardBucket{Diff: diff, Level: level, RewardType: rewardType}
				buckets[key] = bucket
			}
			bucket.Amount += amount
			bucket.Count++
		}
	}
	ordered := make([]*rewardBucket, 0, len(buckets))
	for _, bucket := range buckets {
		ordered = append(ordered, bucket)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if suiteDebugDiffOrder(ordered[i].Diff) != suiteDebugDiffOrder(ordered[j].Diff) {
			return suiteDebugDiffOrder(ordered[i].Diff) < suiteDebugDiffOrder(ordered[j].Diff)
		}
		return ordered[i].Level < ordered[j].Level
	})
	accByDiff := map[string]int{}
	rows := make([]renderer.SuiteSectionRowPayload, 0, len(ordered))
	for _, bucket := range ordered {
		accByDiff[bucket.Diff] += bucket.Amount
		rows = append(rows, renderer.SuiteSectionRowPayload{
			Label: bucket.Diff,
			Value: formatDebugInt(bucket.Amount),
			Meta:  fmt.Sprintf("Lv.%d · 累计 %d · %d谱面", bucket.Level, accByDiff[bucket.Diff], bucket.Count),
			Color: suiteDebugDifficultyColor(bucket.Diff),
			Extra: map[string]interface{}{
				"diff":       bucket.Diff,
				"level":      bucket.Level,
				"amount":     bucket.Amount,
				"accumulate": accByDiff[bucket.Diff],
				"rewardType": bucket.RewardType,
				"count":      bucket.Count,
			},
		})
	}
	return rows
}

func suiteDebugMusicRewardTopRows(store *masterdata.Store, achievements map[int]map[int]bool, limit int) []renderer.SuiteSectionRowPayload {
	rows := make([]suiteDebugMusicRewardRow, 0, len(achievements))
	for mid, ids := range achievements {
		rows = append(rows, suiteDebugMusicRewardRow{MusicID: mid, Count: len(ids)})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Count == rows[j].Count {
			return rows[i].MusicID < rows[j].MusicID
		}
		return rows[i].Count > rows[j].Count
	})
	limit = clampDebugLimit(limit, len(rows))
	out := make([]renderer.SuiteSectionRowPayload, 0, limit)
	for i := 0; i < limit; i++ {
		row := rows[i]
		out = append(out, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: suiteDebugMusicName(store, row.MusicID), Value: formatDebugInt(row.Count), Meta: "已达成奖励", MusicID: row.MusicID})
	}
	return out
}

func suiteDebugMusicDifficultyLevel(store *masterdata.Store, musicID int, diff string) int {
	if store == nil {
		return 0
	}
	for _, item := range store.GetMusicDifficulties(musicID) {
		if strings.EqualFold(item.MusicDifficulty, diff) {
			return item.PlayLevel
		}
	}
	return 0
}

func suiteDebugEventDateRange(startAt int64, endAt int64) string {
	if startAt <= 0 && endAt <= 0 {
		return ""
	}
	if startAt <= 0 {
		return suiteDebugFormatEventDate(endAt)
	}
	if endAt <= 0 {
		return suiteDebugFormatEventDate(startAt)
	}
	return fmt.Sprintf("%s - %s", suiteDebugFormatEventDate(startAt), suiteDebugFormatEventDate(endAt))
}

func suiteDebugFormatEventDate(value int64) string {
	if value <= 0 {
		return "-"
	}
	return time.UnixMilli(normalizeSuiteDebugMillis(value)).Format("2006-01-02")
}

func suiteDebugNormalizedRank(rank int) int {
	if rank <= 0 {
		return 1 << 30
	}
	return rank
}

func suiteDebugProgressRate(value float64, maxValue float64) float64 {
	if maxValue <= 0 || value <= 0 {
		return 0
	}
	if value >= maxValue {
		return 1
	}
	return value / maxValue
}

func suiteDebugDashInt(value int) string {
	if value <= 0 {
		return "-"
	}
	return formatDebugInt(value)
}

func suiteDebugDiffOrder(diff string) int {
	switch strings.ToLower(diff) {
	case "easy":
		return 1
	case "normal":
		return 2
	case "hard":
		return 3
	case "expert":
		return 4
	case "master":
		return 5
	case "append":
		return 6
	default:
		return 99
	}
}

func suiteDebugRewardTotals(rows []renderer.SuiteSectionRowPayload) (int, int) {
	jewel, shard := 0, 0
	for _, row := range rows {
		if row.Extra == nil {
			continue
		}
		amount, _ := row.Extra["amount"].(int)
		rewardType, _ := row.Extra["rewardType"].(string)
		if rewardType == "shard" {
			shard += amount
		} else {
			jewel += amount
		}
	}
	return jewel, shard
}

func suiteDebugFormatRewardTotal(rows []renderer.SuiteSectionRowPayload) string {
	jewel, shard := suiteDebugRewardTotals(rows)
	parts := []string{}
	if jewel > 0 {
		parts = append(parts, fmt.Sprintf("%d水晶", jewel))
	}
	if shard > 0 {
		parts = append(parts, fmt.Sprintf("%d碎片", shard))
	}
	if len(parts) == 0 {
		return "0"
	}
	return strings.Join(parts, " / ")
}

func suiteDebugParseCardBoxOptions(raw string) suiteDebugCardBoxOptions {
	options := suiteDebugCardBoxOptions{FilterText: strings.TrimSpace(raw)}
	tokens := strings.Fields(raw)
	remaining := make([]string, 0, len(tokens))
	for _, token := range tokens {
		lower := strings.ToLower(strings.TrimSpace(token))
		switch lower {
		case "box", "owned", "持有", "已持有":
			options.OwnedOnly = true
		case "id", "ids", "编号", "显示id":
			options.ShowID = true
		case "before", "normal", "花前", "特训前":
			options.UseBeforeTraining = true
		case "time", "created", "createdat", "获取时间", "入手时间", "时间排序":
			options.SortBy = "time"
			options.ShowCreatedAt = true
		case "mr", "master", "masterrank", "专精", "专精排序", "rank":
			options.SortBy = "mr"
		case "sl", "skill", "skilllevel", "技能等级", "技能等级排序":
			options.SortBy = "sl"
		default:
			remaining = append(remaining, token)
		}
	}
	options.FilterText = strings.TrimSpace(strings.Join(remaining, " "))
	return options
}

func suiteDebugCardBoxCards(store *masterdata.Store, options suiteDebugCardBoxOptions) ([]masterdata.CardInfo, string) {
	if store == nil {
		return nil, "卡牌数据不可用"
	}
	filter := strings.TrimSpace(options.FilterText)
	if filter != "" {
		result := cardquery.ResolveAll(store, filter)
		if result.Message != "" {
			return nil, result.Message
		}
		return append([]masterdata.CardInfo(nil), result.Cards...), ""
	}
	cards := store.AllCards()
	suiteDebugSortCards(cards)
	return cards, ""
}

func suiteDebugSortCards(cards []masterdata.CardInfo) {
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].CharacterID != cards[j].CharacterID {
			return cards[i].CharacterID < cards[j].CharacterID
		}
		if cards[i].ReleaseAt != cards[j].ReleaseAt {
			return cards[i].ReleaseAt < cards[j].ReleaseAt
		}
		return cards[i].ID < cards[j].ID
	})
}

func suiteDebugCardBoxDeckSet(profile suiteDebugCardBoxProfile) map[int]struct{} {
	deckCards := renderer.BuildSuiteDeckCards(profile.UserDecks, profile.UserCards, profile.UserGamedata.Deck, nil, nil)
	out := make(map[int]struct{}, len(deckCards))
	for _, card := range deckCards {
		if card.CardID > 0 {
			out[card.CardID] = struct{}{}
		}
	}
	return out
}

func suiteDebugCardBoxSubtitle(options suiteDebugCardBoxOptions, total int, owned int) string {
	parts := []string{fmt.Sprintf("筛选 %d 张", total), fmt.Sprintf("已持有 %d 张", owned)}
	if options.FilterText != "" {
		parts = append(parts, "条件: "+options.FilterText)
	}
	if options.OwnedOnly {
		parts = append(parts, "仅显示持有")
	}
	if options.SortBy != "" {
		parts = append(parts, "排序: "+options.SortBy)
	}
	return strings.Join(parts, " · ")
}

func suiteDebugMusicProgressCounts(profile suiteDebugMusicProgressProfile) map[string]*suiteDebugMusicProgressCount {
	counts := map[string]*suiteDebugMusicProgressCount{}
	best := suiteDebugBestMusicResults(profile.UserMusicResults)
	for _, result := range best {
		diff := suiteDebugMusicResultDifficulty(result)
		if diff == "" {
			continue
		}
		count := counts[diff]
		if count == nil {
			count = &suiteDebugMusicProgressCount{}
			counts[diff] = count
		}
		count.Played++
		if suiteDebugMusicResultCleared(result) {
			count.Clear++
		}
		if suiteDebugMusicResultFullCombo(result) {
			count.FullCombo++
		}
		if suiteDebugMusicResultAllPerfect(result) {
			count.AllPerfect++
		}
	}
	return counts
}

func suiteDebugBestMusicResults(results []suiteDebugMusicResult) map[string]suiteDebugMusicResult {
	best := map[string]suiteDebugMusicResult{}
	for _, result := range results {
		diff := suiteDebugMusicResultDifficulty(result)
		if result.MusicID <= 0 || diff == "" {
			continue
		}
		key := fmt.Sprintf("%d:%s", result.MusicID, diff)
		if prev, ok := best[key]; !ok || suiteDebugMusicResultRank(result) > suiteDebugMusicResultRank(prev) {
			best[key] = result
		}
	}
	return best
}

func suiteDebugMusicResultRank(result suiteDebugMusicResult) int {
	if suiteDebugMusicResultAllPerfect(result) {
		return 3
	}
	if suiteDebugMusicResultFullCombo(result) {
		return 2
	}
	if suiteDebugMusicResultCleared(result) {
		return 1
	}
	return 0
}

func suiteDebugMusicProgressLevelCounts(profile suiteDebugMusicProgressProfile, store *masterdata.Store) map[string]map[int]*suiteDebugMusicProgressCount {
	counts := map[string]map[int]*suiteDebugMusicProgressCount{}
	if store != nil && store.IsLoaded() {
		now := time.Now().UnixMilli()
		for _, music := range store.AllMusics() {
			if music.ID <= 0 || (music.PublishedAt > 0 && music.PublishedAt > now) {
				continue
			}
			for _, diffInfo := range store.GetMusicDifficulties(music.ID) {
				diff := strings.ToLower(diffInfo.MusicDifficulty)
				if diff == "" || diffInfo.PlayLevel <= 0 {
					continue
				}
				suiteDebugMusicProgressLevelCount(counts, diff, diffInfo.PlayLevel).Total++
			}
		}
	}
	for _, result := range suiteDebugBestMusicResults(profile.UserMusicResults) {
		diff := suiteDebugMusicResultDifficulty(result)
		level := suiteDebugMusicDifficultyLevel(store, result.MusicID, diff)
		if diff == "" || level <= 0 {
			continue
		}
		count := suiteDebugMusicProgressLevelCount(counts, diff, level)
		if count.Total <= 0 {
			count.Total = 1
		}
		count.Played++
		if suiteDebugMusicResultCleared(result) {
			count.Clear++
		}
		if suiteDebugMusicResultFullCombo(result) {
			count.FullCombo++
		}
		if suiteDebugMusicResultAllPerfect(result) {
			count.AllPerfect++
		}
	}
	return counts
}

func suiteDebugMusicProgressLevelCount(counts map[string]map[int]*suiteDebugMusicProgressCount, diff string, level int) *suiteDebugMusicProgressCount {
	byLevel := counts[diff]
	if byLevel == nil {
		byLevel = map[int]*suiteDebugMusicProgressCount{}
		counts[diff] = byLevel
	}
	count := byLevel[level]
	if count == nil {
		count = &suiteDebugMusicProgressCount{}
		byLevel[level] = count
	}
	return count
}

func suiteDebugMusicProgressLevelRows(counts map[string]map[int]*suiteDebugMusicProgressCount) []renderer.SuiteSectionRowPayload {
	rows := []renderer.SuiteSectionRowPayload{}
	for _, diff := range suiteDebugMusicDifficultyOrder() {
		byLevel := counts[diff]
		if len(byLevel) == 0 {
			continue
		}
		levels := make([]int, 0, len(byLevel))
		for level := range byLevel {
			levels = append(levels, level)
		}
		sort.Ints(levels)
		for _, level := range levels {
			count := byLevel[level]
			total := count.Total
			if total < count.Played {
				total = count.Played
			}
			notPlayed := maxDebug(total-count.Played, 0)
			clearOnly := maxDebug(count.Clear-count.FullCombo, 0)
			fcOnly := maxDebug(count.FullCombo-count.AllPerfect, 0)
			rows = append(rows, renderer.SuiteSectionRowPayload{Label: strings.ToUpper(diff), Value: fmt.Sprintf("Clear %d / FC %d / AP %d", count.Clear, count.FullCombo, count.AllPerfect), Meta: fmt.Sprintf("Lv.%d · 游玩 %d/%d", level, count.Played, total), Color: suiteDebugDifficultyColor(diff), Extra: map[string]interface{}{"diff": diff, "level": level, "total": total, "played": count.Played, "clear": count.Clear, "fc": count.FullCombo, "ap": count.AllPerfect, "notPlayed": notPlayed, "clearOnly": clearOnly, "fcOnly": fcOnly, "apOnly": count.AllPerfect}})
		}
	}
	return rows
}

func suiteDebugMusicDifficultyOrder() []string {
	return []string{"easy", "normal", "hard", "expert", "master", "append"}
}

func suiteDebugMusicResultDifficulty(result suiteDebugMusicResult) string {
	if result.MusicDifficultyType != "" {
		return strings.ToLower(result.MusicDifficultyType)
	}
	return strings.ToLower(result.MusicDifficulty)
}

func suiteDebugMusicResultCleared(result suiteDebugMusicResult) bool {
	return result.PlayResult != "" && result.PlayResult != "not_clear"
}

func suiteDebugMusicResultFullCombo(result suiteDebugMusicResult) bool {
	return result.FullComboFlg || result.FullPerfectFlg || result.PlayResult == "full_combo" || result.PlayResult == "all_perfect"
}

func suiteDebugMusicResultAllPerfect(result suiteDebugMusicResult) bool {
	return result.FullPerfectFlg || result.PlayResult == "all_perfect"
}

func suiteDebugBondCharacterIDs(bond suiteDebugBond) (int, int) {
	cid1, cid2 := bond.CharacterID1, bond.CharacterID2
	if cid1 == 0 {
		cid1 = bond.GameCharacterID1
	}
	if cid2 == 0 {
		cid2 = bond.GameCharacterID2
	}
	if (cid1 == 0 || cid2 == 0) && bond.BondsGroupID > 0 {
		cid1 = bond.BondsGroupID / 100 % 100
		cid2 = bond.BondsGroupID % 100
	}
	return cid1, cid2
}

func suiteDebugChallengeRowFor(rows map[int]*suiteDebugChallengeRow, characterID int) *suiteDebugChallengeRow {
	row := rows[characterID]
	if row == nil {
		row = &suiteDebugChallengeRow{CharacterID: characterID}
		rows[characterID] = row
	}
	return row
}

func suiteDebugWorldBloomPoint(record suiteDebugWorldBloomRecord) int {
	if record.WorldBloomChapterPoint > 0 {
		return record.WorldBloomChapterPoint
	}
	return record.EventPoint
}

func suiteDebugGachaName(store *masterdata.Store, gachaID int) string {
	if store != nil {
		if gacha := store.GetGacha(gachaID); gacha != nil && strings.TrimSpace(gacha.Name) != "" {
			return fmt.Sprintf("#%d %s", gachaID, gacha.Name)
		}
	}
	return fmt.Sprintf("未知卡池 #%d", gachaID)
}

func suiteDebugEventName(store *masterdata.Store, eventID int) string {
	if store != nil {
		if event := store.GetEvent(eventID); event != nil && strings.TrimSpace(event.Name) != "" {
			return fmt.Sprintf("#%d %s", eventID, event.Name)
		}
	}
	return fmt.Sprintf("活动 #%d", eventID)
}

func suiteDebugMusicName(store *masterdata.Store, musicID int) string {
	if store != nil {
		if music := store.GetMusic(musicID); music != nil && strings.TrimSpace(music.Title) != "" {
			return fmt.Sprintf("#%d %s", musicID, music.Title)
		}
	}
	return fmt.Sprintf("歌曲 #%d", musicID)
}

func suiteDebugCharacterName(characterID int) string {
	names := map[int]string{
		1: "一歌", 2: "咲希", 3: "穗波", 4: "志步", 5: "实乃理", 6: "遥", 7: "爱莉", 8: "雫",
		9: "心羽", 10: "杏", 11: "彰人", 12: "冬弥", 13: "司", 14: "笑梦", 15: "宁宁", 16: "类",
		17: "奏", 18: "真冬", 19: "绘名", 20: "瑞希", 21: "初音未来", 22: "镜音铃", 23: "镜音连", 24: "巡音流歌", 25: "MEIKO", 26: "KAITO",
	}
	if name := names[characterID]; name != "" {
		return name
	}
	if characterID > 0 {
		return fmt.Sprintf("角色 %d", characterID)
	}
	return "角色"
}

func suiteDebugDifficultyColor(diff string) string {
	switch strings.ToLower(diff) {
	case "easy":
		return "#33ccbb"
	case "normal":
		return "#88dd44"
	case "hard":
		return "#ffb000"
	case "expert":
		return "#ff6699"
	case "master":
		return "#a863e8"
	case "append":
		return "#172033"
	default:
		return "#33ccbb"
	}
}

func formatDebugRank(rank int) string {
	if rank <= 0 {
		return ""
	}
	return fmt.Sprintf("Rank %d", rank)
}

func clampDebugLimit(limit int, total int) int {
	if limit <= 0 || limit > total {
		return total
	}
	return limit
}

func minDebug(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxDebug(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func formatDebugInt(value int) string     { return fmt.Sprintf("%d", value) }
func formatDebugInt64(value int64) string { return fmt.Sprintf("%d", value) }

func firstNonEmptyDebug(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

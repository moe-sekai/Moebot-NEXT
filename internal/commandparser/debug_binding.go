package commandparser

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/cardquery"
	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/renderer"
	"moebot-next/internal/servers"
	"moebot-next/internal/suite"
)

const debugDefaultLimit = 10

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

	payload, selected, rows, err := buildSuiteDebugPayloadForDefinition(def, runtime, gameID, argument)
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

func buildSuiteDebugPayloadForDefinition(def Definition, runtime *servers.Runtime, gameID string, argument string) (any, *EntityResult, []EntityResult, error) {
	switch def.ID {
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
	case "gacha-history":
		var profile suiteDebugGachaHistoryProfile
		if err := runtime.Suite.GetUserData(gameID, "", []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserGachas}, &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "抽卡记录", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		sectionRows, stats := suiteDebugRowsFromGachaHistory(profile, runtime.Store, debugDefaultLimit)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), stats...)
		payload.Sections = []renderer.SuiteSectionPayload{{Title: "卡池抽卡记录", Rows: sectionRows}}
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
		payload.Sections = []renderer.SuiteSectionPayload{{Title: "羁绊 TOP", Rows: sectionRows}}
		return payload, suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_panel"), suiteDebugRowsFromSections(payload.Sections), nil
	case "music-progress":
		var profile suiteDebugMusicProgressProfile
		if err := runtime.Suite.GetUserData(gameID, "", []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserMusicResults}, &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "打歌进度", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		sectionRows, stats := suiteDebugRowsFromMusicProgress(profile)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), stats...)
		payload.Sections = []renderer.SuiteSectionPayload{{Title: "难度进度", Rows: sectionRows}}
		return payload, suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_panel"), suiteDebugRowsFromSections(payload.Sections), nil
	case "material-info":
		var profile suiteDebugMaterialProfile
		if err := runtime.Suite.GetUserData(gameID, "", []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserMaterials}, &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "材料信息", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		sectionRows, stats := suiteDebugRowsFromMaterials(profile, 20)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), stats...)
		payload.Sections = []renderer.SuiteSectionPayload{{Title: "持有材料", Rows: sectionRows}}
		return payload, suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_panel"), suiteDebugRowsFromSections(payload.Sections), nil
	case "challenge-info":
		var profile suiteDebugChallengeProfile
		if err := runtime.Suite.GetUserData(gameID, "", suite.Fields(suite.FieldUserChallengeLiveSoloResults, suite.FieldUserChallengeLiveSoloStages, suite.FieldUserChallengeLiveSoloHighScoreRewards), &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "挑战信息", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		sectionRows, stats := suiteDebugRowsFromChallenge(profile, 26)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), stats...)
		payload.Sections = []renderer.SuiteSectionPayload{{Title: "每日挑战 Live", Rows: sectionRows}}
		return payload, suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_panel"), suiteDebugRowsFromSections(payload.Sections), nil
	case "event-record":
		var profile suiteDebugEventRecordProfile
		if err := runtime.Suite.GetUserData(gameID, "", suite.Fields(suite.FieldUserEvents, suite.FieldUserWorldBlooms), &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "活动记录", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		sections, stats := suiteDebugRowsFromEventRecord(profile, runtime.Store, debugDefaultLimit)
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
		sectionRows, stats := suiteDebugRowsFromLeaderCount(profile, 26)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), stats...)
		payload.Sections = []renderer.SuiteSectionPayload{{Title: "角色队长次数", Rows: sectionRows}}
		return payload, suiteDebugSelected(def, runtime, profile.UserGamedata, "suite_panel"), suiteDebugRowsFromSections(payload.Sections), nil
	case "music-reward":
		var profile suiteDebugMusicRewardProfile
		if err := runtime.Suite.GetUserData(gameID, "", []string{suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserMusicAchievements}, &profile); err != nil {
			return nil, nil, nil, err
		}
		payload := newSuiteDebugPanel(runtime, "歌曲奖励", profile.commonSuiteProfile())
		payload.Subtitle = suiteDebugPanelSubtitle(profile.BaseProfile)
		sectionRows, stats := suiteDebugRowsFromMusicReward(profile, runtime.Store, debugDefaultLimit)
		payload.Stats = append(suiteDebugBasicStats(profile.commonSuiteProfile()), stats...)
		payload.Sections = []renderer.SuiteSectionPayload{{Title: "歌曲奖励达成", Rows: sectionRows}}
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
	UserGamedata     suite.UserGamedata      `json:"userGamedata"`
	UserDecks        []suite.UserDeck        `json:"userDecks"`
	UserCards        []suite.UserCard        `json:"userCards"`
	UserMusicResults []suiteDebugMusicResult `json:"userMusicResults"`
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
	CharacterID int `json:"characterId"`
	RewardID    int `json:"challengeLiveHighScoreRewardId"`
}

type suiteDebugChallengeRow struct {
	CharacterID int
	HighScore   int
	Rank        int
	RewardCount int
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
	UserGamedata suite.UserGamedata           `json:"userGamedata"`
	UserDecks    []suite.UserDeck             `json:"userDecks"`
	UserCards    []suite.UserCard             `json:"userCards"`
	Missions     []suiteDebugCharacterMission `json:"userCharacterMissionV2s"`
}

type suiteDebugCharacterMission struct {
	CharacterID          int    `json:"characterId"`
	CharacterMissionType string `json:"characterMissionType"`
	Progress             int    `json:"progress"`
}

type suiteDebugLeaderCountRow struct {
	CharacterID int
	PlayLive    int
	PlayLiveEx  int
}

type suiteDebugMusicRewardProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata           `json:"userGamedata"`
	UserDecks    []suite.UserDeck             `json:"userDecks"`
	UserCards    []suite.UserCard             `json:"userCards"`
	Achievements []suiteDebugMusicAchievement `json:"userMusicAchievements"`
}

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

func (p suiteDebugMusicRewardProfile) commonSuiteProfile() renderer.SuiteCommonProfile {
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
		rows = append(rows, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: suiteDebugCharacterName(cid1) + " × " + suiteDebugCharacterName(cid2), Value: fmt.Sprintf("Lv.%d", bond.Rank), Meta: fmt.Sprintf("EXP %d", bond.Exp)})
	}
	return rows, []renderer.SuiteStatPayload{{Label: "羁绊组数", Value: formatDebugInt(len(bonds))}, {Label: "最高羁绊", Value: fmt.Sprintf("Lv.%d", maxRank)}}
}

func suiteDebugRowsFromMusicProgress(profile suiteDebugMusicProgressProfile) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	counts := suiteDebugMusicProgressCounts(profile)
	rows := make([]renderer.SuiteSectionRowPayload, 0, len(counts))
	totalPlayed, totalClear, totalFC, totalAP := 0, 0, 0, 0
	for _, diff := range []string{"easy", "normal", "hard", "expert", "master", "append"} {
		count := counts[diff]
		if count == nil {
			continue
		}
		totalPlayed += count.Played
		totalClear += count.Clear
		totalFC += count.FullCombo
		totalAP += count.AllPerfect
		rows = append(rows, renderer.SuiteSectionRowPayload{Label: strings.ToUpper(diff), Value: fmt.Sprintf("Clear %d / FC %d / AP %d", count.Clear, count.FullCombo, count.AllPerfect), Meta: fmt.Sprintf("游玩 %d", count.Played), Color: suiteDebugDifficultyColor(diff)})
	}
	return rows, []renderer.SuiteStatPayload{{Label: "游玩", Value: formatDebugInt(totalPlayed)}, {Label: "Clear", Value: formatDebugInt(totalClear)}, {Label: "FC", Value: formatDebugInt(totalFC)}, {Label: "AP", Value: formatDebugInt(totalAP)}}
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

func suiteDebugRowsFromChallenge(profile suiteDebugChallengeProfile, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	rowsByCharacter := map[int]*suiteDebugChallengeRow{}
	for _, result := range profile.Results {
		row := suiteDebugChallengeRowFor(rowsByCharacter, result.CharacterID)
		row.HighScore = maxDebug(row.HighScore, result.HighScore)
	}
	for _, stage := range profile.Stages {
		row := suiteDebugChallengeRowFor(rowsByCharacter, stage.CharacterID)
		row.Rank = maxDebug(row.Rank, stage.Rank)
	}
	for _, reward := range profile.Rewards {
		row := suiteDebugChallengeRowFor(rowsByCharacter, reward.CharacterID)
		row.RewardCount++
	}
	rows := make([]suiteDebugChallengeRow, 0, len(rowsByCharacter))
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
	limit = clampDebugLimit(limit, len(rows))
	out := make([]renderer.SuiteSectionRowPayload, 0, limit)
	maxScore := 0
	for i, row := range rows {
		if row.HighScore > maxScore {
			maxScore = row.HighScore
		}
		if i >= limit {
			continue
		}
		out = append(out, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: suiteDebugCharacterName(row.CharacterID), Value: formatDebugInt(row.HighScore), Meta: fmt.Sprintf("Lv.%d · 奖励 %d", row.Rank, row.RewardCount)})
	}
	return out, []renderer.SuiteStatPayload{{Label: "角色数", Value: formatDebugInt(len(rows))}, {Label: "最高分", Value: formatDebugInt(maxScore)}}
}

func suiteDebugRowsFromEventRecord(profile suiteDebugEventRecordProfile, store *masterdata.Store, limit int) ([]renderer.SuiteSectionPayload, []renderer.SuiteStatPayload) {
	events := append([]suiteDebugEventRecord(nil), profile.UserEvents...)
	sort.SliceStable(events, func(i, j int) bool { return events[i].EventPoint > events[j].EventPoint })
	blooms := append([]suiteDebugWorldBloomRecord(nil), profile.UserWorldBlooms...)
	sort.SliceStable(blooms, func(i, j int) bool {
		return suiteDebugWorldBloomPoint(blooms[i]) > suiteDebugWorldBloomPoint(blooms[j])
	})
	if limit <= 0 {
		limit = maxDebug(len(events), len(blooms))
	}
	sections := make([]renderer.SuiteSectionPayload, 0, 2)
	if len(events) > 0 {
		rows := make([]renderer.SuiteSectionRowPayload, 0, minDebug(limit, len(events)))
		for i := 0; i < minDebug(limit, len(events)); i++ {
			event := events[i]
			rows = append(rows, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: suiteDebugEventName(store, event.EventID), Value: fmt.Sprintf("%dpt", event.EventPoint), Meta: formatDebugRank(event.Rank)})
		}
		sections = append(sections, renderer.SuiteSectionPayload{Title: "活动PT", Rows: rows})
	}
	if len(blooms) > 0 {
		rows := make([]renderer.SuiteSectionRowPayload, 0, minDebug(limit, len(blooms)))
		for i := 0; i < minDebug(limit, len(blooms)); i++ {
			bloom := blooms[i]
			rows = append(rows, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: suiteDebugEventName(store, bloom.EventID), Value: fmt.Sprintf("%dpt", suiteDebugWorldBloomPoint(bloom)), Meta: suiteDebugCharacterName(bloom.GameCharacterID) + " " + formatDebugRank(bloom.WorldBloomChapterRank)})
		}
		sections = append(sections, renderer.SuiteSectionPayload{Title: "WL章节", Rows: rows})
	}
	return sections, []renderer.SuiteStatPayload{{Label: "活动记录", Value: formatDebugInt(len(events))}, {Label: "WL记录", Value: formatDebugInt(len(blooms))}}
}

func suiteDebugRowsFromLeaderCount(profile suiteDebugLeaderCountProfile, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
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
	rows := make([]suiteDebugLeaderCountRow, 0, len(rowsByCharacter))
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
	limit = clampDebugLimit(limit, len(rows))
	out := make([]renderer.SuiteSectionRowPayload, 0, limit)
	total := 0
	for i, row := range rows {
		total += row.PlayLive
		if i >= limit {
			continue
		}
		out = append(out, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: suiteDebugCharacterName(row.CharacterID), Value: formatDebugInt(row.PlayLive), Meta: fmt.Sprintf("EX %d", row.PlayLiveEx)})
	}
	return out, []renderer.SuiteStatPayload{{Label: "总队长次数", Value: formatDebugInt(total)}, {Label: "角色数", Value: formatDebugInt(len(rows))}}
}

func suiteDebugRowsFromMusicReward(profile suiteDebugMusicRewardProfile, store *masterdata.Store, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	counts := map[int]int{}
	for _, achievement := range profile.Achievements {
		if achievement.MusicID <= 0 {
			continue
		}
		counts[achievement.MusicID]++
	}
	rows := make([]suiteDebugMusicRewardRow, 0, len(counts))
	total := 0
	for mid, count := range counts {
		rows = append(rows, suiteDebugMusicRewardRow{MusicID: mid, Count: count})
		total += count
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
		out = append(out, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: suiteDebugMusicName(store, row.MusicID), Value: formatDebugInt(row.Count), Meta: "已达成奖励"})
	}
	return out, []renderer.SuiteStatPayload{{Label: "已达成奖励", Value: formatDebugInt(total)}, {Label: "涉及歌曲", Value: formatDebugInt(len(rows))}}
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
	for _, result := range profile.UserMusicResults {
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

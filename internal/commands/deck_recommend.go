package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"moebot-next/internal/assets"
	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/deckrecommenddata"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/musicsearch"
	"moebot-next/internal/renderer"
	"moebot-next/internal/servers"
	"moebot-next/internal/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

const musicMetaURL = "https://moe.exmeaning.com/data/music_meta/music_metas.json"
const musicMetaCacheTTL = 6 * time.Hour
const deckRecommendDefaultMasterCacheTTL = time.Hour

type deckRecommendUserCardEntry struct {
	CardID int `json:"cardId"`
}

type deckRecommendUserHonorEntry struct {
	HonorID int `json:"honorId"`
}

const (
	deckRecommendDefaultMusicID    = 74
	deckRecommendDefaultDifficulty = "expert"
)

var musicMetaCache struct {
	sync.Mutex
	data      []map[string]any
	updatedAt time.Time
}

type deckMasterCacheEntry struct {
	data      []any
	updatedAt time.Time
}

var deckMasterDataCache struct {
	sync.Mutex
	items map[string]deckMasterCacheEntry
}

var deckRecommendCommonSuiteFields = []string{
	suite.FieldUploadTime,
	suite.FieldUserGamedata,
	suite.FieldUserDecks,
	suite.FieldUserCards,
	suite.FieldUserBonds,
	suite.FieldUserMaterials,
	suite.FieldUserAreas,
	suite.FieldUserCharacters,
	suite.FieldUserChallengeLiveSoloDecks,
	suite.FieldUserChallengeLiveSoloStages,
	suite.FieldUserChallengeLiveSoloResults,
	suite.FieldUserChallengeLiveSoloHighScoreRewards,
	suite.FieldUserCharacterMissionV2s,
	suite.FieldUserCharacterMissionV2Statuses,
	suite.FieldUserMysekaiFixtureGameCharacterPerformanceBonuses,
	suite.FieldUserMysekaiGates,
	"userMusics",
	"userMusicResults",
	"userMysekaiMaterials",
	"userMysekaiCanvases",
	"userWorldBloomSupportDecks",
	"userHonors",
	"userMysekaiCharacterTalks",
	"userEvents",
	"userWorldBlooms",
	"userMusicAchievements",
	"userPlayerFrames",
}

var deckRecommendMinimalSuiteFields = []string{
	suite.FieldUploadTime,
	suite.FieldUserGamedata,
	suite.FieldUserDecks,
	suite.FieldUserCards,
	suite.FieldUserBonds,
	suite.FieldUserMaterials,
	suite.FieldUserAreas,
	suite.FieldUserCharacters,
	suite.FieldUserChallengeLiveSoloDecks,
	suite.FieldUserChallengeLiveSoloStages,
	suite.FieldUserChallengeLiveSoloResults,
	suite.FieldUserChallengeLiveSoloHighScoreRewards,
	suite.FieldUserCharacterMissionV2s,
	suite.FieldUserCharacterMissionV2Statuses,
	suite.FieldUserMysekaiFixtureGameCharacterPerformanceBonuses,
	suite.FieldUserMysekaiGates,
	"userMusics",
	"userMusicResults",
}

var deckRecommendMasterKeys = []string{
	"areaItemLevels", "cards", "cardMysekaiCanvasBonuses", "cardRarities", "characterRanks", "cardEpisodes",
	"events", "eventCards", "eventRarityBonusRates", "eventDeckBonuses", "gameCharacters", "gameCharacterUnits",
	"honors", "masterLessons", "mysekaiGates", "mysekaiGateLevels", "skills", "eventHonorBonuses",
	"eventCardBonusLimits", "eventSkillScoreUpLimits", "worldBlooms", "worldBloomDifferentAttributeBonuses",
	"worldBloomSupportDeckBonuses", "worldBloomSupportDeckBonusesWL1", "worldBloomSupportDeckBonusesWL2",
	"worldBloomSupportDeckBonusesWL3", "worldBloomSupportDeckUnitEventLimitedBonuses",
}

func RegisterDeckRecommend(deps *Deps) {
	registerDeckRecommendMode(deps, "组卡", "event")
	registerDeckRecommendMode(deps, "最强组卡", "strongest")
	registerDeckRecommendMode(deps, "挑战组卡", "challenge")
	registerDeckRecommendMode(deps, "加成组卡", "bonus")
	registerDeckRecommendMode(deps, "烤森组卡", "mysekai")
}

func registerDeckRecommendMode(deps *Deps, primary string, mode string) {
	for _, cmd := range parserCommands(deps, primary) {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		recordCommand := cmd.Primary
		if recordCommand == "" {
			recordCommand = primary
		}
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
				ctx.SendChain(message.Text("Suite API 未启用，无法读取卡牌数据进行组卡"))
				return
			}
			if deps.Renderer == nil || !deps.Renderer.Health() {
				ctx.SendChain(message.Text("渲染/计算服务暂不可用，请稍后再试"))
				return
			}

			options, music, event, err := parseDeckRecommendArgs(commandArgs(ctx), runtime.Store, runtime.MusicAliases, mode)
			if err != nil {
				ctx.SendChain(message.Text(err.Error()))
				return
			}
			ctx.SendChain(message.Text("正在组卡中，请稍等一下喵~"))

			var userData map[string]any
			if err := loadDeckRecommendUserData(runtime.Suite, user.GameID, &userData); err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("读取 Suite 数据失败：%v", err)))
				return
			}
			masterMap, warnings := buildDeckRecommendMasterData(runtime)
			userData = filterDeckRecommendUserDataWithJPMaster(userData, masterMap, runtime.Store)
			musicMetas, err := fetchMusicMetas()
			if err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取歌曲分数元数据失败：%v", err)))
				return
			}
			profile := suiteProfileFromUserData(userData, user.GameID)
			req := renderer.DeckRecommendCalculateRequest{
				Region: runtime.Region, RegionLabel: runtime.Label, UserData: userData, MasterData: masterMap,
				MusicMetas: musicMetas, Options: options, CardAssets: buildDeckRecommendCardAssets(runtime.Store, runtime.Assets),
				Music: deckMusicPayload(runtime.Store, music, runtime.Assets, options.Difficulty, options.IsPresetDefault), Profile: profile,
			}
			if event != nil {
				req.Event = renderer.BuildEventInfoPayloadWithAssets(runtime.Store, *event, runtime.Assets)
			}
			calc, err := deps.Renderer.CalculateDeckRecommend(req)
			if err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("组卡计算失败：%v", err)))
				return
			}
			calc.Warnings = append(calc.Warnings, warnings...)
			payload := buildDeckRecommendPayload(runtime, mode, calc)
			png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "deck_recommend", Data: payload, Width: 980})
			if err != nil {
				ctx.SendChain(message.Text(formatDeckRecommendText(calc)))
			} else {
				ctx.SendChain(message.ImageBytes(png))
			}
			bot.RecordCommandRegion(deps.DB, recordCommand, runtime.Region, ctx, start)
		})
	}
}

func parseDeckRecommendArgs(raw string, store *masterdata.Store, aliases map[int]assets.MusicAlias, mode string) (renderer.DeckRecommendOptions, *masterdata.MusicInfo, *masterdata.EventInfo, error) {
	args := strings.Fields(strings.TrimSpace(raw))
	mode = normalizeDeckRecommendMode(mode)
	options := renderer.DeckRecommendOptions{Mode: mode, MusicID: deckRecommendDefaultMusicID, Difficulty: deckRecommendDefaultDifficulty, LiveType: defaultDeckLiveType(mode), Algorithm: "ga", Target: defaultDeckTarget(mode), Limit: 3, TimeoutMS: 15000, BestSkillAsLeader: true, CardConfig: defaultDeckCardConfig()}
	var eventID, musicID int
	musicExplicit := false
	bonusTargets := []int{}
	challengeSet := false
	wlChapterNo := 0
	remaining := make([]string, 0, len(args))
	fixedMode := false
	for i, rawToken := range args {
		token := strings.ToLower(strings.TrimSpace(rawToken))
		if token == "" {
			continue
		}
		if fixedMode {
			if err := parseDeckFixedToken(token, &options); err != nil {
				return options, nil, nil, err
			}
			continue
		}
		if strings.HasPrefix(token, "#") {
			fixedMode = true
			for _, part := range strings.Split(strings.TrimPrefix(token, "#"), ",") {
				if err := parseDeckFixedToken(part, &options); err != nil {
					return options, nil, nil, err
				}
			}
			continue
		}
		if mode == "event" || mode == "bonus" || mode == "mysekai" {
			if chapterNo, ok := parseDeckWLChapterToken(token); ok {
				wlChapterNo = chapterNo
				continue
			}
		}
		switch token {
		case "多人", "协力", "multi":
			options.LiveType = "multi"
		case "单人", "solo":
			options.LiveType = "solo"
		case "自动", "auto":
			options.LiveType = "auto"
		case "欢乐", "cheerful":
			options.LiveType = "cheerful"
		case "分数", "score", "pt":
			options.Target = "score"
		case "综合力", "综合", "总合力", "总和", "power":
			options.Target = "power"
		case "实效", "skill", "倍率", "时效":
			options.Target = "skill"
		case "dfs":
			options.Algorithm = "dfs"
		case "ga":
			options.Algorithm = "ga"
		case "all", "全部算法":
			options.Algorithm = "all"
		case "技能吸取最大":
			options.SkillReferenceChooseStrategy = "max"
		case "技能吸取最小":
			options.SkillReferenceChooseStrategy = "min"
		case "技能吸取平均":
			options.SkillReferenceChooseStrategy = "average"
		case "不换队长", "固定队长":
			options.BestSkillAsLeader = false
		case "bfes不变", "bf不变":
			options.KeepAfterTrainingState = true
		case "异队":
			options.FilterOtherUnit = true
		case "终章":
			eventID = 180
		case "easy", "ez", "简单":
			options.Difficulty = "easy"
		case "normal", "nm", "普通":
			options.Difficulty = "normal"
		case "hard", "hd", "困难":
			options.Difficulty = "hard"
		case "expert", "ex", "专家":
			options.Difficulty = "expert"
		case "master", "ma", "mas", "大师":
			options.Difficulty = "master"
		case "append", "apd", "ap", "追加":
			options.Difficulty = "append"
		case "满破", "满突破", "rankmax", "mastermax", "5破", "五破":
			setAllCardConfig(options.CardConfig, func(c renderer.DeckCardConfig) renderer.DeckCardConfig { c.MasterMax = true; return c })
		case "满技能", "满技", "skillmax", "技能满级", "slv4":
			setAllCardConfig(options.CardConfig, func(c renderer.DeckCardConfig) renderer.DeckCardConfig { c.SkillMax = true; return c })
		case "已读", "剧情已读", "满剧情", "前后篇已读", "前后篇":
			setAllCardConfig(options.CardConfig, func(c renderer.DeckCardConfig) renderer.DeckCardConfig { c.EpisodeRead = true; return c })
		case "四星满破":
			c := options.CardConfig["rarity_4"]
			c.MasterMax = true
			options.CardConfig["rarity_4"] = c
		case "四星满技能":
			c := options.CardConfig["rarity_4"]
			c.SkillMax = true
			options.CardConfig["rarity_4"] = c
		case "生日满技能":
			c := options.CardConfig["rarity_birthday"]
			c.SkillMax = true
			options.CardConfig["rarity_birthday"] = c
		default:
			if ok, id := parsePrefixedID(token, "event", "活动"); ok {
				eventID = id
				continue
			}
			if ok, id := parsePrefixedID(token, "music", "曲", "歌"); ok {
				musicID = id
				musicExplicit = true
				continue
			}
			if limit, ok := parseLimitToken(token); ok {
				options.Limit = limit
				continue
			}
			if timeout, ok := parseTimeoutToken(token); ok {
				options.TimeoutMS = timeout
				continue
			}
			if id, err := strconv.Atoi(token); err == nil {
				if mode == "bonus" && id >= 0 && id <= 700 {
					bonusTargets = append(bonusTargets, id)
				} else if mode == "challenge" && id >= 1 && id <= 26 && !challengeSet {
					options.ChallengeCharacterID = id
					challengeSet = true
				} else if shouldPreferMusicID(args, i) && store.GetMusic(id) != nil {
					musicID = id
					musicExplicit = true
				} else if store.GetEvent(id) != nil {
					eventID = id
				} else if store.GetMusic(id) != nil {
					musicID = id
					musicExplicit = true
				}
				continue
			}
			if mode == "challenge" && !challengeSet {
				if characterID, ok := deckCharacterAlias(token); ok {
					options.ChallengeCharacterID = characterID
					challengeSet = true
					continue
				}
			}
			remaining = append(remaining, rawToken)
		}
	}
	if len(bonusTargets) > 0 {
		options.TargetBonusList = bonusTargets
		options.TargetBonus = bonusTargets[0]
	}
	if mode == "challenge" && options.ChallengeCharacterID == 0 {
		return options, nil, nil, fmt.Errorf("请输入挑战角色，例如 /挑战组卡 miku")
	}
	if eventID == 0 && mode != "strongest" && mode != "challenge" {
		eventID = currentEventID(store)
	}
	var event *masterdata.EventInfo
	if eventID != 0 {
		event = store.GetEvent(eventID)
		if event == nil {
			return options, nil, nil, fmt.Errorf("找不到活动：%d", eventID)
		}
	}
	if (mode == "event" || mode == "bonus" || mode == "mysekai") && event != nil {
		var err error
		remaining, err = applyDeckWorldBloomChapter(store, event, wlChapterNo, remaining, &options)
		if err != nil {
			return options, nil, nil, err
		}
	}
	if mode == "event" || mode == "bonus" || mode == "mysekai" {
		remaining = applyDeckSupportCharacterAlias(remaining, &options)
	} else if mode != "challenge" {
		remaining = applyDeckFixedCharacterAliases(remaining, &options)
	}
	if mode == "bonus" && len(options.TargetBonusList) == 0 {
		return options, nil, nil, fmt.Errorf("请输入目标活动加成，例如 /加成组卡 300")
	}
	if mode == "mysekai" && event == nil {
		return options, nil, nil, fmt.Errorf("烤森组卡需要指定活动，例如 /烤森组卡 event180")
	}
	if musicID == 0 && mode != "bonus" && mode != "mysekai" && len(remaining) > 0 {
		musicID = searchMusicID(store, aliases, strings.Join(remaining, " "))
		if musicID == 0 {
			return options, nil, nil, fmt.Errorf("找不到曲目关键词：%s", strings.Join(remaining, " "))
		}
		musicExplicit = true
	}
	if musicID != 0 {
		options.MusicID = musicID
	}
	if event != nil {
		options.EventID = event.ID
	}
	options.IsPresetDefault = !musicExplicit && options.MusicID == deckRecommendDefaultMusicID
	music := store.GetMusic(options.MusicID)
	if music == nil && options.MusicID == deckRecommendDefaultMusicID {
		music = &masterdata.MusicInfo{ID: deckRecommendDefaultMusicID, Title: "默认曲目"}
	}
	if options.MusicID == 10000 {
		music = &masterdata.MusicInfo{ID: 10000, Title: "おまかせ"}
	}
	if music == nil {
		return options, nil, nil, fmt.Errorf("找不到曲目：%d", options.MusicID)
	}
	return options, music, event, nil
}

func normalizeDeckRecommendMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "strongest", "challenge", "bonus", "event", "mysekai":
		return strings.ToLower(strings.TrimSpace(mode))
	default:
		return "event"
	}
}

func defaultDeckLiveType(mode string) string {
	if mode == "challenge" {
		return "challenge"
	}
	return "multi"
}

func defaultDeckTarget(mode string) string {
	if mode == "strongest" {
		return "power"
	}
	return "score"
}

func deckRecommendTitle(mode string) string {
	switch normalizeDeckRecommendMode(mode) {
	case "strongest":
		return "最强组卡推荐"
	case "challenge":
		return "挑战组卡推荐"
	case "bonus":
		return "加成/控分组卡推荐"
	case "mysekai":
		return "烤森组卡推荐"
	default:
		return "活动组卡推荐"
	}
}

func deckRecommendModeFields(mode string) []string {
	mode = normalizeDeckRecommendMode(mode)
	if mode == "" {
		return append([]string(nil), deckRecommendCommonSuiteFields...)
	}
	return append([]string(nil), deckRecommendCommonSuiteFields...)
}

func loadDeckRecommendUserData(client *suite.Client, gameID string, out *map[string]any) error {
	if client == nil {
		return fmt.Errorf("suite client is nil")
	}
	if out == nil {
		return fmt.Errorf("suite response output is nil")
	}
	var userData map[string]any
	if err := client.GetUserData(gameID, "", deckRecommendCommonSuiteFields, &userData); err != nil {
		return err
	}
	*out = normalizeDeckRecommendUserData(userData)
	return nil
}

func normalizeDeckRecommendUserData(userData map[string]any) map[string]any {
	normalized := make(map[string]any, len(deckRecommendCommonSuiteFields))
	for key, value := range userData {
		normalized[key] = value
	}
	for _, key := range deckRecommendCommonSuiteFields {
		if _, ok := normalized[key]; ok {
			continue
		}
		normalized[key] = deckRecommendDefaultUserDataValue(key)
	}
	return normalized
}

func deckRecommendDefaultUserDataValue(key string) any {
	switch key {
	case suite.FieldUserGamedata, suite.FieldUploadTime:
		return nil
	default:
		return []any{}
	}
}

func filterDeckRecommendUserData(userData map[string]any, store *masterdata.Store) map[string]any {
	if userData == nil {
		return nil
	}
	filtered := make(map[string]any, len(userData))
	for key, value := range userData {
		filtered[key] = value
	}
	if store == nil {
		return filtered
	}
	if rawCards, ok := filtered[suite.FieldUserCards]; ok {
		filtered[suite.FieldUserCards] = filterDeckRecommendUserCards(rawCards, store)
	}
	if rawHonors, ok := filtered["userHonors"]; ok {
		filtered["userHonors"] = filterDeckRecommendUserHonors(rawHonors, store)
	}
	return filtered
}

func filterDeckRecommendUserDataWithJPMaster(userData map[string]any, masterMap map[string]any, store *masterdata.Store) map[string]any {
	if userData == nil {
		return nil
	}
	filtered := make(map[string]any, len(userData))
	for key, value := range userData {
		filtered[key] = value
	}
	if rawCards, ok := filtered[suite.FieldUserCards]; ok {
		filtered[suite.FieldUserCards] = filterDeckRecommendUserCardsFromJPMaster(rawCards, masterMap)
	}
	if rawHonors, ok := filtered["userHonors"]; ok {
		filtered["userHonors"] = filterDeckRecommendUserHonorsFromJPMaster(rawHonors, masterMap, store)
	}
	if rawAreas, ok := filtered[suite.FieldUserAreas]; ok {
		filtered[suite.FieldUserAreas] = filterDeckRecommendUserAreasFromJPMaster(rawAreas, masterMap)
	}
	if rawGates, ok := filtered[suite.FieldUserMysekaiGates]; ok {
		filtered[suite.FieldUserMysekaiGates] = filterDeckRecommendUserMysekaiGatesFromJPMaster(rawGates, masterMap)
	}
	if rawCharacters, ok := filtered[suite.FieldUserCharacters]; ok {
		filtered[suite.FieldUserCharacters] = filterDeckRecommendUserCharactersFromJPMaster(rawCharacters, masterMap)
	}
	return filtered
}

func filterDeckRecommendUserCardsFromJPMaster(raw any, masterMap map[string]any) any {
	items, ok := raw.([]any)
	if !ok {
		return raw
	}
	jpCards, ok := masterMap["cards"]
	if !ok {
		return raw
	}
	cardList, ok := jpCards.([]any)
	if !ok {
		return raw
	}
	valid := make(map[int]struct{}, len(cardList))
	for _, card := range cardList {
		entry, ok := card.(map[string]any)
		if !ok {
			continue
		}
		cardID := intValueFromAny(entry["id"])
		if cardID != 0 {
			valid[cardID] = struct{}{}
		}
	}
	if len(valid) == 0 {
		return raw
	}
	filtered := make([]any, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		cardID := intValueFromAny(entry["cardId"])
		if _, exists := valid[cardID]; exists {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterDeckRecommendUserCards(raw any, store *masterdata.Store) any {
	items, ok := raw.([]any)
	if !ok {
		return raw
	}
	masterCards := store.AllCards()
	valid := make(map[int]struct{}, len(masterCards))
	for _, card := range masterCards {
		valid[card.ID] = struct{}{}
	}
	filtered := make([]any, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		cardID := intValueFromAny(entry["cardId"])
		if _, exists := valid[cardID]; exists {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterDeckRecommendUserHonors(raw any, store *masterdata.Store) any {
	if store == nil {
		return raw
	}
	masterHonors := store.AllHonors()
	return filterDeckRecommendUserHonorsByValidLevels(raw, deckRecommendValidHonorLevelsFromMaster(masterHonors))
}

func filterDeckRecommendUserHonorsFromJPMaster(raw any, masterMap map[string]any, store *masterdata.Store) any {
	if levels := deckRecommendValidHonorLevelsFromMasterMap(masterMap); len(levels) > 0 {
		return filterDeckRecommendUserHonorsByValidLevels(raw, levels)
	}
	return filterDeckRecommendUserHonors(raw, store)
}

func deckRecommendValidHonorLevelsFromMaster(masterHonors []masterdata.HonorInfo) map[int]map[int]struct{} {
	valid := make(map[int]map[int]struct{}, len(masterHonors))
	for _, honor := range masterHonors {
		levels := make(map[int]struct{}, len(honor.Levels))
		for _, level := range honor.Levels {
			levels[level.Level] = struct{}{}
		}
		valid[honor.ID] = levels
	}
	return valid
}

func deckRecommendValidHonorLevelsFromMasterMap(masterMap map[string]any) map[int]map[int]struct{} {
	jpHonors, ok := masterMap["honors"]
	if !ok {
		return nil
	}
	honorList, ok := jpHonors.([]any)
	if !ok {
		return nil
	}
	valid := make(map[int]map[int]struct{}, len(honorList))
	for _, honor := range honorList {
		entry, ok := honor.(map[string]any)
		if !ok {
			continue
		}
		honorID := intValueFromAny(entry["id"])
		if honorID == 0 {
			continue
		}
		levels := map[int]struct{}{}
		if rawLevels, ok := entry["levels"].([]any); ok {
			for _, rawLevel := range rawLevels {
				levelEntry, ok := rawLevel.(map[string]any)
				if !ok {
					continue
				}
				level := intValueFromAny(levelEntry["level"])
				if level != 0 {
					levels[level] = struct{}{}
				}
			}
		}
		valid[honorID] = levels
	}
	return valid
}

func filterDeckRecommendUserHonorsByValidLevels(raw any, valid map[int]map[int]struct{}) any {
	items, ok := raw.([]any)
	if !ok || len(valid) == 0 {
		return raw
	}
	filtered := make([]any, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		honorID := intValueFromAny(entry["honorId"])
		level := intValueFromAny(entry["level"])
		levels, exists := valid[honorID]
		if !exists {
			continue
		}
		if len(levels) > 0 {
			if _, ok := levels[level]; !ok {
				continue
			}
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func filterDeckRecommendUserAreasFromJPMaster(raw any, masterMap map[string]any) any {
	areas, ok := raw.([]any)
	if !ok {
		return raw
	}
	valid := deckRecommendValidMasterPairs(masterMap, "areaItemLevels", "areaItemId", "level")
	if len(valid) == 0 {
		return raw
	}
	filteredAreas := make([]any, 0, len(areas))
	for _, area := range areas {
		entry, ok := area.(map[string]any)
		if !ok {
			continue
		}
		areaItems, ok := entry["areaItems"].([]any)
		if !ok {
			filteredAreas = append(filteredAreas, area)
			continue
		}
		filteredItems := make([]any, 0, len(areaItems))
		for _, item := range areaItems {
			itemEntry, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if deckRecommendPairExists(valid, intValueFromAny(itemEntry["areaItemId"]), intValueFromAny(itemEntry["level"])) {
				filteredItems = append(filteredItems, item)
			}
		}
		cloned := cloneDeckRecommendMap(entry)
		cloned["areaItems"] = filteredItems
		filteredAreas = append(filteredAreas, cloned)
	}
	return filteredAreas
}

func filterDeckRecommendUserMysekaiGatesFromJPMaster(raw any, masterMap map[string]any) any {
	items, ok := raw.([]any)
	if !ok {
		return raw
	}
	valid := deckRecommendValidMasterPairs(masterMap, "mysekaiGateLevels", "mysekaiGateId", "level")
	if len(valid) == 0 {
		return raw
	}
	filtered := make([]any, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if deckRecommendPairExists(valid, intValueFromAny(entry["mysekaiGateId"]), intValueFromAny(entry["mysekaiGateLevel"])) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterDeckRecommendUserCharactersFromJPMaster(raw any, masterMap map[string]any) any {
	items, ok := raw.([]any)
	if !ok {
		return raw
	}
	ranksByCharacter := deckRecommendValidCharacterRanks(masterMap)
	if len(ranksByCharacter) == 0 {
		return raw
	}
	filtered := make([]any, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		characterID := intValueFromAny(entry["characterId"])
		characterRank := intValueFromAny(entry["characterRank"])
		ranks, exists := ranksByCharacter[characterID]
		if !exists || len(ranks) == 0 {
			continue
		}
		if _, ok := ranks[characterRank]; ok {
			filtered = append(filtered, item)
			continue
		}
		clampedRank := deckRecommendBestCharacterRank(ranks, characterRank)
		if clampedRank == 0 {
			continue
		}
		cloned := cloneDeckRecommendMap(entry)
		cloned["characterRank"] = clampedRank
		filtered = append(filtered, cloned)
	}
	return filtered
}

func deckRecommendValidMasterPairs(masterMap map[string]any, key string, firstField string, secondField string) map[int]map[int]struct{} {
	raw, ok := masterMap[key]
	if !ok {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	valid := make(map[int]map[int]struct{})
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		first := intValueFromAny(entry[firstField])
		second := intValueFromAny(entry[secondField])
		if first == 0 || second == 0 {
			continue
		}
		if valid[first] == nil {
			valid[first] = map[int]struct{}{}
		}
		valid[first][second] = struct{}{}
	}
	return valid
}

func deckRecommendPairExists(valid map[int]map[int]struct{}, first int, second int) bool {
	seconds, ok := valid[first]
	if !ok {
		return false
	}
	_, ok = seconds[second]
	return ok
}

func deckRecommendValidCharacterRanks(masterMap map[string]any) map[int]map[int]struct{} {
	return deckRecommendValidMasterPairs(masterMap, "characterRanks", "characterId", "characterRank")
}

func deckRecommendBestCharacterRank(ranks map[int]struct{}, current int) int {
	bestBelow := 0
	minRank := 0
	for rank := range ranks {
		if minRank == 0 || rank < minRank {
			minRank = rank
		}
		if rank <= current && rank > bestBelow {
			bestBelow = rank
		}
	}
	if bestBelow > 0 {
		return bestBelow
	}
	return minRank
}

func cloneDeckRecommendMap(entry map[string]any) map[string]any {
	cloned := make(map[string]any, len(entry))
	for key, value := range entry {
		cloned[key] = value
	}
	return cloned
}

func intValueFromAny(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i)
		}
		if f, err := strconv.ParseFloat(v.String(), 64); err == nil {
			return int(f)
		}
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return i
		}
	}
	return 0
}

func buildDeckRecommendPayload(runtime *servers.Runtime, mode string, calc *renderer.DeckRecommendCalculateResponse) map[string]any {
	if calc == nil {
		return nil
	}
	return map[string]any{"title": deckRecommendTitle(mode), "regionLabel": runtime.Label, "profile": calc.Profile, "event": calc.Event, "music": calc.Music, "options": calc.Options, "algorithm": calc.Algorithm, "costMs": calc.CostMS, "warnings": calc.Warnings, "decks": calc.Decks, "assetSource": assetSourceForRuntime(runtime.Assets)}
}

func parseDeckFixedToken(token string, options *renderer.DeckRecommendOptions) error {
	token = strings.Trim(strings.TrimSpace(token), ",，")
	if token == "" {
		return nil
	}
	if id, err := strconv.Atoi(token); err == nil {
		options.FixedCards = append(options.FixedCards, id)
		return nil
	}
	if characterID, ok := deckCharacterAlias(token); ok {
		options.FixedCharacters = append(options.FixedCharacters, characterID)
		return nil
	}
	return fmt.Errorf("无法识别固定卡牌/角色：%s", token)
}

func deckCharacterAlias(token string) (int, bool) {
	query := assets.NormalizeAlias(token)
	if query == "" {
		return 0, false
	}
	for _, entry := range assets.CharacterAliasEntries() {
		if entry.Normalized == query {
			return entry.CharacterID, true
		}
	}
	return 0, false
}

func extractDeckCharacterAliasFromText(text string) (int, string, bool) {
	normalized := assets.NormalizeAlias(text)
	if normalized == "" {
		return 0, "", false
	}
	for _, entry := range assets.CharacterAliasEntries() {
		if entry.Normalized == "" {
			continue
		}
		if strings.Contains(normalized, entry.Normalized) {
			return entry.CharacterID, entry.Alias, true
		}
	}
	return 0, "", false
}

func applyDeckFixedCharacterAliases(remaining []string, options *renderer.DeckRecommendOptions) []string {
	if options == nil || len(remaining) == 0 {
		return remaining
	}
	out := make([]string, 0, len(remaining))
	seen := make(map[int]bool, len(options.FixedCharacters))
	for _, id := range options.FixedCharacters {
		seen[id] = true
	}
	for _, token := range remaining {
		if characterID, ok := deckCharacterAlias(token); ok {
			if !seen[characterID] {
				options.FixedCharacters = append(options.FixedCharacters, characterID)
				seen[characterID] = true
			}
			continue
		}
		out = append(out, token)
	}
	return out
}

func applyDeckSupportCharacterAlias(remaining []string, options *renderer.DeckRecommendOptions) []string {
	if options == nil || len(remaining) == 0 || options.SupportCharacterID > 0 {
		return remaining
	}
	out := make([]string, 0, len(remaining))
	consumed := false
	for _, token := range remaining {
		if !consumed {
			if characterID, ok := deckCharacterAlias(token); ok {
				options.SupportCharacterID = characterID
				consumed = true
				continue
			}
		}
		out = append(out, token)
	}
	return out
}

func parseDeckWLChapterToken(token string) (int, bool) {
	clean := strings.TrimSpace(strings.ToLower(token))
	clean = strings.TrimPrefix(clean, "wl")
	clean = strings.TrimPrefix(clean, "章节")
	clean = strings.TrimPrefix(clean, "第")
	clean = strings.TrimSuffix(clean, "章")
	if clean == "" {
		return 0, false
	}
	value, err := strconv.Atoi(clean)
	if err != nil || value <= 0 || value > 99 {
		return 0, false
	}
	return value, true
}

func parsePrefixedID(token string, prefixes ...string) (bool, int) {
	for _, prefix := range prefixes {
		if strings.HasPrefix(token, prefix) {
			id, err := strconv.Atoi(strings.TrimPrefix(token, prefix))
			return err == nil, id
		}
	}
	return false, 0
}
func applyDeckWorldBloomChapter(store *masterdata.Store, event *masterdata.EventInfo, chapterNo int, remaining []string, options *renderer.DeckRecommendOptions) ([]string, error) {
	if event == nil || options == nil {
		return remaining, nil
	}
	if event.EventType != "world_bloom" {
		if chapterNo > 0 {
			return remaining, fmt.Errorf("活动 #%d 不是 WL 活动，无法指定章节 wl%d", event.ID, chapterNo)
		}
		return remaining, nil
	}
	chapters := store.GetWorldBlooms(event.ID)
	if len(chapters) == 0 {
		return remaining, fmt.Errorf("活动 #%d 缺少 WL 章节数据，无法选择章节角色", event.ID)
	}
	var characterID int
	var consumed string
	for _, token := range remaining {
		if id, _, ok := extractDeckCharacterAliasFromText(token); ok {
			characterID = id
			consumed = token
			break
		}
	}
	chapter, err := resolveDeckWorldBloomChapter(event, chapters, chapterNo, characterID)
	if err != nil {
		return remaining, err
	}
	options.SupportCharacterID = chapter.GameCharacterID
	if consumed == "" {
		return remaining, nil
	}
	out := make([]string, 0, len(remaining)-1)
	removed := false
	for _, token := range remaining {
		if !removed && token == consumed {
			removed = true
			continue
		}
		out = append(out, token)
	}
	return out, nil
}

func resolveDeckWorldBloomChapter(event *masterdata.EventInfo, chapters []masterdata.WorldBloom, chapterNo int, characterID int) (*masterdata.WorldBloom, error) {
	if event == nil {
		return nil, fmt.Errorf("活动不存在")
	}
	sorted := append([]masterdata.WorldBloom(nil), chapters...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].ChapterNo != sorted[j].ChapterNo {
			return sorted[i].ChapterNo < sorted[j].ChapterNo
		}
		return sorted[i].ID < sorted[j].ID
	})
	if chapterNo > 0 {
		for i := range sorted {
			if sorted[i].ChapterNo == chapterNo {
				return &sorted[i], nil
			}
		}
		return nil, fmt.Errorf("活动 #%d 没有 WL 第 %d 章", event.ID, chapterNo)
	}
	if characterID > 0 {
		for i := range sorted {
			if sorted[i].GameCharacterID == characterID {
				return &sorted[i], nil
			}
		}
		return nil, fmt.Errorf("活动 #%d 没有 %s 的 WL 章节", event.ID, characterNameByID(characterID))
	}
	if len(sorted) == 1 {
		return &sorted[0], nil
	}
	now := time.Now().UnixMilli()
	for i := range sorted {
		start := sorted[i].ChapterStartAt
		end := sorted[i].ChapterEndAt
		if end <= 0 {
			end = sorted[i].AggregateAt
		}
		if start <= now && (end <= 0 || now <= end) {
			return &sorted[i], nil
		}
	}
	if event.StartAt > 0 && now < event.StartAt {
		return &sorted[0], nil
	}
	if event.ClosedAt > 0 && now > event.ClosedAt {
		return &sorted[len(sorted)-1], nil
	}
	return nil, fmt.Errorf("无法自动判断活动 #%d 的 WL 章节，请指定 wl1/wl2 或章节角色", event.ID)
}

func parseLimitToken(token string) (int, bool) {
	raw := strings.TrimPrefix(token, "limit")
	raw = strings.TrimSuffix(raw, "套")
	if raw == token {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	if value < 1 {
		value = 1
	}
	if value > 10 {
		value = 10
	}
	return value, true
}
func parseTimeoutToken(token string) (int, bool) {
	raw := strings.TrimPrefix(token, "timeout")
	raw = strings.TrimPrefix(raw, "超时")
	raw = strings.TrimSuffix(raw, "秒")
	raw = strings.TrimSuffix(raw, "s")
	if raw == token || raw == "" {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	if value < 5 {
		value = 5
	}
	if value > 60 {
		value = 60
	}
	return value * 1000, true
}
func shouldPreferMusicID(args []string, index int) bool {
	for _, offset := range []int{-1, 1} {
		pos := index + offset
		if pos < 0 || pos >= len(args) {
			continue
		}
		switch strings.ToLower(args[pos]) {
		case "easy", "ez", "normal", "nm", "hard", "hd", "expert", "ex", "master", "ma", "mas", "append", "apd", "ap":
			return true
		}
	}
	return false
}

func defaultDeckCardConfig() map[string]renderer.DeckCardConfig {
	return map[string]renderer.DeckCardConfig{"rarity_1": {RankMax: true, EpisodeRead: true, MasterMax: true, SkillMax: true}, "rarity_2": {RankMax: true, EpisodeRead: true, MasterMax: true, SkillMax: true}, "rarity_3": {RankMax: true}, "rarity_4": {RankMax: true}, "rarity_birthday": {RankMax: true}}
}
func setAllCardConfig(configs map[string]renderer.DeckCardConfig, update func(renderer.DeckCardConfig) renderer.DeckCardConfig) {
	for key, value := range configs {
		configs[key] = update(value)
	}
}
func currentEventID(store *masterdata.Store) int {
	now := time.Now().UnixMilli()
	events := store.AllEvents()
	for _, event := range events {
		if event.StartAt <= now && now <= event.AggregateAt {
			return event.ID
		}
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].StartAt > events[j].StartAt })
	if len(events) > 0 {
		return events[0].ID
	}
	return 0
}
func searchMusicID(store *masterdata.Store, aliases map[int]assets.MusicAlias, keyword string) int {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" || store == nil {
		return 0
	}
	result := musicsearch.Search(store, aliases, keyword, musicsearch.Options{Limit: 1})
	if result.Music != nil {
		return result.Music.ID
	}
	if len(result.Musics) > 0 {
		return result.Musics[0].ID
	}
	return 0
}

func ParseDeckRecommendArgsForDebug(raw string, store *masterdata.Store, aliases map[int]assets.MusicAlias, mode string) (renderer.DeckRecommendOptions, *masterdata.MusicInfo, *masterdata.EventInfo, error) {
	return parseDeckRecommendArgs(raw, store, aliases, mode)
}

func LoadDeckRecommendUserDataForDebug(client *suite.Client, gameID string, out *map[string]any) error {
	return loadDeckRecommendUserData(client, gameID, out)
}

func FilterDeckRecommendUserDataForDebug(userData map[string]any, store *masterdata.Store) map[string]any {
	return filterDeckRecommendUserData(userData, store)
}

func FilterDeckRecommendUserDataWithJPMasterForDebug(userData map[string]any, masterMap map[string]any, store *masterdata.Store) map[string]any {
	return filterDeckRecommendUserDataWithJPMaster(userData, masterMap, store)
}

func BuildDeckRecommendMasterDataForDebug(runtime *servers.Runtime) (map[string]any, []string) {
	return buildDeckRecommendMasterData(runtime)
}

func FetchMusicMetasForDebug() ([]map[string]any, error) {
	return fetchMusicMetas()
}

func BuildDeckRecommendCardAssetsForDebug(store *masterdata.Store, resolver *assets.Resolver) map[int]map[string]any {
	return buildDeckRecommendCardAssets(store, resolver)
}

func SuiteProfileFromUserDataForDebug(userData map[string]any, fallbackUID string) map[string]any {
	return suiteProfileFromUserData(userData, fallbackUID)
}

func DeckMusicPayloadForDebug(store *masterdata.Store, music *masterdata.MusicInfo, resolver *assets.Resolver, difficulty string, isPresetDefault bool) any {
	return deckMusicPayload(store, music, resolver, difficulty, isPresetDefault)
}

func BuildDeckRecommendPayloadForDebug(runtime *servers.Runtime, mode string, calc *renderer.DeckRecommendCalculateResponse) map[string]any {
	return buildDeckRecommendPayload(runtime, mode, calc)
}

func DeckRecommendTitleForDebug(mode string) string {
	return deckRecommendTitle(mode)
}

func buildDeckRecommendMasterData(runtime *servers.Runtime) (map[string]any, []string) {
	out := map[string]any{}
	warnings := []string{}
	if runtime == nil {
		return out, []string{"runtime 不可用"}
	}
	jpCfg := config.MasterdataConfig{
		Region: config.RegionJP,
		Source: config.MasterdataSourceMoeSekai,
	}
	resolved, err := config.ResolveMasterdata(jpCfg, config.RegionJP)
	if err != nil {
		return out, append(warnings, "JP masterdata endpoint 解析失败")
	}
	cacheTTL := deckRecommendDefaultMasterCacheTTL
	if runtime.Profile.Masterdata.RefreshInterval > 0 {
		cacheTTL = time.Duration(runtime.Profile.Masterdata.RefreshInterval) * time.Second
	}
	for _, key := range deckRecommendMasterKeys {
		if data, err := loadDeckRecommendMasterDataAny(key, resolved, cacheTTL); err == nil {
			out[key] = data
		} else {
			out[key] = []any{}
			warnings = append(warnings, fmt.Sprintf("缺少 %s", key))
		}
	}
	return out, warnings
}

func allEventDeckBonuses(store *masterdata.Store) []masterdata.EventDeckBonus {
	var out []masterdata.EventDeckBonus
	for _, event := range store.AllEvents() {
		out = append(out, store.GetEventDeckBonuses(event.ID)...)
	}
	return out
}
func allMusicDifficulties(store *masterdata.Store) []masterdata.MusicDifficulty {
	var out []masterdata.MusicDifficulty
	for _, music := range store.AllMusics() {
		out = append(out, store.GetMusicDifficulties(music.ID)...)
	}
	return out
}

func loadDeckRecommendMasterDataAny(key string, resolved config.ResolvedMasterdata, ttl time.Duration) ([]any, error) {
	if deckrecommenddata.IsLocalMasterKey(key) {
		return deckrecommenddata.LoadLocalMasterData(key)
	}
	return loadMasterDataAny(key, resolved, ttl)
}

func loadMasterDataAny(key string, resolved config.ResolvedMasterdata, ttl time.Duration) ([]any, error) {
	if ttl <= 0 {
		ttl = deckRecommendDefaultMasterCacheTTL
	}
	cacheKey := resolved.Region + "|" + resolved.Source + "|" + resolved.URL + "|" + key
	deckMasterDataCache.Lock()
	if deckMasterDataCache.items == nil {
		deckMasterDataCache.items = map[string]deckMasterCacheEntry{}
	}
	if entry, ok := deckMasterDataCache.items[cacheKey]; ok && time.Since(entry.updatedAt) < ttl {
		data := append([]any(nil), entry.data...)
		deckMasterDataCache.Unlock()
		return data, nil
	}
	deckMasterDataCache.Unlock()
	client := &http.Client{Timeout: 20 * time.Second}
	for _, endpoint := range append([]config.ResolvedEndpoint(nil), resolved.Endpoints...) {
		base := strings.TrimRight(endpoint.URL, "/")
		if base == "" {
			continue
		}
		resp, err := client.Get(base + "/" + url.PathEscape(key) + ".json")
		if err != nil {
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil || resp.StatusCode != http.StatusOK {
			continue
		}
		var data []any
		if err := json.Unmarshal(body, &data); err != nil {
			continue
		}
		deckMasterDataCache.Lock()
		deckMasterDataCache.items[cacheKey] = deckMasterCacheEntry{data: append([]any(nil), data...), updatedAt: time.Now()}
		deckMasterDataCache.Unlock()
		return data, nil
	}
	deckMasterDataCache.Lock()
	deckMasterDataCache.items[cacheKey] = deckMasterCacheEntry{data: []any{}, updatedAt: time.Now()}
	deckMasterDataCache.Unlock()
	return nil, fmt.Errorf("masterdata %s not found", key)
}

func fetchMusicMetas() ([]map[string]any, error) {
	musicMetaCache.Lock()
	if len(musicMetaCache.data) > 0 && time.Since(musicMetaCache.updatedAt) < musicMetaCacheTTL {
		cached := cloneMusicMetas(musicMetaCache.data)
		musicMetaCache.Unlock()
		return cached, nil
	}
	musicMetaCache.Unlock()
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(musicMetaURL)
	if err != nil {
		return fallbackMusicMetas(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fallbackMusicMetas(fmt.Errorf("music meta returned %d", resp.StatusCode))
	}
	var data []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fallbackMusicMetas(err)
	}
	musicMetaCache.Lock()
	musicMetaCache.data = cloneMusicMetas(data)
	musicMetaCache.updatedAt = time.Now()
	musicMetaCache.Unlock()
	return data, nil
}
func fallbackMusicMetas(err error) ([]map[string]any, error) {
	musicMetaCache.Lock()
	defer musicMetaCache.Unlock()
	if len(musicMetaCache.data) > 0 {
		return cloneMusicMetas(musicMetaCache.data), nil
	}
	return nil, err
}
func cloneMusicMetas(input []map[string]any) []map[string]any {
	out := make([]map[string]any, len(input))
	for i, item := range input {
		out[i] = make(map[string]any, len(item))
		for k, v := range item {
			out[i][k] = v
		}
	}
	return out
}

func buildDeckRecommendCardAssets(store *masterdata.Store, resolver *assets.Resolver) map[int]map[string]any {
	out := map[int]map[string]any{}
	assetResolver := resolver
	if assetResolver == nil {
		assetResolver = assets.DefaultResolver()
	}
	for _, card := range store.AllCards() {
		payload := renderer.BuildCardDetailPayloadWithAssets(store, card, assetResolver)
		out[card.ID] = map[string]any{"thumbnailUrl": payload.ThumbnailURL, "trainedThumbnailUrl": payload.TrainedThumbnail, "normalThumbnailUrl": payload.ThumbnailURL, "assetSource": payload.AssetSource, "characterName": payload.CharacterName}
	}
	return out
}
func suiteProfileFromUserData(userData map[string]any, fallbackUID string) map[string]any {
	profile := map[string]any{"userId": fallbackUID}
	if raw, ok := userData["userGamedata"].(map[string]any); ok {
		if name, ok := raw["name"]; ok {
			profile["name"] = name
		}
		if rank, ok := raw["rank"]; ok {
			profile["rank"] = rank
		}
		if uid, ok := raw["userId"]; ok {
			profile["userId"] = uid
		}
	}
	if upload, ok := userData["upload_time"]; ok {
		profile["uploadTime"] = upload
	}
	profile["source"] = suite.PublicSource
	return profile
}
func deckMusicPayload(store *masterdata.Store, music *masterdata.MusicInfo, resolver *assets.Resolver, difficulty string, isPresetDefault bool) any {
	if music == nil {
		return nil
	}
	if music.ID == 10000 {
		out := map[string]any{"id": 10000, "title": "おまかせ", "selectedDifficulty": difficulty}
		if isPresetDefault {
			out["isPresetDefault"] = true
		}
		return out
	}
	payload := renderer.BuildMusicDetailPayloadWithAssets(store, *music, resolver)
	payload.SelectedDifficulty = difficulty
	if !isPresetDefault {
		return payload
	}
	// Attach isPresetDefault as an extra field by converting to map.
	data, err := json.Marshal(payload)
	if err != nil {
		return payload
	}
	out := map[string]any{}
	if err := json.Unmarshal(data, &out); err != nil {
		return payload
	}
	out["isPresetDefault"] = true
	return out
}
func assetSourceForRuntime(resolver *assets.Resolver) string {
	if resolver == nil {
		return assets.DefaultResolver().RendererAssetSource()
	}
	return resolver.RendererAssetSource()
}

func formatDeckRecommendText(result *renderer.DeckRecommendCalculateResponse) string {
	lines := []string{"活动组卡推荐"}
	for _, deck := range result.Decks {
		cards := make([]string, 0, len(deck.Cards))
		for _, card := range deck.Cards {
			cards = append(cards, strconv.Itoa(card.CardID))
		}
		lines = append(lines, fmt.Sprintf("#%d %s:%s 活动PT:%d 加成:%.0f%% 综合力:%v 实效:%.0f 卡组:%s", deck.Rank, deckFirstNonEmptyString(deck.ValueLabel, "主值"), formatFloat(deck.Value), deck.EventPoint, deck.EventBonus, deck.Power["total"], deck.MultiLiveScoreUp, strings.Join(cards, ", ")))
	}
	if len(result.Decks) == 0 {
		lines = append(lines, "没有得到推荐结果")
	}
	return strings.Join(lines, "\n")
}
func formatFloat(value float64) string {
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return fmt.Sprintf("%.2f", value)
}
func deckFirstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

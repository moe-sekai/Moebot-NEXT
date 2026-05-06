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
	"moebot-next/internal/masterdata"
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

var deckRecommendSuiteFields = []string{
	suite.FieldUploadTime, suite.FieldUserGamedata, suite.FieldUserDecks, suite.FieldUserCards,
	suite.FieldUserBonds, suite.FieldUserMaterials, suite.FieldUserAreas, suite.FieldUserCharacters,
	suite.FieldUserChallengeLiveSoloDecks, suite.FieldUserChallengeLiveSoloStages,
	suite.FieldUserChallengeLiveSoloResults, suite.FieldUserChallengeLiveSoloHighScoreRewards,
	suite.FieldUserCharacterMissionV2s, suite.FieldUserCharacterMissionV2Statuses,
	suite.FieldUserMysekaiFixtureGameCharacterPerformanceBonuses, suite.FieldUserMysekaiGates,
	"userMusics", "userMusicResults", "userMysekaiMaterials", "userMysekaiCanvases",
	"userWorldBloomSupportDecks", "userHonors", "userMysekaiCharacterTalks",
	"userEvents", "userWorldBlooms", "userMusicAchievements", "userPlayerFrames",
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

			options, music, event, err := parseDeckRecommendArgs(commandArgs(ctx), runtime.Store, mode)
			if err != nil {
				ctx.SendChain(message.Text(err.Error()))
				return
			}
			ctx.SendChain(message.Text("正在组卡中，请稍等一下喵~"))

			var userData map[string]any
			if err := runtime.Suite.GetUserData(user.GameID, "", deckRecommendSuiteFields, &userData); err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("读取 Suite 数据失败：%v", err)))
				return
			}
			masterMap, warnings := buildDeckRecommendMasterData(runtime)
			musicMetas, err := fetchMusicMetas()
			if err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取歌曲分数元数据失败：%v", err)))
				return
			}
			profile := suiteProfileFromUserData(userData, user.GameID)
			req := renderer.DeckRecommendCalculateRequest{
				Region: runtime.Region, RegionLabel: runtime.Label, UserData: userData, MasterData: masterMap,
				MusicMetas: musicMetas, Options: options, CardAssets: buildDeckRecommendCardAssets(runtime.Store, runtime.Assets),
				Music: deckMusicPayload(runtime.Store, music, runtime.Assets, options.Difficulty), Profile: profile,
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
			payload := map[string]any{"title": deckRecommendTitle(mode), "regionLabel": runtime.Label, "profile": calc.Profile, "event": calc.Event, "music": calc.Music, "options": calc.Options, "algorithm": calc.Algorithm, "costMs": calc.CostMS, "warnings": calc.Warnings, "decks": calc.Decks, "assetSource": assetSourceForRuntime(runtime.Assets)}
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

func parseDeckRecommendArgs(raw string, store *masterdata.Store, mode string) (renderer.DeckRecommendOptions, *masterdata.MusicInfo, *masterdata.EventInfo, error) {
	args := strings.Fields(strings.TrimSpace(raw))
	mode = normalizeDeckRecommendMode(mode)
	options := renderer.DeckRecommendOptions{Mode: mode, MusicID: 10000, Difficulty: "master", LiveType: defaultDeckLiveType(mode), Algorithm: "ga", Target: defaultDeckTarget(mode), Limit: 3, TimeoutMS: 15000, BestSkillAsLeader: true, CardConfig: defaultDeckCardConfig()}
	var eventID, musicID int
	bonusTargets := []int{}
	challengeSet := false
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
		switch token {
		case "多人", "协力", "multi":
			options.LiveType = "multi"
		case "单人", "solo":
			options.LiveType = "solo"
		case "自动", "auto":
			options.LiveType = "auto"
		case "欢乐", "cheerful":
			options.LiveType = "cheerful"
		case "分数", "score":
			options.Target = "score"
		case "综合力", "power":
			options.Target = "power"
		case "实效", "skill", "倍率":
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
		case "easy", "ez":
			options.Difficulty = "easy"
		case "normal", "nm":
			options.Difficulty = "normal"
		case "hard", "hd":
			options.Difficulty = "hard"
		case "expert", "ex":
			options.Difficulty = "expert"
		case "master", "ma", "mas":
			options.Difficulty = "master"
		case "append", "apd", "ap":
			options.Difficulty = "append"
		case "满破":
			setAllCardConfig(options.CardConfig, func(c renderer.DeckCardConfig) renderer.DeckCardConfig { c.MasterMax = true; return c })
		case "满技能":
			setAllCardConfig(options.CardConfig, func(c renderer.DeckCardConfig) renderer.DeckCardConfig { c.SkillMax = true; return c })
		case "已读":
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
				} else if store.GetEvent(id) != nil {
					eventID = id
				} else if store.GetMusic(id) != nil {
					musicID = id
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
	if mode == "bonus" && len(options.TargetBonusList) == 0 {
		return options, nil, nil, fmt.Errorf("请输入目标活动加成，例如 /加成组卡 300")
	}
	if musicID == 0 && len(remaining) > 0 {
		musicID = searchMusicID(store, strings.Join(remaining, " "))
		if musicID == 0 {
			return options, nil, nil, fmt.Errorf("找不到曲目关键词：%s", strings.Join(remaining, " "))
		}
	}
	if musicID != 0 {
		options.MusicID = musicID
	}
	if event != nil {
		options.EventID = event.ID
	}
	music := store.GetMusic(options.MusicID)
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
	case "strongest", "challenge", "bonus", "event":
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
	default:
		return "活动组卡推荐"
	}
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
	aliases := map[string]int{"miku": 1, "初音": 1, "初音未来": 1, "rin": 2, "铃": 2, "len": 3, "连": 3, "luka": 4, "巡音": 4, "meiko": 5, "kaito": 6, "ichika": 7, "一歌": 7, "saki": 8, "咲希": 8, "honami": 9, "穗波": 9, "shiho": 10, "志步": 10, "minori": 11, "实乃理": 11, "haruka": 12, "遥": 12, "airi": 13, "爱莉": 13, "shizuku": 14, "雫": 14, "kohane": 15, "心羽": 15, "an": 16, "杏": 16, "akito": 17, "彰人": 17, "toya": 18, "冬弥": 18, "tsukasa": 19, "司": 19, "emu": 20, "笑梦": 20, "nene": 21, "宁宁": 21, "rui": 22, "类": 22, "kanade": 23, "奏": 23, "mafuyu": 24, "真冬": 24, "ena": 25, "绘名": 25, "mizuki": 26, "瑞希": 26}
	id, ok := aliases[strings.ToLower(strings.TrimSpace(token))]
	return id, ok
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
func searchMusicID(store *masterdata.Store, keyword string) int {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return 0
	}
	musics := store.AllMusics()
	for _, music := range musics {
		if strings.ToLower(music.Title) == keyword || strings.ToLower(music.Pronunciation) == keyword {
			return music.ID
		}
	}
	for _, music := range musics {
		if strings.Contains(strings.ToLower(music.Title), keyword) || strings.Contains(strings.ToLower(music.Pronunciation), keyword) {
			return music.ID
		}
	}
	return 0
}

func buildDeckRecommendMasterData(runtime *servers.Runtime) (map[string]any, []string) {
	out := map[string]any{}
	warnings := []string{}
	if runtime == nil || runtime.Store == nil {
		return out, []string{"masterdata 不可用"}
	}
	store := runtime.Store
	out["cards"] = store.AllCards()
	out["events"] = store.AllEvents()
	out["eventCards"] = store.AllEventCards()
	out["eventDeckBonuses"] = allEventDeckBonuses(store)
	out["gameCharacterUnits"] = store.AllCharacterUnits()
	out["honors"] = store.AllHonors()
	out["skills"] = store.AllSkills()
	out["musics"] = store.AllMusics()
	out["musicDifficulties"] = allMusicDifficulties(store)
	resolved, err := config.ResolveMasterdata(runtime.Profile.Masterdata, runtime.Region)
	if err != nil {
		return out, append(warnings, "部分组卡 masterdata 解析失败")
	}
	cacheTTL := time.Duration(runtime.Profile.Masterdata.RefreshInterval) * time.Second
	if cacheTTL <= 0 {
		cacheTTL = deckRecommendDefaultMasterCacheTTL
	}
	for _, key := range deckRecommendMasterKeys {
		if _, exists := out[key]; exists {
			continue
		}
		if data, err := loadMasterDataAny(key, resolved, cacheTTL); err == nil {
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
func deckMusicPayload(store *masterdata.Store, music *masterdata.MusicInfo, resolver *assets.Resolver, difficulty string) any {
	if music == nil {
		return nil
	}
	if music.ID == 10000 {
		return map[string]any{"id": 10000, "title": "おまかせ", "selectedDifficulty": difficulty}
	}
	payload := renderer.BuildMusicDetailPayloadWithAssets(store, *music, resolver)
	payload.SelectedDifficulty = difficulty
	return payload
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

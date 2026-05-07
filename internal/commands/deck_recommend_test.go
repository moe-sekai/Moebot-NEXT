package commands

import (
	"testing"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/suite"
)

func testDeckRecommendStore() *masterdata.Store {
	store := masterdata.NewStore()
	now := time.Now().UnixMilli()
	store.SetAll(&masterdata.MasterData{
		Events: []masterdata.EventInfo{{ID: 123, Name: "测试活动", EventType: "marathon", StartAt: now - 1000, AggregateAt: now + 100000, ClosedAt: now + 200000}, {ID: 999, Name: "旧活动", EventType: "marathon", StartAt: now - 200000, AggregateAt: now - 100000, ClosedAt: now - 50000}},
		Musics: []masterdata.MusicInfo{{ID: 456, Title: "Test Song", Pronunciation: "test song"}, {ID: 789, Title: "另一个曲目", Pronunciation: "another"}},
	})
	return store
}

func TestParseDeckRecommendArgsFixedCardsAndCharacters(t *testing.T) {
	options, _, event, err := parseDeckRecommendArgs("#123 456 miku 一歌", testDeckRecommendStore(), "event")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if event.ID != 123 {
		t.Fatalf("event = %d", event.ID)
	}
	if got := options.FixedCards; len(got) != 2 || got[0] != 123 || got[1] != 456 {
		t.Fatalf("fixed cards = %#v", got)
	}
	if got := options.FixedCharacters; len(got) != 2 || got[0] != 1 || got[1] != 7 {
		t.Fatalf("fixed characters = %#v", got)
	}
}

func TestParseDeckRecommendArgsEventMusicDifficulty(t *testing.T) {
	options, music, event, err := parseDeckRecommendArgs("event123 music456 expert", testDeckRecommendStore(), "event")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if event.ID != 123 {
		t.Fatalf("event = %d", event.ID)
	}
	if music.ID != 456 || options.MusicID != 456 {
		t.Fatalf("music = %d/%d", music.ID, options.MusicID)
	}
	if options.Difficulty != "expert" {
		t.Fatalf("difficulty = %s", options.Difficulty)
	}
}

func TestParseDeckRecommendArgsDefaultMusicAndDifficulty(t *testing.T) {
	options, music, event, err := parseDeckRecommendArgs("", testDeckRecommendStore(), "event")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if event == nil || event.ID != 123 {
		t.Fatalf("event = %#v", event)
	}
	if options.MusicID != deckRecommendDefaultMusicID {
		t.Fatalf("music id = %d, want %d", options.MusicID, deckRecommendDefaultMusicID)
	}
	if options.Difficulty != deckRecommendDefaultDifficulty {
		t.Fatalf("difficulty = %s, want %s", options.Difficulty, deckRecommendDefaultDifficulty)
	}
	if music == nil || music.ID != deckRecommendDefaultMusicID {
		t.Fatalf("default parse should provide fallback music payload, got %#v", music)
	}
}

func TestParseDeckRecommendArgsExplicitDifficultyOverridesDefault(t *testing.T) {
	options, _, _, err := parseDeckRecommendArgs("master", testDeckRecommendStore(), "strongest")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if options.Difficulty != "master" {
		t.Fatalf("difficulty = %s", options.Difficulty)
	}
}

func TestParseDeckRecommendArgsOptions(t *testing.T) {
	options, _, _, err := parseDeckRecommendArgs("综合力 auto all 3套 timeout30s 技能吸取最大 不换队长", testDeckRecommendStore(), "event")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if options.Target != "power" {
		t.Fatalf("target = %s", options.Target)
	}
	if options.LiveType != "auto" {
		t.Fatalf("live = %s", options.LiveType)
	}
	if options.Algorithm != "all" {
		t.Fatalf("algorithm = %s", options.Algorithm)
	}
	if options.Limit != 3 {
		t.Fatalf("limit = %d", options.Limit)
	}
	if options.TimeoutMS != 30000 {
		t.Fatalf("timeout = %d", options.TimeoutMS)
	}
	if options.SkillReferenceChooseStrategy != "max" {
		t.Fatalf("skill ref = %s", options.SkillReferenceChooseStrategy)
	}
	if options.BestSkillAsLeader {
		t.Fatalf("best skill leader should be false")
	}
}

func TestParseDeckRecommendArgsPreferMusicNearDifficulty(t *testing.T) {
	options, music, _, err := parseDeckRecommendArgs("456 master", testDeckRecommendStore(), "event")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if music.ID != 456 || options.MusicID != 456 {
		t.Fatalf("music = %d/%d", music.ID, options.MusicID)
	}
}

func TestParseStrongestDeckArgs(t *testing.T) {
	options, _, event, err := parseDeckRecommendArgs("实效 5套", testDeckRecommendStore(), "strongest")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if options.Mode != "strongest" {
		t.Fatalf("mode = %s", options.Mode)
	}
	if event != nil {
		t.Fatalf("strongest event should be nil")
	}
	if options.Target != "skill" {
		t.Fatalf("target = %s", options.Target)
	}
	if options.Limit != 5 {
		t.Fatalf("limit = %d", options.Limit)
	}
}

func TestParseChallengeDeckArgs(t *testing.T) {
	options, _, event, err := parseDeckRecommendArgs("miku all", testDeckRecommendStore(), "challenge")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if options.Mode != "challenge" {
		t.Fatalf("mode = %s", options.Mode)
	}
	if event != nil {
		t.Fatalf("challenge event should be nil")
	}
	if options.ChallengeCharacterID != 1 {
		t.Fatalf("character = %d", options.ChallengeCharacterID)
	}
	if options.Algorithm != "all" {
		t.Fatalf("algorithm = %s", options.Algorithm)
	}
}

func TestParseBonusDeckArgs(t *testing.T) {
	options, _, event, err := parseDeckRecommendArgs("event123 250 260 270", testDeckRecommendStore(), "bonus")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if options.Mode != "bonus" {
		t.Fatalf("mode = %s", options.Mode)
	}
	if event == nil || event.ID != 123 {
		t.Fatalf("event = %#v", event)
	}
	if len(options.TargetBonusList) != 3 || options.TargetBonusList[0] != 250 || options.TargetBonusList[2] != 270 {
		t.Fatalf("targets = %#v", options.TargetBonusList)
	}
}

func TestNormalizeDeckRecommendUserDataFillsMissingKeys(t *testing.T) {
	userData := normalizeDeckRecommendUserData(map[string]any{})
	if _, ok := userData[suite.FieldUserCards]; !ok {
		t.Fatalf("missing default userCards: %#v", userData)
	}
	if _, ok := userData[suite.FieldUserGamedata]; !ok {
		t.Fatalf("missing default userGamedata: %#v", userData)
	}
	if userData[suite.FieldUserGamedata] != nil {
		t.Fatalf("userGamedata default = %#v, want nil", userData[suite.FieldUserGamedata])
	}
	if userData[suite.FieldUploadTime] != nil {
		t.Fatalf("upload_time default = %#v, want nil", userData[suite.FieldUploadTime])
	}
}

func TestFilterDeckRecommendUserDataFiltersCardsAndHonors(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		Cards: []masterdata.CardInfo{{ID: 101}},
		Honors: []masterdata.HonorInfo{{
			ID:     201,
			Levels: []masterdata.HonorLevel{{HonorID: 201, Level: 1}, {HonorID: 201, Level: 2}},
		}},
	})
	userData := map[string]any{
		suite.FieldUserCards: []any{map[string]any{"cardId": 101}, map[string]any{"cardId": 999}},
		"userHonors": []any{
			map[string]any{"honorId": 201, "level": 2},
			map[string]any{"honorId": 201, "level": 9},
			map[string]any{"honorId": 999, "level": 1},
		},
	}
	filtered := filterDeckRecommendUserData(userData, store)
	cards, _ := filtered[suite.FieldUserCards].([]any)
	honors, _ := filtered["userHonors"].([]any)
	if len(cards) != 1 {
		t.Fatalf("filtered cards = %#v", cards)
	}
	if len(honors) != 1 {
		t.Fatalf("filtered honors = %#v", honors)
	}
}

func TestAllEventDeckBonusesPreserveAttrOnlyBonus(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		Events: []masterdata.EventInfo{{ID: 203}},
		EventDeckBonuses: []masterdata.EventDeckBonus{
			{ID: 1, EventID: 203, GameCharacterUnitID: 1, BonusRate: 25},
			{ID: 2, EventID: 203, CardAttr: "mysterious", BonusRate: 25},
		},
	})
	bonuses := allEventDeckBonuses(store)
	if len(bonuses) != 2 {
		t.Fatalf("bonuses = %#v", bonuses)
	}
	if bonuses[1].GameCharacterUnitID != 0 || bonuses[1].CardAttr != "mysterious" {
		t.Fatalf("attr-only bonus should remain encoded with zero unit id: %#v", bonuses[1])
	}
}

func TestFilterDeckRecommendUserCardsFromJPMaster(t *testing.T) {
	jpMaster := map[string]any{
		"cards": []any{
			map[string]any{"id": float64(101), "characterId": float64(1)},
			map[string]any{"id": float64(102), "characterId": float64(2)},
			map[string]any{"id": float64(103), "characterId": float64(3)},
		},
	}
	userCards := []any{
		map[string]any{"cardId": float64(101), "level": float64(50)},
		map[string]any{"cardId": float64(102), "level": float64(60)},
		map[string]any{"cardId": float64(999), "level": float64(40)},
	}
	filtered := filterDeckRecommendUserCardsFromJPMaster(userCards, jpMaster)
	result, ok := filtered.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", filtered)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 cards, got %d: %#v", len(result), result)
	}
	first := result[0].(map[string]any)
	second := result[1].(map[string]any)
	if intValueFromAny(first["cardId"]) != 101 {
		t.Fatalf("first card = %v, want 101", first["cardId"])
	}
	if intValueFromAny(second["cardId"]) != 102 {
		t.Fatalf("second card = %v, want 102", second["cardId"])
	}
}

func TestFilterDeckRecommendUserCardsFromJPMasterFallsBackWhenNoCards(t *testing.T) {
	jpMaster := map[string]any{}
	userCards := []any{
		map[string]any{"cardId": float64(101)},
		map[string]any{"cardId": float64(999)},
	}
	filtered := filterDeckRecommendUserCardsFromJPMaster(userCards, jpMaster)
	result, ok := filtered.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", filtered)
	}
	if len(result) != 2 {
		t.Fatalf("should return all cards when JP cards unavailable, got %d", len(result))
	}
}

func TestFilterDeckRecommendUserDataWithJPMaster(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		Honors: []masterdata.HonorInfo{{
			ID:     201,
			Levels: []masterdata.HonorLevel{{HonorID: 201, Level: 1}},
		}},
	})
	jpMaster := map[string]any{
		"cards": []any{
			map[string]any{"id": float64(101)},
		},
		"areaItemLevels": []any{
			map[string]any{"areaItemId": float64(1), "level": float64(5)},
		},
		"mysekaiGateLevels": []any{
			map[string]any{"mysekaiGateId": float64(10), "level": float64(3)},
		},
		"characterRanks": []any{
			map[string]any{"characterId": float64(1), "characterRank": float64(1)},
			map[string]any{"characterId": float64(1), "characterRank": float64(3)},
			map[string]any{"characterId": float64(2), "characterRank": float64(5)},
		},
	}
	userData := map[string]any{
		suite.FieldUserCards: []any{
			map[string]any{"cardId": float64(101)},
			map[string]any{"cardId": float64(999)},
		},
		"userHonors": []any{
			map[string]any{"honorId": 201, "level": 1},
			map[string]any{"honorId": 999, "level": 1},
		},
		suite.FieldUserAreas: []any{
			map[string]any{"areaItems": []any{
				map[string]any{"areaItemId": float64(1), "level": float64(5)},
				map[string]any{"areaItemId": float64(2), "level": float64(1)},
			}},
		},
		suite.FieldUserMysekaiGates: []any{
			map[string]any{"mysekaiGateId": float64(10), "mysekaiGateLevel": float64(3)},
			map[string]any{"mysekaiGateId": float64(10), "mysekaiGateLevel": float64(9)},
		},
		suite.FieldUserCharacters: []any{
			map[string]any{"characterId": float64(1), "characterRank": float64(99)},
			map[string]any{"characterId": float64(2), "characterRank": float64(5)},
			map[string]any{"characterId": float64(99), "characterRank": float64(1)},
		},
	}
	filtered := filterDeckRecommendUserDataWithJPMaster(userData, jpMaster, store)
	cards, _ := filtered[suite.FieldUserCards].([]any)
	honors, _ := filtered["userHonors"].([]any)
	areas, _ := filtered[suite.FieldUserAreas].([]any)
	gates, _ := filtered[suite.FieldUserMysekaiGates].([]any)
	characters, _ := filtered[suite.FieldUserCharacters].([]any)
	if len(cards) != 1 {
		t.Fatalf("filtered cards = %d, want 1", len(cards))
	}
	if len(honors) != 1 {
		t.Fatalf("filtered honors = %d, want 1", len(honors))
	}
	areaItems := areas[0].(map[string]any)["areaItems"].([]any)
	if len(areaItems) != 1 {
		t.Fatalf("filtered area items = %d, want 1", len(areaItems))
	}
	if len(gates) != 1 {
		t.Fatalf("filtered mysekai gates = %d, want 1", len(gates))
	}
	if len(characters) != 2 {
		t.Fatalf("filtered characters = %d, want 2", len(characters))
	}
	if rank := intValueFromAny(characters[0].(map[string]any)["characterRank"]); rank != 3 {
		t.Fatalf("clamped character rank = %d, want 3", rank)
	}
}

func TestLoadDeckRecommendMasterDataUsesLocalWorldBloomData(t *testing.T) {
	data, err := loadDeckRecommendMasterDataAny("worldBloomSupportDeckBonusesWL1", config.ResolvedMasterdata{}, time.Hour)
	if err != nil {
		t.Fatalf("load local WL1 data: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("local WL1 data should not be empty")
	}
}

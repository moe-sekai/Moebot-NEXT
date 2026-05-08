package commandparser

import (
	"testing"
	"time"

	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/masterdata"
)

func TestSuiteDebugDeckMode(t *testing.T) {
	cases := map[string]string{
		"deck-recommend":           "event",
		"strongest-deck-recommend": "strongest",
		"challenge-deck-recommend": "challenge",
		"bonus-deck-recommend":     "bonus",
		"unknown":                  "event",
	}
	for input, want := range cases {
		if got := suiteDebugDeckMode(input); got != want {
			t.Fatalf("suiteDebugDeckMode(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSuiteDebugDeckRecommendArgsDefaultMusicAndDifficulty(t *testing.T) {
	store := masterdata.NewStore()
	options, music, event, err := parseSuiteDebugDeckRecommendArgs("", store, nil, "event")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if options.MusicID != deckRecommendDefaultMusicID {
		t.Fatalf("music id = %d, want %d", options.MusicID, deckRecommendDefaultMusicID)
	}
	if options.Difficulty != deckRecommendDefaultDifficulty {
		t.Fatalf("difficulty = %s, want %s", options.Difficulty, deckRecommendDefaultDifficulty)
	}
	if music == nil || music.ID != deckRecommendDefaultMusicID {
		t.Fatalf("music = %#v, want fallback music id %d", music, deckRecommendDefaultMusicID)
	}
	if event != nil {
		t.Fatalf("event = %#v, want nil", event)
	}
}

func TestSuiteDebugDeckRecommendArgsExplicitDifficultyOverridesDefault(t *testing.T) {
	store := masterdata.NewStore()
	options, _, _, err := parseSuiteDebugDeckRecommendArgs("master", store, nil, "bonus")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if options.Difficulty != "master" {
		t.Fatalf("difficulty = %s", options.Difficulty)
	}
}

func TestSuiteDebugChallengeDeckRequiresCharacter(t *testing.T) {
	store := masterdata.NewStore()
	_, _, _, err := parseSuiteDebugDeckRecommendArgs("挑战组卡", store, nil, "challenge")
	if err == nil {
		t.Fatal("expected missing challenge character error")
	}
}

func TestSuiteDebugChallengeDeckUsesLocalCharacterAlias(t *testing.T) {
	store := masterdata.NewStore()
	options, music, event, err := parseSuiteDebugDeckRecommendArgs("挑战组卡 miku", store, nil, "challenge")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if options.ChallengeCharacterID != 21 {
		t.Fatalf("challenge character = %d, want 21", options.ChallengeCharacterID)
	}
	if options.MusicID != deckRecommendDefaultMusicID || music == nil || music.ID != deckRecommendDefaultMusicID {
		t.Fatalf("music = %#v / %d, want default", music, options.MusicID)
	}
	if event != nil {
		t.Fatalf("event = %#v, want nil", event)
	}
}

func TestSuiteDebugEventDeckCharacterAliasDoesNotOverrideDefaultMusic(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{Events: []masterdata.EventInfo{{ID: 202, Name: "测试活动", EventType: "marathon"}}})
	options, music, event, err := parseSuiteDebugDeckRecommendArgs("活动组卡 202 ick", store, nil, "event")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if event == nil || event.ID != 202 {
		t.Fatalf("event = %#v", event)
	}
	if options.MusicID != deckRecommendDefaultMusicID || music == nil || music.ID != deckRecommendDefaultMusicID {
		t.Fatalf("music = %#v / %d, want default", music, options.MusicID)
	}
	if options.SupportCharacterID != 1 {
		t.Fatalf("support character = %d", options.SupportCharacterID)
	}
	if len(options.FixedCharacters) != 0 {
		t.Fatalf("fixed characters = %#v", options.FixedCharacters)
	}
}

func TestSuiteDebugEventDeckWorldBloomAliasSetsSupportCharacter(t *testing.T) {
	store := masterdata.NewStore()
	now := time.Now().UnixMilli()
	store.SetAll(&masterdata.MasterData{
		Events:      []masterdata.EventInfo{{ID: 202, Name: "WL3", EventType: "world_bloom", StartAt: now - 1000, AggregateAt: now + 1000, ClosedAt: now + 2000}},
		WorldBlooms: []masterdata.WorldBloom{{ID: 1, EventID: 202, GameCharacterID: 1, ChapterNo: 1, ChapterStartAt: now - 1000, ChapterEndAt: now + 1000}},
	})
	options, music, event, err := parseSuiteDebugDeckRecommendArgs("活动组卡 202 ick", store, nil, "event")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if event == nil || event.ID != 202 {
		t.Fatalf("event = %#v", event)
	}
	if options.SupportCharacterID != 1 {
		t.Fatalf("support character = %d", options.SupportCharacterID)
	}
	if len(options.FixedCharacters) != 0 {
		t.Fatalf("fixed characters = %#v", options.FixedCharacters)
	}
	if options.MusicID != deckRecommendDefaultMusicID || music == nil || music.ID != deckRecommendDefaultMusicID {
		t.Fatalf("music = %#v / %d, want default", music, options.MusicID)
	}
}

func TestDebugFilterDeckRecommendUserCardsAndHonors(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		Cards: []masterdata.CardInfo{{ID: 101}},
		Honors: []masterdata.HonorInfo{{
			ID:     201,
			Levels: []masterdata.HonorLevel{{HonorID: 201, Level: 1}, {HonorID: 201, Level: 2}},
		}},
	})
	cards := filterDeckRecommendUserCards([]any{map[string]any{"cardId": 101}, map[string]any{"cardId": 999}}, store).([]any)
	honors := filterDeckRecommendUserHonors([]any{
		map[string]any{"honorId": 201, "level": 2},
		map[string]any{"honorId": 201, "level": 9},
		map[string]any{"honorId": 999, "level": 1},
	}, store).([]any)
	if len(cards) != 1 || len(honors) != 1 {
		t.Fatalf("cards=%#v honors=%#v", cards, honors)
	}
}

func TestSuiteDebugDeckRecommendArgsResolvesMusicViaAlias(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		Musics: []masterdata.MusicInfo{{ID: 789, Title: "另一个曲目", Pronunciation: "another"}},
	})
	aliases := map[int]assets.MusicAlias{
		789: {MusicID: 789, Title: "另一个曲目", Aliases: []string{"龙"}},
	}
	options, music, _, err := parseSuiteDebugDeckRecommendArgs("龙 hd", store, aliases, "event")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if music == nil || music.ID != 789 {
		t.Fatalf("music = %#v, want id 789 from alias", music)
	}
	if options.Difficulty != "hard" {
		t.Fatalf("difficulty = %s, want hard", options.Difficulty)
	}
	if options.IsPresetDefault {
		t.Fatalf("IsPresetDefault should be false when alias resolved")
	}
}

func TestSuiteDebugDeckRecommendArgsDefaultMusicMarksPresetDefault(t *testing.T) {
	store := masterdata.NewStore()
	options, _, _, err := parseSuiteDebugDeckRecommendArgs("", store, nil, "event")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if !options.IsPresetDefault {
		t.Fatalf("IsPresetDefault should be true when no music specified")
	}
}

func TestSuiteDebugDeckRecommendArgsDefaultUsesStoreTitle(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		Musics: []masterdata.MusicInfo{{ID: deckRecommendDefaultMusicID, Title: "RealDefaultSong"}},
	})
	_, music, _, err := parseSuiteDebugDeckRecommendArgs("", store, nil, "event")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if music == nil || music.Title != "RealDefaultSong" {
		t.Fatalf("music = %#v, want real title from store", music)
	}
}

func TestSuiteDebugLoadDeckRecommendMasterDataUsesLocalWorldBloomData(t *testing.T) {
	data, err := suiteDebugLoadDeckRecommendMasterDataAny("worldBloomSupportDeckBonusesWL1", config.ResolvedMasterdata{}, time.Hour)
	if err != nil {
		t.Fatalf("load local WL1 data: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("local WL1 data should not be empty")
	}
}

func TestIsSuiteFetchError(t *testing.T) {
	fetchErrs := []error{
		configError("suite request returned 500"),
		configError("decode suite response: invalid character"),
		configError("read suite response: unexpected EOF"),
	}
	for _, err := range fetchErrs {
		if !isSuiteFetchError(err) {
			t.Fatalf("expected fetch error for %q", err.Error())
		}
	}
	otherErrs := []error{
		configError("组卡计算失败：renderer returned 500"),
		configError("活动组卡 暂未支持临时绑定调试"),
	}
	for _, err := range otherErrs {
		if isSuiteFetchError(err) {
			t.Fatalf("did not expect fetch error for %q", err.Error())
		}
	}
}

type configError string

func (e configError) Error() string { return string(e) }

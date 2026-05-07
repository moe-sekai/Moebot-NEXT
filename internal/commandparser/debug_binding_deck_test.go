package commandparser

import (
	"testing"

	"moebot-next/internal/masterdata"
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
	options, music, event, err := parseSuiteDebugDeckRecommendArgs("", store, "event")
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
	options, _, _, err := parseSuiteDebugDeckRecommendArgs("master", store, "bonus")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if options.Difficulty != "master" {
		t.Fatalf("difficulty = %s", options.Difficulty)
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

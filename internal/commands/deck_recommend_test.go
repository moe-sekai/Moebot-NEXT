package commands

import (
	"testing"
	"time"

	"moebot-next/internal/masterdata"
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

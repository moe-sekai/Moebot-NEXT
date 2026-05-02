package renderer

import (
	"testing"

	"moebot-next/internal/sekai"
)

func TestBuildProfileCardPayloadMapsBasicProfile(t *testing.T) {
	payload := BuildProfileCardPayload(sekai.Profile{
		UserID:     "7485966462906096424",
		Name:       "测试用户",
		Rank:       321,
		Signature:  "你好 SEKAI",
		TotalPower: 123456,
		Stats: sekai.ProfileStats{
			MvpCount:       2516,
			SuperStarCount: 313,
		},
		MusicClearCounts: []sekai.MusicClearCount{
			{Difficulty: "easy", LiveClear: 561, FullCombo: 561, AllPerfect: 561},
			{Difficulty: "master", LiveClear: 493, FullCombo: 493, AllPerfect: 443},
		},
		CharacterRanks: []sekai.CharacterRank{
			{CharacterID: 2, Rank: 55},
			{CharacterID: 1, Rank: 41},
		},
		ChallengeLive: &sekai.ChallengeLiveResult{CharacterID: 20, HighScore: 2063777},
		ProfileHonors: []sekai.ProfileHonor{
			{Seq: 1, HonorID: 136, Level: 1},
			{Seq: 2, HonorID: 79, Level: 2},
		},
		LeaderCard: &sekai.ProfileDeckCard{CardID: 139, Level: 60, Mastery: 0, DefaultImage: "original", SpecialTrained: true},
		DeckCards: []sekai.ProfileDeckCard{
			{CardID: 139, Level: 60, Mastery: 0, DefaultImage: "original", SpecialTrained: true},
			{CardID: 1162, Level: 60, Mastery: 5, DefaultImage: "special_training", SpecialTrained: true},
		},
	})

	if payload.UserID != "7485966462906096424" {
		t.Fatalf("UserID = %q", payload.UserID)
	}
	if payload.Name != "测试用户" || payload.Rank != 321 || payload.Signature != "你好 SEKAI" {
		t.Fatalf("payload = %+v", payload)
	}
	if payload.TotalPower != 123456 {
		t.Fatalf("TotalPower = %d", payload.TotalPower)
	}
	if payload.Stats == nil || payload.Stats.MvpCount != 2516 || payload.Stats.SuperStarCount != 313 {
		t.Fatalf("Stats = %+v", payload.Stats)
	}
	if len(payload.MusicClearCounts) != 2 || payload.MusicClearCounts[1].Difficulty != "master" || payload.MusicClearCounts[1].AllPerfect != 443 {
		t.Fatalf("MusicClearCounts = %+v", payload.MusicClearCounts)
	}
	if len(payload.CharacterRanks) != 2 || payload.CharacterRanks[0].CharacterName == "" || payload.CharacterRanks[0].Rank != 55 {
		t.Fatalf("CharacterRanks = %+v", payload.CharacterRanks)
	}
	if payload.ChallengeLive == nil || payload.ChallengeLive.CharacterName == "" || payload.ChallengeLive.HighScore != 2063777 {
		t.Fatalf("ChallengeLive = %+v", payload.ChallengeLive)
	}
	if len(payload.ProfileHonors) != 2 || payload.ProfileHonors[0].HonorID != 136 || payload.ProfileHonors[1].Level != 2 {
		t.Fatalf("ProfileHonors = %+v", payload.ProfileHonors)
	}
	if payload.LeaderCard == nil || payload.LeaderCard.CardID != 139 || payload.LeaderCard.Level != 60 {
		t.Fatalf("LeaderCard = %+v", payload.LeaderCard)
	}
	if payload.LeaderCard.IsTrained {
		t.Fatalf("LeaderCard IsTrained = true, want false when defaultImage is original")
	}
	if len(payload.DeckCards) != 2 || payload.DeckCards[1].Mastery != 5 {
		t.Fatalf("DeckCards = %+v", payload.DeckCards)
	}
	if !payload.DeckCards[1].IsTrained {
		t.Fatalf("DeckCards[1] IsTrained = false, want true when defaultImage is special_training")
	}
}

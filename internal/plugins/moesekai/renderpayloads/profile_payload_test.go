package renderpayloads

import (
	"strings"
	"testing"

	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/sekai"
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

func TestBuildProfileCardPayloadResolvesHonorLevelAssets(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{Honors: []masterdata.HonorInfo{
		{
			ID:              79,
			GroupID:         10,
			HonorRarity:     "middle",
			Name:            "皆传",
			AssetbundleName: "honor_base",
			Levels: []masterdata.HonorLevel{
				{HonorID: 79, Level: 1, HonorRarity: "middle", AssetbundleName: "honor_base"},
				{HonorID: 79, Level: 12, HonorRarity: "highest", AssetbundleName: "honor_level_12"},
			},
		},
	}})

	payload := BuildProfileCardPayloadWithStore(store, sekai.Profile{
		UserID: "123",
		Name:   "测试用户",
		Rank:   100,
		ProfileHonors: []sekai.ProfileHonor{
			{Seq: 1, HonorID: 79, Level: 12},
		},
	})

	if len(payload.ProfileHonors) != 1 {
		t.Fatalf("ProfileHonors length = %d", len(payload.ProfileHonors))
	}
	honor := payload.ProfileHonors[0]
	if honor.AssetbundleName != "honor_level_12" || honor.HonorRarity != "highest" {
		t.Fatalf("honor asset resolution = %+v", honor)
	}
	if !strings.Contains(honor.ImageURL, "/honor/honor_level_12/degree_main.png") {
		t.Fatalf("ImageURL = %q", honor.ImageURL)
	}
	if !strings.Contains(honor.FrameURL, "/honor/frame/frame_degree_m_4.png") {
		t.Fatalf("FrameURL = %q", honor.FrameURL)
	}
	if !strings.Contains(honor.LevelIconURL, "icon_degreeLv.png") || !strings.Contains(honor.LevelIcon6URL, "icon_degreeLv6.png") {
		t.Fatalf("level icons = %q / %q", honor.LevelIconURL, honor.LevelIcon6URL)
	}
}

func TestBuildProfileCardPayloadResolvesBondsHonorAssets(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		CharacterUnits: []masterdata.GameCharacterUnit{
			{ID: 1, GameCharacterID: 1, Unit: "light_sound", ColorCode: "#33aaee"},
			{ID: 20, GameCharacterID: 20, Unit: "school_refusal", ColorCode: "#ddaacc"},
		},
		BondsHonors: []masterdata.BondsHonorInfo{
			{ID: 101201, BondsGroupID: 1012, GameCharacterUnitID1: 1, GameCharacterUnitID2: 20, HonorRarity: "high", Name: "羁绊称号"},
		},
		BondsHonorWords: []masterdata.BondsHonorWordInfo{
			{ID: 10120101, BondsGroupID: 1012, AssetbundleName: "honorname_0120_01_01", Name: "羁绊文字"},
		},
	})

	payload := BuildProfileCardPayloadWithStore(store, sekai.Profile{
		UserID: "123",
		Name:   "测试用户",
		Rank:   100,
		ProfileHonors: []sekai.ProfileHonor{
			{Seq: 3, HonorType: "bonds", HonorID: 101201, Level: 4, BondsHonorViewType: "reverse", BondsHonorWordID: 10120101},
		},
	})

	if len(payload.ProfileHonors) != 1 {
		t.Fatalf("ProfileHonors length = %d", len(payload.ProfileHonors))
	}
	honor := payload.ProfileHonors[0]
	if honor.LeftCharacterID != 20 || honor.RightCharacterID != 1 {
		t.Fatalf("bonds characters = %+v", honor)
	}
	if honor.LeftColor != "#ddaacc" || honor.RightColor != "#33aaee" {
		t.Fatalf("bonds colors = %+v", honor)
	}
	if !strings.Contains(honor.LeftCharacterURL, "/bonds_honor/character/chr_sd_20_01.png") {
		t.Fatalf("LeftCharacterURL = %q", honor.LeftCharacterURL)
	}
	if honor.BondsHonorWordAssetbundleName != "honorname_0120_01_01" || !strings.Contains(honor.BondsHonorWordURL, "/bonds_honor/word/honorname_0120_01_01_01.png") {
		t.Fatalf("bonds word = %+v", honor)
	}
}

package renderpayloads

import (
	"testing"

	"moebot-next/internal/plugins/moesekai/masterdata"
)

func TestBuildMusicDetailPayloadIncludesDuration(t *testing.T) {
	payload := BuildMusicDetailPayload(nil, masterdata.MusicInfo{
		ID:                    74,
		Title:                 "独りんぼエンヴィー",
		SecForMusicScoreMaker: 74,
		FillerSec:             9,
	})

	if payload.DurationSec != 74 {
		t.Fatalf("DurationSec = %v, want 74", payload.DurationSec)
	}
}

func TestBuildCardDetailPayloadIncludesRelatedEvents(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		Cards: []masterdata.CardInfo{{
			ID:              1001,
			CharacterID:     1,
			CardRarityType:  "rarity_4",
			Attr:            "cute",
			Prefix:          "测试卡牌",
			AssetbundleName: "card_test",
		}},
		Events: []masterdata.EventInfo{{
			ID:              2001,
			Name:            "测试活动",
			EventType:       "marathon",
			AssetbundleName: "event_test",
			StartAt:         1700000000000,
			AggregateAt:     1700100000000,
			ClosedAt:        1700200000000,
			Unit:            "light_sound",
		}},
		EventCards: []masterdata.EventCard{{ID: 1, EventID: 2001, CardID: 1001}},
	})

	card := store.GetCard(1001)
	if card == nil {
		t.Fatal("test card not found")
	}

	payload := BuildCardDetailPayloadWithAssets(store, *card, nil)

	if len(payload.Events) != 1 {
		t.Fatalf("len(payload.Events) = %d, want 1", len(payload.Events))
	}
	if payload.Events[0].ID != 2001 || payload.Events[0].Name != "测试活动" {
		t.Fatalf("payload.Events[0] = %#v, want test event", payload.Events[0])
	}
	if payload.Events[0].AssetbundleName != "event_test" {
		t.Fatalf("payload.Events[0].AssetbundleName = %q, want event_test", payload.Events[0].AssetbundleName)
	}
}

func TestBuildEventInfoPayloadIncludesBonusCards(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		Cards: []masterdata.CardInfo{{
			ID:              1001,
			CharacterID:     1,
			CardRarityType:  "rarity_4",
			Attr:            "cute",
			Prefix:          "测试卡牌",
			AssetbundleName: "card_test",
		}},
		Events: []masterdata.EventInfo{{
			ID:          2001,
			Name:        "测试活动",
			EventType:   "marathon",
			StartAt:     1700000000000,
			AggregateAt: 1700100000000,
			ClosedAt:    1700200000000,
			Unit:        "light_sound",
		}},
		EventCards: []masterdata.EventCard{{ID: 1, EventID: 2001, CardID: 1001}},
	})

	event := store.GetEvent(2001)
	if event == nil {
		t.Fatal("test event not found")
	}

	payload := BuildEventInfoPayloadWithAssets(store, *event, nil)

	if len(payload.BonusCards) != 1 {
		t.Fatalf("len(payload.BonusCards) = %d, want 1", len(payload.BonusCards))
	}
	if payload.BonusCards[0].ID != 1001 || payload.BonusCards[0].Prefix != "测试卡牌" {
		t.Fatalf("payload.BonusCards[0] = %#v, want test card", payload.BonusCards[0])
	}
}

func TestBuildEventInfoPayloadIncludesPickupCards(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		Cards: []masterdata.CardInfo{
			{ID: 1001, CharacterID: 1, CardRarityType: "rarity_4", Attr: "cute", Prefix: "活动卡牌", AssetbundleName: "card_1001"},
			{ID: 1002, CharacterID: 2, CardRarityType: "rarity_4", Attr: "cool", Prefix: "Pickup卡牌", AssetbundleName: "card_1002"},
		},
		Events: []masterdata.EventInfo{{
			ID:          2001,
			Name:        "测试活动",
			EventType:   "marathon",
			StartAt:     1700000000000,
			AggregateAt: 1700100000000,
			ClosedAt:    1700200000000,
			Unit:        "light_sound",
		}},
		EventCards: []masterdata.EventCard{{ID: 1, EventID: 2001, CardID: 1001}},
		Gachas: []masterdata.GachaInfo{{
			ID:           3001,
			GachaType:    "limited",
			Name:         "测试卡池",
			StartAt:      1699990000000,
			EndAt:        1700210000000,
			GachaPickups: []masterdata.GachaPickup{{ID: 1, GachaID: 3001, CardID: 1002, GachaPickupType: "normal"}},
		}},
	})

	event := store.GetEvent(2001)
	if event == nil {
		t.Fatal("test event not found")
	}

	payload := BuildEventInfoPayloadWithAssets(store, *event, nil)

	if len(payload.PickupCards) != 1 {
		t.Fatalf("len(payload.PickupCards) = %d, want 1", len(payload.PickupCards))
	}
	if payload.PickupCards[0].ID != 1002 || payload.PickupCards[0].Prefix != "Pickup卡牌" {
		t.Fatalf("payload.PickupCards[0] = %#v, want pickup card", payload.PickupCards[0])
	}
}

func TestCardSupplyTypeDisplayName(t *testing.T) {
	cases := map[string]string{
		"normal":                    "常驻",
		"birthday":                  "生日",
		"term_limited":              "期间限定",
		"colorful_festival_limited": "CFES限定",
		"bloom_festival_limited":    "BFES限定",
		"unit_event_limited":        "WorldLink限定",
		"collaboration_limited":     "联动限定",
		"":                          "常驻",
	}

	for supplyType, want := range cases {
		if got := CardSupplyTypeDisplayName(supplyType); got != want {
			t.Fatalf("CardSupplyTypeDisplayName(%q) = %q, want %q", supplyType, got, want)
		}
	}
}

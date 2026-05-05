package masterdata

import (
	"encoding/json"
	"testing"
)

func TestCardInfoUnmarshalCardParametersObject(t *testing.T) {
	payload := []byte(`{
		"id": 7,
		"cardParameters": {
			"param1": [100, 200],
			"param2": [10, 20]
		}
	}`)
	var card CardInfo
	if err := json.Unmarshal(payload, &card); err != nil {
		t.Fatalf("unmarshal card: %v", err)
	}
	if len(card.CardParameters) != 4 {
		t.Fatalf("len(cardParameters) = %d, want 4", len(card.CardParameters))
	}
	first := card.CardParameters[0]
	if first.CardID != 7 || first.CardParameterType != "param1" || first.CardLevel != 1 || first.Power != 100 {
		t.Fatalf("first card parameter = %#v", first)
	}
}

func TestCardInfoUnmarshalCardParametersArray(t *testing.T) {
	payload := []byte(`{
		"id": 8,
		"cardParameters": [
			{"id":1,"cardId":8,"cardLevel":1,"cardParameterType":"param1","power":111}
		]
	}`)
	var card CardInfo
	if err := json.Unmarshal(payload, &card); err != nil {
		t.Fatalf("unmarshal card: %v", err)
	}
	if len(card.CardParameters) != 1 || card.CardParameters[0].Power != 111 {
		t.Fatalf("cardParameters = %#v", card.CardParameters)
	}
}

func TestMusicInfoUnmarshalCategoryObjects(t *testing.T) {
	payload := []byte(`{
		"id": 1,
		"title": "Tell Your World",
		"categories": [{"musicCategoryName":"mv"}, {"musicCategoryName":"original"}]
	}`)
	var music MusicInfo
	if err := json.Unmarshal(payload, &music); err != nil {
		t.Fatalf("unmarshal music: %v", err)
	}
	if len(music.Categories) != 2 || music.Categories[0] != "mv" || music.Categories[1] != "original" {
		t.Fatalf("categories = %#v", music.Categories)
	}
}

func TestMusicInfoUnmarshalCategoryStrings(t *testing.T) {
	payload := []byte(`{
		"id": 1,
		"title": "Tell Your World",
		"categories": ["mv", "original"]
	}`)
	var music MusicInfo
	if err := json.Unmarshal(payload, &music); err != nil {
		t.Fatalf("unmarshal music: %v", err)
	}
	if len(music.Categories) != 2 || music.Categories[0] != "mv" || music.Categories[1] != "original" {
		t.Fatalf("categories = %#v", music.Categories)
	}
}

func TestStoreIndexesResourceBoxDetails(t *testing.T) {
	store := NewStore()
	store.SetAll(&MasterData{ResourceBoxDetails: []ResourceBoxDetail{
		{ResourceBoxPurpose: "challenge_live_high_score", ResourceBoxID: 10, ResourceType: "jewel", ResourceQuantity: 100},
		{ResourceBoxPurpose: "challenge_live_high_score", ResourceBoxID: 10, ResourceType: "material", ResourceID: 15, ResourceQuantity: 2},
	}})
	details := store.GetResourceBoxDetails("challenge_live_high_score", 10)
	if len(details) != 2 {
		t.Fatalf("details len = %d, want 2: %#v", len(details), details)
	}
	if details[0].ResourceQuantity != 100 || details[1].ResourceID != 15 {
		t.Fatalf("unexpected details: %#v", details)
	}
}

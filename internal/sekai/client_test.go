package sekai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"moebot-next/internal/config"
)

func TestClientGetProfileUsesRegionPathAndConfiguredHeaders(t *testing.T) {
	var gotPath string
	var gotHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotHeader = r.Header.Get("X-Test-Header")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]any{
				"userId": "7485966462906096424",
				"name":   "测试用户",
				"rank":   321,
			},
			"userProfile": map[string]any{
				"word": "你好 SEKAI",
			},
		})
	}))
	defer server.Close()

	client := NewClient(config.SekaiAPIConfig{
		Enabled: true,
		BaseURL: server.URL,
		Region:  "cn",
		Headers: map[string]string{"X-Test-Header": "test-value"},
		Timeout: 1,
	})

	profile, err := client.GetProfile("7485966462906096424")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/cn/7485966462906096424/profile" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotHeader != "test-value" {
		t.Fatalf("header = %q", gotHeader)
	}
	if profile.Name != "测试用户" || profile.Rank != 321 || profile.UserID != "7485966462906096424" {
		t.Fatalf("profile = %+v", profile)
	}
	if profile.Signature != "你好 SEKAI" {
		t.Fatalf("signature = %q", profile.Signature)
	}
}

func TestProfileResponseNormalizesRealProfileShape(t *testing.T) {
	var result profileResponse
	raw := []byte(`{
		"user":{"userId":7485966462906096424,"name":"luoxia","rank":386},
		"userProfile":{"word":"大概就是这样的结局吧","twitterId":""},
		"totalPower":{"totalPower":317012},
		"userDeck":{"leader":139,"member1":139,"member2":1162,"member3":195,"member4":1164,"member5":238},
		"userCards":[{"cardId":139,"level":60,"masterRank":0,"specialTrainingStatus":"done","defaultImage":"original"}],
		"userCharacters":[{"characterId":1,"characterRank":41},{"characterId":2,"characterRank":55}],
		"userChallengeLiveSoloResult":{"characterId":20,"highScore":2063777},
		"userProfileHonors":[{"seq":1,"profileHonorType":"normal","honorId":136,"honorLevel":1},{"seq":2,"profileHonorType":"normal","honorId":79,"honorLevel":2}],
		"userMultiLiveTopScoreCount":{"mvp":2516,"superStar":313},
		"userMusicDifficultyClearCount":[
			{"musicDifficultyType":"easy","liveClear":561,"fullCombo":561,"allPerfect":561},
			{"musicDifficultyType":"master","liveClear":493,"fullCombo":493,"allPerfect":443}
		]
	}`)
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	profile := result.normalize("fallback")
	if profile.UserID != "7485966462906096424" {
		t.Fatalf("UserID = %q", profile.UserID)
	}
	if profile.TotalPower != 317012 {
		t.Fatalf("TotalPower = %d", profile.TotalPower)
	}
	if len(profile.DeckCards) != 5 || profile.DeckCards[0].CardID != 139 || profile.DeckCards[0].Level != 60 {
		t.Fatalf("DeckCards = %+v", profile.DeckCards)
	}
	if profile.Stats.MvpCount != 2516 || profile.Stats.SuperStarCount != 313 {
		t.Fatalf("Stats = %+v", profile.Stats)
	}
	if len(profile.MusicClearCounts) != 2 || profile.MusicClearCounts[0].Difficulty != "easy" || profile.MusicClearCounts[1].AllPerfect != 443 {
		t.Fatalf("MusicClearCounts = %+v", profile.MusicClearCounts)
	}
	if len(profile.CharacterRanks) != 2 || profile.CharacterRanks[0].CharacterID != 2 || profile.CharacterRanks[0].Rank != 55 {
		t.Fatalf("CharacterRanks = %+v", profile.CharacterRanks)
	}
	if profile.ChallengeLive == nil || profile.ChallengeLive.CharacterID != 20 || profile.ChallengeLive.HighScore != 2063777 {
		t.Fatalf("ChallengeLive = %+v", profile.ChallengeLive)
	}
	if len(profile.ProfileHonors) != 2 || profile.ProfileHonors[0].HonorID != 136 || profile.ProfileHonors[1].Level != 2 {
		t.Fatalf("ProfileHonors = %+v", profile.ProfileHonors)
	}
}

func TestClientDisabledReturnsError(t *testing.T) {
	client := &Client{enabled: false, timeout: time.Second}
	if _, err := client.GetProfile("1"); err == nil {
		t.Fatal("expected disabled client error")
	}
}

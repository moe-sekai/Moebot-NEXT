package suite

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"moebot-next/internal/config"
)

func TestClientGetStatusUsesModeFilterAndBearerToken(t *testing.T) {
	var gotPath string
	var gotMode string
	var gotFilter string
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMode = r.URL.Query().Get("mode")
		gotFilter = r.URL.Query().Get("filter")
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"upload_time": int64(1700000000000),
			"source":      "moesekai",
			"userGamedata": map[string]any{
				"userId": "123456789012345678",
				"name":   "测试玩家",
			},
		})
	}))
	defer server.Close()

	client := NewClient(config.SuiteAPIConfig{
		Enabled:     true,
		URL:         server.URL + "/api/cn/user/{uid}/suite",
		Token:       "secret-token",
		Timeout:     1,
		DefaultMode: config.SuiteModeLatest,
	})

	status, err := client.GetStatus("123456789012345678", config.SuiteModeMoeSekai)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/cn/user/123456789012345678/suite" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotMode != config.SuiteModeMoeSekai {
		t.Fatalf("mode = %q", gotMode)
	}
	if gotFilter != "upload_time,userGamedata" {
		t.Fatalf("filter = %q", gotFilter)
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("authorization = %q", gotAuth)
	}
	if status.UserID != "123456789012345678" || status.Name != "测试玩家" || status.Source != "moesekai" || status.UploadTime != 1700000000000 {
		t.Fatalf("status = %+v", status)
	}
}

func TestClientRejectsUnsupportedMode(t *testing.T) {
	client := NewClient(config.SuiteAPIConfig{Enabled: true, URL: "https://example.test/{uid}"})
	if _, err := client.GetStatus("1", "default"); err == nil {
		t.Fatal("expected unsupported mode error")
	}
}

func TestFieldsAddsCommonProfileFieldsBeforeFeatureFields(t *testing.T) {
	fields := Fields("userGachas", FieldUserMaterials)
	want := []string{"upload_time", "userGamedata", "userDecks", "userCards", "userGachas", "userMaterials"}
	if len(fields) != len(want) {
		t.Fatalf("fields len = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("fields[%d] = %q, want %q; all fields: %#v", i, fields[i], want[i], fields)
		}
	}
}

func TestClientGetUserDataDecodesIntoFeatureProfile(t *testing.T) {
	var gotFilter string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFilter = r.URL.Query().Get("filter")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"upload_time": int64(1700000000000),
			"source":      "moesekai",
			"userGamedata": map[string]any{
				"userId": "123456789012345678",
				"name":   "测试玩家",
				"deck":   1,
				"coin":   12345,
			},
			"userDecks": []map[string]any{{
				"deckId":  1,
				"member1": 101,
			}},
			"userCards": []map[string]any{{
				"cardId": 101,
				"level":  60,
			}},
			"userGachas": []map[string]any{{
				"gachaId": 1001,
				"count":   10,
			}},
		})
	}))
	defer server.Close()

	client := NewClient(config.SuiteAPIConfig{
		Enabled:     true,
		URL:         server.URL + "/api/cn/user/{uid}/suite",
		DefaultMode: config.SuiteModeLatest,
	})

	type featureProfile struct {
		BaseProfile
		UserGamedata UserGamedata `json:"userGamedata"`
		UserDecks    []UserDeck   `json:"userDecks"`
		UserCards    []UserCard   `json:"userCards"`
		UserGachas   []struct {
			GachaID int `json:"gachaId"`
			Count   int `json:"count"`
		} `json:"userGachas"`
	}

	var profile featureProfile
	if err := client.GetUserData("123456789012345678", config.SuiteModeMoeSekai, Fields("userGachas"), &profile); err != nil {
		t.Fatal(err)
	}
	if gotFilter != "upload_time,userGamedata,userDecks,userCards,userGachas" {
		t.Fatalf("filter = %q", gotFilter)
	}
	if profile.UserGamedata.UserID.String() != "123456789012345678" || profile.UserGamedata.Name != "测试玩家" || profile.UserGamedata.Coin != 12345 {
		t.Fatalf("userGamedata = %+v", profile.UserGamedata)
	}
	if len(profile.UserCards) != 1 || profile.UserCards[0].CardID != 101 || profile.UserCards[0].Level != 60 {
		t.Fatalf("userCards = %+v", profile.UserCards)
	}
	if len(profile.UserGachas) != 1 || profile.UserGachas[0].GachaID != 1001 || profile.UserGachas[0].Count != 10 {
		t.Fatalf("userGachas = %+v", profile.UserGachas)
	}
}

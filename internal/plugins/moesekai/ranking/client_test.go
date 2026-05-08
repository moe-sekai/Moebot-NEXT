package ranking

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGetLatestUsesRegionPathAndParsesRankings(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{
			"event_id":   165,
			"region":     "cn",
			"updated_at": 1777744092015,
			"rankings": []map[string]any{{
				"rank":       100,
				"score":      12345678,
				"name":       "测试玩家",
				"userId":     "7486651357831879460",
				"leaderCard": map[string]any{"cardId": 1164, "level": 60, "masterRank": 5, "defaultImage": "special_training", "characterId": 20},
			}},
		})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, Region: "cn"})
	latest, err := client.GetLatest()
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/public/cn/latest" {
		t.Fatalf("path = %q", gotPath)
	}
	if latest.EventID != 165 || len(latest.Rankings) != 1 || latest.Rankings[0].Score != 12345678 {
		t.Fatalf("latest = %+v", latest)
	}
	if latest.Rankings[0].UserID.String() != "7486651357831879460" || latest.Rankings[0].LeaderCard == nil || latest.Rankings[0].LeaderCard.CardID != 1164 || latest.Rankings[0].LeaderCard.CharacterID != 20 {
		t.Fatalf("ranking = %+v", latest.Rankings[0])
	}
}

func TestClientStillParsesNumericUserID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"event_id":   165,
			"region":     "cn",
			"updated_at": 1777744092015,
			"rankings": []map[string]any{{
				"rank":   1,
				"score":  1,
				"name":   "旧响应",
				"userId": 123456789,
			}},
		})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, Region: "cn"})
	latest, err := client.GetLatest()
	if err != nil {
		t.Fatal(err)
	}
	if latest.Rankings[0].UserID.String() != "123456789" {
		t.Fatalf("user id = %q", latest.Rankings[0].UserID.String())
	}
}

func TestClientGetChurnParsesActivityFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"event_id":   165,
			"region":     "cn",
			"updated_at": 1777744092442,
			"rankings": []map[string]any{{
				"rank":            100,
				"score":           12345678,
				"name":            "测试玩家",
				"userId":          "1",
				"churn_48h":       42,
				"last_change":     map[string]any{"time": 1777744017012, "old_score": 12000000, "new_score": 12345678, "delta": 345678},
				"recent_activity": map[string]any{"count": 7},
			}},
		})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, Region: "cn"})
	churn, err := client.GetChurn()
	if err != nil {
		t.Fatal(err)
	}
	entry := churn.Rankings[0]
	if entry.Churn48h != 42 || entry.RecentActivity == nil || entry.RecentActivity.Count != 7 || entry.LastChange == nil || entry.LastChange.Delta != 345678 {
		t.Fatalf("entry = %+v", entry)
	}
}

func TestClientGetWorldLinkLatestAndChurn(t *testing.T) {
	paths := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.String())
		switch r.URL.Path {
		case "/api/public/jp/worldlink-latest":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"event_id":   203,
				"region":     "jp",
				"updated_at": 1,
				"groups": []map[string]any{{
					"event_id":          203,
					"region":            "jp",
					"game_character_id": 1,
					"updated_at":        1,
					"rankings":          []map[string]any{{"rank": 1, "score": 100, "name": "miku", "userId": "42"}},
				}},
			})
		case "/api/public/jp/worldlink-churn":
			if r.URL.Query().Get("gameCharacterId") != "1" {
				t.Fatalf("query = %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"event_id":   203,
				"region":     "jp",
				"board_type": "worldlink",
				"target_id":  1,
				"updated_at": 1,
				"rankings":   []map[string]any{{"rank": 1, "score": 100, "name": "miku", "userId": "42", "churn_48h": 9}},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, Region: "jp"})
	wl, err := client.GetWorldLinkLatest()
	if err != nil {
		t.Fatal(err)
	}
	if len(wl.Groups) != 1 || wl.Groups[0].GameCharacterID != 1 || wl.Groups[0].Rankings[0].UserID.String() != "42" {
		t.Fatalf("wl = %+v", wl)
	}
	churn, err := client.GetWorldLinkChurn(1)
	if err != nil {
		t.Fatal(err)
	}
	if churn.BoardType != "worldlink" || churn.TargetID != 1 || churn.Rankings[0].Churn48h != 9 {
		t.Fatalf("churn = %+v", churn)
	}
	if len(paths) != 2 {
		t.Fatalf("paths = %+v", paths)
	}
}

func TestClientGetWorldLinkLatestNoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "no worldlink data available", "region": "cn"})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, Region: "cn"})
	_, err := client.GetWorldLinkLatest()
	if !errors.Is(err, ErrNoWorldLinkData) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientGetForecastLatest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/public/events":
			if r.URL.Query().Get("region") != "cn" {
				t.Fatalf("events query = %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode([]map[string]any{{"event_id": 165, "name": "活动", "event_type": "marathon", "status": "active", "has_realtime_data": true}})
		case "/public/event/165/latest":
			if r.URL.Query().Get("region") != "cn" {
				t.Fatalf("latest query = %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"event_id":   165,
				"status":     "active",
				"updated_at": "2026-05-04T11:00:00Z",
				"items": []map[string]any{{
					"rank":         100,
					"score":        123,
					"prediction":   "456",
					"collect_time": "2026-05-04T11:00:00Z",
					"is_final":     false,
				}},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: "https://rks.example", Region: "cn"})
	client.forecastURL = server.URL + "/public"
	events, err := client.GetForecastEvents()
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].EventID != 165 {
		t.Fatalf("events = %+v", events)
	}
	board, err := client.GetForecastLatest(165)
	if err != nil {
		t.Fatal(err)
	}
	if board.EventID != 165 || board.Region != "cn" || board.UpdatedUnixMilli() == 0 {
		t.Fatalf("board = %+v", board)
	}
	if score, ok := board.Items[0].PredictedScore(); !ok || score != 456 {
		t.Fatalf("prediction = %d %v", score, ok)
	}
}

func TestClientForecastUnsupportedRegion(t *testing.T) {
	client := NewClient(Config{Region: "en"})
	_, err := client.GetForecastEvents()
	if !errors.Is(err, ErrForecastUnsupportedRegion) {
		t.Fatalf("err = %v", err)
	}
}

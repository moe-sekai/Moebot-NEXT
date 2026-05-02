package ranking

import (
	"encoding/json"
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
				"userId":     7486651357831879460,
				"leaderCard": map[string]any{"cardId": 1164, "level": 60, "masterRank": 5, "defaultImage": "special_training"},
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
	if latest.Rankings[0].UserID != "7486651357831879460" || latest.Rankings[0].LeaderCard == nil || latest.Rankings[0].LeaderCard.CardID != 1164 {
		t.Fatalf("ranking = %+v", latest.Rankings[0])
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
				"userId":          1,
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

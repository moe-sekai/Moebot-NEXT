package renderpayloads

import (
	"testing"
	"time"

	"moebot-next/internal/plugins/moesekai/ranking"
)

func TestBuildRankingListPayloadMapsRankingEntry(t *testing.T) {
	payload := BuildRankingListPayload("活动榜线", ranking.Board{
		EventID:   165,
		Region:    "cn",
		UpdatedAt: 1777744092015,
		Rankings: []ranking.RankingEntry{{
			Rank:       100,
			Score:      12345678,
			Name:       "测试玩家",
			Word:       "签名",
			LeaderCard: &ranking.LeaderCard{CardID: 1164, Level: 60, MasterRank: 5, DefaultImage: "special_training"},
		}},
	})

	if payload.Title != "活动榜线" || payload.EventID != 165 || len(payload.Rankings) != 1 {
		t.Fatalf("payload = %+v", payload)
	}
	entry := payload.Rankings[0]
	if entry.Rank != 100 || entry.Score != 12345678 || entry.Name != "测试玩家" || entry.Signature != "签名" {
		t.Fatalf("entry = %+v", entry)
	}
	if entry.LeaderCard == nil || entry.LeaderCard.CardID != 1164 || !entry.LeaderCard.IsTrained || entry.LeaderCard.Mastery != 5 {
		t.Fatalf("leader = %+v", entry.LeaderCard)
	}
}

func TestBuildChurnRankingListPayloadMapsDelta(t *testing.T) {
	payload := BuildChurnRankingListPayload(ranking.Board{
		EventID: 165,
		Region:  "cn",
		Rankings: []ranking.RankingEntry{{
			Rank:           100,
			Score:          12345678,
			Name:           "测试玩家",
			Churn48h:       42,
			LastChange:     &ranking.LastChange{Delta: 345678},
			RecentActivity: &ranking.RecentActivity{Count: 7},
		}},
	})
	entry := payload.Rankings[0]
	if entry.ScoreDelta != 345678 || entry.Signature == "" {
		t.Fatalf("entry = %+v", entry)
	}
}

func TestBuildChurnRankingListPayloadCalculatesSpeedAndTrend(t *testing.T) {
	now := time.Now().UnixMilli()
	payload := BuildChurnRankingListPayload(ranking.Board{
		EventID: 165,
		Region:  "cn",
		Rankings: []ranking.RankingEntry{{
			Rank:           100,
			Score:          12345678,
			Name:           "测试玩家",
			Churn48h:       42,
			RecentActivity: &ranking.RecentActivity{Count: 7},
			Growth1h:       300000,
			RecentScoreChanges: []ranking.ScoreChange{
				{Time: now - 10*60*1000, Delta: 80000},
				{Time: now - 30*60*1000, Delta: 100000},
				{Time: now - 70*60*1000, Delta: 90000},
			},
		}},
	})

	entry := payload.Rankings[0]
	if entry.Churn48h != 42 || entry.RecentActivityCount != 7 || entry.Growth1h != 300000 {
		t.Fatalf("entry stats = %+v", entry)
	}
	if entry.Churn1h != 2 || entry.Churn20m3 != 3 || entry.Speed20m3 != 240000 || entry.Trend != "down" {
		t.Fatalf("entry speed = %+v", entry)
	}
}

func TestBuildChurnRankingListPayloadMarksTierLine(t *testing.T) {
	payload := BuildChurnRankingListPayload(ranking.Board{
		EventID: 165,
		Region:  "cn",
		Rankings: []ranking.RankingEntry{{
			Rank:     1000,
			Score:    2345678,
			Name:     "TOP1000",
			Growth1h: 120000,
		}},
	})

	entry := payload.Rankings[0]
	if !entry.IsTierLine {
		t.Fatalf("expected tier line entry, got %+v", entry)
	}
}

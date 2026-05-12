package database

import (
	"path/filepath"
	"testing"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/models"
)

func TestEnsureGroupSeparatesOneBotClients(t *testing.T) {
	db, err := New(config.DatabaseConfig{Path: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.EnsureGroup("onebot", "10001", "20001", ""); err != nil {
		t.Fatal(err)
	}
	if err := db.EnsureGroup("onebot", "10002", "20001", ""); err != nil {
		t.Fatal(err)
	}
	if err := db.EnsureGroup("onebot", "10001", "20001", "群备注"); err != nil {
		t.Fatal(err)
	}

	groups, total, err := db.ListGroups(0, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(groups) != 2 {
		t.Fatalf("groups total=%d len=%d, want 2", total, len(groups))
	}

	clientID := "10001"
	groups, total, err = db.ListGroups(0, 10, &clientID)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(groups) != 1 {
		t.Fatalf("client groups total=%d len=%d, want 1", total, len(groups))
	}
	if groups[0].ClientID != "10001" || groups[0].GroupID != "20001" || groups[0].Name != "群备注" {
		t.Fatalf("group = %+v", groups[0])
	}
}

func TestCommandStatsClientFilterAndGrouping(t *testing.T) {
	db, err := New(config.DatabaseConfig{Path: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now()
	rows := []models.CommandStat{
		{Command: "Best30", Platform: "onebot", ClientID: "10001", UserID: "30001", GroupID: "20001", ResponseMs: 100, CreatedAt: now},
		{Command: "Best30", Platform: "onebot", ClientID: "10002", UserID: "30001", GroupID: "20001", ResponseMs: 200, CreatedAt: now},
		{Command: "绑定", Platform: "onebot", ClientID: "10001", UserID: "30002", GroupID: "20002", ResponseMs: 300, CreatedAt: now},
	}
	for i := range rows {
		if err := db.RecordCommandStat(&rows[i]); err != nil {
			t.Fatal(err)
		}
	}

	since := now.Add(-time.Hour)
	byClient, err := db.GetCommandStatsByClient(since, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(byClient) != 2 {
		t.Fatalf("byClient len=%d, want 2: %+v", len(byClient), byClient)
	}
	counts := map[string]int64{}
	for _, row := range byClient {
		counts[row.ClientID] = row.Count
	}
	if counts["10001"] != 2 || counts["10002"] != 1 {
		t.Fatalf("client counts = %+v", counts)
	}

	clientID := "10001"
	totals, err := db.GetCommandStatsTotals(since, &clientID)
	if err != nil {
		t.Fatal(err)
	}
	if totals.Calls != 2 || totals.DistinctGroups != 2 || totals.DistinctUsers != 2 {
		t.Fatalf("totals = %+v", totals)
	}

	byGroup, err := db.GetCommandStatsByGroup(since, nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(byGroup) != 3 {
		t.Fatalf("byGroup len=%d, want 3: %+v", len(byGroup), byGroup)
	}
	seen := map[string]bool{}
	for _, row := range byGroup {
		seen[row.ClientID+"|"+row.GroupID] = true
	}
	for _, key := range []string{"10001|20001", "10002|20001", "10001|20002"} {
		if !seen[key] {
			t.Fatalf("missing group bucket %s in %+v", key, byGroup)
		}
	}
}

func TestBackfillGroupsFromCommandStats(t *testing.T) {
	db, err := New(config.DatabaseConfig{Path: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now()
	rows := []models.CommandStat{
		{Command: "Best30", Platform: "onebot", ClientID: "10001", UserID: "30001", GroupID: "20001", ResponseMs: 100, CreatedAt: now},
		{Command: "Best30", Platform: "onebot", ClientID: "10002", UserID: "30001", GroupID: "20001", ResponseMs: 100, CreatedAt: now},
		{Command: "私聊", Platform: "onebot", ClientID: "10002", UserID: "30001", GroupID: "", ResponseMs: 100, CreatedAt: now},
	}
	for i := range rows {
		if err := db.RecordCommandStat(&rows[i]); err != nil {
			t.Fatal(err)
		}
	}
	if err := backfillClientAwareGroups(db.DB); err != nil {
		t.Fatal(err)
	}

	groups, total, err := db.ListGroups(0, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(groups) != 2 {
		t.Fatalf("groups total=%d len=%d, want 2: %+v", total, len(groups), groups)
	}
}

package database

import (
	"path/filepath"
	"testing"

	"moebot-next/internal/config"
	"moebot-next/internal/models"
)

func TestSuiteSettingUpsertAndLookup(t *testing.T) {
	db, err := New(config.DatabaseConfig{Path: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	setting := &models.SuiteSetting{
		Platform:     "onebot",
		PlatformID:   "10001",
		ServerRegion: config.RegionCN,
		Mode:         config.SuiteModeHaruki,
		Hidden:       true,
	}
	if err := db.UpsertSuiteSetting(setting); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetSuiteSetting("onebot", "10001", config.RegionCN)
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != config.SuiteModeHaruki || !got.Hidden {
		t.Fatalf("setting = %+v", got)
	}

	got.Mode = config.SuiteModeLocal
	got.Hidden = false
	if err := db.UpsertSuiteSetting(got); err != nil {
		t.Fatal(err)
	}

	updated, err := db.GetSuiteSetting("onebot", "10001", config.RegionCN)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Mode != config.SuiteModeLocal || updated.Hidden {
		t.Fatalf("updated = %+v", updated)
	}
}

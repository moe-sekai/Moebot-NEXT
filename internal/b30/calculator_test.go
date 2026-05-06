package b30

import (
	"math"
	"testing"
)

func TestUserRatingFormula(t *testing.T) {
	if got := UserRating(33.2, PlayResultAP); got != 33.2 {
		t.Fatalf("AP rating = %.1f, want 33.2", got)
	}
	if got := UserRating(33.0, PlayResultFC); got != 32.0 {
		t.Fatalf("FC rating >=33 = %.1f, want 32.0", got)
	}
	if got := UserRating(32.9, PlayResultFC); got != 31.4 {
		t.Fatalf("FC rating <33 = %.1f, want 31.4", got)
	}
}

func TestCalculateSkipsClearAndDedupesBestResult(t *testing.T) {
	table := NewConstantsTable([]ChartConstant{{MusicID: 1, Difficulty: "master", Constant: 33.5, Level: 34, NoteCount: 1200}})
	result := Calculate([]UserMusicResult{
		{MusicID: 1, MusicDifficultyType: "master", PlayResult: "clear"},
		{MusicID: 1, MusicDifficultyType: "master", PlayResult: "full_combo"},
		{MusicID: 1, MusicDifficultyType: "master", PlayResult: "all_perfect"},
	}, table, nil)
	if len(result.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(result.Entries))
	}
	entry := result.Entries[0]
	if entry.PlayResult != PlayResultAP || entry.UserRating != 33.5 {
		t.Fatalf("entry = %#v, want AP rating 33.5", entry)
	}
	if result.APCount != 1 || result.FCCount != 0 || result.CandidateCount != 1 {
		t.Fatalf("counts = AP %d FC %d candidate %d", result.APCount, result.FCCount, result.CandidateCount)
	}
}

func TestCalculateClearDoesNotEnterBest30(t *testing.T) {
	table := NewConstantsTable([]ChartConstant{{MusicID: 1, Difficulty: "expert", Constant: 30.5}})
	result := Calculate([]UserMusicResult{{MusicID: 1, MusicDifficultyType: "expert", PlayResult: "clear"}}, table, nil)
	if len(result.Entries) != 0 || result.CandidateCount != 0 {
		t.Fatalf("result = %#v, want no entries", result)
	}
}

func TestCalculateLimitsTop30AndAveragesActualEntries(t *testing.T) {
	constants := make([]ChartConstant, 0, 31)
	results := make([]UserMusicResult, 0, 31)
	for i := 0; i < 31; i++ {
		musicID := i + 1
		constants = append(constants, ChartConstant{MusicID: musicID, Difficulty: "master", Constant: 40 - float64(i)/10})
		results = append(results, UserMusicResult{MusicID: musicID, MusicDifficultyType: "MASTER", PlayResult: "all_perfect"})
	}
	result := Calculate(results, NewConstantsTable(constants), nil)
	if len(result.Entries) != 30 {
		t.Fatalf("entries = %d, want 30", len(result.Entries))
	}
	if result.Entries[0].MusicID != 1 || result.Entries[29].MusicID != 30 {
		t.Fatalf("entry order first=%d last=%d", result.Entries[0].MusicID, result.Entries[29].MusicID)
	}
	wantAverage := 0.0
	for i := 0; i < 30; i++ {
		wantAverage += 40 - float64(i)/10
	}
	wantAverage /= 30
	if math.Abs(result.Average-wantAverage) > 1e-9 {
		t.Fatalf("average = %.6f, want %.6f", result.Average, wantAverage)
	}

	short := Calculate(results[:2], NewConstantsTable(constants[:2]), nil)
	if len(short.Entries) != 2 {
		t.Fatalf("short entries = %d, want 2", len(short.Entries))
	}
	if math.Abs(short.Average-39.95) > 1e-9 {
		t.Fatalf("short average = %.6f, want 39.95", short.Average)
	}
}

func TestMergeLegacyResultsFillsMissingChartsOnly(t *testing.T) {
	merged := MergeLegacyResults(
		[]UserMusicResult{{MusicID: 1, MusicDifficultyType: "master", PlayResult: "full_combo"}},
		[]LegacyUserMusic{{
			MusicID: 1,
			UserMusicResults: []LegacyMusicResult{
				{MusicID: 1, MusicDifficultyType: "master", PlayResult: "all_perfect"},
				{MusicID: 1, MusicDifficultyType: "expert", PlayResult: "full_combo"},
			},
		}},
	)
	if len(merged) != 2 {
		t.Fatalf("merged = %d, want 2: %#v", len(merged), merged)
	}
	if NormalizePlayResult(merged[0]) != PlayResultFC {
		t.Fatalf("primary chart should not be overwritten: %#v", merged[0])
	}
	if merged[1].MusicDifficultyType != "expert" {
		t.Fatalf("legacy missing chart = %#v, want expert", merged[1])
	}
}

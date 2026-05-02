package commands

import (
	"strings"
	"testing"

	"moebot-next/internal/renderer"
)

func TestBuildChartRenderRequestUsesChartTemplate(t *testing.T) {
	req := buildChartRenderRequest(renderer.MusicDetailPayload{ID: 74, Title: "独りんぼエンヴィー"})
	if req.Template != "chart_detail" {
		t.Fatalf("Template = %q, want chart_detail", req.Template)
	}
}

func TestFormatChartTextShowsChartFocusedDifficultyRows(t *testing.T) {
	text := formatChartText(renderer.MusicDetailPayload{
		ID:         88,
		Title:      "テスト楽曲",
		Categories: []string{"mv", "original"},
		Difficulties: []renderer.MusicDifficultyPayload{
			{MusicDifficulty: "easy", PlayLevel: 5, TotalNoteCount: 123},
			{MusicDifficulty: "master", PlayLevel: 30, TotalNoteCount: 1000},
		},
	})

	checks := []string{
		"谱面：テスト楽曲",
		"ID：88",
		"分类：mv / original",
		"EASY：Lv.5 · 123 notes",
		"MASTER：Lv.30 · 1000 notes",
	}
	for _, want := range checks {
		if !strings.Contains(text, want) {
			t.Fatalf("formatChartText() missing %q in:\n%s", want, text)
		}
	}
}

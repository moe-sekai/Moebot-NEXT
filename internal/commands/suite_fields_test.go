package commands

import (
	"testing"

	"moebot-next/internal/suite"
)

func TestSuitePanelFieldsIncludeDecksAndCards(t *testing.T) {
	cases := map[string][]string{
		"bond":           bondFields(),
		"material":       materialFields(),
		"leader_count":   leaderCountFields(),
		"music_overview": musicOverviewFields(),
		"best30":         best30Fields(),
	}
	for name, fields := range cases {
		if !containsField(fields, suite.FieldUserDecks) || !containsField(fields, suite.FieldUserCards) {
			t.Fatalf("%s fields should include deck/card fields, got %#v", name, fields)
		}
	}
}

func TestBest30FieldsIncludeMusicResultCompatibilitySources(t *testing.T) {
	fields := best30Fields()
	for _, want := range []string{suite.FieldUserMusicResults, suite.FieldUserMusics} {
		if !containsField(fields, want) {
			t.Fatalf("best30 fields = %#v, missing %s", fields, want)
		}
	}
}

func containsField(fields []string, want string) bool {
	for _, field := range fields {
		if field == want {
			return true
		}
	}
	return false
}

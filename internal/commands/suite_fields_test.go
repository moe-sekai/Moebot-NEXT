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
	}
	for name, fields := range cases {
		if !containsField(fields, suite.FieldUserDecks) || !containsField(fields, suite.FieldUserCards) {
			t.Fatalf("%s fields should include deck/card fields, got %#v", name, fields)
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

package commands

import (
	"strings"
	"testing"
)

func TestHelpTextIncludesQueryAliasesAndChartCommand(t *testing.T) {
	text := helpText()
	checks := []string{"/查歌", "/查谱", "/查扭蛋"}
	for _, want := range checks {
		if !strings.Contains(text, want) {
			t.Fatalf("helpText() missing %q in:\n%s", want, text)
		}
	}
}

package commands

import (
	"strings"
	"testing"

	"moebot-next/internal/renderer"
)

func TestFormatProfileTextShowsAPIProfile(t *testing.T) {
	text := formatProfileText(renderer.ProfileCardPayload{
		Name:       "测试用户",
		Rank:       321,
		UserID:     "7485966462906096424",
		Signature:  "你好 SEKAI",
		TotalPower: 123456,
	})
	for _, want := range []string{"测试用户", "Rank：321", "7485966462906096424", "你好 SEKAI", "123,456"} {
		if !strings.Contains(text, want) {
			t.Fatalf("formatProfileText() missing %q in:\n%s", want, text)
		}
	}
}

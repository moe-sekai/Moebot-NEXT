package commands

import (
	"strings"
	"testing"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

func TestFormatGachaTextUsesChineseGachaTypeLabels(t *testing.T) {
	text := formatGachaText(renderpayloads.GachaInfoPayload{
		ID:        42,
		Name:      "测试扭蛋",
		GachaType: "birthday",
	})

	if !strings.Contains(text, "类型：生日扭蛋") {
		t.Fatalf("formatGachaText() should use Chinese gacha type label, got:\n%s", text)
	}
}

package renderer

import (
	"testing"

	"moebot-next/internal/masterdata"
)

func TestBuildMusicDetailPayloadIncludesDuration(t *testing.T) {
	payload := BuildMusicDetailPayload(nil, masterdata.MusicInfo{
		ID:                    74,
		Title:                 "独りんぼエンヴィー",
		SecForMusicScoreMaker: 74,
		FillerSec:             9,
	})

	if payload.DurationSec != 74 {
		t.Fatalf("DurationSec = %v, want 74", payload.DurationSec)
	}
}

func TestCardSupplyTypeDisplayName(t *testing.T) {
	cases := map[string]string{
		"normal":                     "常驻",
		"birthday":                   "生日",
		"term_limited":               "期间限定",
		"colorful_festival_limited":  "CFES限定",
		"bloom_festival_limited":     "BFES限定",
		"unit_event_limited":         "WorldLink限定",
		"collaboration_limited":      "联动限定",
		"":                           "常驻",
	}

	for supplyType, want := range cases {
		if got := CardSupplyTypeDisplayName(supplyType); got != want {
			t.Fatalf("CardSupplyTypeDisplayName(%q) = %q, want %q", supplyType, got, want)
		}
	}
}

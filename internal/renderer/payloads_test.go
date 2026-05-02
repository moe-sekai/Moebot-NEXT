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

package commands

import (
	"testing"

	"moebot-next/internal/config"
)

func TestParseRegionalCommandName(t *testing.T) {
	cases := []struct {
		input      string
		wantBase   string
		wantRegion string
	}{
		{input: "绑定", wantBase: "绑定", wantRegion: ""},
		{input: "cn绑定", wantBase: "绑定", wantRegion: config.RegionCN},
		{input: "tw查曲", wantBase: "查曲", wantRegion: config.RegionTW},
		{input: "kr榜线", wantBase: "榜线", wantRegion: config.RegionKR},
		{input: "en个人信息", wantBase: "个人信息", wantRegion: config.RegionEN},
	}
	for _, tc := range cases {
		base, region := parseRegionalCommandName(tc.input)
		if base != tc.wantBase || region != tc.wantRegion {
			t.Fatalf("parseRegionalCommandName(%q) = (%q,%q), want (%q,%q)", tc.input, base, region, tc.wantBase, tc.wantRegion)
		}
	}
}

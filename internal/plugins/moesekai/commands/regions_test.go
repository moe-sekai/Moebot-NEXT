package commands

import (
	"testing"

	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/commandparser"
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

// TestParserCommandsLongerAliasesRegisterFirst guards against regressions where
// a shorter alias (e.g. "谱面") would shadow a longer one (e.g. "谱面预览")
// because ZeroBot's OnCommand uses HasPrefix matching in registration order.
// The longer alias must appear earlier in the slice.
func TestParserCommandsLongerAliasesRegisterFirst(t *testing.T) {
	cmds := parserCommands(&Deps{Definitions: commandparser.BaseDefinitions()}, "查谱")
	indexOf := func(name string) int {
		for i, cmd := range cmds {
			if cmd.Name == name {
				return i
			}
		}
		return -1
	}

	pairs := [][2]string{
		{"谱面预览", "谱面"},
		{"cn谱面预览", "cn谱面"},
		{"jp谱面预览", "jp谱面"},
	}
	for _, p := range pairs {
		long, short := indexOf(p[0]), indexOf(p[1])
		if long < 0 || short < 0 {
			t.Fatalf("expected both %q and %q to be registered, got idx=%d/%d in %#v", p[0], p[1], long, short, cmds)
		}
		if long > short {
			t.Fatalf("%q must register before %q to win HasPrefix matching, got idx=%d>%d", p[0], p[1], long, short)
		}
	}
}

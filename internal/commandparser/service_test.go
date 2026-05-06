package commandparser

import (
	"testing"
	"time"

	"moebot-next/internal/assets"
	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/renderer"
)

func TestParseCardPrimaryAndPresetAliases(t *testing.T) {
	service := NewService("/", nil, nil, nil, nil)

	primary := service.Parse("/查卡 1204")
	if primary.Definition == nil || primary.Definition.ID != "card-detail" {
		t.Fatalf("/查卡 parsed definition = %#v, want card-detail", primary.Definition)
	}
	if primary.Argument != "1204" || primary.MatchSource != MatchPrimary {
		t.Fatalf("primary parse = argument %q source %q", primary.Argument, primary.MatchSource)
	}

	alias := service.Parse("cardinfo 初音")
	if alias.Definition == nil || alias.Definition.PrimaryCommand != "查卡" {
		t.Fatalf("cardinfo parsed definition = %#v, want 查卡", alias.Definition)
	}
	if alias.MatchSource != MatchPresetAlias {
		t.Fatalf("alias source = %q, want %q", alias.MatchSource, MatchPresetAlias)
	}
}

func TestParseCustomAliasAndRegion(t *testing.T) {
	service := NewService("/", map[string][]string{"查卡": []string{"卡牌详情"}}, nil, nil, nil)
	parsed := service.Parse("/cn卡牌详情 1204")
	if parsed.Definition == nil || parsed.Definition.ID != "card-detail" {
		t.Fatalf("custom alias parsed definition = %#v", parsed.Definition)
	}
	if parsed.MatchSource != MatchCustomAlias {
		t.Fatalf("source = %q, want custom_alias", parsed.MatchSource)
	}
	if parsed.Region != config.RegionCN {
		t.Fatalf("region = %q, want cn", parsed.Region)
	}
}

func TestSearchAndBuildCardListParsesLunabotStyleFilters(t *testing.T) {
	store := masterdata.NewStore()
	now := time.Now().UnixMilli()
	store.SetAll(&masterdata.MasterData{Cards: []masterdata.CardInfo{
		{ID: 10, CharacterID: 9, CardRarityType: "rarity_4", Attr: "cute", Prefix: "心羽四星", ReleaseAt: now, SupportUnit: "none"},
		{ID: 11, CharacterID: 9, CardRarityType: "rarity_3", Attr: "cute", Prefix: "心羽三星", ReleaseAt: now - 1, SupportUnit: "none"},
		{ID: 20, CharacterID: 5, CardRarityType: "rarity_4", Attr: "cool", Prefix: "实乃理限定", ReleaseAt: now, SupportUnit: "none", CardSupplyID: 1},
		{ID: 21, CharacterID: 5, CardRarityType: "rarity_4", Attr: "cool", Prefix: "实乃理常驻", ReleaseAt: now - 1, SupportUnit: "none"},
	}, CardSupplies: []masterdata.CardSupplyInfo{{ID: 1, CardSupplyType: "term_limited"}}})

	rows, selected := searchAndBuild(BaseDefinitions()[0], store, nil, nil, "", "khn 四星")
	if selected == nil || selected.Type != "card_list" {
		t.Fatalf("selected = %#v, want card_list", selected)
	}
	if len(rows) != 1 || rows[0].ID != 10 {
		t.Fatalf("rows = %#v, want only khn rarity_4", rows)
	}

	rows, selected = searchAndBuild(BaseDefinitions()[0], store, nil, nil, "", "mnr 4 蓝 限定")
	if selected == nil || selected.Type != "card_list" {
		t.Fatalf("selected = %#v, want card_list", selected)
	}
	if len(rows) != 1 || rows[0].ID != 20 {
		t.Fatalf("rows = %#v, want only limited cool mnr 4*", rows)
	}
}

func TestParseSuiteStatusAliases(t *testing.T) {
	service := NewService("/", nil, nil, nil, nil)
	for _, input := range []string{"/抓包数据", "/抓包信息", "/suite", "/cn抓包数据", "/cnsuite"} {
		parsed := service.Parse(input)
		if parsed.Definition == nil || parsed.Definition.ID != "suite-status" {
			t.Fatalf("%s parsed definition = %#v, want suite-status", input, parsed.Definition)
		}
	}
}

func TestSuiteVisibilityCommandsAreActionOnly(t *testing.T) {
	for _, def := range BaseDefinitions() {
		if def.ID != "suite-hide" && def.ID != "suite-show" {
			continue
		}
		if def.Template != "" || def.PreviewID != "" || def.RenderMode != RenderModeAction {
			t.Fatalf("%s should be action-only, got template=%q preview=%q render=%q", def.ID, def.Template, def.PreviewID, def.RenderMode)
		}
	}
}

func TestParseMusicRewardAliasesUseMusicProgressDefinition(t *testing.T) {
	service := NewService("/", nil, nil, nil, nil)
	for _, input := range []string{"/打歌进度", "/歌曲奖励", "/cn歌曲奖励", "/打歌挖矿"} {
		parsed := service.Parse(input)
		if parsed.Definition == nil || parsed.Definition.ID != "music-progress" {
			t.Fatalf("%s parsed definition = %#v, want music-progress", input, parsed.Definition)
		}
	}
}

func TestParseBest30Aliases(t *testing.T) {
	service := NewService("/", nil, nil, nil, nil)
	for _, input := range []string{"/b30", "/best30", "/B30", "/最佳30", "/cnb30"} {
		parsed := service.Parse(input)
		if parsed.Definition == nil || parsed.Definition.ID != "best30" {
			t.Fatalf("%s parsed definition = %#v, want best30", input, parsed.Definition)
		}
		if !parsed.RequiresBinding || parsed.BindingKind != "suite" || parsed.PreviewFallbackAvailable != true {
			t.Fatalf("%s binding/render flags = requires %v kind %q preview %v", input, parsed.RequiresBinding, parsed.BindingKind, parsed.PreviewFallbackAvailable)
		}
	}
}

func TestRemovedSuiteCommandsDoNotParse(t *testing.T) {
	service := NewService("/", nil, nil, nil, nil)
	for _, input := range []string{"/抽卡记录", "/cn抽卡统计", "/材料信息", "/jp素材", "/materials"} {
		parsed := service.Parse(input)
		if parsed.Definition != nil {
			t.Fatalf("%s parsed definition = %#v, want nil", input, parsed.Definition)
		}
	}
}

func TestParseEventRecordRankingAlias(t *testing.T) {
	service := NewService("/", nil, nil, nil, nil)
	for _, input := range []string{"/冲榜记录", "/cn冲榜记录"} {
		parsed := service.Parse(input)
		if parsed.Definition == nil || parsed.Definition.ID != "event-record" {
			t.Fatalf("%s parsed definition = %#v, want event-record", input, parsed.Definition)
		}
	}
}

func TestParseInlineRankingTargetArgument(t *testing.T) {
	service := NewService("/", nil, nil, nil, nil)
	parsed := service.Parse("/cnsk1-10")
	if parsed.Definition == nil || parsed.Definition.ID != "ranking-target" {
		t.Fatalf("definition = %+v", parsed.Definition)
	}
	if parsed.CommandText != "cnsk" || parsed.Argument != "1-10" || parsed.Region != config.RegionCN {
		t.Fatalf("parsed command = %q argument = %q region = %q", parsed.CommandText, parsed.Argument, parsed.Region)
	}
}

func TestParseInlineForecastDoesNotMatchSk(t *testing.T) {
	service := NewService("/", nil, nil, nil, nil)
	parsed := service.Parse("/skp165")
	if parsed.Definition == nil || parsed.Definition.ID != "forecast-ranking" {
		t.Fatalf("definition = %+v", parsed.Definition)
	}
	if parsed.CommandText != "skp" || parsed.Argument != "165" {
		t.Fatalf("parsed command = %q argument = %q", parsed.CommandText, parsed.Argument)
	}
}

func TestParseInlineSkLineDoesNotMatchSk(t *testing.T) {
	service := NewService("/", nil, nil, nil, nil)
	parsed := service.Parse("/sk线100")
	if parsed.Definition == nil || parsed.Definition.ID != "ranking-list" {
		t.Fatalf("definition = %+v", parsed.Definition)
	}
	if parsed.CommandText != "sk线" || parsed.Argument != "100" {
		t.Fatalf("parsed command = %q argument = %q", parsed.CommandText, parsed.Argument)
	}
}

func TestParseMusicInfoPresetAliases(t *testing.T) {
	service := NewService("/", nil, nil, nil, nil)
	for _, input := range []string{"/songinfo 谷歌", "/musicinfo 谷歌", "/song 谷歌", "/music 谷歌"} {
		parsed := service.Parse(input)
		if parsed.Definition == nil || parsed.Definition.ID != "music-detail" {
			t.Fatalf("%s parsed definition = %#v, want music-detail", input, parsed.Definition)
		}
		if parsed.MatchSource != MatchPresetAlias {
			t.Fatalf("%s source = %q, want preset alias", input, parsed.MatchSource)
		}
	}
}

func TestSearchAndBuildMusicAliasListPayload(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{Musics: []masterdata.MusicInfo{
		{ID: 1, Title: "Alpha"},
		{ID: 2, Title: "Beta"},
	}, MusicDifficulties: []masterdata.MusicDifficulty{
		{MusicID: 1, MusicDifficulty: "master", PlayLevel: 30},
		{MusicID: 2, MusicDifficulty: "master", PlayLevel: 31},
	}})
	aliases := map[int]assets.MusicAlias{
		1: {MusicID: 1, Aliases: []string{"同名别名"}},
		2: {MusicID: 2, Aliases: []string{"同名别名"}},
	}
	var def Definition
	for _, candidate := range BaseDefinitions() {
		if candidate.ID == "music-detail" {
			def = candidate
			break
		}
	}

	rows, selected := searchAndBuild(def, store, nil, aliases, "", "同名别名")
	if selected == nil || selected.Type != "music_list" {
		t.Fatalf("selected = %#v, want music_list", selected)
	}
	payload, ok := selected.Payload.(renderer.MusicListPayload)
	if !ok {
		t.Fatalf("payload type = %T, want MusicListPayload", selected.Payload)
	}
	if len(rows) != 2 || payload.Total != 2 || len(payload.Musics) != 2 {
		t.Fatalf("rows=%d total=%d payload musics=%d, want 2/2/2", len(rows), payload.Total, len(payload.Musics))
	}
}

func TestSearchAndBuildChartPayloadUsesChartSource(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{Musics: []masterdata.MusicInfo{{ID: 739, Title: "Chart Song"}}, MusicDifficulties: []masterdata.MusicDifficulty{
		{MusicID: 739, MusicDifficulty: "easy", PlayLevel: 5},
		{MusicID: 739, MusicDifficulty: "normal", PlayLevel: 12},
		{MusicID: 739, MusicDifficulty: "hard", PlayLevel: 18},
		{MusicID: 739, MusicDifficulty: "expert", PlayLevel: 28},
		{MusicID: 739, MusicDifficulty: "master", PlayLevel: 32},
		{MusicID: 739, MusicDifficulty: "append", PlayLevel: 34},
	}})
	var def Definition
	for _, candidate := range BaseDefinitions() {
		if candidate.ID == "chart-detail" {
			def = candidate
			break
		}
	}

	_, selected := searchAndBuild(def, store, nil, nil, "https://charts.example.test/{id}/{difficulty}.svg", "739")
	if selected == nil {
		t.Fatal("selected is nil")
	}
	payload, ok := selected.Payload.(renderer.MusicDetailPayload)
	if !ok {
		t.Fatalf("payload type = %T, want MusicDetailPayload", selected.Payload)
	}
	if payload.ChartURL != "https://charts.example.test/739/master.svg" || payload.SelectedDifficulty != "master" {
		t.Fatalf("chart url=%q selected=%q", payload.ChartURL, payload.SelectedDifficulty)
	}

	for _, tc := range []struct {
		argument string
		wantDiff string
	}{
		{argument: "739 ex", wantDiff: "expert"},
		{argument: "739 expert", wantDiff: "expert"},
		{argument: "739 ma", wantDiff: "master"},
		{argument: "739 mas", wantDiff: "master"},
		{argument: "739 master", wantDiff: "master"},
		{argument: "739 apd", wantDiff: "append"},
		{argument: "739 ap", wantDiff: "append"},
		{argument: "739 append", wantDiff: "append"},
		{argument: "739 hd", wantDiff: "hard"},
		{argument: "739 hard", wantDiff: "hard"},
		{argument: "739 nm", wantDiff: "normal"},
		{argument: "739 normal", wantDiff: "normal"},
		{argument: "739 ez", wantDiff: "easy"},
		{argument: "739 easy", wantDiff: "easy"},
	} {
		_, selected := searchAndBuild(def, store, nil, nil, "https://charts.example.test/{id}/{difficulty}.svg", tc.argument)
		if selected == nil {
			t.Fatalf("%s selected is nil", tc.argument)
		}
		payload, ok := selected.Payload.(renderer.MusicDetailPayload)
		if !ok {
			t.Fatalf("%s payload type = %T, want MusicDetailPayload", tc.argument, selected.Payload)
		}
		wantURL := "https://charts.example.test/739/" + tc.wantDiff + ".svg"
		if payload.SelectedDifficulty != tc.wantDiff || payload.ChartURL != wantURL {
			t.Fatalf("%s selected=%q url=%q, want %s %s", tc.argument, payload.SelectedDifficulty, payload.ChartURL, tc.wantDiff, wantURL)
		}
	}
}

func TestParseChartChinesePresetAliases(t *testing.T) {
	service := NewService("/", nil, nil, nil, nil)
	for _, input := range []string{"/谱面 739", "/谱面预览 master 739"} {
		parsed := service.Parse(input)
		if parsed.Definition == nil || parsed.Definition.ID != "chart-detail" {
			t.Fatalf("%s parsed definition = %#v, want chart-detail", input, parsed.Definition)
		}
		if parsed.MatchSource != MatchPresetAlias {
			t.Fatalf("%s source = %q, want preset alias", input, parsed.MatchSource)
		}
	}
}

func TestValidateAliasesRejectsProtectedAndShortAlias(t *testing.T) {
	if _, _, err := ValidateAliases(map[string][]string{"查卡": []string{"card"}}); err == nil {
		t.Fatal("expected protected preset alias conflict")
	}
	if _, _, err := ValidateAliases(map[string][]string{"查卡": []string{"卡"}}); err == nil {
		t.Fatal("expected short alias to be rejected")
	}
}

func TestAliasConfigResetClearsCustomAliases(t *testing.T) {
	cfg := AliasConfig(map[string][]string{})
	if len(cfg.Custom) != 0 {
		t.Fatalf("custom aliases = %#v, want empty", cfg.Custom)
	}
}

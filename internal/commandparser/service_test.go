package commandparser

import (
	"testing"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
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

	rows, selected := searchAndBuild(BaseDefinitions()[0], store, nil, "khn 四星")
	if selected == nil || selected.Type != "card_list" {
		t.Fatalf("selected = %#v, want card_list", selected)
	}
	if len(rows) != 1 || rows[0].ID != 10 {
		t.Fatalf("rows = %#v, want only khn rarity_4", rows)
	}

	rows, selected = searchAndBuild(BaseDefinitions()[0], store, nil, "mnr 4 蓝 限定")
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

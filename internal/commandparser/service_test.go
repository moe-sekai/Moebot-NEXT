package commandparser

import (
	"testing"

	"moebot-next/internal/config"
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

package commands

import (
	"testing"
	"time"

	"moebot-next/internal/cardquery"
	"moebot-next/internal/masterdata"
)

func TestParseCardQueryPureIDUsesDetail(t *testing.T) {
	q := cardquery.Parse(nil, "1204")
	if q.Mode != cardquery.ModeDetail || q.DetailID != 1204 {
		t.Fatalf("cardquery.Parse(1204) = mode %v id %d, want detail 1204", q.Mode, q.DetailID)
	}
}

func TestParseCardQueryNonPureIDUsesList(t *testing.T) {
	for _, raw := range []string{"card1204", "id1204", "1204 蓝", "mnr-1"} {
		q := cardquery.Parse(nil, raw)
		if q.Mode != cardquery.ModeList {
			t.Fatalf("cardquery.Parse(%q) mode = %v, want list", raw, q.Mode)
		}
	}
}

func TestParseCardQueryAliasesAndJoinedFilters(t *testing.T) {
	q := cardquery.Parse(nil, "mnr4蓝限")
	if q.CharacterID != 5 {
		t.Fatalf("CharacterID = %d, want 5", q.CharacterID)
	}
	if q.Rarity != "rarity_4" {
		t.Fatalf("Rarity = %q, want rarity_4", q.Rarity)
	}
	if q.Attr != "cool" {
		t.Fatalf("Attr = %q, want cool", q.Attr)
	}
	if q.Supply != "all_limited" {
		t.Fatalf("Supply = %q, want all_limited", q.Supply)
	}
}

func TestParseCardQueryUnitModes(t *testing.T) {
	tests := []struct {
		raw  string
		unit string
		mode cardquery.UnitMode
	}{
		{"纯mmj", "idol", cardquery.UnitOC},
		{"mmjv", "idol", cardquery.UnitVS},
		{"mmjoc", "idol", cardquery.UnitOC},
		{"纯v", "piapro", cardquery.UnitOC},
	}
	for _, tt := range tests {
		q := cardquery.Parse(nil, tt.raw)
		if q.Unit != tt.unit || q.UnitMode != tt.mode {
			t.Fatalf("cardquery.Parse(%q) unit/mode = %q/%q, want %q/%q", tt.raw, q.Unit, q.UnitMode, tt.unit, tt.mode)
		}
	}
}

func TestParseCardQueryLongAliasesWin(t *testing.T) {
	tests := []struct {
		raw   string
		check func(cardquery.Query) bool
		label string
	}{
		{"bfes限定", func(q cardquery.Query) bool { return q.Supply == "bloom_festival_limited" }, "bfes"},
		{"蓝星", func(q cardquery.Query) bool { return q.Attr == "cool" && q.Unparsed == "" }, "蓝星"},
		{"生日卡", func(q cardquery.Query) bool { return q.Rarity == "rarity_birthday" && q.Unparsed == "" }, "生日卡"},
		{"p分", func(q cardquery.Query) bool { _, ok := q.DetailSkillIDs[11]; return ok && q.Unparsed == "" }, "p分"},
	}
	for _, tt := range tests {
		q := cardquery.Parse(nil, tt.raw)
		if !tt.check(q) {
			t.Fatalf("%s parsed as %#v", tt.label, q)
		}
	}
}

func TestParseCardQueryDoesNotTreatSpacedNicknameRarityAsBanEvent(t *testing.T) {
	store := masterdata.NewStore()
	now := time.Now().UnixMilli()
	store.SetAll(&masterdata.MasterData{
		Cards: []masterdata.CardInfo{
			{ID: 517, CharacterID: 5, CardRarityType: "rarity_4", Attr: "cool", Prefix: "箱活卡", ReleaseAt: now, SupportUnit: "none"},
		},
		Events:     []masterdata.EventInfo{{ID: 99, StartAt: now}},
		EventCards: []masterdata.EventCard{{EventID: 99, CardID: 517}},
	})

	spaced := cardquery.Parse(store, "mnr 4")
	if spaced.EventID != 0 || spaced.CharacterID != 5 || spaced.Rarity != "rarity_4" {
		t.Fatalf("spaced mnr 4 parsed as %#v, want character+rarity without event", spaced)
	}

	joined := cardquery.Parse(store, "mnr1")
	if joined.EventID != 99 {
		t.Fatalf("joined mnr1 EventID = %d, want 99", joined.EventID)
	}
}

func TestResolveCardQueryJoinedFiltersMatchExpectedCards(t *testing.T) {
	store := masterdata.NewStore()
	now := time.Now().UnixMilli()
	store.SetAll(&masterdata.MasterData{
		Cards: []masterdata.CardInfo{
			{ID: 1, CharacterID: 5, CardRarityType: "rarity_4", Attr: "cool", Prefix: "实乃理限定", ReleaseAt: now, SupportUnit: "none", CardSupplyID: 1},
			{ID: 2, CharacterID: 5, CardRarityType: "rarity_4", Attr: "cool", Prefix: "实乃理常驻", ReleaseAt: now - 1, SupportUnit: "none"},
			{ID: 3, CharacterID: 6, CardRarityType: "rarity_4", Attr: "cool", Prefix: "遥限定", ReleaseAt: now - 2, SupportUnit: "none", CardSupplyID: 1},
		},
		CardSupplies: []masterdata.CardSupplyInfo{{ID: 1, CardSupplyType: "term_limited"}},
	})
	result := cardquery.Resolve(store, "mnr 4 蓝 限定")
	if result.Mode != cardquery.ModeList || len(result.Cards) != 1 || result.Cards[0].ID != 1 {
		t.Fatalf("result = %#v, want only mnr cool 4* limited", result)
	}
}

func TestResolveCardQueryListSingleResultDoesNotBecomeDetail(t *testing.T) {
	store := masterdata.NewStore()
	now := time.Now().UnixMilli()
	store.SetAll(&masterdata.MasterData{Cards: []masterdata.CardInfo{
		{ID: 1, CharacterID: 5, CardRarityType: "rarity_4", Attr: "cool", Prefix: "测试卡", ReleaseAt: now, SupportUnit: "none"},
	}})
	result := cardquery.Resolve(store, "mnr 4 蓝")
	if result.Mode != cardquery.ModeList {
		t.Fatalf("Mode = %v, want list", result.Mode)
	}
	if len(result.Cards) != 1 || result.Cards[0].ID != 1 {
		t.Fatalf("Cards = %#v, want card #1", result.Cards)
	}
}

func TestResolveCardQueryPureIDDetail(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{Cards: []masterdata.CardInfo{{ID: 1204, CharacterID: 21, CardRarityType: "rarity_4", Attr: "cute", Prefix: "详情卡"}}})
	result := cardquery.Resolve(store, "1204")
	if result.Mode != cardquery.ModeDetail {
		t.Fatalf("Mode = %v, want detail", result.Mode)
	}
	if len(result.Cards) != 1 || result.Cards[0].ID != 1204 {
		t.Fatalf("Cards = %#v, want card #1204", result.Cards)
	}
}

func TestCardBoxCardsUsesCardQueryFilters(t *testing.T) {
	store := masterdata.NewStore()
	now := time.Now().UnixMilli()
	store.SetAll(&masterdata.MasterData{
		Cards: []masterdata.CardInfo{
			{ID: 1, CharacterID: 5, CardRarityType: "rarity_4", Attr: "cool", Prefix: "实乃理限定", ReleaseAt: now, SupportUnit: "none", CardSupplyID: 1},
			{ID: 2, CharacterID: 5, CardRarityType: "rarity_4", Attr: "cool", Prefix: "实乃理常驻", ReleaseAt: now - 1, SupportUnit: "none"},
			{ID: 3, CharacterID: 6, CardRarityType: "rarity_4", Attr: "cool", Prefix: "遥限定", ReleaseAt: now - 2, SupportUnit: "none", CardSupplyID: 1},
		},
		CardSupplies: []masterdata.CardSupplyInfo{{ID: 1, CardSupplyType: "term_limited"}},
	})

	cards, msg := cardBoxCards(store, cardBoxQueryOptions{FilterText: "mnr 4 蓝 限定"})
	if msg != "" {
		t.Fatalf("msg = %q, want empty", msg)
	}
	if len(cards) != 1 || cards[0].ID != 1 {
		t.Fatalf("cards = %#v, want only limited cool mnr 4*", cards)
	}
}

func TestCardBoxCardsPureIDReturnsSingleCard(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{Cards: []masterdata.CardInfo{
		{ID: 1204, CharacterID: 21, CardRarityType: "rarity_4", Attr: "cute", Prefix: "详情卡"},
		{ID: 1205, CharacterID: 21, CardRarityType: "rarity_4", Attr: "cool", Prefix: "其它卡"},
	}})

	cards, msg := cardBoxCards(store, cardBoxQueryOptions{FilterText: "1204"})
	if msg != "" {
		t.Fatalf("msg = %q, want empty", msg)
	}
	if len(cards) != 1 || cards[0].ID != 1204 {
		t.Fatalf("cards = %#v, want card #1204", cards)
	}
}

func TestCardBoxCardsPreservesCardQueryOrderForFilteredResults(t *testing.T) {
	store := masterdata.NewStore()
	now := time.Now().UnixMilli()
	store.SetAll(&masterdata.MasterData{Cards: []masterdata.CardInfo{
		{ID: 1, CharacterID: 1, CardRarityType: "rarity_4", Attr: "cool", Prefix: "蓝卡旧低角色", ReleaseAt: now - 100, SupportUnit: "none"},
		{ID: 2, CharacterID: 26, CardRarityType: "rarity_4", Attr: "cool", Prefix: "蓝卡新高角色", ReleaseAt: now, SupportUnit: "none"},
	}})

	cards, msg := cardBoxCards(store, cardBoxQueryOptions{FilterText: "蓝"})
	if msg != "" {
		t.Fatalf("msg = %q, want empty", msg)
	}
	if len(cards) != 2 || cards[0].ID != 2 || cards[1].ID != 1 {
		t.Fatalf("cards order = %#v, want /查卡 release-desc order [2, 1]", cards)
	}
}

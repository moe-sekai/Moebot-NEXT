package crmission

import (
	"strings"
	"testing"

	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/suite"
)

func TestFieldsUsesMissionData(t *testing.T) {
	fields := Fields()
	want := suite.Fields(suite.FieldUserCharacterMissionV2s, suite.FieldUserCharacterMissionV2Statuses, suite.FieldUserCharacters)
	if len(fields) != len(want) {
		t.Fatalf("fields len = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("fields[%d] = %q, want %q", i, fields[i], want[i])
		}
	}
}

func TestParseArgsOverview(t *testing.T) {
	opts, err := ParseArgs("miku")
	if err != nil {
		t.Fatalf("parse args error: %v", err)
	}
	if opts.CharacterID != 21 || opts.ShowAll || opts.MissionType != "" {
		t.Fatalf("opts = %+v", opts)
	}
}

func TestParseArgsAllMissionAliasAndPage(t *testing.T) {
	opts, err := ParseArgs("miku all 队长 p2")
	if err != nil {
		t.Fatalf("parse args error: %v", err)
	}
	if opts.CharacterID != 21 || !opts.ShowAll || opts.MissionType != "play_live" || opts.Page != 2 {
		t.Fatalf("opts = %+v", opts)
	}

	exOpts, err := ParseArgs("miku all 队长次数EX page3")
	if err != nil {
		t.Fatalf("parse ex args error: %v", err)
	}
	if exOpts.MissionType != "play_live_ex" || exOpts.Page != 3 {
		t.Fatalf("ex opts = %+v", exOpts)
	}
	shortExOpts, err := ParseArgs("miku all 队长EX")
	if err != nil {
		t.Fatalf("parse short ex args error: %v", err)
	}
	if shortExOpts.MissionType != "play_live_ex" {
		t.Fatalf("short ex mission type = %q, want play_live_ex", shortExOpts.MissionType)
	}
}

func TestOverviewComputesNextNeed(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{CharacterMissionV2ParameterGroups: []masterdata.CharacterMissionV2ParameterGroup{
		{ID: 1, Seq: 1, Requirement: 10, Exp: 100},
		{ID: 1, Seq: 2, Requirement: 30, Exp: 200},
		{ID: 101, Seq: 1, Requirement: 5, Exp: 50},
		{ID: 101, Seq: 2, Requirement: 8, Exp: 60},
	}})
	profile := Profile{
		Missions: []MissionV2{
			{CharacterID: 21, CharacterMissionType: "play_live", Progress: 12},
			{CharacterID: 21, CharacterMissionType: "play_live_ex", Progress: 3},
			{CharacterID: 21, CharacterMissionType: "collect_another_vocal", Progress: 7},
		},
		Statuses: []MissionV2Status{{CharacterID: 21, ParameterGroupID: 101, Seq: 1}},
	}
	rows := BuildOverviewRows(store, profile, 21)
	row := findRow(rows, "play_live")
	if row == nil {
		t.Fatalf("play_live row not found: %#v", rows)
	}
	if row.Current != 12 || row.Level != 1 || row.NextNeed != 30 || row.NextExp != 200 {
		t.Fatalf("play_live row = %+v", row)
	}
	ex := findRow(rows, "play_live_ex")
	if ex == nil || ex.Current != 8 || ex.NextNeed != 13 || ex.NextExp != 60 {
		t.Fatalf("play_live_ex row = %+v", ex)
	}
}

func TestAllRows(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{CharacterMissionV2ParameterGroups: []masterdata.CharacterMissionV2ParameterGroup{
		{ID: 1, Seq: 1, Requirement: 10, Exp: 100},
		{ID: 1, Seq: 2, Requirement: 30, Exp: 200},
	}})
	profile := Profile{Missions: []MissionV2{{CharacterID: 21, CharacterMissionType: "play_live", Progress: 12}}}
	rows, _, err := BuildAllRows(store, profile, 21, "play_live")
	if err != nil {
		t.Fatalf("all rows error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d", len(rows))
	}
	if rows[0].Reached != true || rows[1].Reached != false || rows[1].Requirement != 30 || rows[1].AccExp != 300 {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestAllRowsForExUsesAccumulatedRequirement(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{CharacterMissionV2ParameterGroups: []masterdata.CharacterMissionV2ParameterGroup{
		{ID: 101, Seq: 1, Requirement: 5, Exp: 50},
		{ID: 101, Seq: 2, Requirement: 8, Exp: 60},
	}})
	profile := Profile{
		Missions: []MissionV2{{CharacterID: 21, CharacterMissionType: "play_live_ex", Progress: 3}},
		Statuses: []MissionV2Status{{CharacterID: 21, ParameterGroupID: 101, Seq: 1}},
	}
	rows, _, err := BuildAllRows(store, profile, 21, "play_live_ex")
	if err != nil {
		t.Fatalf("all rows error: %v", err)
	}
	if rows[0].AccRequirement != 5 || rows[1].AccRequirement != 13 || !rows[0].Reached || rows[1].Reached {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestBuildPayloadPaginatesAllRowsNearCurrent(t *testing.T) {
	store := masterdata.NewStore()
	groups := make([]masterdata.CharacterMissionV2ParameterGroup, 0, 205)
	for i := 1; i <= 205; i++ {
		groups = append(groups, masterdata.CharacterMissionV2ParameterGroup{ID: 1, Seq: i, Requirement: i * 10, Exp: i})
	}
	store.SetAll(&masterdata.MasterData{CharacterMissionV2ParameterGroups: groups})
	profile := Profile{Missions: []MissionV2{{CharacterID: 21, CharacterMissionType: "play_live", Progress: 950}}}
	payload, fallback, err := BuildPayload("jp", profile, store, nil, Options{CharacterID: 21, ShowAll: true, MissionType: "play_live"})
	if err != nil {
		t.Fatalf("build payload error: %v", err)
	}
	if len(payload.AllRows) != DefaultAllPageSize {
		t.Fatalf("render rows = %d, want %d", len(payload.AllRows), DefaultAllPageSize)
	}
	if payload.AllRowsTotal != 205 || payload.Page != 2 || payload.ShownFrom != 81 || payload.ShownTo != 160 || payload.TotalPages != 3 {
		t.Fatalf("payload page metadata = %+v", payload)
	}
	if payload.AllRows[0].Seq != 81 || payload.AllRows[len(payload.AllRows)-1].Seq != 160 {
		t.Fatalf("shown rows = first %+v last %+v", payload.AllRows[0], payload.AllRows[len(payload.AllRows)-1])
	}
	if !strings.Contains(fallback, "显示: 81-160 / 205") || !strings.Contains(fallback, "追加 p3") {
		t.Fatalf("fallback missing pagination info:\n%s", fallback)
	}
}

func TestBuildPayloadUsesExplicitPage(t *testing.T) {
	store := masterdata.NewStore()
	groups := make([]masterdata.CharacterMissionV2ParameterGroup, 0, 125)
	for i := 1; i <= 125; i++ {
		groups = append(groups, masterdata.CharacterMissionV2ParameterGroup{ID: 1, Seq: i, Requirement: i, Exp: 1})
	}
	store.SetAll(&masterdata.MasterData{CharacterMissionV2ParameterGroups: groups})
	profile := Profile{Missions: []MissionV2{{CharacterID: 21, CharacterMissionType: "play_live", Progress: 1}}}
	payload, _, err := BuildPayload("jp", profile, store, nil, Options{CharacterID: 21, ShowAll: true, MissionType: "play_live", Page: 2})
	if err != nil {
		t.Fatalf("build payload error: %v", err)
	}
	if payload.Page != 2 || payload.ShownFrom != 81 || payload.ShownTo != 125 || len(payload.AllRows) != 45 {
		t.Fatalf("payload = %+v", payload)
	}
}

func findRow(rows []Row, missionType string) *Row {
	for i := range rows {
		if rows[i].MissionType == missionType {
			return &rows[i]
		}
	}
	return nil
}

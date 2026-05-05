package commands

import (
	"testing"

	"moebot-next/internal/masterdata"
	"moebot-next/internal/suite"
)

func TestCharacterRankMissionFieldsUsesMissionData(t *testing.T) {
	fields := characterRankMissionFields()
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

func TestParseCharacterRankMissionArgsOverview(t *testing.T) {
	opts, err := parseCharacterRankMissionArgs("miku")
	if err != nil {
		t.Fatalf("parse args error: %v", err)
	}
	if opts.CharacterID != 21 || opts.ShowAll || opts.MissionType != "" {
		t.Fatalf("opts = %+v", opts)
	}
}

func TestParseCharacterRankMissionArgsAllMissionAlias(t *testing.T) {
	opts, err := parseCharacterRankMissionArgs("miku all 队长")
	if err != nil {
		t.Fatalf("parse args error: %v", err)
	}
	if opts.CharacterID != 21 || !opts.ShowAll || opts.MissionType != "play_live" {
		t.Fatalf("opts = %+v", opts)
	}

	exOpts, err := parseCharacterRankMissionArgs("miku all 队长次数EX")
	if err != nil {
		t.Fatalf("parse ex args error: %v", err)
	}
	if exOpts.MissionType != "play_live_ex" {
		t.Fatalf("ex mission type = %q, want play_live_ex", exOpts.MissionType)
	}
	shortExOpts, err := parseCharacterRankMissionArgs("miku all 队长EX")
	if err != nil {
		t.Fatalf("parse short ex args error: %v", err)
	}
	if shortExOpts.MissionType != "play_live_ex" {
		t.Fatalf("short ex mission type = %q, want play_live_ex", shortExOpts.MissionType)
	}
}

func TestCharacterRankMissionOverviewComputesNextNeed(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{CharacterMissionV2ParameterGroups: []masterdata.CharacterMissionV2ParameterGroup{
		{ID: 1, Seq: 1, Requirement: 10, Exp: 100},
		{ID: 1, Seq: 2, Requirement: 30, Exp: 200},
		{ID: 101, Seq: 1, Requirement: 5, Exp: 50},
		{ID: 101, Seq: 2, Requirement: 8, Exp: 60},
	}})
	profile := characterRankMissionProfile{
		Missions: []characterMissionV2{
			{CharacterID: 21, CharacterMissionType: "play_live", Progress: 12},
			{CharacterID: 21, CharacterMissionType: "play_live_ex", Progress: 3},
			{CharacterID: 21, CharacterMissionType: "collect_another_vocal", Progress: 7},
		},
		Statuses: []characterMissionV2Status{{CharacterID: 21, ParameterGroupID: 101, Seq: 1}},
	}
	rows := buildCharacterRankMissionOverviewRows(store, profile, 21)
	row := findCharacterRankMissionRow(rows, "play_live")
	if row == nil {
		t.Fatalf("play_live row not found: %#v", rows)
	}
	if row.Current != 12 || row.Level != 1 || row.NextNeed != 30 || row.NextExp != 200 {
		t.Fatalf("play_live row = %+v", row)
	}
	ex := findCharacterRankMissionRow(rows, "play_live_ex")
	if ex == nil || ex.Current != 8 || ex.NextNeed != 13 || ex.NextExp != 60 {
		t.Fatalf("play_live_ex row = %+v", ex)
	}
}

func TestCharacterRankMissionAllRows(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{CharacterMissionV2ParameterGroups: []masterdata.CharacterMissionV2ParameterGroup{
		{ID: 1, Seq: 1, Requirement: 10, Exp: 100},
		{ID: 1, Seq: 2, Requirement: 30, Exp: 200},
	}})
	profile := characterRankMissionProfile{Missions: []characterMissionV2{{CharacterID: 21, CharacterMissionType: "play_live", Progress: 12}}}
	rows, err := buildCharacterRankMissionAllRows(store, profile, 21, "play_live")
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

func TestCharacterRankMissionAllRowsForExUsesAccumulatedRequirement(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{CharacterMissionV2ParameterGroups: []masterdata.CharacterMissionV2ParameterGroup{
		{ID: 101, Seq: 1, Requirement: 5, Exp: 50},
		{ID: 101, Seq: 2, Requirement: 8, Exp: 60},
	}})
	profile := characterRankMissionProfile{
		Missions: []characterMissionV2{{CharacterID: 21, CharacterMissionType: "play_live_ex", Progress: 3}},
		Statuses: []characterMissionV2Status{{CharacterID: 21, ParameterGroupID: 101, Seq: 1}},
	}
	rows, err := buildCharacterRankMissionAllRows(store, profile, 21, "play_live_ex")
	if err != nil {
		t.Fatalf("all rows error: %v", err)
	}
	if rows[0].AccRequirement != 5 || rows[1].AccRequirement != 13 || !rows[0].Reached || rows[1].Reached {
		t.Fatalf("rows = %+v", rows)
	}
}

func findCharacterRankMissionRow(rows []characterRankMissionRow, missionType string) *characterRankMissionRow {
	for i := range rows {
		if rows[i].MissionType == missionType {
			return &rows[i]
		}
	}
	return nil
}

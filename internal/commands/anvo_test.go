package commands

import (
	"testing"
	"time"

	"moebot-next/internal/masterdata"
	"moebot-next/internal/suite"
)

func TestAnvoFieldsUsesMusicVocals(t *testing.T) {
	fields := anvoFields()
	want := suite.Fields(suite.FieldUserMusicVocals, suite.FieldUserMusics)
	if len(fields) != len(want) {
		t.Fatalf("fields len = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("fields[%d] = %q, want %q", i, fields[i], want[i])
		}
	}
}

func TestParseAnvoArgsResolvesCharacter(t *testing.T) {
	cid, err := parseAnvoArgs("miku")
	if err != nil {
		t.Fatalf("parseAnvoArgs returned error: %v", err)
	}
	if cid != 21 {
		t.Fatalf("cid = %d, want 21", cid)
	}
}

func TestOwnedMusicVocalIDsIncludesNestedUserMusics(t *testing.T) {
	profile := anvoProfile{
		UserMusicVocals: []userMusicVocal{{MusicVocalID: 10}},
		UserMusics:      []userMusicWithVocals{{UserMusicVocals: []userMusicVocal{{MusicVocalID: 20}}}},
	}
	owned := ownedMusicVocalIDs(profile)
	if !owned[10] || !owned[20] || len(owned) != 2 {
		t.Fatalf("owned = %#v", owned)
	}
}

func TestBuildAnvoEntriesFiltersCharacterAndOwnership(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		Musics: []masterdata.MusicInfo{
			{ID: 1, Title: "Tell Your World", AssetbundleName: "music001", PublishedAt: time.Now().Add(-time.Hour).UnixMilli()},
			{ID: 2, Title: "Future Song", AssetbundleName: "music002", PublishedAt: time.Now().Add(time.Hour).UnixMilli()},
		},
		MusicVocals: []masterdata.MusicVocal{
			{ID: 101, MusicID: 1, Caption: "Another Vocal", Characters: []masterdata.MusicVocalCharacter{{CharacterID: 21, CharacterType: "game_character", Seq: 1}}},
			{ID: 102, MusicID: 1, Caption: "Another Vocal", Characters: []masterdata.MusicVocalCharacter{{CharacterID: 1, CharacterType: "game_character", Seq: 1}}},
			{ID: 103, MusicID: 2, Caption: "Another Vocal", Characters: []masterdata.MusicVocalCharacter{{CharacterID: 21, CharacterType: "game_character", Seq: 1}}},
		},
	})

	entries := buildAnvoEntries(store, nil, 21, map[int]bool{101: true}, time.Now())
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1: %#v", len(entries), entries)
	}
	entry := entries[0]
	if entry.MusicVocalID != 101 || entry.MusicID != 1 || !entry.Owned || entry.Title != "Tell Your World" {
		t.Fatalf("entry = %+v", entry)
	}
	if len(entry.CharacterIDs) != 1 || entry.CharacterIDs[0] != 21 {
		t.Fatalf("CharacterIDs = %#v", entry.CharacterIDs)
	}
}

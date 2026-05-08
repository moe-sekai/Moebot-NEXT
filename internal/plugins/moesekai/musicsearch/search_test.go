package musicsearch

import (
	"testing"

	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/plugins/moesekai/masterdata"
)

func TestSearchExactAliasReturnsDetail(t *testing.T) {
	store := testMusicStore(3)
	aliases := map[int]assets.MusicAlias{
		1: {MusicID: 1, Title: "Tell Your World", Aliases: []string{"谷歌", "tyw"}},
	}

	result := Search(store, aliases, "谷歌", Options{})
	if result.Mode != ModeDetail || result.Music == nil || result.Music.ID != 1 {
		t.Fatalf("result = %#v, want detail #1", result)
	}
}

func TestSearchAliasCanReturnLimitedList(t *testing.T) {
	store := testMusicStore(15)
	aliases := make(map[int]assets.MusicAlias, 15)
	for i := 1; i <= 15; i++ {
		aliases[i] = assets.MusicAlias{MusicID: i, Title: "Song", Aliases: []string{"候选"}}
	}

	result := Search(store, aliases, "候选", Options{Limit: DefaultListLimit})
	if result.Mode != ModeList {
		t.Fatalf("mode = %q, want list", result.Mode)
	}
	if result.Total != 15 || len(result.Musics) != DefaultListLimit || result.TotalPages != 2 {
		t.Fatalf("total=%d shown=%d totalPages=%d, want 15/%d/2", result.Total, len(result.Musics), result.TotalPages, DefaultListLimit)
	}
}

func TestSearchAliasHonorsDifficulty(t *testing.T) {
	store := masterdata.NewStore()
	store.SetAll(&masterdata.MasterData{
		Musics: []masterdata.MusicInfo{
			{ID: 1, Title: "Master Song"},
			{ID: 2, Title: "Expert Song"},
		},
		MusicDifficulties: []masterdata.MusicDifficulty{
			{MusicID: 1, MusicDifficulty: "master", PlayLevel: 30},
			{MusicID: 2, MusicDifficulty: "expert", PlayLevel: 25},
		},
	})
	aliases := map[int]assets.MusicAlias{
		1: {MusicID: 1, Aliases: []string{"同名别名"}},
		2: {MusicID: 2, Aliases: []string{"同名别名"}},
	}

	result := Search(store, aliases, "同名别名", Options{Difficulty: "master"})
	if result.Mode != ModeDetail || result.Music == nil || result.Music.ID != 1 {
		t.Fatalf("result = %#v, want master detail #1", result)
	}
}

func TestParseQueryExtractsDifficulty(t *testing.T) {
	query, options := ParseQuery("mas 谷歌")
	if query != "谷歌" || options.Difficulty != "master" {
		t.Fatalf("query=%q difficulty=%q, want 谷歌/master", query, options.Difficulty)
	}

	for _, tc := range []struct {
		input    string
		wantDiff string
	}{
		{input: "关键词 ex", wantDiff: "expert"},
		{input: "关键词 expert", wantDiff: "expert"},
		{input: "关键词 ma", wantDiff: "master"},
		{input: "关键词 mas", wantDiff: "master"},
		{input: "关键词 master", wantDiff: "master"},
		{input: "关键词 apd", wantDiff: "append"},
		{input: "关键词 ap", wantDiff: "append"},
		{input: "关键词 append", wantDiff: "append"},
		{input: "关键词 hd", wantDiff: "hard"},
		{input: "关键词 hard", wantDiff: "hard"},
		{input: "关键词 nm", wantDiff: "normal"},
		{input: "关键词 normal", wantDiff: "normal"},
		{input: "关键词 ez", wantDiff: "easy"},
		{input: "关键词 easy", wantDiff: "easy"},
	} {
		query, options := ParseQuery(tc.input)
		if query != "关键词" || options.Difficulty != tc.wantDiff {
			t.Fatalf("ParseQuery(%q) = query %q diff %q, want 关键词/%s", tc.input, query, options.Difficulty, tc.wantDiff)
		}
	}
}

func testMusicStore(count int) *masterdata.Store {
	store := masterdata.NewStore()
	musics := make([]masterdata.MusicInfo, 0, count)
	diffs := make([]masterdata.MusicDifficulty, 0, count)
	for i := 1; i <= count; i++ {
		musics = append(musics, masterdata.MusicInfo{ID: i, Title: "Song"})
		diffs = append(diffs, masterdata.MusicDifficulty{MusicID: i, MusicDifficulty: "master", PlayLevel: 30})
	}
	store.SetAll(&masterdata.MasterData{Musics: musics, MusicDifficulties: diffs})
	return store
}

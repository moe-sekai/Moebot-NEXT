package b30

import (
	"strings"
	"testing"
)

func TestParseConstantsCSV(t *testing.T) {
	csv := `Song,JP Name,Constant,Level,Note Count,Difficulty,Song ID,Notes
"Song, With Comma",曲名,33.5,34,"1,234",Master,101,free note
Broken,壊れた,0,0,0,Master,102,
MissingID,missing,31.2,31,900,Expert,,
`
	entries, err := ParseConstantsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1: %#v", len(entries), entries)
	}
	entry := entries[0]
	if entry.MusicID != 101 || entry.Difficulty != "master" || entry.Constant != 33.5 || entry.Level != 34 || entry.NoteCount != 1234 {
		t.Fatalf("entry = %#v", entry)
	}
	if entry.Title != "Song, With Comma" || entry.JPTitle != "曲名" {
		t.Fatalf("titles = %q / %q", entry.Title, entry.JPTitle)
	}
}

func TestNormalizeDifficulty(t *testing.T) {
	cases := map[string]string{
		"MAS":    "master",
		"ma":     "master",
		"APPEND": "append",
		"apd":    "append",
		"EX":     "expert",
		"困难":     "hard",
	}
	for input, want := range cases {
		if got := NormalizeDifficulty(input); got != want {
			t.Fatalf("NormalizeDifficulty(%q) = %q, want %q", input, got, want)
		}
	}
}

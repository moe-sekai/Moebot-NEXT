package assets

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultMusicAliasURL is the canonical remote location of the alias file.
const DefaultMusicAliasURL = "https://moe.exmeaning.com/data/music_alias/music_aliases.json"

// ---------------------------------------------------------------------------
// Music alias types
// ---------------------------------------------------------------------------

// MusicAlias holds a song's alias mappings.
type MusicAlias struct {
	MusicID int      `json:"music_id"`
	Title   string   `json:"title"`
	Aliases []string `json:"aliases"`
}

// musicAliasResponse models the top-level JSON envelope returned by the remote
// alias file.
type musicAliasResponse struct {
	Musics []MusicAlias `json:"musics"`
}

// ---------------------------------------------------------------------------
// Loading
// ---------------------------------------------------------------------------

// LoadMusicAliases fetches aliases from a remote JSON endpoint and returns a
// map keyed by music ID. Pass an empty string to use DefaultMusicAliasURL.
func LoadMusicAliases(url string) (map[int]MusicAlias, error) {
	if url == "" {
		url = DefaultMusicAliasURL
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("assets: fetch music aliases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("assets: fetch music aliases: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("assets: read music aliases body: %w", err)
	}

	var envelope musicAliasResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("assets: decode music aliases: %w", err)
	}

	result := make(map[int]MusicAlias, len(envelope.Musics))
	for _, m := range envelope.Musics {
		result[m.MusicID] = m
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Matching
// ---------------------------------------------------------------------------

// MatchesMusicAlias checks if query matches any alias (or the title) for the
// given music ID using case-insensitive substring matching.
func MatchesMusicAlias(musicID int, query string, aliases map[int]MusicAlias) bool {
	entry, ok := aliases[musicID]
	if !ok {
		return false
	}
	q := strings.ToLower(query)
	if strings.Contains(strings.ToLower(entry.Title), q) {
		return true
	}
	for _, alias := range entry.Aliases {
		if strings.Contains(strings.ToLower(alias), q) {
			return true
		}
	}
	return false
}

// FindMusicByAlias searches ALL entries in the alias map and returns every
// MusicAlias whose title or any alias contains the query (case-insensitive
// substring). Handy for "search by keyword" features.
func FindMusicByAlias(query string, aliases map[int]MusicAlias) []MusicAlias {
	q := strings.ToLower(query)
	var matches []MusicAlias
	for _, entry := range aliases {
		if strings.Contains(strings.ToLower(entry.Title), q) {
			matches = append(matches, entry)
			continue
		}
		for _, alias := range entry.Aliases {
			if strings.Contains(strings.ToLower(alias), q) {
				matches = append(matches, entry)
				break
			}
		}
	}
	return matches
}

// ---------------------------------------------------------------------------
// Character name search (convenience wrapper over constants.go)
// ---------------------------------------------------------------------------

// FindCharacterByName searches character constants by name or alias. It is a
// thin alias for FindCharacterByAlias defined in constants.go, provided here
// so that callers importing only "aliases" functionality get the function name
// they expect.
func FindCharacterByName(query string) *Character {
	return FindCharacterByAlias(query)
}

package musicsearch

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/plugins/moesekai/masterdata"
)

const DefaultListLimit = 12

type Mode string

const (
	ModeNone   Mode = ""
	ModeDetail Mode = "detail"
	ModeList   Mode = "list"
)

type QueryOptions struct {
	Difficulty string
	Refresh    bool
}

type Options struct {
	Difficulty string
	Limit      int
}

type Result struct {
	Query      string
	Mode       Mode
	Music      *masterdata.MusicInfo
	Musics     []masterdata.MusicInfo
	Total      int
	Page       int
	TotalPages int
	Message    string
}

var difficultyAliases = map[string]string{
	"easy":   "easy",
	"ez":     "easy",
	"简单":     "easy",
	"normal": "normal",
	"nm":     "normal",
	"普通":     "normal",
	"hard":   "hard",
	"hd":     "hard",
	"困难":     "hard",
	"expert": "expert",
	"ex":     "expert",
	"专家":     "expert",
	"master": "master",
	"mas":    "master",
	"ma":     "master",
	"大师":     "master",
	"append": "append",
	"apd":    "append",
	"ap":     "append",
	"追加":     "append",
}

func ParseQuery(raw string) (string, QueryOptions) {
	fields := strings.Fields(strings.TrimSpace(raw))
	remaining := make([]string, 0, len(fields))
	var options QueryOptions
	for _, field := range fields {
		key := Normalize(field)
		if key == "refresh" || key == "刷新" {
			options.Refresh = true
			continue
		}
		if diff := difficultyAliases[key]; diff != "" {
			options.Difficulty = diff
			continue
		}
		remaining = append(remaining, field)
	}
	return strings.TrimSpace(strings.Join(remaining, " ")), options
}

func Search(store *masterdata.Store, aliases map[int]assets.MusicAlias, raw string, options Options) Result {
	query := strings.TrimSpace(raw)
	limit := NormalizeLimit(options.Limit)
	result := Result{Query: query, Page: 1}
	if store == nil || query == "" {
		result.Message = "搜索文本为空"
		return result
	}

	if id, ok := parseMusicID(query); ok {
		if music := store.GetMusic(id); music != nil && musicHasDifficulty(store, id, options.Difficulty) {
			return detailResult(query, *music)
		}
		result.Message = fmt.Sprintf("没有找到曲目 #%d", id)
		return result
	}

	if idx, ok := parseRelativeIndex(query); ok && idx < 0 {
		musics := sortedMusicsByPublished(store.AllMusics(), false)
		pos := len(musics) + idx
		if pos >= 0 && pos < len(musics) && musicHasDifficulty(store, musics[pos].ID, options.Difficulty) {
			return detailResult(query, musics[pos])
		}
		result.Message = fmt.Sprintf("找不到倒数第 %d 首已发布曲目", -idx)
		return result
	}

	if eventID, ok := parseEventToken(query); ok {
		links := store.GetEventMusics(eventID)
		musics := make([]masterdata.MusicInfo, 0, len(links))
		for _, link := range links {
			if music := store.GetMusic(link.MusicID); music != nil && musicHasDifficulty(store, music.ID, options.Difficulty) {
				musics = append(musics, *music)
			}
		}
		if len(musics) == 0 {
			result.Message = fmt.Sprintf("活动 #%d 没有关联曲目", eventID)
			return result
		}
		return resultFromMatches(query, musics, limit)
	}

	clean := Normalize(query)
	exacts := make([]masterdata.MusicInfo, 0)
	seenExact := map[int]struct{}{}
	for _, music := range store.AllMusics() {
		if !musicHasDifficulty(store, music.ID, options.Difficulty) {
			continue
		}
		if Normalize(music.Title) == clean || Normalize(music.Pronunciation) == clean || aliasExactMatch(aliases, music.ID, clean) {
			if _, ok := seenExact[music.ID]; !ok {
				exacts = append(exacts, music)
				seenExact[music.ID] = struct{}{}
			}
		}
	}
	if len(exacts) > 0 {
		sortMusicsByIDDesc(exacts)
		return resultFromMatches(query, exacts, limit)
	}

	hits := scoreMusics(store.AllMusics(), aliases, query, options.Difficulty, store)
	if len(hits) == 0 {
		result.Message = fmt.Sprintf("没有找到与「%s」匹配的曲目", query)
		return result
	}
	matches := make([]masterdata.MusicInfo, 0, len(hits))
	for _, hit := range hits {
		matches = append(matches, hit.music)
	}
	return resultFromMatches(query, matches, limit)
}

func (r Result) DisplayMusics() []masterdata.MusicInfo {
	if len(r.Musics) > 0 {
		return r.Musics
	}
	if r.Music != nil {
		return []masterdata.MusicInfo{*r.Music}
	}
	return nil
}

func NormalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultListLimit
	}
	return limit
}

func LimitMusics(musics []masterdata.MusicInfo, limit int) []masterdata.MusicInfo {
	limit = NormalizeLimit(limit)
	if len(musics) <= limit {
		return musics
	}
	return musics[:limit]
}

func TotalPages(total int, limit int) int {
	limit = NormalizeLimit(limit)
	if total <= 0 {
		return 1
	}
	return (total + limit - 1) / limit
}

func Normalize(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Map(func(r rune) rune {
		if r >= 0xFF01 && r <= 0xFF5E {
			r = r - 0xFEE0
		}
		if unicode.IsSpace(r) || r == '_' || r == '-' || r == '・' || r == '·' || r == '.' || r == '/' || r == '／' {
			return -1
		}
		return unicode.ToLower(r)
	}, value)
	return value
}

type musicHit struct {
	music masterdata.MusicInfo
	score int
}

func detailResult(query string, music masterdata.MusicInfo) Result {
	musicCopy := music
	return Result{Query: query, Mode: ModeDetail, Music: &musicCopy, Total: 1, Page: 1, TotalPages: 1}
}

func resultFromMatches(query string, matches []masterdata.MusicInfo, limit int) Result {
	if len(matches) == 0 {
		return Result{Query: query, Page: 1, TotalPages: 1}
	}
	if len(matches) == 1 {
		return detailResult(query, matches[0])
	}
	limit = NormalizeLimit(limit)
	return Result{
		Query:      query,
		Mode:       ModeList,
		Musics:     LimitMusics(matches, limit),
		Total:      len(matches),
		Page:       1,
		TotalPages: TotalPages(len(matches), limit),
	}
}

func scoreMusics(musics []masterdata.MusicInfo, aliases map[int]assets.MusicAlias, query string, diff string, store *masterdata.Store) []musicHit {
	hits := make([]musicHit, 0)
	for _, music := range musics {
		if !musicHasDifficulty(store, music.ID, diff) {
			continue
		}
		score := bestLocalScore(query, music.Title, music.Pronunciation, music.Composer, music.Lyricist, music.Arranger, music.AssetbundleName)
		if aliasScore := bestAliasScore(aliases, music.ID, query); aliasScore > score {
			score = aliasScore
		}
		if score > 0 {
			hits = append(hits, musicHit{music: music, score: score})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return hits[i].music.ID > hits[j].music.ID
	})
	return hits
}

func aliasExactMatch(aliases map[int]assets.MusicAlias, musicID int, cleanQuery string) bool {
	entry, ok := aliases[musicID]
	if !ok {
		return false
	}
	if Normalize(entry.Title) == cleanQuery {
		return true
	}
	for _, alias := range entry.Aliases {
		if Normalize(alias) == cleanQuery {
			return true
		}
	}
	return false
}

func bestAliasScore(aliases map[int]assets.MusicAlias, musicID int, query string) int {
	entry, ok := aliases[musicID]
	if !ok {
		return 0
	}
	values := append([]string{entry.Title}, entry.Aliases...)
	return bestLocalScore(query, values...)
}

func bestLocalScore(keyword string, candidates ...string) int {
	best := 0
	for _, candidate := range candidates {
		if score := localFuzzyScore(candidate, keyword); score > best {
			best = score
		}
	}
	return best
}

func localFuzzyScore(target string, keyword string) int {
	t := Normalize(target)
	k := Normalize(keyword)
	if t == "" || k == "" {
		return 0
	}
	if t == k {
		return 120
	}
	if strings.HasPrefix(t, k) {
		return 100
	}
	if strings.Contains(t, k) {
		return 80 + int(float64(len([]rune(k)))/float64(maxInt(1, len([]rune(t))))*15)
	}
	if isRuneSubsequence(k, t) {
		return 45
	}
	return 0
}

func isRuneSubsequence(needle string, haystack string) bool {
	nr := []rune(needle)
	hr := []rune(haystack)
	idx := 0
	for _, r := range hr {
		if idx < len(nr) && nr[idx] == r {
			idx++
		}
	}
	return idx == len(nr)
}

func parseMusicID(query string) (int, bool) {
	clean := Normalize(query)
	if strings.HasPrefix(clean, "id") {
		clean = strings.TrimPrefix(clean, "id")
	}
	id, err := strconv.Atoi(clean)
	return id, err == nil && id > 0
}

func parseEventToken(query string) (int, bool) {
	clean := Normalize(query)
	if !strings.HasPrefix(clean, "event") {
		return 0, false
	}
	id, err := strconv.Atoi(strings.TrimPrefix(clean, "event"))
	return id, err == nil && id > 0
}

func parseRelativeIndex(query string) (int, bool) {
	query = strings.TrimSpace(query)
	if query == "" {
		return 0, false
	}
	if strings.HasPrefix(query, "+") || strings.HasPrefix(query, "-") {
		value, err := strconv.Atoi(query)
		return value, err == nil
	}
	return 0, false
}

func sortedMusicsByPublished(musics []masterdata.MusicInfo, includeFuture bool) []masterdata.MusicInfo {
	now := time.Now().UnixMilli()
	out := make([]masterdata.MusicInfo, 0, len(musics))
	for _, music := range musics {
		if includeFuture || music.PublishedAt <= 0 || music.PublishedAt <= now {
			out = append(out, music)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PublishedAt != out[j].PublishedAt {
			return out[i].PublishedAt < out[j].PublishedAt
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func musicHasDifficulty(store *masterdata.Store, musicID int, diff string) bool {
	if diff == "" || store == nil {
		return true
	}
	for _, d := range store.GetMusicDifficulties(musicID) {
		if d.MusicDifficulty == diff {
			return true
		}
	}
	return false
}

func sortMusicsByIDDesc(musics []masterdata.MusicInfo) {
	sort.SliceStable(musics, func(i, j int) bool {
		return musics[i].ID > musics[j].ID
	})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

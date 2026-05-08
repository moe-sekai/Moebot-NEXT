package masterdata

import (
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// ---------------------------------------------------------------------------
// search.go — Fuzzy search engine for PJSK masterdata
//
// Provides SearchCards, SearchMusics, SearchEvents, SearchGachas on *Store.
// Uses a multi-field fuzzy matching algorithm with character-name resolution.
//
// TODO(assets): When internal/assets is ready, replace the embedded
//   characterSearchData with assets.GetCharacterByID / assets.GetAliases.
// TODO(utils): The fuzzy helpers here can be extracted to internal/utils/fuzzy.go.
// ---------------------------------------------------------------------------

const maxSearchResults = 25

// searchHit is used internally to pair an index in a data slice with its
// relevance score so we can sort and truncate the results.
type searchHit struct {
	idx   int
	score int
}

// ========================== Character Name Data ============================
// Embedded PJSK character lookup for card-search character resolution.
// IDs match the gameCharacterId in upstream masterdata.

type characterSearchEntry struct {
	ID    int
	Names []string // all searchable names (JP, CN, EN, nicknames)
}

//nolint:gochecknoglobals
var characterSearchData = []characterSearchEntry{
	// --- Leo/need ---
	{1, []string{"星乃一歌", "一歌", "ichika", "hoshino", "一哥"}},
	{2, []string{"天馬咲希", "咲希", "saki", "tenma saki", "小咲", "天马咲希"}},
	{3, []string{"望月穂波", "穂波", "honami", "mochizuki", "穗波"}},
	{4, []string{"日野森志歩", "志歩", "shiho", "hinomori shiho", "志步", "日野森志步"}},
	// --- MORE MORE JUMP! ---
	{5, []string{"花里みのり", "みのり", "minori", "hanasato", "实乃里", "花里实乃里"}},
	{6, []string{"桐谷遥", "遥", "haruka", "kiritani", "遥遥"}},
	{7, []string{"桃井愛莉", "愛莉", "airi", "momoi", "爱莉", "桃井爱莉"}},
	{8, []string{"日野森雫", "雫", "shizuku", "hinomori shizuku", "雫雫"}},
	// --- Vivid BAD SQUAD ---
	{9, []string{"小豆沢こはね", "こはね", "kohane", "azusawa", "小羽", "小豆泽小羽"}},
	{10, []string{"白石杏", "杏", "an", "shiraishi", "杏杏"}},
	{11, []string{"東雲彰人", "彰人", "akito", "shinonome akito", "东云彰人"}},
	{12, []string{"青柳冬弥", "冬弥", "toya", "aoyagi", "冬弥弥"}},
	// --- Wonderlands×Showtime ---
	{13, []string{"天馬司", "司", "tsukasa", "tenma tsukasa", "天马司", "司司"}},
	{14, []string{"鳳えむ", "えむ", "emu", "otori", "笑梦", "凤笑梦"}},
	{15, []string{"草薙寧々", "寧々", "nene", "kusanagi", "宁宁", "草薙宁宁"}},
	{16, []string{"神代類", "類", "rui", "kamishiro", "类", "神代类"}},
	// --- 25時、ナイトコードで。 ---
	{17, []string{"宵崎奏", "奏", "kanade", "yoisaki", "奏酱"}},
	{18, []string{"朝比奈まふゆ", "まふゆ", "mafuyu", "asahina", "真冬", "冬雪"}},
	{19, []string{"東雲絵名", "絵名", "ena", "shinonome ena", "绘名", "东云绘名"}},
	{20, []string{"暁山瑞希", "瑞希", "mizuki", "akiyama", "晓山瑞希"}},
	// --- Virtual Singers ---
	{21, []string{"初音ミク", "ミク", "miku", "hatsune miku", "初音未来", "未来"}},
	{22, []string{"鏡音リン", "リン", "rin", "kagamine rin", "镜音铃", "铃"}},
	{23, []string{"鏡音レン", "レン", "len", "kagamine len", "镜音连", "连"}},
	{24, []string{"巡音ルカ", "ルカ", "luka", "megurine luka", "巡音流歌", "流歌"}},
	{25, []string{"meiko", "MEIKO", "メイコ"}},
	{26, []string{"kaito", "KAITO", "カイト"}},
}

// ========================== Fuzzy Matching ==================================

// normalize lowercases and trims a string for comparison.
func normalize(s string) string {
	s = strings.TrimSpace(s)
	// Convert full-width alphanumerics to half-width.
	s = strings.Map(func(r rune) rune {
		if r >= 0xFF01 && r <= 0xFF5E {
			return r - 0xFEE0 // full-width → half-width
		}
		return unicode.ToLower(r)
	}, s)
	return s
}

// fuzzyScore returns 0-100 indicating how well keyword matches target.
//
//	100  exact match
//	90   target starts with keyword
//	70-89 target contains keyword (bonus by coverage ratio)
//	40   keyword is a subsequence of target
//	0    no match
func fuzzyScore(target, keyword string) int {
	t := normalize(target)
	k := normalize(keyword)

	if len(k) == 0 || len(t) == 0 {
		return 0
	}

	// Exact
	if t == k {
		return 100
	}

	// Prefix
	if strings.HasPrefix(t, k) {
		return 90
	}

	// Contains — score scales with coverage ratio
	if strings.Contains(t, k) {
		ratio := float64(len(k)) / float64(len(t))
		return 70 + int(ratio*19) // 70 .. 89
	}

	// Word-prefix: any whitespace-delimited word starts with keyword
	for _, word := range strings.Fields(t) {
		if strings.HasPrefix(word, k) {
			return 65
		}
	}

	// Subsequence — all runes of k appear in t in order
	if isSubsequence(k, t) {
		return 40
	}

	return 0
}

// isSubsequence reports whether every rune of needle appears in haystack
// in the same order (not necessarily contiguous).
func isSubsequence(needle, haystack string) bool {
	nr := []rune(needle)
	hr := []rune(haystack)
	ni := 0
	for _, r := range hr {
		if ni < len(nr) && r == nr[ni] {
			ni++
		}
	}
	return ni == len(nr)
}

// bestScore returns the highest fuzzyScore for keyword against any of the
// candidate strings.
func bestScore(keyword string, candidates ...string) int {
	best := 0
	for _, c := range candidates {
		if s := fuzzyScore(c, keyword); s > best {
			best = s
		}
	}
	return best
}

// matchCharacterIDs returns a set of character IDs whose names match keyword.
// The returned map values are the match score.
func matchCharacterIDs(keyword string) map[int]int {
	result := make(map[int]int)
	k := normalize(keyword)
	if k == "" {
		return result
	}
	for _, ch := range characterSearchData {
		best := 0
		for _, name := range ch.Names {
			if s := fuzzyScore(name, k); s > best {
				best = s
			}
		}
		if best > 0 {
			result[ch.ID] = best
		}
	}
	return result
}

// ========================== Search Methods =================================

// SearchCards searches cards by keyword, matching against card prefix,
// character name, asset bundle name, and numeric ID.
// Results are sorted by relevance (descending), then by ID descending (newer first).
func (s *Store) SearchCards(keyword string) []CardInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil
	}

	// Direct ID lookup
	if id, err := strconv.Atoi(keyword); err == nil {
		if card, ok := s.cardByID[id]; ok {
			return []CardInfo{*card}
		}
	}

	// Resolve character names
	charScores := matchCharacterIDs(keyword)

	var hits []searchHit

	for i := range s.cards {
		card := &s.cards[i]
		score := 0

		// Character name match
		if cs, ok := charScores[card.CharacterID]; ok {
			score = cs
		}

		// Prefix (card title) match — often the most relevant
		if ps := fuzzyScore(card.Prefix, keyword); ps > score {
			score = ps
		}

		// AssetBundle name match (lower weight)
		if as := fuzzyScore(card.AssetbundleName, keyword); as > 0 && as/2 > score {
			score = as / 2
		}

		// GachaPhrase match
		if gs := fuzzyScore(card.GachaPhrase, keyword); gs > 0 && gs*2/3 > score {
			score = gs * 2 / 3
		}

		// Rarity keyword shortcut: "4星", "3星", "生日"
		score = applyRarityBonus(card, keyword, score)

		if score > 0 {
			hits = append(hits, searchHit{i, score})
		}
	}

	return s.collectCardResults(hits)
}

// applyRarityBonus gives a small bonus when the keyword contains a rarity hint.
func applyRarityBonus(card *CardInfo, keyword string, currentScore int) int {
	kn := normalize(keyword)
	if currentScore == 0 {
		// rarity-only search
		if matchesRarityKeyword(card, kn) {
			return 30
		}
		return 0
	}
	// If the keyword partially matches a rarity, boost the existing score slightly
	if matchesRarityKeyword(card, kn) {
		return currentScore + 5
	}
	return currentScore
}

func matchesRarityKeyword(card *CardInfo, kn string) bool {
	switch {
	case strings.Contains(kn, "4星") || strings.Contains(kn, "四星"):
		return card.CardRarityType == "rarity_4"
	case strings.Contains(kn, "3星") || strings.Contains(kn, "三星"):
		return card.CardRarityType == "rarity_3"
	case strings.Contains(kn, "2星") || strings.Contains(kn, "二星"):
		return card.CardRarityType == "rarity_2"
	case strings.Contains(kn, "1星") || strings.Contains(kn, "一星"):
		return card.CardRarityType == "rarity_1"
	case strings.Contains(kn, "生日") || strings.Contains(kn, "birthday"):
		return card.CardRarityType == "rarity_birthday"
	}
	return false
}

// collectCardResults sorts scored hits and returns the top results as CardInfo slices.
func (s *Store) collectCardResults(hits []searchHit) []CardInfo {
	if len(hits) == 0 {
		return nil
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return s.cards[hits[i].idx].ID > s.cards[hits[j].idx].ID // newer first
	})

	limit := len(hits)
	if limit > maxSearchResults {
		limit = maxSearchResults
	}

	out := make([]CardInfo, limit)
	for i := 0; i < limit; i++ {
		out[i] = s.cards[hits[i].idx]
	}
	return out
}

// SearchMusics searches musics by keyword, matching against title,
// pronunciation, composer, lyricist, arranger, and numeric ID.
func (s *Store) SearchMusics(keyword string) []MusicInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil
	}

	// Direct ID lookup
	if id, err := strconv.Atoi(keyword); err == nil {
		if music, ok := s.musicByID[id]; ok {
			return []MusicInfo{*music}
		}
	}

	type scored struct {
		idx   int
		score int
	}
	var hits []scored

	for i := range s.musics {
		m := &s.musics[i]

		score := bestScore(keyword,
			m.Title,
			m.Pronunciation,
			m.Composer,
			m.Lyricist,
			m.Arranger,
			m.AssetbundleName,
		)

		if score > 0 {
			hits = append(hits, scored{i, score})
		}
	}

	if len(hits) == 0 {
		return nil
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return s.musics[hits[i].idx].ID > s.musics[hits[j].idx].ID
	})

	limit := len(hits)
	if limit > maxSearchResults {
		limit = maxSearchResults
	}

	out := make([]MusicInfo, limit)
	for i := 0; i < limit; i++ {
		out[i] = s.musics[hits[i].idx]
	}
	return out
}

// SearchEvents searches events by keyword, matching against name,
// event type, unit, and numeric ID.
func (s *Store) SearchEvents(keyword string) []EventInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil
	}

	// Direct ID lookup
	if id, err := strconv.Atoi(keyword); err == nil {
		if ev, ok := s.eventByID[id]; ok {
			return []EventInfo{*ev}
		}
	}

	type scored struct {
		idx   int
		score int
	}
	var hits []scored

	// Map friendly event-type keywords
	eventTypeAliases := map[string]string{
		"马拉松":   "marathon",
		"对战":    "cheerful_carnival",
		"欢乐嘉年华": "cheerful_carnival",
		"wl":    "world_bloom",
		"世界开花":  "world_bloom",
	}

	resolvedType := ""
	for alias, etype := range eventTypeAliases {
		if normalize(keyword) == normalize(alias) {
			resolvedType = etype
			break
		}
	}

	for i := range s.events {
		ev := &s.events[i]
		score := bestScore(keyword, ev.Name, ev.AssetbundleName, ev.Unit)

		// Event type match
		if resolvedType != "" && ev.EventType == resolvedType {
			if score < 75 {
				score = 75
			}
		} else if ts := fuzzyScore(ev.EventType, keyword); ts > 0 && ts > score {
			score = ts
		}

		if score > 0 {
			hits = append(hits, scored{i, score})
		}
	}

	if len(hits) == 0 {
		return nil
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return s.events[hits[i].idx].ID > s.events[hits[j].idx].ID
	})

	limit := len(hits)
	if limit > maxSearchResults {
		limit = maxSearchResults
	}

	out := make([]EventInfo, limit)
	for i := 0; i < limit; i++ {
		out[i] = s.events[hits[i].idx]
	}
	return out
}

// SearchGachas searches gachas by keyword, matching against name,
// gacha type, and numeric ID.
func (s *Store) SearchGachas(keyword string) []GachaInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil
	}

	// Direct ID lookup
	if id, err := strconv.Atoi(keyword); err == nil {
		if g, ok := s.gachaByID[id]; ok {
			return []GachaInfo{*g}
		}
	}

	type scored struct {
		idx   int
		score int
	}
	var hits []scored

	// Gacha type aliases
	gachaTypeAliases := map[string]string{
		"限定":  "limited",
		"常驻":  "permanent",
		"fes": "festival",
	}

	resolvedType := ""
	for alias, gtype := range gachaTypeAliases {
		if normalize(keyword) == normalize(alias) {
			resolvedType = gtype
			break
		}
	}

	for i := range s.gachas {
		g := &s.gachas[i]
		score := bestScore(keyword, g.Name, g.AssetbundleName)

		// Gacha type match
		if resolvedType != "" && strings.Contains(normalize(g.GachaType), resolvedType) {
			if score < 70 {
				score = 70
			}
		} else if ts := fuzzyScore(g.GachaType, keyword); ts > 0 && ts > score {
			score = ts
		}

		if score > 0 {
			hits = append(hits, scored{i, score})
		}
	}

	if len(hits) == 0 {
		return nil
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return s.gachas[hits[i].idx].ID > s.gachas[hits[j].idx].ID
	})

	limit := len(hits)
	if limit > maxSearchResults {
		limit = maxSearchResults
	}

	out := make([]GachaInfo, limit)
	for i := 0; i < limit; i++ {
		out[i] = s.gachas[hits[i].idx]
	}
	return out
}

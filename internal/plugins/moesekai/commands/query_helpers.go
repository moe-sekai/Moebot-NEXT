package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

const listPageSize = 12

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

var unitAliases = map[string]string{
	"vs": "piapro", "虚拟歌手": "piapro", "virtual singer": "piapro", "piapro": "piapro",
	"ln": "light_sound", "l/n": "light_sound", "leo/need": "light_sound", "leoneed": "light_sound", "狮子": "light_sound",
	"mmj": "idol", "more more jump": "idol", "moremorejump": "idol", "偶像": "idol",
	"vbs": "street", "vivid bad squad": "street", "vividbadsquad": "street", "街头": "street",
	"wxs": "theme_park", "ws": "theme_park", "wonderlands": "theme_park", "wonderlands×showtime": "theme_park", "马戏团": "theme_park", "游乐园": "theme_park",
	"n25": "school_refusal", "25": "school_refusal", "25时": "school_refusal", "nightcord": "school_refusal", "ニーゴ": "school_refusal",
}

var attrAliases = map[string]string{
	"cute": "cute", "可爱": "cute", "粉": "cute",
	"cool": "cool", "帅气": "cool", "蓝": "cool",
	"pure": "pure", "纯洁": "pure", "绿": "pure",
	"happy": "happy", "快乐": "happy", "橙": "happy", "黄": "happy",
	"mysterious": "mysterious", "神秘": "mysterious", "紫": "mysterious",
}

type searchOptions struct {
	Raw       string
	Keyword   string
	Tokens    []string
	Year      int
	Leak      bool
	Current   bool
	Page      int
	EventID   int
	CardID    int
	Unit      string
	Attr      string
	Rarity    string
	GachaType string
	Rerelease bool
	Recall    bool
}

type musicQueryOptions struct {
	Difficulty string
	Refresh    bool
}

type musicQueryResult struct {
	Music      *masterdata.MusicInfo
	Musics     []masterdata.MusicInfo
	Candidates []masterdata.MusicInfo
	Message    string
}

type searchRuntime interface {
	Store() *masterdata.Store
	Aliases() map[int]assets.MusicAlias
}

func parseSearchOptions(raw string) searchOptions {
	options := searchOptions{Raw: strings.TrimSpace(raw), Keyword: strings.TrimSpace(raw), Page: 1}
	if options.Raw == "" {
		return options
	}
	tokens := strings.Fields(options.Raw)
	remaining := make([]string, 0, len(tokens))
	for _, token := range tokens {
		original := strings.TrimSpace(token)
		lower := normalizeQuery(original)
		switch {
		case lower == "leak" || lower == "剧透" || lower == "未来" || lower == "未实装":
			options.Leak = true
		case lower == "当前" || lower == "现在" || lower == "开放中" || lower == "进行中":
			options.Current = true
		case lower == "复刻":
			options.Rerelease = true
		case lower == "回响" || lower == "echo":
			options.Recall = true
		case strings.HasPrefix(lower, "event") && parsePositiveInt(strings.TrimPrefix(lower, "event")) > 0:
			options.EventID = parsePositiveInt(strings.TrimPrefix(lower, "event"))
		case strings.HasPrefix(lower, "card") && parsePositiveInt(strings.TrimPrefix(lower, "card")) > 0:
			options.CardID = parsePositiveInt(strings.TrimPrefix(lower, "card"))
		case strings.HasPrefix(lower, "p") && parsePositiveInt(strings.TrimPrefix(lower, "p")) > 0:
			options.Page = parsePositiveInt(strings.TrimPrefix(lower, "p"))
		case strings.HasSuffix(lower, "页") && parsePositiveInt(strings.TrimSuffix(lower, "页")) > 0:
			options.Page = parsePositiveInt(strings.TrimSuffix(lower, "页"))
		case looksLikeYear(lower):
			options.Year = parseYear(lower)
		case rarityFromToken(lower) != "":
			options.Rarity = rarityFromToken(lower)
		case attrAliases[lower] != "":
			options.Attr = attrAliases[lower]
		case unitAliases[lower] != "":
			options.Unit = unitAliases[lower]
		case gachaTypeFromToken(lower) != "":
			options.GachaType = gachaTypeFromToken(lower)
		default:
			remaining = append(remaining, original)
		}
	}
	options.Keyword = strings.TrimSpace(strings.Join(remaining, " "))
	options.Tokens = remaining
	return options
}

func parseMusicQuery(raw string) (string, musicQueryOptions) {
	fields := strings.Fields(strings.TrimSpace(raw))
	remaining := make([]string, 0, len(fields))
	var options musicQueryOptions
	for _, field := range fields {
		key := normalizeQuery(field)
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

func searchMusicAdvanced(store *masterdata.Store, aliases map[int]assets.MusicAlias, raw string, diff string) musicQueryResult {
	query := strings.TrimSpace(raw)
	if store == nil || query == "" {
		return musicQueryResult{Message: "搜索文本为空"}
	}
	if id, ok := parseMusicID(query); ok {
		if music := store.GetMusic(id); music != nil && musicHasDifficulty(store, id, diff) {
			return musicQueryResult{Music: music}
		}
		return musicQueryResult{Message: fmt.Sprintf("没有找到曲目 #%d", id)}
	}
	if idx, ok := parseRelativeIndex(query); ok && idx < 0 {
		musics := sortedMusicsByPublished(store.AllMusics(), false)
		pos := len(musics) + idx
		if pos >= 0 && pos < len(musics) && musicHasDifficulty(store, musics[pos].ID, diff) {
			return musicQueryResult{Music: &musics[pos]}
		}
		return musicQueryResult{Message: fmt.Sprintf("找不到倒数第 %d 首已发布曲目", -idx)}
	}
	if eventID, ok := parseEventToken(query); ok {
		links := store.GetEventMusics(eventID)
		musics := make([]masterdata.MusicInfo, 0, len(links))
		for _, link := range links {
			if music := store.GetMusic(link.MusicID); music != nil && musicHasDifficulty(store, music.ID, diff) {
				musics = append(musics, *music)
			}
		}
		if len(musics) > 0 {
			return musicQueryResult{Music: &musics[0], Musics: musics}
		}
		return musicQueryResult{Message: fmt.Sprintf("活动 #%d 没有关联曲目", eventID)}
	}

	clean := normalizeQuery(query)
	for _, music := range store.AllMusics() {
		if !musicHasDifficulty(store, music.ID, diff) {
			continue
		}
		if normalizeQuery(music.Title) == clean || normalizeQuery(music.Pronunciation) == clean || aliasExactMatch(aliases, music.ID, clean) {
			musicCopy := music
			return musicQueryResult{Music: &musicCopy}
		}
	}

	hits := scoreMusics(store.AllMusics(), aliases, query, diff, store)
	if len(hits) == 0 {
		return musicQueryResult{Message: fmt.Sprintf("没有找到与「%s」匹配的曲目", query)}
	}
	best := hits[0].music
	candidates := make([]masterdata.MusicInfo, 0, minInt(4, len(hits)-1))
	for _, hit := range hits[1:minInt(len(hits), 5)] {
		candidates = append(candidates, hit.music)
	}
	message := ""
	if len(candidates) > 0 {
		parts := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			parts = append(parts, fmt.Sprintf("#%d %s", candidate.ID, candidate.Title))
		}
		message = "候选曲目：" + strings.Join(parts, "、")
	}
	return musicQueryResult{Music: &best, Candidates: candidates, Message: message}
}

type musicHit struct {
	music masterdata.MusicInfo
	score int
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
	if normalizeQuery(entry.Title) == cleanQuery {
		return true
	}
	for _, alias := range entry.Aliases {
		if normalizeQuery(alias) == cleanQuery {
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
	t := normalizeQuery(target)
	k := normalizeQuery(keyword)
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

func normalizeQuery(value string) string {
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
	clean := normalizeQuery(query)
	if strings.HasPrefix(clean, "id") {
		clean = strings.TrimPrefix(clean, "id")
	}
	id, err := strconv.Atoi(clean)
	return id, err == nil && id > 0
}

func parseEventToken(query string) (int, bool) {
	clean := normalizeQuery(query)
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

func parsePositiveInt(value string) int {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func parseYear(value string) int {
	value = strings.TrimSuffix(strings.TrimSpace(value), "年")
	if value == "今年" {
		return time.Now().Year()
	}
	if value == "去年" {
		return time.Now().Year() - 1
	}
	if value == "明年" {
		return time.Now().Year() + 1
	}
	year, _ := strconv.Atoi(value)
	return year
}

func looksLikeYear(value string) bool {
	if value == "今年" || value == "去年" || value == "明年" {
		return true
	}
	value = strings.TrimSuffix(value, "年")
	if len(value) != 4 {
		return false
	}
	year, err := strconv.Atoi(value)
	return err == nil && year >= 2020 && year <= 2100
}

func rarityFromToken(token string) string {
	switch token {
	case "1", "1星", "一星", "rarity1", "rarity_1":
		return "rarity_1"
	case "2", "2星", "二星", "rarity2", "rarity_2":
		return "rarity_2"
	case "3", "3星", "三星", "rarity3", "rarity_3":
		return "rarity_3"
	case "4", "4星", "四星", "rarity4", "rarity_4":
		return "rarity_4"
	case "生日", "birthday", "bd", "raritybirthday", "rarity_birthday":
		return "rarity_birthday"
	default:
		return ""
	}
}

func gachaTypeFromToken(token string) string {
	switch token {
	case "fes", "festival", "彩限", "fes限":
		return "festival"
	case "生日", "birthday", "bd":
		return "birthday"
	case "限定", "limited", "限":
		return "limited"
	case "常驻", "normal", "常驻池", "permanent":
		return "normal"
	default:
		return ""
	}
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

func selectedDifficultyPayload(payload renderpayloads.MusicDetailPayload, diff string, chartSourceURL string) renderpayloads.MusicDetailPayload {
	if diff == "" {
		diff = "master"
	}
	for _, d := range payload.Difficulties {
		if d.MusicDifficulty == diff || d.Difficulty == diff {
			payload.SelectedDifficulty = diff
			payload.ChartURL = assets.ChartSourceURL(chartSourceURL, payload.ID, diff)
			break
		}
	}
	return payload
}

func paginate[T any](items []T, page int, pageSize int) ([]T, int, int) {
	if pageSize <= 0 {
		pageSize = listPageSize
	}
	totalPages := maxInt(1, (len(items)+pageSize-1)/pageSize)
	if page <= 0 {
		page = totalPages
	}
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := minInt(len(items), start+pageSize)
	return items[start:end], page, totalPages
}

func sameYear(ms int64, year int) bool {
	if year == 0 {
		return true
	}
	if ms <= 0 {
		return false
	}
	return time.UnixMilli(ms).Year() == year
}

func isFuture(ms int64) bool {
	return ms > time.Now().UnixMilli()
}

func isNowBetween(startAt int64, endAt int64) bool {
	now := time.Now().UnixMilli()
	return startAt <= now && (endAt <= 0 || now <= endAt)
}

func characterNameByID(id int) string {
	if ch := assets.GetCharacterByID(id); ch != nil {
		if ch.NameCN != "" {
			return ch.NameCN
		}
		if ch.NameJP != "" {
			return ch.NameJP
		}
	}
	if id > 0 {
		return fmt.Sprintf("角色 %d", id)
	}
	return "未知角色"
}

func characterDisplayName(id int) string {
	return characterNameByID(id)
}

func cardPayloads(store *masterdata.Store, resolver *assets.Resolver, cards []masterdata.CardInfo) []renderpayloads.CardDetailPayload {
	out := make([]renderpayloads.CardDetailPayload, 0, len(cards))
	for _, card := range cards {
		out = append(out, renderpayloads.BuildCardDetailPayloadWithAssets(store, card, resolver))
	}
	return out
}

func musicPayloads(store *masterdata.Store, resolver *assets.Resolver, musics []masterdata.MusicInfo) []renderpayloads.MusicDetailPayload {
	out := make([]renderpayloads.MusicDetailPayload, 0, len(musics))
	for _, music := range musics {
		out = append(out, renderpayloads.BuildMusicDetailPayloadWithAssets(store, music, resolver))
	}
	return out
}

func eventPayloads(store *masterdata.Store, resolver *assets.Resolver, events []masterdata.EventInfo) []renderpayloads.EventInfoPayload {
	out := make([]renderpayloads.EventInfoPayload, 0, len(events))
	for _, event := range events {
		out = append(out, renderpayloads.BuildEventInfoPayloadWithAssets(store, event, resolver))
	}
	return out
}

func gachaPayloads(store *masterdata.Store, resolver *assets.Resolver, gachas []masterdata.GachaInfo) []renderpayloads.GachaInfoPayload {
	out := make([]renderpayloads.GachaInfoPayload, 0, len(gachas))
	for _, gacha := range gachas {
		out = append(out, renderpayloads.BuildGachaInfoPayloadWithAssets(store, gacha, resolver))
	}
	return out
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

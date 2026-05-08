package cardquery

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

type Mode int

const (
	ModeNone Mode = iota
	ModeDetail
	ModeList
)

type UnitMode string

const (
	UnitAny UnitMode = "any"
	UnitOC  UnitMode = "oc"
	UnitVS  UnitMode = "vs"
)

type Query struct {
	Raw            string
	Mode           Mode
	DetailID       int
	Page           int
	Keyword        string
	Unparsed       string
	CharacterID    int
	Unit           string
	UnitMode       UnitMode
	Attr           string
	Rarity         string
	Supply         string
	Skill          string
	DetailSkillIDs map[int]struct{}
	Year           int
	Leak           bool
	EventID        int
	RelativeCID    int
	RelativeIndex  int
}

type Result struct {
	Cards      []masterdata.CardInfo
	Mode       Mode
	Page       int
	TotalPages int
	Total      int
	Query      Query
	Message    string
}

type cardAlias struct {
	Value string
	Names []string
}

type cardUnitAlias struct {
	Unit string
	Name string
}

var cardAttrAliases = []cardAlias{
	{Value: "cool", Names: []string{"cool", "帅气", "蓝星", "蓝"}},
	{Value: "happy", Names: []string{"happy", "快乐", "橙心", "橙", "黄"}},
	{Value: "mysterious", Names: []string{"mysterious", "神秘", "紫月", "紫"}},
	{Value: "cute", Names: []string{"cute", "可爱", "粉花", "粉"}},
	{Value: "pure", Names: []string{"pure", "纯洁", "绿草", "绿"}},
}

var cardRarityAliases = []cardAlias{
	{Value: "rarity_birthday", Names: []string{"rarity_birthday", "raritybirthday", "生日卡", "birthday", "生日", "bd"}},
	{Value: "rarity_1", Names: []string{"rarity_1", "rarity1", "1星", "一星", "1x", "1"}},
	{Value: "rarity_2", Names: []string{"rarity_2", "rarity2", "2星", "二星", "两星", "2x", "2"}},
	{Value: "rarity_3", Names: []string{"rarity_3", "rarity3", "3星", "三星", "3x", "3"}},
	{Value: "rarity_4", Names: []string{"rarity_4", "rarity4", "4星", "四星", "4x", "4"}},
}

var cardSupplyAliases = []cardAlias{
	{Value: "bloom_festival_limited", Names: []string{"bfes限定", "bfes限", "bfes", "bf"}},
	{Value: "colorful_festival_limited", Names: []string{"cfes限定", "cfes限", "cfes", "cf"}},
	{Value: "festival_limited", Names: []string{"fes限定", "fes限", "fes"}},
	{Value: "unit_event_limited", Names: []string{"worldlink限定", "worldlink", "wl限定", "wl限", "wl"}},
	{Value: "collaboration_limited", Names: []string{"联动限定", "联动"}},
	{Value: "not_limited", Names: []string{"非限定", "非限", "常驻"}},
	{Value: "term_limited", Names: []string{"期间限定", "期间"}},
	{Value: "all_limited", Names: []string{"限定", "限"}},
}

var cardSkillAliases = []cardAlias{
	{Value: "life_recovery", Names: []string{"奶卡", "奶"}},
	{Value: "score_up", Names: []string{"分卡", "分"}},
	{Value: "judgment_up", Names: []string{"判卡", "判"}},
}

var cardDetailSkillAliases = []struct {
	Names []string
	IDs   []int
}{
	{Names: []string{"大分"}, IDs: []int{4}},
	{Names: []string{"p分"}, IDs: []int{11}},
	{Names: []string{"判分"}, IDs: []int{13}},
	{Names: []string{"血分"}, IDs: []int{12}},
	{Names: []string{"组分", "团分"}, IDs: []int{15, 16, 17, 18, 19}},
}

var cardUnitAliases = []cardAlias{
	{Value: "light_sound", Names: []string{"leo/need", "leoneed", "l/n", "ln", "狮子"}},
	{Value: "idol", Names: []string{"moremorejump", "more more jump", "mmj", "偶像"}},
	{Value: "street", Names: []string{"vividbadsquad", "vivid bad squad", "vbs", "街头"}},
	{Value: "theme_park", Names: []string{"wonderlands×showtime", "wonderlandsshowtime", "wonderlands", "wxs", "ws", "马戏团", "游乐园"}},
	{Value: "school_refusal", Names: []string{"nightcord", "25时", "25h", "n25", "25", "ニーゴ"}},
	{Value: "piapro", Names: []string{"virtualsinger", "virtual singer", "虚拟歌手", "piapro", "vs", "v"}},
}

func Resolve(store *masterdata.Store, raw string) Result {
	return resolve(store, raw, true)
}

// ResolveAll resolves a card query without pagination. It is useful for parser
// result rows while Resolve keeps the renderer payload page-sized.
func ResolveAll(store *masterdata.Store, raw string) Result {
	return resolve(store, raw, false)
}

func resolve(store *masterdata.Store, raw string, paged bool) Result {
	query := Parse(store, raw)
	if store == nil || strings.TrimSpace(raw) == "" {
		return Result{Query: query, Mode: query.Mode, Page: 1, TotalPages: 1, Message: "请输入要搜索的卡牌关键词"}
	}
	if query.message() != "" {
		return Result{Query: query, Mode: query.Mode, Page: 1, TotalPages: 1, Message: query.message()}
	}
	if query.Mode == ModeDetail {
		card := store.GetCard(query.DetailID)
		if card == nil {
			return Result{Query: query, Mode: query.Mode, Page: 1, TotalPages: 1, Message: fmt.Sprintf("没有找到卡牌 #%d", query.DetailID)}
		}
		return Result{Cards: []masterdata.CardInfo{*card}, Query: query, Mode: query.Mode, Page: 1, TotalPages: 1, Total: 1}
	}

	cards := collectCardsForListQuery(store, query)
	if len(cards) == 0 {
		return Result{Query: query, Mode: ModeList, Page: 1, TotalPages: 1, Message: fmt.Sprintf("没有找到与「%s」匹配的卡牌", strings.TrimSpace(raw))}
	}
	if !paged {
		_, page, totalPages := paginate(cards, query.Page, listPageSize)
		return Result{Cards: cards, Query: query, Mode: ModeList, Page: page, TotalPages: totalPages, Total: len(cards)}
	}
	pagedCards, page, totalPages := paginate(cards, query.Page, listPageSize)
	return Result{Cards: pagedCards, Query: query, Mode: ModeList, Page: page, TotalPages: totalPages, Total: len(cards)}
}

func (q Query) message() string {
	if strings.TrimSpace(q.Unparsed) != "" && q.Keyword == "" {
		return fmt.Sprintf("无法解析的参数：%s", strings.TrimSpace(q.Unparsed))
	}
	return ""
}

func Parse(store *masterdata.Store, raw string) Query {
	raw = strings.TrimSpace(raw)
	query := Query{Raw: raw, Mode: ModeList, Page: 1, UnitMode: UnitAny}
	if raw == "" {
		query.Mode = ModeNone
		return query
	}
	if id, ok := parseStrictPositiveID(raw); ok {
		query.Mode = ModeDetail
		query.DetailID = id
		return query
	}
	query.Mode = ModeList
	work := raw
	if cid, idx, ok := extractRelativeCharacterCard(work); ok {
		query.RelativeCID = cid
		query.RelativeIndex = idx
		work = removeRelativeCharacterCard(work)
	}
	work = extractCardPage(work, &query)
	work = extractCardLeak(work, &query)
	work = extractCardYear(work, &query)
	work = extractCardEventID(work, &query)
	if query.EventID == 0 && store != nil {
		work = extractCardBanEvent(store, work, &query)
	}
	work = extractCardDetailSkill(work, &query)
	work = extractCardAliasValue(work, cardAttrAliases, func(value string) { query.Attr = value })
	work = extractCardAliasValue(work, cardSupplyAliases, func(value string) { query.Supply = value })
	work = extractCardAliasValue(work, cardSkillAliases, func(value string) { query.Skill = value })
	work = extractCardUnit(work, &query)
	work = extractCardAliasValue(work, cardRarityAliases, func(value string) { query.Rarity = value })
	work = extractCardCharacter(work, &query)
	query.Keyword = strings.TrimSpace(work)
	query.Unparsed = strings.TrimSpace(work)
	return query
}

func parseStrictPositiveID(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	id, err := strconv.Atoi(value)
	return id, err == nil && id > 0
}

func collectCardsForListQuery(store *masterdata.Store, query Query) []masterdata.CardInfo {
	if query.RelativeCID > 0 && query.RelativeIndex < 0 && query.hasOnlyRelativeIndex() {
		if card := relativeCharacterCard(store, query.RelativeCID, query.RelativeIndex); card != nil {
			return []masterdata.CardInfo{*card}
		}
		return nil
	}

	candidateIDs := map[int]struct{}{}
	if query.Keyword != "" {
		for _, card := range store.SearchCards(query.Keyword) {
			candidateIDs[card.ID] = struct{}{}
		}
		if len(candidateIDs) == 0 {
			return nil
		}
	}

	eventCards := map[int]struct{}{}
	if query.EventID > 0 {
		for _, link := range store.GetEventCards(query.EventID) {
			eventCards[link.CardID] = struct{}{}
		}
		if len(eventCards) == 0 {
			return nil
		}
	}

	now := time.Now().UnixMilli()
	out := make([]masterdata.CardInfo, 0)
	for _, card := range store.AllCards() {
		if len(candidateIDs) > 0 {
			if _, ok := candidateIDs[card.ID]; !ok {
				continue
			}
		}
		if query.Leak && card.ReleaseAt <= now {
			continue
		}
		if len(eventCards) > 0 {
			if _, ok := eventCards[card.ID]; !ok {
				continue
			}
		}
		if len(query.DetailSkillIDs) > 0 {
			if _, ok := query.DetailSkillIDs[card.SkillID]; !ok {
				continue
			}
		}
		if query.Attr != "" && card.Attr != query.Attr {
			continue
		}
		if query.Supply != "" && !cardMatchesSupply(store, card, query.Supply) {
			continue
		}
		if query.Skill != "" && !cardMatchesSkill(store, card, query.Skill) {
			continue
		}
		if query.Unit != "" && !cardMatchesUnitMode(card, query.Unit, query.UnitMode) {
			continue
		}
		if query.Year > 0 && !sameYear(card.ReleaseAt, query.Year) {
			continue
		}
		if query.Rarity != "" && card.CardRarityType != query.Rarity {
			continue
		}
		if query.CharacterID > 0 && card.CharacterID != query.CharacterID {
			continue
		}
		out = append(out, card)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ReleaseAt != out[j].ReleaseAt {
			return out[i].ReleaseAt > out[j].ReleaseAt
		}
		return out[i].ID > out[j].ID
	})
	return out
}

func (q Query) hasOnlyRelativeIndex() bool {
	return q.Keyword == "" && q.Unparsed == "" && q.CharacterID == 0 && q.Unit == "" && q.Attr == "" && q.Rarity == "" && q.Supply == "" && q.Skill == "" && q.Year == 0 && !q.Leak && q.EventID == 0 && len(q.DetailSkillIDs) == 0
}

func cardMatchesSupply(store *masterdata.Store, card masterdata.CardInfo, supply string) bool {
	supplyType := renderpayloads.CardSupplyType(store, card)
	if card.CardRarityType == "rarity_birthday" && supply == "birthday" {
		return true
	}
	switch supply {
	case "festival_limited":
		return supplyType == "bloom_festival_limited" || supplyType == "colorful_festival_limited"
	case "all_limited":
		return supplyType == "term_limited" || supplyType == "colorful_festival_limited" || supplyType == "bloom_festival_limited" || supplyType == "unit_event_limited" || supplyType == "collaboration_limited"
	case "not_limited":
		return supplyType == "normal" || supplyType == ""
	default:
		return supplyType == supply
	}
}

func cardMatchesSkill(store *masterdata.Store, card masterdata.CardInfo, skill string) bool {
	if store != nil {
		if s := store.GetSkill(card.SkillID); s != nil {
			if strings.EqualFold(s.DescriptionSpriteName, skill) {
				return true
			}
			desc := strings.ToLower(s.Description)
			switch skill {
			case "life_recovery":
				if strings.Contains(desc, "life") || strings.Contains(desc, "ライフ") || strings.Contains(desc, "生命") || strings.Contains(desc, "体力") {
					return true
				}
			case "judgment_up":
				if strings.Contains(desc, "判定") || strings.Contains(desc, "judgment") || strings.Contains(desc, "perfect") && strings.Contains(desc, "great") {
					return true
				}
			case "score_up":
				if strings.Contains(desc, "score") || strings.Contains(desc, "スコア") || strings.Contains(desc, "得分") || strings.Contains(desc, "分数") {
					return true
				}
			}
		}
	}
	return false
}

func cardMatchesUnitMode(card masterdata.CardInfo, unit string, mode UnitMode) bool {
	mainUnit := ""
	if ch := assets.GetCharacterByID(card.CharacterID); ch != nil {
		mainUnit = string(ch.UnitID)
	}
	supportUnit := card.SupportUnit
	if supportUnit == "" {
		supportUnit = "none"
	}
	switch mode {
	case UnitOC:
		return mainUnit == unit && (supportUnit == "none" || supportUnit == "")
	case UnitVS:
		if unit == "piapro" {
			return mainUnit == "piapro" && (supportUnit == "none" || supportUnit == "")
		}
		return mainUnit == "piapro" && supportUnit == unit
	default:
		return mainUnit == unit || supportUnit == unit
	}
}

func extractCardPage(text string, query *Query) string {
	for _, field := range strings.Fields(text) {
		lower := normalizeCardText(field)
		page := 0
		if strings.HasPrefix(lower, "@") {
			page = parsePositiveInt(strings.TrimPrefix(lower, "@"))
		} else if strings.HasPrefix(lower, "p") {
			page = parsePositiveInt(strings.TrimPrefix(lower, "p"))
		} else if strings.HasSuffix(lower, "页") {
			page = parsePositiveInt(strings.TrimSuffix(lower, "页"))
		}
		if page > 0 {
			query.Page = page
			text = strings.Replace(text, field, "", 1)
			break
		}
	}
	return strings.TrimSpace(text)
}

func extractCardLeak(text string, query *Query) string {
	return extractCardKeyword(text, []string{"未实装", "leak", "剧透", "未来"}, func() { query.Leak = true })
}

func extractCardYear(text string, query *Query) string {
	nowYear := time.Now().Year()
	relative := []struct {
		name string
		year int
	}{
		{"明年", nowYear + 1},
		{"今年", nowYear},
		{"去年", nowYear - 1},
		{"前年", nowYear - 2},
	}
	for _, item := range relative {
		if containsCardText(text, item.name) {
			query.Year = item.year
			return removeCardText(text, item.name)
		}
	}
	for year := nowYear + 1; year >= 2020; year-- {
		full := fmt.Sprintf("%d年", year)
		short := fmt.Sprintf("%02d年", year%100)
		if containsCardText(text, full) {
			query.Year = year
			return removeCardText(text, full)
		}
		if containsCardText(text, short) {
			query.Year = year
			return removeCardText(text, short)
		}
	}
	return text
}

func extractCardEventID(text string, query *Query) string {
	normalized := normalizeCardText(text)
	idx := strings.Index(normalized, "event")
	if idx < 0 {
		return text
	}
	origRunes := []rune(text)
	normRunes := []rune(normalized)
	if idx >= len(normRunes) {
		return text
	}
	j := idx + len([]rune("event"))
	for j < len(normRunes) && normRunes[j] >= '0' && normRunes[j] <= '9' {
		j++
	}
	if j == idx+len([]rune("event")) {
		return text
	}
	id := parsePositiveInt(string(normRunes[idx+len([]rune("event")) : j]))
	if id <= 0 {
		return text
	}
	query.EventID = id
	if len(origRunes) == len(normRunes) {
		return strings.TrimSpace(string(origRunes[:idx]) + string(origRunes[j:]))
	}
	return removeCardText(text, fmt.Sprintf("event%d", id))
}

func extractCardBanEvent(store *masterdata.Store, text string, query *Query) string {
	for _, entry := range assets.CharacterAliasEntries() {
		alias := entry.Normalized
		if alias == "" {
			continue
		}
		for i := 1; i <= 9; i++ {
			needle := fmt.Sprintf("%s%d", alias, i)
			if !containsCardTextStrict(text, needle) {
				continue
			}
			events := banEventsForCharacter(store, entry.CharacterID)
			if i <= len(events) {
				query.EventID = events[i-1].ID
				return removeCardTextStrict(text, needle)
			}
		}
	}
	return text
}

func banEventsForCharacter(store *masterdata.Store, characterID int) []masterdata.EventInfo {
	if store == nil || characterID <= 0 {
		return nil
	}
	events := make([]masterdata.EventInfo, 0)
	for _, event := range store.AllEvents() {
		links := store.GetEventCards(event.ID)
		if len(links) == 0 {
			continue
		}
		cardID := 0
		for _, link := range links {
			if cardID == 0 || link.CardID < cardID {
				cardID = link.CardID
			}
		}
		if card := store.GetCard(cardID); card != nil && card.CharacterID == characterID {
			events = append(events, event)
		}
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].StartAt != events[j].StartAt {
			return events[i].StartAt < events[j].StartAt
		}
		return events[i].ID < events[j].ID
	})
	return events
}

func extractCardDetailSkill(text string, query *Query) string {
	for _, item := range cardDetailSkillAliases {
		for _, name := range item.Names {
			if containsCardText(text, name) {
				query.DetailSkillIDs = make(map[int]struct{}, len(item.IDs))
				for _, id := range item.IDs {
					query.DetailSkillIDs[id] = struct{}{}
				}
				return removeCardText(text, name)
			}
		}
	}
	return text
}

func extractCardAliasValue(text string, aliases []cardAlias, set func(string)) string {
	flat := flattenCardAliases(aliases)
	for _, item := range flat {
		if containsCardText(text, item.Name) {
			set(item.Value)
			return removeCardText(text, item.Name)
		}
	}
	return text
}

func extractCardUnit(text string, query *Query) string {
	vsAliases := make([]cardUnitAlias, 0)
	ocAliases := make([]cardUnitAlias, 0)
	plainAliases := make([]cardUnitAlias, 0)
	for _, alias := range cardUnitAliases {
		for _, name := range alias.Names {
			normalized := normalizeCardText(name)
			vsAliases = append(vsAliases, cardUnitAlias{Unit: alias.Value, Name: normalized + "vs"}, cardUnitAlias{Unit: alias.Value, Name: normalized + "v"})
			ocAliases = append(ocAliases, cardUnitAlias{Unit: alias.Value, Name: normalized + "oc"}, cardUnitAlias{Unit: alias.Value, Name: "纯" + normalized})
			plainAliases = append(plainAliases, cardUnitAlias{Unit: alias.Value, Name: normalized})
		}
	}
	sortCardUnitAliases(vsAliases)
	sortCardUnitAliases(ocAliases)
	sortCardUnitAliases(plainAliases)
	for _, item := range vsAliases {
		if item.Name == "vv" {
			continue
		}
		if strings.Contains(normalizeCardText(text), item.Name) {
			query.Unit = item.Unit
			query.UnitMode = UnitVS
			return removeCardText(text, item.Name)
		}
	}
	for _, item := range ocAliases {
		if strings.Contains(normalizeCardText(text), item.Name) {
			query.Unit = item.Unit
			query.UnitMode = UnitOC
			return removeCardText(text, item.Name)
		}
	}
	for _, item := range plainAliases {
		if strings.Contains(normalizeCardText(text), item.Name) {
			query.Unit = item.Unit
			query.UnitMode = UnitAny
			return removeCardText(text, item.Name)
		}
	}
	return text
}

func extractCardCharacter(text string, query *Query) string {
	normalized := normalizeCardText(text)
	for _, entry := range assets.CharacterAliasEntries() {
		if entry.Normalized == "" {
			continue
		}
		if strings.Contains(normalized, entry.Normalized) {
			query.CharacterID = entry.CharacterID
			return removeCardText(text, entry.Normalized)
		}
	}
	return text
}

func extractRelativeCharacterCard(text string) (int, int, bool) {
	normalized := normalizeCardText(text)
	for _, entry := range assets.CharacterAliasEntries() {
		alias := entry.Normalized
		if alias == "" || !strings.HasPrefix(normalized, alias) {
			continue
		}
		rest := strings.TrimPrefix(normalized, alias)
		idx, err := strconv.Atoi(rest)
		if err == nil && idx < 0 {
			return entry.CharacterID, idx, true
		}
	}
	return 0, 0, false
}

func removeRelativeCharacterCard(text string) string {
	_, _, ok := extractRelativeCharacterCard(text)
	if !ok {
		return text
	}
	return ""
}

func relativeCharacterCard(store *masterdata.Store, characterID int, idx int) *masterdata.CardInfo {
	cards := cardsForCharacter(store, characterID)
	pos := len(cards) + idx
	if pos >= 0 && pos < len(cards) {
		return &cards[pos]
	}
	return nil
}

type flattenedCardAlias struct {
	Value string
	Name  string
}

func flattenCardAliases(aliases []cardAlias) []flattenedCardAlias {
	out := make([]flattenedCardAlias, 0)
	for _, alias := range aliases {
		for _, name := range alias.Names {
			normalized := normalizeCardText(name)
			if normalized == "" {
				continue
			}
			out = append(out, flattenedCardAlias{Value: alias.Value, Name: normalized})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if len([]rune(out[i].Name)) != len([]rune(out[j].Name)) {
			return len([]rune(out[i].Name)) > len([]rune(out[j].Name))
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func sortCardUnitAliases(items []cardUnitAlias) {
	sort.SliceStable(items, func(i, j int) bool {
		if len([]rune(items[i].Name)) != len([]rune(items[j].Name)) {
			return len([]rune(items[i].Name)) > len([]rune(items[j].Name))
		}
		return items[i].Name < items[j].Name
	})
}

func extractCardKeyword(text string, names []string, apply func()) string {
	flat := make([]string, 0, len(names))
	for _, name := range names {
		flat = append(flat, normalizeCardText(name))
	}
	sort.SliceStable(flat, func(i, j int) bool { return len([]rune(flat[i])) > len([]rune(flat[j])) })
	for _, name := range flat {
		if containsCardText(text, name) {
			apply()
			return removeCardText(text, name)
		}
	}
	return text
}

func containsCardText(text string, needle string) bool {
	needle = normalizeCardText(needle)
	return needle != "" && strings.Contains(normalizeCardText(text), needle)
}

func containsCardTextStrict(text string, needle string) bool {
	needle = normalizeCardText(needle)
	return needle != "" && strings.Contains(normalizeCardTextStrict(text), needle)
}

func removeCardTextStrict(text string, needle string) string {
	needle = normalizeCardText(needle)
	if needle == "" {
		return text
	}
	start := strings.Index(normalizeCardTextStrict(text), needle)
	if start < 0 {
		return text
	}
	runes := []rune(text)
	needleLen := len([]rune(needle))
	if start+needleLen > len(runes) {
		return text
	}
	return strings.TrimSpace(string(runes[:start]) + string(runes[start+needleLen:]))
}

func removeCardText(text string, needle string) string {
	needle = normalizeCardText(needle)
	if needle == "" {
		return text
	}
	runes := []rune(text)
	normalizedRunes := make([]rune, 0, len(runes))
	positions := make([]int, 0, len(runes))
	for idx, r := range runes {
		mapped, keep := normalizeCardRune(r)
		if !keep {
			continue
		}
		normalizedRunes = append(normalizedRunes, mapped)
		positions = append(positions, idx)
	}
	normalized := string(normalizedRunes)
	start := strings.Index(normalized, needle)
	if start < 0 {
		return text
	}
	needleLen := len([]rune(needle))
	if start >= len(positions) || start+needleLen-1 >= len(positions) {
		return text
	}
	origStart := positions[start]
	origEnd := positions[start+needleLen-1] + 1
	return strings.TrimSpace(string(runes[:origStart]) + string(runes[origEnd:]))
}

func normalizeCardText(value string) string {
	value = strings.TrimSpace(value)
	var builder strings.Builder
	for _, r := range value {
		mapped, keep := normalizeCardRune(r)
		if keep {
			builder.WriteRune(mapped)
		}
	}
	return builder.String()
}

func normalizeCardTextStrict(value string) string {
	value = strings.TrimSpace(value)
	var builder strings.Builder
	for _, r := range value {
		if r >= 0xFF01 && r <= 0xFF5E {
			r = r - 0xFEE0
		}
		builder.WriteRune(unicode.ToLower(r))
	}
	return builder.String()
}

func normalizeCardRune(r rune) (rune, bool) {
	if r >= 0xFF01 && r <= 0xFF5E {
		r = r - 0xFEE0
	}
	if unicode.IsSpace(r) || r == '_' || r == '-' || r == '・' || r == '·' || r == '.' || r == '/' || r == '／' || r == '!' || r == '！' {
		return 0, false
	}
	return unicode.ToLower(r), true
}

const listPageSize = 100

func parsePositiveInt(value string) int {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0
	}
	return id
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

func cardsForCharacter(store *masterdata.Store, characterID int) []masterdata.CardInfo {
	cards := make([]masterdata.CardInfo, 0)
	if store == nil {
		return cards
	}
	for _, card := range store.AllCards() {
		if card.CharacterID == characterID {
			cards = append(cards, card)
		}
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].ReleaseAt != cards[j].ReleaseAt {
			return cards[i].ReleaseAt < cards[j].ReleaseAt
		}
		return cards[i].ID < cards[j].ID
	})
	return cards
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

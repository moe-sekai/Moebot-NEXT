package assets

import (
	"sort"
	"strings"
	"unicode"
)

// UnitID identifies a PJSK unit / group. Values match upstream masterdata.
type UnitID string

const (
	UnitLeoNeed             UnitID = "light_sound"
	UnitMoreMoreJump        UnitID = "idol"
	UnitVividBadSquad       UnitID = "street"
	UnitWonderlandsShowtime UnitID = "theme_park"
	UnitNightcordAt25       UnitID = "school_refusal"
	UnitPiapro              UnitID = "piapro"
)

// Unit holds display information for a PJSK group.
type Unit struct {
	ID     UnitID
	NameJP string
	NameEN string
	NameCN string
	Color  string
}

// Units is the canonical list of all six groups.
var Units = []Unit{
	{UnitPiapro, "VIRTUAL SINGER", "VIRTUAL SINGER", "虚拟歌手", "#00BBDD"},
	{UnitLeoNeed, "Leo/need", "Leo/need", "Leo/need", "#4455DD"},
	{UnitMoreMoreJump, "MORE MORE JUMP！", "MORE MORE JUMP!", "MORE MORE JUMP!", "#88DD44"},
	{UnitVividBadSquad, "Vivid BAD SQUAD", "Vivid BAD SQUAD", "Vivid BAD SQUAD", "#EE1166"},
	{UnitWonderlandsShowtime, "ワンダーランズ×ショウタイム", "Wonderlands×Showtime", "Wonderlands×Showtime", "#FF9900"},
	{UnitNightcordAt25, "25時、ナイトコードで。", "25-ji, Nightcord de.", "25时，在Night Code。", "#884499"},
}

var unitMap map[UnitID]*Unit

// GetUnitByID returns the Unit for the given id, or nil.
func GetUnitByID(id UnitID) *Unit {
	return unitMap[id]
}

// Attribute represents a card attribute.
type Attribute string

const (
	AttrCute       Attribute = "cute"
	AttrCool       Attribute = "cool"
	AttrPure       Attribute = "pure"
	AttrHappy      Attribute = "happy"
	AttrMysterious Attribute = "mysterious"
)

// AttributeInfo stores display data for an attribute.
type AttributeInfo struct {
	ID      Attribute
	LabelCN string
	Color   string
}

// Attributes lists all five card attributes.
var Attributes = []AttributeInfo{
	{AttrCute, "可爱", "#FF66B9"},
	{AttrCool, "帅气", "#4455DD"},
	{AttrPure, "纯洁", "#44CC88"},
	{AttrHappy, "快乐", "#FF8800"},
	{AttrMysterious, "神秘", "#BB88EE"},
}

var attrMap map[Attribute]*AttributeInfo

// GetAttributeColor returns the hex color code for the given attribute.
func GetAttributeColor(attr Attribute) string {
	if info, ok := attrMap[attr]; ok {
		return info.Color
	}
	return ""
}

// GetAttributeInfo returns full info for the given attribute, or nil.
func GetAttributeInfo(attr Attribute) *AttributeInfo {
	return attrMap[attr]
}

// Character represents a PJSK character.
type Character struct {
	ID        int
	NameJP    string
	NameEN    string
	NameCN    string
	Aliases   []string
	UnitID    UnitID
	ColorCode string
}

// Characters is the canonical ordered list of the 26 game characters.
var Characters = []Character{
	// Leo/need
	{1, "星乃一歌", "Ichika Hoshino", "星乃一歌", []string{"一歌", "ichika", "ick", "1k", "星乃", "一酱", "一哥"}, UnitLeoNeed, "#33AAEE"},
	{2, "天馬咲希", "Saki Tenma", "天马咲希", []string{"咲希", "saki", "sk", "天马咲希", "小咲", "溪"}, UnitLeoNeed, "#FFDD44"},
	{3, "望月穂波", "Honami Mochizuki", "望月穗波", []string{"穂波", "honami", "hnm", "穗波", "望月穗波", "望月", "穗", "萍"}, UnitLeoNeed, "#EE6666"},
	{4, "日野森志歩", "Shiho Hinomori", "日野森志步", []string{"志歩", "shiho", "sh", "志步", "吸", "吸醬", "吸酱"}, UnitLeoNeed, "#44BBAA"},

	// MORE MORE JUMP!
	{5, "花里みのり", "Minori Hanasato", "花里实乃理", []string{"みのり", "minori", "mnr", "实乃理", "实乃里", "花里实乃里", "花里", "花花"}, UnitMoreMoreJump, "#FFCCAA"},
	{6, "桐谷遥", "Haruka Kiritani", "桐谷遥", []string{"遥", "haruka", "hrk", "桐谷", "企鹅", "鹅"}, UnitMoreMoreJump, "#AACCFF"},
	{7, "桃井愛莉", "Airi Momoi", "桃井爱莉", []string{"愛莉", "airi", "ar", "爱莉", "桃井", "桃"}, UnitMoreMoreJump, "#FF8899"},
	{8, "日野森雫", "Shizuku Hinomori", "日野森雫", []string{"雫", "shizuku", "szk", "水滴", "西子"}, UnitMoreMoreJump, "#77DDCC"},

	// Vivid BAD SQUAD
	{9, "小豆沢こはね", "Kohane Azusawa", "小豆泽心羽", []string{"こはね", "kohane", "khn", "心羽", "小羽", "小豆泽小羽", "小豆沢", "小豆泽", "口哈捏", "仓鼠", "豆"}, UnitVividBadSquad, "#FF6688"},
	{10, "白石杏", "An Shiraishi", "白石杏", []string{"杏", "an", "白石", "安酱"}, UnitVividBadSquad, "#FFBB00"},
	{11, "東雲彰人", "Akito Shinonome", "东云彰人", []string{"彰人", "akito", "akt", "弟弟君", "彰"}, UnitVividBadSquad, "#FF7722"},
	{12, "青柳冬弥", "Toya Aoyagi", "青柳冬弥", []string{"冬弥", "toya", "touya", "ty", "青柳", "冬", "董秘"}, UnitVividBadSquad, "#0077BB"},

	// Wonderlands×Showtime
	{13, "天馬司", "Tsukasa Tenma", "天马司", []string{"司", "tsukasa", "tks", "天马司", "tms", "司马天", "大明星"}, UnitWonderlandsShowtime, "#FFBB33"},
	{14, "鳳えむ", "Emu Otori", "凤笑梦", []string{"えむ", "emu", "em", "笑梦", "凤えむ", "鳳", "凤", "冯笑梦", "姆"}, UnitWonderlandsShowtime, "#FF66BB"},
	{15, "草薙寧々", "Nene Kusanagi", "草薙宁宁", []string{"寧々", "nene", "nn", "宁宁", "草薙", "寧", "宁", "捏", "捏捏"}, UnitWonderlandsShowtime, "#33CC88"},
	{16, "神代類", "Rui Kamishiro", "神代类", []string{"類", "rui", "类", "sdl", "神代"}, UnitWonderlandsShowtime, "#BB88EE"},

	// 25時、ナイトコードで。
	{17, "宵崎奏", "Kanade Yoisaki", "宵崎奏", []string{"奏", "kanade", "knd", "k", "宵崎", "小气走", "走"}, UnitNightcordAt25, "#BB6688"},
	{18, "朝比奈まふゆ", "Mafuyu Asahina", "朝比奈真冬", []string{"まふゆ", "mafuyu", "mfy", "真冬", "朝比奈", "雪", "yuki", "马振东", "马"}, UnitNightcordAt25, "#7788CC"},
	{19, "東雲絵名", "Ena Shinonome", "东云绘名", []string{"絵名", "ena", "绘名", "えな", "enana", "董慧敏", "画"}, UnitNightcordAt25, "#CCAA88"},
	{20, "暁山瑞希", "Mizuki Akiyama", "晓山瑞希", []string{"瑞希", "mizuki", "mzk", "晓山", "暁山", "糖", "肖瑞希", "amia"}, UnitNightcordAt25, "#DD8899"},

	// VIRTUAL SINGER
	{21, "初音ミク", "Hatsune Miku", "初音未来", []string{"初音", "miku", "mk", "未来", "ミク", "39", "葱", "初女士", "米库"}, UnitPiapro, "#33CCBB"},
	{22, "鏡音リン", "Kagamine Rin", "镜音铃", []string{"リン", "rin", "铃", "镜音", "橘"}, UnitPiapro, "#FFCC11"},
	{23, "鏡音レン", "Kagamine Len", "镜音连", []string{"レン", "len", "连", "镜音", "蕉"}, UnitPiapro, "#FFEE11"},
	{24, "巡音ルカ", "Megurine Luka", "巡音流歌", []string{"ルカ", "luka", "流歌", "巡音", "露卡", "鱼"}, UnitPiapro, "#FFAACC"},
	{25, "MEIKO", "MEIKO", "MEIKO", []string{"meiko", "メイコ", "大姐", "酒", "mei"}, UnitPiapro, "#DD4444"},
	{26, "KAITO", "KAITO", "KAITO", []string{"kaito", "カイト", "大哥", "冰", "kai"}, UnitPiapro, "#3366CC"},
}

var charByID map[int]*Character

// CharacterAliasEntry maps a searchable character alias to a character ID.
type CharacterAliasEntry struct {
	Alias       string
	Normalized  string
	CharacterID int
}

var characterAliasEntries []CharacterAliasEntry

func init() {
	unitMap = make(map[UnitID]*Unit, len(Units))
	for i := range Units {
		unitMap[Units[i].ID] = &Units[i]
	}

	attrMap = make(map[Attribute]*AttributeInfo, len(Attributes))
	for i := range Attributes {
		attrMap[Attributes[i].ID] = &Attributes[i]
	}

	charByID = make(map[int]*Character, len(Characters))
	for i := range Characters {
		charByID[Characters[i].ID] = &Characters[i]
	}
	characterAliasEntries = buildCharacterAliasEntries()
}

// GetCharacterByID returns the character with the given ID, or nil.
func GetCharacterByID(id int) *Character {
	return charByID[id]
}

// CharacterAliasEntries returns all character names and aliases sorted by
// normalized length descending so longer aliases win before short nicknames.
func CharacterAliasEntries() []CharacterAliasEntry {
	out := make([]CharacterAliasEntry, len(characterAliasEntries))
	copy(out, characterAliasEntries)
	return out
}

func buildCharacterAliasEntries() []CharacterAliasEntry {
	seen := map[string]struct{}{}
	entries := make([]CharacterAliasEntry, 0, len(Characters)*6)
	for i := range Characters {
		ch := &Characters[i]
		aliases := append([]string{ch.NameCN, ch.NameJP, ch.NameEN}, ch.Aliases...)
		for _, alias := range aliases {
			normalized := NormalizeAlias(alias)
			if normalized == "" {
				continue
			}
			key := normalized + "#" + ch.NameCN
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			entries = append(entries, CharacterAliasEntry{Alias: alias, Normalized: normalized, CharacterID: ch.ID})
		}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if len([]rune(entries[i].Normalized)) != len([]rune(entries[j].Normalized)) {
			return len([]rune(entries[i].Normalized)) > len([]rune(entries[j].Normalized))
		}
		return entries[i].CharacterID < entries[j].CharacterID
	})
	return entries
}

// NormalizeAlias normalizes command aliases for robust matching.
func NormalizeAlias(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Map(func(r rune) rune {
		if r >= 0xFF01 && r <= 0xFF5E {
			r = r - 0xFEE0
		}
		if unicode.IsSpace(r) || r == '_' || r == '-' || r == '・' || r == '·' || r == '.' || r == '/' || r == '／' || r == '!' || r == '！' {
			return -1
		}
		return unicode.ToLower(r)
	}, value)
	return value
}

// FindCharacterByAlias performs a case-insensitive exact/substring search.
func FindCharacterByAlias(query string) *Character {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}

	for i := range Characters {
		ch := &Characters[i]
		if strings.EqualFold(ch.NameCN, query) ||
			strings.EqualFold(ch.NameJP, query) ||
			strings.EqualFold(ch.NameEN, query) {
			return ch
		}
		for _, alias := range ch.Aliases {
			if strings.EqualFold(alias, query) {
				return ch
			}
		}
	}

	for i := range Characters {
		ch := &Characters[i]
		if strings.Contains(strings.ToLower(ch.NameCN), q) ||
			strings.Contains(strings.ToLower(ch.NameJP), q) ||
			strings.Contains(strings.ToLower(ch.NameEN), q) {
			return ch
		}
		for _, alias := range ch.Aliases {
			if strings.Contains(strings.ToLower(alias), q) {
				return ch
			}
		}
	}

	return nil
}

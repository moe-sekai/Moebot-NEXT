package assets

import "strings"

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
	{1, "星乃一歌", "Ichika Hoshino", "星乃一歌", []string{"一歌", "ichika", "ick", "1k"}, UnitLeoNeed, "#33AAEE"},
	{2, "天馬咲希", "Saki Tenma", "天马咲希", []string{"咲希", "saki", "sk", "天马咲希"}, UnitLeoNeed, "#FFDD44"},
	{3, "望月穂波", "Honami Mochizuki", "望月穗波", []string{"穂波", "honami", "hnm", "穗波", "望月穗波"}, UnitLeoNeed, "#EE6666"},
	{4, "日野森志歩", "Shiho Hinomori", "日野森志步", []string{"志歩", "shiho", "sh", "志步"}, UnitLeoNeed, "#44BBAA"},

	// MORE MORE JUMP!
	{5, "花里みのり", "Minori Hanasato", "花里实乃理", []string{"みのり", "minori", "mnr", "实乃理", "花里实乃里"}, UnitMoreMoreJump, "#FFCCAA"},
	{6, "桐谷遥", "Haruka Kiritani", "桐谷遥", []string{"遥", "haruka", "hrk"}, UnitMoreMoreJump, "#AACCFF"},
	{7, "桃井愛莉", "Airi Momoi", "桃井爱莉", []string{"愛莉", "airi", "ar", "爱莉"}, UnitMoreMoreJump, "#FF8899"},
	{8, "日野森雫", "Shizuku Hinomori", "日野森雫", []string{"雫", "shizuku", "szk", "水滴"}, UnitMoreMoreJump, "#77DDCC"},

	// Vivid BAD SQUAD
	{9, "小豆沢こはね", "Kohane Azusawa", "小豆泽心羽", []string{"こはね", "kohane", "khn", "心羽", "小羽", "小豆泽小羽"}, UnitVividBadSquad, "#FF6688"},
	{10, "白石杏", "An Shiraishi", "白石杏", []string{"杏", "an"}, UnitVividBadSquad, "#FFBB00"},
	{11, "東雲彰人", "Akito Shinonome", "东云彰人", []string{"彰人", "akito", "akt"}, UnitVividBadSquad, "#FF7722"},
	{12, "青柳冬弥", "Toya Aoyagi", "青柳冬弥", []string{"冬弥", "toya", "touya", "ty"}, UnitVividBadSquad, "#0077BB"},

	// Wonderlands×Showtime
	{13, "天馬司", "Tsukasa Tenma", "天马司", []string{"司", "tsukasa", "tks", "天马司"}, UnitWonderlandsShowtime, "#FFBB33"},
	{14, "鳳えむ", "Emu Otori", "凤笑梦", []string{"えむ", "emu", "em", "笑梦", "凤えむ"}, UnitWonderlandsShowtime, "#FF66BB"},
	{15, "草薙寧々", "Nene Kusanagi", "草薙宁宁", []string{"寧々", "nene", "nn", "宁宁"}, UnitWonderlandsShowtime, "#33CC88"},
	{16, "神代類", "Rui Kamishiro", "神代类", []string{"類", "rui", "类"}, UnitWonderlandsShowtime, "#BB88EE"},

	// 25時、ナイトコードで。
	{17, "宵崎奏", "Kanade Yoisaki", "宵崎奏", []string{"奏", "kanade", "knd", "k"}, UnitNightcordAt25, "#BB6688"},
	{18, "朝比奈まふゆ", "Mafuyu Asahina", "朝比奈真冬", []string{"まふゆ", "mafuyu", "mfy", "真冬"}, UnitNightcordAt25, "#7788CC"},
	{19, "東雲絵名", "Ena Shinonome", "东云绘名", []string{"絵名", "ena", "绘名"}, UnitNightcordAt25, "#CCAA88"},
	{20, "暁山瑞希", "Mizuki Akiyama", "晓山瑞希", []string{"瑞希", "mizuki", "mzk"}, UnitNightcordAt25, "#DD8899"},

	// VIRTUAL SINGER
	{21, "初音ミク", "Hatsune Miku", "初音未来", []string{"初音", "miku", "mk", "未来", "ミク", "39"}, UnitPiapro, "#33CCBB"},
	{22, "鏡音リン", "Kagamine Rin", "镜音铃", []string{"リン", "rin", "铃", "镜音"}, UnitPiapro, "#FFCC11"},
	{23, "鏡音レン", "Kagamine Len", "镜音连", []string{"レン", "len", "连"}, UnitPiapro, "#FFEE11"},
	{24, "巡音ルカ", "Megurine Luka", "巡音流歌", []string{"ルカ", "luka", "流歌", "巡音"}, UnitPiapro, "#FFAACC"},
	{25, "MEIKO", "MEIKO", "MEIKO", []string{"meiko", "メイコ"}, UnitPiapro, "#DD4444"},
	{26, "KAITO", "KAITO", "KAITO", []string{"kaito", "カイト"}, UnitPiapro, "#3366CC"},
}

var charByID map[int]*Character

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
}

// GetCharacterByID returns the character with the given ID, or nil.
func GetCharacterByID(id int) *Character {
	return charByID[id]
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

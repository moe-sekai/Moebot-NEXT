package masterdata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// types.go — PJSK masterdata type definitions
// Migrated from Snowy Viewer TypeScript types to idiomatic Go structs.
// All JSON tags match the upstream Sekai masterdata API field names.
// ---------------------------------------------------------------------------

// MasterData is the aggregate container for a full masterdata snapshot.
// It is passed to Store.SetAll to atomically swap all data.
type MasterData struct {
	Cards                             []CardInfo                         `json:"cards"`
	Musics                            []MusicInfo                        `json:"musics"`
	MusicDifficulties                 []MusicDifficulty                  `json:"musicDifficulties"`
	Events                            []EventInfo                        `json:"events"`
	EventDeckBonuses                  []EventDeckBonus                   `json:"eventDeckBonuses"`
	EventCards                        []EventCard                        `json:"eventCards"`
	EventMusics                       []EventMusic                       `json:"eventMusics"`
	VirtualLives                      []VirtualLive                      `json:"virtualLives"`
	Gachas                            []GachaInfo                        `json:"gachas"`
	CardSupplies                      []CardSupplyInfo                   `json:"cardSupplies"`
	Skills                            []SkillInfo                        `json:"skills"`
	CharacterUnits                    []GameCharacterUnit                `json:"gameCharacterUnits"`
	Honors                            []HonorInfo                        `json:"honors"`
	MusicVocals                       []MusicVocal                       `json:"musicVocals"`
	ChallengeLiveHighScoreRewards     []ChallengeLiveHighScoreReward     `json:"challengeLiveHighScoreRewards"`
	ResourceBoxes                     []ResourceBox                      `json:"resourceBoxes"`
	ResourceBoxDetails                []ResourceBoxDetail                `json:"resourceBoxDetails"`
	CharacterMissionV2ParameterGroups []CharacterMissionV2ParameterGroup `json:"characterMissionV2ParameterGroups"`
}

// ----------------------------- Card types ----------------------------------

// CardInfo represents a PJSK card.
type CardInfo struct {
	ID                              int             `json:"id"`
	Seq                             int             `json:"seq"`
	CharacterID                     int             `json:"characterId"`
	CardRarityType                  string          `json:"cardRarityType"` // rarity_1 .. rarity_4, rarity_birthday
	SpecialTrainingPower1BonusFixed int             `json:"specialTrainingPower1BonusFixed"`
	SpecialTrainingPower2BonusFixed int             `json:"specialTrainingPower2BonusFixed"`
	SpecialTrainingPower3BonusFixed int             `json:"specialTrainingPower3BonusFixed"`
	Attr                            string          `json:"attr"`        // cute, cool, pure, happy, mysterious
	SupportUnit                     string          `json:"supportUnit"` // e.g. "piapro", "light_sound", ...
	SkillID                         int             `json:"skillId"`
	CardSkillName                   string          `json:"cardSkillName"`
	Prefix                          string          `json:"prefix"` // card title / 称号
	AssetbundleName                 string          `json:"assetbundleName"`
	GachaPhrase                     string          `json:"gachaPhrase"`
	FlavorText                      string          `json:"flavorText"`
	CardParameters                  []CardParameter `json:"cardParameters"`
	ReleaseAt                       int64           `json:"releaseAt"`          // Unix ms
	ArchivePublishedAt              int64           `json:"archivePublishedAt"` // Unix ms
	CardSupplyID                    int             `json:"cardSupplyId"`
}

// UnmarshalJSON accepts both JP-style cardParameters arrays and CN/TW-style
// objects like {"param1":[...],"param2":[...],"param3":[...]}.
func (c *CardInfo) UnmarshalJSON(data []byte) error {
	type alias CardInfo
	var raw struct {
		*alias
		CardParameters json.RawMessage `json:"cardParameters"`
	}
	raw.alias = (*alias)(c)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	params, err := parseCardParameters(raw.CardParameters, c.ID)
	if err != nil {
		return err
	}
	c.CardParameters = params
	return nil
}

// RarityStars returns the numeric star count for display, e.g. 4 for "rarity_4".
func (c *CardInfo) RarityStars() int {
	switch c.CardRarityType {
	case "rarity_1":
		return 1
	case "rarity_2":
		return 2
	case "rarity_3":
		return 3
	case "rarity_4":
		return 4
	case "rarity_birthday":
		return 4 // birthday cards are treated as 4-star
	default:
		return 0
	}
}

// IsBirthday reports whether this card is a birthday card.
func (c *CardInfo) IsBirthday() bool {
	return c.CardRarityType == "rarity_birthday"
}

// CardParameter holds per-level power parameters for a card.
type CardParameter struct {
	ID                int    `json:"id"`
	CardID            int    `json:"cardId"`
	CardLevel         int    `json:"cardLevel"`
	CardParameterType string `json:"cardParameterType"`
	Power             int    `json:"power"`
}

// ----------------------------- Music types ---------------------------------

// MusicInfo represents a PJSK song.
type MusicInfo struct {
	ID                    int      `json:"id"`
	Seq                   int      `json:"seq"`
	Title                 string   `json:"title"`
	Pronunciation         string   `json:"pronunciation"`
	Lyricist              string   `json:"lyricist"`
	Composer              string   `json:"composer"`
	Arranger              string   `json:"arranger"`
	AssetbundleName       string   `json:"assetbundleName"`
	Categories            []string `json:"categories"`
	PublishedAt           int64    `json:"publishedAt"` // Unix ms
	ReleasedAt            int64    `json:"releasedAt"`  // Unix ms
	SecForMusicScoreMaker int      `json:"secForMusicScoreMaker"`
	FillerSec             float64  `json:"fillerSec"`
	IsNewlyWrittenMusic   bool     `json:"isNewlyWrittenMusic"`
	IsFullLength          bool     `json:"isFullLength"`
}

// UnmarshalJSON accepts both string-array categories and object-array
// categories like [{"musicCategoryName":"mv"}] used by some regions.
func (m *MusicInfo) UnmarshalJSON(data []byte) error {
	type alias MusicInfo
	var raw struct {
		*alias
		Categories json.RawMessage `json:"categories"`
	}
	raw.alias = (*alias)(m)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	categories, err := parseMusicCategories(raw.Categories)
	if err != nil {
		return err
	}
	m.Categories = categories
	return nil
}

// MusicDifficulty holds note-chart info for one difficulty of a song.
type MusicDifficulty struct {
	ID              int    `json:"id"`
	MusicID         int    `json:"musicId"`
	MusicDifficulty string `json:"musicDifficulty"` // easy, normal, hard, expert, master, append
	PlayLevel       int    `json:"playLevel"`
	TotalNoteCount  int    `json:"totalNoteCount"`
}

// MusicVocal represents a vocal version of a song.
type MusicVocal struct {
	ID              int    `json:"id"`
	MusicID         int    `json:"musicId"`
	Caption         string `json:"caption"`
	AssetbundleName string `json:"assetbundleName"`
}

func parseCardParameters(raw json.RawMessage, cardID int) ([]CardParameter, error) {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}

	var params []CardParameter
	if err := json.Unmarshal(raw, &params); err == nil {
		return params, nil
	}

	var grouped map[string][]int
	if err := json.Unmarshal(raw, &grouped); err == nil {
		keys := make([]string, 0, len(grouped))
		for key := range grouped {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		params = make([]CardParameter, 0)
		for _, key := range keys {
			for i, power := range grouped[key] {
				params = append(params, CardParameter{
					CardID:            cardID,
					CardLevel:         i + 1,
					CardParameterType: key,
					Power:             power,
				})
			}
		}
		return params, nil
	}

	return nil, fmt.Errorf("unsupported cardParameters shape")
}

func parseMusicCategories(raw json.RawMessage) ([]string, error) {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}

	var stringsOnly []string
	if err := json.Unmarshal(raw, &stringsOnly); err == nil {
		return compactStrings(stringsOnly), nil
	}

	var objects []map[string]any
	if err := json.Unmarshal(raw, &objects); err != nil {
		return nil, err
	}

	categories := make([]string, 0, len(objects))
	for _, object := range objects {
		for _, key := range []string{"musicCategoryName", "category", "name"} {
			if value, ok := object[key].(string); ok && strings.TrimSpace(value) != "" {
				categories = append(categories, value)
				break
			}
		}
	}
	return compactStrings(categories), nil
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

// ----------------------------- Event types ---------------------------------

// EventInfo represents a PJSK event.
type EventInfo struct {
	ID                int    `json:"id"`
	EventType         string `json:"eventType"` // marathon, cheerful_carnival, world_bloom
	Name              string `json:"name"`
	AssetbundleName   string `json:"assetbundleName"`
	StartAt           int64  `json:"startAt"`           // Unix ms
	AggregateAt       int64  `json:"aggregateAt"`       // Unix ms
	ClosedAt          int64  `json:"closedAt"`          // Unix ms
	DistributionEndAt int64  `json:"distributionEndAt"` // Unix ms
	VirtualLiveID     int    `json:"virtualLiveId"`
	Unit              string `json:"unit"`
}

// EventDeckBonus defines bonus rates for cards in an event.
type EventDeckBonus struct {
	ID                  int     `json:"id"`
	EventID             int     `json:"eventId"`
	GameCharacterUnitID int     `json:"gameCharacterUnitId"`
	CardAttr            string  `json:"cardAttr"`
	BonusRate           float64 `json:"bonusRate"`
}

// EventCard links an event to its featured cards.
type EventCard struct {
	ID      int `json:"id"`
	EventID int `json:"eventId"`
	CardID  int `json:"cardId"`
}

// EventMusic links an event to its written / associated music.
type EventMusic struct {
	ID      int `json:"id"`
	EventID int `json:"eventId"`
	MusicID int `json:"musicId"`
	Seq     int `json:"seq"`
}

// ----------------------------- Virtual Live types ---------------------------

// VirtualLive represents an in-game virtual live / after live schedule group.
type VirtualLive struct {
	ID                    int                    `json:"id"`
	Name                  string                 `json:"name"`
	AssetbundleName       string                 `json:"assetbundleName"`
	StartAt               int64                  `json:"startAt"`
	EndAt                 int64                  `json:"endAt"`
	VirtualLiveType       string                 `json:"virtualLiveType"`
	VirtualLiveSchedules  []VirtualLiveSchedule  `json:"virtualLiveSchedules"`
	VirtualLiveRewards    []VirtualLiveReward    `json:"virtualLiveRewards"`
	VirtualLiveCharacters []VirtualLiveCharacter `json:"virtualLiveCharacters"`
}

// VirtualLiveSchedule is one playable/viewable time window for a virtual live.
type VirtualLiveSchedule struct {
	ID            int   `json:"id"`
	VirtualLiveID int   `json:"virtualLiveId"`
	Seq           int   `json:"seq"`
	StartAt       int64 `json:"startAt"`
	EndAt         int64 `json:"endAt"`
}

// VirtualLiveReward identifies a reward resource box for a virtual live.
type VirtualLiveReward struct {
	ID              int    `json:"id"`
	VirtualLiveID   int    `json:"virtualLiveId"`
	VirtualLiveType string `json:"virtualLiveType"`
	ResourceBoxID   int    `json:"resourceBoxId"`
}

// VirtualLiveCharacter describes a character appearing in a virtual live.
type VirtualLiveCharacter struct {
	ID                            int    `json:"id"`
	VirtualLiveID                 int    `json:"virtualLiveId"`
	GameCharacterUnitID           int    `json:"gameCharacterUnitId"`
	GameCharacterID               int    `json:"gameCharacterId"`
	SubGameCharacter2dID          int    `json:"subGameCharacter2dId"`
	VirtualLivePerformanceType    string `json:"virtualLivePerformanceType"`
	VirtualLiveCharacterGroupType string `json:"virtualLiveCharacterGroupType"`
}

// ----------------------------- Gacha types ---------------------------------

// GachaInfo represents a gacha / banner.
type GachaInfo struct {
	ID                   int                   `json:"id"`
	GachaType            string                `json:"gachaType"`
	Name                 string                `json:"name"`
	Seq                  int                   `json:"seq"`
	AssetbundleName      string                `json:"assetbundleName"`
	StartAt              int64                 `json:"startAt"` // Unix ms
	EndAt                int64                 `json:"endAt"`   // Unix ms
	IsShowPeriod         bool                  `json:"isShowPeriod"`
	WishSelectCount      int                   `json:"wishSelectCount"`
	GachaPickups         []GachaPickup         `json:"gachaPickups"`
	GachaCardRarityRates []GachaCardRarityRate `json:"gachaCardRarityRates"`
}

// GachaPickup identifies a featured (pick-up) card in a gacha.
type GachaPickup struct {
	ID              int    `json:"id"`
	GachaID         int    `json:"gachaId"`
	CardID          int    `json:"cardId"`
	GachaPickupType string `json:"gachaPickupType"`
}

// GachaCardRarityRate defines the pull-rate for a given rarity in a gacha.
type GachaCardRarityRate struct {
	ID             int     `json:"id"`
	GachaID        int     `json:"gachaId"`
	GroupID        int     `json:"groupId"`
	CardRarityType string  `json:"cardRarityType"`
	LotteryType    string  `json:"lotteryType"`
	Rate           float64 `json:"rate"`
}

// ----------------------------- Card supply types ----------------------------

// CardSupplyInfo describes how a card is supplied, such as normal, limited, or fes.
type CardSupplyInfo struct {
	ID             int    `json:"id"`
	CardSupplyType string `json:"cardSupplyType"`
}

// ----------------------------- Skill types ---------------------------------

// SkillInfo represents a card's skill.
type SkillInfo struct {
	ID                    int           `json:"id"`
	Description           string        `json:"description"`
	DescriptionSpriteName string        `json:"descriptionSpriteName"`
	SkillEffects          []SkillEffect `json:"skillEffects"`
}

// SkillEffect is one effect component of a skill.
type SkillEffect struct {
	ID                        int                 `json:"id"`
	SkillEffectType           string              `json:"skillEffectType"`
	ActivateNotesJudgmentType string              `json:"activateNotesJudgmentType"`
	SkillEffectDetails        []SkillEffectDetail `json:"skillEffectDetails"`
}

// SkillEffectDetail holds level-specific parameters for a skill effect.
type SkillEffectDetail struct {
	ID                     int     `json:"id"`
	Level                  int     `json:"level"`
	ActivateEffectDuration float64 `json:"activateEffectDuration"`
	ActivateEffectValue    int     `json:"activateEffectValue"`
}

// ----------------------------- Challenge & Mission --------------------------

// ChallengeLiveHighScoreReward defines one challenge-live score threshold reward.
type ChallengeLiveHighScoreReward struct {
	ID            int `json:"id"`
	CharacterID   int `json:"characterId"`
	HighScore     int `json:"highScore"`
	ResourceBoxID int `json:"resourceBoxId"`
}

// ResourceBox groups one or more resources under a specific purpose/id pair.
type ResourceBox struct {
	ResourceBoxPurpose string              `json:"resourceBoxPurpose"`
	ID                 int                 `json:"id"`
	ResourceBoxType    string              `json:"resourceBoxType"`
	Description        string              `json:"description"`
	Details            []ResourceBoxDetail `json:"details"`
}

// ResourceBoxDetail is a concrete resource in an expanded resource box.
type ResourceBoxDetail struct {
	ResourceBoxPurpose string `json:"resourceBoxPurpose"`
	ResourceBoxID      int    `json:"resourceBoxId"`
	Seq                int    `json:"seq"`
	ResourceType       string `json:"resourceType"`
	ResourceID         int    `json:"resourceId"`
	ResourceQuantity   int    `json:"resourceQuantity"`
}

// CharacterMissionV2ParameterGroup defines requirement/exp rows for character missions.
type CharacterMissionV2ParameterGroup struct {
	ID          int `json:"id"`
	Seq         int `json:"seq"`
	Requirement int `json:"requirement"`
	Exp         int `json:"exp"`
	Quantity    int `json:"quantity"`
}

// ----------------------------- Character & Honor ----------------------------

// GameCharacterUnit maps a character to a unit with its theme color.
type GameCharacterUnit struct {
	ID              int    `json:"id"`
	GameCharacterID int    `json:"gameCharacterId"`
	Unit            string `json:"unit"`
	ColorCode       string `json:"colorCode"`
}

// HonorInfo represents an achievement / profile title.
type HonorInfo struct {
	ID              int    `json:"id"`
	GroupID         int    `json:"groupId"`
	HonorRarity     string `json:"honorRarity"` // low, middle, high, highest
	Name            string `json:"name"`
	AssetbundleName string `json:"assetbundleName"`
	HonorType       string `json:"honorType"`
}

package renderer

import (
	"fmt"

	"moebot-next/internal/assets"
	"moebot-next/internal/masterdata"
)

const defaultAssetSource = "main-jp"

// CardDetailPayload is the normalized data contract consumed by CardDetail.tsx.
type CardDetailPayload struct {
	ID               int    `json:"id"`
	Prefix           string `json:"prefix"`
	CharacterName    string `json:"characterName"`
	Rarity           string `json:"rarity"`
	CardRarityType   string `json:"cardRarityType"`
	Attr             string `json:"attr"`
	AssetbundleName  string `json:"assetbundleName,omitempty"`
	CharacterID      int    `json:"characterId,omitempty"`
	Power            int    `json:"power,omitempty"`
	SkillName        string `json:"skillName,omitempty"`
	GachaPhrase      string `json:"gachaPhrase,omitempty"`
	SupplyType       string `json:"supplyType,omitempty"`
	CardSupplyID     int    `json:"cardSupplyId,omitempty"`
	AssetSource      string `json:"assetSource,omitempty"`
	NormalFullURL    string `json:"normalFullUrl,omitempty"`
	TrainedFullURL   string `json:"trainedFullUrl,omitempty"`
	ThumbnailURL     string `json:"thumbnailUrl,omitempty"`
	TrainedThumbnail string `json:"trainedThumbnailUrl,omitempty"`
}

// MusicDetailPayload is the normalized data contract consumed by MusicDetail.tsx.
type MusicDetailPayload struct {
	ID                  int                      `json:"id"`
	Title               string                   `json:"title"`
	Pronunciation       string                   `json:"pronunciation,omitempty"`
	Lyricist            string                   `json:"lyricist,omitempty"`
	Composer            string                   `json:"composer,omitempty"`
	Arranger            string                   `json:"arranger,omitempty"`
	Categories          []string                 `json:"categories,omitempty"`
	AssetbundleName     string                   `json:"assetbundleName,omitempty"`
	Difficulties        []MusicDifficultyPayload `json:"difficulties,omitempty"`
	PublishedAt         int64                    `json:"publishedAt,omitempty"`
	ReleasedAt          int64                    `json:"releasedAt,omitempty"`
	FillerSec           float64                  `json:"fillerSec,omitempty"`
	IsNewlyWrittenMusic bool                     `json:"isNewlyWrittenMusic,omitempty"`
	IsFullLength        bool                     `json:"isFullLength,omitempty"`
	AssetSource         string                   `json:"assetSource,omitempty"`
}

// MusicDifficultyPayload mirrors the fields used by MusicDetail.tsx.
type MusicDifficultyPayload struct {
	ID              int    `json:"id"`
	MusicID         int    `json:"musicId"`
	Difficulty      string `json:"difficulty"`
	MusicDifficulty string `json:"musicDifficulty"`
	Level           int    `json:"level"`
	PlayLevel       int    `json:"playLevel"`
	NoteCount       int    `json:"noteCount,omitempty"`
	TotalNoteCount  int    `json:"totalNoteCount,omitempty"`
}

// EventInfoPayload is the normalized data contract consumed by EventInfo.tsx.
type EventInfoPayload struct {
	ID                int                     `json:"id"`
	Name              string                  `json:"name"`
	EventType         string                  `json:"eventType"`
	Unit              string                  `json:"unit,omitempty"`
	AssetbundleName   string                  `json:"assetbundleName,omitempty"`
	StartAt           int64                   `json:"startAt"`
	AggregateAt       int64                   `json:"aggregateAt,omitempty"`
	ClosedAt          int64                   `json:"closedAt"`
	DistributionEndAt int64                   `json:"distributionEndAt,omitempty"`
	DeckBonuses       []EventDeckBonusPayload `json:"deckBonuses,omitempty"`
	BonusAttr         string                  `json:"bonusAttr,omitempty"`
	BonusCharacters   []string                `json:"bonusCharacters,omitempty"`
	AssetSource       string                  `json:"assetSource,omitempty"`
}

// EventDeckBonusPayload enriches an event bonus row with character/unit labels.
type EventDeckBonusPayload struct {
	ID                  int     `json:"id"`
	EventID             int     `json:"eventId"`
	GameCharacterUnitID int     `json:"gameCharacterUnitId,omitempty"`
	GameCharacterID     int     `json:"gameCharacterId,omitempty"`
	CharacterName       string  `json:"characterName,omitempty"`
	Unit                string  `json:"unit,omitempty"`
	CardAttr            string  `json:"cardAttr,omitempty"`
	BonusRate           float64 `json:"bonusRate"`
}

// GachaInfoPayload is the normalized data contract consumed by GachaInfo.tsx.
type GachaInfoPayload struct {
	ID              int                      `json:"id"`
	Name            string                   `json:"name"`
	GachaType       string                   `json:"gachaType,omitempty"`
	AssetbundleName string                   `json:"assetbundleName,omitempty"`
	StartAt         int64                    `json:"startAt"`
	EndAt           int64                    `json:"endAt"`
	IsShowPeriod    bool                     `json:"isShowPeriod,omitempty"`
	WishSelectCount int                      `json:"wishSelectCount,omitempty"`
	PickupCards     []GachaPickupCardPayload `json:"pickupCards,omitempty"`
	Pickups         []GachaPickupPayload     `json:"pickups,omitempty"`
	Rates           []GachaRatePayload       `json:"rates,omitempty"`
	AssetSource     string                   `json:"assetSource,omitempty"`
}

// GachaPickupPayload preserves the raw pickup relation with optional card detail.
type GachaPickupPayload struct {
	ID              int                     `json:"id"`
	GachaID         int                     `json:"gachaId"`
	CardID          int                     `json:"cardId"`
	GachaPickupType string                  `json:"gachaPickupType,omitempty"`
	Card            *GachaPickupCardPayload `json:"card,omitempty"`
}

// GachaPickupCardPayload is the card shape displayed by GachaInfo.tsx.
type GachaPickupCardPayload struct {
	ID                  int    `json:"id"`
	Prefix              string `json:"prefix,omitempty"`
	CharacterName       string `json:"characterName"`
	Rarity              string `json:"rarity"`
	CardRarityType      string `json:"cardRarityType,omitempty"`
	Attr                string `json:"attr"`
	AssetbundleName     string `json:"assetbundleName,omitempty"`
	CharacterID         int    `json:"characterId,omitempty"`
	ThumbnailURL        string `json:"thumbnailUrl,omitempty"`
	TrainedThumbnailURL string `json:"trainedThumbnailUrl,omitempty"`
	IsWish              bool   `json:"isWish,omitempty"`
	GachaPickupType     string `json:"gachaPickupType,omitempty"`
}

// GachaRatePayload normalizes rarity-rate rows for renderer fallback/extension.
type GachaRatePayload struct {
	ID             int     `json:"id"`
	GachaID        int     `json:"gachaId,omitempty"`
	GroupID        int     `json:"groupId,omitempty"`
	CardRarityType string  `json:"cardRarityType"`
	LotteryType    string  `json:"lotteryType,omitempty"`
	Rate           float64 `json:"rate"`
}

// BuildCardDetailPayload adapts a masterdata card into CardDetail renderer props.
func BuildCardDetailPayload(_ *masterdata.Store, card masterdata.CardInfo) CardDetailPayload {
	payload := CardDetailPayload{
		ID:              card.ID,
		Prefix:          card.Prefix,
		CharacterName:   characterName(card.CharacterID),
		Rarity:          card.CardRarityType,
		CardRarityType:  card.CardRarityType,
		Attr:            card.Attr,
		AssetbundleName: card.AssetbundleName,
		CharacterID:     card.CharacterID,
		Power:           maxCardPower(card),
		SkillName:       card.CardSkillName,
		GachaPhrase:     cleanDash(card.GachaPhrase),
		SupplyType:      cardSupplyType(card.CardSupplyID),
		CardSupplyID:    card.CardSupplyID,
		AssetSource:     defaultAssetSource,
	}

	if card.AssetbundleName != "" {
		payload.NormalFullURL = assets.GetCardFullURL(card.AssetbundleName, false)
		payload.TrainedFullURL = assets.GetCardFullURL(card.AssetbundleName, true)
		payload.ThumbnailURL = assets.GetCardThumbnailURL(card.AssetbundleName, false)
		payload.TrainedThumbnail = assets.GetCardThumbnailURL(card.AssetbundleName, true)
	}

	return payload
}

// BuildMusicDetailPayload adapts a masterdata music row into MusicDetail renderer props.
func BuildMusicDetailPayload(store *masterdata.Store, music masterdata.MusicInfo) MusicDetailPayload {
	payload := MusicDetailPayload{
		ID:                  music.ID,
		Title:               music.Title,
		Pronunciation:       music.Pronunciation,
		Lyricist:            music.Lyricist,
		Composer:            music.Composer,
		Arranger:            music.Arranger,
		Categories:          append([]string(nil), music.Categories...),
		AssetbundleName:     music.AssetbundleName,
		PublishedAt:         music.PublishedAt,
		ReleasedAt:          music.ReleasedAt,
		FillerSec:           music.FillerSec,
		IsNewlyWrittenMusic: music.IsNewlyWrittenMusic,
		IsFullLength:        music.IsFullLength,
		AssetSource:         defaultAssetSource,
	}

	if store != nil {
		for _, d := range store.GetMusicDifficulties(music.ID) {
			payload.Difficulties = append(payload.Difficulties, MusicDifficultyPayload{
				ID:              d.ID,
				MusicID:         d.MusicID,
				Difficulty:      d.MusicDifficulty,
				MusicDifficulty: d.MusicDifficulty,
				Level:           d.PlayLevel,
				PlayLevel:       d.PlayLevel,
				NoteCount:       d.TotalNoteCount,
				TotalNoteCount:  d.TotalNoteCount,
			})
		}
	}

	return payload
}

// BuildEventInfoPayload adapts a masterdata event into EventInfo renderer props.
func BuildEventInfoPayload(store *masterdata.Store, event masterdata.EventInfo) EventInfoPayload {
	payload := EventInfoPayload{
		ID:                event.ID,
		Name:              event.Name,
		EventType:         event.EventType,
		Unit:              event.Unit,
		AssetbundleName:   event.AssetbundleName,
		StartAt:           event.StartAt,
		AggregateAt:       event.AggregateAt,
		ClosedAt:          event.ClosedAt,
		DistributionEndAt: event.DistributionEndAt,
		AssetSource:       defaultAssetSource,
	}

	if store == nil {
		return payload
	}

	seenCharacters := make(map[string]struct{})
	for _, b := range store.GetEventDeckBonuses(event.ID) {
		bonus := EventDeckBonusPayload{
			ID:                  b.ID,
			EventID:             b.EventID,
			GameCharacterUnitID: b.GameCharacterUnitID,
			CardAttr:            b.CardAttr,
			BonusRate:           b.BonusRate,
		}

		if unit := store.GetCharacterUnit(b.GameCharacterUnitID); unit != nil {
			bonus.GameCharacterID = unit.GameCharacterID
			bonus.Unit = unit.Unit
			bonus.CharacterName = characterName(unit.GameCharacterID)
		}

		if payload.BonusAttr == "" && b.CardAttr != "" {
			payload.BonusAttr = b.CardAttr
		}
		if bonus.CharacterName != "" {
			if _, ok := seenCharacters[bonus.CharacterName]; !ok {
				seenCharacters[bonus.CharacterName] = struct{}{}
				payload.BonusCharacters = append(payload.BonusCharacters, bonus.CharacterName)
			}
		}

		payload.DeckBonuses = append(payload.DeckBonuses, bonus)
	}

	return payload
}

// BuildGachaInfoPayload adapts a masterdata gacha into GachaInfo renderer props.
func BuildGachaInfoPayload(store *masterdata.Store, gacha masterdata.GachaInfo) GachaInfoPayload {
	payload := GachaInfoPayload{
		ID:              gacha.ID,
		Name:            gacha.Name,
		GachaType:       gacha.GachaType,
		AssetbundleName: gacha.AssetbundleName,
		StartAt:         gacha.StartAt,
		EndAt:           gacha.EndAt,
		IsShowPeriod:    gacha.IsShowPeriod,
		WishSelectCount: gacha.WishSelectCount,
		AssetSource:     defaultAssetSource,
	}

	for _, rate := range gacha.GachaCardRarityRates {
		payload.Rates = append(payload.Rates, GachaRatePayload{
			ID:             rate.ID,
			GachaID:        firstNonZero(rate.GachaID, gacha.ID),
			GroupID:        rate.GroupID,
			CardRarityType: rate.CardRarityType,
			LotteryType:    rate.LotteryType,
			Rate:           rate.Rate,
		})
	}

	for _, pickup := range gacha.GachaPickups {
		pickupPayload := GachaPickupPayload{
			ID:              pickup.ID,
			GachaID:         firstNonZero(pickup.GachaID, gacha.ID),
			CardID:          pickup.CardID,
			GachaPickupType: pickup.GachaPickupType,
		}

		if store != nil {
			if card := store.GetCard(pickup.CardID); card != nil {
				cardPayload := buildGachaPickupCardPayload(*card, pickup.GachaPickupType)
				pickupPayload.Card = &cardPayload
				payload.PickupCards = append(payload.PickupCards, cardPayload)
			}
		}

		payload.Pickups = append(payload.Pickups, pickupPayload)
	}

	return payload
}

func buildGachaPickupCardPayload(card masterdata.CardInfo, pickupType string) GachaPickupCardPayload {
	payload := GachaPickupCardPayload{
		ID:              card.ID,
		Prefix:          card.Prefix,
		CharacterName:   characterName(card.CharacterID),
		Rarity:          card.CardRarityType,
		CardRarityType:  card.CardRarityType,
		Attr:            card.Attr,
		AssetbundleName: card.AssetbundleName,
		CharacterID:     card.CharacterID,
		IsWish:          true,
		GachaPickupType: pickupType,
	}
	if card.AssetbundleName != "" {
		payload.ThumbnailURL = assets.GetCardThumbnailURL(card.AssetbundleName, false)
		payload.TrainedThumbnailURL = assets.GetCardThumbnailURL(card.AssetbundleName, true)
	}
	return payload
}

func maxCardPower(card masterdata.CardInfo) int {
	maxLevel := 0
	for _, parameter := range card.CardParameters {
		if parameter.CardLevel > maxLevel {
			maxLevel = parameter.CardLevel
		}
	}
	if maxLevel == 0 {
		return card.SpecialTrainingPower1BonusFixed + card.SpecialTrainingPower2BonusFixed + card.SpecialTrainingPower3BonusFixed
	}

	total := 0
	for _, parameter := range card.CardParameters {
		if parameter.CardLevel == maxLevel {
			total += parameter.Power
		}
	}

	return total + card.SpecialTrainingPower1BonusFixed + card.SpecialTrainingPower2BonusFixed + card.SpecialTrainingPower3BonusFixed
}

func characterName(characterID int) string {
	if ch := assets.GetCharacterByID(characterID); ch != nil {
		if ch.NameCN != "" {
			return ch.NameCN
		}
		if ch.NameJP != "" {
			return ch.NameJP
		}
		if ch.NameEN != "" {
			return ch.NameEN
		}
	}
	if characterID > 0 {
		return fmt.Sprintf("角色 %d", characterID)
	}
	return "未知角色"
}

func cardSupplyType(cardSupplyID int) string {
	if cardSupplyID <= 0 {
		return ""
	}
	return fmt.Sprintf("供给 #%d", cardSupplyID)
}

func cleanDash(value string) string {
	if value == "-" {
		return ""
	}
	return value
}

func firstNonZero(value int, fallback int) int {
	if value != 0 {
		return value
	}
	return fallback
}

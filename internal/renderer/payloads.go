package renderer

import (
	"fmt"
	"time"

	"moebot-next/internal/assets"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/ranking"
	"moebot-next/internal/sekai"
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
	DurationSec         int                      `json:"durationSec,omitempty"`
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

type RankingListPayload struct {
	Title       string                `json:"title"`
	Subtitle    string                `json:"subtitle,omitempty"`
	Rankings    []RankingEntryPayload `json:"rankings"`
	EventID     int                   `json:"eventId,omitempty"`
	UpdatedAt   int64                 `json:"updatedAt,omitempty"`
	AssetSource string                `json:"assetSource,omitempty"`
}

type RankingEntryPayload struct {
	Rank                int                     `json:"rank"`
	Name                string                  `json:"name,omitempty"`
	DisplayName         string                  `json:"displayName,omitempty"`
	Signature           string                  `json:"signature,omitempty"`
	Score               int64                   `json:"score"`
	UserID              string                  `json:"userId,omitempty"`
	ScoreDelta          int64                   `json:"scoreDelta,omitempty"`
	LeaderCard          *ProfileDeckCardPayload `json:"leaderCard,omitempty"`
	Churn48h            int                     `json:"churn48h,omitempty"`
	RecentActivityCount int                     `json:"recentActivityCount,omitempty"`
	Growth1h            int64                   `json:"growth1h,omitempty"`
	Speed20m3           int64                   `json:"speed20m3,omitempty"`
	Churn1h             int                     `json:"churn1h,omitempty"`
	Churn20m3           int                     `json:"churn20m3,omitempty"`
	Trend               string                  `json:"trend,omitempty"`
	IsTierLine          bool                    `json:"isTierLine,omitempty"`
}

func BuildRankingListPayload(title string, board ranking.Board) RankingListPayload {
	payload := RankingListPayload{Title: title, EventID: board.EventID, UpdatedAt: board.UpdatedAt, AssetSource: defaultAssetSource}
	for _, entry := range board.Rankings {
		payload.Rankings = append(payload.Rankings, buildRankingEntryPayload(entry))
	}
	return payload
}

func BuildChurnRankingListPayload(board ranking.Board) RankingListPayload {
	payload := RankingListPayload{Title: "查房", Subtitle: "活跃度 / 时速 / 最近分数变化", EventID: board.EventID, UpdatedAt: board.UpdatedAt, AssetSource: defaultAssetSource}
	for _, entry := range board.Rankings {
		item := buildRankingEntryPayload(entry)
		if entry.LastChange != nil {
			item.ScoreDelta = entry.LastChange.Delta
		}
		recentCount := 0
		if entry.RecentActivity != nil {
			recentCount = entry.RecentActivity.Count
		}
		item.Churn48h = entry.Churn48h
		item.RecentActivityCount = recentCount
		item.Growth1h = entry.Growth1h
		item.Speed20m3 = calcRecentGrowth(entry.RecentScoreChanges, 20) * 3
		item.Churn1h = calcRecentChurnCount(entry.RecentScoreChanges, 60)
		item.Churn20m3 = calcRecentChurnCount(entry.RecentScoreChanges, 20) * 3
		if item.Churn1h == 0 {
			item.Churn1h = currentHourChurn(entry.HourlyChurn)
		}
		item.Trend = speedTrend(item.Growth1h, item.Speed20m3)
		item.IsTierLine = entry.Rank > 100 && entry.UserID.String() == ""
		item.Signature = fmt.Sprintf("48H %d · 1H速 %s · 20m×3 %s", entry.Churn48h, formatCompactSpeed(item.Growth1h), formatCompactSpeed(item.Speed20m3))
		payload.Rankings = append(payload.Rankings, item)
	}
	return payload
}

func calcRecentGrowth(changes []ranking.ScoreChange, minutes int) int64 {
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute).UnixMilli()
	var total int64
	for _, change := range changes {
		if normalizeTimestampMillis(change.Time) >= cutoff && change.Delta > 0 {
			total += change.Delta
		}
	}
	return total
}

func calcRecentChurnCount(changes []ranking.ScoreChange, minutes int) int {
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute).UnixMilli()
	count := 0
	for _, change := range changes {
		if normalizeTimestampMillis(change.Time) >= cutoff && change.Delta > 0 {
			count++
		}
	}
	return count
}

func currentHourChurn(hourly []ranking.HourlyChurn) int {
	now := time.Now().UTC().Truncate(time.Hour).Format("2006-01-02T15:04:05Z")
	for _, item := range hourly {
		if item.Hour == now {
			return item.Count
		}
	}
	return 0
}

func speedTrend(speed1h int64, speed20m3 int64) string {
	if speed1h == 0 && speed20m3 == 0 {
		return "flat"
	}
	if speed1h == 0 {
		return "up"
	}
	if speed20m3*100 > speed1h*108 {
		return "up"
	}
	if speed20m3*100 < speed1h*92 {
		return "down"
	}
	return "flat"
}

func normalizeTimestampMillis(value int64) int64 {
	if value > 0 && value < 1_000_000_000_000 {
		return value * 1000
	}
	return value
}

func formatCompactSpeed(value int64) string {
	return fmt.Sprintf("%dk", value/1000)
}

func buildRankingEntryPayload(entry ranking.RankingEntry) RankingEntryPayload {
	payload := RankingEntryPayload{
		Rank:      entry.Rank,
		Name:      entry.Name,
		Signature: entry.Word,
		Score:     entry.Score,
		UserID:    entry.UserID.String(),
	}
	if entry.LeaderCard != nil {
		payload.LeaderCard = &ProfileDeckCardPayload{
			CardID:       entry.LeaderCard.CardID,
			ID:           entry.LeaderCard.CardID,
			Level:        entry.LeaderCard.Level,
			Mastery:      entry.LeaderCard.MasterRank,
			IsTrained:    entry.LeaderCard.DefaultImage == "special_training",
			DefaultImage: entry.LeaderCard.DefaultImage,
		}
	}
	return payload
}

// ProfileCardPayload is the normalized data contract consumed by ProfileCard.tsx.
type ProfileCardPayload struct {
	Name             string                   `json:"name"`
	Rank             int                      `json:"rank"`
	UserID           string                   `json:"userId"`
	TwitterID        string                   `json:"twitterId,omitempty"`
	Signature        string                   `json:"signature,omitempty"`
	TotalPower       int                      `json:"totalPower,omitempty"`
	Stats            *ProfileStatsPayload     `json:"stats,omitempty"`
	MusicClearCounts []MusicClearCountPayload `json:"musicClearCounts,omitempty"`
	CharacterRanks   []CharacterRankPayload   `json:"characterRanks,omitempty"`
	ChallengeLive    *ChallengeLivePayload    `json:"challengeLive,omitempty"`
	ProfileHonors    []ProfileHonorPayload    `json:"profileHonors,omitempty"`
	LeaderCard       *ProfileDeckCardPayload  `json:"leaderCard,omitempty"`
	DeckCards        []ProfileDeckCardPayload `json:"deckCards,omitempty"`
	AssetSource      string                   `json:"assetSource,omitempty"`
}

type ProfileStatsPayload struct {
	MvpCount       int `json:"mvpCount,omitempty"`
	SuperStarCount int `json:"superStarCount,omitempty"`
}

type MusicClearCountPayload struct {
	Difficulty string `json:"difficulty"`
	LiveClear  int    `json:"liveClear"`
	FullCombo  int    `json:"fullCombo"`
	AllPerfect int    `json:"allPerfect"`
}

type CharacterRankPayload struct {
	CharacterID   int    `json:"characterId"`
	CharacterName string `json:"characterName"`
	Rank          int    `json:"rank"`
}

type ChallengeLivePayload struct {
	CharacterID   int    `json:"characterId"`
	CharacterName string `json:"characterName"`
	HighScore     int    `json:"highScore"`
}

type ProfileHonorPayload struct {
	Seq             int    `json:"seq"`
	HonorType       string `json:"honorType,omitempty"`
	HonorID         int    `json:"honorId"`
	Level           int    `json:"level"`
	Name            string `json:"name,omitempty"`
	HonorRarity     string `json:"honorRarity,omitempty"`
	AssetbundleName string `json:"assetbundleName,omitempty"`
}

type ProfileDeckCardPayload struct {
	CardID              int    `json:"cardId,omitempty"`
	ID                  int    `json:"id,omitempty"`
	CharacterName       string `json:"characterName,omitempty"`
	Rarity              string `json:"rarity,omitempty"`
	CardRarityType      string `json:"cardRarityType,omitempty"`
	Attr                string `json:"attr,omitempty"`
	AssetbundleName     string `json:"assetbundleName,omitempty"`
	ThumbnailURL        string `json:"thumbnailUrl,omitempty"`
	TrainedThumbnailURL string `json:"trainedThumbnailUrl,omitempty"`
	Level               int    `json:"level,omitempty"`
	Mastery             int    `json:"mastery,omitempty"`
	IsTrained           bool   `json:"isTrained,omitempty"`
	DefaultImage        string `json:"defaultImage,omitempty"`
}

// BuildProfileCardPayload adapts an API profile into ProfileCard renderer props.
func BuildProfileCardPayload(profile sekai.Profile) ProfileCardPayload {
	return BuildProfileCardPayloadWithStore(nil, profile)
}

// BuildProfileCardPayloadWithStore enriches profile card IDs with masterdata card display fields.
func BuildProfileCardPayloadWithStore(store *masterdata.Store, profile sekai.Profile) ProfileCardPayload {
	payload := ProfileCardPayload{
		Name:       profile.Name,
		Rank:       profile.Rank,
		UserID:     profile.UserID,
		TwitterID:  profile.TwitterID,
		Signature:  profile.Signature,
		TotalPower: profile.TotalPower,
		Stats: &ProfileStatsPayload{
			MvpCount:       profile.Stats.MvpCount,
			SuperStarCount: profile.Stats.SuperStarCount,
		},
		AssetSource: defaultAssetSource,
	}
	for _, count := range profile.MusicClearCounts {
		payload.MusicClearCounts = append(payload.MusicClearCounts, MusicClearCountPayload{
			Difficulty: count.Difficulty,
			LiveClear:  count.LiveClear,
			FullCombo:  count.FullCombo,
			AllPerfect: count.AllPerfect,
		})
	}
	for _, rank := range profile.CharacterRanks {
		payload.CharacterRanks = append(payload.CharacterRanks, CharacterRankPayload{
			CharacterID:   rank.CharacterID,
			CharacterName: characterName(rank.CharacterID),
			Rank:          rank.Rank,
		})
	}
	if profile.ChallengeLive != nil {
		payload.ChallengeLive = &ChallengeLivePayload{
			CharacterID:   profile.ChallengeLive.CharacterID,
			CharacterName: characterName(profile.ChallengeLive.CharacterID),
			HighScore:     profile.ChallengeLive.HighScore,
		}
	}
	for _, honor := range profile.ProfileHonors {
		honorPayload := ProfileHonorPayload{
			Seq:       honor.Seq,
			HonorType: honor.HonorType,
			HonorID:   honor.HonorID,
			Level:     honor.Level,
		}
		if store != nil {
			if masterHonor := store.GetHonor(honor.HonorID); masterHonor != nil {
				honorPayload.Name = masterHonor.Name
				honorPayload.HonorRarity = masterHonor.HonorRarity
				honorPayload.AssetbundleName = masterHonor.AssetbundleName
			}
		}
		payload.ProfileHonors = append(payload.ProfileHonors, honorPayload)
	}
	if profile.LeaderCard != nil {
		leader := buildProfileDeckCardPayload(store, *profile.LeaderCard)
		payload.LeaderCard = &leader
	}
	for _, card := range profile.DeckCards {
		payload.DeckCards = append(payload.DeckCards, buildProfileDeckCardPayload(store, card))
	}
	return payload
}

func buildProfileDeckCardPayload(store *masterdata.Store, card sekai.ProfileDeckCard) ProfileDeckCardPayload {
	payload := ProfileDeckCardPayload{
		CardID:       card.CardID,
		ID:           card.CardID,
		Level:        card.Level,
		Mastery:      card.Mastery,
		IsTrained:    card.DefaultImage == "special_training",
		DefaultImage: card.DefaultImage,
	}
	if store != nil {
		if masterCard := store.GetCard(card.CardID); masterCard != nil {
			payload.CharacterName = characterName(masterCard.CharacterID)
			payload.Rarity = masterCard.CardRarityType
			payload.CardRarityType = masterCard.CardRarityType
			payload.Attr = masterCard.Attr
			payload.AssetbundleName = masterCard.AssetbundleName
			payload.ThumbnailURL = assets.GetCardThumbnailURL(masterCard.AssetbundleName, false)
			payload.TrainedThumbnailURL = assets.GetCardThumbnailURL(masterCard.AssetbundleName, true)
		}
	}
	return payload
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
		DurationSec:         music.SecForMusicScoreMaker,
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

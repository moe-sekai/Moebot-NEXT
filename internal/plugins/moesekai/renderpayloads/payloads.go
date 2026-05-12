package renderpayloads

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/plugins/moesekai/b30"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/ranking"
	"moebot-next/internal/plugins/moesekai/sekai"
	"moebot-next/internal/plugins/moesekai/suite"
)

func defaultAssetResolver() *assets.Resolver {
	return assets.DefaultResolver()
}

// CardDetailPayload is the normalized data contract consumed by CardDetail.tsx.
type CardDetailPayload struct {
	ID               int                  `json:"id"`
	Prefix           string               `json:"prefix"`
	CharacterName    string               `json:"characterName"`
	Rarity           string               `json:"rarity"`
	CardRarityType   string               `json:"cardRarityType"`
	Attr             string               `json:"attr"`
	AssetbundleName  string               `json:"assetbundleName,omitempty"`
	CharacterID      int                  `json:"characterId,omitempty"`
	Power            int                  `json:"power,omitempty"`
	SkillName        string               `json:"skillName,omitempty"`
	Skill            *CardSkillPayload    `json:"skill,omitempty"`
	TrainedSkill     *CardSkillPayload    `json:"trainedSkill,omitempty"`
	Costumes         []CardCostumePayload `json:"costumes,omitempty"`
	GachaPhrase      string               `json:"gachaPhrase,omitempty"`
	SupplyType       string               `json:"supplyType,omitempty"`
	CardSupplyID     int                  `json:"cardSupplyId,omitempty"`
	Events           []CardEventPayload   `json:"events,omitempty"`
	AssetSource      string               `json:"assetSource,omitempty"`
	NormalFullURL    string               `json:"normalFullUrl,omitempty"`
	TrainedFullURL   string               `json:"trainedFullUrl,omitempty"`
	ThumbnailURL     string               `json:"thumbnailUrl,omitempty"`
	TrainedThumbnail string               `json:"trainedThumbnailUrl,omitempty"`
}

// CardSkillPayload is a pre-formatted skill description rendered server-side.
type CardSkillPayload struct {
	ID          int    `json:"id"`
	Level       int    `json:"level"`
	SpriteName  string `json:"spriteName,omitempty"`
	Description string `json:"description"`
}

// CardCostumePayload describes one outfit + its thumbnail asset names so the
// renderer can compose preview tiles.
type CardCostumePayload struct {
	CostumeNumber int      `json:"costumeNumber"`
	Name          string   `json:"name"`
	Rarity        string   `json:"rarity,omitempty"`
	Source        string   `json:"source,omitempty"`
	Designer      string   `json:"designer,omitempty"`
	PartTypes     []string `json:"partTypes,omitempty"`
	ThumbnailURLs []string `json:"thumbnailUrls,omitempty"`
}

type CardEventPayload struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	EventType       string `json:"eventType,omitempty"`
	AssetbundleName string `json:"assetbundleName,omitempty"`
	StartAt         int64  `json:"startAt,omitempty"`
	ClosedAt        int64  `json:"closedAt,omitempty"`
	Unit            string `json:"unit,omitempty"`
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
	SelectedDifficulty  string                   `json:"selectedDifficulty,omitempty"`
	ChartURL            string                   `json:"chartUrl,omitempty"`
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
	ID                int                      `json:"id"`
	Name              string                   `json:"name"`
	EventType         string                   `json:"eventType"`
	Unit              string                   `json:"unit,omitempty"`
	AssetbundleName   string                   `json:"assetbundleName,omitempty"`
	StartAt           int64                    `json:"startAt"`
	AggregateAt       int64                    `json:"aggregateAt,omitempty"`
	ClosedAt          int64                    `json:"closedAt"`
	DistributionEndAt int64                    `json:"distributionEndAt,omitempty"`
	DeckBonuses       []EventDeckBonusPayload  `json:"deckBonuses,omitempty"`
	BonusAttr         string                   `json:"bonusAttr,omitempty"`
	BonusCharacters   []string                 `json:"bonusCharacters,omitempty"`
	BonusCards        []CardDetailPayload      `json:"bonusCards,omitempty"`
	PickupCards       []GachaPickupCardPayload `json:"pickupCards,omitempty"`
	BannerURL         string                   `json:"bannerUrl,omitempty"`
	LogoURL           string                   `json:"logoUrl,omitempty"`
	StoryBannerURL    string                   `json:"storyBannerUrl,omitempty"`
	AssetSource       string                   `json:"assetSource,omitempty"`
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

type CardListPayload struct {
	Title       string              `json:"title"`
	Subtitle    string              `json:"subtitle,omitempty"`
	Cards       []CardDetailPayload `json:"cards"`
	Page        int                 `json:"page,omitempty"`
	TotalPages  int                 `json:"totalPages,omitempty"`
	Total       int                 `json:"total,omitempty"`
	AssetSource string              `json:"assetSource,omitempty"`
}

type SuitePanelPayload struct {
	Title       string                 `json:"title"`
	Subtitle    string                 `json:"subtitle,omitempty"`
	Profile     SuiteProfilePayload    `json:"profile"`
	Stats       []SuiteStatPayload     `json:"stats,omitempty"`
	Sections    []SuiteSectionPayload  `json:"sections,omitempty"`
	DeckCards   []SuiteUserCardPayload `json:"deckCards,omitempty"`
	AssetSource string                 `json:"assetSource,omitempty"`
}

type SuiteProfilePayload struct {
	UserID      string `json:"userId,omitempty"`
	Name        string `json:"name"`
	Rank        int    `json:"rank,omitempty"`
	Region      string `json:"region,omitempty"`
	RegionLabel string `json:"regionLabel,omitempty"`
	Mode        string `json:"mode,omitempty"`
	Source      string `json:"source,omitempty"`
	LocalSource string `json:"localSource,omitempty"`
	UploadTime  int64  `json:"uploadTime,omitempty"`
	UpdateText  string `json:"updateText,omitempty"`
	Coin        int64  `json:"coin,omitempty"`
}

type SuiteStatPayload struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Hint  string `json:"hint,omitempty"`
	Color string `json:"color,omitempty"`
}

type SuiteSectionPayload struct {
	Title    string                   `json:"title"`
	Subtitle string                   `json:"subtitle,omitempty"`
	Kind     string                   `json:"kind,omitempty"`
	Note     string                   `json:"note,omitempty"`
	Columns  []SuiteColumnPayload     `json:"columns,omitempty"`
	Rows     []SuiteSectionRowPayload `json:"rows"`
	Extra    map[string]interface{}   `json:"extra,omitempty"`
}

type SuiteColumnPayload struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type SuiteSectionRowPayload struct {
	ID            int                    `json:"id,omitempty"`
	Rank          int                    `json:"rank,omitempty"`
	Label         string                 `json:"label"`
	Value         string                 `json:"value,omitempty"`
	Meta          string                 `json:"meta,omitempty"`
	Color         string                 `json:"color,omitempty"`
	Card          *SuiteUserCardPayload  `json:"card,omitempty"`
	CharacterID   int                    `json:"characterId,omitempty"`
	MusicID       int                    `json:"musicId,omitempty"`
	EventID       int                    `json:"eventId,omitempty"`
	IconURL       string                 `json:"iconUrl,omitempty"`
	ImageURL      string                 `json:"imageUrl,omitempty"`
	BannerURL     string                 `json:"bannerUrl,omitempty"`
	LogoURL       string                 `json:"logoUrl,omitempty"`
	DateText      string                 `json:"dateText,omitempty"`
	StartAt       int64                  `json:"startAt,omitempty"`
	EndAt         int64                  `json:"endAt,omitempty"`
	Progress      float64                `json:"progress,omitempty"`
	ProgressMax   float64                `json:"progressMax,omitempty"`
	ProgressLabel string                 `json:"progressLabel,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

type SuiteCardBoxPayload struct {
	Title       string                  `json:"title"`
	Subtitle    string                  `json:"subtitle,omitempty"`
	Profile     SuiteProfilePayload     `json:"profile"`
	Groups      []SuiteCardGroupPayload `json:"groups"`
	Options     SuiteCardBoxOptions     `json:"options"`
	Total       int                     `json:"total,omitempty"`
	OwnedTotal  int                     `json:"ownedTotal,omitempty"`
	Page        int                     `json:"page,omitempty"`
	TotalPages  int                     `json:"totalPages,omitempty"`
	PageSize    int                     `json:"pageSize,omitempty"`
	TotalAll    int                     `json:"totalAll,omitempty"`
	AssetSource string                  `json:"assetSource,omitempty"`
}

type SuiteCardGroupPayload struct {
	CharacterID   int                    `json:"characterId,omitempty"`
	CharacterName string                 `json:"characterName"`
	Color         string                 `json:"color,omitempty"`
	Cards         []SuiteUserCardPayload `json:"cards"`
}

type SuiteCardBoxOptions struct {
	ShowID            bool   `json:"showId,omitempty"`
	OwnedOnly         bool   `json:"ownedOnly,omitempty"`
	UseBeforeTraining bool   `json:"useBeforeTraining,omitempty"`
	ShowCreatedAt     bool   `json:"showCreatedAt,omitempty"`
	SortBy            string `json:"sortBy,omitempty"`
}

type SuiteUserCardPayload struct {
	CardID                int    `json:"cardId,omitempty"`
	ID                    int    `json:"id,omitempty"`
	Prefix                string `json:"prefix,omitempty"`
	CharacterID           int    `json:"characterId,omitempty"`
	CharacterName         string `json:"characterName,omitempty"`
	Rarity                string `json:"rarity,omitempty"`
	CardRarityType        string `json:"cardRarityType,omitempty"`
	Attr                  string `json:"attr,omitempty"`
	AssetbundleName       string `json:"assetbundleName,omitempty"`
	ThumbnailURL          string `json:"thumbnailUrl,omitempty"`
	TrainedThumbnailURL   string `json:"trainedThumbnailUrl,omitempty"`
	SupplyType            string `json:"supplyType,omitempty"`
	SupplyDisplayName     string `json:"supplyDisplayName,omitempty"`
	Level                 int    `json:"level,omitempty"`
	MasterRank            int    `json:"masterRank,omitempty"`
	Mastery               int    `json:"mastery,omitempty"`
	SkillLevel            int    `json:"skillLevel,omitempty"`
	DefaultImage          string `json:"defaultImage,omitempty"`
	SpecialTrainingStatus string `json:"specialTrainingStatus,omitempty"`
	CreatedAt             int64  `json:"createdAt,omitempty"`
	CreatedAtText         string `json:"createdAtText,omitempty"`
	Owned                 bool   `json:"owned"`
	InDeck                bool   `json:"inDeck,omitempty"`
	IsTrained             bool   `json:"isTrained,omitempty"`
}

type MusicListPayload struct {
	Title       string               `json:"title"`
	Subtitle    string               `json:"subtitle,omitempty"`
	Musics      []MusicDetailPayload `json:"musics"`
	Page        int                  `json:"page,omitempty"`
	TotalPages  int                  `json:"totalPages,omitempty"`
	Total       int                  `json:"total,omitempty"`
	AssetSource string               `json:"assetSource,omitempty"`
}

type Best30Payload struct {
	Title                 string               `json:"title"`
	Subtitle              string               `json:"subtitle,omitempty"`
	Profile               SuiteProfilePayload  `json:"profile"`
	Average               float64              `json:"average"`
	Entries               []Best30EntryPayload `json:"entries"`
	CandidateCount        int                  `json:"candidateCount"`
	APCount               int                  `json:"apCount"`
	FCCount               int                  `json:"fcCount"`
	MissingConstantsCount int                  `json:"missingConstantsCount,omitempty"`
	TotalResultCount      int                  `json:"totalResultCount,omitempty"`
	Region                string               `json:"region,omitempty"`
	RegionLabel           string               `json:"regionLabel,omitempty"`
	UpdatedAt             int64                `json:"updatedAt,omitempty"`
	UpdateText            string               `json:"updateText,omitempty"`
	Formula               string               `json:"formula,omitempty"`
	ConstantsSource       string               `json:"constantsSource,omitempty"`
	AssetSource           string               `json:"assetSource,omitempty"`
}

type Best30EntryPayload struct {
	Rank            int     `json:"rank"`
	MusicID         int     `json:"musicId"`
	Title           string  `json:"title"`
	Difficulty      string  `json:"difficulty"`
	DifficultyLabel string  `json:"difficultyLabel"`
	Level           int     `json:"level,omitempty"`
	Constant        float64 `json:"constant"`
	UserRating      float64 `json:"userRating"`
	PlayResult      string  `json:"playResult"`
	NoteCount       int     `json:"noteCount,omitempty"`
	AssetbundleName string  `json:"assetbundleName,omitempty"`
	JacketURL       string  `json:"jacketUrl,omitempty"`
}

type EventListPayload struct {
	Title       string             `json:"title"`
	Subtitle    string             `json:"subtitle,omitempty"`
	Events      []EventInfoPayload `json:"events"`
	Page        int                `json:"page,omitempty"`
	TotalPages  int                `json:"totalPages,omitempty"`
	Total       int                `json:"total,omitempty"`
	AssetSource string             `json:"assetSource,omitempty"`
}

type GachaListPayload struct {
	Title       string             `json:"title"`
	Subtitle    string             `json:"subtitle,omitempty"`
	Gachas      []GachaInfoPayload `json:"gachas"`
	Page        int                `json:"page,omitempty"`
	TotalPages  int                `json:"totalPages,omitempty"`
	Total       int                `json:"total,omitempty"`
	AssetSource string             `json:"assetSource,omitempty"`
}

type VirtualLiveListPayload struct {
	Title        string               `json:"title"`
	Subtitle     string               `json:"subtitle,omitempty"`
	VirtualLives []VirtualLivePayload `json:"virtualLives"`
	Page         int                  `json:"page,omitempty"`
	TotalPages   int                  `json:"totalPages,omitempty"`
	Total        int                  `json:"total,omitempty"`
	AssetSource  string               `json:"assetSource,omitempty"`
}

type VirtualLivePayload struct {
	ID              int                           `json:"id"`
	Name            string                        `json:"name"`
	AssetbundleName string                        `json:"assetbundleName,omitempty"`
	VirtualLiveType string                        `json:"virtualLiveType,omitempty"`
	StartAt         int64                         `json:"startAt"`
	EndAt           int64                         `json:"endAt"`
	CurrentStartAt  int64                         `json:"currentStartAt,omitempty"`
	CurrentEndAt    int64                         `json:"currentEndAt,omitempty"`
	Living          bool                          `json:"living,omitempty"`
	RestCount       int                           `json:"restCount,omitempty"`
	Schedules       []VirtualLiveSchedulePayload  `json:"schedules,omitempty"`
	Rewards         []VirtualLiveRewardPayload    `json:"rewards,omitempty"`
	Characters      []VirtualLiveCharacterPayload `json:"characters,omitempty"`
	AssetSource     string                        `json:"assetSource,omitempty"`
}

type VirtualLiveSchedulePayload struct {
	ID      int   `json:"id,omitempty"`
	Seq     int   `json:"seq,omitempty"`
	StartAt int64 `json:"startAt"`
	EndAt   int64 `json:"endAt"`
}

type VirtualLiveRewardPayload struct {
	ID              int    `json:"id,omitempty"`
	VirtualLiveType string `json:"virtualLiveType,omitempty"`
	ResourceBoxID   int    `json:"resourceBoxId,omitempty"`
}

type VirtualLiveCharacterPayload struct {
	ID                  int    `json:"id,omitempty"`
	GameCharacterID     int    `json:"gameCharacterId,omitempty"`
	GameCharacterUnitID int    `json:"gameCharacterUnitId,omitempty"`
	CharacterName       string `json:"characterName,omitempty"`
	PerformanceType     string `json:"performanceType,omitempty"`
}

type RankingListPayload struct {
	Title       string                `json:"title"`
	Subtitle    string                `json:"subtitle,omitempty"`
	Rankings    []RankingEntryPayload `json:"rankings"`
	EventID     int                   `json:"eventId,omitempty"`
	UpdatedAt   int64                 `json:"updatedAt,omitempty"`
	AssetSource string                `json:"assetSource,omitempty"`
	Region      string                `json:"region,omitempty"`
	RegionLabel string                `json:"regionLabel,omitempty"`
	BoardType   string                `json:"boardType,omitempty"`
	TargetID    int                   `json:"targetId,omitempty"`
}

type RankingEntryPayload struct {
	Rank                int                     `json:"rank"`
	Name                string                  `json:"name,omitempty"`
	DisplayName         string                  `json:"displayName,omitempty"`
	Signature           string                  `json:"signature,omitempty"`
	Score               int64                   `json:"score"`
	UserID              string                  `json:"userId,omitempty"`
	ScoreDelta          int64                   `json:"scoreDelta,omitempty"`
	AvatarURL           string                  `json:"avatarUrl,omitempty"`
	LeaderCharacterID   int                     `json:"leaderCharacterId,omitempty"`
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
	return BuildRankingListPayloadWithAssets(title, board, defaultAssetResolver())
}

func BuildRankingListPayloadWithAssets(title string, board ranking.Board, resolver *assets.Resolver) RankingListPayload {
	payload := RankingListPayload{Title: title, EventID: board.EventID, UpdatedAt: board.UpdatedAt, AssetSource: assetSourceForResolver(resolver), Region: board.Region, BoardType: board.BoardType, TargetID: board.TargetID}
	for _, entry := range board.Rankings {
		payload.Rankings = append(payload.Rankings, buildRankingEntryPayload(entry))
	}
	return payload
}

func BuildRankingListPayloadWithStore(title string, board ranking.Board, store *masterdata.Store, resolver *assets.Resolver) RankingListPayload {
	payload := RankingListPayload{Title: title, EventID: board.EventID, UpdatedAt: board.UpdatedAt, AssetSource: assetSourceForResolver(resolver), Region: board.Region, BoardType: board.BoardType, TargetID: board.TargetID}
	for _, entry := range board.Rankings {
		payload.Rankings = append(payload.Rankings, buildRankingEntryPayloadWithStore(entry, store, resolver))
	}
	return payload
}

func BuildChurnRankingListPayload(board ranking.Board) RankingListPayload {
	return BuildChurnRankingListPayloadWithAssets(board, defaultAssetResolver())
}

func BuildChurnRankingListPayloadWithAssets(board ranking.Board, resolver *assets.Resolver) RankingListPayload {
	return BuildChurnRankingListPayloadWithStore(board, nil, resolver)
}

func BuildChurnRankingListPayloadWithStore(board ranking.Board, store *masterdata.Store, resolver *assets.Resolver) RankingListPayload {
	payload := RankingListPayload{Title: "查房", Subtitle: "活跃度 / 时速 / 最近分数变化", EventID: board.EventID, UpdatedAt: board.UpdatedAt, AssetSource: assetSourceForResolver(resolver), Region: board.Region, BoardType: board.BoardType, TargetID: board.TargetID}
	for _, entry := range board.Rankings {
		item := buildRankingEntryPayloadWithStore(entry, store, resolver)
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

func BuildWaterTablePayloadWithStore(board ranking.Board, entry ranking.RankingEntry, store *masterdata.Store, resolver *assets.Resolver) WaterTablePayload {
	payload := WaterTablePayload{Title: "查水表", Subtitle: "小时周回 / 停车时间", EventID: board.EventID, UpdatedAt: board.UpdatedAt, Region: board.Region, BoardType: board.BoardType, TargetID: board.TargetID, AssetSource: assetSourceForResolver(resolver)}
	payload.Entry = buildRankingEntryPayloadWithStore(entry, store, resolver)
	payload.Entry.Churn48h = entry.Churn48h
	payload.Entry.Growth1h = entry.Growth1h
	if entry.LastChange != nil {
		payload.Entry.ScoreDelta = entry.LastChange.Delta
	}
	payload.HourlyChurn = append([]ranking.HourlyChurn(nil), entry.HourlyChurn...)
	payload.Parking = append([]ranking.ParkingPeriod(nil), entry.ParkingPeriods...)
	return payload
}

func BuildForecastRankingPayload(board ranking.ForecastBoard, eventName string, region string, regionLabel string) ForecastRankingPayload {
	payload := ForecastRankingPayload{Title: "榜线预测", EventID: board.EventID, EventName: eventName, Region: region, RegionLabel: regionLabel, Status: board.Status, UpdatedAt: board.UpdatedUnixMilli()}
	if eventName != "" {
		payload.Subtitle = eventName
	}
	for _, item := range board.Items {
		prediction, ok := item.PredictedScore()
		payload.Items = append(payload.Items, ForecastRankingEntryPayload{Rank: item.Rank, Score: item.Score, Prediction: prediction, HasPrediction: ok, CollectTime: item.CollectUnixMilli(), IsFinal: item.IsFinal})
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
	return buildRankingEntryPayloadWithStore(entry, nil, nil)
}

func resolveRankingLeaderCard(store *masterdata.Store, cardID int, characterID int) *masterdata.CardInfo {
	if store == nil {
		return nil
	}
	if cardID > 0 {
		if card := store.GetCard(cardID); card != nil && card.AssetbundleName != "" {
			return card
		}
	}
	if characterID <= 0 {
		return nil
	}
	var best *masterdata.CardInfo
	for _, card := range store.AllCards() {
		if card.CharacterID != characterID || card.AssetbundleName == "" {
			continue
		}
		if best == nil || card.ReleaseAt > best.ReleaseAt || (card.ReleaseAt == best.ReleaseAt && card.ID > best.ID) {
			copyCard := card
			best = &copyCard
		}
	}
	return best
}

func buildRankingEntryPayloadWithStore(entry ranking.RankingEntry, store *masterdata.Store, resolver *assets.Resolver) RankingEntryPayload {
	payload := RankingEntryPayload{
		Rank:      entry.Rank,
		Name:      entry.Name,
		Signature: entry.Word,
		Score:     entry.Score,
		UserID:    entry.UserID.String(),
	}
	if entry.LeaderCard != nil {
		payload.LeaderCharacterID = entry.LeaderCard.CharacterID
		payload.LeaderCard = &ProfileDeckCardPayload{
			CardID:       entry.LeaderCard.CardID,
			ID:           entry.LeaderCard.CardID,
			Level:        entry.LeaderCard.Level,
			Mastery:      entry.LeaderCard.MasterRank,
			IsTrained:    entry.LeaderCard.DefaultImage == "special_training",
			DefaultImage: entry.LeaderCard.DefaultImage,
		}
		if entry.LeaderCard.CharacterID > 0 {
			payload.LeaderCard.CharacterName = characterName(entry.LeaderCard.CharacterID)
		}
		if store != nil {
			if card := resolveRankingLeaderCard(store, entry.LeaderCard.CardID, entry.LeaderCard.CharacterID); card != nil {
				payload.LeaderCard.CardID = card.ID
				payload.LeaderCard.ID = card.ID
				payload.LeaderCard.CharacterName = characterName(card.CharacterID)
				payload.LeaderCharacterID = card.CharacterID
				payload.LeaderCard.Rarity = card.CardRarityType
				payload.LeaderCard.CardRarityType = card.CardRarityType
				payload.LeaderCard.Attr = card.Attr
				payload.LeaderCard.AssetbundleName = card.AssetbundleName
				assetResolver := resolverOrDefault(resolver)
				payload.LeaderCard.ThumbnailURL = assetResolver.GetCardThumbnailURL(card.AssetbundleName, false)
				payload.LeaderCard.TrainedThumbnailURL = assetResolver.GetCardThumbnailURL(card.AssetbundleName, true)
				payload.AvatarURL = payload.LeaderCard.ThumbnailURL
			}
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
	Seq                           int    `json:"seq"`
	HonorType                     string `json:"honorType,omitempty"`
	HonorID                       int    `json:"honorId"`
	Level                         int    `json:"level"`
	Name                          string `json:"name,omitempty"`
	HonorRarity                   string `json:"honorRarity,omitempty"`
	AssetbundleName               string `json:"assetbundleName,omitempty"`
	ImageURL                      string `json:"imageUrl,omitempty"`
	FrameURL                      string `json:"frameUrl,omitempty"`
	LevelIconURL                  string `json:"levelIconUrl,omitempty"`
	LevelIcon6URL                 string `json:"levelIcon6Url,omitempty"`
	BondsHonorViewType            string `json:"bondsHonorViewType,omitempty"`
	BondsHonorWordID              int    `json:"bondsHonorWordId,omitempty"`
	BondsHonorWordAssetbundleName string `json:"bondsHonorWordAssetbundleName,omitempty"`
	BondsHonorWordURL             string `json:"bondsHonorWordUrl,omitempty"`
	LeftCharacterID               int    `json:"leftCharacterId,omitempty"`
	RightCharacterID              int    `json:"rightCharacterId,omitempty"`
	LeftCharacterURL              string `json:"leftCharacterUrl,omitempty"`
	RightCharacterURL             string `json:"rightCharacterUrl,omitempty"`
	LeftColor                     string `json:"leftColor,omitempty"`
	RightColor                    string `json:"rightColor,omitempty"`
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
	SupplyType          string `json:"supplyType,omitempty"`
	Level               int    `json:"level,omitempty"`
	Mastery             int    `json:"mastery,omitempty"`
	IsTrained           bool   `json:"isTrained,omitempty"`
	DefaultImage        string `json:"defaultImage,omitempty"`
}

type WaterTablePayload struct {
	Title       string                  `json:"title"`
	Subtitle    string                  `json:"subtitle,omitempty"`
	Entry       RankingEntryPayload     `json:"entry"`
	HourlyChurn []ranking.HourlyChurn   `json:"hourlyChurn,omitempty"`
	Parking     []ranking.ParkingPeriod `json:"parkingPeriods,omitempty"`
	EventID     int                     `json:"eventId,omitempty"`
	UpdatedAt   int64                   `json:"updatedAt,omitempty"`
	Region      string                  `json:"region,omitempty"`
	RegionLabel string                  `json:"regionLabel,omitempty"`
	BoardType   string                  `json:"boardType,omitempty"`
	TargetID    int                     `json:"targetId,omitempty"`
	AssetSource string                  `json:"assetSource,omitempty"`
}

type ForecastRankingPayload struct {
	Title       string                        `json:"title"`
	Subtitle    string                        `json:"subtitle,omitempty"`
	EventID     int                           `json:"eventId,omitempty"`
	EventName   string                        `json:"eventName,omitempty"`
	Region      string                        `json:"region,omitempty"`
	RegionLabel string                        `json:"regionLabel,omitempty"`
	Status      string                        `json:"status,omitempty"`
	UpdatedAt   int64                         `json:"updatedAt,omitempty"`
	Items       []ForecastRankingEntryPayload `json:"items"`
}

type ForecastRankingEntryPayload struct {
	Rank          int   `json:"rank"`
	Score         int64 `json:"score"`
	Prediction    int64 `json:"prediction,omitempty"`
	HasPrediction bool  `json:"hasPrediction,omitempty"`
	CollectTime   int64 `json:"collectTime,omitempty"`
	IsFinal       bool  `json:"isFinal,omitempty"`
}

type SuiteCommonProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	UserDecks    []suite.UserDeck   `json:"userDecks"`
	UserCards    []suite.UserCard   `json:"userCards"`
}

func NewSuitePanelPayload(title string, subtitle string, region string, mode string, profile SuiteCommonProfile, resolver *assets.Resolver) SuitePanelPayload {
	assetResolver := resolverOrDefault(resolver)
	name := strings.TrimSpace(profile.UserGamedata.Name)
	if name == "" {
		name = "未知玩家"
	}
	return SuitePanelPayload{
		Title:       title,
		Subtitle:    subtitle,
		Profile:     BuildSuiteProfilePayload(region, mode, profile.BaseProfile, profile.UserGamedata),
		DeckCards:   BuildSuiteDeckCards(profile.UserDecks, profile.UserCards, profile.UserGamedata.Deck, nil, assetResolver),
		AssetSource: assetSourceForResolver(assetResolver),
	}
}

func BuildSuiteProfilePayload(region string, mode string, base suite.BaseProfile, game suite.UserGamedata) SuiteProfilePayload {
	name := strings.TrimSpace(game.Name)
	if name == "" {
		name = "未知玩家"
	}
	source := strings.TrimSpace(base.Source)
	if source == "" {
		source = suite.PublicSource
	}
	return SuiteProfilePayload{
		UserID:      game.UserID.String(),
		Name:        name,
		Rank:        game.Rank,
		Region:      config.NormalizeRegion(region),
		RegionLabel: config.RegionLabel(region),
		Mode:        strings.TrimSpace(mode),
		Source:      source,
		LocalSource: base.LocalSource,
		UploadTime:  normalizeSuiteTimestamp(base.UploadTime),
		UpdateText:  formatSuiteTimestamp(base.UploadTime),
		Coin:        game.Coin,
	}
}

func Best30MusicMetaResolver(store *masterdata.Store, resolver *assets.Resolver) b30.MetaResolver {
	assetResolver := resolverOrDefault(resolver)
	return func(musicID int, difficulty string, constant b30.ChartConstant) b30.MusicMeta {
		meta := b30.MusicMeta{MusicID: musicID, Level: constant.Level, NoteCount: constant.NoteCount}
		if store != nil {
			if music := store.GetMusic(musicID); music != nil {
				meta.Title = firstNonEmptyString(music.Title, meta.Title)
				meta.AssetbundleName = music.AssetbundleName
				meta.PublishedAt = music.PublishedAt
				if music.AssetbundleName != "" {
					meta.JacketURL = assetResolver.GetMusicJacketURL(music.AssetbundleName)
				}
			}
			for _, diff := range store.GetMusicDifficulties(musicID) {
				if strings.EqualFold(diff.MusicDifficulty, difficulty) {
					if diff.PlayLevel > 0 {
						meta.Level = diff.PlayLevel
					}
					if diff.TotalNoteCount > 0 {
						meta.NoteCount = diff.TotalNoteCount
					}
					break
				}
			}
		}
		return meta
	}
}

func BuildBest30Payload(title string, region string, base suite.BaseProfile, game suite.UserGamedata, result b30.Result, store *masterdata.Store, resolver *assets.Resolver, constantsSource string) Best30Payload {
	assetResolver := resolverOrDefault(resolver)
	region = config.NormalizeRegion(region)
	if region == "" {
		region = config.RegionJP
	}
	if strings.TrimSpace(title) == "" {
		title = fmt.Sprintf("%s Best30", strings.ToUpper(region))
	}
	updatedAt := normalizeSuiteTimestamp(base.UploadTime)
	payload := Best30Payload{
		Title:                 title,
		Subtitle:              fmt.Sprintf("%s · 社区定数 · 仅供参考", config.RegionLabel(region)),
		Profile:               BuildSuiteProfilePayload(region, "", base, game),
		Average:               result.Average,
		CandidateCount:        result.CandidateCount,
		APCount:               result.APCount,
		FCCount:               result.FCCount,
		MissingConstantsCount: result.MissingConstantsCount,
		TotalResultCount:      result.TotalResultCount,
		Region:                region,
		RegionLabel:           config.RegionLabel(region),
		UpdatedAt:             updatedAt,
		UpdateText:            formatSuiteTimestamp(updatedAt),
		Formula:               "AP=定数；FC=定数-1(≥33) / 定数-1.5(<33)",
		ConstantsSource:       strings.TrimSpace(constantsSource),
		AssetSource:           assetSourceForResolver(assetResolver),
	}
	for _, entry := range result.Entries {
		payload.Entries = append(payload.Entries, best30EntryPayload(entry, store, assetResolver))
	}
	return payload
}

func best30EntryPayload(entry b30.Entry, store *masterdata.Store, resolver *assets.Resolver) Best30EntryPayload {
	assetResolver := resolverOrDefault(resolver)
	payload := Best30EntryPayload{
		Rank:            entry.Rank,
		MusicID:         entry.MusicID,
		Title:           firstNonEmptyString(entry.Title, fmt.Sprintf("歌曲 #%d", entry.MusicID)),
		Difficulty:      b30.NormalizeDifficulty(entry.Difficulty),
		DifficultyLabel: best30DifficultyLabel(entry.Difficulty),
		Level:           entry.Level,
		Constant:        entry.Constant,
		UserRating:      entry.UserRating,
		PlayResult:      string(entry.PlayResult),
		NoteCount:       entry.NoteCount,
		AssetbundleName: entry.AssetbundleName,
		JacketURL:       entry.JacketURL,
	}
	if store != nil {
		if music := store.GetMusic(entry.MusicID); music != nil {
			payload.Title = firstNonEmptyString(music.Title, payload.Title)
			if payload.AssetbundleName == "" {
				payload.AssetbundleName = music.AssetbundleName
			}
		}
		for _, diff := range store.GetMusicDifficulties(entry.MusicID) {
			if strings.EqualFold(diff.MusicDifficulty, payload.Difficulty) {
				if payload.Level <= 0 {
					payload.Level = diff.PlayLevel
				}
				if payload.NoteCount <= 0 {
					payload.NoteCount = diff.TotalNoteCount
				}
				break
			}
		}
	}
	if payload.JacketURL == "" && payload.AssetbundleName != "" {
		payload.JacketURL = assetResolver.GetMusicJacketURL(payload.AssetbundleName)
	}
	return payload
}

func best30DifficultyLabel(diff string) string {
	switch b30.NormalizeDifficulty(diff) {
	case "easy":
		return "EAS"
	case "normal":
		return "NOR"
	case "hard":
		return "HRD"
	case "expert":
		return "EXP"
	case "master":
		return "MAS"
	case "append":
		return "APD"
	default:
		return strings.ToUpper(diff)
	}
}

func BuildSuiteDeckCards(decks []suite.UserDeck, userCards []suite.UserCard, deckID int, store *masterdata.Store, resolver *assets.Resolver) []SuiteUserCardPayload {
	deck := selectSuiteDeck(decks, deckID)
	if deck == nil {
		return nil
	}
	memberIDs := suiteDeckMemberIDs(deck)
	owned := SuiteUserCardMap(userCards)
	out := make([]SuiteUserCardPayload, 0, len(memberIDs))
	for _, cardID := range memberIDs {
		if cardID <= 0 {
			continue
		}
		card, ok := owned[cardID]
		if !ok {
			card = suite.UserCard{CardID: cardID}
		}
		payload := BuildSuiteUserCardPayload(store, card, resolver, ok)
		payload.InDeck = true
		out = append(out, payload)
	}
	return out
}

func suiteDeckMemberIDs(deck *suite.UserDeck) []int {
	if deck == nil {
		return nil
	}
	candidates := []int{firstNonZero(deck.Member1, deck.Leader), deck.Member2, deck.Member3, deck.Member4, deck.Member5, deck.Leader, deck.SubLeader}
	ids := make([]int, 0, 5)
	seen := make(map[int]struct{}, len(candidates))
	for _, id := range candidates {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
		if len(ids) >= 5 {
			break
		}
	}
	return ids
}

func selectSuiteDeck(decks []suite.UserDeck, deckID int) *suite.UserDeck {
	if len(decks) == 0 {
		return nil
	}
	if deckID > 0 {
		for i := range decks {
			if decks[i].DeckID == deckID {
				return &decks[i]
			}
		}
	}
	return &decks[0]
}

func SuiteUserCardMap(cards []suite.UserCard) map[int]suite.UserCard {
	out := make(map[int]suite.UserCard, len(cards))
	for _, card := range cards {
		if card.CardID <= 0 {
			continue
		}
		if existing, ok := out[card.CardID]; !ok || card.CreatedAt > existing.CreatedAt || card.Level > existing.Level {
			out[card.CardID] = card
		}
	}
	return out
}

func BuildSuiteUserCardPayload(store *masterdata.Store, card suite.UserCard, resolver *assets.Resolver, owned bool) SuiteUserCardPayload {
	assetResolver := resolverOrDefault(resolver)
	payload := SuiteUserCardPayload{
		CardID:                card.CardID,
		ID:                    card.CardID,
		Level:                 card.Level,
		MasterRank:            card.MasterRank,
		Mastery:               card.MasterRank,
		SkillLevel:            card.SkillLevel,
		DefaultImage:          card.DefaultImage,
		SpecialTrainingStatus: card.SpecialTrainingStatus,
		CreatedAt:             normalizeSuiteTimestamp(card.CreatedAt),
		CreatedAtText:         formatSuiteDate(card.CreatedAt),
		Owned:                 owned,
		IsTrained:             suiteCardUsesTrainedArt(card),
	}
	if store != nil {
		if masterCard := store.GetCard(card.CardID); masterCard != nil {
			EnrichSuiteUserCardPayload(&payload, store, *masterCard, assetResolver)
		}
	}
	return payload
}

func EnrichSuiteUserCardPayload(payload *SuiteUserCardPayload, store *masterdata.Store, card masterdata.CardInfo, resolver *assets.Resolver) {
	if payload == nil {
		return
	}
	assetResolver := resolverOrDefault(resolver)
	payload.CardID = firstNonZero(payload.CardID, card.ID)
	payload.ID = firstNonZero(payload.ID, card.ID)
	payload.Prefix = card.Prefix
	payload.CharacterID = card.CharacterID
	payload.CharacterName = characterName(card.CharacterID)
	payload.Rarity = card.CardRarityType
	payload.CardRarityType = card.CardRarityType
	payload.Attr = card.Attr
	payload.AssetbundleName = card.AssetbundleName
	payload.SupplyType = CardSupplyDisplayName(store, card)
	payload.SupplyDisplayName = payload.SupplyType
	if shouldHideNormalSupply(payload.SupplyType) {
		payload.SupplyType = ""
	}
	if card.AssetbundleName != "" {
		payload.ThumbnailURL = assetResolver.GetCardThumbnailURL(card.AssetbundleName, false)
		if cardCanUseTrainedThumbnail(card.CardRarityType) {
			payload.TrainedThumbnailURL = assetResolver.GetCardThumbnailURL(card.AssetbundleName, true)
		}
	}
}

func BuildSuiteCardBoxPayload(title string, subtitle string, region string, mode string, base suite.BaseProfile, game suite.UserGamedata, cards []masterdata.CardInfo, ownedCards map[int]suite.UserCard, deckCards map[int]struct{}, store *masterdata.Store, resolver *assets.Resolver, options SuiteCardBoxOptions) SuiteCardBoxPayload {
	assetResolver := resolverOrDefault(resolver)
	groupsByCharacter := map[int]*SuiteCardGroupPayload{}
	characterOrder := make([]int, 0)
	ownedTotal := len(ownedCards)
	shownOwned := 0
	for _, masterCard := range cards {
		userCard, owned := ownedCards[masterCard.ID]
		if options.OwnedOnly && !owned {
			continue
		}
		if owned {
			shownOwned++
		} else {
			userCard = suite.UserCard{CardID: masterCard.ID}
		}
		payload := BuildSuiteUserCardPayload(store, userCard, assetResolver, owned)
		EnrichSuiteUserCardPayload(&payload, store, masterCard, assetResolver)
		if _, ok := deckCards[masterCard.ID]; ok {
			payload.InDeck = true
		}
		group := groupsByCharacter[masterCard.CharacterID]
		if group == nil {
			group = &SuiteCardGroupPayload{CharacterID: masterCard.CharacterID, CharacterName: characterName(masterCard.CharacterID), Color: characterColor(masterCard.CharacterID)}
			groupsByCharacter[masterCard.CharacterID] = group
			characterOrder = append(characterOrder, masterCard.CharacterID)
		}
		group.Cards = append(group.Cards, payload)
	}
	sort.Ints(characterOrder)
	groups := make([]SuiteCardGroupPayload, 0, len(characterOrder))
	for _, characterID := range characterOrder {
		group := groupsByCharacter[characterID]
		if group == nil {
			continue
		}
		sort.SliceStable(group.Cards, func(i, j int) bool {
			return suiteCardLess(group.Cards[i], group.Cards[j], options.SortBy)
		})
		groups = append(groups, *group)
	}
	if options.OwnedOnly {
		ownedTotal = shownOwned
	}
	return SuiteCardBoxPayload{
		Title:       title,
		Subtitle:    subtitle,
		Profile:     BuildSuiteProfilePayload(region, mode, base, game),
		Groups:      groups,
		Options:     options,
		Total:       len(cards),
		OwnedTotal:  ownedTotal,
		AssetSource: assetSourceForResolver(assetResolver),
	}
}

func suiteCardLess(a SuiteUserCardPayload, b SuiteUserCardPayload, sortBy string) bool {
	switch sortBy {
	case "mr":
		if a.MasterRank != b.MasterRank {
			return a.MasterRank > b.MasterRank
		}
		if rarityWeight(a.CardRarityType) != rarityWeight(b.CardRarityType) {
			return rarityWeight(a.CardRarityType) > rarityWeight(b.CardRarityType)
		}
		if a.CreatedAt != b.CreatedAt {
			return a.CreatedAt > b.CreatedAt
		}
	case "sl":
		if a.SkillLevel != b.SkillLevel {
			return a.SkillLevel > b.SkillLevel
		}
		if rarityWeight(a.CardRarityType) != rarityWeight(b.CardRarityType) {
			return rarityWeight(a.CardRarityType) > rarityWeight(b.CardRarityType)
		}
		if a.CreatedAt != b.CreatedAt {
			return a.CreatedAt > b.CreatedAt
		}
	case "time":
		if a.CreatedAt != b.CreatedAt {
			return a.CreatedAt > b.CreatedAt
		}
		if rarityWeight(a.CardRarityType) != rarityWeight(b.CardRarityType) {
			return rarityWeight(a.CardRarityType) > rarityWeight(b.CardRarityType)
		}
	default:
		if a.Owned != b.Owned {
			return a.Owned && !b.Owned
		}
	}
	return a.ID < b.ID
}

func suiteCardUsesTrainedArt(card suite.UserCard) bool {
	return card.DefaultImage == "special_training" || card.SpecialTrainingStatus == "done"
}

func normalizeSuiteTimestamp(value int64) int64 {
	if value > 0 && value < 100000000000 {
		return value * 1000
	}
	return value
}

func formatSuiteTimestamp(value int64) string {
	value = normalizeSuiteTimestamp(value)
	if value <= 0 {
		return "未知"
	}
	return time.UnixMilli(value).Format("2006-01-02 15:04:05")
}

func formatSuiteDate(value int64) string {
	value = normalizeSuiteTimestamp(value)
	if value <= 0 {
		return ""
	}
	return time.UnixMilli(value).Format("2006-01-02")
}

func cardCanUseTrainedThumbnail(rarity string) bool {
	return rarity == "rarity_3" || rarity == "rarity_4"
}

func shouldHideNormalSupply(label string) bool {
	return label == "" || label == "常驻" || strings.EqualFold(label, "normal")
}

func rarityWeight(rarity string) int {
	switch rarity {
	case "rarity_birthday":
		return 5
	case "rarity_4":
		return 4
	case "rarity_3":
		return 3
	case "rarity_2":
		return 2
	case "rarity_1":
		return 1
	default:
		return 0
	}
}

func characterColor(characterID int) string {
	colors := []string{
		"#33ccbb", "#4455dd", "#66ccff", "#ee6666", "#bb88ee", "#ffdd44",
		"#88dd44", "#ffccaa", "#99eedd", "#ff6699", "#ffbbcc",
		"#ee1166", "#00bbdd", "#ff7722", "#0077dd", "#884499",
		"#ff9900", "#ffbb00", "#ff66bb", "#33dd99", "#bb88ff",
		"#884499", "#8899cc", "#ccaa88", "#bb6688", "#6699cc",
	}
	if characterID <= 0 {
		return "#33ccbb"
	}
	return colors[(characterID-1)%len(colors)]
}

// BuildProfileCardPayload adapts an API profile into ProfileCard renderer props.
func BuildProfileCardPayload(profile sekai.Profile) ProfileCardPayload {
	return BuildProfileCardPayloadWithStore(nil, profile)
}

// BuildProfileCardPayloadWithStore enriches profile card IDs with masterdata card display fields.
func BuildProfileCardPayloadWithStore(store *masterdata.Store, profile sekai.Profile) ProfileCardPayload {
	return BuildProfileCardPayloadWithAssets(store, profile, defaultAssetResolver())
}

// BuildProfileCardPayloadWithAssets enriches profile card IDs using a region-specific asset resolver.
func BuildProfileCardPayloadWithAssets(store *masterdata.Store, profile sekai.Profile, resolver *assets.Resolver) ProfileCardPayload {
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
		AssetSource: assetSourceForResolver(resolver),
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
			Seq:                honor.Seq,
			HonorType:          honor.HonorType,
			HonorID:            honor.HonorID,
			Level:              honor.Level,
			BondsHonorViewType: honor.BondsHonorViewType,
			BondsHonorWordID:   honor.BondsHonorWordID,
		}
		if store != nil {
			if honor.HonorType == "bonds" {
				if bondsHonor := store.GetBondsHonor(honor.HonorID); bondsHonor != nil {
					applyProfileBondsHonorMasterData(&honorPayload, bondsHonor, store, resolver)
				}
			} else if masterHonor := store.GetHonor(honor.HonorID); masterHonor != nil {
				applyProfileHonorMasterData(&honorPayload, masterHonor, resolver)
			}
		}
		payload.ProfileHonors = append(payload.ProfileHonors, honorPayload)
	}
	if profile.LeaderCard != nil {
		leader := buildProfileDeckCardPayload(store, *profile.LeaderCard, resolver)
		payload.LeaderCard = &leader
	}
	for _, card := range profile.DeckCards {
		payload.DeckCards = append(payload.DeckCards, buildProfileDeckCardPayload(store, card, resolver))
	}
	return payload
}

func applyProfileHonorMasterData(payload *ProfileHonorPayload, honor *masterdata.HonorInfo, resolver *assets.Resolver) {
	if payload == nil || honor == nil {
		return
	}
	assetResolver := resolverOrDefault(resolver)
	payload.Name = honor.Name
	payload.HonorRarity = honor.HonorRarity
	payload.AssetbundleName = honor.AssetbundleName
	if level := findHonorLevel(honor.Levels, payload.Level); level != nil {
		if level.HonorRarity != "" {
			payload.HonorRarity = level.HonorRarity
		}
		if level.AssetbundleName != "" {
			payload.AssetbundleName = level.AssetbundleName
		}
	}
	if payload.AssetbundleName != "" {
		payload.ImageURL = assetResolver.GetHonorBgURL(payload.AssetbundleName, "main")
	}
	if payload.HonorRarity != "" {
		payload.FrameURL = assetResolver.GetHonorFrameURL(payload.HonorRarity, "main")
	}
	payload.LevelIconURL = assetResolver.GetHonorLevelIconURL(false)
	payload.LevelIcon6URL = assetResolver.GetHonorLevelIconURL(true)
}

func findHonorLevel(levels []masterdata.HonorLevel, level int) *masterdata.HonorLevel {
	if level <= 0 {
		return nil
	}
	for i := range levels {
		if levels[i].Level == level {
			return &levels[i]
		}
	}
	return nil
}

func applyProfileBondsHonorMasterData(payload *ProfileHonorPayload, honor *masterdata.BondsHonorInfo, store *masterdata.Store, resolver *assets.Resolver) {
	if payload == nil || honor == nil || store == nil {
		return
	}
	assetResolver := resolverOrDefault(resolver)
	payload.Name = honor.Name
	payload.HonorRarity = honor.HonorRarity
	payload.FrameURL = assetResolver.GetHonorFrameURL(honor.HonorRarity, "main")
	payload.LevelIconURL = assetResolver.GetHonorLevelIconURL(false)
	payload.LevelIcon6URL = assetResolver.GetHonorLevelIconURL(true)

	leftUnit := store.GetCharacterUnit(honor.GameCharacterUnitID1)
	rightUnit := store.GetCharacterUnit(honor.GameCharacterUnitID2)
	if leftUnit == nil || rightUnit == nil {
		return
	}
	if strings.Contains(payload.BondsHonorViewType, "reverse") {
		leftUnit, rightUnit = rightUnit, leftUnit
	}
	payload.LeftCharacterID = leftUnit.GameCharacterID
	payload.RightCharacterID = rightUnit.GameCharacterID
	payload.LeftColor = leftUnit.ColorCode
	payload.RightColor = rightUnit.ColorCode
	payload.LeftCharacterURL = assetResolver.GetBondsHonorCharacterURL(payload.LeftCharacterID)
	payload.RightCharacterURL = assetResolver.GetBondsHonorCharacterURL(payload.RightCharacterID)

	if payload.BondsHonorWordID > 0 {
		if word := store.GetBondsHonorWord(payload.BondsHonorWordID); word != nil {
			payload.BondsHonorWordAssetbundleName = word.AssetbundleName
			payload.BondsHonorWordURL = assetResolver.GetBondsHonorWordURL(word.AssetbundleName)
		}
	}
}

func buildProfileDeckCardPayload(store *masterdata.Store, card sekai.ProfileDeckCard, resolver *assets.Resolver) ProfileDeckCardPayload {
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
			payload.SupplyType = CardSupplyDisplayName(store, *masterCard)
			assetResolver := resolverOrDefault(resolver)
			payload.ThumbnailURL = assetResolver.GetCardThumbnailURL(masterCard.AssetbundleName, false)
			payload.TrainedThumbnailURL = assetResolver.GetCardThumbnailURL(masterCard.AssetbundleName, true)
		}
	}
	return payload
}

// BuildCardDetailPayload adapts a masterdata card into CardDetail renderer props.
func BuildCardDetailPayload(store *masterdata.Store, card masterdata.CardInfo) CardDetailPayload {
	return BuildCardDetailPayloadWithAssets(store, card, defaultAssetResolver())
}

// BuildCardDetailPayloadWithAssets adapts a card using a region-specific asset resolver.
func BuildCardDetailPayloadWithAssets(store *masterdata.Store, card masterdata.CardInfo, resolver *assets.Resolver) CardDetailPayload {
	assetResolver := resolverOrDefault(resolver)
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
		SupplyType:      CardSupplyDisplayName(store, card),
		CardSupplyID:    card.CardSupplyID,
		AssetSource:     assetSourceForResolver(assetResolver),
	}

	if card.AssetbundleName != "" {
		payload.NormalFullURL = assetResolver.GetCardFullURL(card.AssetbundleName, false)
		payload.TrainedFullURL = assetResolver.GetCardFullURL(card.AssetbundleName, true)
		payload.ThumbnailURL = assetResolver.GetCardThumbnailURL(card.AssetbundleName, false)
		payload.TrainedThumbnail = assetResolver.GetCardThumbnailURL(card.AssetbundleName, true)
	}

	if store != nil {
		payload.Events = buildCardEvents(store, card.ID)
		payload.Skill = buildCardSkill(store, card.SkillID, card)
		if card.SpecialTrainingSkillID > 0 && card.SpecialTrainingSkillID != card.SkillID {
			payload.TrainedSkill = buildCardSkill(store, card.SpecialTrainingSkillID, card)
		}
		payload.Costumes = buildCardCostumes(store, card, assetResolver)
	}

	return payload
}

// defaultSkillDisplayLevel is the level we render skill descriptions at.
// Most decks display level 4 (max) which matches sekai.best behaviour.
const defaultSkillDisplayLevel = 4

// buildCardSkill renders a skill description with the placeholder substitution
// rules used by sekai.best / Snowy_Viewer at the maximum displayable level.
func buildCardSkill(store *masterdata.Store, skillID int, card masterdata.CardInfo) *CardSkillPayload {
	if store == nil || skillID <= 0 {
		return nil
	}
	skill := store.GetSkill(skillID)
	if skill == nil {
		return nil
	}
	return &CardSkillPayload{
		ID:          skill.ID,
		Level:       defaultSkillDisplayLevel,
		SpriteName:  skill.DescriptionSpriteName,
		Description: formatSkillDescription(*skill, defaultSkillDisplayLevel, card),
	}
}

// formatSkillDescription replaces {{id;type}} / {{id1,id2;type}} placeholders in
// the skill description with effect values at the requested level. The format
// matches sekai.best's replaceSkillValue() logic.
func formatSkillDescription(skill masterdata.SkillInfo, level int, card masterdata.CardInfo) string {
	if skill.Description == "" {
		return ""
	}

	desc := skill.Description

	single := regexp.MustCompile(`\{\{(\d+);(\w+)\}\}`)
	desc = single.ReplaceAllStringFunc(desc, func(match string) string {
		m := single.FindStringSubmatch(match)
		id, _ := strconv.Atoi(m[1])
		typ := m[2]
		if typ == "c" {
			if name := characterName(card.CharacterID); name != "" {
				return name
			}
			return match
		}
		effect := findSkillEffect(skill.SkillEffects, id)
		if effect == nil {
			return match
		}
		detail := findSkillDetail(effect.SkillEffectDetails, level)
		if detail == nil {
			return match
		}
		switch typ {
		case "d":
			return formatFloat(detail.ActivateEffectDuration)
		case "v":
			return strconv.Itoa(detail.ActivateEffectValue)
		default:
			return match
		}
	})

	double := regexp.MustCompile(`\{\{(\d+),(\d+);(\w+)\}\}`)
	desc = double.ReplaceAllStringFunc(desc, func(match string) string {
		m := double.FindStringSubmatch(match)
		id1, _ := strconv.Atoi(m[1])
		id2, _ := strconv.Atoi(m[2])
		typ := m[3]
		val := func(id int) int {
			effect := findSkillEffect(skill.SkillEffects, id)
			if effect == nil {
				return 0
			}
			detail := findSkillDetail(effect.SkillEffectDetails, level)
			if detail == nil {
				return 0
			}
			return detail.ActivateEffectValue
		}
		v1, v2 := val(id1), val(id2)
		switch typ {
		case "u", "o", "s", "v":
			return strconv.Itoa(v1 + v2)
		case "r":
			if v2 > 0 {
				return strconv.Itoa(v2)
			}
			return strconv.Itoa(v1)
		default:
			return match
		}
	})

	return desc
}

func findSkillEffect(effects []masterdata.SkillEffect, id int) *masterdata.SkillEffect {
	for i := range effects {
		if effects[i].ID == id {
			return &effects[i]
		}
	}
	return nil
}

func findSkillDetail(details []masterdata.SkillEffectDetail, level int) *masterdata.SkillEffectDetail {
	for i := range details {
		if details[i].Level == level {
			return &details[i]
		}
	}
	if len(details) > 0 {
		return &details[len(details)-1]
	}
	return nil
}

func formatFloat(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// buildCardCostumes lists every outfit linked to the card via cardIds and
// resolves thumbnail URLs for the same visual part groups Snowy_Viewer shows:
// shared body/hair/head first, then character-specific extra parts. Hair is often
// stored as a character-specific extraPart instead of a shared part, so we must
// keep the per-card matching extra instead of truncating extras globally.
func buildCardCostumes(store *masterdata.Store, card masterdata.CardInfo, resolver *assets.Resolver) []CardCostumePayload {
	if store == nil || card.ID <= 0 {
		return nil
	}
	costumes := store.GetMoeCostumesByCardID(card.ID)
	if len(costumes) == 0 {
		return nil
	}
	out := make([]CardCostumePayload, 0, len(costumes))
	for _, c := range costumes {
		item := CardCostumePayload{
			CostumeNumber: c.CostumeNumber,
			Name:          c.Name,
			Rarity:        c.Costume3dRarity,
			Source:        c.Source,
			Designer:      c.Designer,
			PartTypes:     costumeDisplayPartTypes(c, card.CharacterID),
		}

		seen := make(map[string]struct{})
		// Shared parts: one assetbundle per display group (first color variant).
		for _, partType := range []string{"body", "hair", "head"} {
			variants, ok := c.Parts[partType]
			if !ok || len(variants) == 0 {
				continue
			}
			for _, assetName := range firstCostumeAssetsByBaseName(variants) {
				appendCostumeThumbnailURL(&item, resolver, seen, assetName)
			}
		}
		// Character-specific extras: keep the current card's character first so
		// unique hairstyles are rendered even when many characters have extra parts.
		for _, extra := range sortedCostumeExtraParts(c.ExtraParts, card.CharacterID) {
			if len(extra.Variants) == 0 {
				continue
			}
			for _, assetName := range firstCostumeAssetsByBaseName(extra.Variants) {
				appendCostumeThumbnailURL(&item, resolver, seen, assetName)
			}
		}
		out = append(out, item)
	}
	return out
}

func appendCostumeThumbnailURL(item *CardCostumePayload, resolver *assets.Resolver, seen map[string]struct{}, assetName string) {
	if assetName == "" {
		return
	}
	if _, ok := seen[assetName]; ok {
		return
	}
	seen[assetName] = struct{}{}
	item.ThumbnailURLs = append(item.ThumbnailURLs, resolver.GetCostumeThumbnailURL(assetName))
}

func firstCostumeAssetsByBaseName(parts []masterdata.MoeCostumePart) []string {
	groups := make(map[string][]masterdata.MoeCostumePart)
	order := make([]string, 0, len(parts))
	for _, part := range parts {
		if part.AssetbundleName == "" {
			continue
		}
		base := costumeVariantBaseName(part.AssetbundleName)
		if _, ok := groups[base]; !ok {
			order = append(order, base)
		}
		groups[base] = append(groups[base], part)
	}

	out := make([]string, 0, len(order))
	for _, base := range order {
		group := groups[base]
		if costumePartGroupHasColorCollision(group) {
			for _, part := range group {
				out = append(out, part.AssetbundleName)
			}
			continue
		}
		out = append(out, group[0].AssetbundleName)
	}
	return out
}

func costumePartGroupHasColorCollision(parts []masterdata.MoeCostumePart) bool {
	seen := make(map[int]struct{}, len(parts))
	for _, part := range parts {
		if _, ok := seen[part.ColorID]; ok {
			return true
		}
		seen[part.ColorID] = struct{}{}
	}
	return false
}

var costumeVariantSuffixRe = regexp.MustCompile(`_\d+$`)

func costumeVariantBaseName(assetName string) string {
	return costumeVariantSuffixRe.ReplaceAllString(assetName, "")
}

func sortedCostumeExtraParts(extras []masterdata.MoeCostumeExtraPart, characterID int) []masterdata.MoeCostumeExtraPart {
	if len(extras) == 0 {
		return nil
	}
	out := append([]masterdata.MoeCostumeExtraPart(nil), extras...)
	sort.SliceStable(out, func(i, j int) bool {
		iMatches := out[i].CharacterID == characterID
		jMatches := out[j].CharacterID == characterID
		if iMatches != jMatches {
			return iMatches
		}
		iScore := costumePartScore(out[i].PartType)
		jScore := costumePartScore(out[j].PartType)
		if iScore != jScore {
			return iScore < jScore
		}
		return false
	})
	return out
}

func costumeDisplayPartTypes(costume masterdata.MoeCostumeInfo, characterID int) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(costume.PartTypes)+len(costume.ExtraParts))
	for _, partType := range costume.PartTypes {
		if partType == "" {
			continue
		}
		seen[partType] = struct{}{}
		out = append(out, partType)
	}
	for _, extra := range sortedCostumeExtraParts(costume.ExtraParts, characterID) {
		partType := extra.PartType
		if partType == "" {
			continue
		}
		if _, ok := seen[partType]; ok {
			continue
		}
		seen[partType] = struct{}{}
		out = append(out, partType)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return costumePartScore(out[i]) < costumePartScore(out[j])
	})
	return out
}

func costumePartScore(partType string) int {
	switch partType {
	case "body":
		return 1
	case "hair":
		return 2
	case "head":
		return 3
	default:
		return 4
	}
}

// BuildMusicDetailPayload adapts a masterdata music row into MusicDetail renderer props.
func buildCardEvents(store *masterdata.Store, cardID int) []CardEventPayload {
	if store == nil || cardID <= 0 {
		return nil
	}
	events := make([]CardEventPayload, 0)
	for _, relation := range store.AllEventCards() {
		if relation.CardID != cardID {
			continue
		}
		event := store.GetEvent(relation.EventID)
		if event == nil {
			continue
		}
		events = append(events, CardEventPayload{
			ID: event.ID, Name: event.Name, EventType: event.EventType,
			AssetbundleName: event.AssetbundleName,
			StartAt:         event.StartAt, ClosedAt: event.ClosedAt, Unit: event.Unit,
		})
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].StartAt != events[j].StartAt {
			return events[i].StartAt < events[j].StartAt
		}
		return events[i].ID < events[j].ID
	})
	return events
}

func BuildMusicDetailPayload(store *masterdata.Store, music masterdata.MusicInfo) MusicDetailPayload {
	return BuildMusicDetailPayloadWithAssets(store, music, defaultAssetResolver())
}

// BuildMusicDetailPayloadWithAssets adapts a music row using a region-specific asset resolver.
func BuildMusicDetailPayloadWithAssets(store *masterdata.Store, music masterdata.MusicInfo, resolver *assets.Resolver) MusicDetailPayload {
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
		AssetSource:         assetSourceForResolver(resolver),
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
	return BuildEventInfoPayloadWithAssets(store, event, defaultAssetResolver())
}

// BuildEventInfoPayloadWithAssets adapts an event using a region-specific asset resolver.
func BuildEventInfoPayloadWithAssets(store *masterdata.Store, event masterdata.EventInfo, resolver *assets.Resolver) EventInfoPayload {
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
		AssetSource:       assetSourceForResolver(resolver),
	}

	if event.AssetbundleName != "" {
		assetResolver := resolverOrDefault(resolver)
		payload.BannerURL = assetResolver.GetEventBannerURL(event.AssetbundleName)
		payload.LogoURL = assetResolver.GetEventLogoURL(event.AssetbundleName)
	}

	if store == nil {
		return payload
	}

	seenCharacters := make(map[string]struct{})
	for _, eventCard := range store.GetEventCards(event.ID) {
		if card := store.GetCard(eventCard.CardID); card != nil {
			payload.BonusCards = append(payload.BonusCards, BuildCardDetailPayloadWithAssets(store, *card, resolver))
		}
	}

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

	// Populate pickup cards from gachas overlapping with the event.
	seenPickupCards := make(map[int]struct{})
	for _, gacha := range store.GetGachasForEvent(event.ID) {
		for _, pickup := range gacha.GachaPickups {
			if _, dup := seenPickupCards[pickup.CardID]; dup {
				continue
			}
			seenPickupCards[pickup.CardID] = struct{}{}
			if card := store.GetCard(pickup.CardID); card != nil {
				payload.PickupCards = append(payload.PickupCards, buildGachaPickupCardPayload(*card, pickup.GachaPickupType, resolver))
			}
		}
	}

	return payload
}

// BuildGachaInfoPayload adapts a masterdata gacha into GachaInfo renderer props.
func BuildGachaInfoPayload(store *masterdata.Store, gacha masterdata.GachaInfo) GachaInfoPayload {
	return BuildGachaInfoPayloadWithAssets(store, gacha, defaultAssetResolver())
}

// BuildGachaInfoPayloadWithAssets adapts a gacha using a region-specific asset resolver.
func BuildGachaInfoPayloadWithAssets(store *masterdata.Store, gacha masterdata.GachaInfo, resolver *assets.Resolver) GachaInfoPayload {
	payload := GachaInfoPayload{
		ID:              gacha.ID,
		Name:            gacha.Name,
		GachaType:       gacha.GachaType,
		AssetbundleName: gacha.AssetbundleName,
		StartAt:         gacha.StartAt,
		EndAt:           gacha.EndAt,
		IsShowPeriod:    gacha.IsShowPeriod,
		WishSelectCount: gacha.WishSelectCount,
		AssetSource:     assetSourceForResolver(resolver),
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
				cardPayload := buildGachaPickupCardPayload(*card, pickup.GachaPickupType, resolver)
				pickupPayload.Card = &cardPayload
				payload.PickupCards = append(payload.PickupCards, cardPayload)
			}
		}

		payload.Pickups = append(payload.Pickups, pickupPayload)
	}

	return payload
}

func BuildCardListPayloadWithAssets(title string, subtitle string, cards []masterdata.CardInfo, store *masterdata.Store, resolver *assets.Resolver, page int, totalPages int, total int) CardListPayload {
	payload := CardListPayload{Title: title, Subtitle: subtitle, Page: page, TotalPages: totalPages, Total: total, AssetSource: assetSourceForResolver(resolver)}
	for _, card := range cards {
		payload.Cards = append(payload.Cards, BuildCardDetailPayloadWithAssets(store, card, resolver))
	}
	return payload
}

func BuildMusicListPayloadWithAssets(title string, subtitle string, musics []masterdata.MusicInfo, store *masterdata.Store, resolver *assets.Resolver, page int, totalPages int, total int) MusicListPayload {
	payload := MusicListPayload{Title: title, Subtitle: subtitle, Page: page, TotalPages: totalPages, Total: total, AssetSource: assetSourceForResolver(resolver)}
	for _, music := range musics {
		payload.Musics = append(payload.Musics, BuildMusicDetailPayloadWithAssets(store, music, resolver))
	}
	return payload
}

func BuildEventListPayloadWithAssets(title string, subtitle string, events []masterdata.EventInfo, store *masterdata.Store, resolver *assets.Resolver, page int, totalPages int, total int) EventListPayload {
	payload := EventListPayload{Title: title, Subtitle: subtitle, Page: page, TotalPages: totalPages, Total: total, AssetSource: assetSourceForResolver(resolver)}
	for _, event := range events {
		eventPayload := BuildEventInfoPayloadWithAssets(store, event, resolver)
		eventPayload.StoryBannerURL = defaultEventListStoryBannerURL(resolver)
		payload.Events = append(payload.Events, eventPayload)
	}
	return payload
}

func defaultEventListStoryBannerURL(resolver *assets.Resolver) string {
	assetResolver := resolverOrDefault(resolver)
	return fmt.Sprintf("%s/event_story/event_show_2026/screen_image/banner_event_story.png", assetResolver.BaseURL())
}

func BuildGachaListPayloadWithAssets(title string, subtitle string, gachas []masterdata.GachaInfo, store *masterdata.Store, resolver *assets.Resolver, page int, totalPages int, total int) GachaListPayload {
	payload := GachaListPayload{Title: title, Subtitle: subtitle, Page: page, TotalPages: totalPages, Total: total, AssetSource: assetSourceForResolver(resolver)}
	for _, gacha := range gachas {
		payload.Gachas = append(payload.Gachas, BuildGachaInfoPayloadWithAssets(store, gacha, resolver))
	}
	return payload
}

func BuildVirtualLiveListPayloadWithAssets(title string, subtitle string, lives []masterdata.VirtualLive, store *masterdata.Store, resolver *assets.Resolver, page int, totalPages int, total int) VirtualLiveListPayload {
	payload := VirtualLiveListPayload{Title: title, Subtitle: subtitle, Page: page, TotalPages: totalPages, Total: total, AssetSource: assetSourceForResolver(resolver)}
	for _, live := range lives {
		payload.VirtualLives = append(payload.VirtualLives, BuildVirtualLivePayloadWithAssets(live, store, resolver))
	}
	return payload
}

func BuildVirtualLivePayloadWithAssets(live masterdata.VirtualLive, store *masterdata.Store, resolver *assets.Resolver) VirtualLivePayload {
	payload := VirtualLivePayload{
		ID:              live.ID,
		Name:            live.Name,
		AssetbundleName: live.AssetbundleName,
		VirtualLiveType: live.VirtualLiveType,
		StartAt:         live.StartAt,
		EndAt:           live.EndAt,
		AssetSource:     assetSourceForResolver(resolver),
	}
	now := time.Now().UnixMilli()
	for _, schedule := range live.VirtualLiveSchedules {
		payload.Schedules = append(payload.Schedules, VirtualLiveSchedulePayload{ID: schedule.ID, Seq: schedule.Seq, StartAt: schedule.StartAt, EndAt: schedule.EndAt})
		if schedule.StartAt > now {
			payload.RestCount++
		}
		if payload.CurrentStartAt == 0 && schedule.EndAt > now {
			payload.CurrentStartAt = schedule.StartAt
			payload.CurrentEndAt = schedule.EndAt
			payload.Living = schedule.StartAt <= now
		}
	}
	for _, reward := range live.VirtualLiveRewards {
		payload.Rewards = append(payload.Rewards, VirtualLiveRewardPayload{ID: reward.ID, VirtualLiveType: reward.VirtualLiveType, ResourceBoxID: reward.ResourceBoxID})
	}
	for _, item := range live.VirtualLiveCharacters {
		characterID := item.GameCharacterID
		if characterID == 0 && store != nil && item.GameCharacterUnitID > 0 {
			if unit := store.GetCharacterUnit(item.GameCharacterUnitID); unit != nil {
				characterID = unit.GameCharacterID
			}
		}
		payload.Characters = append(payload.Characters, VirtualLiveCharacterPayload{
			ID:                  item.ID,
			GameCharacterID:     characterID,
			GameCharacterUnitID: item.GameCharacterUnitID,
			CharacterName:       characterName(characterID),
			PerformanceType:     item.VirtualLivePerformanceType,
		})
	}
	return payload
}

func buildGachaPickupCardPayload(card masterdata.CardInfo, pickupType string, resolver *assets.Resolver) GachaPickupCardPayload {
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
		assetResolver := resolverOrDefault(resolver)
		payload.ThumbnailURL = assetResolver.GetCardThumbnailURL(card.AssetbundleName, false)
		payload.TrainedThumbnailURL = assetResolver.GetCardThumbnailURL(card.AssetbundleName, true)
	}
	return payload
}

func resolverOrDefault(resolver *assets.Resolver) *assets.Resolver {
	if resolver != nil {
		return resolver
	}
	return defaultAssetResolver()
}

func assetSourceForResolver(resolver *assets.Resolver) string {
	return resolverOrDefault(resolver).RendererAssetSource()
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

// CardSupplyDisplayName returns a user-facing card supply label.
func CardSupplyDisplayName(store *masterdata.Store, card masterdata.CardInfo) string {
	if card.CardRarityType == "rarity_birthday" {
		return "生日"
	}
	return CardSupplyTypeDisplayName(CardSupplyType(store, card))
}

// CardSupplyType resolves a card's supply type from loaded masterdata.
func CardSupplyType(store *masterdata.Store, card masterdata.CardInfo) string {
	if card.CardSupplyID <= 0 || store == nil {
		return "normal"
	}
	if supply := store.GetCardSupply(card.CardSupplyID); supply != nil && supply.CardSupplyType != "" {
		return supply.CardSupplyType
	}
	return "normal"
}

// CardSupplyTypeDisplayName converts an upstream cardSupplyType into a label.
func CardSupplyTypeDisplayName(supplyType string) string {
	switch supplyType {
	case "birthday":
		return "生日"
	case "term_limited":
		return "期间限定"
	case "colorful_festival_limited":
		return "CFES限定"
	case "bloom_festival_limited":
		return "BFES限定"
	case "unit_event_limited":
		return "WorldLink限定"
	case "collaboration_limited":
		return "联动限定"
	case "normal", "":
		return "常驻"
	default:
		return supplyType
	}
}

func cleanDash(value string) string {
	if value == "-" {
		return ""
	}
	return value
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonZero(value int, fallback int) int {
	if value != 0 {
		return value
	}
	return fallback
}

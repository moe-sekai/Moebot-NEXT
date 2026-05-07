package renderer

// DeckRecommendCalculateRequest is sent to the Bun renderer's embedded
// TypeScript deck recommender.
type DeckRecommendCalculateRequest struct {
	Region      string                 `json:"region,omitempty"`
	RegionLabel string                 `json:"regionLabel,omitempty"`
	UserData    map[string]any         `json:"userData"`
	MasterData  map[string]any         `json:"masterData"`
	MusicMetas  []map[string]any       `json:"musicMetas"`
	Options     DeckRecommendOptions   `json:"options"`
	CardAssets  map[int]map[string]any `json:"cardAssets,omitempty"`
	MusicAssets map[int]map[string]any `json:"musicAssets,omitempty"`
	Event       any                    `json:"event,omitempty"`
	Music       any                    `json:"music,omitempty"`
	Profile     map[string]any         `json:"profile,omitempty"`
}

type DeckRecommendOptions struct {
	Mode                         string                    `json:"mode,omitempty"`
	EventID                      int                       `json:"eventId"`
	MusicID                      int                       `json:"musicId"`
	Difficulty                   string                    `json:"difficulty"`
	LiveType                     string                    `json:"liveType"`
	Algorithm                    string                    `json:"algorithm,omitempty"`
	Target                       string                    `json:"target,omitempty"`
	Limit                        int                       `json:"limit,omitempty"`
	TimeoutMS                    int                       `json:"timeoutMs,omitempty"`
	FixedCards                   []int                     `json:"fixedCards,omitempty"`
	FixedCharacters              []int                     `json:"fixedCharacters,omitempty"`
	CardConfig                   map[string]DeckCardConfig `json:"cardConfig,omitempty"`
	SkillReferenceChooseStrategy string                    `json:"skillReferenceChooseStrategy,omitempty"`
	KeepAfterTrainingState       bool                      `json:"keepAfterTrainingState,omitempty"`
	BestSkillAsLeader            bool                      `json:"bestSkillAsLeader,omitempty"`
	ChallengeCharacterID         int                       `json:"challengeCharacterId,omitempty"`
	TargetBonus                  int                       `json:"targetBonus,omitempty"`
	TargetBonusList              []int                     `json:"targetBonusList,omitempty"`
	FilterOtherUnit              bool                      `json:"filterOtherUnit,omitempty"`
	SupportCharacterID           int                       `json:"supportCharacterId,omitempty"`
}

type DeckCardConfig struct {
	Disable     bool `json:"disable,omitempty"`
	RankMax     bool `json:"rankMax,omitempty"`
	EpisodeRead bool `json:"episodeRead,omitempty"`
	MasterMax   bool `json:"masterMax,omitempty"`
	SkillMax    bool `json:"skillMax,omitempty"`
	Canvas      bool `json:"canvas,omitempty"`
}

type DeckRecommendCalculateResponse struct {
	OK          bool                  `json:"ok"`
	Region      string                `json:"region,omitempty"`
	RegionLabel string                `json:"regionLabel,omitempty"`
	CostMS      int                   `json:"costMs"`
	Algorithm   string                `json:"algorithm"`
	Warnings    []string              `json:"warnings"`
	Options     DeckRecommendOptions  `json:"options"`
	Profile     map[string]any        `json:"profile,omitempty"`
	Event       any                   `json:"event,omitempty"`
	Music       any                   `json:"music,omitempty"`
	Decks       []DeckRecommendResult `json:"decks"`
	Error       string                `json:"error,omitempty"`
}

type DeckRecommendResult struct {
	Rank             int                 `json:"rank"`
	Value            float64             `json:"value,omitempty"`
	ValueLabel       string              `json:"valueLabel,omitempty"`
	Score            int                 `json:"score"`
	EventPoint       int                 `json:"eventPoint,omitempty"`
	EventBonus       float64             `json:"eventBonus,omitempty"`
	SupportDeckBonus float64             `json:"supportDeckBonus,omitempty"`
	Power            map[string]any      `json:"power,omitempty"`
	MultiLiveScoreUp float64             `json:"multiLiveScoreUp,omitempty"`
	Cards            []DeckRecommendCard `json:"cards"`
}

type DeckRecommendCard struct {
	CardID           int            `json:"cardId"`
	Level            int            `json:"level"`
	SkillLevel       int            `json:"skillLevel"`
	MasterRank       int            `json:"masterRank"`
	DefaultImage     string         `json:"defaultImage,omitempty"`
	EventBonus       string         `json:"eventBonus,omitempty"`
	SupportDeckBonus float64        `json:"supportDeckBonus,omitempty"`
	Power            map[string]any `json:"power,omitempty"`
	Skill            map[string]any `json:"skill,omitempty"`
	Card             map[string]any `json:"card,omitempty"`
}

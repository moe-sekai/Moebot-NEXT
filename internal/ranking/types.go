package ranking

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BaseURL string
	Region  string
	Timeout int
}

type Board struct {
	EventID   int            `json:"event_id"`
	Region    string         `json:"region"`
	BoardType string         `json:"board_type,omitempty"`
	TargetID  int            `json:"target_id,omitempty"`
	StartAt   int64          `json:"start_at,omitempty"`
	EndAt     int64          `json:"end_at,omitempty"`
	UpdatedAt int64          `json:"updated_at"`
	Rankings  []RankingEntry `json:"rankings"`
}

type WorldLinkBoard struct {
	EventID   int              `json:"event_id"`
	Region    string           `json:"region"`
	StartAt   int64            `json:"start_at,omitempty"`
	EndAt     int64            `json:"end_at,omitempty"`
	UpdatedAt int64            `json:"updated_at"`
	Groups    []WorldLinkGroup `json:"groups"`
	Error     string           `json:"error,omitempty"`
}

type WorldLinkGroup struct {
	EventID                      int            `json:"event_id"`
	Region                       string         `json:"region"`
	GameCharacterID              int            `json:"game_character_id"`
	StartAt                      int64          `json:"start_at,omitempty"`
	EndAt                        int64          `json:"end_at,omitempty"`
	UpdatedAt                    int64          `json:"updated_at"`
	UserRankingStatus            string         `json:"user_ranking_status,omitempty"`
	IsWorldBloomChapterAggregate bool           `json:"is_world_bloom_chapter_aggregate,omitempty"`
	Rankings                     []RankingEntry `json:"rankings"`
}

func (g WorldLinkGroup) Board() Board {
	return Board{
		EventID:   g.EventID,
		Region:    g.Region,
		BoardType: "worldlink",
		TargetID:  g.GameCharacterID,
		StartAt:   g.StartAt,
		EndAt:     g.EndAt,
		UpdatedAt: g.UpdatedAt,
		Rankings:  append([]RankingEntry(nil), g.Rankings...),
	}
}

type RankingEntry struct {
	Rank               int             `json:"rank"`
	Score              int64           `json:"score"`
	Name               string          `json:"name"`
	UserID             jsonID          `json:"userId"`
	Word               string          `json:"word,omitempty"`
	ProfileHonors      []ProfileHonor  `json:"profileHonors,omitempty"`
	LeaderCard         *LeaderCard     `json:"leaderCard,omitempty"`
	Churn48h           int             `json:"churn_48h,omitempty"`
	HourlyChurn        []HourlyChurn   `json:"hourly_churn,omitempty"`
	LastChange         *LastChange     `json:"last_change,omitempty"`
	RecentActivity     *RecentActivity `json:"recent_activity,omitempty"`
	RecentScoreChanges []ScoreChange   `json:"recent_score_changes,omitempty"`
	Growth1h           int64           `json:"growth_1h,omitempty"`
	ParkingPeriods     []ParkingPeriod `json:"parking_periods,omitempty"`
	IsTierLine         bool            `json:"isTierLine,omitempty"`
}

type ProfileHonor struct {
	Seq                int    `json:"seq"`
	ProfileHonorType   string `json:"profileHonorType"`
	HonorID            int    `json:"honorId"`
	HonorLevel         int    `json:"honorLevel"`
	BondsHonorViewType string `json:"bondsHonorViewType,omitempty"`
	BondsHonorWordID   int    `json:"bondsHonorWordId,omitempty"`
}

type LeaderCard struct {
	CardID                int    `json:"cardId"`
	Level                 int    `json:"level"`
	MasterRank            int    `json:"masterRank"`
	SpecialTrainingStatus string `json:"specialTrainingStatus"`
	DefaultImage          string `json:"defaultImage"`
	CharacterID           int    `json:"characterId,omitempty"`
}

type LastChange struct {
	Time     int64 `json:"time"`
	OldScore int64 `json:"old_score"`
	NewScore int64 `json:"new_score"`
	Delta    int64 `json:"delta"`
}

type RecentActivity struct {
	Count     int     `json:"count"`
	ChangedAt []int64 `json:"changed_at,omitempty"`
}

type HourlyChurn struct {
	Hour  string `json:"hour"`
	Count int    `json:"count"`
}

type ScoreChange struct {
	Time  int64 `json:"time"`
	Delta int64 `json:"delta"`
}

type ParkingPeriod struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time,omitempty"`
	DurationS int64 `json:"duration_s,omitempty"`
}

type ForecastEvent struct {
	EventID          int    `json:"event_id"`
	Name             string `json:"name"`
	EventType        string `json:"event_type"`
	StartAt          int64  `json:"start_at"`
	EndAt            int64  `json:"end_at"`
	Status           string `json:"status"`
	HasFinalizedData bool   `json:"has_finalized_data"`
	HasRealtimeData  bool   `json:"has_realtime_data"`
}

type ForecastBoard struct {
	EventID   int            `json:"event_id"`
	Region    string         `json:"region,omitempty"`
	Status    string         `json:"status"`
	UpdatedAt flexibleTime   `json:"updated_at"`
	Items     []ForecastItem `json:"items"`
}

func (b ForecastBoard) UpdatedUnixMilli() int64 {
	return b.UpdatedAt.UnixMilli()
}

type ForecastItem struct {
	Rank        int          `json:"rank"`
	Score       int64        `json:"score"`
	Prediction  forecastInt  `json:"prediction"`
	CollectTime flexibleTime `json:"collect_time"`
	IsFinal     bool         `json:"is_final"`
}

func (i ForecastItem) CollectUnixMilli() int64 {
	return i.CollectTime.UnixMilli()
}

func (i ForecastItem) PredictedScore() (int64, bool) {
	return i.Prediction.Int64()
}

type jsonID string

func (id *jsonID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*id = ""
		return nil
	}
	if data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*id = jsonID(strings.TrimSpace(value))
		return nil
	}
	*id = jsonID(strings.TrimSpace(string(data)))
	return nil
}

func (id jsonID) String() string { return string(id) }

type flexibleTime struct {
	value int64
}

func (t *flexibleTime) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		t.value = 0
		return nil
	}
	if data[0] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			t.value = 0
			return nil
		}
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
			t.value = normalizeUnixMillis(n)
			return nil
		}
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return fmt.Errorf("parse time %q: %w", raw, err)
		}
		t.value = parsed.UnixMilli()
		return nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	t.value = normalizeUnixMillis(n)
	return nil
}

func (t flexibleTime) UnixMilli() int64 { return t.value }

type forecastInt struct {
	value int64
	valid bool
}

func (f *forecastInt) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		f.value = 0
		f.valid = false
		return nil
	}
	if data[0] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			f.value = 0
			f.valid = false
			return nil
		}
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		f.value = value
		f.valid = true
		return nil
	}
	var value int64
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	f.value = value
	f.valid = true
	return nil
}

func (f forecastInt) Int64() (int64, bool) { return f.value, f.valid }

func normalizeUnixMillis(value int64) int64 {
	if value > 0 && value < 1_000_000_000_000 {
		return value * 1000
	}
	return value
}

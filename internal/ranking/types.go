package ranking

import (
	"fmt"
)

type Config struct {
	BaseURL string
	Region  string
	Timeout int
}

type Board struct {
	EventID   int            `json:"event_id"`
	Region    string         `json:"region"`
	StartAt   int64          `json:"start_at,omitempty"`
	EndAt     int64          `json:"end_at,omitempty"`
	UpdatedAt int64          `json:"updated_at"`
	Rankings  []RankingEntry `json:"rankings"`
}

type RankingEntry struct {
	Rank               int             `json:"rank"`
	Score              int64           `json:"score"`
	Name               string          `json:"name"`
	UserID             jsonID          `json:"userId"`
	Word               string          `json:"word,omitempty"`
	LeaderCard         *LeaderCard     `json:"leaderCard,omitempty"`
	Churn48h           int             `json:"churn_48h,omitempty"`
	HourlyChurn        []HourlyChurn   `json:"hourly_churn,omitempty"`
	LastChange         *LastChange     `json:"last_change,omitempty"`
	RecentActivity     *RecentActivity `json:"recent_activity,omitempty"`
	RecentScoreChanges []ScoreChange   `json:"recent_score_changes,omitempty"`
	Growth1h           int64           `json:"growth_1h,omitempty"`
	ParkingPeriods     []ParkingPeriod `json:"parking_periods,omitempty"`
}

type LeaderCard struct {
	CardID                int    `json:"cardId"`
	Level                 int    `json:"level"`
	MasterRank            int    `json:"masterRank"`
	SpecialTrainingStatus string `json:"specialTrainingStatus"`
	DefaultImage          string `json:"defaultImage"`
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

type jsonID string

func (id *jsonID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*id = ""
		return nil
	}
	if data[0] == '"' {
		var value string
		if _, err := fmt.Sscanf(string(data), "%q", &value); err != nil {
			return err
		}
		*id = jsonID(value)
		return nil
	}
	*id = jsonID(string(data))
	return nil
}

func (id jsonID) String() string { return string(id) }

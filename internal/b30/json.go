package b30

import (
	"encoding/json"
	"strconv"
	"strings"
)

type userMusicResultJSON struct {
	MusicID             any    `json:"musicId"`
	MusicDifficulty     string `json:"musicDifficulty"`
	MusicDifficultyType string `json:"musicDifficultyType"`
	PlayResult          string `json:"playResult"`
	FullComboFlg        any    `json:"fullComboFlg"`
	FullPerfectFlg      any    `json:"fullPerfectFlg"`
}

func (r *UserMusicResult) UnmarshalJSON(data []byte) error {
	var raw userMusicResultJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.MusicID = flexibleInt(raw.MusicID)
	r.MusicDifficulty = raw.MusicDifficulty
	r.MusicDifficultyType = raw.MusicDifficultyType
	r.PlayResult = raw.PlayResult
	r.FullComboFlg = flexibleBool(raw.FullComboFlg)
	r.FullPerfectFlg = flexibleBool(raw.FullPerfectFlg)
	return nil
}

type legacyUserMusicJSON struct {
	MusicID                     any                 `json:"musicId"`
	UserMusicResults            []LegacyMusicResult `json:"userMusicResults"`
	UserMusicDifficultyStatuses []LegacyMusicStatus `json:"userMusicDifficultyStatuses"`
}

func (m *LegacyUserMusic) UnmarshalJSON(data []byte) error {
	var raw legacyUserMusicJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.MusicID = flexibleInt(raw.MusicID)
	m.UserMusicResults = raw.UserMusicResults
	m.UserMusicDifficultyStatuses = raw.UserMusicDifficultyStatuses
	return nil
}

type legacyMusicStatusJSON struct {
	MusicID             any                 `json:"musicId"`
	MusicDifficulty     string              `json:"musicDifficulty"`
	MusicDifficultyType string              `json:"musicDifficultyType"`
	PlayResult          string              `json:"playResult"`
	FullComboFlg        any                 `json:"fullComboFlg"`
	FullPerfectFlg      any                 `json:"fullPerfectFlg"`
	UserMusicResults    []LegacyMusicResult `json:"userMusicResults"`
}

func (s *LegacyMusicStatus) UnmarshalJSON(data []byte) error {
	var raw legacyMusicStatusJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.MusicID = flexibleInt(raw.MusicID)
	s.MusicDifficulty = raw.MusicDifficulty
	s.MusicDifficultyType = raw.MusicDifficultyType
	s.PlayResult = raw.PlayResult
	s.FullComboFlg = flexibleBool(raw.FullComboFlg)
	s.FullPerfectFlg = flexibleBool(raw.FullPerfectFlg)
	s.UserMusicResults = raw.UserMusicResults
	return nil
}

type legacyMusicResultJSON struct {
	MusicID             any    `json:"musicId"`
	MusicDifficulty     string `json:"musicDifficulty"`
	MusicDifficultyType string `json:"musicDifficultyType"`
	PlayResult          string `json:"playResult"`
	FullComboFlg        any    `json:"fullComboFlg"`
	FullPerfectFlg      any    `json:"fullPerfectFlg"`
}

func (r *LegacyMusicResult) UnmarshalJSON(data []byte) error {
	var raw legacyMusicResultJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.MusicID = flexibleInt(raw.MusicID)
	r.MusicDifficulty = raw.MusicDifficulty
	r.MusicDifficultyType = raw.MusicDifficultyType
	r.PlayResult = raw.PlayResult
	r.FullComboFlg = flexibleBool(raw.FullComboFlg)
	r.FullPerfectFlg = flexibleBool(raw.FullPerfectFlg)
	return nil
}

func flexibleInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		number, _ := typed.Int64()
		return int(number)
	case string:
		text := normalizeNumericString(typed)
		if text == "" {
			return 0
		}
		if number, err := strconv.Atoi(text); err == nil {
			return number
		}
		if number, err := strconv.ParseFloat(text, 64); err == nil {
			return int(number)
		}
	}
	return 0
}

func flexibleBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "1", "yes", "y":
			return true
		}
	case float64:
		return typed != 0
	case int:
		return typed != 0
	}
	return false
}

package b30

// LegacyUserMusic represents older Suite payloads where chart results are nested under userMusics.
type LegacyUserMusic struct {
	MusicID                     int                 `json:"musicId"`
	UserMusicResults            []LegacyMusicResult `json:"userMusicResults"`
	UserMusicDifficultyStatuses []LegacyMusicStatus `json:"userMusicDifficultyStatuses"`
}

type LegacyMusicStatus struct {
	MusicID             int                 `json:"musicId"`
	MusicDifficulty     string              `json:"musicDifficulty"`
	MusicDifficultyType string              `json:"musicDifficultyType"`
	PlayResult          string              `json:"playResult"`
	FullComboFlg        bool                `json:"fullComboFlg"`
	FullPerfectFlg      bool                `json:"fullPerfectFlg"`
	UserMusicResults    []LegacyMusicResult `json:"userMusicResults"`
}

type LegacyMusicResult struct {
	MusicID             int    `json:"musicId"`
	MusicDifficulty     string `json:"musicDifficulty"`
	MusicDifficultyType string `json:"musicDifficultyType"`
	PlayResult          string `json:"playResult"`
	FullComboFlg        bool   `json:"fullComboFlg"`
	FullPerfectFlg      bool   `json:"fullPerfectFlg"`
}

// MergeLegacyResults keeps top-level userMusicResults as source of truth and fills only missing charts from legacy userMusics.
func MergeLegacyResults(primary []UserMusicResult, legacy []LegacyUserMusic) []UserMusicResult {
	out := append([]UserMusicResult(nil), primary...)
	seen := map[string]struct{}{}
	for _, result := range primary {
		if result.MusicID <= 0 || resultDifficulty(result) == "" {
			continue
		}
		seen[resultKey(result.MusicID, resultDifficulty(result))] = struct{}{}
	}
	legacyBest := bestUserMusicResults(legacyResults(legacy))
	for key, result := range legacyBest {
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, result)
	}
	return out
}

func legacyResults(items []LegacyUserMusic) []UserMusicResult {
	out := []UserMusicResult{}
	for _, music := range items {
		musicID := music.MusicID
		for _, result := range music.UserMusicResults {
			out = appendLegacyResult(out, result, musicID, "")
		}
		for _, status := range music.UserMusicDifficultyStatuses {
			diff := NormalizeDifficulty(firstNonEmpty(status.MusicDifficultyType, status.MusicDifficulty))
			out = append(out, UserMusicResult{
				MusicID:             firstNonZero(status.MusicID, musicID),
				MusicDifficulty:     firstNonEmpty(status.MusicDifficulty, diff),
				MusicDifficultyType: firstNonEmpty(status.MusicDifficultyType, diff),
				PlayResult:          status.PlayResult,
				FullComboFlg:        status.FullComboFlg,
				FullPerfectFlg:      status.FullPerfectFlg,
			})
			for _, result := range status.UserMusicResults {
				out = appendLegacyResult(out, result, musicID, diff)
			}
		}
	}
	return out
}

func appendLegacyResult(out []UserMusicResult, result LegacyMusicResult, fallbackMusicID int, fallbackDifficulty string) []UserMusicResult {
	out = append(out, UserMusicResult{
		MusicID:             firstNonZero(result.MusicID, fallbackMusicID),
		MusicDifficulty:     firstNonEmpty(result.MusicDifficulty, fallbackDifficulty),
		MusicDifficultyType: firstNonEmpty(result.MusicDifficultyType, fallbackDifficulty),
		PlayResult:          result.PlayResult,
		FullComboFlg:        result.FullComboFlg,
		FullPerfectFlg:      result.FullPerfectFlg,
	})
	return out
}

func firstNonZero(value int, fallback int) int {
	if value != 0 {
		return value
	}
	return fallback
}

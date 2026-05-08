package b30

import (
	"fmt"
	"sort"
	"strings"
)

const Limit = 30

type PlayResult string

const (
	PlayResultNone  PlayResult = ""
	PlayResultClear PlayResult = "C"
	PlayResultFC    PlayResult = "FC"
	PlayResultAP    PlayResult = "AP"
)

// UserMusicResult is the normalized shape required from Suite user music data.
type UserMusicResult struct {
	MusicID             int    `json:"musicId"`
	MusicDifficulty     string `json:"musicDifficulty"`
	MusicDifficultyType string `json:"musicDifficultyType"`
	PlayResult          string `json:"playResult"`
	FullComboFlg        bool   `json:"fullComboFlg"`
	FullPerfectFlg      bool   `json:"fullPerfectFlg"`
}

// MusicMeta enriches b30 rows with masterdata/asset information.
type MusicMeta struct {
	MusicID          int
	Title            string
	AssetbundleName  string
	JacketURL        string
	Level            int
	NoteCount        int
	PublishedAt      int64
	AdditionalFields map[string]any
}

type MetaResolver func(musicID int, difficulty string, constant ChartConstant) MusicMeta

type Entry struct {
	Rank            int        `json:"rank"`
	MusicID         int        `json:"musicId"`
	Title           string     `json:"title"`
	Difficulty      string     `json:"difficulty"`
	Level           int        `json:"level,omitempty"`
	Constant        float64    `json:"constant"`
	UserRating      float64    `json:"userRating"`
	PlayResult      PlayResult `json:"playResult"`
	NoteCount       int        `json:"noteCount,omitempty"`
	AssetbundleName string     `json:"assetbundleName,omitempty"`
	JacketURL       string     `json:"jacketUrl,omitempty"`
	PublishedAt     int64      `json:"publishedAt,omitempty"`
}

type Result struct {
	Entries               []Entry `json:"entries"`
	Average               float64 `json:"average"`
	CandidateCount        int     `json:"candidateCount"`
	APCount               int     `json:"apCount"`
	FCCount               int     `json:"fcCount"`
	MissingConstantsCount int     `json:"missingConstantsCount"`
	TotalResultCount      int     `json:"totalResultCount"`
}

func Calculate(results []UserMusicResult, table ConstantsTable, resolve MetaResolver) Result {
	best := bestUserMusicResults(results)
	entries := make([]Entry, 0, len(best))
	missing := 0
	apCount := 0
	fcCount := 0
	for _, result := range best {
		playResult := NormalizePlayResult(result)
		if playResult != PlayResultAP && playResult != PlayResultFC {
			continue
		}
		diff := resultDifficulty(result)
		constant, ok := table.Get(result.MusicID, diff)
		if !ok {
			missing++
			continue
		}
		if playResult == PlayResultAP {
			apCount++
		} else {
			fcCount++
		}
		meta := MusicMeta{MusicID: result.MusicID, Title: firstNonEmpty(constant.Title, constant.JPTitle, fmt.Sprintf("歌曲 #%d", result.MusicID)), Level: constant.Level, NoteCount: constant.NoteCount}
		if resolve != nil {
			meta = mergeMeta(meta, resolve(result.MusicID, diff, constant))
		}
		if meta.Title == "" {
			meta.Title = firstNonEmpty(constant.Title, constant.JPTitle, fmt.Sprintf("歌曲 #%d", result.MusicID))
		}
		if meta.Level <= 0 {
			meta.Level = constant.Level
		}
		if meta.NoteCount <= 0 {
			meta.NoteCount = constant.NoteCount
		}
		entries = append(entries, Entry{
			MusicID:         result.MusicID,
			Title:           meta.Title,
			Difficulty:      diff,
			Level:           meta.Level,
			Constant:        constant.Constant,
			UserRating:      UserRating(constant.Constant, playResult),
			PlayResult:      playResult,
			NoteCount:       meta.NoteCount,
			AssetbundleName: meta.AssetbundleName,
			JacketURL:       meta.JacketURL,
			PublishedAt:     meta.PublishedAt,
		})
	}
	SortEntries(entries)
	if len(entries) > Limit {
		entries = entries[:Limit]
	}
	sum := 0.0
	for i := range entries {
		entries[i].Rank = i + 1
		sum += entries[i].UserRating
	}
	average := 0.0
	if len(entries) > 0 {
		average = sum / float64(len(entries))
	}
	return Result{
		Entries:               entries,
		Average:               average,
		CandidateCount:        apCount + fcCount,
		APCount:               apCount,
		FCCount:               fcCount,
		MissingConstantsCount: missing,
		TotalResultCount:      len(best),
	}
}

func UserRating(constant float64, playResult PlayResult) float64 {
	switch playResult {
	case PlayResultAP:
		return constant
	case PlayResultFC:
		if constant >= 33 {
			return constant - 1
		}
		return constant - 1.5
	default:
		return 0
	}
}

func SortEntries(entries []Entry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].UserRating != entries[j].UserRating {
			return entries[i].UserRating > entries[j].UserRating
		}
		if entries[i].Constant != entries[j].Constant {
			return entries[i].Constant > entries[j].Constant
		}
		if playResultPriority(entries[i].PlayResult) != playResultPriority(entries[j].PlayResult) {
			return playResultPriority(entries[i].PlayResult) > playResultPriority(entries[j].PlayResult)
		}
		if entries[i].MusicID != entries[j].MusicID {
			return entries[i].MusicID < entries[j].MusicID
		}
		return difficultyOrder(entries[i].Difficulty) < difficultyOrder(entries[j].Difficulty)
	})
}

func NormalizePlayResult(result UserMusicResult) PlayResult {
	if result.FullPerfectFlg {
		return PlayResultAP
	}
	playResult := strings.ToLower(strings.TrimSpace(result.PlayResult))
	switch playResult {
	case "all_perfect", "full_perfect", "ap", "allperfect", "fullperfect":
		return PlayResultAP
	}
	if result.FullComboFlg {
		return PlayResultFC
	}
	switch playResult {
	case "full_combo", "fc", "fullcombo":
		return PlayResultFC
	case "clear", "c", "live_clear", "success":
		return PlayResultClear
	default:
		return PlayResultNone
	}
}

func bestUserMusicResults(results []UserMusicResult) map[string]UserMusicResult {
	best := map[string]UserMusicResult{}
	for _, result := range results {
		diff := resultDifficulty(result)
		if result.MusicID <= 0 || diff == "" {
			continue
		}
		key := resultKey(result.MusicID, diff)
		if prev, ok := best[key]; !ok || playResultPriority(NormalizePlayResult(result)) > playResultPriority(NormalizePlayResult(prev)) {
			result.MusicDifficultyType = diff
			best[key] = result
		}
	}
	return best
}

func resultDifficulty(result UserMusicResult) string {
	if result.MusicDifficultyType != "" {
		return NormalizeDifficulty(result.MusicDifficultyType)
	}
	return NormalizeDifficulty(result.MusicDifficulty)
}

func resultKey(musicID int, difficulty string) string {
	return fmt.Sprintf("%d:%s", musicID, NormalizeDifficulty(difficulty))
}

func playResultPriority(result PlayResult) int {
	switch result {
	case PlayResultAP:
		return 3
	case PlayResultFC:
		return 2
	case PlayResultClear:
		return 1
	default:
		return 0
	}
}

func difficultyOrder(diff string) int {
	switch NormalizeDifficulty(diff) {
	case "easy":
		return 1
	case "normal":
		return 2
	case "hard":
		return 3
	case "expert":
		return 4
	case "master":
		return 5
	case "append":
		return 6
	default:
		return 99
	}
}

func mergeMeta(base MusicMeta, override MusicMeta) MusicMeta {
	if override.MusicID > 0 {
		base.MusicID = override.MusicID
	}
	if strings.TrimSpace(override.Title) != "" {
		base.Title = strings.TrimSpace(override.Title)
	}
	if strings.TrimSpace(override.AssetbundleName) != "" {
		base.AssetbundleName = strings.TrimSpace(override.AssetbundleName)
	}
	if strings.TrimSpace(override.JacketURL) != "" {
		base.JacketURL = strings.TrimSpace(override.JacketURL)
	}
	if override.Level > 0 {
		base.Level = override.Level
	}
	if override.NoteCount > 0 {
		base.NoteCount = override.NoteCount
	}
	if override.PublishedAt > 0 {
		base.PublishedAt = override.PublishedAt
	}
	if override.AdditionalFields != nil {
		base.AdditionalFields = override.AdditionalFields
	}
	return base
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

package b30

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultConstantsURL = "https://moe.exmeaning.com/data/pjskb30/merged_chart.csv"
	DefaultTimeout      = 10 * time.Second
)

// ChartConstant describes one community chart constant row.
type ChartConstant struct {
	MusicID    int
	Difficulty string
	Constant   float64
	Level      int
	NoteCount  int
	Title      string
	JPTitle    string
	Notes      string
}

// ConstantsTable stores chart constants keyed by music ID and difficulty.
type ConstantsTable struct {
	ByMusic map[int]map[string]ChartConstant
	Entries []ChartConstant
}

func NewConstantsTable(entries []ChartConstant) ConstantsTable {
	table := ConstantsTable{
		ByMusic: make(map[int]map[string]ChartConstant),
		Entries: make([]ChartConstant, 0, len(entries)),
	}
	for _, entry := range entries {
		if entry.MusicID <= 0 || entry.Difficulty == "" || entry.Constant <= 0 {
			continue
		}
		entry.Difficulty = NormalizeDifficulty(entry.Difficulty)
		if entry.Difficulty == "" {
			continue
		}
		if table.ByMusic[entry.MusicID] == nil {
			table.ByMusic[entry.MusicID] = map[string]ChartConstant{}
		}
		table.ByMusic[entry.MusicID][entry.Difficulty] = entry
		table.Entries = append(table.Entries, entry)
	}
	return table
}

func (t ConstantsTable) Get(musicID int, difficulty string) (ChartConstant, bool) {
	byDiff := t.ByMusic[musicID]
	if byDiff == nil {
		return ChartConstant{}, false
	}
	entry, ok := byDiff[NormalizeDifficulty(difficulty)]
	return entry, ok
}

// LoadConstants downloads and parses the community b30 constants CSV.
func LoadConstants(ctx context.Context, url string, timeout time.Duration) (ConstantsTable, error) {
	entries, err := FetchConstants(ctx, url, timeout)
	if err != nil {
		return ConstantsTable{}, err
	}
	return NewConstantsTable(entries), nil
}

func FetchConstants(ctx context.Context, url string, timeout time.Duration) ([]ChartConstant, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		url = DefaultConstantsURL
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build constants request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("constants request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("constants request returned %d", resp.StatusCode)
	}
	entries, err := ParseConstantsCSV(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("constants csv has no usable rows")
	}
	return entries, nil
}

func ParseConstantsCSV(r io.Reader) ([]ChartConstant, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read constants header: %w", err)
	}
	columns := csvHeaderIndex(header)
	required := []string{"constant", "difficulty", "song id"}
	for _, name := range required {
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("constants csv missing %q column", name)
		}
	}
	entries := []ChartConstant{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read constants row: %w", err)
		}
		entry, ok := parseConstantsRecord(record, columns)
		if ok {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func csvHeaderIndex(header []string) map[string]int {
	out := make(map[string]int, len(header))
	for i, name := range header {
		key := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(name, "\ufeff")))
		out[key] = i
	}
	return out
}

func parseConstantsRecord(record []string, columns map[string]int) (ChartConstant, bool) {
	musicID := parseIntColumn(record, columns, "song id")
	constant := parseFloatColumn(record, columns, "constant")
	difficulty := NormalizeDifficulty(stringColumn(record, columns, "difficulty"))
	if musicID <= 0 || constant <= 0 || difficulty == "" {
		return ChartConstant{}, false
	}
	return ChartConstant{
		MusicID:    musicID,
		Difficulty: difficulty,
		Constant:   constant,
		Level:      parseIntColumn(record, columns, "level"),
		NoteCount:  firstPositive(parseIntColumn(record, columns, "note count"), parseIntColumn(record, columns, "notes")),
		Title:      stringColumn(record, columns, "song"),
		JPTitle:    stringColumn(record, columns, "jp name"),
		Notes:      stringColumn(record, columns, "notes"),
	}, true
}

func stringColumn(record []string, columns map[string]int, name string) string {
	idx, ok := columns[name]
	if !ok || idx < 0 || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func parseIntColumn(record []string, columns map[string]int, name string) int {
	value := normalizeNumericString(stringColumn(record, columns, name))
	if value == "" {
		return 0
	}
	if number, err := strconv.Atoi(value); err == nil {
		return number
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return int(floatValue)
}

func parseFloatColumn(record []string, columns map[string]int, name string) float64 {
	value := normalizeNumericString(stringColumn(record, columns, name))
	if value == "" {
		return 0
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return number
}

func normalizeNumericString(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ",", "")
	return value
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func NormalizeDifficulty(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "easy", "ez", "eas", "简单":
		return "easy"
	case "normal", "nm", "nor", "普通":
		return "normal"
	case "hard", "hd", "hrd", "困难":
		return "hard"
	case "expert", "ex", "exp", "专家":
		return "expert"
	case "master", "mas", "ma", "大师":
		return "master"
	case "append", "apd", "app", "追加":
		return "append"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

package assets

import (
	"strconv"
	"strings"

	"moebot-next/internal/config"
)

// ChartSourceURL expands the configured chart SVG URL template.
func ChartSourceURL(template string, musicID int, difficulty string) string {
	template = strings.TrimSpace(template)
	if template == "" {
		template = config.DefaultChartSourceURL
	}
	difficulty = strings.TrimSpace(strings.ToLower(difficulty))
	if difficulty == "" {
		difficulty = "master"
	}
	id := strconv.Itoa(musicID)
	replacer := strings.NewReplacer(
		"{id}", id,
		"{ID}", id,
		"{music_id}", id,
		"{musicId}", id,
		"{difficulty}", difficulty,
		"{Difficulty}", difficulty,
	)
	return replacer.Replace(template)
}

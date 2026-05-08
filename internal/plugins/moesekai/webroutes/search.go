package webroutes

import (
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/musicsearch"

	"github.com/gofiber/fiber/v2"
)

// RegisterSearch registers /api/search/* routes onto the given group. All
// search endpoints query PJSK masterdata and are moesekai-specific.
func RegisterSearch(api fiber.Router, d *Deps) {
	h := &searchHandlers{d: d}
	api.Get("/search/cards", h.cards)
	api.Get("/search/musics", h.musics)
	api.Get("/search/events", h.events)
	api.Get("/search/gachas", h.gachas)
	api.Get("/search/virtual-lives", h.virtualLives)
}

type searchHandlers struct {
	d *Deps
}

// searchStore narrows the masterdata Store interface to the methods the
// search handlers actually call. Mirrors web.Server.defaultStore() but is
// scoped to this package so we don't depend on web internals.
type searchStore interface {
	IsLoaded() bool
	CardCount() int
	SearchCards(q string) []masterdata.CardInfo
	SearchEvents(q string) []masterdata.EventInfo
	SearchGachas(q string) []masterdata.GachaInfo
	AllVirtualLives() []masterdata.VirtualLive
}

func (h *searchHandlers) defaultStore() searchStore {
	if h.d.Servers != nil {
		if runtime := h.d.Servers.Default(); runtime != nil && runtime.Store != nil {
			return runtime.Store
		}
	}
	if h.d.Store == nil {
		return nil
	}
	return h.d.Store
}

func (h *searchHandlers) cards(c *fiber.Ctx) error {
	if err := h.ensureReady(c); err != nil {
		return err
	}
	q := strings.TrimSpace(c.Query("q"))
	results := h.defaultStore().SearchCards(q)
	rows := make([]fiber.Map, 0, len(results))
	for _, card := range results {
		rows = append(rows, fiber.Map{
			"id":              card.ID,
			"title":           card.Prefix,
			"subtitle":        fmt.Sprintf("角色 #%d · %s", card.CharacterID, card.CardRarityType),
			"type":            "card",
			"character_id":    card.CharacterID,
			"rarity":          card.CardRarityType,
			"attr":            card.Attr,
			"assetbundleName": card.AssetbundleName,
		})
	}
	return searchResponse(c, q, rows)
}

func (h *searchHandlers) musics(c *fiber.Ctx) error {
	if err := h.ensureReady(c); err != nil {
		return err
	}
	q := strings.TrimSpace(c.Query("q"))
	store := h.d.Store
	var aliases map[int]assets.MusicAlias
	if h.d.Servers != nil {
		if runtime := h.d.Servers.Default(); runtime != nil {
			store = runtime.Store
			aliases = runtime.MusicAliases
		}
	}
	result := musicsearch.Search(store, aliases, q, musicsearch.Options{Limit: 25})
	results := result.DisplayMusics()
	rows := make([]fiber.Map, 0, len(results))
	for _, music := range results {
		rows = append(rows, fiber.Map{
			"id":              music.ID,
			"title":           music.Title,
			"subtitle":        strings.Join(nonEmptyStrings(music.Composer, music.Lyricist, music.Arranger), " / "),
			"type":            "music",
			"pronunciation":   music.Pronunciation,
			"composer":        music.Composer,
			"lyricist":        music.Lyricist,
			"arranger":        music.Arranger,
			"assetbundleName": music.AssetbundleName,
		})
	}
	return searchResponse(c, q, rows)
}

func (h *searchHandlers) events(c *fiber.Ctx) error {
	if err := h.ensureReady(c); err != nil {
		return err
	}
	q := strings.TrimSpace(c.Query("q"))
	results := h.defaultStore().SearchEvents(q)
	rows := make([]fiber.Map, 0, len(results))
	for _, event := range results {
		rows = append(rows, fiber.Map{
			"id":              event.ID,
			"title":           event.Name,
			"subtitle":        fmt.Sprintf("%s · %s", event.EventType, event.Unit),
			"type":            "event",
			"event_type":      event.EventType,
			"unit":            event.Unit,
			"start_at":        event.StartAt,
			"closed_at":       event.ClosedAt,
			"assetbundleName": event.AssetbundleName,
		})
	}
	return searchResponse(c, q, rows)
}

func (h *searchHandlers) gachas(c *fiber.Ctx) error {
	c.Locals("allow_empty_query", true)
	if err := h.ensureReady(c); err != nil {
		return err
	}
	q := strings.TrimSpace(c.Query("q"))
	store := h.defaultStore()
	results := store.SearchGachas(q)
	if q == "" || strings.EqualFold(q, "当前") {
		if full, ok := store.(interface{ AllGachas() []masterdata.GachaInfo }); ok {
			results = currentGachasForWeb(full.AllGachas())
		}
	}
	rows := make([]fiber.Map, 0, len(results))
	for _, gacha := range results {
		rows = append(rows, fiber.Map{
			"id":              gacha.ID,
			"title":           gacha.Name,
			"subtitle":        gacha.GachaType,
			"type":            "gacha",
			"gacha_type":      gacha.GachaType,
			"start_at":        gacha.StartAt,
			"end_at":          gacha.EndAt,
			"assetbundleName": gacha.AssetbundleName,
		})
	}
	return searchResponse(c, q, rows)
}

func (h *searchHandlers) virtualLives(c *fiber.Ctx) error {
	c.Locals("allow_empty_query", true)
	if err := h.ensureReady(c); err != nil {
		return err
	}
	q := strings.TrimSpace(c.Query("q"))
	results := searchVirtualLivesForWeb(h.defaultStore().AllVirtualLives(), q)
	rows := make([]fiber.Map, 0, len(results))
	for _, live := range results {
		start, end := virtualLiveBoundsForWeb(live)
		rows = append(rows, fiber.Map{
			"id":                live.ID,
			"title":             live.Name,
			"subtitle":          fmt.Sprintf("%s - %s", formatWebMillis(start), formatWebMillis(end)),
			"type":              "virtual_live",
			"virtual_live_type": live.VirtualLiveType,
			"start_at":          start,
			"end_at":            end,
			"assetbundleName":   live.AssetbundleName,
		})
	}
	return searchResponse(c, q, rows)
}

func (h *searchHandlers) ensureReady(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	allowEmpty := c.Locals("allow_empty_query") == true
	if q == "" && !allowEmpty {
		return c.JSON(fiber.Map{
			"data":    []fiber.Map{},
			"total":   0,
			"query":   q,
			"message": "请输入关键词后再搜索。",
		})
	}
	store := h.defaultStore()
	if store == nil || !store.IsLoaded() {
		return c.JSON(fiber.Map{
			"data":    []fiber.Map{},
			"total":   0,
			"query":   q,
			"message": "Masterdata 尚未加载，无法搜索；请先在「数据」页面更新数据。",
		})
	}
	return nil
}

func searchResponse(c *fiber.Ctx, q string, rows []fiber.Map) error {
	message := "搜索完成"
	if len(rows) == 0 {
		message = "没有找到匹配结果。"
	}
	return c.JSON(fiber.Map{
		"data":    rows,
		"total":   len(rows),
		"query":   q,
		"message": message,
	})
}

func nonEmptyStrings(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func currentGachasForWeb(gachas []masterdata.GachaInfo) []masterdata.GachaInfo {
	now := time.Now().UnixMilli()
	out := make([]masterdata.GachaInfo, 0)
	for _, gacha := range gachas {
		if gacha.StartAt <= now && (gacha.EndAt <= 0 || now <= gacha.EndAt) {
			out = append(out, gacha)
		}
	}
	if len(out) > 0 {
		return out
	}
	for _, gacha := range gachas {
		if gacha.StartAt <= now {
			out = append(out, gacha)
		}
	}
	if len(out) > 12 {
		return out[len(out)-12:]
	}
	return out
}

func searchVirtualLivesForWeb(lives []masterdata.VirtualLive, q string) []masterdata.VirtualLive {
	now := time.Now().UnixMilli()
	q = strings.TrimSpace(strings.ToLower(q))
	out := make([]masterdata.VirtualLive, 0)
	for _, live := range lives {
		start, end := virtualLiveBoundsForWeb(live)
		if q == "" {
			if end > now && start-now < int64(7*24*time.Hour/time.Millisecond) {
				out = append(out, live)
			}
			continue
		}
		if fmt.Sprintf("%d", live.ID) == q || strings.Contains(strings.ToLower(live.Name), q) || strings.Contains(strings.ToLower(live.AssetbundleName), q) {
			out = append(out, live)
		}
	}
	return out
}

func virtualLiveBoundsForWeb(live masterdata.VirtualLive) (int64, int64) {
	start, end := live.StartAt, live.EndAt
	for i, schedule := range live.VirtualLiveSchedules {
		if i == 0 || schedule.StartAt < start || start == 0 {
			start = schedule.StartAt
		}
		if schedule.EndAt > end {
			end = schedule.EndAt
		}
	}
	return start, end
}

func formatWebMillis(ms int64) string {
	if ms <= 0 {
		return "-"
	}
	return time.UnixMilli(ms).Format("2006-01-02 15:04")
}

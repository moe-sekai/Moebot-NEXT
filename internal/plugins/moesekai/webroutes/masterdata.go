package webroutes

import (
	"fmt"
	"time"

	"moebot-next/internal/config"

	"github.com/gofiber/fiber/v2"
)

// RegisterMasterdata registers /api/masterdata/{summary,reload}. Both endpoints
// inspect / mutate PJSK masterdata state owned by the moesekai plugin.
func RegisterMasterdata(api fiber.Router, d *Deps) {
	h := &masterdataHandlers{d: d}
	api.Get("/masterdata/summary", h.summary)
	api.Post("/masterdata/reload", h.reload)
}

type masterdataHandlers struct {
	d *Deps
}

// summaryStore is the small interface needed for the summary/reload responses.
type summaryStore interface {
	IsLoaded() bool
	LoadedAt() time.Time
	CardCount() int
	MusicCount() int
	EventCount() int
	GachaCount() int
	VirtualLiveCount() int
}

func (h *masterdataHandlers) defaultStore() summaryStore {
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

func (h *masterdataHandlers) summary(c *fiber.Ctx) error {
	store := h.defaultStore()
	loaded := store != nil && store.IsLoaded()
	loadedAt := time.Time{}
	if store != nil {
		loadedAt = store.LoadedAt()
	}
	return c.JSON(fiber.Map{
		"loaded":    loaded,
		"loaded_at": nullableTime(loadedAt),
		"counts":    summaryCounts(store),
	})
}

func (h *masterdataHandlers) reload(c *fiber.Ctx) error {
	region := config.NormalizeRegion(c.Query("region"))
	started := time.Now()
	if h.d.Servers != nil {
		runtime, err := h.d.Servers.Reload(region)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		return c.JSON(fiber.Map{
			"ok":          true,
			"message":     fmt.Sprintf("%s Masterdata 已重新加载", runtime.Label),
			"region":      runtime.Region,
			"duration_ms": time.Since(started).Milliseconds(),
			"loaded_at":   nullableTime(runtime.Store.LoadedAt()),
			"counts":      summaryCounts(runtime.Store),
		})
	}
	if h.d.Loader == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Masterdata loader is not configured")
	}
	if err := h.d.Loader.LoadAll(); err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	store := h.defaultStore()
	loadedAt := time.Time{}
	if store != nil {
		loadedAt = store.LoadedAt()
	}
	return c.JSON(fiber.Map{
		"ok":          true,
		"message":     "Masterdata 已重新加载",
		"duration_ms": time.Since(started).Milliseconds(),
		"loaded_at":   nullableTime(loadedAt),
		"counts":      summaryCounts(store),
	})
}

func summaryCounts(store summaryStore) fiber.Map {
	if store == nil {
		return fiber.Map{
			"cards":         0,
			"musics":        0,
			"events":        0,
			"gachas":        0,
			"virtual_lives": 0,
		}
	}
	return fiber.Map{
		"cards":         store.CardCount(),
		"musics":        store.MusicCount(),
		"events":        store.EventCount(),
		"gachas":        store.GachaCount(),
		"virtual_lives": store.VirtualLiveCount(),
	}
}

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

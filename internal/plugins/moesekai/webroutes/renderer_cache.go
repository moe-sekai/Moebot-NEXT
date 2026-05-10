package webroutes

import (
	"fmt"
	"sort"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/renderer"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// RegisterRendererCache registers /api/renderer/cache/card-thumbnails* routes
// onto the given group. PJSK card thumbnail preload is moesekai-specific.
func RegisterRendererCache(api fiber.Router, d *Deps) {
	h := &rendererCacheHandlers{d: d}
	api.Get("/renderer/cache/card-thumbnails", h.status)
	api.Post("/renderer/cache/card-thumbnails/preload", h.preload)
}

type rendererCacheHandlers struct {
	d *Deps
}

type cardThumbnailCacheTarget struct {
	Region string
	Label  string
	Store  interface {
		IsLoaded() bool
		CardCount() int
		AllCards() []masterdata.CardInfo
	}
	Resolver *assets.Resolver
}

func (h *rendererCacheHandlers) status(c *fiber.Ctx) error {
	target, urls, cards, err := h.cacheURLs(c.Query("region"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if h.d.Renderer == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Renderer client is not configured")
	}
	status, err := h.d.Renderer.CardThumbnailPreloadStatusWithCards(urls, cards)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(cardThumbnailCacheResponse(target, urls, status, "卡牌缩略图缓存状态已返回"))
}

func (h *rendererCacheHandlers) preload(c *fiber.Ctx) error {
	target, urls, cards, err := h.cacheURLs(c.Query("region"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if h.d.Renderer == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Renderer client is not configured")
	}
	force := c.QueryBool("force", false)
	startedAt := time.Now()
	log.Info().Str("region", target.Region).Int("urls", len(urls)).Int("composites", len(cards)).Bool("force", force).Msg("Starting renderer card thumbnail preload")
	status, err := h.d.Renderer.StartCardThumbnailPreloadWithCards(urls, cards, force)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	log.Info().Str("region", target.Region).Bool("running", status.Running).Int("cached", status.Cached).Int("total", status.Total).Int("composite_cached", status.CompositeCached).Int("composite_total", status.CompositeTotal).Dur("elapsed", time.Since(startedAt)).Msg("Renderer card thumbnail preload accepted")
	return c.JSON(cardThumbnailCacheResponse(target, urls, status, "卡牌缩略图预载已启动"))
}

func (h *rendererCacheHandlers) cacheURLs(rawRegion string) (*cardThumbnailCacheTarget, []string, []renderer.CardThumbnailPreloadCard, error) {
	target, err := h.cacheTarget(rawRegion)
	if err != nil {
		return nil, nil, nil, err
	}
	if target.Store == nil || !target.Store.IsLoaded() {
		return nil, nil, nil, fmt.Errorf("%s Masterdata 尚未加载，无法生成卡牌缩略图列表", target.Label)
	}
	resolver := target.Resolver
	if resolver == nil {
		resolver = assets.DefaultResolver()
	}
	seen := make(map[string]struct{})
	preloadCards := make([]renderer.CardThumbnailPreloadCard, 0, target.Store.CardCount()*2)
	for _, card := range target.Store.AllCards() {
		if card.AssetbundleName == "" {
			continue
		}
		normalURL := resolver.GetCardThumbnailURL(card.AssetbundleName, false)
		seen[normalURL] = struct{}{}
		for _, size := range cardThumbnailCompositeSizes() {
			preloadCards = append(preloadCards, renderer.CardThumbnailPreloadCard{ImageURL: normalURL, Rarity: card.CardRarityType, Attr: card.Attr, Trained: false, Size: size})
		}
		if cardCanUseTrainedThumbnail(card.CardRarityType) {
			trainedURL := resolver.GetCardThumbnailURL(card.AssetbundleName, true)
			seen[trainedURL] = struct{}{}
			for _, size := range cardThumbnailCompositeSizes() {
				preloadCards = append(preloadCards, renderer.CardThumbnailPreloadCard{ImageURL: trainedURL, Rarity: card.CardRarityType, Attr: card.Attr, Trained: true, Size: size})
			}
		}
	}
	urls := make([]string, 0, len(seen))
	for url := range seen {
		urls = append(urls, url)
	}
	sort.Strings(urls)
	return target, urls, preloadCards, nil
}

func (h *rendererCacheHandlers) cacheTarget(rawRegion string) (*cardThumbnailCacheTarget, error) {
	region := config.NormalizeRegion(rawRegion)
	if h.d.Servers != nil {
		runtime := h.d.Servers.Get(region)
		if runtime == nil {
			return nil, fmt.Errorf("未找到可用区服")
		}
		return &cardThumbnailCacheTarget{
			Region:   runtime.Region,
			Label:    runtime.Label,
			Store:    runtime.Store,
			Resolver: runtime.Assets,
		}, nil
	}
	if region == "" {
		region = config.NormalizeRegion(h.d.Config.Server.Region)
	}
	if region == "" {
		region = config.RegionJP
	}
	resolver, err := assets.NewResolver(h.d.Config.Assets, region)
	if err != nil {
		resolver = assets.DefaultResolver()
	}
	return &cardThumbnailCacheTarget{
		Region:   region,
		Label:    config.RegionLabel(region),
		Store:    h.d.Store,
		Resolver: resolver,
	}, nil
}

func cardThumbnailCacheResponse(target *cardThumbnailCacheTarget, urls []string, status *renderer.AssetPreloadStatus, message string) fiber.Map {
	response := fiber.Map{
		"ok":                     status.OK,
		"message":                message,
		"region":                 target.Region,
		"region_label":           target.Label,
		"total_cards":            target.Store.CardCount(),
		"total_urls":             len(urls),
		"total_composite_images": status.CompositeTotal,
	}
	mergeCardThumbnailStatus(response, status)
	return response
}

func mergeCardThumbnailStatus(response fiber.Map, status *renderer.AssetPreloadStatus) {
	response["enabled"] = status.Enabled
	response["running"] = status.Running
	response["cache_dir"] = status.CacheDir
	response["total"] = status.Total
	response["cached"] = status.Cached
	response["missing"] = status.Missing
	response["failed"] = status.Failed
	response["downloaded"] = status.Downloaded
	response["skipped"] = status.Skipped
	response["progress"] = status.Progress
	response["started_at"] = status.StartedAt
	response["completed_at"] = status.CompletedAt
	response["errors"] = status.Errors
	response["composite_total"] = status.CompositeTotal
	response["composite_cached"] = status.CompositeCached
	response["composite_missing"] = status.CompositeMissing
	response["composite_failed"] = status.CompositeFailed
	response["composite_generated"] = status.CompositeGenerated
	response["composite_progress"] = status.CompositeProgress
	response["composite_source_downloaded"] = status.CompositeSourceDownloaded
	response["composite_source_failed"] = status.CompositeSourceFailed
	response["composite_render_ms"] = status.CompositeRenderMS
	if status.Message != "" {
		response["renderer_message"] = status.Message
	}
}

func cardThumbnailCompositeSizes() []int {
	return []int{46, 58, 64, 88, 112, 128}
}

func cardCanUseTrainedThumbnail(rarityType string) bool {
	return rarityType == "rarity_3" || rarityType == "rarity_4"
}

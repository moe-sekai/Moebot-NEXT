package web

import (
	"fmt"
	"sort"

	"moebot-next/internal/assets"
	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/renderer"

	"github.com/gofiber/fiber/v2"
)

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

func (s *Server) handleRendererCardThumbnailCacheStatus(c *fiber.Ctx) error {
	target, urls, err := s.cardThumbnailCacheURLs(c.Query("region"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if s.Renderer == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Renderer client is not configured")
	}
	status, err := s.Renderer.CardThumbnailPreloadStatus(urls)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(cardThumbnailCacheResponse(target, urls, status, "卡牌缩略图缓存状态已返回"))
}

func (s *Server) handleRendererCardThumbnailPreload(c *fiber.Ctx) error {
	target, urls, err := s.cardThumbnailCacheURLs(c.Query("region"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if s.Renderer == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Renderer client is not configured")
	}
	status, err := s.Renderer.StartCardThumbnailPreload(urls)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(cardThumbnailCacheResponse(target, urls, status, "卡牌缩略图预载已启动"))
}

func (s *Server) cardThumbnailCacheURLs(rawRegion string) (*cardThumbnailCacheTarget, []string, error) {
	target, err := s.cardThumbnailCacheTarget(rawRegion)
	if err != nil {
		return nil, nil, err
	}
	if target.Store == nil || !target.Store.IsLoaded() {
		return nil, nil, fmt.Errorf("%s Masterdata 尚未加载，无法生成卡牌缩略图列表", target.Label)
	}
	resolver := target.Resolver
	if resolver == nil {
		resolver = assets.DefaultResolver()
	}
	seen := make(map[string]struct{})
	for _, card := range target.Store.AllCards() {
		if card.AssetbundleName == "" {
			continue
		}
		seen[resolver.GetCardThumbnailURL(card.AssetbundleName, false)] = struct{}{}
		if cardCanUseTrainedThumbnail(card.CardRarityType) {
			seen[resolver.GetCardThumbnailURL(card.AssetbundleName, true)] = struct{}{}
		}
	}
	urls := make([]string, 0, len(seen))
	for url := range seen {
		urls = append(urls, url)
	}
	sort.Strings(urls)
	return target, urls, nil
}

func (s *Server) cardThumbnailCacheTarget(rawRegion string) (*cardThumbnailCacheTarget, error) {
	region := config.NormalizeRegion(rawRegion)
	if s.Servers != nil {
		runtime := s.Servers.Get(region)
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
		region = config.NormalizeRegion(s.Config.Server.Region)
	}
	if region == "" {
		region = config.RegionJP
	}
	resolver, err := assets.NewResolver(s.Config.Assets, region)
	if err != nil {
		resolver = assets.DefaultResolver()
	}
	return &cardThumbnailCacheTarget{
		Region:   region,
		Label:    config.RegionLabel(region),
		Store:    s.Store,
		Resolver: resolver,
	}, nil
}

func cardThumbnailCacheResponse(target *cardThumbnailCacheTarget, urls []string, status *renderer.AssetPreloadStatus, message string) fiber.Map {
	response := fiber.Map{
		"ok":           status.OK,
		"message":      message,
		"region":       target.Region,
		"region_label": target.Label,
		"total_cards":  target.Store.CardCount(),
		"total_urls":   len(urls),
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
	if status.Message != "" {
		response["renderer_message"] = status.Message
	}
}

func cardCanUseTrainedThumbnail(rarityType string) bool {
	return rarityType == "rarity_3" || rarityType == "rarity_4"
}

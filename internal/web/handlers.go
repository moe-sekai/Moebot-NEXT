package web

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/assets"
	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/renderer"

	"github.com/gofiber/fiber/v2"
)

const appVersion = "0.1.0"

// handleHealth returns a lightweight health response for probes and the dashboard.
func (s *Server) handleHealth(c *fiber.Ctx) error {
	now := time.Now()
	return c.JSON(fiber.Map{
		"status":  "ok",
		"version": appVersion,
		"time":    now,
		"uptime":  time.Since(s.startedAt).String(),
	})
}

// handleDashboard returns an overview of the bot's status.
func (s *Server) handleDashboard(c *fiber.Ctx) error {
	cmdCount, userCount, groupCount := s.DB.GetTotalStats()

	return c.JSON(fiber.Map{
		"commands_total": cmdCount,
		"users_total":    userCount,
		"groups_total":   groupCount,
		"uptime":         time.Since(s.startedAt).String(),
		"version":        appVersion,
	})
}

// handleStatus returns a combined runtime status for the admin dashboard.
func (s *Server) handleStatus(c *fiber.Ctx) error {
	rendererOK, rendererCode, rendererErr, rendererLatency := s.checkRenderer()
	databaseOK := true
	databaseMessage := "SQLite 已连接"
	if err := s.DB.Ping(); err != nil {
		databaseOK = false
		databaseMessage = err.Error()
	}

	store := s.defaultStore()
	masterdataLoaded := false
	masterdataLoadedAt := time.Time{}
	if store != nil {
		masterdataLoaded = store.IsLoaded()
		masterdataLoadedAt = store.LoadedAt()
	}

	return c.JSON(fiber.Map{
		"version": appVersion,
		"time":    time.Now(),
		"uptime":  time.Since(s.startedAt).String(),
		"bot": fiber.Map{
			"status":         "configured",
			"ok":             true,
			"message":        "ZeroBot 已配置；等待 OneBot v11 反向 WebSocket 连接",
			"driver_type":    s.Config.Bot.Driver.Type,
			"listen":         s.Config.Bot.Driver.Listen,
			"url_configured": s.Config.Bot.Driver.URL != "",
			"command_prefix": s.Config.Bot.CommandPrefix,
			"nicknames":      s.Config.Bot.Nickname,
		},
		"web": fiber.Map{
			"status":  "ok",
			"ok":      true,
			"message": "Fiber 管理面板运行中",
			"host":    s.Config.Web.Host,
			"port":    s.Config.Web.Port,
		},
		"renderer": fiber.Map{
			"status":         statusText(rendererOK),
			"ok":             rendererOK,
			"message":        rendererMessage(rendererOK, rendererErr),
			"base_url":       rendererBaseURL(s.Renderer),
			"status_code":    rendererCode,
			"latency_ms":     rendererLatency.Milliseconds(),
			"service_port":   s.Config.Renderer.Port,
			"dashboard_port": s.Config.Web.Port,
			"precision":      s.Config.Renderer.Precision,
		},
		"masterdata": fiber.Map{
			"status":    statusText(masterdataLoaded),
			"ok":        masterdataLoaded,
			"message":   masterdataMessage(masterdataLoaded),
			"loaded":    masterdataLoaded,
			"loaded_at": nullableTime(masterdataLoadedAt),
			"counts":    s.masterdataSummaryMap(),
		},
		"database": fiber.Map{
			"status":  statusText(databaseOK),
			"ok":      databaseOK,
			"message": databaseMessage,
			"path":    s.Config.Database.Path,
		},
	})
}

// handleMasterdataSummary returns loaded masterdata counts.
func (s *Server) handleMasterdataSummary(c *fiber.Ctx) error {
	store := s.defaultStore()
	loaded := store != nil && store.IsLoaded()
	loadedAt := time.Time{}
	if store != nil {
		loadedAt = store.LoadedAt()
	}

	return c.JSON(fiber.Map{
		"loaded":    loaded,
		"loaded_at": nullableTime(loadedAt),
		"counts":    s.masterdataSummaryMap(),
	})
}

// handleRendererHealth returns renderer reachability and public renderer info.
func (s *Server) handleRendererHealth(c *fiber.Ctx) error {
	ok, statusCode, err, latency := s.checkRenderer()
	return c.JSON(fiber.Map{
		"ok":             ok,
		"status":         statusText(ok),
		"message":        rendererMessage(ok, err),
		"base_url":       rendererBaseURL(s.Renderer),
		"status_code":    statusCode,
		"latency_ms":     latency.Milliseconds(),
		"renderer_port":  s.Config.Renderer.Port,
		"dashboard_port": s.Config.Web.Port,
		"precision":      s.Config.Renderer.Precision,
		"note":           fmt.Sprintf("%d 是 Satori/Bun 图片渲染服务，%d 才是 Moebot NEXT 管理面板。", s.Config.Renderer.Port, s.Config.Web.Port),
	})
}

// handleRendererPreviews returns Satori preview metadata through the Go admin API.
func (s *Server) handleRendererPreviews(c *fiber.Ctx) error {
	if s.Renderer == nil {
		return c.JSON(fiber.Map{
			"data":    []renderer.PreviewMeta{},
			"total":   0,
			"ok":      false,
			"message": "Renderer client is not configured",
		})
	}

	previews, err := s.Renderer.ListPreviews()
	if err != nil {
		return c.JSON(fiber.Map{
			"data":    []renderer.PreviewMeta{},
			"total":   0,
			"ok":      false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    previews,
		"total":   len(previews),
		"ok":      true,
		"message": "Satori 预览模板已加载",
	})
}

// handleRendererPreviewImage proxies a rendered Satori preview PNG.
func (s *Server) handleRendererPreviewImage(c *fiber.Ctx) error {
	if s.Renderer == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Renderer client is not configured")
	}

	width, _ := strconv.Atoi(c.Query("width"))
	height, _ := strconv.Atoi(c.Query("height"))
	started := time.Now()
	result, err := s.Renderer.RenderPreviewWithTrace(c.Params("id"), width, height)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	c.Set(fiber.HeaderContentType, "image/png")
	c.Set(fiber.HeaderCacheControl, "no-store")
	setOptionalHeader(c, "x-render-total-ms", result.TotalMS)
	setOptionalHeader(c, "x-render-fonts-ms", result.FontsMS)
	setOptionalHeader(c, "x-render-images-ms", result.ImagesMS)
	setOptionalHeader(c, "x-render-satori-ms", result.SatoriMS)
	setOptionalHeader(c, "x-render-resvg-ms", result.ResvgMS)
	setOptionalHeader(c, "x-render-size-bytes", result.SizeBytes)
	setOptionalHeader(c, "x-render-image-total", result.ImageTotal)
	setOptionalHeader(c, "x-render-image-remote", result.ImageRemote)
	setOptionalHeader(c, "x-render-image-cache-hits", result.ImageCacheHits)
	setOptionalHeader(c, "x-render-image-cache-misses", result.ImageCacheMisses)
	setOptionalHeader(c, "x-render-image-cache-errors", result.ImageCacheErrors)
	c.Set("x-render-proxy-ms", strconv.FormatInt(time.Since(started).Milliseconds(), 10))
	return c.Send(result.PNG)
}

// handleRecentCommands returns recent command invocation rows.
func (s *Server) handleRecentCommands(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	commands, err := s.DB.ListRecentCommands(limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list recent commands")
	}

	message := "最近命令记录已返回"
	if len(commands) == 0 {
		message = "暂无命令记录；机器人收到并记录指令后这里会自动显示。"
	}

	return c.JSON(fiber.Map{
		"data":    commands,
		"total":   len(commands),
		"message": message,
	})
}

// handlePublicConfig returns only non-sensitive configuration values.
func (s *Server) handlePublicConfig(c *fiber.Ctx) error {
	return c.JSON(s.publicConfigMap())
}

type updatePublicConfigRequest struct {
	Server            *serverSettingsRequest               `json:"server"`
	Servers           map[string]gameServerSettingsRequest `json:"servers"`
	Masterdata        *masterdataSettingsRequest           `json:"masterdata"`
	Assets            *assetsSettingsRequest               `json:"assets"`
	Renderer          *rendererSettingsRequest             `json:"renderer"`
	ReloadMasterdata  bool                                 `json:"reload_masterdata"`
	SyncClientRegions bool                                 `json:"sync_client_regions"`
}

type serverSettingsRequest struct {
	Region string `json:"region"`
}

type rendererSettingsRequest struct {
	Precision float64 `json:"precision"`
}

type gameServerSettingsRequest struct {
	Enabled    *bool                      `json:"enabled"`
	Masterdata *masterdataSettingsRequest `json:"masterdata"`
	Assets     *assetsSettingsRequest     `json:"assets"`
	SekaiAPI   *sekaiAPISettingsRequest   `json:"sekai_api"`
	SuiteAPI   *suiteAPISettingsRequest   `json:"suite_api"`
	RankingAPI *rankingAPISettingsRequest `json:"ranking_api"`
}

type masterdataSettingsRequest struct {
	Region            string `json:"region"`
	Source            string `json:"source"`
	CustomURL         string `json:"custom_url"`
	CustomFallbackURL string `json:"custom_fallback_url"`
	LocalPath         string `json:"local_path"`
	RefreshInterval   int    `json:"refresh_interval"`
}

type assetsSettingsRequest struct {
	Region        string `json:"region"`
	Source        string `json:"source"`
	Mirror        string `json:"mirror"`
	CustomBaseURL string `json:"custom_base_url"`
	MusicAliasURL string `json:"music_alias_url"`
	StickerPath   string `json:"sticker_path"`
}

type sekaiAPISettingsRequest struct {
	Enabled   *bool  `json:"enabled"`
	Region    string `json:"region"`
	Timeout   int    `json:"timeout"`
	RateLimit int    `json:"rate_limit"`
}

type suiteAPISettingsRequest struct {
	Enabled     *bool  `json:"enabled"`
	URL         string `json:"url"`
	Token       string `json:"token"`
	Timeout     int    `json:"timeout"`
	DefaultMode string `json:"default_mode"`
}

type rankingAPISettingsRequest struct {
	Region  string `json:"region"`
	Timeout int    `json:"timeout"`
}

// handleUpdatePublicConfig saves the editable non-sensitive settings.
func (s *Server) handleUpdatePublicConfig(c *fiber.Ctx) error {
	var req updatePublicConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid settings payload")
	}
	if s.ConfigPath == "" {
		return fiber.NewError(fiber.StatusInternalServerError, "Config path is not configured")
	}

	next := *s.Config
	if req.Server != nil {
		region := config.NormalizeRegion(req.Server.Region)
		if region == "" || !config.IsValidRegion(region) {
			return fiber.NewError(fiber.StatusBadRequest, "Unsupported server region")
		}
		next.Server.Region = region
		if req.SyncClientRegions {
			next.SekaiAPI.Region = region
			next.RankingAPI.Region = region
		}
	}
	if next.Server.Region == "" {
		next.Server.Region = config.RegionJP
	}

	if len(req.Servers) > 0 {
		if next.GameServers == nil {
			next.GameServers = config.DefaultGameServerProfiles()
		}
		for rawRegion, serverReq := range req.Servers {
			region := config.NormalizeRegion(rawRegion)
			if region == "" || !config.IsValidRegion(region) {
				return fiber.NewError(fiber.StatusBadRequest, "Unsupported server region")
			}
			profile := config.ResolveGameServerProfile(&next, region)
			if serverReq.Enabled != nil {
				enabled := *serverReq.Enabled
				if region == config.RegionJP || region == next.Server.Region {
					enabled = true
				}
				profile.Enabled = config.EnabledPtr(enabled)
			}
			if serverReq.Masterdata != nil {
				applyMasterdataSettings(&profile.Masterdata, serverReq.Masterdata)
			}
			masterResolved, err := config.ResolveMasterdata(profile.Masterdata, region)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}
			profile.Masterdata.Region = masterResolved.Region
			profile.Masterdata.Source = masterResolved.Source
			profile.Masterdata.URL = masterResolved.URL
			profile.Masterdata.FallbackURL = masterResolved.FallbackURL
			profile.Masterdata.CustomURL = masterResolved.CustomURL
			profile.Masterdata.CustomFallbackURL = masterResolved.CustomFallbackURL
			profile.Masterdata.LocalPath = masterResolved.LocalPath
			profile.Masterdata.RefreshInterval = masterResolved.RefreshInterval
			if serverReq.Assets != nil {
				applyAssetsSettings(&profile.Assets, serverReq.Assets)
			}
			assetResolved, err := config.ResolveAssets(profile.Assets, region)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}
			profile.Assets.Region = assetResolved.Region
			profile.Assets.Source = assetResolved.Source
			profile.Assets.Mirror = assetResolved.Mirror
			profile.Assets.CDNSource = assetResolved.CDNSource
			profile.Assets.BaseURL = assetResolved.BaseURL
			profile.Assets.CustomBaseURL = assetResolved.CustomBaseURL
			if serverReq.SekaiAPI != nil {
				applySekaiAPISettings(&profile.SekaiAPI, serverReq.SekaiAPI)
			}
			if serverReq.SuiteAPI != nil {
				applySuiteAPISettings(&profile.SuiteAPI, serverReq.SuiteAPI)
				if region == next.Server.Region {
					applySuiteAPISettings(&next.SuiteAPI, serverReq.SuiteAPI)
				}
			}
			if serverReq.RankingAPI != nil {
				applyRankingAPISettings(&profile.RankingAPI, serverReq.RankingAPI)
			}
			next.GameServers[region] = profile
		}
	}

	if req.Masterdata != nil {
		if req.Masterdata.Region != "" {
			next.Masterdata.Region = config.NormalizeRegion(req.Masterdata.Region)
		}
		if req.Masterdata.Source != "" {
			next.Masterdata.Source = config.NormalizeMasterdataSource(req.Masterdata.Source)
		}
		next.Masterdata.CustomURL = strings.TrimSpace(req.Masterdata.CustomURL)
		next.Masterdata.CustomFallbackURL = strings.TrimSpace(req.Masterdata.CustomFallbackURL)
		if req.Masterdata.LocalPath != "" {
			next.Masterdata.LocalPath = strings.TrimSpace(req.Masterdata.LocalPath)
		}
		if req.Masterdata.RefreshInterval >= 0 {
			next.Masterdata.RefreshInterval = req.Masterdata.RefreshInterval
		}
	}

	masterResolved, err := config.ResolveMasterdata(next.Masterdata, next.Server.Region)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	next.Masterdata.Region = masterResolved.Region
	next.Masterdata.Source = masterResolved.Source
	next.Masterdata.URL = masterResolved.URL
	next.Masterdata.FallbackURL = masterResolved.FallbackURL
	next.Masterdata.CustomURL = masterResolved.CustomURL
	next.Masterdata.CustomFallbackURL = masterResolved.CustomFallbackURL
	next.Masterdata.LocalPath = masterResolved.LocalPath
	next.Masterdata.RefreshInterval = masterResolved.RefreshInterval

	if req.Assets != nil {
		if req.Assets.Region != "" {
			next.Assets.Region = config.NormalizeRegion(req.Assets.Region)
		}
		if req.Assets.Source != "" {
			next.Assets.Source = config.NormalizeAssetSource(req.Assets.Source)
		}
		if req.Assets.Mirror != "" {
			next.Assets.Mirror = config.NormalizeAssetMirror(req.Assets.Mirror)
		}
		next.Assets.CustomBaseURL = strings.TrimSpace(req.Assets.CustomBaseURL)
		next.Assets.MusicAliasURL = strings.TrimSpace(req.Assets.MusicAliasURL)
		if req.Assets.StickerPath != "" {
			next.Assets.StickerPath = strings.TrimSpace(req.Assets.StickerPath)
		}
	}

	assetResolved, err := config.ResolveAssets(next.Assets, next.Server.Region)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	next.Assets.Region = assetResolved.Region
	next.Assets.Source = assetResolved.Source
	next.Assets.Mirror = assetResolved.Mirror
	next.Assets.CDNSource = assetResolved.CDNSource
	next.Assets.BaseURL = assetResolved.BaseURL
	next.Assets.CustomBaseURL = assetResolved.CustomBaseURL

	if req.Renderer != nil {
		if req.Renderer.Precision <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "Renderer precision must be greater than 0")
		}
		next.Renderer.Precision = req.Renderer.Precision
	}

	config.NormalizeConfig(&next)

	if err := config.Save(&next, s.ConfigPath); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	*s.Config = next
	if s.Renderer != nil {
		s.Renderer.SetPrecision(s.Config.Renderer.Precision)
	}
	if s.Servers != nil {
		s.Servers.ApplyConfig(s.Config)
		if runtime := s.Servers.Default(); runtime != nil {
			s.Store = runtime.Store
			s.Loader = runtime.Loader
		}
	} else {
		if s.Loader != nil {
			s.Loader.UpdateConfig(s.Config.Masterdata, s.Config.Server.Region)
		}
		if _, err := assets.Configure(s.Config.Assets, s.Config.Server.Region); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	if req.ReloadMasterdata {
		if s.Servers != nil {
			if _, err := s.Servers.Reload(next.Server.Region); err != nil {
				return fiber.NewError(fiber.StatusBadGateway, err.Error())
			}
		} else if s.Loader != nil {
			if err := s.Loader.LoadAll(); err != nil {
				return fiber.NewError(fiber.StatusBadGateway, err.Error())
			}
		}
	}

	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "设置已保存",
		"config":  s.publicConfigMap(),
	})
}

func applyMasterdataSettings(target *config.MasterdataConfig, req *masterdataSettingsRequest) {
	if req == nil || target == nil {
		return
	}
	if req.Region != "" {
		target.Region = config.NormalizeRegion(req.Region)
	}
	if req.Source != "" {
		target.Source = config.NormalizeMasterdataSource(req.Source)
	}
	target.CustomURL = strings.TrimSpace(req.CustomURL)
	target.CustomFallbackURL = strings.TrimSpace(req.CustomFallbackURL)
	if req.LocalPath != "" {
		target.LocalPath = strings.TrimSpace(req.LocalPath)
	}
	if req.RefreshInterval >= 0 {
		target.RefreshInterval = req.RefreshInterval
	}
}

func applyAssetsSettings(target *config.AssetsConfig, req *assetsSettingsRequest) {
	if req == nil || target == nil {
		return
	}
	if req.Region != "" {
		target.Region = config.NormalizeRegion(req.Region)
	}
	if req.Source != "" {
		target.Source = config.NormalizeAssetSource(req.Source)
	}
	if req.Mirror != "" {
		target.Mirror = config.NormalizeAssetMirror(req.Mirror)
	}
	target.CustomBaseURL = strings.TrimSpace(req.CustomBaseURL)
	target.MusicAliasURL = strings.TrimSpace(req.MusicAliasURL)
	if req.StickerPath != "" {
		target.StickerPath = strings.TrimSpace(req.StickerPath)
	}
}

func applySekaiAPISettings(target *config.SekaiAPIConfig, req *sekaiAPISettingsRequest) {
	if req == nil || target == nil {
		return
	}
	if req.Enabled != nil {
		target.Enabled = *req.Enabled
	}
	if req.Region != "" {
		target.Region = config.NormalizeRegion(req.Region)
	}
	if req.Timeout > 0 {
		target.Timeout = req.Timeout
	}
	if req.RateLimit > 0 {
		target.RateLimit = req.RateLimit
	}
}

func applySuiteAPISettings(target *config.SuiteAPIConfig, req *suiteAPISettingsRequest) {
	if req == nil || target == nil {
		return
	}
	if req.Enabled != nil {
		target.Enabled = *req.Enabled
		target.EnabledSet = true
	}
	if strings.TrimSpace(req.URL) != "" {
		target.URL = strings.TrimSpace(req.URL)
	}
	if req.Token != "" {
		target.Token = req.Token
	}
	if req.Timeout > 0 {
		target.Timeout = req.Timeout
	}
	if req.DefaultMode != "" {
		mode := config.NormalizeSuiteMode(req.DefaultMode)
		if config.IsValidSuiteMode(mode) {
			target.DefaultMode = mode
		}
	}
}

func applyRankingAPISettings(target *config.RankingAPIConfig, req *rankingAPISettingsRequest) {
	if req == nil || target == nil {
		return
	}
	if req.Region != "" {
		target.Region = config.NormalizeRegion(req.Region)
	}
	if req.Timeout > 0 {
		target.Timeout = req.Timeout
	}
}

// handleReloadMasterdata forces a full reload from the currently configured source.
func (s *Server) handleReloadMasterdata(c *fiber.Ctx) error {
	region := config.NormalizeRegion(c.Query("region"))
	started := time.Now()
	if s.Servers != nil {
		runtime, err := s.Servers.Reload(region)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		return c.JSON(fiber.Map{
			"ok":          true,
			"message":     fmt.Sprintf("%s Masterdata 已重新加载", runtime.Label),
			"region":      runtime.Region,
			"duration_ms": time.Since(started).Milliseconds(),
			"loaded_at":   nullableTime(runtime.Store.LoadedAt()),
			"counts":      s.masterdataSummaryMapForStore(runtime.Store),
		})
	}
	if s.Loader == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Masterdata loader is not configured")
	}
	if err := s.Loader.LoadAll(); err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(fiber.Map{
		"ok":          true,
		"message":     "Masterdata 已重新加载",
		"duration_ms": time.Since(started).Milliseconds(),
		"loaded_at":   nullableTime(s.defaultStore().LoadedAt()),
		"counts":      s.masterdataSummaryMap(),
	})
}

func (s *Server) publicConfigMap() fiber.Map {
	serverRegion := config.NormalizeRegion(s.Config.Server.Region)
	if serverRegion == "" {
		serverRegion = config.RegionJP
	}
	masterResolved, masterErr := config.ResolveMasterdata(s.Config.Masterdata, serverRegion)
	assetResolved, assetErr := config.ResolveAssets(s.Config.Assets, serverRegion)

	masterMap := fiber.Map{
		"region":                  fallbackString(masterResolved.Region, config.NormalizeRegion(s.Config.Masterdata.Region), serverRegion),
		"region_label":            config.RegionLabel(fallbackString(masterResolved.Region, config.NormalizeRegion(s.Config.Masterdata.Region), serverRegion)),
		"source":                  fallbackString(masterResolved.Source, config.NormalizeMasterdataSource(s.Config.Masterdata.Source)),
		"source_label":            masterResolved.SourceLabel,
		"url":                     fallbackString(masterResolved.URL, s.Config.Masterdata.URL),
		"fallback_url":            fallbackString(masterResolved.FallbackURL, s.Config.Masterdata.FallbackURL),
		"custom_url":              fallbackString(masterResolved.CustomURL, s.Config.Masterdata.CustomURL),
		"custom_fallback_url":     fallbackString(masterResolved.CustomFallbackURL, s.Config.Masterdata.CustomFallbackURL),
		"url_configured":          fallbackString(masterResolved.URL, s.Config.Masterdata.URL) != "",
		"fallback_url_configured": fallbackString(masterResolved.FallbackURL, s.Config.Masterdata.FallbackURL) != "",
		"local_path":              fallbackString(masterResolved.LocalPath, s.Config.Masterdata.LocalPath),
		"refresh_interval":        firstNonZero(masterResolved.RefreshInterval, s.Config.Masterdata.RefreshInterval),
		"endpoints":               masterResolved.Endpoints,
		"supported":               masterErr == nil,
	}
	if masterErr != nil {
		masterMap["error"] = masterErr.Error()
	}

	assetMap := fiber.Map{
		"region":                 fallbackString(assetResolved.Region, config.NormalizeRegion(s.Config.Assets.Region), serverRegion),
		"region_label":           config.RegionLabel(fallbackString(assetResolved.Region, config.NormalizeRegion(s.Config.Assets.Region), serverRegion)),
		"source":                 fallbackString(assetResolved.Source, config.NormalizeAssetSource(s.Config.Assets.Source)),
		"source_label":           assetResolved.SourceLabel,
		"mirror":                 fallbackString(assetResolved.Mirror, config.NormalizeAssetMirror(s.Config.Assets.Mirror)),
		"mirror_label":           assetResolved.MirrorLabel,
		"cdn_source":             fallbackString(assetResolved.CDNSource, s.Config.Assets.CDNSource),
		"base_url":               fallbackString(assetResolved.BaseURL, s.Config.Assets.BaseURL),
		"custom_base_url":        fallbackString(assetResolved.CustomBaseURL, s.Config.Assets.CustomBaseURL),
		"renderer_source":        assetResolved.RendererKey,
		"music_alias_url":        s.Config.Assets.MusicAliasURL,
		"music_alias_configured": s.Config.Assets.MusicAliasURL != "",
		"sticker_path":           s.Config.Assets.StickerPath,
		"supported":              assetErr == nil,
	}
	if assetErr != nil {
		assetMap["error"] = assetErr.Error()
	}

	return fiber.Map{
		"version": appVersion,
		"server": fiber.Map{
			"region": serverRegion,
			"label":  config.RegionLabel(serverRegion),
		},
		"servers": s.publicServerProfilesMap(serverRegion),
		"presets": fiber.Map{
			"regions":            config.RegionOptions(),
			"masterdata_sources": config.MasterdataSourceOptions(),
			"asset_sources":      config.AssetSourceOptions(),
			"asset_mirrors":      config.AssetMirrorOptions(),
		},
		"web": fiber.Map{
			"host": s.Config.Web.Host,
			"port": s.Config.Web.Port,
		},
		"bot": fiber.Map{
			"nickname":        s.Config.Bot.Nickname,
			"command_prefix":  s.Config.Bot.CommandPrefix,
			"command_aliases": s.Config.Bot.CommandAliases,
			"driver_type":     s.Config.Bot.Driver.Type,
			"listen":          s.Config.Bot.Driver.Listen,
			"url_configured":  s.Config.Bot.Driver.URL != "",
			"token_set":       s.Config.Bot.Driver.Token != "",
		},
		"masterdata": masterMap,
		"sekai_api": fiber.Map{
			"enabled":             s.Config.SekaiAPI.Enabled,
			"base_url_configured": s.Config.SekaiAPI.BaseURL != "",
			"region":              s.Config.SekaiAPI.Region,
			"headers_configured":  len(s.Config.SekaiAPI.Headers) > 0,
		},
		"suite_api": publicSuiteAPIMap(s.Config.SuiteAPI),
		"ranking_api": fiber.Map{
			"base_url_configured": s.Config.RankingAPI.BaseURL != "",
			"region":              s.Config.RankingAPI.Region,
			"timeout":             s.Config.RankingAPI.Timeout,
		},
		"renderer": fiber.Map{
			"base_url":  rendererBaseURL(s.Renderer),
			"host":      s.Config.Renderer.Host,
			"port":      s.Config.Renderer.Port,
			"precision": s.Config.Renderer.Precision,
			"cache": fiber.Map{
				"enabled":     s.Config.Renderer.Cache.Enabled,
				"path":        s.Config.Renderer.Cache.Path,
				"max_size_mb": s.Config.Renderer.Cache.MaxSizeMB,
				"ttl_hours":   s.Config.Renderer.Cache.TTLHours,
			},
		},
		"assets": assetMap,
	}
}

func (s *Server) publicServerProfilesMap(defaultRegion string) fiber.Map {
	profiles := fiber.Map{}
	if s.Config == nil {
		return profiles
	}
	for _, region := range config.RegionKeys() {
		profile := config.ResolveGameServerProfile(s.Config, region)
		enabled := config.IsEnabled(profile.Enabled)
		if region == defaultRegion || region == config.RegionJP {
			enabled = true
		}
		masterResolved, masterErr := config.ResolveMasterdata(profile.Masterdata, region)
		assetResolved, assetErr := config.ResolveAssets(profile.Assets, region)
		var store interface {
			IsLoaded() bool
			LoadedAt() time.Time
			CardCount() int
			MusicCount() int
			EventCount() int
			GachaCount() int
			VirtualLiveCount() int
		}
		var loadError string
		if s.Servers != nil {
			if runtime := s.Servers.Get(region); runtime != nil {
				store = runtime.Store
				if runtime.Region != region && !enabled {
					store = nil
				}
				if runtime.Region == region && runtime.LoadError != nil {
					loadError = runtime.LoadError.Error()
				}
			}
		}
		loadedAt := time.Time{}
		loaded := false
		if store != nil {
			loaded = store.IsLoaded()
			loadedAt = store.LoadedAt()
		}
		masterMap := fiber.Map{
			"region":                  fallbackString(masterResolved.Region, profile.Masterdata.Region, region),
			"region_label":            config.RegionLabel(fallbackString(masterResolved.Region, profile.Masterdata.Region, region)),
			"source":                  fallbackString(masterResolved.Source, profile.Masterdata.Source),
			"source_label":            masterResolved.SourceLabel,
			"url":                     fallbackString(masterResolved.URL, profile.Masterdata.URL),
			"fallback_url":            fallbackString(masterResolved.FallbackURL, profile.Masterdata.FallbackURL),
			"custom_url":              fallbackString(masterResolved.CustomURL, profile.Masterdata.CustomURL),
			"custom_fallback_url":     fallbackString(masterResolved.CustomFallbackURL, profile.Masterdata.CustomFallbackURL),
			"url_configured":          fallbackString(masterResolved.URL, profile.Masterdata.URL) != "",
			"fallback_url_configured": fallbackString(masterResolved.FallbackURL, profile.Masterdata.FallbackURL) != "",
			"local_path":              fallbackString(masterResolved.LocalPath, profile.Masterdata.LocalPath),
			"refresh_interval":        firstNonZero(masterResolved.RefreshInterval, profile.Masterdata.RefreshInterval),
			"endpoints":               masterResolved.Endpoints,
			"supported":               masterErr == nil,
		}
		if masterErr != nil {
			masterMap["error"] = masterErr.Error()
		}
		if loadError != "" {
			masterMap["load_error"] = loadError
		}
		assetMap := fiber.Map{
			"region":                 fallbackString(assetResolved.Region, profile.Assets.Region, region),
			"region_label":           config.RegionLabel(fallbackString(assetResolved.Region, profile.Assets.Region, region)),
			"source":                 fallbackString(assetResolved.Source, profile.Assets.Source),
			"source_label":           assetResolved.SourceLabel,
			"mirror":                 assetResolved.Mirror,
			"mirror_label":           assetResolved.MirrorLabel,
			"cdn_source":             fallbackString(assetResolved.CDNSource, profile.Assets.CDNSource),
			"base_url":               fallbackString(assetResolved.BaseURL, profile.Assets.BaseURL),
			"custom_base_url":        fallbackString(assetResolved.CustomBaseURL, profile.Assets.CustomBaseURL),
			"renderer_source":        assetResolved.RendererKey,
			"music_alias_url":        profile.Assets.MusicAliasURL,
			"music_alias_configured": profile.Assets.MusicAliasURL != "",
			"sticker_path":           profile.Assets.StickerPath,
			"supported":              assetErr == nil,
		}
		if assetErr != nil {
			assetMap["error"] = assetErr.Error()
		}
		profiles[region] = fiber.Map{
			"region":      region,
			"label":       config.RegionLabel(region),
			"enabled":     enabled,
			"is_default":  region == defaultRegion,
			"loaded":      loaded,
			"loaded_at":   nullableTime(loadedAt),
			"counts":      s.masterdataSummaryMapForStore(store),
			"masterdata":  masterMap,
			"assets":      assetMap,
			"sekai_api":   publicSekaiAPIMap(profile.SekaiAPI),
			"suite_api":   publicSuiteAPIMap(profile.SuiteAPI),
			"ranking_api": publicRankingAPIMap(profile.RankingAPI),
		}
	}
	return profiles
}

func publicSekaiAPIMap(cfg config.SekaiAPIConfig) fiber.Map {
	return fiber.Map{
		"enabled":             cfg.Enabled,
		"base_url_configured": strings.TrimSpace(cfg.BaseURL) != "",
		"region":              config.NormalizeRegion(cfg.Region),
		"headers_configured":  len(cfg.Headers) > 0,
		"timeout":             cfg.Timeout,
		"rate_limit":          cfg.RateLimit,
	}
}

func publicSuiteAPIMap(cfg config.SuiteAPIConfig) fiber.Map {
	return fiber.Map{
		"enabled":        cfg.Enabled,
		"url_configured": strings.TrimSpace(cfg.URL) != "",
		"token_set":      strings.TrimSpace(cfg.Token) != "",
		"timeout":        cfg.Timeout,
		"default_mode":   config.NormalizeSuiteMode(cfg.DefaultMode),
	}
}

func publicRankingAPIMap(cfg config.RankingAPIConfig) fiber.Map {
	return fiber.Map{
		"base_url_configured": strings.TrimSpace(cfg.BaseURL) != "",
		"region":              config.NormalizeRegion(cfg.Region),
		"timeout":             cfg.Timeout,
	}
}

// handleListGroups returns paginated group list.
func (s *Server) handleListGroups(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	groups, total, err := s.DB.ListGroups(offset, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list groups")
	}

	return c.JSON(fiber.Map{
		"data":  groups,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// handleUpdateGroup updates a group's configuration.
func (s *Server) handleUpdateGroup(c *fiber.Ctx) error {
	// TODO: implement group update
	return c.JSON(fiber.Map{"message": "not implemented"})
}

// handleListUsers returns paginated user list.
func (s *Server) handleListUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	users, total, err := s.DB.ListUsers(offset, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list users")
	}

	return c.JSON(fiber.Map{
		"data":  users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// handleDeleteUser deletes a user by ID.
func (s *Server) handleDeleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	if err := s.DB.DeleteUser(uint(id)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete user")
	}

	return c.JSON(fiber.Map{"message": "deleted"})
}

// handleCommandStats returns command usage statistics.
func (s *Server) handleCommandStats(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days", "7"))
	if days <= 0 {
		days = 7
	}
	if days > 365 {
		days = 365
	}
	since := time.Now().AddDate(0, 0, -days)

	stats, err := s.DB.GetCommandStats(since)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get stats")
	}

	return c.JSON(fiber.Map{
		"data":  stats,
		"since": since,
	})
}

// handleSearchCards searches card masterdata and returns lightweight rows.
func (s *Server) handleSearchCards(c *fiber.Ctx) error {
	if err := s.ensureSearchReady(c); err != nil {
		return err
	}
	q := strings.TrimSpace(c.Query("q"))
	results := s.defaultStore().SearchCards(q)
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

// handleSearchMusics searches music masterdata and returns lightweight rows.
func (s *Server) handleSearchMusics(c *fiber.Ctx) error {
	if err := s.ensureSearchReady(c); err != nil {
		return err
	}
	q := strings.TrimSpace(c.Query("q"))
	results := s.defaultStore().SearchMusics(q)
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

// handleSearchEvents searches event masterdata and returns lightweight rows.
func (s *Server) handleSearchEvents(c *fiber.Ctx) error {
	if err := s.ensureSearchReady(c); err != nil {
		return err
	}
	q := strings.TrimSpace(c.Query("q"))
	results := s.defaultStore().SearchEvents(q)
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

// handleSearchGachas searches gacha masterdata and returns lightweight rows.
func (s *Server) handleSearchGachas(c *fiber.Ctx) error {
	c.Locals("allow_empty_query", true)
	if err := s.ensureSearchReady(c); err != nil {
		return err
	}
	q := strings.TrimSpace(c.Query("q"))
	results := s.defaultStore().SearchGachas(q)
	if q == "" || strings.EqualFold(q, "当前") {
		store := s.defaultStore()
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

// handleSearchVirtualLives searches virtual live masterdata and returns lightweight rows.
func (s *Server) handleSearchVirtualLives(c *fiber.Ctx) error {
	c.Locals("allow_empty_query", true)
	if err := s.ensureSearchReady(c); err != nil {
		return err
	}
	q := strings.TrimSpace(c.Query("q"))
	results := searchVirtualLivesForWeb(s.defaultStore().AllVirtualLives(), q)
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

func (s *Server) masterdataSummaryMap() fiber.Map {
	return s.masterdataSummaryMapForStore(s.defaultStore())
}

func (s *Server) masterdataSummaryMapForStore(store interface {
	CardCount() int
	MusicCount() int
	EventCount() int
	GachaCount() int
	VirtualLiveCount() int
}) fiber.Map {
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

func (s *Server) checkRenderer() (bool, int, error, time.Duration) {
	if s.Renderer == nil {
		return false, 0, fmt.Errorf("renderer client is not configured"), 0
	}
	started := time.Now()
	ok, statusCode, err := s.Renderer.HealthWithTimeout(1200 * time.Millisecond)
	return ok, statusCode, err, time.Since(started)
}

func (s *Server) defaultStore() interface {
	IsLoaded() bool
	LoadedAt() time.Time
	CardCount() int
	MusicCount() int
	EventCount() int
	GachaCount() int
	VirtualLiveCount() int
	SearchCards(string) []masterdata.CardInfo
	SearchMusics(string) []masterdata.MusicInfo
	SearchEvents(string) []masterdata.EventInfo
	SearchGachas(string) []masterdata.GachaInfo
	AllVirtualLives() []masterdata.VirtualLive
} {
	if s.Servers != nil {
		if runtime := s.Servers.Default(); runtime != nil && runtime.Store != nil {
			return runtime.Store
		}
	}
	return s.Store
}

func (s *Server) ensureSearchReady(c *fiber.Ctx) error {
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
	store := s.defaultStore()
	if store == nil || !store.IsLoaded() {
		return c.JSON(fiber.Map{
			"data":    []fiber.Map{},
			"total":   0,
			"query":   q,
			"message": "Masterdata 尚未加载，暂时无法搜索。",
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

func rendererBaseURL(client *renderer.Client) string {
	if client == nil {
		return ""
	}
	return client.BaseURL()
}

func rendererMessage(ok bool, err error) string {
	if ok {
		return "Renderer 渲染服务可用"
	}
	if err != nil {
		return err.Error()
	}
	return "Renderer 渲染服务不可用"
}

func masterdataMessage(loaded bool) string {
	if loaded {
		return "Masterdata 已加载"
	}
	return "Masterdata 未加载或为空"
}

func statusText(ok bool) string {
	if ok {
		return "ok"
	}
	return "error"
}

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func setOptionalHeader(c *fiber.Ctx, key string, value string) {
	if value != "" {
		c.Set(key, value)
	}
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

func fallbackString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

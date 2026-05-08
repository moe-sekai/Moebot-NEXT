package web

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/plugins/moesekai/b30"
	"moebot-next/internal/config"
	"moebot-next/internal/database"
	"moebot-next/internal/plugins/moesekai/masterdata"
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
			"status":          statusText(rendererOK),
			"ok":              rendererOK,
			"message":         rendererMessage(rendererOK, rendererErr),
			"base_url":        rendererBaseURL(s.Renderer),
			"status_code":     rendererCode,
			"latency_ms":      rendererLatency.Milliseconds(),
			"service_port":    s.Config.Renderer.Port,
			"dashboard_port":  s.Config.Web.Port,
			"precision":       s.Config.Renderer.Precision,
			"chart_precision": s.Config.Renderer.ChartPrecision,
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

// handleRendererHealth returns renderer reachability and public renderer info.
func (s *Server) handleRendererHealth(c *fiber.Ctx) error {
	ok, statusCode, err, latency := s.checkRenderer()
	return c.JSON(fiber.Map{
		"ok":              ok,
		"status":          statusText(ok),
		"message":         rendererMessage(ok, err),
		"base_url":        rendererBaseURL(s.Renderer),
		"status_code":     statusCode,
		"latency_ms":      latency.Milliseconds(),
		"renderer_port":   s.Config.Renderer.Port,
		"dashboard_port":  s.Config.Web.Port,
		"precision":       s.Config.Renderer.Precision,
		"chart_precision": s.Config.Renderer.ChartPrecision,
		"note":            fmt.Sprintf("%d 是 Satori/Bun 图片渲染服务，%d 才是 Moebot NEXT 管理面板。", s.Config.Renderer.Port, s.Config.Web.Port),
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
	Bot               *botSettingsRequest                  `json:"bot"`
	Renderer          *rendererSettingsRequest             `json:"renderer"`
	ReloadMasterdata  bool                                 `json:"reload_masterdata"`
	SyncClientRegions bool                                 `json:"sync_client_regions"`
}

type botSettingsRequest struct {
	Nickname      *[]string                 `json:"nickname"`
	CommandPrefix *string                   `json:"command_prefix"`
	Driver        *botDriverSettingsRequest `json:"driver"`
}

type botDriverSettingsRequest struct {
	Type   *string `json:"type"`
	Listen *string `json:"listen"`
	URL    *string `json:"url"`
	Token  *string `json:"token"`
}

type serverSettingsRequest struct {
	Region string `json:"region"`
}

type rendererSettingsRequest struct {
	Precision      *float64 `json:"precision"`
	ChartPrecision *float64 `json:"chart_precision"`
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
	Region         string `json:"region"`
	Source         string `json:"source"`
	Mirror         string `json:"mirror"`
	CustomBaseURL  string `json:"custom_base_url"`
	MusicAliasURL  string `json:"music_alias_url"`
	ChartSourceURL string `json:"chart_source_url"`
	StickerPath    string `json:"sticker_path"`
}

type sekaiAPISettingsRequest struct {
	Enabled   *bool              `json:"enabled"`
	BaseURL   string             `json:"base_url"`
	Region    string             `json:"region"`
	Headers   *map[string]string `json:"headers"`
	Timeout   int                `json:"timeout"`
	RateLimit int                `json:"rate_limit"`
}

type suiteAPISettingsRequest struct {
	Enabled     *bool              `json:"enabled"`
	URL         string             `json:"url"`
	Headers     *map[string]string `json:"headers"`
	Timeout     int                `json:"timeout"`
	DefaultMode string             `json:"default_mode"`
}

type rankingAPISettingsRequest struct {
	Timeout int `json:"timeout"`
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
				if region == next.Server.Region {
					applySekaiAPISettings(&next.SekaiAPI, serverReq.SekaiAPI)
				}
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
		next.Assets.ChartSourceURL = strings.TrimSpace(req.Assets.ChartSourceURL)
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

	if req.Bot != nil {
		if req.Bot.Nickname != nil {
			cleaned := make([]string, 0, len(*req.Bot.Nickname))
			for _, name := range *req.Bot.Nickname {
				if v := strings.TrimSpace(name); v != "" {
					cleaned = append(cleaned, v)
				}
			}
			next.Bot.Nickname = cleaned
		}
		if req.Bot.CommandPrefix != nil {
			prefix := strings.TrimSpace(*req.Bot.CommandPrefix)
			if prefix == "" {
				return fiber.NewError(fiber.StatusBadRequest, "Command prefix must not be empty")
			}
			next.Bot.CommandPrefix = prefix
		}
		if req.Bot.Driver != nil {
			if req.Bot.Driver.Type != nil {
				driverType := strings.TrimSpace(*req.Bot.Driver.Type)
				switch driverType {
				case "ws", "ws-reverse":
					next.Bot.Driver.Type = driverType
				default:
					return fiber.NewError(fiber.StatusBadRequest, "Driver type must be ws or ws-reverse")
				}
			}
			if req.Bot.Driver.Listen != nil {
				next.Bot.Driver.Listen = strings.TrimSpace(*req.Bot.Driver.Listen)
			}
			if req.Bot.Driver.URL != nil {
				next.Bot.Driver.URL = strings.TrimSpace(*req.Bot.Driver.URL)
			}
			if req.Bot.Driver.Token != nil {
				next.Bot.Driver.Token = strings.TrimSpace(*req.Bot.Driver.Token)
			}
		}
	}

	if req.Renderer != nil {
		if req.Renderer.Precision != nil {
			if *req.Renderer.Precision <= 0 {
				return fiber.NewError(fiber.StatusBadRequest, "Renderer precision must be greater than 0")
			}
			next.Renderer.Precision = *req.Renderer.Precision
		}
		if req.Renderer.ChartPrecision != nil {
			if *req.Renderer.ChartPrecision <= 0 {
				return fiber.NewError(fiber.StatusBadRequest, "Chart renderer precision must be greater than 0")
			}
			next.Renderer.ChartPrecision = *req.Renderer.ChartPrecision
		}
	}

	config.NormalizeConfig(&next)

	if err := config.Save(&next, s.ConfigPath); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	*s.Config = next
	s.B30 = b30.NewClient(s.Config.B30)
	if s.Renderer != nil {
		s.Renderer.SetPrecision(s.Config.Renderer.Precision)
		s.Renderer.SetChartPrecision(s.Config.Renderer.ChartPrecision)
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
	target.ChartSourceURL = strings.TrimSpace(req.ChartSourceURL)
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
	target.BaseURL = fallbackString(req.BaseURL, target.BaseURL, config.DefaultSekaiAPIURL)
	if req.Region != "" {
		target.Region = config.NormalizeRegion(req.Region)
	}
	if req.Headers != nil {
		headers := make(map[string]string, len(*req.Headers))
		for key, value := range *req.Headers {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if key != "" && value != "" {
				headers[key] = value
			}
		}
		target.Headers = headers
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
	if req.Headers != nil {
		headers := make(map[string]string, len(*req.Headers))
		for key, value := range *req.Headers {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if key != "" && value != "" {
				headers[key] = value
			}
		}
		target.Headers = headers
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
	if req.Timeout > 0 {
		target.Timeout = req.Timeout
	}
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
		"chart_source_url":       fallbackString(s.Config.Assets.ChartSourceURL, config.DefaultChartSourceURL),
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
			"url":             s.Config.Bot.Driver.URL,
			"url_configured":  s.Config.Bot.Driver.URL != "",
			"token_set":       s.Config.Bot.Driver.Token != "",
		},
		"masterdata":  masterMap,
		"sekai_api":   publicSekaiAPIMap(s.Config.SekaiAPI),
		"suite_api":   publicSuiteAPIMap(s.Config.SuiteAPI),
		"ranking_api": publicRankingAPIMap(s.Config.RankingAPI),
		"renderer": fiber.Map{
			"base_url":        rendererBaseURL(s.Renderer),
			"host":            s.Config.Renderer.Host,
			"port":            s.Config.Renderer.Port,
			"precision":       s.Config.Renderer.Precision,
			"chart_precision": s.Config.Renderer.ChartPrecision,
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
			"chart_source_url":       fallbackString(profile.Assets.ChartSourceURL, config.DefaultChartSourceURL),
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
		"base_url":            fallbackString(cfg.BaseURL, config.DefaultSekaiAPIURL),
		"base_url_configured": strings.TrimSpace(cfg.BaseURL) != "",
		"region":              config.NormalizeRegion(cfg.Region),
		"headers":             copyHeadersMap(cfg.Headers),
		"headers_configured":  len(cfg.Headers) > 0,
		"timeout":             cfg.Timeout,
		"rate_limit":          cfg.RateLimit,
	}
}

func publicSuiteAPIMap(cfg config.SuiteAPIConfig) fiber.Map {
	return fiber.Map{
		"enabled":            cfg.Enabled,
		"url":                fallbackString(cfg.URL, config.DefaultSuiteAPIURL),
		"url_configured":     strings.TrimSpace(cfg.URL) != "",
		"headers":            copyHeadersMap(cfg.Headers),
		"headers_configured": len(cfg.Headers) > 0,
		"timeout":            cfg.Timeout,
		"default_mode":       config.NormalizeSuiteMode(cfg.DefaultMode),
	}
}

func publicRankingAPIMap(cfg config.RankingAPIConfig) fiber.Map {
	return fiber.Map{
		"base_url_configured": strings.TrimSpace(cfg.BaseURL) != "",
		"region":              config.NormalizeRegion(cfg.Region),
		"timeout":             cfg.Timeout,
	}
}

func copyHeadersMap(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(headers))
	for key, value := range headers {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			out[key] = value
		}
	}
	return out
}

// handleListGroups returns paginated group list, optionally including
// command statistics for a sliding window (default: 7 days).
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

	statsDays, _ := strconv.Atoi(c.Query("stats_days", "7"))
	if statsDays <= 0 {
		statsDays = 7
	}
	if statsDays > 90 {
		statsDays = 90
	}
	since := time.Now().AddDate(0, 0, -statsDays)
	statsRows, _ := s.DB.GetGroupCommandStats(since)
	statsMap := make(map[string]database.GroupCommandStat, len(statsRows))
	for _, row := range statsRows {
		statsMap[row.Platform+"|"+row.GroupID] = row
	}

	data := make([]fiber.Map, 0, len(groups))
	for _, g := range groups {
		row := fiber.Map{
			"id":         g.ID,
			"platform":   g.Platform,
			"group_id":   g.GroupID,
			"name":       g.Name,
			"enabled":    g.Enabled,
			"config":     g.Config,
			"created_at": g.CreatedAt,
		}
		if stat, ok := statsMap[g.Platform+"|"+g.GroupID]; ok {
			row["stats"] = fiber.Map{
				"count":     stat.Count,
				"last_used": nullableTime(stat.LastUsed),
				"avg_ms":    stat.AvgMs,
				"days":      statsDays,
			}
		} else {
			row["stats"] = fiber.Map{"count": 0, "last_used": nil, "avg_ms": 0, "days": statsDays}
		}
		data = append(data, row)
	}

	return c.JSON(fiber.Map{
		"data":  data,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

type updateGroupPayload struct {
	Enabled *bool   `json:"enabled"`
	Name    *string `json:"name"`
	Config  *string `json:"config"`
}

// handleUpdateGroup updates a group's configuration.
func (s *Server) handleUpdateGroup(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}

	var payload updateGroupPayload
	if err := c.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	group, err := s.DB.GetGroupByID(uint(id))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Group not found")
	}

	if payload.Enabled != nil {
		group.Enabled = *payload.Enabled
	}
	if payload.Name != nil {
		group.Name = strings.TrimSpace(*payload.Name)
	}
	if payload.Config != nil {
		cfg := strings.TrimSpace(*payload.Config)
		if cfg == "" {
			cfg = "{}"
		}
		// Validate JSON to prevent storing garbage.
		if !json.Valid([]byte(cfg)) {
			return fiber.NewError(fiber.StatusBadRequest, "Config must be valid JSON")
		}
		group.Config = cfg
	}

	if err := s.DB.UpsertGroup(group); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update group")
	}

	return c.JSON(fiber.Map{"data": group, "message": "updated"})
}

// handleDeleteGroup removes a group by primary key.
func (s *Server) handleDeleteGroup(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}
	if err := s.DB.DeleteGroup(uint(id)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete group")
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// handleGroupRecentCommands returns recent command invocations for one group.
func (s *Server) handleGroupRecentCommands(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}
	group, err := s.DB.GetGroupByID(uint(id))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Group not found")
	}
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	rows, err := s.DB.ListGroupRecentCommands(group.Platform, group.GroupID, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to load commands")
	}
	return c.JSON(fiber.Map{"data": rows, "group": group})
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

// handleCommandStats returns command usage statistics for a sliding window.
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
	totals, err := s.DB.GetCommandStatsTotals(since)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get stats totals")
	}
	trend, err := s.DB.GetCommandStatsTrend(since)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get stats trend")
	}
	byPlatform, err := s.DB.GetCommandStatsByPlatform(since)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get stats by platform")
	}

	return c.JSON(fiber.Map{
		"data":        stats,
		"since":       since,
		"days":        days,
		"totals":      totals,
		"trend":       trend,
		"by_platform": byPlatform,
	})
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

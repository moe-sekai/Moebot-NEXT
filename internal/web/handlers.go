package web

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

	masterdataLoaded := false
	masterdataLoadedAt := time.Time{}
	if s.Store != nil {
		masterdataLoaded = s.Store.IsLoaded()
		masterdataLoadedAt = s.Store.LoadedAt()
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
	loaded := s.Store != nil && s.Store.IsLoaded()
	loadedAt := time.Time{}
	if s.Store != nil {
		loadedAt = s.Store.LoadedAt()
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
	setOptionalHeader(c, "x-render-satori-ms", result.SatoriMS)
	setOptionalHeader(c, "x-render-resvg-ms", result.ResvgMS)
	setOptionalHeader(c, "x-render-size-bytes", result.SizeBytes)
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
	return c.JSON(fiber.Map{
		"version": appVersion,
		"web": fiber.Map{
			"host": s.Config.Web.Host,
			"port": s.Config.Web.Port,
		},
		"bot": fiber.Map{
			"nickname":       s.Config.Bot.Nickname,
			"command_prefix": s.Config.Bot.CommandPrefix,
			"driver_type":    s.Config.Bot.Driver.Type,
			"listen":         s.Config.Bot.Driver.Listen,
			"url_configured": s.Config.Bot.Driver.URL != "",
			"token_set":      s.Config.Bot.Driver.Token != "",
		},
		"masterdata": fiber.Map{
			"url_configured":          s.Config.Masterdata.URL != "",
			"fallback_url_configured": s.Config.Masterdata.FallbackURL != "",
			"local_path":              s.Config.Masterdata.LocalPath,
			"refresh_interval":        s.Config.Masterdata.RefreshInterval,
		},
		"renderer": fiber.Map{
			"base_url": rendererBaseURL(s.Renderer),
			"host":     s.Config.Renderer.Host,
			"port":     s.Config.Renderer.Port,
			"cache": fiber.Map{
				"enabled":     s.Config.Renderer.Cache.Enabled,
				"path":        s.Config.Renderer.Cache.Path,
				"max_size_mb": s.Config.Renderer.Cache.MaxSizeMB,
				"ttl_hours":   s.Config.Renderer.Cache.TTLHours,
			},
		},
		"assets": fiber.Map{
			"cdn_source":             s.Config.Assets.CDNSource,
			"music_alias_configured": s.Config.Assets.MusicAliasURL != "",
			"sticker_path":           s.Config.Assets.StickerPath,
		},
	})
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
	results := s.Store.SearchCards(q)
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
	results := s.Store.SearchMusics(q)
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
	results := s.Store.SearchEvents(q)
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
	if err := s.ensureSearchReady(c); err != nil {
		return err
	}
	q := strings.TrimSpace(c.Query("q"))
	results := s.Store.SearchGachas(q)
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

func (s *Server) masterdataSummaryMap() fiber.Map {
	if s.Store == nil {
		return fiber.Map{
			"cards":  0,
			"musics": 0,
			"events": 0,
			"gachas": 0,
		}
	}
	return fiber.Map{
		"cards":  s.Store.CardCount(),
		"musics": s.Store.MusicCount(),
		"events": s.Store.EventCount(),
		"gachas": s.Store.GachaCount(),
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

func (s *Server) ensureSearchReady(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		return c.JSON(fiber.Map{
			"data":    []fiber.Map{},
			"total":   0,
			"query":   q,
			"message": "请输入关键词后再搜索。",
		})
	}
	if s.Store == nil || !s.Store.IsLoaded() {
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

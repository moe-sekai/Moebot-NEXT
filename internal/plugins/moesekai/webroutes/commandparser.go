package webroutes

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/plugins/moesekai/commandparser"

	"github.com/gofiber/fiber/v2"
)

// RegisterCommandParser registers /api/commands/* routes onto the given group.
// The PJSK command parser is moesekai-specific business; this used to live in
// internal/web/command_handlers.go.
func RegisterCommandParser(api fiber.Router, d *Deps) {
	h := &commandParserHandlers{d: d}
	api.Get("/commands/definitions", h.definitions)
	api.Get("/commands/parse", h.parse)
	api.Get("/commands/parse/image", h.parseImage)
	api.Get("/commands/aliases", h.getAliases)
	api.Put("/commands/aliases", h.updateAliases)
	api.Post("/commands/aliases/reset", h.resetAliases)
	api.Get("/commands/aliases/export", h.exportAliases)
	api.Post("/commands/aliases/import", h.importAliases)
}

type commandParserHandlers struct {
	d *Deps
}

func (h *commandParserHandlers) service() *commandparser.Service {
	svc := commandparser.NewService(
		h.d.Config.Bot.CommandPrefix,
		h.d.Config.Bot.CommandAliases,
		h.d.Servers,
		h.d.Store,
		h.d.Renderer,
	)
	svc.B30 = h.d.B30
	return svc
}

func (h *commandParserHandlers) definitions(c *fiber.Ctx) error {
	return c.JSON(h.service().DefinitionsPayload())
}

func (h *commandParserHandlers) parse(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	parsed := h.service().ParseWithOptions(q, commandparser.ParseOptions{DebugBinding: debugBinding(c)})
	return c.JSON(commandparser.ParseResponse{
		OK:      parsed.Definition != nil,
		Parsed:  parsed,
		Message: parsed.Message,
	})
}

func (h *commandParserHandlers) parseImage(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	width, _ := strconv.Atoi(c.Query("width"))
	height, _ := strconv.Atoi(c.Query("height"))
	started := time.Now()
	result, _, err := h.service().RenderWithOptions(q, width, height, commandparser.RenderOptions{DebugBinding: debugBinding(c)})
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

func (h *commandParserHandlers) getAliases(c *fiber.Ctx) error {
	return c.JSON(commandparser.AliasConfig(h.d.Config.Bot.CommandAliases))
}

func (h *commandParserHandlers) updateAliases(c *fiber.Ctx) error {
	var req commandparser.AliasUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid command alias payload")
	}
	return h.saveAliases(c, req.Aliases, "自定义关键词已保存；聊天端通常需要重启后生效。")
}

func (h *commandParserHandlers) resetAliases(c *fiber.Ctx) error {
	return h.saveAliases(c, map[string][]string{}, "已恢复默认关键词；自定义关键词已清空。")
}

func (h *commandParserHandlers) exportAliases(c *fiber.Ctx) error {
	aliases, _, err := commandparser.ValidateAliases(h.d.Config.Bot.CommandAliases)
	if err != nil {
		aliases, _ = commandparser.NormalizeAliases(h.d.Config.Bot.CommandAliases)
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Set("content-disposition", "attachment; filename=moebot-command-aliases.json")
	return c.JSON(commandparser.AliasUpdateRequest{Aliases: aliases})
}

func (h *commandParserHandlers) importAliases(c *fiber.Ctx) error {
	var req commandparser.AliasUpdateRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid alias import JSON")
	}
	return h.saveAliases(c, req.Aliases, "自定义关键词已导入并保存；聊天端通常需要重启后生效。")
}

func (h *commandParserHandlers) saveAliases(c *fiber.Ctx, raw map[string][]string, message string) error {
	aliases, _, err := commandparser.ValidateAliases(raw)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if h.d.SaveConfig == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Config save hook is not wired")
	}
	h.d.Config.Bot.CommandAliases = aliases
	if err := h.d.SaveConfig(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(commandparser.AliasUpdateResponse{
		OK:      true,
		Message: message,
		Aliases: aliases,
		Config:  commandparser.AliasConfig(aliases),
	})
}

func debugBinding(c *fiber.Ctx) commandparser.DebugBinding {
	return commandparser.DebugBinding{
		Region: strings.TrimSpace(c.Query("debug_region")),
		GameID: strings.TrimSpace(c.Query("debug_game_id")),
	}
}

func setOptionalHeader(c *fiber.Ctx, key, value string) {
	if value != "" {
		c.Set(key, value)
	}
}

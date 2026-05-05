package web

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/commandparser"
	"moebot-next/internal/config"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) commandParserService() *commandparser.Service {
	return commandparser.NewService(s.Config.Bot.CommandPrefix, s.Config.Bot.CommandAliases, s.Servers, s.Store, s.Renderer)
}

func (s *Server) handleCommandDefinitions(c *fiber.Ctx) error {
	payload := s.commandParserService().DefinitionsPayload()
	return c.JSON(payload)
}

func (s *Server) handleParseCommand(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	parsed := s.commandParserService().ParseWithOptions(q, commandparser.ParseOptions{DebugBinding: commandDebugBinding(c)})
	return c.JSON(commandparser.ParseResponse{
		OK:      parsed.Definition != nil,
		Parsed:  parsed,
		Message: parsed.Message,
	})
}

func (s *Server) handleParseCommandImage(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	width, _ := strconv.Atoi(c.Query("width"))
	height, _ := strconv.Atoi(c.Query("height"))
	started := time.Now()
	result, _, err := s.commandParserService().RenderWithOptions(q, width, height, commandparser.RenderOptions{DebugBinding: commandDebugBinding(c)})
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

func commandDebugBinding(c *fiber.Ctx) commandparser.DebugBinding {
	return commandparser.DebugBinding{
		Region: strings.TrimSpace(c.Query("debug_region")),
		GameID: strings.TrimSpace(c.Query("debug_game_id")),
	}
}

func (s *Server) handleGetCommandAliases(c *fiber.Ctx) error {
	return c.JSON(commandparser.AliasConfig(s.Config.Bot.CommandAliases))
}

func (s *Server) handleUpdateCommandAliases(c *fiber.Ctx) error {
	var req commandparser.AliasUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid command alias payload")
	}
	return s.saveCommandAliases(c, req.Aliases, "自定义关键词已保存；聊天端通常需要重启后生效。")
}

func (s *Server) handleResetCommandAliases(c *fiber.Ctx) error {
	return s.saveCommandAliases(c, map[string][]string{}, "已恢复默认关键词；自定义关键词已清空。")
}

func (s *Server) handleExportCommandAliases(c *fiber.Ctx) error {
	aliases, _, err := commandparser.ValidateAliases(s.Config.Bot.CommandAliases)
	if err != nil {
		aliases, _ = commandparser.NormalizeAliases(s.Config.Bot.CommandAliases)
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Set("content-disposition", "attachment; filename=moebot-command-aliases.json")
	return c.JSON(commandparser.AliasUpdateRequest{Aliases: aliases})
}

func (s *Server) handleImportCommandAliases(c *fiber.Ctx) error {
	var req commandparser.AliasUpdateRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid alias import JSON")
	}
	return s.saveCommandAliases(c, req.Aliases, "自定义关键词已导入并保存；聊天端通常需要重启后生效。")
}

func (s *Server) saveCommandAliases(c *fiber.Ctx, raw map[string][]string, message string) error {
	aliases, _, err := commandparser.ValidateAliases(raw)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if s.ConfigPath == "" {
		return fiber.NewError(fiber.StatusInternalServerError, "Config path is not configured")
	}
	next := *s.Config
	next.Bot.CommandAliases = aliases
	if err := config.Save(&next, s.ConfigPath); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	*s.Config = next
	return c.JSON(commandparser.AliasUpdateResponse{
		OK:      true,
		Message: message,
		Aliases: aliases,
		Config:  commandparser.AliasConfig(aliases),
	})
}

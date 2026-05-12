package autochat

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"moebot-next/internal/plugin"

	"github.com/gofiber/fiber/v2"
)

// groupSettingPayload 单群配置 DTO。
type groupSettingPayload struct {
	GroupID         int64    `json:"group_id"`
	Persona         string   `json:"persona,omitempty"`           // 覆盖 Chat.Prompt.Persona[gid]
	WillingOverride *float64 `json:"willing_threshold,omitempty"` // 覆盖 Chat.Willing.GroupThresholds[gid]
	Model           string   `json:"model,omitempty"`             // 覆盖 fileDB model_<gid>
	Template        string   `json:"template,omitempty"`          // 绑定的模板名（Chat.GroupTemplates[gid]）
	ChatEnabled     bool     `json:"chat_enabled"`
	AutoEnabled     bool     `json:"auto_enabled"`
}

// registerWebRoutes 在 Fiber 上挂 /api/plugins/autochat/* 路由。
// 仅注册一次（Fiber 不支持取消）；多次 Init 时通过 sync.Once 保护。
func (p *pluginImpl) registerWebRoutes(api fiber.Router) {
	g := api.Group("/plugins/" + PluginName)
	g.Get("/overview", p.handleOverview)
	g.Get("/providers", p.handleGetProviders)
	g.Put("/providers", p.handlePutProviders)
	g.Get("/persona", p.handleGetPersona)
	g.Put("/persona", p.handlePutPersona)
	g.Get("/triggers", p.handleGetTriggers)
	g.Put("/triggers", p.handlePutTriggers)
	g.Get("/yaml", p.handleGetYAML)
	g.Put("/yaml", p.handlePutYAML)
	g.Post("/test-provider", p.handleTestProvider)
	g.Post("/list-models", p.handleListModels)
	g.Get("/groups", p.handleListGroups)
	g.Put("/groups/:gid", p.handleUpsertGroup)
	g.Delete("/groups/:gid", p.handleDeleteGroup)
	g.Get("/templates", p.handleListTemplates)
	g.Put("/templates/:name", p.handleUpsertTemplate)
	g.Delete("/templates/:name", p.handleDeleteTemplate)
	g.Get("/memory/groups", p.handleListMemoryGroups)
	g.Get("/memory", p.handleQueryMemoryItems)
	g.Delete("/memory/:id", p.handleDeleteMemoryItem)
}

// handleListGroups 返回所有“已知群组”的当前配置。
// 已知群组定义为：persona/threshold/whitelist/model 任一存在的群。
func (p *pluginImpl) handleListGroups(c *fiber.Ctx) error {
	cfg := GetConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	ids := map[int64]struct{}{}
	mark := func(s string) {
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			ids[id] = struct{}{}
		}
	}
	for k := range cfg.Chat.Prompt.Persona {
		if k != "default" {
			mark(k)
		}
	}
	for k := range cfg.Chat.Willing.GroupThresholds {
		mark(k)
	}
	for k := range cfg.Chat.GroupTemplates {
		mark(k)
	}
	if p.fileDB != nil {
		for _, k := range p.fileDB.Keys() {
			if strings.HasPrefix(k, "model_") {
				mark(strings.TrimPrefix(k, "model_"))
			}
		}
		for _, gid := range p.fileDB.GetStringSlice("chat_whitelist") {
			mark(gid)
		}
		for _, gid := range p.fileDB.GetStringSlice("autochat_whitelist") {
			mark(gid)
		}
	}
	out := make([]groupSettingPayload, 0, len(ids))
	for id := range ids {
		out = append(out, p.buildGroupPayload(cfg, id))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GroupID < out[j].GroupID })
	return c.JSON(fiber.Map{"groups": out, "default_threshold": cfg.Chat.Willing.Threshold})
}

func (p *pluginImpl) buildGroupPayload(cfg *Config, gid int64) groupSettingPayload {
	gs := strconv.FormatInt(gid, 10)
	out := groupSettingPayload{GroupID: gid}
	if v, ok := cfg.Chat.Prompt.Persona[gs]; ok {
		out.Persona = v
	}
	if v, ok := cfg.Chat.Willing.GroupThresholds[gs]; ok {
		out.WillingOverride = &v
	}
	if p.fileDB != nil {
		out.Model = p.fileDB.GetString("model_" + gs)
	}
	if v, ok := cfg.Chat.GroupTemplates[gs]; ok {
		out.Template = v
	}
	if p.chatWhiteList != nil {
		out.ChatEnabled = p.chatWhiteList.Check(gid)
	}
	if p.autoWhiteList != nil {
		out.AutoEnabled = p.autoWhiteList.Check(gid)
	}
	return out
}

func (p *pluginImpl) handleUpsertGroup(c *fiber.Ctx) error {
	gid, err := strconv.ParseInt(c.Params("gid"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid group id")
	}
	var body struct {
		Persona         *string  `json:"persona,omitempty"`
		WillingOverride *float64 `json:"willing_threshold,omitempty"`
		ClearWilling    bool     `json:"clear_willing,omitempty"`
		Model           *string  `json:"model,omitempty"`
		Template        *string  `json:"template,omitempty"`
		ChatEnabled     *bool    `json:"chat_enabled,omitempty"`
		AutoEnabled     *bool    `json:"auto_enabled,omitempty"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	p.mu.Lock()
	cfg := GetConfig()
	if cfg == nil {
		p.mu.Unlock()
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	gs := strconv.FormatInt(gid, 10)
	if body.Persona != nil {
		if cfg.Chat.Prompt.Persona == nil {
			cfg.Chat.Prompt.Persona = map[string]string{}
		}
		v := strings.TrimSpace(*body.Persona)
		if v == "" {
			delete(cfg.Chat.Prompt.Persona, gs)
		} else {
			cfg.Chat.Prompt.Persona[gs] = v
		}
	}
	if body.ClearWilling {
		delete(cfg.Chat.Willing.GroupThresholds, gs)
	} else if body.WillingOverride != nil {
		if cfg.Chat.Willing.GroupThresholds == nil {
			cfg.Chat.Willing.GroupThresholds = map[string]float64{}
		}
		cfg.Chat.Willing.GroupThresholds[gs] = *body.WillingOverride
	}
	if body.Template != nil {
		if cfg.Chat.GroupTemplates == nil {
			cfg.Chat.GroupTemplates = map[string]string{}
		}
		v := strings.TrimSpace(*body.Template)
		if v == "" {
			delete(cfg.Chat.GroupTemplates, gs)
		} else if _, ok := cfg.Chat.Templates[v]; !ok {
			p.mu.Unlock()
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("template %q does not exist", v))
		} else {
			cfg.Chat.GroupTemplates[gs] = v
		}
	}
	if err := plugin.WriteYAMLFrom(p.configPath, cfg); err != nil {
		p.mu.Unlock()
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("write config: %v", err))
	}
	p.mu.Unlock()

	if body.Model != nil {
		v := strings.TrimSpace(*body.Model)
		if v == "" {
			p.fileDB.Delete("model_" + gs)
		} else {
			_ = p.fileDB.Set("model_"+gs, v)
		}
	}
	if body.ChatEnabled != nil && p.chatWhiteList != nil {
		if *body.ChatEnabled {
			_ = p.chatWhiteList.Add(gid)
		} else {
			_ = p.chatWhiteList.Remove(gid)
		}
	}
	if body.AutoEnabled != nil && p.autoWhiteList != nil {
		if *body.AutoEnabled {
			_ = p.autoWhiteList.Add(gid)
		} else {
			_ = p.autoWhiteList.Remove(gid)
		}
	}
	return c.JSON(p.buildGroupPayload(cfg, gid))
}

func (p *pluginImpl) handleDeleteGroup(c *fiber.Ctx) error {
	gid, err := strconv.ParseInt(c.Params("gid"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid group id")
	}
	gs := strconv.FormatInt(gid, 10)
	p.mu.Lock()
	cfg := GetConfig()
	if cfg != nil {
		delete(cfg.Chat.Prompt.Persona, gs)
		delete(cfg.Chat.Willing.GroupThresholds, gs)
		delete(cfg.Chat.GroupTemplates, gs)
		_ = plugin.WriteYAMLFrom(p.configPath, cfg)
	}
	p.mu.Unlock()
	if p.fileDB != nil {
		p.fileDB.Delete("model_" + gs)
	}
	if p.chatWhiteList != nil {
		_ = p.chatWhiteList.Remove(gid)
	}
	if p.autoWhiteList != nil {
		_ = p.autoWhiteList.Remove(gid)
	}
	return c.JSON(fiber.Map{"ok": true})
}

package web

import (
	"errors"
	"strconv"

	"moebot-next/internal/filter"
	"moebot-next/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type effectiveRules struct {
	UserIDRules         filter.IDRule      `json:"user_id_rules"`
	GroupIDRules        filter.IDRule      `json:"group_id_rules"`
	MessageRules        filter.MessageRule `json:"message_rules"`
	PrivateMessageRules filter.MessageRule `json:"private_message_rules"`
	GroupMessageRules   filter.MessageRule `json:"group_message_rules"`
}

type filterAppPayload struct {
	ID                  uint               `json:"id"`
	Name                string             `json:"name"`
	URI                 string             `json:"uri"`
	AccessToken         string             `json:"access_token"`
	Enabled             bool               `json:"enabled"`
	Builtin             bool               `json:"builtin"`
	SortOrder           int                `json:"sort_order"`
	TemplateID          *uint              `json:"template_id"`
	UserIDRules         filter.IDRule      `json:"user_id_rules"`
	GroupIDRules        filter.IDRule      `json:"group_id_rules"`
	MessageRules        filter.MessageRule `json:"message_rules"`
	PrivateMessageRules filter.MessageRule `json:"private_message_rules"`
	GroupMessageRules   filter.MessageRule `json:"group_message_rules"`
	// EffectiveRules is the set of rules actually used at runtime, after
	// template resolution. Read-only output for the WebUI.
	EffectiveRules effectiveRules `json:"effective_rules"`
}

func computeEffectiveRules(a *models.FilterApp, tplByID map[uint]*models.FilterTemplate) effectiveRules {
	if a.TemplateID != nil {
		if t, ok := tplByID[*a.TemplateID]; ok {
			return effectiveRules{
				UserIDRules:         filter.DecodeIDRule(t.UserIDRules),
				GroupIDRules:        filter.DecodeIDRule(t.GroupIDRules),
				MessageRules:        filter.DecodeMessageRule(t.MessageRules),
				PrivateMessageRules: filter.DecodeMessageRule(t.PrivateMessageRules),
				GroupMessageRules:   filter.DecodeMessageRule(t.GroupMessageRules),
			}
		}
	}
	return effectiveRules{
		UserIDRules:         filter.DecodeIDRule(a.UserIDRules),
		GroupIDRules:        filter.DecodeIDRule(a.GroupIDRules),
		MessageRules:        filter.DecodeMessageRule(a.MessageRules),
		PrivateMessageRules: filter.DecodeMessageRule(a.PrivateMessageRules),
		GroupMessageRules:   filter.DecodeMessageRule(a.GroupMessageRules),
	}
}

func appToPayload(a *models.FilterApp, tplByID map[uint]*models.FilterTemplate) filterAppPayload {
	return filterAppPayload{
		ID: a.ID, Name: a.Name, URI: a.URI, AccessToken: a.AccessToken,
		Enabled: a.Enabled, Builtin: a.Builtin, SortOrder: a.SortOrder,
		TemplateID:          a.TemplateID,
		UserIDRules:         filter.DecodeIDRule(a.UserIDRules),
		GroupIDRules:        filter.DecodeIDRule(a.GroupIDRules),
		MessageRules:        filter.DecodeMessageRule(a.MessageRules),
		PrivateMessageRules: filter.DecodeMessageRule(a.PrivateMessageRules),
		GroupMessageRules:   filter.DecodeMessageRule(a.GroupMessageRules),
		EffectiveRules:      computeEffectiveRules(a, tplByID),
	}
}

func payloadToApp(p *filterAppPayload, dst *models.FilterApp) {
	dst.Name = p.Name
	dst.URI = p.URI
	dst.AccessToken = p.AccessToken
	dst.Enabled = p.Enabled
	dst.SortOrder = p.SortOrder
	dst.TemplateID = p.TemplateID
	dst.UserIDRules = filter.EncodeIDRule(p.UserIDRules)
	dst.GroupIDRules = filter.EncodeIDRule(p.GroupIDRules)
	dst.MessageRules = filter.EncodeMessageRule(p.MessageRules)
	dst.PrivateMessageRules = filter.EncodeMessageRule(p.PrivateMessageRules)
	dst.GroupMessageRules = filter.EncodeMessageRule(p.GroupMessageRules)
}

type gatewayPayload struct {
	Enabled    bool    `json:"enabled"`
	Host       string  `json:"host"`
	Port       int     `json:"port"`
	Suffix     string  `json:"suffix"`
	BotID      string  `json:"bot_id"`
	UserAgent  string  `json:"user_agent"`
	BufferSize int     `json:"buffer_size"`
	SleepTime  float32 `json:"sleep_time"`
	Debug      bool    `json:"debug"`
}

func gatewayToPayload(g *models.FilterGateway) gatewayPayload {
	return gatewayPayload{
		Enabled: g.Enabled, Host: g.Host, Port: g.Port, Suffix: g.Suffix,
		BotID: g.BotID, UserAgent: g.UserAgent, BufferSize: g.BufferSize,
		SleepTime: g.SleepTime, Debug: g.Debug,
	}
}

type filterTemplatePayload struct {
	ID                  uint               `json:"id"`
	Name                string             `json:"name"`
	Description         string             `json:"description"`
	Builtin             bool               `json:"builtin"`
	UserIDRules         filter.IDRule      `json:"user_id_rules"`
	GroupIDRules        filter.IDRule      `json:"group_id_rules"`
	MessageRules        filter.MessageRule `json:"message_rules"`
	PrivateMessageRules filter.MessageRule `json:"private_message_rules"`
	GroupMessageRules   filter.MessageRule `json:"group_message_rules"`
	UsageCount          int64              `json:"usage_count"`
}

func templateToPayload(t *models.FilterTemplate, usage int64) filterTemplatePayload {
	return filterTemplatePayload{
		ID: t.ID, Name: t.Name, Description: t.Description, Builtin: t.Builtin,
		UserIDRules:         filter.DecodeIDRule(t.UserIDRules),
		GroupIDRules:        filter.DecodeIDRule(t.GroupIDRules),
		MessageRules:        filter.DecodeMessageRule(t.MessageRules),
		PrivateMessageRules: filter.DecodeMessageRule(t.PrivateMessageRules),
		GroupMessageRules:   filter.DecodeMessageRule(t.GroupMessageRules),
		UsageCount:          usage,
	}
}

func payloadToTemplate(p *filterTemplatePayload, dst *models.FilterTemplate) {
	if !dst.Builtin || dst.Name == "" {
		dst.Name = p.Name
	}
	dst.Description = p.Description
	dst.UserIDRules = filter.EncodeIDRule(p.UserIDRules)
	dst.GroupIDRules = filter.EncodeIDRule(p.GroupIDRules)
	dst.MessageRules = filter.EncodeMessageRule(p.MessageRules)
	dst.PrivateMessageRules = filter.EncodeMessageRule(p.PrivateMessageRules)
	dst.GroupMessageRules = filter.EncodeMessageRule(p.GroupMessageRules)
}

func (s *Server) handleFilterStatus(c *fiber.Ctx) error {
	if s.Filter == nil {
		return c.JSON(fiber.Map{"running": false, "clients": []any{}})
	}
	return c.JSON(s.Filter.Status())
}

func (s *Server) handleGetFilterGateway(c *fiber.Ctx) error {
	gw, err := s.DB.GetOrCreateFilterGateway()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(gatewayToPayload(gw))
}

func (s *Server) handleUpdateFilterGateway(c *fiber.Ctx) error {
	var p gatewayPayload
	if err := c.BodyParser(&p); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	gw, err := s.DB.GetOrCreateFilterGateway()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	gw.Enabled = p.Enabled
	if p.Host != "" {
		gw.Host = p.Host
	}
	if p.Port > 0 {
		gw.Port = p.Port
	}
	if p.Suffix != "" {
		gw.Suffix = p.Suffix
	}
	if p.BotID != "" {
		gw.BotID = p.BotID
	}
	if p.UserAgent != "" {
		gw.UserAgent = p.UserAgent
	}
	if p.BufferSize > 0 {
		gw.BufferSize = p.BufferSize
	}
	if p.SleepTime > 0 {
		gw.SleepTime = p.SleepTime
	}
	gw.Debug = p.Debug
	if err := s.DB.UpdateFilterGateway(gw); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if s.Filter != nil {
		_ = s.Filter.Reload(c.UserContext())
	}
	return c.JSON(gatewayToPayload(gw))
}

// loadTemplateMap returns a map of all templates by id, plus the slice. The map
// is used by appToPayload to compute effective rules.
func (s *Server) loadTemplateMap() (map[uint]*models.FilterTemplate, []models.FilterTemplate, error) {
	ts, err := s.DB.ListFilterTemplates()
	if err != nil {
		return nil, nil, err
	}
	m := map[uint]*models.FilterTemplate{}
	for i := range ts {
		m[ts[i].ID] = &ts[i]
	}
	return m, ts, nil
}

func (s *Server) handleListFilterApps(c *fiber.Ctx) error {
	apps, err := s.DB.ListFilterApps()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	tplByID, _, err := s.loadTemplateMap()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	out := make([]filterAppPayload, 0, len(apps))
	for i := range apps {
		out = append(out, appToPayload(&apps[i], tplByID))
	}
	return c.JSON(fiber.Map{"items": out})
}

func (s *Server) handleCreateFilterApp(c *fiber.Ctx) error {
	var p filterAppPayload
	if err := c.BodyParser(&p); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if p.Name == "" || p.URI == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name and uri are required")
	}
	if existing, err := s.DB.GetFilterAppByName(p.Name); err == nil && existing != nil {
		return fiber.NewError(fiber.StatusConflict, "name already in use")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	app := &models.FilterApp{}
	payloadToApp(&p, app)
	if err := s.DB.CreateFilterApp(app); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if s.Filter != nil {
		_ = s.Filter.Reload(c.UserContext())
	}
	tplByID, _, _ := s.loadTemplateMap()
	return c.JSON(appToPayload(app, tplByID))
}

func (s *Server) handleUpdateFilterApp(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var p filterAppPayload
	if err := c.BodyParser(&p); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	app, err := s.DB.GetFilterApp(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "filter app not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if app.Builtin {
		// Built-in: keep name/uri/builtin flag, allow rule edits.
		p.Name = app.Name
		p.URI = app.URI
	}
	payloadToApp(&p, app)
	if err := s.DB.UpdateFilterApp(app); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if s.Filter != nil {
		_ = s.Filter.Reload(c.UserContext())
	}
	tplByID, _, _ := s.loadTemplateMap()
	return c.JSON(appToPayload(app, tplByID))
}

func (s *Server) handleDeleteFilterApp(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := s.DB.DeleteFilterApp(uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "filter app not found")
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if s.Filter != nil {
		_ = s.Filter.Reload(c.UserContext())
	}
	return c.JSON(fiber.Map{"ok": true})
}

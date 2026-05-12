package autochat

import (
	"fmt"
	"sort"
	"strings"

	"moebot-next/internal/plugin"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// templatePayload 是 ChatTemplate 的前端 DTO。Multimodal 用 *bool 保留三态：
//   - nil 自动按主选模型是否在 multimodal_models 列表里判定
//   - true / false 强制开/关图片直传
type templatePayload struct {
	Name       string   `json:"name"`
	Persona    string   `json:"persona"`
	Models     []string `json:"models"`
	Multimodal *bool    `json:"multimodal"`
	// 数值字段使用 *float64 以容忍 JSON null（前端清空 input 时会发 null）。
	WillingThreshold *float64 `json:"willing_threshold"`
	AtDelta          *float64 `json:"at_delta"`
	KeywordDelta     *float64 `json:"keyword_delta"`
	RandomDeltaMax   *float64 `json:"random_delta_max"`
	Keywords         []string `json:"keywords"`
	UsedByGroups     []string `json:"used_by_groups"`
}

func ptrFloat(v float64) *float64 { return &v }
func derefFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func buildTemplatePayload(c *Config, name string) templatePayload {
	t := c.Chat.Templates[name]
	used := []string{}
	for gid, tn := range c.Chat.GroupTemplates {
		if tn == name {
			used = append(used, gid)
		}
	}
	sort.Strings(used)
	return templatePayload{
		Name:             name,
		Persona:          t.Persona,
		Models:           append([]string{}, t.Models...),
		Multimodal:       t.Multimodal,
		WillingThreshold: ptrFloat(t.WillingThreshold),
		AtDelta:          ptrFloat(t.AtDelta),
		KeywordDelta:     ptrFloat(t.KeywordDelta),
		RandomDeltaMax:   ptrFloat(t.RandomDeltaMax),
		Keywords:         append([]string{}, t.Keywords...),
		UsedByGroups:     used,
	}
}

func (p *pluginImpl) handleListTemplates(c *fiber.Ctx) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	cfg := GetConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	names := make([]string, 0, len(cfg.Chat.Templates))
	for k := range cfg.Chat.Templates {
		names = append(names, k)
	}
	sort.Strings(names)
	out := make([]templatePayload, 0, len(names))
	for _, n := range names {
		out = append(out, buildTemplatePayload(cfg, n))
	}
	return c.JSON(fiber.Map{"templates": out})
}

func (p *pluginImpl) handleUpsertTemplate(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.Params("name"))
	if name == "" || name == "default" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid template name")
	}
	var body templatePayload
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	// 防御：body.Name 必须与 URL name 一致（前端正常调用一定一致；不一致
	// 多半是异常状态或调用错位，直接拒绝以暴露问题，避免静默写错 key）。
	if bn := strings.TrimSpace(body.Name); bn != "" && bn != name {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("template name mismatch: url=%q body=%q", name, bn))
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	cfg := GetConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	if cfg.Chat.Templates == nil {
		cfg.Chat.Templates = map[string]ChatTemplate{}
	}
	models := make([]string, 0, len(body.Models))
	for _, m := range body.Models {
		if s := strings.TrimSpace(m); s != "" {
			models = append(models, s)
		}
	}
	keywords := make([]string, 0, len(body.Keywords))
	for _, k := range body.Keywords {
		if s := strings.TrimSpace(k); s != "" {
			keywords = append(keywords, s)
		}
	}
	cfg.Chat.Templates[name] = ChatTemplate{
		Persona:          body.Persona,
		Models:           models,
		Multimodal:       body.Multimodal,
		WillingThreshold: derefFloat(body.WillingThreshold),
		AtDelta:          derefFloat(body.AtDelta),
		KeywordDelta:     derefFloat(body.KeywordDelta),
		RandomDeltaMax:   derefFloat(body.RandomDeltaMax),
		Keywords:         keywords,
	}
	logTemplateKeys(cfg, "upsert", name)
	if err := plugin.WriteYAMLFrom(p.configPath, cfg); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("write config: %v", err))
	}
	return c.JSON(buildTemplatePayload(cfg, name))
}

func (p *pluginImpl) handleDeleteTemplate(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.Params("name"))
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid template name")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	cfg := GetConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	delete(cfg.Chat.Templates, name)
	// 清掉所有 GroupTemplates 中指向该模板的绑定，避免悬空引用。
	for gid, tn := range cfg.Chat.GroupTemplates {
		if tn == name {
			delete(cfg.Chat.GroupTemplates, gid)
		}
	}
	logTemplateKeys(cfg, "delete", name)
	if err := plugin.WriteYAMLFrom(p.configPath, cfg); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("write config: %v", err))
	}
	return c.JSON(fiber.Map{"ok": true, "name": name})
}

// logTemplateKeys 把当前 Templates 与 GroupTemplates 的 key 列出来；用于
// 定位"模板被错改/丢失"类问题：每次落盘前留下一条 service-side 真实状态。
func logTemplateKeys(c *Config, op, name string) {
	keys := make([]string, 0, len(c.Chat.Templates))
	for k := range c.Chat.Templates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	log.Info().Str("op", op).Str("name", name).Strs("templates", keys).Interface("group_templates", c.Chat.GroupTemplates).Msg("[autochat] template state")
}

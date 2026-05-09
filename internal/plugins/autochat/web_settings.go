package autochat

import (
	"fmt"
	"os"
	"strings"

	"moebot-next/internal/plugin"

	"github.com/gofiber/fiber/v2"
)

func readFileBytes(path string) ([]byte, error) { return os.ReadFile(path) }

// providerDTO 对应 ProviderConfig，对前端友好。
type providerDTO struct {
	Name             string `json:"name"`
	Type             string `json:"type"` // openai | anthropic
	BaseURL          string `json:"base_url"`
	APIKey           string `json:"api_key"`
	Timeout          int    `json:"timeout"`
	AnthropicVersion string `json:"anthropic_version,omitempty"`
}

// providersPayload 是 /api/plugins/autochat/providers 的结构化 DTO。
type providersPayload struct {
	ProviderList []providerDTO `json:"provider_list"`
	LLM          struct {
		Models           []string `json:"models"`
		MultimodalModels []string `json:"multimodal_models"`
		MaxTokens        int      `json:"max_tokens"`
		Reasoning        bool     `json:"reasoning"`
		Timeout          int      `json:"timeout"`
	} `json:"llm"`
	Embedding struct {
		Enabled    bool   `json:"enabled"`
		Provider   string `json:"provider"` // 引用 provider_list 中的 Name；空则使用下方独立 base_url/api_key
		BaseURL    string `json:"base_url"`
		APIKey     string `json:"api_key"`
		Model      string `json:"model"`
		Dimensions int    `json:"dimensions"`
		Timeout    int    `json:"timeout"`
	} `json:"embedding"`
	Rerank struct {
		Enabled   bool    `json:"enabled"`
		Provider  string  `json:"provider"`
		BaseURL   string  `json:"base_url"`
		APIKey    string  `json:"api_key"`
		Model     string  `json:"model"`
		Threshold float64 `json:"threshold"`
		Timeout   int     `json:"timeout"`
	} `json:"rerank"`
	Vector struct {
		Enabled    bool `json:"enabled"`
		Dimensions int  `json:"dimensions"`
		TopK       int  `json:"top_k"`
	} `json:"vector"`
	ImageCaption struct {
		Enabled   bool   `json:"enabled"`
		Model     string `json:"model"`
		Timeout   int    `json:"timeout"`
		MaxTokens int    `json:"max_tokens"`
		Prompt    string `json:"prompt"`
	} `json:"image_caption"`
	RAGSummary struct {
		Enabled   bool   `json:"enabled"`
		Model     string `json:"model"`
		Timeout   int    `json:"timeout"`
		MaxTokens int    `json:"max_tokens"`
	} `json:"rag_summary"`
}

func buildProvidersPayload(c *Config) providersPayload {
	var p providersPayload
	p.ProviderList = make([]providerDTO, 0, len(c.LLM.ProviderList))
	for _, pc := range c.LLM.ProviderList {
		p.ProviderList = append(p.ProviderList, providerDTO{
			Name: pc.Name, Type: pc.Type,
			BaseURL: pc.BaseURL, APIKey: pc.APIKey, Timeout: pc.Timeout,
			AnthropicVersion: pc.AnthropicVersion,
		})
	}
	p.LLM.Models = append([]string{}, c.LLM.Models...)
	p.LLM.MultimodalModels = append([]string{}, c.LLM.MultimodalModels...)
	p.LLM.MaxTokens = c.LLM.MaxTokens
	p.LLM.Reasoning = c.LLM.Reasoning
	p.LLM.Timeout = c.LLM.Timeout
	p.Embedding.Enabled = c.Embedding.Enabled
	p.Embedding.Provider = c.Embedding.Provider
	p.Embedding.BaseURL = c.Embedding.BaseURL
	p.Embedding.APIKey = c.Embedding.APIKey
	p.Embedding.Model = c.Embedding.Model
	p.Embedding.Dimensions = c.Embedding.Dimensions
	p.Embedding.Timeout = c.Embedding.Timeout
	p.Rerank.Enabled = c.Rerank.Enabled
	p.Rerank.Provider = c.Rerank.Provider
	p.Rerank.BaseURL = c.Rerank.BaseURL
	p.Rerank.APIKey = c.Rerank.APIKey
	p.Rerank.Model = c.Rerank.Model
	p.Rerank.Threshold = c.Rerank.Threshold
	p.Rerank.Timeout = c.Rerank.Timeout
	p.Vector.Enabled = c.Vector.Enabled
	p.Vector.Dimensions = c.Vector.Dimensions
	p.Vector.TopK = c.Vector.TopK
	p.ImageCaption.Enabled = c.ImageCaption.Enabled
	p.ImageCaption.Model = c.ImageCaption.Model
	p.ImageCaption.Timeout = c.ImageCaption.Timeout
	p.ImageCaption.MaxTokens = c.ImageCaption.MaxTokens
	p.ImageCaption.Prompt = c.ImageCaption.Prompt
	p.RAGSummary.Enabled = c.RAGSummary.Enabled
	p.RAGSummary.Model = c.RAGSummary.Model
	p.RAGSummary.Timeout = c.RAGSummary.Timeout
	p.RAGSummary.MaxTokens = c.RAGSummary.MaxTokens
	return p
}

func (p *pluginImpl) handleGetProviders(c *fiber.Ctx) error {
	cfg := GetConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	return c.JSON(buildProvidersPayload(cfg))
}

func (p *pluginImpl) handlePutProviders(c *fiber.Ctx) error {
	var body providersPayload
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	cfg := GetConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	// 重建 ProviderList：去重 Name；强制 Type=openai/anthropic。
	seen := map[string]bool{}
	pl := make([]ProviderConfig, 0, len(body.ProviderList))
	for _, d := range body.ProviderList {
		name := strings.TrimSpace(d.Name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		typ := d.Type
		if typ != "anthropic" {
			typ = "openai"
		}
		pl = append(pl, ProviderConfig{
			Name: name, Type: typ,
			BaseURL: strings.TrimSpace(d.BaseURL), APIKey: d.APIKey,
			Timeout: d.Timeout, AnthropicVersion: d.AnthropicVersion,
		})
	}
	cfg.LLM.ProviderList = pl
	cfg.LLM.Providers = legacyProviders{} // 清空 v0 字段
	cfg.LLM.Models = append([]string{}, body.LLM.Models...)
	cfg.LLM.MultimodalModels = append([]string{}, body.LLM.MultimodalModels...)
	cfg.LLM.MaxTokens = body.LLM.MaxTokens
	cfg.LLM.Reasoning = body.LLM.Reasoning
	cfg.LLM.Timeout = body.LLM.Timeout
	cfg.Embedding.Enabled = body.Embedding.Enabled
	cfg.Embedding.Provider = body.Embedding.Provider
	cfg.Embedding.BaseURL = body.Embedding.BaseURL
	cfg.Embedding.APIKey = body.Embedding.APIKey
	cfg.Embedding.Model = body.Embedding.Model
	cfg.Embedding.Dimensions = body.Embedding.Dimensions
	cfg.Embedding.Timeout = body.Embedding.Timeout
	cfg.Rerank.Enabled = body.Rerank.Enabled
	cfg.Rerank.Provider = body.Rerank.Provider
	cfg.Rerank.BaseURL = body.Rerank.BaseURL
	cfg.Rerank.APIKey = body.Rerank.APIKey
	cfg.Rerank.Model = body.Rerank.Model
	cfg.Rerank.Threshold = body.Rerank.Threshold
	cfg.Rerank.Timeout = body.Rerank.Timeout
	cfg.Vector.Enabled = body.Vector.Enabled
	cfg.Vector.Dimensions = body.Vector.Dimensions
	cfg.Vector.TopK = body.Vector.TopK
	cfg.ImageCaption.Enabled = body.ImageCaption.Enabled
	cfg.ImageCaption.Model = body.ImageCaption.Model
	cfg.ImageCaption.Timeout = body.ImageCaption.Timeout
	cfg.ImageCaption.MaxTokens = body.ImageCaption.MaxTokens
	cfg.ImageCaption.Prompt = body.ImageCaption.Prompt
	cfg.RAGSummary.Enabled = body.RAGSummary.Enabled
	cfg.RAGSummary.Model = body.RAGSummary.Model
	cfg.RAGSummary.Timeout = body.RAGSummary.Timeout
	cfg.RAGSummary.MaxTokens = body.RAGSummary.MaxTokens
	applyDefaults(cfg)

	if err := plugin.WriteYAMLFrom(p.configPath, cfg); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("write config: %v", err))
	}
	p.applyProviders(cfg)
	return c.JSON(buildProvidersPayload(cfg))
}

// personaPayload 对应 prompt + rag_summary 段。
type personaPayload struct {
	DefaultPersona string            `json:"default_persona"`
	Framework      string            `json:"framework"`
	GroupPersonas  map[string]string `json:"group_personas"`
	RAGSummary     struct {
		Enabled   bool   `json:"enabled"`
		Model     string `json:"model"`
		Timeout   int    `json:"timeout"`
		MaxTokens int    `json:"max_tokens"`
		Prompt    string `json:"prompt"`
	} `json:"rag_summary"`
}

func buildPersonaPayload(c *Config) personaPayload {
	var p personaPayload
	if c.Chat.Prompt.Persona != nil {
		p.DefaultPersona = c.Chat.Prompt.Persona["default"]
		p.GroupPersonas = map[string]string{}
		for k, v := range c.Chat.Prompt.Persona {
			if k != "default" {
				p.GroupPersonas[k] = v
			}
		}
	}
	p.Framework = c.Chat.Prompt.Framework
	p.RAGSummary.Enabled = c.RAGSummary.Enabled
	p.RAGSummary.Model = c.RAGSummary.Model
	p.RAGSummary.Timeout = c.RAGSummary.Timeout
	p.RAGSummary.MaxTokens = c.RAGSummary.MaxTokens
	p.RAGSummary.Prompt = c.RAGSummary.Prompt
	return p
}

func (p *pluginImpl) handleGetPersona(c *fiber.Ctx) error {
	cfg := GetConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	return c.JSON(buildPersonaPayload(cfg))
}

func (p *pluginImpl) handlePutPersona(c *fiber.Ctx) error {
	var body personaPayload
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	cfg := GetConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	if cfg.Chat.Prompt.Persona == nil {
		cfg.Chat.Prompt.Persona = map[string]string{}
	}
	cfg.Chat.Prompt.Persona["default"] = body.DefaultPersona
	cfg.Chat.Prompt.Framework = body.Framework
	cfg.RAGSummary.Enabled = body.RAGSummary.Enabled
	cfg.RAGSummary.Model = body.RAGSummary.Model
	cfg.RAGSummary.Timeout = body.RAGSummary.Timeout
	cfg.RAGSummary.MaxTokens = body.RAGSummary.MaxTokens
	cfg.RAGSummary.Prompt = body.RAGSummary.Prompt
	if err := plugin.WriteYAMLFrom(p.configPath, cfg); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("write config: %v", err))
	}
	return c.JSON(buildPersonaPayload(cfg))
}

// triggersPayload 对应 chat 段中触发与缓冲相关字段。
type triggersPayload struct {
	WillingThreshold float64  `json:"willing_threshold"`
	AtDelta          float64  `json:"at_delta"`
	KeywordDelta     float64  `json:"keyword_delta"`
	RandomDeltaMax   float64  `json:"random_delta_max"`
	ChatCDSeconds    int      `json:"chat_cd_seconds"`
	TTSCDSeconds     int      `json:"tts_cd_seconds"`
	ContextSize      int      `json:"context_size"`
	BufferLimit      int      `json:"buffer_limit"`
	ReplyMaxLength   int      `json:"reply_max_length"`
	Keywords         []string `json:"keywords"`
	IgnorePrefixes   []string `json:"ignore_prefixes"`
	IgnorePatterns   []string `json:"ignore_patterns"`
}

func buildTriggersPayload(c *Config) triggersPayload {
	return triggersPayload{
		WillingThreshold: c.Chat.Willing.Threshold,
		AtDelta:          c.Chat.Willing.AtDelta,
		KeywordDelta:     c.Chat.Willing.KeywordDelta,
		RandomDeltaMax:   c.Chat.Willing.RandomDeltaMax,
		ChatCDSeconds:    c.Chat.ChatCDSeconds,
		TTSCDSeconds:     c.Chat.TTSCDSeconds,
		ContextSize:      c.Chat.ContextSize,
		BufferLimit:      c.Chat.BufferLimit,
		ReplyMaxLength:   c.Chat.ReplyMaxLength,
		Keywords:         append([]string{}, c.Chat.Keywords...),
		IgnorePrefixes:   append([]string{}, c.Chat.IgnorePrefixes...),
		IgnorePatterns:   append([]string{}, c.Chat.IgnorePatterns...),
	}
}

func (p *pluginImpl) handleGetTriggers(c *fiber.Ctx) error {
	cfg := GetConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	return c.JSON(buildTriggersPayload(cfg))
}

func (p *pluginImpl) handlePutTriggers(c *fiber.Ctx) error {
	var body triggersPayload
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	cfg := GetConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	cfg.Chat.Willing.Threshold = body.WillingThreshold
	if body.AtDelta > 0 {
		cfg.Chat.Willing.AtDelta = body.AtDelta
	}
	if body.KeywordDelta > 0 {
		cfg.Chat.Willing.KeywordDelta = body.KeywordDelta
	}
	if body.RandomDeltaMax > 0 {
		cfg.Chat.Willing.RandomDeltaMax = body.RandomDeltaMax
	}
	cfg.Chat.ChatCDSeconds = body.ChatCDSeconds
	cfg.Chat.TTSCDSeconds = body.TTSCDSeconds
	cfg.Chat.ContextSize = body.ContextSize
	cfg.Chat.BufferLimit = body.BufferLimit
	cfg.Chat.ReplyMaxLength = body.ReplyMaxLength
	cfg.Chat.Keywords = append([]string{}, body.Keywords...)
	cfg.Chat.IgnorePrefixes = append([]string{}, body.IgnorePrefixes...)
	cfg.Chat.IgnorePatterns = append([]string{}, body.IgnorePatterns...)
	if err := plugin.WriteYAMLFrom(p.configPath, cfg); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("write config: %v", err))
	}
	return c.JSON(buildTriggersPayload(cfg))
}

// handleOverview 汇总 autochat 状态、provider 可用性与 token 统计。
func (p *pluginImpl) handleOverview(c *fiber.Ctx) error {
	cfg := GetConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat config not loaded")
	}
	pt1, ct1, rc1 := 0, 0, 0
	pt7, ct7, rc7 := 0, 0, 0
	if p.tokenStats != nil {
		pt1, ct1, rc1 = p.tokenStats.GetStats(1)
		pt7, ct7, rc7 = p.tokenStats.GetStats(7)
	}
	openaiOK, anthropicOK := false, false
	for _, pc := range cfg.LLM.ProviderList {
		if pc.APIKey == "" || pc.BaseURL == "" {
			continue
		}
		if pc.Type == "anthropic" {
			anthropicOK = true
		} else {
			openaiOK = true
		}
	}
	embAPIKey := cfg.Embedding.APIKey
	if pc := resolveProviderConfig(cfg, cfg.Embedding.Provider); pc != nil {
		embAPIKey = pc.APIKey
	}
	embOK := cfg.Embedding.Enabled && embAPIKey != "" && cfg.Embedding.Model != ""
	rrAPIKey := cfg.Rerank.APIKey
	if pc := resolveProviderConfig(cfg, cfg.Rerank.Provider); pc != nil {
		rrAPIKey = pc.APIKey
	}
	rerankOK := cfg.Rerank.Enabled && rrAPIKey != "" && cfg.Rerank.Model != ""
	captionOK := cfg.ImageCaption.Enabled && cfg.ImageCaption.Model != ""
	primary := ""
	if len(cfg.LLM.Models) > 0 {
		primary = cfg.LLM.Models[0]
	}
	return c.JSON(fiber.Map{
		"primary_model": primary,
		"models_count":  len(cfg.LLM.Models),
		"providers": fiber.Map{
			"openai":        openaiOK,
			"anthropic":     anthropicOK,
			"embedding":     embOK,
			"rerank":        rerankOK,
			"vector":        cfg.Vector.Enabled,
			"image_caption": captionOK,
		},
		"keywords_count":    len(cfg.Chat.Keywords),
		"willing_threshold": cfg.Chat.Willing.Threshold,
		"group_overrides":   len(cfg.Chat.Willing.GroupThresholds),
		"persona_overrides": len(cfg.Chat.Prompt.Persona) - 1, // minus default
		"token_stats_today": fiber.Map{"prompt": pt1, "completion": ct1, "requests": rc1, "total": pt1 + ct1},
		"token_stats_7days": fiber.Map{"prompt": pt7, "completion": ct7, "requests": rc7, "total": pt7 + ct7},
	})
}

// handleGetYAML 返回 autochat.yml 原文（与通用 /api/plugins/:name/config 重叠，
// 但这里方便前端 Settings 页内的"高级"tab 直接调用）。
func (p *pluginImpl) handleGetYAML(c *fiber.Ctx) error {
	if p.configPath == "" {
		return fiber.NewError(fiber.StatusServiceUnavailable, "autochat: plugin not initialized")
	}
	data, err := readFileBytes(p.configPath)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"path": p.configPath, "yaml": string(data)})
}

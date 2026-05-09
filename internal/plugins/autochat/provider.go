package autochat

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// LLMResponse LLM 通用响应
type LLMResponse struct {
	Result           string  `json:"result"`
	Model            string  `json:"model"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	Cost             float64 `json:"cost"`
	Reasoning        string  `json:"reasoning,omitempty"`
}

// LLMProvider 抽象 LLM 客户端。
type LLMProvider interface {
	Name() string
	Chat(ctx context.Context, sess *ChatSession, model string, maxTokens int) (*LLMResponse, error)
}

var (
	providersMu sync.RWMutex
	providers   = map[string]LLMProvider{}
)

func registerProvider(p LLMProvider) {
	if p == nil {
		return
	}
	providersMu.Lock()
	providers[p.Name()] = p
	providersMu.Unlock()
}

func clearProviders() {
	providersMu.Lock()
	providers = map[string]LLMProvider{}
	providersMu.Unlock()
}

func getProvider(name string) (LLMProvider, bool) {
	providersMu.RLock()
	defer providersMu.RUnlock()
	p, ok := providers[name]
	return p, ok
}

// parseModelSpec 解析 `<provider>:<model>` 形式的模型字符串，
// 兼容别名：oa→openai, an/claude→anthropic。无前缀时默认 openai。
func parseModelSpec(spec string) (provider, model string) {
	idx := strings.Index(spec, ":")
	if idx <= 0 {
		return "openai", spec
	}
	provider = strings.ToLower(spec[:idx])
	model = spec[idx+1:]
	switch provider {
	case "oa":
		provider = "openai"
	case "an", "claude":
		provider = "anthropic"
	}
	return
}

// UniversalChat 按模型字符串前缀分发到对应 provider。
func UniversalChat(ctx context.Context, sess *ChatSession, modelSpec string, maxTokens int) (*LLMResponse, error) {
	provName, model := parseModelSpec(modelSpec)
	p, ok := getProvider(provName)
	if !ok {
		return nil, fmt.Errorf("autochat: provider %q 未配置", provName)
	}
	return p.Chat(ctx, sess, model, maxTokens)
}

// UniversalChatWithFallback 按 cfg.LLM.Models 顺序尝试，遇到错误自动切换下一个。
func UniversalChatWithFallback(ctx context.Context, sess *ChatSession, models []string, maxTokens int) (*LLMResponse, error) {
	var lastErr error
	for _, m := range models {
		resp, err := UniversalChat(ctx, sess, m, maxTokens)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("autochat: 没有可用模型")
	}
	return nil, lastErr
}

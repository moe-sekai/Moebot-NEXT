package autochat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AnthropicProvider 实现 Anthropic Messages API（/v1/messages）。
//
// 协议要点：
//   - system 字段独立于 messages 之外
//   - messages 中只允许 user / assistant 两种 role
//   - 多模态：图片以 {"type":"image","source":{"type":"base64",...}} 注入
//   - 鉴权头：x-api-key + anthropic-version
type AnthropicProvider struct {
	baseURL string
	apiKey  string
	version string
	client  *http.Client
}

func newAnthropicProvider(baseURL, apiKey, version string, timeoutSeconds int) *AnthropicProvider {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60
	}
	if version == "" {
		version = "2023-06-01"
	}
	return &AnthropicProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		version: version,
		client:  &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
	}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

func (p *AnthropicProvider) Chat(ctx context.Context, sess *ChatSession, model string, maxTokens int) (*LLMResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("anthropic: api_key 未配置")
	}
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	type imageSource struct {
		Type      string `json:"type"`       // base64
		MediaType string `json:"media_type"` // image/jpeg ...
		Data      string `json:"data"`
	}
	type contentBlock struct {
		Type   string       `json:"type"`             // text | image
		Text   string       `json:"text,omitempty"`
		Source *imageSource `json:"source,omitempty"`
	}
	type message struct {
		Role    string         `json:"role"`
		Content []contentBlock `json:"content"`
	}
	type request struct {
		Model     string    `json:"model"`
		System    string    `json:"system,omitempty"`
		Messages  []message `json:"messages"`
		MaxTokens int       `json:"max_tokens"`
	}

	var msgs []message
	for _, m := range sess.Snapshot() {
		role := "user"
		if m.Role == RoleAssistant {
			role = "assistant"
		}
		blocks := make([]contentBlock, 0, len(m.Images)+1)
		if m.Content != "" {
			blocks = append(blocks, contentBlock{Type: "text", Text: m.Content})
		}
		for _, img := range m.Images {
			mt, raw := splitDataURI(img)
			blocks = append(blocks, contentBlock{
				Type:   "image",
				Source: &imageSource{Type: "base64", MediaType: mt, Data: raw},
			})
		}
		if len(blocks) == 0 {
			continue
		}
		msgs = append(msgs, message{Role: role, Content: blocks})
	}

	req := request{
		Model:     model,
		System:    sess.SystemPrompt,
		Messages:  msgs,
		MaxTokens: maxTokens,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshal: %w", err)
	}

	url := p.baseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", p.version)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic: %d %s", resp.StatusCode, string(respBody))
	}

	var parsed struct {
		Model   string `json:"model"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("anthropic: unmarshal: %w; body=%s", err, string(respBody))
	}
	out := &LLMResponse{
		Model:            parsed.Model,
		PromptTokens:     parsed.Usage.InputTokens,
		CompletionTokens: parsed.Usage.OutputTokens,
	}
	var sb strings.Builder
	for _, c := range parsed.Content {
		if c.Type == "text" {
			sb.WriteString(c.Text)
		}
	}
	out.Result = sb.String()
	return out, nil
}

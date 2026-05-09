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

// OpenAIProvider 实现 OpenAI 兼容的 /v1/chat/completions 协议。
// 适用于 OpenAI 官方 / Azure / SiliconFlow / Ollama / OpenRouter 等。
type OpenAIProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func newOpenAIProvider(baseURL, apiKey string, timeoutSeconds int) *OpenAIProvider {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60
	}
	return &OpenAIProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		client:  &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
	}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Chat(ctx context.Context, sess *ChatSession, model string, maxTokens int) (*LLMResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("openai: api_key 未配置")
	}

	type imageURL struct {
		URL    string `json:"url"`
		Detail string `json:"detail,omitempty"`
	}
	type contentPart struct {
		Type     string    `json:"type"`
		Text     string    `json:"text,omitempty"`
		ImageURL *imageURL `json:"image_url,omitempty"`
	}
	type chatMessage struct {
		Role    string `json:"role"`
		Content any    `json:"content"`
	}
	type chatRequest struct {
		Model       string        `json:"model"`
		Messages    []chatMessage `json:"messages"`
		MaxTokens   int           `json:"max_tokens,omitempty"`
		Temperature float64       `json:"temperature,omitempty"`
		Stream      bool          `json:"stream"`
	}

	var msgs []chatMessage
	if sess.SystemPrompt != "" {
		msgs = append(msgs, chatMessage{Role: "system", Content: sess.SystemPrompt})
	}
	for _, m := range sess.Snapshot() {
		role := "user"
		if m.Role == RoleAssistant {
			role = "assistant"
		}
		if len(m.Images) > 0 {
			parts := make([]contentPart, 0, len(m.Images)+1)
			if m.Content != "" {
				parts = append(parts, contentPart{Type: "text", Text: m.Content})
			}
			for _, img := range m.Images {
				parts = append(parts, contentPart{Type: "image_url", ImageURL: &imageURL{URL: img, Detail: "auto"}})
			}
			msgs = append(msgs, chatMessage{Role: role, Content: parts})
		} else {
			msgs = append(msgs, chatMessage{Role: role, Content: m.Content})
		}
	}
	req := chatRequest{
		Model:       model,
		Messages:    msgs,
		MaxTokens:   maxTokens,
		Temperature: 0.7,
		Stream:      false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("openai: marshal: %w", err)
	}

	url := p.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai: %d %s", resp.StatusCode, string(respBody))
	}

	var parsed struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("openai: unmarshal: %w; body=%s", err, string(respBody))
	}
	out := &LLMResponse{
		Model:            parsed.Model,
		PromptTokens:     parsed.Usage.PromptTokens,
		CompletionTokens: parsed.Usage.CompletionTokens,
	}
	if len(parsed.Choices) > 0 {
		out.Result = parsed.Choices[0].Message.Content
	}
	return out, nil
}

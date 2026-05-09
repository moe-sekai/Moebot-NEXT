package autochat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// EmbeddingClient 包装 OpenAI 兼容的 /v1/embeddings 端点。
type EmbeddingClient struct {
	baseURL    string
	apiKey     string
	model      string
	dimensions int
	client     *http.Client
	enabled    bool
}

var (
	embeddingClient *EmbeddingClient
	embeddingMu     sync.RWMutex
)

func initEmbeddingClient(c *Config) {
	embeddingMu.Lock()
	defer embeddingMu.Unlock()
	if !c.Embedding.Enabled || c.Embedding.APIKey == "" {
		embeddingClient = &EmbeddingClient{enabled: false}
		return
	}
	timeout := c.Embedding.Timeout
	if timeout <= 0 {
		timeout = 30
	}
	embeddingClient = &EmbeddingClient{
		baseURL:    strings.TrimRight(c.Embedding.BaseURL, "/"),
		apiKey:     c.Embedding.APIKey,
		model:      c.Embedding.Model,
		dimensions: c.Embedding.Dimensions,
		client:     &http.Client{Timeout: time.Duration(timeout) * time.Second},
		enabled:    true,
	}
}

// GetEmbeddingClient 返回当前 Embedding 客户端（可能是禁用占位）。
func GetEmbeddingClient() *EmbeddingClient {
	embeddingMu.RLock()
	defer embeddingMu.RUnlock()
	return embeddingClient
}

func (e *EmbeddingClient) IsEnabled() bool { return e != nil && e.enabled }

// GetEmbeddings 批量请求 embedding。
func (e *EmbeddingClient) GetEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if !e.IsEnabled() {
		return nil, fmt.Errorf("embedding: 未启用")
	}
	if len(texts) == 0 {
		return nil, nil
	}
	type request struct {
		Model          string   `json:"model"`
		Input          []string `json:"input"`
		EncodingFormat string   `json:"encoding_format,omitempty"`
		Dimensions     int      `json:"dimensions,omitempty"`
	}
	req := request{
		Model:          e.model,
		Input:          texts,
		EncodingFormat: "float",
		Dimensions:     e.dimensions,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	url := e.baseURL + "/embeddings"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+e.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding: %d %s", resp.StatusCode, string(respBody))
	}
	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}
	out := make([][]float32, len(texts))
	for _, item := range parsed.Data {
		if item.Index >= 0 && item.Index < len(out) {
			out[item.Index] = item.Embedding
		}
	}
	return out, nil
}

func (e *EmbeddingClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	vs, err := e.GetEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vs) == 0 || vs[0] == nil {
		return nil, fmt.Errorf("embedding: 未返回向量")
	}
	return vs[0], nil
}

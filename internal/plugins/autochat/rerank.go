package autochat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// RerankClient 兼容 SiliconFlow / Jina 等 OpenAI 风格的 /v1/rerank 端点。
type RerankClient struct {
	baseURL   string
	apiKey    string
	model     string
	timeout   time.Duration
	threshold float64
	enabled   bool
}

var (
	rerankClient *RerankClient
	rerankMu     sync.RWMutex
)

func initRerankClient(c *Config) {
	rerankMu.Lock()
	defer rerankMu.Unlock()
	if !c.Rerank.Enabled {
		rerankClient = &RerankClient{enabled: false}
		return
	}
	baseURL := c.Rerank.BaseURL
	apiKey := c.Rerank.APIKey
	if pc := resolveProviderConfig(c, c.Rerank.Provider); pc != nil {
		baseURL = pc.BaseURL
		apiKey = pc.APIKey
	}
	if apiKey == "" {
		rerankClient = &RerankClient{enabled: false}
		return
	}
	timeout := c.Rerank.Timeout
	if timeout <= 0 {
		timeout = 15
	}
	rerankClient = &RerankClient{
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		model:     c.Rerank.Model,
		timeout:   time.Duration(timeout) * time.Second,
		threshold: c.Rerank.Threshold,
		enabled:   true,
	}
}

func GetRerankClient() *RerankClient {
	rerankMu.RLock()
	defer rerankMu.RUnlock()
	return rerankClient
}

func (r *RerankClient) IsEnabled() bool { return r != nil && r.enabled }

// RerankItem 单条 rerank 输出。
type RerankItem struct {
	OriginalIndex  int
	Text           string
	RelevanceScore float64
}

func (r *RerankClient) Rerank(ctx context.Context, query string, documents []string, topN int) ([]RerankItem, error) {
	if !r.IsEnabled() || len(documents) == 0 {
		return nil, nil
	}
	if topN <= 0 {
		topN = len(documents)
	}
	type request struct {
		Model     string   `json:"model"`
		Query     string   `json:"query"`
		Documents []string `json:"documents"`
		TopN      int      `json:"top_n,omitempty"`
	}
	body, err := json.Marshal(request{Model: r.model, Query: query, Documents: documents, TopN: topN})
	if err != nil {
		return nil, err
	}
	url := r.baseURL + "/rerank"
	client := &http.Client{Timeout: r.timeout}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rerank: %d %s", resp.StatusCode, string(respBody))
	}
	var parsed struct {
		Results []struct {
			Index          int     `json:"index"`
			RelevanceScore float64 `json:"relevance_score"`
		} `json:"results"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}
	var out []RerankItem
	for _, res := range parsed.Results {
		if res.RelevanceScore < r.threshold {
			continue
		}
		if res.Index < 0 || res.Index >= len(documents) {
			continue
		}
		out = append(out, RerankItem{OriginalIndex: res.Index, Text: documents[res.Index], RelevanceScore: res.RelevanceScore})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RelevanceScore > out[j].RelevanceScore })
	log.Debug().Str("query", query).Int("input", len(documents)).Int("kept", len(out)).Msg("[autochat][rerank] done")
	return out, nil
}

// RerankMemories 对 MemoryVector 列表执行重排，返回新顺序。
func (r *RerankClient) RerankMemories(ctx context.Context, query string, memories []MemoryVector) ([]MemoryVector, error) {
	if !r.IsEnabled() || len(memories) == 0 {
		return memories, nil
	}
	docs := make([]string, len(memories))
	for i, m := range memories {
		docs[i] = m.Text
	}
	res, err := r.Rerank(ctx, query, docs, len(docs))
	if err != nil {
		return memories, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	out := make([]MemoryVector, 0, len(res))
	for _, item := range res {
		mem := memories[item.OriginalIndex]
		mem.Score = item.RelevanceScore
		out = append(out, mem)
	}
	return out, nil
}

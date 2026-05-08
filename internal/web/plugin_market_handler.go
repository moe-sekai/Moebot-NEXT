package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"moebot-next/internal/plugin"

	"github.com/gofiber/fiber/v2"
)

// 插件市场：通过 GitHub Contents API 列出
// FloatTech/ZeroBot-Plugin 仓库的 plugin/ 子目录，作为可选用插件清单。
//
// 由于 GitHub 对未鉴权请求限速 60/IP/小时，且远端列表变化频率低，
// 这里在进程内做 1h TTL 缓存，并把它作为 SSOT 提供给前端。

const (
	marketUpstreamRepo   = "FloatTech/ZeroBot-Plugin"
	marketUpstreamBranch = "master"
	marketUpstreamPath   = "plugin"
	marketCacheTTL       = time.Hour
	marketHTTPTimeout    = 15 * time.Second
)

type marketEntry struct {
	Name       string `json:"name"`        // 目录名，如 "fortune"
	Path       string `json:"path"`        // 仓库内路径，如 "plugin/fortune"
	HTMLURL    string `json:"html_url"`    // GitHub 上的查看链接
	ImportPath string `json:"import_path"` // 形如 github.com/FloatTech/ZeroBot-Plugin/plugin/fortune
	Loaded     bool   `json:"loaded"`      // 当前进程是否已编译加载
	Enabled    bool   `json:"enabled"`     // 当前进程是否已启用
}

type marketResponse struct {
	Source    string        `json:"source"`
	Repo      string        `json:"repo"`
	Branch    string        `json:"branch"`
	FetchedAt time.Time     `json:"fetched_at"`
	CachedFor time.Duration `json:"-"`
	Items     []marketEntry `json:"items"`
}

type marketCache struct {
	mu        sync.Mutex
	fetchedAt time.Time
	items     []marketEntry
}

var globalMarketCache = &marketCache{}

// githubContentItem 是 GitHub Contents API 返回数组元素的最小子集。
type githubContentItem struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Type    string `json:"type"`
	HTMLURL string `json:"html_url"`
}

func (m *marketCache) fetch(ctx context.Context, force bool) ([]marketEntry, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !force && time.Since(m.fetchedAt) < marketCacheTTL && len(m.items) > 0 {
		return m.items, m.fetchedAt, nil
	}
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/contents/%s?ref=%s",
		marketUpstreamRepo, marketUpstreamPath, marketUpstreamBranch,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, time.Time{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "moebot-next")
	if tok := strings.TrimSpace(githubToken()); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	client := &http.Client{Timeout: marketHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, time.Time{}, fmt.Errorf("github api %s: %s", url, resp.Status)
	}
	var raw []githubContentItem
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, time.Time{}, err
	}
	items := make([]marketEntry, 0, len(raw))
	for _, it := range raw {
		if it.Type != "dir" {
			continue
		}
		items = append(items, marketEntry{
			Name:       it.Name,
			Path:       it.Path,
			HTMLURL:    it.HTMLURL,
			ImportPath: fmt.Sprintf("github.com/%s/%s/%s", marketUpstreamRepo, marketUpstreamPath, it.Name),
		})
	}
	m.items = items
	m.fetchedAt = time.Now()
	return items, m.fetchedAt, nil
}

// handleListMarketPlugins 返回 ZeroBot-Plugin 仓库 plugin/* 子目录列表。
//
// 查询参数：
//   - refresh=1 跳过缓存强刷（仍受 GitHub 限速影响）
func (s *Server) handleListMarketPlugins(c *fiber.Ctx) error {
	force := c.Query("refresh") == "1"
	ctx, cancel := context.WithTimeout(c.UserContext(), marketHTTPTimeout)
	defer cancel()
	items, fetchedAt, err := globalMarketCache.fetch(ctx, force)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, fmt.Sprintf("fetch market list: %v", err))
	}
	// 与本地已注册插件交叉，标记 loaded / enabled。
	loadedByImport := map[string]struct {
		loaded  bool
		enabled bool
	}{}
	if reg := plugin.Global(); reg != nil {
		for _, p := range reg.Plugins() {
			m := p.Manifest()
			// Manifest 没有 import 路径字段，只能按 name 匹配（ZeroBot-Plugin 目录名通常 == control 注册名）。
			loadedByImport[m.Name] = struct {
				loaded  bool
				enabled bool
			}{
				loaded:  reg.IsLoaded(m.Name),
				enabled: reg.IsEnabled(m.Name),
			}
		}
	}
	out := make([]marketEntry, len(items))
	for i, it := range items {
		out[i] = it
		if hit, ok := loadedByImport[it.Name]; ok {
			out[i].Loaded = hit.loaded
			out[i].Enabled = hit.enabled
		}
	}
	return c.JSON(marketResponse{
		Source:    "github-contents-api",
		Repo:      marketUpstreamRepo,
		Branch:    marketUpstreamBranch,
		FetchedAt: fetchedAt,
		Items:     out,
	})
}

// githubToken 允许通过环境变量提供 PAT 以缓解未鉴权 60/h 限速。
func githubToken() string {
	// 仅在显式提供时使用，不写入配置文件。
	for _, key := range []string{"MOEBOT_GITHUB_TOKEN", "GITHUB_TOKEN"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

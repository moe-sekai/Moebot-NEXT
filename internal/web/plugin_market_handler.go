package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"moebot-next/internal/plugin"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// 插件市场：通过 GitHub Contents API 列出
// FloatTech/ZeroBot-Plugin 仓库的 plugin/ 子目录，再叠加 README.md 里
// 手写的 <details> 元数据（中文名、优先级分组、指令列表），
// 使前端能展示有含义的中文标题而不是裸目录名。
//
// 由于 GitHub 对未鉴权请求限速 60/IP/小时，且远端列表变化频率低，
// 这里在进程内做 1h TTL 缓存，并把它作为 SSOT 提供给前端。

const (
	marketUpstreamRepo   = "FloatTech/ZeroBot-Plugin"
	marketUpstreamBranch = "master"
	marketUpstreamPath   = "plugin"
	marketCacheTTL       = time.Hour
	marketHTTPTimeout    = 15 * time.Second
	marketReadmeURL      = "https://raw.githubusercontent.com/FloatTech/ZeroBot-Plugin/master/README.md"
)

type marketEntry struct {
	Name        string   `json:"name"`                  // 目录名，如 "fortune"
	Title       string   `json:"title,omitempty"`       // README <summary> 上的中文名
	Path        string   `json:"path"`                  // 仓库内路径，如 "plugin/fortune"
	HTMLURL     string   `json:"html_url"`              // GitHub 上的查看链接
	ImportPath  string   `json:"import_path"`           // 形如 github.com/FloatTech/ZeroBot-Plugin/plugin/fortune
	Priority    string   `json:"priority,omitempty"`    // high | medium | low（来自 README h3 分组）
	Description string   `json:"description,omitempty"` // <details> 块中 import 之后、指令之前的自然语言描述（如有）
	Commands    []string `json:"commands,omitempty"`    // README 中 "- [x] xxx" 提取出的指令摘要
	Source      string   `json:"source,omitempty"`      // "zerobot-plugin" | "zbputils"
	Loaded      bool     `json:"loaded"`                // 当前进程是否已编译加载
	Enabled     bool     `json:"enabled"`               // 当前进程是否已启用
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
			Source:     "zerobot-plugin",
		})
	}
	// 叠加 README 元数据（中文名、指令列表、优先级）。
	// README 抓取失败不阻塞主流程，仅打日志。
	readmeMeta, rerr := fetchReadmeMeta(ctx)
	if rerr != nil {
		log.Warn().Err(rerr).Msg("market: fetch README metadata failed, falling back to dir-only list")
	} else {
		byKey := make(map[string]*readmeEntry, len(readmeMeta))
		for i := range readmeMeta {
			byKey[normalizeName(readmeMeta[i].Name)] = &readmeMeta[i]
		}
		seen := make(map[string]struct{}, len(items))
		for i := range items {
			seen[normalizeName(items[i].Name)] = struct{}{}
			if meta, ok := byKey[normalizeName(items[i].Name)]; ok {
				applyReadmeMeta(&items[i], meta)
			}
		}
		// README 里提到但 Contents API 没返回的条目（例如 zbputils/* 下的 job/chat/vevent），
		// 作为独立条目追加，让用户也能在市场中看到并跳转。
		for i := range readmeMeta {
			m := &readmeMeta[i]
			if _, ok := seen[normalizeName(m.Name)]; ok {
				continue
			}
			e := marketEntry{
				Name:       m.Name,
				Path:       m.ImportPath, // 非仓库路径，仅用于展示链接一致性
				HTMLURL:    "https://" + m.ImportPath,
				ImportPath: m.ImportPath,
				Source:     m.Source,
			}
			applyReadmeMeta(&e, m)
			items = append(items, e)
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		pi, pj := priorityRank(items[i].Priority), priorityRank(items[j].Priority)
		if pi != pj {
			return pi < pj
		}
		return items[i].Name < items[j].Name
	})
	m.items = items
	m.fetchedAt = time.Now()
	return items, m.fetchedAt, nil
}

// normalizeName 在匹配 README 导入路径与 Contents API 目录名时去掉下划线并统一小写，
// 以容忍 README 里的 import 与实际目录名拼写差异（例如 ai_false vs aifalse）。
func normalizeName(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", ""))
}

func priorityRank(p string) int {
	switch p {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	default:
		return 3
	}
}

func applyReadmeMeta(e *marketEntry, m *readmeEntry) {
	if e.Title == "" {
		e.Title = m.Title
	}
	if e.Priority == "" {
		e.Priority = m.Priority
	}
	if e.Description == "" {
		e.Description = m.Description
	}
	if len(e.Commands) == 0 {
		e.Commands = m.Commands
	}
	if e.Source == "" {
		e.Source = m.Source
	}
	// 用 README 里的 import 路径更正 Contents API 合成的那一份（以防 README 与目录名拼写不一致）。
	if m.ImportPath != "" {
		e.ImportPath = m.ImportPath
	}
}

// readmeEntry 是从 ZeroBot-Plugin README 的一个 <details> 块里解析出的原始元数据。
type readmeEntry struct {
	Name        string // 从 import 路径末段提取的目录名（如 fortune）
	Title       string // <summary> 上的中文名
	ImportPath  string // 形如 github.com/FloatTech/ZeroBot-Plugin/plugin/fortune
	Source      string // "zerobot-plugin" | "zbputils"
	Priority    string // high | medium | low
	Description string
	Commands    []string
}

var (
	reDetailsOpen  = regexp.MustCompile(`(?i)<details>`)
	reDetailsClose = regexp.MustCompile(`(?i)</details>`)
	reSummary      = regexp.MustCompile(`(?is)<summary>\s*(.*?)\s*</summary>`)
	reImport       = regexp.MustCompile("`?import\\s+_\\s+\"(github\\.com/FloatTech/(?:ZeroBot-Plugin/plugin|zbputils)/[A-Za-z0-9_/\\-]+)\"")
	reCommandLine  = regexp.MustCompile(`^\s*-\s*\[[ xX]\]\s*(.+?)\s*$`)
	reHeader3      = regexp.MustCompile(`^\s*###\s*(.+?)\s*$`)
)

// fetchReadmeMeta 抓取并解析 ZeroBot-Plugin README.md，按 <details> 块提取每个插件的元数据。
// 未鉴权 raw.githubusercontent.com 不走 API 限速，但会受 CDN 缓存影响。
func fetchReadmeMeta(ctx context.Context) ([]readmeEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, marketReadmeURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "moebot-next")
	client := &http.Client{Timeout: marketHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("readme %s: %s", marketReadmeURL, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseReadme(string(body)), nil
}

// parseReadme 将 ZeroBot-Plugin README 正文切成 <details>...</details> 块并解析。
// 不做完整 Markdown 解析：仅靠少量正则就能覆盖上游手写的固定格式。
func parseReadme(text string) []readmeEntry {
	out := make([]readmeEntry, 0, 128)
	// 优先级分组靠扫描 H3 行跟踪，切块时再赋给当前块。
	lines := strings.Split(text, "\n")
	priority := ""
	var inBlock bool
	var blockLines []string
	flush := func() {
		if !inBlock {
			return
		}
		body := strings.Join(blockLines, "\n")
		blockLines = blockLines[:0]
		inBlock = false
		entry := readmeEntry{Priority: priority}
		if m := reSummary.FindStringSubmatch(body); len(m) == 2 {
			entry.Title = strings.TrimSpace(stripMarkdownInline(m[1]))
		}
		if m := reImport.FindStringSubmatch(body); len(m) == 2 {
			entry.ImportPath = m[1]
			// 末段作为 name 键（例如 plugin/fortune -> fortune；zbputils/job -> job）。
			parts := strings.Split(m[1], "/")
			entry.Name = parts[len(parts)-1]
			if strings.Contains(m[1], "/zbputils/") {
				entry.Source = "zbputils"
			} else {
				entry.Source = "zerobot-plugin"
			}
		}
		// 指令列表 & 描述文本。
		// 指令：`- [x] xxx`；描述：import 之后、第一条指令之前的非空段落（如果有）。
		var desc []string
		seenCmd := false
		for _, ln := range strings.Split(body, "\n") {
			if m := reCommandLine.FindStringSubmatch(ln); len(m) == 2 {
				cmd := strings.TrimSpace(stripMarkdownInline(m[1]))
				if cmd != "" {
					entry.Commands = append(entry.Commands, cmd)
				}
				seenCmd = true
				continue
			}
			if seenCmd {
				continue
			}
			t := strings.TrimSpace(ln)
			if t == "" || strings.HasPrefix(t, "<summary") || strings.HasPrefix(t, "</summary") ||
				strings.HasPrefix(t, "`import") || strings.HasPrefix(t, "<details") {
				continue
			}
			// 忽略 ` ` 反引号包裹的 import 行变体
			if strings.HasPrefix(t, "`") && strings.Contains(t, "import") {
				continue
			}
			desc = append(desc, t)
		}
		if len(desc) > 0 {
			joined := strings.Join(desc, " ")
			joined = strings.TrimSpace(stripMarkdownInline(joined))
			if len([]rune(joined)) > 160 {
				joined = string([]rune(joined)[:160]) + "…"
			}
			entry.Description = joined
		}
		if entry.Name != "" {
			out = append(out, entry)
		}
	}
	for _, ln := range lines {
		if m := reHeader3.FindStringSubmatch(ln); len(m) == 2 {
			title := strings.TrimSpace(strings.Trim(m[1], "*"))
			switch title {
			case "高优先级":
				priority = "high"
			case "中优先级":
				priority = "medium"
			case "低优先级":
				priority = "low"
			}
		}
		if !inBlock {
			if reDetailsOpen.MatchString(ln) {
				inBlock = true
				blockLines = append(blockLines, ln)
			}
			continue
		}
		blockLines = append(blockLines, ln)
		if reDetailsClose.MatchString(ln) {
			flush()
		}
	}
	flush()
	return out
}

// stripMarkdownInline 去掉常见的内联 Markdown 修饰，让 summary / description 清爽一些。
func stripMarkdownInline(s string) string {
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "`", "")
	return s
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

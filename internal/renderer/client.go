package renderer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"moebot-next/internal/config"

	"github.com/rs/zerolog/log"
)

// Client communicates with the Bun renderer microservice.
type Client struct {
	baseURL    string
	httpClient *http.Client
	process    *exec.Cmd
	precision  float64
	cache      config.CacheConfig
}

// RenderRequest is sent to the renderer service.
type RenderRequest struct {
	Template  string      `json:"template"` // e.g. "card_detail", "music_detail"
	Data      interface{} `json:"data"`
	Width     int         `json:"width,omitempty"`
	Height    int         `json:"height,omitempty"`
	Precision float64     `json:"precision,omitempty"`
}

// PreviewMeta describes a sample Satori template exposed by the renderer.
type PreviewMeta struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Command      string `json:"command"`
	TemplatePath string `json:"templatePath"`
	ViewerSource string `json:"viewerSource"`
	Status       string `json:"status"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

// PreviewListResponse is returned by the renderer preview listing endpoint.
type PreviewListResponse struct {
	Data  []PreviewMeta `json:"data"`
	Total int           `json:"total"`
}

// PreviewRenderResult contains a rendered preview image and renderer timing metadata.
type PreviewRenderResult struct {
	PNG              []byte
	TotalMS          string
	FontsMS          string
	ImagesMS         string
	SatoriMS         string
	ResvgMS          string
	SizeBytes        string
	ImageTotal       string
	ImageRemote      string
	ImageCacheHits   string
	ImageCacheMisses string
	ImageCacheErrors string
	StatusCode       int
}

// AssetPreloadStatus mirrors the renderer-side card thumbnail cache status payload.
type AssetPreloadStatus struct {
	OK          bool     `json:"ok"`
	Enabled     bool     `json:"enabled"`
	Running     bool     `json:"running"`
	Message     string   `json:"message"`
	CacheDir    string   `json:"cache_dir"`
	Total       int      `json:"total"`
	Cached      int      `json:"cached"`
	Missing     int      `json:"missing"`
	Failed      int      `json:"failed"`
	Downloaded  int      `json:"downloaded"`
	Skipped     int      `json:"skipped"`
	Progress    float64  `json:"progress"`
	StartedAt   *string  `json:"started_at"`
	CompletedAt *string  `json:"completed_at"`
	Errors      []string `json:"errors"`
}

type assetPreloadRequest struct {
	URLs        []string `json:"urls"`
	Force       bool     `json:"force,omitempty"`
	Concurrency int      `json:"concurrency,omitempty"`
}

// New creates a new renderer client.
func New(cfg config.RendererConfig) *Client {
	precision := cfg.Precision
	if precision <= 0 {
		precision = config.DefaultRendererPrecision
	}
	return &Client{
		baseURL:   fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
		precision: precision,
		cache:     cfg.Cache,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StartProcess starts the Bun renderer as a child process.
func (c *Client) StartProcess(rendererDir string, port int) error {
	// Check if bun is available. The renderer intentionally uses Bun so it can
	// execute TypeScript/TSX Satori templates directly without a build step.
	bunPath, err := exec.LookPath("bun")
	if err != nil {
		return fmt.Errorf("bun not found in PATH; please install Bun or run via Docker")
	}

	cmd := exec.Command(bunPath, "run", "src/server.tsx")
	cmd.Dir = rendererDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%d", port),
		fmt.Sprintf("RENDER_PRECISION=%g", c.precision),
	)
	cmd.Env = append(cmd.Env, rendererCacheEnv(c.cache)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start renderer process: %w", err)
	}

	c.process = cmd
	log.Info().
		Str("runtime", bunPath).
		Int("port", port).
		Msg("Renderer process started")

	// Wait for the renderer to be ready
	return c.waitForReady(10 * time.Second)
}

// waitForReady polls the renderer's health endpoint.
func (c *Client) waitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := c.httpClient.Get(c.baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			log.Info().Msg("Renderer service is ready")
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("renderer did not become ready within %v", timeout)
}

// Render sends a render request and returns PNG bytes.
func (c *Client) Render(req RenderRequest) ([]byte, error) {
	result, err := c.RenderWithTrace(req)
	if err != nil {
		return nil, err
	}
	return result.PNG, nil
}

// RenderWithTrace sends a render request and preserves renderer timing headers.
func (c *Client) RenderWithTrace(req RenderRequest) (*PreviewRenderResult, error) {
	if req.Precision <= 0 {
		req.Precision = c.precision
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal render request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/render", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("render request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("renderer returned %d: %s", resp.StatusCode, string(errBody))
	}

	png, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rendered image: %w", err)
	}

	return &PreviewRenderResult{
		PNG:              png,
		TotalMS:          resp.Header.Get("x-render-total-ms"),
		FontsMS:          resp.Header.Get("x-render-fonts-ms"),
		ImagesMS:         resp.Header.Get("x-render-images-ms"),
		SatoriMS:         resp.Header.Get("x-render-satori-ms"),
		ResvgMS:          resp.Header.Get("x-render-resvg-ms"),
		SizeBytes:        resp.Header.Get("x-render-size-bytes"),
		ImageTotal:       resp.Header.Get("x-render-image-total"),
		ImageRemote:      resp.Header.Get("x-render-image-remote"),
		ImageCacheHits:   resp.Header.Get("x-render-image-cache-hits"),
		ImageCacheMisses: resp.Header.Get("x-render-image-cache-misses"),
		ImageCacheErrors: resp.Header.Get("x-render-image-cache-errors"),
		StatusCode:       resp.StatusCode,
	}, nil
}

// StartCardThumbnailPreload asks the renderer service to preload card thumbnail URLs in the background.
func (c *Client) StartCardThumbnailPreload(urls []string) (*AssetPreloadStatus, error) {
	var result AssetPreloadStatus
	err := c.postJSON("/cache/card-thumbnails/preload", assetPreloadRequest{URLs: urls}, &result)
	return &result, err
}

// CardThumbnailPreloadStatus returns cache coverage for the given card thumbnail URLs.
func (c *Client) CardThumbnailPreloadStatus(urls []string) (*AssetPreloadStatus, error) {
	var result AssetPreloadStatus
	err := c.postJSON("/cache/card-thumbnails/status", assetPreloadRequest{URLs: urls}, &result)
	return &result, err
}

func (c *Client) postJSON(path string, payload interface{}, out interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal renderer request: %w", err)
	}
	resp, err := c.httpClient.Post(c.baseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("renderer request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("renderer returned %d: %s", resp.StatusCode, string(errBody))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode renderer response: %w", err)
	}
	return nil
}

// ListPreviews fetches Satori preview metadata from the renderer service.
func (c *Client) ListPreviews() ([]PreviewMeta, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/previews")
	if err != nil {
		return nil, fmt.Errorf("preview list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("renderer returned %d: %s", resp.StatusCode, string(errBody))
	}

	var result PreviewListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode preview list: %w", err)
	}
	return result.Data, nil
}

// RenderPreview renders a named Satori preview template to PNG.
func (c *Client) RenderPreview(id string, width int, height int) ([]byte, error) {
	result, err := c.RenderPreviewWithTrace(id, width, height)
	if err != nil {
		return nil, err
	}
	return result.PNG, nil
}

// RenderPreviewWithTrace renders a preview and preserves renderer timing headers.
func (c *Client) RenderPreviewWithTrace(id string, width int, height int) (*PreviewRenderResult, error) {
	previewURL := c.baseURL + "/preview/" + url.PathEscape(id)
	query := url.Values{}
	if width > 0 {
		query.Set("width", strconv.Itoa(width))
	}
	if height > 0 {
		query.Set("height", strconv.Itoa(height))
	}
	if c.precision > 0 {
		query.Set("precision", strconv.FormatFloat(c.precision, 'f', -1, 64))
	}
	if encoded := query.Encode(); encoded != "" {
		previewURL += "?" + encoded
	}

	resp, err := c.httpClient.Get(previewURL)
	if err != nil {
		return nil, fmt.Errorf("preview render request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("renderer returned %d: %s", resp.StatusCode, string(errBody))
	}

	png, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read preview image: %w", err)
	}

	return &PreviewRenderResult{
		PNG:              png,
		TotalMS:          resp.Header.Get("x-render-total-ms"),
		FontsMS:          resp.Header.Get("x-render-fonts-ms"),
		ImagesMS:         resp.Header.Get("x-render-images-ms"),
		SatoriMS:         resp.Header.Get("x-render-satori-ms"),
		ResvgMS:          resp.Header.Get("x-render-resvg-ms"),
		SizeBytes:        resp.Header.Get("x-render-size-bytes"),
		ImageTotal:       resp.Header.Get("x-render-image-total"),
		ImageRemote:      resp.Header.Get("x-render-image-remote"),
		ImageCacheHits:   resp.Header.Get("x-render-image-cache-hits"),
		ImageCacheMisses: resp.Header.Get("x-render-image-cache-misses"),
		ImageCacheErrors: resp.Header.Get("x-render-image-cache-errors"),
		StatusCode:       resp.StatusCode,
	}, nil
}

// BaseURL returns the renderer service base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// SetPrecision updates the SVG to PNG render scale used for future requests.
func (c *Client) SetPrecision(precision float64) {
	if precision <= 0 {
		precision = config.DefaultRendererPrecision
	}
	c.precision = precision
}

// Precision returns the current SVG to PNG render scale.
func (c *Client) Precision() float64 {
	if c.precision <= 0 {
		return config.DefaultRendererPrecision
	}
	return c.precision
}

func rendererCacheEnv(cache config.CacheConfig) []string {
	cachePath := cache.Path
	if cachePath == "" {
		cachePath = "./data/cache"
	}
	if absPath, err := filepath.Abs(cachePath); err == nil {
		cachePath = absPath
	}
	maxSizeBytes := int64(cache.MaxSizeMB) * 1024 * 1024
	if cache.MaxSizeMB <= 0 {
		maxSizeBytes = 0
	}
	ttlMS := int64(cache.TTLHours) * 60 * 60 * 1000
	if cache.TTLHours <= 0 {
		ttlMS = 0
	}
	return []string{
		"RENDER_CACHE_ENABLED=" + strconv.FormatBool(cache.Enabled),
		"RENDER_CACHE_DIR=" + cachePath,
		"RENDER_CACHE_MAX_SIZE_BYTES=" + strconv.FormatInt(maxSizeBytes, 10),
		"RENDER_CACHE_TTL_MS=" + strconv.FormatInt(ttlMS, 10),
	}
}

// Health checks if the renderer service is alive.
func (c *Client) Health() bool {
	ok, _, _ := c.HealthWithTimeout(3 * time.Second)
	return ok
}

// HealthWithTimeout checks the renderer health endpoint with a caller-provided timeout.
func (c *Client) HealthWithTimeout(timeout time.Duration) (bool, int, error) {
	client := c.httpClient
	if timeout > 0 {
		client = &http.Client{Timeout: timeout}
	}

	resp, err := client.Get(c.baseURL + "/health")
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, resp.StatusCode, nil
}

// StopProcess gracefully stops the renderer child process.
func (c *Client) StopProcess() error {
	if c.process == nil || c.process.Process == nil {
		return nil
	}
	log.Info().Msg("Stopping renderer process")
	return c.process.Process.Kill()
}

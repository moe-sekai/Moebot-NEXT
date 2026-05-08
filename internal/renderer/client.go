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

const defaultRendererRequestTimeout = 2 * time.Minute
const DefaultChartRenderWidth = 2400

// Client communicates with the Bun renderer microservice.
type Client struct {
	baseURL        string
	httpClient     *http.Client
	process        *exec.Cmd
	precision      float64
	chartPrecision float64
	cache          config.CacheConfig
	fonts          config.RendererFontConfig
}

// RenderRequest is sent to the renderer service.
type RenderRequest struct {
	Template  string      `json:"template"` // e.g. "card_detail", "music_detail"
	Data      interface{} `json:"data"`
	Width     int         `json:"width,omitempty"`
	Height    int         `json:"height,omitempty"`
	Precision float64     `json:"precision,omitempty"`
}

// ChartRenderRequest asks the renderer to convert one remote chart SVG into PNG directly.
type ChartRenderRequest struct {
	URL       string  `json:"url"`
	Width     int     `json:"width,omitempty"`
	Precision float64 `json:"precision,omitempty"`
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
	OK                        bool     `json:"ok"`
	Enabled                   bool     `json:"enabled"`
	Running                   bool     `json:"running"`
	Message                   string   `json:"message"`
	CacheDir                  string   `json:"cache_dir"`
	Total                     int      `json:"total"`
	Cached                    int      `json:"cached"`
	Missing                   int      `json:"missing"`
	Failed                    int      `json:"failed"`
	Downloaded                int      `json:"downloaded"`
	Skipped                   int      `json:"skipped"`
	Progress                  float64  `json:"progress"`
	StartedAt                 *string  `json:"started_at"`
	CompletedAt               *string  `json:"completed_at"`
	Errors                    []string `json:"errors"`
	CompositeTotal            int      `json:"composite_total"`
	CompositeCached           int      `json:"composite_cached"`
	CompositeMissing          int      `json:"composite_missing"`
	CompositeFailed           int      `json:"composite_failed"`
	CompositeGenerated        int      `json:"composite_generated"`
	CompositeProgress         float64  `json:"composite_progress"`
	CompositeSourceDownloaded int      `json:"composite_source_downloaded"`
	CompositeSourceFailed     int      `json:"composite_source_failed"`
	CompositeRenderMS         int      `json:"composite_render_ms"`
}

type CardThumbnailPreloadCard struct {
	ImageURL string `json:"imageUrl,omitempty"`
	Rarity   string `json:"rarity,omitempty"`
	Attr     string `json:"attr,omitempty"`
	Trained  bool   `json:"trained,omitempty"`
	Size     int    `json:"size,omitempty"`
}

type assetPreloadRequest struct {
	URLs        []string                   `json:"urls"`
	Cards       []CardThumbnailPreloadCard `json:"cards,omitempty"`
	Force       bool                       `json:"force,omitempty"`
	Concurrency int                        `json:"concurrency,omitempty"`
}

// New creates a new renderer client.
func New(cfg config.RendererConfig) *Client {
	precision := cfg.Precision
	if precision <= 0 {
		precision = config.DefaultRendererPrecision
	}
	chartPrecision := cfg.ChartPrecision
	if chartPrecision <= 0 {
		chartPrecision = config.DefaultChartRendererPrecision
	}
	return &Client{
		baseURL:        fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
		precision:      precision,
		chartPrecision: chartPrecision,
		cache:          cfg.Cache,
		fonts:          cfg.Fonts,
		httpClient: &http.Client{
			Timeout: defaultRendererRequestTimeout,
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
	cmd.Env = append(cmd.Env, rendererFontEnv(c.fonts)...)
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

// RenderChartURL converts a chart SVG URL to PNG through the renderer's resvg pipeline.
func (c *Client) RenderChartURL(chartURL string) ([]byte, error) {
	result, err := c.RenderChartURLWithTrace(chartURL, 0)
	if err != nil {
		return nil, err
	}
	return result.PNG, nil
}

// RenderChartURLWithTrace converts a chart SVG URL to PNG and preserves renderer timing headers.
func (c *Client) RenderChartURLWithTrace(chartURL string, width int) (*PreviewRenderResult, error) {
	if width <= 0 {
		width = DefaultChartRenderWidth
	}
	req := ChartRenderRequest{URL: chartURL, Width: width, Precision: c.ChartPrecision()}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chart render request: %w", err)
	}
	resp, err := c.httpClient.Post(c.baseURL+"/render/chart", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("chart render request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("renderer returned %d: %s", resp.StatusCode, string(errBody))
	}
	png, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rendered chart image: %w", err)
	}
	return &PreviewRenderResult{
		PNG:        png,
		TotalMS:    resp.Header.Get("x-render-total-ms"),
		ResvgMS:    resp.Header.Get("x-render-resvg-ms"),
		SizeBytes:  resp.Header.Get("x-render-size-bytes"),
		StatusCode: resp.StatusCode,
	}, nil
}

// StartCardThumbnailPreload asks the renderer service to preload card thumbnail URLs in the background.
func (c *Client) StartCardThumbnailPreload(urls []string) (*AssetPreloadStatus, error) {
	return c.StartCardThumbnailPreloadWithCards(urls, nil)
}

// StartCardThumbnailPreloadWithCards preloads source images and renderer-side precomposited card tiles.
func (c *Client) StartCardThumbnailPreloadWithCards(urls []string, cards []CardThumbnailPreloadCard) (*AssetPreloadStatus, error) {
	var result AssetPreloadStatus
	err := c.postJSON("/cache/card-thumbnails/preload", assetPreloadRequest{URLs: urls, Cards: cards}, &result)
	return &result, err
}

// CardThumbnailPreloadStatus returns cache coverage for the given card thumbnail URLs.
func (c *Client) CardThumbnailPreloadStatus(urls []string) (*AssetPreloadStatus, error) {
	return c.CardThumbnailPreloadStatusWithCards(urls, nil)
}

// CardThumbnailPreloadStatusWithCards returns source image and precomposited tile cache coverage.
func (c *Client) CardThumbnailPreloadStatusWithCards(urls []string, cards []CardThumbnailPreloadCard) (*AssetPreloadStatus, error) {
	var result AssetPreloadStatus
	err := c.postJSON("/cache/card-thumbnails/status", assetPreloadRequest{URLs: urls, Cards: cards}, &result)
	return &result, err
}

// CalculateDeckRecommend runs the embedded TypeScript deck recommender in the renderer service.
func (c *Client) CalculateDeckRecommend(req DeckRecommendCalculateRequest) (*DeckRecommendCalculateResponse, error) {
	var result DeckRecommendCalculateResponse
	err := c.postJSON("/deck-recommend/calculate", req, &result)
	if err != nil {
		return nil, err
	}
	if !result.OK {
		if result.Error != "" {
			return &result, fmt.Errorf("%s", result.Error)
		}
		return &result, fmt.Errorf("deck recommend failed")
	}
	return &result, nil
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

// FontEntry describes a single font face loaded by the renderer.
type FontEntry struct {
	Name   string `json:"name"`
	Weight int    `json:"weight"`
	Style  string `json:"style"`
}

// FontDefaults contains the default fontFamily CSS strings used for rendering.
type FontDefaults struct {
	Body  string `json:"body"`
	Score string `json:"score"`
}

// FontConfig contains named font family constants.
type FontConfig struct {
	Score        string `json:"score"`
	Body         string `json:"body"`
	BodyFallback string `json:"bodyFallback"`
	Decorative   string `json:"decorative"`
}

// FontsResponse is the response from the renderer /fonts endpoint.
type FontsResponse struct {
	OK       bool         `json:"ok"`
	Fonts    []FontEntry  `json:"fonts"`
	Families []string     `json:"families"`
	Defaults FontDefaults `json:"defaults"`
	Config   FontConfig   `json:"config"`
	Total    int          `json:"total"`
	Message  string       `json:"message,omitempty"`
}

// GetFonts fetches font information from the renderer service.
func (c *Client) GetFonts() (*FontsResponse, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/fonts")
	if err != nil {
		return nil, fmt.Errorf("fonts request failed: %w", err)
	}
	defer resp.Body.Close()

	var result FontsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode fonts response: %w", err)
	}
	return &result, nil
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

// SetChartPrecision updates the chart SVG to PNG render scale used for future requests.
func (c *Client) SetChartPrecision(precision float64) {
	if precision <= 0 {
		precision = config.DefaultChartRendererPrecision
	}
	c.chartPrecision = precision
}

// Precision returns the current SVG to PNG render scale.
func (c *Client) Precision() float64 {
	if c.precision <= 0 {
		return config.DefaultRendererPrecision
	}
	return c.precision
}

// ChartPrecision returns the current chart SVG to PNG render scale.
func (c *Client) ChartPrecision() float64 {
	if c.chartPrecision <= 0 {
		return config.DefaultChartRendererPrecision
	}
	return c.chartPrecision
}

// UpdateFontsRequest is the body of POST /fonts on the renderer.
type UpdateFontsRequest struct {
	Body  string `json:"body,omitempty"`
	Score string `json:"score,omitempty"`
}

// UpdateFontsResponse mirrors the renderer-side POST /fonts response.
type UpdateFontsResponse struct {
	OK          bool         `json:"ok"`
	Defaults    FontDefaults `json:"defaults"`
	Preferences struct {
		Body  string `json:"body"`
		Score string `json:"score"`
	} `json:"preferences"`
	Message string `json:"message,omitempty"`
}

// UpdateFonts pushes the given font preferences to the running renderer
// without restarting it. Empty values keep the current preference.
func (c *Client) UpdateFonts(body, score string) (*UpdateFontsResponse, error) {
	payload, err := json.Marshal(UpdateFontsRequest{Body: body, Score: score})
	if err != nil {
		return nil, fmt.Errorf("marshal fonts request: %w", err)
	}
	resp, err := c.httpClient.Post(c.baseURL+"/fonts", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("fonts update request failed: %w", err)
	}
	defer resp.Body.Close()
	var result UpdateFontsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode fonts update response: %w", err)
	}
	if resp.StatusCode != http.StatusOK || !result.OK {
		msg := result.Message
		if msg == "" {
			msg = fmt.Sprintf("renderer returned %d", resp.StatusCode)
		}
		return &result, fmt.Errorf("renderer rejected fonts update: %s", msg)
	}
	return &result, nil
}

// SetFonts updates the cached font preferences (used on next process restart)
// and pushes them to the running renderer.
func (c *Client) SetFonts(fonts config.RendererFontConfig) error {
	c.fonts = fonts
	if _, err := c.UpdateFonts(fonts.BodyFamily, fonts.ScoreFamily); err != nil {
		return err
	}
	return nil
}

// Fonts returns the currently cached font preferences.
func (c *Client) Fonts() config.RendererFontConfig {
	return c.fonts
}

func rendererFontEnv(fonts config.RendererFontConfig) []string {
	env := []string{}
	if fonts.BodyFamily != "" {
		env = append(env, "RENDER_FONT_BODY="+fonts.BodyFamily)
	}
	if fonts.ScoreFamily != "" {
		env = append(env, "RENDER_FONT_SCORE="+fonts.ScoreFamily)
	}
	return env
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

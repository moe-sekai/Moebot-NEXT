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
}

// RenderRequest is sent to the renderer service.
type RenderRequest struct {
	Template string      `json:"template"` // e.g. "card_detail", "music_detail"
	Data     interface{} `json:"data"`
	Width    int         `json:"width,omitempty"`
	Height   int         `json:"height,omitempty"`
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

// New creates a new renderer client.
func New(cfg config.RendererConfig) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
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
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port))
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

	return io.ReadAll(resp.Body)
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
	previewURL := c.baseURL + "/preview/" + url.PathEscape(id)
	query := url.Values{}
	if width > 0 {
		query.Set("width", strconv.Itoa(width))
	}
	if height > 0 {
		query.Set("height", strconv.Itoa(height))
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

	return io.ReadAll(resp.Body)
}

// BaseURL returns the renderer service base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
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

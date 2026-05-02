package sekai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"moebot-next/internal/config"
)

// Client talks to the configured SEKAI API endpoint.
type Client struct {
	enabled    bool
	baseURL    string
	region     string
	headers    map[string]string
	timeout    time.Duration
	httpClient *http.Client
}

func NewClient(cfg config.SekaiAPIConfig) *Client {
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	headers := make(map[string]string, len(cfg.Headers))
	for key, value := range cfg.Headers {
		if strings.TrimSpace(key) != "" && value != "" {
			headers[key] = value
		}
	}
	return &Client{
		enabled:    cfg.Enabled,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		region:     strings.Trim(strings.ToLower(cfg.Region), "/"),
		headers:    headers,
		timeout:    timeout,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.enabled && c.baseURL != "" && c.region != ""
}

func (c *Client) GetProfile(userID string) (*Profile, error) {
	if c == nil || !c.Enabled() {
		return nil, fmt.Errorf("sekai api is disabled")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user id is empty")
	}

	endpoint, err := url.JoinPath(c.baseURL, "api", c.region, userID, "profile")
	if err != nil {
		return nil, fmt.Errorf("build profile url: %w", err)
	}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build profile request: %w", err)
	}
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("profile request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("profile request returned %d", resp.StatusCode)
	}

	var result profileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode profile response: %w", err)
	}
	profile := result.normalize(userID)
	return &profile, nil
}

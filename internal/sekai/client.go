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
		region:     config.NormalizeRegion(strings.Trim(strings.ToLower(cfg.Region), "/")),
		headers:    headers,
		timeout:    timeout,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.enabled && c.baseURL != "" && c.region != ""
}

func (c *Client) profileURL(userID string) (string, error) {
	base := strings.ReplaceAll(c.baseURL, "{region}", c.region)
	base = strings.ReplaceAll(base, "{uid}", url.PathEscape(userID))
	if strings.Contains(base, "{user_id}") {
		base = strings.ReplaceAll(base, "{user_id}", url.PathEscape(userID))
	}
	if strings.Contains(base, "{uid}") || strings.Contains(base, "{user_id}") || strings.Contains(base, "{region}") {
		return "", fmt.Errorf("unresolved sekai api placeholder in %q", base)
	}
	if strings.Contains(c.baseURL, "{uid}") || strings.Contains(c.baseURL, "{user_id}") {
		return base, nil
	}
	if strings.Contains(c.baseURL, "{region}") {
		return url.JoinPath(base, userID, "profile")
	}
	return url.JoinPath(base, "api", c.region, userID, "profile")
}

func (c *Client) systemURL() (string, error) {
	base := strings.ReplaceAll(c.baseURL, "{region}", c.region)
	if strings.Contains(base, "{uid}") || strings.Contains(base, "{user_id}") {
		return "", fmt.Errorf("system warmup does not support uid endpoint templates")
	}
	if strings.Contains(base, "{region}") {
		return "", fmt.Errorf("unresolved sekai api region placeholder in %q", base)
	}
	if strings.Contains(c.baseURL, "{region}") {
		return strings.TrimRight(base, "/") + "/system", nil
	}
	return url.JoinPath(base, "api", c.region, "system")
}

func (c *Client) newGETRequest(endpoint string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func (c *Client) doGET(endpoint string) (*http.Response, error) {
	req, err := c.newGETRequest(endpoint)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}

func (c *Client) warmupSystem() error {
	endpoint, err := c.systemURL()
	if err != nil {
		return err
	}
	resp, err := c.doGET(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("system request returned %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) GetProfile(userID string) (*Profile, error) {
	if c == nil || !c.Enabled() {
		return nil, fmt.Errorf("sekai api is disabled")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user id is empty")
	}

	endpoint, err := c.profileURL(userID)
	if err != nil {
		return nil, fmt.Errorf("build profile url: %w", err)
	}
	resp, err := c.doGET(endpoint)
	if err != nil {
		return nil, fmt.Errorf("profile request failed: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		if err := c.warmupSystem(); err == nil {
			resp, err = c.doGET(endpoint)
			if err != nil {
				return nil, fmt.Errorf("profile retry after system warmup failed: %w", err)
			}
		}
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

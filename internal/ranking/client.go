package ranking

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"moebot-next/internal/config"
)

type Client struct {
	baseURL string
	region  string
	http    *http.Client
}

func NewClient(cfg Config) *Client {
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://rks.exmeaning.com"
	}
	region := config.NormalizeRegion(strings.Trim(strings.ToLower(cfg.Region), "/"))
	if region == "" {
		region = config.RegionCN
	}
	return &Client{baseURL: baseURL, region: region, http: &http.Client{Timeout: timeout}}
}

func (c *Client) GetLatest() (*Board, error) {
	return c.getBoard("latest")
}

func (c *Client) GetChurn() (*Board, error) {
	return c.getBoard("churn")
}

func (c *Client) getBoard(path string) (*Board, error) {
	endpoint, err := url.JoinPath(c.baseURL, "api", "public", rankingRegionPath(c.region), path)
	if err != nil {
		return nil, fmt.Errorf("build ranking url: %w", err)
	}
	resp, err := c.http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("ranking request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ranking request returned %d", resp.StatusCode)
	}
	var board Board
	if err := json.NewDecoder(resp.Body).Decode(&board); err != nil {
		return nil, fmt.Errorf("decode ranking response: %w", err)
	}
	return &board, nil
}

func rankingRegionPath(region string) string {
	if config.NormalizeRegion(region) == config.RegionTW {
		return "tw"
	}
	return config.NormalizeRegion(region)
}

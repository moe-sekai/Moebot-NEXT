package ranking

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"moebot-next/internal/config"
)

var ErrNoWorldLinkData = errors.New("no worldlink data available")
var ErrForecastUnsupportedRegion = errors.New("forecast only supports cn and jp")

type Client struct {
	baseURL     string
	forecastURL string
	region      string
	http        *http.Client
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
	return &Client{
		baseURL:     baseURL,
		forecastURL: "https://rk.exmeaning.com/public",
		region:      region,
		http:        &http.Client{Timeout: timeout},
	}
}

func (c *Client) Region() string {
	return c.region
}

func (c *Client) GetLatest() (*Board, error) {
	return c.getBoard("latest", nil)
}

func (c *Client) GetChurn() (*Board, error) {
	return c.getBoard("churn", nil)
}

func (c *Client) GetWorldLinkLatest() (*WorldLinkBoard, error) {
	endpoint, err := c.publicEndpoint("worldlink-latest", nil)
	if err != nil {
		return nil, err
	}
	var board WorldLinkBoard
	if err := c.getJSON(endpoint, &board); err != nil {
		if isNoWorldLinkError(err) {
			return nil, ErrNoWorldLinkData
		}
		return nil, err
	}
	return &board, nil
}

func (c *Client) GetWorldLinkChurn(gameCharacterID int) (*Board, error) {
	query := url.Values{}
	if gameCharacterID > 0 {
		query.Set("gameCharacterId", fmt.Sprintf("%d", gameCharacterID))
	}
	board, err := c.getBoard("worldlink-churn", query)
	if err != nil {
		if isNoWorldLinkError(err) {
			return nil, ErrNoWorldLinkData
		}
		return nil, err
	}
	board.BoardType = "worldlink"
	board.TargetID = gameCharacterID
	return board, nil
}

func (c *Client) GetForecastEvents() ([]ForecastEvent, error) {
	if !c.supportsForecast() {
		return nil, ErrForecastUnsupportedRegion
	}
	endpoint, err := c.forecastEndpoint("events", url.Values{"region": []string{rankingRegionPath(c.region)}})
	if err != nil {
		return nil, err
	}
	var events []ForecastEvent
	if err := c.getJSON(endpoint, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (c *Client) GetForecastLatest(eventID int) (*ForecastBoard, error) {
	if !c.supportsForecast() {
		return nil, ErrForecastUnsupportedRegion
	}
	endpoint, err := c.forecastEndpoint(fmt.Sprintf("event/%d/latest", eventID), url.Values{"region": []string{rankingRegionPath(c.region)}})
	if err != nil {
		return nil, err
	}
	var board ForecastBoard
	if err := c.getJSON(endpoint, &board); err != nil {
		return nil, err
	}
	board.Region = c.region
	return &board, nil
}

func (c *Client) getBoard(path string, query url.Values) (*Board, error) {
	endpoint, err := c.publicEndpoint(path, query)
	if err != nil {
		return nil, err
	}
	var board Board
	if err := c.getJSON(endpoint, &board); err != nil {
		return nil, err
	}
	return &board, nil
}

func (c *Client) publicEndpoint(path string, query url.Values) (string, error) {
	parts := []string{rankingRegionPath(c.region), strings.Trim(path, "/")}
	base := strings.TrimRight(c.baseURL, "/")
	if !strings.HasSuffix(base, "/api/public") && !strings.Contains(base, "/api/public/") {
		parts = append([]string{"api", "public"}, parts...)
	}
	endpoint, err := url.JoinPath(base, parts...)
	if err != nil {
		return "", fmt.Errorf("build ranking url: %w", err)
	}
	return withQuery(endpoint, query), nil
}

func (c *Client) forecastEndpoint(path string, query url.Values) (string, error) {
	endpoint, err := url.JoinPath(strings.TrimRight(c.forecastURL, "/"), strings.Trim(path, "/"))
	if err != nil {
		return "", fmt.Errorf("build forecast url: %w", err)
	}
	return withQuery(endpoint, query), nil
}

func (c *Client) getJSON(endpoint string, out any) error {
	resp, err := c.http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("ranking request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			Error  string `json:"error"`
			Region string `json:"region"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if strings.Contains(strings.ToLower(apiErr.Error), "no worldlink") {
			return ErrNoWorldLinkData
		}
		return fmt.Errorf("ranking request returned %d: %s", resp.StatusCode, strings.TrimSpace(apiErr.Error))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode ranking response: %w", err)
	}
	if apiErr, ok := decodedAPIError(out); ok && strings.Contains(strings.ToLower(apiErr), "no worldlink") {
		return ErrNoWorldLinkData
	}
	return nil
}

func decodedAPIError(out any) (string, bool) {
	if out == nil {
		return "", false
	}
	data, err := json.Marshal(out)
	if err != nil {
		return "", false
	}
	var apiErr struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &apiErr); err != nil || strings.TrimSpace(apiErr.Error) == "" {
		return "", false
	}
	return apiErr.Error, true
}

func (c *Client) supportsForecast() bool {
	region := config.NormalizeRegion(c.region)
	return region == config.RegionCN || region == config.RegionJP
}

func isNoWorldLinkError(err error) bool {
	if errors.Is(err, ErrNoWorldLinkData) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "no worldlink")
}

func withQuery(endpoint string, query url.Values) string {
	if len(query) == 0 {
		return endpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	values := parsed.Query()
	for key, items := range query {
		for _, item := range items {
			values.Add(key, item)
		}
	}
	parsed.RawQuery = values.Encode()
	return parsed.String()
}

func rankingRegionPath(region string) string {
	if config.NormalizeRegion(region) == config.RegionTW {
		return "tw"
	}
	return config.NormalizeRegion(region)
}

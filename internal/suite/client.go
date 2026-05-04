package suite

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"moebot-next/internal/config"
)

// Client queries Suite packet data by game UID.
type Client struct {
	enabled    bool
	urlPattern string
	token      string
	timeout    time.Duration
	mode       string
	httpClient *http.Client
}

type Status struct {
	UserID      string
	Name        string
	Source      string
	LocalSource string
	UploadTime  int64
}

type statusResponse struct {
	BaseProfile
	UserGamedata UserGamedata `json:"userGamedata"`
}

func NewClient(cfg config.SuiteAPIConfig) *Client {
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	mode := config.NormalizeSuiteMode(cfg.DefaultMode)
	return &Client{
		enabled:    cfg.Enabled,
		urlPattern: strings.TrimSpace(cfg.URL),
		token:      cfg.Token,
		timeout:    timeout,
		mode:       mode,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.enabled && c.urlPattern != ""
}

func (c *Client) GetStatus(uid string, mode string) (Status, error) {
	var profile statusResponse
	if err := c.GetUserData(uid, mode, []string{FieldUploadTime, FieldUserGamedata}, &profile); err != nil {
		return Status{}, err
	}
	return Status{
		UserID:      profile.UserGamedata.UserID.String(),
		Name:        profile.UserGamedata.Name,
		Source:      profile.Source,
		LocalSource: profile.LocalSource,
		UploadTime:  profile.UploadTime,
	}, nil
}

func (c *Client) GetUserData(uid string, mode string, filter []string, out any) error {
	if c == nil || !c.Enabled() {
		return fmt.Errorf("suite api is disabled")
	}
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return fmt.Errorf("uid is empty")
	}
	if out == nil {
		return fmt.Errorf("suite response output is nil")
	}
	mode = config.NormalizeSuiteMode(firstNonEmpty(mode, c.mode))
	if !config.IsValidSuiteMode(mode) {
		return fmt.Errorf("unsupported suite mode %q", mode)
	}
	endpoint := strings.ReplaceAll(c.urlPattern, "{uid}", url.PathEscape(uid))
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parse suite url: %w", err)
	}
	query := parsed.Query()
	query.Set("mode", mode)
	if len(filter) > 0 {
		query.Set("filter", strings.Join(filter, ","))
	}
	parsed.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, parsed.String(), nil)
	if err != nil {
		return fmt.Errorf("build suite request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("suite request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("suite request returned %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode suite response: %w", err)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

type jsonID string

func (id *jsonID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*id = ""
		return nil
	}
	if data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*id = jsonID(value)
		return nil
	}
	*id = jsonID(string(data))
	return nil
}

func (id jsonID) String() string { return string(id) }

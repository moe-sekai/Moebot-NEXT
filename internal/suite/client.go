package suite

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
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
	region     string
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

func NewClient(cfg config.SuiteAPIConfig, regions ...string) *Client {
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	region := ""
	if len(regions) > 0 {
		region = config.NormalizeRegion(regions[0])
	}
	return &Client{
		enabled:    cfg.Enabled,
		urlPattern: strings.TrimSpace(firstNonEmpty(cfg.URL, config.DefaultSuiteAPIURL)),
		token:      cfg.Token,
		timeout:    timeout,
		region:     region,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.enabled
}

func (c *Client) GetStatus(uid string, mode string) (Status, error) {
	var profile statusResponse
	if err := c.GetUserData(uid, mode, nil, &profile); err != nil {
		return Status{}, err
	}
	return Status{
		UserID:      profile.UserGamedata.UserID.String(),
		Name:        profile.UserGamedata.Name,
		Source:      firstNonEmpty(profile.Source, PublicSource),
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
	parsed, err := c.buildURL(uid, filter)
	if err != nil {
		return err
	}

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read suite response: %w", err)
	}
	if err := decodeUserDataResponse(body, out); err != nil {
		return fmt.Errorf("decode suite response: %w", err)
	}
	return nil
}

func (c *Client) buildURL(uid string, filter []string) (*url.URL, error) {
	endpoint := firstNonEmpty(c.urlPattern, config.DefaultSuiteAPIURL)
	endpoint = strings.ReplaceAll(endpoint, "{uid}", url.PathEscape(uid))
	region := c.region
	if region == "" {
		region = inferRegionFromSuiteURL(endpoint)
	}
	if region == "" {
		region = config.RegionJP
	}
	endpoint = strings.ReplaceAll(endpoint, "{region}", url.PathEscape(region))
	endpoint = strings.ReplaceAll(endpoint, "{regin}", url.PathEscape(region))

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse suite url: %w", err)
	}
	query := parsed.Query()
	fields := normalizedFields(filter)
	query.Set("key", strings.Join(fields, ","))
	query.Del("filter")
	query.Del("mode")
	parsed.RawQuery = query.Encode()
	return parsed, nil
}

func normalizedFields(fields []string) []string {
	if len(fields) == 0 {
		return DefaultHarukiPublicFields()
	}
	out := make([]string, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		out = append(out, field)
	}
	if len(out) == 0 {
		return DefaultHarukiPublicFields()
	}
	return out
}

func inferRegionFromSuiteURL(endpoint string) string {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	parts := strings.Split(parsed.Path, "/")
	for i, part := range parts {
		if part == "public" && i+1 < len(parts) {
			region := config.NormalizeRegion(parts[i+1])
			if config.IsValidRegion(region) {
				return region
			}
		}
	}
	return ""
}

func decodeUserDataResponse(data []byte, out any) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil {
		if wrapped, ok := raw["data"]; ok && len(wrapped) > 0 && string(wrapped) != "null" {
			if err := json.Unmarshal(wrapped, out); err != nil {
				return err
			}
			mergeTopLevelBaseProfile(data, out)
			return nil
		}
	}
	return json.Unmarshal(data, out)
}

func mergeTopLevelBaseProfile(data []byte, out any) {
	var base BaseProfile
	if err := json.Unmarshal(data, &base); err != nil {
		return
	}
	if base.UploadTime == 0 && base.Source == "" && base.LocalSource == "" {
		return
	}
	value := reflect.ValueOf(out)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return
	}
	field := value.FieldByName("BaseProfile")
	if !field.IsValid() || !field.CanSet() || field.Type() != reflect.TypeOf(BaseProfile{}) {
		return
	}
	current := field.Interface().(BaseProfile)
	if current.UploadTime == 0 {
		current.UploadTime = base.UploadTime
	}
	if current.Source == "" {
		current.Source = base.Source
	}
	if current.LocalSource == "" {
		current.LocalSource = base.LocalSource
	}
	field.Set(reflect.ValueOf(current))
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

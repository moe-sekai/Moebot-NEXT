package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"moebot-next/internal/config"
	"moebot-next/internal/database"
)

func TestPublicConfigIncludesMaskedSuiteAPI(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Server.Region = config.RegionCN
	cfg.SuiteAPI = config.SuiteAPIConfig{
		Enabled:     true,
		URL:         "https://suite.example.test/api/cn/user/{uid}/suite",
		Headers:     map[string]string{"X-Suite-Header": "suite-value"},
		Timeout:     8,
		DefaultMode: config.SuiteModeMoeSekai,
	}
	config.NormalizeConfig(cfg)

	db, err := database.New(config.DatabaseConfig{Path: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	server := New(cfg, db, nil, nil, "", nil)

	resp, err := server.App.Test(httptestRequest(http.MethodGet, "/api/config/public", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	suiteAPI := body["suite_api"].(map[string]any)
	if suiteAPI["token"] != nil {
		t.Fatalf("suite api leaked deprecated token field: %+v", suiteAPI)
	}
	suiteHeaders := suiteAPI["headers"].(map[string]any)
	if suiteHeaders["X-Suite-Header"] != "suite-value" {
		t.Fatalf("suite api headers = %+v", suiteHeaders)
	}
	if suiteAPI["url"] != "https://suite.example.test/api/cn/user/{uid}/suite" {
		t.Fatalf("suite api url = %+v", suiteAPI["url"])
	}
	if suiteAPI["enabled"] != true || suiteAPI["url_configured"] != true || suiteAPI["headers_configured"] != true {
		t.Fatalf("suite api flags = %+v", suiteAPI)
	}
	if suiteAPI["timeout"].(float64) != 8 || suiteAPI["default_mode"] != config.SuiteModeMoeSekai {
		t.Fatalf("suite api values = %+v", suiteAPI)
	}
	servers := body["servers"].(map[string]any)
	cn := servers[config.RegionCN].(map[string]any)
	cnSuite := cn["suite_api"].(map[string]any)
	if cnSuite["enabled"] != true || cnSuite["url_configured"] != true || cnSuite["headers_configured"] != true {
		t.Fatalf("cn suite api = %+v", cnSuite)
	}
}

func TestUpdatePublicConfigSavesSuiteAPISettings(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	cfg := config.DefaultConfig()
	cfg.Server.Region = config.RegionCN
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatal(err)
	}

	db, err := database.New(config.DatabaseConfig{Path: filepath.Join(dir, "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	server := New(cfg, db, nil, nil, cfgPath, nil)

	payload := []byte(`{
		"server": {"region": "cn"},
		"servers": {
			"cn": {
				"enabled": true,
				"suite_api": {
					"enabled": true,
					"url": "https://suite.example.test/api/cn/user/{uid}/suite",
					"headers": {"X-Suite-Header": "suite-value"},
					"timeout": 9,
					"default_mode": "haruki"
				}
			}
		}
	}`)
	resp, err := server.App.Test(httptestRequest(http.MethodPut, "/api/config/public", bytes.NewReader(payload)), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		t.Fatalf("status = %d body=%+v", resp.StatusCode, body)
	}

	saved, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	profile := config.ResolveGameServerProfile(saved, config.RegionCN)
	if !profile.SuiteAPI.Enabled {
		t.Fatal("suite api should be enabled")
	}
	if profile.SuiteAPI.URL != "https://suite.example.test/api/cn/user/{uid}/suite" {
		t.Fatalf("url = %q", profile.SuiteAPI.URL)
	}
	if profile.SuiteAPI.Headers["X-Suite-Header"] != "suite-value" {
		t.Fatalf("headers = %+v", profile.SuiteAPI.Headers)
	}
	if profile.SuiteAPI.Timeout != 9 || profile.SuiteAPI.DefaultMode != config.SuiteModeHaruki {
		t.Fatalf("suite api = %+v", profile.SuiteAPI)
	}
}

func httptestRequest(method string, target string, body *bytes.Reader) *http.Request {
	if body == nil {
		body = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, target, body)
	req.Header.Set("Content-Type", "application/json")
	return req
}

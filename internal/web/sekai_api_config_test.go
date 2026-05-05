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

func TestPublicConfigMasksSekaiAPISensitiveFields(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SekaiAPI.Enabled = true
	cfg.SekaiAPI.BaseURL = "https://sekai.example.test"
	cfg.SekaiAPI.Headers = map[string]string{"X-Test-Header": "test-value", "Authorization": "Bearer secret-token"}
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
	sekaiAPI := body["sekai_api"].(map[string]any)
	if sekaiAPI["token"] != nil {
		t.Fatalf("sekai api leaked deprecated token field: %+v", sekaiAPI)
	}
	sekaiHeaders := sekaiAPI["headers"].(map[string]any)
	if sekaiHeaders["Authorization"] != "Bearer secret-token" {
		t.Fatalf("sekai api headers = %+v", sekaiHeaders)
	}
	if sekaiAPI["base_url"] != "https://sekai.example.test" {
		t.Fatalf("sekai api base url = %+v", sekaiAPI["base_url"])
	}
	if sekaiAPI["enabled"] != true || sekaiAPI["base_url_configured"] != true || sekaiAPI["headers_configured"] != true {
		t.Fatalf("sekai api flags = %+v", sekaiAPI)
	}
}

func TestSekaiSystemURLSupportsRegionPlaceholder(t *testing.T) {
	got, err := sekaiSystemURL("https://seka-api.exmeaning.com/api/{region}", config.RegionCN)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://seka-api.exmeaning.com/api/cn/system" {
		t.Fatalf("system url = %q", got)
	}

	got, err = sekaiSystemURL("https://seka-api.exmeaning.com", config.RegionJP)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://seka-api.exmeaning.com/api/jp/system" {
		t.Fatalf("default system url = %q", got)
	}
}

func TestUpdatePublicConfigSavesSekaiAPISettings(t *testing.T) {
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
				"sekai_api": {
					"enabled": true,
					"base_url": "https://sekai.example.test",
					"region": "cn",
					"headers": {"X-Test-Header": "test-value", "Authorization": "Bearer secret-token"},
					"timeout": 9,
					"rate_limit": 12
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
	if !profile.SekaiAPI.Enabled {
		t.Fatal("sekai api should be enabled")
	}
	if profile.SekaiAPI.BaseURL != "https://sekai.example.test" || profile.SekaiAPI.Region != config.RegionCN {
		t.Fatalf("sekai api endpoint = %+v", profile.SekaiAPI)
	}
	if profile.SekaiAPI.Headers["X-Test-Header"] != "test-value" || profile.SekaiAPI.Headers["Authorization"] != "Bearer secret-token" {
		t.Fatalf("sekai api headers = %+v", profile.SekaiAPI.Headers)
	}
	if profile.SekaiAPI.Timeout != 9 || profile.SekaiAPI.RateLimit != 12 {
		t.Fatalf("sekai api limits = %+v", profile.SekaiAPI)
	}
}

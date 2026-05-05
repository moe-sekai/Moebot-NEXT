package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSuiteAPIConfigLoadedAndMergedIntoGameServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	content := []byte(`server:
  region: "cn"
suite_api:
  enabled: true
  url: "https://suite.example.test/api/cn/user/{uid}/suite"
  headers:
    X-Suite-Header: "suite-value"
  timeout: 7
  default_mode: "moesekai"
game_servers:
  cn:
    enabled: true
`)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.SuiteAPI.Enabled {
		t.Fatal("suite api should be enabled")
	}
	if cfg.SuiteAPI.DefaultMode != SuiteModeHaruki {
		t.Fatalf("default mode = %q, want %q", cfg.SuiteAPI.DefaultMode, SuiteModeHaruki)
	}

	profile := ResolveGameServerProfile(cfg, RegionCN)
	if !profile.SuiteAPI.Enabled {
		t.Fatal("cn suite api should inherit global enabled flag")
	}
	if profile.SuiteAPI.URL != "https://suite.example.test/api/cn/user/{uid}/suite" {
		t.Fatalf("url = %q", profile.SuiteAPI.URL)
	}
	if profile.SuiteAPI.Headers["X-Suite-Header"] != "suite-value" {
		t.Fatalf("headers = %+v", profile.SuiteAPI.Headers)
	}
	if profile.SuiteAPI.Timeout != 7 {
		t.Fatalf("timeout = %d", profile.SuiteAPI.Timeout)
	}
	if profile.SuiteAPI.DefaultMode != SuiteModeHaruki {
		t.Fatalf("mode = %q", profile.SuiteAPI.DefaultMode)
	}
}

func TestSuiteAPIExplicitFalseSurvivesNormalize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	content := []byte(`server:
  region: "cn"
suite_api:
  enabled: false
game_servers:
  cn:
    enabled: true
`)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SuiteAPI.Enabled {
		t.Fatal("suite api enabled=false should survive normalize")
	}
	profile := ResolveGameServerProfile(cfg, RegionCN)
	if profile.SuiteAPI.Enabled {
		t.Fatal("cn suite api should inherit explicit global disabled flag")
	}
}

func TestNormalizeSuiteMode(t *testing.T) {
	cases := map[string]string{
		"":          SuiteModeHaruki,
		"latest":    SuiteModeHaruki,
		"local":     SuiteModeHaruki,
		"haruki":    SuiteModeHaruki,
		"moesekai":  SuiteModeHaruki,
		"moe-sekai": SuiteModeHaruki,
	}
	for input, want := range cases {
		if got := NormalizeSuiteMode(input); got != want {
			t.Fatalf("NormalizeSuiteMode(%q) = %q, want %q", input, got, want)
		}
	}
}

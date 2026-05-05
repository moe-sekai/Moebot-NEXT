package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSekaiAPIHeadersLoadedFromConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	content := []byte(`sekai_api:
  enabled: true
  base_url: "https://example.test"
  region: "cn"
  headers:
    X-Test-Header: "test-value"
    Authorization: "Bearer secret-token"
`)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if got := cfg.SekaiAPI.Headers["X-Test-Header"]; got != "test-value" {
		t.Fatalf("header = %q, want test-value", got)
	}
	if got := cfg.SekaiAPI.Headers["Authorization"]; got != "Bearer secret-token" {
		t.Fatalf("authorization = %q", got)
	}
}

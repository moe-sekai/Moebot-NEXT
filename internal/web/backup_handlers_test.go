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

func TestBackupConfigRedactsSecrets(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Backup.S3.Endpoint = "minio.example.test:9000"
	cfg.Backup.S3.Bucket = "moebot"
	cfg.Backup.S3.AccessKey = "access-secret"
	cfg.Backup.S3.SecretKey = "secret-secret"
	cfg.Backup.S3.SessionToken = "token-secret"
	config.NormalizeConfig(cfg)

	db, err := database.New(config.DatabaseConfig{Path: filepath.Join(dir, "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	server := New(cfg, db, nil, nil, "", nil)

	req := httptestRequest(http.MethodGet, "/api/backup/config", nil)
	mustAuthorizeRequest(t, server, req)
	resp, err := server.App.Test(req, -1)
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
	if body["access_key"] != nil || body["secret_key"] != nil || body["session_token"] != nil {
		t.Fatalf("backup config leaked secrets: %+v", body)
	}
	if body["access_key_set"] != true || body["secret_key_set"] != true || body["session_token_set"] != true || body["configured"] != true {
		t.Fatalf("secret flags = %+v", body)
	}
	excludes := body["exclude_patterns"].([]any)
	if len(excludes) == 0 || excludes[0] != "cache/**" {
		t.Fatalf("exclude patterns = %+v", excludes)
	}
}

func TestUpdateBackupConfigKeepsSecretsWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	cfg := config.DefaultConfig()
	cfg.Backup.S3.AccessKey = "old-ak"
	cfg.Backup.S3.SecretKey = "old-sk"
	cfg.Backup.S3.SessionToken = "old-token"
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatal(err)
	}
	db, err := database.New(config.DatabaseConfig{Path: filepath.Join(dir, "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	server := New(cfg, db, nil, nil, cfgPath, nil)

	payload := []byte(`{"endpoint":"minio.example.test:9000","bucket":"moebot","access_key":"","secret_key":"","session_token":""}`)
	req := httptestRequest(http.MethodPut, "/api/backup/config", bytes.NewReader(payload))
	mustAuthorizeRequest(t, server, req)
	resp, err := server.App.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		t.Fatalf("status = %d body=%+v", resp.StatusCode, body)
	}
	if cfg.Backup.S3.AccessKey != "old-ak" || cfg.Backup.S3.SecretKey != "old-sk" || cfg.Backup.S3.SessionToken != "old-token" {
		t.Fatalf("secrets changed unexpectedly: %+v", cfg.Backup.S3)
	}
}

func TestRestoreBackupRequiresConfirmation(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	db, err := database.New(config.DatabaseConfig{Path: filepath.Join(dir, "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	server := New(cfg, db, nil, nil, filepath.Join(dir, "config.yml"), nil)

	payload := []byte(`{"key":"moebot-next/backups/demo.tar.gz"}`)
	req := httptestRequest(http.MethodPost, "/api/backup/restore", bytes.NewReader(payload))
	mustAuthorizeRequest(t, server, req)
	resp, err := server.App.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

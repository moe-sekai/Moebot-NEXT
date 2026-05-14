package backup

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateAndExtractArchive(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	tempDir := filepath.Join(dataDir, "backups", "tmp")
	if err := os.MkdirAll(filepath.Join(dataDir, "plugins"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "config.yml"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "plugins", "moesekai.yml"), []byte("plugin"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dataDir, "cache", "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "cache", "images", "thumb.png"), []byte("cache"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "skip.tmp"), []byte("skip"), 0o644); err != nil {
		t.Fatal(err)
	}

	archive := filepath.Join(dir, "backup.tar.gz")
	if err := CreateArchiveWithOptions(dataDir, archive, ArchiveOptions{TempDir: tempDir, ExcludePatterns: []string{"cache/**"}}); err != nil {
		t.Fatalf("CreateArchive() error = %v", err)
	}
	extractDir := filepath.Join(dir, "restore")
	if err := ExtractArchive(archive, extractDir); err != nil {
		t.Fatalf("ExtractArchive() error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(extractDir, "plugins", "moesekai.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "plugin" {
		t.Fatalf("restored plugin = %q", got)
	}
	if _, err := os.Stat(filepath.Join(extractDir, "backups", "tmp", "skip.tmp")); !os.IsNotExist(err) {
		t.Fatalf("temp dir should be excluded, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(extractDir, "cache", "images", "thumb.png")); !os.IsNotExist(err) {
		t.Fatalf("cache dir should be excluded, stat err = %v", err)
	}
}

func TestExtractArchiveRejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "evil.tar.gz")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	body := []byte("evil")
	if err := tw.WriteHeader(&tar.Header{Name: "../evil.txt", Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	err = ExtractArchive(archive, filepath.Join(dir, "restore"))
	if err == nil || !strings.Contains(err.Error(), "unsafe archive path") {
		t.Fatalf("ExtractArchive() err = %v, want unsafe archive path", err)
	}
}

func TestValidateObjectKey(t *testing.T) {
	if err := validateObjectKey("moebot-next/backups", "moebot-next/backups/moebot-data-20260101T000000Z.tar.gz"); err != nil {
		t.Fatalf("validateObjectKey() error = %v", err)
	}
	if err := validateObjectKey("moebot-next/backups", "other/moebot-data.tar.gz"); err == nil {
		t.Fatal("validateObjectKey() should reject keys outside prefix")
	}
	if got := buildObjectKey("/a/b/", "/c.tar.gz"); got != "a/b/c.tar.gz" {
		t.Fatalf("buildObjectKey() = %q", got)
	}
}

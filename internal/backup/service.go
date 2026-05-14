package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/database"
)

// Service coordinates data directory archive/restore and S3-compatible storage.
type Service struct {
	cfg *config.Config
	db  *database.DB
}

// New creates a backup service bound to the current core config.
func New(cfg *config.Config, db *database.DB) *Service {
	return &Service{cfg: cfg, db: db}
}

// PublicConfig returns redacted backup settings for the Web UI.
func (s *Service) PublicConfig() PublicConfig {
	cfg := s.effectiveConfig()
	return PublicConfig{
		DataDir:         cfg.DataDir,
		TempDir:         cfg.TempDir,
		Endpoint:        cfg.S3.Endpoint,
		Region:          cfg.S3.Region,
		Bucket:          cfg.S3.Bucket,
		Prefix:          cfg.S3.Prefix,
		UseSSL:          cfg.S3.UseSSL,
		ForcePathStyle:  cfg.S3.ForcePathStyle,
		AccessKeySet:    cfg.S3.AccessKey != "",
		SecretKeySet:    cfg.S3.SecretKey != "",
		SessionTokenSet: cfg.S3.SessionToken != "",
		Configured:      backupS3Configured(cfg.S3),
	}
}

// ValidateConfig checks whether enough S3 settings are present.
func (s *Service) ValidateConfig() error {
	return validateBackupConfig(s.effectiveConfig())
}

// TestConnection validates the bucket by listing at most one object.
func (s *Service) TestConnection(ctx context.Context) error {
	cfg := s.effectiveConfig()
	if err := validateBackupConfig(cfg); err != nil {
		return err
	}
	_, err := newS3Client(cfg.S3).list(ctx, 1)
	return err
}

// List returns remote backup archives sorted newest first.
func (s *Service) List(ctx context.Context) ([]ObjectInfo, error) {
	cfg := s.effectiveConfig()
	if err := validateBackupConfig(cfg); err != nil {
		return nil, err
	}
	objects, err := newS3Client(cfg.S3).list(ctx, 0)
	if err != nil {
		return nil, err
	}
	out := make([]ObjectInfo, 0, len(objects))
	for _, obj := range objects {
		out = append(out, ObjectInfo{
			Key:          obj.Key,
			Name:         filepath.Base(obj.Key),
			Size:         obj.Size,
			LastModified: obj.LastModified,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].LastModified.After(out[j].LastModified)
	})
	return out, nil
}

// Create archives the configured data directory and uploads it to S3.
func (s *Service) Create(ctx context.Context, note string) (*CreateResult, error) {
	started := time.Now()
	cfg := s.effectiveConfig()
	if err := validateBackupConfig(cfg); err != nil {
		return nil, err
	}
	dataDir, err := filepath.Abs(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("resolve data dir: %w", err)
	}
	if err := os.MkdirAll(cfg.TempDir, 0o755); err != nil {
		return nil, fmt.Errorf("create backup temp dir: %w", err)
	}
	if err := s.checkpointSQLiteIfPossible(dataDir); err != nil {
		return nil, err
	}

	stamp := time.Now().UTC().Format("20060102T150405Z")
	archiveName := "moebot-data-" + stamp + ".tar.gz"
	archivePath := filepath.Join(cfg.TempDir, archiveName+".tmp")
	defer os.Remove(archivePath)
	if err := CreateArchive(dataDir, archivePath, cfg.TempDir); err != nil {
		return nil, err
	}
	info, err := os.Stat(archivePath)
	if err != nil {
		return nil, err
	}
	key := buildObjectKey(cfg.S3.Prefix, archiveName)
	obj, err := newS3Client(cfg.S3).putFile(ctx, key, archivePath)
	if err != nil {
		return nil, err
	}
	if obj.LastModified.IsZero() {
		obj.LastModified = time.Now().UTC()
	}
	result := &CreateResult{
		Object: ObjectInfo{
			Key:          obj.Key,
			Name:         filepath.Base(obj.Key),
			Size:         obj.Size,
			LastModified: obj.LastModified,
		},
		ArchiveSize: info.Size(),
		Duration:    time.Since(started),
	}
	result.DurationMS = result.Duration.Milliseconds()
	return result, nil
}

// Restore downloads a backup archive and replaces the configured data directory.
func (s *Service) Restore(ctx context.Context, key string) (*RestoreResult, error) {
	started := time.Now()
	cfg := s.effectiveConfig()
	if err := validateBackupConfig(cfg); err != nil {
		return nil, err
	}
	key = strings.TrimSpace(key)
	if err := validateObjectKey(cfg.S3.Prefix, key); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cfg.TempDir, 0o755); err != nil {
		return nil, fmt.Errorf("create backup temp dir: %w", err)
	}
	stamp := time.Now().UTC().Format("20060102T150405Z")
	downloadPath := filepath.Join(cfg.TempDir, "restore-"+stamp+".tar.gz")
	extractDir := filepath.Join(cfg.TempDir, "restore-extract-"+stamp)
	defer os.Remove(downloadPath)
	defer os.RemoveAll(extractDir)

	if err := newS3Client(cfg.S3).getFile(ctx, key, downloadPath); err != nil {
		return nil, err
	}
	if err := ExtractArchive(downloadPath, extractDir); err != nil {
		return nil, err
	}
	dataDir, err := filepath.Abs(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("resolve data dir: %w", err)
	}
	backupDir := dataDir + ".restore-backup-" + stamp
	if err := replaceDataDir(dataDir, extractDir, backupDir); err != nil {
		return nil, err
	}
	result := &RestoreResult{
		Key:             key,
		BackupDataDir:   backupDir,
		RestoredDataDir: dataDir,
		Duration:        time.Since(started),
		RestartNeeded:   true,
	}
	result.DurationMS = result.Duration.Milliseconds()
	return result, nil
}

// Delete removes one remote backup archive.
func (s *Service) Delete(ctx context.Context, key string) error {
	cfg := s.effectiveConfig()
	if err := validateBackupConfig(cfg); err != nil {
		return err
	}
	key = strings.TrimSpace(key)
	if err := validateObjectKey(cfg.S3.Prefix, key); err != nil {
		return err
	}
	return newS3Client(cfg.S3).delete(ctx, key)
}

func (s *Service) effectiveConfig() config.BackupConfig {
	cfg := config.DefaultConfig().Backup
	if s != nil && s.cfg != nil {
		cfg = s.cfg.Backup
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "./data"
	}
	if cfg.TempDir == "" {
		cfg.TempDir = "./data/backups/tmp"
	}
	cfg.S3.Prefix = strings.Trim(strings.TrimSpace(cfg.S3.Prefix), "/")
	if cfg.S3.Prefix == "" {
		cfg.S3.Prefix = "moebot-next/backups"
	}
	return cfg
}

func validateBackupConfig(cfg config.BackupConfig) error {
	if strings.TrimSpace(cfg.DataDir) == "" {
		return fmt.Errorf("backup data_dir is required")
	}
	if strings.TrimSpace(cfg.TempDir) == "" {
		return fmt.Errorf("backup temp_dir is required")
	}
	if strings.TrimSpace(cfg.S3.Endpoint) == "" {
		return fmt.Errorf("S3 endpoint is required")
	}
	if strings.TrimSpace(cfg.S3.Bucket) == "" {
		return fmt.Errorf("S3 bucket is required")
	}
	if strings.TrimSpace(cfg.S3.AccessKey) == "" {
		return fmt.Errorf("S3 access_key is required")
	}
	if strings.TrimSpace(cfg.S3.SecretKey) == "" {
		return fmt.Errorf("S3 secret_key is required")
	}
	return nil
}

func backupS3Configured(cfg config.BackupS3Config) bool {
	return strings.TrimSpace(cfg.Endpoint) != "" && strings.TrimSpace(cfg.Bucket) != "" && strings.TrimSpace(cfg.AccessKey) != "" && strings.TrimSpace(cfg.SecretKey) != ""
}

func buildObjectKey(prefix, name string) string {
	prefix = strings.Trim(strings.TrimSpace(prefix), "/")
	name = strings.TrimLeft(strings.TrimSpace(name), "/")
	if prefix == "" {
		return name
	}
	return prefix + "/" + name
}

func validateObjectKey(prefix, key string) error {
	if key == "" {
		return fmt.Errorf("backup key is required")
	}
	if strings.Contains(key, "..") || strings.HasPrefix(key, "/") || strings.HasPrefix(key, "\\") {
		return fmt.Errorf("invalid backup key")
	}
	prefix = strings.Trim(strings.TrimSpace(prefix), "/")
	if prefix != "" && key != prefix && !strings.HasPrefix(key, prefix+"/") {
		return fmt.Errorf("backup key is outside configured prefix")
	}
	if !strings.HasSuffix(key, ".tar.gz") {
		return fmt.Errorf("backup key must point to a .tar.gz archive")
	}
	return nil
}

func replaceDataDir(dataDir, extractedDir, backupDir string) error {
	if err := os.MkdirAll(filepath.Dir(dataDir), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(dataDir); err == nil {
		if err := os.Rename(dataDir, backupDir); err != nil {
			return fmt.Errorf("move current data dir aside: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat current data dir: %w", err)
	}
	if err := os.Rename(extractedDir, dataDir); err != nil {
		if _, statErr := os.Stat(backupDir); statErr == nil {
			_ = os.Rename(backupDir, dataDir)
		}
		return fmt.Errorf("activate restored data dir: %w", err)
	}
	return nil
}

func (s *Service) checkpointSQLiteIfPossible(dataDir string) error {
	if s == nil || s.db == nil || s.db.DB == nil || s.cfg == nil {
		return nil
	}
	dbPath := strings.TrimSpace(s.cfg.Database.Path)
	if dbPath == "" {
		return nil
	}
	dbAbs, err := filepath.Abs(dbPath)
	if err != nil {
		return nil
	}
	if dbAbs != dataDir && !isWithin(dbAbs, dataDir) {
		return nil
	}
	return s.db.Exec("PRAGMA wal_checkpoint(PASSIVE)").Error
}

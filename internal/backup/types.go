package backup

import "time"

// PublicConfig is the redacted backup configuration exposed to the Web UI.
type PublicConfig struct {
	DataDir               string   `json:"data_dir"`
	TempDir               string   `json:"temp_dir"`
	ExcludePatterns       []string `json:"exclude_patterns"`
	Endpoint              string   `json:"endpoint"`
	Region                string   `json:"region"`
	Bucket                string   `json:"bucket"`
	Prefix                string   `json:"prefix"`
	UseSSL                bool     `json:"use_ssl"`
	ForcePathStyle        bool     `json:"force_path_style"`
	ScheduleEnabled       bool     `json:"schedule_enabled"`
	ScheduleIntervalHours int      `json:"schedule_interval_hours"`
	AccessKeySet          bool     `json:"access_key_set"`
	SecretKeySet          bool     `json:"secret_key_set"`
	SessionTokenSet       bool     `json:"session_token_set"`
	Configured            bool     `json:"configured"`
}

// ObjectInfo describes one remote backup archive.
type ObjectInfo struct {
	Key          string    `json:"key"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

// CreateResult describes a created backup.
type CreateResult struct {
	Object      ObjectInfo    `json:"object"`
	ArchiveSize int64         `json:"archive_size"`
	Duration    time.Duration `json:"-"`
	DurationMS  int64         `json:"duration_ms"`
}

// RestoreResult describes a completed restore operation.
type RestoreResult struct {
	Key             string        `json:"key"`
	BackupDataDir   string        `json:"backup_data_dir"`
	RestoredDataDir string        `json:"restored_data_dir"`
	Duration        time.Duration `json:"-"`
	DurationMS      int64         `json:"duration_ms"`
	RestartNeeded   bool          `json:"restart_needed"`
}

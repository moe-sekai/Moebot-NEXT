package gallery

import "sync"

type Config struct {
	SizeLimitMB              int    `yaml:"size_limit_mb"`
	PickLimit                int    `yaml:"pick_limit"`
	Hash1DifferenceThreshold int    `yaml:"hash1_difference_threshold"`
	Hash2DifferenceThreshold int    `yaml:"hash2_difference_threshold"`
	RevertExpiredHours       int    `yaml:"revert_expired_hours"`
	DataDir                  string `yaml:"data_dir"`
}

func applyDefaults(c *Config) {
	if c.SizeLimitMB <= 0 {
		c.SizeLimitMB = 1
	}
	if c.PickLimit <= 0 {
		c.PickLimit = 5
	}
	if c.Hash1DifferenceThreshold <= 0 {
		c.Hash1DifferenceThreshold = 5
	}
	if c.Hash2DifferenceThreshold <= 0 {
		c.Hash2DifferenceThreshold = 1000
	}
	if c.RevertExpiredHours <= 0 {
		c.RevertExpiredHours = 24
	}
	if c.DataDir == "" {
		c.DataDir = "data/gallery"
	}
}

var (
	cfgMu  sync.RWMutex
	cfgPtr *Config
)

func setConfig(c *Config) {
	cfgMu.Lock()
	cfgPtr = c
	cfgMu.Unlock()
}

func getConfig() *Config {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return cfgPtr
}

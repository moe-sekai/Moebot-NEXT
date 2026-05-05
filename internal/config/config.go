package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultRendererPrecision = 1.5

// Config is the root configuration for Moebot NEXT.
type Config struct {
	Server      ServerConfig                `yaml:"server"`
	Bot         BotConfig                   `yaml:"bot"`
	Web         WebConfig                   `yaml:"web"`
	Database    DatabaseConfig              `yaml:"database"`
	Masterdata  MasterdataConfig            `yaml:"masterdata"`
	SekaiAPI    SekaiAPIConfig              `yaml:"sekai_api"`
	SuiteAPI    SuiteAPIConfig              `yaml:"suite_api"`
	RankingAPI  RankingAPIConfig            `yaml:"ranking_api"`
	Renderer    RendererConfig              `yaml:"renderer"`
	Assets      AssetsConfig                `yaml:"assets"`
	GameServers map[string]GameServerConfig `yaml:"game_servers"`
	Log         LogConfig                   `yaml:"log"`
}

// ServerConfig selects the default PJSK game server region.
type ServerConfig struct {
	Region string `yaml:"region"` // cn / jp / tw / kr / en
}

// GameServerConfig holds per-region data/API/resource settings.
type GameServerConfig struct {
	Enabled    *bool            `yaml:"enabled"`
	Masterdata MasterdataConfig `yaml:"masterdata"`
	SekaiAPI   SekaiAPIConfig   `yaml:"sekai_api"`
	SuiteAPI   SuiteAPIConfig   `yaml:"suite_api"`
	RankingAPI RankingAPIConfig `yaml:"ranking_api"`
	Assets     AssetsConfig     `yaml:"assets"`
}

// BotConfig holds ZeroBot-related settings.
type BotConfig struct {
	Nickname       []string            `yaml:"nickname"`
	CommandPrefix  string              `yaml:"command_prefix"`
	CommandAliases map[string][]string `yaml:"command_aliases"`
	SuperUsers     []int64             `yaml:"super_users"`
	Driver         DriverConfig        `yaml:"driver"`
}

// DriverConfig specifies the OneBot WebSocket driver.
type DriverConfig struct {
	Type   string `yaml:"type"`   // "ws" or "ws-reverse"
	Listen string `yaml:"listen"` // for ws-reverse: "0.0.0.0:6700"
	URL    string `yaml:"url"`    // for ws (forward): "ws://127.0.0.1:6700"
	Token  string `yaml:"token"`  // optional access token
}

// WebConfig holds the admin panel web server settings.
type WebConfig struct {
	Host string     `yaml:"host"`
	Port int        `yaml:"port"`
	Auth AuthConfig `yaml:"auth"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	JWTSecret string `yaml:"jwt_secret"`
}

// DatabaseConfig holds SQLite database settings.
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// MasterdataConfig holds masterdata loading settings.
type MasterdataConfig struct {
	Region            string `yaml:"region"`
	Source            string `yaml:"source"` // moesekai / haruki / 8823 / custom
	URL               string `yaml:"url"`
	FallbackURL       string `yaml:"fallback_url"`
	CustomURL         string `yaml:"custom_url"`
	CustomFallbackURL string `yaml:"custom_fallback_url"`
	LocalPath         string `yaml:"local_path"`
	RefreshInterval   int    `yaml:"refresh_interval"` // seconds
}

// SekaiAPIConfig holds optional SEKAI API client settings.
type SekaiAPIConfig struct {
	Enabled   bool              `yaml:"enabled"`
	BaseURL   string            `yaml:"base_url"`
	Region    string            `yaml:"region"`     // "jp" or "cn"
	Headers   map[string]string `yaml:"headers"`    // optional API request headers
	Timeout   int               `yaml:"timeout"`    // seconds
	RateLimit int               `yaml:"rate_limit"` // requests per minute
}

type SuiteAPIConfig struct {
	Enabled     bool   `yaml:"enabled"`
	URL         string `yaml:"url"`
	Token       string `yaml:"token"`
	Timeout     int    `yaml:"timeout"`
	DefaultMode string `yaml:"default_mode"`
}

type RankingAPIConfig struct {
	BaseURL string `yaml:"base_url"`
	Region  string `yaml:"region"`
	Timeout int    `yaml:"timeout"`
}

// RendererConfig holds the Bun renderer service settings.
type RendererConfig struct {
	Host      string      `yaml:"host"`      // renderer listen host
	Port      int         `yaml:"port"`      // renderer listen port
	Precision float64     `yaml:"precision"` // SVG -> PNG render scale
	Cache     CacheConfig `yaml:"cache"`
}

// CacheConfig holds image cache settings. Non-positive max size means unlimited;
// non-positive TTL means cached assets never expire automatically.
type CacheConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Path      string `yaml:"path"`
	MaxSizeMB int    `yaml:"max_size_mb"`
	TTLHours  int    `yaml:"ttl_hours"`
}

// AssetsConfig holds asset/resource settings.
type AssetsConfig struct {
	Source        string `yaml:"source"`     // moesekai / sekai_best / custom
	Region        string `yaml:"region"`     // cn / jp / tw / kr / en
	Mirror        string `yaml:"mirror"`     // main / backup / overseas / overseas_backup
	CDNSource     string `yaml:"cdn_source"` // legacy mirror key or custom URL
	BaseURL       string `yaml:"base_url"`
	CustomBaseURL string `yaml:"custom_base_url"`
	MusicAliasURL string `yaml:"music_alias_url"`
	StickerPath   string `yaml:"sticker_path"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `yaml:"level"`  // "debug", "info", "warn", "error"
	Format string `yaml:"format"` // "console" or "json"
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Region: RegionJP,
		},
		Bot: BotConfig{
			Nickname:       []string{"moebot"},
			CommandPrefix:  "/",
			CommandAliases: map[string][]string{},
			SuperUsers:     []int64{},
			Driver: DriverConfig{
				Type:   "ws-reverse",
				Listen: "0.0.0.0:6700",
			},
		},
		Web: WebConfig{
			Host: "0.0.0.0",
			Port: 8080,
			Auth: AuthConfig{
				Username: "admin",
			},
		},
		Database: DatabaseConfig{
			Path: "./data/moebot.db",
		},
		Masterdata: MasterdataConfig{
			URL:             "https://sk.exmeaning.com/master",
			FallbackURL:     "https://sekaimaster.exmeaning.com/master",
			LocalPath:       "./data/master",
			RefreshInterval: 3600,
		},
		SekaiAPI: SekaiAPIConfig{
			BaseURL:   "https://seka-api.exmeaning.com",
			Region:    "cn",
			Headers:   map[string]string{},
			Timeout:   10,
			RateLimit: 30,
		},
		SuiteAPI: SuiteAPIConfig{
			Enabled:     true,
			URL:         DefaultSuiteAPIURL,
			Timeout:     10,
			DefaultMode: SuiteModeHaruki,
		},
		RankingAPI: RankingAPIConfig{
			BaseURL: "https://rks.exmeaning.com",
			Region:  "cn",
			Timeout: 10,
		},
		Renderer: RendererConfig{
			Host:      "127.0.0.1",
			Port:      3001,
			Precision: DefaultRendererPrecision,
			Cache: CacheConfig{
				Enabled:   true,
				Path:      "./data/cache",
				MaxSizeMB: 0,
				TTLHours:  0,
			},
		},
		Assets: AssetsConfig{
			Mirror:        AssetMirrorMain,
			CDNSource:     "cn_main",
			MusicAliasURL: "https://moe.exmeaning.com/data/music_alias/music_aliases.json",
			StickerPath:   "./assets/stickers",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "console",
		},
	}
}

// Load reads the config file and merges with defaults.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Config file %s not found, using defaults\n", path)
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	NormalizeConfig(cfg)

	return cfg, nil
}

// Save writes the config to a YAML file.
func Save(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

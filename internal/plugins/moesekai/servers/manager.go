package servers

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/models"
	"moebot-next/internal/plugins/moesekai/ranking"
	"moebot-next/internal/plugins/moesekai/sekai"
	"moebot-next/internal/plugins/moesekai/suite"

	"github.com/rs/zerolog/log"
)

// Runtime contains all per-server resources used by commands and the web panel.
type Runtime struct {
	Region       string
	Label        string
	Enabled      bool
	Profile      config.GameServerConfig
	Store        *masterdata.Store
	Loader       *masterdata.Loader
	Assets       *assets.Resolver
	MusicAliases map[int]assets.MusicAlias
	Sekai        *sekai.Client
	Suite        *suite.Client
	Ranking      *ranking.Client
	LoadError    error
}

// Manager owns all configured game server runtimes.
type Manager struct {
	mu            sync.RWMutex
	defaultRegion string
	runtimes      map[string]*Runtime
}

// NewManager creates per-server runtimes and loads masterdata for enabled regions.
func NewManager(cfg *config.Config) *Manager {
	config.NormalizeConfig(cfg)
	m := &Manager{}
	m.ApplyConfig(cfg)
	return m
}

// ApplyConfig rebuilds runtime settings from config while preserving loaded stores
// for still-enabled regions where possible. Regions that transition from disabled
// to enabled have their masterdata loaded asynchronously so command handlers
// (e.g. /cn查卡) become functional without requiring a manual reload.
func (m *Manager) ApplyConfig(cfg *config.Config) {
	config.NormalizeConfig(cfg)

	m.mu.Lock()

	old := m.runtimes
	m.defaultRegion = config.NormalizeRegion(cfg.Server.Region)
	if m.defaultRegion == "" {
		m.defaultRegion = config.RegionJP
	}
	m.runtimes = make(map[string]*Runtime, len(config.RegionKeys()))

	// Skip auto-bootstrap on the initial setup call (NewManager): the caller
	// follows up with LoadEnabled + StartPeriodicRefresh, so kicking off the
	// same work here would race / spawn duplicate refresh loops.
	isInitial := old == nil
	var newlyEnabled []*Runtime
	for _, region := range config.RegionKeys() {
		profile := config.ResolveGameServerProfile(cfg, region)
		enabled := config.IsEnabled(profile.Enabled)
		if region == m.defaultRegion || region == config.RegionJP {
			enabled = true
		}
		runtime := &Runtime{
			Region:  region,
			Label:   config.RegionLabel(region),
			Enabled: enabled,
			Profile: profile,
		}
		previous := old[region]
		if previous != nil && previous.Store != nil {
			runtime.Store = previous.Store
		}
		if runtime.Store == nil {
			runtime.Store = masterdata.NewStore()
		}
		if runtime.Enabled {
			runtime.Loader = masterdata.NewLoader(profile.Masterdata, runtime.Store, region)
			runtime.Assets, runtime.LoadError = assets.NewResolver(profile.Assets, region)
			if previous != nil && previous.MusicAliases != nil {
				runtime.MusicAliases = previous.MusicAliases
			}
			runtime.Sekai = sekai.NewClient(profile.SekaiAPI)
			runtime.Suite = suite.NewClient(profile.SuiteAPI, region)
			runtime.Ranking = ranking.NewClient(ranking.Config{
				BaseURL: profile.RankingAPI.BaseURL,
				Region:  profile.RankingAPI.Region,
				Timeout: profile.RankingAPI.Timeout,
			})
			// Detect a region that just came online: either it had no previous
			// runtime, was previously disabled, or its store was never populated.
			justEnabled := previous == nil || !previous.Enabled || (previous.Store != nil && previous.Store.LoadedAt().IsZero())
			if !isInitial && justEnabled {
				newlyEnabled = append(newlyEnabled, runtime)
			}
		}
		m.runtimes[region] = runtime
	}

	m.mu.Unlock()

	for _, runtime := range newlyEnabled {
		go m.bootstrapRuntime(runtime)
	}
}

// bootstrapRuntime performs the first masterdata load for a freshly-enabled
// region and starts its periodic refresh, mirroring what LoadEnabled +
// StartPeriodicRefresh do during initial startup.
func (m *Manager) bootstrapRuntime(runtime *Runtime) {
	if runtime == nil || runtime.Loader == nil {
		return
	}
	log.Info().Str("region", runtime.Region).Msg("Bootstrapping newly-enabled region masterdata")
	if err := runtime.Loader.LoadAll(); err != nil {
		runtime.LoadError = err
		log.Warn().Err(err).Str("region", runtime.Region).Msg("Newly-enabled region masterdata load failed")
	} else {
		runtime.LoadError = nil
	}
	m.loadMusicAliases(runtime)
	if interval := runtime.Profile.Masterdata.RefreshInterval; interval > 0 {
		runtime.Loader.StartPeriodicRefresh(time.Duration(interval) * time.Second)
	}
}

// LoadEnabled loads masterdata for all enabled runtimes.
func (m *Manager) LoadEnabled() {
	for _, runtime := range m.EnabledRuntimes() {
		if runtime.Loader == nil {
			continue
		}
		if err := runtime.Loader.LoadAll(); err != nil {
			runtime.LoadError = err
			log.Warn().Err(err).Str("region", runtime.Region).Msg("Initial regional masterdata load failed")
		} else {
			runtime.LoadError = nil
		}
		m.loadMusicAliases(runtime)
	}
}

// StartPeriodicRefresh starts per-region refresh loops.
func (m *Manager) loadMusicAliases(runtime *Runtime) {
	if runtime == nil || !runtime.Enabled {
		return
	}
	url := runtime.Profile.Assets.MusicAliasURL
	aliases, err := assets.LoadMusicAliases(url)
	if err != nil {
		log.Debug().Err(err).Str("region", runtime.Region).Msg("Music alias load failed; music search will use masterdata only")
		return
	}
	runtime.MusicAliases = aliases
	log.Debug().Str("region", runtime.Region).Int("aliases", len(aliases)).Msg("Music aliases loaded")
}

func (m *Manager) StartPeriodicRefresh() {
	for _, runtime := range m.EnabledRuntimes() {
		interval := runtime.Profile.Masterdata.RefreshInterval
		if runtime.Loader != nil && interval > 0 {
			runtime.Loader.StartPeriodicRefresh(time.Duration(interval) * time.Second)
		}
	}
}

// StopPeriodicRefresh stops all refresh loops.
func (m *Manager) StopPeriodicRefresh() {
	for _, runtime := range m.AllRuntimes() {
		if runtime.Loader != nil {
			runtime.Loader.StopPeriodicRefresh()
		}
	}
}

// Default returns the default server runtime.
func (m *Manager) Default() *Runtime {
	return m.Get("")
}

// Get returns the runtime for a region or the default runtime when region is empty/invalid/disabled.
func (m *Manager) Get(region string) *Runtime {
	m.mu.RLock()
	defer m.mu.RUnlock()

	region = config.NormalizeRegion(region)
	if region == "" || !config.IsValidRegion(region) {
		region = m.defaultRegion
	}
	runtime := m.runtimes[region]
	if runtime == nil || !runtime.Enabled {
		runtime = m.runtimes[m.defaultRegion]
	}
	if runtime == nil {
		runtime = m.runtimes[config.RegionJP]
	}
	return runtime
}

// GetExact returns the runtime for a concrete region without falling back when
// the region is disabled. This is used by explicitly regional commands such as
// /tw组卡 so they do not accidentally run against the default server.
func (m *Manager) GetExact(region string) *Runtime {
	m.mu.RLock()
	defer m.mu.RUnlock()

	region = config.NormalizeRegion(region)
	if region == "" || !config.IsValidRegion(region) {
		return nil
	}
	return m.runtimes[region]
}

// ForUser returns the runtime bound to a user, falling back to default JP.
func (m *Manager) ForUser(user *models.User) *Runtime {
	if user == nil || user.ServerRegion == "" {
		return m.Default()
	}
	return m.Get(user.ServerRegion)
}

// Reload reloads masterdata for one region. Empty region reloads the default region.
func (m *Manager) Reload(region string) (*Runtime, error) {
	runtime := m.Get(region)
	if runtime == nil || runtime.Loader == nil {
		return runtime, fmt.Errorf("server %s masterdata loader is not configured", region)
	}
	if err := runtime.Loader.LoadAll(); err != nil {
		runtime.LoadError = err
		return runtime, err
	}
	runtime.LoadError = nil
	m.loadMusicAliases(runtime)
	return runtime, nil
}

// AllRuntimes returns all region runtimes in canonical order.
func (m *Manager) AllRuntimes() []*Runtime {
	m.mu.RLock()
	defer m.mu.RUnlock()

	runtimes := make([]*Runtime, 0, len(m.runtimes))
	for _, region := range config.RegionKeys() {
		if runtime := m.runtimes[region]; runtime != nil {
			runtimes = append(runtimes, runtime)
		}
	}
	return runtimes
}

// EnabledRuntimes returns enabled runtimes in canonical order.
func (m *Manager) EnabledRuntimes() []*Runtime {
	all := m.AllRuntimes()
	out := make([]*Runtime, 0, len(all))
	for _, runtime := range all {
		if runtime.Enabled {
			out = append(out, runtime)
		}
	}
	return out
}

// DefaultRegion returns the canonical default server key.
func (m *Manager) DefaultRegion() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.defaultRegion == "" {
		return config.RegionJP
	}
	return m.defaultRegion
}

// Regions returns configured region keys in sorted order for stable UI maps.
func (m *Manager) Regions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	regions := make([]string, 0, len(m.runtimes))
	for region := range m.runtimes {
		regions = append(regions, region)
	}
	sort.Strings(regions)
	return regions
}

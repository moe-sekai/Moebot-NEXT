package masterdata

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"moebot-next/internal/config"
)

// ---------------------------------------------------------------------------
// loader.go — Fetches masterdata from remote servers with local fallback
//
// Resolution order for each file:
//   1. Primary URL   — cfg.URL/{file}.json          (sk.exmeaning.com)
//   2. Fallback URL  — cfg.FallbackURL/{file}.json   (sekaimaster.exmeaning.com)
//   3. Local cache   — cfg.LocalPath/{file}.json      (data/master/)
//
// Successfully fetched remote files are cached locally for offline use.
// ---------------------------------------------------------------------------

// masterFiles enumerates every JSON file that should be loaded.
var masterFiles = []struct {
	name string                // filename without .json
	dest func(*MasterData) any // returns a pointer to the target slice for json.Unmarshal
}{
	{"cards", func(d *MasterData) any { return &d.Cards }},
	{"musics", func(d *MasterData) any { return &d.Musics }},
	{"musicDifficulties", func(d *MasterData) any { return &d.MusicDifficulties }},
	{"events", func(d *MasterData) any { return &d.Events }},
	{"eventDeckBonuses", func(d *MasterData) any { return &d.EventDeckBonuses }},
	{"gachas", func(d *MasterData) any { return &d.Gachas }},
	{"skills", func(d *MasterData) any { return &d.Skills }},
	{"gameCharacterUnits", func(d *MasterData) any { return &d.CharacterUnits }},
	{"honors", func(d *MasterData) any { return &d.Honors }},
	{"musicVocals", func(d *MasterData) any { return &d.MusicVocals }},
}

// Loader handles fetching and refreshing masterdata from remote servers.
type Loader struct {
	mu            sync.RWMutex
	loadMu        sync.Mutex
	cfg           config.MasterdataConfig
	defaultRegion string
	store         *Store
	client        *http.Client
	done          chan struct{} // closed to signal periodic-refresh goroutine to stop
}

// NewLoader creates a Loader bound to the given Store.
func NewLoader(cfg config.MasterdataConfig, store *Store, defaultRegion ...string) *Loader {
	region := ""
	if len(defaultRegion) > 0 {
		region = defaultRegion[0]
	}
	return &Loader{
		cfg:           cfg,
		defaultRegion: region,
		store:         store,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UpdateConfig atomically swaps the loader configuration used by future loads.
func (l *Loader) UpdateConfig(cfg config.MasterdataConfig, defaultRegion string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cfg = cfg
	l.defaultRegion = defaultRegion
}

func (l *Loader) configSnapshot() (config.MasterdataConfig, string) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.cfg, l.defaultRegion
}

// ---------- Full Load ------------------------------------------------------

// LoadAll fetches every masterdata file and swaps the Store atomically.
// Errors on individual files are logged but do not abort the whole load;
// an error is only returned if zero files could be loaded at all.
func (l *Loader) LoadAll() error {
	l.loadMu.Lock()
	defer l.loadMu.Unlock()

	cfg, defaultRegion := l.configSnapshot()
	resolved, err := config.ResolveMasterdata(cfg, defaultRegion)
	if err != nil {
		return fmt.Errorf("masterdata: resolve source: %w", err)
	}

	data := &MasterData{}
	loaded := 0

	for _, mf := range masterFiles {
		raw, source, err := l.fetchFile(mf.name, resolved)
		if err != nil {
			log.Warn().Err(err).Str("file", mf.name).Msg("Failed to load masterdata file")
			continue
		}

		target := mf.dest(data)
		if err := json.Unmarshal(raw, target); err != nil {
			resetUnmarshalTarget(target)
			log.Warn().Err(err).Str("file", mf.name).Msg("Failed to parse masterdata JSON")
			continue
		}

		log.Debug().Str("file", mf.name).Str("source", source).Msg("Loaded masterdata file")
		loaded++
	}

	if loaded == 0 {
		return fmt.Errorf("masterdata: failed to load any files (tried remote endpoints and local cache)")
	}

	l.store.SetAll(data)

	log.Info().
		Str("provider", resolved.Source).
		Str("region", resolved.Region).
		Int("files_loaded", loaded).
		Int("total_files", len(masterFiles)).
		Int("cards", len(data.Cards)).
		Int("musics", len(data.Musics)).
		Int("events", len(data.Events)).
		Int("gachas", len(data.Gachas)).
		Msg("Masterdata loaded into store")

	return nil
}

// resetUnmarshalTarget clears a slice/map/etc. target after json.Unmarshal
// returns an error. This prevents partially decoded data from entering Store.
func resetUnmarshalTarget(target any) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return
	}
	e := v.Elem()
	if e.CanSet() {
		e.Set(reflect.Zero(e.Type()))
	}
}

// ---------- File Fetching --------------------------------------------------

// fetchFile tries configured remote endpoints → local cache, returning the raw
// JSON bytes and the source label on success.
func (l *Loader) fetchFile(name string, resolved config.ResolvedMasterdata) ([]byte, string, error) {
	filename := name + ".json"

	for _, endpoint := range resolved.Endpoints {
		remoteURL := strings.TrimRight(endpoint.URL, "/") + "/" + filename
		raw, err := l.fetchRemote(remoteURL)
		if err == nil {
			l.cacheLocally(filename, raw, resolved.LocalPath) // best-effort save
			return raw, endpoint.Key, nil
		}
		log.Debug().Err(err).Str("url", remoteURL).Msg("Remote fetch failed, trying next masterdata source")
	}

	raw, err := l.loadLocal(filename, resolved.LocalPath)
	if err == nil {
		return raw, "local", nil
	}

	return nil, "", fmt.Errorf("all sources exhausted for %s: %w", filename, err)
}

// fetchRemote GETs a URL and returns the response body bytes.
func (l *Loader) fetchRemote(url string) ([]byte, error) {
	resp, err := l.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	return body, nil
}

// loadLocal reads a masterdata file from the local cache directory.
func (l *Loader) loadLocal(filename string, localPath string) ([]byte, error) {
	if localPath == "" {
		return nil, fmt.Errorf("local path not configured")
	}
	path := filepath.Join(localPath, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading local file %s: %w", path, err)
	}
	return data, nil
}

// cacheLocally writes fetched data to the local cache directory for offline use.
func (l *Loader) cacheLocally(filename string, data []byte, localPath string) {
	if localPath == "" {
		return
	}

	dir := localPath
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Debug().Err(err).Str("dir", dir).Msg("Could not create local cache directory")
		return
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Debug().Err(err).Str("path", path).Msg("Could not cache masterdata locally")
	}
}

// ---------- Periodic Refresh -----------------------------------------------

// StartPeriodicRefresh begins a background goroutine that reloads all
// masterdata on the given interval. Call StopPeriodicRefresh to shut it down.
func (l *Loader) StartPeriodicRefresh(interval time.Duration) {
	if interval <= 0 {
		log.Info().Msg("Periodic masterdata refresh is disabled (interval <= 0)")
		return
	}

	l.done = make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Info().Dur("interval", interval).Msg("Periodic masterdata refresh started")

		for {
			select {
			case <-ticker.C:
				log.Info().Msg("Refreshing masterdata...")
				if err := l.LoadAll(); err != nil {
					log.Error().Err(err).Msg("Periodic masterdata refresh failed")
				} else {
					log.Info().Msg("Periodic masterdata refresh completed")
				}
			case <-l.done:
				log.Info().Msg("Periodic masterdata refresh stopped")
				return
			}
		}
	}()
}

// StopPeriodicRefresh signals the background refresh goroutine to stop.
// It is safe to call even if StartPeriodicRefresh was never called.
func (l *Loader) StopPeriodicRefresh() {
	if l.done != nil {
		select {
		case <-l.done:
			// already closed
		default:
			close(l.done)
		}
	}
}

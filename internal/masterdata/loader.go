package masterdata

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
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
	cfg    config.MasterdataConfig
	store  *Store
	client *http.Client
	done   chan struct{} // closed to signal periodic-refresh goroutine to stop
}

// NewLoader creates a Loader bound to the given Store.
func NewLoader(cfg config.MasterdataConfig, store *Store) *Loader {
	return &Loader{
		cfg:   cfg,
		store: store,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ---------- Full Load ------------------------------------------------------

// LoadAll fetches every masterdata file and swaps the Store atomically.
// Errors on individual files are logged but do not abort the whole load;
// an error is only returned if zero files could be loaded at all.
func (l *Loader) LoadAll() error {
	data := &MasterData{}
	loaded := 0

	for _, mf := range masterFiles {
		raw, source, err := l.fetchFile(mf.name)
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
		return fmt.Errorf("masterdata: failed to load any files (tried primary, fallback, and local)")
	}

	l.store.SetAll(data)

	log.Info().
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

// fetchFile tries primary URL → fallback URL → local file, returning the raw
// JSON bytes and the source label on success.
func (l *Loader) fetchFile(name string) ([]byte, string, error) {
	filename := name + ".json"

	// 1. Primary URL
	if l.cfg.URL != "" {
		url := l.cfg.URL + "/" + filename
		raw, err := l.fetchRemote(url)
		if err == nil {
			l.cacheLocally(filename, raw) // best-effort save
			return raw, "primary", nil
		}
		log.Debug().Err(err).Str("url", url).Msg("Primary fetch failed, trying fallback")
	}

	// 2. Fallback URL
	if l.cfg.FallbackURL != "" {
		url := l.cfg.FallbackURL + "/" + filename
		raw, err := l.fetchRemote(url)
		if err == nil {
			l.cacheLocally(filename, raw)
			return raw, "fallback", nil
		}
		log.Debug().Err(err).Str("url", url).Msg("Fallback fetch failed, trying local")
	}

	// 3. Local cache
	raw, err := l.loadLocal(filename)
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
func (l *Loader) loadLocal(filename string) ([]byte, error) {
	if l.cfg.LocalPath == "" {
		return nil, fmt.Errorf("local path not configured")
	}
	path := filepath.Join(l.cfg.LocalPath, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading local file %s: %w", path, err)
	}
	return data, nil
}

// cacheLocally writes fetched data to the local cache directory for offline use.
func (l *Loader) cacheLocally(filename string, data []byte) {
	if l.cfg.LocalPath == "" {
		return
	}

	dir := l.cfg.LocalPath
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

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
	{"eventCards", func(d *MasterData) any { return &d.EventCards }},
	{"eventMusics", func(d *MasterData) any { return &d.EventMusics }},
	{"worldBlooms", func(d *MasterData) any { return &d.WorldBlooms }},
	{"virtualLives", func(d *MasterData) any { return &d.VirtualLives }},
	{"gachas", func(d *MasterData) any { return &d.Gachas }},
	{"cardSupplies", func(d *MasterData) any { return &d.CardSupplies }},
	{"skills", func(d *MasterData) any { return &d.Skills }},
	{"gameCharacterUnits", func(d *MasterData) any { return &d.CharacterUnits }},
	{"honors", func(d *MasterData) any { return &d.Honors }},
	{"bondsHonors", func(d *MasterData) any { return &d.BondsHonors }},
	{"bondsHonorWords", func(d *MasterData) any { return &d.BondsHonorWords }},
	{"musicVocals", func(d *MasterData) any { return &d.MusicVocals }},
	{"challengeLiveHighScoreRewards", func(d *MasterData) any { return &d.ChallengeLiveHighScoreRewards }},
	{"resourceBoxes", func(d *MasterData) any { return &d.ResourceBoxes }},
	{"resourceBoxDetails", func(d *MasterData) any { return &d.ResourceBoxDetails }},
	{"characterMissionV2ParameterGroups", func(d *MasterData) any { return &d.CharacterMissionV2ParameterGroups }},
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

// LoadResult 报告一次 LoadAll/Refresh 的执行情况，用于 /update 命令反馈以及
// 上层判断是否真正发生了文件级更新。
type LoadResult struct {
	Region         string // canonical region key
	Source         string // resolved provider key (moesekai/haruki/...)
	Skipped        bool   // true: 远端 dataVersion 与本地一致，未触发文件下载
	VersionChecked bool   // true: 成功获取到远端版本（无论是否跳过）
	OldDataVersion string // 本地缓存的旧 dataVersion，可能为空
	NewDataVersion string // 远端最新 dataVersion，可能为空（探测失败时）
	FilesLoaded    int    // 实际加载的文件数量；Skipped=true 时为 0
}

// LoadAll fetches every masterdata file and swaps the Store atomically.
// 内部会在 Store 已加载且远端 dataVersion 与本地缓存一致时跳过实际拉取。
// Errors on individual files are logged but do not abort the whole load;
// an error is only returned if zero files could be loaded at all.
func (l *Loader) LoadAll() error {
	_, err := l.loadAllInternal(false)
	return err
}

// Refresh 与 LoadAll 等价，但额外返回 LoadResult，并允许通过 force=true 强制
// 跳过版本一致性检查（用于管理员手动 /update 触发的强制刷新场景）。
func (l *Loader) Refresh(force bool) (LoadResult, error) {
	return l.loadAllInternal(force)
}

func (l *Loader) loadAllInternal(force bool) (LoadResult, error) {
	l.loadMu.Lock()
	defer l.loadMu.Unlock()

	cfg, defaultRegion := l.configSnapshot()
	resolved, err := config.ResolveMasterdata(cfg, defaultRegion)
	if err != nil {
		return LoadResult{}, fmt.Errorf("masterdata: resolve source: %w", err)
	}

	result := LoadResult{
		Region: resolved.Region,
		Source: resolved.Source,
	}
	if stored := loadStoredVersion(resolved.LocalPath); stored != nil {
		result.OldDataVersion = stored.DataVersion
	}

	// ---- 1) 探测远端版本（best-effort，失败时回退到全量拉取） ----
	var remoteVersion *RemoteVersionInfo
	if len(resolved.VersionURLs) > 0 {
		if rv, src, verr := l.fetchRemoteVersion(resolved.VersionURLs); verr == nil {
			remoteVersion = rv
			result.VersionChecked = true
			result.NewDataVersion = rv.DataVersion
			log.Debug().
				Str("region", resolved.Region).
				Str("source", src).
				Str("data_version", rv.DataVersion).
				Msg("masterdata: remote version probed")
		} else {
			log.Debug().Err(verr).Str("region", resolved.Region).Msg("masterdata: version probe failed; will fall back to full fetch")
		}
	}

	// ---- 2) 若版本一致且 Store 已加载，跳过文件下载 ----
	if !force && remoteVersion != nil && result.OldDataVersion != "" &&
		result.OldDataVersion == remoteVersion.DataVersion && l.store.IsLoaded() {
		result.Skipped = true
		log.Info().
			Str("region", resolved.Region).
			Str("data_version", remoteVersion.DataVersion).
			Msg("masterdata: data version unchanged, skip refresh")
		return result, nil
	}

	// ---- 3) 全量拉取 ----
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
		return result, fmt.Errorf("masterdata: failed to load any files (tried remote endpoints and local cache)")
	}

	l.store.SetAll(data)
	result.FilesLoaded = loaded

	// moe_costume.json is JP-only, hosted on the same MoeSekai/Exmeaning master mirror.
	// All regions reuse the JP costume database for rendering, so we always fetch
	// from the JP MoeSekai endpoints regardless of the configured region. Failures
	// are non-fatal — costume info is purely cosmetic.
	if costumes, source, err := l.fetchMoeCostumes(resolved.LocalPath); err != nil {
		log.Debug().Err(err).Msg("masterdata: moe_costume.json not loaded (costume info unavailable)")
	} else {
		l.store.SetMoeCostumes(costumes)
		log.Debug().Str("source", source).Int("costumes", len(costumes)).Msg("Loaded moe_costume.json")
	}

	// ---- 4) 持久化最新版本（best-effort） ----
	if remoteVersion != nil {
		if err := saveStoredVersion(resolved.LocalPath, *remoteVersion); err != nil {
			log.Debug().Err(err).Str("region", resolved.Region).Msg("masterdata: persist data_version.json failed")
		}
	}

	log.Info().
		Str("provider", resolved.Source).
		Str("region", resolved.Region).
		Str("data_version", result.NewDataVersion).
		Int("files_loaded", loaded).
		Int("total_files", len(masterFiles)).
		Int("cards", len(data.Cards)).
		Int("musics", len(data.Musics)).
		Int("events", len(data.Events)).
		Int("gachas", len(data.Gachas)).
		Int("virtual_lives", len(data.VirtualLives)).
		Msg("Masterdata loaded into store")

	return result, nil
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

// jpMoeCostumeEndpoints lists the JP-locked MoeSekai/Exmeaning master mirrors that
// host moe_costume.json. Other regions still reuse the JP costume catalogue.
var jpMoeCostumeEndpoints = []string{
	"https://sk.exmeaning.com/master",
	"https://sekaimaster.exmeaning.com/master",
}

// fetchMoeCostumes loads moe_costume.json from the JP MoeSekai mirrors,
// caching the response under localPath/moe_costume.json. Falls back to the
// local cache if all remote sources fail.
func (l *Loader) fetchMoeCostumes(localPath string) ([]MoeCostumeInfo, string, error) {
	const filename = "moe_costume.json"

	for _, base := range jpMoeCostumeEndpoints {
		url := strings.TrimRight(base, "/") + "/" + filename
		raw, err := l.fetchRemote(url)
		if err != nil {
			log.Debug().Err(err).Str("url", url).Msg("moe_costume.json remote fetch failed")
			continue
		}
		costumes, parseErr := parseMoeCostumes(raw)
		if parseErr != nil {
			log.Debug().Err(parseErr).Str("url", url).Msg("moe_costume.json parse failed")
			continue
		}
		l.cacheLocally(filename, raw, localPath)
		return costumes, "jp:" + base, nil
	}

	if raw, err := l.loadLocal(filename, localPath); err == nil {
		costumes, parseErr := parseMoeCostumes(raw)
		if parseErr != nil {
			return nil, "", fmt.Errorf("local moe_costume.json parse: %w", parseErr)
		}
		return costumes, "local", nil
	}

	return nil, "", fmt.Errorf("all sources exhausted for %s", filename)
}

// parseMoeCostumes accepts both the wrapped {costumes:[…]} shape and a bare
// array, mirroring the structure used by Snowy_Viewer's moe_costume.json.
func parseMoeCostumes(raw []byte) ([]MoeCostumeInfo, error) {
	var wrapped MoeCostumeData
	if err := json.Unmarshal(raw, &wrapped); err == nil && wrapped.Costumes != nil {
		return wrapped.Costumes, nil
	}
	var bare []MoeCostumeInfo
	if err := json.Unmarshal(raw, &bare); err != nil {
		return nil, err
	}
	return bare, nil
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

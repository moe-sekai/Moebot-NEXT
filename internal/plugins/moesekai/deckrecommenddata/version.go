package deckrecommenddata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"moebot-next/internal/config"
)

type RemoteVersionInfo struct {
	AppHash          string `json:"appHash,omitempty"`
	SystemProfile    string `json:"systemProfile,omitempty"`
	AppVersion       string `json:"appVersion,omitempty"`
	MultiPlayVersion string `json:"multiPlayVersion,omitempty"`
	AssetVersion     string `json:"assetVersion,omitempty"`
	AppVersionStatus string `json:"appVersionStatus,omitempty"`
	DataVersion      string `json:"dataVersion,omitempty"`
	AssetHash        string `json:"assetHash,omitempty"`
}

type snakeVersionInfo struct {
	AppHash      string `json:"app_hash"`
	AppVersion   string `json:"app_version"`
	DataVersion  string `json:"data_version"`
	AssetVersion string `json:"asset_version"`
	AssetHash    string `json:"asset_hash"`
}

const (
	remoteVersionCacheTTL = 5 * time.Minute
	remoteVersionTimeout  = 2 * time.Second
)

var remoteVersionCache struct {
	sync.Mutex
	items map[string]remoteVersionCacheEntry
}

type remoteVersionCacheEntry struct {
	info      *RemoteVersionInfo
	updatedAt time.Time
}

// SnapshotVersion returns a content-stable version for deck recommend master
// snapshots. It prefers upstream current_version.json / versions.json and
// appends the local compatibility bundle fingerprint. Unlike the previous
// cache-updatedAt fingerprint, this value stays stable across local cache TTL
// refreshes when upstream data did not change.
func SnapshotVersion(resolved config.ResolvedMasterdata) (string, error) {
	remote, err := FetchRemoteVersion(resolved.VersionURLs, remoteVersionTimeout)
	if err != nil {
		return fallbackResolvedVersion(resolved), err
	}
	parts := []string{
		"region=" + strings.TrimSpace(resolved.Region),
		"source=" + strings.TrimSpace(resolved.Source),
		"url=" + strings.TrimSpace(resolved.URL),
		"data=" + strings.TrimSpace(remote.DataVersion),
	}
	if v := strings.TrimSpace(remote.AppVersion); v != "" {
		parts = append(parts, "app="+v)
	}
	if v := strings.TrimSpace(remote.AppHash); v != "" {
		parts = append(parts, "appHash="+v)
	}
	if v := strings.TrimSpace(LocalMasterBundleVersion); v != "" {
		parts = append(parts, "local="+v)
	}
	return strings.Join(parts, "|"), nil
}

func FetchRemoteVersion(versionURLs []string, timeout time.Duration) (*RemoteVersionInfo, error) {
	if len(versionURLs) == 0 {
		return nil, fmt.Errorf("no version endpoint configured")
	}
	cacheKey := remoteVersionCacheKey(versionURLs)
	if cached := cachedRemoteVersion(cacheKey); cached != nil {
		return cached, nil
	}
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	var lastErr error
	for _, rawURL := range versionURLs {
		rawURL = strings.TrimSpace(rawURL)
		if rawURL == "" {
			continue
		}
		resp, err := client.Get(rawURL)
		if err != nil {
			lastErr = err
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("%s returned %d", rawURL, resp.StatusCode)
			continue
		}
		parsed, err := ParseRemoteVersion(body)
		if err != nil {
			lastErr = fmt.Errorf("parse %s: %w", rawURL, err)
			continue
		}
		storeRemoteVersion(cacheKey, parsed)
		return parsed, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no usable version endpoint")
	}
	return nil, lastErr
}

func cachedRemoteVersion(cacheKey string) *RemoteVersionInfo {
	remoteVersionCache.Lock()
	defer remoteVersionCache.Unlock()
	entry, ok := remoteVersionCache.items[cacheKey]
	if !ok || entry.info == nil || time.Since(entry.updatedAt) >= remoteVersionCacheTTL {
		return nil
	}
	clone := *entry.info
	return &clone
}

func storeRemoteVersion(cacheKey string, info *RemoteVersionInfo) {
	if info == nil {
		return
	}
	clone := *info
	remoteVersionCache.Lock()
	if remoteVersionCache.items == nil {
		remoteVersionCache.items = map[string]remoteVersionCacheEntry{}
	}
	remoteVersionCache.items[cacheKey] = remoteVersionCacheEntry{info: &clone, updatedAt: time.Now()}
	remoteVersionCache.Unlock()
}

func remoteVersionCacheKey(versionURLs []string) string {
	urls := append([]string(nil), versionURLs...)
	for i := range urls {
		urls[i] = strings.TrimSpace(urls[i])
	}
	sort.Strings(urls)
	return strings.Join(urls, "|")
}

func ParseRemoteVersion(raw []byte) (*RemoteVersionInfo, error) {
	var v RemoteVersionInfo
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	if strings.TrimSpace(v.DataVersion) == "" {
		var alt snakeVersionInfo
		if err := json.Unmarshal(raw, &alt); err == nil {
			if v.DataVersion == "" {
				v.DataVersion = alt.DataVersion
			}
			if v.AppVersion == "" {
				v.AppVersion = alt.AppVersion
			}
			if v.AssetVersion == "" {
				v.AssetVersion = alt.AssetVersion
			}
			if v.AppHash == "" {
				v.AppHash = alt.AppHash
			}
			if v.AssetHash == "" {
				v.AssetHash = alt.AssetHash
			}
		}
	}
	if strings.TrimSpace(v.DataVersion) == "" {
		return nil, fmt.Errorf("dataVersion missing")
	}
	return &v, nil
}

func fallbackResolvedVersion(resolved config.ResolvedMasterdata) string {
	parts := []string{
		"fallback",
		strings.TrimSpace(resolved.Region),
		strings.TrimSpace(resolved.Source),
		strings.TrimRight(strings.TrimSpace(resolved.URL), "/"),
		strings.TrimSpace(LocalMasterBundleVersion),
	}
	if len(resolved.VersionURLs) > 0 {
		urls := append([]string(nil), resolved.VersionURLs...)
		sort.Strings(urls)
		parts = append(parts, strings.Join(urls, ","))
	}
	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "fallback-" + hex.EncodeToString(h[:8])
}

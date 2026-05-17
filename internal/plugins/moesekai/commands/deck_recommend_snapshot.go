package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/deckrecommenddata"
)

// deckRecommendMasterDataVersion returns a stable upstream-content fingerprint
// for the JP deck recommender snapshot. It intentionally avoids cache entry
// updatedAt timestamps, because local TTL refreshes should not force a 30MB+
// snapshot re-upload when the upstream masterdata did not change.
func deckRecommendMasterDataVersion(resolved config.ResolvedMasterdata) string {
	version, _ := deckrecommenddata.SnapshotVersion(resolved)
	if version != "" {
		return version
	}
	return deckrecommenddata.LocalMasterBundleVersion
}

// deckRecommendMusicMetaVersion returns a content fingerprint for the in-memory
// music meta cache. It stays stable across TTL refreshes with identical data.
func deckRecommendMusicMetaVersion() string {
	musicMetaCache.Lock()
	defer musicMetaCache.Unlock()
	if len(musicMetaCache.data) == 0 {
		return ""
	}
	body, err := json.Marshal(musicMetaCache.data)
	if err != nil {
		return strconv.FormatInt(musicMetaCache.updatedAt.UnixNano(), 16)
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:8])
}

// deckRecommendResolvedMasterdata returns the JP masterdata endpoint
// resolution used by the deck recommender. Callers use this for snapshot
// version fingerprinting alongside buildDeckRecommendMasterData.
func deckRecommendResolvedMasterdata() (config.ResolvedMasterdata, error) {
	jpCfg := config.MasterdataConfig{
		Region: config.RegionJP,
		Source: config.MasterdataSourceMoeSekai,
	}
	return config.ResolveMasterdata(jpCfg, config.RegionJP)
}

package commandparser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/deckrecommenddata"
)

// suiteDebugDeckRecommendMasterDataVersion mirrors the commands package helper
// and uses the same stable upstream-content version, so debug and OneBot paths
// share renderer snapshots instead of invalidating each other on local TTL
// refreshes.
func suiteDebugDeckRecommendMasterDataVersion(resolved config.ResolvedMasterdata) string {
	version, _ := deckrecommenddata.SnapshotVersion(resolved)
	if version != "" {
		return version
	}
	return deckrecommenddata.LocalMasterBundleVersion
}

func suiteDebugDeckRecommendMusicMetaVersion() string {
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

func suiteDebugDeckRecommendResolvedMasterdata() (config.ResolvedMasterdata, error) {
	jpCfg := config.MasterdataConfig{
		Region: config.RegionJP,
		Source: config.MasterdataSourceMoeSekai,
	}
	return config.ResolveMasterdata(jpCfg, config.RegionJP)
}

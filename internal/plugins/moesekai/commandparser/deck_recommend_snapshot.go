package commandparser

import (
	"encoding/binary"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/config"
)

// suiteDebugDeckRecommendMasterDataVersion mirrors the commands package
// helper of the same purpose; it produces a fingerprint over the cached
// JP master data tables that lets the renderer client decide whether to
// re-upload the per-region snapshot. Versions stay stable for cache hits
// and bump when any entry was refetched.
func suiteDebugDeckRecommendMasterDataVersion(resolved config.ResolvedMasterdata) string {
	prefix := resolved.Region + "|" + resolved.Source + "|" + resolved.URL + "|"
	deckMasterDataCache.Lock()
	defer deckMasterDataCache.Unlock()
	type entry struct {
		key       string
		updatedAt int64
	}
	collected := make([]entry, 0, len(deckMasterDataCache.items))
	for k, v := range deckMasterDataCache.items {
		if strings.HasPrefix(k, prefix) {
			collected = append(collected, entry{k, v.updatedAt.UnixNano()})
		}
	}
	if len(collected) == 0 {
		return "empty-" + strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	sort.Slice(collected, func(i, j int) bool { return collected[i].key < collected[j].key })
	h := fnv.New64a()
	for _, e := range collected {
		h.Write([]byte(e.key))
		h.Write([]byte{0})
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(e.updatedAt))
		h.Write(buf[:])
	}
	return strconv.FormatUint(h.Sum64(), 16)
}

func suiteDebugDeckRecommendMusicMetaVersion() string {
	musicMetaCache.Lock()
	defer musicMetaCache.Unlock()
	if musicMetaCache.updatedAt.IsZero() {
		return ""
	}
	return strconv.FormatInt(musicMetaCache.updatedAt.UnixNano(), 16)
}

func suiteDebugDeckRecommendResolvedMasterdata() (config.ResolvedMasterdata, error) {
	jpCfg := config.MasterdataConfig{
		Region: config.RegionJP,
		Source: config.MasterdataSourceMoeSekai,
	}
	return config.ResolveMasterdata(jpCfg, config.RegionJP)
}

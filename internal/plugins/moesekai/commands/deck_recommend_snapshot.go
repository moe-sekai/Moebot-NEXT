package commands

import (
	"encoding/binary"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/config"
)

// deckRecommendMasterDataVersion returns a stable fingerprint for the cached
// master data tables that back the JP deck recommender. The fingerprint
// changes whenever any cache entry was refetched (its updatedAt advances)
// and stays the same while every entry is still served from cache. The
// renderer client uses it to decide whether to re-upload the snapshot.
func deckRecommendMasterDataVersion(resolved config.ResolvedMasterdata) string {
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
		// Cache empty: caller will populate it; use current time so the version
		// changes once data lands and triggers an upload on the next call.
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

// deckRecommendMusicMetaVersion returns a fingerprint for the in-memory music
// meta cache. It bumps every time fetchMusicMetas refreshes the cache.
func deckRecommendMusicMetaVersion() string {
	musicMetaCache.Lock()
	defer musicMetaCache.Unlock()
	if musicMetaCache.updatedAt.IsZero() {
		return ""
	}
	return strconv.FormatInt(musicMetaCache.updatedAt.UnixNano(), 16)
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

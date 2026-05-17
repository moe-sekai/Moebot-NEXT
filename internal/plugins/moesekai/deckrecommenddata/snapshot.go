package deckrecommenddata

import (
	"encoding/binary"
	"hash/fnv"
	"sort"
	"strconv"
)

// LocalMasterBundleVersion is a stable fingerprint for the embedded
// compatibility master tables bundled with the binary. It lets callers include
// these local tables in a larger snapshot version without tying that version to
// process-local cache refresh times.
var LocalMasterBundleVersion = computeLocalMasterBundleVersion()

func computeLocalMasterBundleVersion() string {
	keys := make([]string, 0, len(localMasterKeys))
	for key := range localMasterKeys {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	h := fnv.New64a()
	for _, key := range keys {
		path := localMasterKeys[key]
		body, err := localMasterData.ReadFile(path)
		if err != nil {
			continue
		}
		h.Write([]byte(key))
		h.Write([]byte{0})
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(body)))
		h.Write(size[:])
		h.Write(body)
		h.Write([]byte{0})
	}
	return strconv.FormatUint(h.Sum64(), 16)
}

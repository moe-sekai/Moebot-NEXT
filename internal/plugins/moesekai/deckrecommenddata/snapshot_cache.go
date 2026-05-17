package deckrecommenddata

import (
	"sync"
	"time"
)

type MasterSnapshot struct {
	Region    string
	Version   string
	Data      map[string]any
	Warnings  []string
	UpdatedAt time.Time
}

var masterSnapshotCache struct {
	sync.Mutex
	items map[string]*MasterSnapshot
}

func GetMasterSnapshot(region string, version string) (*MasterSnapshot, bool) {
	if version == "" {
		return nil, false
	}
	key := masterSnapshotCacheKey(region, version)
	masterSnapshotCache.Lock()
	defer masterSnapshotCache.Unlock()
	if masterSnapshotCache.items == nil {
		return nil, false
	}
	snap, ok := masterSnapshotCache.items[key]
	if !ok || snap == nil {
		return nil, false
	}
	return &MasterSnapshot{
		Region:    snap.Region,
		Version:   snap.Version,
		Data:      snap.Data,
		Warnings:  append([]string(nil), snap.Warnings...),
		UpdatedAt: snap.UpdatedAt,
	}, true
}

func StoreMasterSnapshot(region string, version string, data map[string]any, warnings []string) *MasterSnapshot {
	if version == "" {
		return &MasterSnapshot{Region: region, Version: version, Data: data, Warnings: append([]string(nil), warnings...), UpdatedAt: time.Now()}
	}
	snap := &MasterSnapshot{Region: region, Version: version, Data: data, Warnings: append([]string(nil), warnings...), UpdatedAt: time.Now()}
	key := masterSnapshotCacheKey(region, version)
	masterSnapshotCache.Lock()
	if masterSnapshotCache.items == nil {
		masterSnapshotCache.items = map[string]*MasterSnapshot{}
	}
	masterSnapshotCache.items[key] = snap
	masterSnapshotCache.Unlock()
	return snap
}

func masterSnapshotCacheKey(region string, version string) string {
	if region == "" {
		region = "jp"
	}
	return region + "|" + version
}

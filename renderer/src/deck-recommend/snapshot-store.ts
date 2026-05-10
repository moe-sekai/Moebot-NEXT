// Per-region data snapshot store for the deck recommender.
//
// Background: previously every /deck-recommend/calculate request from Go carried
// the full 27 JP master tables (~32 MB) plus the music meta list (~700 KB).
// Now Go pushes those payloads here once via /deck-recommend/snapshot and
// only re-pushes when the upstream version changes. /deck-recommend/calculate
// then reads from this store, dropping ~700ms of JSON marshal/transmit/parse
// per request.
import type { MusicMeta } from "../sekai-calculator/common/music-meta";
import type { MasterDataMap } from "./data-provider";

interface MasterSnapshot {
	data: MasterDataMap;
	version: string;
	keyCount: number;
	updatedAt: number;
}

interface MusicMetasSnapshot {
	data: MusicMeta[];
	version: string;
	count: number;
	updatedAt: number;
}

interface RegionSnapshot {
	master?: MasterSnapshot;
	musicMetas?: MusicMetasSnapshot;
}

const snapshots = new Map<string, RegionSnapshot>();

function normalizeRegion(region: string | undefined | null): string {
	const value = String(region ?? "jp").trim().toLowerCase();
	return value === "" ? "jp" : value;
}

function getOrInit(region: string): RegionSnapshot {
	const key = normalizeRegion(region);
	let snap = snapshots.get(key);
	if (!snap) {
		snap = {};
		snapshots.set(key, snap);
	}
	return snap;
}

export function setMasterSnapshot(region: string, data: MasterDataMap, version: string): MasterSnapshot {
	const snap = getOrInit(region);
	snap.master = {
		data,
		version,
		keyCount: Object.keys(data ?? {}).length,
		updatedAt: Date.now(),
	};
	return snap.master;
}

export function setMusicMetasSnapshot(region: string, data: MusicMeta[], version: string): MusicMetasSnapshot {
	const snap = getOrInit(region);
	snap.musicMetas = {
		data,
		version,
		count: Array.isArray(data) ? data.length : 0,
		updatedAt: Date.now(),
	};
	return snap.musicMetas;
}

export function getMasterSnapshot(region: string): MasterSnapshot | undefined {
	return snapshots.get(normalizeRegion(region))?.master;
}

export function getMusicMetasSnapshot(region: string): MusicMetasSnapshot | undefined {
	return snapshots.get(normalizeRegion(region))?.musicMetas;
}

export interface SnapshotStatusEntry {
	region: string;
	master?: { version: string; keyCount: number; updatedAt: number };
	musicMetas?: { version: string; count: number; updatedAt: number };
}

export function listSnapshotStatus(): SnapshotStatusEntry[] {
	const out: SnapshotStatusEntry[] = [];
	for (const [region, snap] of snapshots.entries()) {
		out.push({
			region,
			master: snap.master ? { version: snap.master.version, keyCount: snap.master.keyCount, updatedAt: snap.master.updatedAt } : undefined,
			musicMetas: snap.musicMetas ? { version: snap.musicMetas.version, count: snap.musicMetas.count, updatedAt: snap.musicMetas.updatedAt } : undefined,
		});
	}
	return out;
}

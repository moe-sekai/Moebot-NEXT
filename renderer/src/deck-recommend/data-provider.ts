import type { MusicMeta } from "../sekai-calculator/common/music-meta";
import type { DataProvider } from "../sekai-calculator/data-provider/data-provider";

export type MasterDataMap = Record<string, unknown[]>;
export type UserDataMap = Record<string, unknown>;

export const USER_DATA_KEYS = [
	"userCards",
	"userBonds",
	"userDecks",
	"userGamedata",
	"userMusics",
	"userMusicResults",
	"userMysekaiMaterials",
	"userAreas",
	"userChallengeLiveSoloDecks",
	"userCharacters",
	"userCharacterMissionV2Statuses",
	"userMysekaiCanvases",
	"userCharacterMissionV2s",
	"userMysekaiFixtureGameCharacterPerformanceBonuses",
	"userMysekaiGates",
	"userWorldBloomSupportDecks",
	"userHonors",
	"userMysekaiCharacterTalks",
	"userChallengeLiveSoloResults",
	"userChallengeLiveSoloStages",
	"userChallengeLiveSoloHighScoreRewards",
	"userEvents",
	"userWorldBlooms",
	"userMusicAchievements",
	"userPlayerFrames",
	"userMaterials",
	"upload_time",
];

export const PRELOAD_MASTER_KEYS = [
	"areaItemLevels",
	"cards",
	"cardMysekaiCanvasBonuses",
	"cardRarities",
	"characterRanks",
	"cardEpisodes",
	"events",
	"eventCards",
	"eventRarityBonusRates",
	"eventDeckBonuses",
	"gameCharacters",
	"gameCharacterUnits",
	"honors",
	"masterLessons",
	"mysekaiGates",
	"mysekaiGateLevels",
	"skills",
	"worldBloomDifferentAttributeBonuses",
	"worldBloomSupportDeckBonuses",
	"worldBloomSupportDeckBonusesWL1",
	"worldBloomSupportDeckBonusesWL2",
	"worldBloomSupportDeckBonusesWL3",
	"worldBloomSupportDeckUnitEventLimitedBonuses",
];

export const MASTER_DATA_BASES: Record<string, string> = {
	jp: "https://sk.exmeaning.com/master",
	cn: "https://sk-cn.exmeaning.com/master",
	tw: "https://storage.sekai.best/sekai-tc-assets/master-data",
};

interface CardParameterEntry {
	id: number;
	cardId: number;
	cardLevel: number;
	cardParameterType: string;
	power: number;
}

interface CardWithParameters {
	id: number;
	cardParameters?: Record<string, number[]> | CardParameterEntry[];
	[key: string]: unknown;
}

function defaultUserDataValue(key: string): unknown {
	if (key === "userGamedata" || key === "upload_time") return null;
	return [];
}

export function applyDefaultUserDataKeys(data: UserDataMap): UserDataMap {
	const normalized: UserDataMap = { ...(data ?? {}) };
	for (const key of USER_DATA_KEYS) {
		if (!(key in normalized)) normalized[key] = defaultUserDataValue(key);
	}
	return normalized;
}

export function transformCards(cards: CardWithParameters[]): CardWithParameters[] {
	return (cards ?? []).map((card) => {
		if (!card.cardParameters || Array.isArray(card.cardParameters)) return card;
		const transformed: CardParameterEntry[] = [];
		for (const [paramType, powers] of Object.entries(card.cardParameters)) {
			if (!Array.isArray(powers)) continue;
			powers.forEach((power, index) => {
				const cardLevel = index + 1;
				const paramIndex = paramType === "param1" ? 1 : paramType === "param2" ? 2 : 3;
				transformed.push({
					id: paramIndex * 10000 + (card.id % 10000) * 100 + cardLevel,
					cardId: card.id,
					cardLevel,
					cardParameterType: paramType,
					power,
				});
			});
		}
		return { ...card, cardParameters: transformed };
	});
}

function normalizeMasterData(data: MasterDataMap): MasterDataMap {
	const normalized: MasterDataMap = { ...(data ?? {}) };
	const rawEventDeckBonuses = normalized["eventDeckBonuses"];
	if (Array.isArray(rawEventDeckBonuses)) {
		normalized["eventDeckBonuses"] = rawEventDeckBonuses.map((item) => {
			if (!item || typeof item !== "object") return item;
			const entry = { ...(item as Record<string, unknown>) };
			if (entry.gameCharacterUnitId === 0) {
				delete entry.gameCharacterUnitId;
			}
			return entry;
		});
	}
	return normalized;
}

export class MemoryDeckRecommendDataProvider implements DataProvider {
	private readonly userData: UserDataMap;
	private readonly masterData: MasterDataMap;
	private readonly musicMetas: MusicMeta[];
	private readonly masterCache = new Map<string, unknown[]>();

	constructor(input: { userData?: UserDataMap; masterData?: MasterDataMap; musicMetas?: MusicMeta[] }) {
		this.userData = applyDefaultUserDataKeys(input.userData ?? {});
		this.masterData = normalizeMasterData(input.masterData ?? {});
		this.musicMetas = input.musicMetas ?? [];
	}

	async getMasterData<T>(key: string): Promise<T[]> {
		const cached = this.masterCache.get(key);
		if (cached) return cached as T[];
		let data = this.masterData[key] ?? [];
		if (key === "cards") data = transformCards(data as CardWithParameters[]) as unknown[];
		this.masterCache.set(key, data);
		return data as T[];
	}

	getMasterDataSyncLength(key: string): number {
		const data = this.masterData[key];
		return Array.isArray(data) ? data.length : -1;
	}

	async getUserDataAll(): Promise<Record<string, any>> {
		return this.userData as Record<string, any>;
	}

	async getUserData<T>(key: string): Promise<T> {
		return this.userData[key] as T;
	}

	async getMusicMeta(): Promise<MusicMeta[]> {
		return this.musicMetas;
	}
}

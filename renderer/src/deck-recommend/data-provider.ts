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

export class MemoryDeckRecommendDataProvider implements DataProvider {
	private readonly userData: UserDataMap;
	private readonly masterData: MasterDataMap;
	private readonly musicMetas: MusicMeta[];

	constructor(input: { userData?: UserDataMap; masterData?: MasterDataMap; musicMetas?: MusicMeta[] }) {
		this.userData = applyDefaultUserDataKeys(input.userData ?? {});
		this.masterData = input.masterData ?? {};
		this.musicMetas = input.musicMetas ?? [];
	}

	async getMasterData<T>(key: string): Promise<T[]> {
		let data = this.masterData[key] ?? [];
		if (key === "cards") data = transformCards(data as CardWithParameters[]) as unknown[];
		return data as T[];
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

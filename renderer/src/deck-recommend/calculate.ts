import { BaseDeckRecommend, RecommendAlgorithm, RecommendTarget, type DeckRecommendConfig } from "../sekai-calculator/deck-recommend/base-deck-recommend";
import { ChallengeLiveDeckRecommend } from "../sekai-calculator/deck-recommend/challenge-live-deck-recommend";
import { EventBonusDeckRecommend } from "../sekai-calculator/deck-recommend/event-bonus-deck-recommend";
import { EventDeckRecommend } from "../sekai-calculator/deck-recommend/event-deck-recommend";
import { SkillReferenceChooseStrategy } from "../sekai-calculator/deck-information/deck-calculator";
import { LiveCalculator, LiveType } from "../sekai-calculator/live-score/live-calculator";
import { MemoryDeckRecommendDataProvider } from "./data-provider";
import type {
	DeckRecommendCalculateRequest,
	DeckRecommendCalculateResponse,
	DeckRecommendOptions,
	DeckRecommendResultDeck,
} from "./types";

function normalizeMode(value: unknown): string {
	switch (String(value || "event").toLowerCase()) {
		case "strongest":
		case "challenge":
		case "bonus":
		case "event": return String(value || "event").toLowerCase();
		default: return "event";
	}
}

function normalizeLiveType(value: unknown): LiveType {
	switch (String(value || "multi").toLowerCase()) {
		case "solo": return LiveType.SOLO;
		case "auto": return LiveType.AUTO;
		case "cheerful": return LiveType.CHEERFUL;
		case "multi":
		default: return LiveType.MULTI;
	}
}

function normalizeAlgorithm(value: unknown): RecommendAlgorithm | "all" {
	switch (String(value || "ga").toLowerCase()) {
		case "dfs": return RecommendAlgorithm.DFS;
		case "all": return "all";
		case "ga":
		default: return RecommendAlgorithm.GA;
	}
}

function normalizeTarget(value: unknown): RecommendTarget {
	switch (String(value || "score").toLowerCase()) {
		case "power": return RecommendTarget.Power;
		case "skill": return RecommendTarget.Skill;
		case "bonus": return RecommendTarget.Bonus;
		case "score":
		default: return RecommendTarget.Score;
	}
}

function normalizeSkillReferenceStrategy(value: unknown): SkillReferenceChooseStrategy {
	switch (String(value || "average").toLowerCase()) {
		case "max": return SkillReferenceChooseStrategy.Max;
		case "min": return SkillReferenceChooseStrategy.Min;
		case "average":
		default: return SkillReferenceChooseStrategy.Average;
	}
}

function selectMusicMeta(req: DeckRecommendCalculateRequest, options: DeckRecommendOptions) {
	const musicId = Number(options.musicId || 0);
	const difficulty = String(options.difficulty || "master").toLowerCase();
	const metas = req.musicMetas ?? [];
	if (musicId === 10000) {
		const filtered = metas.filter((meta) => String(meta.difficulty).toLowerCase() === difficulty);
		const source = filtered.length > 0 ? filtered : metas;
		if (source.length === 0) throw new Error("music metas are empty");
		const sum = (selector: (meta: any) => number) => source.reduce((v, meta) => v + Number(selector(meta) || 0), 0) / source.length;
		const sumArray = (selector: (meta: any) => number[]) => {
			const maxLen = Math.max(...source.map((meta) => (selector(meta) ?? []).length), 0);
			return Array.from({ length: maxLen }, (_, i) => sum((meta) => selector(meta)?.[i] ?? 0));
		};
		return {
			music_id: 10000,
			difficulty,
			music_time: sum((m) => m.music_time),
			event_rate: sum((m) => m.event_rate),
			base_score: sum((m) => m.base_score),
			base_score_auto: sum((m) => m.base_score_auto),
			skill_score_solo: sumArray((m) => m.skill_score_solo),
			skill_score_auto: sumArray((m) => m.skill_score_auto),
			skill_score_multi: sumArray((m) => m.skill_score_multi),
			fever_score: sum((m) => m.fever_score),
			fever_end_time: sum((m) => m.fever_end_time),
			tap_count: Math.round(sum((m) => m.tap_count)),
		};
	}
	const meta = metas.find((item) => Number(item.music_id) === musicId && String(item.difficulty).toLowerCase() === difficulty);
	if (!meta) throw new Error(`music meta not found: ${musicId}/${difficulty}`);
	return meta;
}

function buildConfig(req: DeckRecommendCalculateRequest): DeckRecommendConfig {
	const options = req.options ?? ({} as DeckRecommendOptions);
	const algorithm = normalizeAlgorithm(options.algorithm);
	return {
		musicMeta: selectMusicMeta(req, options),
		limit: Math.max(1, Math.min(Number(options.limit || 3), 10)),
		fixedCards: options.fixedCards ?? [],
		fixedCharacters: options.fixedCharacters ?? [],
		cardConfig: options.cardConfig ?? {},
		algorithm: algorithm === "all" ? RecommendAlgorithm.GA : algorithm,
		timeoutMs: Math.max(1000, Number(options.timeoutMs || 15000)),
		target: normalizeTarget(options.target),
		skillReferenceChooseStrategy: normalizeSkillReferenceStrategy(options.skillReferenceChooseStrategy),
		keepAfterTrainingState: Boolean(options.keepAfterTrainingState),
		bestSkillAsLeader: options.bestSkillAsLeader !== false,
		filterOtherUnit: Boolean(options.filterOtherUnit),
		debugLog: (line: string) => console.debug(`[DeckRecommend] ${line}`),
	};
}

function resultValue(deck: any, target: unknown, mode: string = "event"): { value: number; valueLabel: string } {
	if (mode === "bonus") return { value: Number(deck.eventBonus ?? deck.score ?? 0), valueLabel: "活动加成" };
	if (mode === "challenge") return { value: Number(deck.score || 0), valueLabel: "挑战分数" };
	switch (String(target || (mode === "strongest" ? "power" : "score")).toLowerCase()) {
		case "power": return { value: Number(deck.power?.total || 0), valueLabel: "综合力" };
		case "skill": return { value: Number(deck.multiLiveScoreUp || 0), valueLabel: "实效" };
		case "bonus": return { value: Number(deck.eventBonus || 0), valueLabel: "活动加成" };
		default: return { value: Number(deck.score || 0), valueLabel: "活动PT" };
	}
}

function deckHash(deck: any): string {
	return (deck.cards ?? []).map((card: any) => Number(card.cardId)).sort((a: number, b: number) => a - b).join("-");
}

function friendlyError(error: unknown): string {
	const message = error instanceof Error ? error.message : String(error);
	if (message.includes("music meta not found")) return "缺少该歌曲/难度的分数元数据";
	if (message.includes("Event type not found")) return "活动类型数据缺失";
	if (message.includes("userCards")) return "Suite 公开数据中没有 userCards，请检查抓包公开设置";
	return message;
}

function logDeckRecommendInputSummary(provider: MemoryDeckRecommendDataProvider, req: DeckRecommendCalculateRequest, options: DeckRecommendOptions, liveType: LiveType, musicMeta: any) {
	const masterData = (req.masterData ?? {}) as Record<string, unknown[]>;
	const userData = (req.userData ?? {}) as Record<string, unknown>;
	const listSize = (value: unknown) => Array.isArray(value) ? value.length : -1;
	console.debug("[DeckRecommend] input summary", {
		region: req.region,
		regionLabel: req.regionLabel,
		mode: normalizeMode(options.mode),
		liveType,
		options: {
			eventId: options.eventId,
			musicId: options.musicId,
			difficulty: options.difficulty,
			target: options.target,
			limit: options.limit,
			bestSkillAsLeader: options.bestSkillAsLeader,
			filterOtherUnit: options.filterOtherUnit,
			supportCharacterId: options.supportCharacterId,
			skillReferenceChooseStrategy: options.skillReferenceChooseStrategy,
		},
		musicMeta: musicMeta ? {
			music_id: musicMeta.music_id,
			difficulty: musicMeta.difficulty,
			base_score: musicMeta.base_score,
			base_score_auto: musicMeta.base_score_auto,
			event_rate: musicMeta.event_rate,
			skill_score_solo_len: Array.isArray(musicMeta.skill_score_solo) ? musicMeta.skill_score_solo.length : -1,
			skill_score_multi_len: Array.isArray(musicMeta.skill_score_multi) ? musicMeta.skill_score_multi.length : -1,
			skill_score_auto_len: Array.isArray(musicMeta.skill_score_auto) ? musicMeta.skill_score_auto.length : -1,
		} : null,
		masterCounts: {
			cards: listSize(masterData.cards),
			events: listSize(masterData.events),
			eventCards: listSize(masterData.eventCards),
			eventDeckBonuses: listSize(masterData.eventDeckBonuses),
			eventRarityBonusRates: listSize(masterData.eventRarityBonusRates),
			characterRanks: listSize(masterData.characterRanks),
			cardRarities: listSize(masterData.cardRarities),
			skills: listSize(masterData.skills),
			gameCharacters: listSize(masterData.gameCharacters),
			gameCharacterUnits: listSize(masterData.gameCharacterUnits),
		},
		userCounts: {
			userCards: listSize(userData.userCards),
			userCharacters: listSize(userData.userCharacters),
			userHonors: listSize(userData.userHonors),
			userAreas: listSize(userData.userAreas),
			userDecks: listSize(userData.userDecks),
			userBonds: listSize(userData.userBonds),
		},
		providerSummary: {
			cards: provider.getMasterDataSyncLength?.("cards") ?? listSize(masterData.cards),
			eventDeckBonuses: provider.getMasterDataSyncLength?.("eventDeckBonuses") ?? listSize(masterData.eventDeckBonuses),
		},
	});
}

async function recommendByMode(mode: string, provider: MemoryDeckRecommendDataProvider, req: DeckRecommendCalculateRequest, config: DeckRecommendConfig, liveType: LiveType): Promise<any[]> {
	if (mode === "challenge") {
		const characterId = Number(req.options.challengeCharacterId || req.options.fixedCharacters?.[0] || 0);
		if (!characterId) throw new Error("请输入挑战角色，例如 /挑战组卡 miku");
		const recommender = new ChallengeLiveDeckRecommend(provider);
		return await recommender.recommendChallengeLiveDeck(characterId, config);
	}
	if (mode === "bonus") {
		const eventID = Number(req.options.eventId || 0);
		if (!eventID) throw new Error("eventId is required");
		const targets = req.options.targetBonusList?.length ? req.options.targetBonusList : req.options.targetBonus ? [req.options.targetBonus] : [];
		if (targets.length === 0) throw new Error("请输入目标活动加成，例如 /加成组卡 300");
		const recommender = new EventBonusDeckRecommend(provider);
		const all: any[] = [];
		for (const target of targets) {
			all.push(...await recommender.recommendEventBonusDeck(eventID, target, liveType, { musicMeta: config.musicMeta, member: config.member, cardConfig: config.cardConfig, specificBonuses: [target], timeoutMs: config.timeoutMs, filterOtherUnit: config.filterOtherUnit }, Number(req.options.supportCharacterId || 0)));
		}
		return all;
	}
	if (mode === "strongest") {
		const userCards = await provider.getUserData<any[]>("userCards");
		const recommender = new BaseDeckRecommend(provider);
		return await recommender.recommendHighScoreDeck(userCards, LiveCalculator.getLiveScoreFunction(liveType), config, liveType, {});
	}
	const recommender = new EventDeckRecommend(provider);
	return await recommender.recommendEventDeck(Number(req.options.eventId), liveType, config, Number(req.options.supportCharacterId || 0));
}

function logDeckRecommendDeckSummary(deck: any, req: DeckRecommendCalculateRequest, mode: string, index: number) {
	const cards = Array.isArray(deck?.cards) ? deck.cards : [];
	console.debug("[DeckRecommend] deck summary", {
		mode,
		index,
		target: req.options?.target,
		score: deck?.score,
		eventBonus: deck?.eventBonus,
		supportDeckBonus: deck?.supportDeckBonus,
		multiLiveScoreUp: deck?.multiLiveScoreUp,
		power: deck?.power,
		cards: cards.map((card: any) => ({
			cardId: card?.cardId,
			characterId: card?.characterId,
			attr: card?.attr,
			level: card?.level,
			masterRank: card?.masterRank,
			skillLevel: card?.skillLevel,
			defaultImage: card?.defaultImage,
			power: card?.power,
			eventBonus: card?.eventBonus,
			supportDeckBonus: card?.supportDeckBonus,
			skill: card?.skill,
		})),
	});
}

export async function calculateDeckRecommend(req: DeckRecommendCalculateRequest): Promise<DeckRecommendCalculateResponse> {
	const start = performance.now();
	const warnings: string[] = [];
	try {
		if (!req?.userData) throw new Error("userData is required");
		if (!req?.masterData) throw new Error("masterData is required");
		if (!req?.options?.eventId && normalizeMode(req?.options?.mode) !== "strongest" && normalizeMode(req?.options?.mode) !== "challenge") throw new Error("eventId is required");
		const provider = new MemoryDeckRecommendDataProvider({
			userData: req.userData,
			masterData: req.masterData,
			musicMetas: req.musicMetas,
		});
		const config = buildConfig(req);
		let liveType = normalizeLiveType(req.options.liveType);
		const mode = normalizeMode(req.options.mode);
		if (mode === "event") {
			const events = await provider.getMasterData<any>("events");
			const event = events.find((it) => Number(it.id) === Number(req.options.eventId));
			if (event && String(event.eventType || "").toLowerCase() === "cheerful_carnival" && liveType === LiveType.MULTI) {
				liveType = LiveType.CHEERFUL;
			}
		}
		logDeckRecommendInputSummary(provider, req, req.options, liveType, config.musicMeta);
		const algorithms = String(req.options.algorithm || "ga").toLowerCase() === "all" && mode !== "bonus"
			? [RecommendAlgorithm.GA, RecommendAlgorithm.DFS]
			: [config.algorithm ?? RecommendAlgorithm.GA];
		const merged = new Map<string, any>();
		for (const algorithm of algorithms) {
			const decks = await recommendByMode(mode, provider, req, { ...config, algorithm }, liveType);
			decks.forEach((deck, index) => logDeckRecommendDeckSummary(deck, req, mode, index));
			for (const deck of decks) {
				const hash = deckHash(deck);
				const prev = merged.get(hash);
				if (!prev || resultValue(deck, req.options.target, mode).value > resultValue(prev, req.options.target, mode).value) {
					merged.set(hash, { ...deck, algorithm });
				}
			}
		}
		const decks = Array.from(merged.values()).sort((a, b) => resultValue(b, req.options.target, mode).value - resultValue(a, req.options.target, mode).value).slice(0, config.limit ?? 3);
		const masterCards = await provider.getMasterData<any>("cards");
		const resultDecks: DeckRecommendResultDeck[] = decks.map((deck, index) => {
			const { value, valueLabel } = resultValue(deck, req.options.target, mode);
			return {
			rank: index + 1,
			value,
			valueLabel,
			score: Math.round(Number(deck.score || 0)),
			eventPoint: mode === "event" && String(req.options.target || "score").toLowerCase() === "score" ? Math.round(Number(deck.score || 0)) : undefined,
				eventBonus: deck.eventBonus,
				supportDeckBonus: deck.supportDeckBonus,
				power: deck.power,
				multiLiveScoreUp: deck.multiLiveScoreUp,
			cards: deck.cards.map((card: any) => {
				const master = masterCards.find((it) => Number(it.id) === Number(card.cardId));
				const asset = req.cardAssets?.[Number(card.cardId)] ?? {};
				return {
					...card,
					card: {
						id: card.cardId,
						cardId: card.cardId,
						prefix: master?.prefix,
						characterId: master?.characterId,
						cardRarityType: master?.cardRarityType,
						rarity: master?.cardRarityType,
						attr: master?.attr,
						assetbundleName: master?.assetbundleName,
						defaultImage: card.defaultImage,
						level: card.level,
						masterRank: card.masterRank,
						skillLevel: card.skillLevel,
						...asset,
					},
				};
			}),
		};
		});
		return {
			ok: true,
			region: req.region,
			regionLabel: req.regionLabel,
			costMs: Math.round(performance.now() - start),
			algorithm: String(req.options.algorithm || config.algorithm),
			warnings,
			options: req.options,
			profile: req.profile,
			event: req.event,
			music: req.music,
			decks: resultDecks,
		};
	} catch (error) {
		return {
			ok: false,
			region: req?.region,
			regionLabel: req?.regionLabel,
			costMs: Math.round(performance.now() - start),
			algorithm: String(req?.options?.algorithm || "ga"),
			warnings,
			options: req?.options ?? ({} as DeckRecommendOptions),
			profile: req?.profile,
			event: req?.event,
			music: req?.music,
			decks: [],
			error: friendlyError(error),
		};
	}
}

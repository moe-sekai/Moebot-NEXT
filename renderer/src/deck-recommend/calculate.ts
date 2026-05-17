import { createHash } from "node:crypto";
import { BaseDeckRecommend, RecommendAlgorithm, RecommendTarget, type DeckRecommendConfig } from "../sekai-calculator/deck-recommend/base-deck-recommend";
import { ChallengeLiveDeckRecommend } from "../sekai-calculator/deck-recommend/challenge-live-deck-recommend";
import { EventBonusDeckRecommend } from "../sekai-calculator/deck-recommend/event-bonus-deck-recommend";
import { EventDeckRecommend } from "../sekai-calculator/deck-recommend/event-deck-recommend";
import { MysekaiDeckRecommend } from "../sekai-calculator/deck-recommend/mysekai-deck-recommend";
import { MysekaiEventCalculator } from "../sekai-calculator/mysekai-information/mysekai-event-calculator";
import { SkillReferenceChooseStrategy } from "../sekai-calculator/deck-information/deck-calculator";
import { LiveCalculator, LiveType } from "../sekai-calculator/live-score/live-calculator";
import { MemoryDeckRecommendDataProvider } from "./data-provider";
import { getMasterSnapshot, getMusicMetasSnapshot } from "./snapshot-store";
import type {
	DeckRecommendCalculateRequest,
	DeckRecommendCalculateResponse,
	DeckRecommendOptions,
	DeckRecommendResultDeck,
} from "./types";

const CALCULATE_CACHE_TTL_MS = 5 * 60 * 1000;
const CALCULATE_CACHE_MAX_ENTRIES = 128;

const calculateCache = new Map<string, { expiresAt: number; response: DeckRecommendCalculateResponse }>();

function normalizeMode(value: unknown): string {
	switch (String(value || "event").toLowerCase()) {
		case "strongest":
		case "challenge":
		case "bonus":
		case "mysekai":
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
	// 烤森模式不依赖歌曲元数据进行计分，使用首条可用元数据兜底
	if (normalizeMode(options.mode) === "mysekai") {
		const byDifficulty = metas.find((item) => String(item.difficulty).toLowerCase() === difficulty);
		if (byDifficulty) return byDifficulty;
		if (metas.length > 0) return metas[0];
		throw new Error("music metas are empty");
	}
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
		gaConfig: {
			seed: stableDeckRecommendSeed(req),
			popSize: 2000,
			parentSize: 240,
			eliteSize: 8,
			maxIter: 400,
			maxIterNoImprove: 6,
		},
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
	if (mode === "mysekai") return { value: Number(deck.mysekaiPt ?? deck.score ?? 0), valueLabel: "烤森PT" };
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
	if (mode === "mysekai") {
		const eventID = Number(req.options.eventId || 0);
		if (!eventID) throw new Error("烤森组卡需要指定活动 eventId");
		const recommender = new MysekaiDeckRecommend(provider);
		const decks = await recommender.recommendMysekaiDeck(eventID, config, Number(req.options.supportCharacterId || 0));
		// 重新计算分段后的烤森 PT作为实际得分
		return decks.map((deck: any) => {
			const pt = MysekaiEventCalculator.getDeckMysekaiEventPoint(deck);
			return { ...deck, score: pt.mysekaiEventPoint, mysekaiPt: pt.mysekaiEventPoint };
		});
	}
	if (mode === "strongest") {
		const userCards = await provider.getUserData<any[]>("userCards");
		const recommender = new BaseDeckRecommend(provider);
		return await recommender.recommendHighScoreDeck(userCards, LiveCalculator.getLiveScoreFunction(liveType), config, liveType, {});
	}
	const recommender = new EventDeckRecommend(provider);
	return await recommender.recommendEventDeck(Number(req.options.eventId), liveType, config, Number(req.options.supportCharacterId || 0));
}

function stableDeckRecommendSeed(req: DeckRecommendCalculateRequest): number {
	let hash = 2166136261;
	const update = (value: unknown) => {
		const text = typeof value === "string" ? value : JSON.stringify(value ?? null);
		for (let i = 0; i < text.length; i++) {
			hash ^= text.charCodeAt(i);
			hash = Math.imul(hash, 16777619) >>> 0;
		}
	};
	update(req.region ?? "jp");
	update(req.options ?? {});
	update((req.userData as any)?.upload_time ?? null);
	update((req.userData as any)?.userGamedata?.userId ?? (req.profile as any)?.userId ?? null);
	return hash === 0 ? 1 : hash;
}

function calculateCacheKey(req: DeckRecommendCalculateRequest, masterVersion: string, musicMetasVersion: string): string {
	const payload = JSON.stringify({
		region: req.region ?? "jp",
		masterVersion,
		musicMetasVersion,
		options: req.options ?? {},
		userData: req.userData ?? {},
		cardAssets: req.cardAssets ?? {},
		event: req.event ?? null,
		music: req.music ?? null,
		profile: req.profile ?? null,
	});
	return createHash("sha256").update(payload).digest("hex");
}

function getCachedCalculation(cacheKey: string): DeckRecommendCalculateResponse | undefined {
	const entry = calculateCache.get(cacheKey);
	if (!entry) return undefined;
	if (entry.expiresAt <= Date.now()) {
		calculateCache.delete(cacheKey);
		return undefined;
	}
	calculateCache.delete(cacheKey);
	calculateCache.set(cacheKey, entry);
	return {
		...entry.response,
		costMs: 0,
		trace: { ...(entry.response.trace ?? {}), cache: "hit" },
	};
}

function storeCachedCalculation(cacheKey: string, response: DeckRecommendCalculateResponse) {
	if (!response.ok) return;
	calculateCache.set(cacheKey, {
		expiresAt: Date.now() + CALCULATE_CACHE_TTL_MS,
		response: { ...response, trace: { ...(response.trace ?? {}), cache: "miss" } },
	});
	while (calculateCache.size > CALCULATE_CACHE_MAX_ENTRIES) {
		const oldest = calculateCache.keys().next().value;
		if (oldest === undefined) break;
		calculateCache.delete(oldest);
	}
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
	let stageStart = start;
	const trace: Record<string, number | string> = {};
	const mark = (name: string) => {
		trace[name] = Math.round(performance.now() - stageStart);
		stageStart = performance.now();
	};
	const warnings: string[] = [];
	try {
		if (!req?.userData) throw new Error("userData is required");
		const _mode = normalizeMode(req?.options?.mode);
		if (!req?.options?.eventId && _mode !== "strongest" && _mode !== "challenge") throw new Error("eventId is required");
		// Resolve masterData/musicMetas: prefer inline body (legacy), fall back to
		// the per-region snapshot store populated via /deck-recommend/snapshot.
		const region = String(req.region ?? "jp");
		let masterVersion = "inline";
		let musicMetasVersion = "inline";
		let masterData = req.masterData;
		if (!masterData || Object.keys(masterData).length === 0) {
			const snap = getMasterSnapshot(region);
			if (!snap) throw new Error(`masterData snapshot is not registered for region "${region}"; upload via POST /deck-recommend/snapshot first`);
			masterData = snap.data;
			masterVersion = snap.version;
			req.masterData = masterData;
		}
		let musicMetas = req.musicMetas;
		if (!musicMetas || musicMetas.length === 0) {
			const snap = getMusicMetasSnapshot(region);
			if (snap) {
				musicMetas = snap.data;
				musicMetasVersion = snap.version;
				req.musicMetas = musicMetas;
			}
		}
		mark("snapshotResolveMs");
		const cacheKey = calculateCacheKey(req, masterVersion, musicMetasVersion);
		const cached = getCachedCalculation(cacheKey);
		if (cached) {
			cached.trace = { ...(cached.trace ?? {}), snapshotResolveMs: trace.snapshotResolveMs ?? 0, totalMs: Math.round(performance.now() - start) };
			return cached;
		}
		const provider = new MemoryDeckRecommendDataProvider({
			userData: req.userData,
			masterData,
			musicMetas,
		});
		const config = buildConfig(req);
		mark("providerConfigMs");
		let liveType = normalizeLiveType(req.options.liveType);
		const mode = normalizeMode(req.options.mode);
		if (mode === "event") {
			const events = await provider.getMasterData<any>("events");
			const event = events.find((it) => Number(it.id) === Number(req.options.eventId));
			if (event && String(event.eventType || "").toLowerCase() === "cheerful_carnival" && liveType === LiveType.MULTI) {
				liveType = LiveType.CHEERFUL;
			}
		}
		mark("eventResolveMs");
		logDeckRecommendInputSummary(provider, req, req.options, liveType, config.musicMeta);
		mark("inputLogMs");
		const algorithms = String(req.options.algorithm || "ga").toLowerCase() === "all" && mode !== "bonus" && mode !== "mysekai"
			? [RecommendAlgorithm.GA, RecommendAlgorithm.DFS]
			: [config.algorithm ?? RecommendAlgorithm.GA];
		const merged = new Map<string, any>();
		for (const algorithm of algorithms) {
			const algorithmStart = performance.now();
			const decks = await recommendByMode(mode, provider, req, { ...config, algorithm }, liveType);
			trace[`algorithm_${algorithm}_ms`] = Math.round(performance.now() - algorithmStart);
			decks.forEach((deck, index) => logDeckRecommendDeckSummary(deck, req, mode, index));
			for (const deck of decks) {
				const hash = deckHash(deck);
				const prev = merged.get(hash);
				if (!prev || resultValue(deck, req.options.target, mode).value > resultValue(prev, req.options.target, mode).value) {
					merged.set(hash, { ...deck, algorithm });
				}
			}
		}
		mark("recommendMergeMs");
		const decks = Array.from(merged.values()).sort((a, b) => resultValue(b, req.options.target, mode).value - resultValue(a, req.options.target, mode).value).slice(0, config.limit ?? 3);
		const masterCards = await provider.getMasterData<any>("cards");
		const masterCardsById = new Map(masterCards.map((card: any) => [Number(card.id), card]));
		const resultDecks: DeckRecommendResultDeck[] = decks.map((deck, index) => {
			const { value, valueLabel } = resultValue(deck, req.options.target, mode);
			return {
			rank: index + 1,
			value,
			valueLabel,
			score: Math.round(Number(deck.score || 0)),
			eventPoint: (mode === "event" && String(req.options.target || "score").toLowerCase() === "score") || mode === "mysekai"
				? Math.round(Number(deck.score || 0))
				: undefined,
				eventBonus: deck.eventBonus,
				supportDeckBonus: deck.supportDeckBonus,
				power: deck.power,
				multiLiveScoreUp: deck.multiLiveScoreUp,
			cards: deck.cards.map((card: any) => {
				const master = masterCardsById.get(Number(card.cardId));
				const asset = req.cardAssets?.[Number(card.cardId)] ?? {};
				const rarity = master?.cardRarityType;
				const isTrained = card.defaultImage === "special_training"
					|| (card.defaultImage !== "original" && (rarity === "rarity_3" || rarity === "rarity_4"));
				return {
					...card,
					card: {
						id: card.cardId,
						cardId: card.cardId,
						prefix: master?.prefix,
						characterId: master?.characterId,
						cardRarityType: rarity,
						rarity,
						attr: master?.attr,
						assetbundleName: master?.assetbundleName,
						defaultImage: card.defaultImage,
						isTrained,
						level: card.level,
						masterRank: card.masterRank,
						skillLevel: card.skillLevel,
						...asset,
					},
				};
			}),
		};
		});
		mark("resultBuildMs");
		trace.totalMs = Math.round(performance.now() - start);
		trace.cache = "miss";
		const response: DeckRecommendCalculateResponse = {
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
			trace,
		};
		storeCachedCalculation(cacheKey, response);
		return response;
	} catch (error) {
		trace.totalMs = Math.round(performance.now() - start);
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
			trace,
			error: friendlyError(error),
		};
	}
}

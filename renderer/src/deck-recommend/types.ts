import type { CardConfig } from "../sekai-calculator/card-information/card-calculator";
import type { MusicMeta } from "../sekai-calculator/common/music-meta";
import type { RecommendAlgorithm, RecommendTarget } from "../sekai-calculator/deck-recommend/base-deck-recommend";
import type { SkillReferenceChooseStrategy } from "../sekai-calculator/deck-information/deck-calculator";
import type { LiveType } from "../sekai-calculator/live-score/live-calculator";
import type { MasterDataMap, UserDataMap } from "./data-provider";

export interface DeckRecommendOptions {
	mode?: "event" | "strongest" | "challenge" | "bonus" | "mysekai" | string;
	eventId: number;
	musicId: number;
	difficulty: string;
	liveType: LiveType | string;
	algorithm?: RecommendAlgorithm | string;
	target?: RecommendTarget | string;
	limit?: number;
	timeoutMs?: number;
	fixedCards?: number[];
	fixedCharacters?: number[];
	cardConfig?: Record<string, CardConfig>;
	skillReferenceChooseStrategy?: SkillReferenceChooseStrategy | string;
	keepAfterTrainingState?: boolean;
	bestSkillAsLeader?: boolean;
	challengeCharacterId?: number;
	targetBonus?: number;
	targetBonusList?: number[];
	filterOtherUnit?: boolean;
	supportCharacterId?: number;
}

export interface DeckRecommendCalculateRequest {
	region?: string;
	regionLabel?: string;
	userData: UserDataMap;
	masterData: MasterDataMap;
	musicMetas: MusicMeta[];
	options: DeckRecommendOptions;
	cardAssets?: Record<number, any>;
	musicAssets?: Record<number, any>;
	event?: any;
	music?: any;
	profile?: any;
}

export interface DeckRecommendResultCard {
	cardId: number;
	level: number;
	skillLevel: number;
	masterRank: number;
	defaultImage?: string;
	eventBonus?: number | string;
	supportDeckBonus?: number;
	power?: any;
	skill?: any;
	card?: any;
}

export interface DeckRecommendResultDeck {
	rank: number;
	value?: number;
	valueLabel?: string;
	score: number;
	eventPoint?: number;
	eventBonus?: number;
	supportDeckBonus?: number;
	power: any;
	multiLiveScoreUp: number;
	cards: DeckRecommendResultCard[];
}

export interface DeckRecommendCalculateResponse {
	ok: boolean;
	region?: string;
	regionLabel?: string;
	costMs: number;
	algorithm: string;
	warnings: string[];
	options: DeckRecommendOptions;
	profile?: any;
	event?: any;
	music?: any;
	decks: DeckRecommendResultDeck[];
	error?: string;
}

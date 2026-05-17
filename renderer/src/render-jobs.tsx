import { getAssetBaseUrl } from "../shared";
import { getCardThumbnailCompositeDataUri, getCardThumbnailCompositeLayersFromSvg, getCardThumbnailCompositeSvg, type CardThumbnailCompositeLayer, type CardThumbnailCompositeRequest } from "./card-thumbnail-composites";
import { calculateDeckRecommend } from "./deck-recommend/calculate";
import type { DeckRecommendCalculateRequest, DeckRecommendCalculateResponse } from "./deck-recommend/types";
import { renderWithTrace } from "./engine";
import type { RenderTrace } from "./engine";
import { renderChartSvg } from "./chart-svg-renderer";
import type { ChartSvgRenderTrace } from "./chart-svg-renderer";
import { renderPreviewTemplate } from "./preview";
import type { RenderPreviewResult } from "./preview";
import {
	DeckRecommend,
	AnvoList,
	Best30,
	CardDetail,
	CharacterRankMission,
	CardList,
	ChartDetail,
	ChurnRankingList,
	ForecastRankingList,
	EventInfo,
	EventList,
	GachaInfo,
	GachaList,
	GalleryGrid,
	GalleryList,
	HelpCard,
	SkillCalc,
	MemoryCard,
	MusicDetail,
	MusicList,
	ProfileCard,
	RankingList,
	SuiteCardBox,
	SuitePanel,
	VirtualLiveList,
	WaterTable,
} from "./templates";

export interface RenderRequest {
	template: string;
	data: any;
	width?: number;
	height?: number;
	precision?: number;
}

export interface ChartRenderRequest {
	url?: string;
	svg?: string;
	width?: number;
	precision?: number;
}


let renderPrepareConcurrency = normalizeRenderPrepareConcurrency(process.env.RENDER_PREPARE_CONCURRENCY, 4);

export interface RenderJobConfigUpdate {
	prepareConcurrency?: number;
}

export function setRenderJobConfig(update: RenderJobConfigUpdate): { prepareConcurrency: number } {
	if (update.prepareConcurrency !== undefined) {
		renderPrepareConcurrency = normalizeRenderPrepareConcurrency(update.prepareConcurrency, renderPrepareConcurrency);
	}
	return { prepareConcurrency: renderPrepareConcurrency };
}

function normalizeRenderPrepareConcurrency(value: unknown, fallback: number): number {
	const n = typeof value === "number" ? value : Number(value);
	if (!Number.isFinite(n) || n <= 0) return fallback;
	return Math.max(1, Math.min(32, Math.floor(n)));
}

function parsePositiveNumber(value: unknown, fallback: number): number {
	const numberValue = typeof value === "number" ? value : Number(value);
	return Number.isFinite(numberValue) && numberValue > 0 ? numberValue : fallback;
}

function defaultHelpData() {
	return {
		commands: [
			{
				name: "查卡",
				usage: "/查卡 初音未来",
				description: "搜索卡牌信息，支持角色名/卡名/ID 模糊匹配",
			},
			{
				name: "查曲/查歌",
				usage: "/查歌 千本樱",
				description: "搜索曲目信息，支持别名/日文/罗马音",
			},
			{
				name: "查谱",
				usage: "/查谱 千本樱",
				description: "查询谱面等级与 notes",
			},
			{
				name: "查活动",
				usage: "/查活动 周年",
				description: "查询活动时间、类型、加成等信息",
			},
			{
				name: "查卡池/查扭蛋",
				usage: "/查扭蛋 限定",
				description: "查询扭蛋和 Pick Up 信息",
			},
			{
				name: "绑定",
				usage: "/绑定 123456789",
				description: "绑定 PJSK 游戏账号",
			},
			{ name: "生日", usage: "/生日", description: "查看今日和近期角色生日" },
		],
		version: "0.1.0",
	};
}

function toPngImageUrl(value: any): string | undefined {
	if (typeof value !== "string" || value.length === 0) return undefined;
	return value.replace(/\.webp(?=([?#]|$))/i, ".png");
}

function sanitizeImageUrls<T>(value: T): T {
	if (Array.isArray(value)) {
		return value.map((item) => sanitizeImageUrls(item)) as T;
	}
	if (!value || typeof value !== "object") {
		return value;
	}
	const out: Record<string, any> = {};
	for (const [key, item] of Object.entries(value as Record<string, any>)) {
		out[key] =
			typeof item === "string" && /url$/i.test(key)
				? toPngImageUrl(item)
				: sanitizeImageUrls(item);
	}
	return out as T;
}

function normalizeCard(data: any) {
	return {
		id: data.id ?? data.ID ?? 0,
		prefix: data.prefix ?? data.Prefix ?? "未知卡牌",
		characterName: data.characterName ?? data.CharacterName ?? "未知角色",
		rarity:
			data.rarity ??
			data.cardRarityType ??
			data.CardRarityType ??
			"rarity_unknown",
		attr: data.attr ?? data.Attr ?? "cute",
		thumbnailUrl: toPngImageUrl(data.thumbnailUrl ?? data.ThumbnailURL),
		normalThumbnailUrl: toPngImageUrl(
			data.normalThumbnailUrl ?? data.NormalThumbnailURL,
		),
		trainedThumbnailUrl: toPngImageUrl(
			data.trainedThumbnailUrl ??
				data.TrainedThumbnailURL ??
				data.TrainedThumbnail,
		),
		normalFullUrl: toPngImageUrl(data.normalFullUrl ?? data.NormalFullURL),
		trainedFullUrl: toPngImageUrl(data.trainedFullUrl ?? data.TrainedFullURL),
		assetbundleName: data.assetbundleName ?? data.AssetbundleName,
		characterId: data.characterId ?? data.CharacterID,
		cardRarityType: data.cardRarityType ?? data.CardRarityType,
		assetSource: data.assetSource ?? data.AssetSource,
		power: data.power ?? data.Power,
		skillName: data.skillName ?? data.SkillName,
		skill: data.skill ?? data.Skill,
		trainedSkill: data.trainedSkill ?? data.TrainedSkill,
		costumes: data.costumes ?? data.Costumes ?? [],
		gachaPhrase: data.gachaPhrase ?? data.GachaPhrase,
		supplyType: data.supplyType ?? data.SupplyType,
		compositeThumbnailUrl: data.compositeThumbnailUrl ?? data.CompositeThumbnailURL,
		compositeLayers: data.compositeLayers ?? data.CompositeLayers,
		normalCompositeLayers: data.normalCompositeLayers ?? data.NormalCompositeLayers,
		trainedCompositeLayers: data.trainedCompositeLayers ?? data.TrainedCompositeLayers,
		events: data.events ?? data.Events ?? [],
	};
}


function normalizeMusic(data: any) {
	return {
		id: data.id ?? data.ID ?? 0,
		title: data.title ?? data.Title ?? "未知曲目",
		pronunciation: data.pronunciation ?? data.Pronunciation,
		lyricist: data.lyricist ?? data.Lyricist,
		composer: data.composer ?? data.Composer,
		arranger: data.arranger ?? data.Arranger,
		categories: data.categories ?? data.Categories ?? [],
		assetbundleName: data.assetbundleName ?? data.AssetbundleName,
		jacketUrl: toPngImageUrl(data.jacketUrl ?? data.JacketURL),
		assetSource: data.assetSource ?? data.AssetSource,
		difficulties: data.difficulties ?? data.Difficulties ?? [],
		publishedAt: data.publishedAt ?? data.PublishedAt,
		releasedAt: data.releasedAt ?? data.ReleasedAt,
		durationSec:
			data.durationSec ??
			data.DurationSec ??
			data.secForMusicScoreMaker ??
			data.SecForMusicScoreMaker,
		isNewlyWrittenMusic: data.isNewlyWrittenMusic ?? data.IsNewlyWrittenMusic,
		isFullLength: data.isFullLength ?? data.IsFullLength,
	};
}

function normalizeEvent(data: any) {
	return {
		id: data.id ?? data.ID ?? 0,
		name: data.name ?? data.Name ?? "未知活动",
		eventType: data.eventType ?? data.EventType,
		unit: data.unit ?? data.Unit,
		assetbundleName: data.assetbundleName ?? data.AssetbundleName,
		assetSource: data.assetSource ?? data.AssetSource,
		bannerUrl: toPngImageUrl(data.bannerUrl ?? data.BannerURL),
		logoUrl: toPngImageUrl(data.logoUrl ?? data.LogoURL),
		storyBannerUrl: toPngImageUrl(data.storyBannerUrl ?? data.StoryBannerURL),
		startAt: data.startAt ?? data.StartAt,
		aggregateAt: data.aggregateAt ?? data.AggregateAt,
		closedAt: data.closedAt ?? data.ClosedAt,
		distributionEndAt: data.distributionEndAt ?? data.DistributionEndAt,
		deckBonuses: data.deckBonuses ?? data.DeckBonuses ?? [],
		bonusAttr: data.bonusAttr ?? data.BonusAttr,
		bonusCharacters:
			data.bonusCharacters ??
			data.BonusCharacters ??
			deriveBonusCharacters(data.deckBonuses ?? data.DeckBonuses ?? []),
		bonusCards: (data.bonusCards ?? data.BonusCards ?? []).map(normalizeCard),
		pickupCards: (data.pickupCards ?? data.PickupCards ?? []).map(normalizeCard),
	};
}

function normalizeRankingList(data: any) {
	return {
		title: data.title ?? data.Title ?? "活动榜线",
		subtitle: data.subtitle ?? data.Subtitle,
		rankings: (data.rankings ?? data.Rankings ?? []).map(normalizeRankingEntry),
		eventId: data.eventId ?? data.EventID,
		eventName: data.eventName ?? data.EventName,
		updatedAt: data.updatedAt ?? data.UpdatedAt,
		assetSource: data.assetSource ?? data.AssetSource,
		region: data.region ?? data.Region,
		regionLabel: data.regionLabel ?? data.RegionLabel,
		boardType: data.boardType ?? data.BoardType,
		targetId: data.targetId ?? data.TargetID,
	};
}

function normalizeRankingEntry(entry: any) {
	return {
		...entry,
		leaderCard: entry?.leaderCard ? normalizeSuiteCard(entry.leaderCard) : entry?.LeaderCard ? normalizeSuiteCard(entry.LeaderCard) : undefined,
	};
}

function normalizeWaterTable(data: any) {
	return {
		title: data.title ?? data.Title ?? "查水表",
		subtitle: data.subtitle ?? data.Subtitle,
		entry: data.entry ?? data.Entry ?? {},
		hourlyChurn: data.hourlyChurn ?? data.HourlyChurn ?? [],
		parkingPeriods: data.parkingPeriods ?? data.Parking ?? [],
		eventId: data.eventId ?? data.EventID,
		updatedAt: data.updatedAt ?? data.UpdatedAt,
		regionLabel: data.regionLabel ?? data.RegionLabel,
		boardType: data.boardType ?? data.BoardType,
		targetId: data.targetId ?? data.TargetID,
	};
}

export function renderDataSummary(template: string, data: any) {
	if (template === "character_rank_mission" || template === "cr_mission") {
		return `mode=${data?.mode ?? data?.Mode ?? ""}, character=${data?.characterId ?? data?.CharacterID ?? ""}, mission=${data?.missionType ?? data?.MissionType ?? ""}, rows=${(data?.rows ?? data?.Rows ?? []).length}, allRows=${(data?.allRows ?? data?.AllRows ?? []).length}/${data?.allRowsTotal ?? data?.AllRowsTotal ?? ""}, page=${data?.page ?? data?.Page ?? ""}/${data?.totalPages ?? data?.TotalPages ?? ""}`;
	}
	return "";
}

function normalizeForecastRanking(data: any) {
	return {
		title: data.title ?? data.Title ?? "榜线预测",
		subtitle: data.subtitle ?? data.Subtitle,
		eventId: data.eventId ?? data.EventID,
		eventName: data.eventName ?? data.EventName,
		region: data.region ?? data.Region,
		regionLabel: data.regionLabel ?? data.RegionLabel,
		status: data.status ?? data.Status,
		updatedAt: data.updatedAt ?? data.UpdatedAt,
		items: data.items ?? data.Items ?? [],
	};
}

function normalizeProfile(data: any) {
	return {
		name: data.name ?? data.Name ?? "未知玩家",
		rank: data.rank ?? data.Rank ?? 0,
		userId:
			data.userId ?? data.UserID ?? data.userID ?? data.ID ?? data.id ?? "-",
		twitterId: data.twitterId ?? data.TwitterID,
		bio: data.bio ?? data.Bio,
		signature: data.signature ?? data.Signature,
		totalPower: data.totalPower ?? data.TotalPower,
		characterId: data.characterId ?? data.CharacterID,
		avatarUrl: toPngImageUrl(data.avatarUrl ?? data.AvatarURL),
		assetSource: data.assetSource ?? data.AssetSource,
		stats: data.stats ?? data.Stats,
		musicClearCounts: data.musicClearCounts ?? data.MusicClearCounts,
		characterRanks: data.characterRanks ?? data.CharacterRanks,
		challengeLive: data.challengeLive ?? data.ChallengeLive,
		profileHonors: (data.profileHonors ?? data.ProfileHonors ?? []).map(normalizeProfileHonor),
		leaderCard: data.leaderCard ? normalizeSuiteCard(data.leaderCard) : data.LeaderCard ? normalizeSuiteCard(data.LeaderCard) : undefined,
		deckCards: (data.deckCards ?? data.DeckCards ?? []).map(normalizeSuiteCard),
		honors: (data.honors ?? data.Honors ?? []).map(normalizeProfileHonor),
	};
}

function normalizeProfileHonor(honor: any) {
	honor = honor ?? {};
	return {
		seq: honor.seq ?? honor.Seq,
		honorType: honor.honorType ?? honor.HonorType,
		honorId: honor.honorId ?? honor.HonorID,
		level: honor.level ?? honor.Level,
		name: honor.name ?? honor.Name,
		honorRarity: honor.honorRarity ?? honor.HonorRarity,
		assetbundleName: honor.assetbundleName ?? honor.AssetbundleName,
		imageUrl: toPngImageUrl(honor.imageUrl ?? honor.ImageURL),
		frameUrl: toPngImageUrl(honor.frameUrl ?? honor.FrameURL),
		levelIconUrl: toPngImageUrl(honor.levelIconUrl ?? honor.LevelIconURL),
		levelIcon6Url: toPngImageUrl(honor.levelIcon6Url ?? honor.LevelIcon6URL),
		bondsHonorViewType: honor.bondsHonorViewType ?? honor.BondsHonorViewType,
		bondsHonorWordId: honor.bondsHonorWordId ?? honor.BondsHonorWordID,
		bondsHonorWordAssetbundleName: honor.bondsHonorWordAssetbundleName ?? honor.BondsHonorWordAssetbundleName,
		bondsHonorWordUrl: toPngImageUrl(honor.bondsHonorWordUrl ?? honor.BondsHonorWordURL),
		leftCharacterId: honor.leftCharacterId ?? honor.LeftCharacterID,
		rightCharacterId: honor.rightCharacterId ?? honor.RightCharacterID,
		leftCharacterUrl: toPngImageUrl(honor.leftCharacterUrl ?? honor.LeftCharacterURL),
		rightCharacterUrl: toPngImageUrl(honor.rightCharacterUrl ?? honor.RightCharacterURL),
		leftColor: honor.leftColor ?? honor.LeftColor,
		rightColor: honor.rightColor ?? honor.RightColor,
	};
}

function normalizeGacha(data: any) {
	const pickups =
		data.pickups ??
		data.Pickups ??
		data.gachaPickups ??
		data.GachaPickups ??
		[];
	return {
		id: data.id ?? data.ID ?? 0,
		name: data.name ?? data.Name ?? "未知卡池",
		gachaType: data.gachaType ?? data.GachaType,
		assetbundleName: data.assetbundleName ?? data.AssetbundleName,
		assetSource: data.assetSource ?? data.AssetSource,
		logoUrl: toPngImageUrl(data.logoUrl ?? data.LogoURL),
		bannerUrl: toPngImageUrl(data.bannerUrl ?? data.BannerURL),
		screenUrl: toPngImageUrl(data.screenUrl ?? data.ScreenURL),
		startAt: data.startAt ?? data.StartAt,
		endAt: data.endAt ?? data.EndAt,
		isShowPeriod: data.isShowPeriod ?? data.IsShowPeriod,
		wishSelectCount: data.wishSelectCount ?? data.WishSelectCount,
		pickupCards: normalizeGachaPickupCards(
			data.pickupCards ?? data.PickupCards,
			pickups,
		),
		pickups,
		rates:
			data.rates ??
			data.Rates ??
			data.gachaCardRarityRates ??
			data.GachaCardRarityRates ??
			[],
	};
}

function normalizeCardList(data: any) {
	return {
		title: data.title ?? data.Title ?? "卡牌列表",
		subtitle: data.subtitle ?? data.Subtitle,
		cards: (data.cards ?? data.Cards ?? []).map(normalizeCard),
		page: data.page ?? data.Page,
		totalPages: data.totalPages ?? data.TotalPages,
		total: data.total ?? data.Total,
		assetSource: data.assetSource ?? data.AssetSource,
	};
}

function normalizeMusicList(data: any) {
	return {
		title: data.title ?? data.Title ?? "曲目列表",
		subtitle: data.subtitle ?? data.Subtitle,
		musics: (data.musics ?? data.Musics ?? []).map(normalizeMusic),
		page: data.page ?? data.Page,
		totalPages: data.totalPages ?? data.TotalPages,
		total: data.total ?? data.Total,
		assetSource: data.assetSource ?? data.AssetSource,
	};
}

function normalizeBest30(data: any) {
	return {
		title: data.title ?? data.Title ?? "Best30",
		subtitle: data.subtitle ?? data.Subtitle,
		profile: normalizeSuiteProfile(data.profile ?? data.Profile),
		average: data.average ?? data.Average,
		entries: (data.entries ?? data.Entries ?? []).map(normalizeBest30Entry),
		candidateCount: data.candidateCount ?? data.CandidateCount,
		apCount: data.apCount ?? data.APCount,
		fcCount: data.fcCount ?? data.FCCount,
		missingConstantsCount: data.missingConstantsCount ?? data.MissingConstantsCount,
		totalResultCount: data.totalResultCount ?? data.TotalResultCount,
		region: data.region ?? data.Region,
		regionLabel: data.regionLabel ?? data.RegionLabel,
		updatedAt: data.updatedAt ?? data.UpdatedAt,
		updateText: data.updateText ?? data.UpdateText,
		formula: data.formula ?? data.Formula,
		constantsSource: data.constantsSource ?? data.ConstantsSource,
		assetSource: data.assetSource ?? data.AssetSource,
	};
}

function normalizeBest30Entry(entry: any) {
	entry = entry ?? {};
	return {
		rank: entry.rank ?? entry.Rank,
		musicId: entry.musicId ?? entry.MusicID,
		title: entry.title ?? entry.Title,
		difficulty: entry.difficulty ?? entry.Difficulty,
		difficultyLabel: entry.difficultyLabel ?? entry.DifficultyLabel,
		level: entry.level ?? entry.Level,
		constant: entry.constant ?? entry.Constant,
		userRating: entry.userRating ?? entry.UserRating,
		playResult: entry.playResult ?? entry.PlayResult,
		noteCount: entry.noteCount ?? entry.NoteCount,
		assetbundleName: entry.assetbundleName ?? entry.AssetbundleName,
		jacketUrl: toPngImageUrl(entry.jacketUrl ?? entry.JacketURL),
	};
}

function normalizeEventList(data: any) {
	return {
		title: data.title ?? data.Title ?? "活动列表",
		subtitle: data.subtitle ?? data.Subtitle,
		events: (data.events ?? data.Events ?? []).map(normalizeEvent),
		page: data.page ?? data.Page,
		totalPages: data.totalPages ?? data.TotalPages,
		total: data.total ?? data.Total,
		assetSource: data.assetSource ?? data.AssetSource,
	};
}

function normalizeGachaList(data: any) {
	return {
		title: data.title ?? data.Title ?? "卡池列表",
		subtitle: data.subtitle ?? data.Subtitle,
		gachas: (data.gachas ?? data.Gachas ?? []).map(normalizeGacha),
		page: data.page ?? data.Page,
		totalPages: data.totalPages ?? data.TotalPages,
		total: data.total ?? data.Total,
		assetSource: data.assetSource ?? data.AssetSource,
	};
}

function normalizeDeckRecommend(data: any) {
	return {
		title: data.title ?? data.Title ?? "活动组卡推荐",
		subtitle: data.subtitle ?? data.Subtitle,
		regionLabel: data.regionLabel ?? data.RegionLabel,
		profile: normalizeSuiteProfile(data.profile ?? data.Profile),
		event: data.event ?? data.Event,
		music: data.music ?? data.Music,
		options: data.options ?? data.Options,
		algorithm: data.algorithm ?? data.Algorithm,
		costMs: data.costMs ?? data.CostMS,
		warnings: data.warnings ?? data.Warnings ?? [],
		decks: data.decks ?? data.Decks ?? [],
		assetSource: data.assetSource ?? data.AssetSource,
	};
}

async function prepareDeckRecommend(data: ReturnType<typeof normalizeDeckRecommend>) {
	const cards = (data.decks ?? []).flatMap((deck: any) => (deck.cards ?? []).map((entry: any) => entry.card ?? entry));
	await hydrateCardCompositeDataUrisForCards(cards, {
		assetSource: data.assetSource,
		size: 58,
		allowDownload: true,
	});
	return data;
}

function normalizeSuitePanel(data: any) {
	return {
		title: data.title ?? data.Title ?? "Suite 数据面板",
		subtitle: data.subtitle ?? data.Subtitle,
		profile: normalizeSuiteProfile(data.profile ?? data.Profile),
		stats: data.stats ?? data.Stats ?? [],
		sections: (data.sections ?? data.Sections ?? []).map(normalizeSuiteSection),
		deckCards: (data.deckCards ?? data.DeckCards ?? []).map(normalizeSuiteCard),
		assetSource: data.assetSource ?? data.AssetSource,
	};
}

function normalizeSuiteCardBox(data: any) {
	return {
		title: data.title ?? data.Title ?? "卡牌一览",
		subtitle: data.subtitle ?? data.Subtitle,
		profile: normalizeSuiteProfile(data.profile ?? data.Profile),
		groups: (data.groups ?? data.Groups)?.map((group: any) => ({
			title: group.title ?? group.Title,
			name: group.name ?? group.Name,
			characterName: group.characterName ?? group.CharacterName,
			cards: (group.cards ?? group.Cards ?? []).map(normalizeSuiteCard),
		})),
		cards: (data.cards ?? data.Cards ?? []).map(normalizeSuiteCard),
		options: data.options ?? data.Options,
		assetSource: data.assetSource ?? data.AssetSource,
		total: data.total ?? data.Total,
		ownedTotal: data.ownedTotal ?? data.OwnedTotal,
		page: data.page ?? data.Page,
		totalPages: data.totalPages ?? data.TotalPages,
		pageSize: data.pageSize ?? data.PageSize,
		totalAll: data.totalAll ?? data.TotalAll,
	};
}

function normalizeSuiteSection(section: any) {
	section = section ?? {};
	return {
		title: section.title ?? section.Title,
		subtitle: section.subtitle ?? section.Subtitle,
		kind: section.kind ?? section.Kind,
		note: section.note ?? section.Note,
		columns: section.columns ?? section.Columns,
		items: section.items ?? section.Items,
		extra: section.extra ?? section.Extra,
		rows: (section.rows ?? section.Rows ?? []).map(normalizeSuiteSectionRow),
	};
}

function normalizeSuiteSectionRow(row: any) {
	if (Array.isArray(row)) return row;
	row = row ?? {};
	return {
		...row,
		id: row.id ?? row.ID,
		rank: row.rank ?? row.Rank,
		label: row.label ?? row.Label,
		value: row.value ?? row.Value,
		meta: row.meta ?? row.Meta,
		color: row.color ?? row.Color,
		card: row.card ?? row.Card,
		characterId: row.characterId ?? row.CharacterID,
		musicId: row.musicId ?? row.MusicID,
		eventId: row.eventId ?? row.EventID,
		iconUrl: toPngImageUrl(row.iconUrl ?? row.IconURL),
		imageUrl: toPngImageUrl(row.imageUrl ?? row.ImageURL),
		bannerUrl: toPngImageUrl(row.bannerUrl ?? row.BannerURL),
		logoUrl: toPngImageUrl(row.logoUrl ?? row.LogoURL),
		dateText: row.dateText ?? row.DateText,
		startAt: row.startAt ?? row.StartAt,
		endAt: row.endAt ?? row.EndAt,
		progress: row.progress ?? row.Progress,
		progressMax: row.progressMax ?? row.ProgressMax,
		progressLabel: row.progressLabel ?? row.ProgressLabel,
		extra: row.extra ?? row.Extra,
	};
}

function normalizeSuiteProfile(profile: any) {
	if (!profile) return undefined;
	return {
		name: profile.name ?? profile.Name,
		displayName: profile.displayName ?? profile.DisplayName,
		rank: profile.rank ?? profile.Rank,
		userId: profile.userId ?? profile.UserID ?? profile.userID ?? profile.ID ?? profile.id,
		uid: profile.uid ?? profile.UID,
		bio: profile.bio ?? profile.Bio,
		signature: profile.signature ?? profile.Signature,
		source: profile.source ?? profile.Source,
		updatedAt: profile.updatedAt ?? profile.UpdatedAt,
		uploadTime: profile.uploadTime ?? profile.UploadTime,
		updateText: profile.updateText ?? profile.UpdateText,
		avatarUrl: toPngImageUrl(profile.avatarUrl ?? profile.AvatarURL),
	};
}

function normalizeSuiteCard(card: any) {
	card = card ?? {};
	return {
		id: card.id ?? card.ID,
		cardId: card.cardId ?? card.CardID,
		prefix: card.prefix ?? card.Prefix,
		characterName: card.characterName ?? card.CharacterName,
		rarity: card.rarity ?? card.Rarity,
		cardRarityType: card.cardRarityType ?? card.CardRarityType,
		attr: card.attr ?? card.Attr,
		assetbundleName: card.assetbundleName ?? card.AssetbundleName,
		thumbnailUrl: toPngImageUrl(card.thumbnailUrl ?? card.ThumbnailURL),
		trainedThumbnailUrl: toPngImageUrl(card.trainedThumbnailUrl ?? card.TrainedThumbnailURL),
		compositeThumbnailUrl: card.compositeThumbnailUrl ?? card.CompositeThumbnailURL,
		compositeImageUrl: card.compositeImageUrl ?? card.CompositeImageURL,
		isTrained: card.isTrained ?? card.IsTrained,
		defaultImage: card.defaultImage ?? card.DefaultImage,
		mastery: card.mastery ?? card.Mastery,
		masterRank: card.masterRank ?? card.MasterRank,
		skillLevel: card.skillLevel ?? card.SkillLevel,
		level: card.level ?? card.Level,
		createdAt: card.createdAt ?? card.CreatedAt,
		obtainedAt: card.obtainedAt ?? card.ObtainedAt,
		acquiredAt: card.acquiredAt ?? card.AcquiredAt,
		owned: card.owned ?? card.Owned,
		isOwned: card.isOwned ?? card.IsOwned,
		supplyType: card.supplyType ?? card.SupplyType,
		limitedType: card.limitedType ?? card.LimitedType,
		isLimited: card.isLimited ?? card.IsLimited,
		isBirthday: card.isBirthday ?? card.IsBirthday,
	};
}

async function prepareCardDetail(data: ReturnType<typeof normalizeCard>) {
	await hydrateCardCompositeLayers(data, {
		assetSource: data.assetSource,
		sizes: [128],
		allowDownload: true,
		bothTrainingStates: true,
	});
	return data;
}

async function prepareCardList(data: ReturnType<typeof normalizeCardList>) {
	await hydrateCardCompositeLayersForCards(data.cards ?? [], {
		assetSource: data.assetSource,
		sizes: [112],
		allowDownload: true,
	});
	return data;
}

async function prepareProfileCard(data: ReturnType<typeof normalizeProfile>) {
	await hydrateCardCompositeLayersForCards([data.leaderCard, ...(data.deckCards ?? [])], {
		assetSource: data.assetSource,
		sizes: [112],
		allowDownload: true,
	});
	return data;
}

async function prepareSuitePanel(data: ReturnType<typeof normalizeSuitePanel>) {
	await hydrateCardCompositeLayersForCards(data.deckCards ?? [], {
		assetSource: data.assetSource,
		sizes: [112],
		allowDownload: true,
	});
	return data;
}

async function prepareSuiteCardBox(data: ReturnType<typeof normalizeSuiteCardBox>) {
	const options = data.options ?? {};
	const allCards = [
		...(data.cards ?? []),
		...(data.groups ?? []).flatMap((group: any) => group.cards ?? []),
	];
	options.totalCardsForLayout = allCards.length;
	await hydrateCardCompositeLayersForCards(allCards, {
		assetSource: data.assetSource,
		sizes: [allCards.length >= 80 ? 88 : 112],
		allowDownload: true,
		useBeforeTraining: Boolean(options.useBeforeTraining),
	});
	return data;
}

async function prepareGachaInfo(data: ReturnType<typeof normalizeGacha>) {
	await hydrateCardCompositeLayersForCards(data.pickupCards ?? [], {
		assetSource: data.assetSource,
		sizes: [112],
		allowDownload: true,
	});
	return data;
}

async function prepareRankingList(data: ReturnType<typeof normalizeRankingList>) {
	await hydrateCardCompositeLayersForCards((data.rankings ?? []).map((entry: any) => entry.leaderCard), {
		assetSource: data.assetSource,
		sizes: [46, 64, 88],
		allowDownload: true,
	});
	return data;
}

async function prepareChurnRankingList(data: ReturnType<typeof normalizeRankingList>) {
	await hydrateCardCompositeLayersForCards((data.rankings ?? []).map((entry: any) => entry.leaderCard), {
		assetSource: data.assetSource,
		sizes: [58],
		allowDownload: true,
	});
	return data;
}

async function hydrateCardCompositeDataUrisForCards(cards: any[], options: { assetSource?: any; size: number; allowDownload: boolean }) {
	const entries: Array<{ card: any; source: any }> = [];
	const unique = new Map<string, { card: any; request: CardThumbnailCompositeRequest }>();
	for (const card of (cards ?? []).filter(Boolean)) {
		const trained = shouldUseTrainedThumbnail(card);
		const request = cardCompositeRequest(card, { assetSource: options.assetSource, trained, size: options.size });
		if (!request) continue;
		const key = JSON.stringify(request);
		let item = unique.get(key);
		if (!item) {
			item = { card, request };
			unique.set(key, item);
		}
		entries.push({ card, source: item.card });
	}
	await runPool(Array.from(unique.values()), renderPrepareConcurrency, async (item) => {
		const dataUri = await getCardThumbnailCompositeDataUri(item.request, options.allowDownload);
		if (dataUri) item.card.compositeImageUrl = dataUri;
	});
	for (const entry of entries) {
		if (entry.source !== entry.card && entry.source.compositeImageUrl) {
			entry.card.compositeImageUrl = entry.source.compositeImageUrl;
		}
	}
}

async function hydrateCardCompositeLayersForCards(cards: any[], options: { assetSource?: any; sizes: number[]; allowDownload: boolean; useBeforeTraining?: boolean }) {
	await runPool((cards ?? []).filter(Boolean), renderPrepareConcurrency, (card) => hydrateCardCompositeLayers(card, options));
}

async function hydrateCardCompositeLayers(card: any, options: { assetSource?: any; sizes: number[]; allowDownload: boolean; bothTrainingStates?: boolean; useBeforeTraining?: boolean }) {
	if (!card) return;
	const sizes = Array.from(new Set(options.sizes.filter((size) => Number.isFinite(size) && size > 0)));
	if (sizes.length === 0) return;
	if (options.bothTrainingStates) {
		const [normal, trained] = await Promise.all([
			compositeLayersForCard(card, { ...options, trained: false, size: sizes[0] }),
			compositeLayersForCard(card, { ...options, trained: true, size: sizes[0] }),
		]);
		if (normal) card.normalCompositeLayers = normal;
		if (trained) card.trainedCompositeLayers = trained;
		return;
	}
	const trained = !Boolean(options.useBeforeTraining) && shouldUseTrainedThumbnail(card);
	const bySize = await Promise.all(sizes.map((size) => compositeLayersForCard(card, { ...options, trained, size })));
	const first = bySize.find(Boolean);
	if (first) card.compositeLayers = first;
}

async function compositeLayersForCard(card: any, options: { assetSource?: any; allowDownload: boolean; trained: boolean; size: number }): Promise<CardThumbnailCompositeLayer[] | undefined> {
	const composite = cardCompositeRequest(card, options);
	if (!composite) return undefined;
	const svg = await getCardThumbnailCompositeSvg(composite, options.allowDownload);
	return svg ? getCardThumbnailCompositeLayersFromSvg(svg) : undefined;
}

function cardCompositeRequest(card: any, options: { assetSource?: any; trained: boolean; size: number }): CardThumbnailCompositeRequest | null {
	const rarity = card.cardRarityType ?? card.rarity ?? (card.isBirthday ? "rarity_birthday" : "rarity_1");
	const source = card.assetSource ?? options.assetSource;
	const imageUrl = imageUrlForComposite(card, options.trained, source);
	if (!imageUrl) return null;
	return {
		imageUrl,
		rarity,
		attr: card.attr ?? "cute",
		trained: options.trained,
		size: options.size,
	};
}

function imageUrlForComposite(card: any, trained: boolean, source: any): string | undefined {
	const normal = card.thumbnailUrl ?? card.normalThumbnailUrl ?? (card.assetbundleName ? cardThumbnailUrl(card.assetbundleName, false, source) : undefined);
	const trainedUrl = card.trainedThumbnailUrl ?? (card.assetbundleName ? cardThumbnailUrl(card.assetbundleName, true, source) : undefined);
	return trained ? trainedUrl ?? normal : normal;
}

function shouldUseTrainedThumbnail(card: any): boolean {
	const rarity = card.cardRarityType ?? card.rarity ?? (card.isBirthday ? "rarity_birthday" : "rarity_1");
	if (card.defaultImage === "special_training") return true;
	if (card.defaultImage === "original") return false;
	if (typeof card.isTrained === "boolean") return card.isTrained;
	return rarity === "rarity_3" || rarity === "rarity_4";
}

function cardThumbnailUrl(assetbundleName: string, trained: boolean, source: any): string {
	const base = assetBaseUrl(source).replace(/\/$/, "");
	return `${base}/thumbnail/chara/${assetbundleName}_${trained ? "after_training" : "normal"}.png`;
}

async function runPool<T>(items: T[], concurrency: number, worker: (item: T) => Promise<void>): Promise<void> {
	let nextIndex = 0;
	const workerCount = Math.min(Math.max(1, concurrency), Math.max(1, items.length));
	const workers = Array.from({ length: workerCount }, async () => {
		while (nextIndex < items.length) {
			const currentIndex = nextIndex;
			nextIndex += 1;
			await worker(items[currentIndex] as T);
		}
	});
	await Promise.all(workers);
}

function assetBaseUrl(source: any): string {
	return getAssetBaseUrl(typeof source === "string" && source.trim() ? source.trim() : "main-jp");
}

function normalizeVirtualLiveList(data: any) {
	return {
		title: data.title ?? data.Title ?? "虚拟 Live",
		subtitle: data.subtitle ?? data.Subtitle,
		virtualLives: (data.virtualLives ?? data.VirtualLives ?? []).map(
			normalizeVirtualLive,
		),
		page: data.page ?? data.Page,
		totalPages: data.totalPages ?? data.TotalPages,
		total: data.total ?? data.Total,
		assetSource: data.assetSource ?? data.AssetSource,
	};
}

function normalizeVirtualLive(data: any) {
	return {
		id: data.id ?? data.ID ?? 0,
		name: data.name ?? data.Name ?? "未知虚拟 Live",
		assetbundleName: data.assetbundleName ?? data.AssetbundleName,
		virtualLiveType: data.virtualLiveType ?? data.VirtualLiveType,
		startAt: data.startAt ?? data.StartAt,
		endAt: data.endAt ?? data.EndAt,
		currentStartAt: data.currentStartAt ?? data.CurrentStartAt,
		currentEndAt: data.currentEndAt ?? data.CurrentEndAt,
		living: data.living ?? data.Living,
		restCount: data.restCount ?? data.RestCount,
		schedules: data.schedules ?? data.Schedules ?? [],
		rewards: data.rewards ?? data.Rewards ?? [],
		characters: data.characters ?? data.Characters ?? [],
		assetSource: data.assetSource ?? data.AssetSource,
	};
}

function deriveBonusCharacters(deckBonuses: any): string[] {
	if (!Array.isArray(deckBonuses)) return [];
	const seen = new Set<string>();
	const result: string[] = [];
	for (const bonus of deckBonuses) {
		const name = bonus?.characterName ?? bonus?.CharacterName;
		if (typeof name === "string" && name && !seen.has(name)) {
			seen.add(name);
			result.push(name);
		}
	}
	return result;
}

function normalizeGachaPickupCards(directCards: any, pickups: any): any[] {
	if (Array.isArray(directCards)) {
		return directCards.map(normalizeGachaPickupCard);
	}
	if (!Array.isArray(pickups)) return [];
	return pickups
		.map((pickup: any) => pickup?.card ?? pickup?.Card)
		.filter(Boolean)
		.map(normalizeGachaPickupCard);
}

function normalizeGachaPickupCard(card: any) {
	card = card ?? {};
	const characterId = card.characterId ?? card.CharacterID;
	return {
		id: card.id ?? card.ID ?? card.cardId ?? card.CardID ?? 0,
		prefix: card.prefix ?? card.Prefix,
		characterName:
			card.characterName ?? card.CharacterName ?? `角色 ${characterId ?? "?"}`,
		rarity:
			card.rarity ??
			card.cardRarityType ??
			card.CardRarityType ??
			"rarity_unknown",
		cardRarityType: card.cardRarityType ?? card.CardRarityType,
		attr: card.attr ?? card.Attr ?? "cute",
		assetbundleName: card.assetbundleName ?? card.AssetbundleName,
		characterId,
		thumbnailUrl: toPngImageUrl(card.thumbnailUrl ?? card.ThumbnailURL),
		trainedThumbnailUrl: toPngImageUrl(
			card.trainedThumbnailUrl ?? card.TrainedThumbnailURL,
		),
		isWish: card.isWish ?? card.IsWish ?? true,
		weight: card.weight ?? card.Weight,
	};
}

async function createElement(req: RenderRequest) {
	const data = sanitizeImageUrls(req.data);
	switch (req.template) {
		case "help_card":
		case "help":
			return <HelpCard {...(data ?? defaultHelpData())} />;
		case "skill_calc":
		case "skillcalc":
			return <SkillCalc {...(data ?? {})} />;
		case "autochat_memory":
		case "memory_card":
			return <MemoryCard {...(data ?? {})} />;
		case "card_detail":
		case "card":
			return <CardDetail card={await prepareCardDetail(normalizeCard(data))} />;
		case "card_list":
		case "cards":
			return <CardList {...(await prepareCardList(normalizeCardList(data)))} />;
		case "music_detail":
		case "music":
			return <MusicDetail music={normalizeMusic(data)} />;
		case "music_list":
		case "musics":
			return <MusicList {...normalizeMusicList(data)} />;
		case "best30":
		case "b30":
			return <Best30 {...normalizeBest30(data ?? {})} />;
		case "chart_detail":
		case "chart":
			return <ChartDetail music={normalizeMusic(data)} />;
		case "event_info":
		case "event":
			return <EventInfo event={normalizeEvent(data)} />;
		case "event_list":
		case "events":
			return <EventList {...normalizeEventList(data)} />;
		case "gacha_info":
		case "gacha":
			return <GachaInfo gacha={await prepareGachaInfo(normalizeGacha(data))} />;
		case "gacha_list":
		case "gachas":
			return <GachaList {...normalizeGachaList(data)} />;
		case "gallery_grid":
		case "gallery":
			return <GalleryGrid {...(data ?? { title: "画廊", pics: [] })} />;
		case "gallery_list":
		case "galleries":
			return <GalleryList {...(data ?? { galleries: [] })} />;
		case "virtual_live_list":
		case "virtual-lives":
		case "vlive":
			return <VirtualLiveList {...normalizeVirtualLiveList(data)} />;
		case "profile_card":
		case "profile":
			return <ProfileCard profile={await prepareProfileCard(normalizeProfile(data))} />;
		case "suite_panel":
		case "suite_status":
			return <SuitePanel {...(await prepareSuitePanel(normalizeSuitePanel(data ?? {})))} />;
		case "suite_card_box":
		case "suite_cards":
			return <SuiteCardBox {...(await prepareSuiteCardBox(normalizeSuiteCardBox(data ?? {})))} />;
		case "anvo_list":
		case "anvo":
			return <AnvoList {...(data ?? {})} />;
		case "character_rank_mission":
		case "cr_mission":
			return <CharacterRankMission {...(data ?? {})} />;
		case "deck_recommend":
		case "deck-recommend":
			return <DeckRecommend {...(await prepareDeckRecommend(normalizeDeckRecommend(data ?? {})))} />;
		case "ranking_list":
		case "ranking":
			return <RankingList {...(await prepareRankingList(normalizeRankingList(data)))} />;
		case "churn_ranking_list":
		case "churn_ranking":
			return <ChurnRankingList {...(await prepareChurnRankingList(normalizeRankingList(data)))} />;
		case "water_table":
		case "csb":
			return <WaterTable {...normalizeWaterTable(data)} />;
		case "forecast_ranking_list":
		case "forecast_ranking":
			return <ForecastRankingList {...normalizeForecastRanking(data)} />;
		default:
			return <HelpCard {...defaultHelpData()} />;
	}
}


export interface RenderTemplateJobResult {
	png: Buffer;
	headers: Record<string, string>;
}

export async function renderTemplateJob(body: RenderRequest): Promise<RenderTemplateJobResult> {
	const width = body.width ?? 800;
	const height = body.height;
	const precision = parsePositiveNumber(body.precision, 1.5);
	const element = await createElement(body);
	const trace = await renderWithTrace(element, {
		width,
		height,
		precision,
	});
	const headers = headersForRenderTrace(trace);
	return { headers, png: Buffer.from(trace.png) };
}

export async function renderPreviewJob(input: { id: string; width?: number; height?: number; precision?: number }): Promise<RenderPreviewResult> {
	return renderPreviewTemplate(input.id, {
		...(input.width && input.width > 0 ? { width: input.width } : {}),
		...(input.height && input.height > 0 ? { height: input.height } : {}),
		precision: parsePositiveNumber(input.precision, 1.5),
	});
}

export async function renderChartJob(body: ChartRenderRequest): Promise<ChartSvgRenderTrace> {
	return renderChartSvg({
		url: body.url,
		svg: body.svg,
		width: body.width,
		precision: parsePositiveNumber(body.precision, 1.5),
	});
}

export async function calculateDeckRecommendJob(req: DeckRecommendCalculateRequest): Promise<DeckRecommendCalculateResponse> {
	return calculateDeckRecommend(req);
}

function headersForRenderTrace(trace: RenderTrace): Record<string, string> {
	return {
		"content-type": "image/png",
		"cache-control": "no-store",
		"x-render-total-ms": String(trace.timings.totalMs),
		"x-render-fonts-ms": String(trace.timings.fontsMs),
		"x-render-images-ms": String(trace.timings.imagesMs),
		"x-render-satori-ms": String(trace.timings.satoriMs),
		"x-render-resvg-ms": String(trace.timings.resvgMs),
		"x-render-size-bytes": String(trace.sizeBytes),
		"x-render-image-total": String(trace.imageCache.total),
		"x-render-image-remote": String(trace.imageCache.remote),
		"x-render-image-cache-hits": String(trace.imageCache.hits),
		"x-render-image-cache-misses": String(trace.imageCache.misses),
		"x-render-image-cache-errors": String(trace.imageCache.errors),
	};
}

import {
	getCardThumbnailUrl,
	getGachaLogoUrl,
	getMusicJacketUrl,
} from "../shared";
import {
	renderWithTrace,
	type RenderOptions,
	type RenderTrace,
} from "./engine";
import { Best30 } from "./templates/Best30";
import { CardDetail } from "./templates/CardDetail";
import { CardList } from "./templates/CardList";
import { MusicDetail } from "./templates/MusicDetail";
import { MusicList } from "./templates/MusicList";
import { ChartDetail } from "./templates/ChartDetail";
import { EventInfo } from "./templates/EventInfo";
import { EventList } from "./templates/EventList";
import { RankingList } from "./templates/RankingList";
import { ProfileCard } from "./templates/ProfileCard";
import { getBondsHonorCharacterUrl, getHonorBgUrl, getHonorFrameUrl, getHonorLevelIconUrl } from "../shared";
import { HelpCard } from "./templates/HelpCard";
import { GachaInfo } from "./templates/GachaInfo";
import { GachaList } from "./templates/GachaList";
import { GachaResult } from "./templates/GachaResult";
import { VirtualLiveList } from "./templates/VirtualLiveList";
import { SuitePanel } from "./templates/SuitePanel";
import { SuiteCardBox } from "./templates/SuiteCardBox";

export type RenderPreviewStatus = "ready" | "draft";

export interface RenderPreviewMeta {
	id: string;
	name: string;
	description: string;
	command: string;
	templatePath: string;
	viewerSource: string;
	status: RenderPreviewStatus;
	width: number;
	height: number;
}

export interface RenderPreviewResult {
	meta: RenderPreviewMeta;
	trace: RenderTrace;
}

const PREVIEW_META: RenderPreviewMeta[] = [
	{
		id: "card-detail",
		name: "查卡详情图",
		description:
			"用于 /查卡 的卡牌详情图片，展示花前/花后卡面、卡牌缩略图、角色、稀有度、属性与综合力。",
		command: "/查卡 [关键词]",
		templatePath: "packages/renderer/src/templates/CardDetail.tsx",
		viewerSource:
			"Snowy Viewer: lib/assets.ts + components/cards/SekaiCardThumbnail.tsx",
		status: "ready",
		width: 800,
		height: 620,
	},
	{
		id: "card-list",
		name: "查卡列表图",
		description: "用于 /查卡 过滤器命中多张卡时的列表图片。",
		command: "/查卡 miku 4星",
		templatePath: "packages/renderer/src/templates/CardList.tsx",
		viewerSource: "Moebot Renderer: card list payload",
		status: "ready",
		width: 800,
		height: 760,
	},
	{
		id: "music-detail",
		name: "查曲详情图",
		description:
			"用于 /查曲 的曲目信息图片，展示远程 jacket、分类、作者信息与彩色难度表。",
		command: "/查曲 [关键词]",
		templatePath: "packages/renderer/src/templates/MusicDetail.tsx",
		viewerSource:
			"Snowy Viewer: components/music/MusicItem.tsx + lib/assets.ts",
		status: "ready",
		width: 800,
		height: 650,
	},
	{
		id: "music-list",
		name: "查曲列表图",
		description: "用于 /查曲 leak、多 ID 和活动关联曲目列表。",
		command: "/查曲 leak",
		templatePath: "packages/renderer/src/templates/MusicList.tsx",
		viewerSource: "Moebot Renderer: music list payload",
		status: "ready",
		width: 800,
		height: 760,
	},
	{
		id: "chart-detail",
		name: "查谱详情图",
		description:
			"用于 /查谱 的谱面详情图片，展示曲绘、难度、等级、Notes 与歌曲长度。",
		command: "/查谱 [关键词/ID]",
		templatePath: "packages/renderer/src/templates/ChartDetail.tsx",
		viewerSource: "Moebot Renderer: MusicDetail payload + ChartDetail.tsx",
		status: "ready",
		width: 800,
		height: 650,
	},
	{
		id: "event-info",
		name: "活动信息图",
		description:
			"用于 /查活动 的活动卡片，展示远程活动 banner、logo、角色图、状态、团组与时间范围。",
		command: "/查活动 [关键词/ID]",
		templatePath: "packages/renderer/src/templates/EventInfo.tsx",
		viewerSource:
			"Snowy Viewer: components/events/EventItem.tsx + lib/assets.ts",
		status: "ready",
		width: 800,
		height: 720,
	},
	{
		id: "event-list",
		name: "活动列表图",
		description: "用于 /查活动 当前、年份、leak 等列表查询。",
		command: "/查活动 2025",
		templatePath: "packages/renderer/src/templates/EventList.tsx",
		viewerSource: "Moebot Renderer: event list payload",
		status: "ready",
		width: 800,
		height: 760,
	},
	{
		id: "gacha-info",
		name: "卡池信息图",
		description:
			"用于 /查卡池 的卡池图片，展示远程卡池 logo、banner/screen、开放时间与 pickup 卡。",
		command: "/查卡池 [关键词/ID]",
		templatePath: "packages/renderer/src/templates/GachaInfo.tsx",
		viewerSource:
			"Snowy Viewer: components/gacha/GachaItem.tsx + lib/assets.ts",
		status: "ready",
		width: 800,
		height: 760,
	},
	{
		id: "gacha-list",
		name: "卡池列表图",
		description: "用于 /查扭蛋 当前、fes、card123 等列表查询。",
		command: "/查扭蛋 当前",
		templatePath: "packages/renderer/src/templates/GachaList.tsx",
		viewerSource: "Moebot Renderer: gacha list payload",
		status: "ready",
		width: 800,
		height: 760,
	},
	{
		id: "virtual-live-list",
		name: "虚拟 Live 列表图",
		description: "用于 /查演唱会 的近期虚拟 Live 列表图片。",
		command: "/查演唱会",
		templatePath: "packages/renderer/src/templates/VirtualLiveList.tsx",
		viewerSource: "Moebot Renderer: virtual live payload",
		status: "ready",
		width: 800,
		height: 760,
	},
	{
		id: "ranking-list",
		name: "实时排行图",
		description:
			"用于 /排行 的排行榜图片，展示排名、玩家名、队长卡/头像、分数与变动。",
		command: "/排行 [排名]",
		templatePath: "packages/renderer/src/templates/RankingList.tsx",
		viewerSource: "Snowy Viewer: components/realtime-ranking/",
		status: "ready",
		width: 800,
		height: 760,
	},
	{
		id: "profile-card",
		name: "个人信息图",
		description:
			"用于 /个人信息 的用户画像卡片，展示头像、昵称、Rank、简介、统计与队伍卡。",
		command: "/个人信息",
		templatePath: "packages/renderer/src/templates/ProfileCard.tsx",
		viewerSource:
			"Snowy Viewer: components/profile/ + CardItem thumbnail stack",
		status: "ready",
		width: 800,
		height: 760,
	},
	{
		id: "gacha-result",
		name: "抽卡结果图",
		description:
			"用于 /抽卡模拟 的十连结果图片，复用 Snowy 卡牌缩略图层级展示属性、稀有度、花后与 NEW 标记。",
		command: "/抽卡模拟",
		templatePath: "packages/renderer/src/templates/GachaResult.tsx",
		viewerSource:
			"Snowy Viewer: components/cards/CardItem.tsx + SekaiCardThumbnail.tsx",
		status: "ready",
		width: 800,
		height: 650,
	},
	{
		id: "suite-panel",
		name: "Suite 数据面板",
		description: "通用 Suite 状态面板，展示玩家资料、当前队伍、统计卡与分段数据。",
		command: "/suite status",
		templatePath: "packages/renderer/src/templates/SuitePanel.tsx",
		viewerSource: "Moebot Renderer: Suite payload",
		status: "ready",
		width: 800,
		height: 760,
	},
	{
		id: "best30",
		name: "Best30 / b30",
		description: "基于 Haruki Suite AP/FC 成绩与社区定数表的 Best30 分享图。",
		command: "/b30",
		templatePath: "packages/renderer/src/templates/Best30.tsx",
		viewerSource: "pjsk.moe /my-musics Best30ShareImage + Moebot Renderer",
		status: "ready",
		width: 800,
		height: 1320,
	},
	{
		id: "suite-card-box",
		name: "Suite 卡牌一览",
		description: "Suite 持有卡牌盒子，支持按角色分组、未持有遮罩与限定/生日标记。",
		command: "/suite cards",
		templatePath: "packages/renderer/src/templates/SuiteCardBox.tsx",
		viewerSource: "Moebot Renderer: Suite card collection payload",
		status: "ready",
		width: 800,
		height: 980,
	},
	{
		id: "help-card",
		name: "帮助菜单图",
		description:
			"用于 /帮助 的指令列表图片，展示基础查询、个人数据与娱乐指令。",
		command: "/帮助",
		templatePath: "packages/renderer/src/templates/HelpCard.tsx",
		viewerSource: "Moebot Console / command registry",
		status: "ready",
		width: 800,
		height: 880,
	},
];

export function listRenderPreviews(): RenderPreviewMeta[] {
	return PREVIEW_META.map((item) => ({ ...item }));
}

export async function renderPreviewTemplate(
	id: string,
	options: RenderOptions = {},
): Promise<RenderPreviewResult> {
	const meta = PREVIEW_META.find((item) => item.id === id);
	if (!meta) {
		throw new Error(`Unknown render preview template: ${id}`);
	}

	const trace = await renderWithTrace(createPreviewElement(id), {
		width: meta.width,
		// Height is not passed — Satori auto-computes height from template content.
		...options,
	});

	return {
		meta: { ...meta },
		trace,
	};
}

export function createPreviewElementForTest(id: string) {
	return createPreviewElement(id);
}

function createPreviewElement(id: string) {
	switch (id) {
		case "card-detail":
			return (
				<CardDetail
					card={{
						id: 1204,
						prefix: "在舞台上绽放的微笑",
						characterName: "初音未来",
						rarity: "rarity_4",
						cardRarityType: "rarity_4",
						attr: "cute",
						assetbundleName: "res001_no003",
						assetSource: "main-jp",
						power: 33956,
						skillName: "5秒内得分提升 120%",
						gachaPhrase: "要把这份心情，传达到舞台的每一个角落。",
						supplyType: "CFES限定",
					}}
				/>
			);
		case "card-list":
			return (
				<CardList
					title="卡牌查询"
					subtitle="关键词：miku 4星"
					page={1}
					totalPages={2}
					total={18}
					assetSource="main-jp"
					cards={[
						{
							id: 1204,
							prefix: "在舞台上绽放的微笑",
							characterName: "初音未来",
							rarity: "rarity_4",
							cardRarityType: "rarity_4",
							attr: "cute",
							assetbundleName: "res001_no003",
							supplyType: "常驻",
						},
						{
							id: 1301,
							prefix: "闪闪发光的歌声",
							characterName: "镜音铃",
							rarity: "rarity_4",
							cardRarityType: "rarity_4",
							attr: "happy",
							assetbundleName: "res002_no003",
							supplyType: "CFES限定",
						},
						{
							id: 1310,
							prefix: "街角的相遇",
							characterName: "巡音流歌",
							rarity: "rarity_4",
							cardRarityType: "rarity_4",
							attr: "cool",
							assetbundleName: "res003_no003",
							supplyType: "WorldLink限定",
						},
					]}
				/>
			);
		case "music-detail":
			return (
				<MusicDetail
					music={{
						id: 1,
						title: "Tell Your World",
						pronunciation: "テル ユア ワールド",
						lyricist: "kz",
						composer: "kz",
						arranger: "kz",
						categories: ["mv", "original"],
						assetbundleName: "jacket_s_001",
						assetSource: "main-jp",
						jacketUrl: getMusicJacketUrl("jacket_s_001", "main-jp"),
						publishedAt: Date.UTC(2020, 8, 30, 6, 0, 0),
						durationSec: 127,
						difficulties: [
							{ musicDifficulty: "easy", playLevel: 5, totalNoteCount: 158 },
							{ musicDifficulty: "normal", playLevel: 10, totalNoteCount: 305 },
							{ musicDifficulty: "hard", playLevel: 16, totalNoteCount: 719 },
							{ musicDifficulty: "expert", playLevel: 22, totalNoteCount: 961 },
							{
								musicDifficulty: "master",
								playLevel: 26,
								totalNoteCount: 1147,
							},
						],
					}}
				/>
			);
		case "music-list":
			return (
				<MusicList
					title="未发布曲目"
					subtitle="leak"
					page={1}
					totalPages={1}
					total={3}
					assetSource="main-jp"
					musics={[
						{
							id: 1,
							title: "Tell Your World",
							pronunciation: "テル ユア ワールド",
							composer: "kz",
							assetbundleName: "jacket_s_001",
							difficulties: [
								{ musicDifficulty: "easy", playLevel: 5 },
								{ musicDifficulty: "expert", playLevel: 22 },
								{ musicDifficulty: "master", playLevel: 26 },
							],
						},
						{
							id: 2,
							title: "Brand New Day",
							composer: "いるかアイス",
							assetbundleName: "jacket_s_002",
							difficulties: [
								{ musicDifficulty: "expert", playLevel: 28 },
								{ musicDifficulty: "master", playLevel: 32 },
							],
						},
					]}
				/>
			);
		case "chart-detail":
			return (
				<ChartDetail
					music={{
						id: 1,
						title: "Tell Your World",
						pronunciation: "テル ユア ワールド",
						assetbundleName: "jacket_s_001",
						assetSource: "main-jp",
						jacketUrl: getMusicJacketUrl("jacket_s_001", "main-jp"),
						durationSec: 127,
						selectedDifficulty: "master",
						chartUrl: "https://charts-new.unipjsk.com/moe/svg/1/master.svg",
						difficulties: [
							{ musicDifficulty: "easy", playLevel: 5, totalNoteCount: 158 },
							{ musicDifficulty: "normal", playLevel: 10, totalNoteCount: 305 },
							{ musicDifficulty: "hard", playLevel: 16, totalNoteCount: 719 },
							{ musicDifficulty: "expert", playLevel: 22, totalNoteCount: 961 },
							{
								musicDifficulty: "master",
								playLevel: 26,
								totalNoteCount: 1147,
							},
						],
					}}
				/>
			);
		case "event-info":
			return (
				<EventInfo
					event={{
						id: 136,
						name: "闪耀于夜空的音色",
						eventType: "marathon",
						assetbundleName: "event_stella_2020",
						assetSource: "main-jp",
						startAt: Date.UTC(2026, 3, 20, 6, 0, 0),
						aggregateAt: Date.UTC(2026, 3, 29, 20, 59, 59),
						closedAt: Date.UTC(2026, 3, 30, 11, 0, 0),
						distributionEndAt: Date.UTC(2026, 4, 6, 11, 0, 0),
						unit: "light_sound",
						bonusAttr: "cool",
						bonusCharacters: ["一歌", "咲希", "穗波", "志步"],
						bonusCards: [
							{
								id: 1001,
								prefix: "被星光照亮的舞台",
								characterName: "初音未来",
								rarity: "rarity_4",
								cardRarityType: "rarity_4",
								attr: "cool",
								assetbundleName: "res001_no003",
								supplyType: "期间限定",
							},
						],
					}}
				/>
			);
		case "event-list":
			return (
				<EventList
					title="活动查询"
					subtitle="2026"
					page={1}
					totalPages={1}
					total={2}
					assetSource="main-jp"
					events={[
						{
							id: 136,
							name: "闪耀于夜空的音色",
							eventType: "marathon",
							unit: "light_sound",
							assetbundleName: "event_stella_2020",
							storyBannerUrl:
								"https://storage.exmeaning.com/sekai-jp-assets/event_story/event_show_2026/screen_image/banner_event_story.png",
							startAt: Date.UTC(2026, 3, 20, 6, 0, 0),
							closedAt: Date.UTC(2026, 3, 30, 11, 0, 0),
							bonusAttr: "cool",
							bonusCharacters: ["一歌", "咲希", "穗波", "志步"],
							bonusCards: [
								{ id: 1001, prefix: "被星光照亮的舞台", characterName: "初音未来" },
							],
						},
						{
							id: 137,
							name: "在世界中心唱响",
							eventType: "cheerful_carnival",
							unit: "piapro",
							assetbundleName: "event_stella_2020",
							storyBannerUrl:
								"https://storage.exmeaning.com/sekai-jp-assets/event_story/event_show_2026/screen_image/banner_event_story.png",
							startAt: Date.UTC(2026, 4, 1, 6, 0, 0),
							closedAt: Date.UTC(2026, 4, 9, 11, 0, 0),
							bonusAttr: "cute",
							bonusCharacters: ["初音未来", "镜音铃"],
						},
					]}
				/>
			);
		case "gacha-info":
			return (
				<GachaInfo
					gacha={{
						id: 700,
						name: "闪耀舞台招募",
						gachaType: "ceil",
						assetbundleName: "ab_gacha_900",
						assetSource: "main-jp",
						logoUrl: getGachaLogoUrl("ab_gacha_900", "main-jp"),
						startAt: Date.UTC(2026, 3, 20, 6, 0, 0),
						endAt: Date.UTC(2026, 3, 30, 11, 0, 0),
						wishSelectCount: 1,
						pickupCards: createPickupPreviewCards(),
					}}
				/>
			);
		case "gacha-list":
			return (
				<GachaList
					title="卡池查询"
					subtitle="当前/最近卡池"
					page={1}
					totalPages={1}
					total={2}
					assetSource="main-jp"
					gachas={[
						{
							id: 700,
							name: "闪耀舞台招募",
							gachaType: "ceil",
							assetbundleName: "ab_gacha_900",
							startAt: Date.UTC(2026, 3, 20, 6, 0, 0),
							endAt: Date.UTC(2026, 3, 30, 11, 0, 0),
							pickupCards: createPickupPreviewCards(),
						},
						{
							id: 701,
							name: "Birthday Gift 招募",
							gachaType: "birthday",
							assetbundleName: "ab_gacha_901",
							startAt: Date.UTC(2026, 4, 1, 6, 0, 0),
							endAt: Date.UTC(2026, 4, 8, 11, 0, 0),
							pickupCards: createPickupPreviewCards().slice(0, 1),
						},
					]}
				/>
			);
		case "virtual-live-list":
			return (
				<VirtualLiveList
					title="虚拟 Live"
					subtitle="未来 7 天内的近期虚拟 Live"
					page={1}
					totalPages={1}
					total={2}
					assetSource="main-jp"
					virtualLives={[
						{
							id: 1,
							name: "HAPPY ANNIVERSARY Virtual Live",
							assetbundleName: "vlentrance_00001_re",
							virtualLiveType: "normal",
							startAt: Date.now() + 3600000,
							endAt: Date.now() + 86400000,
							currentStartAt: Date.now() + 3600000,
							currentEndAt: Date.now() + 5400000,
							restCount: 4,
							characters: [
								{ characterName: "初音未来" },
								{ characterName: "星乃一歌" },
							],
							rewards: [{ resourceBoxId: 101 }],
						},
						{
							id: 2,
							name: "After Event Live",
							assetbundleName: "vlentrance_00002",
							virtualLiveType: "after_live",
							startAt: Date.now() + 7200000,
							endAt: Date.now() + 172800000,
							restCount: 6,
							characters: [
								{ characterName: "花里实乃理" },
								{ characterName: "小豆泽心羽" },
							],
							rewards: [{ resourceBoxId: 102 }],
						},
					]}
				/>
			);
		case "ranking-list":
			return (
				<RankingList
					title="闪耀于夜空的音色"
					eventId={136}
					updatedAt={Date.now()}
					rankings={[
						{
							rank: 1,
							displayName: "Miku Fan",
							signature: "目标 T1!",
							score: 112345678,
							scoreDelta: 123456,
							rankDelta: 0,
							leaderCard: previewLeaderCard("res001_no003", "初音未来", "cute"),
						},
						{
							rank: 2,
							displayName: "Sekai Runner",
							signature: "周回中",
							score: 98765432,
							scoreDelta: 45678,
							rankDelta: 1,
							leaderCard: previewLeaderCard("res002_no003", "镜音铃", "happy"),
						},
						{
							rank: 3,
							displayName: "Night Melody",
							score: 87654321,
							scoreDelta: -1200,
							rankDelta: -1,
							leaderCard: previewLeaderCard("res003_no003", "镜音连", "cool"),
						},
						{
							rank: 10,
							displayName: "Wonder Stage",
							score: 65432100,
							userId: "10010",
							leaderCharacterId: 9,
						},
						{
							rank: 100,
							displayName: "Virtual Singer",
							score: 43210000,
							userId: "10100",
							leaderCharacterId: 1,
						},
						{
							rank: 1000,
							displayName: "Moebot Tester",
							score: 21000000,
							scoreDelta: 8000,
							userId: "11000",
							leaderCharacterId: 21,
						},
					]}
				/>
			);
		case "profile-card":
			return (
				<ProfileCard
					profile={{
						name: "Moebot Tester",
						rank: 398,
						userId: "1234567890",
						bio: "今天也要在 SEKAI 里闪闪发光。",
						totalPower: 352198,
						assetSource: "main-jp",
						stats: {
							multiLiveCount: 8888,
							mvpCount: 1234,
							superStarCount: 567,
						},
						deckCards: createDeckPreviewCards(),
						characterRanks: [
							{ characterId: 1, characterName: "一歌", rank: 72 },
							{ characterId: 20, characterName: "瑞希", rank: 68 },
							{ characterId: 9, characterName: "遥", rank: 64 },
							{ characterId: 14, characterName: "杏", rank: 58 },
							{ characterId: 17, characterName: "司", rank: 55 },
							{ characterId: 3, characterName: "穗波", rank: 51 },
						],
						profileHonors: [
							{
								honorId: 1,
								name: "一歌ファン",
								level: 5,
								honorRarity: "low",
								assetbundleName: "honor_0001",
								imageUrl: getHonorBgUrl("honor_0001", false, "main-jp"),
								frameUrl: getHonorFrameUrl("low", false, "main-jp"),
								levelIconUrl: getHonorLevelIconUrl("main-jp"),
								levelIcon6Url: getHonorLevelIconUrl("main-jp").replace("icon_degreeLv.png", "icon_degreeLv6.png"),
							},
							{
								honorId: 79,
								name: "皆传",
								level: 12,
								honorRarity: "highest",
								assetbundleName: "honor_0034",
								imageUrl: getHonorBgUrl("honor_0034", false, "main-jp"),
								frameUrl: getHonorFrameUrl("highest", false, "main-jp"),
								levelIconUrl: getHonorLevelIconUrl("main-jp"),
								levelIcon6Url: getHonorLevelIconUrl("main-jp").replace("icon_degreeLv.png", "icon_degreeLv6.png"),
							},
							{
								honorId: 101201,
								honorType: "bonds",
								name: "羁绊称号",
								level: 4,
								honorRarity: "high",
								leftCharacterId: 1,
								rightCharacterId: 20,
								leftCharacterUrl: getBondsHonorCharacterUrl(1, "main-jp"),
								rightCharacterUrl: getBondsHonorCharacterUrl(20, "main-jp"),
								leftColor: "#33aaee",
								rightColor: "#ddaacc",
							},
						],
					}}
				/>
			);
		case "gacha-result":
			return (
				<GachaResult
					pullType="multi"
					assetSource="main-jp"
					results={createGachaPreviewResults()}
				/>
			);
		case "suite-panel":
			return (
				<SuitePanel
					title="Suite 状态"
					subtitle="自动同步 · 主线 JP"
					assetSource="main-jp"
					profile={{
						name: "Moebot Tester",
						rank: 398,
						userId: "1234567890",
						bio: "今天也要把缺的卡慢慢补齐。",
						source: "Sekai Viewer",
						updatedAt: "2026-05-05 12:30",
					}}
					stats={[
						{ label: "总综合力", value: 352198, highlight: true },
						{ label: "持有卡牌", value: "128/420" },
						{ label: "四星", value: 56 },
						{ label: "满破", value: 18 },
					]}
					deckCards={createDeckPreviewCards()}
					sections={[
						{
							title: "羁绊 TOP",
							kind: "bond_list",
							note: "角色头像来自本地 assets/characters。",
							rows: [
								{ rank: 1, label: "穗波 × 镜音连", value: "Lv.55", meta: "EXP 820", extra: { characterId1: 3, characterId2: 23, characterName1: "穗波", characterName2: "镜音连", rankLevel: 55, exp: 820 } },
								{ rank: 2, label: "咲希 × 镜音铃", value: "Lv.48", meta: "EXP 360", extra: { characterId1: 2, characterId2: 22, characterName1: "咲希", characterName2: "镜音铃", rankLevel: 48, exp: 360 } },
								{ rank: 3, label: "瑞希 × 绘名", value: "Lv.42", meta: "EXP 120", extra: { characterId1: 20, characterId2: 19, characterName1: "瑞希", characterName2: "绘名", rankLevel: 42, exp: 120 } },
							],
						},
						{
							title: "资源概览",
							items: [
								{ label: "水晶", value: 45320 },
								{ label: "想法碎片", value: 1288 },
								{ label: "技能书", value: 42 },
							],
						},
					]}
				/>
			);
		case "best30":
			return (
				<Best30
					title="JP Best30"
					subtitle="社区定数 · 仅供参考"
					regionLabel="日服"
					updateText="2026-05-06 12:30:00"
					assetSource="main-jp"
					profile={{ name: "Moebot Tester", rank: 398, userId: "1234567890", source: "Haruki Suite" }}
					average={32.84}
					candidateCount={74}
					apCount={28}
					fcCount={46}
					missingConstantsCount={2}
					totalResultCount={512}
					entries={createBest30PreviewEntries()}
					constantsSource="https://moe.exmeaning.com/data/pjskb30/merged_chart.csv"
				/>
			);
		case "suite-card-box":
			return (
				<SuiteCardBox
					title="Suite 卡牌盒"
					profile={{ name: "Moebot Tester", rank: 398, userId: "1234567890" }}
					assetSource="main-jp"
					total={12}
					ownedTotal={8}
					options={{ groupByCharacter: true, showId: true, showCreatedAt: true }}
					cards={createSuiteCardBoxPreviewCards()}
				/>
			);
		case "help-card":
			return (
				<HelpCard
					version="0.1.0"
					footer={
						"🌐 区服前缀: jp / cn / tw / kr / en；WL 加在区服与命令之间\n" +
						"💡 例: /cn查卡 1204、/krsk 1k、/cnwlcf、/wlcsb 1 100\n" +
						"💡 无前缀按你的绑定服务器，未绑定则默认日服"
					}
					groups={[
						{
							label: "查询 / 榜线",
							commands: [
								{ name: "查卡" },
								{ name: "查曲" },
								{ name: "查谱" },
								{ name: "查活动" },
								{ name: "查卡池" },
								{ name: "查演唱会" },
								{ name: "榜线" },
								{ name: "sk" },
								{ name: "查房" },
								{ name: "查水表" },
								{ name: "榜线预测" },
							],
						},
						{
							label: "账号 / Profile",
							commands: [
								{ name: "绑定" },
								{ name: "解绑" },
								{ name: "个人信息" },
							],
						},
						{
							label: "Suite 数据",
							commands: [
								{ name: "抓包状态" },
								{ name: "隐藏抓包" },
								{ name: "展示抓包" },
								{ name: "羁绊" },
								{ name: "打歌进度" },
								{ name: "best30" },
								{ name: "挑战信息" },
								{ name: "活动记录" },
								{ name: "队长次数" },
								{ name: "CR任务" },
								{ name: "ANVO持有" },
								{ name: "卡牌一览" },
							],
						},
						{
							label: "组卡推荐",
							commands: [
								{ name: "组卡" },
								{ name: "最强组卡" },
								{ name: "挑战组卡" },
								{ name: "加成组卡" },
								{ name: "烤森组卡" },
							],
						},
						{
							label: "其它",
							commands: [
								{ name: "抽卡模拟" },
								{ name: "帮助" },
							],
						},
					]}
				/>
			);
		default:
			throw new Error(`Unknown render preview template: ${id}`);
	}
}

function createPickupPreviewCards() {
	return [
		{
			id: 3001,
			characterName: "初音未来",
			rarity: "rarity_4",
			attr: "cute",
			assetbundleName: "res001_no003",
			isWish: true,
			weight: 400,
		},
		{
			id: 3002,
			characterName: "镜音铃",
			rarity: "rarity_4",
			attr: "happy",
			assetbundleName: "res002_no003",
			isWish: true,
			weight: 400,
		},
		{
			id: 3003,
			characterName: "镜音连",
			rarity: "rarity_4",
			attr: "cool",
			assetbundleName: "res003_no003",
			isWish: true,
			weight: 400,
		},
		{
			id: 3004,
			characterName: "巡音流歌",
			rarity: "rarity_3",
			attr: "pure",
			assetbundleName: "res004_no003",
			weight: 1200,
		},
	];
}

function createBest30PreviewEntries() {
	const titles = [
		"Hatsune Creation Myth",
		"What's up? Pop!",
		"the EmpErroR",
		"Yaminabe!!!!",
		"Don't Fight The Music",
		"嬢王",
		"六兆年と一夜物語",
		"Brand New Day",
		"初音ミクの消失",
		"ÅMARA(大未来電脳)",
	];
	const diffs = ["append", "master", "master", "master", "master"];
	return Array.from({ length: 30 }, (_, index) => {
		const diff = diffs[index % diffs.length];
		const constant = 35.2 - index * 0.12;
		const ap = index % 4 === 0;
		return {
			rank: index + 1,
			musicId: 100 + index,
			title: titles[index % titles.length],
			difficulty: diff,
			difficultyLabel: diff === "append" ? "APD" : "MAS",
			level: Math.max(29, Math.round(constant)),
			constant,
			userRating: ap ? constant : constant >= 33 ? constant - 1 : constant - 1.5,
			playResult: ap ? "AP" : "FC",
			noteCount: 980 + index * 23,
			assetbundleName: `jacket_s_${String((index % 20) + 1).padStart(3, "0")}`,
			jacketUrl: getMusicJacketUrl(`jacket_s_${String((index % 20) + 1).padStart(3, "0")}`, "main-jp"),
		};
	});
}

function createDeckPreviewCards() {
	return [
		{
			cardId: 3001,
			characterName: "初音未来",
			rarity: "rarity_4",
			attr: "cute",
			assetbundleName: "res001_no003",
			isTrained: true,
			mastery: 5,
			level: 60,
		},
		{
			cardId: 3002,
			characterName: "镜音铃",
			rarity: "rarity_4",
			attr: "happy",
			assetbundleName: "res002_no003",
			isTrained: true,
			mastery: 3,
			level: 60,
		},
		{
			cardId: 3003,
			characterName: "镜音连",
			rarity: "rarity_4",
			attr: "cool",
			assetbundleName: "res003_no003",
			isTrained: true,
			mastery: 2,
			level: 60,
		},
		{
			cardId: 3004,
			characterName: "巡音流歌",
			rarity: "rarity_3",
			attr: "pure",
			assetbundleName: "res004_no003",
			isTrained: true,
			mastery: 1,
			level: 50,
		},
		{
			cardId: 3005,
			characterName: "MEIKO",
			rarity: "rarity_3",
			attr: "mysterious",
			assetbundleName: "res005_no003",
			isTrained: true,
			level: 50,
		},
	];
}

function createSuiteCardBoxPreviewCards() {
	return createGachaPreviewResults().slice(0, 12).map((card, index) => ({
		...card,
		id: card.cardId,
		prefix: index % 2 === 0 ? "闪耀的舞台" : "与你相连的歌声",
		owned: index % 4 !== 1,
		level: index % 4 !== 1 ? (card.rarity === "rarity_4" ? 60 : 50) : undefined,
		mastery: index % 4 !== 1 ? index % 6 : undefined,
		skillLevel: index % 4 !== 1 ? 1 + (index % 4) : undefined,
		createdAt: `2026-05-${String(5 - (index % 5)).padStart(2, "0")}`,
		supplyType: index === 2 ? "CFES限定" : index === 6 ? "期间限定" : undefined,
		isBirthday: index === 8,
		characterName: ["初音未来", "初音未来", "镜音铃", "镜音铃", "镜音连", "巡音流歌", "MEIKO", "KAITO", "星乃一歌", "天马咲希", "望月穗波", "日野森志步"][index],
	}));
}

function createGachaPreviewResults() {
	const bundles = [
		"res001_no003",
		"res002_no003",
		"res003_no003",
		"res004_no003",
		"res005_no003",
	];
	const names = [
		"一歌",
		"咲希",
		"穗波",
		"志步",
		"实乃理",
		"遥",
		"爱莉",
		"雫",
		"心羽",
		"杏",
	];
	const attrs = ["cute", "cool", "pure", "happy", "mysterious"];

	return Array.from({ length: 10 }, (_, index) => {
		const assetbundleName = bundles[index % bundles.length];
		const rarity =
			index === 2 || index === 8
				? "rarity_4"
				: index % 3 === 0
					? "rarity_3"
					: "rarity_2";
		const isTrained = rarity === "rarity_3" || rarity === "rarity_4";

		return {
			cardId: 3000 + index,
			characterName: names[index],
			rarity,
			attr: attrs[index % attrs.length],
			isNew: index === 2 || index === 8,
			assetbundleName,
			thumbnailUrl: getCardThumbnailUrl(
				assetbundleName,
				false,
				"main-jp",
				"png",
			),
			trainedThumbnailUrl: getCardThumbnailUrl(
				assetbundleName,
				true,
				"main-jp",
				"png",
			),
			isTrained,
		};
	});
}

function previewLeaderCard(
	assetbundleName: string,
	characterName: string,
	attr: string,
) {
	return {
		cardId: Number(assetbundleName.match(/\d+/)?.[0] ?? 1),
		characterName,
		rarity: "rarity_4",
		attr,
		assetbundleName,
		thumbnailUrl: getCardThumbnailUrl(assetbundleName, false, "main-jp", "png"),
		trainedThumbnailUrl: getCardThumbnailUrl(
			assetbundleName,
			true,
			"main-jp",
			"png",
		),
		isTrained: true,
		mastery: 5,
	};
}

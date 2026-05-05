import { getCardThumbnailUrl, type AssetSourceType } from "../../shared";
import { getLocalCharacterIconAssetDataUri } from "../styles/assets";
import { theme } from "../styles/theme";
import { BaseCard } from "./base";
import { SekaiCardThumbnail } from "./SekaiCardThumbnail";

export interface SuitePanelProps {
	title?: string;
	subtitle?: string;
	profile?: SuiteProfile;
	stats?: SuiteStat[] | Record<string, unknown>;
	sections?: SuiteSection[];
	deckCards?: SuiteDeckCard[];
	assetSource?: AssetSourceType | string;
}

interface SuiteProfile {
	name?: string;
	displayName?: string;
	rank?: number | string;
	userId?: number | string;
	uid?: number | string;
	bio?: string;
	signature?: string;
	source?: string;
	updatedAt?: number | string;
	uploadTime?: number | string;
	updateText?: string;
	avatarUrl?: string;
}

interface SuiteStat {
	label?: string;
	name?: string;
	value?: unknown;
	description?: string;
	hint?: string;
	color?: string;
	highlight?: boolean;
}

interface SuiteSectionRow {
	id?: number | string;
	rank?: number | string;
	label?: string;
	value?: unknown;
	meta?: string;
	color?: string;
	card?: SuiteDeckCard;
	characterId?: number | string;
	musicId?: number | string;
	eventId?: number | string;
	iconUrl?: string;
	imageUrl?: string;
	bannerUrl?: string;
	logoUrl?: string;
	dateText?: string;
	startAt?: number | string;
	endAt?: number | string;
	progress?: number | string;
	progressMax?: number | string;
	progressLabel?: string;
	extra?: Record<string, unknown>;
	[key: string]: unknown;
}

interface SuiteSection {
	title?: string;
	subtitle?: string;
	kind?: string;
	note?: string;
	columns?: Array<string | { key?: string; label?: string }>;
	rows?: Array<SuiteSectionRow | unknown[]>;
	items?: Array<string | { label?: string; value?: unknown; description?: string }>;
	extra?: Record<string, unknown>;
}

interface SuiteDeckCard {
	id?: number | string;
	cardId?: number | string;
	characterName?: string;
	rarity?: string;
	cardRarityType?: string;
	attr?: string;
	assetbundleName?: string;
	thumbnailUrl?: string;
	trainedThumbnailUrl?: string;
	compositeLayers?: import("../card-thumbnail-composites").CardThumbnailCompositeLayer[];
	isTrained?: boolean;
	defaultImage?: string;
	mastery?: number;
	level?: number;
}

export function SuitePanel({
	title = "Suite 数据面板",
	subtitle,
	profile,
	stats,
	sections = [],
	deckCards = [],
	assetSource = "main-jp",
}: SuitePanelProps) {
	const normalizedStats = normalizeStats(stats);
	const updated = profile?.updateText ?? profile?.uploadTime ?? profile?.updatedAt;
	const meta = [
		profile?.source ? `来源：${profile.source}` : undefined,
		updated ? `更新：${formatValue(updated)}` : undefined,
	].filter(Boolean).join(" · ");

	return (
		<BaseCard title={title} subtitle={subtitle ?? meta} accentColor={theme.colors.accentLight}>
			<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.lg }}>
				<ProfileBlock profile={profile} meta={meta} leaderCard={deckCards[0]} assetSource={assetSource} />

				{deckCards.length > 0 && (
					<Panel title="当前队伍">
						<div style={{ display: "flex", gap: theme.spacing.sm, justifyContent: "space-between" }}>
							{deckCards.slice(0, 5).map((card, index) => (
								<DeckThumb key={`${card.cardId ?? card.id ?? index}`} card={card} source={assetSource} leader={index === 0} />
							))}
						</div>
					</Panel>
				)}

				{normalizedStats.length > 0 && (
					<div style={{ display: "flex", flexWrap: "wrap", gap: theme.spacing.md }}>
						{normalizedStats.map((stat, index) => (
							<StatBox key={`${stat.label ?? index}`} stat={stat} />
						))}
					</div>
				)}

				{sections.map((section, index) => (
					<SectionBlock key={`${section.title ?? index}`} section={section} />
				))}
			</div>
		</BaseCard>
	);
}

function ProfileBlock({ profile, meta, leaderCard, assetSource }: { profile?: SuiteProfile; meta?: string; leaderCard?: SuiteDeckCard; assetSource: AssetSourceType | string }) {
	const name = profile?.displayName ?? profile?.name ?? "未知玩家";
	const uid = profile?.userId ?? profile?.uid;
	const bio = profile?.bio ?? profile?.signature;
	const avatarUrl = getSuiteCardImageUrl(leaderCard, assetSource) ?? profile?.avatarUrl;
	return (
		<div
			style={{
				display: "flex",
				gap: theme.spacing.lg,
				alignItems: "stretch",
				backgroundColor: theme.colors.surface,
				border: `1px solid ${theme.colors.border}`,
				borderRadius: theme.borderRadius.xl,
				padding: theme.spacing.lg,
			}}
		>
			<div
				style={{
					display: "flex",
					width: 96,
					height: 96,
					borderRadius: theme.borderRadius.xl,
					backgroundColor: theme.colors.surfaceAccent,
					border: `1px solid ${theme.colors.borderStrong}`,
					overflow: "hidden",
					alignItems: "center",
					justifyContent: "center",
					flexShrink: 0,
				}}
			>
				{avatarUrl ? (
					<img src={avatarUrl} width={96} height={96} style={{ objectFit: "cover" }} />
				) : (
					<span style={{ display: "flex", fontSize: theme.fontSize.xxl, color: theme.colors.accent, fontWeight: 900 }}>{name.slice(0, 1)}</span>
				)}
			</div>
			<div style={{ display: "flex", flexDirection: "column", flex: 1, justifyContent: "center", gap: theme.spacing.sm }}>
				<div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: theme.spacing.md }}>
					<div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
						<span style={{ display: "flex", color: theme.colors.text, fontSize: theme.fontSize.xl, fontWeight: 900 }}>{name}</span>
						{uid !== undefined && <span style={{ display: "flex", color: theme.colors.textMuted, fontSize: theme.fontSize.sm, fontWeight: 800 }}>UID: {String(uid)}</span>}
					</div>
					{profile?.rank !== undefined && (
						<span style={{ display: "flex", padding: "7px 14px", borderRadius: theme.borderRadius.round, backgroundColor: theme.colors.accentSoft, color: theme.colors.accent, fontSize: theme.fontSize.md, fontWeight: 900 }}>Rank {String(profile.rank)}</span>
					)}
				</div>
				{bio && <span style={{ display: "flex", color: theme.colors.textSecondary, fontSize: theme.fontSize.sm, lineHeight: 1.5 }}>{bio}</span>}
				{meta && <span style={{ display: "flex", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 800 }}>{meta}</span>}
			</div>
		</div>
	);
}

function Panel({ title, children }: { title: string; children: any }) {
	return (
		<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.md, backgroundColor: theme.colors.surface, border: `1px solid ${theme.colors.border}`, borderRadius: theme.borderRadius.xl, padding: theme.spacing.md }}>
			<span style={{ display: "flex", color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>{title}</span>
			{children}
		</div>
	);
}

function getSuiteCardImageUrl(card: SuiteDeckCard | undefined, source: AssetSourceType | string): string | undefined {
	if (!card) return undefined;
	const trained = card.defaultImage === "special_training" || Boolean(card.isTrained);
	const normalUrl = card.thumbnailUrl ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, source, "png") : undefined);
	const trainedUrl = card.trainedThumbnailUrl ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, true, source, "png") : undefined);
	return trained ? trainedUrl ?? normalUrl : normalUrl;
}

function DeckThumb({ card, source, leader }: { card: SuiteDeckCard; source: AssetSourceType | string; leader?: boolean }) {
	const rarity = card.cardRarityType ?? card.rarity ?? "rarity_1";
	const trained = card.defaultImage === "special_training" || Boolean(card.isTrained);
	const normalUrl = card.thumbnailUrl ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, source, "png") : undefined);
	const trainedUrl = card.trainedThumbnailUrl ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, true, source, "png") : undefined);
	return (
		<div style={{ display: "flex", flexDirection: "column", alignItems: "center", width: 128, gap: 5 }}>
			<div style={{ display: "flex", position: "relative" }}>
				<SekaiCardThumbnail imageUrl={trained ? trainedUrl ?? normalUrl : normalUrl} compositeLayers={card.compositeLayers} rarity={rarity} attr={card.attr ?? "cute"} isTrained={trained} mastery={card.mastery} characterName={card.characterName} size={112} />
				{leader && <span style={{ display: "flex", position: "absolute", right: -6, top: -6, padding: "3px 7px", borderRadius: theme.borderRadius.round, backgroundColor: theme.colors.accent, color: "#fff", fontSize: 10, fontWeight: 900 }}>LEADER</span>}
			</div>
			<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 800 }}>{card.level ? `Lv.${card.level}` : `#${card.cardId ?? card.id ?? "-"}`}</span>
		</div>
	);
}

function StatBox({ stat }: { stat: SuiteStat }) {
	const accent = stat.color ?? (stat.highlight ? theme.colors.accent : theme.colors.text);
	const description = stat.description ?? stat.hint;
	return (
		<div style={{ display: "flex", flexDirection: "column", width: 164, gap: 5, backgroundColor: stat.highlight || stat.color ? theme.colors.surfaceAccent : theme.colors.surface, border: `1px solid ${stat.highlight || stat.color ? theme.colors.borderStrong : theme.colors.border}`, borderRadius: theme.borderRadius.lg, padding: theme.spacing.md }}>
			<span style={{ display: "flex", color: theme.colors.textSecondary, fontSize: theme.fontSize.xs, fontWeight: 800 }}>{stat.label ?? stat.name ?? "统计"}</span>
			<span style={{ display: "flex", color: accent, fontSize: theme.fontSize.lg, fontWeight: 900 }}>{formatValue(stat.value ?? "-")}</span>
			{description && <span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 11 }}>{description}</span>}
		</div>
	);
}

function SectionBlock({ section }: { section: SuiteSection }) {
	const rows = (section.rows ?? []).filter((row): row is SuiteSectionRow => !Array.isArray(row));
	switch (section.kind) {
		case "leader_count":
			return <CharacterProgressSection section={section} rows={rows} mode="leader" />;
		case "challenge_info":
			return <CharacterProgressSection section={section} rows={rows} mode="challenge" />;
		case "bond_list":
			return <BondSection section={section} rows={rows} />;
		case "event_record":
		case "event_record_wl":
			return <EventRecordSection section={section} rows={rows} />;
		case "music_progress_summary":
		case "music_progress_level":
			return <MusicProgressSection section={section} rows={rows} />;
		case "music_reward_summary":
		case "music_reward_combo":
		case "music_reward_achieved":
			return <MusicRewardSection section={section} rows={rows} />;
		default:
			return <GenericSectionBlock section={section} />;
	}
}

function GenericSectionBlock({ section }: { section: SuiteSection }) {
	const columns = section.columns ?? [];
	const rows = section.rows ?? [];
	const items = section.items ?? [];
	return (
		<Panel title={section.title ?? "详情"}>
			<SectionNote section={section} />
			{rows.length > 0 ? (
				<div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
					{columns.length > 0 && <Row values={columns.map((col) => typeof col === "string" ? col : col.label ?? col.key ?? "-")} header />}
					{rows.map((row, index) => <Row key={index} values={rowValues(row, columns)} row={Array.isArray(row) ? undefined : row} />)}
				</div>
			) : (
				<ItemsList items={items} />
			)}
		</Panel>
	);
}

function SectionNote({ section }: { section: SuiteSection }) {
	return (
		<>
			{section.subtitle && <span style={{ display: "flex", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 800 }}>{section.subtitle}</span>}
			{section.note && <span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 11, lineHeight: 1.45 }}>{section.note}</span>}
		</>
	);
}

function SummaryChips({ items }: { items: Array<{ label: string; value: unknown; color?: string }> }) {
	const visible = items.filter((item) => item.value !== undefined && item.value !== null && item.value !== "");
	if (visible.length === 0) return null;
	return (
		<div style={{ display: "flex", flexWrap: "wrap", gap: 7 }}>
			{visible.map((item, index) => (
				<div key={index} style={{ display: "flex", flexDirection: "column", minWidth: 118, gap: 3, padding: "7px 10px", borderRadius: theme.borderRadius.lg, backgroundColor: theme.colors.surfaceLight, border: `1px solid ${item.color ?? theme.colors.border}` }}>
					<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900 }}>{item.label}</span>
					<span style={{ display: "flex", color: item.color ?? theme.colors.accent, fontSize: theme.fontSize.sm, fontWeight: 900 }}>{formatValue(item.value)}</span>
				</div>
			))}
		</div>
	);
}

function ItemsList({ items }: { items: NonNullable<SuiteSection["items"]> }) {
	return (
		<div style={{ display: "flex", flexDirection: "column", gap: 7 }}>
			{items.map((item, index) => typeof item === "string" ? (
				<span key={index} style={{ display: "flex", color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>{item}</span>
			) : (
				<div key={index} style={{ display: "flex", justifyContent: "space-between", gap: theme.spacing.md }}>
					<span style={{ display: "flex", color: theme.colors.textSecondary, fontSize: theme.fontSize.sm, fontWeight: 800 }}>{item.label ?? item.description ?? "-"}</span>
					<span style={{ display: "flex", color: theme.colors.text, fontSize: theme.fontSize.sm, fontWeight: 900 }}>{formatValue(item.value ?? "-")}</span>
				</div>
			))}
		</div>
	);
}

function BondSection({ section, rows }: { section: SuiteSection; rows: SuiteSectionRow[] }) {
	return (
		<Panel title={section.title ?? "羁绊 TOP"}>
			<SectionNote section={section} />
			<div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
				{rows.map((row, index) => <BondRow key={`${row.rank ?? index}`} row={row} />)}
			</div>
		</Panel>
	);
}

function BondRow({ row }: { row: SuiteSectionRow }) {
	const cid1 = toNumber(row.extra?.characterId1);
	const cid2 = toNumber(row.extra?.characterId2);
	const name1 = String(row.extra?.characterName1 ?? "");
	const name2 = String(row.extra?.characterName2 ?? "");
	const rankLevel = row.extra?.rankLevel ?? String(row.value ?? "").replace(/^Lv\.?/i, "");
	const exp = row.extra?.exp;
	const image1 = cid1 ? getLocalCharacterIconAssetDataUri(cid1) : undefined;
	const image2 = cid2 ? getLocalCharacterIconAssetDataUri(cid2) : undefined;
	return (
		<div style={{ display: "flex", alignItems: "center", gap: theme.spacing.md, padding: "10px 12px", borderRadius: theme.borderRadius.xl, backgroundColor: "rgba(242, 247, 251, 0.74)", border: `1px solid ${theme.colors.border}` }}>
			<span style={{ display: "flex", width: 28, justifyContent: "center", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 900 }}>#{formatValue(row.rank)}</span>
			<div style={{ display: "flex", alignItems: "center", width: 136, height: 58, position: "relative", flexShrink: 0 }}>
				<div style={{ display: "flex", position: "absolute", left: 0, top: 0, zIndex: 2 }}><CharacterIcon src={image1} label={name1 || row.label} /></div>
				<div style={{ display: "flex", position: "absolute", left: 54, top: 0, zIndex: 1 }}><CharacterIcon src={image2} label={name2 || row.label} /></div>
				<span style={{ display: "flex", position: "absolute", left: 44, top: 17, zIndex: 3, width: 24, height: 24, borderRadius: theme.borderRadius.round, alignItems: "center", justifyContent: "center", backgroundColor: "#ffffff", border: `1px solid ${theme.colors.borderStrong}`, color: theme.colors.accent, fontSize: 13, fontWeight: 900 }}>×</span>
			</div>
			<div style={{ display: "flex", flexDirection: "column", flex: 1, gap: 5 }}>
				<span style={{ display: "flex", color: theme.colors.text, fontSize: theme.fontSize.sm, fontWeight: 900 }}>{name1 && name2 ? `${name1} × ${name2}` : row.label ?? "羁绊组合"}</span>
				<div style={{ display: "flex", flexWrap: "wrap", gap: 6 }}>
					<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900, padding: "3px 7px", borderRadius: theme.borderRadius.round, backgroundColor: "rgba(255,255,255,0.72)" }}>角色头像 · assets/characters</span>
					<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900, padding: "3px 7px", borderRadius: theme.borderRadius.round, backgroundColor: "rgba(255,255,255,0.72)" }}>EXP {formatValue(exp ?? row.meta?.replace(/^EXP\s*/i, ""))}</span>
				</div>
			</div>
			<div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", gap: 3, minWidth: 78 }}>
				<span style={{ display: "flex", color: theme.colors.accent, fontSize: theme.fontSize.lg, fontWeight: 900 }}>Lv.{formatValue(rankLevel)}</span>
				<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900 }}>羁绊等级</span>
			</div>
		</div>
	);
}

function CharacterProgressSection({ section, rows, mode }: { section: SuiteSection; rows: SuiteSectionRow[]; mode: "leader" | "challenge" }) {
	if (mode === "challenge") return <ChallengeInfoSection section={section} rows={rows} />;
	return (
		<Panel title={section.title ?? "角色队长次数"}>
			<SectionNote section={section} />
			<SummaryChips items={characterSummaryItems(section, mode)} />
			<div style={{ display: "flex", flexDirection: "column", gap: 7 }}>
				{rows.map((row, index) => <CharacterProgressRow key={`${row.characterId ?? index}`} row={row} mode={mode} />)}
			</div>
		</Panel>
	);
}

function characterSummaryItems(section: SuiteSection, mode: "leader" | "challenge") {
	const extra = section.extra ?? {};
	if (mode === "leader") {
		return [
			{ label: "剩余总次数", value: extra.totalRemain },
			{ label: "普通档位", value: `${formatValue(extra.totalMissionLevel)}/${formatValue(extra.totalMissionMax)}` },
			{ label: "剩余档位", value: extra.totalMissionRemain },
			{ label: "EX总次数", value: extra.totalEx, color: "#a863e8" },
		];
	}
	const distribution = Array.isArray(extra.rankDistribution)
		? extra.rankDistribution.map((item: any) => `${item.label ?? `Lv.${item.level ?? 0}`}×${item.count ?? 0}`).join(" / ")
		: undefined;
	return [
		{ label: "剩余水晶", value: extra.totalRemainJewel },
		{ label: "剩余碎片", value: extra.totalRemainFragment, color: "#ff8fb3" },
		{ label: "剩余奖励档", value: extra.totalRemainRewards },
		{ label: "等级数量", value: distribution },
	];
}

function ChallengeInfoSection({ section, rows }: { section: SuiteSection; rows: SuiteSectionRow[] }) {
	return (
		<Panel title={section.title ?? "每日挑战 Live"}>
			<SectionNote section={section} />
			<SummaryChips items={characterSummaryItems(section, "challenge")} />
			<ChartLegend items={[{ label: "水晶", color: theme.colors.accent }, { label: "碎片", color: "#ff8fb3" }, { label: "分数进度", color: "#a863e8" }]} />
			<div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
				<ChallengeHeader />
				{rows.map((row, index) => <ChallengeInfoRow key={`${row.characterId ?? index}`} row={row} />)}
			</div>
		</Panel>
	);
}

function ChallengeHeader() {
	return (
		<div style={{ display: "flex", alignItems: "center", gap: theme.spacing.sm, padding: "5px 9px", borderRadius: theme.borderRadius.md, backgroundColor: theme.colors.surfaceLight, color: theme.colors.textMuted, fontSize: 10, fontWeight: 900 }}>
			<span style={{ display: "flex", width: 104 }}>角色</span>
			<span style={{ display: "flex", width: 48, justifyContent: "center" }}>等级</span>
			<span style={{ display: "flex", width: 100, justifyContent: "flex-end" }}>最高分</span>
			<span style={{ display: "flex", flex: 1 }}>进度</span>
			<span style={{ display: "flex", width: 54, justifyContent: "center" }}>剩余档</span>
			<span style={{ display: "flex", width: 64, justifyContent: "center" }}>水晶</span>
			<span style={{ display: "flex", width: 64, justifyContent: "center" }}>碎片</span>
		</div>
	);
}

function ChallengeInfoRow({ row }: { row: SuiteSectionRow }) {
	const cid = toNumber(row.characterId ?? row.id ?? row.rank);
	const image = row.iconUrl || (cid ? getLocalCharacterIconAssetDataUri(cid) : undefined);
	const rate = progressRate(row.progress, row.progressMax);
	const rankLevel = toNumber(row.extra?.rankLevel) ?? 0;
	const rewardRemain = toNumber(row.extra?.rewardRemain) ?? 0;
	const jewel = toNumber(row.extra?.remainJewel ?? row.extra?.jewel) ?? 0;
	const shard = toNumber(row.extra?.remainFragment ?? row.extra?.shard) ?? 0;
	const score = toNumber(row.extra?.highScore ?? row.progress) ?? 0;
	const accent = challengeColor(rate);
	return (
		<div style={{ display: "flex", alignItems: "center", gap: theme.spacing.sm, padding: "7px 9px", borderRadius: theme.borderRadius.lg, backgroundColor: "rgba(242, 247, 251, 0.72)", border: `1px solid ${theme.colors.border}` }}>
			<div style={{ display: "flex", alignItems: "center", gap: 7, width: 104 }}>
				<CharacterIcon src={image} label={row.label} />
				<span style={{ display: "flex", color: theme.colors.text, fontSize: 11, fontWeight: 900 }}>{row.label ?? `角色 ${cid ?? "-"}`}</span>
			</div>
			<span style={{ display: "flex", width: 48, justifyContent: "center", color: rankLevel > 0 ? theme.colors.text : theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 900 }}>{rankLevel > 0 ? `Lv.${formatValue(rankLevel)}` : "-"}</span>
			<span style={{ display: "flex", width: 100, justifyContent: "flex-end", color: score > 0 ? theme.colors.text : theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 900 }}>{score > 0 ? formatValue(score) : "-"}</span>
			<div style={{ display: "flex", flex: 1, flexDirection: "column", gap: 3 }}>
				<ProgressBar value={rate} color={accent} />
				<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 9, fontWeight: 800 }}>{row.progressLabel ?? `${Math.round(rate * 100)}%`}</span>
			</div>
			<ResourcePill value={rewardRemain} color="#a863e8" width={54} />
			<ResourcePill value={jewel} color={theme.colors.accent} width={64} />
			<ResourcePill value={shard} color="#ff8fb3" width={64} />
		</div>
	);
}

function CharacterProgressRow({ row, mode }: { row: SuiteSectionRow; mode: "leader" | "challenge" }) {
	const cid = toNumber(row.characterId ?? row.id ?? row.rank);
	const image = row.iconUrl || (cid ? getLocalCharacterIconAssetDataUri(cid) : undefined);
	const rate = progressRate(row.progress, row.progressMax);
	const exLevel = row.extra?.exLevel ?? 0;
	const playLiveEx = row.extra?.playLiveEx ?? 0;
	const playLiveRemain = row.extra?.playLiveRemain ?? 0;
	const missionLevel = row.extra?.missionLevel ?? 0;
	const missionLevelMax = row.extra?.missionLevelMax ?? 0;
	const nextNeed = row.extra?.nextNeed ?? 0;
	const rankLevel = row.extra?.rankLevel ?? 0;
	const rewardCount = row.extra?.rewardCount ?? 0;
	const rewardRemain = row.extra?.rewardRemain ?? 0;
	const jewel = row.extra?.jewel ?? 0;
	const shard = row.extra?.shard ?? 0;
	const accent = mode === "leader" ? theme.colors.accent : challengeColor(rate);
	const detail = mode === "leader"
		? [`剩余 ${formatValue(playLiveRemain)}`, `档位 ${formatValue(missionLevel)}/${formatValue(missionLevelMax)}`, `下一档 ${formatValue(nextNeed)}`, `EX x${formatValue(exLevel)} / ${formatValue(playLiveEx)}`]
		: [`挑战 Lv.${formatValue(rankLevel)}`, `剩余奖励 ${formatValue(rewardRemain)}`, `水晶 ${formatValue(jewel)}`, `碎片 ${formatValue(shard)}`, `已领 ${formatValue(rewardCount)}`];
	return (
		<div style={{ display: "flex", alignItems: "center", gap: theme.spacing.sm, padding: "8px 10px", borderRadius: theme.borderRadius.lg, backgroundColor: "rgba(242, 247, 251, 0.7)", border: `1px solid ${theme.colors.border}` }}>
			<CharacterIcon src={image} label={row.label} />
			<div style={{ display: "flex", flexDirection: "column", flex: 1, gap: 5 }}>
				<div style={{ display: "flex", justifyContent: "space-between", gap: theme.spacing.md, alignItems: "baseline" }}>
					<span style={{ display: "flex", color: theme.colors.text, fontSize: theme.fontSize.sm, fontWeight: 900 }}>{row.label ?? `角色 ${cid ?? "-"}`}</span>
					<span style={{ display: "flex", color: accent, fontSize: theme.fontSize.md, fontWeight: 900 }}>{formatValue(row.value)}</span>
				</div>
				<ProgressBar value={rate} color={accent} />
				<div style={{ display: "flex", flexWrap: "wrap", gap: 5 }}>
					{detail.map((item, index) => <span key={index} style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900, padding: "2px 6px", borderRadius: theme.borderRadius.round, backgroundColor: "rgba(255,255,255,0.68)" }}>{item}</span>)}
				</div>
			</div>
		</div>
	);
}

function EventRecordSection({ section, rows }: { section: SuiteSection; rows: SuiteSectionRow[] }) {
	return (
		<Panel title={section.title ?? "活动记录"}>
			<SectionNote section={section} />
			<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.sm }}>
				{rows.map((row, index) => <EventRecordRow key={`${row.eventId ?? index}`} row={row} />)}
			</div>
		</Panel>
	);
}

function EventRecordRow({ row }: { row: SuiteSectionRow }) {
	const cid = toNumber(row.characterId);
	const chara = cid ? getLocalCharacterIconAssetDataUri(cid) : undefined;
	return (
		<div style={{ display: "flex", overflow: "hidden", borderRadius: theme.borderRadius.lg, backgroundColor: "rgba(242, 247, 251, 0.78)", border: `1px solid ${theme.colors.border}` }}>
			<div style={{ display: "flex", position: "relative", width: 176, height: 78, flexShrink: 0, backgroundColor: theme.colors.surfaceAccent, overflow: "hidden", alignItems: "center", justifyContent: "center" }}>
				{row.bannerUrl ? <img src={row.bannerUrl} width={176} height={78} style={{ objectFit: "cover" }} /> : <span style={{ display: "flex", color: theme.colors.accent, fontWeight: 900 }}>EVENT</span>}
				{chara && <div style={{ display: "flex", position: "absolute", right: 5, bottom: 5, width: 34, height: 34, borderRadius: theme.borderRadius.round, backgroundColor: "rgba(255,255,255,0.86)", overflow: "hidden", border: "2px solid white" }}><img src={chara} width={34} height={34} style={{ objectFit: "cover" }} /></div>}
			</div>
			<div style={{ display: "flex", flexDirection: "column", flex: 1, gap: 5, padding: "9px 12px", justifyContent: "center" }}>
				<div style={{ display: "flex", justifyContent: "space-between", gap: theme.spacing.md }}>
					<span style={{ display: "flex", flex: 1, color: theme.colors.text, fontSize: theme.fontSize.sm, fontWeight: 900, lineHeight: 1.25 }}>{row.label ?? `活动 ${row.eventId ?? "-"}`}</span>
					<span style={{ display: "flex", color: theme.colors.accent, fontSize: theme.fontSize.md, fontWeight: 900 }}>{formatValue(row.value)}</span>
				</div>
				<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 11, fontWeight: 800 }}>{[row.meta, row.dateText].filter(Boolean).join(" · ") || "-"}</span>
			</div>
		</div>
	);
}

function MusicProgressSection({ section, rows }: { section: SuiteSection; rows: SuiteSectionRow[] }) {
	if (section.kind === "music_progress_level") return <MusicProgressLevelChart section={section} rows={rows} />;
	return (
		<Panel title={section.title ?? "打歌进度"}>
			<SectionNote section={section} />
			<SummaryChips items={[{ label: "游玩谱面", value: section.extra?.totalPlayed }, { label: "Clear", value: section.extra?.totalClear }, { label: "FC", value: section.extra?.totalFC }, { label: "AP", value: section.extra?.totalAP }]} />
			<ChartLegend items={playResultLegend()} />
			<div style={{ display: "flex", flexDirection: "column", gap: 7 }}>
				{rows.map((row, index) => <MusicProgressChartRow key={index} row={row} />)}
			</div>
		</Panel>
	);
}

function MusicProgressChartRow({ row }: { row: SuiteSectionRow }) {
	const diff = String(row.extra?.diff ?? row.label ?? "-").toUpperCase();
	const played = toNumber(row.extra?.played) ?? 0;
	const clear = toNumber(row.extra?.clear) ?? 0;
	const fc = toNumber(row.extra?.fc) ?? 0;
	const ap = toNumber(row.extra?.ap) ?? 0;
	const total = Math.max(played, clear, fc, ap, 1);
	return (
		<div style={{ display: "flex", alignItems: "center", gap: theme.spacing.sm, padding: "8px 10px", borderRadius: theme.borderRadius.lg, backgroundColor: "rgba(242, 247, 251, 0.72)", border: `1px solid ${row.color ?? theme.colors.border}` }}>
			<DiffBadge diff={diff} color={row.color} width={86} />
			<StackedResultBar total={total} clear={clear} fc={fc} ap={ap} />
			<div style={{ display: "flex", gap: 5, width: 260, justifyContent: "flex-end" }}>
				<MetricMini label="游玩" value={played} />
				<MetricMini label="Clear" value={clear} color={PLAY_RESULT_COLORS.clear} />
				<MetricMini label="FC" value={fc} color={PLAY_RESULT_COLORS.fc} />
				<MetricMini label="AP" value={ap} color={PLAY_RESULT_COLORS.ap} />
			</div>
		</div>
	);
}

function MusicProgressLevelChart({ section, rows }: { section: SuiteSection; rows: SuiteSectionRow[] }) {
	const groups = groupRowsByDiff(rows);
	return (
		<Panel title={section.title ?? "等级数量"}>
			<SectionNote section={section} />
			<ChartLegend items={playResultLegend()} />
			<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.sm }}>
				{groups.map((group) => <MusicLevelGroup key={group.diff} diff={group.diff} rows={group.rows} />)}
			</div>
		</Panel>
	);
}

function MusicLevelGroup({ diff, rows }: { diff: string; rows: SuiteSectionRow[] }) {
	const color = rows[0]?.color ?? theme.colors.accent;
	return (
		<div style={{ display: "flex", flexDirection: "column", gap: 6, padding: theme.spacing.sm, borderRadius: theme.borderRadius.lg, backgroundColor: "rgba(242, 247, 251, 0.72)", border: `1px solid ${theme.colors.border}` }}>
			<div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
				<DiffBadge diff={diff.toUpperCase()} color={color} width={112} />
				<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900 }}>{rows.length} 个等级</span>
			</div>
			<div style={{ display: "flex", flexDirection: "column", gap: 5 }}>
				{rows.map((row, index) => <MusicLevelChartRow key={index} row={row} />)}
			</div>
		</div>
	);
}

function MusicLevelChartRow({ row }: { row: SuiteSectionRow }) {
	const level = row.extra?.level ?? "-";
	const total = Math.max(toNumber(row.extra?.total) ?? toNumber(row.extra?.played) ?? 0, 1);
	const played = toNumber(row.extra?.played) ?? 0;
	const clear = toNumber(row.extra?.clear) ?? 0;
	const fc = toNumber(row.extra?.fc) ?? 0;
	const ap = toNumber(row.extra?.ap) ?? 0;
	return (
		<div style={{ display: "flex", alignItems: "center", gap: theme.spacing.sm }}>
			<span style={{ display: "flex", width: 52, justifyContent: "center", color: row.color ?? theme.colors.accent, fontSize: 11, fontWeight: 900, padding: "3px 7px", borderRadius: theme.borderRadius.round, backgroundColor: "rgba(255,255,255,0.84)", border: `1px solid ${row.color ?? theme.colors.border}` }}>Lv.{formatValue(level)}</span>
			<StackedResultBar total={total} clear={clear} fc={fc} ap={ap} />
			<span style={{ display: "flex", width: 172, justifyContent: "flex-end", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900 }}>{formatValue(played)}/{formatValue(total)} 谱面 · C {formatValue(clear)} / FC {formatValue(fc)} / AP {formatValue(ap)}</span>
		</div>
	);
}

function MusicRewardSection({ section, rows }: { section: SuiteSection; rows: SuiteSectionRow[] }) {
	if (section.kind === "music_reward_combo") return <MusicRewardComboSection section={section} rows={rows} />;
	if (section.kind === "music_reward_summary") return <MusicRewardSummarySection section={section} />;
	return (
		<Panel title={section.title ?? "已达成奖励 TOP"}>
			<SectionNote section={section} />
			<div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
				{rows.map((row, index) => <AchievedRewardRow key={index} row={row} />)}
			</div>
		</Panel>
	);
}

function MusicRewardSummarySection({ section }: { section: SuiteSection }) {
	const extra = section.extra ?? {};
	const rankRemain = toNumber(extra.rankRemainCount) ?? 0;
	const validMusic = Math.max(toNumber(extra.validMusicCount) ?? 0, rankRemain, 1);
	const achieved = Math.max(validMusic - rankRemain, 0);
	return (
		<Panel title={section.title ?? "歌曲评级奖励(S)"}>
			<SectionNote section={section} />
			<SummaryChips items={[
				{ label: "总剩余水晶", value: extra.totalJewelRemain ?? extra.rankJewelRemain },
				{ label: "总剩余碎片", value: extra.totalShardRemain, color: "#ff8fb3" },
				{ label: "S剩余水晶", value: extra.rankJewelRemain },
				{ label: "已达成奖励", value: extra.achievementTotal },
			]} />
			<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.sm, padding: theme.spacing.sm, borderRadius: theme.borderRadius.lg, backgroundColor: "rgba(242, 247, 251, 0.72)", border: `1px solid ${theme.colors.border}` }}>
				<div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
					<span style={{ display: "flex", color: theme.colors.text, fontSize: theme.fontSize.sm, fontWeight: 900 }}>S 评级完成概览</span>
					<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 11, fontWeight: 900 }}>{formatValue(achieved)} / {formatValue(validMusic)} 首已达成</span>
				</div>
				<RatioBar segments={[{ value: achieved, color: theme.colors.accent }, { value: rankRemain, color: "rgba(148, 163, 184, 0.35)" }]} total={validMusic} height={14} />
				<div style={{ display: "flex", gap: theme.spacing.sm }}>
					<ResourceSummaryCard label="未 S 歌曲" value={rankRemain} suffix="首" color={theme.colors.warning} />
					<ResourceSummaryCard label="有效歌曲" value={validMusic} suffix="首" color={theme.colors.accent} />
					<ResourceSummaryCard label="涉及歌曲" value={extra.achievedMusicCount} suffix="首" color="#a863e8" />
				</div>
			</div>
		</Panel>
	);
}

function MusicRewardComboSection({ section, rows }: { section: SuiteSection; rows: SuiteSectionRow[] }) {
	const groups = groupRowsByDiff(rows);
	return (
		<Panel title={section.title ?? "连击奖励剩余"}>
			<SectionNote section={section} />
			<SummaryChips items={[
				{ label: "剩余总量", value: section.extra?.total },
				{ label: "连击水晶", value: section.extra?.comboJewelRemain },
				{ label: "连击碎片", value: section.extra?.comboShardRemain, color: "#ff8fb3" },
			]} />
			<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.sm }}>
				{groups.map((group) => <ComboRewardGroup key={group.diff} diff={group.diff} rows={group.rows} />)}
			</div>
		</Panel>
	);
}

function ComboRewardGroup({ diff, rows }: { diff: string; rows: SuiteSectionRow[] }) {
	const color = rows[0]?.color ?? theme.colors.accent;
	const maxAmount = Math.max(...rows.map((row) => toNumber(row.extra?.amount ?? row.value) ?? 0), 1);
	const total = rows.reduce((sum, row) => sum + (toNumber(row.extra?.amount ?? row.value) ?? 0), 0);
	const type = String(rows[0]?.extra?.rewardType ?? "jewel");
	return (
		<div style={{ display: "flex", flexDirection: "column", gap: 6, padding: theme.spacing.sm, borderRadius: theme.borderRadius.lg, backgroundColor: "rgba(242, 247, 251, 0.72)", border: `1px solid ${theme.colors.border}` }}>
			<div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
				<DiffBadge diff={diff.toUpperCase()} color={color} width={112} />
				<div style={{ display: "flex", gap: 6, alignItems: "center" }}>
					<RewardBadge type={type} />
					<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900 }}>合计 {formatValue(total)}</span>
				</div>
			</div>
			<div style={{ display: "flex", flexDirection: "column", gap: 5 }}>
				{rows.map((row, index) => <ComboRewardChartRow key={index} row={row} maxAmount={maxAmount} />)}
			</div>
		</div>
	);
}

function ComboRewardChartRow({ row, maxAmount }: { row: SuiteSectionRow; maxAmount: number }) {
	const level = row.extra?.level ?? "-";
	const count = row.extra?.count ?? "-";
	const accumulate = row.extra?.accumulate ?? "-";
	const amount = toNumber(row.extra?.amount ?? row.value) ?? 0;
	const type = String(row.extra?.rewardType ?? "jewel");
	const color = type === "shard" ? "#ff8fb3" : theme.colors.accent;
	return (
		<div style={{ display: "flex", alignItems: "center", gap: theme.spacing.sm }}>
			<span style={{ display: "flex", width: 52, justifyContent: "center", color: row.color ?? theme.colors.accent, fontSize: 11, fontWeight: 900, padding: "3px 7px", borderRadius: theme.borderRadius.round, backgroundColor: "rgba(255,255,255,0.84)", border: `1px solid ${row.color ?? theme.colors.border}` }}>Lv.{formatValue(level)}</span>
			<div style={{ display: "flex", flex: 1, alignItems: "center", gap: 8 }}>
				<SimpleBar value={amount} max={maxAmount} color={color} />
				<span style={{ display: "flex", width: 78, justifyContent: "flex-end", color, fontSize: theme.fontSize.xs, fontWeight: 900 }}>{formatValue(amount)}</span>
			</div>
			<span style={{ display: "flex", width: 148, justifyContent: "flex-end", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900 }}>累计 {formatValue(accumulate)} · {formatValue(count)}谱面</span>
		</div>
	);
}

function AchievedRewardRow({ row }: { row: SuiteSectionRow }) {
	return (
		<div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: theme.spacing.md, padding: "7px 10px", borderRadius: theme.borderRadius.lg, backgroundColor: "rgba(242, 247, 251, 0.72)", border: `1px solid ${theme.colors.border}` }}>
			<span style={{ display: "flex", flex: 1, color: theme.colors.textSecondary, fontSize: theme.fontSize.xs, fontWeight: 900 }}>{row.rank ? `#${formatValue(row.rank)} ` : ""}{row.label ?? "奖励"}</span>
			<span style={{ display: "flex", color: theme.colors.accent, fontSize: theme.fontSize.sm, fontWeight: 900 }}>{formatValue(row.value)}</span>
		</div>
	);
}

function groupRowsByDiff(rows: SuiteSectionRow[]): Array<{ diff: string; rows: SuiteSectionRow[] }> {
	const grouped = new Map<string, SuiteSectionRow[]>();
	for (const row of rows) {
		const diff = String(row.extra?.diff ?? row.label ?? "-");
		grouped.set(diff, [...(grouped.get(diff) ?? []), row]);
	}
	return Array.from(grouped.entries()).map(([diff, groupRows]) => ({ diff, rows: groupRows }));
}

const PLAY_RESULT_COLORS = {
	notPlayed: "rgba(148, 163, 184, 0.34)",
	clear: "#33ccbb",
	fc: "#ffb000",
	ap: "#ff6699",
};

function playResultLegend() {
	return [
		{ label: "未游玩/未完成", color: PLAY_RESULT_COLORS.notPlayed },
		{ label: "Clear", color: PLAY_RESULT_COLORS.clear },
		{ label: "FC", color: PLAY_RESULT_COLORS.fc },
		{ label: "AP", color: PLAY_RESULT_COLORS.ap },
	];
}

function ChartLegend({ items }: { items: Array<{ label: string; color: string }> }) {
	return (
		<div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
			{items.map((item, index) => (
				<div key={index} style={{ display: "flex", alignItems: "center", gap: 4 }}>
					<span style={{ display: "flex", width: 10, height: 10, borderRadius: theme.borderRadius.round, backgroundColor: item.color }} />
					<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900 }}>{item.label}</span>
				</div>
			))}
		</div>
	);
}

function DiffBadge({ diff, color, width = 78 }: { diff: string; color?: string; width?: number }) {
	return <span style={{ display: "flex", width, justifyContent: "center", color: "#fff", fontSize: 11, fontWeight: 900, padding: "5px 9px", borderRadius: theme.borderRadius.round, backgroundColor: color ?? theme.colors.accent }}>{diff}</span>;
}

function MetricMini({ label, value, color }: { label: string; value: unknown; color?: string }) {
	return (
		<div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", minWidth: 46, gap: 1 }}>
			<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 9, fontWeight: 900 }}>{label}</span>
			<span style={{ display: "flex", color: color ?? theme.colors.text, fontSize: 11, fontWeight: 900 }}>{formatValue(value)}</span>
		</div>
	);
}

function StackedResultBar({ total, clear, fc, ap }: { total: number; clear: number; fc: number; ap: number }) {
	const safeTotal = Math.max(total, clear, fc, ap, 1);
	const apOnly = Math.max(ap, 0);
	const fcOnly = Math.max(fc - ap, 0);
	const clearOnly = Math.max(clear - fc, 0);
	const notClear = Math.max(safeTotal - clear, 0);
	return <RatioBar segments={[{ value: notClear, color: PLAY_RESULT_COLORS.notPlayed }, { value: clearOnly, color: PLAY_RESULT_COLORS.clear }, { value: fcOnly, color: PLAY_RESULT_COLORS.fc }, { value: apOnly, color: PLAY_RESULT_COLORS.ap }]} total={safeTotal} height={13} />;
}

function RatioBar({ segments, total, height = 10 }: { segments: Array<{ value: number; color: string }>; total?: number; height?: number }) {
	const sum = total && total > 0 ? total : segments.reduce((acc, item) => acc + Math.max(item.value, 0), 0);
	return (
		<div style={{ display: "flex", flex: 1, height, borderRadius: theme.borderRadius.round, backgroundColor: "rgba(148, 163, 184, 0.18)", overflow: "hidden" }}>
			{segments.map((segment, index) => {
				const width = sum > 0 ? `${Math.max(0, segment.value) / sum * 100}%` : "0%";
				return <div key={index} style={{ display: "flex", width, backgroundColor: segment.color }} />;
			})}
		</div>
	);
}

function SimpleBar({ value, max, color }: { value: number; max: number; color: string }) {
	const width = max > 0 ? Math.max(4, Math.min(100, value / max * 100)) : 0;
	return (
		<div style={{ display: "flex", flex: 1, height: 12, borderRadius: theme.borderRadius.round, backgroundColor: "rgba(148, 163, 184, 0.18)", overflow: "hidden" }}>
			<div style={{ display: "flex", width: `${width}%`, backgroundColor: color, borderRadius: theme.borderRadius.round }} />
		</div>
	);
}

function ResourcePill({ value, color, width }: { value: unknown; color: string; width: number }) {
	return <span style={{ display: "flex", width, justifyContent: "center", color, fontSize: 11, fontWeight: 900, padding: "4px 6px", borderRadius: theme.borderRadius.round, backgroundColor: "rgba(255,255,255,0.82)", border: `1px solid ${color}` }}>{formatValue(value)}</span>;
}

function ResourceSummaryCard({ label, value, suffix, color }: { label: string; value: unknown; suffix?: string; color: string }) {
	return (
		<div style={{ display: "flex", flexDirection: "column", flex: 1, gap: 3, padding: "8px 10px", borderRadius: theme.borderRadius.lg, backgroundColor: "rgba(255,255,255,0.76)", border: `1px solid ${color}` }}>
			<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900 }}>{label}</span>
			<span style={{ display: "flex", color, fontSize: theme.fontSize.sm, fontWeight: 900 }}>{formatValue(value)}{suffix ?? ""}</span>
		</div>
	);
}

function Row({ values, header, row }: { values: unknown[]; header?: boolean; row?: SuiteSectionRow }) {
	return (
		<div style={{ display: "flex", gap: theme.spacing.sm, padding: "7px 10px", borderRadius: theme.borderRadius.md, backgroundColor: header ? theme.colors.surfaceLight : "rgba(242, 247, 251, 0.55)" }}>
			{values.map((value, index) => <span key={index} style={{ display: "flex", flex: 1, color: row?.color ?? (header ? theme.colors.text : theme.colors.textSecondary), fontSize: theme.fontSize.xs, fontWeight: header ? 900 : 800 }}>{formatValue(value)}</span>)}
		</div>
	);
}

function CharacterIcon({ src, label }: { src?: string; label?: string }) {
	return (
		<div style={{ display: "flex", width: 48, height: 48, borderRadius: theme.borderRadius.round, backgroundColor: theme.colors.surfaceAccent, border: `1px solid ${theme.colors.borderStrong}`, overflow: "hidden", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
			{src ? <img src={src} width={48} height={48} style={{ objectFit: "cover" }} /> : <span style={{ display: "flex", color: theme.colors.accent, fontWeight: 900 }}>{(label ?? "?").slice(0, 1)}</span>}
		</div>
	);
}

function ProgressBar({ value, color }: { value: number; color: string }) {
	return (
		<div style={{ display: "flex", height: 8, borderRadius: theme.borderRadius.round, backgroundColor: "rgba(148, 163, 184, 0.20)", overflow: "hidden" }}>
			<div style={{ display: "flex", width: `${Math.round(value * 100)}%`, backgroundColor: color, borderRadius: theme.borderRadius.round }} />
		</div>
	);
}

function RewardBadge({ type }: { type: string }) {
	const shard = type === "shard";
	return <span style={{ display: "flex", padding: "3px 8px", borderRadius: theme.borderRadius.round, backgroundColor: shard ? "rgba(255, 143, 179, 0.18)" : theme.colors.accentSoft, color: shard ? "#e84f82" : theme.colors.accent, fontSize: 10, fontWeight: 900 }}>{shard ? "碎片" : "水晶"}</span>;
}

function challengeColor(rate: number): string {
	if (rate >= 0.9) return "#ff6699";
	if (rate >= 0.65) return "#a863e8";
	if (rate >= 0.4) return "#33ccbb";
	return "#ffb000";
}

function progressRate(value: unknown, maxValue: unknown): number {
	const max = toNumber(maxValue) ?? 0;
	const current = toNumber(value) ?? 0;
	if (max <= 0 || current <= 0) return 0;
	return Math.max(0, Math.min(1, current / max));
}

function toNumber(value: unknown): number | undefined {
	if (typeof value === "number" && Number.isFinite(value)) return value;
	if (typeof value === "string" && value.trim() !== "") {
		const parsed = Number(value);
		if (Number.isFinite(parsed)) return parsed;
	}
	return undefined;
}

function rowValues(row: SuiteSectionRow | unknown[], columns: SuiteSection["columns"] = []): unknown[] {
	if (Array.isArray(row)) return row;
	if (columns.length > 0) return columns.map((col) => row[typeof col === "string" ? col : col.key ?? col.label ?? ""]);
	return [row.rank, row.label, row.value, row.meta, formatCardLabel(row.card)].filter((value) => value !== undefined && value !== "");
}

function formatCardLabel(card?: SuiteDeckCard): string | undefined {
	if (!card) return undefined;
	return [card.characterName, card.cardId ?? card.id ? `#${card.cardId ?? card.id}` : undefined, card.level ? `Lv.${card.level}` : undefined].filter(Boolean).join(" · ");
}

function normalizeStats(stats?: SuitePanelProps["stats"]): SuiteStat[] {
	if (!stats) return [];
	if (Array.isArray(stats)) return stats;
	return Object.entries(stats).map(([label, value], index) => ({ label, value, highlight: index === 0 }));
}

function formatValue(value: unknown): string {
	if (value === undefined || value === null || value === "") return "-";
	if (typeof value === "number") return Number.isFinite(value) ? value.toLocaleString() : "-";
	return String(value);
}

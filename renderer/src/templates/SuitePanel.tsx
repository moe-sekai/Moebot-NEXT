import { getCardThumbnailUrl, type AssetSourceType } from "../../shared";
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
	avatarUrl?: string;
}

interface SuiteStat {
	label?: string;
	name?: string;
	value?: unknown;
	description?: string;
	highlight?: boolean;
}

interface SuiteSection {
	title?: string;
	subtitle?: string;
	columns?: Array<string | { key?: string; label?: string }>;
	rows?: Array<Record<string, unknown> | unknown[]>;
	items?: Array<string | { label?: string; value?: unknown; description?: string }>;
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
	const meta = [
		profile?.source ? `来源：${profile.source}` : undefined,
		profile?.updatedAt ? `更新：${formatValue(profile.updatedAt)}` : undefined,
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
				<SekaiCardThumbnail imageUrl={trained ? trainedUrl ?? normalUrl : normalUrl} rarity={rarity} attr={card.attr ?? "cute"} isTrained={trained} mastery={card.mastery} characterName={card.characterName} size={112} />
				{leader && <span style={{ display: "flex", position: "absolute", right: -6, top: -6, padding: "3px 7px", borderRadius: theme.borderRadius.round, backgroundColor: theme.colors.accent, color: "#fff", fontSize: 10, fontWeight: 900 }}>LEADER</span>}
			</div>
			<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 800 }}>{card.level ? `Lv.${card.level}` : `#${card.cardId ?? card.id ?? "-"}`}</span>
		</div>
	);
}

function StatBox({ stat }: { stat: SuiteStat }) {
	return (
		<div style={{ display: "flex", flexDirection: "column", width: 164, gap: 5, backgroundColor: stat.highlight ? theme.colors.surfaceAccent : theme.colors.surface, border: `1px solid ${stat.highlight ? theme.colors.borderStrong : theme.colors.border}`, borderRadius: theme.borderRadius.lg, padding: theme.spacing.md }}>
			<span style={{ display: "flex", color: theme.colors.textSecondary, fontSize: theme.fontSize.xs, fontWeight: 800 }}>{stat.label ?? stat.name ?? "统计"}</span>
			<span style={{ display: "flex", color: stat.highlight ? theme.colors.accent : theme.colors.text, fontSize: theme.fontSize.lg, fontWeight: 900 }}>{formatValue(stat.value ?? "-")}</span>
			{stat.description && <span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 11 }}>{stat.description}</span>}
		</div>
	);
}

function SectionBlock({ section }: { section: SuiteSection }) {
	const columns = section.columns ?? [];
	const rows = section.rows ?? [];
	const items = section.items ?? [];
	return (
		<Panel title={section.title ?? "详情"}>
			{section.subtitle && <span style={{ display: "flex", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 800 }}>{section.subtitle}</span>}
			{rows.length > 0 ? (
				<div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
					{columns.length > 0 && <Row values={columns.map((col) => typeof col === "string" ? col : col.label ?? col.key ?? "-")} header />}
					{rows.map((row, index) => <Row key={index} values={rowValues(row, columns)} />)}
				</div>
			) : (
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
			)}
		</Panel>
	);
}

function Row({ values, header }: { values: unknown[]; header?: boolean }) {
	return (
		<div style={{ display: "flex", gap: theme.spacing.sm, padding: "7px 10px", borderRadius: theme.borderRadius.md, backgroundColor: header ? theme.colors.surfaceLight : "rgba(242, 247, 251, 0.55)" }}>
			{values.map((value, index) => <span key={index} style={{ display: "flex", flex: 1, color: header ? theme.colors.text : theme.colors.textSecondary, fontSize: theme.fontSize.xs, fontWeight: header ? 900 : 800 }}>{formatValue(value)}</span>)}
		</div>
	);
}

function rowValues(row: Record<string, unknown> | unknown[], columns: SuiteSection["columns"] = []): unknown[] {
	if (Array.isArray(row)) return row;
	if (columns.length === 0) return Object.values(row);
	return columns.map((col) => row[typeof col === "string" ? col : col.key ?? col.label ?? ""]);
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

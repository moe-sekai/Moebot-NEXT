import { theme } from "../styles/theme";
import { BaseCard } from "./base";
import { SekaiCardThumbnail } from "./SekaiCardThumbnail";

export interface DeckRecommendProps {
	title?: string;
	subtitle?: string;
	regionLabel?: string;
	profile?: any;
	event?: any;
	music?: any;
	options?: any;
	algorithm?: string;
	costMs?: number;
	warnings?: string[];
	decks?: DeckRecommendDeck[];
	assetSource?: string;
}

interface DeckRecommendDeck {
	rank?: number;
	value?: number;
	valueLabel?: string;
	score?: number;
	eventPoint?: number;
	eventBonus?: number;
	multiLiveScoreUp?: number;
	power?: { total?: number } | Record<string, unknown>;
	cards?: Array<any>;
}

export function DeckRecommend({
	title = "活动组卡推荐",
	subtitle,
	regionLabel,
	profile,
	event,
	music,
	options,
	algorithm,
	costMs,
	warnings = [],
	decks = [],
	assetSource = "main-jp",
}: DeckRecommendProps) {
	const summary = subtitle ?? [
		regionLabel,
		profile?.name ? `${profile.name}${profile.rank ? ` · Rank ${profile.rank}` : ""}` : undefined,
		event?.name,
	].filter(Boolean).join(" · ");
	return (
		<BaseCard title={title} subtitle={summary} accentColor="#55c7f7" footer={`Moebot NEXT · ${algorithm ?? "ga"} · ${costMs ?? 0}ms`}>
			<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.lg }}>
				<InfoPanel event={event} music={music} options={options} />
				{decks.length === 0 ? (
					<div style={emptyStyle}>没有得到推荐结果，请检查 Suite 数据或换用更宽松的参数。</div>
				) : decks.map((deck, index) => (
					<DeckBlock key={index} deck={deck} source={assetSource} />
				))}
				{warnings.length > 0 && (
					<div style={warningStyle}>{warnings.slice(0, 4).join(" · ")}</div>
				)}
			</div>
		</BaseCard>
	);
}

function InfoPanel({ event, music, options }: { event?: any; music?: any; options?: any }) {
	const fixed = [
		...(options?.fixedCards ?? []).map((id: number) => `卡${id}`),
		...(options?.fixedCharacters ?? []).map((id: number) => `角色${id}`),
	];
	const items = [
		["模式", modeLabel(options?.mode)],
		["活动", event?.name ?? (options?.eventId ? `#${options.eventId}` : "无")],
		["歌曲", music?.title ?? (options?.musicId === 10000 ? "おまかせ" : `#${options?.musicId ?? "-"}`)],
		["难度", String(options?.difficulty ?? "master").toUpperCase()],
		["Live", liveTypeLabel(options?.liveType)],
		["目标", targetLabel(options?.target)],
		["算法", String(options?.algorithm ?? "ga").toUpperCase()],
		["固定", fixed.length > 0 ? fixed.join(" / ") : "无"],
		...(options?.challengeCharacterId ? [["挑战角色", `角色${options.challengeCharacterId}`]] : []),
		...(options?.targetBonusList?.length ? [["目标加成", `${options.targetBonusList.join(" / ")}%`]] : []),
	];
	return (
		<div style={panelStyle}>
			{items.map(([label, value]) => (
				<div key={label} style={{ display: "flex", flexDirection: "column", gap: 4, flex: 1, minWidth: 112 }}>
					<div style={{ display: "flex", fontSize: 18, color: theme.colors.textMuted }}>{label}</div>
					<div style={{ display: "flex", fontSize: 24, fontWeight: 800, color: theme.colors.text }}>{value}</div>
				</div>
			))}
		</div>
	);
}

function DeckBlock({ deck, source }: { deck: DeckRecommendDeck; source: string }) {
	return (
		<div style={deckStyle}>
			<div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: theme.spacing.md }}>
				<div style={{ display: "flex", alignItems: "baseline", gap: theme.spacing.sm }}>
					<div style={{ display: "flex", fontSize: 28, fontWeight: 900, color: "#2196f3" }}>#{deck.rank ?? "-"}</div>
					<div style={{ display: "flex", fontSize: 18, color: theme.colors.textSecondary }}>推荐卡组</div>
				</div>
				<div style={{ display: "flex", gap: theme.spacing.md, fontSize: 19, color: theme.colors.textSecondary }}>
					<span style={{ display: "flex" }}>{deck.valueLabel ?? "主值"}：{formatNumber(deck.value ?? deck.eventPoint ?? deck.score)}</span>
					<span style={{ display: "flex" }}>活动PT：{formatNumber(deck.eventPoint ?? deck.score)}</span>
					<span style={{ display: "flex" }}>加成：{formatPercent(deck.eventBonus)}</span>
					<span style={{ display: "flex" }}>综合力：{formatNumber((deck.power as any)?.total)}</span>
					<span style={{ display: "flex" }}>实效：{formatNumber(deck.multiLiveScoreUp)}</span>
				</div>
			</div>
			<div style={{ display: "flex", gap: theme.spacing.sm, justifyContent: "space-between" }}>
				{(deck.cards ?? []).slice(0, 5).map((entry, index) => (
					<CardBox key={`${entry.cardId ?? index}`} entry={entry} source={source} leader={index === 0} />
				))}
			</div>
		</div>
	);
}

function CardBox({ entry, source, leader }: { entry: any; source: string; leader?: boolean }) {
	const card = entry.card ?? entry;
	return (
		<div style={{ display: "flex", flexDirection: "column", alignItems: "center", gap: 6, width: 128 }}>
			<div style={{ display: "flex", position: "relative" }}>
				<SekaiCardThumbnail
					size={112}
					rarity={card.cardRarityType ?? card.rarity ?? "rarity_1"}
					attr={card.attr ?? "cute"}
					imageUrl={card.thumbnailUrl ?? card.normalThumbnailUrl}
					compositeLayers={card.compositeLayers}
					isTrained={card.isTrained}
					mastery={card.masterRank ?? card.mastery ?? 0}
					characterName={card.characterName}
				/>
				{leader && <div style={leaderBadge}>队长</div>}
			</div>
			<div style={{ display: "flex", fontSize: 16, fontWeight: 700, color: theme.colors.text, textAlign: "center" }}>{card.characterName ?? card.prefix ?? `#${entry.cardId}`}</div>
			<div style={{ display: "flex", fontSize: 15, fontWeight: 700, color: theme.colors.text }}>#{entry.cardId}</div>
			<div style={{ display: "flex", fontSize: 14, color: theme.colors.textSecondary }}>
				Lv.{entry.level ?? "-"} · SLv.{entry.skillLevel ?? "-"} · MR{entry.masterRank ?? 0}
			</div>
			{entry.eventBonus && <div style={{ display: "flex", fontSize: 14, color: "#e67e22" }}>+{entry.eventBonus}</div>}
		</div>
	);
}

const panelStyle = {
	display: "flex",
	gap: theme.spacing.md,
	padding: theme.spacing.md,
	backgroundColor: theme.colors.surface,
	border: `1px solid ${theme.colors.border}`,
	borderRadius: theme.borderRadius.lg,
	flexWrap: "wrap" as const,
};

const deckStyle = {
	display: "flex",
	flexDirection: "column" as const,
	gap: theme.spacing.md,
	padding: theme.spacing.md,
	backgroundColor: "#ffffff",
	border: `1px solid ${theme.colors.border}`,
	borderRadius: theme.borderRadius.lg,
};

const emptyStyle = {
	display: "flex",
	padding: theme.spacing.lg,
	backgroundColor: theme.colors.surface,
	borderRadius: theme.borderRadius.lg,
	fontSize: 22,
	color: theme.colors.textSecondary,
};

const warningStyle = {
	display: "flex",
	padding: theme.spacing.md,
	backgroundColor: "#fff8e1",
	border: "1px solid #ffe082",
	borderRadius: theme.borderRadius.md,
	fontSize: 17,
	color: "#8d6e00",
};

const leaderBadge = {
	position: "absolute" as const,
	top: 4,
	left: 4,
	display: "flex",
	padding: "2px 6px",
	backgroundColor: "rgba(33, 150, 243, 0.9)",
	borderRadius: 999,
	fontSize: 13,
	fontWeight: 800,
	color: "#fff",
};

function formatNumber(value: unknown): string {
	const n = Number(value);
	return Number.isFinite(n) ? Math.round(n).toLocaleString("zh-CN") : "-";
}

function formatPercent(value: unknown): string {
	const n = Number(value);
	return Number.isFinite(n) ? `${Math.round(n)}%` : "-";
}

function liveTypeLabel(value: unknown): string {
	switch (String(value || "multi").toLowerCase()) {
		case "solo": return "单人";
		case "auto": return "自动";
		case "cheerful": return "欢乐嘉年华";
		default: return "多人";
	}
}

function modeLabel(value: unknown): string {
	switch (String(value || "event").toLowerCase()) {
		case "strongest": return "最强/长草";
		case "challenge": return "挑战Live";
		case "bonus": return "加成/控分";
		default: return "活动";
	}
}

function targetLabel(value: unknown): string {
	switch (String(value || "score").toLowerCase()) {
		case "power": return "综合力";
		case "skill": return "实效";
		case "bonus": return "加成";
		default: return "活动点/分数";
	}
}

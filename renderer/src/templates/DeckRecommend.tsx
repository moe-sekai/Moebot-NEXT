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
	supportDeckBonus?: number;
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
		profile?.userId ? `UID ${maskUID(String(profile.userId))}` : undefined,
		event?.name,
	].filter(Boolean).join(" · ");
	const footer = [
		"功能移植并修改自 33Kit",
		algorithm ? `${String(algorithm).toUpperCase()} 算法` : undefined,
		Number.isFinite(Number(costMs)) ? `耗时 ${Math.round(Number(costMs))}ms` : undefined,
	].filter(Boolean).join(" · ");

	return (
		<BaseCard title={title} subtitle={summary} accentColor={accentPink} footer={footer}>
			<div style={pageStyle}>
				<ScenarioPanel event={event} music={music} options={options} />
				<ResultTable decks={decks} source={assetSource} />
				{warnings.length > 0 && <div style={warningStyle}>{warnings.slice(0, 4).join(" · ")}</div>}
			</div>
		</BaseCard>
	);
}

function ScenarioPanel({ event, music, options }: { event?: any; music?: any; options?: any }) {
	const eventId = options?.eventId ?? event?.id;
	const eventName = event?.name ?? (eventId ? `活动 #${eventId}` : "无活动");
	const musicTitle = music?.title ?? (options?.musicId === 10000 ? "おまかせ（所有歌曲平均）" : `#${options?.musicId ?? "-"}`);
	const difficulty = String(options?.difficulty ?? "master").toUpperCase();
	const isPresetDefault = Boolean(music?.isPresetDefault ?? options?.isPresetDefault);

	return (
		<div style={scenarioStyle}>
			<div style={scenarioTopStyle}>
				<div style={scenarioMainStyle}>
					<div style={scenarioTitleRowStyle}>
						<span style={scenarioTitleStyle}>{modeLabel(options?.mode)}</span>
						{eventId ? <span style={scenarioTitleAccentStyle}>#{eventId}</span> : null}
					</div>
					<div style={scenarioMetaRowStyle}>
						<span style={scenarioStrongTextStyle}>{eventName}</span>
						<span style={scenarioDividerStyle}>·</span>
						<span style={scenarioTextStyle}>{musicTitle}</span>
						{isPresetDefault ? <span style={defaultBadgeStyle}>默认</span> : null}
					</div>
				</div>
				<div style={scenarioSideStyle}>
					<MetaPill label="难度" value={difficulty} />
					<MetaPill label="Live" value={liveTypeLabel(options?.liveType)} />
					<MetaPill label="目标" value={targetLabel(options?.target)} />
				</div>
			</div>
			<div style={skillOrderStyle}>
				<span style={skillOrderLabelStyle}>技能顺序</span>
				<span style={skillOrderTextStyle}>平均情况 → BloomFes 花前技能吸取 → 平均值</span>
			</div>
		</div>
	);
}

function MetaPill({ label, value }: { label: string; value: string }) {
	return (
		<div style={metaPillStyle}>
			<span style={metaPillLabelStyle}>{label}</span>
			<span style={metaPillValueStyle}>{value}</span>
		</div>
	);
}

function ResultTable({ decks, source }: { decks: DeckRecommendDeck[]; source: string }) {
	if (decks.length === 0) {
		return <div style={emptyStyle}>没有得到推荐结果，请检查 Suite 数据或换用更宽松的参数。</div>;
	}
	return (
		<div style={tableStyle}>
			<div style={headerRowStyle}>
				<div style={ptvHeaderCellStyle}>PT</div>
				<div style={cardsHeaderCellStyle}>卡组</div>
				<div style={metricHeaderCellStyle}>加成</div>
				<div style={metricHeaderCellStyle}>实效</div>
				<div style={metricHeaderCellStyle}>综合力</div>
			</div>
			<div style={tableBodyStyle}>
				{decks.map((deck, index) => <DeckRow key={index} deck={deck} source={source} index={index} />)}
			</div>
		</div>
	);
}

function DeckRow({ deck, source, index }: { deck: DeckRecommendDeck; source: string; index: number }) {
	const rank = Number(deck.rank ?? index + 1);
	const support = Number(deck.supportDeckBonus);
	const hasSupport = Number.isFinite(support) && support > 0;
	return (
		<div style={index === 0 ? rowStyleTop : rowStyle}>
			<div style={ptvCellStyle}>
				<div style={rankBadgeStyle(rank)}>{rank}</div>
				<div style={ptvValueStyle}>{formatNumber(deck.value ?? deck.eventPoint ?? deck.score)}</div>
				<div style={ptvLabelStyle}>{String(deck.valueLabel ?? "活动PT")}</div>
			</div>
			<div style={cardsCellStyle}>
				{(deck.cards ?? []).slice(0, 5).map((entry, i) => <CardMini key={`${entry.cardId ?? i}`} entry={entry} source={source} />)}
			</div>
			<MetricCell value={formatPercent(deck.eventBonus)} sub={hasSupport ? `主 ${formatPercent(Math.max(0, Number(deck.eventBonus) - support))} + 支援 ${formatPercent(support)}` : undefined} />
			<MetricCell value={formatPercent(deck.multiLiveScoreUp)} sub={undefined} />
			<MetricCell value={formatNumber((deck.power as any)?.total)} sub={undefined} />
		</div>
	);
}

function CardMini({ entry, source }: { entry: any; source: string }) {
	const card = entry.card ?? entry;
	const rarity = card.cardRarityType ?? card.rarity ?? "rarity_1";
	const trained = resolveCardTrained(card, entry, rarity);
	const normalUrl = card.thumbnailUrl ?? card.normalThumbnailUrl;
	const trainedUrl = card.trainedThumbnailUrl ?? normalUrl;
	const skillLevel = Number(entry.skillLevel ?? card.skillLevel ?? 0);
	const masterRank = Number(card.masterRank ?? card.mastery ?? entry.masterRank ?? 0);
	const bonus = Number(entry.eventBonus ?? card.eventBonus ?? 0);
	return (
		<div style={cardMiniStyle}>
			<SekaiCardThumbnail
				size={58}
				rarity={rarity}
				attr={card.attr ?? "cute"}
				imageUrl={trained ? trainedUrl : normalUrl}
				compositeLayers={card.compositeLayers}
				isTrained={trained}
				mastery={masterRank}
				characterName={card.characterName}
			/>
			<div style={cardSubStyle}>{cardSubText(skillLevel, masterRank)}</div>
			{Number.isFinite(bonus) && bonus > 0 ? <div style={cardBonusStyle}>+{Math.round(bonus)}%</div> : <div style={cardBonusEmptyStyle}>·</div>}
		</div>
	);
}

function cardSubText(skillLevel: number, masterRank: number): string {
	const parts: string[] = [];
	if (Number.isFinite(skillLevel) && skillLevel > 0) parts.push(`SLv.${skillLevel}`);
	if (Number.isFinite(masterRank) && masterRank > 0) parts.push(`M${masterRank}`);
	return parts.length > 0 ? parts.join(" · ") : "—";
}

function resolveCardTrained(card: any, entry: any, rarity: string): boolean {
	if (typeof card.isTrained === "boolean") return card.isTrained;
	if (typeof entry.isTrained === "boolean") return entry.isTrained;
	const defaultImage = card.defaultImage ?? entry.defaultImage;
	if (defaultImage === "special_training") return true;
	if (defaultImage === "original") return false;
	return rarity === "rarity_3" || rarity === "rarity_4";
}

function MetricCell({ value, sub }: { value: string; sub?: string }) {
	return (
		<div style={metricCellStyle}>
			<div style={metricValueStyle}>{value}</div>
			{sub ? <div style={metricSubStyle}>{sub}</div> : null}
		</div>
	);
}

const accentPink = "#f38ab8";

const pageStyle = {
	display: "flex",
	flexDirection: "column" as const,
	gap: 16,
};

const scenarioStyle = {
	display: "flex",
	flexDirection: "column" as const,
	gap: 12,
	padding: "18px 20px",
	backgroundColor: "#ffffff",
	border: "1px solid rgba(232,174,204,.45)",
	borderRadius: 18,
	boxShadow: "0 6px 18px rgba(187,112,154,.10)",
};

const scenarioTopStyle = { display: "flex", alignItems: "flex-start", justifyContent: "space-between", gap: 16, flexWrap: "wrap" as const };
const scenarioMainStyle = { display: "flex", flexDirection: "column" as const, gap: 6, flex: 1, minWidth: 320 };
const scenarioTitleRowStyle = { display: "flex", alignItems: "baseline", gap: 8 };
const scenarioTitleStyle = { display: "flex", fontSize: 26, fontWeight: 950, color: "#2b2330", letterSpacing: "0.5px" };
const scenarioTitleAccentStyle = { display: "flex", fontSize: 22, fontWeight: 900, color: accentPink };
const scenarioMetaRowStyle = { display: "flex", alignItems: "baseline", gap: 8, flexWrap: "wrap" as const };
const scenarioStrongTextStyle = { display: "flex", fontSize: 17, fontWeight: 800, color: "#3d3640" };
const scenarioTextStyle = { display: "flex", fontSize: 16, fontWeight: 700, color: "#5c535e" };
const scenarioDividerStyle = { display: "flex", fontSize: 16, color: "#c2a5b6" };
const scenarioSideStyle = { display: "flex", gap: 8, alignItems: "stretch", flexWrap: "wrap" as const };
const defaultBadgeStyle = {
	display: "flex",
	padding: "2px 7px",
	borderRadius: 999,
	backgroundColor: "rgba(200,188,198,.22)",
	color: "#9b8a96",
	fontSize: 11,
	fontWeight: 800,
	letterSpacing: "0.5px",
};

const metaPillStyle = {
	display: "flex",
	flexDirection: "column" as const,
	alignItems: "center",
	gap: 2,
	padding: "6px 12px",
	minWidth: 64,
	backgroundColor: "rgba(255,242,247,.9)",
	borderRadius: 10,
	border: "1px solid rgba(232,174,204,.45)",
};
const metaPillLabelStyle = { display: "flex", color: "#a07b91", fontSize: 11, fontWeight: 800, letterSpacing: "1px" };
const metaPillValueStyle = { display: "flex", color: "#2f2930", fontSize: 16, fontWeight: 900 };

const skillOrderStyle = {
	display: "flex",
	alignItems: "center",
	gap: 8,
	paddingTop: 10,
	borderTop: "1px dashed rgba(232,174,204,.55)",
};
const skillOrderLabelStyle = {
	display: "flex",
	padding: "3px 9px",
	borderRadius: 999,
	backgroundColor: "rgba(243,138,184,.14)",
	color: accentPink,
	fontSize: 12,
	fontWeight: 900,
	letterSpacing: "1px",
};
const skillOrderTextStyle = { display: "flex", fontSize: 14, fontWeight: 700, color: "#5c535b" };

const tableStyle = {
	display: "flex",
	flexDirection: "column" as const,
	backgroundColor: "#ffffff",
	border: "1px solid rgba(232,174,204,.45)",
	borderRadius: 20,
	boxShadow: "0 8px 22px rgba(187,112,154,.12)",
	overflow: "hidden" as const,
};
const tableBodyStyle = { display: "flex", flexDirection: "column" as const };

const ptvColumnWidth = 72;
const cardsColumnWidth = 332;
const metricColumnWidth = 88;

const headerRowStyle = {
	display: "flex",
	alignItems: "center",
	gap: 8,
	padding: "11px 16px",
	backgroundColor: "rgba(255,238,246,.9)",
	borderBottom: "1px solid rgba(232,174,204,.45)",
	fontSize: 16,
	fontWeight: 900,
	color: "#71465c",
	letterSpacing: "1px",
};

const rowStyle = {
	display: "flex",
	gap: 8,
	alignItems: "center",
	padding: "12px 16px",
	borderTop: "1px solid rgba(232,174,204,.28)",
};
const rowStyleTop = {
	...rowStyle,
	borderTop: "none",
	background: "linear-gradient(90deg, rgba(255,232,243,.5), rgba(255,255,255,0))",
};

const ptvHeaderCellStyle = { display: "flex", width: ptvColumnWidth, justifyContent: "center" };
const cardsHeaderCellStyle = { display: "flex", width: cardsColumnWidth, justifyContent: "flex-start", paddingLeft: 4 };
const metricHeaderCellStyle = { display: "flex", width: metricColumnWidth, justifyContent: "center" };

const ptvCellStyle = { display: "flex", flexDirection: "column" as const, alignItems: "center", gap: 4, width: ptvColumnWidth };
const ptvValueStyle = { display: "flex", fontSize: 22, fontWeight: 950, color: "#2b2330" };
const ptvLabelStyle = { display: "flex", fontSize: 11, fontWeight: 800, color: "#9b8a96", letterSpacing: "0.5px" };

const cardsCellStyle = { display: "flex", gap: 6, alignItems: "flex-start", width: cardsColumnWidth, paddingLeft: 4 };

const cardMiniStyle = { display: "flex", flexDirection: "column" as const, alignItems: "center", gap: 3, width: 60 };
const cardSubStyle = { display: "flex", fontSize: 10, color: "#4b4650", fontWeight: 700, whiteSpace: "nowrap" as const };
const cardBonusStyle = {
	display: "flex",
	padding: "1px 6px",
	borderRadius: 999,
	backgroundColor: "rgba(10,143,98,.12)",
	color: "#0a8f62",
	fontSize: 10,
	fontWeight: 900,
	whiteSpace: "nowrap" as const,
};
const cardBonusEmptyStyle = { display: "flex", fontSize: 10, color: "#cdc1cb", fontWeight: 700 };

const metricCellStyle = { display: "flex", flexDirection: "column" as const, alignItems: "center", gap: 3, width: metricColumnWidth };
const metricValueStyle = { display: "flex", fontSize: 21, fontWeight: 950, color: "#2b2330" };
const metricSubStyle = { display: "flex", fontSize: 10, color: "#9b8a96", fontWeight: 700, textAlign: "center" as const };

const rankBadgeBase = {
	display: "flex",
	alignItems: "center",
	justifyContent: "center",
	width: 22,
	height: 22,
	borderRadius: 999,
	fontSize: 13,
	fontWeight: 900,
};
function rankBadgeStyle(rank: number) {
	if (rank === 1) return { ...rankBadgeBase, backgroundColor: "#f7c95a", color: "#5a3d00", boxShadow: "0 2px 6px rgba(247,201,90,.6)" };
	if (rank === 2) return { ...rankBadgeBase, backgroundColor: "#d6d6e1", color: "#3a3947" };
	if (rank === 3) return { ...rankBadgeBase, backgroundColor: "#e7b58d", color: "#5a3300" };
	return { ...rankBadgeBase, backgroundColor: "#efe2ea", color: "#7a6072" };
}

const emptyStyle = { display: "flex", padding: theme.spacing.lg, backgroundColor: "rgba(255,255,255,.78)", borderRadius: 18, fontSize: 22, color: theme.colors.textSecondary };
const warningStyle = { display: "flex", padding: "10px 14px", backgroundColor: "rgba(255,248,225,.75)", border: "1px solid rgba(255,224,130,.85)", borderRadius: 14, fontSize: 15, color: "#705800", lineHeight: 1.45 };
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
		case "strongest": return "最强/长草组卡";
		case "challenge": return "挑战 Live 组卡";
		case "bonus": return "加成/控分组卡";
		case "mysekai": return "烤森组卡";
		default: return "活动组卡";
	}
}

function targetLabel(value: unknown): string {
	switch (String(value || "score").toLowerCase()) {
		case "power": return "综合力";
		case "skill": return "实效";
		case "bonus": return "加成";
		case "mysekai": return "烤森PT";
		default: return "活动点/分数";
	}
}

function maskUID(value: string): string {
	if (value.length <= 6) return value;
	return `${value.slice(0, 2)}******${value.slice(-4)}`;
}

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
		event?.name,
	].filter(Boolean).join(" · ");
	const footer = [
		"功能移植并修改自 33Kit",
		algorithm ? `${String(algorithm).toUpperCase()} 算法` : undefined,
		Number.isFinite(Number(costMs)) ? `耗时 ${Math.round(Number(costMs))}ms` : undefined,
	].filter(Boolean).join(" · ");

	return (
		<BaseCard title={title} subtitle={summary} accentColor="#f38ab8" footer={footer}>
			<div style={pageStyle}>
				<ProfileStrip profile={profile} regionLabel={regionLabel} />
				<ScenarioPanel event={event} music={music} options={options} />
				<ResultTable decks={decks} source={assetSource} />
				{warnings.length > 0 && <div style={warningStyle}>{warnings.slice(0, 4).join(" · ")}</div>}
			</div>
		</BaseCard>
	);
}

function ProfileStrip({ profile, regionLabel }: { profile?: any; regionLabel?: string }) {
	if (!profile?.name && !regionLabel) return null;
	return (
		<div style={profileStripStyle}>
			<div style={avatarStyle}>{String(profile?.name ?? regionLabel ?? "M").slice(0, 1).toUpperCase()}</div>
			<div style={{ display: "flex", flexDirection: "column", gap: 3 }}>
				<div style={{ display: "flex", fontSize: 24, fontWeight: 900, color: "#2b2b32" }}>{profile?.name ?? "Moebot"}</div>
				<div style={{ display: "flex", fontSize: 15, color: "#56515c" }}>
					{regionLabel ?? "SEKAI"}{profile?.userId ? ` · UID ${maskUID(String(profile.userId))}` : ""}{profile?.source ? ` · ${profile.source}` : " Suite数据"}
				</div>
			</div>
		</div>
	);
}

function ScenarioPanel({ event, music, options }: { event?: any; music?: any; options?: any }) {
	return (
		<div style={scenarioStyle}>
			<div style={scenarioMainStyle}>
				<div style={scenarioTitleStyle}>{modeLabel(options?.mode)}{options?.eventId || event?.id ? ` #${options?.eventId ?? event?.id}` : ""}</div>
				<div style={scenarioTextStyle}>{event?.name ?? (options?.eventId ? `活动 #${options.eventId}` : "无活动")}</div>
				<div style={scenarioTextStyle}>歌曲：{music?.title ?? (options?.musicId === 10000 ? "おまかせ（所有歌曲平均）" : `#${options?.musicId ?? "-"}`)}</div>
			</div>
			<div style={scenarioSideStyle}>
				<MetaPill label="难度" value={String(options?.difficulty ?? "master").toUpperCase()} />
				<MetaPill label="Live" value={liveTypeLabel(options?.liveType)} />
				<MetaPill label="目标" value={targetLabel(options?.target)} />
			</div>
			<div style={skillOrderStyle}>技能顺序：平均情况 → BloomFes 花前技能吸取 → 平均值</div>
		</div>
	);
}

function MetaPill({ label, value }: { label: string; value: string }) {
	return (
		<div style={metaPillStyle}>
			<span style={{ display: "flex", color: "#8a7b86", fontSize: 12, fontWeight: 800 }}>{label}</span>
			<span style={{ display: "flex", color: "#2f2930", fontSize: 15, fontWeight: 900 }}>{value}</span>
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
				<div style={ptvHeaderCellStyle}>PTV</div>
				<div style={cardsHeaderCellStyle}>卡组</div>
				<div style={metricHeaderCellStyle}>加成</div>
				<div style={metricHeaderCellStyle}>实效</div>
				<div style={metricHeaderCellStyle}>综合力</div>
			</div>
			{decks.map((deck, index) => <DeckRow key={index} deck={deck} source={source} />)}
		</div>
	);
}

function DeckRow({ deck, source }: { deck: DeckRecommendDeck; source: string }) {
	return (
		<div style={rowStyle}>
			<div style={ptvCellStyle}>
				<div style={ptvValueStyle}>{formatNumber(deck.value ?? deck.eventPoint ?? deck.score)}</div>
				<div style={mutedSmallStyle}>{String(deck.valueLabel ?? "GA")}</div>
			</div>
			<div style={cardsCellStyle}>
				{(deck.cards ?? []).slice(0, 5).map((entry, index) => <CardMini key={`${entry.cardId ?? index}`} entry={entry} source={source} />)}
			</div>
			<MetricCell value={formatPercent(deck.eventBonus)} sub={supportBonusText(deck)} />
			<MetricCell value={formatPercent(deck.multiLiveScoreUp)} sub="队友 200" />
			<MetricCell value={formatNumber((deck.power as any)?.total)} sub="队友 250000" />
		</div>
	);
}

function CardMini({ entry, source }: { entry: any; source: string }) {
	const card = entry.card ?? entry;
	return (
		<div style={cardMiniStyle}>
			<SekaiCardThumbnail
				size={56}
				rarity={card.cardRarityType ?? card.rarity ?? "rarity_1"}
				attr={card.attr ?? "cute"}
				imageUrl={card.thumbnailUrl ?? card.normalThumbnailUrl}
				compositeLayers={card.compositeLayers}
				isTrained={card.isTrained}
				mastery={card.masterRank ?? card.mastery ?? entry.masterRank ?? 0}
				characterName={card.characterName}
			/>
			<div style={cardSubStyle}>SLv.{entry.skillLevel ?? 1} · {formatPercent(entry.skillScoreUp ?? entry.scoreUp ?? 0)}</div>
			<div style={cardBonusStyle}>+{formatPercent(entry.eventBonus ?? 0)}</div>
		</div>
	);
}

function MetricCell({ value, sub }: { value: string; sub: string }) {
	return (
		<div style={metricCellStyle}>
			<div style={metricValueStyle}>{value}</div>
			<div style={mutedSmallStyle}>{sub}</div>
		</div>
	);
}

const pageStyle = {
	display: "flex",
	flexDirection: "column" as const,
	gap: 14,
	padding: 2,
	background: "linear-gradient(180deg, rgba(255,241,248,.72), rgba(255,255,255,.15))",
};

const profileStripStyle = {
	display: "flex",
	alignItems: "center",
	gap: 12,
	alignSelf: "flex-start",
	padding: "12px 16px",
	backgroundColor: "rgba(255,255,255,.82)",
	border: "1px solid rgba(232, 174, 204, .42)",
	borderRadius: 16,
	boxShadow: "0 8px 20px rgba(187, 112, 154, .12)",
};

const avatarStyle = {
	display: "flex",
	alignItems: "center",
	justifyContent: "center",
	width: 54,
	height: 54,
	borderRadius: 12,
	background: "linear-gradient(135deg, #ffd6e8, #bfe8ff)",
	color: "#fff",
	fontSize: 28,
	fontWeight: 900,
	boxShadow: "inset 0 0 0 2px rgba(255,255,255,.7)",
};

const scenarioStyle = {
	display: "flex",
	position: "relative" as const,
	gap: 14,
	padding: 16,
	backgroundColor: "rgba(255,255,255,.78)",
	border: "1px solid rgba(232, 174, 204, .38)",
	borderRadius: 18,
	boxShadow: "0 8px 24px rgba(187, 112, 154, .11)",
	flexWrap: "wrap" as const,
};

const scenarioMainStyle = { display: "flex", flexDirection: "column" as const, gap: 7, flex: 1, minWidth: 360 };
const scenarioTitleStyle = { display: "flex", fontSize: 25, fontWeight: 950, color: "#2f2930" };
const scenarioTextStyle = { display: "flex", fontSize: 17, fontWeight: 800, color: "#4a424a" };
const scenarioSideStyle = { display: "flex", gap: 8, alignItems: "flex-start", flexWrap: "wrap" as const };
const metaPillStyle = { display: "flex", flexDirection: "column" as const, gap: 2, padding: "7px 10px", backgroundColor: "rgba(255,246,250,.9)", borderRadius: 12, border: "1px solid rgba(232,174,204,.35)" };
const skillOrderStyle = { display: "flex", width: "100%", paddingTop: 8, borderTop: "1px solid rgba(232,174,204,.3)", fontSize: 15, fontWeight: 800, color: "#5c535b" };

const tableStyle = {
	display: "flex",
	flexDirection: "column" as const,
	padding: "18px 18px 8px",
	backgroundColor: "rgba(255,255,255,.84)",
	border: "1px solid rgba(232,174,204,.42)",
	borderRadius: 20,
	boxShadow: "0 10px 30px rgba(187,112,154,.14)",
};

const ptvColumnWidth = 72;
const cardsColumnWidth = 330;
const metricColumnWidth = 88;
const headerRowStyle = {
	display: "flex",
	alignItems: "center",
	gap: 10,
	padding: "0 4px 10px",
	fontSize: 23,
	fontWeight: 950,
	color: "#2f2930",
};
const rowStyle = {
	display: "flex",
	gap: 10,
	alignItems: "flex-start",
	padding: "10px 4px",
	borderTop: "1px solid rgba(232,174,204,.28)",
};
const ptvHeaderCellStyle = { display: "flex", width: ptvColumnWidth };
const cardsHeaderCellStyle = { display: "flex", width: cardsColumnWidth };
const metricHeaderCellStyle = { display: "flex", width: metricColumnWidth, justifyContent: "center" };
const ptvCellStyle = { display: "flex", flexDirection: "column" as const, alignItems: "center", gap: 2, width: ptvColumnWidth };
const ptvValueStyle = { display: "flex", fontSize: 22, fontWeight: 950, color: "#363039" };
const cardsCellStyle = { display: "flex", gap: 7, alignItems: "flex-start", width: cardsColumnWidth };
const cardMiniStyle = { display: "flex", flexDirection: "column" as const, alignItems: "center", gap: 1, width: 60 };
const cardSubStyle = { display: "flex", fontSize: 8, color: "#4b4650", whiteSpace: "nowrap" as const };
const cardBonusStyle = { display: "flex", fontSize: 8, color: "#0a8f62", whiteSpace: "nowrap" as const };
const metricCellStyle = { display: "flex", flexDirection: "column" as const, alignItems: "center", gap: 3, width: metricColumnWidth };
const metricValueStyle = { display: "flex", fontSize: 22, fontWeight: 950, color: "#363039", textAlign: "center" as const };
const mutedSmallStyle = { display: "flex", fontSize: 10, color: "#9b929c", fontWeight: 700, textAlign: "center" as const };
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

function formatDeckBonus(deck: DeckRecommendDeck): string {
	const support = Number(deck.supportDeckBonus);
	if (Number.isFinite(support) && support > 0) {
		return `${formatPercent(deck.eventBonus)} + ${formatPercent(support)}`;
	}
	return formatPercent(deck.eventBonus);
}

function supportBonusText(deck: DeckRecommendDeck): string {
	const base = Number(deck.eventBonus);
	const support = Number(deck.supportDeckBonus);
	if (Number.isFinite(base) && Number.isFinite(support) && support > 0) {
		return `${Math.max(0, Math.round(base - support))}.0+${Math.round(support)}.0%`;
	}
	return "活动加成";
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
		default: return "活动组卡";
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

function maskUID(value: string): string {
	if (value.length <= 6) return value;
	return `${value.slice(0, 2)}******${value.slice(-4)}`;
}

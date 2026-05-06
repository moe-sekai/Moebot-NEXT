import { getMusicJacketUrl, type AssetSourceType } from "../../shared";
import { theme } from "../styles/theme";
import { BaseCard } from "./base";

export interface Best30Props {
	title?: string;
	subtitle?: string;
	profile?: Best30Profile;
	average?: number;
	entries?: Best30Entry[];
	candidateCount?: number;
	apCount?: number;
	fcCount?: number;
	missingConstantsCount?: number;
	totalResultCount?: number;
	region?: string;
	regionLabel?: string;
	updatedAt?: number | string;
	updateText?: string;
	formula?: string;
	constantsSource?: string;
	assetSource?: AssetSourceType | string;
}

interface Best30Profile {
	name?: string;
	displayName?: string;
	rank?: number | string;
	userId?: number | string;
	uid?: number | string;
	updateText?: string;
	uploadTime?: number | string;
	updatedAt?: number | string;
	source?: string;
}

interface Best30Entry {
	rank?: number;
	musicId?: number;
	title?: string;
	difficulty?: string;
	difficultyLabel?: string;
	level?: number;
	constant?: number;
	userRating?: number;
	playResult?: string;
	noteCount?: number;
	assetbundleName?: string;
	jacketUrl?: string;
}

const DIFF_COLORS: Record<string, string> = {
	easy: "#6fd58e",
	normal: "#4fb5ff",
	hard: "#ffb347",
	expert: "#ff5d8f",
	master: "#a855f7",
	append: "#fb7185",
};

export function Best30({
	title = "Best30",
	subtitle,
	profile,
	average = 0,
	entries = [],
	candidateCount = entries.length,
	apCount = 0,
	fcCount = 0,
	missingConstantsCount = 0,
	totalResultCount = 0,
	regionLabel,
	updateText,
	formula = "AP=定数；FC=定数-1(≥33) / 定数-1.5(<33)",
	constantsSource,
	assetSource = "main-jp",
}: Best30Props) {
	const name = profile?.displayName ?? profile?.name ?? "未知玩家";
	const uid = profile?.userId ?? profile?.uid;
	const updated = updateText ?? profile?.updateText ?? profile?.uploadTime ?? profile?.updatedAt;
	const meta = [regionLabel, updated ? `更新：${updated}` : undefined, profile?.source ? `来源：${profile.source}` : undefined]
		.filter(Boolean)
		.join(" · ");
	const visibleEntries = entries.slice(0, 30);

	return (
		<BaseCard title={title} subtitle={subtitle ?? meta} accentColor={theme.colors.accent}>
			<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.lg }}>
				<div style={{ display: "flex", gap: theme.spacing.md, alignItems: "stretch" }}>
					<div style={{ display: "flex", flex: 1, flexDirection: "column", gap: theme.spacing.sm, padding: theme.spacing.lg, borderRadius: theme.borderRadius.xl, backgroundColor: theme.colors.surface, border: `1px solid ${theme.colors.border}` }}>
						<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 900 }}>PLAYER</span>
						<div style={{ display: "flex", alignItems: "baseline", justifyContent: "space-between", gap: theme.spacing.md }}>
							<span style={{ display: "flex", color: theme.colors.text, fontSize: theme.fontSize.xl, fontWeight: 900 }}>{name}</span>
							{profile?.rank !== undefined && <span style={{ display: "flex", color: theme.colors.accent, fontSize: theme.fontSize.md, fontWeight: 900 }}>Rank {String(profile.rank)}</span>}
						</div>
						{uid !== undefined && <span style={{ display: "flex", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 800 }}>UID: {String(uid)}</span>}
						{meta && <span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 11, fontWeight: 800 }}>{meta}</span>}
					</div>
					<div style={{ display: "flex", flexDirection: "column", width: 198, gap: theme.spacing.sm, padding: theme.spacing.lg, borderRadius: theme.borderRadius.xl, background: `linear-gradient(135deg, ${theme.colors.accentSoft}, #ffffff)`, border: `1px solid ${theme.colors.borderStrong}` }}>
						<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 900 }}>B30 AVG</span>
						<span style={{ display: "flex", color: theme.colors.accent, fontSize: 38, lineHeight: 1, fontWeight: 900 }}>{fixed2(average)}</span>
						<span style={{ display: "flex", color: theme.colors.textSecondary, fontSize: 11, fontWeight: 800 }}>计入 {visibleEntries.length}/30 首</span>
					</div>
				</div>

				<div style={{ display: "flex", gap: theme.spacing.sm }}>
					<Metric label="AP" value={apCount} color="#ff5d8f" />
					<Metric label="FC" value={fcCount} color="#f59e0b" />
					<Metric label="候选" value={candidateCount} color={theme.colors.accent} />
					<Metric label="总成绩" value={totalResultCount} color="#64748b" />
					<Metric label="缺定数" value={missingConstantsCount} color={missingConstantsCount > 0 ? theme.colors.warning : "#33ccbb"} />
				</div>

				<div style={{ display: "flex", flexWrap: "wrap", gap: theme.spacing.sm }}>
					{visibleEntries.map((entry, index) => <Best30Row key={`${entry.rank ?? index}-${entry.musicId ?? index}-${entry.difficulty ?? ""}`} entry={entry} assetSource={assetSource} />)}
				</div>

				<div style={{ display: "flex", flexDirection: "column", gap: 4, padding: theme.spacing.md, borderRadius: theme.borderRadius.lg, backgroundColor: "rgba(242, 247, 251, 0.74)", border: `1px solid ${theme.colors.border}` }}>
					<span style={{ display: "flex", color: theme.colors.textSecondary, fontSize: theme.fontSize.xs, fontWeight: 900 }}>{formula}</span>
					{constantsSource && <span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 700 }}>定数来源：{constantsSource}</span>}
				</div>
			</div>
		</BaseCard>
	);
}

function Best30Row({ entry, assetSource }: { entry: Best30Entry; assetSource: AssetSourceType | string }) {
	const diff = normalizeDiff(entry.difficulty ?? entry.difficultyLabel ?? "master");
	const diffColor = DIFF_COLORS[diff] ?? theme.colors.accent;
	const jacket = entry.jacketUrl ?? (entry.assetbundleName ? getMusicJacketUrl(entry.assetbundleName, assetSource) : undefined);
	const resultColor = entry.playResult === "AP" ? "#ff5d8f" : "#f59e0b";
	const rank = entry.rank ?? 0;
	const highlight = rank > 0 && rank <= 3;
	return (
		<div style={{ display: "flex", width: 359, minHeight: 88, gap: theme.spacing.sm, padding: 8, borderRadius: theme.borderRadius.lg, backgroundColor: highlight ? "rgba(255, 245, 250, 0.94)" : theme.colors.surface, border: `1px solid ${highlight ? resultColor : theme.colors.border}` }}>
			<div style={{ display: "flex", width: 72, height: 72, borderRadius: theme.borderRadius.md, overflow: "hidden", backgroundColor: theme.colors.surfaceAccent, flexShrink: 0, alignItems: "center", justifyContent: "center" }}>
				{jacket ? <img src={jacket} width={72} height={72} style={{ objectFit: "cover" }} /> : <span style={{ display: "flex", color: theme.colors.accent, fontSize: 12, fontWeight: 900 }}>MUSIC</span>}
			</div>
			<div style={{ display: "flex", flex: 1, flexDirection: "column", gap: 5, minWidth: 0 }}>
				<div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 6 }}>
					<div style={{ display: "flex", gap: 5, alignItems: "center" }}>
						<span style={{ display: "flex", width: 31, justifyContent: "center", color: highlight ? resultColor : theme.colors.textMuted, fontSize: 12, fontWeight: 900 }}>#{String(rank).padStart(2, "0")}</span>
						<span style={{ display: "flex", padding: "3px 7px", borderRadius: theme.borderRadius.round, backgroundColor: diffColor, color: "white", fontSize: 10, fontWeight: 900 }}>{entry.difficultyLabel ?? diff.toUpperCase()}</span>
						{entry.level ? <span style={{ display: "flex", color: diffColor, fontSize: 11, fontWeight: 900 }}>Lv.{entry.level}</span> : null}
					</div>
					<span style={{ display: "flex", color: resultColor, fontSize: 11, fontWeight: 900 }}>{entry.playResult ?? "FC"}</span>
				</div>
				<span style={{ display: "flex", color: theme.colors.text, fontSize: 12, fontWeight: 900, lineHeight: 1.25, maxHeight: 30, overflow: "hidden" }}>#{entry.musicId ?? "-"} {entry.title ?? "未知曲目"}</span>
				<div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", gap: 6 }}>
					<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 800 }}>定数 {fixed1(entry.constant)} · Notes {entry.noteCount ?? "-"}</span>
					<span style={{ display: "flex", color: resultColor, fontSize: theme.fontSize.md, fontWeight: 900 }}>{fixed1(entry.userRating)}</span>
				</div>
			</div>
		</div>
	);
}

function Metric({ label, value, color }: { label: string; value?: number; color: string }) {
	return (
		<div style={{ display: "flex", flex: 1, flexDirection: "column", gap: 4, padding: "10px 12px", borderRadius: theme.borderRadius.lg, backgroundColor: theme.colors.surface, border: `1px solid ${theme.colors.border}` }}>
			<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: 10, fontWeight: 900 }}>{label}</span>
			<span style={{ display: "flex", color, fontSize: theme.fontSize.lg, fontWeight: 900 }}>{Number(value ?? 0).toLocaleString()}</span>
		</div>
	);
}

function normalizeDiff(value: string): string {
	const diff = value.toLowerCase();
	if (diff === "mas" || diff === "ma") return "master";
	if (diff === "exp" || diff === "ex") return "expert";
	if (diff === "apd") return "append";
	return diff;
}

function fixed1(value?: number): string {
	return typeof value === "number" && Number.isFinite(value) ? value.toFixed(1) : "-";
}

function fixed2(value?: number): string {
	return typeof value === "number" && Number.isFinite(value) ? value.toFixed(2) : "0.00";
}

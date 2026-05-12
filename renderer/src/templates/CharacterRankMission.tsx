import { getCharacterIconUrl } from "../../shared";
import { getLocalCharacterIconAssetDataUri } from "../styles/assets";
import { theme } from "../styles/theme";
import { BaseCard } from "./base";

export interface CharacterRankMissionProps {
	title?: string;
	subtitle?: string;
	profile?: { name?: string; rank?: number | string; updateText?: string; source?: string };
	characterId?: number;
	character?: string;
	mode?: "overview" | "all" | string;
	rows?: MissionRow[];
	allRows?: MissionAllRow[];
	assetSource?: string;
}

interface MissionRow {
	missionType?: string;
	title?: string;
	current?: number;
	upper?: number;
	level?: number;
	levelMax?: number;
	nextNeed?: number;
	nextExp?: number;
	progress?: number;
	isEx?: boolean;
}

interface MissionAllRow {
	seq?: number;
	requirement?: number;
	accRequirement?: number;
	exp?: number;
	accExp?: number;
	reached?: boolean;
}

export function CharacterRankMission({ title = "CR任务", subtitle, profile, characterId, character, mode = "overview", rows = [], allRows = [], assetSource = "main-jp" }: CharacterRankMissionProps) {
	const meta = [
		profile?.name ? `玩家：${profile.name}` : undefined,
		profile?.rank ? `Rank ${profile.rank}` : undefined,
		profile?.updateText ? `更新：${profile.updateText}` : undefined,
		profile?.source ? `来源：${profile.source}` : undefined,
	].filter(Boolean).join(" · ");
	const isAllMode = mode === "all";

	return (
		<BaseCard title={title} subtitle={subtitle ?? meta} accentColor={theme.colors.accentLight}>
			<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.lg }}>
				<div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: theme.spacing.md, padding: theme.spacing.md, borderRadius: theme.borderRadius.xl, backgroundColor: theme.colors.surface, border: `1px solid ${theme.colors.border}` }}>
					<div style={{ display: "flex", alignItems: "center", gap: theme.spacing.md, minWidth: 0 }}>
						{characterId ? <CharacterIcon id={characterId} assetSource={assetSource} /> : <CharacterFallback />}
						<div style={{ display: "flex", flexDirection: "column", gap: 4, minWidth: 0 }}>
							<span style={{ display: "flex", color: theme.colors.text, fontSize: theme.fontSize.lg, fontWeight: 900 }}>{character ?? "角色"}</span>
							{meta ? <span style={{ display: "flex", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 800 }}>{meta}</span> : null}
						</div>
					</div>
					<span style={{ display: "flex", flexShrink: 0, padding: "7px 12px", borderRadius: theme.borderRadius.round, backgroundColor: isAllMode ? theme.colors.accentSoft : theme.colors.surfaceAccent, color: theme.colors.accent, fontSize: theme.fontSize.sm, fontWeight: 900 }}>
						{isAllMode ? "档位表" : `${rows.length} 项任务`}
					</span>
				</div>
				{isAllMode ? <AllTable rows={allRows} /> : <OverviewGrid rows={rows} />}
			</div>
		</BaseCard>
	);
}

function OverviewGrid({ rows }: { rows: MissionRow[] }) {
	if (rows.length === 0) {
		return <EmptyState text="暂无可展示的 CR 任务进度" />;
	}
	return (
		<div style={{ display: "flex", flexWrap: "wrap", gap: theme.spacing.md }}>
			{rows.map((row, index) => <MissionCard key={`${row.missionType ?? row.title ?? index}`} row={row} />)}
		</div>
	);
}

function MissionCard({ row }: { row: MissionRow }) {
	const progress = Math.max(0, Math.min(1, Number(row.progress ?? 0)));
	const current = Number(row.current ?? 0);
	const upper = Number(row.upper ?? 0);
	const nextNeed = Number(row.nextNeed ?? 0);
	const accent = row.isEx ? theme.colors.warning : theme.colors.accent;
	return (
		<div style={{ display: "flex", flexDirection: "column", width: 351, gap: theme.spacing.sm, padding: theme.spacing.md, borderRadius: theme.borderRadius.lg, backgroundColor: theme.colors.surface, border: `1px solid ${theme.colors.border}` }}>
			<div style={{ display: "flex", justifyContent: "space-between", gap: theme.spacing.sm, alignItems: "center" }}>
				<span style={{ display: "flex", flex: 1, minWidth: 0, fontSize: theme.fontSize.md, fontWeight: 900, color: theme.colors.text, lineHeight: 1.35 }}>{row.title ?? "任务"}</span>
				<span style={{ display: "flex", flexShrink: 0, color: accent, fontWeight: 900, fontSize: theme.fontSize.sm }}>Lv.{row.level ?? 0}/{row.levelMax ?? 0}</span>
			</div>
			<div style={{ display: "flex", height: 10, borderRadius: 999, backgroundColor: theme.colors.surfaceLight, overflow: "hidden" }}>
				<div style={{ display: "flex", width: `${progress * 100}%`, height: "100%", backgroundColor: accent }} />
			</div>
			<div style={{ display: "flex", justifyContent: "space-between", gap: theme.spacing.sm, color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>
				<span style={{ display: "flex", fontWeight: 800 }}>{current.toLocaleString()} / {upper.toLocaleString()}</span>
				<span style={{ display: "flex", color: nextNeed > 0 ? theme.colors.textSecondary : theme.colors.success, fontWeight: 800 }}>
					{nextNeed > 0 ? `下一档 ${nextNeed.toLocaleString()} EXP+${row.nextExp ?? "?"}` : "已满"}
				</span>
			</div>
		</div>
	);
}

function AllTable({ rows }: { rows: MissionAllRow[] }) {
	if (rows.length === 0) {
		return <EmptyState text="暂无可展示的任务档位" />;
	}
	return (
		<div style={{ display: "flex", flexDirection: "column", borderRadius: theme.borderRadius.xl, overflow: "hidden", border: `1px solid ${theme.colors.border}` }}>
			<TableHeader />
			{rows.map((row, index) => <AllTableRow key={`${row.seq ?? index}`} row={row} />)}
		</div>
	);
}

function TableHeader() {
	return (
		<div style={{ display: "flex", alignItems: "center", gap: theme.spacing.sm, padding: "10px 12px", backgroundColor: theme.colors.surfaceAccent, color: theme.colors.textSecondary, fontSize: theme.fontSize.xs, fontWeight: 900 }}>
			<span style={{ display: "flex", width: 70 }}>档位</span>
			<span style={{ display: "flex", width: 190 }}>累计需求</span>
			<span style={{ display: "flex", width: 130 }}>获得 EXP</span>
			<span style={{ display: "flex", flex: 1 }}>状态</span>
		</div>
	);
}

function AllTableRow({ row }: { row: MissionAllRow }) {
	const reached = Boolean(row.reached);
	const requirement = Number(row.accRequirement ?? row.requirement ?? 0);
	return (
		<div style={{ display: "flex", alignItems: "center", gap: theme.spacing.sm, padding: "10px 12px", backgroundColor: reached ? theme.colors.accentSoft : theme.colors.surface, borderTop: `1px solid ${theme.colors.border}`, color: theme.colors.text, fontSize: theme.fontSize.sm }}>
			<strong style={{ display: "flex", width: 70, color: reached ? theme.colors.accent : theme.colors.textMuted }}>#{row.seq ?? "-"}</strong>
			<span style={{ display: "flex", width: 190, fontWeight: 900 }}>{requirement.toLocaleString()}</span>
			<span style={{ display: "flex", width: 130, color: theme.colors.textSecondary, fontWeight: 800 }}>+{Number(row.exp ?? 0).toLocaleString()}</span>
			<span style={{ display: "flex", flex: 1, justifyContent: "flex-end", color: reached ? theme.colors.success : theme.colors.textMuted, fontWeight: 900 }}>{reached ? "已达成" : "未达成"}</span>
		</div>
	);
}

function EmptyState({ text }: { text: string }) {
	return (
		<div style={{ display: "flex", alignItems: "center", justifyContent: "center", minHeight: 96, padding: theme.spacing.lg, borderRadius: theme.borderRadius.xl, backgroundColor: theme.colors.surface, border: `1px dashed ${theme.colors.borderStrong}`, color: theme.colors.textMuted, fontSize: theme.fontSize.sm, fontWeight: 900 }}>
			{text}
		</div>
	);
}

function CharacterIcon({ id, assetSource: _assetSource }: { id: number; assetSource?: string }) {
	const src = getLocalCharacterIconAssetDataUri(id) ?? getCharacterIconUrl(id);
	return <img src={src} width={52} height={52} style={{ borderRadius: 999, objectFit: "cover", backgroundColor: theme.colors.surfaceLight, border: `2px solid ${theme.colors.borderStrong}` }} />;
}

function CharacterFallback() {
	return (
		<div style={{ display: "flex", width: 52, height: 52, alignItems: "center", justifyContent: "center", borderRadius: 999, backgroundColor: theme.colors.surfaceAccent, color: theme.colors.accent, border: `2px solid ${theme.colors.borderStrong}`, fontSize: theme.fontSize.md, fontWeight: 900 }}>
			CR
		</div>
	);
}

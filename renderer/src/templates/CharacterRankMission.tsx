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

export function CharacterRankMission({ title = "CR任务", subtitle, profile, characterId, character, mode = "overview", rows = [], allRows = [] }: CharacterRankMissionProps) {
	const meta = [
		profile?.name ? `玩家：${profile.name}` : undefined,
		profile?.rank ? `Rank ${profile.rank}` : undefined,
		profile?.updateText ? `更新：${profile.updateText}` : undefined,
		profile?.source ? `来源：${profile.source}` : undefined,
	].filter(Boolean).join(" · ");

	return (
		<BaseCard title={title} subtitle={subtitle ?? meta} accentColor={theme.colors.accentLight}>
			<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.lg }}>
				<div style={{ display: "flex", alignItems: "center", gap: theme.spacing.md, color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>
					{characterId ? <CharacterIcon id={characterId} /> : null}
					<span>{character ?? "角色"}</span>
					{meta && <span>{meta}</span>}
				</div>
				{mode === "all" ? <AllTable rows={allRows} /> : <OverviewGrid rows={rows} />}
			</div>
		</BaseCard>
	);
}

function OverviewGrid({ rows }: { rows: MissionRow[] }) {
	return <div style={{ display: "grid", gridTemplateColumns: "repeat(2, 1fr)", gap: theme.spacing.md }}>{rows.map((row) => <MissionCard key={row.missionType} row={row} />)}</div>;
}

function MissionCard({ row }: { row: MissionRow }) {
	const progress = Math.max(0, Math.min(1, Number(row.progress ?? 0)));
	return (
		<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.sm, padding: theme.spacing.md, borderRadius: theme.borderRadius.lg, backgroundColor: theme.colors.surface, border: `1px solid ${theme.colors.border}` }}>
			<div style={{ display: "flex", justifyContent: "space-between", gap: theme.spacing.sm, alignItems: "center" }}>
				<span style={{ fontSize: theme.fontSize.md, fontWeight: 900, color: theme.colors.text }}>{row.title}</span>
				<span style={{ color: row.isEx ? theme.colors.warning : theme.colors.accent, fontWeight: 900, fontSize: theme.fontSize.sm }}>Lv.{row.level ?? 0}/{row.levelMax ?? 0}</span>
			</div>
			<div style={{ height: 10, borderRadius: 999, backgroundColor: theme.colors.surfaceLight, overflow: "hidden" }}>
				<div style={{ width: `${progress * 100}%`, height: "100%", backgroundColor: row.isEx ? theme.colors.warning : theme.colors.accent }} />
			</div>
			<div style={{ display: "flex", justifyContent: "space-between", color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>
				<span>{(row.current ?? 0).toLocaleString()} / {(row.upper ?? 0).toLocaleString()}</span>
				<span>{row.nextNeed ? `下一档 ${row.nextNeed.toLocaleString()} EXP+${row.nextExp ?? "?"}` : "已满"}</span>
			</div>
		</div>
	);
}

function AllTable({ rows }: { rows: MissionAllRow[] }) {
	return (
		<div style={{ display: "flex", flexDirection: "column", borderRadius: theme.borderRadius.lg, overflow: "hidden", border: `1px solid ${theme.colors.border}` }}>
			{rows.map((row) => (
				<div key={row.seq} style={{ display: "grid", gridTemplateColumns: "80px 1fr 1fr 1fr", gap: theme.spacing.sm, padding: theme.spacing.sm, backgroundColor: row.reached ? theme.colors.accentSoft : theme.colors.surface, color: theme.colors.text }}>
					<strong>#{row.seq}</strong>
					<span>需求 {(row.accRequirement ?? row.requirement)?.toLocaleString()}</span>
					<span>EXP +{row.exp}</span>
					<span>{row.reached ? "已达成" : "未达成"}</span>
				</div>
			))}
		</div>
	);
}

function CharacterIcon({ id }: { id: number }) {
	const src = getLocalCharacterIconAssetDataUri(id) ?? getCharacterIconUrl(id);
	return <img src={src} width={40} height={40} style={{ borderRadius: 999, objectFit: "cover" }} />;
}

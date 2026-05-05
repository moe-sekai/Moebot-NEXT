import { getCharacterIconUrl } from "../../shared";
import { getLocalCharacterIconAssetDataUri } from "../styles/assets";
import { theme } from "../styles/theme";
import { BaseCard } from "./base";

export interface AnvoListProps {
	title?: string;
	subtitle?: string;
	profile?: {
		name?: string;
		rank?: number | string;
		updateText?: string;
		source?: string;
	};
	characterId?: number;
	entries?: AnvoEntry[];
	ownedCount?: number;
	totalCount?: number;
}

interface AnvoEntry {
	musicVocalId?: number;
	musicId?: number;
	title?: string;
	characterIds?: number[];
	coverUrl?: string;
	owned?: boolean;
}

export function AnvoList({ title = "Another Vocal 持有情况", subtitle, profile, entries = [], ownedCount, totalCount }: AnvoListProps) {
	const total = totalCount ?? entries.length;
	const owned = ownedCount ?? entries.filter((entry) => entry.owned).length;
	const meta = [
		profile?.name ? `玩家：${profile.name}` : undefined,
		profile?.rank ? `Rank ${profile.rank}` : undefined,
		profile?.updateText ? `更新：${profile.updateText}` : undefined,
		profile?.source ? `来源：${profile.source}` : undefined,
	].filter(Boolean).join(" · ");

	return (
		<BaseCard title={title} subtitle={subtitle ?? `已持有 ${owned}/${total} 首`} accentColor={theme.colors.accentLight}>
			<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.lg }}>
				{meta && <div style={{ color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>{meta}</div>}
				<div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: theme.spacing.md }}>
					{entries.map((entry, index) => <AnvoCard key={`${entry.musicVocalId ?? index}`} entry={entry} />)}
				</div>
			</div>
		</BaseCard>
	);
}

function AnvoCard({ entry }: { entry: AnvoEntry }) {
	const owned = Boolean(entry.owned);
	return (
		<div
			style={{
				display: "flex",
				flexDirection: "column",
				gap: theme.spacing.sm,
				padding: theme.spacing.sm,
				borderRadius: theme.borderRadius.lg,
				backgroundColor: owned ? theme.colors.surface : theme.colors.surfaceLight,
				border: `1px solid ${owned ? theme.colors.border : theme.colors.borderStrong}`,
				opacity: owned ? 1 : 0.58,
			}}
		>
			<div style={{ width: "100%", aspectRatio: "1 / 1", borderRadius: theme.borderRadius.md, overflow: "hidden", backgroundColor: theme.colors.surfaceMuted }}>
				{entry.coverUrl ? <img src={entry.coverUrl} width="100%" height="100%" style={{ objectFit: "cover" }} /> : null}
			</div>
			<div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
				<div style={{ fontSize: theme.fontSize.sm, fontWeight: 900, color: theme.colors.text, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
					#{entry.musicId} {entry.title ?? "未知曲目"}
				</div>
				<div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: theme.spacing.sm }}>
					<div style={{ display: "flex", gap: 3 }}>{(entry.characterIds ?? []).map((id) => <CharacterIcon key={id} id={id} />)}</div>
					<span style={{ fontSize: theme.fontSize.xs, fontWeight: 900, color: owned ? theme.colors.success : theme.colors.textMuted }}>
						{owned ? "已持有" : "未持有"}
					</span>
				</div>
			</div>
		</div>
	);
}

function CharacterIcon({ id }: { id: number }) {
	const src = getLocalCharacterIconAssetDataUri(id) ?? getCharacterIconUrl(id);
	return <img src={src} width={24} height={24} style={{ borderRadius: 999, objectFit: "cover" }} />;
}

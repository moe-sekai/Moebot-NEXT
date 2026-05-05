import { getMusicJacketUrl, type AssetSourceType } from "../../shared";
import { BaseCard } from "./base";
import { theme } from "../styles/theme";

export interface MusicListProps {
	title: string;
	subtitle?: string;
	musics: Array<{
		id: number;
		title: string;
		pronunciation?: string;
		assetbundleName?: string;
		composer?: string;
		lyricist?: string;
		arranger?: string;
		publishedAt?: number;
		difficulties?: Array<{
			musicDifficulty?: string;
			difficulty?: string;
			playLevel?: number;
			level?: number;
		}>;
	}>;
	page?: number;
	totalPages?: number;
	total?: number;
	assetSource?: AssetSourceType | string;
}

export function MusicList({
	title,
	subtitle,
	musics,
	page,
	totalPages,
	total,
	assetSource = "main-jp",
}: MusicListProps) {
	const visibleMusics = (musics ?? []).slice(0, 12);
	return (
		<BaseCard
			title={title}
			subtitle={subtitle ?? pageText(page, totalPages, total)}
			accentColor={theme.colors.accent}
		>
			<div
				style={{
					display: "flex",
					flexDirection: "column",
					gap: theme.spacing.sm,
				}}
			>
				<div
					style={{
						display: "flex",
						color: theme.colors.textMuted,
						fontSize: theme.fontSize.xs,
						fontWeight: 800,
					}}
				>
					{pageText(page, totalPages, total)}
				</div>
				{visibleMusics.map((music) => {
					const jacket = music.assetbundleName
						? getMusicJacketUrl(music.assetbundleName, assetSource)
						: undefined;
					return (
						<div
							key={music.id}
							style={{
								display: "flex",
								gap: theme.spacing.md,
								padding: theme.spacing.md,
								borderRadius: theme.borderRadius.lg,
								backgroundColor: theme.colors.surface,
								border: `1px solid ${theme.colors.border}`,
							}}
						>
							<img
								src={jacket ?? placeholder("MUSIC")}
								width={86}
								height={86}
								style={{
									objectFit: "cover",
									borderRadius: theme.borderRadius.md,
								}}
							/>
							<div
								style={{
									display: "flex",
									flexDirection: "column",
									flex: 1,
									gap: 5,
								}}
							>
								<span
									style={{
										display: "flex",
										color: theme.colors.text,
										fontSize: theme.fontSize.md,
										fontWeight: 900,
									}}
								>
									#{music.id} {music.title}
								</span>
								<span
									style={{
										display: "flex",
										color: theme.colors.textMuted,
										fontSize: theme.fontSize.xs,
									}}
								>
									{music.pronunciation ??
										[music.composer, music.lyricist, music.arranger]
											.filter(Boolean)
											.join(" / ")}
								</span>
								<div style={{ display: "flex", flexWrap: "wrap", gap: 5 }}>
									{(music.difficulties ?? []).slice(0, 6).map((d) => (
										<DiffChip
											key={d.musicDifficulty ?? d.difficulty}
											diff={d.musicDifficulty ?? d.difficulty ?? "?"}
											level={d.playLevel ?? d.level ?? 0}
										/>
									))}
								</div>
							</div>
						</div>
					);
				})}
			</div>
		</BaseCard>
	);
}

function DiffChip({ diff, level }: { diff: string; level: number }) {
	return (
		<span
			style={{
				display: "flex",
				padding: "3px 8px",
				borderRadius: theme.borderRadius.round,
				backgroundColor: theme.colors.accentSoft,
				color: theme.colors.textSecondary,
				fontSize: 11,
				fontWeight: 900,
			}}
		>
			{diff.toUpperCase()} {level}
		</span>
	);
}

function pageText(page?: number, totalPages?: number, total?: number) {
	return `第 ${page ?? 1}/${totalPages ?? 1} 页 · 共 ${total ?? 0} 条`;
}

function placeholder(label: string) {
	const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="172" height="172"><rect width="172" height="172" rx="24" fill="#eef6ff"/><text x="50%" y="52%" dominant-baseline="middle" text-anchor="middle" font-family="Arial" font-size="28" font-weight="900" fill="#38bdf8">${label}</text></svg>`;
	return `data:image/svg+xml;base64,${Buffer.from(svg).toString("base64")}`;
}

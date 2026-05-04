import { getMusicJacketUrl, type AssetSourceType } from "../../shared";
import { BaseCard } from "./base";
import { theme } from "../styles/theme";

const DIFFICULTY_NAMES: Record<string, string> = {
	easy: "EASY",
	normal: "NORMAL",
	hard: "HARD",
	expert: "EXPERT",
	master: "MASTER",
	append: "APPEND",
};

const DIFFICULTY_COLORS: Record<string, string> = {
	easy: "#5AC06E",
	normal: "#56A4D4",
	hard: "#EFAF28",
	expert: "#E84D53",
	master: "#BB58B8",
	append: "#EE92BC",
};

const DIFFICULTY_ORDER = [
	"easy",
	"normal",
	"hard",
	"expert",
	"master",
	"append",
];

export interface ChartDetailProps {
	music: {
		id: number;
		title: string;
		pronunciation?: string;
		assetbundleName?: string;
		jacketUrl?: string;
		assetSource?: AssetSourceType | string;
		selectedDifficulty?: string;
		chartUrl?: string;
		durationSec?: number;
		secForMusicScoreMaker?: number;
		difficulties?: Array<{
			difficulty?: string;
			musicDifficulty?: string;
			level?: number;
			playLevel?: number;
			noteCount?: number;
			totalNoteCount?: number;
		}>;
	};
}

export function ChartDetail({ music }: ChartDetailProps) {
	const source = music.assetSource ?? "main-jp";
	const jacketUrl =
		music.jacketUrl ??
		(music.assetbundleName
			? getMusicJacketUrl(music.assetbundleName, source)
			: undefined);
	const difficulties = normalizeDifficulties(music.difficulties ?? []);
	const durationSec = music.durationSec ?? music.secForMusicScoreMaker;
	const selectedDifficulty = music.selectedDifficulty;
	const selectedChart = selectedDifficulty
		? difficulties.find((d) => d.difficulty === selectedDifficulty)
		: undefined;
	const maxNotes = Math.max(1, ...difficulties.map((d) => d.noteCount ?? 0));

	return (
		<BaseCard
			title={`${music.title} 谱面`}
			subtitle={
				music.pronunciation
					? `${music.pronunciation} · ID: ${music.id}`
					: `ID: ${music.id}`
			}
			accentColor={theme.colors.accent}
		>
			<div
				style={{
					display: "flex",
					flexDirection: "column",
					gap: theme.spacing.lg,
				}}
			>
				<div
					style={{
						display: "flex",
						gap: theme.spacing.lg,
						alignItems: "stretch",
					}}
				>
					<div
						style={{
							display: "flex",
							width: 196,
							height: 196,
							borderRadius: theme.borderRadius.xl,
							overflow: "hidden",
							flexShrink: 0,
							backgroundColor: theme.colors.surface,
							border: `1px solid ${theme.colors.border}`,
						}}
					>
						<img
							src={
								jacketUrl ??
								placeholderImage("CHART", theme.colors.accent, 392, 392)
							}
							width={196}
							height={196}
							style={{ objectFit: "cover" }}
						/>
					</div>

					<div
						style={{
							display: "flex",
							flexDirection: "column",
							gap: theme.spacing.md,
							flex: 1,
						}}
					>
						<MetricGrid
							values={[
								{ label: "谱面数", value: String(difficulties.length) },
								{
									label: "最高等级",
									value: difficulties.length
										? `Lv.${Math.max(...difficulties.map((d) => d.level))}`
										: "-",
								},
								{
									label: "最多 Notes",
									value: difficulties.length
										? Math.max(
												...difficulties.map((d) => d.noteCount ?? 0),
											).toLocaleString()
										: "-",
								},
								{
									label: "歌曲长度",
									value:
										typeof durationSec === "number"
											? formatDuration(durationSec)
											: "-",
								},
							]}
						/>
					</div>
				</div>

				{selectedChart && music.chartUrl ? (
					<div
						style={{
							display: "flex",
							flexDirection: "column",
							gap: theme.spacing.sm,
							padding: theme.spacing.md,
							borderRadius: theme.borderRadius.lg,
							backgroundColor: theme.colors.surface,
							border: `1px solid ${theme.colors.border}`,
						}}
					>
						<div
							style={{
								display: "flex",
								justifyContent: "space-between",
								color: theme.colors.textSecondary,
								fontSize: theme.fontSize.sm,
								fontWeight: 900,
							}}
						>
							<span>
								{DIFFICULTY_NAMES[selectedChart.difficulty] ??
									selectedChart.difficulty.toUpperCase()}{" "}
								Lv.{selectedChart.level}
							</span>
							<span>
								{(selectedChart.noteCount ?? 0).toLocaleString()} notes
							</span>
						</div>
						<img
							src={music.chartUrl}
							width={700}
							style={{
								objectFit: "contain",
								borderRadius: theme.borderRadius.md,
								backgroundColor: "#ffffff",
							}}
						/>
					</div>
				) : null}

				<div
					style={{
						display: "flex",
						flexDirection: "column",
						gap: theme.spacing.sm,
					}}
				>
					{difficulties.map((d) => (
						<ChartDifficultyRow
							key={d.difficulty}
							difficulty={d.difficulty}
							level={d.level}
							noteCount={d.noteCount}
							maxNotes={maxNotes}
							selected={d.difficulty === selectedDifficulty}
						/>
					))}
				</div>
			</div>
		</BaseCard>
	);
}

function MetricGrid({
	values,
}: {
	values: Array<{ label: string; value: string }>;
}) {
	return (
		<div style={{ display: "flex", flexWrap: "wrap", gap: theme.spacing.sm }}>
			{values.map((item) => (
				<div
					key={item.label}
					style={{
						display: "flex",
						flexDirection: "column",
						width: 132,
						padding: theme.spacing.md,
						borderRadius: theme.borderRadius.lg,
						backgroundColor: theme.colors.surface,
						border: `1px solid ${theme.colors.border}`,
					}}
				>
					<span
						style={{
							display: "flex",
							color: theme.colors.textMuted,
							fontSize: theme.fontSize.xs,
						}}
					>
						{item.label}
					</span>
					<span
						style={{
							display: "flex",
							color: theme.colors.text,
							fontSize: theme.fontSize.lg,
							fontWeight: 900,
							marginTop: 4,
						}}
					>
						{item.value}
					</span>
				</div>
			))}
		</div>
	);
}

function ChartDifficultyRow({
	difficulty,
	level,
	noteCount,
	maxNotes,
	selected,
}: {
	difficulty: string;
	level: number;
	noteCount?: number;
	maxNotes: number;
	selected?: boolean;
}) {
	const color = DIFFICULTY_COLORS[difficulty] ?? theme.colors.textMuted;
	const notes = noteCount ?? 0;
	const width = Math.max(8, Math.round((notes / maxNotes) * 100));
	return (
		<div
			style={{
				display: "flex",
				flexDirection: "column",
				gap: theme.spacing.xs,
				padding: theme.spacing.md,
				borderRadius: theme.borderRadius.lg,
				backgroundColor: selected ? `${color}18` : theme.colors.surface,
				border: selected
					? `2px solid ${color}`
					: `1px solid ${theme.colors.border}`,
			}}
		>
			<div
				style={{
					display: "flex",
					justifyContent: "space-between",
					alignItems: "baseline",
				}}
			>
				<div
					style={{
						display: "flex",
						alignItems: "baseline",
						gap: theme.spacing.sm,
					}}
				>
					<span
						style={{
							display: "flex",
							width: 96,
							color,
							fontSize: theme.fontSize.md,
							fontWeight: 900,
						}}
					>
						{DIFFICULTY_NAMES[difficulty] ?? difficulty.toUpperCase()}
					</span>
					<span
						style={{
							display: "flex",
							color,
							fontSize: theme.fontSize.xl,
							fontWeight: 900,
						}}
					>
						Lv.{level}
					</span>
				</div>
				<span
					style={{
						display: "flex",
						color: theme.colors.textSecondary,
						fontSize: theme.fontSize.sm,
						fontWeight: 800,
					}}
				>
					{notes.toLocaleString()} notes
				</span>
			</div>
			<div
				style={{
					display: "flex",
					height: 10,
					borderRadius: theme.borderRadius.round,
					backgroundColor: theme.colors.surfaceLight,
					overflow: "hidden",
				}}
			>
				<div
					style={{
						display: "flex",
						width: `${width}%`,
						backgroundColor: color,
					}}
				/>
			</div>
		</div>
	);
}

function normalizeDifficulties(
	input: NonNullable<ChartDetailProps["music"]["difficulties"]>,
) {
	return input
		.map((d) => ({
			difficulty: d.musicDifficulty ?? d.difficulty ?? "unknown",
			level: d.playLevel ?? d.level ?? 0,
			noteCount: d.totalNoteCount ?? d.noteCount,
		}))
		.filter((d) => d.level > 0)
		.sort(
			(a, b) =>
				DIFFICULTY_ORDER.indexOf(a.difficulty) -
				DIFFICULTY_ORDER.indexOf(b.difficulty),
		);
}

function formatDuration(seconds: number): string {
	const total = Math.max(0, Math.round(seconds));
	const minutes = Math.floor(total / 60);
	const rest = total % 60;
	return `${minutes}:${String(rest).padStart(2, "0")}`;
}

function placeholderImage(
	label: string,
	color: string,
	width: number,
	height: number,
): string {
	const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}"><rect width="${width}" height="${height}" rx="42" fill="#f8fafc"/><text x="50%" y="52%" dominant-baseline="middle" text-anchor="middle" font-family="Arial,sans-serif" font-size="48" font-weight="900" fill="${color}">${label}</text></svg>`;
	return `data:image/svg+xml;base64,${Buffer.from(svg, "utf8").toString("base64")}`;
}

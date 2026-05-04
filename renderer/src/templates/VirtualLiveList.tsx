import { getVirtualLiveBannerUrl, type AssetSourceType } from "../../shared";
import { BaseCard } from "./base";
import { theme } from "../styles/theme";

export interface VirtualLiveListProps {
	title: string;
	subtitle?: string;
	virtualLives: Array<{
		id: number;
		name: string;
		assetbundleName?: string;
		virtualLiveType?: string;
		startAt?: number;
		endAt?: number;
		currentStartAt?: number;
		currentEndAt?: number;
		living?: boolean;
		restCount?: number;
		schedules?: Array<{ startAt?: number; endAt?: number }>;
		rewards?: Array<{ resourceBoxId?: number; virtualLiveType?: string }>;
		characters?: Array<{ characterName?: string; performanceType?: string }>;
	}>;
	page?: number;
	totalPages?: number;
	total?: number;
	assetSource?: AssetSourceType | string;
}

export function VirtualLiveList({
	title,
	subtitle,
	virtualLives,
	page,
	totalPages,
	total,
	assetSource = "main-jp",
}: VirtualLiveListProps) {
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
				{virtualLives.map((live) => {
					const banner = live.assetbundleName
						? getVirtualLiveBannerUrl(live.assetbundleName, assetSource)
						: undefined;
					const nextStart =
						live.currentStartAt ||
						live.schedules?.find((s) => normalize(s.endAt) > Date.now())
							?.startAt ||
						live.startAt;
					const nextEnd =
						live.currentEndAt ||
						live.schedules?.find((s) => normalize(s.endAt) > Date.now())
							?.endAt ||
						live.endAt;
					return (
						<div
							key={live.id}
							style={{
								display: "flex",
								gap: theme.spacing.md,
								padding: theme.spacing.md,
								borderRadius: theme.borderRadius.lg,
								backgroundColor: live.living ? "#e8fff9" : theme.colors.surface,
								border: `1px solid ${theme.colors.border}`,
							}}
						>
							<img
								src={banner ?? placeholder("VIRTUAL LIVE")}
								width={248}
								height={78}
								style={{
									objectFit: "cover",
									borderRadius: theme.borderRadius.md,
									flexShrink: 0,
								}}
							/>
							<div
								style={{
									display: "flex",
									flexDirection: "column",
									flex: 1,
									gap: 6,
								}}
							>
								<div
									style={{
										display: "flex",
										justifyContent: "space-between",
										gap: theme.spacing.sm,
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
										#{live.id} {live.name}
									</span>
									{live.living ? (
										<span style={badgeStyle("#10b981")}>LIVE</span>
									) : (
										<span style={badgeStyle("#38bdf8")}>
											{live.restCount ?? 0} 场
										</span>
									)}
								</div>
								<span
									style={{
										display: "flex",
										color: theme.colors.textMuted,
										fontSize: theme.fontSize.xs,
									}}
								>
									{live.virtualLiveType ?? "virtual_live"} ·{" "}
									{formatDateTime(nextStart)} - {formatDateTime(nextEnd)}
								</span>
								<span
									style={{
										display: "flex",
										color: theme.colors.textSecondary,
										fontSize: 11,
									}}
								>
									出演{" "}
									{(live.characters ?? [])
										.slice(0, 8)
										.map((c) => c.characterName)
										.filter(Boolean)
										.join("、") || "-"}
								</span>
								<span
									style={{
										display: "flex",
										color: theme.colors.textSecondary,
										fontSize: 11,
									}}
								>
									奖励{" "}
									{(live.rewards ?? [])
										.map((r) =>
											r.resourceBoxId
												? `Box#${r.resourceBoxId}`
												: r.virtualLiveType,
										)
										.filter(Boolean)
										.join("、") || "-"}
								</span>
							</div>
						</div>
					);
				})}
			</div>
		</BaseCard>
	);
}

function badgeStyle(color: string) {
	return {
		display: "flex",
		alignItems: "center",
		padding: "3px 9px",
		borderRadius: theme.borderRadius.round,
		backgroundColor: `${color}22`,
		color,
		fontSize: 11,
		fontWeight: 900,
	};
}

function pageText(page?: number, totalPages?: number, total?: number) {
	return `第 ${page ?? 1}/${totalPages ?? 1} 页 · 共 ${total ?? 0} 条`;
}

function normalize(value?: number) {
	if (!value) return 0;
	return value < 1_000_000_000_000 ? value * 1000 : value;
}

function formatDateTime(value?: number) {
	const ms = normalize(value);
	if (!ms) return "-";
	const d = new Date(ms);
	return `${d.toLocaleDateString("zh-CN")} ${d.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" })}`;
}

function placeholder(label: string) {
	const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="360" height="144"><rect width="360" height="144" rx="22" fill="#ecfeff"/><text x="50%" y="52%" dominant-baseline="middle" text-anchor="middle" font-family="Arial" font-size="30" font-weight="900" fill="#06b6d4">${label}</text></svg>`;
	return `data:image/svg+xml;base64,${Buffer.from(svg).toString("base64")}`;
}

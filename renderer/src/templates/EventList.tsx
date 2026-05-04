import { getEventBannerUrl, type AssetSourceType } from "../../shared";
import { BaseCard } from "./base";
import { theme } from "../styles/theme";

export interface EventListProps {
	title: string;
	subtitle?: string;
	events: Array<{
		id: number;
		name: string;
		eventType?: string;
		unit?: string;
		assetbundleName?: string;
		bannerUrl?: string;
		storyBannerUrl?: string;
		startAt?: number;
		aggregateAt?: number;
		closedAt?: number;
		bonusAttr?: string;
		bonusCharacters?: string[];
	}>;
	page?: number;
	totalPages?: number;
	total?: number;
	assetSource?: AssetSourceType | string;
}

export function EventList({
	title,
	subtitle,
	events,
	page,
	totalPages,
	total,
	assetSource = "main-jp",
}: EventListProps) {
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
				{events.map((event) => {
					const banner =
						event.storyBannerUrl ??
						event.bannerUrl ??
						defaultStoryBannerUrl(assetSource) ??
						(event.assetbundleName
							? getEventBannerUrl(event.assetbundleName, assetSource)
							: undefined);
					return (
						<div
							key={event.id}
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
								src={banner ?? placeholder("EVENT")}
								width={150}
								height={52}
								style={{
									objectFit: "cover",
									borderRadius: theme.borderRadius.md,
								}}
							/>
							<div
								style={{
									display: "flex",
									flexDirection: "column",
									gap: 5,
									flex: 1,
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
									#{event.id} {event.name}
								</span>
								<span
									style={{
										display: "flex",
										color: theme.colors.textMuted,
										fontSize: theme.fontSize.xs,
									}}
								>
									{event.eventType ?? "event"} · {event.unit ?? "none"} ·{" "}
									{formatDate(event.startAt)} - {formatDate(event.closedAt)}
								</span>
								<span
									style={{
										display: "flex",
										color: theme.colors.textSecondary,
										fontSize: 11,
									}}
								>
									{event.bonusAttr ? `属性 ${event.bonusAttr}` : ""}{" "}
									{(event.bonusCharacters ?? []).slice(0, 5).join("、")}
								</span>
							</div>
						</div>
					);
				})}
			</div>
		</BaseCard>
	);
}

function defaultStoryBannerUrl(source: AssetSourceType | string) {
	const base =
		source.startsWith("http://") || source.startsWith("https://")
			? source.replace(/\/$/, "")
			: source === "main-jp"
				? "https://storage.exmeaning.com/sekai-jp-assets"
				: undefined;
	return base
		? `${base}/event_story/event_show_2026/screen_image/banner_event_story.png`
		: "https://storage.exmeaning.com/sekai-jp-assets/event_story/event_show_2026/screen_image/banner_event_story.png";
}

function pageText(page?: number, totalPages?: number, total?: number) {
	return `第 ${page ?? 1}/${totalPages ?? 1} 页 · 共 ${total ?? 0} 条`;
}

function formatDate(value?: number) {
	if (!value) return "-";
	const ms = value < 1_000_000_000_000 ? value * 1000 : value;
	return new Date(ms).toLocaleDateString("zh-CN");
}

function placeholder(label: string) {
	const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="300" height="104"><rect width="300" height="104" rx="18" fill="#fff1f7"/><text x="50%" y="52%" dominant-baseline="middle" text-anchor="middle" font-family="Arial" font-size="32" font-weight="900" fill="#f472b6">${label}</text></svg>`;
	return `data:image/svg+xml;base64,${Buffer.from(svg).toString("base64")}`;
}

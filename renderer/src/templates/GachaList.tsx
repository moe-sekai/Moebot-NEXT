import { getGachaLogoUrl, type AssetSourceType } from "../../shared";
import { BaseCard } from "./base";
import { theme } from "../styles/theme";

export interface GachaListProps {
	title: string;
	subtitle?: string;
	gachas: Array<{
		id: number;
		name: string;
		gachaType?: string;
		assetbundleName?: string;
		startAt?: number;
		endAt?: number;
		pickupCards?: Array<{ id: number; characterName?: string }>;
	}>;
	page?: number;
	totalPages?: number;
	total?: number;
	assetSource?: AssetSourceType | string;
}

export function GachaList({
	title,
	subtitle,
	gachas,
	page,
	totalPages,
	total,
	assetSource = "main-jp",
}: GachaListProps) {
	return (
		<BaseCard
			title={title}
			subtitle={subtitle ?? pageText(page, totalPages, total)}
			accentColor={theme.colors.warning}
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
				<div
					style={{ display: "flex", flexWrap: "wrap", gap: theme.spacing.sm }}
				>
					{gachas.map((gacha) => {
						const logo = gacha.assetbundleName
							? getGachaLogoUrl(gacha.assetbundleName, assetSource)
							: undefined;
						return (
							<div
								key={gacha.id}
								style={{
									display: "flex",
									flexDirection: "column",
									width: 346,
									gap: 7,
									padding: theme.spacing.md,
									borderRadius: theme.borderRadius.lg,
									backgroundColor: isCurrent(gacha)
										? "#fff7db"
										: theme.colors.surface,
									border: `1px solid ${theme.colors.border}`,
								}}
							>
								<div
									style={{
										display: "flex",
										height: 72,
										alignItems: "center",
										justifyContent: "center",
									}}
								>
									<img
										src={logo ?? placeholder("GACHA")}
										width={220}
										height={72}
										style={{ objectFit: "contain" }}
									/>
								</div>
								<span
									style={{
										display: "flex",
										color: theme.colors.text,
										fontSize: theme.fontSize.sm,
										fontWeight: 900,
									}}
								>
									#{gacha.id} {gacha.name}
								</span>
								<span
									style={{
										display: "flex",
										color: theme.colors.textMuted,
										fontSize: 11,
									}}
								>
									{gacha.gachaType ?? "gacha"} · {formatDate(gacha.startAt)} -{" "}
									{formatDate(gacha.endAt)}
								</span>
								<span
									style={{
										display: "flex",
										color: theme.colors.textSecondary,
										fontSize: 11,
									}}
								>
									Pickup{" "}
									{(gacha.pickupCards ?? [])
										.slice(0, 3)
										.map((c) => c.characterName ?? `#${c.id}`)
										.join("、") || "-"}
								</span>
							</div>
						);
					})}
				</div>
			</div>
		</BaseCard>
	);
}

function pageText(page?: number, totalPages?: number, total?: number) {
	return `第 ${page ?? 1}/${totalPages ?? 1} 页 · 共 ${total ?? 0} 条`;
}

function isCurrent(gacha: { startAt?: number; endAt?: number }) {
	const now = Date.now();
	const start = normalize(gacha.startAt);
	const end = normalize(gacha.endAt);
	return Boolean(start && end && start <= now && now <= end);
}

function normalize(value?: number) {
	if (!value) return 0;
	return value < 1_000_000_000_000 ? value * 1000 : value;
}

function formatDate(value?: number) {
	const ms = normalize(value);
	return ms ? new Date(ms).toLocaleDateString("zh-CN") : "-";
}

function placeholder(label: string) {
	const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="440" height="144"><rect width="440" height="144" rx="22" fill="#fff7db"/><text x="50%" y="52%" dominant-baseline="middle" text-anchor="middle" font-family="Arial" font-size="40" font-weight="900" fill="#f59e0b">${label}</text></svg>`;
	return `data:image/svg+xml;base64,${Buffer.from(svg).toString("base64")}`;
}

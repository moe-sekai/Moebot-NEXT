import { BaseCard } from "./base";
import { theme } from "../styles/theme";

export interface GalleryListItem {
	name: string;
	mode: string; // edit / view / off
	picCount: number;
	aliases?: string[];
	coverDataUri?: string; // data:image/jpeg;base64,... 由 Go 端把封面缩略图字节嵌入；缺省则显示占位
}

export interface GalleryListProps {
	title?: string;
	subtitle?: string;
	galleries: GalleryListItem[];
}

const COLS = 4;
const CARD_W = 220;
const COVER_H = 140;
const GAP = 12;

export function GalleryList({
	title = "画廊总览",
	subtitle,
	galleries,
}: GalleryListProps) {
	const sub = subtitle ?? `共 ${galleries.length} 个画廊`;
	const totalW = COLS * CARD_W + (COLS + 1) * GAP + 32;

	return (
		<BaseCard
			title={title}
			subtitle={sub}
			width={totalW}
			accentColor={theme.colors.accent}
		>
			<div
				style={{
					display: "flex",
					flexWrap: "wrap",
					gap: GAP,
				}}
			>
				{galleries.map((g) => (
					<div
						key={g.name}
						style={{
							display: "flex",
							flexDirection: "column",
							width: CARD_W,
							borderRadius: 12,
							overflow: "hidden",
							border: `1px solid ${theme.colors.border}`,
							backgroundColor: theme.colors.surface,
						}}
					>
						{/* 封面 */}
						<div
							style={{
								display: "flex",
								width: CARD_W,
								height: COVER_H,
								alignItems: "center",
								justifyContent: "center",
								backgroundColor: "#0f1115",
							}}
						>
							{g.coverDataUri ? (
								<img
									src={g.coverDataUri}
									width={CARD_W}
									height={COVER_H}
									style={{ objectFit: "cover" }}
								/>
							) : (
								<div
									style={{
										display: "flex",
										color: "#9ca3af",
										fontSize: 14,
										fontWeight: 700,
									}}
								>
									无封面
								</div>
							)}
						</div>

						{/* 信息条 */}
						<div
							style={{
								display: "flex",
								flexDirection: "column",
								padding: "10px 12px",
								gap: 4,
							}}
						>
							<div
								style={{
									display: "flex",
									alignItems: "center",
									justifyContent: "space-between",
									gap: 8,
								}}
							>
								<span
									style={{
										display: "flex",
										color: theme.colors.text,
										fontSize: 16,
										fontWeight: 800,
										overflow: "hidden",
										textOverflow: "ellipsis",
									}}
								>
									{g.name}
								</span>
								<span
									style={{
										display: "flex",
										alignItems: "center",
										justifyContent: "center",
										padding: "1px 8px",
										fontSize: 11,
										fontWeight: 800,
										borderRadius: 6,
										...modeStyle(g.mode),
									}}
								>
									{g.mode}
								</span>
							</div>
							<div
								style={{
									display: "flex",
									color: theme.colors.textSecondary,
									fontSize: 12,
								}}
							>
								{g.picCount} 张
							</div>
							{g.aliases && g.aliases.length > 0 && (
								<div
									style={{
										display: "flex",
										color: theme.colors.textMuted,
										fontSize: 11,
										overflow: "hidden",
										textOverflow: "ellipsis",
									}}
								>
									别名：{g.aliases.join("、")}
								</div>
							)}
						</div>
					</div>
				))}
			</div>
		</BaseCard>
	);
}

function modeStyle(mode: string) {
	switch (mode) {
		case "edit":
			return { backgroundColor: "#dcfce7", color: "#166534" };
		case "view":
			return { backgroundColor: "#dbeafe", color: "#1e40af" };
		case "off":
			return { backgroundColor: "#f3f4f6", color: "#6b7280" };
		default:
			return { backgroundColor: "#f3f4f6", color: "#6b7280" };
	}
}

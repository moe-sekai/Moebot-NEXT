import { BaseCard } from "./base";
import { theme } from "../styles/theme";

export interface GalleryGridPic {
	pid: number;
	dataUri: string; // data:image/jpeg;base64,... 由 Go 端把缩略图字节嵌入
}

export interface GalleryGridProps {
	title: string;
	subtitle?: string;
	pics: GalleryGridPic[];
	page?: number;
	totalPages?: number;
	total?: number;
}

const COLS = 10;
const CELL = 96;
const GAP = 8;

export function GalleryGrid({
	title,
	subtitle,
	pics,
	page,
	totalPages,
	total,
}: GalleryGridProps) {
	const sub =
		subtitle ??
		`第 ${page ?? 1}/${totalPages ?? 1} 页 · 共 ${total ?? pics.length} 张`;

	return (
		<BaseCard
			title={title}
			subtitle={sub}
			width={COLS * CELL + (COLS + 1) * GAP + 32}
			accentColor={theme.colors.accent}
			footer="Moebot NEXT · Gallery"
		>
			<div
				style={{
					display: "flex",
					flexWrap: "wrap",
					gap: GAP,
					justifyContent: "flex-start",
				}}
			>
				{pics.map((p) => (
					<div
						key={p.pid}
						style={{
							display: "flex",
							flexDirection: "column",
							width: CELL,
							border: `1px solid ${theme.colors.border}`,
							borderRadius: 8,
							overflow: "hidden",
							backgroundColor: theme.colors.surface,
						}}
					>
						<div
							style={{
								display: "flex",
								width: CELL,
								height: CELL,
								alignItems: "center",
								justifyContent: "center",
								backgroundColor: "#0f1115",
							}}
						>
							<img
								src={p.dataUri}
								width={CELL}
								height={CELL}
								style={{ objectFit: "cover" }}
							/>
						</div>
						<div
							style={{
								display: "flex",
								justifyContent: "center",
								alignItems: "center",
								padding: "2px 0",
								color: theme.colors.text,
								fontSize: 12,
								fontWeight: 700,
							}}
						>
							#{p.pid}
						</div>
					</div>
				))}
			</div>
		</BaseCard>
	);
}

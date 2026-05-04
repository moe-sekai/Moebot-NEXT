import { BaseCard } from "./base";
import { theme } from "../styles/theme";
import { SekaiCardThumbnail, canUseTrainedArt } from "./SekaiCardThumbnail";
import { getCardThumbnailUrl, type AssetSourceType } from "../../shared";

export interface CardListProps {
	title: string;
	subtitle?: string;
	cards: Array<{
		id: number;
		prefix?: string;
		characterName?: string;
		rarity?: string;
		cardRarityType?: string;
		attr?: string;
		assetbundleName?: string;
		thumbnailUrl?: string;
		normalThumbnailUrl?: string;
		trainedThumbnailUrl?: string;
	}>;
	page?: number;
	totalPages?: number;
	total?: number;
	assetSource?: AssetSourceType | string;
}

export function CardList({
	title,
	subtitle,
	cards,
	page,
	totalPages,
	total,
	assetSource = "main-jp",
}: CardListProps) {
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
					gap: theme.spacing.md,
				}}
			>
				<div
					style={{
						display: "flex",
						justifyContent: "space-between",
						color: theme.colors.textMuted,
						fontSize: theme.fontSize.xs,
						fontWeight: 800,
					}}
				>
					<span>{pageText(page, totalPages, total)}</span>
					<span>{cards.length} shown</span>
				</div>
				<div
					style={{ display: "flex", flexWrap: "wrap", gap: theme.spacing.sm }}
				>
					{cards.map((card) => {
						const rarity =
							card.cardRarityType ?? card.rarity ?? "rarity_unknown";
						const trained = canUseTrainedArt(rarity);
						const imageUrl =
							card.thumbnailUrl ??
							card.normalThumbnailUrl ??
							(card.assetbundleName
								? getCardThumbnailUrl(
										card.assetbundleName,
										false,
										assetSource,
										"png",
									)
								: undefined);
						const trainedUrl =
							card.trainedThumbnailUrl ??
							(card.assetbundleName
								? getCardThumbnailUrl(
										card.assetbundleName,
										true,
										assetSource,
										"png",
									)
								: undefined);
						return (
							<div
								key={card.id}
								style={{
									display: "flex",
									flexDirection: "column",
									width: 164,
									gap: 6,
									padding: 8,
									borderRadius: theme.borderRadius.lg,
									border: `1px solid ${theme.colors.border}`,
									backgroundColor: theme.colors.surface,
								}}
							>
								<div style={{ display: "flex", justifyContent: "center" }}>
									<SekaiCardThumbnail
										imageUrl={trained ? (trainedUrl ?? imageUrl) : imageUrl}
										rarity={rarity}
										attr={card.attr ?? "cute"}
										isTrained={trained}
										characterName={card.characterName}
										size={112}
									/>
								</div>
								<span
									style={{
										display: "flex",
										justifyContent: "center",
										color: theme.colors.text,
										fontSize: theme.fontSize.xs,
										fontWeight: 900,
										textAlign: "center",
									}}
								>
									#{card.id} {card.characterName ?? ""}
								</span>
								<span
									style={{
										display: "flex",
										justifyContent: "center",
										color: theme.colors.textSecondary,
										fontSize: 11,
										textAlign: "center",
									}}
								>
									{card.prefix ?? rarity}
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

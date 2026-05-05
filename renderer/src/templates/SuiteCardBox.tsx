import { getCardThumbnailUrl, type AssetSourceType } from "../../shared";
import { theme } from "../styles/theme";
import { BaseCard } from "./base";
import { SekaiCardThumbnail, canUseTrainedArt } from "./SekaiCardThumbnail";

export interface SuiteCardBoxProps {
	title?: string;
	subtitle?: string;
	profile?: { name?: string; displayName?: string; rank?: number | string; userId?: number | string; uid?: number | string };
	groups?: SuiteCardGroup[];
	cards?: SuiteCard[];
	options?: {
		useBeforeTraining?: boolean;
		showId?: boolean;
		showCreatedAt?: boolean;
		groupByCharacter?: boolean;
	};
	assetSource?: AssetSourceType | string;
	total?: number;
	ownedTotal?: number;
}

interface SuiteCardGroup {
	title?: string;
	name?: string;
	characterName?: string;
	cards?: SuiteCard[];
}

interface SuiteCard {
	id?: number | string;
	cardId?: number | string;
	prefix?: string;
	characterName?: string;
	rarity?: string;
	cardRarityType?: string;
	attr?: string;
	assetbundleName?: string;
	thumbnailUrl?: string;
	trainedThumbnailUrl?: string;
	isTrained?: boolean;
	defaultImage?: string;
	mastery?: number;
	masterRank?: number;
	skillLevel?: number;
	level?: number;
	createdAt?: number | string;
	obtainedAt?: number | string;
	acquiredAt?: number | string;
	owned?: boolean;
	isOwned?: boolean;
	supplyType?: string;
	limitedType?: string;
	isLimited?: boolean;
	isBirthday?: boolean;
}

export function SuiteCardBox({
	title = "卡牌一览",
	subtitle,
	profile,
	groups,
	cards = [],
	options = {},
	assetSource = "main-jp",
	total,
	ownedTotal,
}: SuiteCardBoxProps) {
	const normalizedGroups = normalizeGroups(groups, cards, options.groupByCharacter);
	const allCards = normalizedGroups.flatMap((group) => group.cards ?? []);
	const owned = ownedTotal ?? allCards.filter((card) => isOwned(card)).length;
	const count = total ?? allCards.length;
	const profileText = profile ? `${profile.displayName ?? profile.name ?? "未知玩家"}${profile.rank !== undefined ? ` · Rank ${profile.rank}` : ""}` : undefined;
	const summary = [profileText, `持有 ${owned}/${count}`].filter(Boolean).join(" · ");

	return (
		<BaseCard title={title} subtitle={subtitle ?? summary} accentColor={theme.colors.accent}>
			<div style={{ display: "flex", flexDirection: "column", gap: theme.spacing.lg }}>
				<div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 800 }}>
					<span style={{ display: "flex" }}>{summary}</span>
					{profile?.userId || profile?.uid ? <span style={{ display: "flex" }}>UID: {profile.userId ?? profile.uid}</span> : <span style={{ display: "flex" }}>{allCards.length} shown</span>}
				</div>

				{normalizedGroups.map((group, index) => (
					<div key={`${group.title ?? group.name ?? index}`} style={{ display: "flex", flexDirection: "column", gap: theme.spacing.md }}>
						<div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", borderBottom: `1px solid ${theme.colors.border}`, paddingBottom: theme.spacing.sm }}>
							<span style={{ display: "flex", color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>{group.title ?? group.name ?? group.characterName ?? "未分组"}</span>
							<span style={{ display: "flex", color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 800 }}>{(group.cards ?? []).filter(isOwned).length}/{group.cards?.length ?? 0}</span>
						</div>
						<div style={{ display: "flex", flexWrap: "wrap", gap: theme.spacing.sm }}>
							{(group.cards ?? []).map((card, cardIndex) => <CardTile key={`${card.cardId ?? card.id ?? cardIndex}`} card={card} options={options} source={assetSource} />)}
						</div>
					</div>
				))}
			</div>
		</BaseCard>
	);
}

function CardTile({ card, options, source }: { card: SuiteCard; options: SuiteCardBoxProps["options"]; source: AssetSourceType | string }) {
	const owned = isOwned(card);
	const rarity = card.cardRarityType ?? card.rarity ?? (card.isBirthday ? "rarity_birthday" : "rarity_1");
	const useBefore = Boolean(options?.useBeforeTraining);
	const trained = !useBefore && (card.defaultImage === "special_training" || card.isTrained || canUseTrainedArt(rarity));
	const normalUrl = card.thumbnailUrl ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, source, "png") : undefined);
	const trainedUrl = card.trainedThumbnailUrl ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, true, source, "png") : undefined);
	const supplyType = badgeSupplyType(card);
	const obtained = card.createdAt ?? card.obtainedAt ?? card.acquiredAt;

	return (
		<div style={{ display: "flex", flexDirection: "column", width: 132, gap: 6, padding: 8, borderRadius: theme.borderRadius.lg, border: `1px solid ${owned ? theme.colors.border : theme.colors.borderStrong}`, backgroundColor: owned ? theme.colors.surface : theme.colors.surfaceLight }}>
			<div style={{ display: "flex", justifyContent: "center", position: "relative" }}>
				<SekaiCardThumbnail imageUrl={trained ? trainedUrl ?? normalUrl : normalUrl} rarity={rarity} attr={card.attr ?? "cute"} isTrained={trained} mastery={owned ? card.mastery ?? card.masterRank : undefined} characterName={card.characterName} supplyType={supplyType} size={112} />
				{!owned && (
					<div style={{ display: "flex", position: "absolute", left: 10, top: 0, width: 112, height: 112, borderRadius: 9, backgroundColor: "rgba(23, 32, 51, 0.52)", alignItems: "center", justifyContent: "center" }}>
						<span style={{ display: "flex", color: "#ffffff", fontSize: theme.fontSize.xs, fontWeight: 900 }}>未持有</span>
					</div>
				)}
			</div>
			<div style={{ display: "flex", flexDirection: "column", gap: 3 }}>
				<span style={{ display: "flex", justifyContent: "center", color: theme.colors.text, fontSize: theme.fontSize.xs, fontWeight: 900, textAlign: "center" }}>
					{options?.showId ? `#${card.cardId ?? card.id ?? "-"} ` : ""}{card.characterName ?? "未知角色"}
				</span>
				{owned ? (
					<span style={{ display: "flex", justifyContent: "center", color: theme.colors.textSecondary, fontSize: 11, fontWeight: 800, textAlign: "center" }}>
						{[
							card.level !== undefined ? `Lv.${card.level}` : undefined,
							(card.mastery ?? card.masterRank) !== undefined ? `MR${card.mastery ?? card.masterRank}` : undefined,
							card.skillLevel !== undefined ? `SL${card.skillLevel}` : undefined,
						].filter(Boolean).join(" · ") || "已持有"}
					</span>
				) : (
					<span style={{ display: "flex", justifyContent: "center", color: theme.colors.textMuted, fontSize: 11, fontWeight: 800 }}>未解锁</span>
				)}
				{options?.showCreatedAt && obtained !== undefined && <span style={{ display: "flex", justifyContent: "center", color: theme.colors.textMuted, fontSize: 10, fontWeight: 700 }}>{String(obtained)}</span>}
			</div>
		</div>
	);
}

function normalizeGroups(groups?: SuiteCardGroup[], cards: SuiteCard[] = [], groupByCharacter?: boolean): SuiteCardGroup[] {
	if (groups && groups.length > 0) return groups.map((group) => ({ ...group, cards: group.cards ?? [] }));
	if (!groupByCharacter) return [{ title: "全部卡牌", cards }];
	const map = new Map<string, SuiteCard[]>();
	for (const card of cards) {
		const key = card.characterName ?? "未知角色";
		map.set(key, [...(map.get(key) ?? []), card]);
	}
	return Array.from(map.entries()).map(([title, groupCards]) => ({ title, cards: groupCards }));
}

function isOwned(card: SuiteCard): boolean {
	return card.owned ?? card.isOwned ?? card.level !== undefined;
}

function badgeSupplyType(card: SuiteCard): string | undefined {
	const raw = card.supplyType ?? card.limitedType;
	if (card.isBirthday || card.cardRarityType === "rarity_birthday" || card.rarity === "rarity_birthday") return "生日";
	if (card.isLimited) return raw ?? "期间限定";
	if (!raw || raw === "常驻" || raw === "normal") return undefined;
	return raw;
}

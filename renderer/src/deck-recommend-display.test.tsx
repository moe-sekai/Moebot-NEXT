import { describe, expect, test } from "bun:test";
import React from "react";
import { DeckRecommend } from "./templates/DeckRecommend";
import { SekaiCardThumbnail } from "./templates/SekaiCardThumbnail";

describe("deck recommend display", () => {
	test("world link deck displays main and support deck bonus separately", () => {
		const text = collectText(<DeckRecommend
			options={{ mode: "event" }}
			decks={[{
				rank: 1,
				value: 12345,
				valueLabel: "活动PT",
				score: 12345,
				eventPoint: 12345,
				eventBonus: 123,
				supportDeckBonus: 231,
				power: { total: 456789 },
				multiLiveScoreUp: 500,
				cards: [],
			}]}
		/>);

		expect(text).toContain("123%");
		expect(text).toContain("主 0%");
		expect(text).toContain("支援 231%");
	});

	test("deck result uses compact table headers", () => {
		const text = collectText(<DeckRecommend
			options={{ mode: "event", liveType: "multi", target: "score" }}
			decks={[{
				rank: 1,
				value: 12345,
				valueLabel: "PTV",
				eventPoint: 12345,
				eventBonus: 300,
				power: { total: 234567 },
				multiLiveScoreUp: 200,
				cards: [{ cardId: 1001, card: { characterName: "初音未来", rarity: "rarity_4", attr: "cool" } }],
			}]}
		/>);

		expect(text).toContain("PTV");
		expect(text).toContain("卡组");
		expect(text).toContain("加成");
		expect(text).toContain("实效");
		expect(text).toContain("综合力");
		expect(text).toContain("技能顺序");
	});

	test("preset default music shows real title plus '默认' tag", () => {
		const text = collectText(<DeckRecommend
			options={{ mode: "event", musicId: 74, isPresetDefault: true }}
			music={{ id: 74, title: "RealDefaultSong", isPresetDefault: true }}
			decks={[]}
		/>);
		expect(text).toContain("RealDefaultSong");
		expect(text).toContain("默认");
		expect(text).not.toContain("默认曲目");
	});

	test("explicitly chosen music does not show '默认' tag", () => {
		const text = collectText(<DeckRecommend
			options={{ mode: "event", musicId: 456 }}
			music={{ id: 456, title: "ChosenSong" }}
			decks={[]}
		/>);
		expect(text).toContain("ChosenSong");
		expect(text).not.toContain("默认");
	});

	test("card mini renders pre-cached 58px thumbnail with isTrained and event bonus", () => {
		const tree = (<DeckRecommend
			options={{ mode: "event" }}
			decks={[{
				rank: 1,
				value: 1,
				valueLabel: "PTV",
				score: 1,
				eventPoint: 1,
				eventBonus: 0,
				power: { total: 0 },
				multiLiveScoreUp: 0,
				cards: [
					{
						cardId: 1001,
						skillLevel: 4,
						skillScoreUp: 120,
						eventBonus: 55,
						defaultImage: "special_training",
						card: {
							cardId: 1001,
							rarity: "rarity_4",
							cardRarityType: "rarity_4",
							attr: "cool",
							thumbnailUrl: "https://example.com/normal.png",
							trainedThumbnailUrl: "https://example.com/trained.png",
							isTrained: true,
							compositeLayers: [
								{ type: "rect", width: 58, height: 58, rx: 4, fill: "#fff" },
							],
						},
					},
					{
						cardId: 1002,
						skillLevel: 1,
						skillScoreUp: 80,
						eventBonus: 15,
						defaultImage: "original",
						card: {
							cardId: 1002,
							rarity: "rarity_4",
							cardRarityType: "rarity_4",
							attr: "happy",
							thumbnailUrl: "https://example.com/n2.png",
							trainedThumbnailUrl: "https://example.com/t2.png",
							isTrained: false,
						},
					},
				],
			}]}
		/>);

		const text = collectText(tree);
		expect(text).toContain("+55%");
		expect(text).toContain("+15%");
		expect(text).toContain("SLv.4");

		const thumbs = collectThumbnails(tree);
		expect(thumbs.length).toBe(2);
		expect(thumbs[0]?.size).toBe(58);
		expect(thumbs[0]?.isTrained).toBe(true);
		expect(thumbs[0]?.imageUrl).toBe("https://example.com/trained.png");
		expect(thumbs[0]?.compositeLayers?.length).toBe(1);
		expect(thumbs[1]?.size).toBe(58);
		expect(thumbs[1]?.isTrained).toBe(false);
		expect(thumbs[1]?.imageUrl).toBe("https://example.com/n2.png");
	});
});

function collectText(node: any): string {
	if (node == null || typeof node === "boolean") return "";
	if (typeof node === "string" || typeof node === "number") return String(node);
	if (Array.isArray(node)) return node.map(collectText).join("");
	if (React.isValidElement(node)) {
		if (typeof node.type === "function") {
			return collectText((node.type as any)(node.props));
		}
		return collectText((node.props as any).children);
	}
	return "";
}

function collectThumbnails(node: any): any[] {
	if (node == null || typeof node === "boolean") return [];
	if (typeof node === "string" || typeof node === "number") return [];
	if (Array.isArray(node)) return node.flatMap(collectThumbnails);
	if (React.isValidElement(node)) {
		const matches: any[] = [];
		if (node.type === SekaiCardThumbnail) {
			matches.push(node.props);
			return matches;
		}
		if (typeof node.type === "function") {
			return collectThumbnails((node.type as any)(node.props));
		}
		return collectThumbnails((node.props as any).children);
	}
	return [];
}

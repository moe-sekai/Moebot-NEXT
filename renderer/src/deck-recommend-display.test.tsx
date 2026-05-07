import { describe, expect, test } from "bun:test";
import React from "react";
import { DeckRecommend } from "./templates/DeckRecommend";

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
		expect(text).toContain("0.0+231.0%");
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

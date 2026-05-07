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

		expect(text).toContain("加成：123% + 231%");
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

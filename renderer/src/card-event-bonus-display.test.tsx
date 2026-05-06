import { describe, expect, test } from "bun:test";
import React from "react";
import { CardDetail } from "./templates/CardDetail";
import { CardList } from "./templates/CardList";
import { EventInfo } from "./templates/EventInfo";
import { EventList } from "./templates/EventList";

const card = {
	id: 1001,
	prefix: "测试卡牌",
	characterName: "初音未来",
	rarity: "rarity_4",
	cardRarityType: "rarity_4",
	attr: "cute",
	events: [{ id: 2001, name: "测试活动", eventType: "marathon" }],
};

const event = {
	id: 2001,
	name: "测试活动",
	eventType: "marathon",
	startAt: 1700000000000,
	closedAt: 1700200000000,
	bonusCards: [card],
};

describe("card and event bonus display", () => {
	test("card detail displays related event", () => {
		const text = collectText(<CardDetail card={card} />);
		expect(text).toContain("关联活动");
		expect(text).toContain("测试活动");
	});

	test("card list displays related event summary", () => {
		const text = collectText(<CardList title="卡牌查询" cards={[card]} />);
		expect(text).toContain("活动：测试活动");
	});

	test("event detail displays bonus cards", () => {
		const text = collectText(<EventInfo event={event} />);
		expect(text).toContain("加成卡片");
		expect(text).toContain("测试卡牌");
	});

	test("event list displays bonus card summary", () => {
		const text = collectText(<EventList title="活动查询" events={[event]} />);
		expect(text).toContain("加成卡：初音未来");
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

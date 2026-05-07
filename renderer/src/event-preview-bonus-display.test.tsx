import { describe, expect, test } from "bun:test";
import React from "react";
import { createPreviewElementForTest } from "./preview";

describe("event preview bonus display", () => {
	test("event info preview data can show bonus cards", () => {
		const text = collectText(createPreviewElementForTest("event-info"));
		expect(text).toContain("加成卡片");
		expect(text).toContain("被星光照亮的舞台");
	});

	test("event list preview data can show bonus card summary", () => {
		const text = collectText(createPreviewElementForTest("event-list"));
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

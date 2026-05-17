import { rendererAssetCache } from "./asset-cache";
import { startCardThumbnailCompositePreload, statusForCardThumbnailComposites, type CardThumbnailCompositeRequest } from "./card-thumbnail-composites";
import {
	getMasterSnapshot,
	getMusicMetasSnapshot,
	listSnapshotStatus,
	setMasterSnapshot,
	setMusicMetasSnapshot,
} from "./deck-recommend/snapshot-store";
import { setDeployer, getDeployer } from "./deployer";
import { getCachedRender, setCachedRender, renderCacheStats, clearRenderCache, updateRenderCacheConfig } from "./render-cache";
import { BudgetRejectedError } from "./render-budget";
import { RenderWorkerPool } from "./render-worker-pool";
import { loadFonts, FONT_FAMILY, defaultFontFamily, scoreFontFamily, setFontPreferences, fontPreferences } from "./fonts";
import { preloadFixedChartNoteAssets } from "./svg-assets";
import { listRenderPreviews } from "./preview";
import { renderDataSummary, type ChartRenderRequest, type RenderRequest } from "./render-jobs";

interface CachePreloadRequest {
	urls?: string[];
	cards?: CardThumbnailCompositeRequest[];
	force?: boolean;
	concurrency?: number;
}

const port = Number(process.env.PORT ?? 13001);
const defaultPrecision = parsePositiveNumber(process.env.RENDER_PRECISION, 1.5);
let renderPrepareConcurrency = Math.max(
	1,
	Math.min(32, Math.floor(parsePositiveNumber(process.env.RENDER_PREPARE_CONCURRENCY, 4))),
);

const workerPool = new RenderWorkerPool({
	maxConcurrency: parsePositiveNumber(process.env.RENDER_MAX_CONCURRENCY, 2),
	queueLimit: Number.isFinite(Number(process.env.RENDER_QUEUE_LIMIT))
		? Math.max(0, Math.floor(Number(process.env.RENDER_QUEUE_LIMIT)))
		: 8,
	config: { prepareConcurrency: renderPrepareConcurrency },
});

function budgetRejectionResponse(error: unknown): Response | null {
	if (error instanceof BudgetRejectedError) {
		const stats = workerPool.stats();
		return Response.json(
			{
				error: true,
				message: error.message,
				budget: stats,
			},
			{ status: 503, headers: { "retry-after": "1" } },
		);
	}
	return null;
}

void preloadFixedChartNoteAssets().catch((error) => {
	console.warn("[renderer] fixed chart note asset preload failed:", error);
});

function parsePositiveNumber(value: unknown, fallback = 0): number {
	const numberValue = typeof value === "number" ? value : Number(value);
	return Number.isFinite(numberValue) && numberValue > 0
		? numberValue
		: fallback;
}

function mergePreloadStatuses(status: any, composite: any) {
	const total = (status.total ?? 0) + (composite.composite_total ?? 0);
	const cached = (status.cached ?? 0) + (composite.composite_cached ?? 0);
	const running = Boolean(status.running || composite.composite_running);
	const progress = running && total > 0
		? Math.min(0.999, ((status.progress ?? 0) * (status.total ?? 0) + (composite.composite_progress ?? 0) * (composite.composite_total ?? 0)) / total)
		: total === 0 ? 1 : cached / total;
	return {
		...status,
		...composite,
		running,
		progress,
		cached,
		missing: Math.max(0, total - cached),
		total,
		errors: [...(status.errors ?? []), ...(composite.composite_errors ?? [])].slice(0, 8),
	};
}

Bun.serve({
	port,
	async fetch(request) {
		const url = new URL(request.url);

		if (url.pathname === "/" || url.pathname === "/health") {
			return Response.json({
				status: "ok",
				service: "moebot-next-renderer",
				version: "0.1.0",
				endpoints: [
					"GET /health",
					"GET /fonts",
					"GET /previews",
					"GET /preview/:id",
					"POST /render",
					"POST /render/chart",
					"GET /budget",
					"PUT /budget",
					"POST /cache/card-thumbnails/preload",
					"POST /cache/card-thumbnails/status",
				],
				note: "这是内部 Satori 渲染服务；管理面板请访问 http://127.0.0.1:8080/",
			});
		}

		if (url.pathname === "/fonts" && request.method === "GET") {
			try {
				const fonts = await loadFonts();
				const fontList = fonts.map((f) => ({
					name: f.name,
					weight: f.weight,
					style: f.style,
				}));
				const uniqueFamilies = [...new Set(fonts.map((f) => f.name))];
				return Response.json({
					ok: true,
					fonts: fontList,
					families: uniqueFamilies,
					defaults: {
						body: defaultFontFamily,
						score: scoreFontFamily,
					},
					preferences: {
						body: fontPreferences.body,
						score: fontPreferences.score,
					},
					config: FONT_FAMILY,
					total: fonts.length,
				});
			} catch (error) {
				return Response.json(
					{
						ok: false,
						fonts: [],
						families: [],
						defaults: { body: defaultFontFamily, score: scoreFontFamily },
						preferences: { body: fontPreferences.body, score: fontPreferences.score },
						config: FONT_FAMILY,
						total: 0,
						message: error instanceof Error ? error.message : String(error),
					},
					{ status: 500 },
				);
			}
		}

		if (url.pathname === "/fonts" && request.method === "POST") {
			try {
				const body = (await request.json()) as { body?: string; score?: string };
				setFontPreferences(body.body, body.score);
				workerPool.updateConfig({ fontBody: fontPreferences.body, fontScore: fontPreferences.score });
				return Response.json({
					ok: true,
					defaults: {
						body: defaultFontFamily,
						score: scoreFontFamily,
					},
					preferences: {
						body: fontPreferences.body,
						score: fontPreferences.score,
					},
				});
			} catch (error) {
				return Response.json(
					{
						ok: false,
						message: error instanceof Error ? error.message : String(error),
					},
					{ status: 400 },
				);
			}
		}

		if (url.pathname === "/previews" && request.method === "GET") {
			return Response.json({
				data: listRenderPreviews(),
				total: listRenderPreviews().length,
			});
		}

		if (url.pathname.startsWith("/preview/") && request.method === "GET") {
			try {
				const id = decodeURIComponent(
					url.pathname.replace("/preview/", "").replace(/\/$/, ""),
				);
				const width = Number(url.searchParams.get("width") || 0);
				const height = Number(url.searchParams.get("height") || 0);
				const precision = parsePositiveNumber(
					url.searchParams.get("precision"),
					defaultPrecision,
				);
				const result = await workerPool.run({ type: "preview", id, width, height, precision }) as any;
				return new Response(new Uint8Array(result.trace.png), {
					headers: {
						"content-type": "image/png",
						"cache-control": "no-store",
						"x-render-total-ms": String(result.trace.timings.totalMs),
						"x-render-fonts-ms": String(result.trace.timings.fontsMs),
						"x-render-images-ms": String(result.trace.timings.imagesMs),
						"x-render-satori-ms": String(result.trace.timings.satoriMs),
						"x-render-resvg-ms": String(result.trace.timings.resvgMs),
						"x-render-size-bytes": String(result.trace.sizeBytes),
						"x-render-image-total": String(result.trace.imageCache.total),
						"x-render-image-remote": String(result.trace.imageCache.remote),
						"x-render-image-cache-hits": String(result.trace.imageCache.hits),
						"x-render-image-cache-misses": String(result.trace.imageCache.misses),
						"x-render-image-cache-errors": String(result.trace.imageCache.errors),
					},
				});
			} catch (error) {
				const rejected = budgetRejectionResponse(error);
				if (rejected) return rejected;
				console.error("[renderer] preview render failed:", error);
				return Response.json(
					{
						error: true,
						message: error instanceof Error ? error.message : String(error),
					},
					{ status: 500 },
				);
			}
		}

		if (url.pathname === "/cache/render/stats" && request.method === "GET") {
			return Response.json(renderCacheStats());
		}

		if (url.pathname === "/budget" && request.method === "GET") {
			return Response.json({ ok: true, ...workerPool.stats() });
		}

		if (url.pathname === "/budget" && request.method === "PUT") {
			try {
				const body = (await request.json()) as {
					maxConcurrency?: number;
					queueLimit?: number;
					prepareConcurrency?: number;
				};
				const applied = workerPool.configure({
					maxConcurrency:
						typeof body.maxConcurrency === "number" ? body.maxConcurrency : undefined,
					queueLimit:
						typeof body.queueLimit === "number" ? body.queueLimit : undefined,
					prepareConcurrency:
						typeof body.prepareConcurrency === "number" ? body.prepareConcurrency : undefined,
				});
				if (typeof applied.prepareConcurrency === "number" && applied.prepareConcurrency > 0) {
					renderPrepareConcurrency = applied.prepareConcurrency;
				}
				return Response.json({ ok: true, applied, ...workerPool.stats() });
			} catch (error) {
				return Response.json(
					{
						ok: false,
						error: true,
						message: error instanceof Error ? error.message : String(error),
					},
					{ status: 400 },
				);
			}
		}

		if (url.pathname === "/cache/render" && request.method === "DELETE") {
			clearRenderCache();
			return Response.json({ ok: true, ...renderCacheStats() });
		}

		if (url.pathname === "/cache/render/config" && request.method === "PUT") {
			try {
				const body = (await request.json()) as { maxBytes?: number; maxEntries?: number };
				const applied = updateRenderCacheConfig({
					maxBytes: typeof body.maxBytes === "number" ? body.maxBytes : undefined,
					maxEntries: typeof body.maxEntries === "number" ? body.maxEntries : undefined,
				});
				return Response.json({ ok: true, applied, ...renderCacheStats() });
			} catch (error) {
				return Response.json(
					{ ok: false, error: true, message: error instanceof Error ? error.message : String(error) },
					{ status: 400 },
				);
			}
		}

		if (url.pathname === "/cache/card-thumbnails/preload" && request.method === "POST") {
			try {
				const body = (await request.json()) as CachePreloadRequest;
				console.info(`[renderer] starting card thumbnail preload: urls=${body.urls?.length ?? 0}, composites=${body.cards?.length ?? 0}`);
				const status = await rendererAssetCache.startPreload(body.urls ?? [], {
					force: body.force,
					concurrency: body.concurrency,
				});
				const composite = await startCardThumbnailCompositePreload(body.cards ?? [], {
					force: body.force,
					concurrency: body.concurrency,
				});
				return Response.json(mergePreloadStatuses(status, composite));
			} catch (error) {
				console.error("[renderer] preload card thumbnails failed:", error);
				return Response.json(
					{
						error: true,
						message: error instanceof Error ? error.message : String(error),
					},
					{ status: 500 },
				);
			}
		}

		if (url.pathname === "/cache/card-thumbnails/status" && request.method === "POST") {
			try {
				const body = (await request.json()) as CachePreloadRequest;
				const [status, composite] = await Promise.all([
					rendererAssetCache.statusForUrls(body.urls ?? []),
					statusForCardThumbnailComposites(body.cards ?? []),
				]);
				return Response.json(mergePreloadStatuses(status, composite));
			} catch (error) {
				console.error("[renderer] card thumbnail cache status failed:", error);
				return Response.json(
					{
						error: true,
						message: error instanceof Error ? error.message : String(error),
					},
					{ status: 500 },
				);
			}
		}

		if (url.pathname === "/render/chart" && request.method === "POST") {
			try {
				const body = (await request.json()) as ChartRenderRequest;
				const trace = await workerPool.run({
					type: "chart",
					request: {
						url: body.url,
						svg: body.svg,
						width: body.width,
						precision: parsePositiveNumber(body.precision, defaultPrecision),
					},
				}) as any;
				return new Response(new Uint8Array(trace.png), {
					headers: {
						"content-type": "image/png",
						"cache-control": "no-store",
						"x-render-total-ms": String(trace.timings.totalMs),
						"x-render-resvg-ms": String(trace.timings.resvgMs),
						"x-render-size-bytes": String(trace.sizeBytes),
						"x-render-width": String(trace.width),
						"x-render-height": String(trace.height),
					},
				});
			} catch (error) {
				const rejected = budgetRejectionResponse(error);
				if (rejected) return rejected;
				console.error("[renderer] chart render failed:", error);
				return Response.json(
					{
						error: true,
						message: error instanceof Error ? error.message : String(error),
					},
					{ status: 500 },
				);
			}
		}

		// Push or update the per-region master/musicMetas snapshot used by the
		// deck recommender. Go calls this once on startup (and when the cached
		// version changes) so /deck-recommend/calculate can stop carrying the
		// 32MB master payload on every request.
		if (url.pathname === "/deck-recommend/snapshot" && request.method === "POST") {
			try {
				const body = await request.json() as {
					region?: string;
					master?: { version?: string; data?: Record<string, unknown[]> } | null;
					musicMetas?: { version?: string; data?: any[] } | null;
				};
				const region = String(body?.region ?? "jp").trim() || "jp";
				const out: Record<string, unknown> = { ok: true, region };
				const snapshotRequest: any = { region };
				if (body?.master && body.master.data) {
					const snap = setMasterSnapshot(region, body.master.data as any, String(body.master.version ?? Date.now()));
					out.master = { version: snap.version, keyCount: snap.keyCount, updatedAt: snap.updatedAt };
					snapshotRequest.master = { version: snap.version, data: body.master.data };
				}
				if (body?.musicMetas && body.musicMetas.data) {
					const snap = setMusicMetasSnapshot(region, body.musicMetas.data, String(body.musicMetas.version ?? Date.now()));
					out.musicMetas = { version: snap.version, count: snap.count, updatedAt: snap.updatedAt };
					snapshotRequest.musicMetas = { version: snap.version, data: body.musicMetas.data };
				}
				await workerPool.broadcastSnapshot(snapshotRequest);
				return Response.json(out);
			} catch (error) {
				console.error("[renderer] deck recommend snapshot upload failed:", error);
				return Response.json(
					{ ok: false, error: true, message: error instanceof Error ? error.message : String(error) },
					{ status: 500 },
				);
			}
		}

		if (url.pathname === "/deck-recommend/snapshot/status" && request.method === "GET") {
			const queryRegion = url.searchParams.get("region");
			if (queryRegion) {
				const region = queryRegion.trim().toLowerCase() || "jp";
				const master = getMasterSnapshot(region);
				const musicMetas = getMusicMetasSnapshot(region);
				return Response.json({
					ok: true,
					region,
					master: master ? { version: master.version, keyCount: master.keyCount, updatedAt: master.updatedAt } : null,
					musicMetas: musicMetas ? { version: musicMetas.version, count: musicMetas.count, updatedAt: musicMetas.updatedAt } : null,
				});
			}
			return Response.json({ ok: true, snapshots: listSnapshotStatus() });
		}

		if (url.pathname === "/deck-recommend/calculate" && request.method === "POST") {
			try {
				const body = await request.json();
				const result = await workerPool.run({ type: "deck-calculate", request: body as any }) as any;
				return Response.json(result, { status: result.ok ? 200 : 400 });
			} catch (error) {
				const rejected = budgetRejectionResponse(error);
				if (rejected) return rejected;
				console.error("[renderer] deck recommend calculation failed:", error);
				return Response.json(
					{
						ok: false,
						error: true,
						message: error instanceof Error ? error.message : String(error),
					},
					{ status: 500 },
				);
			}
		}

		if (url.pathname === "/config" && request.method === "POST") {
			try {
				const body = (await request.json()) as { deployer?: string | null };
				const next = typeof body.deployer === "string" ? body.deployer.trim() : "";
				setDeployer(next);
				workerPool.updateConfig({ deployer: getDeployer() });
				return Response.json({ ok: true, deployer: getDeployer() });
			} catch (error) {
				return Response.json(
					{ error: true, message: error instanceof Error ? error.message : String(error) },
					{ status: 400 },
				);
			}
		}

		if (url.pathname === "/render" && request.method === "POST") {
			let body: RenderRequest | undefined;
			let renderStarted = Date.now();
			try {
				body = (await request.json()) as RenderRequest;
				const precision = parsePositiveNumber(body.precision, defaultPrecision);
				const width = body.width ?? 800;
				const height = body.height;
				renderStarted = Date.now();
				const summary = renderDataSummary(body.template, body.data);
				console.info(`[renderer] render start template=${body.template} width=${width} height=${height ?? "auto"}${summary ? ` ${summary}` : ""}`);

				// 命中缓存则直接返回，省掉 worker + createElement + satori + resvg 整条链路。
				const cached = getCachedRender(body.template, body.data, width, height, precision);
				if (cached) {
					const elapsed = Date.now() - renderStarted;
					console.info(`[renderer] render cache hit template=${body.template} elapsed=${elapsed}ms bytes=${cached.png.length}`);
					return new Response(new Uint8Array(cached.png), {
						headers: {
							...cached.headers,
							"x-render-cache": "hit",
							"x-render-total-ms": String(elapsed),
							"x-render-fonts-ms": "0",
							"x-render-images-ms": "0",
							"x-render-satori-ms": "0",
							"x-render-resvg-ms": "0",
						},
					});
				}

				const rendered = await workerPool.run({ type: "render", request: { ...body, precision } }) as any;
				setCachedRender(body.template, body.data, width, height, precision, rendered.png, rendered.headers);
				console.info(`[renderer] render ok template=${body.template} elapsed=${Date.now() - renderStarted}ms bytes=${rendered.png.length}`);
				return new Response(new Uint8Array(rendered.png), {
					headers: { ...rendered.headers, "x-render-cache": "miss" },
				});
			} catch (error) {
				const rejected = budgetRejectionResponse(error);
				if (rejected) return rejected;
				console.error(`[renderer] render failed template=${body?.template ?? "unknown"} elapsed=${Date.now() - renderStarted}ms:`, error);
				return Response.json(
					{
						error: true,
						message: error instanceof Error ? error.message : String(error),
					},
					{ status: 500 },
				);
			}
		}

		return Response.json(
			{ error: true, message: "not found" },
			{ status: 404 },
		);
	},
});

console.log(`[renderer] Moebot renderer listening on http://127.0.0.1:${port}`);

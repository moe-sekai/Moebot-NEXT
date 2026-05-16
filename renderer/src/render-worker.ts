import { parentPort } from "node:worker_threads";
import { setDeployer } from "./deployer";
import { setFontPreferences } from "./fonts";
import { setMasterSnapshot, setMusicMetasSnapshot } from "./deck-recommend/snapshot-store";
import {
	calculateDeckRecommendJob,
	renderChartJob,
	renderPreviewJob,
	renderTemplateJob,
	setRenderJobConfig,
} from "./render-jobs";
import type { WorkerConfigState, WorkerJobPayload, WorkerRequestMessage, WorkerResponseMessage } from "./render-worker-types";

if (!parentPort) {
	throw new Error("render-worker must run inside a Worker thread");
}

let currentConfig: WorkerConfigState = {};

function applyConfig(config?: WorkerConfigState) {
	if (!config) return;
	currentConfig = { ...currentConfig, ...config };
	if (typeof config.deployer === "string") setDeployer(config.deployer);
	if (typeof config.fontBody === "string" || typeof config.fontScore === "string") {
		setFontPreferences(config.fontBody, config.fontScore);
	}
	if (typeof config.prepareConcurrency === "number") {
		setRenderJobConfig({ prepareConcurrency: config.prepareConcurrency });
	}
}

async function runJob(job: WorkerJobPayload) {
	applyConfig(job.config);
	switch (job.type) {
		case "render":
			return renderTemplateJob(job.request);
		case "preview":
			return renderPreviewJob({ id: job.id, width: job.width, height: job.height, precision: job.precision });
		case "chart":
			return renderChartJob(job.request);
		case "deck-calculate":
			return calculateDeckRecommendJob(job.request);
		case "snapshot": {
			const region = String(job.request.region ?? "jp").trim() || "jp";
			const out: { ok: true; region: string; master?: { version: string; keyCount: number; updatedAt: number }; musicMetas?: { version: string; count: number; updatedAt: number } } = { ok: true, region };
			if (job.request.master?.data) {
				const snap = setMasterSnapshot(region, job.request.master.data, String(job.request.master.version ?? Date.now()));
				out.master = { version: snap.version, keyCount: snap.keyCount, updatedAt: snap.updatedAt };
			}
			if (job.request.musicMetas?.data) {
				const snap = setMusicMetasSnapshot(region, job.request.musicMetas.data, String(job.request.musicMetas.version ?? Date.now()));
				out.musicMetas = { version: snap.version, count: snap.count, updatedAt: snap.updatedAt };
			}
			return out;
		}
		case "config":
			applyConfig(job.config);
			return { ok: true as const, config: currentConfig };
		default: {
			const neverJob: never = job;
			throw new Error(`unknown worker job: ${JSON.stringify(neverJob)}`);
		}
	}
}

parentPort.on("message", (message: WorkerRequestMessage) => {
	void (async () => {
		const response: WorkerResponseMessage = await runJob(message.job)
			.then((result) => ({ id: message.id, ok: true as const, result }))
			.catch((error) => ({
				id: message.id,
				ok: false as const,
				error: error instanceof Error ? error.message : String(error),
				stack: error instanceof Error ? error.stack : undefined,
			}));
		parentPort!.postMessage(response);
	})();
});

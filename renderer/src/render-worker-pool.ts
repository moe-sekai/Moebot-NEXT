import { Worker } from "node:worker_threads";
import type { DeckRecommendSnapshotRequest } from "./deck-recommend/types";
import { BudgetRejectedError } from "./render-budget";
import type { SnapshotReplay, WorkerConfigState, WorkerJobPayload, WorkerJobResult, WorkerRequestMessage, WorkerResponseMessage } from "./render-worker-types";

const HARD_MAX_WORKERS = 32;
const HARD_MAX_QUEUE = 1024;
const DEFAULT_WORKERS = 2;
const DEFAULT_QUEUE = 8;

interface PoolState {
	maxConcurrency: number;
	queueLimit: number;
	inFlight: number;
	completed: number;
	rejected: number;
	totalWaitMs: number;
	peakInFlight: number;
	peakQueued: number;
	spawned: number;
	restarted: number;
}

interface PendingJob {
	id: number;
	job: WorkerJobPayload;
	enqueuedAt: number;
	resolve: (result: WorkerJobResult) => void;
	reject: (error: Error) => void;
}

interface WorkerSlot {
	id: number;
	worker: Worker;
	busy: boolean;
	current?: PendingJob;
}

export class RenderWorkerPool {
	private readonly state: PoolState;
	private readonly workers: WorkerSlot[] = [];
	private readonly queue: PendingJob[] = [];
	private nextJobId = 1;
	private nextWorkerId = 1;
	private config: WorkerConfigState = {};
	private snapshots = new Map<string, SnapshotReplay>();

	constructor(opts: { maxConcurrency?: number; queueLimit?: number; config?: WorkerConfigState } = {}) {
		this.state = {
			maxConcurrency: clampInt(opts.maxConcurrency, DEFAULT_WORKERS, 1, HARD_MAX_WORKERS),
			queueLimit: clampInt(opts.queueLimit, DEFAULT_QUEUE, 0, HARD_MAX_QUEUE),
			inFlight: 0,
			completed: 0,
			rejected: 0,
			totalWaitMs: 0,
			peakInFlight: 0,
			peakQueued: 0,
			spawned: 0,
			restarted: 0,
		};
		this.config = { ...opts.config };
		this.ensureWorkerCount();
	}

	configure(opts: { maxConcurrency?: number; queueLimit?: number; prepareConcurrency?: number }) {
		if (opts.queueLimit !== undefined) {
			this.state.queueLimit = clampInt(opts.queueLimit, this.state.queueLimit, 0, HARD_MAX_QUEUE);
		}
		if (opts.prepareConcurrency !== undefined) {
			this.updateConfig({ prepareConcurrency: opts.prepareConcurrency });
		}
		if (opts.maxConcurrency !== undefined) {
			this.state.maxConcurrency = clampInt(opts.maxConcurrency, this.state.maxConcurrency, 1, HARD_MAX_WORKERS);
			this.ensureWorkerCount();
		}
		this.drain();
		return {
			maxConcurrency: this.state.maxConcurrency,
			queueLimit: this.state.queueLimit,
			prepareConcurrency: this.config.prepareConcurrency ?? 0,
		};
	}

	updateConfig(config: WorkerConfigState) {
		this.config = { ...this.config, ...config };
		for (const slot of this.workers) {
			this.postControl(slot, { type: "config", config: this.config });
		}
		return { ...this.config };
	}

	rememberSnapshot(req: DeckRecommendSnapshotRequest) {
		const region = String(req.region ?? "jp").trim() || "jp";
		const previous = this.snapshots.get(region) ?? { region };
		this.snapshots.set(region, {
			region,
			master: req.master ?? previous.master,
			musicMetas: req.musicMetas ?? previous.musicMetas,
		});
	}

	async broadcastSnapshot(req: DeckRecommendSnapshotRequest): Promise<void> {
		this.rememberSnapshot(req);
		await Promise.all(this.workers.map((slot) => this.runOnSlot(slot, { type: "snapshot", request: req, config: this.config }, true).then(() => undefined)));
	}

	run(job: WorkerJobPayload): Promise<WorkerJobResult> {
		if (this.queue.length >= this.state.queueLimit && !this.idleWorker()) {
			this.state.rejected += 1;
			throw new BudgetRejectedError(`渲染队列已满 (workers=${this.state.maxConcurrency}, queue=${this.state.queueLimit}); 请稍后重试`);
		}
		return new Promise((resolve, reject) => {
			const pending: PendingJob = {
				id: this.nextJobId++,
				job: withConfig(job, this.config),
				enqueuedAt: Date.now(),
				resolve,
				reject,
			};
			this.queue.push(pending);
			if (this.queue.length > this.state.peakQueued) this.state.peakQueued = this.queue.length;
			this.drain();
		});
	}

	stats() {
		return {
			maxConcurrency: this.state.maxConcurrency,
			queueLimit: this.state.queueLimit,
			inFlight: this.state.inFlight,
			queued: this.queue.length,
			completed: this.state.completed,
			rejected: this.state.rejected,
			avgWaitMs: this.state.completed > 0 ? Math.round(this.state.totalWaitMs / this.state.completed) : 0,
			peakInFlight: this.state.peakInFlight,
			peakQueued: this.state.peakQueued,
			workerCount: this.workers.length,
			busyWorkers: this.workers.filter((worker) => worker.busy).length,
			spawned: this.state.spawned,
			restarted: this.state.restarted,
			prepareConcurrency: this.config.prepareConcurrency ?? 0,
			limits: {
				hardMaxConcurrency: HARD_MAX_WORKERS,
				hardMaxQueue: HARD_MAX_QUEUE,
				hardMaxPrepareConcurrency: 32,
			},
		};
	}

	private drain() {
		while (this.queue.length > 0) {
			const slot = this.idleWorker();
			if (!slot) return;
			const pending = this.queue.shift()!;
			void this.dispatch(slot, pending);
		}
	}

	private async dispatch(slot: WorkerSlot, pending: PendingJob) {
		slot.busy = true;
		slot.current = pending;
		this.state.inFlight += 1;
		if (this.state.inFlight > this.state.peakInFlight) this.state.peakInFlight = this.state.inFlight;
		const waitMs = Date.now() - pending.enqueuedAt;
		try {
			const result = await this.runOnSlot(slot, pending.job, false, pending.id);
			this.state.completed += 1;
			this.state.totalWaitMs += waitMs;
			pending.resolve(result);
		} catch (error) {
			pending.reject(error instanceof Error ? error : new Error(String(error)));
		} finally {
			this.state.inFlight -= 1;
			slot.busy = false;
			slot.current = undefined;
			this.ensureWorkerCount();
			this.drain();
		}
	}

	private runOnSlot(slot: WorkerSlot, job: WorkerJobPayload, _control = false, id = this.nextJobId++): Promise<WorkerJobResult> {
		return new Promise((resolve, reject) => {
			const onMessage = (message: WorkerResponseMessage) => {
				if (message.id !== id) return;
				cleanup();
				if (message.ok) resolve(message.result);
				else reject(new Error(message.error));
			};
			const onError = (error: Error) => {
				cleanup();
				reject(error);
			};
			const cleanup = () => {
				slot.worker.off("message", onMessage);
				slot.worker.off("error", onError);
			};
			slot.worker.on("message", onMessage);
			slot.worker.on("error", onError);
			const message: WorkerRequestMessage = { id, job: withConfig(job, this.config) };
			try {
				slot.worker.postMessage(message);
			} catch (error) {
				cleanup();
				reject(error instanceof Error ? error : new Error(String(error)));
			}
		});
	}

	private postControl(slot: WorkerSlot, job: WorkerJobPayload) {
		void this.runOnSlot(slot, job, true).catch((error) => {
			console.warn(`[renderer] worker ${slot.id} control message failed:`, error);
		});
	}

	private idleWorker(): WorkerSlot | undefined {
		return this.workers.find((worker) => !worker.busy);
	}

	private ensureWorkerCount() {
		while (this.workers.length < this.state.maxConcurrency) {
			this.workers.push(this.createWorker());
		}
		while (this.workers.length > this.state.maxConcurrency) {
			const removableIndex = this.workers.findIndex((worker) => !worker.busy);
			if (removableIndex < 0) break;
			const [slot] = this.workers.splice(removableIndex, 1);
			void slot.worker.terminate();
		}
	}

	private createWorker(): WorkerSlot {
		const slot: WorkerSlot = {
			id: this.nextWorkerId++,
			worker: new Worker(new URL("./render-worker.ts", import.meta.url)),
			busy: false,
		};
		this.state.spawned += 1;
		slot.worker.on("error", (error) => this.handleWorkerCrash(slot, error instanceof Error ? error : new Error(String(error))));
		slot.worker.on("exit", (code) => {
			if (code !== 0) this.handleWorkerCrash(slot, new Error(`worker exited with code ${code}`));
		});
		this.postControl(slot, { type: "config", config: this.config });
		for (const snap of this.snapshots.values()) {
			this.postControl(slot, { type: "snapshot", request: snapshotToRequest(snap), config: this.config });
		}
		return slot;
	}

	private handleWorkerCrash(slot: WorkerSlot, error: Error) {
		console.error(`[renderer] worker ${slot.id} crashed:`, error);
		const index = this.workers.indexOf(slot);
		if (index >= 0) this.workers.splice(index, 1);
		if (slot.current) {
			slot.current.reject(error);
			slot.current = undefined;
			if (this.state.inFlight > 0) this.state.inFlight -= 1;
		}
		void slot.worker.terminate().catch(() => {});
		this.state.restarted += 1;
		this.ensureWorkerCount();
		this.drain();
	}
}

function withConfig(job: WorkerJobPayload, config: WorkerConfigState): WorkerJobPayload {
	if (job.type === "config") return { type: "config", config: { ...config, ...job.config } };
	return { ...job, config: { ...config, ...job.config } } as WorkerJobPayload;
}

function snapshotToRequest(snap: SnapshotReplay): DeckRecommendSnapshotRequest {
	return {
		region: snap.region,
		...(snap.master ? { master: snap.master } : {}),
		...(snap.musicMetas ? { musicMetas: snap.musicMetas } : {}),
	};
}

function clampInt(value: unknown, fallback: number, min: number, max: number): number {
	const n = typeof value === "number" ? value : Number(value);
	if (!Number.isFinite(n)) return fallback;
	const r = Math.floor(n);
	if (r < min) return min;
	if (r > max) return max;
	return r;
}

/**
 * Render concurrency budget.
 *
 * Bun.serve 是单进程单事件循环；当多个 /render 请求同时打进来时，satori 会
 * 在事件循环里互相挤占内存（每张卡 SVG/PNG 都是几十 MB），在 16G 灵车上很容
 * 易 OOM。这里实现一个轻量信号量：
 *
 *   - 同时运行的渲染数 ≤ maxConcurrency
 *   - 等待队列长度 ≤ queueLimit；超过直接拒绝（503），保护服务不雪崩
 *   - 上限可在运行时通过控制台前端热更新
 *
 * 这个模块只关心“并发预算”，不关心具体渲染逻辑，调用方用 runWithBudget()
 * 包住任何 CPU 密集的渲染入口即可。
 */

const HARD_MAX_CONCURRENCY = 64;
const HARD_MAX_QUEUE = 1024;

interface BudgetState {
  maxConcurrency: number;
  queueLimit: number;
  inFlight: number;
  queued: number;
  completed: number;
  rejected: number;
  totalWaitMs: number;
  peakInFlight: number;
  peakQueued: number;
}

const state: BudgetState = {
  maxConcurrency: 2,
  queueLimit: 8,
  inFlight: 0,
  queued: 0,
  completed: 0,
  rejected: 0,
  totalWaitMs: 0,
  peakInFlight: 0,
  peakQueued: 0,
};

type Waiter = () => void;
const waiters: Waiter[] = [];

function clampInt(value: unknown, fallback: number, min: number, max: number): number {
  const n = typeof value === 'number' ? value : Number(value);
  if (!Number.isFinite(n)) return fallback;
  const r = Math.floor(n);
  if (r < min) return min;
  if (r > max) return max;
  return r;
}

export function configureRenderBudget(opts: { maxConcurrency?: number; queueLimit?: number }): {
  maxConcurrency: number;
  queueLimit: number;
} {
  if (opts.maxConcurrency !== undefined) {
    state.maxConcurrency = clampInt(opts.maxConcurrency, state.maxConcurrency, 1, HARD_MAX_CONCURRENCY);
  }
  if (opts.queueLimit !== undefined) {
    state.queueLimit = clampInt(opts.queueLimit, state.queueLimit, 0, HARD_MAX_QUEUE);
  }
  // 上调 maxConcurrency 时，立刻唤醒等待者填满空位。
  drainWaiters();
  return { maxConcurrency: state.maxConcurrency, queueLimit: state.queueLimit };
}

export function getRenderBudgetStats() {
  return {
    maxConcurrency: state.maxConcurrency,
    queueLimit: state.queueLimit,
    inFlight: state.inFlight,
    queued: state.queued,
    completed: state.completed,
    rejected: state.rejected,
    avgWaitMs: state.completed > 0 ? Math.round(state.totalWaitMs / state.completed) : 0,
    peakInFlight: state.peakInFlight,
    peakQueued: state.peakQueued,
    limits: {
      hardMaxConcurrency: HARD_MAX_CONCURRENCY,
      hardMaxQueue: HARD_MAX_QUEUE,
    },
  };
}

function drainWaiters() {
  while (waiters.length > 0 && state.inFlight < state.maxConcurrency) {
    const fn = waiters.shift()!;
    state.queued = waiters.length;
    fn();
  }
}

/**
 * BudgetRejectedError：队列已满，调用方应返回 503。
 */
export class BudgetRejectedError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'BudgetRejectedError';
  }
}

/**
 * runWithBudget(fn)：拿到许可后执行 fn，否则进入队列；队列满则抛出
 * BudgetRejectedError。
 */
export async function runWithBudget<T>(fn: () => Promise<T>): Promise<T> {
  const enqueueStart = Date.now();
  if (state.inFlight >= state.maxConcurrency) {
    if (waiters.length >= state.queueLimit) {
      state.rejected += 1;
      throw new BudgetRejectedError(
        `渲染队列已满 (max=${state.maxConcurrency}, queue=${state.queueLimit}); 请稍后重试`,
      );
    }
    await new Promise<void>((resolve) => {
      waiters.push(resolve);
      state.queued = waiters.length;
      if (state.queued > state.peakQueued) state.peakQueued = state.queued;
    });
  }
  const waitMs = Date.now() - enqueueStart;
  state.inFlight += 1;
  if (state.inFlight > state.peakInFlight) state.peakInFlight = state.inFlight;
  try {
    const result = await fn();
    state.completed += 1;
    state.totalWaitMs += waitMs;
    return result;
  } finally {
    state.inFlight -= 1;
    drainWaiters();
  }
}

import { createHash } from 'node:crypto'
import { mkdir, readFile, readdir, stat, unlink, writeFile } from 'node:fs/promises'
import { join } from 'node:path'

export interface AssetCacheOptions {
  enabled: boolean
  cacheDir: string
  // 0 means unlimited; card thumbnail assets are immutable enough to keep forever.
  maxSizeBytes: number
  // 0 means never expire.
  ttlMs: number
  requestTimeoutMs: number
  preloadConcurrency: number
}

export interface CachedImageResult {
  hit: boolean
  dataUri?: string
  skipped?: boolean
  error?: string
}

export interface CachedBytesResult {
  hit: boolean
  data?: Buffer
  skipped?: boolean
  error?: string
}

export interface ImageHydrationStats {
  total: number
  remote: number
  hits: number
  misses: number
  errors: number
  skipped: number
}

export interface AssetPreloadStatus {
  ok: boolean
  enabled: boolean
  running: boolean
  message: string
  cache_dir: string
  total: number
  cached: number
  missing: number
  failed: number
  downloaded: number
  skipped: number
  progress: number
  started_at: string | null
  completed_at: string | null
  errors: string[]
  composite_total?: number
  composite_cached?: number
  composite_missing?: number
  composite_failed?: number
  composite_generated?: number
  composite_running?: boolean
  composite_errors?: string[]
}

type PreloadOutcomeStatus = 'cached' | 'downloaded' | 'failed' | 'skipped' | 'disabled'

interface PreloadOutcome {
  url: string
  status: PreloadOutcomeStatus
  error?: string
}

interface PreloadJob {
  urls: string[]
  running: boolean
  startedAt: string
  completedAt: string | null
  downloaded: number
  failed: number
  skipped: number
  processed: number
}

interface StartPreloadOptions {
  force?: boolean
  concurrency?: number
}

const DEFAULT_MAX_SIZE_BYTES = 0
const DEFAULT_TTL_MS = 0
const DEFAULT_TIMEOUT_MS = 12_000
const DEFAULT_PRELOAD_CONCURRENCY = 10
const CLEANUP_INTERVAL_MS = 30 * 60 * 1000

export function createImageHydrationStats(): ImageHydrationStats {
  return {
    total: 0,
    remote: 0,
    hits: 0,
    misses: 0,
    errors: 0,
    skipped: 0,
  }
}

export class RendererAssetCache {
  private initialized = false
  private inflight = new Map<string, Promise<PreloadOutcome>>()
  private failedUrls = new Map<string, string>()
  private currentJob: PreloadJob | null = null
  private lastJob: PreloadJob | null = null
  private lastCleanupAt = 0

  constructor(private readonly options: AssetCacheOptions = cacheOptionsFromEnv()) {}

  get enabled(): boolean {
    return this.options.enabled
  }

  get cacheDir(): string {
    return this.options.cacheDir
  }

  async init(): Promise<void> {
    if (!this.options.enabled || this.initialized) return
    await mkdir(this.options.cacheDir, { recursive: true })
    this.initialized = true
  }

  isRemoteUrl(value: unknown): value is string {
    return typeof value === 'string' && /^https?:\/\//i.test(value)
  }

  async getBytes(url: string): Promise<CachedBytesResult> {
    if (!this.options.enabled) return { hit: false, skipped: true }
    if (!this.isRemoteUrl(url)) return { hit: false, skipped: true }

    try {
      const data = await this.readFresh(url)
      if (!data) return { hit: false }
      return { hit: true, data }
    } catch (error) {
      return {
        hit: false,
        error: error instanceof Error ? error.message : String(error),
      }
    }
  }

  async getDataUri(url: string): Promise<CachedImageResult> {
    const cached = await this.getBytes(url)
    if (!cached.hit || !cached.data) {
      return { hit: false, skipped: cached.skipped, error: cached.error }
    }
    return {
      hit: true,
      dataUri: `data:${mimeFromUrl(url)};base64,${cached.data.toString('base64')}`,
    }
  }

  async startPreload(urls: string[], options: StartPreloadOptions = {}): Promise<AssetPreloadStatus> {
    const uniqueUrls = uniqueRemoteUrls(urls)
    if (!this.options.enabled) {
      return this.statusForUrls(uniqueUrls, 'Renderer 图片缓存未启用')
    }
    if (this.currentJob?.running) {
      return this.statusForUrls(uniqueUrls, '已有卡牌缩略图预载任务运行中')
    }

    const job: PreloadJob = {
      urls: uniqueUrls,
      running: true,
      startedAt: new Date().toISOString(),
      completedAt: null,
      downloaded: 0,
      failed: 0,
      skipped: 0,
      processed: 0,
    }
    this.currentJob = job
    this.lastJob = job

    void this.runPreload(job, options).catch((error) => {
      const message = error instanceof Error ? error.message : String(error)
      for (const preloadUrl of job.urls) {
        if (!this.failedUrls.has(preloadUrl)) {
          this.failedUrls.set(preloadUrl, message)
        }
      }
      job.running = false
      job.completedAt = new Date().toISOString()
      this.currentJob = null
    })

    return this.statusForUrls(uniqueUrls, '卡牌缩略图预载已启动')
  }

  async statusForUrls(urls: string[], message = '卡牌缩略图缓存状态已返回'): Promise<AssetPreloadStatus> {
    const uniqueUrls = uniqueRemoteUrls(urls)
    const cachedUrls = new Set<string>()
    for (const url of uniqueUrls) {
      if (await this.hasFresh(url)) cachedUrls.add(url)
    }

    const failedUrls = uniqueUrls.filter((url) => !cachedUrls.has(url) && this.failedUrls.has(url))
    const job = this.currentJob?.running ? this.currentJob : this.lastJob
    const total = uniqueUrls.length
    const cached = cachedUrls.size
    const failed = failedUrls.length
    const running = Boolean(this.currentJob?.running)
    const displayCached = running ? Math.max(cached, job?.processed ?? 0) : cached

    return {
      ok: this.options.enabled,
      enabled: this.options.enabled,
      running,
      message,
      cache_dir: this.options.cacheDir,
      total,
      cached: displayCached,
      missing: Math.max(0, total - displayCached),
      failed,
      downloaded: job?.downloaded ?? 0,
      skipped: job?.skipped ?? 0,
      progress: job?.running && total > 0 ? Math.min(0.999, job.processed / total) : total === 0 ? 1 : cached / total,
      started_at: job?.startedAt ?? null,
      completed_at: job?.completedAt ?? null,
      errors: failedUrls.slice(0, 8).map((url) => `${url}: ${this.failedUrls.get(url)}`),
    }
  }

  async prefetch(url: string, force = false): Promise<PreloadOutcome> {
    if (!this.options.enabled) return { url, status: 'disabled' }
    if (!this.isRemoteUrl(url)) return { url, status: 'skipped' }
    if (!force && await this.hasFresh(url)) return { url, status: 'cached' }

    const running = this.inflight.get(url)
    if (running) return running

    const task = this.downloadAndStore(url)
    this.inflight.set(url, task)
    try {
      return await task
    } finally {
      this.inflight.delete(url)
    }
  }

  private async runPreload(job: PreloadJob, options: StartPreloadOptions): Promise<void> {
    await this.init()
    const concurrency = normalizeConcurrency(options.concurrency ?? this.options.preloadConcurrency)
    await runPool(job.urls, concurrency, async (url) => {
      const outcome = await this.prefetch(url, Boolean(options.force))
      switch (outcome.status) {
        case 'downloaded':
          job.downloaded += 1
          break
        case 'failed':
          job.failed += 1
          break
        case 'skipped':
        case 'disabled':
          job.skipped += 1
          break
      }
      job.processed += 1
    })
    job.running = false
    job.completedAt = new Date().toISOString()
    this.currentJob = null
    await this.cleanup(true)
  }

  private async downloadAndStore(url: string): Promise<PreloadOutcome> {
    try {
      await this.init()
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), this.options.requestTimeoutMs)
      try {
        const response = await fetch(url, { signal: controller.signal })
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`)
        }
        const buffer = Buffer.from(await response.arrayBuffer())
        if (buffer.length === 0) {
          throw new Error('empty image body')
        }
        await writeFile(this.getPath(url), buffer)
        this.failedUrls.delete(url)
        void this.cleanup(false)
        return { url, status: 'downloaded' }
      } finally {
        clearTimeout(timeout)
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error)
      this.failedUrls.set(url, message)
      return { url, status: 'failed', error: message }
    }
  }

  private async hasFresh(url: string): Promise<boolean> {
    if (!this.options.enabled || !this.isRemoteUrl(url)) return false
    return Boolean(await this.freshStat(url))
  }

  private async readFresh(url: string): Promise<Buffer | null> {
    const fresh = await this.freshStat(url)
    if (!fresh) return null
    return readFile(this.getPath(url))
  }

  private async freshStat(url: string) {
    try {
      const filePath = this.getPath(url)
      const fileStat = await stat(filePath)
      if (this.options.ttlMs > 0 && Date.now() - fileStat.mtimeMs > this.options.ttlMs) {
        await unlink(filePath).catch(() => {})
        return null
      }
      return fileStat
    } catch {
      return null
    }
  }

  private getPath(url: string): string {
    const hash = createHash('sha256').update(url).digest('hex')
    return join(this.options.cacheDir, `${hash}${extensionFromUrl(url)}`)
  }

  private async cleanup(force: boolean): Promise<void> {
    if (!this.options.enabled) return
    if (this.options.ttlMs <= 0 && this.options.maxSizeBytes <= 0) return
    const now = Date.now()
    if (!force && now - this.lastCleanupAt < CLEANUP_INTERVAL_MS) return
    this.lastCleanupAt = now

    try {
      await this.init()
      const files = await readdir(this.options.cacheDir)
      const fileStats = await Promise.all(
        files.map(async (file) => {
          const path = join(this.options.cacheDir, file)
          const fileStat = await stat(path)
          return { path, size: fileStat.size, mtime: fileStat.mtimeMs }
        }),
      )

      const freshFiles = [] as Array<{ path: string; size: number; mtime: number }>
      for (const file of fileStats) {
        if (this.options.ttlMs > 0 && now - file.mtime > this.options.ttlMs) {
          await unlink(file.path).catch(() => {})
        } else {
          freshFiles.push(file)
        }
      }

      if (this.options.maxSizeBytes <= 0) return
      freshFiles.sort((a, b) => a.mtime - b.mtime)
      let totalSize = freshFiles.reduce((sum, file) => sum + file.size, 0)
      for (const file of freshFiles) {
        if (totalSize <= this.options.maxSizeBytes) break
        await unlink(file.path).catch(() => {})
        totalSize -= file.size
      }
    } catch {
      // Cache cleanup is best-effort.
    }
  }
}

export const rendererAssetCache = new RendererAssetCache()

function cacheOptionsFromEnv(): AssetCacheOptions {
  const rootDir = process.env.RENDER_CACHE_DIR?.trim() || join(process.cwd(), 'data', 'cache')
  return {
    enabled: envBoolean(process.env.RENDER_CACHE_ENABLED, true),
    cacheDir: join(rootDir, 'renderer-assets'),
    maxSizeBytes: envNumber(process.env.RENDER_CACHE_MAX_SIZE_BYTES, DEFAULT_MAX_SIZE_BYTES),
    ttlMs: envNumber(process.env.RENDER_CACHE_TTL_MS, DEFAULT_TTL_MS),
    requestTimeoutMs: envNumber(process.env.RENDER_IMAGE_TIMEOUT_MS, DEFAULT_TIMEOUT_MS),
    preloadConcurrency: normalizeConcurrency(envNumber(process.env.RENDER_PRELOAD_CONCURRENCY, DEFAULT_PRELOAD_CONCURRENCY)),
  }
}

function uniqueRemoteUrls(urls: string[]): string[] {
  const seen = new Set<string>()
  const result: string[] = []
  for (const raw of urls) {
    const url = typeof raw === 'string' ? raw.trim() : ''
    if (!/^https?:\/\//i.test(url) || seen.has(url)) continue
    seen.add(url)
    result.push(url)
  }
  return result
}

async function runPool<T>(items: T[], concurrency: number, worker: (item: T) => Promise<void>): Promise<void> {
  let nextIndex = 0
  const workers = Array.from({ length: Math.min(concurrency, Math.max(1, items.length)) }, async () => {
    while (nextIndex < items.length) {
      const currentIndex = nextIndex
      nextIndex += 1
      await worker(items[currentIndex] as T)
    }
  })
  await Promise.all(workers)
}

function envBoolean(value: string | undefined, fallback: boolean): boolean {
  if (value === undefined || value === '') return fallback
  return !['0', 'false', 'no', 'off'].includes(value.toLowerCase())
}

function envNumber(value: string | undefined, fallback: number): number {
  if (value === undefined || value === '') return fallback
  const numberValue = Number(value)
  return Number.isFinite(numberValue) && numberValue >= 0 ? numberValue : fallback
}

function normalizeConcurrency(value: number): number {
  if (!Number.isFinite(value) || value <= 0) return DEFAULT_PRELOAD_CONCURRENCY
  return Math.max(1, Math.min(32, Math.round(value)))
}

function extensionFromUrl(url: string): string {
  try {
    const pathname = new URL(url).pathname.toLowerCase()
    const match = pathname.match(/\.(png|jpg|jpeg|webp|gif|svg)$/)
    if (match) return `.${match[1]}`
  } catch {
    // Fall through to PNG because PJSK renderer assets are PNG by default.
  }
  return '.png'
}

function mimeFromUrl(url: string): string {
  switch (extensionFromUrl(url)) {
    case '.jpg':
    case '.jpeg':
      return 'image/jpeg'
    case '.webp':
      return 'image/webp'
    case '.gif':
      return 'image/gif'
    case '.svg':
      return 'image/svg+xml'
    default:
      return 'image/png'
  }
}

import { createHash } from 'node:crypto'
import { mkdir, readFile, rename, stat, unlink, writeFile } from 'node:fs/promises'
import { performance } from 'node:perf_hooks'
import { setImmediate as setImmediatePromise } from 'node:timers/promises'
import { join } from 'node:path'
import { rendererAssetCache } from './asset-cache'
import { getSekaiCardUiAssetDataUri } from './styles/assets'

export interface CardThumbnailCompositeRequest {
  imageUrl?: string
  rarity?: string
  attr?: string
  trained?: boolean
  size?: number
}

export interface CardThumbnailCompositeStatus {
  composite_total: number
  composite_cached: number
  composite_missing: number
  composite_failed: number
  composite_generated: number
  composite_progress: number
  composite_running: boolean
  composite_source_downloaded: number
  composite_source_failed: number
  composite_render_ms: number
  composite_errors: string[]
}

export type CardThumbnailCompositeLayer =
  | { type: 'rect'; width: number; height: number; rx?: number; fill?: string }
  | { type: 'image'; href: string; x: number; y: number; width: number; height: number; preserveAspectRatio?: string }

interface NormalizedCardThumbnailCompositeRequest {
  imageUrl: string
  rarity: string
  attr: string
  trained: boolean
  size: number
}

interface CompositeJob {
  inputs: NormalizedCardThumbnailCompositeRequest[]
  running: boolean
  generated: number
  failed: number
  skipped: number
  processed: number
  sourceDownloads: number
  sourceFailed: number
  renderMs: number
  svgBytes: number
  startedAt: string
  completedAt: string | null
  lastLogAt: number
}

interface CompositeGroup {
  imageUrl: string
  inputs: NormalizedCardThumbnailCompositeRequest[]
}

const DEFAULT_CARD_THUMBNAIL_SIZE = 156
const COMPOSITE_CACHE_VERSION = 'cardtile-svg-v1'
const PROGRESS_LOG_INTERVAL_MS = 5_000
const YIELD_EVERY_COMPOSITES = 64

const failedComposites = new Map<string, string>()
const compositeInflight = new Map<string, Promise<boolean>>()
const memoryComposites = new Set<string>()
let currentCompositeJob: CompositeJob | null = null
let lastCompositeJob: CompositeJob | null = null

export async function getCardThumbnailCompositeDataUri(input: CardThumbnailCompositeRequest, allowDownload = false): Promise<string | undefined> {
  const svg = await getCardThumbnailCompositeSvg(input, allowDownload)
  return svg ? svgToDataUri(svg) : undefined
}

export async function getCardThumbnailCompositeSvg(input: CardThumbnailCompositeRequest, allowDownload = false): Promise<string | undefined> {
  const normalized = normalizeCompositeRequest(input)
  if (!rendererAssetCache.enabled || !normalized) return undefined

  const cached = await readComposite(normalized)
  if (cached) return cached.toString('utf8')

  if (!allowDownload) return undefined

  const generated = await generateComposite(normalized, true)
  if (!generated) console.debug(`[renderer] card thumbnail composite miss: ${normalized.imageUrl} (${normalized.trained ? 'after' : 'before'} ${normalized.size}px) reason=${failedComposites.get(compositeKey(normalized)) ?? 'unknown'}`)
  return generated ? generated.toString('utf8') : undefined
}

export function getCardThumbnailCompositeLayersFromSvg(svg: string): CardThumbnailCompositeLayer[] {
  const layers: CardThumbnailCompositeLayer[] = []
  const rect = svg.match(/<rect\s+([^>]*)\/>/)
  if (rect) {
    const attrs = parseSvgAttrs(rect[1] ?? '')
    layers.push({
      type: 'rect',
      width: numberAttr(attrs.width),
      height: numberAttr(attrs.height),
      rx: optionalNumberAttr(attrs.rx),
      fill: attrs.fill,
    })
  }
  for (const match of svg.matchAll(/<image\s+([^>]*)\/>/g)) {
    const attrs = parseSvgAttrs(match[1] ?? '')
    const href = attrs.href ?? attrs['xlink:href']
    if (!href) continue
    layers.push({
      type: 'image',
      href,
      x: numberAttr(attrs.x),
      y: numberAttr(attrs.y),
      width: numberAttr(attrs.width),
      height: numberAttr(attrs.height),
      preserveAspectRatio: attrs.preserveAspectRatio,
    })
  }
  return layers
}

export async function startCardThumbnailCompositePreload(inputs: CardThumbnailCompositeRequest[], options: { force?: boolean; concurrency?: number } = {}): Promise<CardThumbnailCompositeStatus> {
  const normalized = uniqueCompositeRequests(inputs)
  if (!rendererAssetCache.enabled) return statusForCardThumbnailComposites(normalized)
  if (currentCompositeJob?.running) return statusForCardThumbnailComposites(normalized)

  const job: CompositeJob = {
    inputs: normalized,
    running: true,
    generated: 0,
    failed: 0,
    skipped: 0,
    processed: 0,
    sourceDownloads: 0,
    sourceFailed: 0,
    renderMs: 0,
    svgBytes: 0,
    startedAt: new Date().toISOString(),
    completedAt: null,
    lastLogAt: Date.now(),
  }
  currentCompositeJob = job
  lastCompositeJob = job

  void runCompositePreload(job, options).catch((error) => {
    const message = error instanceof Error ? error.message : String(error)
    for (const input of job.inputs) {
      if (!failedComposites.has(compositeKey(input))) failedComposites.set(compositeKey(input), message)
    }
    job.running = false
    job.completedAt = new Date().toISOString()
    currentCompositeJob = null
    console.error('[renderer] card thumbnail composite preload crashed:', message)
  })

  return statusForCardThumbnailComposites(normalized)
}

export async function statusForCardThumbnailComposites(inputs: CardThumbnailCompositeRequest[]): Promise<CardThumbnailCompositeStatus> {
  const normalized = uniqueCompositeRequests(inputs)
  const job = currentCompositeJob?.running ? currentCompositeJob : lastCompositeJob
  const running = Boolean(currentCompositeJob?.running)

  let cached = 0
  if (running && job) {
    cached = Math.max(job.processed, countMemoryComposites(normalized))
  } else {
    for (const input of normalized) {
      if (await hasComposite(input)) cached += 1
    }
  }

  const failed = normalized.filter((input) => failedComposites.has(compositeKey(input)))
  const jobTotal = running ? (job?.inputs.length ?? normalized.length) : normalized.length
  const jobProcessed = job?.processed ?? 0
  return {
    composite_total: normalized.length,
    composite_cached: Math.min(cached, normalized.length),
    composite_missing: Math.max(0, normalized.length - cached),
    composite_failed: running ? job?.failed ?? failed.length : failed.length,
    composite_generated: job?.generated ?? 0,
    composite_progress: running && jobTotal > 0 ? Math.min(0.999, jobProcessed / jobTotal) : normalized.length === 0 ? 1 : cached / normalized.length,
    composite_running: running,
    composite_source_downloaded: job?.sourceDownloads ?? 0,
    composite_source_failed: job?.sourceFailed ?? 0,
    composite_render_ms: Math.round(job?.renderMs ?? 0),
    composite_errors: failed.slice(0, 8).map((input) => `${input.imageUrl} (${input.trained ? '花后' : '花前'} ${input.size}px): ${failedComposites.get(compositeKey(input))}`),
  }
}

async function runCompositePreload(job: CompositeJob, options: { force?: boolean; concurrency?: number }): Promise<void> {
  console.info(`[renderer] card thumbnail composite preload started: total=${job.inputs.length}`)
  await mkdir(rendererAssetCache.cacheDir, { recursive: true })

  const groups = groupCompositeRequestsByImageUrl(job.inputs)
  const concurrency = normalizeConcurrency(options.concurrency)
  try {
    await runPool(groups, concurrency, async (group) => {
      const pending = [] as NormalizedCardThumbnailCompositeRequest[]
      for (const input of group.inputs) {
        try {
          if (!options.force && await hasComposite(input)) {
            job.skipped += 1
            markCompositeProcessed(job)
          } else {
            pending.push(input)
          }
        } catch (error) {
          failedComposites.set(compositeKey(input), error instanceof Error ? error.message : String(error))
          job.failed += 1
          markCompositeProcessed(job)
        }
      }
      if (pending.length === 0) return

      const source = await sourceImageDataUri(group.imageUrl, true)
      if (!source) {
        job.sourceFailed += 1
        for (const input of pending) {
          failedComposites.set(compositeKey(input), 'source image is not cached')
          job.failed += 1
          markCompositeProcessed(job)
        }
        return
      }
      job.sourceDownloads += 1

      for (const input of pending) {
        const started = performance.now()
        const result = await generateCompositeFromSource(input, source)
        job.renderMs += performance.now() - started
        if (result.ok) {
          job.generated += 1
          job.svgBytes += result.bytes
        } else {
          job.failed += 1
        }
        markCompositeProcessed(job)
        if (job.processed % YIELD_EVERY_COMPOSITES === 0) await setImmediatePromise()
      }
    })
  } finally {
    job.running = false
    job.completedAt = new Date().toISOString()
    currentCompositeJob = null
    console.info(`[renderer] card thumbnail composite preload completed: total=${job.inputs.length}, generated=${job.generated}, skipped=${job.skipped}, failed=${job.failed}, source_failed=${job.sourceFailed}, svg_ms=${Math.round(job.renderMs)}, svg_kb=${Math.round(job.svgBytes / 1024)}`)
  }
}

function markCompositeProcessed(job: CompositeJob): void {
  job.processed += 1
  const now = Date.now()
  if (job.processed === job.inputs.length || now - job.lastLogAt >= PROGRESS_LOG_INTERVAL_MS) {
    job.lastLogAt = now
    console.info(`[renderer] card thumbnail composite preload progress: ${job.processed}/${job.inputs.length}, generated=${job.generated}, skipped=${job.skipped}, failed=${job.failed}, source_failed=${job.sourceFailed}`)
  }
}

async function prefetchCardThumbnailComposite(input: NormalizedCardThumbnailCompositeRequest, force = false): Promise<boolean> {
  const key = compositeKey(input)
  if (!force && await hasComposite(input)) return true
  const running = compositeInflight.get(key)
  if (running) return running

  const task = generateComposite(input, true).then(Boolean)
  compositeInflight.set(key, task)
  try {
    return await task
  } finally {
    compositeInflight.delete(key)
  }
}

async function generateComposite(input: NormalizedCardThumbnailCompositeRequest, allowDownload: boolean): Promise<Buffer | null> {
  try {
    const source = await sourceImageDataUri(input.imageUrl, allowDownload)
    if (!source) {
      failedComposites.set(compositeKey(input), 'source image is not cached')
      return null
    }
    const svg = renderCompositeSvg(input, source)
    await writeFileAtomic(compositePath(input), svg)
    memoryComposites.add(compositeKey(input))
    failedComposites.delete(compositeKey(input))
    return Buffer.from(svg)
  } catch (error) {
    memoryComposites.delete(compositeKey(input))
    failedComposites.set(compositeKey(input), error instanceof Error ? error.message : String(error))
    return null
  }
}

async function generateCompositeFromSource(input: NormalizedCardThumbnailCompositeRequest, source: string): Promise<{ ok: boolean; bytes: number }> {
  try {
    const svg = renderCompositeSvg(input, source)
    await writeFileAtomic(compositePath(input), svg)
    memoryComposites.add(compositeKey(input))
    failedComposites.delete(compositeKey(input))
    return { ok: true, bytes: Buffer.byteLength(svg) }
  } catch (error) {
    memoryComposites.delete(compositeKey(input))
    failedComposites.set(compositeKey(input), error instanceof Error ? error.message : String(error))
    return { ok: false, bytes: 0 }
  }
}

function renderCompositeSvg(input: NormalizedCardThumbnailCompositeRequest, source: string): string {
  return buildCardThumbnailCompositeSvg(input, source)
}

async function sourceImageDataUri(url: string, allowDownload: boolean): Promise<string | undefined> {
  let cached = await rendererAssetCache.getDataUri(url)
  if (cached.hit && cached.dataUri) return cached.dataUri
  if (!allowDownload) return undefined

  const outcome = await rendererAssetCache.prefetch(url)
  if (outcome.status === 'failed' || outcome.status === 'skipped' || outcome.status === 'disabled') return undefined
  cached = await rendererAssetCache.getDataUri(url)
  return cached.hit ? cached.dataUri : undefined
}

async function readComposite(input: NormalizedCardThumbnailCompositeRequest): Promise<Buffer | null> {
  const key = compositeKey(input)
  try {
    const path = compositePath(input)
    const fileStat = await stat(path)
    if (!fileStat.isFile()) return null
    memoryComposites.add(key)
    return readFile(path)
  } catch {
    memoryComposites.delete(key)
    return null
  }
}

async function hasComposite(input: NormalizedCardThumbnailCompositeRequest): Promise<boolean> {
  const key = compositeKey(input)
  if (memoryComposites.has(key)) return true
  try {
    const fileStat = await stat(compositePath(input))
    if (!fileStat.isFile()) return false
    memoryComposites.add(key)
    return true
  } catch {
    return false
  }
}

function compositePath(input: NormalizedCardThumbnailCompositeRequest): string {
  return join(rendererAssetCache.cacheDir, `cardtile-${compositeKey(input)}.svg`)
}

function compositeKey(input: NormalizedCardThumbnailCompositeRequest): string {
  return createHash('sha256').update(JSON.stringify({ version: COMPOSITE_CACHE_VERSION, ...input })).digest('hex')
}

function groupCompositeRequestsByImageUrl(inputs: NormalizedCardThumbnailCompositeRequest[]): CompositeGroup[] {
  const map = new Map<string, NormalizedCardThumbnailCompositeRequest[]>()
  for (const input of inputs) {
    map.set(input.imageUrl, [...(map.get(input.imageUrl) ?? []), input])
  }
  return Array.from(map.entries()).map(([imageUrl, groupInputs]) => ({ imageUrl, inputs: groupInputs }))
}

function uniqueCompositeRequests(inputs: CardThumbnailCompositeRequest[]): NormalizedCardThumbnailCompositeRequest[] {
  const seen = new Set<string>()
  const result: NormalizedCardThumbnailCompositeRequest[] = []
  for (const input of inputs) {
    const normalized = normalizeCompositeRequest(input)
    if (!normalized) continue
    const key = JSON.stringify(normalized)
    if (seen.has(key)) continue
    seen.add(key)
    result.push(normalized)
  }
  return result
}

function normalizeCompositeRequest(input: CardThumbnailCompositeRequest): NormalizedCardThumbnailCompositeRequest | null {
  const imageUrl = typeof input.imageUrl === 'string' ? input.imageUrl.trim() : ''
  if (!/^https?:\/\//i.test(imageUrl)) return null
  return {
    imageUrl,
    rarity: typeof input.rarity === 'string' && input.rarity.trim() ? input.rarity.trim() : 'rarity_1',
    attr: typeof input.attr === 'string' && input.attr.trim() ? input.attr.trim() : 'cute',
    trained: Boolean(input.trained),
    size: normalizeCompositeSize(input.size),
  }
}

function buildCardThumbnailCompositeSvg(input: NormalizedCardThumbnailCompositeRequest, imageDataUri: string): string {
  const size = input.size
  const scale = size / DEFAULT_CARD_THUMBNAIL_SIZE
  const starCount = getRarityNumber(input.rarity)
  const birthday = input.rarity === 'rarity_birthday'
  const raritySuffix = birthday ? 'birthday' : String(starCount)
  const frameUrl = getSekaiCardUiAssetDataUri(`frame_rarity_${raritySuffix}.png`)
  const attrUrl = getSekaiCardUiAssetDataUri(`attr_${input.attr}.png`)
  const starUrl = getSekaiCardUiAssetDataUri(birthday ? 'rare_birthday.png' : input.trained ? 'rare_star_after_training.png' : 'rare_star_normal.png')
  const layers = [
    `<rect width="${size}" height="${size}" rx="${12 * scale}" fill="#f8fbff"/>`,
    `<image href="${escapeXmlAttr(imageDataUri)}" x="${2 * scale}" y="${2 * scale}" width="${152 * scale}" height="${152 * scale}" preserveAspectRatio="xMidYMid slice"/>`,
  ]

  if (frameUrl) layers.push(`<image href="${escapeXmlAttr(frameUrl)}" x="0" y="0" width="${size}" height="${size}" preserveAspectRatio="none"/>`)
  if (attrUrl) layers.push(`<image href="${escapeXmlAttr(attrUrl)}" x="0" y="0" width="${35 * scale}" height="${35 * scale}" preserveAspectRatio="xMidYMid meet"/>`)
  if (starUrl) {
    for (let index = 0; index < starCount; index += 1) {
      layers.push(`<image href="${escapeXmlAttr(starUrl)}" x="${(birthday ? 10 : 5 + index * 24) * scale}" y="${125 * scale}" width="${24 * scale}" height="${24 * scale}" preserveAspectRatio="xMidYMid meet"/>`)
    }
  }

  return `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 ${size} ${size}">${layers.join('')}</svg>`
}

function getRarityNumber(rarityType: string): number {
  if (rarityType === 'rarity_birthday') return 1
  const matched = rarityType.match(/(\d+)/)
  return matched ? Number(matched[1]) : 1
}

function normalizeCompositeSize(value: unknown): number {
  const size = typeof value === 'number' ? value : Number(value)
  if (!Number.isFinite(size) || size <= 0) return DEFAULT_CARD_THUMBNAIL_SIZE
  return Math.max(32, Math.min(320, Math.round(size)))
}

function normalizeConcurrency(value: unknown): number {
  const numberValue = typeof value === 'number' ? value : Number(value)
  if (!Number.isFinite(numberValue) || numberValue <= 0) return 10
  return Math.max(1, Math.min(32, Math.round(numberValue)))
}

async function writeFileAtomic(path: string, data: string): Promise<void> {
  const tmpPath = `${path}.${process.pid}.${Date.now()}.${Math.random().toString(16).slice(2)}.tmp`
  await writeFile(tmpPath, data)
  await rename(tmpPath, path).catch(async (error) => {
    await unlink(tmpPath).catch(() => {})
    throw error
  })
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

function countMemoryComposites(inputs: NormalizedCardThumbnailCompositeRequest[]): number {
  let count = 0
  for (const input of inputs) {
    if (memoryComposites.has(compositeKey(input))) count += 1
  }
  return count
}

function svgToDataUri(svg: string): string {
  return `data:image/svg+xml;base64,${Buffer.from(svg).toString('base64')}`
}

function parseSvgAttrs(value: string): Record<string, string> {
  const attrs: Record<string, string> = {}
  for (const match of value.matchAll(/([\w:-]+)="([^"]*)"/g)) {
    attrs[match[1] as string] = unescapeXmlAttr(match[2] as string)
  }
  return attrs
}

function numberAttr(value: string | undefined): number {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : 0
}

function optionalNumberAttr(value: string | undefined): number | undefined {
  if (value === undefined) return undefined
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : undefined
}

function escapeXmlAttr(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function unescapeXmlAttr(value: string): string {
  return value
    .replace(/&quot;/g, '"')
    .replace(/&gt;/g, '>')
    .replace(/&lt;/g, '<')
    .replace(/&amp;/g, '&')
}

import { createHash } from 'node:crypto'

// 渲染结果缓存：避免高峰期同一查询重复跑 satori + resvg 流水线。
//
// 设计要点：
// - 按模板分层 TTL：详情类近似永久（365d），列表类 10 分钟，用户相关 5 分钟，动态默认 10 秒
//   缓存键包含 data 哈希，所以即使 TTL 长，data 变化也会自动 miss，正确性由 data 保证
// - 容量按"字节预算 + 条数硬上限"双重控制，LRU 淘汰
// - 黑名单可通过环境变量 RENDER_CACHE_BLACKLIST=template1,template2 排除特定模板
// - 进程重启即清空，无需持久化

const SECOND = 1_000
const MINUTE = 60 * SECOND
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

// 详情类：单条数据，几乎不变
const TTL_DETAIL = 365 * DAY
// 列表类：随新增/编辑会变；data 哈希已经能反映变化，TTL 仅作回收兜底
const TTL_LIST = 10 * MINUTE
// 用户相关：用户卡组/档案变化频率中等
const TTL_USER = 5 * MINUTE
// 动态/默认：榜线、查水表等
const TTL_DYNAMIC = 10 * SECOND

const TEMPLATE_TTL: Record<string, number> = {
  // —— 详情类 ——
  card_detail: TTL_DETAIL,
  card: TTL_DETAIL,
  music_detail: TTL_DETAIL,
  music: TTL_DETAIL,
  chart_detail: TTL_DETAIL,
  chart: TTL_DETAIL,
  event_info: TTL_DETAIL,
  event: TTL_DETAIL,
  gacha_info: TTL_DETAIL,
  gacha: TTL_DETAIL,
  help_card: TTL_DETAIL,
  help: TTL_DETAIL,

  // —— 列表类 ——
  card_list: TTL_LIST,
  cards: TTL_LIST,
  music_list: TTL_LIST,
  musics: TTL_LIST,
  event_list: TTL_LIST,
  events: TTL_LIST,
  gacha_list: TTL_LIST,
  gachas: TTL_LIST,
  virtual_live_list: TTL_LIST,
  'virtual-lives': TTL_LIST,
  vlive: TTL_LIST,
  gallery_list: TTL_LIST,
  galleries: TTL_LIST,

  // —— 用户相关 ——
  profile_card: TTL_USER,
  profile: TTL_USER,
  suite_panel: TTL_USER,
  suite_status: TTL_USER,
  suite_card_box: TTL_USER,
  suite_cards: TTL_USER,
  best30: TTL_USER,
  b30: TTL_USER,
  deck_recommend: TTL_USER,
  'deck-recommend': TTL_USER,
  character_rank_mission: TTL_USER,
  cr_mission: TTL_USER,
  autochat_memory: TTL_USER,
  memory_card: TTL_USER,
  anvo_list: TTL_USER,
  anvo: TTL_USER,

  // —— 动态类显式声明（可省略，默认就是动态 TTL） ——
  ranking_list: TTL_DYNAMIC,
  ranking: TTL_DYNAMIC,
  churn_ranking_list: TTL_DYNAMIC,
  churn_ranking: TTL_DYNAMIC,
  forecast_ranking_list: TTL_DYNAMIC,
  forecast_ranking: TTL_DYNAMIC,
  water_table: TTL_DYNAMIC,
  csb: TTL_DYNAMIC,
}

const DEFAULT_TTL_MS = parsePositive(process.env.RENDER_CACHE_TTL_MS_DEFAULT, TTL_DYNAMIC)
let MAX_BYTES = parsePositive(process.env.RENDER_CACHE_MAX_BYTES, 256 * 1024 * 1024) // 256MB
let MAX_ENTRIES = parsePositive(process.env.RENDER_CACHE_MAX_ENTRIES, 1024)

// 上限的安全边界
const MIN_MAX_BYTES = 1024 * 1024 // 1MB
const HARD_MAX_BYTES = 4 * 1024 * 1024 * 1024 // 4GB
const MIN_MAX_ENTRIES = 16
const HARD_MAX_ENTRIES = 100_000

const BLACKLIST: Set<string> = new Set(
  (process.env.RENDER_CACHE_BLACKLIST ?? '')
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean),
)

function parsePositive(value: unknown, fallback: number): number {
  const n = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(n) && n > 0 ? n : fallback
}

function ttlForTemplate(template: string): number {
  return TEMPLATE_TTL[template] ?? DEFAULT_TTL_MS
}

export interface CachedRender {
  png: Buffer
  headers: Record<string, string>
  expiresAt: number
  bytes: number
}

const store = new Map<string, CachedRender>()
let totalBytes = 0
let hits = 0
let misses = 0
let evictions = 0

function makeKey(
  template: string,
  data: unknown,
  width: number | undefined,
  height: number | undefined,
  precision: number | undefined,
): string {
  const payload = JSON.stringify({
    t: template,
    d: data ?? null,
    w: width ?? 0,
    h: height ?? 0,
    p: precision ?? 0,
  })
  return createHash('sha256').update(payload).digest('hex')
}

export function isCacheable(template: string): boolean {
  if (!template) return false
  return !BLACKLIST.has(template)
}

export function getCachedRender(
  template: string,
  data: unknown,
  width?: number,
  height?: number,
  precision?: number,
): CachedRender | null {
  if (!isCacheable(template)) return null
  const key = makeKey(template, data, width, height, precision)
  const entry = store.get(key)
  if (!entry) {
    misses++
    return null
  }
  if (entry.expiresAt < Date.now()) {
    store.delete(key)
    totalBytes -= entry.bytes
    misses++
    return null
  }
  // LRU：命中后移到末尾（Map 保留插入顺序）
  store.delete(key)
  store.set(key, entry)
  hits++
  return entry
}

export function setCachedRender(
  template: string,
  data: unknown,
  width: number | undefined,
  height: number | undefined,
  precision: number | undefined,
  png: Buffer,
  headers: Record<string, string>,
  ttlMs?: number,
): void {
  if (!isCacheable(template)) return
  const effectiveTtl = ttlMs ?? ttlForTemplate(template)
  const key = makeKey(template, data, width, height, precision)

  // 若已有同键先扣除旧字节数
  const existing = store.get(key)
  if (existing) {
    totalBytes -= existing.bytes
    store.delete(key)
  }

  // 单条超过预算的一半就不缓存（避免一张大图把缓存挤爆）
  if (png.byteLength > MAX_BYTES / 2) return

  const entry: CachedRender = {
    png,
    headers: { ...headers },
    expiresAt: Date.now() + effectiveTtl,
    bytes: png.byteLength,
  }
  store.set(key, entry)
  totalBytes += entry.bytes

  // 按字节预算 + 条数硬上限淘汰最久未用的项
  while ((totalBytes > MAX_BYTES || store.size > MAX_ENTRIES) && store.size > 0) {
    const firstKey = store.keys().next().value
    if (firstKey === undefined) break
    const evicted = store.get(firstKey)
    store.delete(firstKey)
    if (evicted) totalBytes -= evicted.bytes
    evictions++
  }
}

export function renderCacheStats() {
  const total = hits + misses
  return {
    size: store.size,
    bytes: totalBytes,
    maxBytes: MAX_BYTES,
    byteUsageRatio: MAX_BYTES > 0 ? totalBytes / MAX_BYTES : 0,
    maxEntries: MAX_ENTRIES,
    limits: {
      minMaxBytes: MIN_MAX_BYTES,
      hardMaxBytes: HARD_MAX_BYTES,
      minMaxEntries: MIN_MAX_ENTRIES,
      hardMaxEntries: HARD_MAX_ENTRIES,
    },
    hits,
    misses,
    hitRate: total > 0 ? hits / total : 0,
    evictions,
    defaultTtlMs: DEFAULT_TTL_MS,
    tiers: {
      detailMs: TTL_DETAIL,
      listMs: TTL_LIST,
      userMs: TTL_USER,
      dynamicMs: TTL_DYNAMIC,
    },
    blacklist: [...BLACKLIST],
  }
}

export function clearRenderCache(): void {
  store.clear()
  totalBytes = 0
  hits = 0
  misses = 0
  evictions = 0
}

export interface RenderCacheConfigUpdate {
  maxBytes?: number
  maxEntries?: number
}

/**
 * 运行时调整缓存上限。会立即按新上限做一次 LRU 淘汰。
 * 返回应用后的实际上限（已 clamp 到安全边界）。
 */
export function updateRenderCacheConfig(update: RenderCacheConfigUpdate) {
  if (typeof update.maxBytes === 'number' && Number.isFinite(update.maxBytes) && update.maxBytes > 0) {
    MAX_BYTES = clamp(Math.floor(update.maxBytes), MIN_MAX_BYTES, HARD_MAX_BYTES)
  }
  if (typeof update.maxEntries === 'number' && Number.isFinite(update.maxEntries) && update.maxEntries > 0) {
    MAX_ENTRIES = clamp(Math.floor(update.maxEntries), MIN_MAX_ENTRIES, HARD_MAX_ENTRIES)
  }
  // 立即按新上限淘汰
  while ((totalBytes > MAX_BYTES || store.size > MAX_ENTRIES) && store.size > 0) {
    const firstKey = store.keys().next().value
    if (firstKey === undefined) break
    const evicted = store.get(firstKey)
    store.delete(firstKey)
    if (evicted) totalBytes -= evicted.bytes
    evictions++
  }
  return { maxBytes: MAX_BYTES, maxEntries: MAX_ENTRIES }
}

function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value))
}

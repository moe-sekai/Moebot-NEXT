import { fetchRemoteBytes, chartSvgRequestHeadersFor, hydrateSvgAssets, svgAssetFetcherForBase } from './svg-assets'
import { renderSvgToPngWithTrace } from './engine'
export interface ChartSvgRenderOptions {
  url?: string
  svg?: string
  width?: number
  precision?: number
}
export interface ChartSvgRenderTrace {
  png: Buffer
  timings: {
    resvgMs: number
    totalMs: number
  }
  sizeBytes: number
  width: number
  height: number
}
interface Viewport {
  width: number
  height: number
}
const DEFAULT_CHART_WIDTH = 5248
const DEFAULT_CHART_HEIGHT = 1920
const DEFAULT_CHART_PRECISION = 4
export async function renderChartSvg(options: ChartSvgRenderOptions): Promise<ChartSvgRenderTrace> {
  const totalStart = Date.now()
  const precision = normalizePrecision(options.precision)
  const source = await chartSource(options)
  const viewport = chartViewportFromSvg(source.svg, precision, options.width)
  const trace = await renderSvgToPngWithTrace(source.svg, {
    width: chartLogicalWidth(source.svg, options.width),
    precision,
  })
  return {
    png: trace.png,
    timings: {
      resvgMs: trace.timings.resvgMs,
      totalMs: Date.now() - totalStart,
    },
    sizeBytes: trace.sizeBytes,
    width: trace.width ?? viewport.width,
    height: viewport.height,
  }
}
export function chartViewportFromSvg(svg: string, precision: number, requestedWidth = 0): Viewport {
  const base = parseSvgSize(svg)
  const logicalWidth = requestedWidth > 0 ? requestedWidth : base.width
  const scale = logicalWidth / base.width * normalizePrecision(precision)
  return {
    width: Math.max(1, Math.round(base.width * scale)),
    height: Math.max(1, Math.round(base.height * scale)),
  }
}
export function chartLogicalWidth(svg: string, requestedWidth = 0): number {
  if (requestedWidth > 0) return requestedWidth
  return parseSvgSize(svg).width
}
async function chartSource(options: ChartSvgRenderOptions): Promise<{ svg: string; url?: string }> {
  const explicitSvg = typeof options.svg === 'string' && options.svg.trim() ? options.svg : ''
  const chartUrl = parseChartUrl(options.url)
  if (explicitSvg) {
    const hydrated = chartUrl
      ? await hydrateSvgAssets(explicitSvg, chartUrl.toString(), svgAssetFetcherForBase(chartUrl.toString()))
      : explicitSvg
    return { svg: prepareChartSvgForResvg(hydrated), ...(chartUrl ? { url: chartUrl.toString() } : {}) }
  }
  if (!chartUrl) {
    throw new Error('chart svg url is required')
  }
  const response = await fetchRemoteBytes(chartUrl.toString(), chartSvgRequestHeadersFor(chartUrl.toString()))
  const svg = response.data.toString('utf8')
  return {
    svg: prepareChartSvgForResvg(await hydrateSvgAssets(svg, chartUrl.toString(), svgAssetFetcherForBase(chartUrl.toString()))),
    url: chartUrl.toString(),
  }
}
export function prepareChartSvgForResvg(svg: string): string {
  return removeChartNoteSymbols(inlineChartNoteUsesWithAssets(normalizeChartSvgPaints(svg), extractChartNoteAssets(svg)))
}

function normalizeChartSvgPaints(svg: string): string {
  return svg
    .replaceAll('stop-color="var(--color-start)"', 'stop-color="#c9fce299"')
    .replaceAll('stop-color="var(--color-stop)"', 'stop-color="#c9fce233"')
    .replaceAll('<linearGradient id="decoration-critical-gradient" x1="0" x2="0" y1="1" y2="0"><stop offset="0" stop-color="#c9fce299" /><stop offset="1" stop-color="#c9fce233" /></linearGradient>', '<linearGradient id="decoration-critical-gradient" x1="0" x2="0" y1="1" y2="0"><stop offset="0" stop-color="#fcf1c399" /><stop offset="1" stop-color="#fcf1c333" /></linearGradient>')
    .replace(/class=(["'])decoration\1/g, 'class=$1decoration$1 fill="url(#decoration-gradient)"')
    .replace(/class=(["'])decoration-critical\1/g, 'class=$1decoration-critical$1 fill="url(#decoration-critical-gradient)"')
    .replace(/class=(["'])slide\1/g, 'class=$1slide$1 fill="#c9fce2cc"')
    .replace(/class=(["'])slide-critical\1/g, 'class=$1slide-critical$1 fill="#fcf1c3cc"')
}

interface ChartNoteAsset {
  href: string
  width: string
  height: string
  fallbackFill: string
  fallbackStroke: string
}

const NOTE_SYMBOL_RE = /<symbol\b(?=[^>]*\bid=(["'])notes-[^"']+\1)[\s\S]*?<\/symbol>/g
const NOTE_USE_RE = /<use\b(?=[^>]*(?:href|xlink:href)=(["'])#notes-[^"']+\1)[^>]*\/>/g
const NOTE_PALETTE = [
  { fill: '#dff8ff', stroke: '#69cfff' },
  { fill: '#ffe1f0', stroke: '#f08ac4' },
  { fill: '#dff8ec', stroke: '#68d7aa' },
  { fill: '#e6e4ff', stroke: '#9f9bea' },
  { fill: '#fff2ce', stroke: '#ffc858' },
  { fill: '#e9f3ff', stroke: '#7bbcff' },
  { fill: '#f3e6ff', stroke: '#c08cff' },
]

function extractChartNoteAssets(svg: string): Map<string, ChartNoteAsset> {
  const assets = new Map<string, ChartNoteAsset>()
  for (const match of svg.matchAll(NOTE_SYMBOL_RE)) {
    const symbol = match[0]
    const attrs = svgAttributes(symbol)
    const id = attrs.id
    const noteType = noteTypeFromID(id)
    if (noteType == null || !/^notes-\d+(?:-middle)?$/.test(id)) continue

    const image = symbol.match(/<image\b[^>]*\/>/i)?.[0]
    if (!image) continue
    const imageAttrs = svgAttributes(image)
    const href = imageAttrs.href ?? imageAttrs['xlink:href']
    if (!href) continue

    const fallback = NOTE_PALETTE[noteType] ?? NOTE_PALETTE[0]
    assets.set(id, {
      href,
      width: dimensionFromViewBox(attrs.viewBox, 2) ?? imageAttrs.width ?? '112',
      height: dimensionFromViewBox(attrs.viewBox, 3) ?? imageAttrs.height ?? '56',
      fallbackFill: fallback.fill,
      fallbackStroke: fallback.stroke,
    })
  }
  return assets
}

function inlineChartNoteUsesWithAssets(svg: string, assets: Map<string, ChartNoteAsset>): string {
  return svg.replace(NOTE_USE_RE, (tag) => {
    const attrs = svgAttributes(tag)
    const id = noteIDFromUseAttrs(attrs)
    const noteType = noteTypeFromID(id)
    const asset = assetForNoteID(id, assets)
    if (!asset) return noteFallbackRect(attrs, noteType)

    const x = attrs.x ?? '0'
    const y = attrs.y ?? '0'
    const width = attrs.width ?? asset.width
    const height = attrs.height ?? asset.height
    const transform = attrs.transform ? ` transform="${escapeSvgAttr(attrs.transform)}"` : ''
    const opacity = attrs.opacity ? ` opacity="${escapeSvgAttr(attrs.opacity)}"` : ''
    return `<image class="note-asset" x="${escapeSvgAttr(x)}" y="${escapeSvgAttr(y)}" width="${escapeSvgAttr(width)}" height="${escapeSvgAttr(height)}" preserveAspectRatio="none" href="${escapeSvgAttr(asset.href)}"${transform}${opacity} />`
  })
}

function removeChartNoteSymbols(svg: string): string {
  return svg.replace(NOTE_SYMBOL_RE, '')
}

function assetForNoteID(id: string, assets: Map<string, ChartNoteAsset>): ChartNoteAsset | undefined {
  if (assets.has(id)) return assets.get(id)
  const type = noteTypeFromID(id)
  if (type == null) return undefined
  return assets.get(`notes-${type}`) ?? assets.get(`notes-${type}-middle`)
}

function noteFallbackRect(attrs: Record<string, string>, noteType: number | null): string {
  const palette = noteType == null ? NOTE_PALETTE[0] : (NOTE_PALETTE[noteType] ?? NOTE_PALETTE[0])
  const transform = attrs.transform ? ` transform="${escapeSvgAttr(attrs.transform)}"` : ''
  return `<rect class="note-placeholder" x="${escapeSvgAttr(attrs.x ?? '0')}" y="${escapeSvgAttr(attrs.y ?? '0')}" width="${escapeSvgAttr(attrs.width ?? '18')}" height="${escapeSvgAttr(attrs.height ?? '18')}" rx="5" ry="5" fill="${palette.fill}" stroke="${palette.stroke}" stroke-width="2"${transform} />`
}

function noteIDFromUseAttrs(attrs: Record<string, string>): string {
  return (attrs.href ?? attrs['xlink:href'] ?? '').replace(/^#/, '')
}

function noteTypeFromID(id: string): number | null {
  const raw = id.match(/^notes-(\d+)/)?.[1]
  if (!raw) return null
  const value = Number(raw)
  return Number.isFinite(value) ? value : null
}

function dimensionFromViewBox(viewBox: string | undefined, index: number): string | undefined {
  const value = viewBox?.trim().split(/[\s,]+/)[index]
  return value && Number.isFinite(Number(value)) && Number(value) > 0 ? value : undefined
}

function svgAttributes(tag: string): Record<string, string> {
  const attrs: Record<string, string> = {}
  for (const match of tag.matchAll(/([\w:-]+)=(["'])(.*?)\2/g)) {
    attrs[match[1]] = match[3]
  }
  return attrs
}

function escapeSvgAttr(value: string): string {
  return value.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function parseChartUrl(value: string | undefined): URL | null {
  if (typeof value !== 'string' || !value.trim()) return null
  const chartUrl = new URL(value)
  if (chartUrl.protocol !== 'http:' && chartUrl.protocol !== 'https:') {
    throw new Error('chart svg url must be http(s)')
  }
  return chartUrl
}
function parseSvgSize(svg: string): Viewport {
  const root = svg.match(/<svg\b([^>]*)>/i)?.[1] ?? ''
  const width = parseSvgNumber(root.match(/\bwidth=["']([^"']+)["']/i)?.[1])
  const height = parseSvgNumber(root.match(/\bheight=["']([^"']+)["']/i)?.[1])
  if (width > 0 && height > 0) return { width, height }
  const viewBox = root.match(/\bviewBox=["']([^"']+)["']/i)?.[1]
  if (viewBox) {
    const parts = viewBox.trim().split(/[\s,]+/).map(Number)
    if (parts.length >= 4 && Number.isFinite(parts[2]) && Number.isFinite(parts[3]) && parts[2] > 0 && parts[3] > 0) {
      return { width: parts[2], height: parts[3] }
    }
  }
  return { width: DEFAULT_CHART_WIDTH, height: DEFAULT_CHART_HEIGHT }
}
function parseSvgNumber(value: string | undefined): number {
  if (!value) return 0
  const match = value.trim().match(/^([0-9.]+)/)
  return match ? Number(match[1]) : 0
}
function normalizePrecision(value: number | undefined): number {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : DEFAULT_CHART_PRECISION
}


import satori from 'satori'
import { Resvg } from '@resvg/resvg-js'
import { join } from 'path'
import { type ImageHydrationStats } from './asset-cache'
import { hydrateCachedImages } from './hydrate-images'
import { loadFonts, type FontData, FONT_FAMILY, fontPreferences } from './fonts'

// Resvg renders raw SVG strings (e.g. game chart SVGs that go through
// renderSvgToPngWithTrace) without going through Satori first, so it cannot
// rely on Satori's text-to-path conversion. We must point resvg at our
// bundled font directory so chart text such as the song title can render.
const RESVG_FONT_DIR = join(process.cwd(), 'assets', 'fonts')

function resvgFontConfig() {
  const primaryFamily = fontPreferences.body || FONT_FAMILY.body
  return {
    fontDirs: [RESVG_FONT_DIR],
    loadSystemFonts: true,
    defaultFontFamily: primaryFamily,
    serifFamily: primaryFamily,
    sansSerifFamily: primaryFamily,
    cursiveFamily: primaryFamily,
    fantasyFamily: primaryFamily,
    monospaceFamily: primaryFamily,
  }
}

export interface RenderOptions {
  width?: number
  height?: number
  precision?: number
  quality?: number
  debug?: boolean
}

export interface RenderTrace {
  svg: string
  png: Buffer
  timings: {
    fontsMs: number
    imagesMs: number
    satoriMs: number
    resvgMs: number
    totalMs: number
  }
  imageCache: ImageHydrationStats
  sizeBytes: number
  width: number
  height?: number
}

const DEFAULT_RENDER_PRECISION = 1.5

const DEFAULT_OPTIONS: RenderOptions = {
  width: 800,
  precision: DEFAULT_RENDER_PRECISION,
  quality: 80,
}

let fontsCache: FontData[] | null = null

async function getFonts(): Promise<FontData[]> {
  if (!fontsCache) {
    fontsCache = await loadFonts()
  }
  return fontsCache
}

function toSatoriFonts(fonts: FontData[]) {
  return fonts.map(f => ({
    name: f.name,
    data: f.data,
    weight: f.weight as any,
    style: f.style as any,
  }))
}

function normalizePrecision(value?: number): number {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : DEFAULT_RENDER_PRECISION
}

export interface SvgRenderTrace {
  png: Buffer
  timings: {
    resvgMs: number
    totalMs: number
  }
  sizeBytes: number
  width?: number
}

/**
 * Render a raw SVG string to PNG with resvg only.
 */
export async function renderSvgToPngWithTrace(svg: string, options: RenderOptions = {}): Promise<SvgRenderTrace> {
  const totalStart = Date.now()
  const resvgStart = Date.now()
  const precision = normalizePrecision(options.precision)
  const width = typeof options.width === 'number' && Number.isFinite(options.width) && options.width > 0
    ? Math.max(1, Math.round(options.width * precision))
    : 0
  const resvg = new Resvg(svg, {
    font: resvgFontConfig(),
    ...(width > 0 ? { fitTo: { mode: 'width' as const, value: width } } : {}),
  })
  const png = Buffer.from(resvg.render().asPng())
  const resvgMs = Date.now() - resvgStart
  return {
    png,
    timings: {
      resvgMs,
      totalMs: Date.now() - totalStart,
    },
    sizeBytes: png.length,
    ...(width > 0 ? { width } : {}),
  }
}

/**
 * Render JSX to SVG and PNG with timing metadata for debugging and preview pages.
 */
export async function renderWithTrace(
  element: any,
  options: RenderOptions = {},
): Promise<RenderTrace> {
  const opts = { ...DEFAULT_OPTIONS, ...options }
  const totalStart = Date.now()

  const fontsStart = Date.now()
  const fonts = await getFonts()
  const fontsMs = Date.now() - fontsStart

  const { element: hydratedElement, stats: imageCache, ms: imagesMs } = await hydrateCachedImages(element)

  const satoriStart = Date.now()
  // Height is intentionally omitted — Satori auto-computes height from content.
  // Only width is constrained; the card grows vertically to fit all content.
  const svg = await satori(hydratedElement, {
    width: opts.width!,
    fonts: toSatoriFonts(fonts),
    debug: opts.debug,
  })
  const satoriMs = Date.now() - satoriStart

  const resvgStart = Date.now()
  const precision = normalizePrecision(opts.precision)
  const outputWidth = Math.max(1, Math.round(opts.width! * precision))
  const resvg = new Resvg(svg, {
    fitTo: { mode: 'width', value: outputWidth },
  })
  const pngData = resvg.render()
  const png = Buffer.from(pngData.asPng())
  const resvgMs = Date.now() - resvgStart

  return {
    svg,
    png,
    timings: {
      fontsMs,
      imagesMs,
      satoriMs,
      resvgMs,
      totalMs: Date.now() - totalStart,
    },
    imageCache,
    sizeBytes: png.length,
    width: opts.width!,
    height: opts.height,
  }
}

/**
 * Render a JSX element to PNG buffer.
 * Pipeline: JSX -> Satori -> SVG string -> resvg -> PNG Buffer
 */
export async function renderToImage(
  element: any, // React JSX element (Satori compatible)
  options: RenderOptions = {},
): Promise<Buffer> {
  return (await renderWithTrace(element, options)).png
}

/**
 * Render JSX to SVG string only (for debugging)
 */
export async function renderToSvg(
  element: any,
  options: RenderOptions = {},
): Promise<string> {
  const opts = { ...DEFAULT_OPTIONS, ...options }
  const fonts = await getFonts()
  const { element: hydratedElement } = await hydrateCachedImages(element)

  // Height is intentionally omitted — auto-computed by Satori from content.
  return satori(hydratedElement, {
    width: opts.width!,
    fonts: toSatoriFonts(fonts),
    debug: opts.debug,
  })
}

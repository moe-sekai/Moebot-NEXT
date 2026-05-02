import satori from 'satori'
import { Resvg } from '@resvg/resvg-js'
import { loadFonts, type FontData } from './fonts'

export interface RenderOptions {
  width?: number
  height?: number
  quality?: number
  debug?: boolean
}

export interface RenderTrace {
  svg: string
  png: Buffer
  timings: {
    fontsMs: number
    satoriMs: number
    resvgMs: number
    totalMs: number
  }
  sizeBytes: number
  width: number
  height?: number
}

const DEFAULT_OPTIONS: RenderOptions = {
  width: 800,
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

  const satoriStart = Date.now()
  const svg = await satori(element, {
    width: opts.width!,
    height: opts.height,
    fonts: toSatoriFonts(fonts),
    debug: opts.debug,
  })
  const satoriMs = Date.now() - satoriStart

  const resvgStart = Date.now()
  const resvg = new Resvg(svg, {
    fitTo: { mode: 'width', value: opts.width! },
  })
  const pngData = resvg.render()
  const png = Buffer.from(pngData.asPng())
  const resvgMs = Date.now() - resvgStart

  return {
    svg,
    png,
    timings: {
      fontsMs,
      satoriMs,
      resvgMs,
      totalMs: Date.now() - totalStart,
    },
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

  return satori(element, {
    width: opts.width!,
    height: opts.height,
    fonts: toSatoriFonts(fonts),
    debug: opts.debug,
  })
}

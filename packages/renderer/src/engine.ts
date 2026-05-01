import satori from 'satori'
import { Resvg } from '@resvg/resvg-js'
import { loadFonts, type FontData } from './fonts'

export interface RenderOptions {
  width?: number
  height?: number
  quality?: number
  debug?: boolean
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

/**
 * Render a JSX element to PNG buffer.
 * Pipeline: JSX -> Satori -> SVG string -> resvg -> PNG Buffer
 */
export async function renderToImage(
  element: any, // React JSX element (Satori compatible)
  options: RenderOptions = {},
): Promise<Buffer> {
  const opts = { ...DEFAULT_OPTIONS, ...options }
  const fonts = await getFonts()

  // Step 1: JSX -> SVG via Satori
  const svg = await satori(element, {
    width: opts.width!,
    height: opts.height,
    fonts: fonts.map(f => ({
      name: f.name,
      data: f.data,
      weight: f.weight as any,
      style: f.style as any,
    })),
    debug: opts.debug,
  })

  // Step 2: SVG -> PNG via resvg
  const resvg = new Resvg(svg, {
    fitTo: { mode: 'width', value: opts.width! },
  })
  const pngData = resvg.render()
  return Buffer.from(pngData.asPng())
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
    fonts: fonts.map(f => ({
      name: f.name,
      data: f.data,
      weight: f.weight as any,
      style: f.style as any,
    })),
  })
}

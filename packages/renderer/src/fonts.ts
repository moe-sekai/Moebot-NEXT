import { readFile } from 'fs/promises'
import { join } from 'path'

export interface FontData {
  name: string
  data: ArrayBuffer
  weight: number
  style: 'normal' | 'italic'
}

const FONTS_DIR = join(process.cwd(), 'assets', 'fonts')

/**
 * Load fonts from assets/fonts directory.
 * Falls back to fetching Noto Sans CJK from Google Fonts CDN if local fonts not found.
 */
export async function loadFonts(): Promise<FontData[]> {
  const fonts: FontData[] = []

  // Try loading local font first
  try {
    const regularPath = join(FONTS_DIR, 'NotoSansCJKsc-Regular.otf')
    const boldPath = join(FONTS_DIR, 'NotoSansCJKsc-Bold.otf')

    const [regular, bold] = await Promise.allSettled([
      readFile(regularPath),
      readFile(boldPath),
    ])

    if (regular.status === 'fulfilled') {
      fonts.push({
        name: 'Noto Sans CJK SC',
        data: regular.value.buffer as ArrayBuffer,
        weight: 400,
        style: 'normal',
      })
    }

    if (bold.status === 'fulfilled') {
      fonts.push({
        name: 'Noto Sans CJK SC',
        data: bold.value.buffer as ArrayBuffer,
        weight: 700,
        style: 'normal',
      })
    }
  } catch {
    // Local fonts not available
  }

  // Fallback: fetch from CDN if no local fonts
  if (fonts.length === 0) {
    console.log('[renderer] No local fonts found, fetching from Google Fonts CDN...')
    try {
      const response = await fetch(
        'https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;700&display=swap'
      )
      // For production, download and bundle fonts locally
      // This is a fallback for development
      console.warn('[renderer] Please download Noto Sans CJK SC to assets/fonts/ for production use')

      // Use a simpler fallback - fetch the actual font file
      const fontResponse = await fetch(
        'https://cdn.jsdelivr.net/npm/@fontsource/noto-sans-sc@5.0.0/files/noto-sans-sc-chinese-simplified-400-normal.woff'
      )
      if (fontResponse.ok) {
        const buffer = await fontResponse.arrayBuffer()
        fonts.push({
          name: 'Noto Sans SC',
          data: buffer,
          weight: 400,
          style: 'normal',
        })
      }
    } catch (err) {
      console.error('[renderer] Failed to fetch fallback fonts:', err)
    }
  }

  if (fonts.length === 0) {
    throw new Error('[renderer] No fonts available. Please place font files in assets/fonts/')
  }

  console.log(`[renderer] Loaded ${fonts.length} font(s)`)
  return fonts
}

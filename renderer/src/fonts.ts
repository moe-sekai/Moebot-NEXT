import { readdir, readFile } from 'fs/promises'
import { join } from 'path'

export interface FontData {
  name: string
  data: ArrayBuffer
  weight: number
  style: 'normal' | 'italic'
}

const FONTS_DIR = join(process.cwd(), 'assets', 'fonts')
// Satori currently accepts TTF/OTF/WOFF font data, but not WOFF2.
// WOFF2 files are allowed to live in assets/fonts for humans, but must be
// skipped at runtime or Satori will fail with `Unsupported OpenType signature wOF2`.
const SUPPORTED_FONT_EXTENSIONS = new Set(['.otf', '.ttf', '.woff'])
const UNSUPPORTED_FONT_EXTENSIONS = new Set(['.woff2'])

/**
 * Load fonts from assets/fonts directory.
 * Falls back to fetching Noto Sans SC from CDN only if local fonts are absent.
 */
export async function loadFonts(): Promise<FontData[]> {
  const localFonts = await loadLocalFonts()
  if (localFonts.length > 0) {
    console.log(`[renderer] Loaded ${localFonts.length} local font(s) from ${FONTS_DIR}`)
    return localFonts
  }

  const fallbackFonts = await loadFallbackFonts()
  if (fallbackFonts.length === 0) {
    throw new Error('[renderer] No fonts available. Please place font files in renderer/assets/fonts/')
  }

  console.log(`[renderer] Loaded ${fallbackFonts.length} fallback font(s)`)
  return fallbackFonts
}

async function loadLocalFonts(): Promise<FontData[]> {
  const fonts: FontData[] = []

  try {
    const entries = await readdir(FONTS_DIR, { withFileTypes: true })
    const unsupportedFiles = entries
      .filter(entry => entry.isFile() && UNSUPPORTED_FONT_EXTENSIONS.has(fileExtension(entry.name)))
      .map(entry => entry.name)
      .sort((a, b) => a.localeCompare(b))
    if (unsupportedFiles.length > 0) {
      console.warn(`[renderer] Skipping unsupported WOFF2 font(s) because Satori only supports TTF/OTF/WOFF: ${unsupportedFiles.join(', ')}`)
    }

    const files = entries
      .filter(entry => entry.isFile() && SUPPORTED_FONT_EXTENSIONS.has(fileExtension(entry.name)))
      .map(entry => entry.name)
      .sort((a, b) => scoreFontFile(a) - scoreFontFile(b) || a.localeCompare(b))

    for (const fileName of files) {
      try {
        const data = await readFile(join(FONTS_DIR, fileName))
        fonts.push({
          name: inferFontFamily(fileName),
          data: data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength) as ArrayBuffer,
          weight: inferFontWeight(fileName),
          style: /italic|oblique/i.test(fileName) ? 'italic' : 'normal',
        })
      } catch (err) {
        console.warn(`[renderer] Failed to load local font ${fileName}:`, err)
      }
    }
  } catch {
    // Font directory not available; CDN fallback below will handle it.
  }

  return normalizeFontSet(fonts)
}

function normalizeFontSet(fonts: FontData[]): FontData[] {
  if (fonts.length === 0) return fonts

  const hasRegular = fonts.some(font => font.weight === 400 && font.style === 'normal')
  const hasBold = fonts.some(font => font.weight === 700 && font.style === 'normal')
  const firstNormal = fonts.find(font => font.style === 'normal') ?? fonts[0]
  const normalized = [...fonts]

  if (!hasRegular && firstNormal) {
    normalized.unshift({ ...firstNormal, weight: 400 })
  }
  if (!hasBold && firstNormal) {
    normalized.push({ ...firstNormal, weight: 700 })
  }

  return normalized
}

async function loadFallbackFonts(): Promise<FontData[]> {
  console.log('[renderer] No local fonts found, fetching Noto Sans SC from CDN...')
  console.warn('[renderer] Please keep renderer/assets/fonts populated for offline production use')

  const fonts: FontData[] = []
  try {
    const regular = await fetchFont(
      'https://cdn.jsdelivr.net/npm/@fontsource/noto-sans-sc@5.0.0/files/noto-sans-sc-chinese-simplified-400-normal.woff',
      400,
    )
    if (regular) fonts.push(regular)

    const bold = await fetchFont(
      'https://cdn.jsdelivr.net/npm/@fontsource/noto-sans-sc@5.0.0/files/noto-sans-sc-chinese-simplified-700-normal.woff',
      700,
    )
    if (bold) fonts.push(bold)
  } catch (err) {
    console.error('[renderer] Failed to fetch fallback fonts:', err)
  }

  return fonts
}

async function fetchFont(url: string, weight: number): Promise<FontData | null> {
  const response = await fetch(url)
  if (!response.ok) return null
  return {
    name: 'Noto Sans SC',
    data: await response.arrayBuffer(),
    weight,
    style: 'normal',
  }
}

function fileExtension(fileName: string): string {
  const index = fileName.lastIndexOf('.')
  return index >= 0 ? fileName.slice(index).toLowerCase() : ''
}

function inferFontFamily(fileName: string): string {
  const normalized = fileName.toLowerCase()
  if (normalized.includes('moebotscoresans')) return 'Moebot Score Sans'
  if (normalized.includes('noto')) return 'Noto Sans CJK SC'
  if (normalized.includes('plex')) return 'IBM Plex Sans'
  if (normalized.includes('yuruka')) return 'Yuruka Std'
  if (normalized.includes('fangtang') || normalized.includes('shangshou')) return 'ShangShou FangTangTi'
  if (normalized.includes('maoken')) return 'Maoken Assorted Sans'
  return fileName.replace(/\.(otf|ttf|woff2?)$/i, '')
}

function inferFontWeight(fileName: string): number {
  const normalized = fileName.toLowerCase()
  if (/black|heavy|900/.test(normalized)) return 900
  if (/extra[-_ ]?bold|800/.test(normalized)) return 800
  if (/bold|700/.test(normalized)) return 700
  if (/semi[-_ ]?bold|600/.test(normalized)) return 600
  if (/medium|500/.test(normalized)) return 500
  if (/light|300/.test(normalized)) return 300
  if (/thin|200/.test(normalized)) return 200
  return 400
}

function scoreFontFile(fileName: string): number {
  const normalized = fileName.toLowerCase()
  if (normalized.includes('moebotscoresans')) return 0
  if (normalized.includes('noto')) return 1
  if (normalized.includes('plex')) return 2
  if (normalized.includes('maoken')) return 3
  if (normalized.includes('fangtang') || normalized.includes('shangshou')) return 4
  if (normalized.includes('yuruka')) return 5
  return 10
}

import { readdir, readFile } from 'fs/promises'
import { join } from 'path'

export interface FontData {
  name: string
  data: ArrayBuffer
  weight: number
  style: 'normal' | 'italic'
}

// Well-known font family names used throughout the renderer.
// Components reference these constants for their fontFamily CSS values.
export const FONT_FAMILY = {
  /** Score/number display font — bold/black weight, used for PT scores */
  score: 'Moebot Score Sans',
  /** Primary CJK body font */
  body: 'LXGW WenKai Lite',
  /** Fallback body CJK font */
  bodyFallback: 'Noto Sans CJK SC',
  /** Decorative / hand-written style */
  decorative: 'Maoken Assorted Sans',
} as const

/**
 * Build a fontFamily CSS string with the given primary family followed by safe fallbacks.
 */
function buildFontFamilyChain(primary: string): string {
  const fallbacks = [FONT_FAMILY.bodyFallback, 'Noto Sans SC', 'sans-serif']
  const seen = new Set<string>()
  const parts: string[] = []
  for (const part of [primary, ...fallbacks]) {
    const trimmed = part.trim()
    if (!trimmed || seen.has(trimmed)) continue
    seen.add(trimmed)
    parts.push(trimmed)
  }
  return parts.join(', ')
}

/**
 * Currently selected primary font families. Mutable so the running renderer can
 * switch fonts via POST /fonts without restarting.
 */
export const fontPreferences = {
  body: process.env.RENDER_FONT_BODY?.trim() || FONT_FAMILY.body,
  score: process.env.RENDER_FONT_SCORE?.trim() || FONT_FAMILY.score,
}

/**
 * Default fontFamily value for card body text.
 * Prioritizes the configured primary family, falls back to system CJK fonts.
 * Mutable (export let) so changes propagate via ESM live bindings.
 */
export let defaultFontFamily = buildFontFamilyChain(fontPreferences.body)

/**
 * fontFamily value used for rendering numeric PT scores (ranking cutoff lines, etc.).
 * Uses a black-weight sans-serif (黑体) for clarity.
 */
export let scoreFontFamily = buildFontFamilyChain(fontPreferences.score)

/**
 * Update the default fontFamily strings used by templates. Pass undefined/empty
 * to leave that preference unchanged.
 */
export function setFontPreferences(body?: string | null, score?: string | null): void {
  if (typeof body === 'string' && body.trim()) {
    fontPreferences.body = body.trim()
    defaultFontFamily = buildFontFamilyChain(fontPreferences.body)
  }
  if (typeof score === 'string' && score.trim()) {
    fontPreferences.score = score.trim()
    scoreFontFamily = buildFontFamilyChain(fontPreferences.score)
  }
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
  if (normalized.includes('moebotscoresans')) return FONT_FAMILY.score
  if (normalized.includes('lxgw') || normalized.includes('wenkai')) return FONT_FAMILY.body
  if (normalized.includes('noto')) return FONT_FAMILY.bodyFallback
  if (normalized.includes('plex')) return 'IBM Plex Sans'
  if (normalized.includes('yuruka')) return 'Yuruka Std'
  if (normalized.includes('fangtang') || normalized.includes('shangshou')) return 'ShangShou FangTangTi'
  if (normalized.includes('maoken')) return FONT_FAMILY.decorative
  // Custom fonts: use filename (without extension) as family name
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
  if (normalized.includes('lxgw') || normalized.includes('wenkai')) return 1
  if (normalized.includes('noto')) return 2
  if (normalized.includes('plex')) return 3
  if (normalized.includes('maoken')) return 4
  if (normalized.includes('fangtang') || normalized.includes('shangshou')) return 5
  if (normalized.includes('yuruka')) return 6
  return 10
}

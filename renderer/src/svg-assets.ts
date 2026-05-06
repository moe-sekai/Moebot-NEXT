import { spawn } from 'node:child_process'
import { rendererAssetCache } from './asset-cache'

export interface SvgAssetFetchResult {
  data: Buffer
  mime: string
}

export type SvgAssetFetcher = (url: string) => Promise<SvgAssetFetchResult>

export interface RemoteBytesResult {
  data: Buffer
  contentType: string | null
}

export type RequestHeaders = Record<string, string>

const SVG_ASSET_REF_RE = /\b(href|xlink:href)=(['"])([^'"]+)\2/g
const DEFAULT_SVG_ASSET_TIMEOUT_MS = 12_000

const defaultChartReferer = 'https://charts-new.unipjsk.com/'

export const chartSvgRequestHeaders = {
  accept: 'image/svg+xml,text/plain;q=0.9,*/*;q=0.8',
  'user-agent': 'Mozilla/5.0 (compatible; Moebot-NEXT Renderer)',
  referer: defaultChartReferer,
}

export const chartAssetRequestHeaders = {
  accept: 'image/avif,image/webp,image/png,image/jpeg,image/svg+xml,*/*;q=0.8',
  'user-agent': chartSvgRequestHeaders['user-agent'],
  referer: defaultChartReferer,
}

export function chartSvgRequestHeadersFor(baseUrl: string): RequestHeaders {
  return requestHeadersWithReferer(chartSvgRequestHeaders, refererForSvgBase(baseUrl))
}

export function chartAssetRequestHeadersFor(baseUrl: string): RequestHeaders {
  return requestHeadersWithReferer(chartAssetRequestHeaders, refererForSvgBase(baseUrl))
}

export function svgAssetFetcherForBase(baseUrl: string): SvgAssetFetcher {
  return async (url: string) => {
    const cached = await rendererAssetCache.getBytes(url)
    if (cached.hit && cached.data) {
      return { data: cached.data, mime: mimeFromResponse(url, null) }
    }

    const result = await fetchRemoteBytes(url, chartAssetRequestHeadersFor(baseUrl))
    void rendererAssetCache.prefetch(url).catch((error) => {
      console.warn(`[renderer] failed to persist svg asset cache ${url}:`, error)
    })
    return { data: result.data, mime: mimeFromResponse(url, result.contentType) }
  }
}

export const fixedChartNoteAssetUrls = buildFixedChartNoteAssetUrls()

export async function preloadFixedChartNoteAssets(): Promise<void> {
  await rendererAssetCache.startPreload(fixedChartNoteAssetUrls, { concurrency: 8 })
}

export async function hydrateSvgAssets(
  svg: string,
  baseUrl: string,
  fetchAsset: SvgAssetFetcher = fetchSvgAsset,
): Promise<string> {
  if (!svg || !baseUrl) return svg

  const replacements = new Map<string, string>()
  const refsByRaw = collectSvgAssetRefs(svg, baseUrl)

  await Promise.all(Array.from(refsByRaw.entries()).map(async ([raw, absoluteUrl]) => {
    try {
      const asset = await fetchAsset(absoluteUrl)
      if (asset.data.length === 0) return
      replacements.set(raw, `data:${asset.mime};base64,${asset.data.toString('base64')}`)
    } catch (error) {
      console.warn(`[renderer] failed to inline svg asset ${absoluteUrl}:`, error)
    }
  }))

  if (replacements.size === 0) return svg

  return svg.replace(SVG_ASSET_REF_RE, (full, attr: string, quote: string, raw: string) => {
    const replacement = replacements.get(raw)
    return replacement ? `${attr}=${quote}${replacement}${quote}` : full
  })
}

export function collectSvgAssetRefs(svg: string, baseUrl: string): Map<string, string> {
  const refsByRaw = new Map<string, string>()
  if (!svg || !baseUrl) return refsByRaw
  for (const match of svg.matchAll(SVG_ASSET_REF_RE)) {
    const raw = match[3]
    if (!raw || shouldSkipSvgAssetRef(raw)) continue
    const absoluteUrl = resolveSvgAssetUrl(raw, baseUrl)
    if (absoluteUrl) refsByRaw.set(raw, absoluteUrl)
  }
  return refsByRaw
}

function shouldSkipSvgAssetRef(value: string): boolean {
  const trimmed = value.trim()
  return trimmed === ''
    || trimmed.startsWith('#')
    || /^data:/i.test(trimmed)
    || /^blob:/i.test(trimmed)
    || /^javascript:/i.test(trimmed)
}

function resolveSvgAssetUrl(value: string, baseUrl: string): string | null {
  try {
    const url = new URL(value, baseUrl)
    return url.protocol === 'http:' || url.protocol === 'https:' ? url.toString() : null
  } catch {
    return null
  }
}

async function fetchSvgAsset(url: string): Promise<SvgAssetFetchResult> {
  const cached = await rendererAssetCache.getBytes(url)
  if (cached.hit && cached.data) {
    return { data: cached.data, mime: mimeFromResponse(url, null) }
  }

  const result = await fetchRemoteBytes(url, chartAssetRequestHeaders)
  void rendererAssetCache.prefetch(url).catch((error) => {
    console.warn(`[renderer] failed to persist svg asset cache ${url}:`, error)
  })
  return { data: result.data, mime: mimeFromResponse(url, result.contentType) }
}

function buildFixedChartNoteAssetUrls(): string[] {
  const base = `${defaultChartReferer}moe/notes_new/custom01`
  const names = [
    ...Array.from({ length: 7 }, (_, index) => `notes_${index}.png`),
    'notes_long_among.png',
    'notes_friction_among_long.png',
    'notes_friction_among_crtcl.png',
    ...Array.from({ length: 7 }, (_, index) => `notes_flick_arrow_${String(index).padStart(2, '0')}.png`),
    ...Array.from({ length: 7 }, (_, index) => `notes_flick_arrow_${String(index).padStart(2, '0')}_diagonal.png`),
    ...Array.from({ length: 7 }, (_, index) => `notes_flick_arrow_crtcl_${String(index).padStart(2, '0')}.png`),
    ...Array.from({ length: 7 }, (_, index) => `notes_flick_arrow_crtcl_${String(index).padStart(2, '0')}_diagonal.png`),
  ]
  return Array.from(new Set(names.map((name) => `${base}/${name}`)))
}

function requestHeadersWithReferer(headers: RequestHeaders, referer: string): RequestHeaders {
  return { ...headers, referer }
}

function refererForSvgBase(baseUrl: string): string {
  try {
    const url = new URL(baseUrl)
    return `${url.protocol}//${url.host}/`
  } catch {
    return defaultChartReferer
  }
}

export async function fetchRemoteBytes(
  url: string,
  headers: RequestHeaders,
  options: {
    primaryFetch?: (url: string, headers: RequestHeaders) => Promise<RemoteBytesResult>
    fallbackFetch?: (url: string, headers: RequestHeaders) => Promise<RemoteBytesResult>
  } = {},
): Promise<RemoteBytesResult> {
  const primaryFetch = options.primaryFetch ?? fetchRemoteBytesWithBun
  const fallbackFetch = options.fallbackFetch ?? fetchRemoteBytesWithCurl
  try {
    return await primaryFetch(url, headers)
  } catch (error) {
    console.warn(`[renderer] primary fetch failed for ${url}, fallback to curl:`, error)
    return fallbackFetch(url, headers)
  }
}

async function fetchRemoteBytesWithBun(url: string, headers: RequestHeaders): Promise<RemoteBytesResult> {
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), DEFAULT_SVG_ASSET_TIMEOUT_MS)
  try {
    const response = await fetch(url, { signal: controller.signal, headers })
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }
    return {
      data: Buffer.from(await response.arrayBuffer()),
      contentType: response.headers.get('content-type'),
    }
  } finally {
    clearTimeout(timeout)
  }
}

async function fetchRemoteBytesWithCurl(url: string, headers: RequestHeaders): Promise<RemoteBytesResult> {
  const args = [
    '-L',
    '--fail',
    '--silent',
    '--show-error',
    '--max-time',
    String(Math.ceil(DEFAULT_SVG_ASSET_TIMEOUT_MS / 1000)),
    '-H',
    `Accept: ${headers.accept ?? '*/*'}`,
    '-A',
    headers['user-agent'] ?? 'Mozilla/5.0',
  ]
  if (headers.referer) {
    args.push('-e', headers.referer)
  }
  args.push('-w', '\n%{content_type}', url)

  const output = await runCurl(args)
  const marker = output.lastIndexOf('\n')
  if (marker < 0) return { data: output, contentType: null }
  return {
    data: output.subarray(0, marker),
    contentType: output.subarray(marker + 1).toString('utf8').trim() || null,
  }
}

function runCurl(args: string[]): Promise<Buffer> {
  return new Promise((resolve, reject) => {
    const child = spawn('curl', args, { stdio: ['ignore', 'pipe', 'pipe'] })
    const stdout: Buffer[] = []
    const stderr: Buffer[] = []
    child.stdout.on('data', (chunk) => stdout.push(Buffer.from(chunk)))
    child.stderr.on('data', (chunk) => stderr.push(Buffer.from(chunk)))
    child.on('error', reject)
    child.on('close', (code) => {
      if (code === 0) {
        resolve(Buffer.concat(stdout))
        return
      }
      reject(new Error(Buffer.concat(stderr).toString('utf8') || `curl exited with ${code}`))
    })
  })
}

function mimeFromResponse(url: string, contentType: string | null): string {
  const cleanType = contentType?.split(';')[0]?.trim()
  if (cleanType?.startsWith('image/')) return cleanType

  const pathname = new URL(url).pathname.toLowerCase()
  if (pathname.endsWith('.svg')) return 'image/svg+xml'
  if (pathname.endsWith('.jpg') || pathname.endsWith('.jpeg')) return 'image/jpeg'
  if (pathname.endsWith('.webp')) return 'image/webp'
  if (pathname.endsWith('.gif')) return 'image/gif'
  return 'image/png'
}

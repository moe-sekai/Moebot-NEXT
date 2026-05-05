import { mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { spawn } from 'node:child_process'
import { fetchRemoteBytes, chartSvgRequestHeaders } from './svg-assets'

export interface ChartBrowserRenderOptions {
  url?: string
  svg?: string
  width?: number
  precision?: number
  chromeTimeoutMs?: number
}

export interface ChartBrowserRenderTrace {
  png: Buffer
  timings: {
    chromeMs: number
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
const DEFAULT_CHROME_TIMEOUT_MS = 45_000

export async function renderChartWithBrowser(options: ChartBrowserRenderOptions): Promise<ChartBrowserRenderTrace> {
  const totalStart = Date.now()
  const precision = normalizePrecision(options.precision)
  const source = await chartSource(options)
  const viewport = chartViewportFromSvg(source.svg, precision, options.width)
  const tempDir = await mkdtemp(join(tmpdir(), 'moebot-chart-'))
  const outPath = join(tempDir, 'chart.png')

  try {
    const target = await chromeTargetFromSvg(source.svg, tempDir)
    const chromeStart = Date.now()
    await runChrome(chromeArgsForChart(target, outPath, viewport), normalizeChromeTimeoutMs(options.chromeTimeoutMs))
    const chromeMs = Date.now() - chromeStart
    const png = await readFile(outPath)
    return {
      png,
      timings: { chromeMs, totalMs: Date.now() - totalStart },
      sizeBytes: png.length,
      width: viewport.width,
      height: viewport.height,
    }
  } finally {
    await rm(tempDir, { recursive: true, force: true }).catch(() => {})
  }
}

export function chartViewportFromSvg(svg: string, precision: number, requestedWidth = 0): Viewport {
  const base = parseSvgSize(svg)
  const scale = requestedWidth > 0 ? requestedWidth / base.width : normalizePrecision(precision)
  return {
    width: Math.max(1, Math.round(base.width * scale)),
    height: Math.max(1, Math.round(base.height * scale)),
  }
}

export async function chromeTargetFromSvg(svg: string, tempDir: string): Promise<string> {
  await mkdir(tempDir, { recursive: true })
  const svgPath = join(tempDir, 'chart.svg')
  await writeFile(svgPath, svg)
  return `file://${svgPath}`
}

export function chromeArgsForChart(targetUrl: string, outputPath: string, viewport: Viewport): string[] {
  return [
    '--headless=new',
    '--disable-gpu',
    '--no-sandbox',
    '--hide-scrollbars',
    '--force-device-scale-factor=1',
    `--window-size=${viewport.width},${viewport.height}`,
    `--screenshot=${outputPath}`,
    targetUrl,
  ]
}

async function chartSource(options: ChartBrowserRenderOptions): Promise<{ svg: string; url?: string }> {
  if (typeof options.svg === 'string' && options.svg.trim()) {
    return { svg: options.svg }
  }
  if (typeof options.url !== 'string' || !options.url.trim()) {
    throw new Error('chart svg url is required')
  }
  const chartUrl = new URL(options.url)
  if (chartUrl.protocol !== 'http:' && chartUrl.protocol !== 'https:') {
    throw new Error('chart svg url must be http(s)')
  }
  const response = await fetchRemoteBytes(chartUrl.toString(), chartSvgRequestHeaders)
  return { svg: response.data.toString('utf8'), url: chartUrl.toString() }
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

export function normalizeChromeTimeoutMs(value: number | undefined): number {
  if (typeof value !== 'number' || !Number.isFinite(value) || value <= 0) return DEFAULT_CHROME_TIMEOUT_MS
  return Math.max(1_000, Math.round(value))
}

function runChrome(args: string[], timeoutMs: number): Promise<void> {
  return new Promise((resolve, reject) => {
    const child = spawn(process.env.CHROME_PATH || 'google-chrome', args, { stdio: ['ignore', 'ignore', 'pipe'] })
    const stderr: Buffer[] = []
    let settled = false
    const timer = setTimeout(() => {
      if (settled) return
      settled = true
      child.kill('SIGKILL')
      reject(new Error(`chrome screenshot timed out after ${timeoutMs}ms`))
    }, timeoutMs)

    child.stderr.on('data', (chunk) => {
      if (stderr.reduce((sum, item) => sum + item.length, 0) < 64_000) {
        stderr.push(Buffer.from(chunk))
      }
    })
    child.on('error', (error) => {
      if (settled) return
      settled = true
      clearTimeout(timer)
      reject(error)
    })
    child.on('close', (code) => {
      if (settled) return
      settled = true
      clearTimeout(timer)
      if (code === 0) {
        resolve()
        return
      }
      reject(new Error(Buffer.concat(stderr).toString('utf8') || `chrome exited with ${code}`))
    })
  })
}

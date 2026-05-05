import { describe, expect, test } from 'bun:test'
import { chartViewportFromSvg, chromeArgsForChart, chromeTargetFromSvg, normalizeChromeTimeoutMs } from './chart-browser-renderer'

describe('chart browser renderer', () => {
  test('scales svg dimensions by precision', () => {
    const viewport = chartViewportFromSvg('<svg width="5248" height="1920"></svg>', 4)
    expect(viewport).toEqual({ width: 20992, height: 7680 })
  })

  test('uses requested width when provided', () => {
    const viewport = chartViewportFromSvg('<svg width="5248" height="1920"></svg>', 4, 1600)
    expect(viewport).toEqual({ width: 1600, height: 585 })
  })

  test('chrome args include screenshot output and target url', () => {
    const args = chromeArgsForChart('file:///tmp/chart.svg', '/tmp/out.png', { width: 800, height: 600 })
    expect(args).toContain('--headless=new')
    expect(args).toContain('--no-sandbox')
    expect(args).toContain('--screenshot=/tmp/out.png')
    expect(args).toContain('--window-size=800,600')
    expect(args.at(-1)).toBe('file:///tmp/chart.svg')
  })

  test('always screenshots a local temporary svg file', async () => {
    const tempDir = '/tmp/moebot-chart-unit'
    const target = await chromeTargetFromSvg('<svg width="1" height="1"></svg>', tempDir)
    expect(target).toBe('file:///tmp/moebot-chart-unit/chart.svg')
  })

  test('normalizes chrome timeout with safe default', () => {
    expect(normalizeChromeTimeoutMs(undefined)).toBe(45000)
    expect(normalizeChromeTimeoutMs(500)).toBe(1000)
    expect(normalizeChromeTimeoutMs(120000)).toBe(120000)
  })
})

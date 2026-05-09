import { describe, expect, test } from 'bun:test'
import { chartLogicalWidth, chartViewportFromSvg, prepareChartSvgForResvg, rewriteChartSvgFontFamilies } from './chart-svg-renderer'

describe('chart svg renderer helpers', () => {
  test('scales svg dimensions by precision', () => {
    const viewport = chartViewportFromSvg('<svg width="5248" height="1920"></svg>', 4)
    expect(viewport).toEqual({ width: 20992, height: 7680 })
  })

  test('uses requested logical width before precision scaling', () => {
    const viewport = chartViewportFromSvg('<svg width="5248" height="1920"></svg>', 4, 1600)
    expect(viewport).toEqual({ width: 6400, height: 2341 })
  })

  test('returns original svg width when no requested width is provided', () => {
    expect(chartLogicalWidth('<svg width="3616" height="2227"></svg>')).toBe(3616)
    expect(chartLogicalWidth('<svg viewBox="0 0 123 456"></svg>')).toBe(123)
  })

  test('normalizes chart slide and decoration paints for resvg', () => {
    const svg = '<svg><defs><linearGradient id="decoration-gradient" x1="0" x2="0" y1="1" y2="0"><stop offset="0" stop-color="var(--color-start)" /><stop offset="1" stop-color="var(--color-stop)" /></linearGradient><linearGradient id="decoration-critical-gradient" x1="0" x2="0" y1="1" y2="0"><stop offset="0" stop-color="var(--color-start)" /><stop offset="1" stop-color="var(--color-stop)" /></linearGradient></defs><path class="slide" d="M0 0"/><path class="slide-critical" d="M0 0"/><path class="decoration" d="M0 0"/><path class="decoration-critical" d="M0 0"/></svg>'
    const prepared = prepareChartSvgForResvg(svg)

    expect(prepared).toContain('class="slide" fill="#c9fce2cc"')
    expect(prepared).toContain('class="slide-critical" fill="#fcf1c3cc"')
    expect(prepared).toContain('class="decoration" fill="url(#decoration-gradient)"')
    expect(prepared).toContain('class="decoration-critical" fill="url(#decoration-critical-gradient)"')
    expect(prepared).toContain('stop-color="#c9fce299"')
    expect(prepared).toContain('stop-color="#fcf1c399"')
    expect(prepared).not.toContain('var(--color-start)')
  })

  test('inlines note symbol uses with original note assets for resvg', () => {
    const svg = '<svg><defs><symbol id="notes-2" viewBox="0 0 112 56"><image width="118" height="62" href="data:image/png;base64,note2"/></symbol><symbol id="notes-2-3" viewBox="0 0 64 18"><use href="#notes-2"/></symbol></defs><use x="10" y="20" width="30" height="40" xlink:href="#notes-2-3"/><use href="#other"/></svg>'
    const prepared = prepareChartSvgForResvg(svg)

    expect(prepared).toContain('<image class="note-asset" x="10" y="20" width="30" height="40"')
    expect(prepared).toContain('href="data:image/png;base64,note2"')
    expect(prepared).toContain('<use href="#other"/>')
    expect(prepared).not.toContain('xlink:href="#notes-2-3"')
    expect(prepared).not.toContain('<symbol id="notes-2"')
  })

  test('rewrites unavailable chart svg font-family declarations to the bundled body font', () => {
    const svg = '<svg><defs><style>.title { font-family: ヒラギノ角ゴシック; } .subtitle { font-family: FOT-RodinNTLG Pro DB; } .speed-text { font-family: Avenir; }</style></defs><text font-family="ヒラギノ角ゴシック">千本桜</text></svg>'
    const rewritten = rewriteChartSvgFontFamilies(svg)

    expect(rewritten).not.toContain('ヒラギノ角ゴシック')
    expect(rewritten).not.toContain('FOT-RodinNTLG Pro DB')
    expect(rewritten).not.toContain('Avenir')
    expect(rewritten).toContain("'LXGW WenKai Lite'")
    expect(rewritten).toContain('font-family="')
    // attribute substitution should keep the same quoting style
    expect(rewritten).toMatch(/<text font-family="[^"]+">千本桜<\/text>/)
  })

  test('chart svg preparation rewrites font-family along with other normalisations', () => {
    const svg = '<svg><defs><style>.title { font-family: ヒラギノ角ゴシック; }</style></defs><text class="title">タイトル</text></svg>'
    const prepared = prepareChartSvgForResvg(svg)
    expect(prepared).not.toContain('ヒラギノ角ゴシック')
    expect(prepared).toContain("'LXGW WenKai Lite'")
  })
})

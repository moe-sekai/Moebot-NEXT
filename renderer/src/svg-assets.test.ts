import { describe, expect, test } from 'bun:test'
import { chartAssetRequestHeaders, chartAssetRequestHeadersFor, chartSvgRequestHeaders, chartSvgRequestHeadersFor, collectSvgAssetRefs, fetchRemoteBytes, fixedChartNoteAssetUrls, hydrateSvgAssets } from './svg-assets'

const onePixelPng = Buffer.from('iVBORw0KGgo=', 'base64')

describe('hydrateSvgAssets', () => {
  test('inlines relative image hrefs using the source svg url', async () => {
    const requested: string[] = []
    const svg = '<svg xmlns="http://www.w3.org/2000/svg"><image href="assets/note.png" width="10" height="10"/></svg>'

    const hydrated = await hydrateSvgAssets(svg, 'https://charts.example.test/moe/svg/739/master.svg', async (url) => {
      requested.push(url)
      return { data: onePixelPng, mime: 'image/png' }
    })

    expect(requested).toEqual(['https://charts.example.test/moe/svg/739/assets/note.png'])
    expect(hydrated).toContain('href="data:image/png;base64,iVBORw0KGgo="')
  })

  test('inlines xlink hrefs and keeps existing data uris unchanged', async () => {
    const requested: string[] = []
    const svg = '<svg xmlns:xlink="http://www.w3.org/1999/xlink"><image xlink:href="/assets/hold.png"/><image href="data:image/png;base64,abc"/></svg>'

    const hydrated = await hydrateSvgAssets(svg, 'https://charts.example.test/moe/svg/739/master.svg', async (url) => {
      requested.push(url)
      return { data: onePixelPng, mime: 'image/png' }
    })

    expect(requested).toEqual(['https://charts.example.test/assets/hold.png'])
    expect(hydrated).toContain('xlink:href="data:image/png;base64,iVBORw0KGgo="')
    expect(hydrated).toContain('href="data:image/png;base64,abc"')
  })

  test('uses browser-like headers for unipjsk chart svg and assets', () => {
    expect(chartSvgRequestHeaders['user-agent']).toContain('Mozilla/5.0')
    expect(chartSvgRequestHeaders.referer).toBe('https://charts-new.unipjsk.com/')
    expect(chartAssetRequestHeaders['user-agent']).toContain('Mozilla/5.0')
    expect(chartAssetRequestHeaders.referer).toBe('https://charts-new.unipjsk.com/')
    expect(chartSvgRequestHeadersFor('https://charts.example.test/moe/svg/739/master.svg').referer).toBe('https://charts.example.test/')
    expect(chartAssetRequestHeadersFor('https://charts.example.test/moe/svg/739/master.svg').referer).toBe('https://charts.example.test/')
  })

  test('fetchRemoteBytes falls back when primary fetch fails', async () => {
    const calls: string[] = []
    const result = await fetchRemoteBytes('https://charts.example.test/1.svg', chartSvgRequestHeaders, {
      primaryFetch: async () => {
        calls.push('primary')
        throw new Error('certificate failed')
      },
      fallbackFetch: async (url, headers) => {
        calls.push(`${url}:${headers.referer}`)
        return { data: Buffer.from('<svg/>'), contentType: 'image/svg+xml' }
      },
    })

    expect(calls).toEqual(['primary', 'https://charts.example.test/1.svg:https://charts-new.unipjsk.com/'])
    expect(result.data.toString()).toBe('<svg/>')
    expect(result.contentType).toBe('image/svg+xml')
  })

  test('fetches each repeated asset reference only once', async () => {
    const requested: string[] = []
    const svg = '<svg><image href="notes.png"/><image href="notes.png"/><image xlink:href="notes.png"/></svg>'

    const hydrated = await hydrateSvgAssets(svg, 'https://charts.example.test/moe/svg/1/master.svg', async (url) => {
      requested.push(url)
      return { data: onePixelPng, mime: 'image/png' }
    })

    expect(requested).toEqual(['https://charts.example.test/moe/svg/1/notes.png'])
    expect(hydrated.match(/data:image\/png;base64/g)?.length).toBe(3)
  })

  test('resolves unipjsk note assets relative to the source svg url', () => {
    const refs = collectSvgAssetRefs(
      '<svg><image xlink:href="../../notes_new/custom01/notes_0.png"/><use href="#notes-0"/></svg>',
      'https://charts-new.unipjsk.com/moe/svg/739/master.svg',
    )

    expect(refs.get('../../notes_new/custom01/notes_0.png')).toBe('https://charts-new.unipjsk.com/moe/notes_new/custom01/notes_0.png')
    expect(refs.has('#notes-0')).toBe(false)
  })

  test('predefines fixed unipjsk note assets for local preload cache', () => {
    expect(fixedChartNoteAssetUrls).toContain('https://charts-new.unipjsk.com/moe/notes_new/custom01/notes_2.png')
    expect(fixedChartNoteAssetUrls).toContain('https://charts-new.unipjsk.com/moe/notes_new/custom01/notes_flick_arrow_crtcl_06_diagonal.png')
    expect(new Set(fixedChartNoteAssetUrls).size).toBe(fixedChartNoteAssetUrls.length)
  })
})

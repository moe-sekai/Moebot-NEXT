import { Context } from 'koishi'
import { MoebotConfig } from '../config'
import { renderToImage, ImageCache, type RenderOptions } from '@moebot/renderer'

export class RendererService {
  private cache: ImageCache

  constructor(
    private ctx: Context,
    private config: MoebotConfig,
  ) {
    this.cache = new ImageCache({
      cacheDir: undefined, // Use default
      maxSizeBytes: config.imageCache.maxSizeBytes,
      ttlMs: config.imageCache.ttlMs,
    })
  }

  /**
   * Render a JSX template to PNG, with caching
   */
  async render(
    cacheKey: string,
    element: any,
    options?: RenderOptions,
  ): Promise<Buffer> {
    // Check cache first
    if (this.config.imageCache.enabled) {
      const cached = await this.cache.get(cacheKey)
      if (cached) {
        this.ctx.logger('moebot').debug(`Cache hit: ${cacheKey}`)
        return cached
      }
    }

    // Render
    const startTime = Date.now()
    const png = await renderToImage(element, options)
    const elapsed = Date.now() - startTime
    this.ctx.logger('moebot').debug(`Rendered ${cacheKey} in ${elapsed}ms (${png.length} bytes)`)

    // Cache the result
    if (this.config.imageCache.enabled) {
      await this.cache.set(cacheKey, png)
    }

    return png
  }

  /**
   * Render without caching (for dynamic content like rankings)
   */
  async renderDirect(element: any, options?: RenderOptions): Promise<Buffer> {
    return renderToImage(element, options)
  }
}

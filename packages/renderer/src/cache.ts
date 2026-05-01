import { readFile, writeFile, mkdir, stat, readdir, unlink } from 'fs/promises'
import { join } from 'path'

export interface CacheOptions {
  cacheDir: string
  maxSizeBytes: number // Max total cache size (default 1GB)
  ttlMs: number // Time to live (default 24h)
}

const DEFAULT_CACHE_OPTIONS: CacheOptions = {
  cacheDir: join(process.cwd(), 'data', 'cache'),
  maxSizeBytes: 1024 * 1024 * 1024, // 1GB
  ttlMs: 24 * 60 * 60 * 1000, // 24 hours
}

export class ImageCache {
  private options: CacheOptions

  constructor(options?: Partial<CacheOptions>) {
    this.options = { ...DEFAULT_CACHE_OPTIONS, ...options }
  }

  async init(): Promise<void> {
    await mkdir(this.options.cacheDir, { recursive: true })
  }

  private getPath(key: string): string {
    // Sanitize key for filesystem
    const safeKey = key.replace(/[^a-zA-Z0-9_-]/g, '_')
    return join(this.options.cacheDir, `${safeKey}.png`)
  }

  async get(key: string): Promise<Buffer | null> {
    try {
      const filePath = this.getPath(key)
      const fileStat = await stat(filePath)

      // Check TTL
      const age = Date.now() - fileStat.mtimeMs
      if (age > this.options.ttlMs) {
        await unlink(filePath).catch(() => {})
        return null
      }

      return await readFile(filePath)
    } catch {
      return null
    }
  }

  async set(key: string, data: Buffer): Promise<void> {
    await this.init()
    const filePath = this.getPath(key)
    await writeFile(filePath, data)

    // Async cleanup (don't await)
    this.cleanup().catch(() => {})
  }

  async has(key: string): Promise<boolean> {
    try {
      const filePath = this.getPath(key)
      await stat(filePath)
      return true
    } catch {
      return false
    }
  }

  async delete(key: string): Promise<void> {
    try {
      await unlink(this.getPath(key))
    } catch {
      // File doesn't exist, ignore
    }
  }

  async cleanup(): Promise<void> {
    try {
      const files = await readdir(this.options.cacheDir)
      const fileStats = await Promise.all(
        files.map(async (f) => {
          const fullPath = join(this.options.cacheDir, f)
          const s = await stat(fullPath)
          return { path: fullPath, size: s.size, mtime: s.mtimeMs }
        })
      )

      // Remove expired files
      const now = Date.now()
      for (const file of fileStats) {
        if (now - file.mtime > this.options.ttlMs) {
          await unlink(file.path).catch(() => {})
        }
      }

      // If still over size limit, remove oldest files
      const remaining = fileStats
        .filter(f => now - f.mtime <= this.options.ttlMs)
        .sort((a, b) => a.mtime - b.mtime)

      let totalSize = remaining.reduce((sum, f) => sum + f.size, 0)
      for (const file of remaining) {
        if (totalSize <= this.options.maxSizeBytes) break
        await unlink(file.path).catch(() => {})
        totalSize -= file.size
      }
    } catch {
      // Cache dir doesn't exist yet, ignore
    }
  }
}

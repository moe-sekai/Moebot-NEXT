import { Schema } from 'koishi'

export interface MoebotConfig {
  /** Masterdata source URL */
  masterDataUrl: string
  /** Masterdata refresh interval in ms (default: 1 hour) */
  masterDataRefreshInterval: number
  /** Image cache settings */
  imageCache: {
    enabled: boolean
    maxSizeBytes: number
    ttlMs: number
  }
  /** Command prefix (default: /) */
  commandPrefix: string
  /** SEKAI API configuration (optional — bot works without it) */
  sekaiApi: {
    enabled: boolean
    baseUrl: string
    region: 'jp' | 'en' | 'tw' | 'kr'
    headers: Record<string, string>
    timeout: number
    rateLimit: number
    proxy?: string
  }
}

export const Config: Schema<MoebotConfig> = Schema.object({
  masterDataUrl: Schema.string()
    .default('https://sk.exmeaning.com/master')
    .description('Masterdata JSON source URL'),
  masterDataRefreshInterval: Schema.number()
    .default(3600000)
    .description('Masterdata auto-refresh interval (ms)'),
  imageCache: Schema.object({
    enabled: Schema.boolean().default(true),
    maxSizeBytes: Schema.number().default(1073741824).description('Max cache size in bytes (default 1GB)'),
    ttlMs: Schema.number().default(86400000).description('Cache TTL in ms (default 24h)'),
  }),
  commandPrefix: Schema.string().default('/'),
  sekaiApi: Schema.object({
    enabled: Schema.boolean().default(false).description('Enable SEKAI API integration (optional)'),
    baseUrl: Schema.string().default('').description('SEKAI API base URL'),
    region: Schema.union(['jp', 'en', 'tw', 'kr'] as const).default('jp'),
    headers: Schema.dict(Schema.string()).default({}).description('Custom request headers (auth tokens etc)'),
    timeout: Schema.number().default(10000).description('Request timeout (ms)'),
    rateLimit: Schema.number().default(30).description('Max requests per minute'),
    proxy: Schema.string().description('Optional proxy URL'),
  }).description('SEKAI game API integration (不接入也不影响基本功能)'),
})

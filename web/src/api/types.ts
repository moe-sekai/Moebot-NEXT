export interface DashboardData {
  commands_total: number
  users_total: number
  groups_total: number
  uptime: string
  version: string
}

export interface HealthResponse {
  status: string
  version: string
  time: string
  uptime: string
}

export interface MasterdataCounts {
  cards: number
  musics: number
  events: number
  gachas: number
}

export interface MasterdataSummary {
  loaded: boolean
  loaded_at: string | null
  counts: MasterdataCounts
}

export interface StatusBlock {
  status: string
  ok: boolean
  message: string
  [key: string]: unknown
}

export interface BotStatus extends StatusBlock {
  driver_type: string
  listen: string
  url_configured: boolean
  command_prefix: string
  nicknames: string[]
}

export interface WebStatus extends StatusBlock {
  host: string
  port: number
}

export interface RendererStatus extends StatusBlock {
  base_url: string
  status_code: number
  latency_ms: number
  service_port: number
  dashboard_port: number
}

export interface MasterdataStatus extends StatusBlock {
  loaded: boolean
  loaded_at: string | null
  counts: MasterdataCounts
}

export interface DatabaseStatus extends StatusBlock {
  path: string
}

export interface RuntimeStatus {
  version: string
  time: string
  uptime: string
  bot: BotStatus
  web: WebStatus
  renderer: RendererStatus
  masterdata: MasterdataStatus
  database: DatabaseStatus
}

export interface RendererHealth {
  ok: boolean
  status: string
  message: string
  base_url: string
  status_code: number
  latency_ms: number
  renderer_port: number
  dashboard_port: number
  note: string
}

export interface RecentCommand {
  id: number
  command: string
  platform: string
  user_id: string
  group_id: string
  args: string
  response_ms: number
  created_at: string
}

export interface RecentCommandsResponse {
  data: RecentCommand[]
  total: number
  message: string
}

export interface PublicConfig {
  version: string
  web: {
    host: string
    port: number
  }
  bot: {
    nickname: string[]
    command_prefix: string
    driver_type: string
    listen: string
    url_configured: boolean
    token_set?: boolean
  }
  masterdata: {
    url_configured: boolean
    fallback_url_configured: boolean
    local_path: string
    refresh_interval: number
  }
  sekai_api: {
    enabled: boolean
    base_url_configured: boolean
    region: string
    headers_configured: boolean
  }
  renderer: {
    base_url: string
    host: string
    port: number
    cache: {
      enabled: boolean
      path: string
      max_size_mb: number
      ttl_hours: number
    }
  }
  assets: {
    cdn_source: string
    music_alias_configured: boolean
    sticker_path: string
  }
}

export type SearchType = 'cards' | 'musics' | 'events' | 'gachas'

export interface SearchResult {
  id: number
  title: string
  subtitle: string
  type: string
  [key: string]: unknown
}

export interface SearchResponse {
  data: SearchResult[]
  total: number
  query: string
  message: string
}

export interface RenderPreviewMeta {
  id: string
  name: string
  description: string
  command: string
  templatePath: string
  viewerSource: string
  status: 'ready' | 'draft' | string
  width: number
  height: number
}

export interface RenderTiming {
  fonts_ms: number | null
  satori_ms: number | null
  resvg_ms: number | null
  total_ms: number | null
  proxy_ms: number | null
  network_ms: number | null
  size_bytes: number | null
}

export interface RendererPreviewImageResult {
  url: string
  blob: Blob
  timings: RenderTiming
}

export interface RendererPreviewsResponse {
  data: RenderPreviewMeta[]
  total: number
  ok: boolean
  message: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
}

export interface GroupRow {
  id: number
  platform: string
  group_id: string
  name: string
  enabled: boolean
}

export interface UserRow {
  id: number
  platform: string
  platform_id: string
  game_id: string
  nickname: string
}

export interface CommandStatRow {
  command: string
  count: number
  avg_ms: number
}

export interface CommandStatsResponse {
  data: CommandStatRow[]
  since: string
}

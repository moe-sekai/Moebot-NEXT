import axios from 'axios'
import type {
  CommandStatsResponse,
  DashboardData,
  GroupRow,
  HealthResponse,
  ConfigUpdateResponse,
  MasterdataReloadResponse,
  MasterdataSummary,
  PaginatedResponse,
  PublicConfig,
  RecentCommandsResponse,
  UpdatePublicConfigPayload,
  RendererHealth,
  RendererPreviewImageResult,
  RendererPreviewsResponse,
  RuntimeStatus,
  SearchResponse,
  SearchType,
  UserRow,
} from './types'

export const api = axios.create({
  baseURL: '/api',
  timeout: 10_000,
})

export type { DashboardData } from './types'

export async function getHealth() {
  const { data } = await api.get<HealthResponse>('/health')
  return data
}

export async function getDashboard() {
  const { data } = await api.get<DashboardData>('/dashboard')
  return data
}

export async function getStatus() {
  const { data } = await api.get<RuntimeStatus>('/status')
  return data
}

export async function getMasterdataSummary() {
  const { data } = await api.get<MasterdataSummary>('/masterdata/summary')
  return data
}

export async function getRendererHealth() {
  const { data } = await api.get<RendererHealth>('/renderer/health')
  return data
}

export async function getRecentCommands(limit = 10) {
  const { data } = await api.get<RecentCommandsResponse>('/commands/recent', {
    params: { limit },
  })
  return data
}

export async function getPublicConfig() {
  const { data } = await api.get<PublicConfig>('/config/public')
  return data
}

export async function updatePublicConfig(payload: UpdatePublicConfigPayload) {
  const { data } = await api.put<ConfigUpdateResponse>('/config/public', payload)
  return data
}

export async function reloadMasterdata(region?: string) {
  const { data } = await api.post<MasterdataReloadResponse>('/masterdata/reload', null, {
    params: region ? { region } : undefined,
  })
  return data
}

export async function searchMasterdata(type: SearchType, q: string) {
  const { data } = await api.get<SearchResponse>(`/search/${type}`, {
    params: { q },
  })
  return data
}

export async function getRendererPreviews() {
  const { data } = await api.get<RendererPreviewsResponse>('/renderer/previews')
  return data
}

export function getRendererPreviewImageUrl(id: string, width?: number, height?: number) {
  const params = new URLSearchParams({ ts: String(Date.now()) })
  if (width) params.set('width', String(width))
  if (height) params.set('height', String(height))
  return `/api/renderer/previews/${encodeURIComponent(id)}/image?${params.toString()}`
}

export async function renderRendererPreview(id: string, width?: number, height?: number): Promise<RendererPreviewImageResult> {
  const startedAt = performance.now()
  const response = await api.get<Blob>(`/renderer/previews/${encodeURIComponent(id)}/image`, {
    params: {
      ...(width ? { width } : {}),
      ...(height ? { height } : {}),
      ts: Date.now(),
    },
    responseType: 'blob',
  })
  const networkMs = Math.round(performance.now() - startedAt)
  const blob = response.data
  return {
    blob,
    url: URL.createObjectURL(blob),
    timings: {
      fonts_ms: parseHeaderNumber(response.headers['x-render-fonts-ms']),
      satori_ms: parseHeaderNumber(response.headers['x-render-satori-ms']),
      resvg_ms: parseHeaderNumber(response.headers['x-render-resvg-ms']),
      total_ms: parseHeaderNumber(response.headers['x-render-total-ms']),
      proxy_ms: parseHeaderNumber(response.headers['x-render-proxy-ms']),
      network_ms: networkMs,
      size_bytes: parseHeaderNumber(response.headers['x-render-size-bytes']) ?? blob.size,
    },
  }
}

export async function getGroups(page = 1, limit = 20) {
  const { data } = await api.get<PaginatedResponse<GroupRow>>('/groups', { params: { page, limit } })
  return data
}

export async function getUsers(page = 1, limit = 20) {
  const { data } = await api.get<PaginatedResponse<UserRow>>('/users', { params: { page, limit } })
  return data
}

export async function getCommandStats(days = 7) {
  const { data } = await api.get<CommandStatsResponse>('/stats/commands', { params: { days } })
  return data
}

function parseHeaderNumber(value: unknown) {
  const raw = Array.isArray(value) ? value[0] : value
  if (raw === undefined || raw === null || raw === '') return null
  const numberValue = Number(raw)
  return Number.isFinite(numberValue) ? numberValue : null
}

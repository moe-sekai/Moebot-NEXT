import axios from 'axios'
import type {
  DashboardData,
  HealthResponse,
  MasterdataSummary,
  PublicConfig,
  RecentCommandsResponse,
  RendererHealth,
  RendererPreviewsResponse,
  RuntimeStatus,
  SearchResponse,
  SearchType,
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

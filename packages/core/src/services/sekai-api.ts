import { Context } from 'koishi'
import type { SekaiApiConfig, SekaiUserProfile, SekaiApiResponse } from '@moebot/shared'

/**
 * Optional SEKAI game API client.
 * When not configured, commands that depend on it will show a friendly message
 * telling the user that this feature requires SEKAI API configuration.
 */
export class SekaiApiService {
  private requestCount = 0
  private requestResetTime = 0

  constructor(
    private ctx: Context,
    private config: SekaiApiConfig,
  ) {}

  /** Check if rate limit allows a request */
  private checkRateLimit(): boolean {
    const now = Date.now()
    if (now - this.requestResetTime > 60000) {
      this.requestCount = 0
      this.requestResetTime = now
    }
    return this.requestCount < this.config.rateLimit
  }

  /** Make an authenticated request to SEKAI API */
  private async request<T>(path: string): Promise<SekaiApiResponse<T>> {
    if (!this.checkRateLimit()) {
      return { httpStatus: 429, errorCode: 'rate_limit', message: 'Rate limit exceeded' }
    }

    this.requestCount++

    try {
      const url = `${this.config.baseUrl}${path}`
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), this.config.timeout)

      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          ...this.config.headers,
        },
        signal: controller.signal,
      })
      clearTimeout(timeout)

      if (!response.ok) {
        return {
          httpStatus: response.status,
          errorCode: 'http_error',
          message: `HTTP ${response.status}: ${response.statusText}`,
        }
      }

      const data = await response.json() as T
      return { httpStatus: 200, data }
    } catch (err: any) {
      this.ctx.logger('moebot').warn(`SEKAI API request failed: ${path}`, err.message)
      return {
        httpStatus: 0,
        errorCode: 'network_error',
        message: err.message,
      }
    }
  }

  /** Fetch user profile by game user ID */
  async getUserProfile(userId: number): Promise<SekaiApiResponse<SekaiUserProfile>> {
    return this.request<SekaiUserProfile>(`/api/user/${userId}/profile`)
  }

  /** Fetch event ranking */
  async getEventRanking(eventId: number, rank: number): Promise<SekaiApiResponse<any>> {
    return this.request(`/api/user/%7BuserId%7D/event/${eventId}/ranking?targetRank=${rank}`)
  }

  /** Check API connectivity */
  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(this.config.baseUrl, {
        headers: this.config.headers,
        signal: AbortSignal.timeout(5000),
      })
      return response.ok
    } catch {
      return false
    }
  }
}

/**
 * Helper to check if SEKAI API is available and show a friendly message if not
 */
export function requireSekaiApi(sekaiApi: SekaiApiService | null): SekaiApiService {
  if (!sekaiApi) {
    throw new SekaiApiNotConfiguredError()
  }
  return sekaiApi
}

export class SekaiApiNotConfiguredError extends Error {
  constructor() {
    super('此功能需要配置 SEKAI API 端点。请在管理面板中配置 SEKAI API 连接信息。')
    this.name = 'SekaiApiNotConfiguredError'
  }
}

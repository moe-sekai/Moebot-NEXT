import { Context, Schema } from 'koishi'
import { listRenderPreviews, renderPreviewTemplate, type RenderOptions } from '@moebot/renderer'

export const name = 'moebot-console'
export const inject = {
  required: ['console', 'database'],
}
export const usage = `
## Moebot 管理面板

提供 Web 管理界面，包括：
- 运行状态监控
- SEKAI API 端点配置
- 群组管理
- 用户绑定管理
- 指令使用统计
`

export interface ConsoleConfig {
  auth: {
    username: string
    password: string
  }
}

export const Config: Schema<ConsoleConfig> = Schema.object({
  auth: Schema.object({
    username: Schema.string().default('admin'),
    password: Schema.string().default('').description('管理面板密码（留空则首次访问时设置）'),
  }),
})

interface RendererPreviewRequest {
  id: string
  width?: number
  height?: number
  debug?: boolean
  includeSvg?: boolean
}

export function apply(ctx: Context, config: ConsoleConfig) {
  const logger = ctx.logger('moebot-console')
  const koishi = ctx as any
  const clientEntry = getClientEntry()

  // ======== Dashboard API ========
  koishi.console.addEntry({
    dev: clientEntry,
    prod: clientEntry,
  })

  // ======== Renderer Preview API ========
  koishi.console.addListener('moebot/renderer/templates', async () => {
    return listRenderPreviews()
  })

  koishi.console.addListener('moebot/renderer/preview', async (payload: RendererPreviewRequest) => {
    try {
      if (!payload?.id) {
        return { success: false, error: '缺少模板 ID' }
      }

      const options: RenderOptions = {
        debug: Boolean(payload.debug),
      }
      if (isPositiveNumber(payload.width)) options.width = payload.width
      if (isPositiveNumber(payload.height)) options.height = payload.height

      const result = await renderPreviewTemplate(payload.id, options)
      const pngBase64 = result.trace.png.toString('base64')
      const includeSvg = payload.includeSvg !== false

      return {
        success: true,
        meta: result.meta,
        image: `data:image/png;base64,${pngBase64}`,
        svg: includeSvg ? result.trace.svg : undefined,
        svgDataUrl: includeSvg ? `data:image/svg+xml;utf8,${encodeURIComponent(result.trace.svg)}` : undefined,
        timings: result.trace.timings,
        sizeBytes: result.trace.sizeBytes,
        width: result.trace.width,
        height: result.trace.height,
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err)
      logger.warn('Renderer preview failed:', message)
      return { success: false, error: message }
    }
  })

  // API: Get bot status
  koishi.console.addListener('moebot/status', async () => {
    // Get stats from database
    let totalUsers = 0
    let totalGroups = 0
    let totalCommands = 0

    try {
      const users = await koishi.database.get('moebot.users', {})
      totalUsers = users.length
      const groups = await koishi.database.get('moebot.groups', {})
      totalGroups = groups.length
      const stats = await koishi.database.get('moebot.stats', {})
      totalCommands = stats.length
    } catch (err) {
      logger.warn('Failed to fetch stats:', err)
    }

    return {
      uptime: process.uptime(),
      totalUsers,
      totalGroups,
      totalCommands,
      memoryUsage: process.memoryUsage(),
      nodeVersion: process.version,
      platform: process.platform,
    }
  })

  // ======== SEKAI API Configuration API ========

  // API: Get current SEKAI API config
  koishi.console.addListener('moebot/sekai-api/config', async () => {
    try {
      const [config] = await koishi.database.get('moebot.groups', { groupId: '__sekai_api_config__' })
      if (config) {
        return JSON.parse(config.config)
      }
    } catch {}

    // Return default config
    return {
      endpoints: [],
    }
  })

  // API: Save SEKAI API config
  koishi.console.addListener('moebot/sekai-api/save', async (data: any) => {
    const configStr = JSON.stringify(data)

    try {
      const [existing] = await koishi.database.get('moebot.groups', { groupId: '__sekai_api_config__' })
      if (existing) {
        await koishi.database.set('moebot.groups', existing.id, { config: configStr })
      } else {
        await koishi.database.create('moebot.groups', {
          platform: 'system',
          groupId: '__sekai_api_config__',
          name: 'SEKAI API Configuration',
          enabled: true,
          config: configStr,
          createdAt: new Date(),
        })
      }

      logger.info('SEKAI API configuration saved')
      return { success: true }
    } catch (err: any) {
      logger.error('Failed to save SEKAI API config:', err)
      return { success: false, error: err.message }
    }
  })

  // API: Test SEKAI API connection
  koishi.console.addListener('moebot/sekai-api/test', async (endpoint: any) => {
    try {
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), endpoint.timeout || 10000)

      const response = await fetch(endpoint.baseUrl, {
        headers: endpoint.headers || {},
        signal: controller.signal,
      })
      clearTimeout(timeout)

      return {
        success: response.ok,
        status: response.status,
        statusText: response.statusText,
        responseTime: Date.now(),
      }
    } catch (err: any) {
      return {
        success: false,
        error: err.message,
      }
    }
  })

  // ======== Groups API ========
  koishi.console.addListener('moebot/groups', async () => {
    try {
      const groups = await koishi.database.get('moebot.groups', {
        groupId: { $ne: '__sekai_api_config__' },
      })
      return groups
    } catch {
      return []
    }
  })

  koishi.console.addListener('moebot/groups/toggle', async ({ id, enabled }: { id: number; enabled: boolean }) => {
    await koishi.database.set('moebot.groups', id, { enabled })
    return { success: true }
  })

  // ======== Users API ========
  koishi.console.addListener('moebot/users', async () => {
    try {
      return await koishi.database.get('moebot.users', {})
    } catch {
      return []
    }
  })

  // ======== Command Stats API ========
  koishi.console.addListener('moebot/stats', async () => {
    try {
      const stats = await koishi.database.get('moebot.stats', {})

      // Aggregate by command
      const commandCounts: Record<string, number> = {}
      const commandAvgMs: Record<string, number[]> = {}

      for (const stat of stats) {
        commandCounts[stat.command] = (commandCounts[stat.command] || 0) + 1
        if (!commandAvgMs[stat.command]) commandAvgMs[stat.command] = []
        commandAvgMs[stat.command].push(stat.responseMs)
      }

      return {
        total: stats.length,
        commands: Object.entries(commandCounts).map(([name, count]) => ({
          name,
          count,
          avgResponseMs: Math.round(
            commandAvgMs[name].reduce((a, b) => a + b, 0) / commandAvgMs[name].length
          ),
        })).sort((a, b) => b.count - a.count),
      }
    } catch {
      return { total: 0, commands: [] }
    }
  })

  logger.info('Console management panel loaded')
}

function isPositiveNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value) && value > 0
}

function resolve(...args: string[]) {
  return require('path').resolve(...args)
}

function getClientEntry(): string {
  const { existsSync } = require('fs')
  const candidates = [
    // compiled runtime: dist/index.js -> ../../client/index.ts
    resolve(__dirname, '../../client/index.ts'),
    // source runtime: src/index.ts -> ../client/index.ts
    resolve(__dirname, '../client/index.ts'),
  ]
  return candidates.find((file) => existsSync(file)) ?? candidates[0]
}

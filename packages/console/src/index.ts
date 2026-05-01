import { Context, Schema } from 'koishi'

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

export function apply(ctx: Context, config: ConsoleConfig) {
  const logger = ctx.logger('moebot-console')

  // ======== Dashboard API ========
  ctx.console.addEntry(process.env.KOISHI_BASE ? [
    process.env.KOISHI_BASE + '/dist/index.js',
  ] : process.env.KOISHI_ENV === 'browser' ? [
    // @ts-ignore
    import.meta.url.replace(/\/src\/.*$/, '/client/index.ts'),
  ] : {
    dev: resolve(__dirname, '../client/index.ts'),
    prod: resolve(__dirname, '../dist'),
  })

  // API: Get bot status
  ctx.console.addListener('moebot/status', async () => {
    // Get stats from database
    let totalUsers = 0
    let totalGroups = 0
    let totalCommands = 0

    try {
      const users = await ctx.database.get('moebot.users', {})
      totalUsers = users.length
      const groups = await ctx.database.get('moebot.groups', {})
      totalGroups = groups.length
      const stats = await ctx.database.get('moebot.stats', {})
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
  ctx.console.addListener('moebot/sekai-api/config', async () => {
    try {
      const [config] = await ctx.database.get('moebot.groups', { groupId: '__sekai_api_config__' })
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
  ctx.console.addListener('moebot/sekai-api/save', async (data: any) => {
    const configStr = JSON.stringify(data)

    try {
      const [existing] = await ctx.database.get('moebot.groups', { groupId: '__sekai_api_config__' })
      if (existing) {
        await ctx.database.set('moebot.groups', existing.id, { config: configStr })
      } else {
        await ctx.database.create('moebot.groups', {
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
  ctx.console.addListener('moebot/sekai-api/test', async (endpoint: any) => {
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
  ctx.console.addListener('moebot/groups', async () => {
    try {
      const groups = await ctx.database.get('moebot.groups', {
        groupId: { $ne: '__sekai_api_config__' },
      })
      return groups
    } catch {
      return []
    }
  })

  ctx.console.addListener('moebot/groups/toggle', async ({ id, enabled }: { id: number; enabled: boolean }) => {
    await ctx.database.set('moebot.groups', id, { enabled })
    return { success: true }
  })

  // ======== Users API ========
  ctx.console.addListener('moebot/users', async () => {
    try {
      return await ctx.database.get('moebot.users', {})
    } catch {
      return []
    }
  })

  // ======== Command Stats API ========
  ctx.console.addListener('moebot/stats', async () => {
    try {
      const stats = await ctx.database.get('moebot.stats', {})

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

function resolve(...args: string[]) {
  return require('path').resolve(...args)
}

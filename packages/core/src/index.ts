import { Context } from 'koishi'
import { Config, MoebotConfig } from './config'
import { MasterdataService } from './services/masterdata'
import { DatabaseService } from './services/database'
import { RendererService } from './services/renderer'
import { SekaiApiService } from './services/sekai-api'
import { registerCommands } from './commands'

export const name = 'moebot-core'
export { Config }
export const inject = {
  required: ['database'],
  optional: [],
}
export const usage = `
## Moebot NEXT 核心插件

PJSK (Project SEKAI) 查询机器人核心功能。

### 功能列表
- 查卡 / 查曲 / 查活动 / 查卡池
- 绑定游戏账号 / 个人信息查询
- 表情贴纸 / 猜曲游戏
- 帮助指令

### SEKAI API (可选)
接入 SEKAI API 后可解锁：
- 实时排行查询
- 玩家详细数据查询
- Best 30 成绩展示

不接入也不影响基础查询功能。
`

export function apply(ctx: Context, config: MoebotConfig) {
  // Initialize services
  const masterdata = new MasterdataService(ctx, config)
  const database = new DatabaseService(ctx)
  const renderer = new RendererService(ctx, config)

  // Optional SEKAI API service
  let sekaiApi: SekaiApiService | null = null
  if (config.sekaiApi.enabled) {
    sekaiApi = new SekaiApiService(ctx, {
      name: 'default',
      ...config.sekaiApi,
    })
    ctx.logger('moebot').info('SEKAI API integration enabled')
  } else {
    ctx.logger('moebot').info('SEKAI API not configured — basic features only')
  }

  // Provide services to context for commands
  ctx.set('moebot.masterdata', masterdata)
  ctx.set('moebot.database', database)
  ctx.set('moebot.renderer', renderer)
  ctx.set('moebot.sekaiApi', sekaiApi)

  // Initialize masterdata
  masterdata.init()

  // Initialize database tables
  database.init()

  // Register all commands
  registerCommands(ctx, config, { masterdata, database, renderer, sekaiApi })
}

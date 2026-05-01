import { Context } from 'koishi'
import { MoebotConfig } from '../config'
import { MasterdataService } from '../services/masterdata'
import { DatabaseService } from '../services/database'
import { RendererService } from '../services/renderer'
import { SekaiApiService } from '../services/sekai-api'

export interface CommandServices {
  masterdata: MasterdataService
  database: DatabaseService
  renderer: RendererService
  sekaiApi: SekaiApiService | null
}

export function registerCommands(
  ctx: Context,
  config: MoebotConfig,
  services: CommandServices,
) {
  const { masterdata, database, renderer, sekaiApi } = services
  const logger = ctx.logger('moebot')

  // Middleware: log command usage
  ctx.before('command/execute', async ({ command, session }) => {
    const startTime = Date.now()
    // Store start time for later logging
    if (session) {
      (session as any)._moebotStartTime = startTime
    }
  })

  // Register commands
  registerCardCommand(ctx, services)
  registerMusicCommand(ctx, services)
  registerEventCommand(ctx, services)
  registerGachaCommand(ctx, services)
  registerProfileCommand(ctx, services)
  registerHelpCommand(ctx, services)
  registerStickerCommand(ctx, services)
  registerGuessCommand(ctx, services)

  logger.info('All commands registered')
}

// Import and re-export individual command registrations
function registerCardCommand(ctx: Context, services: CommandServices) {
  const { masterdata, renderer, database } = services

  ctx.command('查卡 <keyword:text>', '搜索 PJSK 卡牌')
    .alias('card')
    .alias('卡牌')
    .action(async ({ session }, keyword) => {
      if (!keyword) return '请输入搜索关键词，例如：/查卡 初音未来'
      if (!masterdata.isReady) return 'Masterdata 正在加载中，请稍后再试...'

      const startTime = Date.now()

      try {
        // TODO: Implement full search logic with fuzzy matching
        const cards = masterdata.cards.filter((card: any) => {
          const char = masterdata.gameCharacters.find((c: any) => c.id === card.characterId)
          const searchStr = `${card.prefix} ${char?.givenName ?? ''} ${char?.firstName ?? ''}`
          return searchStr.toLowerCase().includes(keyword.toLowerCase())
        })

        if (cards.length === 0) {
          return `未找到与「${keyword}」相关的卡牌`
        }

        // For now, return text result. Image rendering will be added later.
        const card = cards[0]
        const char = masterdata.gameCharacters.find((c: any) => c.id === card.characterId)

        const responseMs = Date.now() - startTime
        await database.logCommand('查卡', session?.platform ?? 'unknown', session?.userId, session?.guildId, keyword, responseMs)

        return [
          `🎴 ${card.prefix}`,
          `角色: ${char?.givenName ?? '未知'}`,
          `稀有度: ${card.cardRarityType}`,
          `属性: ${card.attr}`,
          `ID: ${card.id}`,
          cards.length > 1 ? `\n还有 ${cards.length - 1} 张相关卡牌` : '',
        ].filter(Boolean).join('\n')
      } catch (err: any) {
        return `查询出错: ${err.message}`
      }
    })
}

function registerMusicCommand(ctx: Context, services: CommandServices) {
  const { masterdata, database } = services

  ctx.command('查曲 <keyword:text>', '搜索 PJSK 曲目')
    .alias('music')
    .alias('曲目')
    .action(async ({ session }, keyword) => {
      if (!keyword) return '请输入搜索关键词，例如：/查曲 テルテル'
      if (!masterdata.isReady) return 'Masterdata 正在加载中，请稍后再试...'

      const startTime = Date.now()

      try {
        const musics = masterdata.musics.filter((m: any) =>
          m.title?.toLowerCase().includes(keyword.toLowerCase()) ||
          m.pronunciation?.toLowerCase().includes(keyword.toLowerCase())
        )

        if (musics.length === 0) {
          return `未找到与「${keyword}」相关的曲目`
        }

        const music = musics[0]
        const responseMs = Date.now() - startTime
        await database.logCommand('查曲', session?.platform ?? 'unknown', session?.userId, session?.guildId, keyword, responseMs)

        return [
          `🎵 ${music.title}`,
          music.pronunciation ? `读音: ${music.pronunciation}` : '',
          `作词: ${music.lyricist}`,
          `作曲: ${music.composer}`,
          `编曲: ${music.arranger}`,
          `ID: ${music.id}`,
          musics.length > 1 ? `\n还有 ${musics.length - 1} 首相关曲目` : '',
        ].filter(Boolean).join('\n')
      } catch (err: any) {
        return `查询出错: ${err.message}`
      }
    })
}

function registerEventCommand(ctx: Context, services: CommandServices) {
  const { masterdata, database } = services

  ctx.command('查活动 [keyword:text]', '搜索 PJSK 活动')
    .alias('event')
    .alias('活动')
    .action(async ({ session }, keyword) => {
      if (!masterdata.isReady) return 'Masterdata 正在加载中，请稍后再试...'

      const startTime = Date.now()

      try {
        let events = masterdata.events
        if (keyword) {
          events = events.filter((e: any) =>
            e.name?.toLowerCase().includes(keyword.toLowerCase()) ||
            String(e.id) === keyword
          )
        } else {
          // Show latest event
          events = [...events].sort((a: any, b: any) => b.startAt - a.startAt).slice(0, 1)
        }

        if (events.length === 0) {
          return `未找到与「${keyword}」相关的活动`
        }

        const event = events[0]
        const startDate = new Date(event.startAt).toLocaleDateString('zh-CN')
        const endDate = new Date(event.closedAt).toLocaleDateString('zh-CN')

        const responseMs = Date.now() - startTime
        await database.logCommand('查活动', session?.platform ?? 'unknown', session?.userId, session?.guildId, keyword, responseMs)

        return [
          `📅 ${event.name}`,
          `类型: ${event.eventType}`,
          `时间: ${startDate} ~ ${endDate}`,
          `ID: ${event.id}`,
        ].join('\n')
      } catch (err: any) {
        return `查询出错: ${err.message}`
      }
    })
}

function registerGachaCommand(ctx: Context, services: CommandServices) {
  ctx.command('查卡池 [keyword:text]', '搜索 PJSK 卡池')
    .alias('gacha')
    .alias('卡池')
    .action(async (_argv, keyword) => {
      // TODO: Implement gacha search
      return '🎰 查卡池功能开发中...'
    })
}

function registerProfileCommand(ctx: Context, services: CommandServices) {
  const { database, sekaiApi } = services

  ctx.command('绑定 <gameId:string>', '绑定 PJSK 游戏账号')
    .alias('bind')
    .action(async ({ session }, gameId) => {
      if (!gameId) return '请输入你的游戏ID，例如：/绑定 1234567890'
      if (!session) return '无法获取用户信息'

      try {
        await database.bindUser(session.platform, session.userId, gameId)
        return `✅ 绑定成功！\n游戏ID: ${gameId}\n\n使用 /个人信息 查看你的游戏数据`
      } catch (err: any) {
        return `绑定失败: ${err.message}`
      }
    })

  ctx.command('解绑', '解除 PJSK 账号绑定')
    .alias('unbind')
    .action(async ({ session }) => {
      if (!session) return '无法获取用户信息'

      const success = await database.unbindUser(session.platform, session.userId)
      return success ? '✅ 已解除绑定' : '❌ 你还没有绑定游戏账号'
    })

  ctx.command('个人信息', '查看绑定的 PJSK 账号信息')
    .alias('profile')
    .alias('我的')
    .action(async ({ session }) => {
      if (!session) return '无法获取用户信息'

      const user = await database.findUser(session.platform, session.userId)
      if (!user?.gameId) {
        return '你还没有绑定游戏账号，使用 /绑定 <游戏ID> 来绑定'
      }

      // If SEKAI API is configured, try to fetch detailed profile
      if (sekaiApi) {
        try {
          const result = await sekaiApi.getUserProfile(parseInt(user.gameId))
          if (result.data) {
            const profile = result.data
            return [
              `👤 ${profile.name}`,
              `Rank: ${profile.rank}`,
              `综合力: ${profile.userGamedata.totalPower}`,
              `协力次数: ${profile.userGamedata.multiLiveCount}`,
              `MVP: ${profile.userGamedata.mvpCount}`,
              `SS: ${profile.userGamedata.superStarCount}`,
            ].join('\n')
          }
        } catch {
          // Fall through to basic info
        }
      }

      // Basic info without SEKAI API
      return [
        `👤 游戏信息`,
        `游戏ID: ${user.gameId}`,
        `服务器: ${user.region?.toUpperCase() ?? 'JP'}`,
        '',
        sekaiApi ? '' : '💡 提示: 配置 SEKAI API 后可查看更详细的游戏数据',
      ].filter(Boolean).join('\n')
    })
}

function registerHelpCommand(ctx: Context, services: CommandServices) {
  ctx.command('帮助', '显示 Moebot NEXT 帮助信息')
    .alias('help')
    .alias('菜单')
    .action(async () => {
      return [
        '🤖 Moebot NEXT — PJSK 查询助手',
        '',
        '📋 基础指令:',
        '  /查卡 <关键词>    搜索卡牌',
        '  /查曲 <关键词>    搜索曲目',
        '  /查活动 [关键词]  搜索活动',
        '  /查卡池 [关键词]  搜索卡池',
        '  /表情 <编号>      发送表情贴纸',
        '',
        '👤 个人指令:',
        '  /绑定 <游戏ID>   绑定游戏账号',
        '  /解绑             解除绑定',
        '  /个人信息         查看个人数据',
        '',
        '🎮 娱乐指令:',
        '  /猜曲             猜歌游戏',
        '  /猜角色           猜角色游戏',
        '',
        '💡 Powered by pjsk.moe',
      ].join('\n')
    })
}

function registerStickerCommand(ctx: Context, services: CommandServices) {
  ctx.command('表情 <id:number>', '发送 PJSK 游戏表情贴纸')
    .alias('sticker')
    .alias('贴纸')
    .action(async ({ session }, id) => {
      if (!id) return '请输入表情编号，例如：/表情 1'
      // TODO: Implement sticker sending with image
      return `表情 #${id} (功能开发中...)`
    })
}

function registerGuessCommand(ctx: Context, services: CommandServices) {
  ctx.command('猜曲', '开始猜歌游戏')
    .alias('guess-music')
    .action(async () => {
      // TODO: Implement guess music game
      return '🎵 猜曲游戏开发中...'
    })

  ctx.command('猜角色', '开始猜角色游戏')
    .alias('guess-who')
    .action(async () => {
      // TODO: Implement guess character game
      return '🎭 猜角色游戏开发中...'
    })
}

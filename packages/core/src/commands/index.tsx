import { Context, h } from 'koishi'
import {
  CardDetail,
  EventInfo,
  GachaInfo,
  HelpCard,
  MusicDetail,
  ProfileCard,
  RankingList,
} from '@moebot/renderer'
import {
  getCardThumbnailUrl,
  getEventBannerUrl,
  getEventLogoUrl,
  getGachaLogoUrl,
  getMusicJacketUrl,
} from '@moebot/shared'
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

const ASSET_SOURCE = 'main-jp'

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
  registerRankingCommand(ctx, services)
  registerProfileCommand(ctx, services)
  registerHelpCommand(ctx, services)
  registerStickerCommand(ctx, services)
  registerGuessCommand(ctx, services)

  logger.info('All commands registered')
}

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
        const cards = searchCards(masterdata, keyword)

        if (cards.length === 0) {
          return `未找到与「${keyword}」相关的卡牌`
        }

        const card = cards[0]
        const char = getGameCharacter(masterdata, card.characterId)
        const fallback = formatCardText(card, char, cards.length)
        const png = await renderer.render(
          `card_detail_${card.id}_${card.assetbundleName ?? 'noasset'}`,
          <CardDetail
            card={{
              id: card.id,
              prefix: card.prefix ?? `Card #${card.id}`,
              characterName: characterName(char),
              rarity: card.cardRarityType ?? card.rarity ?? 'rarity_1',
              cardRarityType: card.cardRarityType,
              attr: card.attr ?? 'cute',
              assetbundleName: card.assetbundleName,
              characterId: card.characterId,
              assetSource: ASSET_SOURCE,
              power: calcCardPower(card),
              skillName: findSkillName(masterdata, card.skillId),
              gachaPhrase: card.gachaPhrase,
              supplyType: card.supplyType,
            }}
          />,
          { width: 800, height: 620 },
        ).catch((err) => {
          ctx.logger('moebot').warn(`Render /查卡 failed: ${err?.message ?? err}`)
          return null
        })

        const responseMs = Date.now() - startTime
        await database.logCommand('查卡', session?.platform ?? 'unknown', session?.userId, session?.guildId, keyword, responseMs)

        return png ? imageReply(png) : fallback
      } catch (err: any) {
        return `查询出错: ${err.message}`
      }
    })
}

function registerMusicCommand(ctx: Context, services: CommandServices) {
  const { masterdata, database, renderer } = services

  ctx.command('查曲 <keyword:text>', '搜索 PJSK 曲目')
    .alias('music')
    .alias('曲目')
    .action(async ({ session }, keyword) => {
      if (!keyword) return '请输入搜索关键词，例如：/查曲 テルテル'
      if (!masterdata.isReady) return 'Masterdata 正在加载中，请稍后再试...'

      const startTime = Date.now()

      try {
        const musics = searchMusics(masterdata, keyword)

        if (musics.length === 0) {
          return `未找到与「${keyword}」相关的曲目`
        }

        const music = musics[0]
        const difficulties = (masterdata.data.musicDifficulties ?? [])
          .filter((d: any) => Number(d.musicId) === Number(music.id))
        const fallback = formatMusicText(music, musics.length)
        const png = await renderer.render(
          `music_detail_${music.id}_${music.assetbundleName ?? 'noasset'}`,
          <MusicDetail
            music={{
              id: music.id,
              title: music.title,
              pronunciation: music.pronunciation,
              lyricist: music.lyricist,
              composer: music.composer,
              arranger: music.arranger,
              categories: music.categories,
              assetbundleName: music.assetbundleName,
              assetSource: ASSET_SOURCE,
              jacketUrl: music.assetbundleName ? getMusicJacketUrl(music.assetbundleName, ASSET_SOURCE) : undefined,
              difficulties,
              publishedAt: music.publishedAt ?? music.releasedAt,
              isNewlyWrittenMusic: music.isNewlyWrittenMusic,
              isFullLength: music.isFullLength,
              fillerSec: music.fillerSec,
            }}
          />,
          { width: 800, height: 650 },
        ).catch((err) => {
          ctx.logger('moebot').warn(`Render /查曲 failed: ${err?.message ?? err}`)
          return null
        })

        const responseMs = Date.now() - startTime
        await database.logCommand('查曲', session?.platform ?? 'unknown', session?.userId, session?.guildId, keyword, responseMs)

        return png ? imageReply(png) : fallback
      } catch (err: any) {
        return `查询出错: ${err.message}`
      }
    })
}

function registerEventCommand(ctx: Context, services: CommandServices) {
  const { masterdata, database, renderer } = services

  ctx.command('查活动 [keyword:text]', '搜索 PJSK 活动')
    .alias('event')
    .alias('活动')
    .action(async ({ session }, keyword) => {
      if (!masterdata.isReady) return 'Masterdata 正在加载中，请稍后再试...'

      const startTime = Date.now()

      try {
        const events = searchEvents(masterdata, keyword)

        if (events.length === 0) {
          return `未找到与「${keyword}」相关的活动`
        }

        const event = events[0]
        const fallback = formatEventText(event)
        const eventBonus = getEventBonus(masterdata, event.id)
        const png = await renderer.render(
          `event_info_${event.id}_${event.assetbundleName ?? 'noasset'}`,
          <EventInfo
            event={{
              id: event.id,
              name: event.name,
              eventType: event.eventType,
              assetbundleName: event.assetbundleName,
              assetSource: ASSET_SOURCE,
              bannerUrl: event.assetbundleName ? getEventBannerUrl(event.assetbundleName, ASSET_SOURCE) : undefined,
              logoUrl: event.assetbundleName ? getEventLogoUrl(event.assetbundleName, ASSET_SOURCE) : undefined,
              startAt: event.startAt,
              aggregateAt: event.aggregateAt,
              closedAt: event.closedAt,
              distributionEndAt: event.distributionEndAt,
              unit: event.unit,
              bonusAttr: eventBonus.attr,
              bonusCharacters: eventBonus.characters,
            }}
          />,
          { width: 800, height: 720 },
        ).catch((err) => {
          ctx.logger('moebot').warn(`Render /查活动 failed: ${err?.message ?? err}`)
          return null
        })

        const responseMs = Date.now() - startTime
        await database.logCommand('查活动', session?.platform ?? 'unknown', session?.userId, session?.guildId, keyword, responseMs)

        return png ? imageReply(png) : fallback
      } catch (err: any) {
        return `查询出错: ${err.message}`
      }
    })
}

function registerGachaCommand(ctx: Context, services: CommandServices) {
  const { masterdata, database, renderer } = services

  ctx.command('查卡池 [keyword:text]', '搜索 PJSK 卡池')
    .alias('gacha')
    .alias('卡池')
    .action(async ({ session }, keyword) => {
      if (!masterdata.isReady) return 'Masterdata 正在加载中，请稍后再试...'
      const startTime = Date.now()

      try {
        const gachas = searchGachas(masterdata, keyword)
        if (gachas.length === 0) return `未找到与「${keyword ?? '最新'}」相关的卡池`

        const gacha = gachas[0]
        const pickupCards = getGachaPickupCards(masterdata, gacha)
        const fallback = formatGachaText(gacha, pickupCards.length)
        const png = await renderer.render(
          `gacha_info_${gacha.id}_${gacha.assetbundleName ?? 'noasset'}`,
          <GachaInfo
            gacha={{
              id: gacha.id,
              name: gacha.name,
              gachaType: gacha.gachaType,
              assetbundleName: gacha.assetbundleName,
              assetSource: ASSET_SOURCE,
              logoUrl: gacha.assetbundleName ? getGachaLogoUrl(gacha.assetbundleName, ASSET_SOURCE) : undefined,
              startAt: gacha.startAt,
              endAt: gacha.endAt,
              isShowPeriod: gacha.isShowPeriod,
              wishSelectCount: gacha.wishSelectCount,
              pickupCards,
            }}
          />,
          { width: 800, height: 760 },
        ).catch((err) => {
          ctx.logger('moebot').warn(`Render /查卡池 failed: ${err?.message ?? err}`)
          return null
        })

        const responseMs = Date.now() - startTime
        await database.logCommand('查卡池', session?.platform ?? 'unknown', session?.userId, session?.guildId, keyword, responseMs)

        return png ? imageReply(png) : fallback
      } catch (err: any) {
        return `查询出错: ${err.message}`
      }
    })
}

function registerRankingCommand(ctx: Context, services: CommandServices) {
  const { masterdata, sekaiApi, renderer, database } = services

  ctx.command('排行 [rank:number]', '查询活动实时排行')
    .alias('ranking')
    .alias('活动排行')
    .action(async ({ session }, rank = 100) => {
      if (!masterdata.isReady) return 'Masterdata 正在加载中，请稍后再试...'
      if (!sekaiApi) return '此功能需要配置 SEKAI API 端点。请在管理面板中配置后再使用 /排行。'

      const startTime = Date.now()
      try {
        const event = getLatestEvent(masterdata)
        if (!event) return '未找到当前活动数据'

        const result = await sekaiApi.getEventRanking(event.id, rank)
        if (!result.data) return `排行查询失败：${result.message ?? result.errorCode ?? 'API 未返回数据'}`

        const rankings = normalizeRankingData(result.data)
        if (rankings.length === 0) return 'API 没有返回可展示的排行数据'

        const png = await renderer.renderDirect(
          <RankingList
            title={event.name}
            eventId={event.id}
            eventName={event.name}
            updatedAt={Date.now()}
            assetSource={ASSET_SOURCE}
            rankings={rankings}
          />,
          { width: 800, height: 760 },
        )

        const responseMs = Date.now() - startTime
        await database.logCommand('排行', session?.platform ?? 'unknown', session?.userId, session?.guildId, String(rank), responseMs)

        return imageReply(png)
      } catch (err: any) {
        return `排行查询出错: ${err.message}`
      }
    })
}

function registerProfileCommand(ctx: Context, services: CommandServices) {
  const { database, sekaiApi, renderer, masterdata } = services

  ctx.command('绑定 <gameId:string>', '绑定 PJSK 游戏账号')
    .alias('bind')
    .action(async ({ session }, gameId) => {
      if (!gameId) return '请输入你的游戏ID，例如：/绑定 1234567890'
      if (!session) return '无法获取用户信息'

      try {
        await database.bindUser(session.platform, (session.userId ?? ''), gameId)
        return `✅ 绑定成功！\n游戏ID: ${gameId}\n\n使用 /个人信息 查看你的游戏数据`
      } catch (err: any) {
        return `绑定失败: ${err.message}`
      }
    })

  ctx.command('解绑', '解除 PJSK 账号绑定')
    .alias('unbind')
    .action(async ({ session }) => {
      if (!session) return '无法获取用户信息'

      const success = await database.unbindUser(session.platform, (session.userId ?? ''))
      return success ? '✅ 已解除绑定' : '❌ 你还没有绑定游戏账号'
    })

  ctx.command('个人信息', '查看绑定的 PJSK 账号信息')
    .alias('profile')
    .alias('我的')
    .action(async ({ session }) => {
      if (!session) return '无法获取用户信息'

      const user = await database.findUser(session.platform, (session.userId ?? ''))
      if (!user?.gameId) {
        return '你还没有绑定游戏账号，使用 /绑定 <游戏ID> 来绑定'
      }

      const fallback = [
        '👤 游戏信息',
        `游戏ID: ${user.gameId}`,
        `服务器: ${user.region?.toUpperCase() ?? 'JP'}`,
        '',
        sekaiApi ? '' : '💡 提示: 配置 SEKAI API 后可查看更详细的游戏数据',
      ].filter(Boolean).join('\n')

      try {
        if (sekaiApi) {
          const result = await sekaiApi.getUserProfile(parseInt(user.gameId))
          if (result.data) {
            const profile = result.data
            const deckCards = buildProfileDeckCards(masterdata, profile)
            const png = await renderer.render(
              `profile_${profile.userId}_${profile.rank}_${profile.userGamedata?.totalPower ?? 0}`,
              <ProfileCard
                profile={{
                  name: profile.name,
                  rank: profile.rank,
                  userId: profile.userId,
                  twitterId: profile.twitterId,
                  totalPower: profile.userGamedata?.totalPower,
                  stats: {
                    multiLiveCount: profile.userGamedata?.multiLiveCount,
                    mvpCount: profile.userGamedata?.mvpCount,
                    superStarCount: profile.userGamedata?.superStarCount,
                  },
                  assetSource: ASSET_SOURCE,
                  deckCards,
                  honors: profile.userHonors?.slice(0, 3).map((honor: any) => ({
                    honorId: honor.honorId,
                    level: honor.level,
                  })),
                }}
              />,
              { width: 800, height: 760 },
            ).catch((err) => {
              ctx.logger('moebot').warn(`Render /个人信息 failed: ${err?.message ?? err}`)
              return null
            })
            return png ? imageReply(png) : [
              `👤 ${profile.name}`,
              `Rank: ${profile.rank}`,
              `综合力: ${profile.userGamedata.totalPower}`,
              `协力次数: ${profile.userGamedata.multiLiveCount}`,
              `MVP: ${profile.userGamedata.mvpCount}`,
              `SS: ${profile.userGamedata.superStarCount}`,
            ].join('\n')
          }
        }

        const png = await renderer.render(
          `profile_basic_${user.platform}_${user.platformId}_${user.gameId}`,
          <ProfileCard
            profile={{
              name: user.nickname ?? '游戏信息',
              rank: 0,
              userId: user.gameId,
              bio: sekaiApi ? 'SEKAI API 暂未返回资料。' : '配置 SEKAI API 后可查看更详细的游戏数据。',
              assetSource: ASSET_SOURCE,
            }}
          />,
          { width: 800, height: 430 },
        ).catch((err) => {
          ctx.logger('moebot').warn(`Render basic /个人信息 failed: ${err?.message ?? err}`)
          return null
        })

        return png ? imageReply(png) : fallback
      } catch (err: any) {
        return `个人信息查询出错: ${err.message}`
      }
    })
}

function registerHelpCommand(ctx: Context, services: CommandServices) {
  const { renderer } = services

  ctx.command('帮助', '显示 Moebot NEXT 帮助信息')
    .alias('help')
    .alias('菜单')
    .action(async () => {
      const commands = getHelpCommands()
      const fallback = formatHelpText(commands)
      const png = await renderer.render(
        'help_card_v2',
        <HelpCard version="0.1.0" commands={commands} />,
        { width: 800, height: 880 },
      ).catch((err) => {
        ctx.logger('moebot').warn(`Render /帮助 failed: ${err?.message ?? err}`)
        return null
      })
      return png ? imageReply(png) : fallback
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

function imageReply(buffer: Buffer) {
  return h.image(buffer, 'image/png')
}

function normalized(value: unknown): string {
  return String(value ?? '').trim().toLowerCase()
}

function searchCards(masterdata: MasterdataService, keyword: string): any[] {
  const q = normalized(keyword)
  return masterdata.cards.filter((card: any) => {
    const char = getGameCharacter(masterdata, card.characterId)
    const searchStr = `${card.id} ${card.prefix ?? ''} ${card.assetbundleName ?? ''} ${characterName(char)} ${char?.givenName ?? ''} ${char?.firstName ?? ''}`
    return normalized(searchStr).includes(q)
  })
}

function searchMusics(masterdata: MasterdataService, keyword: string): any[] {
  const q = normalized(keyword)
  return masterdata.musics.filter((m: any) =>
    normalized(`${m.id}`).includes(q) ||
    normalized(m.title).includes(q) ||
    normalized(m.pronunciation).includes(q) ||
    normalized(m.assetbundleName).includes(q),
  )
}

function searchEvents(masterdata: MasterdataService, keyword?: string): any[] {
  if (!keyword) return getSortedByStart(masterdata.events).slice(0, 1)
  const q = normalized(keyword)
  return masterdata.events.filter((e: any) =>
    normalized(`${e.id}`).includes(q) ||
    normalized(e.name).includes(q) ||
    normalized(e.assetbundleName).includes(q),
  )
}

function searchGachas(masterdata: MasterdataService, keyword?: string): any[] {
  const gachas = masterdata.gachas ?? []
  if (!keyword) return getSortedByStart(gachas, 'startAt').slice(0, 1)
  const q = normalized(keyword)
  return gachas.filter((g: any) =>
    normalized(`${g.id}`).includes(q) ||
    normalized(g.name).includes(q) ||
    normalized(g.assetbundleName).includes(q),
  )
}

function getSortedByStart(items: any[], field = 'startAt'): any[] {
  return [...items].sort((a: any, b: any) => Number(b[field] ?? 0) - Number(a[field] ?? 0))
}

function getLatestEvent(masterdata: MasterdataService): any | null {
  return getSortedByStart(masterdata.events).find((event: any) => Number(event.startAt ?? 0) <= Date.now())
    ?? getSortedByStart(masterdata.events)[0]
    ?? null
}

function getGameCharacter(masterdata: MasterdataService, id: number): any | undefined {
  return masterdata.gameCharacters.find((c: any) => Number(c.id) === Number(id))
}

function characterName(char: any): string {
  if (!char) return '未知'
  return char.givenName ?? char.firstName ?? char.fullName ?? char.name ?? `角色 #${char.id}`
}

function calcCardPower(card: any): number | undefined {
  if (typeof card.power === 'number') return card.power
  if (Array.isArray(card.cardParameters)) {
    return card.cardParameters.reduce((sum: number, item: any) => sum + Number(item.power ?? item.param ?? 0), 0)
  }
  return undefined
}

function findSkillName(masterdata: MasterdataService, skillId: number | undefined): string | undefined {
  if (!skillId) return undefined
  return (masterdata.data.skills ?? []).find((skill: any) => Number(skill.id) === Number(skillId))?.name
}

function getEventBonus(masterdata: MasterdataService, eventId: number): { attr?: string; characters: string[] } {
  const bonuses = masterdata.data.eventDeckBonuses?.filter((bonus: any) => Number(bonus.eventId) === Number(eventId)) ?? []
  const attr = bonuses.find((bonus: any) => bonus.cardAttr)?.cardAttr
  const characters = bonuses
    .map((bonus: any) => {
      const unit = masterdata.data.gameCharacterUnits?.find((item: any) => Number(item.id) === Number(bonus.gameCharacterUnitId))
      const char = getGameCharacter(masterdata, unit?.gameCharacterId ?? bonus.gameCharacterId)
      return characterName(char)
    })
    .filter((name: string) => name && name !== '未知')
  return { attr, characters: Array.from(new Set(characters)) }
}

function getGachaPickupCards(masterdata: MasterdataService, gacha: any) {
  const detailCards = Array.isArray(gacha.gachaDetails) ? gacha.gachaDetails : []
  const externalCards = (masterdata.data.gachaCards ?? []).filter((item: any) => Number(item.gachaId) === Number(gacha.id))
  const details = [...detailCards, ...externalCards]
    .filter((detail: any) => detail.cardId)
    .sort((a: any, b: any) => Number(b.weight ?? 0) - Number(a.weight ?? 0))

  const seen = new Set<number>()
  return details
    .map((detail: any) => {
      const card = masterdata.cards.find((item: any) => Number(item.id) === Number(detail.cardId))
      if (!card || seen.has(card.id)) return null
      seen.add(card.id)
      const char = getGameCharacter(masterdata, card.characterId)
      return {
        id: card.id,
        prefix: card.prefix,
        characterName: characterName(char),
        rarity: card.cardRarityType ?? card.rarity ?? 'rarity_1',
        attr: card.attr ?? 'cute',
        assetbundleName: card.assetbundleName,
        thumbnailUrl: card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, ASSET_SOURCE, 'png') : undefined,
        trainedThumbnailUrl: card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, true, ASSET_SOURCE, 'png') : undefined,
        isWish: Boolean(detail.isWish),
        weight: detail.weight,
      }
    })
    .filter(Boolean)
    .slice(0, 8) as any[]
}

function buildProfileDeckCards(masterdata: MasterdataService, profile: any) {
  const activeDeck = profile.userDecks?.[0]
  if (!activeDeck) return []
  const ids = [activeDeck.leader, activeDeck.member1, activeDeck.member2, activeDeck.member3, activeDeck.member4, activeDeck.member5].filter(Boolean)
  return ids.map((cardId: number) => {
    const userCard = profile.userCards?.find((item: any) => Number(item.cardId) === Number(cardId))
    const card = masterdata.cards.find((item: any) => Number(item.id) === Number(cardId))
    const char = getGameCharacter(masterdata, card?.characterId)
    return {
      cardId,
      characterName: characterName(char),
      rarity: card?.cardRarityType ?? 'rarity_1',
      attr: card?.attr ?? 'cute',
      assetbundleName: card?.assetbundleName,
      isTrained: userCard?.specialTrainingStatus === 'done' || userCard?.defaultImage === 'special_training',
      mastery: userCard?.masterRank,
      level: userCard?.level,
      thumbnailUrl: card?.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, ASSET_SOURCE, 'png') : undefined,
      trainedThumbnailUrl: card?.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, true, ASSET_SOURCE, 'png') : undefined,
    }
  })
}

function normalizeRankingData(data: any): any[] {
  const list = Array.isArray(data)
    ? data
    : data.rankings ?? data.ranking ?? data.eventRankings ?? data.data ?? []
  return (Array.isArray(list) ? list : [])
    .map((entry: any) => ({
      rank: Number(entry.rank ?? entry.targetRank ?? 0),
      displayName: entry.displayName ?? entry.name ?? entry.userName,
      signature: entry.signature ?? entry.comment,
      score: Number(entry.score ?? entry.eventPoint ?? entry.point ?? 0),
      userId: entry.userId,
      scoreDelta: entry.scoreDelta,
      rankDelta: entry.rankDelta,
      leaderCharacterId: entry.leaderCharacterId,
      leaderCard: entry.leaderCardId ? { cardId: entry.leaderCardId, defaultImage: entry.leaderCardDefaultImage } : entry.leaderCard,
    }))
    .filter((entry: any) => entry.rank > 0 && entry.score >= 0)
}

function formatCardText(card: any, char: any, total: number): string {
  return [
    `🎴 ${card.prefix}`,
    `角色: ${characterName(char)}`,
    `稀有度: ${card.cardRarityType}`,
    `属性: ${card.attr}`,
    `ID: ${card.id}`,
    total > 1 ? `\n还有 ${total - 1} 张相关卡牌` : '',
  ].filter(Boolean).join('\n')
}

function formatMusicText(music: any, total: number): string {
  return [
    `🎵 ${music.title}`,
    music.pronunciation ? `读音: ${music.pronunciation}` : '',
    `作词: ${music.lyricist}`,
    `作曲: ${music.composer}`,
    `编曲: ${music.arranger}`,
    `ID: ${music.id}`,
    total > 1 ? `\n还有 ${total - 1} 首相关曲目` : '',
  ].filter(Boolean).join('\n')
}

function formatEventText(event: any): string {
  const startDate = new Date(normalizeTimestamp(event.startAt)).toLocaleDateString('zh-CN')
  const endDate = new Date(normalizeTimestamp(event.closedAt)).toLocaleDateString('zh-CN')
  return [
    `📅 ${event.name}`,
    `类型: ${event.eventType}`,
    `时间: ${startDate} ~ ${endDate}`,
    `ID: ${event.id}`,
  ].join('\n')
}

function formatGachaText(gacha: any, pickupCount: number): string {
  return [
    `🎰 ${gacha.name}`,
    `类型: ${gacha.gachaType ?? '招募'}`,
    `时间: ${new Date(normalizeTimestamp(gacha.startAt)).toLocaleDateString('zh-CN')} ~ ${new Date(normalizeTimestamp(gacha.endAt)).toLocaleDateString('zh-CN')}`,
    `Pickup: ${pickupCount} 张`,
    `ID: ${gacha.id}`,
  ].join('\n')
}

function getHelpCommands() {
  return [
    { name: '查卡', usage: '<关键词>', description: '搜索卡牌，支持角色名、卡名与 ID。' },
    { name: '查曲', usage: '<关键词>', description: '搜索曲目，支持别名、日文、罗马音与模糊匹配。' },
    { name: '查活动', usage: '[关键词/ID]', description: '查询活动信息、活动类型与时间范围。' },
    { name: '查卡池', usage: '[关键词/ID]', description: '查询招募卡池、开放时间与 pickup 卡。' },
    { name: '排行', usage: '[排名]', description: '查询活动实时排行榜（需要 SEKAI API）。' },
    { name: '绑定', usage: '<游戏ID>', description: '绑定 Project SEKAI 游戏账号。' },
    { name: '个人信息', usage: '', description: '查看已绑定账号的玩家资料。' },
    { name: '表情', usage: '<编号>', description: '发送 PJSK 表情贴纸。' },
  ]
}

function formatHelpText(commands: ReturnType<typeof getHelpCommands>): string {
  return [
    '🤖 Moebot NEXT — PJSK 查询助手',
    '',
    ...commands.map(command => `/${command.name} ${command.usage}  ${command.description}`),
    '',
    '💡 Powered by pjsk.moe',
  ].join('\n')
}

function normalizeTimestamp(timestamp: number): number {
  return timestamp < 1_000_000_000_000 ? timestamp * 1000 : timestamp
}

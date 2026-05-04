import { renderWithTrace } from './engine'
import { listRenderPreviews, renderPreviewTemplate } from './preview'
import { CardDetail, ChartDetail, ChurnRankingList, EventInfo, GachaInfo, GachaResult, HelpCard, MusicDetail, ProfileCard, RankingList } from './templates'

interface RenderRequest {
  template: string
  data: any
  width?: number
  height?: number
  precision?: number
}

const port = Number(process.env.PORT ?? 3001)
const defaultPrecision = parsePositiveNumber(process.env.RENDER_PRECISION, 1.5)

function parsePositiveNumber(value: unknown, fallback = 0): number {
  const numberValue = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(numberValue) && numberValue > 0 ? numberValue : fallback
}

function defaultHelpData() {
  return {
    commands: [
      { name: '查卡', usage: '/查卡 初音未来', description: '搜索卡牌信息，支持角色名/卡名/ID 模糊匹配' },
      { name: '查曲/查歌', usage: '/查歌 千本樱', description: '搜索曲目信息，支持别名/日文/罗马音' },
      { name: '查谱', usage: '/查谱 千本樱', description: '查询谱面等级与 notes' },
      { name: '查活动', usage: '/查活动 周年', description: '查询活动时间、类型、加成等信息' },
      { name: '查卡池/查扭蛋', usage: '/查扭蛋 限定', description: '查询扭蛋和 Pick Up 信息' },
      { name: '绑定', usage: '/绑定 123456789', description: '绑定 PJSK 游戏账号' },
      { name: '生日', usage: '/生日', description: '查看今日和近期角色生日' },
    ],
    version: '0.1.0',
  }
}

function toPngImageUrl(value: any): string | undefined {
  if (typeof value !== 'string' || value.length === 0) return undefined
  return value.replace(/\.webp(?=([?#]|$))/i, '.png')
}

function sanitizeImageUrls<T>(value: T): T {
  if (Array.isArray(value)) {
    return value.map(item => sanitizeImageUrls(item)) as T
  }
  if (!value || typeof value !== 'object') {
    return value
  }
  const out: Record<string, any> = {}
  for (const [key, item] of Object.entries(value as Record<string, any>)) {
    out[key] = typeof item === 'string' && /url$/i.test(key) ? toPngImageUrl(item) : sanitizeImageUrls(item)
  }
  return out as T
}

function normalizeCard(data: any) {
  return {
    id: data.id ?? data.ID ?? 0,
    prefix: data.prefix ?? data.Prefix ?? '未知卡牌',
    characterName: data.characterName ?? data.CharacterName ?? `角色 ${data.characterId ?? data.CharacterID ?? '?'}`,
    rarity: data.rarity ?? data.cardRarityType ?? data.CardRarityType ?? 'rarity_unknown',
    attr: data.attr ?? data.Attr ?? 'cute',
    thumbnailUrl: toPngImageUrl(data.thumbnailUrl ?? data.ThumbnailURL),
    normalThumbnailUrl: toPngImageUrl(data.normalThumbnailUrl ?? data.NormalThumbnailURL),
    trainedThumbnailUrl: toPngImageUrl(data.trainedThumbnailUrl ?? data.TrainedThumbnailURL ?? data.TrainedThumbnail),
    normalFullUrl: toPngImageUrl(data.normalFullUrl ?? data.NormalFullURL),
    trainedFullUrl: toPngImageUrl(data.trainedFullUrl ?? data.TrainedFullURL),
    assetbundleName: data.assetbundleName ?? data.AssetbundleName,
    characterId: data.characterId ?? data.CharacterID,
    cardRarityType: data.cardRarityType ?? data.CardRarityType,
    assetSource: data.assetSource ?? data.AssetSource,
    power: data.power ?? data.Power,
    skillName: data.skillName ?? data.SkillName ?? data.cardSkillName ?? data.CardSkillName,
    gachaPhrase: data.gachaPhrase ?? data.GachaPhrase,
    supplyType: data.supplyType ?? data.SupplyType,
  }
}

function normalizeMusic(data: any) {
  return {
    id: data.id ?? data.ID ?? 0,
    title: data.title ?? data.Title ?? '未知曲目',
    pronunciation: data.pronunciation ?? data.Pronunciation,
    lyricist: data.lyricist ?? data.Lyricist,
    composer: data.composer ?? data.Composer,
    arranger: data.arranger ?? data.Arranger,
    categories: data.categories ?? data.Categories ?? [],
    assetbundleName: data.assetbundleName ?? data.AssetbundleName,
    jacketUrl: toPngImageUrl(data.jacketUrl ?? data.JacketURL),
    assetSource: data.assetSource ?? data.AssetSource,
    difficulties: data.difficulties ?? data.Difficulties ?? [],
    publishedAt: data.publishedAt ?? data.PublishedAt,
    releasedAt: data.releasedAt ?? data.ReleasedAt,
    durationSec: data.durationSec ?? data.DurationSec ?? data.secForMusicScoreMaker ?? data.SecForMusicScoreMaker,
    isNewlyWrittenMusic: data.isNewlyWrittenMusic ?? data.IsNewlyWrittenMusic,
    isFullLength: data.isFullLength ?? data.IsFullLength,
  }
}

function normalizeEvent(data: any) {
  return {
    id: data.id ?? data.ID ?? 0,
    name: data.name ?? data.Name ?? '未知活动',
    eventType: data.eventType ?? data.EventType,
    unit: data.unit ?? data.Unit,
    assetbundleName: data.assetbundleName ?? data.AssetbundleName,
    assetSource: data.assetSource ?? data.AssetSource,
    bannerUrl: toPngImageUrl(data.bannerUrl ?? data.BannerURL),
    logoUrl: toPngImageUrl(data.logoUrl ?? data.LogoURL),
    startAt: data.startAt ?? data.StartAt,
    aggregateAt: data.aggregateAt ?? data.AggregateAt,
    closedAt: data.closedAt ?? data.ClosedAt,
    distributionEndAt: data.distributionEndAt ?? data.DistributionEndAt,
    deckBonuses: data.deckBonuses ?? data.DeckBonuses ?? [],
    bonusAttr: data.bonusAttr ?? data.BonusAttr,
    bonusCharacters: data.bonusCharacters ?? data.BonusCharacters ?? deriveBonusCharacters(data.deckBonuses ?? data.DeckBonuses ?? []),
  }
}

function normalizeRankingList(data: any) {
  return {
    title: data.title ?? data.Title ?? '活动榜线',
    subtitle: data.subtitle ?? data.Subtitle,
    rankings: data.rankings ?? data.Rankings ?? [],
    eventId: data.eventId ?? data.EventID,
    eventName: data.eventName ?? data.EventName,
    updatedAt: data.updatedAt ?? data.UpdatedAt,
    assetSource: data.assetSource ?? data.AssetSource,
  }
}

function normalizeProfile(data: any) {
  return {
    name: data.name ?? data.Name ?? '未知玩家',
    rank: data.rank ?? data.Rank ?? 0,
    userId: data.userId ?? data.UserID ?? data.userID ?? data.ID ?? data.id ?? '-',
    twitterId: data.twitterId ?? data.TwitterID,
    bio: data.bio ?? data.Bio,
    signature: data.signature ?? data.Signature,
    totalPower: data.totalPower ?? data.TotalPower,
    characterId: data.characterId ?? data.CharacterID,
    avatarUrl: toPngImageUrl(data.avatarUrl ?? data.AvatarURL),
    assetSource: data.assetSource ?? data.AssetSource,
    stats: data.stats ?? data.Stats,
    musicClearCounts: data.musicClearCounts ?? data.MusicClearCounts,
    characterRanks: data.characterRanks ?? data.CharacterRanks,
    challengeLive: data.challengeLive ?? data.ChallengeLive,
    profileHonors: data.profileHonors ?? data.ProfileHonors,
    leaderCard: data.leaderCard ?? data.LeaderCard,
    deckCards: data.deckCards ?? data.DeckCards,
    honors: data.honors ?? data.Honors,
  }
}

function normalizeGacha(data: any) {
  const pickups = data.pickups ?? data.Pickups ?? data.gachaPickups ?? data.GachaPickups ?? []
  return {
    id: data.id ?? data.ID ?? 0,
    name: data.name ?? data.Name ?? '未知卡池',
    gachaType: data.gachaType ?? data.GachaType,
    assetbundleName: data.assetbundleName ?? data.AssetbundleName,
    assetSource: data.assetSource ?? data.AssetSource,
    logoUrl: toPngImageUrl(data.logoUrl ?? data.LogoURL),
    bannerUrl: toPngImageUrl(data.bannerUrl ?? data.BannerURL),
    screenUrl: toPngImageUrl(data.screenUrl ?? data.ScreenURL),
    startAt: data.startAt ?? data.StartAt,
    endAt: data.endAt ?? data.EndAt,
    isShowPeriod: data.isShowPeriod ?? data.IsShowPeriod,
    wishSelectCount: data.wishSelectCount ?? data.WishSelectCount,
    pickupCards: normalizeGachaPickupCards(data.pickupCards ?? data.PickupCards, pickups),
    pickups,
    rates: data.rates ?? data.Rates ?? data.gachaCardRarityRates ?? data.GachaCardRarityRates ?? [],
  }
}

function deriveBonusCharacters(deckBonuses: any): string[] {
  if (!Array.isArray(deckBonuses)) return []
  const seen = new Set<string>()
  const result: string[] = []
  for (const bonus of deckBonuses) {
    const name = bonus?.characterName ?? bonus?.CharacterName
    if (typeof name === 'string' && name && !seen.has(name)) {
      seen.add(name)
      result.push(name)
    }
  }
  return result
}

function normalizeGachaPickupCards(directCards: any, pickups: any): any[] {
  if (Array.isArray(directCards)) {
    return directCards.map(normalizeGachaPickupCard)
  }
  if (!Array.isArray(pickups)) return []
  return pickups
    .map((pickup: any) => pickup?.card ?? pickup?.Card)
    .filter(Boolean)
    .map(normalizeGachaPickupCard)
}

function normalizeGachaPickupCard(card: any) {
  card = card ?? {}
  const characterId = card.characterId ?? card.CharacterID
  return {
    id: card.id ?? card.ID ?? card.cardId ?? card.CardID ?? 0,
    prefix: card.prefix ?? card.Prefix,
    characterName: card.characterName ?? card.CharacterName ?? `角色 ${characterId ?? '?'}`,
    rarity: card.rarity ?? card.cardRarityType ?? card.CardRarityType ?? 'rarity_unknown',
    cardRarityType: card.cardRarityType ?? card.CardRarityType,
    attr: card.attr ?? card.Attr ?? 'cute',
    assetbundleName: card.assetbundleName ?? card.AssetbundleName,
    characterId,
    thumbnailUrl: toPngImageUrl(card.thumbnailUrl ?? card.ThumbnailURL),
    trainedThumbnailUrl: toPngImageUrl(card.trainedThumbnailUrl ?? card.TrainedThumbnailURL),
    isWish: card.isWish ?? card.IsWish ?? true,
    weight: card.weight ?? card.Weight,
  }
}

function createElement(req: RenderRequest) {
  const data = sanitizeImageUrls(req.data)
  switch (req.template) {
    case 'help_card':
    case 'help':
      return <HelpCard {...(data ?? defaultHelpData())} />
    case 'card_detail':
    case 'card':
      return <CardDetail card={normalizeCard(data)} />
    case 'music_detail':
    case 'music':
      return <MusicDetail music={normalizeMusic(data)} />
    case 'chart_detail':
    case 'chart':
      return <ChartDetail music={normalizeMusic(data)} />
    case 'event_info':
    case 'event':
      return <EventInfo event={normalizeEvent(data)} />
    case 'gacha_info':
    case 'gacha':
      return <GachaInfo gacha={normalizeGacha(data)} />
    case 'gacha_result':
    case 'gacha-result':
      return <GachaResult {...(data ?? { pullType: 'multi', results: [] })} />
    case 'profile_card':
    case 'profile':
      return <ProfileCard profile={normalizeProfile(data)} />
    case 'ranking_list':
    case 'ranking':
      return <RankingList {...normalizeRankingList(data)} />
    case 'churn_ranking_list':
    case 'churn_ranking':
      return <ChurnRankingList {...normalizeRankingList(data)} />
    default:
      return <HelpCard {...defaultHelpData()} />
  }
}

Bun.serve({
  port,
  async fetch(request) {
    const url = new URL(request.url)

    if (url.pathname === '/' || url.pathname === '/health') {
      return Response.json({
        status: 'ok',
        service: 'moebot-next-renderer',
        version: '0.1.0',
        endpoints: ['GET /health', 'GET /previews', 'GET /preview/:id', 'POST /render'],
        note: '这是内部 Satori 渲染服务；管理面板请访问 http://127.0.0.1:8080/',
      })
    }

    if (url.pathname === '/previews' && request.method === 'GET') {
      return Response.json({
        data: listRenderPreviews(),
        total: listRenderPreviews().length,
      })
    }

    if (url.pathname.startsWith('/preview/') && request.method === 'GET') {
      try {
        const id = decodeURIComponent(url.pathname.replace('/preview/', '').replace(/\/$/, ''))
        const width = Number(url.searchParams.get('width') || 0)
        const height = Number(url.searchParams.get('height') || 0)
        const precision = parsePositiveNumber(url.searchParams.get('precision'), defaultPrecision)
        const result = await renderPreviewTemplate(id, {
          ...(width > 0 ? { width } : {}),
          ...(height > 0 ? { height } : {}),
          precision,
        })
        return new Response(new Uint8Array(result.trace.png), {
          headers: {
            'content-type': 'image/png',
            'cache-control': 'no-store',
            'x-render-total-ms': String(result.trace.timings.totalMs),
            'x-render-fonts-ms': String(result.trace.timings.fontsMs),
            'x-render-satori-ms': String(result.trace.timings.satoriMs),
            'x-render-resvg-ms': String(result.trace.timings.resvgMs),
            'x-render-size-bytes': String(result.trace.sizeBytes),
          },
        })
      } catch (error) {
        console.error('[renderer] preview render failed:', error)
        return Response.json({ error: true, message: error instanceof Error ? error.message : String(error) }, { status: 500 })
      }
    }

    if (url.pathname === '/render' && request.method === 'POST') {
      try {
        const body = await request.json() as RenderRequest
        const trace = await renderWithTrace(createElement(body), {
          width: body.width ?? 800,
          height: body.height,
          precision: parsePositiveNumber(body.precision, defaultPrecision),
        })
        return new Response(new Uint8Array(trace.png), {
          headers: {
            'content-type': 'image/png',
            'cache-control': 'no-store',
            'x-render-total-ms': String(trace.timings.totalMs),
            'x-render-fonts-ms': String(trace.timings.fontsMs),
            'x-render-satori-ms': String(trace.timings.satoriMs),
            'x-render-resvg-ms': String(trace.timings.resvgMs),
            'x-render-size-bytes': String(trace.sizeBytes),
          },
        })
      } catch (error) {
        console.error('[renderer] render failed:', error)
        return Response.json({ error: true, message: error instanceof Error ? error.message : String(error) }, { status: 500 })
      }
    }

    return Response.json({ error: true, message: 'not found' }, { status: 404 })
  },
})

console.log(`[renderer] Moebot renderer listening on http://127.0.0.1:${port}`)

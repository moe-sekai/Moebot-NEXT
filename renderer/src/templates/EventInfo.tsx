import {
  getCardThumbnailUrl,
  getEventBannerUrl,
  getEventLogoUrl,
  type AssetSourceType,
} from '../../shared'
import { BaseCard } from './base'
import { getLocalIconAssetDataUri } from '../styles/assets'
import { theme } from '../styles/theme'
import { SekaiCardThumbnail, canUseTrainedArt } from './SekaiCardThumbnail'

const EVENT_TYPE_NAMES: Record<string, string> = {
  marathon: '马拉松',
  cheerful_carnival: '欢乐嘉年华',
  world_bloom: '世界绽放',
}

const EVENT_TYPE_COLORS: Record<string, string> = {
  marathon: '#F06292',
  cheerful_carnival: '#FFB74D',
  world_bloom: '#81C784',
}

const EVENT_STATUS_DISPLAY: Record<string, { label: string; color: string }> = {
  upcoming: { label: '即将开始', color: '#42A5F5' },
  ongoing: { label: '进行中', color: '#66BB6A' },
  ended: { label: '已结束', color: '#9E9E9E' },
}

const UNIT_LABELS: Record<string, string> = {
  piapro: 'VIRTUAL SINGER',
  light_sound: 'Leo/need',
  idol: 'MORE MORE JUMP!',
  street: 'Vivid BAD SQUAD',
  theme_park: 'Wonderlands×Showtime',
  school_refusal: '25時、ナイトコードで。',
  vs: 'VIRTUAL SINGER',
  ln: 'Leo/need',
  mmj: 'MORE MORE JUMP!',
  vbs: 'Vivid BAD SQUAD',
  wxs: 'Wonderlands×Showtime',
  n25: '25時、ナイトコードで。',
}

export interface EventInfoProps {
  event: {
    id: number
    name: string
    eventType: string
    assetbundleName?: string
    assetSource?: AssetSourceType | string
    bannerUrl?: string
    storyBannerUrl?: string
    logoUrl?: string
    characterUrl?: string
    startAt: number
    aggregateAt?: number
    closedAt: number
    distributionEndAt?: number
    unit?: string
    bonusAttr?: string
    bonusCharacters?: string[]
    bonusCards?: Array<{
      id: number
      prefix?: string
      characterName?: string
      rarity?: string
      cardRarityType?: string
      attr?: string
      assetbundleName?: string
      thumbnailUrl?: string
      trainedThumbnailUrl?: string
      supplyType?: string
    }>
  }
}

export function EventInfo({ event }: EventInfoProps) {
  const source = event.assetSource ?? 'main-jp'
  const accent = EVENT_TYPE_COLORS[event.eventType] ?? theme.colors.accent
  const status = getEventStatus(event)
  const statusDisplay = EVENT_STATUS_DISPLAY[status]
  const bannerUrl = event.bannerUrl
    ?? (event.assetbundleName ? getEventBannerUrl(event.assetbundleName, source) : undefined)
  const logoUrl = event.logoUrl
    ?? (event.assetbundleName ? getEventLogoUrl(event.assetbundleName, source) : undefined)
  const unitLabel = event.unit ? UNIT_LABELS[event.unit] ?? event.unit : undefined
  const unitLogoUrl = event.unit ? resolveUnitLogo(event.unit, source) : undefined

  return (
    <BaseCard
      title={event.name}
      subtitle={`${EVENT_TYPE_NAMES[event.eventType] ?? event.eventType} · ID: ${event.id}`}
      accentColor={accent}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.lg }}>
        <div
          style={{
            display: 'flex',
            position: 'relative',
            width: '100%',
            height: 252,
            borderRadius: theme.borderRadius.xl,
            overflow: 'hidden',
            backgroundColor: theme.colors.surface,
            border: `1px solid ${theme.colors.border}`,
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <img
            src={bannerUrl ?? placeholderImage('EVENT', accent, 1472, 504)}
            width={736}
            height={252}
            style={{ position: 'absolute', inset: 0, objectFit: 'cover', width: '100%', height: '100%' }}
          />
          {logoUrl && (
            <div
              style={{
                display: 'flex',
                position: 'relative',
                width: 460,
                height: 150,
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <img src={logoUrl} width={460} height={150} style={{ objectFit: 'contain' }} />
            </div>
          )}
        </div>

        <div style={{ display: 'flex', gap: theme.spacing.md }}>
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              flex: 1,
              gap: theme.spacing.sm,
              backgroundColor: theme.colors.surface,
              border: `1px solid ${theme.colors.border}`,
              borderRadius: theme.borderRadius.xl,
              padding: theme.spacing.md,
            }}
          >
            <SectionTitle title="活动时间" color={accent} />
            <InfoRow label="开始" value={formatDateTime(event.startAt)} />
            <InfoRow label="结算" value={formatDateTime(event.aggregateAt ?? event.closedAt)} />
            <InfoRow label="关闭" value={formatDateTime(event.closedAt)} />
            {event.distributionEndAt && <InfoRow label="领取截止" value={formatDateTime(event.distributionEndAt)} />}
          </div>

          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              width: 260,
              gap: theme.spacing.sm,
              backgroundColor: theme.colors.surface,
              border: `1px solid ${theme.colors.border}`,
              borderRadius: theme.borderRadius.xl,
              padding: theme.spacing.md,
            }}
          >
            <SectionTitle title="加成信息" color={accent} />
            {unitLabel ? (
              <div style={{ display: 'flex', alignItems: 'center', gap: theme.spacing.sm }}>
                {unitLogoUrl && <img src={unitLogoUrl} width={36} height={36} style={{ objectFit: 'contain' }} />}
                <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.sm, fontWeight: 900 }}>{unitLabel}</span>
              </div>
            ) : (
              <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.sm }}>未标注团组</span>
            )}
            {event.bonusAttr && <InfoRow label="属性" value={event.bonusAttr} />}
            {event.bonusCharacters && event.bonusCharacters.length > 0 && (
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: theme.spacing.xs }}>
                {event.bonusCharacters.slice(0, 6).map((name) => (
                  <span
                    key={name}
                    style={{
                      display: 'flex',
                      padding: '4px 8px',
                      borderRadius: theme.borderRadius.round,
                      backgroundColor: theme.colors.accentSoft,
                      color: theme.colors.textSecondary,
                      fontSize: theme.fontSize.xs,
                      fontWeight: 800,
                    }}
                  >
                    {name}
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>

        {event.bonusCards && event.bonusCards.length > 0 && (
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: theme.spacing.sm,
              backgroundColor: theme.colors.surface,
              border: `1px solid ${theme.colors.border}`,
              borderRadius: theme.borderRadius.xl,
              padding: theme.spacing.md,
            }}
          >
            <SectionTitle title="加成卡片" color={accent} />
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: theme.spacing.sm }}>
              {event.bonusCards.slice(0, 6).map((card) => {
                const rarity = card.cardRarityType ?? card.rarity ?? 'rarity_unknown'
                const trained = canUseTrainedArt(rarity)
                const imageUrl = card.thumbnailUrl ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, trained, source, 'png') : undefined)
                return (
                  <div key={card.id} style={{ display: 'flex', alignItems: 'center', gap: theme.spacing.xs, width: 224 }}>
                    <SekaiCardThumbnail imageUrl={imageUrl} rarity={rarity} attr={card.attr ?? 'cute'} isTrained={trained} characterName={card.characterName} supplyType={card.supplyType} size={64} />
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 2, flex: 1 }}>
                      <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.xs, fontWeight: 900 }}>#{card.id} {card.characterName ?? ''}</span>
                      <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: 10, lineHeight: 1.25 }}>{card.prefix ?? rarity}</span>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        )}
      </div>
    </BaseCard>
  )
}

function SectionTitle({ title, color }: { title: string; color: string }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: theme.spacing.sm, marginBottom: 2 }}>
      <div style={{ display: 'flex', width: 6, height: 18, borderRadius: theme.borderRadius.round, backgroundColor: color }} />
      <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>{title}</span>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', gap: theme.spacing.md, alignItems: 'center' }}>
      <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.sm, flexShrink: 0 }}>{label}</span>
      <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.sm, fontWeight: 800, textAlign: 'right' }}>{value}</span>
    </div>
  )
}

function Chip({ text, color, style }: { text: string; color: string; style: Record<string, number> }) {
  return (
    <div
      style={{
        display: 'flex',
        position: 'absolute',
        padding: '5px 11px',
        borderRadius: theme.borderRadius.round,
        backgroundColor: color,
        color: '#fff',
        fontSize: theme.fontSize.xs,
        fontWeight: 900,
        boxShadow: '0 3px 12px rgba(0,0,0,0.14)',
        ...style,
      }}
    >
      {text}
    </div>
  )
}

function getEventStatus(event: EventInfoProps['event']): 'upcoming' | 'ongoing' | 'ended' {
  const now = Date.now()
  if (now < normalizeTimestamp(event.startAt)) return 'upcoming'
  if (now > normalizeTimestamp(event.aggregateAt ?? event.closedAt)) return 'ended'
  return 'ongoing'
}

function normalizeTimestamp(timestamp: number): number {
  return timestamp < 1_000_000_000_000 ? timestamp * 1000 : timestamp
}

function formatDateTime(timestamp: number): string {
  return new Date(normalizeTimestamp(timestamp)).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function resolveUnitLogo(unit: string, source: AssetSourceType | string): string | undefined {
  if (!unit || unit === 'none') return undefined
  const localMap: Record<string, string> = {
    piapro: 'vs.png',
    light_sound: 'ln.png',
    idol: 'mmj.png',
    street: 'vbs.png',
    theme_park: 'wxs.png',
    school_refusal: 'n25.png',
    vs: 'vs.png',
    ln: 'ln.png',
    mmj: 'mmj.png',
    vbs: 'vbs.png',
    wxs: 'wxs.png',
    n25: 'n25.png',
  }
  return getLocalIconAssetDataUri(localMap[unit] ?? `${unit}.png`)
}

function placeholderImage(label: string, color: string, width: number, height: number): string {
  const safeLabel = escapeXml(label)
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">
  <defs><linearGradient id="g" x1="0" x2="1" y1="0" y2="1"><stop offset="0" stop-color="#ffffff"/><stop offset="1" stop-color="${color}" stop-opacity="0.35"/></linearGradient></defs>
  <rect width="${width}" height="${height}" rx="42" fill="url(#g)"/>
  <circle cx="${Math.round(width * 0.82)}" cy="${Math.round(height * 0.24)}" r="${Math.round(height * 0.36)}" fill="${color}" opacity="0.18"/>
  <text x="50%" y="52%" dominant-baseline="middle" text-anchor="middle" font-family="Arial, sans-serif" font-size="64" font-weight="900" fill="${color}">${safeLabel}</text>
</svg>`
  return `data:image/svg+xml;base64,${Buffer.from(svg, 'utf8').toString('base64')}`
}

function escapeXml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

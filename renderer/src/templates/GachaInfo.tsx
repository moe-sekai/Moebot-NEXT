import {
  getCardThumbnailUrl,
  getGachaLogoUrl,
  type AssetSourceType,
} from '../../shared'
import { BaseCard } from './base'
import { theme } from '../styles/theme'
import { SekaiCardThumbnail, canUseTrainedArt, getAttributeColor } from './SekaiCardThumbnail'

export interface GachaInfoProps {
  gacha: {
    id: number
    name: string
    gachaType?: string
    assetbundleName?: string
    assetSource?: AssetSourceType | string
    logoUrl?: string
    bannerUrl?: string
    screenUrl?: string
    startAt: number
    endAt: number
    isShowPeriod?: boolean
    wishSelectCount?: number
    pickupCards?: Array<{
      id: number
      prefix?: string
      characterName: string
      rarity: string
      attr: string
      assetbundleName?: string
      thumbnailUrl?: string
      trainedThumbnailUrl?: string
      compositeLayers?: import('../card-thumbnail-composites').CardThumbnailCompositeLayer[]
      isWish?: boolean
      weight?: number
    }>
  }
}

export function GachaInfo({ gacha }: GachaInfoProps) {
  const source = gacha.assetSource ?? 'main-jp'
  const logoUrl = gacha.logoUrl
    ?? (gacha.assetbundleName ? getGachaLogoUrl(gacha.assetbundleName, source) : undefined)
  const pickups = gacha.pickupCards ?? []

  return (
    <BaseCard
      title={gacha.name}
      subtitle={`${formatGachaType(gacha.gachaType)} · ID: ${gacha.id}`}
      accentColor={theme.colors.warning}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.lg }}>
        <div
          style={{
            display: 'flex',
            position: 'relative',
            width: '100%',
            height: 250,
            borderRadius: theme.borderRadius.xl,
            overflow: 'hidden',
            backgroundColor: theme.colors.surfaceLight,
            border: `1px solid ${theme.colors.border}`,
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {logoUrl ? (
            <img src={logoUrl} width={560} height={170} style={{ objectFit: 'contain' }} />
          ) : (
            <img src={placeholderImage('GACHA', theme.colors.warning, 1120, 340)} width={560} height={170} style={{ objectFit: 'contain' }} />
          )}
        </div>

        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: theme.spacing.md,
            backgroundColor: theme.colors.surface,
            border: `1px solid ${theme.colors.border}`,
            borderRadius: theme.borderRadius.xl,
            padding: theme.spacing.md,
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>Pickup 卡牌</span>
            <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.xs }}>{pickups.length} cards</span>
          </div>

          {pickups.length > 0 ? (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: theme.spacing.sm }}>
              {pickups.slice(0, 8).map(card => {
                const isTrained = canUseTrainedArt(card.rarity)
                const thumbnailUrl = card.thumbnailUrl
                  ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, source, 'png') : undefined)
                const trainedThumbnailUrl = card.trainedThumbnailUrl
                  ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, true, source, 'png') : undefined)
                const attrColor = getAttributeColor(card.attr)
                return (
                  <div
                    key={card.id}
                    style={{
                      display: 'flex',
                      flexDirection: 'column',
                      width: 162,
                      gap: 6,
                      padding: 7,
                      borderRadius: theme.borderRadius.lg,
                      backgroundColor: theme.colors.surfaceLight,
                      border: `1px solid ${card.isWish ? theme.colors.warning : theme.colors.border}`,
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'center', position: 'relative' }}>
                      <SekaiCardThumbnail
                        imageUrl={isTrained ? trainedThumbnailUrl ?? thumbnailUrl : thumbnailUrl}
                        compositeLayers={card.compositeLayers}
                        rarity={card.rarity}
                        attr={card.attr}
                        isTrained={isTrained}
                        characterName={card.characterName}
                        size={112}
                      />
                      {card.isWish && <SmallBadge text="PICK" color={theme.colors.warning} />}
                    </div>
                    <span style={{ display: 'flex', justifyContent: 'center', color: theme.colors.text, fontSize: theme.fontSize.xs, fontWeight: 900, textAlign: 'center' }}>
                      {card.characterName}
                    </span>
                    <span style={{ display: 'flex', justifyContent: 'center', color: attrColor, fontSize: 11, fontWeight: 800 }}>
                      #{card.id}{typeof card.weight === 'number' ? ` · ${card.weight}` : ''}
                    </span>
                  </div>
                )
              })}
            </div>
          ) : (
            <div
              style={{
                display: 'flex',
                justifyContent: 'center',
                padding: theme.spacing.lg,
                borderRadius: theme.borderRadius.lg,
                backgroundColor: theme.colors.surfaceLight,
                color: theme.colors.textMuted,
                fontSize: theme.fontSize.sm,
              }}
            >
              暂无 pickup 明细，等待 masterdata 补充。
            </div>
          )}
        </div>
      </div>
    </BaseCard>
  )
}

function SmallBadge({ text, color }: { text: string; color: string }) {
  return (
    <div
      style={{
        display: 'flex',
        position: 'absolute',
        right: 16,
        top: -4,
        padding: '3px 7px',
        borderRadius: theme.borderRadius.round,
        backgroundColor: color,
        color: '#fff',
        fontSize: 10,
        fontWeight: 900,
      }}
    >
      {text}
    </div>
  )
}

function formatGachaType(type?: string): string {
  const map: Record<string, string> = {
    ceil: '招募',
    normal: '普通招募',
    birthday: '生日招募',
    limited: '限定招募',
  }
  return type ? map[type] ?? type : '招募'
}

function normalizeTimestamp(timestamp: number): number {
  return timestamp < 1_000_000_000_000 ? timestamp * 1000 : timestamp
}

function formatDate(timestamp: number): string {
  return new Date(normalizeTimestamp(timestamp)).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

function placeholderImage(label: string, color: string, width: number, height: number): string {
  const safeLabel = escapeXml(label)
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">
  <defs><linearGradient id="g" x1="0" x2="1" y1="0" y2="1"><stop offset="0" stop-color="#ffffff"/><stop offset="1" stop-color="${color}" stop-opacity="0.35"/></linearGradient></defs>
  <rect width="${width}" height="${height}" rx="42" fill="url(#g)"/>
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

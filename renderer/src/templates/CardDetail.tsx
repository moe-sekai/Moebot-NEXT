import { getCardFullUrl, getCardThumbnailUrl, type AssetSourceType } from '../../shared'
import { BaseCard } from './base'
import { theme } from '../styles/theme'
import { SekaiCardThumbnail, canUseTrainedArt, getAttributeColor } from './SekaiCardThumbnail'

interface CardDetailProps {
  card: {
    id: number
    prefix: string
    characterName: string
    rarity: string
    attr: string
    thumbnailUrl?: string
    normalThumbnailUrl?: string
    trainedThumbnailUrl?: string
    normalFullUrl?: string
    trainedFullUrl?: string
    assetbundleName?: string
    characterId?: number
    cardRarityType?: string
    assetSource?: AssetSourceType | string
    power?: number
    skillName?: string
    gachaPhrase?: string
    supplyType?: string
    trained?: boolean
  }
}

export function CardDetail({ card }: CardDetailProps) {
  const rarity = card.cardRarityType ?? card.rarity
  const attrColor = getAttributeColor(card.attr)
  const canRenderTrained = canUseTrainedArt(rarity)
  const source = card.assetSource ?? 'main-jp'

  const normalThumbnailUrl = card.normalThumbnailUrl
    ?? card.thumbnailUrl
    ?? resolveCardThumbnail(card.assetbundleName, false, source)
  const trainedThumbnailUrl = card.trainedThumbnailUrl
    ?? resolveCardThumbnail(card.assetbundleName, true, source)
  const normalFullUrl = card.normalFullUrl
    ?? resolveCardFull(card.assetbundleName, false, source)
    ?? normalThumbnailUrl
  const trainedFullUrl = card.trainedFullUrl
    ?? resolveCardFull(card.assetbundleName, true, source)
    ?? trainedThumbnailUrl
  const showTrained = canRenderTrained && Boolean(trainedFullUrl)

  return (
    <BaseCard
      title={card.prefix}
      subtitle={`${card.characterName} · ID: ${card.id} · ${rarity}`}
      accentColor={attrColor}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.md }}>
        <div style={{ display: 'flex', gap: theme.spacing.md }}>
          <ArtPanel title="花前卡面" subtitle="card_normal" imageUrl={normalFullUrl} accentColor={attrColor} />
          <ArtPanel
            title={showTrained ? '花后卡面' : '未开放花后'}
            subtitle={showTrained ? 'card_after_training' : 'normal only'}
            imageUrl={showTrained ? trainedFullUrl : normalFullUrl}
            accentColor={showTrained ? theme.colors.accent : theme.colors.textMuted}
            muted={!showTrained}
          />
        </div>

        <div style={{ display: 'flex', gap: theme.spacing.md, alignItems: 'stretch' }}>
          <div style={{ display: 'flex', gap: theme.spacing.sm, flexShrink: 0 }}>
            <SekaiCardThumbnail
              imageUrl={normalThumbnailUrl}
              rarity={rarity}
              attr={card.attr}
              isTrained={false}
              label="花前"
              characterName={card.characterName}
              size={128}
            />
            {showTrained && (
              <SekaiCardThumbnail
                imageUrl={trainedThumbnailUrl}
                rarity={rarity}
                attr={card.attr}
                isTrained
                label="花后"
                characterName={card.characterName}
                size={128}
              />
            )}
          </div>

          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: theme.spacing.sm,
              flex: 1,
              backgroundColor: theme.colors.surface,
              border: `1px solid ${theme.colors.border}`,
              borderRadius: theme.borderRadius.lg,
              padding: theme.spacing.md,
            }}
          >
            <InfoRow label="属性" value={card.attr} color={attrColor} />
            <InfoRow label="稀有度" value={rarity} />
            <InfoRow label="卡面资源" value={card.assetbundleName ? `${card.assetbundleName}_{normal|after_training}` : 'mock/offline'} />
            {card.power && <InfoRow label="综合力" value={card.power.toLocaleString()} />}
            {card.skillName && <InfoRow label="技能" value={card.skillName} />}
            {card.supplyType && <InfoRow label="获取类型" value={card.supplyType} />}
          </div>
        </div>

        {card.gachaPhrase && (
          <div
            style={{
              display: 'flex',
              padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
              borderRadius: theme.borderRadius.lg,
              backgroundColor: theme.colors.surfaceAccent,
              border: `1px solid ${theme.colors.borderStrong}`,
              color: theme.colors.textSecondary,
              fontSize: theme.fontSize.sm,
              lineHeight: 1.5,
            }}
          >
            “{card.gachaPhrase}”
          </div>
        )}
      </div>
    </BaseCard>
  )
}

function ArtPanel({
  title,
  subtitle,
  imageUrl,
  accentColor,
  muted = false,
}: {
  title: string
  subtitle: string
  imageUrl?: string
  accentColor: string
  muted?: boolean
}) {
  return (
    <div
      style={{
        display: 'flex',
        position: 'relative',
        flexDirection: 'column',
        width: 360,
        height: 214,
        borderRadius: theme.borderRadius.lg,
        overflow: 'hidden',
        backgroundColor: theme.colors.surface,
        border: `1px solid ${muted ? theme.colors.border : accentColor}`,
      }}
    >
      <img
        src={imageUrl ?? panelPlaceholder(title, accentColor)}
        width={360}
        height={214}
        style={{ objectFit: 'cover', objectPosition: 'center top', opacity: muted ? 0.72 : 1 }}
      />
      <div
        style={{
          display: 'flex',
          position: 'absolute',
          left: 12,
          bottom: 12,
          flexDirection: 'column',
          gap: 2,
          padding: '7px 11px',
          borderRadius: theme.borderRadius.md,
          backgroundColor: 'rgba(255,255,255,0.9)',
          border: `1px solid ${theme.colors.border}`,
        }}
      >
        <span style={{ display: 'flex', color: muted ? theme.colors.textMuted : accentColor, fontSize: theme.fontSize.sm, fontWeight: 900 }}>
          {title}
        </span>
        <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.xs }}>
          {subtitle}
        </span>
      </div>
    </div>
  )
}

function InfoRow({ label, value, color }: { label: string; value: string; color?: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: theme.spacing.md }}>
      <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>{label}</span>
      <span
        style={{
          display: 'flex',
          color: color ?? theme.colors.text,
          fontSize: theme.fontSize.sm,
          fontWeight: 800,
          textAlign: 'right',
          maxWidth: 330,
        }}
      >
        {value}
      </span>
    </div>
  )
}

function resolveCardThumbnail(assetbundleName: string | undefined, trained: boolean, source: AssetSourceType | string): string | undefined {
  return assetbundleName ? getCardThumbnailUrl(assetbundleName, trained, source, 'png') : undefined
}

function resolveCardFull(assetbundleName: string | undefined, trained: boolean, source: AssetSourceType | string): string | undefined {
  return assetbundleName ? getCardFullUrl(assetbundleName, trained, source, 'png') : undefined
}

function panelPlaceholder(label: string, color: string): string {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="720" height="428" viewBox="0 0 720 428">
  <defs>
    <linearGradient id="g" x1="0" x2="1" y1="0" y2="1">
      <stop offset="0" stop-color="#ffffff"/>
      <stop offset="1" stop-color="${color}" stop-opacity="0.36"/>
    </linearGradient>
  </defs>
  <rect width="720" height="428" rx="30" fill="url(#g)"/>
  <circle cx="584" cy="86" r="102" fill="${color}" opacity="0.2"/>
  <circle cx="126" cy="340" r="126" fill="#fff" opacity="0.55"/>
  <text x="50%" y="52%" dominant-baseline="middle" text-anchor="middle" font-family="Arial, sans-serif" font-size="54" font-weight="800" fill="${color}">${escapeXml(label)}</text>
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

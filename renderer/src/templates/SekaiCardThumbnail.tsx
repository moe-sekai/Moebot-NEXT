import type { CardThumbnailCompositeLayer } from '../card-thumbnail-composites'
import { getSekaiCardUiAssetDataUri } from '../styles/assets'
import { theme } from '../styles/theme'

export interface SekaiCardThumbnailProps {
  imageUrl?: string
  compositeImageUrl?: string
  compositeLayers?: CardThumbnailCompositeLayer[]
  rarity: string
  attr: string
  label?: string
  size?: number
  isTrained?: boolean
  mastery?: number
  characterName?: string
  supplyType?: string
}

export function getRarityNumber(rarityType: string): number {
  if (rarityType === 'rarity_birthday') return 1
  const matched = rarityType.match(/(\d+)/)
  return matched ? Number(matched[1]) : 1
}

export function canUseTrainedArt(rarityType: string): boolean {
  return rarityType === 'rarity_3' || rarityType === 'rarity_4'
}

export function getAttributeColor(attr: string): string {
  return (theme.colors as Record<string, string>)[attr] ?? theme.colors.accent
}

export function getAttributeLabel(attr: string): string {
  const labels: Record<string, string> = {
    cute: 'Cute',
    cool: 'Cool',
    pure: 'Pure',
    happy: 'Happy',
    mysterious: 'Myst',
  }
  return labels[attr] ?? attr
}

export function SekaiCardThumbnail({
  imageUrl,
  compositeImageUrl,
  compositeLayers,
  rarity,
  attr,
  label,
  size = 156,
  isTrained = false,
  mastery = 0,
  characterName,
  supplyType,
}: SekaiCardThumbnailProps) {
  const attrColor = getAttributeColor(attr)
  const starCount = getRarityNumber(rarity)
  const birthday = rarity === 'rarity_birthday'
  const raritySuffix = birthday ? 'birthday' : String(getRarityNumber(rarity))
  const frameUrl = getSekaiCardUiAssetDataUri(`frame_rarity_${raritySuffix}.png`)
  const attrUrl = getSekaiCardUiAssetDataUri(`attr_${attr}.png`)
  const starUrl = getSekaiCardUiAssetDataUri(birthday ? 'rare_birthday.png' : isTrained ? 'rare_star_after_training.png' : 'rare_star_normal.png')
  const masteryUrl = mastery > 0 ? getSekaiCardUiAssetDataUri(`train_rank_${mastery}.png`) : undefined
  const scale = size / 156
  const supplyBadge = supplyType ? getSupplyBadgeStyle(supplyType) : undefined
  const baseImageUrl = compositeImageUrl ?? imageUrl ?? placeholderCardImage(characterName ?? 'CARD', attrColor)
  const useComposite = Boolean(compositeImageUrl || compositeLayers?.length)
  // Composite layers are pre-rendered at a fixed size (encoded in the rect's width).
  // When the requested render size differs from that, scale every layer dimension
  // accordingly so the canvas always fills the container without leaving blank space.
  const layerBaseSize = compositeLayers?.find((l) => l.type === 'rect')?.width
  const layerScale = layerBaseSize && layerBaseSize > 0 ? size / layerBaseSize : 1

  return (
    <div
      style={{
        display: 'flex',
        position: 'relative',
        width: size,
        height: size,
        borderRadius: 12 * scale,
        overflow: 'hidden',
        backgroundColor: theme.colors.surface,
      }}
    >
      {compositeLayers?.length ? (
        <div style={{ display: 'flex', position: 'absolute', left: 0, top: 0, width: size, height: size }}>
          {compositeLayers.map((layer, index) => layer.type === 'rect' ? (
            <div
              key={index}
              style={{
                display: 'flex',
                position: 'absolute',
                left: 0,
                top: 0,
                width: layer.width * layerScale,
                height: layer.height * layerScale,
                borderRadius: (layer.rx ?? 0) * layerScale,
                backgroundColor: layer.fill ?? 'transparent',
              }}
            />
          ) : (
            <img
              key={index}
              src={layer.href}
              width={layer.width * layerScale}
              height={layer.height * layerScale}
              style={{
                position: 'absolute',
                left: layer.x * layerScale,
                top: layer.y * layerScale,
                width: layer.width * layerScale,
                height: layer.height * layerScale,
                objectFit: preserveAspectRatioToObjectFit(layer.preserveAspectRatio),
              }}
            />
          ))}
        </div>
      ) : (
        <img
          src={baseImageUrl}
          width={size}
          height={size}
          style={useComposite ? { position: 'absolute', left: 0, top: 0, width: size, height: size, objectFit: 'fill' } : { position: 'absolute', left: 2 * scale, top: 2 * scale, width: 152 * scale, height: 152 * scale, objectFit: 'cover' }}
        />
      )}

      {!useComposite && frameUrl && (
        <img
          src={frameUrl}
          width={size}
          height={size}
          style={{ position: 'absolute', left: 0, top: 0, width: size, height: size, objectFit: 'fill' }}
        />
      )}

      {!useComposite && attrUrl && (
        <img
          src={attrUrl}
          width={35 * scale}
          height={35 * scale}
          style={{ position: 'absolute', left: 0, top: 0, objectFit: 'contain' }}
        />
      )}

      {!useComposite && starUrl && Array.from({ length: starCount }).map((_, index) => (
        <img
          key={index}
          src={starUrl}
          width={24 * scale}
          height={24 * scale}
          style={{
            position: 'absolute',
            left: (birthday ? 10 : 5 + index * 24) * scale,
            top: 125 * scale,
            objectFit: 'contain',
          }}
        />
      ))}

      {masteryUrl && (
        <img
          src={masteryUrl}
          width={56 * scale}
          height={56 * scale}
          style={{
            position: 'absolute',
            left: 100 * scale,
            top: 100 * scale,
            objectFit: 'contain',
          }}
        />
      )}

      {supplyBadge && (
        <div
          style={{
            display: 'flex',
            position: 'absolute',
            right: 6 * scale,
            top: (label ? 28 : 7) * scale,
            maxWidth: 94 * scale,
            minHeight: 18 * scale,
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: supplyBadge.backgroundColor,
            color: supplyBadge.color,
            border: `1px solid ${supplyBadge.borderColor}`,
            borderRadius: theme.borderRadius.round,
            padding: `${2 * scale}px ${7 * scale}px`,
            fontSize: supplyBadge.fontSize * scale,
            lineHeight: 1,
            fontWeight: 900,
            letterSpacing: supplyBadge.letterSpacing ?? '0px',
            overflow: 'hidden',
          }}
        >
          {supplyBadge.label}
        </div>
      )}

      {label && (
        <div
          style={{
            display: 'flex',
            position: 'absolute',
            right: 7 * scale,
            top: 7 * scale,
            backgroundColor: isTrained ? theme.colors.accent : theme.colors.surface,
            color: isTrained ? '#ffffff' : theme.colors.textSecondary,
            border: `1px solid ${isTrained ? theme.colors.accent : theme.colors.border}`,
            borderRadius: theme.borderRadius.round,
            padding: `${2 * scale}px ${8 * scale}px`,
            fontSize: 11 * scale,
            fontWeight: 900,
          }}
        >
          {label}
        </div>
      )}

    </div>
  )
}

function preserveAspectRatioToObjectFit(value?: string): 'fill' | 'contain' | 'cover' {
  if (!value || value === 'none') return 'fill'
  return value.includes('slice') ? 'cover' : 'contain'
}

interface SupplyBadgeStyle {
  label: string
  color: string
  backgroundColor: string
  borderColor: string
  fontSize: number
  letterSpacing?: string
}

const SUPPLY_BADGES: Record<string, SupplyBadgeStyle> = {
  常驻: {
    label: '常驻',
    color: theme.colors.textSecondary,
    backgroundColor: 'rgba(255, 255, 255, 0.92)',
    borderColor: theme.colors.borderStrong,
    fontSize: 10,
  },
  生日: {
    label: '生日',
    color: '#ffffff',
    backgroundColor: '#ff7eb6',
    borderColor: '#ffb3d2',
    fontSize: 10,
  },
  期间限定: {
    label: '期间限定',
    color: '#ffffff',
    backgroundColor: '#ff8f3f',
    borderColor: '#ffc08a',
    fontSize: 9.5,
  },
  CFES限定: {
    label: 'CFES限定',
    color: '#ffffff',
    backgroundColor: '#7c5cff',
    borderColor: '#b7a7ff',
    fontSize: 9.5,
    letterSpacing: '0.1px',
  },
  BFES限定: {
    label: 'BFES限定',
    color: '#ffffff',
    backgroundColor: '#00a6d6',
    borderColor: '#7bdcf4',
    fontSize: 9.5,
    letterSpacing: '0.1px',
  },
  WorldLink限定: {
    label: 'WorldLink限定',
    color: '#ffffff',
    backgroundColor: '#2f6df6',
    borderColor: '#9fbdff',
    fontSize: 8.5,
    letterSpacing: '-0.15px',
  },
  联动限定: {
    label: '联动限定',
    color: '#ffffff',
    backgroundColor: '#e84b6b',
    borderColor: '#ff9caf',
    fontSize: 9.5,
  },
}

const SUPPLY_TYPE_ALIASES: Record<string, string> = {
  normal: '常驻',
  birthday: '生日',
  term_limited: '期间限定',
  colorful_festival_limited: 'CFES限定',
  bloom_festival_limited: 'BFES限定',
  unit_event_limited: 'WorldLink限定',
  collaboration_limited: '联动限定',
  CFes限定: 'CFES限定',
  BFes限定: 'BFES限定',
  WL限定: 'WorldLink限定',
}

function getSupplyBadgeStyle(supplyType: string): SupplyBadgeStyle {
  const normalized = SUPPLY_TYPE_ALIASES[supplyType] ?? supplyType
  return SUPPLY_BADGES[normalized] ?? {
    label: normalized,
    color: theme.colors.text,
    backgroundColor: 'rgba(255, 255, 255, 0.92)',
    borderColor: theme.colors.borderStrong,
    fontSize: normalized.length > 7 ? 8.5 : 9.5,
  }
}

function placeholderCardImage(label: string, color: string): string {
  const safeLabel = escapeXml(label.slice(0, 8))
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="320" height="320" viewBox="0 0 320 320">
  <defs>
    <linearGradient id="g" x1="0" x2="1" y1="0" y2="1">
      <stop offset="0" stop-color="#ffffff"/>
      <stop offset="1" stop-color="${color}" stop-opacity="0.35"/>
    </linearGradient>
  </defs>
  <rect width="320" height="320" rx="36" fill="url(#g)"/>
  <circle cx="248" cy="64" r="60" fill="${color}" opacity="0.22"/>
  <circle cx="74" cy="254" r="78" fill="#ffffff" opacity="0.55"/>
  <text x="50%" y="51%" dominant-baseline="middle" text-anchor="middle" font-family="Arial, sans-serif" font-size="38" font-weight="800" fill="${color}">${safeLabel}</text>
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

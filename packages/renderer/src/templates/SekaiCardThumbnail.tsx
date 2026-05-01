import { getSekaiCardUiAssetDataUri } from '../styles/assets'
import { theme } from '../styles/theme'

export interface SekaiCardThumbnailProps {
  imageUrl?: string
  rarity: string
  attr: string
  label?: string
  size?: number
  isTrained?: boolean
  mastery?: number
  characterName?: string
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
  rarity,
  attr,
  label,
  size = 156,
  isTrained = false,
  mastery = 0,
  characterName,
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

  return (
    <div
      style={{
        display: 'flex',
        position: 'relative',
        width: size,
        height: size,
        borderRadius: 14 * scale,
        overflow: 'hidden',
        backgroundColor: theme.colors.surface,
      }}
    >
      <img
        src={imageUrl ?? placeholderCardImage(characterName ?? 'CARD', attrColor)}
        width={size}
        height={size}
        style={{ objectFit: 'cover' }}
      />

      {frameUrl && (
        <img
          src={frameUrl}
          width={size}
          height={size}
          style={{ position: 'absolute', left: 0, top: 0, objectFit: 'fill' }}
        />
      )}

      {attrUrl && (
        <img
          src={attrUrl}
          width={35 * scale}
          height={35 * scale}
          style={{ position: 'absolute', left: 0, top: 0, objectFit: 'contain' }}
        />
      )}

      {starUrl && Array.from({ length: starCount }).map((_, index) => (
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
    </div>
  )
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

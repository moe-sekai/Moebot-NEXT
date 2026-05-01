import { BaseCard } from './base'
import { theme } from '../styles/theme'

interface CardDetailProps {
  card: {
    id: number
    prefix: string
    characterName: string
    rarity: string
    attr: string
    thumbnailUrl: string
    power?: number
  }
}

export function CardDetail({ card }: CardDetailProps) {
  const attrColor = (theme.colors as any)[card.attr] ?? theme.colors.accent

  return (
    <BaseCard title={card.prefix} subtitle={`${card.characterName} · ID: ${card.id}`} accentColor={attrColor}>
      <div style={{ display: 'flex', gap: theme.spacing.lg, alignItems: 'flex-start' }}>
        {/* Thumbnail */}
        <div
          style={{
            display: 'flex',
            width: 200,
            height: 200,
            borderRadius: theme.borderRadius.lg,
            overflow: 'hidden',
            flexShrink: 0,
            backgroundColor: theme.colors.surface,
          }}
        >
          <img src={card.thumbnailUrl} width={200} height={200} style={{ objectFit: 'cover' }} />
        </div>

        {/* Info */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm, flex: 1 }}>
          <InfoRow label="稀有度" value={card.rarity} />
          <InfoRow label="属性" value={card.attr} color={attrColor} />
          {card.power && <InfoRow label="综合力" value={String(card.power)} />}
        </div>
      </div>
    </BaseCard>
  )
}

function InfoRow({ label, value, color }: { label: string; value: string; color?: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <span style={{ color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>{label}</span>
      <span style={{ color: color ?? theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 600 }}>{value}</span>
    </div>
  )
}

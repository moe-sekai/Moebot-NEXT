import { BaseCard } from './base'
import { theme } from '../styles/theme'

// TODO: Implement full gacha result template
// Should display pull results with card thumbnails, rarity stars, and summary

interface GachaResultProps {
  results: Array<{
    cardId: number
    characterName: string
    rarity: string
    attr: string
    isNew: boolean
    thumbnailUrl: string
  }>
  pullType: string // 'single' | 'multi'
}

export function GachaResult({ results, pullType }: GachaResultProps) {
  return (
    <BaseCard
      title="抽卡结果"
      subtitle={pullType === 'multi' ? '十连抽卡' : '单抽'}
      accentColor={theme.colors.warning}
    >
      <div
        style={{
          display: 'flex',
          flexWrap: 'wrap',
          gap: theme.spacing.sm,
          justifyContent: 'center',
        }}
      >
        {results.map((card, i) => {
          const attrColor = (theme.colors as any)[card.attr] ?? theme.colors.accent
          return (
            <div
              key={i}
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                width: 130,
                gap: 4,
              }}
            >
              {/* Card thumbnail */}
              <div
                style={{
                  display: 'flex',
                  width: 120,
                  height: 120,
                  borderRadius: theme.borderRadius.md,
                  overflow: 'hidden',
                  border: `2px solid ${attrColor}`,
                  backgroundColor: theme.colors.surface,
                  position: 'relative',
                }}
              >
                <img src={card.thumbnailUrl} width={120} height={120} style={{ objectFit: 'cover' }} />
                {card.isNew && (
                  <div
                    style={{
                      display: 'flex',
                      position: 'absolute',
                      top: 4,
                      right: 4,
                      backgroundColor: theme.colors.error,
                      color: '#fff',
                      fontSize: 10,
                      padding: '2px 6px',
                      borderRadius: theme.borderRadius.sm,
                      fontWeight: 700,
                    }}
                  >
                    NEW
                  </div>
                )}
              </div>

              {/* Card info */}
              <span style={{ color: theme.colors.text, fontSize: theme.fontSize.xs, textAlign: 'center' }}>
                {card.characterName}
              </span>
              <span style={{ color: attrColor, fontSize: theme.fontSize.xs, fontWeight: 600 }}>
                {card.rarity}
              </span>
            </div>
          )
        })}
      </div>
    </BaseCard>
  )
}

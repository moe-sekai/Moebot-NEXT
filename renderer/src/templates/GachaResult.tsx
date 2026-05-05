import { getCardThumbnailUrl, type AssetSourceType } from '../../shared'
import { BaseCard } from './base'
import { theme } from '../styles/theme'
import { SekaiCardThumbnail, getAttributeColor } from './SekaiCardThumbnail'

interface GachaResultProps {
  results: Array<{
    cardId: number
    characterName: string
    rarity: string
    attr: string
    isNew: boolean
    thumbnailUrl?: string
    trainedThumbnailUrl?: string
    assetbundleName?: string
    assetSource?: AssetSourceType | string
    isTrained?: boolean
    compositeLayers?: import('../card-thumbnail-composites').CardThumbnailCompositeLayer[]
  }>
  pullType: string // 'single' | 'multi'
  assetSource?: AssetSourceType | string
}

export function GachaResult({ results, pullType, assetSource = 'main-jp' }: GachaResultProps) {
  const highest = results.filter(card => card.rarity === 'rarity_4' || card.rarity === 'rarity_birthday').length

  return (
    <BaseCard
      title="抽卡结果"
      subtitle={pullType === 'multi' ? `十连抽卡 · ★4/${highest}` : '单抽'}
      accentColor={theme.colors.warning}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.md }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            backgroundColor: theme.colors.surface,
            border: `1px solid ${theme.colors.border}`,
            borderRadius: theme.borderRadius.lg,
            padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
          }}
        >
          <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>
            使用 Snowy Viewer 卡面层级：缩略图 + 属性 + 稀有度 + NEW
          </span>
          <span style={{ display: 'flex', color: theme.colors.warning, fontSize: theme.fontSize.md, fontWeight: 900 }}>
            {results.length} pulls
          </span>
        </div>

        <div
          style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: theme.spacing.sm,
            justifyContent: 'center',
          }}
        >
          {results.map((card, i) => {
            const attrColor = getAttributeColor(card.attr)
            const isTrained = card.isTrained ?? (card.rarity === 'rarity_3' || card.rarity === 'rarity_4')
            const source = card.assetSource ?? assetSource
            const thumbnailUrl = card.thumbnailUrl
              ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, source, 'png') : undefined)
            const trainedThumbnailUrl = card.trainedThumbnailUrl
              ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, true, source, 'png') : undefined)
            return (
              <div
                key={i}
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  width: 132,
                  gap: 5,
                  backgroundColor: theme.colors.surface,
                  border: `1px solid ${card.isNew ? attrColor : theme.colors.border}`,
                  borderRadius: theme.borderRadius.lg,
                  padding: 6,
                }}
              >
                <div style={{ display: 'flex', position: 'relative' }}>
                  <SekaiCardThumbnail
                    imageUrl={isTrained ? trainedThumbnailUrl ?? thumbnailUrl : thumbnailUrl}
                    compositeLayers={card.compositeLayers}
                    rarity={card.rarity}
                    attr={card.attr}
                    isTrained={isTrained}
                    characterName={card.characterName}
                    size={112}
                  />
                  {card.isNew && (
                    <div
                      style={{
                        display: 'flex',
                        position: 'absolute',
                        right: -5,
                        top: -5,
                        backgroundColor: theme.colors.error,
                        color: '#fff',
                        fontSize: 10,
                        padding: '3px 7px',
                        borderRadius: theme.borderRadius.round,
                        fontWeight: 900,
                      }}
                    >
                      NEW
                    </div>
                  )}
                </div>

                <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.xs, fontWeight: 800, textAlign: 'center' }}>
                  {card.characterName}
                </span>
                <span style={{ display: 'flex', color: attrColor, fontSize: 11, fontWeight: 800 }}>
                  ID:{card.cardId}
                </span>
              </div>
            )
          })}
        </div>
      </div>
    </BaseCard>
  )
}

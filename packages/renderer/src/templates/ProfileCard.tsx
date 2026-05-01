import { getCardThumbnailUrl, getCharacterIconUrl, type AssetSourceType } from '@moebot/shared'
import { BaseCard } from './base'
import { theme } from '../styles/theme'
import { SekaiCardThumbnail } from './SekaiCardThumbnail'

export interface ProfileCardProps {
  profile: {
    name: string
    rank: number
    userId: string | number
    twitterId?: string
    bio?: string
    signature?: string
    totalPower?: number
    characterId?: number
    avatarUrl?: string
    assetSource?: AssetSourceType | string
    stats?: {
      multiLiveCount?: number
      mvpCount?: number
      superStarCount?: number
      fullComboCount?: number
      allPerfectCount?: number
    }
    leaderCard?: ProfileDeckCard
    deckCards?: ProfileDeckCard[]
    honors?: Array<{
      honorId?: number
      name?: string
      level?: number
      assetbundleName?: string
      imageUrl?: string
    }>
  }
}

interface ProfileDeckCard {
  cardId?: number
  id?: number
  characterName?: string
  rarity?: string
  cardRarityType?: string
  attr?: string
  assetbundleName?: string
  thumbnailUrl?: string
  trainedThumbnailUrl?: string
  isTrained?: boolean
  defaultImage?: string
  mastery?: number
  level?: number
}

export function ProfileCard({ profile }: ProfileCardProps) {
  const source = profile.assetSource ?? 'main-jp'
  const leader = profile.leaderCard ?? profile.deckCards?.[0]
  const avatarUrl = profile.avatarUrl
    ?? resolveCardThumbnail(leader, source)
    ?? (profile.characterId ? getCharacterIconUrl(profile.characterId) : undefined)
  const deckCards = profile.deckCards ?? (leader ? [leader] : [])
  const bio = profile.bio ?? profile.signature

  return (
    <BaseCard title={profile.name} subtitle={`UID: ${profile.userId}`} accentColor={theme.colors.accentLight} footer="Player profile · Moebot NEXT">
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.lg }}>
        <div
          style={{
            display: 'flex',
            gap: theme.spacing.lg,
            alignItems: 'stretch',
            backgroundColor: theme.colors.surface,
            border: `1px solid ${theme.colors.border}`,
            borderRadius: theme.borderRadius.xl,
            padding: theme.spacing.lg,
          }}
        >
          <div
            style={{
              display: 'flex',
              width: 128,
              height: 128,
              borderRadius: theme.borderRadius.xl,
              backgroundColor: theme.colors.surfaceLight,
              border: `1px solid ${theme.colors.borderStrong}`,
              overflow: 'hidden',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
            }}
          >
            {avatarUrl ? (
              <img src={avatarUrl} width={128} height={128} style={{ objectFit: 'cover' }} />
            ) : (
              <span style={{ display: 'flex', fontSize: theme.fontSize.title, color: theme.colors.textMuted, fontWeight: 900 }}>
                {profile.name.slice(0, 1)}
              </span>
            )}
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm, flex: 1, justifyContent: 'center' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: theme.spacing.md }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                <span style={{ display: 'flex', fontSize: theme.fontSize.xl, fontWeight: 900, color: theme.colors.text }}>{profile.name}</span>
                {profile.twitterId && <span style={{ display: 'flex', fontSize: theme.fontSize.sm, color: theme.colors.textSecondary }}>@{profile.twitterId}</span>}
              </div>
              <div
                style={{
                  display: 'flex',
                  padding: '7px 14px',
                  borderRadius: theme.borderRadius.round,
                  backgroundColor: theme.colors.accentSoft,
                  color: theme.colors.accent,
                  fontSize: theme.fontSize.md,
                  fontWeight: 900,
                }}
              >
                Rank {profile.rank}
              </div>
            </div>
            {bio && (
              <div
                style={{
                  display: 'flex',
                  color: theme.colors.textSecondary,
                  fontSize: theme.fontSize.sm,
                  lineHeight: 1.55,
                  padding: theme.spacing.sm,
                  borderRadius: theme.borderRadius.md,
                  backgroundColor: theme.colors.surfaceLight,
                }}
              >
                {bio}
              </div>
            )}
          </div>
        </div>

        <div style={{ display: 'flex', gap: theme.spacing.md }}>
          <StatCard label="总综合力" value={profile.totalPower?.toLocaleString() ?? '-'} highlight />
          <StatCard label="协力次数" value={profile.stats?.multiLiveCount?.toLocaleString() ?? '-'} />
          <StatCard label="MVP" value={profile.stats?.mvpCount?.toLocaleString() ?? '-'} />
          <StatCard label="Super Star" value={profile.stats?.superStarCount?.toLocaleString() ?? '-'} />
        </div>

        {deckCards.length > 0 && (
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
            <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>当前队伍</span>
            <div style={{ display: 'flex', gap: theme.spacing.sm, justifyContent: 'space-between' }}>
              {deckCards.slice(0, 5).map((card, index) => (
                <DeckCard key={`${card.cardId ?? card.id ?? index}`} card={card} source={source} leader={index === 0} />
              ))}
            </div>
          </div>
        )}

        {profile.honors && profile.honors.length > 0 && (
          <div style={{ display: 'flex', gap: theme.spacing.sm, flexWrap: 'wrap' }}>
            {profile.honors.slice(0, 3).map((honor, index) => (
              <div
                key={`${honor.honorId ?? index}`}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: theme.spacing.sm,
                  padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
                  borderRadius: theme.borderRadius.round,
                  backgroundColor: theme.colors.surface,
                  border: `1px solid ${theme.colors.border}`,
                }}
              >
                {honor.imageUrl && <img src={honor.imageUrl} width={36} height={36} style={{ objectFit: 'contain' }} />}
                <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.sm, fontWeight: 800 }}>
                  {honor.name ?? `称号 #${honor.honorId ?? '-'}`}{honor.level ? ` Lv.${honor.level}` : ''}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </BaseCard>
  )
}

function DeckCard({ card, source, leader }: { card: ProfileDeckCard; source: AssetSourceType | string; leader?: boolean }) {
  const rarity = card.cardRarityType ?? card.rarity ?? 'rarity_1'
  const attr = card.attr ?? 'cute'
  const isTrained = card.isTrained ?? card.defaultImage === 'special_training'
  const thumbnailUrl = card.thumbnailUrl
    ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, source, 'png') : undefined)
  const trainedThumbnailUrl = card.trainedThumbnailUrl
    ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, true, source, 'png') : undefined)

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 5, width: 128 }}>
      <div style={{ display: 'flex', position: 'relative' }}>
        <SekaiCardThumbnail
          imageUrl={isTrained ? trainedThumbnailUrl ?? thumbnailUrl : thumbnailUrl}
          rarity={rarity}
          attr={attr}
          isTrained={isTrained}
          mastery={card.mastery}
          characterName={card.characterName}
          size={112}
        />
        {leader && (
          <div
            style={{
              display: 'flex',
              position: 'absolute',
              right: -6,
              top: -6,
              padding: '3px 7px',
              borderRadius: theme.borderRadius.round,
              backgroundColor: theme.colors.accent,
              color: '#fff',
              fontSize: 10,
              fontWeight: 900,
            }}
          >
            LEADER
          </div>
        )}
      </div>
      <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 800 }}>
        {card.level ? `Lv.${card.level}` : `#${card.cardId ?? card.id ?? '-'}`}
      </span>
    </div>
  )
}

function StatCard({ label, value, highlight }: { label: string; value: string; highlight?: boolean }) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        gap: 5,
        backgroundColor: highlight ? theme.colors.surfaceAccent : theme.colors.surface,
        border: `1px solid ${highlight ? theme.colors.borderStrong : theme.colors.border}`,
        borderRadius: theme.borderRadius.lg,
        padding: theme.spacing.md,
      }}
    >
      <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.xs, fontWeight: 800 }}>{label}</span>
      <span style={{ display: 'flex', color: highlight ? theme.colors.accent : theme.colors.text, fontSize: theme.fontSize.lg, fontWeight: 900 }}>{value}</span>
    </div>
  )
}

function resolveCardThumbnail(card: ProfileDeckCard | undefined, source: AssetSourceType | string): string | undefined {
  if (!card) return undefined
  if (card.thumbnailUrl) return card.thumbnailUrl
  return card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, source, 'png') : undefined
}

import { getCardThumbnailUrl, getCharacterIconUrl, type AssetSourceType } from '../../shared'
import { BaseCard } from './base'
import { theme } from '../styles/theme'
import { SekaiCardThumbnail } from './SekaiCardThumbnail'
import { ScoreDeltaText, ScoreText, scoreTextStyle } from './ScoreText'

export interface RankingListProps {
  title: string
  subtitle?: string
  rankings: Array<{
    rank: number
    name?: string
    displayName?: string
    signature?: string
    score: number
    userId?: string | number
    scoreDelta?: number
    rankDelta?: number
    avatarUrl?: string
    churn48h?: number
    recentActivityCount?: number
    growth1h?: number
    speed20m3?: number
    churn1h?: number
    churn20m3?: number
    trend?: string
    isTierLine?: boolean
    leaderCharacterId?: number
    leaderCard?: {
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
    }
  }>
  eventId?: number
  eventName?: string
  updatedAt?: number | string
  assetSource?: AssetSourceType | string
  region?: string
  regionLabel?: string
  boardType?: string
  targetId?: number
}

const TOP_COLORS: Record<number, { bg: string; text: string; border: string }> = {
  1: { bg: '#fff6cf', text: '#9a6400', border: '#f4c542' },
  2: { bg: '#f4f7fb', text: '#667085', border: '#c9d3df' },
  3: { bg: '#fff0df', text: '#a7561b', border: '#e8a56b' },
}

export function RankingList({ title, subtitle, rankings, eventId, eventName, updatedAt, assetSource = 'main-jp', regionLabel, boardType, targetId }: RankingListProps) {
  const showPodium = rankings.length >= 3 && rankings.slice(0, 3).every((entry) => entry.rank >= 1 && entry.rank <= 3)
  const topThree = showPodium ? rankings.slice(0, 3) : []
  const autoSubtitle = subtitle ?? [eventName ?? '活动实时排行', eventId ? `Event #${eventId}` : undefined, boardType === 'worldlink' && targetId ? `WL 角色 ${targetId}` : undefined, regionLabel, updatedAt ? `更新 ${fmtTime(updatedAt)}` : undefined].filter(Boolean).join(' · ')
  const rest = showPodium ? rankings.slice(3) : rankings

  return (
    <BaseCard
      title={title}
      subtitle={autoSubtitle}
      accentColor={theme.colors.accent}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.md }}>
        {topThree.length > 0 && (
          <div style={{ display: 'flex', gap: theme.spacing.sm, alignItems: 'stretch' }}>
            {topThree.map((entry) => (
              <TopRankingCard key={entry.rank} entry={entry} assetSource={assetSource} />
            ))}
          </div>
        )}

        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: theme.spacing.xs,
            backgroundColor: theme.colors.surface,
            border: `1px solid ${theme.colors.border}`,
            borderRadius: theme.borderRadius.xl,
            padding: theme.spacing.sm,
          }}
        >
          {(rest.length > 0 ? rest : rankings).map((entry) => (
            <RankingRow key={`${entry.rank}-${entry.userId ?? entry.name}`} entry={entry} assetSource={assetSource} compact={rankings.length > 16} />
          ))}
        </div>
      </div>
    </BaseCard>
  )
}

function TopRankingCard({ entry, assetSource }: { entry: RankingListProps['rankings'][number]; assetSource: AssetSourceType | string }) {
  const top = TOP_COLORS[entry.rank] ?? TOP_COLORS[3]
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        alignItems: 'center',
        gap: theme.spacing.sm,
        padding: theme.spacing.md,
        borderRadius: theme.borderRadius.xl,
        backgroundColor: top.bg,
        border: `1px solid ${top.border}`,
      }}
    >
      <span style={{ display: 'flex', color: top.text, fontSize: theme.fontSize.lg, fontWeight: 900 }}>#{entry.rank}</span>
      <RankingAvatar entry={entry} size={88} assetSource={assetSource} />
      <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.sm, fontWeight: 900, textAlign: 'center', maxWidth: 170 }}>
        {entry.displayName ?? entry.name ?? 'Unknown'}
      </span>
      <ScoreText value={entry.score} color={top.text} fontSize={theme.fontSize.md} />
      <DeltaLine scoreDelta={entry.scoreDelta} rankDelta={entry.rankDelta} compact />
    </div>
  )
}

function RankingRow({ entry, assetSource, compact = false }: { entry: RankingListProps['rankings'][number]; assetSource: AssetSourceType | string; compact?: boolean }) {
  const top = TOP_COLORS[entry.rank]
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: theme.spacing.md,
        padding: `${compact ? 6 : theme.spacing.sm}px ${theme.spacing.md}px`,
        borderRadius: theme.borderRadius.lg,
        backgroundColor: top?.bg ?? theme.colors.surfaceLight,
        border: `1px solid ${top?.border ?? theme.colors.border}`,
      }}
    >
      <div style={{ display: 'flex', width: 58, justifyContent: 'center' }}>
        <span
          style={{
            display: 'flex',
            minWidth: 46,
            justifyContent: 'center',
            padding: '4px 7px',
            borderRadius: theme.borderRadius.md,
            backgroundColor: top?.border ?? theme.colors.surface,
            color: top ? top.text : theme.colors.textSecondary,
            fontSize: theme.fontSize.sm,
            fontWeight: 900,
          }}
        >
          #{entry.rank}
        </span>
      </div>

      <RankingAvatar entry={entry} size={compact ? 46 : 64} assetSource={assetSource} />

      <div style={{ display: 'flex', flexDirection: 'column', gap: 4, minWidth: 0, flex: 1 }}>
        <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>
          {entry.displayName ?? entry.name ?? 'Unknown'}
        </span>
        {entry.signature ? (
          <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.xs, maxWidth: 350 }}>
            {entry.signature}
          </span>
        ) : (
          <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.xs }}>
            UID: {entry.userId ?? '-'}
          </span>
        )}
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 4, width: 168 }}>
        <ScoreText value={entry.score} fontSize={theme.fontSize.lg} />
        <DeltaLine scoreDelta={entry.scoreDelta} rankDelta={entry.rankDelta} />
      </div>
    </div>
  )
}

function RankingAvatar({ entry, size, assetSource }: { entry: RankingListProps['rankings'][number]; size: number; assetSource: AssetSourceType | string }) {
  const card = entry.leaderCard
  if (card) {
    const rarity = card.cardRarityType ?? card.rarity ?? 'rarity_1'
    const attr = card.attr ?? 'cute'
    const isTrained = shouldUseTrainedImage(card)
    const thumbnailUrl = card.thumbnailUrl
      ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, assetSource, 'png') : undefined)
    const trainedThumbnailUrl = card.trainedThumbnailUrl
      ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, true, assetSource, 'png') : undefined)
    return (
      <SekaiCardThumbnail
        imageUrl={isTrained ? trainedThumbnailUrl ?? thumbnailUrl : thumbnailUrl}
        rarity={rarity}
        attr={attr}
        isTrained={isTrained}
        mastery={card.mastery}
        characterName={card.characterName ?? entry.displayName ?? entry.name}
        size={size}
      />
    )
  }

  const avatarUrl = entry.avatarUrl ?? (entry.leaderCharacterId ? getCharacterIconUrl(entry.leaderCharacterId) : undefined)
  return (
    <div
      style={{
        display: 'flex',
        width: size,
        height: size,
        borderRadius: theme.borderRadius.lg,
        overflow: 'hidden',
        backgroundColor: theme.colors.surface,
        border: `1px solid ${theme.colors.border}`,
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      {avatarUrl ? (
        <img src={avatarUrl} width={size} height={size} style={{ objectFit: 'cover' }} />
      ) : (
        <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.sm, fontWeight: 900 }}>#{entry.rank}</span>
      )}
    </div>
  )
}

function fmtTime(value: number | string) {
  const date = new Date(typeof value === 'number' && value < 1_000_000_000_000 ? value * 1000 : value)
  if (Number.isNaN(date.getTime())) return '—'
  return `${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

function shouldUseTrainedImage(card: NonNullable<RankingListProps['rankings'][number]['leaderCard']>): boolean {
  if (card.defaultImage === 'special_training') return true
  if (card.defaultImage === 'original') return false
  return Boolean(card.isTrained)
}

function DeltaLine({ scoreDelta, rankDelta, compact = false }: { scoreDelta?: number; rankDelta?: number; compact?: boolean }) {
  const hasScore = typeof scoreDelta === 'number' && scoreDelta !== 0
  const hasRank = typeof rankDelta === 'number' && rankDelta !== 0
  if (!hasScore && !hasRank) {
    return <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: compact ? 10 : theme.fontSize.xs }}>—</span>
  }
  const color = (scoreDelta ?? 0) >= 0 ? theme.colors.success : theme.colors.error
  return (
    <div style={{ display: 'flex', gap: 5, alignItems: 'center', justifyContent: 'flex-end' }}>
      {hasRank && (
        <span style={scoreTextStyle({ color, fontSize: compact ? 10 : theme.fontSize.xs })}>
          {rankDelta! > 0 ? '▲' : '▼'}{Math.abs(rankDelta!)}
        </span>
      )}
      {hasScore && (
        <ScoreDeltaText value={scoreDelta!} color={color} fontSize={compact ? 10 : theme.fontSize.xs} />
      )}
    </div>
  )
}

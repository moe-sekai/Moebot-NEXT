import { getCardThumbnailUrl, getCharacterIconUrl, type AssetSourceType } from '../../shared'
import { BaseCard } from './base'
import { theme } from '../styles/theme'
import { SekaiCardThumbnail } from './SekaiCardThumbnail'
import type { RankingListProps } from './RankingList'

export function ChurnRankingList({ title, subtitle, rankings, eventId, eventName, assetSource = 'main-jp' }: RankingListProps) {
  const shown = rankings.slice(0, 10)
  return (
    <BaseCard
      title={title}
      subtitle={subtitle ?? `${eventName ?? '实时查房'}${eventId ? ` · Event #${eventId}` : ''}`}
      accentColor={theme.colors.accent}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm }}>
        {shown.map((entry) => (
          <ChurnRow key={`${entry.rank}-${entry.userId ?? entry.name}`} entry={entry} assetSource={assetSource} />
        ))}
        <div style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.xs, justifyContent: 'center', paddingTop: 4 }}>
          TOP100 后为扩展榜线，数据可能存在延迟
        </div>
      </div>
    </BaseCard>
  )
}

function ChurnRow({ entry, assetSource }: { entry: RankingListProps['rankings'][number]; assetSource: AssetSourceType | string }) {
  const isTierLine = Boolean(entry.isTierLine)
  const deltaColor = (entry.scoreDelta ?? 0) >= 0 ? theme.colors.success : theme.colors.error
  return (
    <div
      style={{
        display: 'flex',
        gap: theme.spacing.md,
        alignItems: 'center',
        padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
        borderRadius: theme.borderRadius.xl,
        backgroundColor: isTierLine ? theme.colors.surfaceLight : theme.colors.surface,
        border: `1px solid ${isTierLine ? theme.colors.borderStrong : theme.colors.border}`,
      }}
    >
      <div style={{ display: 'flex', width: 56, justifyContent: 'center' }}>
        <span
          style={{
            display: 'flex',
            minWidth: 48,
            justifyContent: 'center',
            padding: '5px 8px',
            borderRadius: theme.borderRadius.md,
            backgroundColor: rankColor(entry.rank).bg,
            color: rankColor(entry.rank).text,
            fontSize: theme.fontSize.sm,
            fontWeight: 900,
          }}
        >
          #{entry.rank}
        </span>
      </div>

      <RankingAvatar entry={entry} size={58} assetSource={assetSource} />

      <div style={{ display: 'flex', flexDirection: 'column', gap: 6, minWidth: 0, flex: 1 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>
            {entry.displayName ?? entry.name ?? (isTierLine ? `TOP${entry.rank} 榜线` : 'Unknown')}
          </span>
          {isTierLine && <Badge text="榜线" color={theme.colors.accent} />}
        </div>
        <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
          <Badge text={`48H ${entry.churn48h ?? 0}`} color={theme.colors.accent} />
          <Badge text={`1H速 ${fmtSpeed(entry.growth1h ?? 0)}`} color={theme.colors.textSecondary} />
          <Badge text={`20m×3 ${fmtSpeed(entry.speed20m3 ?? 0)} ${trendIcon(entry.trend)}`} color={trendColor(entry.trend)} />
          {!isTierLine && <Badge text={`1H周回 ${entry.churn1h ?? 0}`} color={theme.colors.warning} />}
          {entry.recentActivityCount ? <Badge text={`近期 ${entry.recentActivityCount}`} color={theme.colors.textMuted} /> : null}
        </div>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 5, width: 160 }}>
        <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.lg, fontWeight: 900 }}>
          {entry.score.toLocaleString()}P
        </span>
        {entry.scoreDelta ? (
          <span style={{ display: 'flex', color: deltaColor, fontSize: theme.fontSize.sm, fontWeight: 900 }}>
            {entry.scoreDelta > 0 ? '+' : ''}{entry.scoreDelta.toLocaleString()}
          </span>
        ) : (
          <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.xs }}>—</span>
        )}
      </div>
    </div>
  )
}

function RankingAvatar({ entry, size, assetSource }: { entry: RankingListProps['rankings'][number]; size: number; assetSource: AssetSourceType | string }) {
  const card = entry.leaderCard
  if (card) {
    const rarity = card.cardRarityType ?? card.rarity ?? 'rarity_1'
    const attr = card.attr ?? 'cute'
    const isTrained = card.defaultImage === 'special_training' || Boolean(card.isTrained)
    const thumbnailUrl = card.thumbnailUrl ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, assetSource, 'png') : undefined)
    const trainedThumbnailUrl = card.trainedThumbnailUrl ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, true, assetSource, 'png') : undefined)
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
        backgroundColor: theme.colors.surfaceLight,
        border: `1px solid ${theme.colors.border}`,
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      {avatarUrl ? <img src={avatarUrl} width={size} height={size} style={{ objectFit: 'cover' }} /> : <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.sm, fontWeight: 900 }}>#{entry.rank}</span>}
    </div>
  )
}

function Badge({ text, color }: { text: string; color: string }) {
  return (
    <span
      style={{
        display: 'flex',
        alignItems: 'center',
        padding: '3px 7px',
        borderRadius: theme.borderRadius.md,
        backgroundColor: theme.colors.surfaceLight,
        color,
        fontSize: theme.fontSize.xs,
        fontWeight: 800,
      }}
    >
      {text}
    </span>
  )
}

function rankColor(rank: number) {
  if (rank === 1) return { bg: '#fff0b8', text: '#9a6400' }
  if (rank === 2) return { bg: '#eef2f7', text: '#667085' }
  if (rank === 3) return { bg: '#ffe1bd', text: '#a7561b' }
  return { bg: theme.colors.accentSoft, text: theme.colors.accent }
}

function fmtSpeed(value: number) {
  return `${Math.round(value / 1000)}k`
}

function trendIcon(trend?: string) {
  if (trend === 'up') return '▲'
  if (trend === 'down') return '▼'
  return '—'
}

function trendColor(trend?: string) {
  if (trend === 'up') return theme.colors.success
  if (trend === 'down') return theme.colors.error
  return theme.colors.textSecondary
}

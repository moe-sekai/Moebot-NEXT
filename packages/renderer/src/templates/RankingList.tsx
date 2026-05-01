import { BaseCard } from './base'
import { theme } from '../styles/theme'

// TODO: Implement full ranking list template
// Should display a leaderboard with rank, player name, score, and optional tier icons

interface RankingListProps {
  title: string
  rankings: Array<{
    rank: number
    name: string
    score: number
    userId?: string
  }>
  eventId?: number
}

export function RankingList({ title, rankings }: RankingListProps) {
  return (
    <BaseCard title={title} subtitle="排行榜">
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.xs }}>
        {rankings.map((entry) => (
          <div
            key={entry.rank}
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              backgroundColor: theme.colors.surface,
              padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
              borderRadius: theme.borderRadius.md,
            }}
          >
            <div style={{ display: 'flex', gap: theme.spacing.md, alignItems: 'center' }}>
              <span
                style={{
                  color: entry.rank <= 3 ? theme.colors.warning : theme.colors.textSecondary,
                  fontSize: theme.fontSize.lg,
                  fontWeight: 700,
                  width: 40,
                }}
              >
                #{entry.rank}
              </span>
              <span style={{ color: theme.colors.text, fontSize: theme.fontSize.md }}>
                {entry.name}
              </span>
            </div>
            <span style={{ color: theme.colors.accentLight, fontSize: theme.fontSize.md, fontWeight: 600 }}>
              {entry.score.toLocaleString()}
            </span>
          </div>
        ))}
      </div>
    </BaseCard>
  )
}

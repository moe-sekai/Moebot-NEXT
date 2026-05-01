import { BaseCard } from './base'
import { theme } from '../styles/theme'

// TODO: Implement full profile card template
// Should display user avatar, name, rank, play stats, and equipped title/deck

interface ProfileCardProps {
  profile: {
    name: string
    rank: number
    userId: string
    twitterId?: string
    bio?: string
    totalPower?: number
    characterId?: number
  }
}

export function ProfileCard({ profile }: ProfileCardProps) {
  return (
    <BaseCard title={profile.name} subtitle={`UID: ${profile.userId}`} accentColor={theme.colors.accentLight}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.md }}>
        {/* Avatar placeholder + basic info */}
        <div style={{ display: 'flex', gap: theme.spacing.lg, alignItems: 'center' }}>
          {/* Avatar placeholder */}
          <div
            style={{
              display: 'flex',
              width: 100,
              height: 100,
              borderRadius: theme.borderRadius.round,
              backgroundColor: theme.colors.surface,
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: theme.fontSize.xxl,
              color: theme.colors.textMuted,
            }}
          >
            ?
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.xs }}>
            <div style={{ display: 'flex', fontSize: theme.fontSize.lg, fontWeight: 700, color: theme.colors.text }}>
              {profile.name}
            </div>
            <div style={{ display: 'flex', fontSize: theme.fontSize.sm, color: theme.colors.textSecondary }}>
              Rank {profile.rank}
            </div>
            {profile.bio && (
              <div style={{ display: 'flex', fontSize: theme.fontSize.sm, color: theme.colors.textMuted, marginTop: 4 }}>
                {profile.bio}
              </div>
            )}
          </div>
        </div>

        {/* Stats */}
        {profile.totalPower && (
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              backgroundColor: theme.colors.surface,
              padding: theme.spacing.md,
              borderRadius: theme.borderRadius.md,
            }}
          >
            <span style={{ color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>总综合力</span>
            <span style={{ color: theme.colors.accent, fontSize: theme.fontSize.md, fontWeight: 700 }}>
              {profile.totalPower.toLocaleString()}
            </span>
          </div>
        )}
      </div>
    </BaseCard>
  )
}

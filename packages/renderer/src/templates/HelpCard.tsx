import { BaseCard } from './base'
import { theme } from '../styles/theme'

interface HelpCardProps {
  commands: Array<{
    name: string
    description: string
    usage: string
  }>
  version: string
}

export function HelpCard({ commands, version }: HelpCardProps) {
  return (
    <BaseCard title="Moebot NEXT 帮助" subtitle={`v${version}`} accentColor={theme.colors.accent}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm }}>
        {commands.map((cmd) => (
          <div
            key={cmd.name}
            style={{
              display: 'flex',
              flexDirection: 'column',
              backgroundColor: theme.colors.surface,
              padding: theme.spacing.md,
              borderRadius: theme.borderRadius.md,
              gap: 4,
            }}
          >
            <div style={{ display: 'flex', gap: theme.spacing.sm, alignItems: 'center' }}>
              <span style={{ color: theme.colors.accent, fontSize: theme.fontSize.md, fontWeight: 700 }}>
                /{cmd.name}
              </span>
              <span style={{ color: theme.colors.textMuted, fontSize: theme.fontSize.xs }}>
                {cmd.usage}
              </span>
            </div>
            <span style={{ color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>
              {cmd.description}
            </span>
          </div>
        ))}
      </div>
    </BaseCard>
  )
}

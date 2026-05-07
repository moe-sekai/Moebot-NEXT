import { BaseCard } from './base'
import { theme } from '../styles/theme'

interface HelpCommand {
  name: string
  description?: string
  usage?: string
}

interface HelpGroup {
  label: string
  commands: HelpCommand[]
}

interface HelpCardProps {
  commands?: HelpCommand[]
  groups?: HelpGroup[]
  footer?: string
  version: string
}

export function HelpCard({ commands, groups, footer, version }: HelpCardProps) {
  const hasGroups = Array.isArray(groups) && groups.length > 0
  return (
    <BaseCard title="Moebot NEXT 帮助" subtitle={`v${version}`} accentColor={theme.colors.accent}>
      {hasGroups ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.md }}>
          {groups!.map((group) => (
            <div
              key={group.label}
              style={{
                display: 'flex',
                flexDirection: 'column',
                backgroundColor: theme.colors.surface,
                padding: theme.spacing.md,
                borderRadius: theme.borderRadius.md,
                gap: theme.spacing.sm,
              }}
            >
              <span
                style={{
                  color: theme.colors.accent,
                  fontSize: theme.fontSize.md,
                  fontWeight: 700,
                }}
              >
                {group.label}
              </span>
              <div
                style={{
                  display: 'flex',
                  flexWrap: 'wrap',
                  gap: theme.spacing.xs,
                }}
              >
                {group.commands.map((cmd) => (
                  <span
                    key={cmd.name}
                    style={{
                      color: theme.colors.text,
                      fontSize: theme.fontSize.sm,
                      backgroundColor: theme.colors.surfaceLight,
                      borderRadius: theme.borderRadius.sm,
                      padding: '4px 10px',
                    }}
                  >
                    /{cmd.name}
                  </span>
                ))}
              </div>
            </div>
          ))}
          {footer ? (
            <span
              style={{
                color: theme.colors.textMuted,
                fontSize: theme.fontSize.xs,
                whiteSpace: 'pre-wrap',
              }}
            >
              {footer}
            </span>
          ) : null}
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm }}>
          {(commands ?? []).map((cmd) => (
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
                {cmd.usage ? (
                  <span style={{ color: theme.colors.textMuted, fontSize: theme.fontSize.xs }}>
                    {cmd.usage}
                  </span>
                ) : null}
              </div>
              {cmd.description ? (
                <span style={{ color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>
                  {cmd.description}
                </span>
              ) : null}
            </div>
          ))}
        </div>
      )}
    </BaseCard>
  )
}

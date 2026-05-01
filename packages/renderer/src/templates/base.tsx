import type { ReactNode } from 'react'
import { theme } from '../styles/theme'

interface BaseCardProps {
  title?: string
  subtitle?: string
  children: ReactNode
  width?: number
  accentColor?: string
  footer?: string
}

export function BaseCard({ title, subtitle, children, width, accentColor, footer }: BaseCardProps) {
  const w = width ?? theme.cardWidth
  const accent = accentColor ?? theme.colors.accent

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        width: w,
        backgroundColor: theme.colors.background,
        borderRadius: theme.borderRadius.xl,
        overflow: 'hidden',
        fontFamily: 'Noto Sans CJK SC, Noto Sans SC, sans-serif',
      }}
    >
      {/* Accent top bar */}
      <div style={{ display: 'flex', height: 4, backgroundColor: accent, width: '100%' }} />

      {/* Header */}
      {title && (
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            padding: `${theme.spacing.lg}px ${theme.cardPadding}px ${theme.spacing.md}px`,
          }}
        >
          <div style={{ display: 'flex', fontSize: theme.fontSize.xl, fontWeight: 700, color: theme.colors.text }}>
            {title}
          </div>
          {subtitle && (
            <div style={{ display: 'flex', fontSize: theme.fontSize.sm, color: theme.colors.textSecondary, marginTop: 4 }}>
              {subtitle}
            </div>
          )}
        </div>
      )}

      {/* Content */}
      <div style={{ display: 'flex', flexDirection: 'column', padding: `0 ${theme.cardPadding}px ${theme.spacing.lg}px` }}>
        {children}
      </div>

      {/* Footer */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'flex-end',
          padding: `${theme.spacing.sm}px ${theme.cardPadding}px`,
          backgroundColor: theme.colors.surface,
          fontSize: theme.fontSize.xs,
          color: theme.colors.textMuted,
        }}
      >
        {footer ?? 'Moebot NEXT · pjsk.moe'}
      </div>
    </div>
  )
}

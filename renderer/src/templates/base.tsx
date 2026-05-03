import type { ReactNode } from 'react'
import { theme } from '../styles/theme'
import { getMoebotLogoDataUri } from '../styles/logo'

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
  const logo = getMoebotLogoDataUri(accent)

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        width: w,
        backgroundColor: theme.colors.background,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: theme.borderRadius.xl,
        overflow: 'hidden',
        fontFamily: 'Noto Sans CJK SC, Noto Sans SC, sans-serif',
        color: theme.colors.text,
      }}
    >
      {/* Brand header — title (left) + logo (right) on one row */}
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          padding: `${theme.spacing.lg}px ${theme.cardPadding}px ${theme.spacing.md}px`,
          backgroundColor: theme.colors.surface,
          borderTop: `6px solid ${accent}`,
          borderBottom: `1px solid ${theme.colors.border}`,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: theme.spacing.lg }}>
          {/* Left: title + subtitle */}
          <div style={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
            {title && (
              <div
                style={{
                  display: 'flex',
                  fontSize: theme.fontSize.xl,
                  fontWeight: 800,
                  color: theme.colors.text,
                  lineHeight: 1.25,
                }}
              >
                {title}
              </div>
            )}
            {subtitle && (
              <div
                style={{
                  display: 'flex',
                  fontSize: theme.fontSize.sm,
                  color: theme.colors.textSecondary,
                  marginTop: 5,
                  lineHeight: 1.45,
                }}
              >
                {subtitle}
              </div>
            )}
          </div>

          {/* Right: logo — aspect ratio 1536 × 699 ≈ 2.197:1  →  97 × 44 px */}
          <div style={{ display: 'flex', alignItems: 'center', flexShrink: 0 }}>
            <img
              src={logo}
              width={97}
              height={44}
              style={{ objectFit: 'contain', objectPosition: 'right center' }}
            />
          </div>
        </div>
      </div>

      {/* Content */}
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          padding: `${theme.spacing.lg}px ${theme.cardPadding}px`,
          backgroundColor: theme.colors.background,
        }}
      >
        {children}
      </div>

      {/* Footer */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'flex-end',
          alignItems: 'center',
          padding: `${theme.spacing.sm}px ${theme.cardPadding}px`,
          backgroundColor: theme.colors.surface,
          borderTop: `1px solid ${theme.colors.border}`,
          fontSize: theme.fontSize.xs,
          color: theme.colors.textMuted,
        }}
      >
        <span style={{ display: 'flex' }}>{footer ?? 'Moebot NEXT'}</span>
      </div>
    </div>
  )
}

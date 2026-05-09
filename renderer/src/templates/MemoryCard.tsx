import { BaseCard } from './base'
import { theme } from '../styles/theme'

export interface MemorySegment {
  timestamp?: string
  text: string
}

export interface MemoryResult {
  index?: number
  name?: string
  userId?: number | string
  timestamp?: string
  text: string
}

export interface MemoryCardProps {
  title?: string
  subtitle?: string
  profileName?: string
  profileUserId?: number | string
  profile?: string
  segments?: MemorySegment[]
  results?: MemoryResult[]
  emptyText?: string
  footer?: string
  hint?: string
}

const sectionStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  backgroundColor: theme.colors.surface,
  padding: theme.spacing.md,
  borderRadius: theme.borderRadius.md,
  gap: theme.spacing.sm,
}

const labelStyle: React.CSSProperties = {
  display: 'flex',
  color: theme.colors.accent,
  fontSize: theme.fontSize.md,
  fontWeight: 700,
}

const bodyTextStyle: React.CSSProperties = {
  display: 'flex',
  color: theme.colors.text,
  fontSize: theme.fontSize.sm,
  whiteSpace: 'pre-wrap',
  lineHeight: 1.5,
}

const metaTextStyle: React.CSSProperties = {
  display: 'flex',
  color: theme.colors.textMuted,
  fontSize: theme.fontSize.xs,
}

export function MemoryCard(props: MemoryCardProps) {
  const {
    title,
    subtitle,
    profileName,
    profileUserId,
    profile,
    segments,
    results,
    emptyText,
    footer,
    hint,
  } = props

  const hasProfile = Boolean(profile && profile.trim().length > 0)
  const hasSegments = Array.isArray(segments) && segments.length > 0
  const hasResults = Array.isArray(results) && results.length > 0
  const showEmpty = !hasProfile && !hasSegments && !hasResults

  return (
    <BaseCard
      title={title ?? '记忆查询'}
      subtitle={subtitle}
      accentColor={theme.colors.accent}
      footer={footer}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.md }}>
        {(profileName || profileUserId !== undefined) && (
          <div style={sectionStyle}>
            <div style={labelStyle}>用户</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: theme.spacing.sm }}>
              {profileName ? (
                <span
                  style={{
                    display: 'flex',
                    color: theme.colors.text,
                    fontSize: theme.fontSize.lg,
                    fontWeight: 700,
                  }}
                >
                  {profileName}
                </span>
              ) : null}
              {profileUserId !== undefined ? (
                <span style={metaTextStyle}>QQ {String(profileUserId)}</span>
              ) : null}
            </div>
          </div>
        )}

        {hasProfile ? (
          <div style={sectionStyle}>
            <div style={labelStyle}>用户画像</div>
            <div style={bodyTextStyle}>{profile}</div>
          </div>
        ) : null}

        {hasSegments ? (
          <div style={sectionStyle}>
            <div style={labelStyle}>历史记忆片段</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm }}>
              {segments!.map((seg, i) => (
                <div
                  key={i}
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 4,
                    backgroundColor: theme.colors.surfaceLight,
                    padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
                    borderRadius: theme.borderRadius.sm,
                  }}
                >
                  <div style={{ display: 'flex', gap: theme.spacing.sm, alignItems: 'center' }}>
                    <span
                      style={{
                        display: 'flex',
                        color: theme.colors.accent,
                        fontSize: theme.fontSize.sm,
                        fontWeight: 700,
                      }}
                    >
                      {`#${i + 1}`}
                    </span>
                    {seg.timestamp ? <span style={metaTextStyle}>{seg.timestamp}</span> : null}
                  </div>
                  <div style={bodyTextStyle}>{seg.text}</div>
                </div>
              ))}
            </div>
          </div>
        ) : null}

        {hasResults ? (
          <div style={sectionStyle}>
            <div style={labelStyle}>命中结果</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm }}>
              {results!.map((r, i) => (
                <div
                  key={i}
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 4,
                    backgroundColor: theme.colors.surfaceLight,
                    padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
                    borderRadius: theme.borderRadius.sm,
                  }}
                >
                  <div style={{ display: 'flex', gap: theme.spacing.sm, alignItems: 'center', flexWrap: 'wrap' }}>
                    <span
                      style={{
                        display: 'flex',
                        color: theme.colors.accent,
                        fontSize: theme.fontSize.sm,
                        fontWeight: 700,
                      }}
                    >
                      {`#${r.index ?? i + 1}`}
                    </span>
                    {r.name ? (
                      <span
                        style={{
                          display: 'flex',
                          color: theme.colors.text,
                          fontSize: theme.fontSize.sm,
                          fontWeight: 600,
                        }}
                      >
                        {r.name}
                      </span>
                    ) : null}
                    {r.userId !== undefined ? (
                      <span style={metaTextStyle}>{`QQ ${String(r.userId)}`}</span>
                    ) : null}
                    {r.timestamp ? <span style={metaTextStyle}>{r.timestamp}</span> : null}
                  </div>
                  <div style={bodyTextStyle}>{r.text}</div>
                </div>
              ))}
            </div>
          </div>
        ) : null}

        {showEmpty ? (
          <div style={sectionStyle}>
            <div style={bodyTextStyle}>{emptyText ?? '暂无记忆数据。'}</div>
          </div>
        ) : null}

        {hint ? (
          <div
            style={{
              display: 'flex',
              color: theme.colors.textMuted,
              fontSize: theme.fontSize.xs,
              whiteSpace: 'pre-wrap',
            }}
          >
            {hint}
          </div>
        ) : null}
      </div>
    </BaseCard>
  )
}

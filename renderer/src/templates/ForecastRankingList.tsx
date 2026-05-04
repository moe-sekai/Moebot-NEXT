import { BaseCard } from './base'
import { theme } from '../styles/theme'
import { ScoreText } from './ScoreText'

export interface ForecastRankingListProps {
  title: string
  subtitle?: string
  eventId?: number
  eventName?: string
  region?: string
  regionLabel?: string
  status?: string
  updatedAt?: number | string
  items: Array<{
    rank: number
    score: number
    prediction?: number
    hasPrediction?: boolean
    collectTime?: number | string
    isFinal?: boolean
  }>
}

export function ForecastRankingList({ title, subtitle, eventId, eventName, regionLabel, status, updatedAt, items }: ForecastRankingListProps) {
  const shown = items.slice(0, 18)
  const sub = subtitle ?? [regionLabel, eventName, eventId ? `Event #${eventId}` : undefined, statusLabel(status), updatedAt ? `更新 ${fmtTime(updatedAt)}` : undefined].filter(Boolean).join(' · ')
  return (
    <BaseCard title={title} subtitle={sub} accentColor={theme.colors.accent}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.xs }}>
        <HeaderRow />
        {shown.map((item) => (
          <ForecastRow key={item.rank} item={item} />
        ))}
      </div>
    </BaseCard>
  )
}

function HeaderRow() {
  return (
    <div style={{ display: 'flex', padding: '6px 12px', color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 900 }}>
      <span style={{ display: 'flex', width: 80 }}>档位</span>
      <span style={{ display: 'flex', flex: 1 }}>当前分数</span>
      <span style={{ display: 'flex', flex: 1 }}>预测 / 最终</span>
      <span style={{ display: 'flex', width: 120, justifyContent: 'flex-end' }}>采集</span>
    </div>
  )
}

function ForecastRow({ item }: { item: ForecastRankingListProps['items'][number] }) {
  const final = Boolean(item.isFinal)
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        padding: '10px 12px',
        borderRadius: theme.borderRadius.lg,
        border: `1px solid ${theme.colors.border}`,
        backgroundColor: final ? theme.colors.surface : theme.colors.surfaceLight,
        gap: theme.spacing.sm,
      }}
    >
      <span style={{ display: 'flex', width: 80, color: theme.colors.accent, fontSize: theme.fontSize.md, fontWeight: 900 }}>#{item.rank}</span>
      <span style={{ display: 'flex', flex: 1 }}><ScoreText value={item.score} fontSize={theme.fontSize.md} /></span>
      <span style={{ display: 'flex', flex: 1, alignItems: 'center', gap: 5, color: final ? theme.colors.success : theme.colors.warning, fontSize: theme.fontSize.md, fontWeight: 900 }}>
        {final ? (
          <><ScoreText value={item.score} color={theme.colors.success} fontSize={theme.fontSize.md} /><span style={{ display: 'flex' }}>最终</span></>
        ) : item.hasPrediction && item.prediction ? (
          <ScoreText value={item.prediction} color={theme.colors.warning} fontSize={theme.fontSize.md} />
        ) : '—'}
      </span>
      <span style={{ display: 'flex', width: 120, justifyContent: 'flex-end', color: theme.colors.textMuted, fontSize: theme.fontSize.xs }}>
        {item.collectTime ? fmtTime(item.collectTime) : '—'}
      </span>
    </div>
  )
}

function fmtTime(value: number | string) {
  const date = new Date(typeof value === 'number' && value < 1_000_000_000_000 ? value * 1000 : value)
  if (Number.isNaN(date.getTime())) return '—'
  const mm = String(date.getMonth() + 1).padStart(2, '0')
  const dd = String(date.getDate()).padStart(2, '0')
  const hh = String(date.getHours()).padStart(2, '0')
  const min = String(date.getMinutes()).padStart(2, '0')
  return `${mm}-${dd} ${hh}:${min}`
}

function statusLabel(status?: string) {
  if (status === 'active') return '进行中'
  if (status === 'finished') return '已结束'
  return status
}

import type { RankingListProps } from './RankingList'
import { BaseCard } from './base'
import { theme } from '../styles/theme'
import { ScoreText, fmtScore, scoreTextStyle } from './ScoreText'

type ParkingPeriod = { start_time?: number; startTime?: number; end_time?: number; endTime?: number; duration_s?: number; durationS?: number }

export interface WaterTableProps {
  title: string
  subtitle?: string
  entry: RankingListProps['rankings'][number]
  hourlyChurn?: Array<{ hour: string; count: number }>
  parkingPeriods?: ParkingPeriod[]
  eventId?: number
  updatedAt?: number | string
  regionLabel?: string
  boardType?: string
  targetId?: number
}

export function WaterTable({ title, subtitle, entry, hourlyChurn = [], parkingPeriods = [], eventId, updatedAt, regionLabel, boardType, targetId }: WaterTableProps) {
  const recentHours = hourlyChurn.slice(-24)
  const sub = subtitle ?? [regionLabel, eventId ? `Event #${eventId}` : undefined, boardType === 'worldlink' && targetId ? `WL 角色 ${targetId}` : undefined, updatedAt ? `更新 ${fmtTime(updatedAt)}` : undefined].filter(Boolean).join(' · ')
  return (
    <BaseCard title={title} subtitle={sub} accentColor={theme.colors.accent}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.md }}>
        <PlayerSummary entry={entry} />
        <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.xs }}>
          <SectionTitle text="近 24 小时周回" />
          <div style={{ display: 'flex', gap: 4, alignItems: 'flex-end', height: 96, padding: 10, borderRadius: theme.borderRadius.lg, backgroundColor: theme.colors.surface, border: `1px solid ${theme.colors.border}` }}>
            {recentHours.length > 0 ? recentHours.map((item, index) => (
              <HourBar key={`${item.hour}-${index}`} item={item} />
            )) : <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.sm }}>暂无小时周回数据</span>}
          </div>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.xs }}>
          <SectionTitle text="停车区间" />
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            {parkingPeriods.length > 0 ? parkingPeriods.slice(-5).reverse().map((period, index) => (
              <ParkingRow key={index} period={period} />
            )) : <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.sm }}>暂无停车记录</span>}
          </div>
        </div>
      </div>
    </BaseCard>
  )
}

function PlayerSummary({ entry }: { entry: RankingListProps['rankings'][number] }) {
  const name = entry.displayName ?? entry.name ?? 'Unknown'
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm, padding: theme.spacing.md, borderRadius: theme.borderRadius.xl, backgroundColor: theme.colors.surface, border: `1px solid ${theme.colors.border}` }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          <span style={{ display: 'flex', color: theme.colors.accent, fontSize: theme.fontSize.sm, fontWeight: 900 }}>#{entry.rank}</span>
          <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.lg, fontWeight: 900 }}>{name}</span>
          {entry.userId ? <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.xs }}>UID {entry.userId}</span> : null}
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 4 }}>
          <ScoreText value={entry.score} fontSize={theme.fontSize.xl} />
          <div style={{ display: 'flex', alignItems: 'center', gap: 4, color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>
            <span style={{ display: 'flex' }}>48H {entry.churn48h ?? 0} · 1H</span>
            <span style={scoreTextStyle({ color: theme.colors.textSecondary, fontSize: theme.fontSize.sm })}>{fmtScore(entry.growth1h ?? 0)}</span>
          </div>
        </div>
      </div>
    </div>
  )
}

function SectionTitle({ text }: { text: string }) {
  return <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.sm, fontWeight: 900 }}>{text}</span>
}

function HourBar({ item }: { item: { hour: string; count: number } }) {
  const capped = Math.min(33, Math.max(0, item.count ?? 0))
  const height = Math.max(8, Math.round((capped / 33) * 72))
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'flex-end', flex: 1, gap: 4 }}>
      <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: 9 }}>{item.count}</span>
      <div style={{ display: 'flex', width: '100%', height, borderRadius: 5, backgroundColor: theme.colors.accent }} />
    </div>
  )
}

function ParkingRow({ period }: { period: ParkingPeriod }) {
  const start = period.start_time ?? period.startTime
  const end = period.end_time ?? period.endTime
  const duration = period.duration_s ?? period.durationS
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 10px', borderRadius: theme.borderRadius.lg, backgroundColor: theme.colors.surfaceLight, border: `1px solid ${theme.colors.border}` }}>
      <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.xs }}>{start ? fmtTime(start) : '—'} → {end ? fmtTime(end) : '进行中'}</span>
      <span style={{ display: 'flex', color: theme.colors.warning, fontSize: theme.fontSize.xs, fontWeight: 900 }}>{fmtDuration(duration)}</span>
    </div>
  )
}

function fmtTime(value: number | string) {
  const date = new Date(typeof value === 'number' && value < 1_000_000_000_000 ? value * 1000 : value)
  if (Number.isNaN(date.getTime())) return '—'
  return `${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

function fmtDuration(seconds?: number) {
  if (!seconds || seconds <= 0) return '进行中'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return h > 0 ? `${h}h${m}m` : `${m}m`
}

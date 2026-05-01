import { BaseCard } from './base'
import { theme } from '../styles/theme'

interface EventInfoProps {
  event: {
    id: number
    name: string
    eventType: string
    bannerUrl: string
    startAt: number
    closedAt: number
    unit?: string
  }
}

export function EventInfo({ event }: EventInfoProps) {
  const startDate = new Date(event.startAt).toLocaleDateString('zh-CN')
  const endDate = new Date(event.closedAt).toLocaleDateString('zh-CN')

  return (
    <BaseCard title={event.name} subtitle={`${event.eventType} · ID: ${event.id}`}>
      {/* Banner */}
      <div style={{ display: 'flex', width: '100%', borderRadius: theme.borderRadius.md, overflow: 'hidden', backgroundColor: theme.colors.surface }}>
        <img src={event.bannerUrl} width={736} height={184} style={{ objectFit: 'cover', width: '100%' }} />
      </div>

      {/* Info */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm, marginTop: theme.spacing.md }}>
        <InfoRow label="活动类型" value={event.eventType} />
        {event.unit && <InfoRow label="团组" value={event.unit} />}
        <InfoRow label="开始时间" value={startDate} />
        <InfoRow label="结束时间" value={endDate} />
      </div>
    </BaseCard>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
      <span style={{ color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>{label}</span>
      <span style={{ color: theme.colors.text, fontSize: theme.fontSize.md }}>{value}</span>
    </div>
  )
}

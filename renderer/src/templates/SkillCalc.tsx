import { BaseCard } from './base'
import { theme } from '../styles/theme'

export interface SkillCalcProps {
  inputs?: number[]
  chariotHead?: number
  internal?: number
  multiplier?: number
  actualValue?: number
  othersAvg?: number
  title?: string
  subtitle?: string
  usageHint?: string
}

function trim(value: number | undefined): string {
  if (value === undefined || value === null || !Number.isFinite(value)) return '-'
  if (Math.abs(value - Math.round(value)) < 1e-9) return String(Math.round(value))
  return value.toFixed(1).replace(/\.0$/, '')
}

function fmtMultiplier(value: number | undefined): string {
  if (value === undefined || value === null || !Number.isFinite(value)) return '-'
  return value.toFixed(2)
}

export function SkillCalc(props: SkillCalcProps) {
  const inputs = Array.isArray(props.inputs) ? props.inputs : []
  const chariot = inputs[0]
  const others = inputs.slice(1, 5)
  while (others.length < 4) others.push(0)

  const stats: Array<{ label: string; value: string; tone?: 'accent' | 'warning' | 'success' }> = [
    { label: '车头自身', value: trim(props.chariotHead ?? chariot) },
    { label: '内部合计', value: trim(props.internal) },
    { label: '倍率', value: fmtMultiplier(props.multiplier), tone: 'accent' },
    { label: '技能实际值', value: `${trim(props.actualValue)}%`, tone: 'success' },
  ]

  return (
    <BaseCard
      title={props.title ?? '卡组技能效果计算'}
      subtitle={props.subtitle ?? '车头 + 队友4 张 → 倍率 / 技能实际值'}
      accentColor={theme.colors.accent}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.md }}>
        {/* Input chips: chariot + 4 teammates */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            backgroundColor: theme.colors.surface,
            borderRadius: theme.borderRadius.md,
            padding: theme.spacing.md,
            gap: theme.spacing.sm,
            border: `1px solid ${theme.colors.border}`,
          }}
        >
          <div
            style={{
              display: 'flex',
              color: theme.colors.textSecondary,
              fontSize: theme.fontSize.sm,
              fontWeight: 700,
              letterSpacing: 0.5,
            }}
          >
            输入技能值
          </div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: theme.spacing.sm }}>
            <InputChip label="车头" value={trim(chariot)} highlight />
            {others.map((v, i) => (
              <InputChip key={i} label={`队友${i + 1}`} value={trim(v)} />
            ))}
          </div>
        </div>

        {/* Result grid: 2x2 */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: theme.spacing.sm,
          }}
        >
          {stats.map((stat) => (
            <StatCell key={stat.label} label={stat.label} value={stat.value} tone={stat.tone} />
          ))}
        </div>

        {/* Formula hint */}
        <div
          style={{
            display: 'flex',
            backgroundColor: theme.colors.surfaceLight,
            borderRadius: theme.borderRadius.md,
            padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
            color: theme.colors.textMuted,
            fontSize: theme.fontSize.xs,
            border: `1px dashed ${theme.colors.border}`,
            lineHeight: 1.5,
          }}
        >
          {props.usageHint ?? '倍率 = (车头 + 其余技能值平均/5 + 100) / 100；技能实际值 = 车头 + 其余平均/5'}
        </div>
      </div>
    </BaseCard>
  )
}

function InputChip({ label, value, highlight }: { label: string; value: string; highlight?: boolean }) {
  const accent = highlight ? theme.colors.accent : theme.colors.borderStrong
  const bg = highlight ? theme.colors.accentSoft : theme.colors.surfaceLight
  const fg = highlight ? theme.colors.accent : theme.colors.text
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: theme.spacing.xs,
        padding: `6px 12px`,
        borderRadius: theme.borderRadius.round,
        border: `1px solid ${accent}`,
        backgroundColor: bg,
      }}
    >
      <span
        style={{
          display: 'flex',
          color: fg,
          fontSize: theme.fontSize.xs,
          fontWeight: 700,
          letterSpacing: 0.5,
        }}
      >
        {label}
      </span>
      <span
        style={{
          display: 'flex',
          color: theme.colors.text,
          fontSize: theme.fontSize.md,
          fontWeight: 700,
        }}
      >
        {value}
      </span>
    </div>
  )
}

function StatCell({
  label,
  value,
  tone,
}: {
  label: string
  value: string
  tone?: 'accent' | 'warning' | 'success'
}) {
  const toneColor =
    tone === 'accent'
      ? theme.colors.accent
      : tone === 'warning'
      ? theme.colors.warning
      : tone === 'success'
      ? theme.colors.success
      : theme.colors.text
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 4,
        padding: theme.spacing.md,
        backgroundColor: theme.colors.surface,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: theme.borderRadius.md,
      }}
    >
      <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.sm, fontWeight: 600 }}>
        {label}
      </span>
      <span
        style={{
          display: 'flex',
          color: toneColor,
          fontSize: theme.fontSize.xxl,
          fontWeight: 800,
          lineHeight: 1.1,
        }}
      >
        {value}
      </span>
    </div>
  )
}

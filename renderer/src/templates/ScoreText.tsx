import { theme } from '../styles/theme'
import { scoreFontFamily } from '../fonts'

export { scoreFontFamily }

export function scoreTextStyle({
  color = theme.colors.text,
  fontSize = theme.fontSize.md,
  fontWeight = 900,
}: {
  color?: string
  fontSize?: number
  fontWeight?: number
} = {}) {
  return {
    display: 'flex',
    color,
    fontSize,
    fontWeight,
    fontFamily: scoreFontFamily,
    letterSpacing: '-0.01em',
  }
}

export function ScoreText({
  value,
  color = theme.colors.text,
  fontSize = theme.fontSize.md,
  suffix = 'P',
}: {
  value?: number
  color?: string
  fontSize?: number
  suffix?: string
}) {
  return <span style={scoreTextStyle({ color, fontSize })}>{fmtScore(value)}{suffix}</span>
}

export function ScoreDeltaText({
  value,
  color,
  fontSize = theme.fontSize.xs,
}: {
  value: number
  color: string
  fontSize?: number
}) {
  return <span style={scoreTextStyle({ color, fontSize })}>{value > 0 ? '+' : ''}{fmtScore(value)}</span>
}

export function fmtScore(value?: number) {
  return Number(value ?? 0).toLocaleString()
}

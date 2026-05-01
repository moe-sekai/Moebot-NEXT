import { getMusicJacketUrl, type AssetSourceType } from '@moebot/shared'
import { BaseCard } from './base'
import { theme } from '../styles/theme'

const MUSIC_CATEGORY_NAMES: Record<string, string> = {
  mv: '3D MV',
  mv_2d: '2D MV',
  original: '原创MV',
  image: '静态图片',
}

const MUSIC_CATEGORY_COLORS: Record<string, string> = {
  mv: '#4488DD',
  mv_2d: '#44BB88',
  original: '#FF9900',
  image: '#888888',
}

const DIFFICULTY_NAMES: Record<string, string> = {
  easy: 'EASY',
  normal: 'NORMAL',
  hard: 'HARD',
  expert: 'EXPERT',
  master: 'MASTER',
  append: 'APPEND',
}

const DIFFICULTY_COLORS: Record<string, string> = {
  easy: '#5AC06E',
  normal: '#56A4D4',
  hard: '#EFAF28',
  expert: '#E84D53',
  master: '#BB58B8',
  append: '#EE92BC',
}

const DIFFICULTY_ORDER = ['easy', 'normal', 'hard', 'expert', 'master', 'append']

export interface MusicDetailProps {
  music: {
    id: number
    title: string
    pronunciation?: string
    lyricist?: string
    composer?: string
    arranger?: string
    categories?: string[]
    assetbundleName?: string
    jacketUrl?: string
    assetSource?: AssetSourceType | string
    difficulties?: Array<{
      difficulty?: string
      musicDifficulty?: string
      level?: number
      playLevel?: number
      noteCount?: number
      totalNoteCount?: number
    }>
    publishedAt?: number
    releasedAt?: number
    isNewlyWrittenMusic?: boolean
    isFullLength?: boolean
    fillerSec?: number
  }
}

export function MusicDetail({ music }: MusicDetailProps) {
  const source = music.assetSource ?? 'main-jp'
  const jacketUrl = music.jacketUrl
    ?? (music.assetbundleName ? getMusicJacketUrl(music.assetbundleName, source) : undefined)
  const categories = Array.from(new Set(music.categories ?? []))
  const difficulties = normalizeDifficulties(music.difficulties ?? [])
  const publishedAt = music.publishedAt ?? music.releasedAt
  const accent = categories[0] ? MUSIC_CATEGORY_COLORS[categories[0]] ?? theme.colors.accent : theme.colors.accent

  return (
    <BaseCard
      title={music.title}
      subtitle={music.pronunciation ? `${music.pronunciation} · ID: ${music.id}` : `ID: ${music.id}`}
      accentColor={accent}
      footer="Snowy Viewer music assets · Satori renderer"
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.lg }}>
        <div style={{ display: 'flex', gap: theme.spacing.lg, alignItems: 'stretch' }}>
          <div
            style={{
              display: 'flex',
              position: 'relative',
              width: 224,
              height: 224,
              borderRadius: theme.borderRadius.xl,
              overflow: 'hidden',
              flexShrink: 0,
              backgroundColor: theme.colors.surface,
              border: `1px solid ${theme.colors.border}`,
            }}
          >
            <img
              src={jacketUrl ?? placeholderImage('MUSIC', accent, 448, 448)}
              width={224}
              height={224}
              style={{ objectFit: 'cover' }}
            />
            <Badge text={`#${music.id}`} color="rgba(0,0,0,0.62)" style={{ right: 10, top: 10 }} />
            {music.isNewlyWrittenMusic && <Badge text="原创" color={theme.colors.warning} style={{ left: 10, top: 10 }} />}
            {music.isFullLength && <Badge text="FULL" color={theme.colors.accent} style={{ left: 10, bottom: 10 }} />}
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.md, flex: 1 }}>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: theme.spacing.xs }}>
              {categories.length > 0
                ? categories.map((cat) => (
                  <span
                    key={cat}
                    style={{
                      display: 'flex',
                      padding: '5px 10px',
                      borderRadius: theme.borderRadius.round,
                      backgroundColor: MUSIC_CATEGORY_COLORS[cat] ?? theme.colors.textMuted,
                      color: '#fff',
                      fontSize: theme.fontSize.xs,
                      fontWeight: 900,
                    }}
                  >
                    {MUSIC_CATEGORY_NAMES[cat] ?? cat}
                  </span>
                ))
                : <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.sm }}>未标注分类</span>}
            </div>

            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: theme.spacing.sm,
                backgroundColor: theme.colors.surface,
                border: `1px solid ${theme.colors.border}`,
                borderRadius: theme.borderRadius.lg,
                padding: theme.spacing.md,
              }}
            >
              <InfoRow label="作词" value={music.lyricist || '-'} />
              <InfoRow label="作曲" value={music.composer || '-'} />
              <InfoRow label="编曲" value={music.arranger || '-'} />
              {publishedAt && <InfoRow label="发布时间" value={formatDate(publishedAt)} />}
              {typeof music.fillerSec === 'number' && <InfoRow label="Filler" value={`${music.fillerSec}s`} />}
            </div>
          </div>
        </div>

        {difficulties.length > 0 && (
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: theme.spacing.sm,
              backgroundColor: theme.colors.surface,
              border: `1px solid ${theme.colors.border}`,
              borderRadius: theme.borderRadius.xl,
              padding: theme.spacing.md,
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>难度信息</span>
              <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.xs }}>Lv / Notes</span>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.xs }}>
              {difficulties.map((d) => (
                <DifficultyRow key={d.difficulty} difficulty={d.difficulty} level={d.level} noteCount={d.noteCount} />
              ))}
            </div>
          </div>
        )}
      </div>
    </BaseCard>
  )
}

function DifficultyRow({ difficulty, level, noteCount }: { difficulty: string; level: number; noteCount?: number }) {
  const color = DIFFICULTY_COLORS[difficulty] ?? theme.colors.textMuted
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        borderRadius: theme.borderRadius.md,
        backgroundColor: theme.colors.surfaceLight,
        overflow: 'hidden',
        border: `1px solid ${theme.colors.border}`,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: theme.spacing.sm, flex: 1 }}>
        <div style={{ display: 'flex', width: 7, height: 42, backgroundColor: color }} />
        <span style={{ display: 'flex', width: 86, color, fontSize: theme.fontSize.sm, fontWeight: 900 }}>
          {DIFFICULTY_NAMES[difficulty] ?? difficulty.toUpperCase()}
        </span>
        <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.xs }}>
          {difficulty}
        </span>
      </div>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: theme.spacing.sm, paddingRight: theme.spacing.md }}>
        <span style={{ display: 'flex', color: color, fontSize: theme.fontSize.lg, fontWeight: 900 }}>Lv.{level}</span>
        {typeof noteCount === 'number' && (
          <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.xs }}>{noteCount.toLocaleString()} notes</span>
        )}
      </div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', gap: theme.spacing.md, alignItems: 'flex-start' }}>
      <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.sm, flexShrink: 0 }}>{label}</span>
      <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.sm, fontWeight: 800, maxWidth: 360, textAlign: 'right' }}>{value}</span>
    </div>
  )
}

function Badge({ text, color, style }: { text: string; color: string; style: Record<string, number> }) {
  return (
    <div
      style={{
        display: 'flex',
        position: 'absolute',
        padding: '4px 9px',
        borderRadius: theme.borderRadius.round,
        backgroundColor: color,
        color: '#fff',
        fontSize: theme.fontSize.xs,
        fontWeight: 900,
        ...style,
      }}
    >
      {text}
    </div>
  )
}

function normalizeDifficulties(input: NonNullable<MusicDetailProps['music']['difficulties']>) {
  return input
    .map((d) => ({
      difficulty: d.musicDifficulty ?? d.difficulty ?? 'unknown',
      level: d.playLevel ?? d.level ?? 0,
      noteCount: d.totalNoteCount ?? d.noteCount,
    }))
    .filter(d => d.level > 0)
    .sort((a, b) => DIFFICULTY_ORDER.indexOf(a.difficulty) - DIFFICULTY_ORDER.indexOf(b.difficulty))
}

function formatDate(timestamp: number): string {
  const normalized = timestamp < 1_000_000_000_000 ? timestamp * 1000 : timestamp
  return new Date(normalized).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

function placeholderImage(label: string, color: string, width: number, height: number): string {
  const safeLabel = escapeXml(label)
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">
  <defs><linearGradient id="g" x1="0" x2="1" y1="0" y2="1"><stop offset="0" stop-color="#ffffff"/><stop offset="1" stop-color="${color}" stop-opacity="0.36"/></linearGradient></defs>
  <rect width="${width}" height="${height}" rx="42" fill="url(#g)"/>
  <circle cx="${Math.round(width * 0.78)}" cy="${Math.round(height * 0.22)}" r="${Math.round(Math.min(width, height) * 0.18)}" fill="${color}" opacity="0.20"/>
  <circle cx="${Math.round(width * 0.2)}" cy="${Math.round(height * 0.82)}" r="${Math.round(Math.min(width, height) * 0.24)}" fill="#fff" opacity="0.55"/>
  <text x="50%" y="52%" dominant-baseline="middle" text-anchor="middle" font-family="Arial, sans-serif" font-size="${Math.max(34, Math.round(width / 9))}" font-weight="900" fill="${color}">${safeLabel}</text>
</svg>`
  return `data:image/svg+xml;base64,${Buffer.from(svg, 'utf8').toString('base64')}`
}

function escapeXml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

import { BaseCard } from './base'
import { theme } from '../styles/theme'

interface MusicDetailProps {
  music: {
    id: number
    title: string
    pronunciation?: string
    lyricist: string
    composer: string
    arranger: string
    jacketUrl: string
    difficulties?: Array<{
      difficulty: string
      level: number
      noteCount: number
    }>
    publishedAt?: number
  }
}

export function MusicDetail({ music }: MusicDetailProps) {
  return (
    <BaseCard title={music.title} subtitle={music.pronunciation ?? `ID: ${music.id}`}>
      <div style={{ display: 'flex', gap: theme.spacing.lg, alignItems: 'flex-start' }}>
        {/* Jacket */}
        <div
          style={{
            display: 'flex',
            width: 180,
            height: 180,
            borderRadius: theme.borderRadius.lg,
            overflow: 'hidden',
            flexShrink: 0,
            backgroundColor: theme.colors.surface,
          }}
        >
          <img src={music.jacketUrl} width={180} height={180} style={{ objectFit: 'cover' }} />
        </div>

        {/* Info */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm, flex: 1 }}>
          <InfoRow label="作词" value={music.lyricist} />
          <InfoRow label="作曲" value={music.composer} />
          <InfoRow label="编曲" value={music.arranger} />
        </div>
      </div>

      {/* Difficulty table */}
      {music.difficulties && music.difficulties.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', marginTop: theme.spacing.lg, gap: theme.spacing.xs }}>
          <div style={{ display: 'flex', fontSize: theme.fontSize.sm, fontWeight: 700, color: theme.colors.text, marginBottom: 4 }}>
            难度
          </div>
          {music.difficulties.map((d) => (
            <div key={d.difficulty} style={{ display: 'flex', justifyContent: 'space-between', padding: `${theme.spacing.xs}px 0` }}>
              <span style={{ color: theme.colors.textSecondary, fontSize: theme.fontSize.sm, textTransform: 'uppercase' }}>
                {d.difficulty}
              </span>
              <span style={{ color: theme.colors.text, fontSize: theme.fontSize.sm }}>
                Lv.{d.level} ({d.noteCount} notes)
              </span>
            </div>
          ))}
        </div>
      )}
    </BaseCard>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
      <span style={{ color: theme.colors.textSecondary, fontSize: theme.fontSize.sm }}>{label}</span>
      <span style={{ color: theme.colors.text, fontSize: theme.fontSize.sm, maxWidth: 350, textAlign: 'right' }}>{value}</span>
    </div>
  )
}

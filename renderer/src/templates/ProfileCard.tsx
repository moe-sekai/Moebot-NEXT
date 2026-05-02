import { getCardThumbnailUrl, getCharacterIconUrl, type AssetSourceType } from '../../shared'
import { BaseCard } from './base'
import { theme } from '../styles/theme'
import { SekaiCardThumbnail } from './SekaiCardThumbnail'

export interface ProfileCardProps {
  profile: {
    name: string
    rank: number
    userId: string | number
    twitterId?: string
    bio?: string
    signature?: string
    totalPower?: number
    characterId?: number
    avatarUrl?: string
    assetSource?: AssetSourceType | string
    stats?: {
      multiLiveCount?: number
      mvpCount?: number
      superStarCount?: number
      fullComboCount?: number
      allPerfectCount?: number
    }
    musicClearCounts?: MusicClearCount[]
    characterRanks?: CharacterRank[]
    challengeLive?: ChallengeLive
    profileHonors?: ProfileHonor[]
    leaderCard?: ProfileDeckCard
    deckCards?: ProfileDeckCard[]
    honors?: Array<{
      honorId?: number
      name?: string
      level?: number
      assetbundleName?: string
      imageUrl?: string
    }>
  }
}

interface MusicClearCount {
  difficulty: string
  liveClear: number
  fullCombo: number
  allPerfect: number
}

interface CharacterRank {
  characterId: number
  characterName: string
  rank: number
}

interface ChallengeLive {
  characterId: number
  characterName: string
  highScore: number
}

interface ProfileHonor {
  seq?: number
  honorType?: string
  honorId: number
  level: number
  name?: string
  honorRarity?: string
  assetbundleName?: string
}

interface ProfileDeckCard {
  cardId?: number
  id?: number
  characterName?: string
  rarity?: string
  cardRarityType?: string
  attr?: string
  assetbundleName?: string
  thumbnailUrl?: string
  trainedThumbnailUrl?: string
  isTrained?: boolean
  defaultImage?: string
  mastery?: number
  level?: number
}

export function ProfileCard({ profile }: ProfileCardProps) {
  const source = profile.assetSource ?? 'main-jp'
  const leader = profile.leaderCard ?? profile.deckCards?.[0]
  const avatarUrl = profile.avatarUrl
    ?? resolveCardThumbnail(leader, source)
    ?? (profile.characterId ? getCharacterIconUrl(profile.characterId) : undefined)
  const deckCards = profile.deckCards ?? (leader ? [leader] : [])
  const bio = profile.bio ?? profile.signature
  const musicClearCounts = normalizeMusicClearCounts(profile.musicClearCounts ?? [])
  const characterRanks = (profile.characterRanks ?? []).slice(0, 6)
  const profileHonors = profile.profileHonors ?? []

  return (
    <BaseCard title={profile.name} subtitle={`UID: ${profile.userId}`} accentColor={theme.colors.accentLight}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.lg }}>
        <div
          style={{
            display: 'flex',
            gap: theme.spacing.lg,
            alignItems: 'stretch',
            backgroundColor: theme.colors.surface,
            border: `1px solid ${theme.colors.border}`,
            borderRadius: theme.borderRadius.xl,
            padding: theme.spacing.lg,
          }}
        >
          <div
            style={{
              display: 'flex',
              width: 128,
              height: 128,
              borderRadius: theme.borderRadius.xl,
              backgroundColor: theme.colors.surfaceLight,
              border: `1px solid ${theme.colors.borderStrong}`,
              overflow: 'hidden',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
            }}
          >
            {avatarUrl ? (
              <img src={avatarUrl} width={128} height={128} style={{ objectFit: 'cover' }} />
            ) : (
              <span style={{ display: 'flex', fontSize: theme.fontSize.title, color: theme.colors.textMuted, fontWeight: 900 }}>
                {profile.name.slice(0, 1)}
              </span>
            )}
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm, flex: 1, justifyContent: 'center' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: theme.spacing.md }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                <span style={{ display: 'flex', fontSize: theme.fontSize.xl, fontWeight: 900, color: theme.colors.text }}>{profile.name}</span>
                {profile.twitterId && <span style={{ display: 'flex', fontSize: theme.fontSize.sm, color: theme.colors.textSecondary }}>@{profile.twitterId}</span>}
              </div>
              <div
                style={{
                  display: 'flex',
                  padding: '7px 14px',
                  borderRadius: theme.borderRadius.round,
                  backgroundColor: theme.colors.accentSoft,
                  color: theme.colors.accent,
                  fontSize: theme.fontSize.md,
                  fontWeight: 900,
                }}
              >
                Rank {profile.rank}
              </div>
            </div>
            {bio && (
              <div
                style={{
                  display: 'flex',
                  color: theme.colors.textSecondary,
                  fontSize: theme.fontSize.sm,
                  lineHeight: 1.55,
                  padding: theme.spacing.sm,
                  borderRadius: theme.borderRadius.md,
                  backgroundColor: theme.colors.surfaceLight,
                }}
              >
                {bio}
              </div>
            )}
          </div>
        </div>

        <div style={{ display: 'flex', gap: theme.spacing.md }}>
          <StatCard label="总综合力" value={profile.totalPower?.toLocaleString() ?? '-'} highlight />
          <StatCard label="MVP" value={profile.stats?.mvpCount?.toLocaleString() ?? '-'} />
          <StatCard label="Super Star" value={profile.stats?.superStarCount?.toLocaleString() ?? '-'} />
          <StatCard label="挑战 Live" value={profile.challengeLive?.highScore?.toLocaleString() ?? '-'} />
        </div>

        {deckCards.length > 0 && (
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: theme.spacing.md,
              backgroundColor: theme.colors.surface,
              border: `1px solid ${theme.colors.border}`,
              borderRadius: theme.borderRadius.xl,
              padding: theme.spacing.md,
            }}
          >
            <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>当前队伍</span>
            <div style={{ display: 'flex', gap: theme.spacing.sm, justifyContent: 'space-between' }}>
              {deckCards.slice(0, 5).map((card, index) => (
                <DeckCard key={`${card.cardId ?? card.id ?? index}`} card={card} source={source} leader={index === 0} />
              ))}
            </div>
          </div>
        )}

        {(musicClearCounts.length > 0 || characterRanks.length > 0) && (
          <div style={{ display: 'flex', gap: theme.spacing.md, alignItems: 'stretch' }}>
            {musicClearCounts.length > 0 && <MusicClearPanel counts={musicClearCounts} />}
            {characterRanks.length > 0 && <CharacterRankPanel ranks={characterRanks} />}
          </div>
        )}

        {(profile.challengeLive || profileHonors.length > 0) && (
          <div style={{ display: 'flex', gap: theme.spacing.md, alignItems: 'stretch' }}>
            {profile.challengeLive && <ChallengeLivePanel challenge={profile.challengeLive} />}
            {profileHonors.length > 0 && <ProfileHonorPanel honors={profileHonors} />}
          </div>
        )}

        {profile.honors && profile.honors.length > 0 && (
          <div style={{ display: 'flex', gap: theme.spacing.sm, flexWrap: 'wrap' }}>
            {profile.honors.slice(0, 3).map((honor, index) => (
              <div
                key={`${honor.honorId ?? index}`}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: theme.spacing.sm,
                  padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
                  borderRadius: theme.borderRadius.round,
                  backgroundColor: theme.colors.surface,
                  border: `1px solid ${theme.colors.border}`,
                }}
              >
                {honor.imageUrl && <img src={honor.imageUrl} width={36} height={36} style={{ objectFit: 'contain' }} />}
                <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.sm, fontWeight: 800 }}>
                  {honor.name ?? `称号 #${honor.honorId ?? '-'}`}{honor.level ? ` Lv.${honor.level}` : ''}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </BaseCard>
  )
}

function ChallengeLivePanel({ challenge }: { challenge: ChallengeLive }) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        gap: theme.spacing.sm,
        backgroundColor: theme.colors.surface,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: theme.borderRadius.xl,
        padding: theme.spacing.md,
      }}
    >
      <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>挑战 Live</span>
      <div style={{ display: 'flex', justifyContent: 'space-between', gap: theme.spacing.md, alignItems: 'baseline' }}>
        <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.sm, fontWeight: 800 }}>{challenge.characterName}</span>
        <span style={{ display: 'flex', color: theme.colors.accent, fontSize: theme.fontSize.lg, fontWeight: 900 }}>{challenge.highScore.toLocaleString()}</span>
      </div>
    </div>
  )
}

function ProfileHonorPanel({ honors }: { honors: ProfileHonor[] }) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1.6,
        gap: theme.spacing.sm,
        backgroundColor: theme.colors.surface,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: theme.borderRadius.xl,
        padding: theme.spacing.md,
      }}
    >
      <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>展示称号</span>
      <div style={{ display: 'flex', gap: theme.spacing.sm, flexWrap: 'wrap' }}>
        {honors.slice(0, 3).map((honor) => (
          <span
            key={`${honor.seq ?? honor.honorId}`}
            style={{
              display: 'flex',
              padding: '6px 10px',
              borderRadius: theme.borderRadius.round,
              backgroundColor: theme.colors.surfaceLight,
              border: `1px solid ${theme.colors.border}`,
              color: theme.colors.textSecondary,
              fontSize: theme.fontSize.xs,
              fontWeight: 800,
            }}
          >
            {honor.name ?? `称号 #${honor.honorId}`}{honor.level ? ` Lv.${honor.level}` : ''}
          </span>
        ))}
      </div>
    </div>
  )
}

function MusicClearPanel({ counts }: { counts: MusicClearCount[] }) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1.4,
        gap: theme.spacing.sm,
        backgroundColor: theme.colors.surface,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: theme.borderRadius.xl,
        padding: theme.spacing.md,
      }}
    >
      <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>通关统计</span>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        {counts.map((count) => (
          <div key={count.difficulty} style={{ display: 'flex', alignItems: 'center', gap: theme.spacing.sm }}>
            <span style={{ display: 'flex', width: 66, color: difficultyColor(count.difficulty), fontSize: theme.fontSize.xs, fontWeight: 900 }}>
              {difficultyLabel(count.difficulty)}
            </span>
            <span style={{ display: 'flex', flex: 1, color: theme.colors.textSecondary, fontSize: 11, fontWeight: 800 }}>
              Clear {count.liveClear.toLocaleString()}
            </span>
            <span style={{ display: 'flex', flex: 1, color: theme.colors.textSecondary, fontSize: 11, fontWeight: 800 }}>
              FC {count.fullCombo.toLocaleString()}
            </span>
            <span style={{ display: 'flex', flex: 1, color: theme.colors.textSecondary, fontSize: 11, fontWeight: 800 }}>
              AP {count.allPerfect.toLocaleString()}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

function CharacterRankPanel({ ranks }: { ranks: CharacterRank[] }) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        gap: theme.spacing.sm,
        backgroundColor: theme.colors.surface,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: theme.borderRadius.xl,
        padding: theme.spacing.md,
      }}
    >
      <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>角色等级 TOP</span>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        {ranks.map((rank) => (
          <div key={rank.characterId} style={{ display: 'flex', justifyContent: 'space-between', gap: theme.spacing.sm }}>
            <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.xs, fontWeight: 800, maxWidth: 130 }}>{rank.characterName}</span>
            <span style={{ display: 'flex', color: theme.colors.accent, fontSize: theme.fontSize.xs, fontWeight: 900 }}>Lv.{rank.rank}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

function DeckCard({ card, source, leader }: { card: ProfileDeckCard; source: AssetSourceType | string; leader?: boolean }) {
  const rarity = card.cardRarityType ?? card.rarity ?? 'rarity_1'
  const attr = card.attr ?? 'cute'
  const isTrained = shouldUseTrainedImage(card)
  const thumbnailUrl = card.thumbnailUrl
    ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, source, 'png') : undefined)
  const trainedThumbnailUrl = card.trainedThumbnailUrl
    ?? (card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, true, source, 'png') : undefined)

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 5, width: 128 }}>
      <div style={{ display: 'flex', position: 'relative' }}>
        <SekaiCardThumbnail
          imageUrl={isTrained ? trainedThumbnailUrl ?? thumbnailUrl : thumbnailUrl}
          rarity={rarity}
          attr={attr}
          isTrained={isTrained}
          mastery={card.mastery}
          characterName={card.characterName}
          size={112}
        />
        {leader && (
          <div
            style={{
              display: 'flex',
              position: 'absolute',
              right: -6,
              top: -6,
              padding: '3px 7px',
              borderRadius: theme.borderRadius.round,
              backgroundColor: theme.colors.accent,
              color: '#fff',
              fontSize: 10,
              fontWeight: 900,
            }}
          >
            LEADER
          </div>
        )}
      </div>
      <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: theme.fontSize.xs, fontWeight: 800 }}>
        {card.level ? `Lv.${card.level}` : `#${card.cardId ?? card.id ?? '-'}`}
      </span>
    </div>
  )
}

const DIFFICULTY_ORDER = ['easy', 'normal', 'hard', 'expert', 'master', 'append']

function shouldUseTrainedImage(card: ProfileDeckCard): boolean {
  if (card.defaultImage === 'special_training') return true
  if (card.defaultImage === 'original') return false
  return Boolean(card.isTrained)
}

function normalizeMusicClearCounts(counts: MusicClearCount[]): MusicClearCount[] {
  return [...counts].sort((a, b) => DIFFICULTY_ORDER.indexOf(a.difficulty) - DIFFICULTY_ORDER.indexOf(b.difficulty))
}

function difficultyLabel(difficulty: string): string {
  const labels: Record<string, string> = {
    easy: 'EASY',
    normal: 'NORMAL',
    hard: 'HARD',
    expert: 'EXPERT',
    master: 'MASTER',
    append: 'APPEND',
  }
  return labels[difficulty] ?? difficulty.toUpperCase()
}

function difficultyColor(difficulty: string): string {
  const colors: Record<string, string> = {
    easy: '#5AC06E',
    normal: '#56A4D4',
    hard: '#EFAF28',
    expert: '#E84D53',
    master: '#BB58B8',
    append: '#EE92BC',
  }
  return colors[difficulty] ?? theme.colors.textMuted
}

function StatCard({ label, value, highlight }: { label: string; value: string; highlight?: boolean }) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        gap: 5,
        backgroundColor: highlight ? theme.colors.surfaceAccent : theme.colors.surface,
        border: `1px solid ${highlight ? theme.colors.borderStrong : theme.colors.border}`,
        borderRadius: theme.borderRadius.lg,
        padding: theme.spacing.md,
      }}
    >
      <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.xs, fontWeight: 800 }}>{label}</span>
      <span style={{ display: 'flex', color: highlight ? theme.colors.accent : theme.colors.text, fontSize: theme.fontSize.lg, fontWeight: 900 }}>{value}</span>
    </div>
  )
}

function resolveCardThumbnail(card: ProfileDeckCard | undefined, source: AssetSourceType | string): string | undefined {
  if (!card) return undefined
  if (card.thumbnailUrl) return card.thumbnailUrl
  return card.assetbundleName ? getCardThumbnailUrl(card.assetbundleName, false, source, 'png') : undefined
}

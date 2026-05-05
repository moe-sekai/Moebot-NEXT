import { getBondsHonorCharacterUrl, getCardThumbnailUrl, getCharacterIconUrl, type AssetSourceType } from '../../shared'
import { BaseCard } from './base'
import { theme } from '../styles/theme'
import { getLocalFrameAssetDataUri } from '../styles/assets'
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
  imageUrl?: string
  frameUrl?: string
  levelIconUrl?: string
  levelIcon6Url?: string
  bondsHonorViewType?: string
  bondsHonorWordId?: number
  bondsHonorWordAssetbundleName?: string
  bondsHonorWordUrl?: string
  leftCharacterId?: number
  rightCharacterId?: number
  leftCharacterUrl?: string
  rightCharacterUrl?: string
  leftColor?: string
  rightColor?: string
  assetSource?: AssetSourceType | string
}

const DEFAULT_HONOR_FRAME_URLS: Record<string, string | undefined> = {
  low: getLocalFrameAssetDataUri('frame_degree_m_1.png'),
  middle: getLocalFrameAssetDataUri('frame_degree_m_2.png'),
  high: getLocalFrameAssetDataUri('frame_degree_m_3.png'),
  highest: getLocalFrameAssetDataUri('frame_degree_m_4.png'),
}
const DEFAULT_HONOR_LEVEL_ICON_URL = getLocalFrameAssetDataUri('icon_degreeLv.png')
const DEFAULT_HONOR_LEVEL_ICON6_URL = getLocalFrameAssetDataUri('icon_degreeLv6.png')
const HONOR_BADGE_WIDTH = 224
const HONOR_BADGE_HEIGHT = Math.round(HONOR_BADGE_WIDTH * 80 / 380)

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
  compositeLayers?: import('../card-thumbnail-composites').CardThumbnailCompositeLayer[]
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
  const characterRanks = profile.characterRanks ?? []
  const profileHonors = (profile.profileHonors ?? []).map(honor => ({ ...honor, assetSource: source }))

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
          {profile.challengeLive && <ChallengeLiveMiniCard challenge={profile.challengeLive} />}
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

        {profileHonors.length > 0 && <ProfileHonorPanel honors={profileHonors} />}

        {musicClearCounts.length > 0 && <MusicClearPanel counts={musicClearCounts} />}

        {characterRanks.length > 0 && <CharacterRankPanel ranks={characterRanks} />}

        {profileHonors.length === 0 && profile.honors && profile.honors.length > 0 && (
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

function ChallengeLiveMiniCard({ challenge }: { challenge: ChallengeLive }) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        gap: 5,
        backgroundColor: theme.colors.surface,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: theme.borderRadius.lg,
        padding: theme.spacing.md,
      }}
    >
      <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.xs, fontWeight: 800 }}>挑战 Live · {challenge.characterName}</span>
      <span style={{ display: 'flex', color: theme.colors.accent, fontSize: theme.fontSize.lg, fontWeight: 900 }}>{challenge.highScore.toLocaleString()}</span>
    </div>
  )
}

function ProfileHonorPanel({ honors }: { honors: ProfileHonor[] }) {
  const hasRenderableHonor = honors.some(honor => Boolean(honor.imageUrl) || (honor.honorType === 'bonds' && honor.leftCharacterId && honor.rightCharacterId))

  if (!hasRenderableHonor) {
    return <ProfileHonorFallbackPanel honors={honors} />
  }

  return (
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
      <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>展示称号</span>
      <div style={{ display: 'flex', gap: theme.spacing.sm, flexWrap: 'nowrap', alignItems: 'center' }}>
        {honors.slice(0, 3).map((honor, index) => (
          <HonorBadge key={`${honor.seq ?? honor.honorId}-${index}`} honor={honor} />
        ))}
      </div>
    </div>
  )
}

function ProfileHonorFallbackPanel({ honors }: { honors: ProfileHonor[] }) {
  return (
    <div style={{ display: 'flex', gap: theme.spacing.sm, flexWrap: 'wrap' }}>
      {honors.slice(0, 3).map((honor, index) => (
        <HonorTextFallback key={`${honor.seq ?? honor.honorId}-${index}`} honor={honor} />
      ))}
    </div>
  )
}

function HonorBadge({ honor }: { honor: ProfileHonor }) {
  if (honor.honorType === 'bonds' && honor.leftCharacterId && honor.rightCharacterId) {
    return <BondsHonorBadge honor={honor} />
  }
  if (!honor.imageUrl) {
    return <HonorTextFallback honor={honor} />
  }

  return (
    <HonorBadgeFrame honor={honor}>
      <img src={honor.imageUrl} width={HONOR_BADGE_WIDTH} height={HONOR_BADGE_HEIGHT} style={{ objectFit: 'fill' }} />
    </HonorBadgeFrame>
  )
}

function HonorBadgeFrame({ honor, children }: { honor: ProfileHonor; children: any }) {
  const scale = HONOR_BADGE_WIDTH / 380
  const displayLevel = honorLevelDisplay(honor.level)
  const normalStars = Math.min(displayLevel, 5)
  const plusStars = Math.max(displayLevel - 5, 0)
  const frameUrl = DEFAULT_HONOR_FRAME_URLS[honor.honorRarity ?? ''] ?? honor.frameUrl
  const levelIconUrl = DEFAULT_HONOR_LEVEL_ICON_URL ?? honor.levelIconUrl
  const levelIcon6Url = DEFAULT_HONOR_LEVEL_ICON6_URL ?? honor.levelIcon6Url
  const starSize = Math.max(7, Math.round(16 * scale))
  const starX = Math.round(50 * scale)
  const starY = HONOR_BADGE_HEIGHT - starSize - 2

  return (
    <div style={{ display: 'flex', position: 'relative', width: HONOR_BADGE_WIDTH, height: HONOR_BADGE_HEIGHT, overflow: 'hidden', borderRadius: HONOR_BADGE_HEIGHT / 2, flexShrink: 0 }}>
      {children}
      {frameUrl && <img src={frameUrl} width={HONOR_BADGE_WIDTH} height={HONOR_BADGE_HEIGHT} style={{ position: 'absolute', left: 0, top: 0, objectFit: 'fill' }} />}
      {levelIconUrl && normalStars > 0 && Array.from({ length: normalStars }).map((_, i) => (
        <img key={`lv-${i}`} src={levelIconUrl} width={starSize} height={starSize} style={{ position: 'absolute', left: starX + i * starSize, top: starY }} />
      ))}
      {levelIcon6Url && plusStars > 0 && Array.from({ length: plusStars }).map((_, i) => (
        <img key={`lv6-${i}`} src={levelIcon6Url} width={starSize} height={starSize} style={{ position: 'absolute', left: starX + i * starSize, top: starY }} />
      ))}
    </div>
  )
}

function BondsHonorBadge({ honor }: { honor: ProfileHonor }) {
  const leftColor = honor.leftColor ?? theme.colors.accentLight
  const rightColor = honor.rightColor ?? theme.colors.accentSoft
  const leftUrl = honor.leftCharacterUrl ?? getBondsHonorCharacterUrl(honor.leftCharacterId!, honorAssetSource(honor))
  const rightUrl = honor.rightCharacterUrl ?? getBondsHonorCharacterUrl(honor.rightCharacterId!, honorAssetSource(honor))
  const scale = HONOR_BADGE_WIDTH / 380
  const maskLeft = Math.round(10 * scale)
  const maskWidth = Math.round(360 * scale)

  return (
    <HonorBadgeFrame honor={honor}>
      <div
        style={{
          display: 'flex',
          position: 'absolute',
          left: maskLeft,
          top: 0,
          width: maskWidth,
          height: HONOR_BADGE_HEIGHT,
          borderRadius: HONOR_BADGE_HEIGHT / 2,
          overflow: 'hidden',
        }}
      >
        <div style={{ display: 'flex', position: 'absolute', left: 0, top: 0, width: Math.round(180 * scale), height: HONOR_BADGE_HEIGHT, backgroundColor: leftColor }} />
        <div style={{ display: 'flex', position: 'absolute', left: Math.round(180 * scale), top: 0, width: Math.round(180 * scale), height: HONOR_BADGE_HEIGHT, backgroundColor: rightColor }} />
        <div
          style={{
            display: 'flex',
            position: 'absolute',
            left: Math.round(6 * scale),
            top: Math.round(6 * scale),
            width: Math.round(348 * scale),
            height: Math.round(68 * scale),
            borderRadius: Math.round(34 * scale),
            border: `${Math.max(2, Math.round(8 * scale))}px solid #fff`,
          }}
        />
        <img
          src={leftUrl}
          width={Math.round(160 * scale)}
          height={Math.round(136 * scale)}
          style={{
            position: 'absolute',
            left: Math.round(10 * scale),
            top: Math.round(-43 * scale),
            objectFit: 'contain',
          }}
        />
        <img
          src={rightUrl}
          width={Math.round(160 * scale)}
          height={Math.round(136 * scale)}
          style={{
            position: 'absolute',
            left: Math.round(190 * scale),
            top: Math.round(-43 * scale),
            objectFit: 'contain',
          }}
        />
        {honor.bondsHonorWordUrl && (
          <img
            src={honor.bondsHonorWordUrl}
            width={Math.round(200 * scale)}
            height={Math.round(44 * scale)}
            style={{
              position: 'absolute',
              left: Math.round(80 * scale),
              top: Math.round(18 * scale),
              objectFit: 'contain',
            }}
          />
        )}
      </div>
    </HonorBadgeFrame>
  )
}

function honorAssetSource(honor: ProfileHonor): AssetSourceType | string {
  return honor.assetSource ?? 'main-jp'
}

function HonorTextFallback({ honor }: { honor: ProfileHonor }) {
  return (
    <span
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
  )
}

function honorLevelDisplay(level: number | undefined): number {
  if (!level || level <= 0) return 0
  return level > 10 ? level - 10 : level
}

function MusicClearPanel({ counts }: { counts: MusicClearCount[] }) {
  const maxValue = Math.max(1, ...counts.flatMap(count => [count.liveClear, count.fullCombo, count.allPerfect]))

  return (
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
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
        <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>打歌统计</span>
        <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: 10, fontWeight: 800 }}>Clear / FC / AP</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm }}>
        {counts.map((count) => (
          <MusicClearRow key={count.difficulty} count={count} maxValue={maxValue} />
        ))}
      </div>
    </div>
  )
}

function MusicClearRow({ count, maxValue }: { count: MusicClearCount; maxValue: number }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: theme.spacing.sm }}>
      <span style={{ display: 'flex', width: 70, color: difficultyColor(count.difficulty), fontSize: theme.fontSize.xs, fontWeight: 900 }}>
        {difficultyLabel(count.difficulty)}
      </span>
      <MusicStatBar label="Clear" value={count.liveClear} maxValue={maxValue} color={difficultyColor(count.difficulty)} />
      <MusicStatBar label="FC" value={count.fullCombo} maxValue={maxValue} color={theme.colors.success} />
      <MusicStatBar label="AP" value={count.allPerfect} maxValue={maxValue} color={theme.colors.warning} />
    </div>
  )
}

function MusicStatBar({ label, value, maxValue, color }: { label: string; value: number; maxValue: number; color: string }) {
  const width = Math.max(5, Math.round((value / maxValue) * 92))

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 3, flex: 1 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', gap: 4 }}>
        <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: 9, fontWeight: 800 }}>{label}</span>
        <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: 9, fontWeight: 900 }}>{value.toLocaleString()}</span>
      </div>
      <div style={{ display: 'flex', width: 96, height: 5, borderRadius: theme.borderRadius.round, backgroundColor: theme.colors.surfaceLight, overflow: 'hidden' }}>
        <div style={{ display: 'flex', width, height: 5, borderRadius: theme.borderRadius.round, backgroundColor: color }} />
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
        gap: theme.spacing.md,
        backgroundColor: theme.colors.surface,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: theme.borderRadius.xl,
        padding: theme.spacing.md,
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
        <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.md, fontWeight: 900 }}>角色等级</span>
        <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: 10, fontWeight: 800 }}>{ranks.length} CHARACTERS</span>
      </div>
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: theme.spacing.sm }}>
        {ranks.map((rank) => (
          <CharacterRankBadge key={rank.characterId} rank={rank} />
        ))}
      </div>
    </div>
  )
}

function CharacterRankBadge({ rank }: { rank: CharacterRank }) {
  const iconUrl = getCharacterIconUrl(rank.characterId)

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 6,
        width: 128,
        padding: 6,
        borderRadius: theme.borderRadius.lg,
        backgroundColor: theme.colors.surfaceLight,
        border: `1px solid ${theme.colors.border}`,
      }}
    >
      <img src={iconUrl} width={32} height={32} style={{ objectFit: 'contain', borderRadius: 16, flexShrink: 0 }} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 1, minWidth: 0 }}>
        <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: 10, fontWeight: 800 }}>{shortCharacterName(rank.characterName)}</span>
        <span style={{ display: 'flex', color: theme.colors.accent, fontSize: theme.fontSize.xs, fontWeight: 900 }}>Lv.{rank.rank}</span>
      </div>
    </div>
  )
}

function shortCharacterName(name: string): string {
  if (!name) return '-'
  return name.length > 5 ? name.slice(0, 5) : name
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
          compositeLayers={card.compositeLayers}
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

import type { ReactNode } from 'react'
import { getCardFullUrl, getCardThumbnailUrl, getEventBannerUrl, type AssetSourceType } from '../../shared'
import { BaseCard } from './base'
import { getLocalIconAssetDataUri, getSekaiCardUiAssetDataUri } from '../styles/assets'
import { theme } from '../styles/theme'
import { SekaiCardThumbnail, canUseTrainedArt, getAttributeColor } from './SekaiCardThumbnail'

// Mirrors the attribute icon mapping used in SekaiCardThumbnail (assets/icon
// uses lowercase for `cute`, capitalized first letter for the rest).
const ATTRIBUTE_ICON_FILES: Record<string, string> = {
  cute: 'cute.png',
  cool: 'Cool.png',
  pure: 'Pure.png',
  happy: 'Happy.png',
  mysterious: 'Mysterious.png',
}

function attrIconUrl(attr: string): string | undefined {
  const file = ATTRIBUTE_ICON_FILES[attr]
  return (file && getLocalIconAssetDataUri(file)) || getSekaiCardUiAssetDataUri(`attr_${attr}.png`)
}

function starIconUrl(birthday: boolean): string | undefined {
  return getLocalIconAssetDataUri(birthday ? 'birthday.png' : 'star.png')
    ?? getSekaiCardUiAssetDataUri(birthday ? 'rare_birthday.png' : 'rare_star_normal.png')
}

function rarityStarCount(rarity: string): number {
  if (rarity === 'rarity_birthday') return 1
  const m = /rarity[_-]?(\d+)/.exec(rarity)
  return m ? Number(m[1]) : 0
}

interface CardSkill {
  id?: number
  level?: number
  description?: string
  spriteName?: string
}

interface CardCostume {
  costumeNumber?: number
  name?: string
  rarity?: string
  source?: string
  designer?: string
  partTypes?: string[]
  thumbnailUrls?: string[]
}

interface CardDetailProps {
  card: {
    id: number
    prefix: string
    characterName: string
    rarity: string
    attr: string
    thumbnailUrl?: string
    normalThumbnailUrl?: string
    trainedThumbnailUrl?: string
    normalFullUrl?: string
    trainedFullUrl?: string
    assetbundleName?: string
    characterId?: number
    cardRarityType?: string
    assetSource?: AssetSourceType | string
    power?: number
    skillName?: string
    skill?: CardSkill
    trainedSkill?: CardSkill
    costumes?: CardCostume[]
    gachaPhrase?: string
    supplyType?: string
    trained?: boolean
    events?: Array<{
      id: number
      name: string
      eventType?: string
      assetbundleName?: string
      startAt?: number
      closedAt?: number
      unit?: string
    }>
    compositeLayers?: import('../card-thumbnail-composites').CardThumbnailCompositeLayer[]
    normalCompositeLayers?: import('../card-thumbnail-composites').CardThumbnailCompositeLayer[]
    trainedCompositeLayers?: import('../card-thumbnail-composites').CardThumbnailCompositeLayer[]
  }
}

export function CardDetail({ card }: CardDetailProps) {
  const rarity = card.cardRarityType ?? card.rarity
  const attrColor = getAttributeColor(card.attr)
  const canRenderTrained = canUseTrainedArt(rarity)
  const source = card.assetSource ?? 'main-jp'

  const normalThumbnailUrl = card.normalThumbnailUrl
    ?? card.thumbnailUrl
    ?? resolveCardThumbnail(card.assetbundleName, false, source)
  const trainedThumbnailUrl = card.trainedThumbnailUrl
    ?? resolveCardThumbnail(card.assetbundleName, true, source)
  const normalFullUrl = card.normalFullUrl
    ?? resolveCardFull(card.assetbundleName, false, source)
    ?? normalThumbnailUrl
  const trainedFullUrl = card.trainedFullUrl
    ?? resolveCardFull(card.assetbundleName, true, source)
    ?? trainedThumbnailUrl
  const showTrained = canRenderTrained && Boolean(trainedFullUrl)

  const hasSkill = !!(card.skill?.description || card.trainedSkill?.description)
  const hasCostumes = !!(card.costumes && card.costumes.length > 0)
  const hasEvents = !!(card.events && card.events.length > 0)

  return (
    <BaseCard
      title={card.prefix}
      subtitle={`${card.characterName} · ID: ${card.id} · ${rarity}`}
      accentColor={attrColor}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: theme.spacing.sm }}>
        {/* ── Hero: two art panels ── */}
        <div style={{ display: 'flex', gap: theme.spacing.sm }}>
          <ArtPanel imageUrl={normalFullUrl} accentColor={attrColor} />
          <ArtPanel
            imageUrl={showTrained ? trainedFullUrl : normalFullUrl}
            accentColor={showTrained ? theme.colors.accent : theme.colors.textMuted}
            muted={!showTrained}
          />
        </div>

        {/* ── Info strip: thumbnails + metadata grid ── */}
        <div style={{ display: 'flex', gap: theme.spacing.sm, alignItems: 'stretch' }}>
          <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
            <SekaiCardThumbnail
              imageUrl={normalThumbnailUrl}
              compositeLayers={card.normalCompositeLayers ?? card.compositeLayers}
              rarity={rarity}
              attr={card.attr}
              isTrained={false}
              characterName={card.characterName}
              supplyType={card.supplyType}
              size={116}
            />
            {showTrained && (
              <SekaiCardThumbnail
                imageUrl={trainedThumbnailUrl}
                compositeLayers={card.trainedCompositeLayers}
                rarity={rarity}
                attr={card.attr}
                isTrained
                characterName={card.characterName}
                supplyType={card.supplyType}
                size={116}
              />
            )}
          </div>

          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              gap: 6,
              flex: 1,
              backgroundColor: theme.colors.surface,
              border: `1px solid ${theme.colors.border}`,
              borderRadius: theme.borderRadius.md,
              padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
            }}
          >
            <InfoRow label="属性" value={
              <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
                {attrIconUrl(card.attr) && (
                  <img src={attrIconUrl(card.attr)} width={16} height={16} style={{ objectFit: 'contain' }} />
                )}
                <span style={{ display: 'flex', color: attrColor, fontWeight: 800, fontSize: theme.fontSize.sm }}>{card.attr}</span>
              </span>
            } />
            <InfoRow label="稀有度" value={
              <span style={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                {(() => {
                  const birthday = rarity === 'rarity_birthday'
                  const count = birthday ? 1 : rarityStarCount(rarity)
                  const url = starIconUrl(birthday)
                  if (!url || count <= 0) return <span style={{ display: 'flex', color: theme.colors.text, fontWeight: 800, fontSize: theme.fontSize.sm }}>{rarity}</span>
                  return Array.from({ length: count }).map((_, i) => (
                    <img key={i} src={url} width={14} height={14} style={{ objectFit: 'contain' }} />
                  ))
                })()}
              </span>
            } />
            {card.power != null && card.power > 0 && <InfoRow label="综合力" value={card.power.toLocaleString()} />}
            {card.skillName && <InfoRow label="技能" value={card.skillName} />}
            {card.supplyType && <InfoRow label="获取类型" value={card.supplyType} />}
          </div>
        </div>

        {/* ── Skill description (full width, compact) ── */}
        {hasSkill && (
          <SkillSection
            accentColor={attrColor}
            skillName={card.skillName}
            normal={card.skill}
            trained={card.trainedSkill}
          />
        )}

        {/* ── Bottom row: costumes (left) + events (right) ── */}
        {(hasCostumes || hasEvents) && (
          <div style={{ display: 'flex', gap: theme.spacing.sm, alignItems: 'stretch' }}>
            {hasCostumes && (
              <div style={{ display: 'flex', flex: 1, minWidth: 0 }}>
                <CostumeSection costumes={card.costumes!} accentColor={attrColor} />
              </div>
            )}
            {hasEvents && (
              <div style={{ display: 'flex', flex: 1, minWidth: 0 }}>
                <EventSection events={card.events!} accentColor={attrColor} source={source} />
              </div>
            )}
          </div>
        )}

        {/* ── Gacha phrase footer ── */}
        {card.gachaPhrase && (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              padding: `6px ${theme.spacing.md}px`,
              borderRadius: theme.borderRadius.md,
              backgroundColor: theme.colors.surfaceAccent,
              border: `1px solid ${theme.colors.borderStrong}`,
            }}
          >
            <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.xs, lineHeight: 1.4 }}>
              "{card.gachaPhrase}"
            </span>
          </div>
        )}
      </div>
    </BaseCard>
  )
}

function ArtPanel({
  imageUrl,
  accentColor,
  muted = false,
}: {
  imageUrl?: string
  accentColor: string
  muted?: boolean
}) {
  return (
    <div
      style={{
        display: 'flex',
        position: 'relative',
        flex: 1,
        height: 216,
        borderRadius: theme.borderRadius.md,
        overflow: 'hidden',
        backgroundColor: theme.colors.surface,
        border: `1px solid ${muted ? theme.colors.border : accentColor}`,
      }}
    >
      <img
        src={imageUrl ?? panelPlaceholder(accentColor)}
        width={364}
        height={216}
        style={{ width: '100%', height: '100%', objectFit: 'cover', objectPosition: 'center top', opacity: muted ? 0.72 : 1 }}
      />
    </div>
  )
}

function InfoRow({ label, value, color }: { label: string; value: ReactNode; color?: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: theme.spacing.sm }}>
      <span style={{ display: 'flex', color: theme.colors.textSecondary, fontSize: theme.fontSize.xs, flexShrink: 0 }}>{label}</span>
      <span
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'flex-end',
          color: color ?? theme.colors.text,
          fontSize: theme.fontSize.xs,
          fontWeight: 800,
          textAlign: 'right',
          maxWidth: 320,
        }}
      >
        {value}
      </span>
    </div>
  )
}

function resolveCardThumbnail(assetbundleName: string | undefined, trained: boolean, source: AssetSourceType | string): string | undefined {
  return assetbundleName ? getCardThumbnailUrl(assetbundleName, trained, source, 'png') : undefined
}

function resolveCardFull(assetbundleName: string | undefined, trained: boolean, source: AssetSourceType | string): string | undefined {
  return assetbundleName ? getCardFullUrl(assetbundleName, trained, source, 'png') : undefined
}

function SectionHeader({ title, badge, accentColor }: { title: string; badge?: string; accentColor: string }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
        <div style={{ display: 'flex', width: 4, height: 14, borderRadius: theme.borderRadius.round, backgroundColor: accentColor }} />
        <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.sm, fontWeight: 900 }}>{title}</span>
      </div>
      {badge && <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: 11, fontWeight: 700 }}>{badge}</span>}
    </div>
  )
}

function SkillSection({
  accentColor,
  skillName,
  normal,
  trained,
}: {
  accentColor: string
  skillName?: string
  normal?: CardSkill
  trained?: CardSkill
}) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 6,
        borderRadius: theme.borderRadius.md,
        backgroundColor: theme.colors.surface,
        border: `1px solid ${theme.colors.border}`,
        padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <div style={{ display: 'flex', width: 4, height: 14, borderRadius: theme.borderRadius.round, backgroundColor: accentColor }} />
          <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.sm, fontWeight: 900 }}>技能描述</span>
          {skillName && (
            <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: 11, fontWeight: 700 }}>{skillName}</span>
          )}
        </div>
        <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: 11, fontWeight: 700 }}>
          Lv.{normal?.level ?? trained?.level ?? 4}
        </span>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        {normal?.description && (
          <SkillDescriptionRow label="开花前" description={normal.description} accentColor={theme.colors.textMuted} />
        )}
        {trained?.description && (
          <SkillDescriptionRow label="开花后" description={trained.description} accentColor={accentColor} />
        )}
      </div>
    </div>
  )
}

function SkillDescriptionRow({ label, description, accentColor }: { label: string; description: string; accentColor: string }) {
  return (
    <div
      style={{
        display: 'flex',
        gap: theme.spacing.sm,
        alignItems: 'baseline',
        padding: `4px ${theme.spacing.sm}px`,
        borderRadius: theme.borderRadius.sm,
        backgroundColor: theme.colors.background,
        borderLeft: `3px solid ${accentColor}`,
      }}
    >
      <span style={{ display: 'flex', color: accentColor, fontSize: 11, fontWeight: 800, flexShrink: 0 }}>{label}</span>
      <span style={{ display: 'flex', color: theme.colors.text, fontSize: theme.fontSize.xs, lineHeight: 1.5 }}>{description}</span>
    </div>
  )
}

function CostumeSection({ costumes, accentColor }: { costumes: CardCostume[]; accentColor: string }) {
  const visible = costumes.slice(0, 3)
  const overflow = costumes.length - visible.length
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        gap: 6,
        borderRadius: theme.borderRadius.md,
        backgroundColor: theme.colors.surface,
        border: `1px solid ${theme.colors.border}`,
        padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
      }}
    >
      <SectionHeader title="关联服装" badge={`共 ${costumes.length} 套`} accentColor={accentColor} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        {visible.map((costume, idx) => {
          const thumbs = (costume.thumbnailUrls ?? []).slice(0, 2)
          const partsLabel = (costume.partTypes ?? []).map(p => COSTUME_PART_LABELS[p] ?? p).join('/')
          return (
            <div
              key={`${costume.costumeNumber ?? idx}-${idx}`}
              style={{
                display: 'flex',
                gap: theme.spacing.sm,
                alignItems: 'center',
                padding: `4px ${theme.spacing.sm}px`,
                borderRadius: theme.borderRadius.sm,
                backgroundColor: theme.colors.background,
              }}
            >
              <div style={{ display: 'flex', gap: 3, flexShrink: 0 }}>
                {thumbs.length > 0 ? thumbs.map((url, i) => (
                  <div key={i} style={{ display: 'flex', width: 36, height: 36, borderRadius: 4, overflow: 'hidden', backgroundColor: theme.colors.surface, border: `1px solid ${theme.colors.border}` }}>
                    <img src={url} width={36} height={36} style={{ objectFit: 'cover' }} />
                  </div>
                )) : (
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', width: 36, height: 36, borderRadius: 4, backgroundColor: theme.colors.surface, border: `1px solid ${theme.colors.border}`, color: theme.colors.textMuted, fontSize: 9, fontWeight: 800 }}>N/A</div>
                )}
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', flex: 1, minWidth: 0, gap: 1 }}>
                <span style={{ display: 'flex', color: theme.colors.text, fontSize: 11, fontWeight: 800, lineHeight: 1.2 }}>
                  {costume.name ?? `#${costume.costumeNumber ?? ''}`}
                </span>
                {(partsLabel || costume.rarity) && (
                  <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: 10, fontWeight: 700 }}>
                    {[partsLabel, costume.rarity ? COSTUME_RARITY_LABELS[costume.rarity] ?? costume.rarity : ''].filter(Boolean).join(' · ')}
                  </span>
                )}
              </div>
            </div>
          )
        })}
        {overflow > 0 && (
          <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: 11, fontWeight: 700, paddingLeft: theme.spacing.sm }}>+{overflow} 套</span>
        )}
      </div>
    </div>
  )
}

function EventSection({
  events,
  accentColor,
  source,
}: {
  events: NonNullable<CardDetailProps['card']['events']>
  accentColor: string
  source: AssetSourceType | string
}) {
  const visible = events.slice(0, 3)
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        gap: 6,
        borderRadius: theme.borderRadius.md,
        backgroundColor: theme.colors.surface,
        border: `1px solid ${theme.colors.border}`,
        padding: `${theme.spacing.sm}px ${theme.spacing.md}px`,
      }}
    >
      <SectionHeader title="关联活动" badge={`共 ${events.length} 个`} accentColor={accentColor} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        {visible.map((event) => {
          const bannerUrl = event.assetbundleName ? getEventBannerUrl(event.assetbundleName, source) : undefined
          const typeLabel = EVENT_TYPE_LABELS[event.eventType ?? ''] ?? event.eventType ?? ''
          return (
            <div
              key={event.id}
              style={{
                display: 'flex',
                gap: theme.spacing.sm,
                alignItems: 'center',
                borderRadius: theme.borderRadius.sm,
                backgroundColor: theme.colors.background,
                overflow: 'hidden',
              }}
            >
              {bannerUrl && (
                <div style={{ display: 'flex', width: 100, height: 44, flexShrink: 0, overflow: 'hidden' }}>
                  <img src={bannerUrl} width={200} height={44} style={{ width: 100, height: 44, objectFit: 'cover', objectPosition: 'center' }} />
                </div>
              )}
              <div style={{ display: 'flex', flexDirection: 'column', flex: 1, minWidth: 0, gap: 1, padding: `4px ${bannerUrl ? 0 : theme.spacing.sm}px 4px ${bannerUrl ? 0 : theme.spacing.sm}px` }}>
                <span style={{ display: 'flex', color: theme.colors.text, fontSize: 11, fontWeight: 800, lineHeight: 1.2 }}>{event.name}</span>
                <span style={{ display: 'flex', color: theme.colors.textMuted, fontSize: 10, fontWeight: 700 }}>#{event.id}{typeLabel ? ` · ${typeLabel}` : ''}</span>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

const EVENT_TYPE_LABELS: Record<string, string> = {
  marathon: '马拉松',
  cheerful_carnival: '欢乐嘉年华',
  world_bloom: '世界绽放',
}

const COSTUME_PART_LABELS: Record<string, string> = {
  head: '发饰',
  hair: '发型',
  body: '服装',
}

const COSTUME_RARITY_LABELS: Record<string, string> = {
  rare: '稀有',
  normal: '普通',
}

function panelPlaceholder(color: string): string {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="720" height="428" viewBox="0 0 720 428">
  <defs>
    <linearGradient id="g" x1="0" x2="1" y1="0" y2="1">
      <stop offset="0" stop-color="#ffffff"/>
      <stop offset="1" stop-color="${color}" stop-opacity="0.36"/>
    </linearGradient>
  </defs>
  <rect width="720" height="428" rx="30" fill="url(#g)"/>
  <circle cx="584" cy="86" r="102" fill="${color}" opacity="0.2"/>
  <circle cx="126" cy="340" r="126" fill="#fff" opacity="0.55"/>
</svg>`
  return `data:image/svg+xml;base64,${Buffer.from(svg, 'utf8').toString('base64')}`
}


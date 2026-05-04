// Snowy Viewer-compatible asset URL helpers for PJSK assets.
// The path rules mirror moesekai / Snowy Viewer so card normal / after_training
// art can be shared between bot renderers and the web viewer.

export type AssetSourceType =
  | 'main-jp'
  | 'backup-jp'
  | 'overseas-jp'
  | 'overseas-backup-jp'
  | 'main-cn'
  | 'backup-cn'
  | 'overseas-cn'
  | 'overseas-backup-cn'
  | 'sekai-best-jp'
  | 'sekai-best-cn'
  | 'sekai-best-tw'
  | 'sekai-best-kr'
  | 'sekai-best-en'

export const MOE_STATIC_BASE_URL = 'https://moe.exmeaning.com'
export const MOE_ASSETS_BASE_URL = `${MOE_STATIC_BASE_URL}/assets`
export const MOE_LOGO_URL = `${MOE_ASSETS_BASE_URL}/logo.svg`

export const ASSET_DOMAIN_MAIN = 'https://storage.exmeaning.com'
export const ASSET_DOMAIN_BACKUP = 'https://storage2.exmeaning.com'
export const ASSET_DOMAIN_OVERSEAS = 'https://storage.pjsk.moe'
export const ASSET_DOMAIN_OVERSEAS_BACKUP = 'https://storage2.pjsk.moe'

export const ASSET_BASE_URL_MAP: Record<AssetSourceType, string> = {
  'main-jp': `${ASSET_DOMAIN_MAIN}/sekai-jp-assets`,
  'backup-jp': `${ASSET_DOMAIN_BACKUP}/sekai-jp-assets`,
  'overseas-jp': `${ASSET_DOMAIN_OVERSEAS}/sekai-jp-assets`,
  'overseas-backup-jp': `${ASSET_DOMAIN_OVERSEAS_BACKUP}/sekai-jp-assets`,
  'main-cn': `${ASSET_DOMAIN_MAIN}/sekai-cn-assets`,
  'backup-cn': `${ASSET_DOMAIN_BACKUP}/sekai-cn-assets`,
  'overseas-cn': `${ASSET_DOMAIN_OVERSEAS}/sekai-cn-assets`,
  'overseas-backup-cn': `${ASSET_DOMAIN_OVERSEAS_BACKUP}/sekai-cn-assets`,
  'sekai-best-jp': 'https://storage.sekai.best/sekai-jp-assets',
  'sekai-best-cn': 'https://storage.sekai.best/sekai-cn-assets',
  'sekai-best-tw': 'https://storage.sekai.best/sekai-tc-assets',
  'sekai-best-kr': 'https://storage.sekai.best/sekai-kr-assets',
  'sekai-best-en': 'https://storage.sekai.best/sekai-en-assets',
}

export function getAssetCdnBase(source: AssetSourceType | string = 'main-jp'): string {
  return getAssetBaseUrl(source)
}

export function getAssetBaseUrl(source: AssetSourceType | string = 'main-jp'): string {
  if (source.startsWith('http://') || source.startsWith('https://')) {
    return source.replace(/\/$/, '')
  }
  return ASSET_BASE_URL_MAP[source as AssetSourceType] ?? ASSET_BASE_URL_MAP['main-jp']
}

function buildImageAssetUrl(source: AssetSourceType | string, assetPath: string, format: 'png' | 'webp' = 'png'): string {
  const safeFormat = format === 'webp' ? 'png' : format
  return `${getAssetBaseUrl(source)}/${assetPath}.${safeFormat}`
}

function buildAudioAssetUrl(source: AssetSourceType | string, assetPath: string): string {
  return `${getAssetBaseUrl(source)}/${assetPath}.mp3`
}

function buildJsonAssetUrl(source: AssetSourceType | string, assetPath: string): string {
  return `${getAssetBaseUrl(source)}/${assetPath}.json`
}

function buildPngImageAssetUrl(source: AssetSourceType | string, assetPath: string): string {
  return `${getAssetBaseUrl(source)}/${assetPath}.png`
}

function withImageExtension(url: string, _format: 'webp' | 'png' = 'png'): string {
  return url.replace(/\.webp(?=([?#]|$))/i, '.png')
}

function resolveCardAssetArgs(
  first: string | number,
  second?: string | boolean,
  third?: boolean | AssetSourceType | string,
  fourth?: AssetSourceType | string,
): { assetbundleName: string; trained: boolean; source: AssetSourceType | string } {
  if (typeof first === 'number') {
    return {
      assetbundleName: String(second ?? ''),
      trained: typeof third === 'boolean' ? third : false,
      source: typeof third === 'string' ? third : fourth ?? 'main-jp',
    }
  }

  return {
    assetbundleName: first,
    trained: typeof second === 'boolean' ? second : false,
    source: typeof second === 'string' ? second : typeof third === 'string' ? third : 'main-jp',
  }
}

export function getCharacterIconUrl(characterId: number): string {
  return `${MOE_ASSETS_BASE_URL}/chr_ts_${characterId}.png`
}

export function getAreaItemThumbnailUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `thumbnail/areaitem/${assetbundleName}`)
}

export function getMaterialThumbnailUrl(materialId: number, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `thumbnail/material/material${materialId}`)
}

export function getCommonMaterialThumbnailUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `thumbnail/common_material/${assetbundleName}`)
}

export function getAttrIconUrl(attr: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `thumbnail/common/attribute/${attr}`)
}

export function getUnitLogoUrl(unitId: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `thumbnail/common/unit/${unitId}`)
}

export function getCardThumbnailUrl(assetbundleName: string, trained?: boolean, source?: AssetSourceType | string, format?: 'webp' | 'png'): string
export function getCardThumbnailUrl(characterId: number, assetbundleName: string, trained?: boolean, source?: AssetSourceType | string, format?: 'webp' | 'png'): string
export function getCardThumbnailUrl(
  first: string | number,
  second?: string | boolean,
  third?: boolean | AssetSourceType | string,
  fourth?: AssetSourceType | string,
  fifth?: 'webp' | 'png',
): string {
  const { assetbundleName, trained, source } = resolveCardAssetArgs(first, second, third, fourth)
  const format = (typeof first === 'number' ? fifth : typeof fourth === 'string' && (fourth === 'webp' || fourth === 'png') ? fourth : fifth) ?? 'png'
  const status = trained ? 'after_training' : 'normal'
  return withImageExtension(buildImageAssetUrl(source, `thumbnail/chara/${assetbundleName}_${status}`), format)
}

export function getCardFullUrl(assetbundleName: string, trained?: boolean, source?: AssetSourceType | string, format?: 'webp' | 'png'): string
export function getCardFullUrl(characterId: number, assetbundleName: string, trained?: boolean, source?: AssetSourceType | string, format?: 'webp' | 'png'): string
export function getCardFullUrl(
  first: string | number,
  second?: string | boolean,
  third?: boolean | AssetSourceType | string,
  fourth?: AssetSourceType | string,
  fifth?: 'webp' | 'png',
): string {
  const { assetbundleName, trained, source } = resolveCardAssetArgs(first, second, third, fourth)
  const format = (typeof first === 'number' ? fifth : typeof fourth === 'string' && (fourth === 'webp' || fourth === 'png') ? fourth : fifth) ?? 'png'
  const status = trained ? 'after_training' : 'normal'
  return withImageExtension(buildImageAssetUrl(source, `character/member/${assetbundleName}/card_${status}`), format)
}

export function getCostumeThumbnailUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `thumbnail/costume/${assetbundleName}`)
}

export function getMusicScoreUrl(musicId: number, difficulty: string, source: AssetSourceType | string = 'main-jp'): string {
  const paddedId = String(musicId).padStart(4, '0')
  return `${getAssetBaseUrl(source)}/music/music_score/${paddedId}_01/${difficulty}.txt`
}

export function getChartSvgUrl(musicId: number, difficulty: string): string {
  return `https://charts-new.unipjsk.com/moe/svg/${musicId}/${difficulty}.svg`
}

export function getMusicJacketUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `music/jacket/${assetbundleName}/${assetbundleName}`)
}

export function getMusicVocalAudioUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildAudioAssetUrl(source, `music/long/${assetbundleName}/${assetbundleName}`)
}

export function getEventBannerUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `event/${assetbundleName}/screen/bg`)
}

export function getEventCharacterUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `event/${assetbundleName}/screen/character`)
}

export function getEventLogoUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `event/${assetbundleName}/logo/logo`)
}

export function getEventStoryBannerUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `event_story/${assetbundleName}/screen_image/banner_event_story`)
}

export function getEventBgmUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildAudioAssetUrl(source, `event/${assetbundleName}/bgm/${assetbundleName}_top`)
}

export function getGachaLogoUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `gacha/${assetbundleName}/logo/logo`)
}

export function getGachaBannerUrl(gachaId: number, source?: AssetSourceType | string): string
export function getGachaBannerUrl(assetbundleName: string, source?: AssetSourceType | string): string
export function getGachaBannerUrl(gachaIdOrBundle: number | string, source: AssetSourceType | string = 'main-jp'): string {
  const assetbundleName = typeof gachaIdOrBundle === 'number'
    ? `banner_gacha${gachaIdOrBundle}`
    : gachaIdOrBundle
  return buildImageAssetUrl(source, `home/banner/${assetbundleName}/${assetbundleName}`)
}

export function getGachaScreenUrl(assetbundleName: string, gachaId: number, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `gacha/${assetbundleName}/screen/texture/bg_gacha${gachaId}`)
}

export function getCardGachaVoiceUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildAudioAssetUrl(source, `sound/gacha/get_voice/${assetbundleName}/${assetbundleName}`)
}

export function getComicUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `comic/one_frame/${assetbundleName}`)
}

export function getMangaImageUrl(id: number): string {
  return `${MOE_STATIC_BASE_URL}/mangas/${id}.png`
}

export function getStampUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildPngImageAssetUrl(source, `stamp/${assetbundleName}/${assetbundleName}`)
}

export function getVirtualLiveBannerUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `virtual_live/select/banner/${assetbundleName}/${assetbundleName}`)
}

export function getCharacterTrimUrl(characterId: number, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `character/character_select/chr_tl_${characterId}`)
}

export function getCharacterLabelHUrl(characterId: number, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `character/label/chr_h_lb_${characterId}`)
}

export function getCharacterLabelVUrl(characterId: number, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `character/label_vertical/chr_v_lb_${characterId}`)
}

export function getCharacterSelectUrl(characterId: number, source: AssetSourceType | string = 'main-jp'): string {
  return getCharacterTrimUrl(characterId, source)
}

export function getHonorBgUrl(assetbundleName: string, sub = false, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `honor/${assetbundleName}/degree_${sub ? 'sub' : 'main'}`)
}

export function getHonorFrameUrl(rarity: string, sub = false, source: AssetSourceType | string = 'main-jp'): string {
  const rarityMap: Record<string, number> = { low: 1, middle: 2, high: 3, highest: 4 }
  const num = rarityMap[rarity] || 1
  const size = sub ? 's' : 'm'
  return buildImageAssetUrl(source, `honor/frame/frame_degree_${size}_${num}`)
}

export function getHonorCustomFrameUrl(frameName: string, rarity: string, sub = false, source: AssetSourceType | string = 'main-jp'): string {
  const rarityMap: Record<string, number> = { low: 1, middle: 2, high: 3, highest: 4 }
  const num = rarityMap[rarity] || 1
  const size = sub ? 's' : 'm'
  return buildImageAssetUrl(source, `honor_frame/${frameName}/frame_degree_${size}_${num}`)
}

export function getHonorRankUrl(assetbundleName: string, type: 'rank' | 'scroll' = 'rank', sub = false, source: AssetSourceType | string = 'main-jp'): string {
  const suffix = type === 'rank' ? `rank_${sub ? 'sub' : 'main'}` : 'scroll'
  return buildImageAssetUrl(source, `honor/${assetbundleName}/${suffix}`)
}

export function getHonorRankMatchBgUrl(assetbundleName: string, sub = false, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `rank_live/honor/${assetbundleName}/degree_${sub ? 'sub' : 'main'}`)
}

export function getBondsHonorWordUrl(assetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `bonds_honor/word/${assetbundleName}_01`)
}

export function getBondsHonorCharacterUrl(characterId: number, source: AssetSourceType | string = 'main-jp'): string {
  const paddedId = String(characterId).padStart(2, '0')
  return buildImageAssetUrl(source, `bonds_honor/character/chr_sd_${paddedId}_01`)
}

export function getHonorLevelIconUrl(source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, 'honor/frame/icon_degreeLv')
}

export function getHonorUrl(assetbundleName: string, _level: string = 'rank1', source: AssetSourceType | string = 'main-jp'): string {
  return getHonorBgUrl(assetbundleName, false, source)
}

export function getScenarioJsonUrl(scenarioPath: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildJsonAssetUrl(source, scenarioPath)
}

export function getBackgroundImageUrl(bgName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `scenario/background/${bgName}/${bgName}`)
}

export function getStoryVoiceUrl(scenarioId: string, voiceId: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildAudioAssetUrl(source, `sound/scenario/voice/${scenarioId}/${voiceId}`)
}

export function getCardStoryVoiceUrl(scenarioId: string, voiceId: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildAudioAssetUrl(source, `sound/card_scenario/voice/${scenarioId}/${voiceId}`)
}

export function getAreaTalkVoiceUrl(scenarioId: string, voiceId: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildAudioAssetUrl(source, `sound/actionset/voice/${scenarioId}/${voiceId}`)
}

export function getSpecialStoryVoiceUrl(scenarioId: string, voiceId: string, source: AssetSourceType | string = 'main-jp'): string {
  return getStoryVoiceUrl(scenarioId, voiceId, source)
}

export function getStoryBgmUrl(bgmName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildAudioAssetUrl(source, `sound/scenario/bgm/${bgmName}/${bgmName}`)
}

export function getStorySoundEffectUrl(seName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildAudioAssetUrl(source, `sound/scenario/se/${seName}`)
}

export function getStoryEpisodeImageUrl(assetbundleName: string, episodeNo: number, source: AssetSourceType | string = 'main-jp'): string {
  const paddedNo = String(episodeNo).padStart(2, '0')
  return buildImageAssetUrl(source, `event_story/${assetbundleName}/episode_image/${assetbundleName}_${paddedNo}`)
}

export function getUnitStoryEpisodeImageUrl(chapterAssetbundleName: string, episodeAssetbundleName: string, source: AssetSourceType | string = 'main-jp'): string {
  return buildImageAssetUrl(source, `story/episode_image/${chapterAssetbundleName}/${episodeAssetbundleName}`)
}

export function getStickerUrl(stickerId: number, source: AssetSourceType | string = 'main-jp'): string {
  const padded = String(stickerId).padStart(4, '0')
  return buildPngImageAssetUrl(source, `stamp/${padded}/${padded}`)
}

// Base CDN URLs for PJSK assets
const ASSET_CDN_BASE = 'https://storage.sekai.best/sekai-jp-assets'
const ASSET_CDN_BASE_EN = 'https://storage.sekai.best/sekai-en-assets'

export function getAssetCdnBase(region: 'jp' | 'en' = 'jp'): string {
  return region === 'en' ? ASSET_CDN_BASE_EN : ASSET_CDN_BASE
}

export function getCardThumbnailUrl(
  assetbundleName: string,
  trained: boolean = false,
): string {
  const suffix = trained ? 'after_training' : 'normal'
  return `${ASSET_CDN_BASE}/thumbnail/chara_rip/${assetbundleName}_${suffix}.webp`
}

export function getCardFullUrl(
  assetbundleName: string,
  trained: boolean = false,
): string {
  const suffix = trained ? 'after_training' : 'normal'
  return `${ASSET_CDN_BASE}/character/member/${assetbundleName}_rip/card_${suffix}.webp`
}

export function getMusicJacketUrl(assetbundleName: string): string {
  return `${ASSET_CDN_BASE}/music/jacket/${assetbundleName}_rip/${assetbundleName}.webp`
}

export function getEventBannerUrl(assetbundleName: string): string {
  return `${ASSET_CDN_BASE}/event/${assetbundleName}/logo_rip/logo.webp`
}

export function getGachaBannerUrl(assetbundleName: string): string {
  return `${ASSET_CDN_BASE}/home/banner/${assetbundleName}_rip/${assetbundleName}.webp`
}

export function getHonorUrl(assetbundleName: string, _level: string = 'rank1'): string {
  return `${ASSET_CDN_BASE}/honor/${assetbundleName}_rip/degree_main.webp`
}

export function getStickerUrl(stickerId: number): string {
  const padded = String(stickerId).padStart(4, '0')
  return `${ASSET_CDN_BASE}/stamp/${padded}_rip/stamp${padded}/stamp${padded}.webp`
}

export function getCharacterIconUrl(characterId: number): string {
  return `${ASSET_CDN_BASE}/thumbnail/chara_rip/chr_ts_${characterId}.webp`
}

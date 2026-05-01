import { getAttrIconUrl, getCardFullUrl, getCardThumbnailUrl, getCharacterIconUrl, getEventBannerUrl, getEventCharacterUrl, getEventLogoUrl, getEventStoryBannerUrl, getGachaBannerUrl, getGachaLogoUrl, getMusicJacketUrl, getStickerUrl, getUnitLogoUrl } from '@moebot/shared'

// Re-export shared asset utilities with any bot-specific overrides
export {
  getAttrIconUrl,
  getCardThumbnailUrl,
  getCardFullUrl,
  getMusicJacketUrl,
  getEventBannerUrl,
  getEventCharacterUrl,
  getEventLogoUrl,
  getEventStoryBannerUrl,
  getGachaBannerUrl,
  getGachaLogoUrl,
  getStickerUrl,
  getCharacterIconUrl,
  getUnitLogoUrl,
}

/** Fetch an image from URL and return as Buffer */
export async function fetchImageBuffer(url: string): Promise<Buffer> {
  const response = await fetch(url)
  if (!response.ok) {
    throw new Error(`Failed to fetch image: ${response.status}`)
  }
  const arrayBuffer = await response.arrayBuffer()
  return Buffer.from(arrayBuffer)
}

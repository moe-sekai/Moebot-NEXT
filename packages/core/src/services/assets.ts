import { getCardThumbnailUrl, getCardFullUrl, getMusicJacketUrl, getEventBannerUrl, getGachaBannerUrl, getStickerUrl, getCharacterIconUrl } from '@moebot/shared'

// Re-export shared asset utilities with any bot-specific overrides
export {
  getCardThumbnailUrl,
  getCardFullUrl,
  getMusicJacketUrl,
  getEventBannerUrl,
  getGachaBannerUrl,
  getStickerUrl,
  getCharacterIconUrl,
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

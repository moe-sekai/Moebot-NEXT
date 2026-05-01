import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const imageCache = new Map<string, string>()

function assetPath(...segments: string[]): string | null {
  const candidates = [
    resolve(process.cwd(), 'assets', ...segments),
    resolve(__dirname, '../../../../assets', ...segments),
  ]
  return candidates.find(path => existsSync(path)) ?? null
}

export function getLocalAssetDataUri(...segments: string[]): string | undefined {
  const key = segments.join('/')
  const cached = imageCache.get(key)
  if (cached) return cached

  const path = assetPath(...segments)
  if (!path) return undefined

  const ext = path.split('.').pop()?.toLowerCase()
  if (ext === 'webp') {
    return undefined
  }

  const mime = ext === 'svg'
    ? 'image/svg+xml'
    : ext === 'jpg' || ext === 'jpeg'
      ? 'image/jpeg'
      : 'image/png'
  const dataUri = `data:${mime};base64,${readFileSync(path).toString('base64')}`
  imageCache.set(key, dataUri)
  return dataUri
}

export function getSekaiCardUiAssetDataUri(fileName: string): string | undefined {
  return getLocalAssetDataUri('sekai_cards_assets', fileName)
}

export function getLocalIconAssetDataUri(fileName: string): string | undefined {
  return getLocalAssetDataUri('icon', fileName)
}

export function getLocalFrameAssetDataUri(fileName: string): string | undefined {
  return getLocalAssetDataUri('frame', fileName)
}

export function getLocalMusicAssetDataUri(fileName: string): string | undefined {
  return getLocalAssetDataUri('music', fileName)
}

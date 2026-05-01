/** Format timestamp to readable date string */
export function formatDate(timestamp: number, locale: string = 'zh-CN'): string {
  return new Date(timestamp).toLocaleDateString(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

/** Format timestamp to readable datetime string */
export function formatDateTime(timestamp: number, locale: string = 'zh-CN'): string {
  return new Date(timestamp).toLocaleString(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/** Format card rarity type to display text */
export function formatRarity(rarityType: string): string {
  const map: Record<string, string> = {
    'rarity_1': '★',
    'rarity_2': '★★',
    'rarity_3': '★★★',
    'rarity_4': '★★★★',
    'rarity_birthday': '🎂 Birthday',
  }
  return map[rarityType] ?? rarityType
}

/** Format number with comma separators */
export function formatNumber(n: number): string {
  return n.toLocaleString()
}

/** Format milliseconds to human-readable duration */
export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  const minutes = Math.floor(ms / 60000)
  const seconds = Math.floor((ms % 60000) / 1000)
  return `${minutes}m${seconds}s`
}

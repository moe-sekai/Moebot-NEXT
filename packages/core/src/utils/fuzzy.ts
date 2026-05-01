/**
 * Simple fuzzy search implementation
 * Supports: exact match, prefix match, contains, pinyin-like matching
 */

export interface FuzzyResult<T> {
  item: T
  score: number
  matches: string[]
}

export function fuzzySearch<T>(
  items: T[],
  query: string,
  getSearchFields: (item: T) => string[],
  limit: number = 10,
): FuzzyResult<T>[] {
  const normalizedQuery = query.toLowerCase().trim()
  if (!normalizedQuery) return []

  const results: FuzzyResult<T>[] = []

  for (const item of items) {
    const fields = getSearchFields(item)
    let bestScore = 0
    const matches: string[] = []

    for (const field of fields) {
      if (!field) continue
      const normalizedField = field.toLowerCase()

      // Exact match
      if (normalizedField === normalizedQuery) {
        bestScore = Math.max(bestScore, 100)
        matches.push(field)
        continue
      }

      // Prefix match
      if (normalizedField.startsWith(normalizedQuery)) {
        bestScore = Math.max(bestScore, 80)
        matches.push(field)
        continue
      }

      // Contains match
      if (normalizedField.includes(normalizedQuery)) {
        bestScore = Math.max(bestScore, 60)
        matches.push(field)
        continue
      }

      // Word boundary match
      const words = normalizedField.split(/[\s_-]+/)
      if (words.some(w => w.startsWith(normalizedQuery))) {
        bestScore = Math.max(bestScore, 70)
        matches.push(field)
      }
    }

    if (bestScore > 0) {
      results.push({ item, score: bestScore, matches })
    }
  }

  return results
    .sort((a, b) => b.score - a.score)
    .slice(0, limit)
}

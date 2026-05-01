import { CHARACTERS } from '@moebot/shared'

/**
 * Character alias matching
 * Maps common aliases (Chinese, English, nicknames) to character IDs
 */
export function findCharacterByAlias(query: string): number | null {
  const normalized = query.toLowerCase().trim()

  for (const char of CHARACTERS) {
    // Check all aliases
    if (char.aliases.some(alias => alias.toLowerCase() === normalized)) {
      return char.id
    }
    // Check names
    if (
      char.firstName.toLowerCase() === normalized ||
      char.givenName.toLowerCase() === normalized ||
      char.fullName.toLowerCase() === normalized ||
      char.enName.toLowerCase() === normalized
    ) {
      return char.id
    }
  }

  return null
}

/**
 * Music title alias matching
 * TODO: Import music aliases from Snowy Viewer's musicAliases.ts
 */
export function findMusicByAlias(query: string, musics: any[]): any[] {
  const normalized = query.toLowerCase().trim()

  return musics.filter(m => {
    if (m.title?.toLowerCase().includes(normalized)) return true
    if (m.pronunciation?.toLowerCase().includes(normalized)) return true
    return false
  })
}

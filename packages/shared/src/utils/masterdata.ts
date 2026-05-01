export interface MasterdataStore {
  cards: any[]
  musics: any[]
  events: any[]
  gachas: any[]
  honors: any[]
  costumes: any[]
  gameCharacters: any[]
  musicDifficulties: any[]
  skills: any[]
  [key: string]: any[]
}

const MASTERDATA_FILES = [
  'cards',
  'musics',
  'events',
  'gachas',
  'honors',
  'costumes',
  'gameCharacters',
  'musicDifficulties',
  'skills',
  'eventCards',
  'gachaCards',
  'cardEpisodes',
  'musicVocals',
  'outsideCharacters',
] as const

export type MasterdataFile = (typeof MASTERDATA_FILES)[number]

export async function fetchMasterdata(
  baseUrl: string,
  file: MasterdataFile,
): Promise<any[]> {
  const url = `${baseUrl}/${file}.json`
  const response = await fetch(url)
  if (!response.ok) {
    throw new Error(`Failed to fetch ${file}: ${response.status} ${response.statusText}`)
  }
  return response.json()
}

export async function fetchAllMasterdata(baseUrl: string): Promise<MasterdataStore> {
  const store: Partial<MasterdataStore> = {}

  const results = await Promise.allSettled(
    MASTERDATA_FILES.map(async (file) => {
      store[file] = await fetchMasterdata(baseUrl, file)
    }),
  )

  const failures = results.filter((r) => r.status === 'rejected')
  if (failures.length > 0) {
    console.warn(
      `[masterdata] ${failures.length}/${MASTERDATA_FILES.length} files failed to fetch`,
    )
  }

  return store as MasterdataStore
}

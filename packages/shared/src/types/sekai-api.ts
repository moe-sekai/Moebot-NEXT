/** Configuration for connecting to a SEKAI game API endpoint */
export interface SekaiApiConfig {
  /** Whether this API endpoint is enabled */
  enabled: boolean
  /** Display name for this endpoint (e.g. "JP Server", "EN Server") */
  name: string
  /** Base URL of the API endpoint */
  baseUrl: string
  /** API version or region identifier */
  region: 'jp' | 'en' | 'tw' | 'kr'
  /** Request headers to include (e.g. authorization tokens, cookies) */
  headers: Record<string, string>
  /** Request timeout in ms */
  timeout: number
  /** Rate limit: max requests per minute */
  rateLimit: number
  /** Optional proxy URL */
  proxy?: string
}

/** User profile fetched from SEKAI API */
export interface SekaiUserProfile {
  userId: number
  name: string
  rank: number
  twitterId?: string
  profileImageType: string
  userGamedata: {
    totalPower: number
    multiLiveCount: number
    mvpCount: number
    superStarCount: number
  }
  userDecks: SekaiUserDeck[]
  userCards: SekaiUserCard[]
  userHonors: SekaiUserHonor[]
}

export interface SekaiUserDeck {
  deckId: number
  leader: number
  member1: number
  member2: number
  member3: number
  member4: number
  member5: number
}

export interface SekaiUserCard {
  cardId: number
  level: number
  masterRank: number
  specialTrainingStatus: string
  defaultImage: string
}

export interface SekaiUserHonor {
  honorId: number
  level: number
  obtainedAt: number
}

/** Response wrapper from SEKAI API */
export interface SekaiApiResponse<T> {
  httpStatus: number
  errorCode?: string
  message?: string
  data?: T
}

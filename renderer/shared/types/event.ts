export interface GameEvent {
  id: number
  eventType: EventType
  name: string
  assetbundleName: string
  bgmAssetbundleName: string
  startAt: number
  aggregateAt: number
  rankingAnnounceAt: number
  distributionStartAt: number
  closedAt: number
  distributionEndAt: number
  virtualLiveId: number
  unit: string
  eventRankingRewardRanges: EventRankingRewardRange[]
}

export type EventType = 'marathon' | 'cheerful_carnival' | 'world_bloom'

export interface EventRankingRewardRange {
  id: number
  eventId: number
  fromRank: number
  toRank: number
}

export interface EventRanking {
  userId: number
  score: number
  rank: number
  name: string
  userCard?: {
    cardId: number
    level: number
    masterRank: number
    specialTrainingStatus: string
  }
}

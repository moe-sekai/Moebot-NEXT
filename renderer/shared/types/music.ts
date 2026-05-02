export interface Music {
  id: number
  seq: number
  releaseConditionId: number
  categories: string[]
  title: string
  pronunciation: string
  creatorArtistId: number
  lyricist: string
  composer: string
  arranger: string
  dancerCount: number
  selfDancerPosition: number
  assetbundleName: string
  liveTalkBackgroundAssetbundleName: string
  publishedAt: number
  liveStageId: number
  fillerSec: number
  musicCollaborationId?: number
  isNewlyWrittenMusic?: boolean
}

export interface MusicDifficulty {
  id: number
  musicId: number
  musicDifficulty: DifficultyType
  playLevel: number
  noteCount: number
  totalNoteCount: number
}

export type DifficultyType =
  | 'easy'
  | 'normal'
  | 'hard'
  | 'expert'
  | 'master'
  | 'append'

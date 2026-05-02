export interface UnitStory {
  unit: string
  seq: number
  unitEpisodeCategory: string
  chapterNo: number
  episodes: StoryEpisode[]
}

export interface StoryEpisode {
  id: number
  episodeNo: number
  title: string
  assetbundleName: string
  scenarioId: string
  releaseConditionId: number
}

export interface Card {
  id: number
  seq: number
  characterId: number
  cardRarityType: CardRarityType
  specialTrainingPower: number
  attr: CardAttr
  supportUnit: string
  skillId: number
  cardSkillName: string
  prefix: string
  assetbundleName: string
  gachaPhrase: string
  flavorText: string
  releaseAt: number
  archivePublishedAt: number
  cardParameters: CardParameter[]
}

export type CardRarityType =
  | 'rarity_1'
  | 'rarity_2'
  | 'rarity_3'
  | 'rarity_4'
  | 'rarity_birthday'

export type CardAttr = 'cute' | 'cool' | 'pure' | 'happy' | 'mysterious'

export interface CardParameter {
  id: number
  cardId: number
  cardLevel: number
  cardParameterType: string
  power: number
}

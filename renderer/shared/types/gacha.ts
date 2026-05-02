export interface Gacha {
  id: number
  gachaType: string
  name: string
  seq: number
  assetbundleName: string
  gachaCeilItemId?: number
  startAt: number
  endAt: number
  isShowPeriod: boolean
  gachaCardRarityRateGroupId: number
  wishSelectCount: number
  gachaDetails: GachaDetail[]
}

export interface GachaDetail {
  id: number
  gachaId: number
  cardId: number
  weight: number
  isWish: boolean
}

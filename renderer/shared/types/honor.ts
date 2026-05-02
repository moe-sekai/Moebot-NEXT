export interface Honor {
  id: number
  seq: number
  groupId: number
  honorRarity: HonorRarity
  name: string
  assetbundleName: string
  levels: HonorLevel[]
}

export type HonorRarity = 'low' | 'middle' | 'high' | 'highest'

export interface HonorLevel {
  honorId: number
  level: number
  bonus: number
  description: string
}

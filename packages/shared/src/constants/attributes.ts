export const CARD_ATTRIBUTES = ['cute', 'cool', 'pure', 'happy', 'mysterious'] as const
export type CardAttribute = (typeof CARD_ATTRIBUTES)[number]

export const ATTRIBUTE_COLORS: Record<CardAttribute, string> = {
  cute: '#FF6699',
  cool: '#0077DD',
  pure: '#00BB33',
  happy: '#FFAA00',
  mysterious: '#BB44DD',
}

export const ATTRIBUTE_LABELS: Record<CardAttribute, { ja: string; zh: string; en: string }> = {
  cute: { ja: 'キュート', zh: '可爱', en: 'Cute' },
  cool: { ja: 'クール', zh: '帅气', en: 'Cool' },
  pure: { ja: 'ピュア', zh: '纯洁', en: 'Pure' },
  happy: { ja: 'ハッピー', zh: '快乐', en: 'Happy' },
  mysterious: { ja: 'ミステリアス', zh: '神秘', en: 'Mysterious' },
}

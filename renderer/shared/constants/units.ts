export type UnitId =
  | 'piapro'
  | 'light_sound'
  | 'idol'
  | 'street'
  | 'theme_park'
  | 'school_refusal'

export const UNITS: Record<UnitId, { name: string; fullName: string; color: string }> = {
  piapro: { name: 'VS', fullName: 'VIRTUAL SINGER', color: '#00BBDD' },
  light_sound: { name: 'Leo/need', fullName: 'Leo/need', color: '#4455DD' },
  idol: { name: 'MORE MORE JUMP!', fullName: 'MORE MORE JUMP!', color: '#88DD44' },
  street: { name: 'Vivid BAD SQUAD', fullName: 'Vivid BAD SQUAD', color: '#EE1166' },
  theme_park: {
    name: 'ワンダーランズ×ショウタイム',
    fullName: 'Wonderlands×Showtime',
    color: '#FF9900',
  },
  school_refusal: {
    name: '25時、ナイトコードで。',
    fullName: '25-ji, Nightcord de.',
    color: '#884499',
  },
}

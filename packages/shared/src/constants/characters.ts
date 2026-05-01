import type { UnitId } from './units'

export interface CharacterMeta {
  id: number
  firstName: string
  givenName: string
  fullName: string
  enName: string
  unit: UnitId
  color: string
  aliases: string[]
}

/** All 26 game characters + 6 Virtual Singers */
export const CHARACTERS: CharacterMeta[] = [
  // ── Leo/need ──
  {
    id: 1,
    firstName: '星乃',
    givenName: '一歌',
    fullName: '星乃一歌',
    enName: 'Ichika Hoshino',
    unit: 'light_sound',
    color: '#33AAEE',
    aliases: ['ichika', '一歌', 'ick'],
  },
  {
    id: 2,
    firstName: '天馬',
    givenName: '咲希',
    fullName: '天馬咲希',
    enName: 'Saki Tenma',
    unit: 'light_sound',
    color: '#FFDD44',
    aliases: ['saki', '咲希', '天马咲希'],
  },
  {
    id: 3,
    firstName: '望月',
    givenName: '穂波',
    fullName: '望月穂波',
    enName: 'Honami Mochizuki',
    unit: 'light_sound',
    color: '#EE6666',
    aliases: ['honami', '穂波', '望月�的波'],
  },
  {
    id: 4,
    firstName: '日野森',
    givenName: '志歩',
    fullName: '日野森志歩',
    enName: 'Shiho Hinomori',
    unit: 'light_sound',
    color: '#44BBAA',
    aliases: ['shiho', '志歩', '志步'],
  },

  // ── MORE MORE JUMP! ──
  {
    id: 5,
    firstName: '花里',
    givenName: 'みのり',
    fullName: '花里みのり',
    enName: 'Minori Hanasato',
    unit: 'idol',
    color: '#FFCCAA',
    aliases: ['minori', 'みのり', '花里实乃理', '实乃理'],
  },
  {
    id: 6,
    firstName: '桐谷',
    givenName: '遥',
    fullName: '桐谷遥',
    enName: 'Haruka Kiritani',
    unit: 'idol',
    color: '#AACCFF',
    aliases: ['haruka', '遥', '桐谷遥'],
  },
  {
    id: 7,
    firstName: '桃井',
    givenName: '愛莉',
    fullName: '桃井愛莉',
    enName: 'Airi Momoi',
    unit: 'idol',
    color: '#FF8899',
    aliases: ['airi', '愛莉', '爱莉', '桃井爱莉'],
  },
  {
    id: 8,
    firstName: '日野森',
    givenName: '雫',
    fullName: '日野森雫',
    enName: 'Shizuku Hinomori',
    unit: 'idol',
    color: '#77DDCC',
    aliases: ['shizuku', '雫', '日野森雫'],
  },

  // ── Vivid BAD SQUAD ──
  {
    id: 9,
    firstName: '小豆沢',
    givenName: 'こはね',
    fullName: '小豆沢こはね',
    enName: 'Kohane Azusawa',
    unit: 'street',
    color: '#FF6688',
    aliases: ['kohane', 'こはね', '小�的泽', '小豆泽心羽', '心羽'],
  },
  {
    id: 10,
    firstName: '白石',
    givenName: '杏',
    fullName: '白石杏',
    enName: 'An Shiraishi',
    unit: 'street',
    color: '#FFBB00',
    aliases: ['an', '杏', '白石杏'],
  },
  {
    id: 11,
    firstName: '東雲',
    givenName: '彰人',
    fullName: '東雲彰人',
    enName: 'Akito Shinonome',
    unit: 'street',
    color: '#FF7722',
    aliases: ['akito', '彰人', '东云彰人'],
  },
  {
    id: 12,
    firstName: '青柳',
    givenName: '冬弥',
    fullName: '青柳冬弥',
    enName: 'Touya Aoyagi',
    unit: 'street',
    color: '#0077BB',
    aliases: ['touya', 'toya', '冬弥', '青柳冬弥'],
  },

  // ── Wonderlands×Showtime ──
  {
    id: 13,
    firstName: '天馬',
    givenName: '司',
    fullName: '天馬司',
    enName: 'Tsukasa Tenma',
    unit: 'theme_park',
    color: '#FFBB33',
    aliases: ['tsukasa', '司', '天马司'],
  },
  {
    id: 14,
    firstName: '鳳',
    givenName: 'えむ',
    fullName: '鳳えむ',
    enName: 'Emu Otori',
    unit: 'theme_park',
    color: '#FF66BB',
    aliases: ['emu', 'えむ', '凤笑梦', '笑梦', '凤えむ'],
  },
  {
    id: 15,
    firstName: '草薙',
    givenName: '寧々',
    fullName: '草薙寧々',
    enName: 'Nene Kusanagi',
    unit: 'theme_park',
    color: '#33CC88',
    aliases: ['nene', '寧々', '宁宁', '草薙宁宁'],
  },
  {
    id: 16,
    firstName: '神代',
    givenName: '類',
    fullName: '神代類',
    enName: 'Rui Kamishiro',
    unit: 'theme_park',
    color: '#BB88EE',
    aliases: ['rui', '類', '类', '神代类'],
  },

  // ── 25時、ナイトコードで。 ──
  {
    id: 17,
    firstName: '宵崎',
    givenName: '奏',
    fullName: '宵崎奏',
    enName: 'Kanade Yoisaki',
    unit: 'school_refusal',
    color: '#BB6688',
    aliases: ['kanade', '奏', '宵崎奏', 'k'],
  },
  {
    id: 18,
    firstName: '朝比奈',
    givenName: 'まふゆ',
    fullName: '朝比奈まふゆ',
    enName: 'Mafuyu Asahina',
    unit: 'school_refusal',
    color: '#7788CC',
    aliases: ['mafuyu', 'まふゆ', '朝比奈真冬', '真冬'],
  },
  {
    id: 19,
    firstName: '東雲',
    givenName: '絵名',
    fullName: '東雲絵名',
    enName: 'Ena Shinonome',
    unit: 'school_refusal',
    color: '#CCAA88',
    aliases: ['ena', '絵名', '绘名', '东云绘名'],
  },
  {
    id: 20,
    firstName: '暁山',
    givenName: '瑞希',
    fullName: '暁山瑞希',
    enName: 'Mizuki Akiyama',
    unit: 'school_refusal',
    color: '#DD8899',
    aliases: ['mizuki', '瑞希', '晓山瑞希'],
  },

  // ── VIRTUAL SINGER ──
  {
    id: 21,
    firstName: '初音',
    givenName: 'ミク',
    fullName: '初音ミク',
    enName: 'Hatsune Miku',
    unit: 'piapro',
    color: '#33CCBB',
    aliases: ['miku', '初音未来', '初音', 'ミク', '39'],
  },
  {
    id: 22,
    firstName: '鏡音',
    givenName: 'リン',
    fullName: '鏡音リン',
    enName: 'Kagamine Rin',
    unit: 'piapro',
    color: '#FFCC11',
    aliases: ['rin', '镜音铃', '镜音リン', '铃'],
  },
  {
    id: 23,
    firstName: '鏡音',
    givenName: 'レン',
    fullName: '鏡音レン',
    enName: 'Kagamine Len',
    unit: 'piapro',
    color: '#FFEE11',
    aliases: ['len', '镜音连', '镜音レン', '连'],
  },
  {
    id: 24,
    firstName: '巡音',
    givenName: 'ルカ',
    fullName: '巡音ルカ',
    enName: 'Megurine Luka',
    unit: 'piapro',
    color: '#FFAACC',
    aliases: ['luka', '巡音流歌', '巡音ルカ', '流歌'],
  },
  {
    id: 25,
    firstName: '',
    givenName: 'MEIKO',
    fullName: 'MEIKO',
    enName: 'MEIKO',
    unit: 'piapro',
    color: '#DD4444',
    aliases: ['meiko', 'メイコ'],
  },
  {
    id: 26,
    firstName: '',
    givenName: 'KAITO',
    fullName: 'KAITO',
    enName: 'KAITO',
    unit: 'piapro',
    color: '#3366CC',
    aliases: ['kaito', 'カイト'],
  },
]

/**
 * Lookup a character by ID.
 */
export function getCharacterById(id: number): CharacterMeta | undefined {
  return CHARACTERS.find((c) => c.id === id)
}

/**
 * Search characters by alias (case-insensitive).
 */
export function findCharacterByAlias(alias: string): CharacterMeta | undefined {
  const lower = alias.toLowerCase()
  return CHARACTERS.find(
    (c) =>
      c.enName.toLowerCase().includes(lower) ||
      c.fullName.includes(alias) ||
      c.givenName.includes(alias) ||
      c.aliases.some((a) => a.toLowerCase() === lower),
  )
}

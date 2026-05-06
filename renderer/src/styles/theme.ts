// Light design tokens for Satori-rendered Moebot cards.
// Keep this palette bright: Console preview and bot images should not use dark surfaces.
export const theme = {
  colors: {
    background: '#f7fbff',
    backgroundSoft: '#eef9fb',
    surface: '#ffffff',
    surfaceLight: '#f2f7fb',
    surfaceMuted: '#e8eef5',
    surfaceAccent: '#e8fbf8',
    text: '#172033',
    textSecondary: '#526277',
    textMuted: '#8a98aa',
    accent: '#33ccbb',
    accentLight: '#73e1d4',
    accentSoft: '#dff8f5',
    success: '#21b37b',
    warning: '#ffb23f',
    error: '#ff5d7a',
    border: '#dce8f2',
    borderStrong: '#b8d7e6',

    // Card attribute colors
    cute: '#FF6699',
    cool: '#1C8CFF',
    pure: '#35C86A',
    happy: '#FFB000',
    mysterious: '#A863E8',

    // Unit colors
    vs: '#33CCBB',
    ln: '#4455DD',
    mmj: '#88DD44',
    vbs: '#EE1166',
    wxs: '#FF9900',
    n25: '#884499',
  },

  spacing: {
    xs: 4,
    sm: 8,
    md: 16,
    lg: 24,
    xl: 32,
    xxl: 48,
  },

  fontSize: {
    xs: 12,
    sm: 14,
    md: 16,
    lg: 20,
    xl: 24,
    xxl: 32,
    title: 40,
  },

  borderRadius: {
    sm: 6,
    md: 10,
    lg: 14,
    xl: 22,
    round: 9999,
  },

  cardWidth: 800,
  cardPadding: 32,
} as const

export type ThemeColorName = keyof typeof theme.colors

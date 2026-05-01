// Design tokens for card templates
export const theme = {
  colors: {
    background: '#1a1b2e',
    surface: '#252641',
    surfaceLight: '#2f3050',
    text: '#ffffff',
    textSecondary: '#a0a0b8',
    textMuted: '#6b6b80',
    accent: '#7c5cfc',
    accentLight: '#9b85fc',
    success: '#44cc88',
    warning: '#ffaa44',
    error: '#ff4466',
    border: '#3a3b55',

    // Card attribute colors
    cute: '#FF6699',
    cool: '#0077DD',
    pure: '#00BB33',
    happy: '#FFAA00',
    mysterious: '#BB44DD',

    // Unit colors
    vs: '#00BBDD',
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
    sm: 4,
    md: 8,
    lg: 12,
    xl: 16,
    round: 9999,
  },

  cardWidth: 800,
  cardPadding: 32,
} as const

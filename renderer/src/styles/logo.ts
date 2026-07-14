import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'

let cachedLogoSvg: string | null = null
let cachedStarMoeSvg: string | null = null

const fallbackLogoSvg = `<?xml version="1.0" encoding="UTF-8"?>
<svg width="512" height="120" viewBox="0 0 512 120" xmlns="http://www.w3.org/2000/svg">
  <rect width="512" height="120" rx="36" fill="var(--theme-color,#33ccbb)" opacity="0.12"/>
  <text x="50%" y="54%" dominant-baseline="middle" text-anchor="middle" font-family="Arial,sans-serif" font-size="52" font-weight="800" fill="var(--theme-color,#33ccbb)">Moebot</text>
</svg>`

const fallbackStarMoeSvg = `<?xml version="1.0" encoding="UTF-8"?>
<svg width="512" height="120" viewBox="0 0 512 120" xmlns="http://www.w3.org/2000/svg">
  <rect width="512" height="120" rx="36" fill="var(--theme-color,#39C5BB)" opacity="0.12"/>
  <text x="50%" y="54%" dominant-baseline="middle" text-anchor="middle" font-family="Arial,sans-serif" font-size="52" font-weight="800" fill="var(--theme-color,#39C5BB)">StarMoe</text>
</svg>`

function loadLogoSvg(): string {
  if (cachedLogoSvg) return cachedLogoSvg

  const candidates = [
    resolve(process.cwd(), 'assets/moebot.svg'),
    resolve(process.cwd(), '..', 'assets/moebot.svg'),
    resolve(process.cwd(), '..', '..', 'assets/moebot.svg'),
    resolve(__dirname, '../../../../assets/moebot.svg'),
    resolve(__dirname, '../../../../../assets/moebot.svg'),
  ]

  const logoPath = candidates.find(path => existsSync(path))
  cachedLogoSvg = logoPath ? readFileSync(logoPath, 'utf8') : fallbackLogoSvg
  return cachedLogoSvg
}

function loadStarMoeSvg(): string {
  if (cachedStarMoeSvg) return cachedStarMoeSvg

  const candidates = [
    resolve(process.cwd(), 'assets/starmoe.svg'),
    resolve(process.cwd(), '..', 'assets/starmoe.svg'),
    resolve(process.cwd(), '..', '..', 'assets/starmoe.svg'),
    resolve(__dirname, '../../../../assets/starmoe.svg'),
    resolve(__dirname, '../../../../../assets/starmoe.svg'),
  ]

  const logoPath = candidates.find(path => existsSync(path))
  cachedStarMoeSvg = logoPath ? readFileSync(logoPath, 'utf8') : fallbackStarMoeSvg
  return cachedStarMoeSvg
}

function sanitizeCssColor(color: string, defaultColor = '#33ccbb'): string {
  return color.replace(/[^#%,.()\w\s-]/g, '').trim() || defaultColor
}

export function getMoebotLogoDataUri(themeColor = '#33ccbb'): string {
  const color = sanitizeCssColor(themeColor, '#33ccbb')
  const svg = loadLogoSvg()
  const styledSvg = svg.includes('<svg ')
    ? svg.replace('<svg ', `<svg style="--theme-color: ${color}; color: ${color};" `)
    : svg
  const resolvedSvg = styledSvg.replace(/var\(--theme-color,\s*#[0-9a-fA-F]{3,8}\)/g, color)

  return `data:image/svg+xml;base64,${Buffer.from(resolvedSvg, 'utf8').toString('base64')}`
}

export function getStarMoeLogoDataUri(themeColor = '#39C5BB'): string {
  const color = sanitizeCssColor(themeColor, '#39C5BB')
  const svg = loadStarMoeSvg()
  const styledSvg = svg.includes('<svg ')
    ? svg.replace('<svg ', `<svg style="--theme-color: ${color}; color: ${color};" `)
    : svg
  const resolvedSvg = styledSvg.replace(/var\(--theme-color,\s*#[0-9a-fA-F]{3,8}\)/g, color)

  return `data:image/svg+xml;base64,${Buffer.from(resolvedSvg, 'utf8').toString('base64')}`
}

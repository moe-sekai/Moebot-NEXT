/**
 * Moebot NEXT — Asset Sync Script
 *
 * Syncs static assets from the Snowy Viewer repository.
 * 
 * Usage: npx tsx scripts/sync-assets.ts [--viewer-path <path>]
 * 
 * Default viewer path: ../Snowy_Viewer
 */

import { cpSync, existsSync, mkdirSync } from 'fs'
import { join, resolve } from 'path'

const args = process.argv.slice(2)
const viewerPathIdx = args.indexOf('--viewer-path')
const VIEWER_PATH = viewerPathIdx >= 0 
  ? resolve(args[viewerPathIdx + 1])
  : resolve(process.cwd(), '..', 'Snowy_Viewer')

const ASSETS_DIR = resolve(process.cwd(), 'assets')

interface SyncTask {
  src: string
  dest: string
  description: string
}

const SYNC_TASKS: SyncTask[] = [
  {
    src: 'web/public/sticker-maker',
    dest: 'stickers/sticker-maker',
    description: 'Sticker maker assets',
  },
  {
    src: 'refer/sekai-stickers',
    dest: 'stickers/sekai-stickers',
    description: 'SEKAI sticker images',
  },
]

async function main() {
  console.log('Moebot NEXT — Asset Sync')
  console.log('========================')
  console.log(`Viewer path: ${VIEWER_PATH}`)
  console.log(`Assets dir:  ${ASSETS_DIR}`)
  console.log('')

  if (!existsSync(VIEWER_PATH)) {
    console.error(`[ERROR] Snowy Viewer not found at: ${VIEWER_PATH}`)
    console.error('Use --viewer-path to specify the correct path')
    process.exit(1)
  }

  for (const task of SYNC_TASKS) {
    const srcPath = join(VIEWER_PATH, task.src)
    const destPath = join(ASSETS_DIR, task.dest)

    if (!existsSync(srcPath)) {
      console.log(`[SKIP] ${task.description} — source not found: ${srcPath}`)
      continue
    }

    console.log(`[SYNC] ${task.description}`)
    console.log(`       ${srcPath} → ${destPath}`)
    
    mkdirSync(destPath, { recursive: true })
    cpSync(srcPath, destPath, { recursive: true, force: true })
    console.log(`       ✓ Done`)
  }

  console.log('')
  console.log('Asset sync complete!')
}

main().catch(console.error)

import React, { cloneElement, isValidElement } from 'react'
import { createImageHydrationStats, rendererAssetCache, type ImageHydrationStats } from './asset-cache'

export interface ImageHydrationResult {
  element: React.ReactNode
  stats: ImageHydrationStats
  ms: number
}

export async function hydrateCachedImages(element: React.ReactNode): Promise<ImageHydrationResult> {
  const startedAt = performance.now()
  const stats = createImageHydrationStats()
  const hydrated = await hydrateNode(element, stats)
  return {
    element: hydrated,
    stats,
    ms: Math.round(performance.now() - startedAt),
  }
}

async function hydrateNode(node: React.ReactNode, stats: ImageHydrationStats): Promise<React.ReactNode> {
  if (node === null || node === undefined || typeof node === 'boolean') return node
  if (typeof node === 'string' || typeof node === 'number' || typeof node === 'bigint') return node
  if (Array.isArray(node)) return Promise.all(node.map((child) => hydrateNode(child, stats)))
  if (!isValidElement(node)) return node

  const props = node.props as Record<string, unknown>

  if (typeof node.type === 'function') {
    const rendered = (node.type as (props: Record<string, unknown>) => React.ReactNode)(props)
    return hydrateNode(rendered, stats)
  }

  const nextProps: Record<string, unknown> = {}
  let changed = false

  if (typeof node.type === 'string' && node.type.toLowerCase() === 'img') {
    stats.total += 1
    const src = props.src
    if (rendererAssetCache.isRemoteUrl(src)) {
      stats.remote += 1
      const cached = await rendererAssetCache.getDataUri(src)
      if (cached.hit && cached.dataUri) {
        nextProps.src = cached.dataUri
        stats.hits += 1
        changed = true
      } else if (cached.skipped) {
        stats.skipped += 1
      } else {
        stats.misses += 1
        if (cached.error) stats.errors += 1
      }
    } else {
      stats.skipped += 1
    }
  }

  if ('children' in props) {
    const children = props.children as React.ReactNode
    const hydratedChildren = await hydrateNode(children, stats)
    if (hydratedChildren !== children) {
      nextProps.children = hydratedChildren
      changed = true
    }
  }

  return changed ? cloneElement(node, nextProps) : node
}

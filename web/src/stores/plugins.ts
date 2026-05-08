import { defineStore } from 'pinia'
import { listPlugins, type PluginListItem } from '../api/client'

interface State {
  items: PluginListItem[]
  loaded: boolean
  loading: boolean
  error: string
  lastFetch: number
}

const STALE_MS = 15_000

export const usePluginsStore = defineStore('plugins', {
  state: (): State => ({
    items: [],
    loaded: false,
    loading: false,
    error: '',
    lastFetch: 0,
  }),
  getters: {
    isLoaded: state => (name: string) =>
      state.items.some(p => p.name === name && p.loaded),
    isEnabled: state => (name: string) =>
      state.items.some(p => p.name === name && p.enabled),
    byName: state => (name: string) => state.items.find(p => p.name === name),
  },
  actions: {
    async fetch(force = false) {
      if (!force && this.loaded && Date.now() - this.lastFetch < STALE_MS) return
      this.loading = true
      this.error = ''
      try {
        this.items = await listPlugins()
        this.loaded = true
        this.lastFetch = Date.now()
      } catch (err) {
        this.error = err instanceof Error ? err.message : '加载插件失败'
      } finally {
        this.loading = false
      }
    },
    async refresh() {
      return this.fetch(true)
    },
  },
})

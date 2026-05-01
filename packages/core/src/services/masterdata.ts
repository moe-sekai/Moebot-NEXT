import { Context } from 'koishi'
import { MoebotConfig } from '../config'
import { fetchAllMasterdata, MasterdataStore } from '@moebot/shared'

export class MasterdataService {
  private store: MasterdataStore | null = null
  private refreshTimer: NodeJS.Timeout | null = null
  private loading = false

  constructor(
    private ctx: Context,
    private config: MoebotConfig,
  ) {}

  async init(): Promise<void> {
    await this.refresh()

    // Set up auto-refresh
    if (this.config.masterDataRefreshInterval > 0) {
      this.refreshTimer = setInterval(
        () => this.refresh(),
        this.config.masterDataRefreshInterval,
      )
      this.ctx.on('dispose', () => {
        if (this.refreshTimer) clearInterval(this.refreshTimer)
      })
    }
  }

  async refresh(): Promise<void> {
    if (this.loading) return
    this.loading = true

    try {
      this.ctx.logger('moebot').info('Refreshing masterdata...')
      this.store = await fetchAllMasterdata(this.config.masterDataUrl)
      this.ctx.logger('moebot').info('Masterdata refreshed successfully')
    } catch (err) {
      this.ctx.logger('moebot').error('Failed to refresh masterdata:', err)
    } finally {
      this.loading = false
    }
  }

  get data(): MasterdataStore {
    if (!this.store) {
      throw new Error('Masterdata not loaded yet')
    }
    return this.store
  }

  get isReady(): boolean {
    return this.store !== null
  }

  // Convenience getters
  get cards() { return this.data.cards }
  get musics() { return this.data.musics }
  get events() { return this.data.events }
  get gachas() { return this.data.gachas }
  get gameCharacters() { return this.data.gameCharacters }
}

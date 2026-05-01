import { Context } from 'koishi'

// Koishi uses its own database abstraction.
// We declare the tables using Koishi's model system.

declare module 'koishi' {
  interface Tables {
    'moebot.users': MoebotUser
    'moebot.groups': MoebotGroup
    'moebot.stats': MoebotStat
  }
}

export interface MoebotUser {
  id: number
  platform: string
  platformId: string
  gameId: string | null
  nickname: string | null
  region: string
  createdAt: Date
  updatedAt: Date
}

export interface MoebotGroup {
  id: number
  platform: string
  groupId: string
  name: string | null
  enabled: boolean
  config: string  // JSON string
  createdAt: Date
}

export interface MoebotStat {
  id: number
  command: string
  platform: string
  userId: string | null
  groupId: string | null
  args: string | null
  responseMs: number
  createdAt: Date
}

export class DatabaseService {
  constructor(private ctx: Context) {}

  async init(): Promise<void> {
    // Extend Koishi database with our tables
    this.ctx.model.extend('moebot.users', {
      id: 'unsigned',
      platform: 'string',
      platformId: 'string',
      gameId: { type: 'string', nullable: true },
      nickname: { type: 'string', nullable: true },
      region: { type: 'string', initial: 'jp' },
      createdAt: 'timestamp',
      updatedAt: 'timestamp',
    }, {
      autoInc: true,
      unique: [['platform', 'platformId']],
    })

    this.ctx.model.extend('moebot.groups', {
      id: 'unsigned',
      platform: 'string',
      groupId: 'string',
      name: { type: 'string', nullable: true },
      enabled: { type: 'boolean', initial: true },
      config: { type: 'string', initial: '{}' },
      createdAt: 'timestamp',
    }, {
      autoInc: true,
      unique: [['platform', 'groupId']],
    })

    this.ctx.model.extend('moebot.stats', {
      id: 'unsigned',
      command: 'string',
      platform: 'string',
      userId: { type: 'string', nullable: true },
      groupId: { type: 'string', nullable: true },
      args: { type: 'string', nullable: true },
      responseMs: { type: 'unsigned', initial: 0 },
      createdAt: 'timestamp',
    }, {
      autoInc: true,
    })

    this.ctx.logger('moebot').info('Database tables initialized')
  }

  // User operations
  async findUser(platform: string, platformId: string): Promise<MoebotUser | null> {
    const [user] = await this.ctx.database.get('moebot.users', { platform, platformId })
    return user ?? null
  }

  async bindUser(platform: string, platformId: string, gameId: string, region: string = 'jp'): Promise<MoebotUser> {
    const existing = await this.findUser(platform, platformId)
    if (existing) {
      await this.ctx.database.set('moebot.users', existing.id, {
        gameId,
        region,
        updatedAt: new Date(),
      })
      return { ...existing, gameId, region, updatedAt: new Date() }
    }

    const user = await this.ctx.database.create('moebot.users', {
      platform,
      platformId,
      gameId,
      region,
      createdAt: new Date(),
      updatedAt: new Date(),
    })
    return user
  }

  async unbindUser(platform: string, platformId: string): Promise<boolean> {
    const user = await this.findUser(platform, platformId)
    if (!user) return false
    await this.ctx.database.set('moebot.users', user.id, {
      gameId: null,
      updatedAt: new Date(),
    })
    return true
  }

  // Group operations
  async getGroupConfig(platform: string, groupId: string): Promise<MoebotGroup | null> {
    const [group] = await this.ctx.database.get('moebot.groups', { platform, groupId })
    return group ?? null
  }

  async setGroupConfig(platform: string, groupId: string, config: Record<string, any>): Promise<void> {
    const existing = await this.getGroupConfig(platform, groupId)
    if (existing) {
      await this.ctx.database.set('moebot.groups', existing.id, {
        config: JSON.stringify(config),
      })
    } else {
      await this.ctx.database.create('moebot.groups', {
        platform,
        groupId,
        config: JSON.stringify(config),
        createdAt: new Date(),
      })
    }
  }

  // Stats
  async logCommand(
    command: string,
    platform: string,
    userId?: string,
    groupId?: string,
    args?: string,
    responseMs?: number,
  ): Promise<void> {
    await this.ctx.database.create('moebot.stats', {
      command,
      platform,
      userId: userId ?? null,
      groupId: groupId ?? null,
      args: args ?? null,
      responseMs: responseMs ?? 0,
      createdAt: new Date(),
    })
  }
}

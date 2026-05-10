<template>
  <main class="page-stack">
    <PageHeader eyebrow="Command Stats" title="调用统计" subtitle="查看指令调用次数、响应耗时、活跃群组与用户分布。">
      <template #actions>
        <select v-model.number="days" class="stats-select" @change="loadAll">
          <option :value="1">最近 1 天</option>
          <option :value="7">最近 7 天</option>
          <option :value="14">最近 14 天</option>
          <option :value="30">最近 30 天</option>
          <option :value="90">最近 90 天</option>
        </select>
        <select v-model.number="topN" class="stats-select" @change="loadAll">
          <option :value="5">Top 5</option>
          <option :value="10">Top 10</option>
          <option :value="20">Top 20</option>
          <option :value="50">Top 50</option>
        </select>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadAll">刷新</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="统计加载失败">{{ error }}</UiAlert>

    <div v-if="loading && !stats" class="status-grid">
      <UiSkeleton v-for="item in 4" :key="item" height="108px" />
    </div>
    <div v-else class="status-grid status-grid--metrics">
      <MetricCard label="调用总数" :value="formatNumber(stats?.totals.calls ?? 0)" :hint="windowHint" icon="stats" />
      <MetricCard label="活跃用户" :value="formatNumber(stats?.totals.users ?? 0)" hint="去重后的用户数" icon="users" />
      <MetricCard label="活跃群组" :value="formatNumber(stats?.totals.groups ?? 0)" hint="含私聊（empty group）" icon="groups" />
      <MetricCard label="平均响应" :value="`${formatMs(stats?.totals.avg_ms ?? 0)} ms`" hint="所有指令端到端耗时" icon="clock" />
    </div>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>每日调用趋势</h2>
          <p>过去 {{ days }} 天每日调用次数与平均耗时。</p>
        </div>
      </div>
      <div v-if="loading && !stats" class="table-skeleton">
        <UiSkeleton height="180px" />
      </div>
      <div v-else-if="!stats || stats.trend.length === 0" class="empty-state compact">
        <div class="empty-state__icon"><SvgIcon name="stats" :size="22" /></div>
        <p>暂无调用数据。</p>
      </div>
      <div v-else class="trend-chart" :style="{ '--cols': stats.trend.length }">
        <div
          v-for="point in stats.trend"
          :key="point.date"
          class="trend-bar"
          :title="`${point.date} · ${point.count} 次 · 平均 ${formatMs(point.avg_ms)} ms`"
        >
          <div class="trend-bar__fill" :style="{ height: barHeight(point.count) }">
            <span class="trend-bar__value">{{ point.count }}</span>
          </div>
          <div class="trend-bar__label">{{ shortDate(point.date) }}</div>
        </div>
      </div>
    </UiCard>

    <div class="dashboard-grid dashboard-grid--main stats-grid">
      <UiCard>
        <div class="card-heading">
          <div>
            <h2>指令排行</h2>
            <p>调用次数最多的指令 · Top {{ topN }}</p>
          </div>
        </div>
        <div v-if="!stats || stats.data.length === 0" class="empty-state compact">
          <div class="empty-state__icon"><SvgIcon name="command" :size="22" /></div>
          <p>窗口内暂无指令调用。</p>
        </div>
        <div v-else class="table-wrap">
          <table class="ui-table">
            <thead>
              <tr>
                <th style="width: 32px">#</th>
                <th>指令</th>
                <th class="text-right">次数</th>
                <th class="text-right">平均耗时</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, idx) in topCommands" :key="row.command">
                <td class="muted-text">{{ idx + 1 }}</td>
                <td class="font-medium">{{ row.command }}</td>
                <td class="text-right">{{ formatNumber(row.count) }}</td>
                <td class="text-right">{{ formatMs(row.avg_ms) }} ms</td>
              </tr>
            </tbody>
          </table>
        </div>
      </UiCard>

      <UiCard>
        <div class="card-heading">
          <div>
            <h2>群组排行</h2>
            <p>调用最频繁的群组 · Top {{ topN }}</p>
          </div>
        </div>
        <div v-if="!stats || stats.by_group.length === 0" class="empty-state compact">
          <div class="empty-state__icon"><SvgIcon name="groups" :size="22" /></div>
          <p>窗口内暂无群组调用。</p>
        </div>
        <div v-else class="rank-list">
          <div v-for="(row, idx) in stats.by_group" :key="row.group_id" class="rank-row">
            <span class="rank-row__idx">{{ idx + 1 }}</span>
            <span class="rank-row__id" :title="row.group_id">{{ row.group_id === 'private' ? '私聊' : row.group_id }}</span>
            <div class="rank-row__bar">
              <div class="rank-row__fill" :style="{ width: pct(row.count, maxGroupCount) + '%' }"></div>
            </div>
            <span class="rank-row__count">{{ formatNumber(row.count) }}</span>
          </div>
        </div>
      </UiCard>

      <UiCard>
        <div class="card-heading">
          <div>
            <h2>用户排行</h2>
            <p>调用最频繁的用户 · Top {{ topN }}</p>
          </div>
        </div>
        <div v-if="!stats || stats.by_user.length === 0" class="empty-state compact">
          <div class="empty-state__icon"><SvgIcon name="users" :size="22" /></div>
          <p>窗口内暂无用户调用。</p>
        </div>
        <div v-else class="rank-list">
          <div v-for="(row, idx) in stats.by_user" :key="row.user_id" class="rank-row">
            <span class="rank-row__idx">{{ idx + 1 }}</span>
            <span class="rank-row__id" :title="row.user_id">{{ row.user_id }}</span>
            <div class="rank-row__bar">
              <div class="rank-row__fill" :style="{ width: pct(row.count, maxUserCount) + '%' }"></div>
            </div>
            <span class="rank-row__count">{{ formatNumber(row.count) }}</span>
          </div>
        </div>
      </UiCard>

      <UiCard>
        <div class="card-heading">
          <div>
            <h2>平台分布</h2>
            <p>不同平台来源的调用占比。</p>
          </div>
        </div>
        <div v-if="!stats || stats.by_platform.length === 0" class="empty-state compact">
          <div class="empty-state__icon"><SvgIcon name="bot" :size="22" /></div>
          <p>暂无平台数据。</p>
        </div>
        <div v-else class="rank-list">
          <div v-for="row in stats.by_platform" :key="row.platform" class="rank-row">
            <span class="rank-row__id">{{ row.platform }}</span>
            <div class="rank-row__bar">
              <div class="rank-row__fill" :style="{ width: pct(row.count, totalPlatformCount) + '%' }"></div>
            </div>
            <span class="rank-row__count">{{ formatNumber(row.count) }}（{{ pct(row.count, totalPlatformCount).toFixed(1) }}%）</span>
          </div>
        </div>
      </UiCard>
    </div>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>最近调用</h2>
          <p>最近的指令调用流水（仅展示最近 {{ recentLimit }} 条，用于排查与审计）。</p>
        </div>
        <select v-model.number="recentLimit" class="stats-select" @change="loadRecent">
          <option :value="20">20 条</option>
          <option :value="50">50 条</option>
        </select>
      </div>
      <div v-if="recentLoading" class="table-skeleton">
        <UiSkeleton v-for="item in 6" :key="item" height="32px" />
      </div>
      <div v-else-if="recent.length === 0" class="empty-state compact">
        <div class="empty-state__icon"><SvgIcon name="logs" :size="22" /></div>
        <p>暂无调用记录。</p>
      </div>
      <div v-else class="table-wrap">
        <table class="ui-table">
          <thead>
            <tr>
              <th>时间</th>
              <th>指令</th>
              <th>平台</th>
              <th>群组</th>
              <th>用户</th>
              <th>参数</th>
              <th class="text-right">耗时</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in recent" :key="row.id">
              <td class="muted-text" :title="row.created_at">{{ formatTime(row.created_at) }}</td>
              <td class="font-medium">{{ row.command }}</td>
              <td>{{ row.platform || '-' }}</td>
              <td>{{ row.group_id || '私聊' }}</td>
              <td>{{ row.user_id || '-' }}</td>
              <td class="muted-text args-cell" :title="row.args">{{ row.args || '-' }}</td>
              <td class="text-right">{{ row.response_ms }} ms</td>
            </tr>
          </tbody>
        </table>
      </div>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { getCommandStats, getRecentCommands } from '../api/client'
import type { CommandStatsResponse, RecentCommand } from '../api/types'
import MetricCard from '../components/MetricCard.vue'
import PageHeader from '../components/PageHeader.vue'
import SvgIcon from '../components/icons/SvgIcon.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

const days = ref(7)
const topN = ref(10)
const recentLimit = ref(20)

const stats = ref<CommandStatsResponse | null>(null)
const recent = ref<RecentCommand[]>([])
const loading = ref(false)
const recentLoading = ref(false)
const error = ref('')

const windowHint = computed(() => `过去 ${days.value} 天`)

const topCommands = computed(() => {
  if (!stats.value) return []
  return stats.value.data.slice(0, topN.value)
})

const maxTrendCount = computed(() => {
  if (!stats.value || stats.value.trend.length === 0) return 1
  return Math.max(1, ...stats.value.trend.map((p) => p.count))
})

const maxGroupCount = computed(() => {
  if (!stats.value || stats.value.by_group.length === 0) return 1
  return Math.max(1, ...stats.value.by_group.map((p) => p.count))
})

const maxUserCount = computed(() => {
  if (!stats.value || stats.value.by_user.length === 0) return 1
  return Math.max(1, ...stats.value.by_user.map((p) => p.count))
})

const totalPlatformCount = computed(() => {
  if (!stats.value) return 1
  const sum = stats.value.by_platform.reduce((acc, p) => acc + p.count, 0)
  return sum > 0 ? sum : 1
})

watch(recentLimit, () => loadRecent())

onMounted(loadAll)

async function loadAll() {
  await Promise.all([loadStats(), loadRecent()])
}

async function loadStats() {
  loading.value = true
  error.value = ''
  try {
    stats.value = await getCommandStats(days.value, topN.value)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载统计失败。'
  } finally {
    loading.value = false
  }
}

async function loadRecent() {
  recentLoading.value = true
  try {
    const result = await getRecentCommands(recentLimit.value)
    recent.value = result.data ?? []
  } catch (err) {
    if (!error.value) {
      error.value = err instanceof Error ? err.message : '加载最近调用失败。'
    }
  } finally {
    recentLoading.value = false
  }
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat('zh-CN').format(value)
}

function formatMs(value: number): string {
  if (!Number.isFinite(value)) return '0'
  if (value >= 1000) return value.toFixed(0)
  if (value >= 100) return value.toFixed(1)
  return value.toFixed(2)
}

function pct(value: number, total: number): number {
  if (total <= 0) return 0
  return (value / total) * 100
}

function barHeight(count: number): string {
  const ratio = count / maxTrendCount.value
  const pctValue = Math.max(ratio * 100, count > 0 ? 4 : 0)
  return `${pctValue}%`
}

function shortDate(date: string): string {
  const m = /^\d{4}-(\d{2})-(\d{2})$/.exec(date)
  if (!m) return date
  return `${m[1]}/${m[2]}`
}

function formatTime(value: string): string {
  if (!value) return '-'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  return d.toLocaleString('zh-CN', { hour12: false })
}
</script>

<style scoped>
.stats-select {
  border-radius: 12px;
  border: 1px solid rgba(165, 180, 252, 0.5);
  background: rgba(255, 255, 255, 0.86);
  padding: 6px 10px;
  font-size: 13px;
  color: inherit;
}
.stats-select:focus {
  outline: none;
  border-color: rgba(99, 102, 241, 0.85);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.18);
}
.status-grid--metrics {
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
}
.stats-grid {
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
}
.text-right { text-align: right; }

.trend-chart {
  display: grid;
  grid-template-columns: repeat(var(--cols, 7), minmax(0, 1fr));
  gap: 6px;
  align-items: end;
  height: 200px;
  padding: 12px 4px 0;
  border-radius: 14px;
  background: rgba(99, 102, 241, 0.06);
}
.trend-bar {
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%;
  min-width: 0;
}
.trend-bar__fill {
  width: 80%;
  max-width: 36px;
  margin-top: auto;
  background: linear-gradient(180deg, rgba(99, 102, 241, 0.85), rgba(129, 140, 248, 0.55));
  border-radius: 6px 6px 2px 2px;
  position: relative;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  min-height: 0;
  transition: filter .15s;
}
.trend-bar__fill:hover { filter: brightness(1.05); }
.trend-bar__value {
  position: absolute;
  top: -18px;
  font-size: 11px;
  font-weight: 600;
  color: var(--muted-foreground);
  white-space: nowrap;
}
.trend-bar__label {
  margin-top: 6px;
  font-size: 11px;
  color: var(--muted-foreground);
  white-space: nowrap;
}

.rank-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.rank-row {
  display: grid;
  grid-template-columns: 22px minmax(80px, 1fr) minmax(0, 2fr) auto;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}
.rank-row__idx {
  color: var(--muted-foreground);
  font-weight: 700;
  font-size: 12px;
}
.rank-row__id {
  font-family: ui-monospace, "SFMono-Regular", "JetBrains Mono", Menlo, Consolas, monospace;
  font-size: 12.5px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.rank-row__bar {
  height: 8px;
  background: rgba(99, 102, 241, 0.1);
  border-radius: 999px;
  overflow: hidden;
}
.rank-row__fill {
  height: 100%;
  background: linear-gradient(90deg, rgba(99, 102, 241, 0.85), rgba(129, 140, 248, 0.55));
  border-radius: 999px;
  transition: width .25s ease;
}
.rank-row__count {
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  color: var(--muted-foreground);
  white-space: nowrap;
}

.args-cell {
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: ui-monospace, "SFMono-Regular", "JetBrains Mono", Menlo, Consolas, monospace;
  font-size: 12.5px;
}
</style>

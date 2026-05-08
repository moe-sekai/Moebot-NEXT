<template>
  <main class="page-stack">
    <PageHeader eyebrow="Command Stats" title="指令统计" :subtitle="`查看最近 ${days} 天的指令调用、耗时与平台分布。`">
      <template #actions>
        <div class="stats-range">
          <button
            v-for="option in RANGE_OPTIONS"
            :key="option.value"
            type="button"
            class="stats-range__btn"
            :class="{ 'stats-range__btn--active': days === option.value }"
            @click="setDays(option.value)"
          >{{ option.label }}</button>
        </div>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadStats">刷新统计</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="统计加载失败">{{ error }}</UiAlert>

    <section class="stats-metrics">
      <MetricCard label="调用总数" :value="totals.calls.toLocaleString()" :hint="`自 ${sinceLabel}`" icon="command" />
      <MetricCard label="活跃用户" :value="totals.users.toLocaleString()" :hint="userHint" icon="users" />
      <MetricCard label="活跃群组" :value="totals.groups.toLocaleString()" :hint="groupHint" icon="groups" />
      <MetricCard label="平均耗时" :value="`${Math.round(totals.avg_ms)} ms`" :hint="avgHint" icon="clock" />
    </section>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>每日趋势</h2>
          <p>统计起点：{{ sinceLabel }}</p>
        </div>
        <UiBadge variant="secondary">{{ trend.length }} 个数据点</UiBadge>
      </div>
      <div v-if="loading && trend.length === 0" class="table-skeleton">
        <UiSkeleton height="180px" />
      </div>
      <div v-else-if="trend.length === 0" class="empty-state compact">
        <div class="empty-state__icon"><SvgIcon name="stats" :size="22" /></div>
        <p>所选时间段尚无调用数据。</p>
      </div>
      <div v-else class="stats-trend">
        <svg :viewBox="trendViewBox" preserveAspectRatio="none" class="stats-trend__svg">
          <line v-for="(line, idx) in trendGridLines" :key="`grid-${idx}`" :x1="0" :x2="trendChartWidth" :y1="line" :y2="line" class="stats-trend__grid" />
          <polyline :points="trendPoints" class="stats-trend__line" />
          <circle
            v-for="(point, idx) in trendCoords"
            :key="`pt-${idx}`"
            :cx="point.x"
            :cy="point.y"
            r="3"
            class="stats-trend__dot"
          >
            <title>{{ point.label }}</title>
          </circle>
        </svg>
        <div class="stats-trend__axis">
          <span v-for="(point, idx) in trendAxisLabels" :key="`axis-${idx}`">{{ point }}</span>
        </div>
      </div>
    </UiCard>

    <div class="dashboard-grid dashboard-grid--main">
      <UiCard>
        <div class="card-heading">
          <div>
            <h2>Top 指令</h2>
            <p>按调用次数倒序排列。</p>
          </div>
          <UiBadge variant="secondary">{{ rows.length }} 条指令</UiBadge>
        </div>
        <div v-if="loading && rows.length === 0" class="table-skeleton">
          <UiSkeleton v-for="item in 6" :key="item" height="42px" />
        </div>
        <div v-else-if="rows.length === 0" class="empty-state compact">
          <div class="empty-state__icon"><SvgIcon name="stats" :size="22" /></div>
          <p>暂无统计记录，命令被调用后这里会自动显示。</p>
        </div>
        <div v-else class="table-wrap">
          <table class="ui-table">
            <thead>
              <tr>
                <th style="width:42px">#</th>
                <th>指令</th>
                <th>调用次数</th>
                <th>平均耗时</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, idx) in rows" :key="row.command">
                <td>{{ idx + 1 }}</td>
                <td class="font-medium">{{ row.command }}</td>
                <td>{{ row.count.toLocaleString() }}</td>
                <td>{{ Math.round(row.avg_ms) }} ms</td>
              </tr>
            </tbody>
          </table>
        </div>
      </UiCard>

      <UiCard>
        <div class="card-heading">
          <div>
            <h2>平台分布</h2>
            <p>不同平台的指令调用比例。</p>
          </div>
          <UiBadge variant="secondary">{{ byPlatform.length }} 个平台</UiBadge>
        </div>
        <div v-if="byPlatform.length === 0" class="empty-state compact">
          <div class="empty-state__icon"><SvgIcon name="bot" :size="22" /></div>
          <p>暂无平台调用数据。</p>
        </div>
        <ul v-else class="stats-platform">
          <li v-for="row in byPlatform" :key="row.platform">
            <div class="stats-platform__row">
              <span class="font-medium">{{ row.platform }}</span>
              <span class="muted-text">{{ row.count.toLocaleString() }} · {{ percentLabel(row.count) }}</span>
            </div>
            <div class="stats-platform__bar">
              <div class="stats-platform__fill" :style="{ width: `${percent(row.count)}%` }" />
            </div>
          </li>
        </ul>
      </UiCard>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getCommandStats } from '../api/client'
import type { CommandStatRow, CommandStatsPlatformPoint, CommandStatsTotals, CommandStatsTrendPoint } from '../api/types'
import SvgIcon from '../components/icons/SvgIcon.vue'
import MetricCard from '../components/MetricCard.vue'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

const RANGE_OPTIONS = [
  { value: 1, label: '1 天' },
  { value: 7, label: '7 天' },
  { value: 30, label: '30 天' },
] as const

const rows = ref<CommandStatRow[]>([])
const trend = ref<CommandStatsTrendPoint[]>([])
const byPlatform = ref<CommandStatsPlatformPoint[]>([])
const totals = ref<CommandStatsTotals>({ calls: 0, users: 0, groups: 0, avg_ms: 0 })
const since = ref('')
const days = ref<number>(7)
const loading = ref(false)
const error = ref('')

const trendChartWidth = 600
const trendChartHeight = 160

const sinceLabel = computed(() => since.value ? new Date(since.value).toLocaleString() : '-')

const userHint = computed(() => totals.value.calls > 0 ? `共 ${totals.value.calls.toLocaleString()} 次调用` : '等待数据')
const groupHint = computed(() => byPlatform.value.length > 0 ? `${byPlatform.value.length} 个平台` : '等待数据')
const avgHint = computed(() => totals.value.calls > 0 ? `按 ${days.value} 天均值` : '等待数据')

const trendMaxCount = computed(() => trend.value.reduce((max, point) => Math.max(max, point.count), 0))

const trendCoords = computed(() => {
  if (trend.value.length === 0) return []
  const max = Math.max(trendMaxCount.value, 1)
  const stepX = trend.value.length === 1 ? 0 : trendChartWidth / (trend.value.length - 1)
  return trend.value.map((point, idx) => {
    const x = trend.value.length === 1 ? trendChartWidth / 2 : idx * stepX
    const y = trendChartHeight - (point.count / max) * (trendChartHeight - 12) - 6
    return {
      x,
      y,
      label: `${point.date} · ${point.count} 次 · 平均 ${Math.round(point.avg_ms)} ms`,
    }
  })
})

const trendPoints = computed(() => trendCoords.value.map((p) => `${p.x},${p.y}`).join(' '))
const trendViewBox = `0 0 ${trendChartWidth} ${trendChartHeight}`

const trendGridLines = computed(() => {
  const lines: number[] = []
  for (let i = 1; i < 4; i++) {
    lines.push((trendChartHeight / 4) * i)
  }
  return lines
})

const trendAxisLabels = computed(() => {
  if (trend.value.length === 0) return []
  if (trend.value.length <= 7) return trend.value.map((p) => formatDateShort(p.date))
  const idxs = [0, Math.floor(trend.value.length / 3), Math.floor(trend.value.length * 2 / 3), trend.value.length - 1]
  return idxs.map((idx) => formatDateShort(trend.value[idx].date))
})

const platformTotal = computed(() => byPlatform.value.reduce((sum, row) => sum + row.count, 0))

function percent(value: number): number {
  if (platformTotal.value === 0) return 0
  return (value / platformTotal.value) * 100
}

function percentLabel(value: number): string {
  return `${percent(value).toFixed(1)}%`
}

function formatDateShort(value: string): string {
  if (!value) return ''
  const parts = value.split('-')
  if (parts.length === 3) return `${parts[1]}-${parts[2]}`
  return value
}

function setDays(value: number) {
  if (days.value === value) return
  days.value = value
  loadStats()
}

onMounted(loadStats)

async function loadStats() {
  loading.value = true
  error.value = ''
  try {
    const result = await getCommandStats(days.value)
    rows.value = result.data ?? []
    trend.value = result.trend ?? []
    byPlatform.value = result.by_platform ?? []
    totals.value = result.totals ?? { calls: 0, users: 0, groups: 0, avg_ms: 0 }
    since.value = result.since
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载指令统计失败。'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.stats-range {
  display: inline-flex;
  border-radius: 999px;
  padding: 3px;
  background: rgba(165, 180, 252, 0.18);
  border: 1px solid rgba(165, 180, 252, 0.4);
}
.stats-range__btn {
  border: none;
  background: transparent;
  border-radius: 999px;
  padding: 5px 14px;
  font-size: 12px;
  font-weight: 700;
  color: var(--muted-foreground);
  cursor: pointer;
  transition: background .15s, color .15s;
}
.stats-range__btn:hover { color: #4338ca; }
.stats-range__btn--active {
  background: rgba(255, 255, 255, 0.95);
  color: #312e81;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.08);
}
.stats-metrics {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 14px;
}
.stats-trend {
  display: grid;
  gap: 8px;
}
.stats-trend__svg {
  width: 100%;
  height: 200px;
  background: rgba(99, 102, 241, 0.05);
  border-radius: 14px;
  padding: 8px;
}
.stats-trend__grid { stroke: rgba(99, 102, 241, 0.18); stroke-width: 1; stroke-dasharray: 4 6; }
.stats-trend__line { fill: none; stroke: #6366f1; stroke-width: 2; stroke-linejoin: round; stroke-linecap: round; }
.stats-trend__dot { fill: #6366f1; }
.stats-trend__axis {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--muted-foreground);
  padding: 0 4px;
}
.stats-platform {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 12px;
}
.stats-platform__row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 8px;
  font-size: 13px;
}
.stats-platform__bar {
  height: 8px;
  border-radius: 999px;
  background: rgba(165, 180, 252, 0.25);
  overflow: hidden;
  margin-top: 4px;
}
.stats-platform__fill {
  height: 100%;
  background: linear-gradient(90deg, #818cf8, #6366f1);
  border-radius: 999px;
}
</style>

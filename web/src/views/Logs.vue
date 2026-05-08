<template>
  <main class="page-stack logs-page">
    <PageHeader eyebrow="Logs" title="日志" subtitle="实时查看 zerolog 缓冲中的运行日志，支持等级与关键字过滤。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading && !paused" @click="manualRefresh">立即刷新</UiButton>
        <UiButton :variant="paused ? 'default' : 'outline'" size="sm" @click="togglePause">{{ paused ? '继续' : '暂停' }}</UiButton>
        <UiButton variant="ghost" size="sm" @click="clearEntries">清空视图</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="日志加载失败">{{ error }}</UiAlert>
    <UiAlert v-else-if="lastResponse && !lastResponse.available" variant="warning" title="日志缓冲未就绪">
      {{ lastResponse.message || '后端日志缓冲尚未初始化。' }}
    </UiAlert>

    <UiCard>
      <div class="logs-toolbar">
        <div class="logs-toolbar__group">
          <span class="logs-toolbar__label">等级</span>
          <button
            v-for="level in LEVELS"
            :key="level.key"
            type="button"
            class="logs-chip"
            :class="[`logs-chip--${level.key}`, { 'logs-chip--active': activeLevels.has(level.key) }]"
            @click="toggleLevel(level.key)"
          >{{ level.label }}</button>
        </div>
        <div class="logs-toolbar__group logs-toolbar__group--grow">
          <span class="logs-toolbar__label">关键字</span>
          <input
            v-model="queryInput"
            class="logs-input"
            type="search"
            placeholder="按消息或字段内容过滤…"
          />
        </div>
        <div class="logs-toolbar__group">
          <span class="logs-toolbar__label">间隔</span>
          <select v-model.number="intervalMs" class="logs-select">
            <option :value="2000">2 秒</option>
            <option :value="5000">5 秒</option>
            <option :value="10000">10 秒</option>
            <option :value="0">手动</option>
          </select>
        </div>
      </div>

      <div class="logs-meta">
        <UiBadge variant="outline">缓冲 {{ bufferedLabel }}</UiBadge>
        <UiBadge v-if="droppedCount > 0" variant="warning">已丢弃 {{ droppedCount }} 条</UiBadge>
        <UiBadge :variant="paused ? 'warning' : 'secondary'">{{ paused ? '已暂停' : autoRefreshLabel }}</UiBadge>
        <span class="muted-text">显示 {{ visibleEntries.length }} / 缓存 {{ entries.length }}（最近优先）</span>
      </div>

      <div v-if="loading && entries.length === 0" class="table-skeleton">
        <UiSkeleton v-for="item in 6" :key="item" height="32px" />
      </div>
      <div v-else-if="visibleEntries.length === 0" class="empty-state compact">
        <div class="empty-state__icon"><SvgIcon name="logs" :size="22" /></div>
        <p>{{ entries.length === 0 ? '暂无日志条目，机器人产生日志后会自动出现。' : '当前过滤条件下没有匹配条目。' }}</p>
      </div>
      <div v-else class="logs-stream">
        <div
          v-for="entry in visibleEntries"
          :key="entry.seq"
          class="logs-row"
          :class="`logs-row--${normaliseLevel(entry.level)}`"
        >
          <span class="logs-row__time">{{ formatTime(entry.time) }}</span>
          <span class="logs-row__level" :class="`logs-chip--${normaliseLevel(entry.level)}`">{{ normaliseLevel(entry.level).toUpperCase() }}</span>
          <span class="logs-row__message">{{ entry.message || '(空消息)' }}</span>
          <span v-if="entry.fields && Object.keys(entry.fields).length > 0" class="logs-row__fields">{{ formatFields(entry.fields) }}</span>
        </div>
      </div>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { getLogs } from '../api/client'
import type { LogEntry, LogsResponse } from '../api/types'
import SvgIcon from '../components/icons/SvgIcon.vue'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

interface LevelOption { key: string; label: string }

const LEVELS: LevelOption[] = [
  { key: 'debug', label: 'Debug' },
  { key: 'info', label: 'Info' },
  { key: 'warn', label: 'Warn' },
  { key: 'error', label: 'Error' },
  { key: 'fatal', label: 'Fatal' },
]

const MAX_ENTRIES = 1000

const entries = ref<LogEntry[]>([])
const lastResponse = ref<LogsResponse | null>(null)
const lastSeq = ref(0)
const droppedCount = ref(0)
const bufferedTotal = ref(0)
const bufferCapacity = ref(0)
const loading = ref(false)
const error = ref('')
const paused = ref(false)
const queryInput = ref('')
const debouncedQuery = ref('')
const intervalMs = ref<number>(5000)
const activeLevels = ref<Set<string>>(new Set())

let pollTimer: ReturnType<typeof setTimeout> | null = null
let queryTimer: ReturnType<typeof setTimeout> | null = null

const visibleEntries = computed(() => {
  const q = debouncedQuery.value.trim().toLowerCase()
  const levels = activeLevels.value
  return entries.value.filter((entry) => {
    if (levels.size > 0 && !levels.has(normaliseLevel(entry.level))) return false
    if (!q) return true
    if (entry.message && entry.message.toLowerCase().includes(q)) return true
    if (entry.fields) {
      try {
        if (JSON.stringify(entry.fields).toLowerCase().includes(q)) return true
      } catch (_) {
        return false
      }
    }
    return false
  })
})

const bufferedLabel = computed(() => {
  if (bufferCapacity.value > 0) return `${bufferedTotal.value} / ${bufferCapacity.value}`
  return `${bufferedTotal.value}`
})

const autoRefreshLabel = computed(() => {
  if (intervalMs.value === 0) return '手动刷新'
  return `每 ${(intervalMs.value / 1000).toFixed(0)}s 刷新`
})

watch(queryInput, (val) => {
  if (queryTimer) clearTimeout(queryTimer)
  queryTimer = setTimeout(() => {
    debouncedQuery.value = val
  }, 300)
})

watch(intervalMs, () => schedulePoll())
watch(paused, () => schedulePoll())

onMounted(async () => {
  await fetchLogs(true)
  schedulePoll()
})

onUnmounted(() => {
  if (pollTimer) clearTimeout(pollTimer)
  if (queryTimer) clearTimeout(queryTimer)
})

function schedulePoll() {
  if (pollTimer) {
    clearTimeout(pollTimer)
    pollTimer = null
  }
  if (paused.value || intervalMs.value <= 0) return
  pollTimer = setTimeout(async () => {
    await fetchLogs(false)
    schedulePoll()
  }, intervalMs.value)
}

async function fetchLogs(initial: boolean) {
  loading.value = true
  if (initial) error.value = ''
  try {
    const result = await getLogs({
      sinceSeq: initial ? 0 : lastSeq.value,
      limit: initial ? 500 : 200,
    })
    lastResponse.value = result
    bufferedTotal.value = result.total ?? 0
    bufferCapacity.value = result.capacity ?? 0
    droppedCount.value = result.dropped ?? 0

    const incoming = (result.data ?? []).slice().sort((a, b) => a.seq - b.seq)
    if (incoming.length > 0) {
      if (initial) {
        entries.value = incoming.slice().reverse()
      } else {
        entries.value = [...incoming.slice().reverse(), ...entries.value]
      }
      lastSeq.value = Math.max(lastSeq.value, ...incoming.map((entry) => entry.seq))
      if (entries.value.length > MAX_ENTRIES) {
        entries.value = entries.value.slice(0, MAX_ENTRIES)
      }
    } else if (initial) {
      entries.value = []
    }

    if (result.next_seq && result.next_seq > lastSeq.value) {
      // ensure sinceSeq stays consistent even when filter hides some entries
      lastSeq.value = Math.max(lastSeq.value, lastResponse.value?.next_seq ?? 0)
    }
    error.value = ''
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载日志失败。'
  } finally {
    loading.value = false
  }
}

function manualRefresh() {
  fetchLogs(false)
}

function togglePause() {
  paused.value = !paused.value
}

function clearEntries() {
  entries.value = []
}

function toggleLevel(level: string) {
  const next = new Set(activeLevels.value)
  if (next.has(level)) next.delete(level)
  else next.add(level)
  activeLevels.value = next
}

function normaliseLevel(level: string): string {
  const lower = (level || 'info').toLowerCase()
  if (lower === 'warning') return 'warn'
  return lower
}

function formatTime(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleTimeString([], { hour12: false }) + '.' + String(date.getMilliseconds()).padStart(3, '0')
}

function formatFields(fields: Record<string, unknown>) {
  const parts: string[] = []
  for (const [key, value] of Object.entries(fields)) {
    if (value === null || value === undefined) continue
    let display: string
    if (typeof value === 'object') {
      try { display = JSON.stringify(value) } catch (_) { display = String(value) }
    } else {
      display = String(value)
    }
    if (display.length > 120) display = display.slice(0, 117) + '…'
    parts.push(`${key}=${display}`)
  }
  return parts.join(' · ')
}
</script>

<style scoped>
.logs-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 14px 18px;
  align-items: center;
  margin-bottom: 14px;
}
.logs-toolbar__group {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.logs-toolbar__group--grow { flex: 1 1 280px; }
.logs-toolbar__label {
  font-size: 12px;
  font-weight: 700;
  color: var(--muted-foreground);
  letter-spacing: .02em;
}
.logs-input,
.logs-select {
  border-radius: 12px;
  border: 1px solid rgba(165, 180, 252, 0.5);
  background: rgba(255, 255, 255, 0.86);
  padding: 7px 12px;
  font-size: 13px;
  color: inherit;
  min-width: 160px;
  transition: border-color .2s, box-shadow .2s;
}
.logs-input { width: 100%; }
.logs-input:focus,
.logs-select:focus {
  outline: none;
  border-color: rgba(99, 102, 241, 0.85);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.18);
}
.logs-chip {
  border: 1px solid rgba(165, 180, 252, 0.5);
  background: rgba(255, 255, 255, 0.6);
  border-radius: 999px;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  color: var(--muted-foreground);
  transition: background .15s, color .15s, border-color .15s;
}
.logs-chip:hover { border-color: rgba(99, 102, 241, 0.6); }
.logs-chip--active { background: rgba(99, 102, 241, 0.18); color: #4338ca; border-color: rgba(99, 102, 241, 0.55); }
.logs-chip--debug.logs-chip--active { background: rgba(148, 163, 184, 0.25); color: #475569; border-color: rgba(148, 163, 184, 0.7); }
.logs-chip--info.logs-chip--active { background: rgba(56, 189, 248, 0.18); color: #075985; border-color: rgba(56, 189, 248, 0.6); }
.logs-chip--warn.logs-chip--active { background: rgba(251, 191, 36, 0.22); color: #92400e; border-color: rgba(251, 191, 36, 0.7); }
.logs-chip--error.logs-chip--active,
.logs-chip--fatal.logs-chip--active { background: rgba(248, 113, 113, 0.2); color: #991b1b; border-color: rgba(248, 113, 113, 0.65); }

.logs-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 12px;
  margin-bottom: 12px;
}

.logs-stream {
  display: grid;
  gap: 4px;
  max-height: 60vh;
  overflow: auto;
  padding: 6px 4px;
  border-radius: 14px;
  background: rgba(15, 23, 42, 0.04);
}
.logs-row {
  display: grid;
  grid-template-columns: 96px 64px minmax(0, 1fr);
  gap: 10px;
  padding: 6px 10px;
  border-radius: 10px;
  font-family: ui-monospace, "SFMono-Regular", "JetBrains Mono", Menlo, Consolas, monospace;
  font-size: 12.5px;
  line-height: 1.5;
  color: #1f2937;
  background: rgba(255, 255, 255, 0.78);
  border: 1px solid transparent;
  align-items: baseline;
}
.logs-row__time { color: var(--muted-foreground); white-space: nowrap; }
.logs-row__level {
  text-align: center;
  border-radius: 999px;
  padding: 1px 8px;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: .04em;
  white-space: nowrap;
}
.logs-row__message { word-break: break-word; }
.logs-row__fields {
  grid-column: 3;
  color: var(--muted-foreground);
  font-size: 11.5px;
  word-break: break-word;
}
.logs-row--debug { background: rgba(148, 163, 184, 0.12); }
.logs-row--info { background: rgba(255, 255, 255, 0.78); }
.logs-row--warn { background: rgba(251, 191, 36, 0.12); border-color: rgba(251, 191, 36, 0.35); }
.logs-row--error,
.logs-row--fatal { background: rgba(248, 113, 113, 0.12); border-color: rgba(248, 113, 113, 0.45); }
</style>

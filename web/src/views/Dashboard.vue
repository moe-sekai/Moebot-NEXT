<template>
  <main class="dashboard-page">
    <section class="hero-panel">
      <div>
        <UiBadge variant="secondary">Moebot NEXT Console</UiBadge>
        <h1>Moebot NEXT</h1>
        <p>Go + ZeroBot + Fiber + Satori Renderer，一眼看到机器人、数据与渲染链路状态。</p>
        <div class="hero-panel__meta">
          <UiBadge variant="outline">管理面板 {{ webPortLabel }}</UiBadge>
          <UiBadge variant="outline">Renderer {{ rendererUrl }}</UiBadge>
          <UiBadge variant="outline">v{{ health?.version ?? '0.1.0' }}</UiBadge>
        </div>
      </div>
      <div class="hero-panel__aside">
        <div class="hero-metric">
          <span>Renderer</span>
          <strong>{{ rendererHealth?.ok ? 'Ready' : 'Check' }}</strong>
        </div>
        <div class="hero-metric">
          <span>Masterdata</span>
          <strong>{{ summary?.loaded ? 'Loaded' : 'Pending' }}</strong>
        </div>
      </div>
    </section>

    <UiAlert v-if="pageError" variant="destructive" title="状态加载失败">{{ pageError }}</UiAlert>

    <div v-if="statusLoading" class="status-grid">
      <UiSkeleton v-for="item in 5" :key="item" height="148px" />
    </div>
    <div v-else class="status-grid">
      <StatusCard title="Bot 状态" icon="🤖" :ok="status?.bot.ok" :status="status?.bot.status" :message="status?.bot.message" :meta="botMeta" />
      <StatusCard title="Web 状态" icon="🌐" :ok="status?.web.ok" :status="status?.web.status" :message="status?.web.message" :meta="webMeta" />
      <StatusCard title="Renderer 状态" icon="🎨" :ok="status?.renderer.ok" :status="status?.renderer.status" :message="status?.renderer.message" :meta="rendererMeta" />
      <StatusCard title="Masterdata 状态" icon="📚" :ok="status?.masterdata.ok" :status="status?.masterdata.status" :message="status?.masterdata.message" :meta="masterdataMeta" />
      <StatusCard title="Database 状态" icon="🗄️" :ok="status?.database.ok" :status="status?.database.status" :message="status?.database.message" :meta="status?.database.path" />
    </div>

    <div class="dashboard-grid dashboard-grid--main">
      <MasterdataSummary :summary="summary" :loading="summaryLoading" :error="summaryError" />

      <UiCard class-name="renderer-info">
        <div class="card-heading">
          <div>
            <h2>Renderer 信息</h2>
            <p>Satori 渲染服务健康检查</p>
          </div>
          <UiBadge :variant="rendererHealth?.ok ? 'success' : 'destructive'">{{ rendererHealth?.ok ? '可用' : '不可用' }}</UiBadge>
        </div>
        <UiAlert v-if="rendererError" variant="destructive" title="检查失败">{{ rendererError }}</UiAlert>
        <dl v-else class="info-list">
          <div><dt>Renderer 地址</dt><dd>{{ rendererUrl }}</dd></div>
          <div><dt>健康状态</dt><dd>{{ rendererHealth?.message ?? '等待检查' }}</dd></div>
          <div><dt>响应耗时</dt><dd>{{ rendererHealth?.latency_ms ?? 0 }} ms</dd></div>
          <div><dt>服务说明</dt><dd>{{ rendererHealth?.note ?? '3001 是渲染服务，8080 才是管理面板。' }}</dd></div>
        </dl>
      </UiCard>
    </div>

    <SatoriPreview />
    <CommandList />

    <div class="dashboard-grid dashboard-grid--bottom">
      <RecentCommands
        :commands="recentCommands"
        :loading="recentLoading"
        :error="recentError"
        :message="recentMessage"
        @refresh="loadRecentCommands"
      />
      <SearchPanel />
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  getHealth,
  getMasterdataSummary,
  getPublicConfig,
  getRecentCommands,
  getRendererHealth,
  getStatus,
} from '../api/client'
import type {
  HealthResponse,
  MasterdataSummary as MasterdataSummaryData,
  PublicConfig,
  RecentCommand,
  RendererHealth,
  RuntimeStatus,
} from '../api/types'
import CommandList from '../components/CommandList.vue'
import MasterdataSummary from '../components/MasterdataSummary.vue'
import RecentCommands from '../components/RecentCommands.vue'
import SatoriPreview from '../components/SatoriPreview.vue'
import SearchPanel from '../components/SearchPanel.vue'
import StatusCard from '../components/StatusCard.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

const health = ref<HealthResponse | null>(null)
const status = ref<RuntimeStatus | null>(null)
const summary = ref<MasterdataSummaryData | null>(null)
const rendererHealth = ref<RendererHealth | null>(null)
const publicConfig = ref<PublicConfig | null>(null)
const recentCommands = ref<RecentCommand[]>([])
const recentMessage = ref('')

const statusLoading = ref(false)
const summaryLoading = ref(false)
const rendererLoading = ref(false)
const recentLoading = ref(false)
const pageError = ref('')
const summaryError = ref('')
const rendererError = ref('')
const recentError = ref('')

const webPortLabel = computed(() => {
  const port = publicConfig.value?.web.port ?? status.value?.web.port ?? 8080
  return `:${port}`
})

const rendererUrl = computed(() => {
  return publicConfig.value?.renderer.base_url || status.value?.renderer.base_url || 'http://127.0.0.1:3001'
})

const botMeta = computed(() => {
  if (!status.value) return 'OneBot v11 反向 WebSocket 默认监听 :6700'
  return `${status.value.bot.driver_type} · ${status.value.bot.listen || '未配置监听地址'}`
})

const webMeta = computed(() => {
  if (!status.value) return 'Fiber + Vue 3 + shadcn-vue style'
  return `${status.value.web.host}:${status.value.web.port}`
})

const rendererMeta = computed(() => {
  if (!status.value) return rendererUrl.value
  return `${status.value.renderer.base_url} · ${status.value.renderer.latency_ms} ms`
})

const masterdataMeta = computed(() => {
  const counts = status.value?.masterdata.counts ?? summary.value?.counts
  if (!counts) return '等待数据加载'
  return `卡牌 ${counts.cards} / 曲目 ${counts.musics} / 活动 ${counts.events} / 卡池 ${counts.gachas}`
})

onMounted(async () => {
  await Promise.all([loadOverview(), loadSummary(), loadRendererHealth(), loadRecentCommands()])
})

async function loadOverview() {
  statusLoading.value = true
  pageError.value = ''
  try {
    const [healthData, statusData, configData] = await Promise.all([getHealth(), getStatus(), getPublicConfig()])
    health.value = healthData
    status.value = statusData
    publicConfig.value = configData
  } catch (err) {
    pageError.value = normalizeError(err, '加载运行状态失败')
  } finally {
    statusLoading.value = false
  }
}

async function loadSummary() {
  summaryLoading.value = true
  summaryError.value = ''
  try {
    summary.value = await getMasterdataSummary()
  } catch (err) {
    summaryError.value = normalizeError(err, '加载 Masterdata 统计失败')
  } finally {
    summaryLoading.value = false
  }
}

async function loadRendererHealth() {
  rendererLoading.value = true
  rendererError.value = ''
  try {
    rendererHealth.value = await getRendererHealth()
  } catch (err) {
    rendererError.value = normalizeError(err, '检查 Renderer 失败')
  } finally {
    rendererLoading.value = false
  }
}

async function loadRecentCommands() {
  recentLoading.value = true
  recentError.value = ''
  try {
    const result = await getRecentCommands(10)
    recentCommands.value = result.data ?? []
    recentMessage.value = result.message
  } catch (err) {
    recentError.value = normalizeError(err, '加载最近命令失败')
  } finally {
    recentLoading.value = false
  }
}

function normalizeError(err: unknown, fallback: string) {
  return err instanceof Error ? `${fallback}：${err.message}` : fallback
}
</script>

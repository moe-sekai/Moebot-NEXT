<template>
  <main class="page-stack">
    <PageHeader eyebrow="About" title="关于 Moebot NEXT-Go" subtitle="Project SEKAI 查询与渲染机器人控制台。" />

    <section class="about-hero ui-card">
      <div class="about-hero__logo">
        <MoebotLogo color="var(--accent-pink)" :height="70" />
      </div>
      <div>
        <UiBadge variant="secondary">Moebot NEXT-Go</UiBadge>
        <h2>面向可维护运营的轻量控制台</h2>
        <p>本控制台沿用 Vue / Vite / TypeScript，后端由 Go Fiber 提供 API，Renderer 使用 Bun + Satori + resvg 生成图片。</p>
      </div>
    </section>

    <div class="dashboard-grid dashboard-grid--main">
      <UiCard>
        <div class="card-heading">
          <div>
            <h2>版本与运行时</h2>
            <p>来自 /api/health 与 /api/status。</p>
          </div>
          <UiButton variant="outline" size="sm" :loading="loading" @click="loadAbout">刷新</UiButton>
        </div>
        <UiAlert v-if="error" variant="destructive" title="加载失败">{{ error }}</UiAlert>
        <dl class="info-list">
          <div><dt>版本</dt><dd>{{ status?.version ?? health?.version ?? '0.1.0' }}</dd></div>
          <div><dt>健康状态</dt><dd>{{ health?.status ?? '-' }}</dd></div>
          <div><dt>当前时间</dt><dd>{{ formatTime(status?.time ?? health?.time) }}</dd></div>
          <div><dt>运行时长</dt><dd>{{ status?.uptime ?? health?.uptime ?? '-' }}</dd></div>
          <div><dt>Web 监听</dt><dd>{{ status ? `${status.web.host}:${status.web.port}` : '-' }}</dd></div>
        </dl>
      </UiCard>

      <UiCard>
        <div class="card-heading">
          <div>
            <h2>依赖说明</h2>
            <p>控制台与渲染链路的关键组件。</p>
          </div>
          <UiBadge variant="secondary">Stack</UiBadge>
        </div>
        <div class="dependency-list">
          <div v-for="item in dependencies" :key="item.name" class="dependency-item">
            <span><SvgIcon :name="item.icon" :size="18" /></span>
            <div>
              <strong>{{ item.name }}</strong>
              <p>{{ item.description }}</p>
            </div>
          </div>
        </div>
      </UiCard>
    </div>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>项目说明</h2>
          <p>目前已拆分为清晰分页，并保留后续扩展入口。</p>
        </div>
        <UiBadge variant="outline">Console UI</UiBadge>
      </div>
      <p class="muted-text about-text">
        已提供概览、状态、设置、渲染预览、Masterdata、Bot、日志、关于等页面。设置与日志保存/流式能力依赖后端新增接口，当前以只读与占位方式呈现，避免误导为已支持写入。
      </p>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getHealth, getStatus } from '../api/client'
import type { HealthResponse, RuntimeStatus } from '../api/types'
import SvgIcon, { type IconName } from '../components/icons/SvgIcon.vue'
import MoebotLogo from '../components/MoebotLogo.vue'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'

const health = ref<HealthResponse | null>(null)
const status = ref<RuntimeStatus | null>(null)
const loading = ref(false)
const error = ref('')

const dependencies: Array<{ name: string; description: string; icon: IconName }> = [
  { name: 'Go + Fiber', description: '提供控制台 API、静态资源 embed 与运行状态聚合。', icon: 'web' },
  { name: 'ZeroBot / OneBot', description: '负责机器人连接、命令入口与平台事件处理。', icon: 'bot' },
  { name: 'Vue 3 + Vite', description: '控制台前端技术栈，保留 TypeScript 与 vue-router。', icon: 'dashboard' },
  { name: 'Bun + Satori + resvg', description: '图片渲染微服务，预览页展示各步骤耗时。', icon: 'renderer' },
  { name: 'SQLite + Masterdata', description: '存储运行记录，并提供 Project SEKAI 基础数据查询。', icon: 'database' },
]

onMounted(loadAbout)

async function loadAbout() {
  loading.value = true
  error.value = ''
  try {
    const [healthData, statusData] = await Promise.all([getHealth(), getStatus()])
    health.value = healthData
    status.value = statusData
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载运行信息失败。'
  } finally {
    loading.value = false
  }
}

function formatTime(value?: string) {
  return value ? new Date(value).toLocaleString() : '-'
}
</script>

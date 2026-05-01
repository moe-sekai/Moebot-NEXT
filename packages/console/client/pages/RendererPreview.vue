<template>
  <k-layout>
    <div class="renderer-page">
      <header class="page-header">
        <div class="brand-heading">
          <img class="brand-logo" :src="logoUrl" alt="Moebot" />
          <div>
            <h1>Satori 渲染预览</h1>
            <p class="description">集中查看 Moebot 各功能图片的 JSX 模板、Satori SVG、resvg PNG 与耗时链路。</p>
          </div>
        </div>
        <button class="btn btn-primary" :disabled="loading || !selected" @click="renderPreview">
          {{ loading ? '渲染中...' : '重新渲染' }}
        </button>
      </header>

      <div class="workspace">
        <aside class="template-list">
          <div class="section-title">功能图片</div>
          <button
            v-for="item in templates"
            :key="item.id"
            class="template-item"
            :class="{ active: item.id === selectedId }"
            @click="selectTemplate(item)"
          >
            <div class="template-main">
              <span class="template-name">{{ item.name }}</span>
              <span class="badge" :class="item.status === 'ready' ? 'badge-ready' : 'badge-draft'">
                {{ item.status === 'ready' ? 'ready' : 'draft' }}
              </span>
            </div>
            <span class="template-command">{{ item.command }}</span>
          </button>
        </aside>

        <main class="preview-main" v-if="selected">
          <section class="panel controls-panel">
            <div class="control-group">
              <label>宽度</label>
              <input type="number" min="320" max="2000" step="20" v-model.number="width" />
            </div>
            <div class="control-group">
              <label>高度</label>
              <input type="number" min="240" max="2400" step="20" v-model.number="height" />
            </div>
            <label class="check-row">
              <input type="checkbox" v-model="debug" />
              <span>Satori debug 边框</span>
            </label>
          </section>

          <section class="panel preview-panel">
            <div class="panel-header">
              <div>
                <h2>{{ selected.name }}</h2>
                <p>{{ selected.description }}</p>
              </div>
              <div class="meta-pills" v-if="result?.success">
                <span>{{ result.width }} × {{ result.height }}</span>
                <span>{{ formatBytes(result.sizeBytes) }}</span>
                <span>{{ result.timings.totalMs }}ms</span>
              </div>
            </div>

            <div class="image-stage">
              <div v-if="loading" class="state-box">Satori 正在画图...</div>
              <div v-else-if="error" class="state-box error-box">{{ error }}</div>
              <img v-else-if="result?.image" :src="result.image" :alt="selected.name" />
              <div v-else class="state-box">请选择模板并点击渲染。</div>
            </div>
          </section>

          <section class="grid-row">
            <div class="panel">
              <h3>渲染流程</h3>
              <div class="timeline">
                <div v-for="step in flowSteps" :key="step.name" class="timeline-step">
                  <div class="dot" />
                  <div class="step-body">
                    <div class="step-title">
                      <span>{{ step.name }}</span>
                      <code v-if="step.time !== null">{{ step.time }}ms</code>
                    </div>
                    <p>{{ step.description }}</p>
                  </div>
                </div>
              </div>
            </div>

            <div class="panel">
              <h3>对应关系</h3>
              <table class="info-table">
                <tbody>
                  <tr><td>BOT 指令</td><td><code>{{ selected.command }}</code></td></tr>
                  <tr><td>模板文件</td><td><code>{{ selected.templatePath }}</code></td></tr>
                  <tr><td>Viewer 来源</td><td>{{ selected.viewerSource }}</td></tr>
                  <tr><td>状态</td><td>{{ selected.status === 'ready' ? '可预览' : '草稿模板' }}</td></tr>
                </tbody>
              </table>
            </div>
          </section>

          <section class="panel" v-if="result?.svg">
            <details>
              <summary>查看 Satori 生成的 SVG 源码</summary>
              <textarea readonly :value="result.svg" />
            </details>
          </section>
        </main>

        <main class="preview-main" v-else>
          <div class="panel state-box">正在加载模板列表...</div>
        </main>
      </div>
    </div>
  </k-layout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { send } from '@koishijs/client'

interface RenderPreviewMeta {
  id: string
  name: string
  description: string
  command: string
  templatePath: string
  viewerSource: string
  status: 'ready' | 'draft'
  width: number
  height: number
}

interface PreviewResponse {
  success: boolean
  error?: string
  meta?: RenderPreviewMeta
  image?: string
  svg?: string
  timings?: {
    fontsMs: number
    satoriMs: number
    resvgMs: number
    totalMs: number
  }
  sizeBytes?: number
  width?: number
  height?: number
}

const logoUrl = new URL('../../../../assets/moebot.svg', import.meta.url).href

const templates = ref<RenderPreviewMeta[]>([])
const selectedId = ref('')
const width = ref(800)
const height = ref(420)
const debug = ref(false)
const loading = ref(false)
const error = ref('')
const result = ref<PreviewResponse | null>(null)

const selected = computed(() => templates.value.find(item => item.id === selectedId.value) ?? null)

const flowSteps = computed(() => {
  const timings = result.value?.timings
  return [
    { name: '示例数据 Props', description: '控制台使用安全 mock 数据喂给模板，避免依赖真实 Masterdata 和外网图片。', time: null },
    { name: 'React/JSX 模板', description: selected.value?.templatePath ?? '等待选择模板', time: null },
    { name: '字体加载', description: '加载 assets/fonts 下的 Noto Sans CJK 字体，首次渲染会更慢。', time: timings?.fontsMs ?? null },
    { name: 'Satori SVG', description: 'Satori 将 JSX 和内联样式转换为 SVG 字符串。', time: timings?.satoriMs ?? null },
    { name: 'resvg PNG', description: 'resvg-js 将 SVG 栅格化为可发送的 PNG Buffer。', time: timings?.resvgMs ?? null },
    { name: '缓存/发送', description: 'BOT 指令中会进入图片缓存，再由 Koishi 发送到聊天平台。', time: timings?.totalMs ?? null },
  ]
})

onMounted(async () => {
  templates.value = await send('moebot/renderer/templates') || []
  if (templates.value.length > 0) {
    selectTemplate(templates.value[0])
  }
})

function selectTemplate(item: RenderPreviewMeta) {
  selectedId.value = item.id
  width.value = item.width
  height.value = item.height
  renderPreview()
}

async function renderPreview() {
  if (!selected.value) return
  loading.value = true
  error.value = ''
  result.value = null

  try {
    const response = await send('moebot/renderer/preview', {
      id: selected.value.id,
      width: Number(width.value) || selected.value.width,
      height: Number(height.value) || selected.value.height,
      debug: debug.value,
      includeSvg: true,
    }) as PreviewResponse

    if (!response?.success) {
      error.value = response?.error || '渲染失败，但服务端没有返回错误信息。'
      return
    }

    result.value = response
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
  } finally {
    loading.value = false
  }
}

function formatBytes(bytes?: number): string {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = bytes
  let index = 0
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024
    index++
  }
  return `${value.toFixed(1)} ${units[index]}`
}
</script>

<style scoped>
.renderer-page {
  --preview-accent: #33ccbb;
  --preview-bg: #f7fbff;
  --preview-surface: #ffffff;
  --preview-surface-soft: #f2f7fb;
  --preview-border: #dce8f2;
  --preview-border-strong: #b8d7e6;
  --preview-text: #172033;
  --preview-muted: #6b7b90;
  padding: 24px;
  max-width: 1440px;
  margin: 0 auto;
  min-height: 100vh;
  background: linear-gradient(180deg, #ffffff 0%, var(--preview-bg) 100%);
  color: var(--preview-text);
}

.page-header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 20px;
}

h1, h2, h3 { margin: 0; color: var(--preview-text); }
h1 { font-size: 28px; }
h2 { font-size: 20px; }
h3 { font-size: 16px; margin-bottom: 12px; }
.brand-heading {
  display: flex;
  gap: 16px;
  align-items: center;
}
.brand-logo {
  width: 160px;
  height: auto;
  padding: 10px 12px;
  border-radius: 16px;
  background: var(--preview-surface);
  border: 1px solid var(--preview-border);
}
.description { color: var(--preview-muted); margin: 6px 0 0; line-height: 1.6; }

.workspace {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  gap: 20px;
}

.template-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.section-title {
  color: var(--preview-muted);
  font-size: 13px;
  font-weight: 700;
  margin-bottom: 4px;
}

.template-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  text-align: left;
  padding: 12px;
  border-radius: 12px;
  border: 1px solid var(--preview-border);
  background: var(--preview-surface);
  color: var(--preview-text);
  cursor: pointer;
}

.template-item:hover,
.template-item.active {
  border-color: var(--preview-accent);
  background: #f4fffd;
}

.template-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.template-name { font-weight: 700; }
.template-command { color: var(--preview-muted); font-size: 12px; }

.badge {
  padding: 2px 7px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
}
.badge-ready { background: rgba(33, 179, 123, 0.12); color: #21a875; }
.badge-draft { background: rgba(255, 178, 63, 0.16); color: #d98a00; }

.preview-main {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.panel {
  background: var(--preview-surface);
  border: 1px solid var(--preview-border);
  border-radius: 16px;
  padding: 16px;
}

.controls-panel {
  display: flex;
  gap: 16px;
  align-items: center;
  flex-wrap: wrap;
}

.control-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.control-group label,
.check-row {
  color: var(--preview-muted);
  font-size: 13px;
}

.control-group input {
  width: 110px;
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid var(--preview-border);
  background: var(--preview-surface);
  color: var(--preview-text);
}

.check-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 16px;
}

.panel-header p {
  color: var(--preview-muted);
  margin: 6px 0 0;
  line-height: 1.6;
}

.meta-pills {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.meta-pills span {
  padding: 4px 8px;
  border-radius: 999px;
  background: var(--preview-surface-soft);
  color: var(--preview-muted);
  font-size: 12px;
}

.image-stage {
  display: flex;
  min-height: 360px;
  align-items: center;
  justify-content: center;
  border-radius: 14px;
  background: var(--preview-surface-soft);
  border: 1px dashed var(--preview-border-strong);
  overflow: auto;
  padding: 20px;
}

.image-stage img {
  max-width: 100%;
  height: auto;
  border-radius: 12px;
  box-shadow: 0 16px 36px rgba(51, 204, 187, 0.18);
}

.state-box {
  display: flex;
  min-height: 160px;
  align-items: center;
  justify-content: center;
  color: var(--preview-muted);
  text-align: center;
}

.error-box {
  color: #ff4466;
  white-space: pre-wrap;
}

.grid-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(320px, 0.8fr);
  gap: 16px;
}

.timeline {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.timeline-step {
  display: flex;
  gap: 10px;
}

.dot {
  width: 10px;
  height: 10px;
  margin-top: 6px;
  border-radius: 50%;
  background: var(--preview-accent);
  flex-shrink: 0;
}

.step-body {
  flex: 1;
  border-bottom: 1px solid var(--preview-border);
  padding-bottom: 10px;
}

.step-title {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-weight: 700;
}

.step-body p {
  color: var(--preview-muted);
  margin: 4px 0 0;
  font-size: 13px;
  line-height: 1.5;
}

.info-table {
  width: 100%;
  border-collapse: collapse;
}

.info-table td {
  padding: 8px 0;
  border-bottom: 1px solid var(--preview-border);
  vertical-align: top;
}

.info-table td:first-child {
  width: 88px;
  color: var(--preview-muted);
}

code {
  background: var(--preview-surface-soft);
  padding: 2px 6px;
  border-radius: 5px;
  font-size: 12px;
  word-break: break-all;
}

summary {
  cursor: pointer;
  font-weight: 700;
}

textarea {
  width: 100%;
  min-height: 240px;
  margin-top: 12px;
  padding: 12px;
  border-radius: 10px;
  border: 1px solid var(--preview-border);
  background: var(--preview-surface-soft);
  color: var(--preview-text);
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  font-size: 12px;
  resize: vertical;
}

.btn {
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid var(--preview-border);
  background: var(--preview-surface);
  color: var(--preview-text);
  cursor: pointer;
}

.btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--preview-accent);
  color: white;
  border-color: transparent;
}

@media (max-width: 980px) {
  .workspace,
  .grid-row {
    grid-template-columns: 1fr;
  }

  .page-header,
  .panel-header {
    flex-direction: column;
  }
}
</style>
/style>

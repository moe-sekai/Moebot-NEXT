<template>
  <UiCard class-name="satori-preview">
    <div class="card-heading">
      <div>
        <h2>Satori 渲染预览</h2>
        <p>Bun renderer + Satori + resvg 实时生成 PNG，便于确认模板可用性。</p>
      </div>
      <UiButton variant="outline" size="sm" :loading="loading" @click="loadPreviews">刷新模板</UiButton>
    </div>

    <UiAlert v-if="error" variant="destructive" title="Renderer 预览不可用">{{ error }}</UiAlert>

    <div v-if="loading && previews.length === 0" class="preview-layout">
      <div class="preview-list"><UiSkeleton v-for="item in 5" :key="item" height="84px" /></div>
      <UiSkeleton height="420px" radius="1rem" />
    </div>

    <div v-else-if="previews.length === 0" class="empty-state">
      <div class="empty-state__icon"><SvgIcon name="preview" :size="22" /></div>
      <p>{{ message || '暂无可预览模板。' }}</p>
    </div>

    <div v-else class="preview-layout">
      <div class="preview-list" role="listbox" aria-label="Satori 模板列表">
        <button
          v-for="item in previews"
          :key="item.id"
          class="preview-item"
          :class="{ 'preview-item--active': item.id === selectedId }"
          type="button"
          @click="selectPreview(item.id)"
        >
          <span class="preview-item__top">
            <span class="preview-item__name">{{ item.name }}</span>
            <UiBadge :variant="item.status === 'ready' ? 'success' : 'warning'">{{ item.status }}</UiBadge>
          </span>
          <span class="preview-item__desc">{{ item.command }} · {{ item.width }}px 宽</span>
        </button>
      </div>

      <div class="preview-stage">
        <div class="preview-toolbar">
          <div>
            <div class="preview-title">{{ selectedPreview?.name }}</div>
            <div class="preview-subtitle">{{ selectedPreview?.description }}</div>
          </div>
          <UiButton variant="secondary" size="sm" :loading="imageLoading" @click="refreshImage">重新渲染</UiButton>
        </div>

        <div class="preview-image-wrap">
          <UiSkeleton v-if="imageLoading" height="420px" radius="1rem" />
          <UiAlert v-if="imageError" variant="destructive" title="图片加载失败">{{ imageError }}</UiAlert>
          <img
            v-show="!imageLoading && !imageError && imageUrl"
            class="preview-image"
            :src="imageUrl"
            :alt="selectedPreview?.name || 'Satori preview'"
            @load="imageLoading = false"
            @error="handleImageError"
          />
        </div>

        <div v-if="timingItems.length > 0" class="timing-panel">
          <div class="timing-panel__header">
            <div>
              <div class="timing-panel__title">渲染时间</div>
              <div class="timing-panel__subtitle">从 Renderer 响应头读取，网络耗时由浏览器侧估算。</div>
            </div>
            <UiBadge variant="secondary">{{ formatMs(timings.total_ms) }}</UiBadge>
          </div>
          <div class="timing-grid">
            <div v-for="item in timingItems" :key="item.key" class="timing-item">
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
            </div>
          </div>
        </div>

        <dl v-if="selectedPreview" class="preview-meta">
          <div><dt>命令</dt><dd>{{ selectedPreview.command }}</dd></div>
          <div><dt>模板</dt><dd>{{ selectedPreview.templatePath }}</dd></div>
          <div><dt>来源</dt><dd>{{ selectedPreview.viewerSource }}</dd></div>
          <div><dt>图片大小</dt><dd>{{ formatBytes(timings.size_bytes) }}</dd></div>
        </dl>

        <div v-if="selectedCommandDefinition" class="command-preview-info">
          <div class="timing-panel__header">
            <div>
              <div class="timing-panel__title">指令解析</div>
              <div class="timing-panel__subtitle">{{ selectedCommandDefinition.description }}</div>
            </div>
            <UiButton variant="outline" size="sm" @click="openCommandParser">去解析</UiButton>
          </div>
          <dl class="preview-meta command-parse-meta">
            <div><dt>标准用法</dt><dd>{{ selectedCommandDefinition.usage }}</dd></div>
            <div><dt>官方别名</dt><dd>{{ joinList(selectedCommandDefinition.preset_aliases) }}</dd></div>
            <div><dt>自定义关键词</dt><dd>{{ joinList(selectedCommandDefinition.custom_aliases) }}</dd></div>
            <div><dt>示例</dt><dd>{{ selectedCommandDefinition.examples?.[0] || '-' }}</dd></div>
          </dl>
        </div>
      </div>
    </div>
  </UiCard>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getCommandDefinitions, getRendererPreviews, renderRendererPreview } from '../api/client'
import type { CommandDefinition, RenderPreviewMeta, RenderTiming } from '../api/types'
import SvgIcon from './icons/SvgIcon.vue'
import UiAlert from './ui/UiAlert.vue'
import UiBadge from './ui/UiBadge.vue'
import UiButton from './ui/UiButton.vue'
import UiCard from './ui/UiCard.vue'
import UiSkeleton from './ui/UiSkeleton.vue'

const router = useRouter()
const previews = ref<RenderPreviewMeta[]>([])
const selectedId = ref('')
const imageUrl = ref('')
const loading = ref(false)
const imageLoading = ref(false)
const error = ref('')
const imageError = ref('')
const message = ref('')
const timings = ref<RenderTiming>(emptyTiming())
const commandDefinitions = ref<CommandDefinition[]>([])

const selectedPreview = computed(() => previews.value.find(item => item.id === selectedId.value) ?? null)
const selectedCommandDefinition = computed(() => commandDefinitions.value.find(item => item.preview_id === selectedId.value) ?? null)
const timingItems = computed(() => [
  { key: 'fonts', label: '字体加载', value: formatMs(timings.value.fonts_ms) },
  { key: 'satori', label: 'Satori', value: formatMs(timings.value.satori_ms) },
  { key: 'resvg', label: 'resvg', value: formatMs(timings.value.resvg_ms) },
  { key: 'proxy', label: 'Go 代理', value: formatMs(timings.value.proxy_ms) },
  { key: 'network', label: '浏览器请求', value: formatMs(timings.value.network_ms) },
].filter(item => item.value !== '-'))

onMounted(loadPreviews)
onBeforeUnmount(() => revokeImageUrl())

async function loadPreviews() {
  loading.value = true
  error.value = ''
  try {
    const [result, commands] = await Promise.all([getRendererPreviews(), getCommandDefinitions().catch(() => null)])
    previews.value = result.data ?? []
    commandDefinitions.value = commands?.data ?? []
    message.value = result.message
    if (!selectedId.value && previews.value.length > 0) {
      selectedId.value = previews.value[0].id
    }
    await refreshImage()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载预览模板失败。'
  } finally {
    loading.value = false
  }
}

function selectPreview(id: string) {
  selectedId.value = id
  void refreshImage()
}

async function refreshImage() {
  const preview = selectedPreview.value
  if (!preview) return
  imageError.value = ''
  imageLoading.value = true
  try {
    const result = await renderRendererPreview(preview.id, preview.width)
    revokeImageUrl()
    imageUrl.value = result.url
    timings.value = result.timings
  } catch (err) {
    imageError.value = err instanceof Error ? err.message : '无法获取预览图，请确认 renderer 进程已启动且模板渲染没有报错。'
    timings.value = emptyTiming()
  } finally {
    imageLoading.value = false
  }
}

function handleImageError() {
  imageLoading.value = false
  imageError.value = '预览图已返回，但浏览器无法解码图片。'
}

function revokeImageUrl() {
  if (imageUrl.value) {
    URL.revokeObjectURL(imageUrl.value)
    imageUrl.value = ''
  }
}

function emptyTiming(): RenderTiming {
  return {
    fonts_ms: null,
    satori_ms: null,
    resvg_ms: null,
    total_ms: null,
    proxy_ms: null,
    network_ms: null,
    size_bytes: null,
  }
}

function formatMs(value: number | null) {
  return typeof value === 'number' ? `${value} ms` : '-'
}

function formatBytes(value: number | null) {
  if (typeof value !== 'number') return '-'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  return `${(value / 1024 / 1024).toFixed(2)} MB`
}

function joinList(values: string[]) {
  return values?.length ? values.join(' / ') : '-'
}

function openCommandParser() {
  const example = selectedCommandDefinition.value?.examples?.[0] || selectedCommandDefinition.value?.usage || ''
  void router.push({ path: '/commands', query: example ? { q: example } : undefined })
}
</script>

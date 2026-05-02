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
      <div class="empty-state__icon">🖼</div>
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
          <span class="preview-item__desc">{{ item.command }} · {{ item.width }}×{{ item.height }}</span>
        </button>
      </div>

      <div class="preview-stage">
        <div class="preview-toolbar">
          <div>
            <div class="preview-title">{{ selectedPreview?.name }}</div>
            <div class="preview-subtitle">{{ selectedPreview?.description }}</div>
          </div>
          <UiButton variant="secondary" size="sm" @click="refreshImage">重新渲染</UiButton>
        </div>

        <div class="preview-image-wrap">
          <UiSkeleton v-if="imageLoading" height="420px" radius="1rem" />
          <UiAlert v-if="imageError" variant="destructive" title="图片加载失败">{{ imageError }}</UiAlert>
          <img
            v-show="!imageLoading && !imageError"
            class="preview-image"
            :src="imageUrl"
            :alt="selectedPreview?.name || 'Satori preview'"
            @load="imageLoading = false"
            @error="handleImageError"
          />
        </div>

        <dl v-if="selectedPreview" class="preview-meta">
          <div><dt>命令</dt><dd>{{ selectedPreview.command }}</dd></div>
          <div><dt>模板</dt><dd>{{ selectedPreview.templatePath }}</dd></div>
          <div><dt>来源</dt><dd>{{ selectedPreview.viewerSource }}</dd></div>
        </dl>
      </div>
    </div>
  </UiCard>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getRendererPreviewImageUrl, getRendererPreviews } from '../api/client'
import type { RenderPreviewMeta } from '../api/types'
import UiAlert from './ui/UiAlert.vue'
import UiBadge from './ui/UiBadge.vue'
import UiButton from './ui/UiButton.vue'
import UiCard from './ui/UiCard.vue'
import UiSkeleton from './ui/UiSkeleton.vue'

const previews = ref<RenderPreviewMeta[]>([])
const selectedId = ref('')
const imageUrl = ref('')
const loading = ref(false)
const imageLoading = ref(false)
const error = ref('')
const imageError = ref('')
const message = ref('')

const selectedPreview = computed(() => previews.value.find(item => item.id === selectedId.value) ?? null)

onMounted(loadPreviews)

async function loadPreviews() {
  loading.value = true
  error.value = ''
  try {
    const result = await getRendererPreviews()
    previews.value = result.data ?? []
    message.value = result.message
    if (!selectedId.value && previews.value.length > 0) {
      selectedId.value = previews.value[0].id
    }
    refreshImage()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载预览模板失败。'
  } finally {
    loading.value = false
  }
}

function selectPreview(id: string) {
  selectedId.value = id
  refreshImage()
}

function refreshImage() {
  const preview = selectedPreview.value
  if (!preview) return
  imageError.value = ''
  imageLoading.value = true
  imageUrl.value = getRendererPreviewImageUrl(preview.id, preview.width, preview.height)
}

function handleImageError() {
  imageLoading.value = false
  imageError.value = '无法获取预览图，请确认 renderer 进程已启动且模板渲染没有报错。'
}
</script>

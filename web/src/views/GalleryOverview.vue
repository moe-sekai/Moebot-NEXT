<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="Plugin · Gallery"
      title="画廊管理"
      subtitle="创建与管理图片画廊，浏览缩略图、上传与删除图片。"
    >
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="reload">刷新</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="操作失败">{{ error }}</UiAlert>
    <UiAlert v-if="successMsg" variant="info" title="成功">{{ successMsg }}</UiAlert>

    <!-- 画廊列表模式 -->
    <template v-if="!selectedGallery">
      <UiCard>
        <div class="card-heading">
          <div>
            <h2>画廊列表</h2>
            <p>共 {{ galleries.length }} 个画廊</p>
          </div>
          <div class="actions">
            <UiButton variant="default" size="sm" @click="showCreateDialog = true">+ 新建画廊</UiButton>
          </div>
        </div>

        <div v-if="!galleries.length && !loading" class="empty">暂无画廊，点击右上角"新建画廊"开始。</div>

        <div class="gallery-grid">
          <div v-for="g in galleries" :key="g.name" class="gallery-card" @click="openGallery(g)">
            <div class="gallery-card-cover">
              <img v-if="g.cover_pid" :src="thumbUrl(g.cover_pid)" alt="cover" loading="lazy" />
              <div v-else class="gallery-card-cover-empty">
                <SvgIcon name="gallery" :size="32" />
              </div>
            </div>
            <div class="gallery-card-body">
              <div class="gallery-card-title">{{ g.name }}</div>
              <div class="gallery-card-meta">
                <span class="tag" :class="'tag-' + g.mode">{{ g.mode }}</span>
                <span>{{ g.pic_count }} 张</span>
              </div>
              <div v-if="g.aliases?.length" class="gallery-card-aliases">
                别名: {{ g.aliases.join(', ') }}
              </div>
            </div>
            <div class="gallery-card-actions" @click.stop>
              <button class="icon-btn" title="切换模式" @click="cycleMode(g)">模式</button>
              <button class="icon-btn icon-btn--danger" title="删除" @click="removeGallery(g.name)">删除</button>
            </div>
          </div>
        </div>
      </UiCard>
    </template>

    <!-- 画廊详情模式（图片列表） -->
    <template v-if="selectedGallery">
      <UiCard>
        <div class="card-heading">
          <div>
            <h2>
              <button class="back-btn" @click="closeGallery">&larr;</button>
              {{ selectedGallery.name }}
            </h2>
            <p>共 {{ picTotal }} 张图片 · 模式: {{ selectedGallery.mode }}</p>
          </div>
          <div class="actions">
            <UiButton variant="outline" size="sm" @click="triggerUpload">上传图片</UiButton>
          </div>
        </div>

        <input ref="fileInput" type="file" accept="image/*" multiple hidden @change="handleUpload" />

        <div v-if="!pics.length && !loading" class="empty">该画廊暂无图片。</div>

        <div class="pic-grid">
          <div v-for="pic in pics" :key="pic.PID" class="pic-item">
            <img :src="thumbUrl(pic.PID)" :alt="'PID ' + pic.PID" loading="lazy" @click="previewPic = pic" />
            <div class="pic-pid">{{ pic.PID }}</div>
            <button class="pic-del" title="删除" @click="removePic(pic.PID)">&times;</button>
          </div>
        </div>

        <div v-if="picTotal > pics.length" class="load-more">
          <UiButton variant="outline" size="sm" :loading="loading" @click="loadMorePics">加载更多</UiButton>
        </div>
      </UiCard>
    </template>

    <!-- 新建画廊弹窗 -->
    <div v-if="showCreateDialog" class="dialog-overlay" @click.self="showCreateDialog = false">
      <div class="dialog">
        <h3>新建画廊</h3>
        <input v-model="newGalleryName" type="text" placeholder="画廊名称（英文/中文，不含空格）" @keydown.enter="doCreate" />
        <div class="dialog-actions">
          <UiButton variant="ghost" size="sm" @click="showCreateDialog = false">取消</UiButton>
          <UiButton variant="default" size="sm" :loading="loading" @click="doCreate">创建</UiButton>
        </div>
      </div>
    </div>

    <!-- 图片预览弹窗 -->
    <div v-if="previewPic" class="dialog-overlay" @click.self="previewPic = null">
      <div class="preview-dialog">
        <img :src="imageUrl(previewPic.PID)" :alt="'PID ' + previewPic.PID" />
        <div class="preview-info">PID: {{ previewPic.PID }}</div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import SvgIcon from '../components/icons/SvgIcon.vue'
import type { GalleryDTO, GalleryPic } from '../api/types'
import {
  listGalleries, createGallery, deleteGallery, updateGallery,
  listGalleryPics, deleteGalleryPic, uploadGalleryPic,
  galleryPicThumbUrl, galleryPicImageUrl,
} from '../api/client'

const loading = ref(false)
const error = ref('')
const successMsg = ref('')
const galleries = ref<GalleryDTO[]>([])
const selectedGallery = ref<GalleryDTO | null>(null)
const pics = ref<GalleryPic[]>([])
const picTotal = ref(0)
const showCreateDialog = ref(false)
const newGalleryName = ref('')
const previewPic = ref<GalleryPic | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

const thumbUrl = galleryPicThumbUrl
const imageUrl = galleryPicImageUrl

function flash(msg: string) {
  successMsg.value = msg
  setTimeout(() => { successMsg.value = '' }, 3000)
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    galleries.value = await listGalleries()
  } catch (e: any) {
    error.value = e.response?.data?.message || e.message
  } finally {
    loading.value = false
  }
}

onMounted(reload)

async function doCreate() {
  if (!newGalleryName.value.trim()) return
  loading.value = true
  error.value = ''
  try {
    await createGallery(newGalleryName.value.trim())
    showCreateDialog.value = false
    newGalleryName.value = ''
    flash('画廊创建成功')
    await reload()
  } catch (e: any) {
    error.value = e.response?.data?.message || e.message
  } finally {
    loading.value = false
  }
}

async function removeGallery(name: string) {
  if (!confirm(`确定要删除画廊 "${name}" 吗？`)) return
  loading.value = true
  error.value = ''
  try {
    await deleteGallery(name)
    flash('画廊已删除')
    await reload()
  } catch (e: any) {
    error.value = e.response?.data?.message || e.message
  } finally {
    loading.value = false
  }
}

const modeOrder = ['edit', 'view', 'off']
async function cycleMode(g: GalleryDTO) {
  const next = modeOrder[(modeOrder.indexOf(g.mode) + 1) % modeOrder.length]
  loading.value = true
  error.value = ''
  try {
    await updateGallery(g.name, { mode: next })
    g.mode = next
    flash(`模式已切换为 ${next}`)
  } catch (e: any) {
    error.value = e.response?.data?.message || e.message
  } finally {
    loading.value = false
  }
}

async function openGallery(g: GalleryDTO) {
  selectedGallery.value = g
  pics.value = []
  picTotal.value = 0
  await loadPics(0)
}

function closeGallery() {
  selectedGallery.value = null
  pics.value = []
  reload()
}

async function loadPics(offset: number) {
  if (!selectedGallery.value) return
  loading.value = true
  error.value = ''
  try {
    const data = await listGalleryPics(selectedGallery.value.name, offset, 100)
    if (offset === 0) {
      pics.value = data.pics || []
    } else {
      pics.value.push(...(data.pics || []))
    }
    picTotal.value = data.total
  } catch (e: any) {
    error.value = e.response?.data?.message || e.message
  } finally {
    loading.value = false
  }
}

async function loadMorePics() {
  await loadPics(pics.value.length)
}

async function removePic(pid: number) {
  if (!confirm(`确定要删除图片 PID=${pid} 吗？`)) return
  loading.value = true
  error.value = ''
  try {
    await deleteGalleryPic(pid)
    pics.value = pics.value.filter(p => p.PID !== pid)
    picTotal.value--
    flash('图片已删除')
  } catch (e: any) {
    error.value = e.response?.data?.message || e.message
  } finally {
    loading.value = false
  }
}

function triggerUpload() {
  fileInput.value?.click()
}

async function handleUpload(e: Event) {
  const input = e.target as HTMLInputElement
  const files = input.files
  if (!files?.length || !selectedGallery.value) return
  loading.value = true
  error.value = ''
  let ok = 0
  let fail = 0
  for (const file of Array.from(files)) {
    try {
      await uploadGalleryPic(selectedGallery.value.name, file)
      ok++
    } catch {
      fail++
    }
  }
  input.value = ''
  flash(`上传完成: 成功 ${ok}, 失败 ${fail}`)
  await loadPics(0)
  loading.value = false
}
</script>

<style scoped>
.gallery-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}

.gallery-card {
  border: 1px solid var(--border);
  border-radius: 0.75rem;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.15s;
  background: var(--bg-card, var(--bg-secondary));
}
.gallery-card:hover {
  box-shadow: 0 2px 12px rgba(0,0,0,0.08);
}

.gallery-card-cover {
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary, #f5f5f5);
  overflow: hidden;
}
.gallery-card-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.gallery-card-cover-empty {
  color: var(--text-tertiary);
  opacity: 0.4;
}

.gallery-card-body {
  padding: 0.75rem;
}
.gallery-card-title {
  font-weight: 600;
  font-size: 1rem;
  margin-bottom: 0.25rem;
}
.gallery-card-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: var(--text-secondary);
}
.gallery-card-aliases {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  margin-top: 0.25rem;
}

.gallery-card-actions {
  display: flex;
  gap: 0.5rem;
  padding: 0 0.75rem 0.75rem;
}

.tag {
  display: inline-block;
  padding: 0.1rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.7rem;
  font-weight: 500;
}
.tag-edit { background: #dcfce7; color: #166534; }
.tag-view { background: #dbeafe; color: #1e40af; }
.tag-off  { background: #f3f4f6; color: #6b7280; }

.pic-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 0.5rem;
  margin-top: 1rem;
}

.pic-item {
  position: relative;
  aspect-ratio: 1;
  border-radius: 0.5rem;
  overflow: hidden;
  border: 1px solid var(--border);
}
.pic-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  cursor: pointer;
}

.pic-pid {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(0,0,0,0.5);
  color: #fff;
  font-size: 0.7rem;
  text-align: center;
  padding: 0.1rem;
}
.pic-del {
  position: absolute;
  top: 2px;
  right: 2px;
  width: 20px;
  height: 20px;
  border: none;
  border-radius: 50%;
  background: rgba(220,38,38,0.8);
  color: #fff;
  font-size: 14px;
  line-height: 1;
  cursor: pointer;
  display: none;
}
.pic-item:hover .pic-del {
  display: block;
}

.load-more {
  margin-top: 1rem;
  text-align: center;
}

.back-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.2rem;
  color: var(--text-secondary);
  margin-right: 0.5rem;
  padding: 0;
}
.back-btn:hover {
  color: var(--text-primary);
}

.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.dialog {
  background: var(--bg-primary, #fff);
  border-radius: 0.75rem;
  padding: 1.5rem;
  min-width: 320px;
  max-width: 90vw;
}
.dialog h3 {
  margin: 0 0 1rem;
}
.dialog input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid var(--border);
  border-radius: 0.5rem;
  font-size: 0.9rem;
  margin-bottom: 1rem;
  box-sizing: border-box;
}
.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.preview-dialog {
  max-width: 90vw;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.preview-dialog img {
  max-width: 90vw;
  max-height: 80vh;
  object-fit: contain;
  border-radius: 0.5rem;
}
.preview-info {
  color: #fff;
  margin-top: 0.5rem;
  font-size: 0.85rem;
}

.icon-btn {
  padding: 0.2rem 0.5rem;
  border: 1px solid var(--border);
  border-radius: 0.25rem;
  background: transparent;
  cursor: pointer;
  font-size: 0.75rem;
  color: var(--text-secondary);
}
.icon-btn:hover {
  background: var(--bg-secondary);
}
.icon-btn--danger:hover {
  background: #fee2e2;
  color: #dc2626;
  border-color: #fca5a5;
}

.card-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}
.card-heading h2 {
  margin: 0;
  font-size: 1.1rem;
  display: flex;
  align-items: center;
}
.card-heading p {
  margin: 0.25rem 0 0;
  font-size: 0.85rem;
  color: var(--text-secondary);
}
.empty {
  padding: 2rem;
  text-align: center;
  color: var(--text-tertiary);
}
</style>

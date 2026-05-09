<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="Plugin · AutoChat"
      title="AutoChat 记忆管理"
      subtitle="检索向量库中的用户画像（user_memory）与对话总结（summary）；语义检索需启用 embedding。"
    >
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="search">查询</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="操作失败">{{ error }}</UiAlert>
    <UiAlert v-if="!vectorEnabled" variant="warning" title="向量库未启用">
      记忆面板依赖 sqlite-vec 向量库，请前往「概览」启用 Vector + Embedding。
    </UiAlert>

    <UiCard>
      <SectionHeader title="过滤条件" desc="关键词非空 → 走 embedding 语义检索（必须选群组与类型）；为空 → 按时间倒序列出。" />
      <div class="form-grid">
        <Field label="群组">
          <select v-model="form.group_id">
            <option :value="0">全部</option>
            <option v-for="g in groupOptions" :key="g.group_id" :value="g.group_id">
              {{ g.group_id }}（{{ g.count }} 条）
            </option>
          </select>
        </Field>
        <Field label="类型">
          <select v-model="form.type">
            <option value="">全部</option>
            <option value="user_memory">用户画像</option>
            <option value="summary">对话总结</option>
          </select>
        </Field>
        <Field label="用户 ID">
          <input v-model.number="form.user_id" type="number" placeholder="0 表示不过滤" />
        </Field>
        <Field label="Top K">
          <select v-model.number="form.limit">
            <option :value="10">10</option>
            <option :value="20">20</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
          </select>
        </Field>
        <Field label="语义关键词" full hint="留空表示按时间排序；非空时必须选择群组和类型">
          <input v-model="form.q" type="text" placeholder="例如：游戏 / 心情 / 周末计划" @keydown.enter.prevent="search" />
        </Field>
      </div>
    </UiCard>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>结果（{{ items.length }} 条 · {{ mode === 'semantic' ? '语义模式' : '时间序列' }}）</h2>
        </div>
      </div>

      <div v-if="!items.length && !loading" class="empty">无结果。</div>

      <div v-for="m in items" :key="m.id" class="memory-card">
        <div class="memory-head">
          <span class="badge" :class="m.type === 'summary' ? 'badge-summary' : 'badge-user'">
            {{ m.type === 'summary' ? '总结' : '用户画像' }}
          </span>
          <span class="meta">群 {{ m.group_id }}</span>
          <span v-if="m.type === 'user_memory'" class="meta">
            用户 {{ m.user_name || m.user_id }}
          </span>
          <span class="meta">{{ formatTime(m.timestamp) }}</span>
          <span v-if="m.score !== undefined" class="meta score">相似度 {{ (m.score * 100).toFixed(1) }}%</span>
          <span class="grow" />
          <UiButton variant="destructive" size="sm" @click="remove(m)">删除</UiButton>
        </div>
        <p class="memory-text">{{ m.text }}</p>
      </div>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import {
  listAutochatMemoryGroups,
  queryAutochatMemory,
  deleteAutochatMemory,
  type AutochatMemoryGroup,
  type AutochatMemoryItem,
} from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import SectionHeader from '../components/autochat/AutochatSectionHeader.vue'
import Field from '../components/autochat/AutochatField.vue'

const groupOptions = ref<AutochatMemoryGroup[]>([])
const vectorEnabled = ref(true)
const items = ref<AutochatMemoryItem[]>([])
const mode = ref<'semantic' | 'recent'>('recent')
const loading = ref(false)
const error = ref('')

const form = reactive({
  group_id: 0,
  user_id: 0,
  type: '' as '' | 'user_memory' | 'summary',
  q: '',
  limit: 20,
})

onMounted(async () => {
  await loadGroups()
  await search()
})

async function loadGroups() {
  try {
    const data = await listAutochatMemoryGroups()
    groupOptions.value = data.groups || []
    vectorEnabled.value = data.vector_enabled
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function search() {
  loading.value = true
  error.value = ''
  try {
    const data = await queryAutochatMemory({
      group_id: form.group_id || undefined,
      user_id: form.user_id || undefined,
      type: form.type || undefined,
      q: form.q.trim() || undefined,
      limit: form.limit,
    })
    items.value = data.items
    mode.value = data.mode
    vectorEnabled.value = data.vector_enabled
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function remove(m: AutochatMemoryItem) {
  if (!window.confirm(`删除该条${m.type === 'summary' ? '总结' : '用户画像'}？此操作不可恢复。`)) return
  try {
    await deleteAutochatMemory(m.id)
    items.value = items.value.filter(x => x.id !== m.id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

function formatTime(ts: number) {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleString()
}
</script>

<style scoped>
.form-grid {
  display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 14px;
}
.card-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; margin-bottom: 14px; }
.card-heading h2 { margin: 0; font-size: 16px; font-weight: 700; color: var(--foreground); }
.empty { padding: 24px 0; text-align: center; color: var(--muted-foreground); font-size: 13px; }

.memory-card {
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 14px 16px;
  margin-top: 10px;
  background: rgba(255, 255, 255, 0.7);
}
.memory-head { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; flex-wrap: wrap; }
.memory-head .grow { flex: 1; }
.memory-text { margin: 0; font-size: 13px; line-height: 1.6; white-space: pre-wrap; word-break: break-word; color: var(--foreground); }
.meta { font-size: 11px; color: var(--muted-foreground); font-weight: 500; }
.meta.score { color: var(--primary, #ff78b7); font-weight: 700; }
.badge { font-size: 11px; padding: 3px 10px; border-radius: 999px; font-weight: 600; }
.badge-summary { background: rgba(120, 140, 240, 0.2); color: #5868c5; }
.badge-user { background: rgba(80, 200, 120, 0.18); color: #1e8a4a; }
</style>

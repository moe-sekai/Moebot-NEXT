<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="Marketplace"
      title="插件市场"
      subtitle="官方插件 + FloatTech/ZeroBot-Plugin 上游目录。点击「跳转」打开设置或 GitHub 源码。"
    >
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="load(true)">刷新</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="加载失败">
      {{ error }}
      <p style="margin-top:6px">未鉴权 GitHub API 限速 60 次/小时；可设置 <code>MOEBOT_GITHUB_TOKEN</code> 环境变量缓解。</p>
    </UiAlert>

    <div class="market-toolbar">
      <input
        v-model="query"
        type="search"
        class="market-search"
        placeholder="按插件名搜索（如 fortune / chouxianghua）"
      />
      <div class="market-filters">
        <label><input type="radio" value="all" v-model="filter" /> 全部 ({{ rows.length }})</label>
        <label><input type="radio" value="loaded" v-model="filter" /> 已编译加载 ({{ loadedCount }})</label>
        <label><input type="radio" value="missing" v-model="filter" /> 未编译 ({{ rows.length - loadedCount }})</label>
      </div>
      <span v-if="fetchedAt" class="market-meta">
        缓存于 {{ fetchedAtLabel }}（TTL 1h）
      </span>
    </div>

    <UiCard v-if="loading && !rows.length">
      <UiSkeleton style="height: 64px" />
    </UiCard>

    <div v-else class="market-list">
      <div v-if="filtered.length === 0 && !error" class="market-empty">
        没有匹配的插件。
      </div>

      <div v-for="row in filtered" :key="`${row.kind}:${row.name}`" class="market-row">
        <div class="market-row__main">
          <div class="market-row__title">
            <span class="market-row__name">{{ row.name }}</span>
            <UiBadge :variant="row.kind === 'official' ? 'success' : 'secondary'">
              {{ row.kind === 'official' ? '官方' : '市场' }}
            </UiBadge>
            <UiBadge v-if="row.loaded" :variant="row.enabled ? 'success' : 'outline'">
              {{ row.enabled ? '已启用' : '已加载' }}
            </UiBadge>
            <UiBadge v-else-if="row.kind === 'market'" variant="outline">未编译</UiBadge>
          </div>
          <p v-if="row.description" class="market-row__desc">{{ row.description }}</p>
        </div>
        <div class="market-row__actions">
          <RouterLink
            v-if="row.route"
            :to="row.route"
            class="ui-button ui-button--default ui-button--sm"
          >跳转</RouterLink>
          <a
            v-else-if="row.href"
            :href="row.href"
            target="_blank"
            rel="noopener"
            class="ui-button ui-button--outline ui-button--sm"
          >跳转 ↗</a>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import {
  listMarketPlugins,
  listPlugins,
  type MarketPluginEntry,
  type PluginListItem,
} from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

interface Row {
  kind: 'official' | 'market'
  name: string
  description?: string
  loaded: boolean
  enabled: boolean
  route?: string
  href?: string
}

const market = ref<MarketPluginEntry[]>([])
const locals = ref<PluginListItem[]>([])
const loading = ref(false)
const error = ref('')
const fetchedAt = ref<string>('')
const query = ref('')
const filter = ref<'all' | 'loaded' | 'missing'>('all')

const fetchedAtLabel = computed(() => {
  if (!fetchedAt.value) return ''
  const d = new Date(fetchedAt.value)
  if (Number.isNaN(d.getTime())) return fetchedAt.value
  return d.toLocaleString()
})

const rows = computed<Row[]>(() => {
  const officials: Row[] = locals.value
    .filter((p) => p.category === 'official')
    .map((p) => ({
      kind: 'official' as const,
      name: p.title || p.name,
      description: p.description,
      loaded: p.loaded,
      enabled: p.enabled,
      route: p.settings_route || '/plugins',
    }))

  const marketRows: Row[] = market.value
    // 市场列表里若与官方插件同名则隐藏，避免重复。
    .filter((m) => !locals.value.some((p) => p.category === 'official' && p.name === m.name))
    .map((m) => ({
      kind: 'market' as const,
      name: m.name,
      loaded: m.loaded,
      enabled: m.enabled,
      href: m.html_url,
    }))

  return [...officials, ...marketRows]
})

const loadedCount = computed(() => rows.value.filter((r) => r.loaded).length)

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  return rows.value.filter((r) => {
    if (filter.value === 'loaded' && !r.loaded) return false
    if (filter.value === 'missing' && r.loaded) return false
    if (q && !r.name.toLowerCase().includes(q)) return false
    return true
  })
})

async function load(force = false) {
  loading.value = true
  error.value = ''
  try {
    const [resp, local] = await Promise.all([
      listMarketPlugins(force),
      listPlugins().catch(() => [] as PluginListItem[]),
    ])
    market.value = resp.items ?? []
    locals.value = local
    fetchedAt.value = resp.fetched_at
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : '加载市场列表失败。'
  } finally {
    loading.value = false
  }
}

onMounted(() => load(false))
</script>

<style scoped>
code {
  background: var(--surface-soft, rgba(255, 255, 255, 0.04));
  padding: 0 4px;
  border-radius: 4px;
}
.market-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
  align-items: center;
  padding: 12px 16px;
  background: var(--surface-soft, rgba(255, 255, 255, 0.04));
  border-radius: 10px;
}
.market-search {
  flex: 1 1 240px;
  min-width: 200px;
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid var(--border, rgba(255, 255, 255, 0.12));
  background: var(--surface, transparent);
  color: inherit;
  font: inherit;
}
.market-filters {
  display: flex;
  gap: 12px;
  font-size: 0.9em;
  color: var(--text-muted);
}
.market-filters label {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
}
.market-meta {
  font-size: 0.85em;
  color: var(--text-muted);
  margin-left: auto;
}
.market-list {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border, rgba(255, 255, 255, 0.08));
  border-radius: 12px;
  overflow: hidden;
  background: var(--surface, transparent);
}
.market-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border, rgba(255, 255, 255, 0.06));
}
.market-row:last-child {
  border-bottom: none;
}
.market-row:hover {
  background: var(--surface-soft, rgba(255, 255, 255, 0.03));
}
.market-row__main {
  flex: 1 1 auto;
  min-width: 0;
}
.market-row__title {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.market-row__name {
  font-weight: 600;
  font-size: 1rem;
}
.market-row__desc {
  margin: 4px 0 0;
  color: var(--text-muted);
  font-size: 0.9em;
}
.market-row__actions {
  flex: 0 0 auto;
}
.market-row__actions .ui-button {
  text-decoration: none;
}
.market-empty {
  padding: 24px;
  text-align: center;
  color: var(--text-muted);
}
</style>

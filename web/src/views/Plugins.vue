<template>
  <main class="page-stack">
    <PageHeader eyebrow="Plugins" title="插件" subtitle="管理已编译进当前进程的插件，启用/禁用并跳转到各插件设置。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="load">刷新</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="加载失败">{{ error }}</UiAlert>
    <UiAlert v-if="restartNotice" variant="info" title="需要重启">
      已修改启用状态，重启 moebot 进程后变更才会生效。
    </UiAlert>

    <UiCard v-if="loading && !plugins.length">
      <UiSkeleton style="height: 64px" />
    </UiCard>

    <div class="dashboard-grid dashboard-grid--main">
      <UiCard v-for="p in plugins" :key="p.name">
        <div class="card-heading">
          <div>
            <h2>{{ p.title }} <span class="version">v{{ p.version }}</span></h2>
            <p>{{ p.description || '—' }}</p>
          </div>
          <UiBadge :variant="categoryVariant(p.category)">{{ categoryLabel(p.category) }}</UiBadge>
        </div>
        <dl class="info-list">
          <div><dt>名称</dt><dd>{{ p.name }}</dd></div>
          <div v-if="p.author"><dt>作者</dt><dd>{{ p.author }}</dd></div>
          <div v-if="p.repo"><dt>仓库</dt><dd><a :href="p.repo" target="_blank" rel="noopener">{{ p.repo }}</a></dd></div>
          <div><dt>启用状态</dt><dd>{{ p.enabled ? '已启用' : '已禁用' }}</dd></div>
        </dl>
        <div class="actions">
          <UiButton
            :variant="p.enabled ? 'destructive' : 'default'"
            size="sm"
            :loading="busy === p.name"
            @click="toggle(p)"
          >{{ p.enabled ? '禁用' : '启用' }}</UiButton>
          <RouterLink
            v-if="p.settings_route"
            :to="p.settings_route"
            class="ui-button ui-button--outline ui-button--sm"
          >设置</RouterLink>
        </div>
      </UiCard>
    </div>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { listPlugins, setPluginEnabled, type PluginListItem } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

const plugins = ref<PluginListItem[]>([])
const loading = ref(false)
const error = ref('')
const busy = ref<string | null>(null)
const restartNotice = ref(false)

onMounted(load)

async function load() {
  loading.value = true
  error.value = ''
  try {
    plugins.value = await listPlugins()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载插件列表失败。'
  } finally {
    loading.value = false
  }
}

async function toggle(p: PluginListItem) {
  busy.value = p.name
  try {
    const result = await setPluginEnabled(p.name, !p.enabled)
    p.enabled = result.enabled
    if (result.requires_restart) restartNotice.value = true
  } catch (err) {
    error.value = err instanceof Error ? err.message : '操作失败。'
  } finally {
    busy.value = null
  }
}

function categoryLabel(c: PluginListItem['category']) {
  switch (c) {
    case 'official':
      return '官方'
    case 'market':
      return '市场'
    case 'third':
      return '第三方'
    default:
      return c
  }
}

function categoryVariant(c: PluginListItem['category']) {
  switch (c) {
    case 'official':
      return 'success' as const
    case 'market':
      return 'secondary' as const
    default:
      return 'outline' as const
  }
}
</script>

<style scoped>
.version {
  font-size: 0.85em;
  color: var(--text-muted);
  margin-left: 6px;
}
.actions {
  display: flex;
  gap: 12px;
  margin-top: 12px;
  align-items: center;
}
.actions .ui-button {
  text-decoration: none;
}
</style>

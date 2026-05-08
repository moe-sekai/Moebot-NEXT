<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="Core Settings"
      title="核心设置"
      subtitle="只覆盖框架自身的运行时配置；业务设置由各插件提供其专属设置页。"
    />

    <UiAlert variant="info" title="业务配置已下沉到插件">
      原"区服 / Masterdata / Assets / Sekai API"等设置属于
      <strong>MoeSekai</strong> 插件，已迁移至
      <RouterLink to="/plugins/moesekai">/plugins/moesekai</RouterLink>。
      本页仅保留对所有插件通用的核心运行时配置。
    </UiAlert>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>插件管理</h2>
          <p>启用 / 禁用插件、跳转到各插件设置、浏览插件市场。</p>
        </div>
      </div>
      <div class="quick-links">
        <RouterLink to="/plugins" class="ui-button ui-button--default ui-button--sm">插件列表</RouterLink>
        <RouterLink to="/plugins/market" class="ui-button ui-button--outline ui-button--sm">插件市场</RouterLink>
      </div>
    </UiCard>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>已加载插件</h2>
          <p>下表反映当前进程实际加载的插件，供快速判断状态。</p>
        </div>
        <UiButton variant="outline" size="sm" :loading="loading" @click="load">刷新</UiButton>
      </div>

      <UiAlert v-if="error" variant="destructive" title="加载失败">{{ error }}</UiAlert>

      <table v-if="plugins.length" class="plugins-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>分类</th>
            <th>启用偏好</th>
            <th>运行状态</th>
            <th>设置</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in plugins" :key="p.name">
            <td>
              <div class="plugin-name">{{ p.title }}</div>
              <div class="plugin-id">{{ p.name }} · v{{ p.version }}</div>
            </td>
            <td>{{ categoryLabel(p.category) }}</td>
            <td>{{ p.enabled ? '已启用' : '已禁用' }}</td>
            <td>
              <span :class="['dot', p.loaded ? 'dot--ok' : 'dot--off']" /> {{ p.loaded ? '运行中' : '已停止' }}
            </td>
            <td>
              <RouterLink
                v-if="p.settings_route"
                :to="p.settings_route"
                class="ui-button ui-button--outline ui-button--sm"
              >打开</RouterLink>
              <span v-else class="muted">—</span>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else-if="!loading" class="muted">暂无已注册插件。</div>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { listPlugins, type PluginListItem } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'

const plugins = ref<PluginListItem[]>([])
const loading = ref(false)
const error = ref('')

onMounted(load)

async function load() {
  loading.value = true
  error.value = ''
  try {
    plugins.value = await listPlugins()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败。'
  } finally {
    loading.value = false
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
</script>

<style scoped>
.card-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; margin-bottom: 8px; }
.quick-links { display: flex; gap: 12px; flex-wrap: wrap; }
.quick-links .ui-button { text-decoration: none; }
.plugins-table { width: 100%; border-collapse: collapse; margin-top: 12px; font-size: 13px; }
.plugins-table th, .plugins-table td {
  text-align: left; padding: 10px 12px;
  border-bottom: 1px solid var(--border-default, rgba(255,255,255,0.06));
}
.plugins-table th { color: var(--text-muted); font-weight: 500; font-size: 12px; }
.plugin-name { font-weight: 500; }
.plugin-id { color: var(--text-muted); font-size: 11px; }
.dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; vertical-align: middle; }
.dot--ok { background: #5fd49a; box-shadow: 0 0 6px rgba(95,212,154,0.6); }
.dot--off { background: #6c707a; }
.muted { color: var(--text-muted); }
</style>

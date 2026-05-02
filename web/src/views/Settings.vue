<template>
  <main class="page-stack">
    <PageHeader eyebrow="Settings" title="设置" subtitle="按模块查看当前公开配置；敏感字段不会在控制台暴露。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadConfig">刷新配置</UiButton>
      </template>
    </PageHeader>

    <UiAlert variant="warning" title="只读设置">
      当前后端暂未提供保存设置接口，本页仅展示 /api/config/public 返回的非敏感配置。TODO：新增设置保存 API、权限校验与敏感字段保护。
    </UiAlert>
    <UiAlert v-if="error" variant="destructive" title="配置加载失败">{{ error }}</UiAlert>

    <div v-if="loading" class="settings-grid">
      <UiSkeleton v-for="item in 5" :key="item" height="230px" />
    </div>
    <div v-else class="settings-grid">
      <ConfigSection title="Bot" description="OneBot 驱动、命令前缀与昵称。" icon="bot" :items="botItems" />
      <ConfigSection title="Renderer" description="Satori 渲染服务与缓存配置。" icon="renderer" :items="rendererItems" />
      <ConfigSection title="Web" description="Fiber 管理控制台监听配置。" icon="web" :items="webItems" />
      <ConfigSection title="Masterdata" description="基础数据来源、刷新与本地路径。" icon="masterdata" :items="masterdataItems" />
      <ConfigSection title="资源" description="CDN、别名与贴纸资源配置。" icon="resources" :items="assetItems" />
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from 'vue'
import { getPublicConfig } from '../api/client'
import type { PublicConfig } from '../api/types'
import SvgIcon, { type IconName } from '../components/icons/SvgIcon.vue'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

interface ConfigItem {
  label: string
  value: string | number | boolean
  badge?: boolean
}

const config = ref<PublicConfig | null>(null)
const loading = ref(false)
const error = ref('')

const webItems = computed<ConfigItem[]>(() => [
  { label: 'Host', value: config.value?.web.host ?? '-' },
  { label: 'Port', value: config.value?.web.port ?? '-' },
])

const botItems = computed<ConfigItem[]>(() => [
  { label: '驱动类型', value: config.value?.bot.driver_type ?? '-' },
  { label: '监听地址', value: config.value?.bot.listen ?? '-' },
  { label: '命令前缀', value: config.value?.bot.command_prefix ?? '-' },
  { label: '昵称', value: config.value?.bot.nickname?.join(' / ') || '-' },
  { label: 'URL 已配置', value: Boolean(config.value?.bot.url_configured), badge: true },
  { label: 'Token 已设置', value: Boolean(config.value?.bot.token_set), badge: true },
])

const rendererItems = computed<ConfigItem[]>(() => [
  { label: 'Base URL', value: config.value?.renderer.base_url ?? '-' },
  { label: 'Host', value: config.value?.renderer.host ?? '-' },
  { label: 'Port', value: config.value?.renderer.port ?? '-' },
  { label: '缓存启用', value: Boolean(config.value?.renderer.cache.enabled), badge: true },
  { label: '缓存路径', value: config.value?.renderer.cache.path ?? '-' },
  { label: '缓存上限', value: `${config.value?.renderer.cache.max_size_mb ?? '-'} MB` },
  { label: '缓存 TTL', value: `${config.value?.renderer.cache.ttl_hours ?? '-'} 小时` },
])

const masterdataItems = computed<ConfigItem[]>(() => [
  { label: '主 URL 已配置', value: Boolean(config.value?.masterdata.url_configured), badge: true },
  { label: '备用 URL 已配置', value: Boolean(config.value?.masterdata.fallback_url_configured), badge: true },
  { label: '本地路径', value: config.value?.masterdata.local_path ?? '-' },
  { label: '刷新间隔', value: `${config.value?.masterdata.refresh_interval ?? '-'} 秒` },
])

const assetItems = computed<ConfigItem[]>(() => [
  { label: 'CDN Source', value: config.value?.assets.cdn_source ?? '-' },
  { label: '曲名别名配置', value: Boolean(config.value?.assets.music_alias_configured), badge: true },
  { label: '贴纸路径', value: config.value?.assets.sticker_path ?? '-' },
  { label: '版本', value: config.value?.version ?? '-' },
])

const ConfigSection = defineComponent({
  props: {
    title: { type: String, required: true },
    description: { type: String, required: true },
    icon: { type: String as () => IconName, required: true },
    items: { type: Array as () => ConfigItem[], required: true },
  },
  setup(props) {
    return () => h(UiCard, { className: 'settings-card' }, () => [
      h('div', { class: 'settings-card__heading' }, [
        h('div', { class: 'settings-card__icon' }, [h(SvgIcon, { name: props.icon, size: 22 })]),
        h('div', null, [h('h2', props.title), h('p', props.description)]),
      ]),
      h('dl', { class: 'settings-list' }, props.items.map(item => h('div', { key: item.label }, [
        h('dt', item.label),
        h('dd', item.badge
          ? h(UiBadge, { variant: item.value ? 'success' : 'warning' }, () => item.value ? '是' : '否')
          : String(item.value)),
      ]))),
    ])
  },
})

onMounted(loadConfig)

async function loadConfig() {
  loading.value = true
  error.value = ''
  try {
    config.value = await getPublicConfig()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载配置失败。'
  } finally {
    loading.value = false
  }
}
</script>

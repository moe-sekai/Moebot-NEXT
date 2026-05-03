<template>
  <main class="page-stack">
    <PageHeader eyebrow="Settings" title="设置" subtitle="配置五服目标、Masterdata 来源与资源服务器；敏感字段不会在控制台暴露。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadConfig">刷新配置</UiButton>
        <UiButton size="sm" :loading="saving" :disabled="!dirty || !canSave" @click="saveSettings">保存设置</UiButton>
      </template>
    </PageHeader>

    <UiAlert variant="info" title="多服务器设置">
      支持 JP/CN/TW/KR/EN 独立 Masterdata 与 Assets。无前缀命令按用户绑定服，/cn查卡 等前缀会临时切换服务器；敏感字段仍不会暴露或被本页覆盖。
    </UiAlert>
    <UiAlert v-if="success" variant="info" title="操作完成">{{ success }}</UiAlert>
    <UiAlert v-if="error" variant="destructive" title="配置操作失败">{{ error }}</UiAlert>

    <div v-if="loading" class="settings-grid">
      <UiSkeleton v-for="item in 6" :key="item" height="230px" />
    </div>
    <template v-else>
      <div class="settings-editor-grid">
        <UiCard className="settings-card settings-card--wide">
          <div class="settings-card__heading">
            <div class="settings-card__icon"><SvgIcon name="web" :size="22" /></div>
            <div>
              <h2>目标服务器</h2>
              <p>统一选择 Project SEKAI 服务器；切换后会同步 Masterdata 与 Assets 的地区，也可在各卡片中单独覆盖。</p>
            </div>
          </div>
          <div class="settings-form settings-form--inline">
            <label class="settings-field">
              <span>服务器</span>
              <select v-model="form.server.region" class="ui-select" @change="syncServerRegion">
                <option v-for="option in regionOptions" :key="option.key" :value="option.key">
                  {{ option.label }} · {{ option.key.toUpperCase() }}
                </option>
              </select>
            </label>
            <div class="settings-hint">当前保存值：{{ config?.server.label ?? '-' }} / {{ config?.server.region?.toUpperCase() ?? '-' }}</div>
          </div>
        </UiCard>

        <UiCard className="settings-card">
          <div class="settings-card__heading">
            <div class="settings-card__icon"><SvgIcon name="masterdata" :size="22" /></div>
            <div>
              <h2>Masterdata</h2>
              <p>MoeSekai 支持 JP/CN，Haruki 支持五服，8823 支持 JP/CN/TW，自定义源需填写到 JSON 文件目录。</p>
            </div>
          </div>
          <div class="settings-form">
            <label class="settings-field">
              <span>地区</span>
              <select v-model="form.masterdata.region" class="ui-select" @change="() => normalizeMasterdataSource()">
                <option v-for="option in regionOptions" :key="option.key" :value="option.key">{{ option.label }} · {{ option.key }}</option>
              </select>
            </label>
            <label class="settings-field">
              <span>来源</span>
              <select v-model="form.masterdata.source" class="ui-select">
                <option v-for="option in masterdataSourceOptions" :key="option.key" :value="option.key" :disabled="!optionAvailable(option, form.masterdata.region)">
                  {{ option.label }}
                </option>
              </select>
            </label>
            <UiAlert v-if="!masterdataSupported" variant="warning" title="组合不可用">
              {{ selectedMasterdataLabel }} 暂不支持 {{ regionLabel(form.masterdata.region) }}，保存前请更换来源或地区。
            </UiAlert>
            <label v-if="form.masterdata.source === 'custom'" class="settings-field settings-field--full">
              <span>自定义主 URL</span>
              <input v-model.trim="form.masterdata.custom_url" class="ui-input" placeholder="https://example.com/master" />
            </label>
            <label v-if="form.masterdata.source === 'custom'" class="settings-field settings-field--full">
              <span>自定义备用 URL</span>
              <input v-model.trim="form.masterdata.custom_fallback_url" class="ui-input" placeholder="可选" />
            </label>
            <label class="settings-field">
              <span>本地缓存路径</span>
              <input v-model.trim="form.masterdata.local_path" class="ui-input" placeholder="./data/master" />
            </label>
            <label class="settings-field">
              <span>刷新间隔（秒）</span>
              <input v-model.number="form.masterdata.refresh_interval" class="ui-input" type="number" min="0" />
            </label>
          </div>
          <div class="settings-preview">
            <div><span>当前主 URL</span><code>{{ masterdataPrimaryPreview }}</code></div>
            <div><span>当前备用 URL</span><code>{{ masterdataFallbackPreview }}</code></div>
          </div>
          <div class="settings-actions-row">
            <UiButton variant="outline" size="sm" :loading="reloading === form.server.region" @click="() => reloadMasterdataNow()">立即重载 Masterdata</UiButton>
            <span class="settings-hint">保存来源后可立即重载；失败时会继续保留本地缓存兜底。</span>
          </div>
        </UiCard>
        <UiCard className="settings-card">
          <div class="settings-card__heading">
            <div class="settings-card__icon"><SvgIcon name="resources" :size="22" /></div>
            <div>
              <h2>Assets 资源服务器</h2>
              <p>MoeSekai 仅 JP/CN；sekai.best 支持 JP/CN/TW/KR/EN；自定义源会直接作为 renderer 资源 base URL。</p>
            </div>
          </div>
          <div class="settings-form">
            <label class="settings-field">
              <span>地区</span>
              <select v-model="form.assets.region" class="ui-select" @change="() => normalizeAssetSource()">
                <option v-for="option in regionOptions" :key="option.key" :value="option.key">{{ option.label }} · {{ option.key }}</option>
              </select>
            </label>
            <label class="settings-field">
              <span>来源</span>
              <select v-model="form.assets.source" class="ui-select" @change="() => normalizeAssetSource()">
                <option v-for="option in assetSourceOptions" :key="option.key" :value="option.key" :disabled="!optionAvailable(option, form.assets.region)">
                  {{ option.label }}
                </option>
              </select>
            </label>
            <label v-if="form.assets.source === 'moesekai'" class="settings-field">
              <span>镜像</span>
              <select v-model="form.assets.mirror" class="ui-select">
                <option v-for="option in assetMirrorOptions" :key="option.key" :value="option.key">{{ option.label }}</option>
              </select>
            </label>
            <UiAlert v-if="!assetSupported" variant="warning" title="组合不可用">
              {{ selectedAssetLabel }} 暂不支持 {{ regionLabel(form.assets.region) }}，保存前请更换来源或地区。
            </UiAlert>
            <label v-if="form.assets.source === 'custom'" class="settings-field settings-field--full">
              <span>自定义 Base URL</span>
              <input v-model.trim="form.assets.custom_base_url" class="ui-input" placeholder="https://example.com/sekai-jp-assets" />
            </label>
            <label class="settings-field settings-field--full">
              <span>曲名别名 URL</span>
              <input v-model.trim="form.assets.music_alias_url" class="ui-input" placeholder="https://.../music_aliases.json" />
            </label>
            <label class="settings-field">
              <span>贴纸路径</span>
              <input v-model.trim="form.assets.sticker_path" class="ui-input" placeholder="./assets/stickers" />
            </label>
          </div>
          <div class="settings-preview">
            <div><span>当前 Base URL</span><code>{{ assetBasePreview }}</code></div>
            <div><span>Renderer Source</span><code>{{ config?.assets.renderer_source || '-' }}</code></div>
          </div>
        </UiCard>
      </div>

      <div class="settings-editor-grid">
        <UiCard v-for="entry in serverEntries" :key="entry.option.key" className="settings-card">
          <div class="settings-card__heading">
            <div class="settings-card__icon"><SvgIcon name="web" :size="22" /></div>
            <div>
              <h2>{{ entry.option.label }} · {{ entry.option.key.toUpperCase() }}</h2>
              <p>命令前缀 /{{ entry.option.key }}，例如 /{{ entry.option.key }}绑定、/{{ entry.option.key }}查曲。</p>
            </div>
          </div>
          <div class="settings-form">
            <label class="settings-field">
              <span>启用该服</span>
              <select v-model="entry.form.enabled" class="ui-select" :disabled="entry.option.key === 'jp' || entry.option.key === form.server.region">
                <option :value="true">启用</option>
                <option :value="false">停用</option>
              </select>
            </label>
            <label class="settings-field">
              <span>Masterdata 来源</span>
              <select v-model="entry.form.masterdata.source" class="ui-select" @change="() => normalizeServerProfile(entry.option.key)">
                <option v-for="option in masterdataSourceOptions" :key="option.key" :value="option.key" :disabled="!optionAvailable(option, entry.form.masterdata.region)">{{ option.label }}</option>
              </select>
            </label>
            <label class="settings-field">
              <span>Masterdata 缓存</span>
              <input v-model.trim="entry.form.masterdata.local_path" class="ui-input" />
            </label>
            <label class="settings-field">
              <span>刷新间隔（秒）</span>
              <input v-model.number="entry.form.masterdata.refresh_interval" class="ui-input" type="number" min="0" />
            </label>
            <label v-if="entry.form.masterdata.source === 'custom'" class="settings-field settings-field--full">
              <span>自定义 Master URL</span>
              <input v-model.trim="entry.form.masterdata.custom_url" class="ui-input" placeholder="https://example.com/master" />
            </label>
            <label class="settings-field">
              <span>Assets 来源</span>
              <select v-model="entry.form.assets.source" class="ui-select" @change="() => normalizeServerProfile(entry.option.key)">
                <option v-for="option in assetSourceOptions" :key="option.key" :value="option.key" :disabled="!optionAvailable(option, entry.form.assets.region)">{{ option.label }}</option>
              </select>
            </label>
            <label v-if="entry.form.assets.source === 'moesekai'" class="settings-field">
              <span>Assets 镜像</span>
              <select v-model="entry.form.assets.mirror" class="ui-select">
                <option v-for="option in assetMirrorOptions" :key="option.key" :value="option.key">{{ option.label }}</option>
              </select>
            </label>
            <label v-if="entry.form.assets.source === 'custom'" class="settings-field settings-field--full">
              <span>自定义 Assets URL</span>
              <input v-model.trim="entry.form.assets.custom_base_url" class="ui-input" placeholder="https://example.com/sekai-assets" />
            </label>
            <label class="settings-field">
              <span>SEKAI API</span>
              <select v-model="entry.form.sekai_api.enabled" class="ui-select">
                <option :value="true">启用</option>
                <option :value="false">关闭</option>
              </select>
            </label>
            <label class="settings-field">
              <span>SEKAI API 地区</span>
              <select v-model="entry.form.sekai_api.region" class="ui-select">
                <option v-for="option in regionOptions" :key="option.key" :value="option.key">{{ option.key }}</option>
              </select>
            </label>
            <label class="settings-field">
              <span>Ranking 地区</span>
              <select v-model="entry.form.ranking_api.region" class="ui-select">
                <option v-for="option in regionOptions" :key="option.key" :value="option.key">{{ option.key }}</option>
              </select>
            </label>
          </div>
          <UiAlert v-if="!serverProfileSupported(entry.form)" variant="warning" title="组合不可用">
            当前 Masterdata 或 Assets 来源不支持 {{ entry.option.label }}，请更换来源。
          </UiAlert>
          <div class="settings-preview">
            <div><span>加载状态</span><code>{{ entry.state?.loaded ? '已加载' : '未加载' }} · 卡 {{ entry.state?.counts.cards ?? 0 }} / 曲 {{ entry.state?.counts.musics ?? 0 }}</code></div>
            <div><span>Master URL</span><code>{{ entry.state?.masterdata.url || '保存后由后端解析' }}</code></div>
            <div><span>Asset Base</span><code>{{ entry.state?.assets.base_url || '保存后由后端解析' }}</code></div>
            <div><span>Renderer Source</span><code>{{ entry.state?.assets.renderer_source || '-' }}</code></div>
          </div>
          <div class="settings-actions-row">
            <UiButton variant="outline" size="sm" :loading="reloading === entry.option.key" @click="() => reloadMasterdataNow(entry.option.key)">重载该服 Masterdata</UiButton>
            <span class="settings-hint">{{ entry.state?.masterdata.error || entry.state?.masterdata.load_error || '保存后新命令会立即使用该服务器配置。' }}</span>
          </div>
        </UiCard>
      </div>

      <div class="settings-grid">
        <ConfigSection title="Bot" description="OneBot 驱动、命令前缀与昵称。" icon="bot" :items="botItems" />
        <ConfigSection title="Renderer" description="Satori 渲染服务与缓存配置。" icon="renderer" :items="rendererItems" />
        <ConfigSection title="Web" description="Fiber 管理控制台监听配置。" icon="web" :items="webItems" />
        <ConfigSection title="Masterdata 状态" description="当前生效的数据来源、刷新与本地路径。" icon="masterdata" :items="masterdataItems" />
        <ConfigSection title="SEKAI API" description="玩家资料接口开关、地区与请求头配置状态。" icon="web" :items="sekaiApiItems" />
        <ConfigSection title="资源状态" description="CDN、别名与贴纸资源配置。" icon="resources" :items="assetItems" />
      </div>
    </template>
  </main>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from 'vue'
import { getPublicConfig, reloadMasterdata, updatePublicConfig } from '../api/client'
import type { ConfigOption, PublicConfig, PublicServerProfile, UpdatePublicConfigPayload } from '../api/types'
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

interface MasterdataForm {
  region: string
  source: string
  custom_url: string
  custom_fallback_url: string
  local_path: string
  refresh_interval: number
}

interface AssetsForm {
  region: string
  source: string
  mirror: string
  custom_base_url: string
  music_alias_url: string
  sticker_path: string
}

interface ServerProfileForm {
  enabled: boolean
  masterdata: MasterdataForm
  assets: AssetsForm
  sekai_api: {
    enabled: boolean
    region: string
    timeout: number
    rate_limit: number
  }
  ranking_api: {
    region: string
    timeout: number
  }
}

interface SettingsForm {
  server: { region: string }
  masterdata: MasterdataForm
  assets: AssetsForm
  servers: Record<string, ServerProfileForm>
}

const fallbackRegions: ConfigOption[] = [
  { key: 'cn', label: '国服' },
  { key: 'jp', label: '日服' },
  { key: 'tw', label: '台服' },
  { key: 'kr', label: '韩服' },
  { key: 'en', label: '国际服' },
]
const fallbackMasterdataSources: ConfigOption[] = [
  { key: 'moesekai', label: 'MoeSekai', regions: ['jp', 'cn'] },
  { key: 'haruki', label: 'Haruki GitHub', regions: ['jp', 'cn', 'tw', 'kr', 'en'] },
  { key: '8823', label: '8823 GitHub', regions: ['jp', 'cn', 'tw'] },
  { key: 'custom', label: '自定义', regions: ['jp', 'cn', 'tw', 'kr', 'en'] },
]
const fallbackAssetSources: ConfigOption[] = [
  { key: 'moesekai', label: 'MoeSekai', regions: ['jp', 'cn'] },
  { key: 'sekai_best', label: 'sekai.best', regions: ['jp', 'cn', 'tw', 'kr', 'en'] },
  { key: 'custom', label: '自定义', regions: ['jp', 'cn', 'tw', 'kr', 'en'] },
]
const fallbackAssetMirrors: ConfigOption[] = [
  { key: 'main', label: '主镜像' },
  { key: 'backup', label: '备用镜像' },
  { key: 'overseas', label: '海外镜像' },
  { key: 'overseas_backup', label: '海外备用' },
]

const config = ref<PublicConfig | null>(null)
const form = ref<SettingsForm>(createEmptyForm())
const savedSnapshot = ref('')
const loading = ref(false)
const saving = ref(false)
const reloading = ref('')
const error = ref('')
const success = ref('')

const regionOptions = computed(() => config.value?.presets.regions ?? fallbackRegions)
const masterdataSourceOptions = computed(() => config.value?.presets.masterdata_sources ?? fallbackMasterdataSources)
const assetSourceOptions = computed(() => config.value?.presets.asset_sources ?? fallbackAssetSources)
const assetMirrorOptions = computed(() => config.value?.presets.asset_mirrors ?? fallbackAssetMirrors)

const dirty = computed(() => {
  if (!config.value) return false
  return JSON.stringify(buildPayload()) !== savedSnapshot.value
})
const masterdataSupported = computed(() => optionAvailable(findOption(masterdataSourceOptions.value, form.value.masterdata.source), form.value.masterdata.region))
const assetSupported = computed(() => optionAvailable(findOption(assetSourceOptions.value, form.value.assets.source), form.value.assets.region))
const serverEntries = computed(() => regionOptions.value.map(region => ({ option: region, form: ensureServerForm(region.key), state: config.value?.servers?.[region.key] })))
const serverProfilesSupported = computed(() => serverEntries.value.every(entry => serverProfileSupported(entry.form)))
const canSave = computed(() => masterdataSupported.value && assetSupported.value && serverProfilesSupported.value)
const selectedMasterdataLabel = computed(() => findOption(masterdataSourceOptions.value, form.value.masterdata.source)?.label ?? form.value.masterdata.source)
const selectedAssetLabel = computed(() => findOption(assetSourceOptions.value, form.value.assets.source)?.label ?? form.value.assets.source)
const masterdataSelectionChanged = computed(() => {
  const current = config.value?.masterdata
  return !current || current.region !== form.value.masterdata.region || current.source !== form.value.masterdata.source
})
const assetSelectionChanged = computed(() => {
  const current = config.value?.assets
  return !current || current.region !== form.value.assets.region || current.source !== form.value.assets.source || current.mirror !== form.value.assets.mirror
})
const masterdataPrimaryPreview = computed(() => {
  if (form.value.masterdata.source === 'custom') return form.value.masterdata.custom_url || '-'
  return masterdataSelectionChanged.value ? '保存后由后端解析' : config.value?.masterdata.url || '-'
})
const masterdataFallbackPreview = computed(() => {
  if (form.value.masterdata.source === 'custom') return form.value.masterdata.custom_fallback_url || '-'
  return masterdataSelectionChanged.value ? '保存后由后端解析' : config.value?.masterdata.fallback_url || '-'
})
const assetBasePreview = computed(() => {
  if (form.value.assets.source === 'custom') return form.value.assets.custom_base_url || '-'
  return assetSelectionChanged.value ? '保存后由后端解析' : config.value?.assets.base_url || '-'
})

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
  { label: '地区', value: `${config.value?.masterdata.region_label ?? '-'} (${config.value?.masterdata.region ?? '-'})` },
  { label: '来源', value: config.value?.masterdata.source_label || config.value?.masterdata.source || '-' },
  { label: '主 URL', value: config.value?.masterdata.url || '-' },
  { label: '备用 URL', value: config.value?.masterdata.fallback_url || '-' },
  { label: '源可用', value: Boolean(config.value?.masterdata.supported), badge: true },
  { label: '本地路径', value: config.value?.masterdata.local_path ?? '-' },
  { label: '刷新间隔', value: `${config.value?.masterdata.refresh_interval ?? '-'} 秒` },
])

const sekaiApiItems = computed<ConfigItem[]>(() => [
  { label: '启用', value: Boolean(config.value?.sekai_api.enabled), badge: true },
  { label: 'Base URL 已配置', value: Boolean(config.value?.sekai_api.base_url_configured), badge: true },
  { label: '地区', value: config.value?.sekai_api.region ?? '-' },
  { label: '请求头已配置', value: Boolean(config.value?.sekai_api.headers_configured), badge: true },
  { label: 'Ranking 地区', value: config.value?.ranking_api?.region ?? '-' },
])

const assetItems = computed<ConfigItem[]>(() => [
  { label: '地区', value: `${config.value?.assets.region_label ?? '-'} (${config.value?.assets.region ?? '-'})` },
  { label: '来源', value: config.value?.assets.source_label || config.value?.assets.source || '-' },
  { label: '镜像', value: config.value?.assets.mirror_label || config.value?.assets.mirror || '-' },
  { label: 'Base URL', value: config.value?.assets.base_url || '-' },
  { label: 'Renderer Source', value: config.value?.assets.renderer_source || '-' },
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
  success.value = ''
  try {
    const data = await getPublicConfig()
    config.value = data
    applyConfigToForm(data)
  } catch (err) {
    error.value = getErrorMessage(err, '加载配置失败。')
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    const response = await updatePublicConfig(buildPayload())
    config.value = response.config
    applyConfigToForm(response.config)
    success.value = response.message || '设置已保存。'
  } catch (err) {
    error.value = getErrorMessage(err, '保存设置失败。')
  } finally {
    saving.value = false
  }
}

async function reloadMasterdataNow(region = form.value.server.region) {
  reloading.value = region
  error.value = ''
  success.value = ''
  try {
    const result = await reloadMasterdata(region)
    success.value = `${result.message}：卡牌 ${result.counts.cards} / 曲目 ${result.counts.musics} / 活动 ${result.counts.events} / 卡池 ${result.counts.gachas}`
    await loadConfig()
  } catch (err) {
    error.value = getErrorMessage(err, '重载 Masterdata 失败。')
  } finally {
    reloading.value = ''
  }
}
function createEmptyForm(): SettingsForm {
  return {
    server: { region: 'jp' },
    masterdata: {
      region: 'jp',
      source: 'moesekai',
      custom_url: '',
      custom_fallback_url: '',
      local_path: './data/master',
      refresh_interval: 3600,
    },
    assets: {
      region: 'jp',
      source: 'moesekai',
      mirror: 'main',
      custom_base_url: '',
      music_alias_url: '',
      sticker_path: './assets/stickers',
    },
    servers: {},
  }
}

function applyConfigToForm(data: PublicConfig) {
  form.value = {
    server: { region: data.server.region || 'jp' },
    masterdata: {
      region: data.masterdata.region || data.server.region || 'jp',
      source: data.masterdata.source || 'custom',
      custom_url: data.masterdata.custom_url || (data.masterdata.source === 'custom' ? data.masterdata.url : ''),
      custom_fallback_url: data.masterdata.custom_fallback_url || (data.masterdata.source === 'custom' ? data.masterdata.fallback_url : ''),
      local_path: data.masterdata.local_path || './data/master',
      refresh_interval: data.masterdata.refresh_interval ?? 3600,
    },
    assets: {
      region: data.assets.region || data.server.region || 'jp',
      source: data.assets.source || 'custom',
      mirror: data.assets.mirror || 'main',
      custom_base_url: data.assets.custom_base_url || (data.assets.source === 'custom' ? data.assets.base_url : ''),
      music_alias_url: data.assets.music_alias_url || '',
      sticker_path: data.assets.sticker_path || './assets/stickers',
    },
    servers: {},
  }
  for (const option of regionOptions.value) {
    const server = data.servers?.[option.key]
    form.value.servers[option.key] = createServerForm(option.key, server)
    normalizeServerProfile(option.key, false)
  }
  normalizeSourceSelections(false)
  savedSnapshot.value = JSON.stringify(buildPayload())
}

function buildPayload(): UpdatePublicConfigPayload {
  return {
    server: { region: form.value.server.region },
    masterdata: {
      region: form.value.masterdata.region,
      source: form.value.masterdata.source,
      custom_url: form.value.masterdata.custom_url,
      custom_fallback_url: form.value.masterdata.custom_fallback_url,
      local_path: form.value.masterdata.local_path,
      refresh_interval: Number(form.value.masterdata.refresh_interval) || 0,
    },
    assets: {
      region: form.value.assets.region,
      source: form.value.assets.source,
      mirror: form.value.assets.mirror,
      custom_base_url: form.value.assets.custom_base_url,
      music_alias_url: form.value.assets.music_alias_url,
      sticker_path: form.value.assets.sticker_path,
    },
    servers: Object.fromEntries(regionOptions.value.map(option => [option.key, buildServerPayload(option.key)])),
  }
}

function syncServerRegion() {
  form.value.masterdata.region = form.value.server.region
  form.value.assets.region = form.value.server.region
  const defaultProfile = ensureServerForm(form.value.server.region)
  defaultProfile.enabled = true
  normalizeSourceSelections()
}

function normalizeSourceSelections(clearSuccess = true) {
  normalizeMasterdataSource(clearSuccess)
  normalizeAssetSource(clearSuccess)
}

function normalizeMasterdataSource(clearSuccess = true) {
  const current = findOption(masterdataSourceOptions.value, form.value.masterdata.source)
  if (!optionAvailable(current, form.value.masterdata.region)) {
    form.value.masterdata.source = firstAvailableOption(masterdataSourceOptions.value, form.value.masterdata.region)?.key ?? 'custom'
  }
  if (clearSuccess) success.value = ''
}

function normalizeAssetSource(clearSuccess = true) {
  const current = findOption(assetSourceOptions.value, form.value.assets.source)
  if (!optionAvailable(current, form.value.assets.region)) {
    form.value.assets.source = firstAvailableOption(assetSourceOptions.value, form.value.assets.region)?.key ?? 'custom'
  }
  if (form.value.assets.source !== 'moesekai') {
    form.value.assets.mirror = ''
  } else if (!form.value.assets.mirror) {
    form.value.assets.mirror = 'main'
  }
  if (clearSuccess) success.value = ''
}

function createServerForm(region: string, server?: PublicServerProfile): ServerProfileForm {
  const masterdata = server?.masterdata
  const assets = server?.assets
  return {
    enabled: server?.enabled ?? region === 'jp',
    masterdata: {
      region: masterdata?.region || region,
      source: masterdata?.source || (region === 'jp' || region === 'cn' ? 'moesekai' : 'haruki'),
      custom_url: masterdata?.custom_url || (masterdata?.source === 'custom' ? masterdata?.url || '' : ''),
      custom_fallback_url: masterdata?.custom_fallback_url || (masterdata?.source === 'custom' ? masterdata?.fallback_url || '' : ''),
      local_path: masterdata?.local_path || `./data/master/${region}`,
      refresh_interval: masterdata?.refresh_interval ?? 3600,
    },
    assets: {
      region: assets?.region || region,
      source: assets?.source || (region === 'jp' || region === 'cn' ? 'moesekai' : 'sekai_best'),
      mirror: assets?.mirror || 'main',
      custom_base_url: assets?.custom_base_url || (assets?.source === 'custom' ? assets?.base_url || '' : ''),
      music_alias_url: assets?.music_alias_url || '',
      sticker_path: assets?.sticker_path || './assets/stickers',
    },
    sekai_api: {
      enabled: server?.sekai_api.enabled ?? false,
      region: server?.sekai_api.region || region,
      timeout: server?.sekai_api.timeout ?? 10,
      rate_limit: server?.sekai_api.rate_limit ?? 30,
    },
    ranking_api: {
      region: server?.ranking_api.region || region,
      timeout: server?.ranking_api.timeout ?? 10,
    },
  }
}

function ensureServerForm(region: string) {
  if (!form.value.servers[region]) {
    form.value.servers[region] = createServerForm(region)
  }
  return form.value.servers[region]
}

function buildServerPayload(region: string) {
  const profile = ensureServerForm(region)
  return {
    enabled: region === 'jp' || region === form.value.server.region ? true : profile.enabled,
    masterdata: {
      region: profile.masterdata.region,
      source: profile.masterdata.source,
      custom_url: profile.masterdata.custom_url,
      custom_fallback_url: profile.masterdata.custom_fallback_url,
      local_path: profile.masterdata.local_path,
      refresh_interval: Number(profile.masterdata.refresh_interval) || 0,
    },
    assets: {
      region: profile.assets.region,
      source: profile.assets.source,
      mirror: profile.assets.mirror,
      custom_base_url: profile.assets.custom_base_url,
      music_alias_url: profile.assets.music_alias_url,
      sticker_path: profile.assets.sticker_path,
    },
    sekai_api: {
      enabled: profile.sekai_api.enabled,
      region: profile.sekai_api.region,
      timeout: Number(profile.sekai_api.timeout) || 10,
      rate_limit: Number(profile.sekai_api.rate_limit) || 30,
    },
    ranking_api: {
      region: profile.ranking_api.region,
      timeout: Number(profile.ranking_api.timeout) || 10,
    },
  }
}

function normalizeServerProfile(region: string, clearSuccess = true) {
  const profile = ensureServerForm(region)
  profile.masterdata.region = region
  profile.assets.region = region
  if (!optionAvailable(findOption(masterdataSourceOptions.value, profile.masterdata.source), region)) {
    profile.masterdata.source = firstAvailableOption(masterdataSourceOptions.value, region)?.key ?? 'custom'
  }
  if (!optionAvailable(findOption(assetSourceOptions.value, profile.assets.source), region)) {
    profile.assets.source = firstAvailableOption(assetSourceOptions.value, region)?.key ?? 'custom'
  }
  if (profile.assets.source !== 'moesekai') {
    profile.assets.mirror = ''
  } else if (!profile.assets.mirror) {
    profile.assets.mirror = 'main'
  }
  if (clearSuccess) success.value = ''
}

function serverProfileSupported(profile: ServerProfileForm) {
  return optionAvailable(findOption(masterdataSourceOptions.value, profile.masterdata.source), profile.masterdata.region)
    && optionAvailable(findOption(assetSourceOptions.value, profile.assets.source), profile.assets.region)
}

function optionAvailable(option: ConfigOption | undefined, region: string) {
  if (!option) return false
  return !option.regions?.length || option.regions.includes(region)
}

function findOption(options: ConfigOption[], key: string) {
  return options.find(option => option.key === key)
}

function firstAvailableOption(options: ConfigOption[], region: string) {
  return options.find(option => optionAvailable(option, region))
}

function regionLabel(region: string) {
  return findOption(regionOptions.value, region)?.label ?? region
}

function getErrorMessage(err: unknown, fallback: string) {
  const maybeAxios = err as { response?: { data?: { message?: string } }; message?: string }
  return maybeAxios.response?.data?.message || maybeAxios.message || fallback
}
</script>

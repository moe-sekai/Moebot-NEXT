<template>
  <div class="filter-view">
    <header class="page-head">
      <div>
        <h1>Filter · OneBot 网关</h1>
        <p>统一接入 OneBot 客户端，按规则过滤事件并分发给下游 bot 应用。</p>
      </div>
      <div class="page-head-actions">
        <UiButton size="sm" variant="ghost" :loading="refreshing" @click="loadAll">刷新</UiButton>
        <UiButton size="sm" variant="outline" @click="exportYAML">导出 YAML</UiButton>
        <UiButton size="sm" variant="outline" @click="showImport = !showImport">
          {{ showImport ? '收起导入' : '导入 YAML' }}
        </UiButton>
        <UiButton size="sm" variant="ghost" @click="showRegex = !showRegex">
          {{ showRegex ? '关闭正则测试' : '正则测试' }}
        </UiButton>
      </div>
    </header>

    <UiAlert v-if="loadError" variant="destructive" title="加载失败">{{ loadError }}</UiAlert>

    <!-- 1. 连接指南 -->
    <ConnectionGuide
      :status="status"
      :host="gatewayDraft?.host"
      :port="gatewayDraft?.port"
      :suffix="gatewayDraft?.suffix"
      :apps="connectionGuideApps"
    />

    <!-- 2. 网关设置 -->
    <UiCard>
      <header class="card-heading">
        <div>
          <h2>网关设置</h2>
          <p>监听地址、Bot ID、连接参数等。修改后请保存以生效。</p>
        </div>
        <UiBadge :variant="status?.running ? 'success' : 'secondary'">
          {{ status?.running ? `已监听 ${status?.listen ?? ''}` : '未启动' }}
        </UiBadge>
      </header>
      <div v-if="gatewayDraft" class="gw-grid">
        <label class="gw-field gw-field--toggle">
          <input type="checkbox" v-model="gatewayDraft.enabled" />
          <span>启用网关</span>
        </label>
        <label class="gw-field">
          <span>Host</span>
          <input type="text" v-model="gatewayDraft.host" placeholder="0.0.0.0" />
        </label>
        <label class="gw-field">
          <span>Port</span>
          <input type="number" v-model.number="gatewayDraft.port" min="1" max="65535" />
        </label>
        <label class="gw-field">
          <span>Suffix</span>
          <input type="text" v-model="gatewayDraft.suffix" placeholder="/ws" />
        </label>
        <label class="gw-field">
          <span>Bot ID</span>
          <input type="text" v-model="gatewayDraft.bot_id" />
        </label>
        <label class="gw-field">
          <span>
            Access Token
            <small class="gw-field__hint">留空表示不鉴权；OneBot 客户端通过 Authorization 头或 ?access_token= 传入</small>
          </span>
          <input type="text" v-model="gatewayDraft.access_token" placeholder="留空 = 不鉴权" autocomplete="off" />
        </label>
        <label class="gw-field">
          <span>User-Agent</span>
          <input type="text" v-model="gatewayDraft.user_agent" />
        </label>
        <label class="gw-field">
          <span>Buffer Size</span>
          <input type="number" v-model.number="gatewayDraft.buffer_size" min="256" />
        </label>
        <label class="gw-field">
          <span>Sleep Time (s)</span>
          <input type="number" v-model.number="gatewayDraft.sleep_time" min="0" step="0.5" />
        </label>
        <label class="gw-field gw-field--toggle">
          <input type="checkbox" v-model="gatewayDraft.debug" />
          <span>Debug 日志</span>
        </label>
      </div>
      <div class="gw-actions">
        <UiButton size="sm" :loading="savingGateway" @click="saveGateway">保存网关</UiButton>
        <span class="gw-note">默认放行/拦截规则现在由「规则模板 / default」统一管理。</span>
      </div>
    </UiCard>

    <!-- 3. 规则模板 -->
    <TemplatesPanel :templates="templates" :loading="refreshing" @changed="onTemplatesChanged" />

    <!-- 4. 下游应用 -->
    <UiCard>
      <header class="card-heading">
        <div>
          <h2>下游 Bot 应用</h2>
          <p>每个应用对应一个反向 WS 链接。可直接编辑独立规则，或绑定一个模板复用规则集。</p>
        </div>
        <UiButton size="sm" @click="addNew">新增应用</UiButton>
      </header>

      <ul class="app-list" v-if="apps.length">
        <li v-for="a in apps" :key="appKey(a)" class="app-item" :class="{ 'app-item--open': openApp === appKey(a) }">
          <details :open="openApp === appKey(a)" @toggle="onToggle(a, $event)">
            <summary class="app-summary">
              <UiBadge :variant="connectedFor(a) ? 'success' : 'secondary'">
                {{ connectedFor(a) ? '已连接' : '未连接' }}
              </UiBadge>
              <strong class="app-name">
                {{ a.name || '未命名' }}
                <UiBadge v-if="a.builtin" variant="secondary">内置</UiBadge>
                <UiBadge v-if="!a.enabled" variant="secondary">已禁用</UiBadge>
              </strong>
              <code class="app-uri">{{ a.uri || '（未填）' }}</code>
              <span class="app-spacer" />
              <UiBadge v-if="templateLabel(a)" variant="secondary">模板：{{ templateLabel(a) }}</UiBadge>
              <EffectiveRulesChips v-if="a.effective_rules" :rules="a.effective_rules" />
            </summary>

            <div class="app-body">
              <div class="app-form-grid">
                <label>
                  <span>名称</span>
                  <input type="text" v-model="a.name" :disabled="a.builtin" />
                </label>
                <label>
                  <span>URI</span>
                  <input type="text" v-model="a.uri" :disabled="a.builtin" placeholder="ws://127.0.0.1:8080/ws" />
                </label>
                <label>
                  <span>Access Token</span>
                  <input type="text" v-model="a.access_token" placeholder="（可选）" />
                </label>
                <label>
                  <span>排序</span>
                  <input type="number" v-model.number="a.sort_order" />
                </label>
                <label class="gw-field--toggle">
                  <input type="checkbox" v-model="a.enabled" />
                  <span>启用此应用</span>
                </label>
              </div>

              <div class="app-template-row">
                <label>
                  <span>引用规则模板</span>
                  <select :value="a.template_id ?? ''" @change="onTemplateSelect(a, $event)">
                    <option value="">— 不引用（使用下方独立规则）—</option>
                    <option v-for="t in templates" :key="t.id" :value="t.id">
                      {{ t.name }}{{ t.builtin ? '（内置）' : '' }}
                    </option>
                  </select>
                </label>
                <p v-if="a.template_id != null" class="app-template-hint">
                  当前应用引用模板，下方规则编辑器仅用于「取消引用后」的备用值，编辑不会影响生效规则。
                </p>
              </div>

              <RuleTabs
                :model-value="ruleSetOf(a)"
                :allow-default="true"
                :disabled="a.template_id != null"
                @update:model-value="updateAppRules(a, $event)"
              />

              <div class="app-actions">
                <UiButton size="sm" :loading="savingId === (a.id || -1)" @click="save(a)">保存</UiButton>
                <UiButton
                  v-if="!a.builtin"
                  size="sm"
                  variant="destructive"
                  @click="remove(a)"
                >
                  {{ a.id ? '删除' : '丢弃草稿' }}
                </UiButton>
                <span v-if="a._error" class="app-error">{{ a._error }}</span>
              </div>
            </div>
          </details>
        </li>
      </ul>
      <p v-else class="empty-line">暂无下游应用，点击「新增应用」开始配置。</p>
    </UiCard>

    <!-- 正则测试 -->
    <UiCard v-if="showRegex">
      <header class="card-heading">
        <div>
          <h2>正则测试 · regexp2</h2>
          <p>用于验证 message 规则中的正则表达式，语法与 .NET / dlclark/regexp2 兼容。</p>
        </div>
      </header>
      <div class="regex-grid">
        <label>
          <span>正则</span>
          <input type="text" v-model="regexInput.pattern" placeholder="例如：b\\d+" />
        </label>
        <label>
          <span>测试文本</span>
          <textarea rows="3" v-model="regexInput.text" />
        </label>
      </div>
      <div class="app-actions">
        <UiButton size="sm" :loading="regexBusy" @click="runRegexTest">测试</UiButton>
      </div>
      <pre v-if="regexResult" class="regex-result">{{ JSON.stringify(regexResult, null, 2) }}</pre>
    </UiCard>

    <!-- YAML 导入 -->
    <UiCard v-if="showImport">
      <header class="card-heading">
        <div>
          <h2>导入 YAML</h2>
          <p>兼容 OneBotFilter 的 YAML 配置。导入将更新网关、默认模板及下游应用。</p>
        </div>
      </header>
      <div class="import-grid">
        <input type="file" accept=".yml,.yaml" @change="onYamlFile" />
        <textarea rows="8" v-model="yamlInput" placeholder="或粘贴 YAML 内容到此处..." />
        <label class="gw-field--toggle">
          <input type="checkbox" v-model="importReplaceAll" />
          <span>覆盖式导入（删除当前未在 YAML 中出现的非内置应用）</span>
        </label>
      </div>
      <div class="app-actions">
        <UiButton size="sm" :loading="importBusy" @click="runImport">开始导入</UiButton>
        <span v-if="importMsg" class="import-msg">{{ importMsg }}</span>
      </div>
    </UiCard>

    <!-- 实时事件 -->
    <UiCard>
      <header class="card-heading">
        <div>
          <h2>实时事件</h2>
          <p>来自 SSE 推送，最多保留 {{ MAX_EVENTS }} 条。</p>
        </div>
        <div class="ev-toolbar">
          <label class="gw-field--toggle">
            <input type="checkbox" v-model="liveStream" @change="toggleStream" />
            <span>实时</span>
          </label>
          <UiBadge :variant="sseConnected ? 'success' : 'secondary'">
            {{ sseConnected ? 'SSE 已连接' : 'SSE 离线' }}
          </UiBadge>
        </div>
      </header>
      <div class="ev-filters">
        <label class="ev-filter">
          <span>类型</span>
          <select multiple v-model="selectedKinds" class="ev-multi">
            <option v-for="k in eventKinds" :key="k" :value="k">{{ k }}</option>
          </select>
        </label>
        <label class="ev-filter">
          <span>过滤器名称</span>
          <select v-model="filterByName">
            <option value="">全部</option>
            <option v-for="n in eventFilterNames" :key="n" :value="n">{{ n }}</option>
          </select>
        </label>
        <label class="ev-filter ev-filter--grow">
          <span>搜索</span>
          <input type="text" v-model="eventSearch" placeholder="搜索 raw / reason / user / group" />
        </label>
      </div>
      <div class="filter-events">
        <p v-if="!filteredEvents.length" class="filter-events-empty">暂无事件。</p>
        <div v-for="(ev, i) in filteredEvents" :key="i" class="filter-event">
          <UiBadge :variant="eventBadgeVariant(ev.kind)">{{ ev.kind }}</UiBadge>
          <span class="filter-event-time">{{ formatEventTime(ev.time) }}</span>
          <span v-if="ev.filter">[{{ ev.filter }}]</span>
          <span v-if="ev.user_id">u:{{ ev.user_id }}</span>
          <span v-if="ev.group_id">g:{{ ev.group_id }}</span>
          <span v-if="ev.reason" class="ev-reason">{{ ev.reason }}</span>
          <span v-if="ev.raw" class="ev-raw">{{ ev.raw }}</span>
        </div>
      </div>
    </UiCard>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import ConnectionGuide from '../components/filter/ConnectionGuide.vue'
import EffectiveRulesChips from '../components/filter/EffectiveRulesChips.vue'
import RuleTabs, { type RuleSet } from '../components/filter/RuleTabs.vue'
import TemplatesPanel from '../components/filter/TemplatesPanel.vue'
import {
  createFilterApp,
  deleteFilterApp,
  exportFilterYAML,
  getFilterGateway,
  getFilterRecentEvents,
  getFilterStatus,
  importFilterYAML,
  listFilterApps,
  listFilterTemplates,
  openFilterEvents,
  testFilterRegex,
  updateFilterApp,
  updateFilterGateway,
} from '../api/client'
import type {
  FilterAppPayload,
  FilterClientStatus,
  FilterEffectiveRules,
  FilterEvent,
  FilterEventKind,
  FilterGatewayPayload,
  FilterRegexTestResponse,
  FilterStatus,
  FilterTemplatePayload,
} from '../api/types'

type AppDraft = FilterAppPayload & {
  _isNew?: boolean
  _error?: string
  _tempKey?: string
}

const eventKinds: FilterEventKind[] = [
  'allow',
  'block',
  'prefix_pass',
  'client_up',
  'client_down',
  'upstream_up',
  'upstream_down',
]

const status = ref<FilterStatus | null>(null)
const gatewayDraft = ref<FilterGatewayPayload | null>(null)
const apps = ref<AppDraft[]>([])
const templates = ref<FilterTemplatePayload[]>([])
const refreshing = ref(false)
const savingGateway = ref(false)
const savingId = ref<number | null>(null)
const loadError = ref('')
const openApp = ref<string | null>(null)

const events = ref<FilterEvent[]>([])
const liveStream = ref(true)
const sseConnected = ref(false)
let eventSource: EventSource | null = null
const MAX_EVENTS = 200

const showRegex = ref(false)
const regexInput = ref({ pattern: '', text: '' })
const regexBusy = ref(false)
const regexResult = ref<FilterRegexTestResponse | null>(null)

const showImport = ref(false)
const yamlInput = ref('')
const importReplaceAll = ref(false)
const importBusy = ref(false)
const importMsg = ref('')

const selectedKinds = ref<FilterEventKind[]>([...eventKinds])
const filterByName = ref('')
const eventSearch = ref('')

const eventFilterNames = computed(() => {
  const set = new Set<string>()
  for (const ev of events.value) if (ev.filter) set.add(ev.filter)
  return [...set].sort()
})

const filteredEvents = computed(() => {
  const kinds = new Set(selectedKinds.value)
  const name = filterByName.value
  const q = eventSearch.value.trim().toLowerCase()
  return events.value.filter((ev) => {
    if (!kinds.has(ev.kind)) return false
    if (name && ev.filter !== name) return false
    if (!q) return true
    const haystack = [ev.raw, ev.reason, ev.user_id, ev.group_id, ev.filter]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()
    return haystack.includes(q)
  })
})

const connectionGuideApps = computed<FilterClientStatus[]>(() => {
  const byName = new Map<string, FilterClientStatus>()
  for (const c of status.value?.clients || []) byName.set(c.name, c)
  return apps.value.map((a) => ({
    name: a.name,
    uri: a.uri,
    builtin: a.builtin,
    connected: byName.get(a.name)?.connected ?? false,
  }))
})

let tempCounter = 0

onMounted(async () => {
  await loadAll()
  await loadRecentEvents()
  if (liveStream.value) connectStream()
})

onBeforeUnmount(() => {
  closeStream()
})

async function loadAll() {
  refreshing.value = true
  loadError.value = ''
  try {
    const [s, gw, list, tpls] = await Promise.all([
      getFilterStatus(),
      getFilterGateway(),
      listFilterApps(),
      listFilterTemplates(),
    ])
    status.value = s
    gatewayDraft.value = gw
    apps.value = list.map((a) => normalizeApp(a))
    templates.value = tpls
  } catch (err) {
    loadError.value = err instanceof Error ? err.message : '加载 Filter 数据失败。'
  } finally {
    refreshing.value = false
  }
}

function normalizeApp(a: FilterAppPayload): AppDraft {
  const empty: FilterEffectiveRules = {
    user_id_rules: { mode: 'on', ids: [] },
    group_id_rules: { mode: 'on', ids: [] },
    message_rules: { mode: 'on', filters: [], prefix: [], prefix_replace: '' },
    private_message_rules: { mode: 'default', filters: [], prefix: [], prefix_replace: '' },
    group_message_rules: { mode: 'default', filters: [], prefix: [], prefix_replace: '' },
  }
  return {
    ...a,
    template_id: a.template_id ?? null,
    user_id_rules: a.user_id_rules || { mode: '', ids: [] },
    group_id_rules: a.group_id_rules || { mode: '', ids: [] },
    message_rules: a.message_rules || { mode: '', filters: [], prefix: [], prefix_replace: '' },
    private_message_rules:
      a.private_message_rules || { mode: '', filters: [], prefix: [], prefix_replace: '' },
    group_message_rules:
      a.group_message_rules || { mode: '', filters: [], prefix: [], prefix_replace: '' },
    effective_rules: a.effective_rules || empty,
  }
}

function appKey(a: AppDraft): string {
  if (a.id) return `id-${a.id}`
  if (!a._tempKey) a._tempKey = `tmp-${++tempCounter}`
  return a._tempKey
}

function onToggle(a: AppDraft, ev: Event) {
  if ((ev.target as HTMLDetailsElement).open) {
    openApp.value = appKey(a)
  } else if (openApp.value === appKey(a)) {
    openApp.value = null
  }
}

function connectedFor(app: AppDraft): boolean {
  if (!app.name) return false
  return Boolean(status.value?.clients?.find((c) => c.name === app.name && c.connected))
}

function templateLabel(a: AppDraft): string {
  if (a.template_id == null) return ''
  const t = templates.value.find((x) => x.id === a.template_id)
  return t ? t.name : `#${a.template_id}`
}

function ruleSetOf(a: AppDraft): RuleSet {
  if (a.template_id != null) {
    // When bound to a template, show the template's rules so the user can
    // preview without editing. Editor is disabled.
    const t = templates.value.find((x) => x.id === a.template_id)
    if (t) {
      return {
        user_id_rules: { ...t.user_id_rules },
        group_id_rules: { ...t.group_id_rules },
        message_rules: { ...t.message_rules },
        private_message_rules: { ...t.private_message_rules },
        group_message_rules: { ...t.group_message_rules },
      }
    }
  }
  return {
    user_id_rules: a.user_id_rules,
    group_id_rules: a.group_id_rules,
    message_rules: a.message_rules,
    private_message_rules: a.private_message_rules,
    group_message_rules: a.group_message_rules,
  }
}

function updateAppRules(a: AppDraft, v: RuleSet) {
  a.user_id_rules = v.user_id_rules
  a.group_id_rules = v.group_id_rules
  a.message_rules = v.message_rules
  a.private_message_rules = v.private_message_rules
  a.group_message_rules = v.group_message_rules
}

function onTemplateSelect(a: AppDraft, ev: Event) {
  const v = (ev.target as HTMLSelectElement).value
  a.template_id = v ? Number(v) : null
}

function onTemplatesChanged() {
  loadAll()
}

function addNew() {
  const draft: AppDraft = {
    id: 0,
    name: '',
    uri: '',
    access_token: '',
    enabled: true,
    builtin: false,
    sort_order: apps.value.length,
    template_id: null,
    user_id_rules: { mode: 'default', ids: [] },
    group_id_rules: { mode: 'default', ids: [] },
    message_rules: { mode: 'on', filters: [], prefix: [], prefix_replace: '' },
    private_message_rules: { mode: 'default', filters: [], prefix: [], prefix_replace: '' },
    group_message_rules: { mode: 'default', filters: [], prefix: [], prefix_replace: '' },
    effective_rules: {
      user_id_rules: { mode: 'default', ids: [] },
      group_id_rules: { mode: 'default', ids: [] },
      message_rules: { mode: 'on', filters: [], prefix: [], prefix_replace: '' },
      private_message_rules: { mode: 'default', filters: [], prefix: [], prefix_replace: '' },
      group_message_rules: { mode: 'default', filters: [], prefix: [], prefix_replace: '' },
    },
    _isNew: true,
  }
  apps.value.push(draft)
  openApp.value = appKey(draft)
}

async function saveGateway() {
  if (!gatewayDraft.value) return
  savingGateway.value = true
  loadError.value = ''
  try {
    gatewayDraft.value = await updateFilterGateway(gatewayDraft.value)
    status.value = await getFilterStatus()
  } catch (err) {
    loadError.value = err instanceof Error ? err.message : '保存网关失败。'
  } finally {
    savingGateway.value = false
  }
}

async function save(app: AppDraft) {
  app._error = ''
  if (!app.name || !app.uri) {
    app._error = '请填写名称和 URI。'
    return
  }
  savingId.value = app.id || -1
  try {
    let saved: FilterAppPayload
    if (app.id) {
      saved = await updateFilterApp(app.id, app)
    } else {
      saved = await createFilterApp(app)
    }
    Object.assign(app, normalizeApp(saved))
    app._isNew = false
    status.value = await getFilterStatus()
  } catch (err) {
    const detail =
      (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
      (err instanceof Error ? err.message : '保存失败。')
    app._error = detail
  } finally {
    savingId.value = null
  }
}

async function remove(app: AppDraft) {
  if (!app.id) {
    apps.value = apps.value.filter((a) => a !== app)
    return
  }
  if (!confirm(`确认删除下游应用「${app.name}」？`)) return
  app._error = ''
  try {
    await deleteFilterApp(app.id)
    apps.value = apps.value.filter((a) => a.id !== app.id)
    status.value = await getFilterStatus()
  } catch (err) {
    const detail =
      (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
      (err instanceof Error ? err.message : '删除失败。')
    app._error = detail
  }
}

function formatEventTime(s: string) {
  try {
    const d = new Date(s)
    return d.toLocaleTimeString() + '.' + String(d.getMilliseconds()).padStart(3, '0')
  } catch {
    return s
  }
}

function eventBadgeVariant(kind: FilterEventKind) {
  switch (kind) {
    case 'allow':
    case 'prefix_pass':
    case 'client_up':
    case 'upstream_up':
      return 'success' as const
    case 'block':
    case 'client_down':
    case 'upstream_down':
      return 'destructive' as const
    default:
      return 'secondary' as const
  }
}

async function loadRecentEvents() {
  try {
    const items = await getFilterRecentEvents(100)
    events.value = items.slice().reverse()
  } catch {
    /* ignore */
  }
}

function connectStream() {
  closeStream()
  try {
    const es = openFilterEvents(0)
    eventSource = es
    es.onopen = () => {
      sseConnected.value = true
    }
    es.onerror = () => {
      sseConnected.value = false
    }
    const handler = (msg: MessageEvent) => {
      try {
        const ev = JSON.parse(msg.data) as FilterEvent
        events.value.unshift(ev)
        if (events.value.length > MAX_EVENTS) events.value.length = MAX_EVENTS
      } catch {
        /* ignore */
      }
    }
    for (const k of eventKinds) es.addEventListener(k, handler as EventListener)
  } catch {
    sseConnected.value = false
  }
}

function closeStream() {
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
  sseConnected.value = false
}

function toggleStream() {
  if (liveStream.value) connectStream()
  else closeStream()
}

async function runRegexTest() {
  regexBusy.value = true
  try {
    regexResult.value = await testFilterRegex(regexInput.value)
  } catch (err) {
    regexResult.value = {
      compiled: false,
      matched: false,
      error: err instanceof Error ? err.message : '请求失败',
    }
  } finally {
    regexBusy.value = false
  }
}

async function exportYAML() {
  try {
    const blob = await exportFilterYAML()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'moebot-filter.yaml'
    a.click()
    URL.revokeObjectURL(url)
  } catch (err) {
    loadError.value = err instanceof Error ? err.message : '导出失败。'
  }
}

function onYamlFile(ev: Event) {
  const file = (ev.target as HTMLInputElement).files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = () => {
    yamlInput.value = String(reader.result || '')
  }
  reader.readAsText(file)
}

async function runImport() {
  if (!yamlInput.value.trim()) {
    importMsg.value = '请粘贴 YAML 内容或选择文件。'
    return
  }
  importBusy.value = true
  importMsg.value = ''
  try {
    const res = await importFilterYAML(yamlInput.value, importReplaceAll.value)
    importMsg.value = `导入成功：新建 ${res.created}、更新 ${res.updated}、共 ${res.total}。`
    await loadAll()
  } catch (err) {
    const detail =
      (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
      (err instanceof Error ? err.message : '导入失败。')
    importMsg.value = detail
  } finally {
    importBusy.value = false
  }
}
</script>

<style scoped>
.filter-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 4px;
}
.page-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  flex-wrap: wrap;
}
.page-head h1 { margin: 0; font-size: 20px; }
.page-head p { margin: 4px 0 0; font-size: 13px; color: var(--muted-foreground, #71717a); }
.page-head-actions { display: flex; gap: 6px; flex-wrap: wrap; }

.card-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  flex-wrap: wrap;
}
.card-heading h2 { margin: 0; font-size: 15px; }
.card-heading p { margin: 4px 0 0; font-size: 12px; color: var(--muted-foreground, #71717a); }

.gw-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 10px;
  margin-top: 12px;
}
.gw-field { display: flex; flex-direction: column; gap: 4px; font-size: 12px; }
.gw-field input[type='text'],
.gw-field input[type='number'] {
  padding: 6px 8px;
  border: 1px solid var(--border, #d4d4d8);
  border-radius: 6px;
  font: inherit;
  font-size: 13px;
  background: var(--background, #fff);
}
.gw-field--toggle { flex-direction: row; align-items: center; gap: 6px; padding-top: 18px; }
.gw-field--toggle input { margin: 0; }
.gw-field__hint { display: block; margin-top: 2px; font-size: 11px; font-weight: normal; color: var(--muted-foreground, #71717a); }
.gw-actions { display: flex; gap: 10px; align-items: center; margin-top: 12px; flex-wrap: wrap; }
.gw-note { font-size: 12px; color: var(--muted-foreground, #71717a); }

.app-list { list-style: none; padding: 0; margin: 12px 0 0; display: flex; flex-direction: column; gap: 8px; }
.app-item {
  border: 1px solid var(--border, #e4e4e7);
  border-radius: 10px;
  background: var(--background, #fff);
  overflow: hidden;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.app-item--open { border-color: rgba(59,130,246,0.5); box-shadow: 0 1px 4px rgba(59,130,246,0.1); }
.app-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  cursor: pointer;
  user-select: none;
  flex-wrap: wrap;
  list-style: none;
}
.app-summary::-webkit-details-marker { display: none; }
.app-name { display: inline-flex; gap: 4px; align-items: center; font-size: 14px; }
.app-uri {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  color: var(--muted-foreground, #71717a);
}
.app-spacer { flex: 1 1 auto; }

.app-body {
  border-top: 1px solid var(--border, #e4e4e7);
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: var(--muted, #f9fafb);
}
.app-form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 10px;
}
.app-form-grid label { display: flex; flex-direction: column; gap: 4px; font-size: 12px; }
.app-form-grid input {
  padding: 6px 8px;
  border: 1px solid var(--border, #d4d4d8);
  border-radius: 6px;
  font: inherit;
  font-size: 13px;
  background: var(--background, #fff);
}
.app-template-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 8px;
  background: rgba(59,130,246,0.06);
  border: 1px solid rgba(59,130,246,0.15);
}
.app-template-row label { display: flex; flex-direction: column; gap: 4px; font-size: 12px; }
.app-template-row select {
  padding: 6px 8px;
  border-radius: 6px;
  border: 1px solid var(--border, #d4d4d8);
  background: var(--background, #fff);
  font: inherit;
  font-size: 13px;
}
.app-template-hint { margin: 0; font-size: 11px; color: var(--muted-foreground, #71717a); }

.app-actions { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
.app-error { color: #b91c1c; font-size: 12px; }
.empty-line { font-size: 13px; color: var(--muted-foreground, #71717a); margin: 12px 0 0; font-style: italic; }

.regex-grid { display: flex; flex-direction: column; gap: 10px; margin-top: 10px; }
.regex-grid label { display: flex; flex-direction: column; gap: 4px; font-size: 12px; }
.regex-grid input,
.regex-grid textarea {
  padding: 6px 8px;
  border: 1px solid var(--border, #d4d4d8);
  border-radius: 6px;
  font: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  background: var(--background, #fff);
  resize: vertical;
}
.regex-result {
  margin: 10px 0 0;
  padding: 10px;
  background: var(--muted, #f4f4f5);
  border-radius: 8px;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}

.import-grid { display: flex; flex-direction: column; gap: 10px; margin-top: 10px; }
.import-grid textarea {
  padding: 8px;
  border: 1px solid var(--border, #d4d4d8);
  border-radius: 6px;
  font: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  background: var(--background, #fff);
  resize: vertical;
}
.import-msg { font-size: 12px; }

.ev-toolbar { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
.ev-filters { display: flex; gap: 12px; flex-wrap: wrap; margin: 12px 0 6px; }
.ev-filter { display: flex; flex-direction: column; gap: 4px; font-size: 12px; }
.ev-filter--grow { flex: 1 1 200px; }
.ev-filter input,
.ev-filter select {
  padding: 4px 8px;
  border: 1px solid var(--border, #d4d4d8);
  border-radius: 6px;
  font: inherit;
  font-size: 12px;
  background: var(--background, #fff);
}
.ev-multi { min-height: 90px; min-width: 140px; }

.filter-events {
  margin-top: 4px;
  max-height: 320px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
}
.filter-events-empty {
  color: var(--muted-foreground, #71717a);
  padding: 8px 0;
  font-style: italic;
}
.filter-event {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
  padding: 4px 6px;
  border-radius: 6px;
  background: var(--muted, #f4f4f5);
}
.filter-event-time { color: var(--muted-foreground, #71717a); }
.ev-reason { color: #b45309; }
.ev-raw {
  color: var(--foreground, #18181b);
  word-break: break-all;
  flex: 1 1 100%;
  background: var(--background, #fff);
  padding: 4px 6px;
  border-radius: 4px;
}
</style>

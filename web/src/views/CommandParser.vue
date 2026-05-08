<template>
  <main class="page-stack">
    <PageHeader eyebrow="Command Parser" title="指令解析" subtitle="测试聊天指令、官方预设别名与自定义关键词，并查看解析结果和 Satori 渲染预览。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loadingDefinitions" @click="loadAll">刷新</UiButton>
      </template>
    </PageHeader>

    <div class="command-parser-layout">
      <UiCard class-name="command-parser-main">
        <div class="card-heading">
          <div>
            <h2>通用输入栏</h2>
            <p>输入完整命令后会解析功能、参数、区服，并尝试生成渲染预览。</p>
          </div>
          <UiBadge variant="secondary">{{ definitions?.command_prefix || '/' }}</UiBadge>
        </div>

        <div class="command-input-row">
          <div class="command-input-wrap" @focusout="handleInputBlur">
            <input
              ref="inputEl"
              v-model="input"
              class="ui-input command-input"
              placeholder="例如 /查卡 1204、/card 初音未来、/cn查曲 Tell Your World"
              autocomplete="off"
              spellcheck="false"
              @focus="suggestionsOpen = true"
              @input="onInputChange"
              @keydown="onInputKeydown"
              @keyup.enter.exact="onInputEnter"
            />
            <ul v-if="suggestionsOpen && suggestions.length" class="command-suggest" role="listbox">
              <li
                v-for="(item, index) in suggestions"
                :key="item.key"
                class="command-suggest__item"
                :class="{ 'command-suggest__item--active': index === activeSuggestion }"
                role="option"
                :aria-selected="index === activeSuggestion"
                @mousedown.prevent="applySuggestion(item)"
                @mouseenter="activeSuggestion = index"
              >
                <span class="command-suggest__cmd">{{ item.text }}</span>
                <span class="command-suggest__meta">
                  <span class="command-suggest__name">{{ item.name }}</span>
                  <UiBadge variant="outline">{{ item.kind }}</UiBadge>
                </span>
              </li>
            </ul>
          </div>
          <UiButton :loading="parsing" @click="runParse">解析</UiButton>
          <UiButton variant="secondary" :loading="rendering" :disabled="!parsed?.definition" @click="runRender">渲染</UiButton>
        </div>

        <div v-if="showDebugBindingPanel" class="debug-binding-panel">
          <div class="debug-binding-panel__header">
            <div>
              <strong>临时绑定信息</strong>
              <span>{{ parsed?.definition?.binding_hint || '仅用于本次 WebUI 调试，不会写入数据库或影响聊天端绑定。' }}</span>
            </div>
            <UiBadge :variant="parsed?.debug_binding_used ? 'success' : 'warning'">{{ parsed?.debug_binding_used ? 'debug active' : 'debug optional' }}</UiBadge>
          </div>
          <div class="debug-binding-form">
            <label class="settings-field">
              区服
              <select v-model="debugBinding.region" class="ui-select">
                <option v-for="region in regionOptions" :key="region.key" :value="region.key">{{ region.label }} ({{ region.key.toUpperCase() }})</option>
              </select>
            </label>
            <label class="settings-field">
              游戏 UID
              <input v-model.trim="debugBinding.gameId" class="ui-input" placeholder="例如 123456789012345678" @keyup.enter="runParse" />
            </label>
          </div>
          <p class="debug-binding-hint">{{ debugBindingHint }}</p>
        </div>

        <UiAlert v-if="parseError" variant="destructive" title="解析失败">{{ parseError }}</UiAlert>

        <div v-if="parsed" class="parse-result-panel">
          <div class="parse-result-header">
            <div>
              <div class="preview-title">{{ parsed.definition?.name || '未匹配指令' }}</div>
              <div class="preview-subtitle">{{ parsed.message }}</div>
            </div>
            <UiBadge :variant="parsed.definition ? 'success' : 'warning'">{{ parsed.definition ? 'matched' : 'unknown' }}</UiBadge>
          </div>

          <div v-if="parsed.warnings?.length" class="warning-list">
            <UiAlert v-for="warning in parsed.warnings" :key="warning" variant="warning">{{ warning }}</UiAlert>
          </div>

          <dl class="preview-meta command-parse-meta">
            <div><dt>输入</dt><dd>{{ parsed.raw_input || '-' }}</dd></div>
            <div><dt>命令</dt><dd>{{ parsed.matched_command || parsed.command_text || '-' }}</dd></div>
            <div><dt>来源</dt><dd>{{ matchSourceLabel(parsed.match_source) }}</dd></div>
            <div><dt>区服</dt><dd>{{ parsed.region_label }} ({{ parsed.region?.toUpperCase() }})</dd></div>
            <div v-if="parsed.requires_binding"><dt>绑定调试</dt><dd>{{ parsed.debug_binding_used ? '已使用临时绑定' : '未使用临时绑定' }}</dd></div>
            <div><dt>参数</dt><dd>{{ parsed.argument || '-' }}</dd></div>
            <div><dt>模板</dt><dd>{{ parsed.selected?.type?.endsWith('_list') ? parsed.selected.type : (parsed.definition?.template || '-') }}</dd></div>
          </dl>

          <div v-if="parsed.selected" class="selected-result">
            <span>命中结果</span>
            <strong>#{{ parsed.selected.id }} {{ parsed.selected.title }}</strong>
            <p>{{ parsed.selected.subtitle }}</p>
          </div>

          <div v-if="parsed.results?.length" class="table-wrap compact-table">
            <table class="ui-table">
              <thead><tr><th>ID</th><th>类型</th><th>名称</th><th>摘要</th></tr></thead>
              <tbody>
                <tr v-for="row in parsed.results.slice(0, 8)" :key="`${row.type}-${row.id}`">
                  <td>{{ row.id }}</td>
                  <td><UiBadge variant="outline">{{ row.type }}</UiBadge></td>
                  <td class="font-medium">{{ row.title }}</td>
                  <td>{{ row.subtitle }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </UiCard>

      <UiCard class-name="command-render-card">
        <div class="card-heading">
          <div>
            <h2>渲染预览</h2>
            <p>真实数据优先，无法搜索或缺少上下文时使用对应 Satori 静态预览。</p>
          </div>
          <UiBadge variant="secondary">PNG</UiBadge>
        </div>
        <div class="preview-image-wrap command-preview-wrap">
          <UiSkeleton v-if="rendering" height="420px" radius="1rem" />
          <UiAlert v-else-if="renderError" variant="destructive" title="渲染失败">{{ renderError }}</UiAlert>
          <img v-else-if="imageUrl" class="preview-image" :src="imageUrl" alt="指令渲染预览" />
          <div v-else class="empty-state compact">
            <div class="empty-state__icon"><SvgIcon name="command" :size="22" /></div>
            <p>解析后点击「渲染」查看预览图。</p>
          </div>
        </div>

        <div v-if="timingItems.length" class="timing-panel">
          <div class="timing-panel__header">
            <div>
              <div class="timing-panel__title">渲染时间</div>
              <div class="timing-panel__subtitle">从 Renderer / Go 代理响应头读取。</div>
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
      </UiCard>
    </div>

    <UiCard class-name="command-tabs-card">
      <div class="card-heading">
        <div>
          <h2>指令分类</h2>
          <p>按功能类别浏览所有命令，点击 Tab 切换当前分类。选中指令可编辑自定义关键词。</p>
        </div>
        <UiBadge variant="secondary">{{ commandDefinitions.length }} 项</UiBadge>
      </div>

      <div class="command-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          class="command-tab"
          :class="{ 'command-tab--active': activeCategory === tab.key }"
          @click="setActiveCategory(tab.key)"
        >
          <span class="command-tab__icon" aria-hidden="true"><SvgIcon :name="tab.icon" :size="16" /></span>
          <span class="command-tab__label">{{ tab.label }}</span>
          <UiBadge :variant="activeCategory === tab.key ? 'secondary' : 'outline'">{{ tab.count }}</UiBadge>
        </button>
      </div>

      <p v-if="activeTabHint" class="command-tab__hint">{{ activeTabHint }}</p>

      <div class="command-tab-section">
        <h3 class="command-section-title">功能列表</h3>
        <div v-if="activeDefinitions.length" class="command-definition-grid">
          <button v-for="definition in activeDefinitions" :key="definition.id" type="button" class="command-definition-card" @click="useExample(definition.examples?.[0] || definition.usage)">
            <span class="command-definition-card__top">
              <strong>{{ definition.name }}</strong>
              <UiBadge variant="outline">{{ definition.render_mode }}</UiBadge>
            </span>
            <span class="command-definition-card__usage">{{ definition.usage }}</span>
            <span class="command-definition-card__desc">{{ definition.description }}</span>
            <span v-if="definitionExamples(definition).length" class="command-definition-card__examples">
              <span v-for="example in definitionExamples(definition).slice(0, 3)" :key="example" class="alias-chip alias-chip--example">{{ example }}</span>
            </span>
            <span v-if="definitionAliases(definition).length" class="alias-chip-row">
              <span v-for="alias in definitionAliases(definition).slice(0, 8)" :key="alias" class="alias-chip">{{ alias }}</span>
            </span>
          </button>
        </div>
        <div v-else class="empty-state compact">
          <p>该分类暂无指令。</p>
        </div>
      </div>

      <div class="command-tab-section">
        <div class="command-section-heading">
          <h3 class="command-section-title">自定义关键词</h3>
          <div class="command-alias-actions">
            <UiButton variant="outline" size="sm" @click="exportAliasesToFile">导出</UiButton>
            <UiButton variant="outline" size="sm" @click="triggerImport">导入</UiButton>
            <UiButton variant="secondary" size="sm" :loading="savingAliases" @click="resetAliasesToDefault">恢复默认</UiButton>
            <UiButton size="sm" :loading="savingAliases" @click="saveAliases">保存</UiButton>
            <input ref="importInput" class="sr-only-input" type="file" accept="application/json" @change="handleImportFile" />
          </div>
        </div>

        <UiAlert v-if="aliasMessage" variant="info">{{ aliasMessage }}</UiAlert>
        <UiAlert v-if="aliasError" variant="destructive" title="关键词配置错误">{{ aliasError }}</UiAlert>

        <div v-if="aliasEditDefinition" class="alias-editor-card">
          <div class="alias-editor-card__row">
            <label class="settings-field alias-editor-card__select">
              选择指令
              <select v-model="aliasEditCommand" class="ui-select">
                <optgroup v-for="group in aliasEditOptions" :key="group.category" :label="group.label">
                  <option v-for="def in group.items" :key="def.id" :value="def.primary_command">{{ def.name }} · /{{ def.primary_command }}</option>
                </optgroup>
              </select>
            </label>
            <UiBadge variant="outline">{{ aliasEditDefinition.template || aliasEditDefinition.render_mode }}</UiBadge>
          </div>

          <p class="alias-editor-card__desc">{{ aliasEditDefinition.usage }} · {{ aliasEditDefinition.description }}</p>

          <div class="alias-chip-row">
            <span v-for="alias in (aliasEditDefinition.preset_aliases || [])" :key="`p-${alias}`" class="alias-chip alias-chip--preset">{{ alias }}</span>
            <span v-for="alias in aliasDraft[aliasEditDefinition.primary_command] || []" :key="`c-${alias}`" class="alias-chip alias-chip--custom">
              {{ alias }}
              <button type="button" @click="removeAlias(aliasEditDefinition.primary_command, alias)">×</button>
            </span>
            <span v-if="!(aliasEditDefinition.preset_aliases || []).length && !(aliasDraft[aliasEditDefinition.primary_command] || []).length" class="alias-editor-card__empty">暂无别名，可在下方添加。</span>
          </div>

          <div class="alias-add-row">
            <input v-model="aliasInputs[aliasEditDefinition.primary_command]" class="ui-input" placeholder="添加自定义关键词，例如 cardx" @keyup.enter="addAlias(aliasEditDefinition.primary_command)" />
            <UiButton variant="outline" size="sm" @click="addAlias(aliasEditDefinition.primary_command)">添加</UiButton>
          </div>
        </div>
      </div>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  downloadCommandAliases,
  exportCommandAliases,
  getCommandAliases,
  getCommandDefinitions,
  importCommandAliases,
  parseCommand,
  renderParsedCommand,
  resetCommandAliases,
  updateCommandAliases,
} from '../api/client'
import type { CommandAliasConfig, CommandCategory, CommandDefinition, ParsedCommand, RenderTiming } from '../api/types'
import PageHeader from '../components/PageHeader.vue'
import SvgIcon from '../components/icons/SvgIcon.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

import type { IconName } from '../components/icons/SvgIcon.vue'

interface CategoryTab {
  key: CommandCategory
  label: string
  icon: IconName
  hint: string
}

const CATEGORY_TABS: CategoryTab[] = [
  { key: 'profile', label: '账号 / Profile', icon: 'users', hint: '账号绑定与个人资料图，使用 Sekai Profile API。' },
  { key: 'suite', label: 'Suite 数据', icon: 'database', hint: '需绑定账号且 Suite 公开，使用 Haruki Suite 公开 API 拉取真实数据。' },
  { key: 'deck', label: '组卡推荐', icon: 'resources', hint: '基于 Suite 卡组数据，由内置 sekai-calculator 推荐配队。' },
  { key: 'query', label: '查询 / 榜线', icon: 'search', hint: '查卡 / 查曲 / 活动 / 卡池 / 演唱会以及实时榜线、预测、查房、水表。' },
  { key: 'misc', label: '其它', icon: 'sparkle', hint: '帮助菜单等系统类指令。' },
]

const route = useRoute()
const router = useRouter()
const initialQuery = pickStringFromQuery(route.query.q)
const initialCategory = pickCategoryFromRoute(route.query.cat)

const input = ref(initialQuery || '/查卡 1204')
type CommandDefinitionsData = Awaited<ReturnType<typeof getCommandDefinitions>>

const definitions = ref<CommandDefinitionsData | null>(null)
const aliasConfig = ref<CommandAliasConfig | null>(null)
const parsed = ref<ParsedCommand | null>(null)
const imageUrl = ref('')
const loadingDefinitions = ref(false)
const parsing = ref(false)
const rendering = ref(false)
const savingAliases = ref(false)
const parseError = ref('')
const renderError = ref('')
const aliasMessage = ref('')
const aliasError = ref('')
const aliasDraft = reactive<Record<string, string[]>>({})
const aliasInputs = reactive<Record<string, string>>({})
const importInput = ref<HTMLInputElement | null>(null)
const timings = ref<RenderTiming>(emptyTiming())
const debugBinding = reactive({ region: '', gameId: '' })
const activeCategory = ref<CommandCategory>(initialCategory)
const aliasEditCommand = ref('')
const inputEl = ref<HTMLInputElement | null>(null)
const suggestionsOpen = ref(false)
const activeSuggestion = ref(0)

interface SuggestionItem {
  key: string
  text: string
  name: string
  kind: string
  category: CommandCategory
  score: number
}

const commandDefinitions = computed<CommandDefinition[]>(() => aliasConfig.value?.data ?? definitions.value?.data ?? [])
const definitionsByCategory = computed(() => {
  const groups: Record<string, CommandDefinition[]> = {}
  for (const tab of CATEGORY_TABS) groups[tab.key] = []
  for (const def of commandDefinitions.value) {
    const key = def.category || 'misc'
    if (!groups[key]) groups[key] = []
    groups[key].push(def)
  }
  return groups
})
const tabs = computed(() => CATEGORY_TABS.map(tab => ({ ...tab, count: definitionsByCategory.value[tab.key]?.length ?? 0 })))
const activeDefinitions = computed(() => definitionsByCategory.value[activeCategory.value] ?? [])
const activeTabHint = computed(() => CATEGORY_TABS.find(tab => tab.key === activeCategory.value)?.hint ?? '')
const aliasEditOptions = computed(() => CATEGORY_TABS
  .map(tab => ({ category: tab.key, label: tab.label, items: definitionsByCategory.value[tab.key] ?? [] }))
  .filter(group => group.items.length > 0))
const aliasEditDefinition = computed<CommandDefinition | null>(() => {
  if (!aliasEditCommand.value) return null
  return commandDefinitions.value.find(def => def.primary_command === aliasEditCommand.value) ?? null
})
const suggestions = computed<SuggestionItem[]>(() => buildSuggestions(input.value, commandDefinitions.value))
const regionOptions = computed(() => definitions.value?.regions ?? [])
const showDebugBindingPanel = computed(() => Boolean(parsed.value?.definition?.requires_binding))
const debugBindingHint = computed(() => {
  if (!parsed.value?.definition?.requires_binding) return ''
  if (!debugBinding.gameId) return '填写 UID 后再次解析或渲染，就会尝试拉取真实数据；留空时仍使用静态样例。'
  const regionLabel = regionOptions.value.find(region => region.key === debugBinding.region)?.label ?? debugBinding.region.toUpperCase()
  return `本次调试：${regionLabel} · UID ${debugBinding.gameId}`
})
const timingItems = computed(() => [
  { key: 'fonts', label: '字体加载', value: formatMs(timings.value.fonts_ms) },
  { key: 'images', label: '图片缓存', value: formatMs(timings.value.images_ms) },
  { key: 'satori', label: 'Satori', value: formatMs(timings.value.satori_ms) },
  { key: 'resvg', label: 'resvg', value: formatMs(timings.value.resvg_ms) },
  { key: 'cache', label: '图片命中', value: formatCacheHits() },
  { key: 'proxy', label: 'Go 代理', value: formatMs(timings.value.proxy_ms) },
  { key: 'network', label: '浏览器请求', value: formatMs(timings.value.network_ms) },
].filter(item => item.value !== '-'))

watch(activeCategory, value => {
  void router.replace({ query: { ...route.query, cat: value } })
  ensureAliasEditSelection()
})

watch(commandDefinitions, () => {
  ensureAliasEditSelection()
}, { deep: false })

function ensureAliasEditSelection() {
  const activeList = activeDefinitions.value
  if (activeList.length === 0) return
  const exists = activeList.some(def => def.primary_command === aliasEditCommand.value)
  if (!exists) {
    aliasEditCommand.value = activeList[0].primary_command
  }
}

onMounted(async () => {
  await loadAll()
  if (input.value) await runParse()
})
onBeforeUnmount(revokeImageUrl)

async function loadAll() {
  loadingDefinitions.value = true
  try {
    const [defs, aliases] = await Promise.all([getCommandDefinitions(), getCommandAliases()])
    definitions.value = defs
    aliasConfig.value = aliases
    if (!debugBinding.region) debugBinding.region = defs.regions?.[0]?.key || 'jp'
    applyAliasDraft(aliases)
  } catch (err) {
    parseError.value = err instanceof Error ? err.message : '加载指令定义失败。'
  } finally {
    loadingDefinitions.value = false
  }
}

async function runParse() {
  parsing.value = true
  parseError.value = ''
  parsed.value = null
  try {
    const response = await parseCommand(input.value, currentDebugBindingPayload())
    parsed.value = response.parsed
    syncDebugRegionFromParsed(response.parsed)
    if (response.parsed?.definition?.category) {
      activeCategory.value = response.parsed.definition.category
    }
    void router.replace({ query: { ...route.query, q: input.value, cat: activeCategory.value } })
  } catch (err) {
    parseError.value = err instanceof Error ? err.message : '解析失败。'
  } finally {
    parsing.value = false
  }
}

async function runRender() {
  renderError.value = ''
  rendering.value = true
  try {
    if (!parsed.value || parsed.value.raw_input !== input.value) {
      await runParse()
    }
    const result = await renderParsedCommand(input.value, parsed.value?.definition ? 800 : undefined, undefined, currentDebugBindingPayload())
    revokeImageUrl()
    imageUrl.value = result.url
    timings.value = result.timings
  } catch (err) {
    renderError.value = err instanceof Error ? err.message : '渲染失败。'
    timings.value = emptyTiming()
  } finally {
    rendering.value = false
  }
}

function setActiveCategory(category: CommandCategory) {
  activeCategory.value = category
}

function useExample(example: string) {
  input.value = example
  suggestionsOpen.value = false
  void runParse()
}

function definitionExamples(def: CommandDefinition): string[] {
  return Array.isArray(def?.examples) ? def.examples : []
}

function definitionAliases(def: CommandDefinition): string[] {
  const preset = Array.isArray(def?.preset_aliases) ? def.preset_aliases : []
  const custom = Array.isArray(def?.custom_aliases) ? def.custom_aliases : []
  return [...preset, ...custom]
}

function onInputChange() {
  activeSuggestion.value = 0
  suggestionsOpen.value = true
}

function onInputKeydown(event: KeyboardEvent) {
  if (!suggestionsOpen.value || suggestions.value.length === 0) return
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    activeSuggestion.value = (activeSuggestion.value + 1) % suggestions.value.length
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    activeSuggestion.value = (activeSuggestion.value - 1 + suggestions.value.length) % suggestions.value.length
  } else if (event.key === 'Escape') {
    suggestionsOpen.value = false
  } else if (event.key === 'Tab') {
    const picked = suggestions.value[activeSuggestion.value]
    if (picked) {
      event.preventDefault()
      applySuggestionText(picked)
    }
  }
}

function onInputEnter(event: KeyboardEvent) {
  if (suggestionsOpen.value && suggestions.value.length > 0 && activeSuggestion.value >= 0) {
    const picked = suggestions.value[activeSuggestion.value]
    const typed = input.value.trim().replace(/^\//, '').toLowerCase()
    if (picked && typed !== picked.text.replace(/^\//, '').toLowerCase()) {
      event.preventDefault()
      applySuggestion(picked)
      return
    }
  }
  suggestionsOpen.value = false
  void runParse()
}

function applySuggestion(item: SuggestionItem) {
  applySuggestionText(item)
  void runParse()
}

function applySuggestionText(item: SuggestionItem) {
  input.value = item.text
  suggestionsOpen.value = false
  requestAnimationFrame(() => {
    inputEl.value?.focus()
    inputEl.value?.setSelectionRange(item.text.length, item.text.length)
  })
}

function handleInputBlur(event: FocusEvent) {
  const next = event.relatedTarget as HTMLElement | null
  if (next && next.closest('.command-input-wrap')) return
  suggestionsOpen.value = false
}

function buildSuggestions(raw: string, defs: CommandDefinition[]): SuggestionItem[] {
  if (!defs.length) return []
  const trimmed = raw.trim()
  const stripped = trimmed.replace(/^\//, '')
  const firstToken = stripped.split(/\s+/, 1)[0] ?? ''
  const needle = firstToken.toLowerCase()
  const out: SuggestionItem[] = []
  const seen = new Set<string>()
  const push = (text: string, name: string, kind: string, category: CommandCategory, score: number) => {
    const normalized = text.startsWith('/') ? text : `/${text}`
    const key = normalized.toLowerCase()
    if (seen.has(key)) return
    seen.add(key)
    out.push({ key, text: normalized, name, kind, category, score })
  }

  for (const def of defs) {
    const category = (def.category || 'misc') as CommandCategory
    const primary = def.primary_command
    score(primary, def.name, '主指令', category, 120)
    for (const cmd of def.commands ?? []) {
      if (cmd === primary) continue
      score(cmd, def.name, '指令', category, 110)
    }
    for (const alias of def.preset_aliases ?? []) {
      score(alias, def.name, '预设别名', category, 95)
    }
    for (const alias of def.custom_aliases ?? []) {
      score(alias, def.name, '自定义', category, 100)
    }
    for (const example of def.examples ?? []) {
      score(example, def.name, '示例', category, 70)
    }
  }

  function score(candidate: string, name: string, kind: string, category: CommandCategory, baseScore: number) {
    const canon = candidate.startsWith('/') ? candidate.slice(1) : candidate
    const canonLower = canon.toLowerCase()
    let weight = baseScore
    if (!needle) {
      // Empty input: surface primary/preset only, keep list short.
      if (kind !== '主指令' && kind !== '预设别名') return
    } else if (canonLower.startsWith(needle)) {
      weight += 40
    } else if (canonLower.includes(needle)) {
      weight += 15
    } else {
      return
    }
    push(canon, name, kind, category, weight)
  }

  out.sort((a, b) => b.score - a.score || a.text.localeCompare(b.text))
  return out.slice(0, needle ? 10 : 8)
}

function currentDebugBindingPayload() {
  if (!showDebugBindingPanel.value && !debugBinding.gameId) return undefined
  return {
    region: debugBinding.region || parsed.value?.region || definitions.value?.regions?.[0]?.key || 'jp',
    game_id: debugBinding.gameId.trim(),
  }
}

function syncDebugRegionFromParsed(nextParsed: ParsedCommand) {
  if (!nextParsed?.definition?.requires_binding) return
  if (nextParsed.region) debugBinding.region = nextParsed.region
}

function applyAliasDraft(config: CommandAliasConfig) {
  Object.keys(aliasDraft).forEach(key => delete aliasDraft[key])
  for (const definition of config.data) {
    aliasDraft[definition.primary_command] = [...(config.custom[definition.primary_command] ?? [])]
    aliasInputs[definition.primary_command] = ''
  }
}

function addAlias(command: string) {
  const value = (aliasInputs[command] || '').trim().replace(/^\//, '')
  if (!value) return
  const list = aliasDraft[command] ?? (aliasDraft[command] = [])
  if (!list.some(item => item.toLowerCase() === value.toLowerCase())) {
    list.push(value)
  }
  aliasInputs[command] = ''
}

function removeAlias(command: string, alias: string) {
  aliasDraft[command] = (aliasDraft[command] ?? []).filter(item => item !== alias)
}

async function saveAliases() {
  savingAliases.value = true
  aliasError.value = ''
  aliasMessage.value = ''
  try {
    const response = await updateCommandAliases({ aliases: cloneAliases() })
    aliasConfig.value = response.config
    applyAliasDraft(response.config)
    aliasMessage.value = response.message
  } catch (err) {
    aliasError.value = err instanceof Error ? err.message : '保存关键词失败。'
  } finally {
    savingAliases.value = false
  }
}

async function resetAliasesToDefault() {
  savingAliases.value = true
  aliasError.value = ''
  aliasMessage.value = ''
  try {
    const response = await resetCommandAliases()
    aliasConfig.value = response.config
    applyAliasDraft(response.config)
    aliasMessage.value = response.message
  } catch (err) {
    aliasError.value = err instanceof Error ? err.message : '恢复默认失败。'
  } finally {
    savingAliases.value = false
  }
}

async function exportAliasesToFile() {
  const payload = await exportCommandAliases()
  downloadCommandAliases(payload)
}

function triggerImport() {
  importInput.value?.click()
}

async function handleImportFile(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  aliasError.value = ''
  aliasMessage.value = ''
  try {
    const text = await file.text()
    const payload = JSON.parse(text)
    const response = await importCommandAliases(payload)
    aliasConfig.value = response.config
    applyAliasDraft(response.config)
    aliasMessage.value = response.message
  } catch (err) {
    aliasError.value = err instanceof Error ? err.message : '导入失败，请检查 JSON 文件。'
  } finally {
    if (importInput.value) importInput.value.value = ''
  }
}

function cloneAliases() {
  const out: Record<string, string[]> = {}
  for (const [command, aliases] of Object.entries(aliasDraft)) {
    out[command] = aliases.map(alias => alias.trim()).filter(Boolean)
  }
  return out
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
    images_ms: null,
    satori_ms: null,
    resvg_ms: null,
    total_ms: null,
    proxy_ms: null,
    network_ms: null,
    size_bytes: null,
    image_total: null,
    image_remote: null,
    image_cache_hits: null,
    image_cache_misses: null,
    image_cache_errors: null,
  }
}

function formatMs(value: number | null) {
  return typeof value === 'number' ? `${value} ms` : '-'
}

function formatCacheHits() {
  const total = timings.value.image_remote
  if (typeof total !== 'number' || total <= 0) return '-'
  return `${timings.value.image_cache_hits ?? 0}/${total}`
}

function matchSourceLabel(source: string) {
  const labels: Record<string, string> = {
    primary: '原始指令',
    preset_alias: '官方预设别名',
    custom_alias: '用户自定义关键词',
  }
  return labels[source] ?? (source || '-')
}

function pickCategoryFromRoute(value: unknown): CommandCategory {
  const known = CATEGORY_TABS.map(tab => tab.key)
  const candidate = pickStringFromQuery(value)
  return known.includes(candidate as CommandCategory) ? (candidate as CommandCategory) : 'profile'
}

function pickStringFromQuery(value: unknown): string {
  if (typeof value === 'string') return value
  if (Array.isArray(value)) {
    const first = value[0]
    if (typeof first === 'string') return first
  }
  return ''
}
</script>

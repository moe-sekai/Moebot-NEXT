<template>
  <main class="page-stack">
    <PageHeader eyebrow="Command Parser" title="指令解析" subtitle="测试聊天指令、官方预设别名与自定义关键词，并查看解析结果和 Satori 渲染预览。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loadingDefinitions" @click="loadAll">刷新</UiButton>
      </template>
    </PageHeader>

    <UiAlert variant="warning" title="自定义关键词提示">
      {{ aliasConfig?.risk_message || definitions?.risk_message || '自定义关键词会影响聊天端指令触发，请避免使用常见聊天词。' }}
      <br />
      {{ aliasConfig?.restart_note || definitions?.restart_note || '聊天端通常需要重启后生效。' }}
    </UiAlert>

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
          <input v-model="input" class="ui-input command-input" placeholder="例如 /查卡 1204、/card 初音未来、/cn查曲 Tell Your World" @keyup.enter="runParse" />
          <UiButton :loading="parsing" @click="runParse">解析</UiButton>
          <UiButton variant="secondary" :loading="rendering" :disabled="!parsed?.definition" @click="runRender">渲染</UiButton>
        </div>

        <div v-if="exampleCommands.length" class="command-examples">
          <button v-for="example in exampleCommands" :key="example" class="alias-chip" type="button" @click="useExample(example)">{{ example }}</button>
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
            <div><dt>参数</dt><dd>{{ parsed.argument || '-' }}</dd></div>
            <div><dt>模板</dt><dd>{{ parsed.definition?.template || '-' }}</dd></div>
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

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>自定义关键词</h2>
          <p>只填写不带 / 的关键词；官方预设和原始指令受保护，恢复默认会清空自定义关键词。</p>
        </div>
        <div class="command-alias-actions">
          <UiButton variant="outline" size="sm" @click="exportAliasesToFile">导出</UiButton>
          <UiButton variant="outline" size="sm" @click="triggerImport">导入</UiButton>
          <UiButton variant="secondary" size="sm" :loading="savingAliases" @click="resetAliasesToDefault">恢复默认</UiButton>
          <UiButton size="sm" :loading="savingAliases" @click="saveAliases">保存关键词</UiButton>
          <input ref="importInput" class="sr-only-input" type="file" accept="application/json" @change="handleImportFile" />
        </div>
      </div>

      <UiAlert v-if="aliasMessage" variant="info">{{ aliasMessage }}</UiAlert>
      <UiAlert v-if="aliasError" variant="destructive" title="关键词配置错误">{{ aliasError }}</UiAlert>

      <div class="alias-editor-grid">
        <div v-for="definition in commandDefinitions" :key="definition.id" class="alias-editor-item">
          <div class="alias-editor-item__header">
            <div>
              <strong>{{ definition.name }}</strong>
              <span>{{ definition.primary_command }} · {{ definition.usage }}</span>
            </div>
            <UiBadge variant="outline">{{ definition.template }}</UiBadge>
          </div>
          <div class="alias-chip-row">
            <span v-for="alias in definition.preset_aliases" :key="alias" class="alias-chip alias-chip--preset">{{ alias }}</span>
            <span v-for="alias in aliasDraft[definition.primary_command] || []" :key="alias" class="alias-chip alias-chip--custom">
              {{ alias }}
              <button type="button" @click="removeAlias(definition.primary_command, alias)">×</button>
            </span>
          </div>
          <div class="alias-add-row">
            <input v-model="aliasInputs[definition.primary_command]" class="ui-input" placeholder="添加自定义关键词" @keyup.enter="addAlias(definition.primary_command)" />
            <UiButton variant="outline" size="sm" @click="addAlias(definition.primary_command)">添加</UiButton>
          </div>
        </div>
      </div>
    </UiCard>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>功能列表</h2>
          <p>每个 Satori 渲染功能对应的解析方法、预设别名和示例。</p>
        </div>
        <UiBadge variant="secondary">{{ commandDefinitions.length }} 项</UiBadge>
      </div>
      <div class="command-definition-grid">
        <button v-for="definition in commandDefinitions" :key="definition.id" type="button" class="command-definition-card" @click="useExample(definition.examples?.[0] || definition.usage)">
          <span class="command-definition-card__top">
            <strong>{{ definition.name }}</strong>
            <UiBadge variant="outline">{{ definition.render_mode }}</UiBadge>
          </span>
          <span class="command-definition-card__usage">{{ definition.usage }}</span>
          <span class="command-definition-card__desc">{{ definition.description }}</span>
          <span class="alias-chip-row">
            <span v-for="alias in [...definition.preset_aliases, ...definition.custom_aliases].slice(0, 6)" :key="alias" class="alias-chip">{{ alias }}</span>
          </span>
        </button>
      </div>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
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
import type { CommandAliasConfig, CommandDefinition, ParsedCommand, RenderTiming } from '../api/types'
import PageHeader from '../components/PageHeader.vue'
import SvgIcon from '../components/icons/SvgIcon.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

const route = useRoute()
const router = useRouter()
const initialQuery = route.query.q
const input = ref(typeof initialQuery === 'string' ? initialQuery : Array.isArray(initialQuery) && typeof initialQuery[0] === 'string' ? initialQuery[0] : '/查卡 1204')
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

const commandDefinitions = computed<CommandDefinition[]>(() => aliasConfig.value?.data ?? definitions.value?.data ?? [])
const exampleCommands = computed(() => commandDefinitions.value.flatMap(definition => definition.examples ?? []).slice(0, 12))
const timingItems = computed(() => [
  { key: 'fonts', label: '字体加载', value: formatMs(timings.value.fonts_ms) },
  { key: 'satori', label: 'Satori', value: formatMs(timings.value.satori_ms) },
  { key: 'resvg', label: 'resvg', value: formatMs(timings.value.resvg_ms) },
  { key: 'proxy', label: 'Go 代理', value: formatMs(timings.value.proxy_ms) },
  { key: 'network', label: '浏览器请求', value: formatMs(timings.value.network_ms) },
].filter(item => item.value !== '-'))

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
    const response = await parseCommand(input.value)
    parsed.value = response.parsed
    void router.replace({ query: { ...route.query, q: input.value } })
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
    const result = await renderParsedCommand(input.value, parsed.value?.definition ? 800 : undefined)
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

function useExample(example: string) {
  input.value = example
  void runParse()
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
  return { fonts_ms: null, satori_ms: null, resvg_ms: null, total_ms: null, proxy_ms: null, network_ms: null, size_bytes: null }
}

function formatMs(value: number | null) {
  return typeof value === 'number' ? `${value} ms` : '-'
}

function matchSourceLabel(source: string) {
  const labels: Record<string, string> = {
    primary: '原始指令',
    preset_alias: '官方预设别名',
    custom_alias: '用户自定义关键词',
  }
  return labels[source] ?? (source || '-')
}
</script>

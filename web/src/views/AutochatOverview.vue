<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="Plugin · AutoChat"
      title="AutoChat 概览与提供商"
      subtitle="管理已接入的 LLM 提供商，并为各类任务指派模型。"
    >
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="reload">刷新</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="操作失败">{{ error }}</UiAlert>
    <UiAlert v-if="successMsg" variant="info" title="已保存">{{ successMsg }}</UiAlert>

    <!-- ===== 概览卡片 ===== -->
    <UiCard v-if="overview">
      <h2 class="card-title">运行概览</h2>
      <div class="stat-grid">
        <div class="stat">
          <div class="stat-label">首选模型</div>
          <div class="stat-value mono">{{ overview.primary_model || '未设置' }}</div>
          <div class="stat-sub">候选 {{ overview.models_count }} 个</div>
        </div>
        <div class="stat">
          <div class="stat-label">已接入提供商</div>
          <div class="stat-value">{{ model?.provider_list.length || 0 }}</div>
          <div class="stat-sub">默认阈值 {{ overview.willing_threshold.toFixed(2) }} · 单群覆盖 {{ overview.group_overrides }}</div>
        </div>
        <div class="stat">
          <div class="stat-label">今日 Token</div>
          <div class="stat-value">{{ overview.token_stats_today.total.toLocaleString() }}</div>
          <div class="stat-sub">{{ overview.token_stats_today.requests }} 次</div>
        </div>
        <div class="stat">
          <div class="stat-label">近 7 日 Token</div>
          <div class="stat-value">{{ overview.token_stats_7days.total.toLocaleString() }}</div>
          <div class="stat-sub">{{ overview.token_stats_7days.requests }} 次</div>
        </div>
      </div>
    </UiCard>

    <template v-if="model">
      <!-- ===== 已接入提供商列表 ===== -->
      <UiCard>
        <div class="card-heading">
          <div>
            <h2>已接入提供商</h2>
            <p>每个提供商有唯一 Name；模型用 <code>name:model</code> 形式被引用，例如 <code>openai:gpt-4o</code>。</p>
          </div>
          <div class="actions">
            <UiButton variant="default" size="sm" @click="openAddDialog">＋ 添加提供商</UiButton>
          </div>
        </div>

        <div v-if="!model.provider_list.length" class="empty">
          暂无已接入提供商。点击右上角"添加提供商"开始配置 OpenAI 兼容 / Anthropic endpoint。
        </div>

        <div class="provider-grid">
          <div v-for="(prov, idx) in model.provider_list" :key="prov.name + idx" class="provider-card">
            <div class="provider-card-head">
              <div class="provider-card-title">
                <span class="provider-name">{{ prov.name }}</span>
                <span class="provider-type" :class="prov.type === 'anthropic' ? 'tag-anthropic' : 'tag-openai'">
                  {{ prov.type === 'anthropic' ? 'Anthropic' : 'OpenAI 兼容' }}
                </span>
              </div>
              <div class="provider-card-actions">
                <button class="icon-btn" title="编辑" @click="openEditDialog(idx)">编辑</button>
                <button class="icon-btn icon-btn--danger" title="删除" @click="deleteProvider(idx)">删除</button>
              </div>
            </div>
            <div class="provider-meta">
              <div class="meta-row"><span class="meta-key">Base URL</span><span class="mono">{{ prov.base_url || '—' }}</span></div>
              <div class="meta-row"><span class="meta-key">API Key</span><span class="mono">{{ maskKey(prov.api_key) }}</span></div>
              <div class="meta-row"><span class="meta-key">超时</span><span>{{ prov.timeout }} s</span></div>
              <div class="meta-row" v-if="prov.type === 'anthropic'"><span class="meta-key">version</span><span>{{ prov.anthropic_version || '—' }}</span></div>
              <div class="meta-row"><span class="meta-key">已选模型</span><span>{{ countModelsForProvider(prov.name) }} 个</span></div>
            </div>
          </div>
        </div>
      </UiCard>

      <!-- ===== 任务模型指派 ===== -->
      <UiCard>
        <SectionHeader title="任务模型指派" desc="所有任务直接选择已接入提供商和模型，不再重复填写 endpoint。" />
        <div class="form-grid">
          <Field label="主对话模型">
            <select v-model="primaryModel">
              <option value="">— 请选择 —</option>
              <option v-for="m in allModels" :key="m" :value="m">{{ m }}</option>
            </select>
          </Field>
          <Field label="max_tokens"><input v-model.number="model.llm.max_tokens" type="number" /></Field>
          <Field label="对话超时 (秒)"><input v-model.number="model.llm.timeout" type="number" /></Field>
          <Field label="启用 reasoning">
            <label class="check"><input type="checkbox" v-model="model.llm.reasoning" /> 启用扩展思考</label>
          </Field>
        </div>
        <div class="form-grid" style="margin-top:12px">
          <Field label="视觉/图片描述模型">
            <select v-model="model.image_caption.model">
              <option value="">— 不启用 —</option>
              <option v-for="m in allModels" :key="m" :value="m">{{ m }}</option>
            </select>
          </Field>
          <Field label="RAG 总结模型">
            <select v-model="model.rag_summary.model">
              <option value="">— 不启用 —</option>
              <option v-for="m in allModels" :key="m" :value="m">{{ m }}</option>
            </select>
          </Field>
        </div>
      </UiCard>

      <!-- ===== 嵌入 / 重排 ===== -->
      <UiCard>
        <SectionHeader title="嵌入 (Embedding)" desc="选择哪个已接入提供商提供 /embeddings 端点，并填入模型名。" />
        <div class="form-grid">
          <Field label="启用">
            <label class="check"><input type="checkbox" v-model="model.embedding.enabled" /> 启用</label>
          </Field>
          <Field label="使用提供商">
            <select v-model="model.embedding.provider">
              <option value="">— 自定义 endpoint —</option>
              <option v-for="p in openaiCompatProviders" :key="p.name" :value="p.name">{{ p.name }}</option>
            </select>
          </Field>
          <Field label="模型名"><input v-model="model.embedding.model" type="text" placeholder="text-embedding-3-small" /></Field>
          <Field label="维度"><input v-model.number="model.embedding.dimensions" type="number" /></Field>
          <Field label="超时 (秒)"><input v-model.number="model.embedding.timeout" type="number" /></Field>
        </div>
        <template v-if="!model.embedding.provider">
          <div class="form-grid" style="margin-top:8px">
            <Field label="自定义 Base URL"><input v-model="model.embedding.base_url" type="text" placeholder="https://api.siliconflow.cn/v1" /></Field>
            <Field label="自定义 API Key"><input v-model="model.embedding.api_key" type="text" /></Field>
          </div>
        </template>
        <div class="btn-row">
          <UiButton variant="outline" size="sm" :loading="testingEmbedding" @click="testEmbedding">测试可用性</UiButton>
          <span v-if="embeddingTestResult" class="test-result" :class="embeddingTestResult.ok ? 'ok' : 'fail'">
            {{ embeddingTestResult.ok ? embeddingTestResult.message : embeddingTestResult.error }}
          </span>
        </div>
      </UiCard>

      <UiCard>
        <SectionHeader title="重排 (Rerank)" desc="重排接口使用 /rerank 端点；threshold 越高越严格。" />
        <div class="form-grid">
          <Field label="启用">
            <label class="check"><input type="checkbox" v-model="model.rerank.enabled" /> 启用</label>
          </Field>
          <Field label="使用提供商">
            <select v-model="model.rerank.provider">
              <option value="">— 自定义 endpoint —</option>
              <option v-for="p in openaiCompatProviders" :key="p.name" :value="p.name">{{ p.name }}</option>
            </select>
          </Field>
          <Field label="模型名"><input v-model="model.rerank.model" type="text" /></Field>
          <Field label="threshold"><input v-model.number="model.rerank.threshold" type="number" step="0.05" /></Field>
          <Field label="超时 (秒)"><input v-model.number="model.rerank.timeout" type="number" /></Field>
        </div>
        <template v-if="!model.rerank.provider">
          <div class="form-grid" style="margin-top:8px">
            <Field label="自定义 Base URL"><input v-model="model.rerank.base_url" type="text" /></Field>
            <Field label="自定义 API Key"><input v-model="model.rerank.api_key" type="text" /></Field>
          </div>
        </template>
        <div class="btn-row">
          <UiButton variant="outline" size="sm" :loading="testingRerank" @click="testRerank">测试可用性</UiButton>
          <span v-if="rerankTestResult" class="test-result" :class="rerankTestResult.ok ? 'ok' : 'fail'">
            {{ rerankTestResult.ok ? rerankTestResult.message : rerankTestResult.error }}
          </span>
        </div>
      </UiCard>

      <!-- ===== 管理员指令说明 ===== -->
      <UiCard>
        <SectionHeader
          title="管理员指令"
          desc="以下指令仅限「超级管理员（SuperUser）」使用。请前往 核心设置 → 超级管理员 配置 QQ 名单。" />
        <div class="cmd-grid">
          <div class="cmd-row">
            <code>/开启聊天</code>
            <span>把当前群加入 <code>/chat</code> 白名单（手动调用 LLM 模式）。</span>
          </div>
          <div class="cmd-row">
            <code>/关闭聊天</code>
            <span>从 <code>/chat</code> 白名单移除当前群。</span>
          </div>
          <div class="cmd-row">
            <code>/开启autochat</code>
            <span>把当前群加入主动发言白名单（按发言倾向阈值自动插嘴）。</span>
          </div>
          <div class="cmd-row">
            <code>/关闭autochat</code>
            <span>从主动发言白名单移除当前群。</span>
          </div>
          <div class="cmd-row">
            <code>/模型 &lt;model&gt;</code>
            <span>切换当前群使用的模型（管理员或 SuperUser 均可）。<code>/模型</code> 单独发查看当前。</span>
          </div>
          <div class="cmd-row">
            <code>/模型列表</code>
            <span>列出已配置的所有候选模型。</span>
          </div>
          <div class="cmd-row">
            <code>/消耗统计</code>
            <span>查看 24 小时 / 7 日 token 用量与请求次数。</span>
          </div>
          <div class="cmd-row">
            <code>/查询记忆</code>
            <span>从向量库检索本群历史画像与对话总结。</span>
          </div>
        </div>
      </UiCard>

      <UiCard>
        <SectionHeader title="向量库 (sqlite-vec)" desc="维度变化会清空向量虚表（原文保留）；topK 控制单次召回条数。" />
        <div class="form-grid">
          <Field label="启用">
            <label class="check"><input type="checkbox" v-model="model.vector.enabled" /> 启用</label>
          </Field>
          <Field label="维度"><input v-model.number="model.vector.dimensions" type="number" /></Field>
          <Field label="top_k"><input v-model.number="model.vector.top_k" type="number" /></Field>
        </div>
      </UiCard>
    </template>

    <!-- ===== Sticky 保存栏 ===== -->
    <div class="sticky-save" v-if="model">
      <div class="sticky-inner">
        <span v-if="dirty" class="dirty-hint">有未保存的更改</span>
        <UiButton variant="default" size="sm" :loading="saving" :disabled="!dirty" @click="save">保存所有配置</UiButton>
      </div>
    </div>

    <!-- ===== 添加/编辑 提供商 对话框 ===== -->
    <div v-if="dialog.open" class="dialog-mask" @click.self="closeDialog">
      <div class="dialog">
        <div class="dialog-head">
          <h3>{{ dialog.editIdx === null ? '添加提供商' : '编辑提供商' }}</h3>
          <button class="icon-btn" @click="closeDialog">关闭</button>
        </div>

        <div class="dialog-body">
          <div class="type-tabs">
            <button class="type-tab" :class="{ active: dialog.draft.type === 'openai' }" @click="dialog.draft.type = 'openai'">OpenAI 兼容</button>
            <button class="type-tab" :class="{ active: dialog.draft.type === 'anthropic' }" @click="dialog.draft.type = 'anthropic'">Anthropic</button>
          </div>

          <div class="form-grid">
            <Field label="名称（唯一标识）" hint="模型用 name:model 形式引用，建议英文小写">
              <input v-model="dialog.draft.name" type="text" placeholder="openai / siliconflow / claude" />
            </Field>
            <Field label="超时 (秒)"><input v-model.number="dialog.draft.timeout" type="number" /></Field>
            <Field label="Base URL" full>
              <input v-model="dialog.draft.base_url" type="text" :placeholder="dialog.draft.type === 'anthropic' ? 'https://api.anthropic.com' : 'https://api.openai.com/v1'" />
            </Field>
            <Field label="API Key" full><input v-model="dialog.draft.api_key" type="text" /></Field>
            <Field v-if="dialog.draft.type === 'anthropic'" label="anthropic-version" full>
              <input v-model="dialog.draft.anthropic_version" type="text" placeholder="2023-06-01" />
            </Field>
          </div>

          <div class="btn-row">
            <UiButton variant="outline" size="sm" :loading="dialog.testing" @click="testDraft">测试连通性</UiButton>
            <UiButton variant="outline" size="sm" :loading="dialog.fetching" @click="fetchDraftModels">获取模型列表</UiButton>
            <span v-if="dialog.testResult" class="test-result" :class="dialog.testResult.ok ? 'ok' : 'fail'">{{ dialog.testResult.ok ? dialog.testResult.message : dialog.testResult.error }}</span>
          </div>

          <div v-if="dialog.modelOptions.length" class="model-list">
            <div class="model-list-header">
              可用模型（{{ dialog.modelOptions.length }} 个） ·
              已选 {{ dialog.selectedDraftModels.size }} 个
              <button class="link-btn" @click="selectAllDraftModels">全选</button>
              <button class="link-btn" @click="clearDraftModels">清空</button>
            </div>
            <div class="model-chips">
              <label v-for="m in dialog.modelOptions" :key="m" class="model-chip" :class="{ selected: dialog.selectedDraftModels.has(m) }">
                <input type="checkbox" :checked="dialog.selectedDraftModels.has(m)" @change="toggleDraftModel(m)" />
                {{ m }}
              </label>
            </div>
          </div>

          <Field label="自定义模型（每行一个，可选）" full hint="未在列表中的模型可手填，会带上 name: 前缀">
            <textarea v-model="dialog.customModelsText" rows="2" placeholder="gpt-4o&#10;o1-preview" />
          </Field>
        </div>

        <div class="dialog-foot">
          <UiButton variant="outline" size="sm" @click="closeDialog">取消</UiButton>
          <UiButton variant="default" size="sm" @click="confirmDialog">{{ dialog.editIdx === null ? '添加' : '保存' }}</UiButton>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  getAutochatOverview,
  getAutochatProviders,
  updateAutochatProviders,
  testAutochatProvider,
  listAutochatModels,
  type AutochatOverview,
  type AutochatProviders,
  type AutochatProvider,
  type TestProviderResult,
} from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import SectionHeader from '../components/autochat/AutochatSectionHeader.vue'
import Field from '../components/autochat/AutochatField.vue'

const overview = ref<AutochatOverview | null>(null)
const model = ref<AutochatProviders | null>(null)
const original = ref('')
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const successMsg = ref('')

// ---------- Provider Add/Edit Dialog ----------
interface DialogState {
  open: boolean
  editIdx: number | null
  draft: AutochatProvider
  testing: boolean
  fetching: boolean
  testResult: TestProviderResult | null
  modelOptions: string[]            // 完整 id 字符串（含 name: 前缀）
  selectedDraftModels: Set<string>  // 用户在弹窗里勾选的模型
  customModelsText: string
}

const dialog = reactive<DialogState>({
  open: false,
  editIdx: null,
  draft: emptyProvider(),
  testing: false,
  fetching: false,
  testResult: null,
  modelOptions: [],
  selectedDraftModels: new Set<string>(),
  customModelsText: '',
})

function emptyProvider(): AutochatProvider {
  return { name: '', type: 'openai', base_url: '', api_key: '', timeout: 60, anthropic_version: '2023-06-01' }
}

// ---------- Embedding / Rerank availability test ----------
const testingEmbedding = ref(false)
const testingRerank = ref(false)
const embeddingTestResult = ref<TestProviderResult | null>(null)
const rerankTestResult = ref<TestProviderResult | null>(null)

// 解析 embedding/rerank 的有效端点：优先 provider 引用；否则用独立 base_url/api_key。
function resolveEndpoint(provName: string, fallbackBaseURL: string, fallbackKey: string) {
  if (provName) {
    const p = (model.value?.provider_list || []).find(x => x.name === provName)
    if (p) return { base_url: p.base_url, api_key: p.api_key, timeout: p.timeout }
  }
  return { base_url: fallbackBaseURL, api_key: fallbackKey, timeout: 15 }
}

async function testEmbedding() {
  if (!model.value) return
  testingEmbedding.value = true
  embeddingTestResult.value = null
  try {
    const ep = resolveEndpoint(
      model.value.embedding.provider,
      model.value.embedding.base_url,
      model.value.embedding.api_key,
    )
    embeddingTestResult.value = await testAutochatProvider({
      type: 'openai', // /models 端点对 OpenAI 兼容供应商通用
      base_url: ep.base_url,
      api_key: ep.api_key,
      timeout: model.value.embedding.timeout || ep.timeout,
    })
  } catch (e) {
    embeddingTestResult.value = { ok: false, error: e instanceof Error ? e.message : String(e) }
  } finally {
    testingEmbedding.value = false
  }
}

async function testRerank() {
  if (!model.value) return
  testingRerank.value = true
  rerankTestResult.value = null
  try {
    const ep = resolveEndpoint(
      model.value.rerank.provider,
      model.value.rerank.base_url,
      model.value.rerank.api_key,
    )
    rerankTestResult.value = await testAutochatProvider({
      type: 'openai',
      base_url: ep.base_url,
      api_key: ep.api_key,
      timeout: model.value.rerank.timeout || ep.timeout,
    })
  } catch (e) {
    rerankTestResult.value = { ok: false, error: e instanceof Error ? e.message : String(e) }
  } finally {
    testingRerank.value = false
  }
}

const primaryModel = computed({
  get: () => model.value?.llm.models?.[0] || '',
  set: (v: string) => {
    if (!model.value) return
    const rest = model.value.llm.models.filter(m => m !== v)
    model.value.llm.models = v ? [v, ...rest] : rest
  },
})

const allModels = computed(() => model.value?.llm.models || [])

const openaiCompatProviders = computed(() =>
  (model.value?.provider_list || []).filter(p => p.type === 'openai'),
)

const dirty = computed(() =>
  model.value !== null && JSON.stringify(model.value) !== original.value,
)

onMounted(reload)
watch(model, () => { successMsg.value = '' }, { deep: true })

// normalizeProviders 给 payload 补齐所有字段默认值，避免后端旧二进制 / 字段缺失
// 时 Vue 模板里 model.provider_list.length / model.embedding.* 抛 TypeError
// 导致整张子卡片静默不渲染（典型症状：刚把插件由关闭切到开启后，部分卡片消失）。
function normalizeProviders(pv: Partial<AutochatProviders> | null | undefined): AutochatProviders {
  const p = (pv || {}) as Partial<AutochatProviders>
  return {
    provider_list: Array.isArray(p.provider_list) ? p.provider_list : [],
    llm: {
      models: Array.isArray(p.llm?.models) ? p.llm!.models : [],
      max_tokens: p.llm?.max_tokens ?? 2048,
      reasoning: !!p.llm?.reasoning,
      timeout: p.llm?.timeout ?? 120,
    },
    embedding: {
      enabled: !!p.embedding?.enabled,
      provider: p.embedding?.provider ?? '',
      base_url: p.embedding?.base_url ?? '',
      api_key: p.embedding?.api_key ?? '',
      model: p.embedding?.model ?? '',
      dimensions: p.embedding?.dimensions ?? 1536,
      timeout: p.embedding?.timeout ?? 30,
    },
    rerank: {
      enabled: !!p.rerank?.enabled,
      provider: p.rerank?.provider ?? '',
      base_url: p.rerank?.base_url ?? '',
      api_key: p.rerank?.api_key ?? '',
      model: p.rerank?.model ?? '',
      threshold: p.rerank?.threshold ?? 0.3,
      timeout: p.rerank?.timeout ?? 15,
    },
    vector: {
      enabled: !!p.vector?.enabled,
      dimensions: p.vector?.dimensions ?? 1536,
      top_k: p.vector?.top_k ?? 5,
    },
    image_caption: {
      enabled: !!p.image_caption?.enabled,
      model: p.image_caption?.model ?? '',
      timeout: p.image_caption?.timeout ?? 20,
      max_tokens: p.image_caption?.max_tokens ?? 80,
      prompt: p.image_caption?.prompt ?? '',
    },
    rag_summary: {
      enabled: !!p.rag_summary?.enabled,
      model: p.rag_summary?.model ?? '',
      timeout: p.rag_summary?.timeout ?? 30,
      max_tokens: p.rag_summary?.max_tokens ?? 256,
    },
  }
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    // 拆开两个请求：overview 失败时不影响下方表单加载；
    // 这样在插件刚由关闭切到开启时，至少配置面板可以正常显示。
    const ovP = getAutochatOverview().catch(err => {
      console.warn('[autochat] overview load failed', err)
      return null
    })
    const pvP = getAutochatProviders()
    const [ov, pv] = await Promise.all([ovP, pvP])
    overview.value = ov
    const normalized = normalizeProviders(pv)
    model.value = normalized
    original.value = JSON.stringify(normalized)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败。'
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!model.value) return
  saving.value = true
  error.value = ''
  successMsg.value = ''
  try {
    const data = await updateAutochatProviders(model.value)
    model.value = data
    original.value = JSON.stringify(data)
    successMsg.value = '所有配置已写回 autochat.yml 并即时生效。'
    overview.value = await getAutochatOverview()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存失败。'
  } finally {
    saving.value = false
  }
}

// ---- Provider list ops ----
function maskKey(k: string) {
  if (!k) return '—'
  if (k.length <= 8) return '****'
  return k.slice(0, 4) + '****' + k.slice(-4)
}

function countModelsForProvider(name: string) {
  if (!model.value) return 0
  const prefix = name + ':'
  return model.value.llm.models.filter(m => m.startsWith(prefix)).length
}

function deleteProvider(idx: number) {
  if (!model.value) return
  const prov = model.value.provider_list[idx]
  if (!prov) return
  if (!window.confirm(`确认删除提供商 "${prov.name}"？\n它的所有模型 (${countModelsForProvider(prov.name)} 个) 也会被移除。`)) return
  const prefix = prov.name + ':'
  model.value.llm.models = model.value.llm.models.filter(m => !m.startsWith(prefix))
  // 清空依赖该 provider 的引用
  if (model.value.embedding.provider === prov.name) model.value.embedding.provider = ''
  if (model.value.rerank.provider === prov.name) model.value.rerank.provider = ''
  model.value.provider_list.splice(idx, 1)
}

// ---- Dialog ----
function openAddDialog() {
  dialog.editIdx = null
  dialog.draft = emptyProvider()
  resetDialogTransient()
  dialog.open = true
}

function openEditDialog(idx: number) {
  if (!model.value) return
  const prov = model.value.provider_list[idx]
  dialog.editIdx = idx
  dialog.draft = JSON.parse(JSON.stringify(prov))
  resetDialogTransient()
  // 预填已选模型
  const prefix = prov.name + ':'
  const existing = (model.value.llm.models || []).filter(m => m.startsWith(prefix))
  existing.forEach(m => dialog.selectedDraftModels.add(m))
  dialog.open = true
}

function resetDialogTransient() {
  dialog.testResult = null
  dialog.modelOptions = []
  dialog.selectedDraftModels = new Set<string>()
  dialog.customModelsText = ''
}

function closeDialog() { dialog.open = false }

async function testDraft() {
  dialog.testing = true; dialog.testResult = null
  try {
    dialog.testResult = await testAutochatProvider({
      type: dialog.draft.type,
      base_url: dialog.draft.base_url,
      api_key: dialog.draft.api_key,
      timeout: dialog.draft.timeout,
    })
  } catch (e) {
    dialog.testResult = { ok: false, error: e instanceof Error ? e.message : String(e) }
  } finally { dialog.testing = false }
}

async function fetchDraftModels() {
  if (!dialog.draft.name.trim()) {
    error.value = '请先填写提供商名称（用作模型前缀）。'
    return
  }
  dialog.fetching = true
  try {
    const res = await listAutochatModels({
      type: dialog.draft.type,
      base_url: dialog.draft.base_url,
      api_key: dialog.draft.api_key,
      timeout: dialog.draft.timeout,
      prefix: dialog.draft.name,
    })
    dialog.modelOptions = res.models.map(m => m.id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally { dialog.fetching = false }
}

function toggleDraftModel(id: string) {
  if (dialog.selectedDraftModels.has(id)) dialog.selectedDraftModels.delete(id)
  else dialog.selectedDraftModels.add(id)
}
function selectAllDraftModels() { dialog.modelOptions.forEach(m => dialog.selectedDraftModels.add(m)) }
function clearDraftModels() { dialog.selectedDraftModels.clear() }

function confirmDialog() {
  if (!model.value) return
  const draft = dialog.draft
  const name = draft.name.trim()
  if (!name) { error.value = '提供商名称不能为空。'; return }
  if (!/^[a-zA-Z0-9_\-]+$/.test(name)) { error.value = '名称仅允许英文/数字/_/-。'; return }

  // 检测重名（编辑时排除自身）
  const dupIdx = model.value.provider_list.findIndex(p => p.name === name)
  if (dupIdx !== -1 && dupIdx !== dialog.editIdx) {
    error.value = `名称 "${name}" 已存在。`
    return
  }
  draft.name = name

  // 处理模型：旧前缀 → 新前缀（编辑时改名）
  const oldName = dialog.editIdx !== null ? model.value.provider_list[dialog.editIdx].name : null
  if (oldName && oldName !== name) {
    const oldPrefix = oldName + ':'
    model.value.llm.models = model.value.llm.models.map(m =>
      m.startsWith(oldPrefix) ? name + ':' + m.slice(oldPrefix.length) : m,
    )
    if (model.value.embedding.provider === oldName) model.value.embedding.provider = name
    if (model.value.rerank.provider === oldName) model.value.rerank.provider = name
  }

  // 写入 / 更新 provider
  if (dialog.editIdx === null) {
    model.value.provider_list.push({ ...draft })
  } else {
    model.value.provider_list[dialog.editIdx] = { ...draft }
  }

  // 合并 selected + custom 模型，并把当前 provider 的旧条目替换掉
  const prefix = name + ':'
  const customs = dialog.customModelsText.split('\n').map(s => s.trim()).filter(Boolean)
    .map(m => m.includes(':') ? m : prefix + m)
  const newSet = new Set<string>([...dialog.selectedDraftModels, ...customs])
  const others = model.value.llm.models.filter(m => !m.startsWith(prefix))
  model.value.llm.models = [...others, ...newSet]

  error.value = ''
  closeDialog()
}
</script>

<style scoped>
.card-title { margin: 0 0 12px 0; font-size: 16px; font-weight: 700; color: var(--foreground); }
.card-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; margin-bottom: 14px; }
.card-heading h2 { margin: 0 0 4px 0; font-size: 16px; font-weight: 700; color: var(--foreground); }
.card-heading p { margin: 0; font-size: 12px; color: var(--muted-foreground); line-height: 1.6; }
.card-heading .actions { display: flex; gap: 8px; flex-shrink: 0; }

.stat-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 12px; }
.stat { border: 1px solid var(--border); border-radius: 16px; padding: 12px 14px; background: rgba(255, 255, 255, 0.7); }
.stat-label { font-size: 12px; color: var(--muted-foreground); }
.stat-value { font-size: 22px; font-weight: 700; margin-top: 4px; color: var(--foreground); }
.stat-value.mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 14px; }
.stat-sub { font-size: 11px; color: var(--muted-foreground); margin-top: 4px; }

.empty { padding: 24px 0; color: var(--muted-foreground); font-size: 13px; text-align: center; }

.provider-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 12px; }
.provider-card {
  border: 1px solid var(--border); border-radius: 16px;
  background: rgba(255, 255, 255, 0.75);
  padding: 14px 16px; display: flex; flex-direction: column; gap: 10px;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.provider-card:hover { border-color: var(--input); box-shadow: 0 2px 8px rgba(165, 180, 252, 0.15); }
.provider-card-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 8px; }
.provider-card-title { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.provider-name { font-weight: 700; color: var(--foreground); font-size: 15px; }
.provider-type { font-size: 11px; padding: 2px 8px; border-radius: 999px; font-weight: 600; }
.tag-openai { background: rgba(80, 200, 120, 0.18); color: #1e8a4a; }
.tag-anthropic { background: rgba(255, 140, 80, 0.18); color: #b8590f; }
.provider-card-actions { display: flex; gap: 4px; flex-shrink: 0; }
.icon-btn {
  border: 1px solid var(--border); background: rgba(255,255,255,0.85); color: var(--foreground);
  border-radius: 999px; padding: 4px 10px; font-size: 12px; cursor: pointer;
  transition: all 0.15s;
}
.icon-btn:hover { border-color: var(--input); background: #fff; }
.icon-btn--danger { color: #b04040; }
.icon-btn--danger:hover { background: rgba(220, 80, 80, 0.1); border-color: rgba(220, 80, 80, 0.4); }

.provider-meta { display: flex; flex-direction: column; gap: 4px; }
.meta-row { display: flex; justify-content: space-between; gap: 12px; font-size: 12px; }
.meta-key { color: var(--muted-foreground); flex-shrink: 0; }
.meta-row .mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; color: var(--foreground); word-break: break-all; text-align: right; }

.form-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 14px; }
.check { display: inline-flex; align-items: center; gap: 6px; cursor: pointer; color: var(--foreground); font-size: 13px; }

.btn-row { display: flex; align-items: center; gap: 10px; margin-top: 12px; flex-wrap: wrap; }
.test-result { font-size: 13px; font-weight: 600; }
.test-result.ok { color: #1e8a4a; }
.test-result.fail { color: #b04040; }

.model-list { margin-top: 12px; padding: 12px; border: 1px solid var(--border); border-radius: 16px; background: rgba(255,255,255,0.5); }
.model-list-header { font-size: 13px; font-weight: 600; margin-bottom: 8px; color: var(--foreground); display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.link-btn { background: none; border: none; color: var(--primary, #ff78b7); font-size: 12px; cursor: pointer; font-weight: 600; }
.model-chips { display: flex; flex-wrap: wrap; gap: 6px; max-height: 200px; overflow-y: auto; }
.model-chip {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 4px 10px; border-radius: 999px; font-size: 12px; cursor: pointer;
  border: 1px solid var(--border); background: rgba(255,255,255,0.8); color: var(--foreground);
  transition: all 0.15s; user-select: none;
}
.model-chip input { display: none; }
.model-chip.selected { background: rgba(80, 200, 120, 0.18); border-color: rgba(80, 200, 120, 0.4); color: #1e8a4a; font-weight: 600; }

.sticky-save {
  position: sticky; bottom: 0; z-index: 50;
  padding: 12px 0;
  background: linear-gradient(transparent, var(--background, #fff7fb) 30%);
}
.sticky-inner {
  display: flex; align-items: center; justify-content: flex-end; gap: 12px;
  padding: 10px 16px;
  border: 1px solid var(--border); border-radius: 999px;
  background: rgba(255,255,255,0.92); backdrop-filter: blur(8px);
}
.dirty-hint { font-size: 13px; color: var(--primary, #ff78b7); font-weight: 600; }

/* ---- Dialog ---- */
.dialog-mask {
  position: fixed; inset: 0; background: rgba(38, 48, 79, 0.42);
  backdrop-filter: blur(2px);
  display: flex; align-items: center; justify-content: center;
  z-index: 100; padding: 20px;
}
.dialog {
  background: #fff; border-radius: 18px;
  width: 100%; max-width: 640px; max-height: 90vh;
  display: flex; flex-direction: column;
  box-shadow: 0 12px 36px rgba(38, 48, 79, 0.18);
}
.dialog-head { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border); }
.dialog-head h3 { margin: 0; font-size: 16px; font-weight: 700; color: var(--foreground); }
.dialog-body { padding: 16px 20px; overflow-y: auto; flex: 1; display: flex; flex-direction: column; gap: 14px; }
.dialog-foot { padding: 12px 20px; border-top: 1px solid var(--border); display: flex; justify-content: flex-end; gap: 8px; }

.type-tabs { display: flex; gap: 6px; }
.type-tab {
  flex: 1; padding: 8px 14px; border-radius: 999px;
  border: 1px solid var(--border); background: rgba(255,255,255,0.7);
  color: var(--foreground); font-size: 13px; font-weight: 600; cursor: pointer;
  transition: all 0.15s;
}
.type-tab:hover { border-color: var(--input); }
.type-tab.active { background: var(--primary, #ff78b7); color: #fff; border-color: var(--primary, #ff78b7); }

/* ---- 管理员指令表 ---- */
.cmd-grid { display: flex; flex-direction: column; gap: 8px; margin-top: 6px; }
.cmd-row {
  display: grid; grid-template-columns: minmax(160px, 220px) 1fr; gap: 16px;
  padding: 10px 14px; border-radius: 12px;
  background: rgba(255, 255, 255, 0.55); border: 1px solid var(--border);
  font-size: 13px; align-items: baseline;
}
.cmd-row code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  background: rgba(165, 180, 252, 0.18); padding: 2px 8px; border-radius: 6px;
  color: var(--foreground); font-weight: 600; font-size: 12px;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.cmd-row span { color: var(--muted-foreground); line-height: 1.55; }
.cmd-row span code { background: rgba(255,165,200,0.2); font-size: 11px; padding: 1px 6px; }
</style>

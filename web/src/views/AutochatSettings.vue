<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="Plugin · AutoChat"
      title="AutoChat 设置"
      subtitle="人设 / 触发与阈值 / 单群覆盖 / YAML 高级编辑，所有更改即时写回 autochat.yml。"
    />

    <UiAlert v-if="error" variant="destructive" title="操作失败">{{ error }}</UiAlert>

    <UiCard>
      <div class="tabs">
        <button v-for="t in tabs" :key="t.id" class="tab" :class="{ active: tab === t.id }" @click="tab = t.id">
          {{ t.label }}
        </button>
      </div>
    </UiCard>

    <!-- ================= 人设 ================= -->
    <template v-if="tab === 'persona'">
      <UiCard v-if="persona">
        <div class="card-heading">
          <div>
            <h2>默认人设与对话框架</h2>
            <p>默认 persona 应用于未单独配置的群；framework 是组装最终 system prompt 的模板，包含 <code>{persona}</code>、<code>{recent_text}</code> 等占位符。</p>
          </div>
          <div class="actions">
            <UiButton variant="outline" size="sm" @click="loadPersona">刷新</UiButton>
            <UiButton variant="default" size="sm" :loading="savingPersona" @click="savePersona">保存</UiButton>
          </div>
        </div>
        <div class="form-grid">
          <Field label="默认 Persona" full>
            <textarea v-model="persona.default_persona" rows="8" placeholder="你是一个有用的 AI 助手……" />
          </Field>
          <Field label="Framework 模板" full hint="支持 {self_id} {self_name} {persona} {recent_text} {um_text} {sm_text} {rag_mem_text} {rag_summary_text} 等占位符">
            <textarea v-model="persona.framework" rows="10" />
          </Field>
        </div>
      </UiCard>

      <UiCard v-if="persona">
        <SectionHeader title="RAG Summary（对话总结）" desc="用便宜模型生成对话摘要写入向量库，供后续语义检索；空闲时段自动触发。" />
        <div class="form-grid">
          <Field label="启用">
            <label class="check"><input type="checkbox" v-model="persona.rag_summary.enabled" /> 启用</label>
          </Field>
          <Field label="模型"><input v-model="persona.rag_summary.model" type="text" placeholder="openai:gpt-4o-mini" /></Field>
          <Field label="max_tokens"><input v-model.number="persona.rag_summary.max_tokens" type="number" /></Field>
          <Field label="超时 (秒)"><input v-model.number="persona.rag_summary.timeout" type="number" /></Field>
          <Field label="Prompt 模板（{text} 占位）" full>
            <textarea v-model="persona.rag_summary.prompt" rows="4" />
          </Field>
        </div>
      </UiCard>
    </template>

    <!-- ================= 触发 ================= -->
    <template v-if="tab === 'triggers'">
      <UiCard v-if="triggers">
        <div class="card-heading">
          <div>
            <h2>触发与阈值</h2>
            <p>willing_threshold 越低越爱说话；keywords 命中即直接显著加权；ignore_prefixes/patterns 屏蔽其它插件命令。</p>
          </div>
          <div class="actions">
            <UiButton variant="outline" size="sm" @click="loadTriggers">刷新</UiButton>
            <UiButton variant="default" size="sm" :loading="savingTriggers" @click="saveTriggers">保存</UiButton>
          </div>
        </div>
        <div class="form-grid">
          <Field label="willing_threshold（触发阈值）" hint="累计达到该值即主动发言；越低越爱说话">
            <input v-model.number="triggers.willing_threshold" type="number" step="0.1" />
          </Field>
          <Field label="at_delta（被 @ 的增量）" hint="默认 2.5；通常 ≥ threshold 即被 @ 必回复">
            <input v-model.number="triggers.at_delta" type="number" step="0.1" />
          </Field>
          <Field label="keyword_delta（命中关键词的增量）" hint="默认 1.0；命中下方关键词列表时累加，热重载，无需重启">
            <input v-model.number="triggers.keyword_delta" type="number" step="0.1" />
          </Field>
          <Field label="random_delta_max（随机加权上限）" hint="默认 0.2；普通文本随机叠加 [0, max)">
            <input v-model.number="triggers.random_delta_max" type="number" step="0.05" />
          </Field>
          <Field label="/chat 冷却（秒）"><input v-model.number="triggers.chat_cd_seconds" type="number" /></Field>
          <Field label="/tts 冷却（秒）"><input v-model.number="triggers.tts_cd_seconds" type="number" /></Field>
          <Field label="context_size 上下文条数"><input v-model.number="triggers.context_size" type="number" /></Field>
          <Field label="buffer_limit 缓冲容量"><input v-model.number="triggers.buffer_limit" type="number" /></Field>
          <Field label="reply_max_length"><input v-model.number="triggers.reply_max_length" type="number" /></Field>
          <Field label="关键词（每行一个）" full>
            <textarea v-model="keywordsText" rows="4" />
          </Field>
          <Field label="命令前缀屏蔽（每行一个）" full hint="以这些字符/字串开头的纯文本不会触发自动对话">
            <textarea v-model="ignorePrefixesText" rows="3" />
          </Field>
          <Field label="正则屏蔽（每行一个）" full hint="额外的正则表达式列表，用于覆盖纯中文指令">
            <textarea v-model="ignorePatternsText" rows="3" />
          </Field>
        </div>
      </UiCard>
    </template>

    <!-- ================= 模板 ================= -->
    <template v-if="tab === 'templates'">
      <UiCard>
        <div class="card-heading">
          <div>
            <h2>对话模板</h2>
            <p>每个模板包含独立的人设、首选模型、触发倾向（at/关键词/随机增量）、专属关键词和多模态开关；在「单群配置」里把模板分配给一个或多个群聊即可。</p>
          </div>
          <div class="actions">
            <UiButton variant="outline" size="sm" :loading="templatesLoading" @click="loadTemplates">刷新</UiButton>
            <UiButton variant="default" size="sm" @click="addTemplate">新建模板</UiButton>
          </div>
        </div>

        <UiAlert v-if="templatesError" variant="destructive" title="加载/保存失败">{{ templatesError }}</UiAlert>

        <div v-if="!templates.length && !templatesLoading" class="empty">
          暂无模板。点击"新建模板"创建。
        </div>

        <div v-for="t in templates" :key="t.name" class="group-card">
          <div class="group-card-head">
            <div>
              <span class="group-id">{{ t.name }}</span>
              <span v-if="t.isNew" class="badge badge-auto">未保存</span>
              <span v-if="t.used_by_groups?.length" class="badge badge-on">绑定群 {{ t.used_by_groups.length }}</span>
            </div>
            <div class="actions">
              <UiButton variant="outline" size="sm" :loading="t.saving" @click="saveTemplate(t)">保存</UiButton>
              <UiButton variant="destructive" size="sm" @click="removeTemplate(t)">删除</UiButton>
            </div>
          </div>
          <div class="form-grid">
            <Field label="人设 Persona" full hint="留空则继承全局默认 persona">
              <textarea v-model="t.persona" rows="4" />
            </Field>
            <Field label="首选模型（每行一个）" hint="按顺序 fallback；与全局 models 拼接，模板项在前">
              <textarea :value="templateModelsText(t)" @input="setTemplateModelsText(t, ($event.target as HTMLTextAreaElement).value)" rows="3" placeholder="openai:gpt-4o-mini" />
            </Field>
            <Field label="多模态" hint="auto = 看首选模型是否在 multimodal_models 列表；on/off 强制覆盖">
              <select v-model="t.multimodalMode">
                <option value="auto">auto（按模型判定）</option>
                <option value="on">强制开（图片直传 LLM）</option>
                <option value="off">强制关（走 image_caption）</option>
              </select>
            </Field>
            <Field label="willing_threshold 覆盖" hint="0 = 沿用全局/单群设置">
              <input v-model.number="t.willing_threshold" type="number" step="0.1" />
            </Field>
            <Field label="at_delta（被 @ 增量）" hint="0 = 沿用全局">
              <input v-model.number="t.at_delta" type="number" step="0.1" />
            </Field>
            <Field label="keyword_delta（关键词增量）" hint="0 = 沿用全局">
              <input v-model.number="t.keyword_delta" type="number" step="0.1" />
            </Field>
            <Field label="random_delta_max（随机加权上限）" hint="0 = 沿用全局">
              <input v-model.number="t.random_delta_max" type="number" step="0.05" />
            </Field>
            <Field label="模板专属关键词（每行一个）" full hint="会与全局 keywords 合并">
              <textarea :value="templateKeywordsText(t)" @input="setTemplateKeywordsText(t, ($event.target as HTMLTextAreaElement).value)" rows="3" />
            </Field>
            <Field v-if="t.used_by_groups?.length" label="绑定群" full>
              <div class="badge-list">
                <span v-for="gid in t.used_by_groups" :key="gid" class="badge badge-auto">{{ gid }}</span>
              </div>
            </Field>
          </div>
        </div>
      </UiCard>
    </template>

    <!-- ================= 单群 ================= -->
    <template v-if="tab === 'groups'">
      <UiCard>
        <div class="card-heading">
          <div>
            <h2>单群组配置</h2>
            <p>每个群可独立设置：人设覆盖、阈值覆盖、首选模型覆盖、命令/自动回复开关。空白字段表示沿用默认。</p>
          </div>
          <div class="actions">
            <UiButton variant="outline" size="sm" :loading="groupsLoading" @click="loadGroups">刷新</UiButton>
            <UiButton variant="default" size="sm" @click="addGroup">添加群</UiButton>
          </div>
        </div>

        <UiAlert v-if="groupsError" variant="destructive" title="加载/保存失败">{{ groupsError }}</UiAlert>

        <div v-if="!groups.length && !groupsLoading" class="empty">
          暂无单群覆盖配置。点击"添加群"录入群号即可。默认阈值：<code>{{ defaultThreshold }}</code>
        </div>

        <div v-for="g in groups" :key="g.group_id" class="group-card">
          <div class="group-card-head">
            <div>
              <span class="group-id">群 {{ g.group_id }}</span>
              <span v-if="g.chat_enabled" class="badge badge-on">/chat 启用</span>
              <span v-if="g.auto_enabled" class="badge badge-auto">自动回复启用</span>
            </div>
            <div class="actions">
              <UiButton variant="outline" size="sm" :loading="g.saving" @click="saveGroup(g)">保存</UiButton>
              <UiButton variant="destructive" size="sm" @click="removeGroup(g)">移除覆盖</UiButton>
            </div>
          </div>
          <div class="form-grid">
            <Field label="人设覆盖" full>
              <textarea v-model="g.persona" rows="3" placeholder="留空表示使用默认 persona" />
            </Field>
            <Field label="阈值覆盖" hint="留空恢复默认">
              <input
                type="number" step="0.1" :value="g.willing_threshold ?? ''"
                :placeholder="`默认 ${defaultThreshold}`"
                @input="g.willing_threshold = ($event.target as HTMLInputElement).value === '' ? null : Number(($event.target as HTMLInputElement).value)"
              />
            </Field>
            <Field label="首选模型覆盖" hint="等价于群内 /模型 xxx">
              <input v-model="g.model" type="text" placeholder="如 openai:gpt-4o-mini" />
            </Field>
            <Field label="使用模板" hint="留空表示不使用模板（沿用全局/单群覆盖）">
              <select v-model="g.template">
                <option value="">（无）</option>
                <option v-for="n in templateNames" :key="n" :value="n">{{ n }}</option>
              </select>
            </Field>
            <Field label="开关">
              <div class="check-row">
                <label class="check"><input type="checkbox" v-model="g.chat_enabled" /> /chat 命令</label>
                <label class="check"><input type="checkbox" v-model="g.auto_enabled" /> 阈值/关键词自动回复</label>
              </div>
            </Field>
          </div>
        </div>
      </UiCard>
    </template>

    <!-- ================= 高级 (YAML) ================= -->
    <template v-if="tab === 'advanced'">
      <UiCard>
        <div class="card-heading">
          <div>
            <h2>autochat.yml</h2>
            <p>路径：<code>{{ yamlPath || '(尚未生成)' }}</code></p>
            <p>包含本页所有可视化字段以及更冷门的高级字段；保存即覆盖整个文件。</p>
          </div>
          <div class="actions">
            <UiButton variant="outline" size="sm" :loading="loadingYAML" @click="loadYAML">刷新</UiButton>
            <UiButton variant="default" size="sm" :loading="savingYAML" @click="saveYAML">保存 YAML</UiButton>
          </div>
        </div>
        <textarea v-model="yamlText" class="yaml-editor" spellcheck="false" />
      </UiCard>
    </template>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  getAutochatPersona,
  updateAutochatPersona,
  getAutochatTriggers,
  updateAutochatTriggers,
  listAutochatGroups,
  upsertAutochatGroup,
  deleteAutochatGroup,
  getPluginConfig,
  updatePluginConfig,
  listAutochatTemplates,
  upsertAutochatTemplate,
  deleteAutochatTemplate,
  type AutochatPersona,
  type AutochatTriggers,
  type AutochatGroupSetting,
  type AutochatTemplate,
} from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import SectionHeader from '../components/autochat/AutochatSectionHeader.vue'
import Field from '../components/autochat/AutochatField.vue'

type TabId = 'persona' | 'triggers' | 'templates' | 'groups' | 'advanced'
const tabs: { id: TabId; label: string }[] = [
  { id: 'persona', label: '人设与提示词' },
  { id: 'triggers', label: '触发与阈值' },
  { id: 'templates', label: '模板' },
  { id: 'groups', label: '单群配置' },
  { id: 'advanced', label: 'YAML 高级' },
]
const tab = ref<TabId>('persona')

const error = ref('')

// ----- Persona -----
const persona = ref<AutochatPersona | null>(null)
const savingPersona = ref(false)
async function loadPersona() {
  try { persona.value = await getAutochatPersona() } catch (e) { error.value = String(e) }
}
async function savePersona() {
  if (!persona.value) return
  savingPersona.value = true
  error.value = ''
  try { persona.value = await updateAutochatPersona(persona.value) }
  catch (e) { error.value = e instanceof Error ? e.message : String(e) }
  finally { savingPersona.value = false }
}

// ----- Triggers -----
const triggers = ref<AutochatTriggers | null>(null)
const savingTriggers = ref(false)
const keywordsText = computed({
  get: () => (triggers.value?.keywords || []).join('\n'),
  set: (v: string) => { if (triggers.value) triggers.value.keywords = splitLines(v) },
})
const ignorePrefixesText = computed({
  get: () => (triggers.value?.ignore_prefixes || []).join('\n'),
  set: (v: string) => { if (triggers.value) triggers.value.ignore_prefixes = splitLines(v) },
})
const ignorePatternsText = computed({
  get: () => (triggers.value?.ignore_patterns || []).join('\n'),
  set: (v: string) => { if (triggers.value) triggers.value.ignore_patterns = splitLines(v) },
})
function splitLines(v: string) { return v.split('\n').map(s => s.trim()).filter(Boolean) }
async function loadTriggers() {
  try { triggers.value = await getAutochatTriggers() } catch (e) { error.value = String(e) }
}
async function saveTriggers() {
  if (!triggers.value) return
  savingTriggers.value = true
  error.value = ''
  try { triggers.value = await updateAutochatTriggers(triggers.value) }
  catch (e) { error.value = e instanceof Error ? e.message : String(e) }
  finally { savingTriggers.value = false }
}

// ----- Groups -----
interface GroupRow extends AutochatGroupSetting { saving?: boolean }
const groups = ref<GroupRow[]>([])
const defaultThreshold = ref(0)
const groupsLoading = ref(false)
const groupsError = ref('')
async function loadGroups() {
  groupsLoading.value = true
  groupsError.value = ''
  try {
    const data = await listAutochatGroups()
    groups.value = (data.groups || []).map(g => ({ ...g }))
    defaultThreshold.value = data.default_threshold
  } catch (e) {
    groupsError.value = e instanceof Error ? e.message : String(e)
  } finally {
    groupsLoading.value = false
  }
}
function addGroup() {
  const input = window.prompt('请输入要添加的群号（QQ 群 ID）：')
  if (!input) return
  const gid = Number(input.trim())
  if (!Number.isFinite(gid) || gid <= 0) { groupsError.value = '群号无效。'; return }
  if (groups.value.some(g => g.group_id === gid)) return
  groups.value.unshift({
    group_id: gid, persona: '', willing_threshold: null, model: '', template: '',
    chat_enabled: false, auto_enabled: false,
  })
}
async function saveGroup(g: GroupRow) {
  g.saving = true
  groupsError.value = ''
  try {
    const willing = g.willing_threshold
    const updated = await upsertAutochatGroup(g.group_id, {
      persona: g.persona ?? '',
      clear_willing: willing === null || willing === undefined,
      willing_threshold: willing === null || willing === undefined ? undefined : willing,
      model: g.model ?? '',
      template: g.template ?? '',
      chat_enabled: g.chat_enabled,
      auto_enabled: g.auto_enabled,
    })
    Object.assign(g, updated)
  } catch (e) { groupsError.value = e instanceof Error ? e.message : String(e) }
  finally { g.saving = false }
}
async function removeGroup(g: GroupRow) {
  if (!window.confirm(`移除群 ${g.group_id} 的所有覆盖配置？默认值会重新生效。`)) return
  try {
    await deleteAutochatGroup(g.group_id)
    groups.value = groups.value.filter(x => x.group_id !== g.group_id)
  } catch (e) { groupsError.value = e instanceof Error ? e.message : String(e) }
}

// ----- Templates -----
interface TemplateRow extends AutochatTemplate {
  saving?: boolean
  isNew?: boolean
  // 因 multimodal 是三态(null/true/false)，UI 用单独的字符串字段绑定
  multimodalMode?: 'auto' | 'on' | 'off'
}
const templates = ref<TemplateRow[]>([])
const templatesLoading = ref(false)
const templatesError = ref('')

function applyMultimodalMode(t: TemplateRow) {
  if (t.multimodal === true) t.multimodalMode = 'on'
  else if (t.multimodal === false) t.multimodalMode = 'off'
  else t.multimodalMode = 'auto'
}
function syncMultimodalFromMode(t: TemplateRow) {
  if (t.multimodalMode === 'on') t.multimodal = true
  else if (t.multimodalMode === 'off') t.multimodal = false
  else t.multimodal = null
}

async function loadTemplates() {
  templatesLoading.value = true
  templatesError.value = ''
  try {
    const data = await listAutochatTemplates()
    templates.value = (data.templates || []).map(t => {
      const row: TemplateRow = { ...t, models: t.models || [], keywords: t.keywords || [], used_by_groups: t.used_by_groups || [] }
      applyMultimodalMode(row)
      return row
    })
  } catch (e) {
    templatesError.value = e instanceof Error ? e.message : String(e)
  } finally {
    templatesLoading.value = false
  }
}
function addTemplate() {
  const name = window.prompt('请输入新模板名（仅字母/数字/中划线，不能是 default）')
  if (!name) return
  const trimmed = name.trim()
  if (!trimmed || trimmed === 'default') { templatesError.value = '模板名无效'; return }
  if (templates.value.some(t => t.name === trimmed)) return
  const row: TemplateRow = {
    name: trimmed,
    persona: '',
    models: [],
    multimodal: null,
    willing_threshold: 0,
    at_delta: 0,
    keyword_delta: 0,
    random_delta_max: 0,
    keywords: [],
    used_by_groups: [],
    isNew: true,
    multimodalMode: 'auto',
  }
  templates.value.unshift(row)
}
async function saveTemplate(t: TemplateRow) {
  syncMultimodalFromMode(t)
  t.saving = true
  templatesError.value = ''
  try {
    const updated = await upsertAutochatTemplate(t.name, {
      name: t.name,
      persona: t.persona,
      models: t.models,
      multimodal: t.multimodal,
      willing_threshold: t.willing_threshold || 0,
      at_delta: t.at_delta || 0,
      keyword_delta: t.keyword_delta || 0,
      random_delta_max: t.random_delta_max || 0,
      keywords: t.keywords,
      used_by_groups: t.used_by_groups,
    })
    Object.assign(t, updated, { isNew: false })
    applyMultimodalMode(t)
  } catch (e) { templatesError.value = e instanceof Error ? e.message : String(e) }
  finally { t.saving = false }
}
async function removeTemplate(t: TemplateRow) {
  if (t.isNew) {
    templates.value = templates.value.filter(x => x !== t)
    return
  }
  if (!window.confirm(`删除模板 "${t.name}"？所有绑定该模板的群将自动解绑。`)) return
  try {
    await deleteAutochatTemplate(t.name)
    templates.value = templates.value.filter(x => x.name !== t.name)
    // 同步刷新 groups（解除绑定）
    await loadGroups()
  } catch (e) { templatesError.value = e instanceof Error ? e.message : String(e) }
}
function templateModelsText(t: TemplateRow) {
  return (t.models || []).join('\n')
}
function setTemplateModelsText(t: TemplateRow, v: string) {
  t.models = splitLines(v)
}
function templateKeywordsText(t: TemplateRow) {
  return (t.keywords || []).join('\n')
}
function setTemplateKeywordsText(t: TemplateRow, v: string) {
  t.keywords = splitLines(v)
}

const templateNames = computed(() => templates.value.map(t => t.name))

// ----- YAML -----
const yamlText = ref('')
const yamlPath = ref('')
const loadingYAML = ref(false)
const savingYAML = ref(false)
async function loadYAML() {
  loadingYAML.value = true
  error.value = ''
  try {
    const data = await getPluginConfig('autochat')
    yamlText.value = data.yaml
    yamlPath.value = data.path
  } catch (e) { error.value = e instanceof Error ? e.message : String(e) }
  finally { loadingYAML.value = false }
}
async function saveYAML() {
  savingYAML.value = true
  error.value = ''
  try { await updatePluginConfig('autochat', yamlText.value) }
  catch (e) { error.value = e instanceof Error ? e.message : String(e) }
  finally { savingYAML.value = false }
}

onMounted(() => { loadPersona(); loadTriggers(); loadGroups(); loadTemplates(); loadYAML() })
watch(tab, () => { error.value = '' })
</script>

<style scoped>
.tabs { display: flex; gap: 8px; flex-wrap: wrap; }
.tab {
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid var(--border);
  color: var(--foreground); border-radius: 999px;
  padding: 6px 14px; font-size: 13px; font-weight: 600;
  cursor: pointer; transition: all 0.15s;
}
.tab:hover { background: rgba(255, 255, 255, 0.95); border-color: var(--input); }
.tab.active {
  background: var(--primary, #ff78b7); color: #fff;
  border-color: var(--primary, #ff78b7);
}

.card-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; margin-bottom: 14px; }
.card-heading h2 { margin: 0 0 4px 0; font-size: 16px; font-weight: 700; color: var(--foreground); }
.card-heading p { margin: 0; font-size: 12px; color: var(--muted-foreground); line-height: 1.6; }
.card-heading .actions { display: flex; gap: 8px; flex-shrink: 0; }

.form-grid {
  display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 14px;
}
.check { display: inline-flex; align-items: center; gap: 6px; cursor: pointer; color: var(--foreground); font-size: 13px; }
.check-row { display: flex; gap: 16px; flex-wrap: wrap; padding-top: 6px; }
.empty { padding: 16px 0; color: var(--muted-foreground); font-size: 13px; }

.group-card {
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 14px 16px; margin-top: 12px;
  background: rgba(255, 255, 255, 0.7);
}
.group-card-head { display: flex; justify-content: space-between; align-items: center; gap: 8px; margin-bottom: 12px; }
.group-card-head .actions { display: flex; gap: 8px; }
.group-id { font-weight: 700; margin-right: 8px; color: var(--foreground); }
.badge { font-size: 11px; padding: 3px 8px; border-radius: 999px; margin-right: 4px; background: rgba(165, 180, 252, 0.18); color: var(--foreground); font-weight: 600; }
.badge-on { background: rgba(80, 200, 120, 0.18); color: #1e8a4a; }
.badge-auto { background: rgba(120, 140, 240, 0.2); color: #5868c5; }
.badge-list { display: flex; gap: 6px; flex-wrap: wrap; padding-top: 4px; }

.yaml-editor {
  width: 100%; min-height: 480px;
  background: rgba(255, 255, 255, 0.9); color: var(--foreground);
  border: 1px solid var(--input);
  border-radius: 16px; padding: 12px 14px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px; line-height: 1.55; resize: vertical;
}
.yaml-editor:focus { outline: none; border-color: var(--primary, #ff78b7); box-shadow: 0 0 0 3px rgba(255, 120, 183, 0.18); }
code { background: rgba(165, 180, 252, 0.18); padding: 1px 6px; border-radius: 6px; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 12px; }
</style>

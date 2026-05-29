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
          <UiButton variant="outline" size="sm" @click="expandAllTemplates">全部展开</UiButton>
          <UiButton variant="outline" size="sm" @click="collapseAllTemplates">全部折叠</UiButton>
          <UiButton variant="outline" size="sm" :loading="templatesLoading" @click="loadTemplates">刷新</UiButton>
          <UiButton variant="default" size="sm" @click="addTemplate">新建模板</UiButton>
        </div>
        </div>

        <UiAlert v-if="templatesError" variant="destructive" title="加载/保存失败">{{ templatesError }}</UiAlert>

        <div v-if="!templates.length && !templatesLoading" class="empty">
          暂无模板。点击"新建模板"创建。
        </div>

        <div v-for="t in templates" :key="t.rowKey" class="group-card">
          <div class="group-card-head clickable" @click="toggleTemplateCollapse(t.rowKey)">
            <div>
              <span class="collapse-arrow">{{ isTemplateCollapsed(t.rowKey) ? '▶' : '▼' }}</span>
              <span class="group-id">{{ displayTemplateName(t) }}</span>
              <span v-if="t.isNew" class="badge badge-auto">未保存</span>
              <span v-if="t.used_by_groups?.length" class="badge badge-on">绑定群 {{ t.used_by_groups.length }}</span>
              <span v-if="(t.models || []).length" class="badge">{{ (t.models || []).length }} 模型</span>
              <span v-if="(t.keywords || []).length" class="badge">{{ (t.keywords || []).length }} 关键词</span>
            </div>
            <div class="actions" @click.stop>
              <UiButton variant="outline" size="sm" :loading="t.saving" @click="saveTemplate(t)">保存</UiButton>
              <UiButton variant="destructive" size="sm" @click="removeTemplate(t)">删除</UiButton>
            </div>
          </div>
          <div v-show="!isTemplateCollapsed(t.rowKey)" class="form-grid">
            <Field label="人设 Persona" full hint="留空则继承全局默认 persona">
              <textarea v-model="t.persona" rows="4" />
            </Field>
            <Field label="首选模型（按顺序 fallback）" full hint="勾选已接入的模型；如需手动输入未列出的模型，点 自定义 按钮">
              <div v-if="t.models?.length" class="badge-list" style="margin-bottom: 8px;">
                <span v-for="m in t.models" :key="m" class="badge badge-on">
                  {{ m }}
                  <button type="button" class="chip-x" @click="removeTemplateModel(t, m)" title="移除">×</button>
                </span>
              </div>
              <div v-if="availableModels.length" class="model-chips">
                <label v-for="m in availableModels" :key="m" class="model-chip" :class="{ selected: (t.models || []).includes(m) }">
                  <input type="checkbox" :checked="(t.models || []).includes(m)" @change="toggleTemplateModel(t, m)" />
                  {{ m }}
                </label>
              </div>
              <div v-else class="empty-hint">暂无已接入模型。请先在「概览」页配置 Provider 与模型。</div>
              <div style="margin-top: 8px;">
                <UiButton variant="outline" size="sm" @click="addCustomTemplateModel(t)">＋ 自定义</UiButton>
              </div>
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
            <Field label="模板专属关键词（按 Enter 添加）" full hint="会与全局 keywords 合并；命中任一即触发回复">
              <div v-if="t.keywords?.length" class="badge-list" style="margin-bottom: 8px;">
                <span v-for="k in t.keywords" :key="k" class="badge badge-auto">
                  {{ k }}
                  <button type="button" class="chip-x" @click="removeKeyword(t, k)" title="移除">×</button>
                </span>
              </div>
              <input
                type="text"
                :value="keywordDraft[t.rowKey] || ''"
                @input="keywordDraft[t.rowKey] = ($event.target as HTMLInputElement).value"
                @keydown.enter.prevent="addKeyword(t)"
                placeholder="输入关键词后按 Enter 添加"
              />
            </Field>
            <Field label="绑定群聊" full hint="在此直接管理使用该模板的群，输入群号回车添加">
              <div v-if="t.used_by_groups?.length" class="badge-list" style="margin-bottom: 8px;">
                <span v-for="gid in t.used_by_groups" :key="gid" class="badge badge-auto">
                  {{ gid }}
                  <button type="button" class="chip-x" @click.stop="unbindGroupFromTemplate(t, gid)" title="解绑（群将变为默认配置）">×</button>
                </span>
              </div>
              <div v-else class="empty-hint" style="margin-bottom: 6px;">暂无群聊绑定此模板</div>
              <input
                type="text"
                placeholder="输入群号按 Enter 绑定"
                @keydown.enter.prevent="bindGroupToTemplate(t, $event)"
              />
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
            <p>仅控制 /chat 命令和阈值自动回复开关；模板绑定请在「模板」页操作。默认阈值：<code>{{ defaultThreshold }}</code></p>
          </div>
          <div class="actions">
            <UiButton variant="outline" size="sm" :loading="groupsLoading" @click="loadGroups">刷新</UiButton>
          </div>
        </div>

        <UiAlert v-if="groupsError" variant="destructive" title="加载/保存失败">{{ groupsError }}</UiAlert>

        <Field label="批量添加群" full hint="每行一个群号，或逗号/空格分隔，回车提交">
          <textarea v-model="addGroupsText" rows="4" placeholder="123456&#10;789012, 345678" />
        </Field>
        <div style="margin-bottom: 16px;">
          <UiButton variant="default" size="sm" @click="addGroupsBatch">批量添加</UiButton>
        </div>

        <div v-if="!groups.length && !groupsLoading" class="empty">
          暂无单群覆盖配置。在上方输入群号添加。
        </div>

        <table v-if="groups.length" class="group-table">
          <thead>
            <tr>
              <th><input type="checkbox" :checked="isAllGroupsSelected" @change="toggleSelectAllGroups($event)" /></th>
              <th>群号</th>
              <th>绑定模板</th>
              <th>/chat</th>
              <th>自动回复</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="g in groups" :key="g.group_id">
              <td><input type="checkbox" :checked="selectedGroups.has(g.group_id)" @change="toggleSelectGroup(g.group_id)" /></td>
              <td><span class="group-id">{{ g.group_id }}</span></td>
              <td>
                <span v-if="g.template" class="badge badge-auto">{{ g.template }}</span>
                <span v-else class="muted-text">（默认）</span>
              </td>
              <td><label class="check"><input type="checkbox" v-model="g.chat_enabled" @change="saveGroupSilent(g)" /></label></td>
              <td><label class="check"><input type="checkbox" v-model="g.auto_enabled" @change="saveGroupSilent(g)" /></label></td>
              <td><UiButton variant="destructive" size="sm" @click="removeGroup(g)">移除</UiButton></td>
            </tr>
          </tbody>
        </table>

        <div v-if="groups.length && selectedGroups.size > 0" class="batch-actions" style="margin-top: 12px;">
          <span>已选 {{ selectedGroups.size }} 项</span>
          <UiButton size="sm" variant="destructive" @click="batchRemoveAllSelected">批量移除</UiButton>
          <UiButton size="sm" variant="outline" @click="batchToggleChat">批量开关 /chat</UiButton>
          <UiButton size="sm" variant="outline" @click="batchToggleAuto">批量开关 自动回复</UiButton>
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
  updateAutochatYAML,
  listAutochatTemplates,
  upsertAutochatTemplate,
  deleteAutochatTemplate,
  getAutochatProviders,
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
const addGroupsText = ref('')

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

function parseGroupIDs(text: string): number[] {
  return text.split(/[\n,，\s]+/)
    .map(s => s.trim())
    .filter(Boolean)
    .map(Number)
    .filter(n => Number.isFinite(n) && n > 0)
}

async function addGroupsBatch() {
  const ids = parseGroupIDs(addGroupsText.value)
  if (!ids.length) { groupsError.value = '未检测到有效群号'; return }
  const existing = new Set(groups.value.map(g => g.group_id))
  const newIds = ids.filter(id => !existing.has(id))
  if (!newIds.length) { groupsError.value = '所有群号已存在'; return }
  groupsError.value = ''
  for (const gid of newIds) {
    groups.value.push({
      group_id: gid, persona: '', willing_threshold: null, model: '', template: '',
      chat_enabled: false, auto_enabled: false,
    })
  }
  addGroupsText.value = ''
}

async function saveGroupSilent(g: GroupRow) {
  g.saving = true
  groupsError.value = ''
  try {
    await upsertAutochatGroup(g.group_id, {
      template: g.template ?? '',
      chat_enabled: g.chat_enabled,
      auto_enabled: g.auto_enabled,
    })
  } catch (e) { groupsError.value = e instanceof Error ? e.message : String(e) }
  finally { g.saving = false }
}

async function removeGroup(g: GroupRow) {
  if (!window.confirm(`移除群 ${g.group_id} 的所有覆盖配置？`)) return
  try {
    await deleteAutochatGroup(g.group_id)
    groups.value = groups.value.filter(x => x.group_id !== g.group_id)
    await loadTemplates()
  } catch (e) { groupsError.value = e instanceof Error ? e.message : String(e) }
}

// ----- Group selection & batch ops -----
const selectedGroups = ref(new Set<number>())
function toggleSelectGroup(gid: number) {
  const s = new Set(selectedGroups.value)
  s.has(gid) ? s.delete(gid) : s.add(gid)
  selectedGroups.value = s
}
function toggleSelectAllGroups(event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  selectedGroups.value = checked ? new Set(groups.value.map(g => g.group_id)) : new Set()
}
const isAllGroupsSelected = computed(() =>
  groups.value.length > 0 && groups.value.every(g => selectedGroups.value.has(g.group_id))
)

async function batchRemoveAllSelected() {
  if (!selectedGroups.value.size) return
  if (!window.confirm(`确定移除选中的 ${selectedGroups.value.size} 个群？`)) return
  groupsError.value = ''
  try {
    for (const gid of selectedGroups.value) {
      await deleteAutochatGroup(gid)
    }
    selectedGroups.value = new Set()
    await loadGroups()
    await loadTemplates()
  } catch (e) { groupsError.value = e instanceof Error ? e.message : String(e) }
}

async function batchToggleChat() {
  const list = groups.value.filter(g => selectedGroups.value.has(g.group_id))
  if (!list.length) return
  const target = !list.every(g => g.chat_enabled)
  for (const g of list) {
    g.chat_enabled = target
    await saveGroupSilent(g)
  }
}

async function batchToggleAuto() {
  const list = groups.value.filter(g => selectedGroups.value.has(g.group_id))
  if (!list.length) return
  const target = !list.every(g => g.auto_enabled)
  for (const g of list) {
    g.auto_enabled = target
    await saveGroupSilent(g)
  }
}

// ----- Template group binding (在模板卡片内操作) -----
async function bindGroupToTemplate(t: TemplateRow, event: KeyboardEvent) {
  const input = event.target as HTMLInputElement
  const gid = Number(input.value.trim())
  if (!Number.isFinite(gid) || gid <= 0) return
  input.value = ''
  // 如果该群已在 groups 列表中，直接更新其 template
  const existing = groups.value.find(g => g.group_id === gid)
  if (existing) {
    existing.template = t.name
    await upsertAutochatGroup(gid, {
      template: t.name,
      chat_enabled: existing.chat_enabled,
      auto_enabled: existing.auto_enabled,
    })
  } else {
    // 新建群覆盖
    await upsertAutochatGroup(gid, {
      template: t.name,
      chat_enabled: false,
      auto_enabled: false,
    })
    groups.value.push({
      group_id: gid, persona: '', willing_threshold: null, model: '',
      template: t.name, chat_enabled: false, auto_enabled: false,
    })
  }
  if (!t.used_by_groups) t.used_by_groups = []
  const gidStr = String(gid)
  if (!t.used_by_groups.includes(gidStr)) t.used_by_groups.push(gidStr)
}

async function unbindGroupFromTemplate(t: TemplateRow, gid: string) {
  const numGid = Number(gid)
  // 将群的 template 清空
  const existing = groups.value.find(g => g.group_id === numGid)
  if (existing) {
    existing.template = ''
    await upsertAutochatGroup(numGid, {
      template: '',
      chat_enabled: existing.chat_enabled,
      auto_enabled: existing.auto_enabled,
    })
  } else {
    await upsertAutochatGroup(numGid, { template: '', chat_enabled: false, auto_enabled: false })
  }
  t.used_by_groups = (t.used_by_groups || []).filter(x => x !== gid)
}

// ----- Templates -----
interface TemplateRow extends AutochatTemplate {
  saving?: boolean
  isNew?: boolean
  rowKey: string
  originalName?: string
  // 因 multimodal 是三态(null/true/false)，UI 用单独的字符串字段绑定
  multimodalMode?: 'auto' | 'on' | 'off'
}
const templates = ref<TemplateRow[]>([])
let nextTemplateRowID = 1
function createTemplateRowKey(name: string, isNew = false): string {
  return `${isNew ? 'new' : 'saved'}:${name}:${nextTemplateRowID++}`
}
function displayTemplateName(t: TemplateRow): string {
  return t.name || t.originalName || '(未命名模板)'
}
const templatesLoading = ref(false)
const templatesError = ref('')
const availableModels = ref<string[]>([])
// 关键词 chip-input 的临时输入文本（每个模板一个）
const keywordDraft = ref<Record<string, string>>({})

// ----- Template collapse -----
const collapsedTemplates = ref(new Set<string>())
function toggleTemplateCollapse(rowKey: string) {
  const s = new Set(collapsedTemplates.value)
  s.has(rowKey) ? s.delete(rowKey) : s.add(rowKey)
  collapsedTemplates.value = s
}
function isTemplateCollapsed(rowKey: string) {
  return collapsedTemplates.value.has(rowKey)
}
function collapseAllTemplates() {
  collapsedTemplates.value = new Set(templates.value.map(t => t.rowKey))
}
function expandAllTemplates() {
  collapsedTemplates.value = new Set()
}

async function loadAvailableModels() {
  try {
    const data = await getAutochatProviders()
    availableModels.value = data.llm?.models || []
  } catch { /* 忽略：模型列表可选 */ }
}

function toggleTemplateModel(t: TemplateRow, m: string) {
  const list = t.models || []
  if (list.includes(m)) t.models = list.filter(x => x !== m)
  else t.models = [...list, m]
}
function addCustomTemplateModel(t: TemplateRow) {
  const v = window.prompt('输入自定义模型 spec（如 openai:gpt-4o-mini）')
  if (!v) return
  const s = v.trim()
  if (!s) return
  if (!(t.models || []).includes(s)) t.models = [...(t.models || []), s]
}
function removeTemplateModel(t: TemplateRow, m: string) {
  t.models = (t.models || []).filter(x => x !== m)
}

function addKeyword(t: TemplateRow) {
  const v = (keywordDraft.value[t.rowKey] || '').trim()
  if (!v) return
  if (!(t.keywords || []).includes(v)) t.keywords = [...(t.keywords || []), v]
  keywordDraft.value[t.rowKey] = ''
}
function removeKeyword(t: TemplateRow, k: string) {
  t.keywords = (t.keywords || []).filter(x => x !== k)
}

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
      const row: TemplateRow = {
        ...t,
        models: t.models || [],
        keywords: t.keywords || [],
        used_by_groups: t.used_by_groups || [],
        rowKey: createTemplateRowKey(t.name),
        originalName: t.name,
      }
      applyMultimodalMode(row)
      return row
    })
    // 已有模板默认折叠，避免页面过长
    collapsedTemplates.value = new Set(templates.value.map(t => t.rowKey))
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
  if (templates.value.some(t => t.name === trimmed)) {
    templatesError.value = `模板 "${trimmed}" 已存在`
    return
  }
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
    rowKey: createTemplateRowKey(trimmed, true),
    originalName: trimmed,
    multimodalMode: 'auto',
  }
  templates.value.unshift(row)
}
// 把可能为 NaN/undefined/null 的输入归一为有限数。
function safeNum(v: unknown): number {
  const n = typeof v === 'number' ? v : Number(v)
  return Number.isFinite(n) ? n : 0
}
async function saveTemplate(t: TemplateRow) {
  syncMultimodalFromMode(t)
  const templateName = (t.originalName || t.name || '').trim()
  if (!templateName || templateName === 'default') {
    templatesError.value = '模板名无效'
    return
  }
  t.saving = true
  templatesError.value = ''
  try {
    await upsertAutochatTemplate(templateName, {
      name: templateName,
      persona: t.persona ?? '',
      models: t.models || [],
      multimodal: t.multimodal ?? null,
      willing_threshold: safeNum(t.willing_threshold),
      at_delta: safeNum(t.at_delta),
      keyword_delta: safeNum(t.keyword_delta),
      random_delta_max: safeNum(t.random_delta_max),
      keywords: t.keywords || [],
      used_by_groups: t.used_by_groups || [],
    })
    t.isNew = false
    t.originalName = templateName
    t.name = templateName
    // 保存成功后强制从后端 reload，避免前端 state 与后端漂移导致幽灵 row。
    await loadTemplates()
  } catch (e) { templatesError.value = e instanceof Error ? e.message : String(e) }
  finally { t.saving = false }
}
async function removeTemplate(t: TemplateRow) {
  if (t.isNew) {
    templates.value = templates.value.filter(x => x !== t)
    return
  }
  const templateName = (t.originalName || t.name || '').trim()
  if (!templateName) {
    templatesError.value = '模板名无效'
    return
  }
  if (!window.confirm(`删除模板 "${templateName}"？所有绑定该模板的群将自动解绑。`)) return
  try {
    await deleteAutochatTemplate(templateName)
    // 删除成功后强制从后端 reload，避免本地按 name 过滤把同名 row 一并误删
    // 或漏掉真实后端状态。
    await loadTemplates()
    // 同步刷新 groups（解除绑定）
    await loadGroups()
  } catch (e) { templatesError.value = e instanceof Error ? e.message : String(e) }
}
const templateNames = computed(() => templates.value.filter(t => !t.isNew).map(t => t.name))

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
  try {
    await updateAutochatYAML(yamlText.value)
    // YAML 覆写整个配置，保存后重新加载所有数据以保持内存/UI 同步
    await Promise.all([loadPersona(), loadTriggers(), loadGroups(), loadTemplates(), loadYAML()])
  }
  catch (e) { error.value = e instanceof Error ? e.message : String(e) }
  finally { savingYAML.value = false }
}

onMounted(() => { loadPersona(); loadTriggers(); loadGroups(); loadTemplates(); loadAvailableModels(); loadYAML() })
watch(tab, (newTab) => {
  error.value = ''
  // 切换 tab 时重新加载对应数据，避免跨 tab 操作后看到过期状态
  if (newTab === 'persona') loadPersona()
  else if (newTab === 'triggers') loadTriggers()
  else if (newTab === 'templates') loadTemplates()
  else if (newTab === 'groups') { loadGroups(); loadTemplates() }
  else if (newTab === 'advanced') loadYAML()
})
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
.group-card-head.clickable { cursor: pointer; user-select: none; }
.group-card-head.clickable:hover { opacity: 0.85; }
.group-card-head .actions { display: flex; gap: 8px; }
.group-id { font-weight: 700; margin-right: 8px; color: var(--foreground); }
.collapse-arrow { display: inline-block; width: 16px; font-size: 11px; color: var(--muted-foreground); margin-right: 4px; }
.badge { font-size: 11px; padding: 3px 8px; border-radius: 999px; margin-right: 4px; background: rgba(165, 180, 252, 0.18); color: var(--foreground); font-weight: 600; }

/* 按模板分组 */
.template-group-section {
  border: 1px solid var(--border);
  border-radius: 16px;
  margin-top: 12px;
  overflow: hidden;
}
.template-group-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  user-select: none;
  font-size: 14px;
}
.template-group-header:hover { background: rgba(255, 255, 255, 0.8); }
.template-group-body { padding: 0 12px 12px; }

/* 批量操作栏 */
.batch-actions {
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 8px 4px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 8px;
}

/* 群表格 */
.group-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.group-table th {
  text-align: left;
  padding: 8px 10px;
  font-weight: 600;
  color: var(--muted-foreground);
  border-bottom: 2px solid var(--border);
  font-size: 12px;
  white-space: nowrap;
}
.group-table td {
  padding: 8px 10px;
  border-bottom: 1px solid var(--border);
  vertical-align: middle;
}
.group-table tr:hover td {
  background: rgba(255, 255, 255, 0.5);
}
.group-table th:first-child,
.group-table td:first-child {
  width: 36px;
  text-align: center;
}
.muted-text {
  color: var(--muted-foreground);
  font-size: 12px;
}
.badge-on { background: rgba(80, 200, 120, 0.18); color: #1e8a4a; }
.badge-auto { background: rgba(120, 140, 240, 0.2); color: #5868c5; }
.badge-list { display: flex; gap: 6px; flex-wrap: wrap; padding-top: 4px; }
.chip-x {
  margin-left: 6px; border: none; background: transparent; cursor: pointer;
  font-size: 13px; line-height: 1; padding: 0; color: inherit; opacity: 0.6;
}
.chip-x:hover { opacity: 1; }
.empty-hint { color: var(--muted-foreground); font-size: 12px; padding: 8px 0; }
.model-chips { display: flex; flex-wrap: wrap; gap: 6px; max-height: 160px; overflow-y: auto; }
.model-chip {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 4px 10px; border-radius: 999px; font-size: 12px; cursor: pointer;
  border: 1px solid var(--border); background: rgba(255,255,255,0.8); color: var(--foreground);
  user-select: none;
}
.model-chip input { display: none; }
.model-chip.selected { background: rgba(80, 200, 120, 0.18); border-color: rgba(80, 200, 120, 0.4); color: #1e8a4a; font-weight: 600; }

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

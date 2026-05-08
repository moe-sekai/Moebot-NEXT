<template>
  <UiCard>
    <header class="card-heading">
      <div>
        <h2>规则模板</h2>
        <p>把常用规则封装成模板；下游应用可直接引用，避免重复配置。内置 <code>default</code> 模板提供 ID 默认回退。</p>
      </div>
      <UiButton size="sm" :disabled="creating" @click="onNew">新建模板</UiButton>
    </header>

    <UiAlert v-if="error" variant="destructive" title="模板操作失败">{{ error }}</UiAlert>

    <div v-if="!templates.length && !loading" class="tpl-empty">
      暂无模板，点击「新建模板」添加。
    </div>

    <ul class="tpl-list" v-if="templates.length">
      <li v-for="t in templates" :key="t.id" class="tpl-item" :class="{ 'tpl-item--active': editingId === t.id }">
        <div class="tpl-header" @click="toggle(t)">
          <UiBadge v-if="t.builtin" variant="secondary">内置</UiBadge>
          <strong>{{ t.name }}</strong>
          <span v-if="t.description" class="tpl-desc">{{ t.description }}</span>
          <span class="tpl-spacer" />
          <UiBadge variant="secondary">引用 × {{ t.usage_count }}</UiBadge>
          <button type="button" class="tpl-toggle" :aria-expanded="editingId === t.id">
            {{ editingId === t.id ? '收起' : '编辑' }}
          </button>
        </div>

        <div v-if="editingId === t.id && draft" class="tpl-body">
          <div class="tpl-form-grid">
            <label>
              <span>名称</span>
              <input type="text" v-model="draft.name" :disabled="t.builtin" />
            </label>
            <label class="tpl-desc-label">
              <span>描述</span>
              <input type="text" v-model="draft.description" placeholder="给同事一句话解释这个模板用途" />
            </label>
          </div>

          <RuleTabs
            :model-value="draftRules"
            :allow-default="!t.builtin"
            @update:model-value="onRulesChange"
          />

          <div class="tpl-actions">
            <UiButton size="sm" :loading="savingId === t.id" @click="save(t)">保存模板</UiButton>
            <UiButton
              v-if="!t.builtin"
              size="sm"
              variant="destructive"
              :disabled="t.usage_count > 0 || savingId === t.id"
              :title="t.usage_count > 0 ? '被引用中，不能删除' : ''"
              @click="confirmDelete(t)"
            >
              删除
            </UiButton>
            <UiButton size="sm" variant="ghost" @click="editingId = null">取消</UiButton>
            <span v-if="rowError" class="tpl-error">{{ rowError }}</span>
          </div>
        </div>
      </li>
    </ul>

    <!-- 新建模板弹层（简易行内） -->
    <div v-if="creating" class="tpl-create">
      <h3>新建模板</h3>
      <div class="tpl-form-grid">
        <label>
          <span>名称</span>
          <input type="text" v-model="newDraft.name" placeholder="例如：strict-prod" />
        </label>
        <label class="tpl-desc-label">
          <span>描述</span>
          <input type="text" v-model="newDraft.description" placeholder="（可选）说明用途" />
        </label>
      </div>
      <RuleTabs v-model="newDraftRules" allow-default />
      <div class="tpl-actions">
        <UiButton size="sm" :loading="busy" @click="submitNew">创建</UiButton>
        <UiButton size="sm" variant="ghost" :disabled="busy" @click="cancelNew">取消</UiButton>
        <span v-if="newError" class="tpl-error">{{ newError }}</span>
      </div>
    </div>
  </UiCard>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import UiAlert from '../ui/UiAlert.vue'
import UiBadge from '../ui/UiBadge.vue'
import UiButton from '../ui/UiButton.vue'
import UiCard from '../ui/UiCard.vue'
import RuleTabs, { type RuleSet } from './RuleTabs.vue'
import type { FilterTemplatePayload } from '../../api/types'
import {
  createFilterTemplate,
  deleteFilterTemplate,
  updateFilterTemplate,
} from '../../api/client'

const props = defineProps<{ templates: FilterTemplatePayload[]; loading?: boolean }>()
const emit = defineEmits<{ (e: 'changed'): void }>()

const editingId = ref<number | null>(null)
const draft = ref<FilterTemplatePayload | null>(null)
const draftRules = ref<RuleSet>(emptyRuleSet())
const savingId = ref<number | null>(null)
const error = ref('')
const rowError = ref('')

const creating = ref(false)
const newDraft = ref({ name: '', description: '' })
const newDraftRules = ref<RuleSet>(emptyRuleSet())
const busy = ref(false)
const newError = ref('')

watch(
  () => props.templates,
  () => {
    if (editingId.value != null) {
      const t = props.templates.find((x) => x.id === editingId.value)
      if (t) {
        draft.value = { ...t }
        draftRules.value = ruleSetFromTemplate(t)
      }
    }
  },
)

function toggle(t: FilterTemplatePayload) {
  if (editingId.value === t.id) {
    editingId.value = null
    draft.value = null
    return
  }
  editingId.value = t.id
  draft.value = { ...t }
  draftRules.value = ruleSetFromTemplate(t)
  rowError.value = ''
}

function onRulesChange(v: RuleSet) {
  draftRules.value = v
}

async function save(t: FilterTemplatePayload) {
  if (!draft.value) return
  savingId.value = t.id
  rowError.value = ''
  try {
    await updateFilterTemplate(t.id, {
      ...draft.value,
      ...draftRules.value,
    })
    emit('changed')
  } catch (err) {
    rowError.value = errMsg(err)
  } finally {
    savingId.value = null
  }
}

async function confirmDelete(t: FilterTemplatePayload) {
  if (!window.confirm(`确认删除模板「${t.name}」？此操作不可撤销。`)) return
  rowError.value = ''
  try {
    await deleteFilterTemplate(t.id)
    editingId.value = null
    emit('changed')
  } catch (err) {
    rowError.value = errMsg(err)
  }
}

function onNew() {
  creating.value = true
  newDraft.value = { name: '', description: '' }
  newDraftRules.value = emptyRuleSet()
  newError.value = ''
}

function cancelNew() {
  creating.value = false
}

async function submitNew() {
  newError.value = ''
  if (!newDraft.value.name.trim()) {
    newError.value = '名称不能为空。'
    return
  }
  busy.value = true
  try {
    await createFilterTemplate({
      name: newDraft.value.name.trim(),
      description: newDraft.value.description,
      ...newDraftRules.value,
    })
    creating.value = false
    emit('changed')
  } catch (err) {
    newError.value = errMsg(err)
  } finally {
    busy.value = false
  }
}

function ruleSetFromTemplate(t: FilterTemplatePayload): RuleSet {
  return {
    user_id_rules: { ...t.user_id_rules },
    group_id_rules: { ...t.group_id_rules },
    message_rules: { ...t.message_rules },
    private_message_rules: { ...t.private_message_rules },
    group_message_rules: { ...t.group_message_rules },
  }
}

function emptyRuleSet(): RuleSet {
  return {
    user_id_rules: { mode: 'on', ids: [] },
    group_id_rules: { mode: 'on', ids: [] },
    message_rules: { mode: 'on', filters: [], prefix: [], prefix_replace: '' },
    private_message_rules: { mode: 'default', filters: [], prefix: [], prefix_replace: '' },
    group_message_rules: { mode: 'default', filters: [], prefix: [], prefix_replace: '' },
  }
}

function errMsg(err: unknown): string {
  const ax = err as { response?: { data?: { message?: string } }; message?: string }
  return ax?.response?.data?.message || ax?.message || '操作失败。'
}
</script>

<style scoped>
.card-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}
.card-heading h2 { margin: 0; font-size: 15px; }
.card-heading p { margin: 4px 0 0; font-size: 12px; color: var(--muted-foreground, #71717a); }

.tpl-empty {
  padding: 16px;
  border: 1px dashed var(--border, #e4e4e7);
  border-radius: 10px;
  font-size: 13px;
  color: var(--muted-foreground, #71717a);
  margin-top: 12px;
}

.tpl-list { list-style: none; padding: 0; margin: 12px 0 0; display: flex; flex-direction: column; gap: 8px; }
.tpl-item {
  border: 1px solid var(--border, #e4e4e7);
  border-radius: 10px;
  background: var(--background, #fff);
  overflow: hidden;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.tpl-item--active { border-color: rgba(59,130,246,0.5); box-shadow: 0 1px 4px rgba(59,130,246,0.1); }

.tpl-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  cursor: pointer;
  user-select: none;
  flex-wrap: wrap;
}
.tpl-header strong { font-size: 14px; }
.tpl-desc { font-size: 12px; color: var(--muted-foreground, #71717a); }
.tpl-spacer { flex: 1 1 auto; }
.tpl-toggle {
  border: 1px solid var(--border, #e4e4e7);
  background: var(--background, #fff);
  font: inherit;
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 6px;
  cursor: pointer;
}

.tpl-body {
  border-top: 1px solid var(--border, #e4e4e7);
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: var(--muted, #f9fafb);
}
.tpl-form-grid {
  display: grid;
  grid-template-columns: minmax(140px, 240px) 1fr;
  gap: 10px;
}
.tpl-form-grid label { display: flex; flex-direction: column; gap: 4px; font-size: 12px; }
.tpl-form-grid input {
  padding: 6px 8px;
  border: 1px solid var(--border, #d4d4d8);
  border-radius: 6px;
  font: inherit;
  font-size: 13px;
  background: var(--background, #fff);
}
.tpl-form-grid input:disabled { opacity: 0.7; background: var(--muted, #f4f4f5); }
@media (max-width: 600px) {
  .tpl-form-grid { grid-template-columns: 1fr; }
}

.tpl-actions {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.tpl-error { color: #b91c1c; font-size: 12px; }

.tpl-create {
  margin-top: 12px;
  padding: 14px;
  border: 1px dashed rgba(34,197,94,0.5);
  border-radius: 10px;
  background: rgba(34,197,94,0.05);
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.tpl-create h3 { margin: 0; font-size: 14px; }
</style>

<template>
  <UiCard>
    <div class="card-heading">
      <div>
        <h2>{{ title }}</h2>
        <p>由插件 <code>{{ pluginName }}</code> 通过 <code>plugin.Configurable</code> 暴露。</p>
      </div>
      <div class="actions">
        <UiButton variant="outline" size="sm" :loading="loading" @click="load">刷新</UiButton>
        <UiButton variant="default" size="sm" :loading="saving" :disabled="!dirty" @click="save">保存</UiButton>
      </div>
    </div>

    <UiAlert v-if="error" variant="destructive" title="操作失败">{{ error }}</UiAlert>
    <UiAlert v-if="successMsg" variant="info" title="已保存">{{ successMsg }}</UiAlert>

    <div v-if="loading && !schema.length" class="loading-placeholder">加载中…</div>
    <div v-else-if="!schema.length && !loading" class="empty">该插件未声明任何 schema 字段。</div>

    <form v-else class="schema-form" @submit.prevent="save">
      <div v-for="field in schema" :key="field.key" class="field">
        <label :for="`pf-${field.key}`">
          <span class="label">{{ field.label || field.key }}</span>
          <span v-if="field.group" class="group-tag">{{ field.group }}</span>
        </label>

        <select
          v-if="field.type === 'select'"
          :id="`pf-${field.key}`"
          v-model="model[field.key]"
        >
          <option v-for="opt in field.options || []" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
        </select>

        <input
          v-else-if="field.type === 'bool'"
          :id="`pf-${field.key}`"
          type="checkbox"
          :checked="!!model[field.key]"
          @change="model[field.key] = ($event.target as HTMLInputElement).checked"
        />

        <input
          v-else-if="field.type === 'int'"
          :id="`pf-${field.key}`"
          type="number"
          :value="model[field.key] ?? ''"
          @input="model[field.key] = Number(($event.target as HTMLInputElement).value)"
        />

        <input
          v-else-if="field.type === 'float'"
          :id="`pf-${field.key}`"
          type="number"
          step="0.01"
          :value="model[field.key] ?? ''"
          @input="model[field.key] = Number(($event.target as HTMLInputElement).value)"
        />

        <textarea
          v-else-if="field.type === 'textarea'"
          :id="`pf-${field.key}`"
          rows="4"
          :value="(model[field.key] as string) ?? ''"
          @input="model[field.key] = ($event.target as HTMLTextAreaElement).value"
        />

        <input
          v-else
          :id="`pf-${field.key}`"
          type="text"
          :value="(model[field.key] as string) ?? ''"
          @input="model[field.key] = ($event.target as HTMLInputElement).value"
        />

        <p v-if="field.description" class="hint">{{ field.description }}</p>
      </div>
    </form>
  </UiCard>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  getPluginSettings,
  updatePluginSettings,
  type PluginSettingField,
} from '../api/client'
import UiAlert from './ui/UiAlert.vue'
import UiButton from './ui/UiButton.vue'
import UiCard from './ui/UiCard.vue'

const props = withDefaults(
  defineProps<{ pluginName: string; title?: string }>(),
  { title: '插件设置' },
)

const schema = ref<PluginSettingField[]>([])
const model = reactive<Record<string, unknown>>({})
const original = ref<Record<string, unknown>>({})
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const successMsg = ref('')

const dirty = computed(() => JSON.stringify(model) !== JSON.stringify(original.value))

watch(() => props.pluginName, () => load(), { immediate: false })
onMounted(load)

async function load() {
  loading.value = true
  error.value = ''
  successMsg.value = ''
  try {
    const data = await getPluginSettings(props.pluginName)
    schema.value = data.schema || []
    Object.keys(model).forEach(k => delete model[k])
    Object.assign(model, data.values || {})
    original.value = JSON.parse(JSON.stringify(data.values || {}))
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败。'
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  error.value = ''
  successMsg.value = ''
  try {
    const data = await updatePluginSettings(props.pluginName, { ...model })
    Object.keys(model).forEach(k => delete model[k])
    Object.assign(model, data.values || {})
    original.value = JSON.parse(JSON.stringify(data.values || {}))
    successMsg.value = '设置已生效（已同步写回 data/plugins/' + props.pluginName + '.yml）。'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存失败。'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.card-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; }
.actions { display: flex; gap: 8px; }
.schema-form { display: flex; flex-direction: column; gap: 16px; margin-top: 12px; }
.field { display: flex; flex-direction: column; gap: 6px; }
.field label { display: flex; align-items: center; gap: 8px; font-weight: 500; }
.field .group-tag {
  font-size: 11px; padding: 2px 6px; border-radius: 999px;
  background: rgba(165, 180, 252, 0.18); color: var(--muted-foreground);
  font-weight: 600;
}
.field input[type="text"], .field input[type="number"], .field select {
  height: 38px;
  background: rgba(255, 255, 255, 0.9); color: var(--foreground);
  border: 1px solid var(--input); border-radius: 999px;
  padding: 0 13px; font-size: 14px; font-weight: 600;
}
.field textarea {
  background: rgba(255, 255, 255, 0.9); color: var(--foreground);
  border: 1px solid var(--input); border-radius: 16px;
  padding: 10px 13px; font-size: 14px; line-height: 1.5;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  resize: vertical;
}
.field input:focus, .field textarea:focus, .field select:focus {
  outline: none; border-color: var(--primary, #ff78b7);
  box-shadow: 0 0 0 3px rgba(255, 120, 183, 0.18);
}
.hint { font-size: 12px; color: var(--muted-foreground); margin: 0; }
.loading-placeholder, .empty { padding: 16px 0; color: var(--muted-foreground); font-size: 13px; }
code { background: rgba(165, 180, 252, 0.18); padding: 1px 6px; border-radius: 6px; font-size: 12px; }
</style>

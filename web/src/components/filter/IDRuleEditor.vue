<template>
  <div class="rule-editor" :class="{ 'rule-editor--disabled': disabled }">
    <header class="rule-editor-head">
      <div class="rule-editor-title">
        <strong>{{ label }}</strong>
        <p v-if="hint" class="rule-editor-hint-line">{{ hint }}</p>
      </div>
      <div class="mode-segments" role="tablist">
        <button
          v-for="opt in modeOptions"
          :key="opt.value"
          type="button"
          class="mode-seg"
          :class="{ 'mode-seg--active': modelValue.mode === opt.value, [`mode-seg--${opt.tone}`]: modelValue.mode === opt.value }"
          :disabled="disabled"
          @click="setMode(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>
    </header>
    <div v-if="needIDs" class="rule-editor-body">
      <label class="block-label">
        <span>ID 列表（逗号或换行分隔）</span>
        <textarea
          rows="3"
          :value="idsText"
          :disabled="disabled"
          @input="onIdsInput"
          placeholder="例如：&#10;123456789&#10;987654321"
          spellcheck="false"
          autocomplete="off"
        />
      </label>
      <small class="rule-editor-hint">已解析 {{ modelValue.ids?.length ?? 0 }} 个 ID。</small>
    </div>
    <p v-else-if="modelValue.mode === 'on'" class="rule-editor-summary tone-success">所有 ID 都会放行。</p>
    <p v-else-if="modelValue.mode === 'off'" class="rule-editor-summary tone-danger">所有 ID 都会被拦截。</p>
    <p v-else-if="modelValue.mode === 'default'" class="rule-editor-summary tone-muted">回退到默认模板的同字段规则。</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { FilterIDRule } from '../../api/types'

const props = withDefaults(
  defineProps<{
    label: string
    hint?: string
    modelValue: FilterIDRule
    allowDefault?: boolean
    disabled?: boolean
  }>(),
  { allowDefault: false, disabled: false, hint: '' },
)
const emit = defineEmits<{ (e: 'update:modelValue', v: FilterIDRule): void }>()

type ModeOption = { value: FilterIDRule['mode']; label: string; tone: 'muted' | 'success' | 'danger' | 'warning' | 'info' }

const modeOptions = computed<ModeOption[]>(() => {
  const opts: ModeOption[] = [
    { value: 'on', label: '全放行', tone: 'success' },
    { value: 'off', label: '全拦截', tone: 'danger' },
    { value: 'whitelist', label: '白名单', tone: 'info' },
    { value: 'blacklist', label: '黑名单', tone: 'warning' },
  ]
  if (props.allowDefault) {
    opts.unshift({ value: 'default', label: '默认', tone: 'muted' })
  }
  return opts
})

const needIDs = computed(
  () => props.modelValue.mode === 'whitelist' || props.modelValue.mode === 'blacklist',
)

function parseIds(raw: string): number[] {
  return raw
    .split(/[\s,，;；]+/)
    .map((s) => s.trim())
    .filter(Boolean)
    .map((s) => Number(s))
    .filter((n) => Number.isFinite(n))
}

// Local raw text preserves user typing (newlines, trailing whitespace) so that
// the controlled textarea does not strip newlines on the way through the
// `ids` array round-trip.
const idsText = ref<string>((props.modelValue.ids || []).join('\n'))

watch(
  () => props.modelValue.ids,
  (ids) => {
    const next = ids || []
    const parsed = parseIds(idsText.value)
    const same = parsed.length === next.length && parsed.every((v, i) => v === next[i])
    if (!same) idsText.value = next.join('\n')
  },
)

function setMode(mode: FilterIDRule['mode']) {
  if (props.disabled) return
  emit('update:modelValue', { ...props.modelValue, mode })
}

function onIdsInput(ev: Event) {
  const raw = (ev.target as HTMLTextAreaElement).value
  idsText.value = raw
  emit('update:modelValue', { ...props.modelValue, ids: parseIds(raw) })
}
</script>

<style scoped>
.rule-editor {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  border: 1px solid var(--border, #e4e4e7);
  border-radius: 10px;
  background: var(--card, #fff);
}
.rule-editor--disabled { opacity: 0.6; pointer-events: none; }
.rule-editor-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}
.rule-editor-title strong { font-size: 13px; font-weight: 600; }
.rule-editor-hint-line { margin: 2px 0 0; font-size: 11px; color: var(--muted-foreground, #71717a); }
.mode-segments {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 2px;
  padding: 2px;
  background: var(--muted, #f4f4f5);
  border-radius: 8px;
}
.mode-seg {
  border: none;
  background: transparent;
  font: inherit;
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 6px;
  cursor: pointer;
  color: var(--foreground, #18181b);
  transition: background 0.15s, color 0.15s;
}
.mode-seg:hover:not(:disabled) { background: rgba(0,0,0,0.06); }
.mode-seg:disabled { cursor: not-allowed; }
.mode-seg--active { background: var(--background, #fff); box-shadow: 0 1px 3px rgba(0,0,0,0.08); }
.mode-seg--success { color: #15803d; }
.mode-seg--danger { color: #b91c1c; }
.mode-seg--warning { color: #b45309; }
.mode-seg--info { color: #1d4ed8; }
.mode-seg--muted { color: #71717a; }
.rule-editor-body { display: flex; flex-direction: column; gap: 6px; }
.block-label { display: flex; flex-direction: column; gap: 4px; font-size: 12px; }
.block-label textarea {
  padding: 6px 8px;
  border: 1px solid var(--border, #d4d4d8);
  border-radius: 6px;
  font: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  background: var(--background, #fff);
  resize: vertical;
}
.rule-editor-hint {
  color: var(--muted-foreground, #71717a);
  font-size: 11px;
}
.rule-editor-summary {
  margin: 0;
  font-size: 12px;
  padding: 6px 10px;
  border-radius: 6px;
}
.tone-success { background: rgba(34,197,94,0.1); color: #15803d; }
.tone-danger { background: rgba(239,68,68,0.1); color: #b91c1c; }
.tone-muted { background: var(--muted, #f4f4f5); color: var(--muted-foreground, #71717a); }
</style>

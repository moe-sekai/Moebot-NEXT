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
    <template v-if="needRules">
      <label class="block-label">
        <span>正则列表（每行一条，使用 .NET 兼容语法 / dlclark/regexp2）</span>
        <textarea
          rows="3"
          :value="filtersText"
          :disabled="disabled"
          @input="onFiltersInput"
          placeholder="如：&#10;b\\d+&#10;查看\\s*[\\d.]{2,}"
        />
      </label>
      <details class="rule-editor-advanced">
        <summary>前缀放行（高级）</summary>
        <label class="block-label">
          <span>前缀列表（每行一条，匹配后会替换/移除前缀）</span>
          <textarea
            rows="2"
            :value="prefixText"
            :disabled="disabled"
            @input="onPrefixInput"
            placeholder="如：&#10;b1&#10;#"
          />
        </label>
        <label class="block-label">
          <span>前缀替换为</span>
          <input
            type="text"
            :value="modelValue.prefix_replace"
            :disabled="disabled"
            @input="onPrefixReplaceInput"
            placeholder="留空表示移除前缀"
          />
        </label>
      </details>
    </template>
    <p v-else-if="modelValue.mode === 'on'" class="rule-editor-summary tone-success">所有消息都会放行。</p>
    <p v-else-if="modelValue.mode === 'off'" class="rule-editor-summary tone-danger">所有消息都会被拦截。</p>
    <p v-else-if="modelValue.mode === 'default'" class="rule-editor-summary tone-muted">回退到通用消息规则。</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { FilterMessageRule } from '../../api/types'

const props = withDefaults(
  defineProps<{
    label: string
    hint?: string
    modelValue: FilterMessageRule
    allowDefault?: boolean
    disabled?: boolean
  }>(),
  { allowDefault: false, disabled: false, hint: '' },
)
const emit = defineEmits<{ (e: 'update:modelValue', v: FilterMessageRule): void }>()

type ModeOption = { value: FilterMessageRule['mode']; label: string; tone: 'muted' | 'success' | 'danger' | 'warning' | 'info' }

const modeOptions = computed<ModeOption[]>(() => {
  const opts: ModeOption[] = [
    { value: 'on', label: '全放行', tone: 'success' },
    { value: 'off', label: '全拦截', tone: 'danger' },
    { value: 'whitelist', label: '白名单', tone: 'info' },
    { value: 'blacklist', label: '黑名单', tone: 'warning' },
  ]
  if (props.allowDefault) {
    opts.unshift({ value: 'default', label: '继承通用', tone: 'muted' })
  }
  return opts
})

const needRules = computed(
  () => props.modelValue.mode === 'whitelist' || props.modelValue.mode === 'blacklist',
)
function parseLines(raw: string): string[] {
  return raw.split(/\n+/).map((s) => s.trim()).filter(Boolean)
}

// Local raw text preserves user typing (newlines, trailing whitespace) so that
// the controlled textarea does not strip newlines on the way through the
// array round-trip.
const filtersText = ref<string>((props.modelValue.filters || []).join('\n'))
const prefixText = ref<string>((props.modelValue.prefix || []).join('\n'))

function sameStrArr(a: string[], b: string[]): boolean {
  return a.length === b.length && a.every((v, i) => v === b[i])
}

watch(
  () => props.modelValue.filters,
  (v) => {
    const next = v || []
    if (!sameStrArr(parseLines(filtersText.value), next)) filtersText.value = next.join('\n')
  },
)

watch(
  () => props.modelValue.prefix,
  (v) => {
    const next = v || []
    if (!sameStrArr(parseLines(prefixText.value), next)) prefixText.value = next.join('\n')
  },
)

function setMode(mode: FilterMessageRule['mode']) {
  if (props.disabled) return
  emit('update:modelValue', { ...props.modelValue, mode })
}

function onFiltersInput(ev: Event) {
  const raw = (ev.target as HTMLTextAreaElement).value
  filtersText.value = raw
  emit('update:modelValue', { ...props.modelValue, filters: parseLines(raw) })
}

function onPrefixInput(ev: Event) {
  const raw = (ev.target as HTMLTextAreaElement).value
  prefixText.value = raw
  emit('update:modelValue', { ...props.modelValue, prefix: parseLines(raw) })
}

function onPrefixReplaceInput(ev: Event) {
  emit('update:modelValue', { ...props.modelValue, prefix_replace: (ev.target as HTMLInputElement).value })
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
.block-label { display: flex; flex-direction: column; gap: 4px; font-size: 12px; }
.block-label textarea,
.block-label input {
  padding: 6px 8px;
  border: 1px solid var(--border, #d4d4d8);
  border-radius: 6px;
  font: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  background: var(--background, #fff);
  resize: vertical;
}
.rule-editor-advanced {
  border-top: 1px dashed var(--border, #e4e4e7);
  padding-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.rule-editor-advanced summary {
  cursor: pointer;
  font-size: 12px;
  color: var(--muted-foreground, #71717a);
  user-select: none;
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

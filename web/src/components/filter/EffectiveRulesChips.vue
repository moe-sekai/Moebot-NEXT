<template>
  <div class="rule-chips" role="list">
    <span class="rule-chip" :class="`tone-${toneOf(rules.user_id_rules.mode)}`" role="listitem">
      <span class="chip-key">user-id</span>
      <span class="chip-val">{{ labelOf(rules.user_id_rules.mode) }}</span>
    </span>
    <span class="rule-chip" :class="`tone-${toneOf(rules.group_id_rules.mode)}`" role="listitem">
      <span class="chip-key">group-id</span>
      <span class="chip-val">{{ labelOf(rules.group_id_rules.mode) }}</span>
    </span>
    <span class="rule-chip" :class="`tone-${toneOf(rules.message_rules.mode)}`" role="listitem">
      <span class="chip-key">消息</span>
      <span class="chip-val">{{ labelOf(rules.message_rules.mode) }}</span>
    </span>
    <span class="rule-chip" :class="`tone-${toneOf(rules.private_message_rules.mode)}`" role="listitem">
      <span class="chip-key">私聊</span>
      <span class="chip-val">{{ labelOf(rules.private_message_rules.mode) }}</span>
    </span>
    <span class="rule-chip" :class="`tone-${toneOf(rules.group_message_rules.mode)}`" role="listitem">
      <span class="chip-key">群聊</span>
      <span class="chip-val">{{ labelOf(rules.group_message_rules.mode) }}</span>
    </span>
  </div>
</template>

<script setup lang="ts">
import type { FilterEffectiveRules, FilterMode } from '../../api/types'

defineProps<{ rules: FilterEffectiveRules }>()

type Mode = FilterMode | ''

function labelOf(mode: Mode) {
  switch (mode) {
    case 'on': return '全放行'
    case 'off': return '全拦截'
    case 'whitelist': return '白名单'
    case 'blacklist': return '黑名单'
    case 'default': return '默认'
    default: return '未设'
  }
}

function toneOf(mode: Mode) {
  switch (mode) {
    case 'on': return 'success'
    case 'off': return 'danger'
    case 'whitelist': return 'info'
    case 'blacklist': return 'warning'
    case 'default': return 'muted'
    default: return 'muted'
  }
}
</script>

<style scoped>
.rule-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.rule-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  line-height: 18px;
  border: 1px solid transparent;
}
.chip-key { font-weight: 600; opacity: 0.85; }
.chip-val { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.tone-success { background: rgba(34,197,94,0.12); color: #15803d; border-color: rgba(34,197,94,0.3); }
.tone-danger { background: rgba(239,68,68,0.12); color: #b91c1c; border-color: rgba(239,68,68,0.3); }
.tone-info { background: rgba(59,130,246,0.12); color: #1d4ed8; border-color: rgba(59,130,246,0.3); }
.tone-warning { background: rgba(245,158,11,0.12); color: #b45309; border-color: rgba(245,158,11,0.3); }
.tone-muted { background: var(--muted, #f4f4f5); color: var(--muted-foreground, #71717a); border-color: var(--border, #e4e4e7); }
</style>

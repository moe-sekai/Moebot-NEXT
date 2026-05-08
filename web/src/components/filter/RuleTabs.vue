<template>
  <div class="rule-tabs">
    <nav class="rule-tabs-nav" role="tablist">
      <button
        v-for="t in tabs"
        :key="t.id"
        type="button"
        class="rule-tab"
        :class="{ 'rule-tab--active': activeTab === t.id }"
        @click="activeTab = t.id"
      >
        <span class="rule-tab-icon" :class="`tab-icon--${t.tone}`" aria-hidden="true">{{ t.icon }}</span>
        <span class="rule-tab-label">{{ t.label }}</span>
      </button>
    </nav>

    <div class="rule-tabs-body">
      <div v-show="activeTab === 'id'" class="rule-tabs-grid">
        <IDRuleEditor
          label="user-id 名单"
          hint="按 QQ 号过滤来源用户"
          :model-value="modelValue.user_id_rules"
          :allow-default="allowDefault"
          :disabled="disabled"
          @update:model-value="update('user_id_rules', $event)"
        />
        <IDRuleEditor
          label="group-id 名单"
          hint="按群号过滤来源群"
          :model-value="modelValue.group_id_rules"
          :allow-default="allowDefault"
          :disabled="disabled"
          @update:model-value="update('group_id_rules', $event)"
        />
      </div>

      <div v-show="activeTab === 'message'">
        <MessageRuleEditor
          label="消息·通用（兜底）"
          hint="对所有消息生效；当下面「私聊 / 群聊」未指定（mode=default）时回退到这里。"
          :model-value="modelValue.message_rules"
          :disabled="disabled"
          @update:model-value="update('message_rules', $event)"
        />
      </div>

      <div v-show="activeTab === 'private'">
        <MessageRuleEditor
          label="消息·私聊"
          hint="仅对私聊消息生效；选择「继承通用」可使用通用规则。"
          :model-value="modelValue.private_message_rules"
          allow-default
          :disabled="disabled"
          @update:model-value="update('private_message_rules', $event)"
        />
      </div>

      <div v-show="activeTab === 'group'">
        <MessageRuleEditor
          label="消息·群聊"
          hint="仅对群聊消息生效；选择「继承通用」可使用通用规则。"
          :model-value="modelValue.group_message_rules"
          allow-default
          :disabled="disabled"
          @update:model-value="update('group_message_rules', $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import IDRuleEditor from './IDRuleEditor.vue'
import MessageRuleEditor from './MessageRuleEditor.vue'
import type { FilterIDRule, FilterMessageRule } from '../../api/types'

export type RuleSet = {
  user_id_rules: FilterIDRule
  group_id_rules: FilterIDRule
  message_rules: FilterMessageRule
  private_message_rules: FilterMessageRule
  group_message_rules: FilterMessageRule
}

const props = withDefaults(
  defineProps<{
    modelValue: RuleSet
    /** When true, ID rules show a "default" mode option that falls back to the default template. */
    allowDefault?: boolean
    /** Read-only mode (e.g. when an app is using a template). */
    disabled?: boolean
  }>(),
  { allowDefault: true, disabled: false },
)

const emit = defineEmits<{
  (e: 'update:modelValue', v: RuleSet): void
}>()

type TabID = 'id' | 'message' | 'private' | 'group'

const tabs: { id: TabID; label: string; icon: string; tone: string }[] = [
  { id: 'id', label: 'ID 名单', icon: '#', tone: 'info' },
  { id: 'message', label: '消息·通用', icon: '✉', tone: 'muted' },
  { id: 'private', label: '私聊', icon: '◎', tone: 'success' },
  { id: 'group', label: '群聊', icon: '◇', tone: 'warning' },
]

const activeTab = ref<TabID>('id')

function update<K extends keyof RuleSet>(key: K, val: RuleSet[K]) {
  emit('update:modelValue', { ...props.modelValue, [key]: val })
}
</script>

<style scoped>
.rule-tabs {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.rule-tabs-nav {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 4px;
  background: var(--muted, #f4f4f5);
  border-radius: 10px;
}
.rule-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: none;
  border-radius: 7px;
  background: transparent;
  font: inherit;
  font-size: 13px;
  color: var(--muted-foreground, #71717a);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}
.rule-tab:hover { color: var(--foreground, #18181b); }
.rule-tab--active {
  background: var(--background, #fff);
  color: var(--foreground, #18181b);
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}
.rule-tab-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  background: var(--background, #fff);
}
.tab-icon--info { color: #1d4ed8; }
.tab-icon--success { color: #15803d; }
.tab-icon--warning { color: #b45309; }
.tab-icon--muted { color: #71717a; }

.rule-tabs-body { display: flex; flex-direction: column; gap: 8px; }
.rule-tabs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 10px;
}
</style>

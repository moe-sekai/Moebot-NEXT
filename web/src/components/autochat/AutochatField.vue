<template>
  <div class="ac-field" :class="{ 'ac-field--full': full }">
    <label v-if="label">{{ label }}</label>
    <slot />
    <p v-if="hint" class="ac-hint">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{ label?: string; hint?: string; full?: boolean }>()
</script>

<style scoped>
.ac-field { display: flex; flex-direction: column; gap: 6px; }
.ac-field--full { grid-column: 1 / -1; }
.ac-field > label {
  font-size: 13px;
  font-weight: 600;
  color: var(--foreground);
}
/* 与全局 .ui-input / .ui-textarea 保持一致：白底圆角，避免黑色填充 */
.ac-field :slotted(input[type="text"]),
.ac-field :slotted(input[type="number"]),
.ac-field :slotted(select) {
  height: 38px;
  width: 100%;
  border: 1px solid var(--input);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.9);
  color: var(--foreground);
  padding: 0 13px;
  font-size: 14px;
  font-weight: 600;
  font-family: inherit;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.ac-field :slotted(textarea) {
  width: 100%;
  min-height: 86px;
  border: 1px solid var(--input);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.9);
  color: var(--foreground);
  padding: 10px 13px;
  font-size: 14px;
  line-height: 1.5;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  resize: vertical;
}
.ac-field :slotted(input:focus),
.ac-field :slotted(textarea:focus),
.ac-field :slotted(select:focus) {
  outline: none;
  border-color: var(--primary, #ff78b7);
  box-shadow: 0 0 0 3px rgba(255, 120, 183, 0.18);
}
.ac-hint {
  margin: 0;
  font-size: 12px;
  color: var(--muted-foreground);
  line-height: 1.5;
}
</style>

<template>
  <UiCard class-name="metric-card">
    <div class="metric-card__icon"><SvgIcon :name="icon" :size="22" /></div>
    <div class="metric-card__body">
      <div class="metric-card__label">{{ label }}</div>
      <div class="metric-card__value">{{ value }}</div>
      <div v-if="hint" class="metric-card__hint">{{ hint }}</div>
    </div>
  </UiCard>
</template>

<script setup lang="ts">
import type { IconName } from './icons/SvgIcon.vue'
import SvgIcon from './icons/SvgIcon.vue'
import UiCard from './ui/UiCard.vue'

defineProps<{
  label: string
  value: string | number
  hint?: string
  icon: IconName
}>()
</script>

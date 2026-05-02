<template>
  <UiCard class-name="status-card">
    <div class="status-card__header">
      <div class="status-card__icon"><SvgIcon :name="icon" :size="22" /></div>
      <UiBadge :variant="badgeVariant">{{ statusLabel }}</UiBadge>
    </div>
    <div class="status-card__title">{{ title }}</div>
    <p class="status-card__message">{{ message }}</p>
    <p v-if="meta" class="status-card__meta">{{ meta }}</p>
  </UiCard>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { IconName } from './icons/SvgIcon.vue'
import SvgIcon from './icons/SvgIcon.vue'
import UiBadge from './ui/UiBadge.vue'
import UiCard from './ui/UiCard.vue'

const props = defineProps<{
  title: string
  icon: IconName
  ok?: boolean
  status?: string
  message?: string
  meta?: string
}>()

const badgeVariant = computed(() => {
  if (props.ok === undefined) return 'outline'
  return props.ok ? 'success' : 'destructive'
})

const statusLabel = computed(() => {
  if (props.status) return props.status
  if (props.ok === undefined) return 'unknown'
  return props.ok ? 'ok' : 'error'
})

const message = computed(() => props.message || '等待状态上报')
</script>

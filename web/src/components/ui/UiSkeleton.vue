<template>
  <div class="ui-skeleton" :style="styleValue" />
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  width?: string
  height?: string
  radius?: string
}>()

const styleValue = computed(() => ({
  width: props.width ?? '100%',
  height: props.height ?? '1rem',
  borderRadius: props.radius ?? 'var(--radius)',
}))
</script>

<template>
  <span class="moebot-logo" :style="logoStyle" aria-label="Moebot NEXT">
    <span class="moebot-logo__svg" aria-hidden="true" v-html="logoSvg" />
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import logoSvg from '../../../assets/moebot.svg?raw'

const props = withDefaults(
  defineProps<{
    color?: string
    height?: number
  }>(),
  {
    color: 'var(--accent-pink)',
    height: 40,
  },
)

const logoStyle = computed(() => ({
  '--theme-color': props.color,
  height: `${props.height}px`,
}))
</script>

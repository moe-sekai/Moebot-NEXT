<template>
  <span class="ui-badge" :class="`ui-badge--${variant}`"><slot /></span>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'default' | 'secondary' | 'success' | 'warning' | 'destructive' | 'outline'
  }>(),
  { variant: 'default' },
)
</script>

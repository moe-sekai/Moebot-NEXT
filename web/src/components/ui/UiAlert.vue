<template>
  <div class="ui-alert" :class="`ui-alert--${variant}`">
    <div v-if="title" class="ui-alert__title">{{ title }}</div>
    <div class="ui-alert__content"><slot /></div>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string
    variant?: 'default' | 'info' | 'warning' | 'destructive'
  }>(),
  { variant: 'default' },
)
</script>

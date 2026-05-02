<template>
  <button class="ui-button" :class="[`ui-button--${variant}`, `ui-button--${size}`]" :disabled="disabled || loading">
    <span v-if="loading" class="ui-spinner" />
    <slot />
  </button>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'default' | 'secondary' | 'outline' | 'ghost' | 'destructive'
    size?: 'sm' | 'md' | 'lg'
    disabled?: boolean
    loading?: boolean
  }>(),
  {
    variant: 'default',
    size: 'md',
  },
)
</script>

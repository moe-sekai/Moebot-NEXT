<template>
  <section class="ui-card" :class="className">
    <slot />
  </section>
</template>

<script setup lang="ts">
defineProps<{ className?: string }>()
</script>

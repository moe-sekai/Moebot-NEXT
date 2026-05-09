<template>
  <div class="section-header">
    <h3>{{ title }}</h3>
    <p v-if="desc">{{ desc }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string; desc?: string }>()
</script>

<style scoped>
.section-header { margin-bottom: 12px; }
.section-header h3 { margin: 0 0 4px 0; font-size: 15px; font-weight: 600; }
.section-header p { margin: 0; font-size: 12px; color: var(--text-muted); line-height: 1.6; }
</style>

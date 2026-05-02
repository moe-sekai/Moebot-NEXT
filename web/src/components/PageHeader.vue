<template>
  <section class="page-header">
    <div class="page-header__content">
      <div class="page-header__eyebrow">{{ eyebrow }}</div>
      <h1>{{ title }}</h1>
      <p>{{ subtitle }}</p>
    </div>
    <div v-if="$slots.actions" class="page-header__actions">
      <slot name="actions" />
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  eyebrow: string
  title: string
  subtitle: string
}>()
</script>

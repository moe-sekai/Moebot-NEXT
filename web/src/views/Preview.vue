<template>
  <main class="page-stack">
    <PageHeader eyebrow="Renderer Preview" title="渲染预览" subtitle="查看模板输出、预览图与 Renderer 各步骤耗时。" />
    <SatoriPreview />
  </main>
</template>

<script setup lang="ts">
import PageHeader from '../components/PageHeader.vue'
import SatoriPreview from '../components/SatoriPreview.vue'
</script>

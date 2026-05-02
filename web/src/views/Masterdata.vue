<template>
  <main class="page-stack">
    <PageHeader eyebrow="Data Console" title="Masterdata" subtitle="查看基础数据加载情况，并通过只读接口搜索卡牌、曲目、活动与卡池。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadSummary">刷新数据</UiButton>
      </template>
    </PageHeader>

    <MasterdataSummary :summary="summary" :loading="loading" :error="error" />
    <SearchPanel />
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getMasterdataSummary } from '../api/client'
import type { MasterdataSummary as MasterdataSummaryData } from '../api/types'
import MasterdataSummary from '../components/MasterdataSummary.vue'
import PageHeader from '../components/PageHeader.vue'
import SearchPanel from '../components/SearchPanel.vue'
import UiButton from '../components/ui/UiButton.vue'

const summary = ref<MasterdataSummaryData | null>(null)
const loading = ref(false)
const error = ref('')

onMounted(loadSummary)

async function loadSummary() {
  loading.value = true
  error.value = ''
  try {
    summary.value = await getMasterdataSummary()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载 Masterdata 统计失败。'
  } finally {
    loading.value = false
  }
}
</script>

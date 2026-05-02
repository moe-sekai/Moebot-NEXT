<template>
  <n-space vertical size="large">
    <n-page-header title="指令统计" subtitle="最近 7 天指令调用概览" />
    <n-card>
      <n-data-table :columns="columns" :data="rows" :bordered="false" />
    </n-card>
  </n-space>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { api } from '../api/client'

interface StatRow {
  command: string
  count: number
  avg_ms: number
}

const rows = ref<StatRow[]>([])
const columns: DataTableColumns<StatRow> = [
  { title: '指令', key: 'command' },
  { title: '调用次数', key: 'count' },
  { title: '平均耗时(ms)', key: 'avg_ms' },
]

onMounted(async () => {
  const { data } = await api.get('/stats/commands')
  rows.value = data.data ?? []
})
</script>

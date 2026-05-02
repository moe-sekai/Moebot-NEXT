<template>
  <n-space vertical size="large">
    <n-page-header title="群组管理" subtitle="管理 BOT 启用状态和群配置" />
    <n-card>
      <n-data-table :columns="columns" :data="rows" :bordered="false" />
    </n-card>
  </n-space>
</template>

<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NSwitch, type DataTableColumns } from 'naive-ui'
import { api } from '../api/client'

interface GroupRow {
  id: number
  platform: string
  group_id: string
  name: string
  enabled: boolean
}

const rows = ref<GroupRow[]>([])
const columns: DataTableColumns<GroupRow> = [
  { title: '平台', key: 'platform' },
  { title: '群号', key: 'group_id' },
  { title: '群名', key: 'name' },
  { title: '启用', key: 'enabled', render: row => h(NSwitch, { value: row.enabled, disabled: true }) },
]

onMounted(async () => {
  const { data } = await api.get('/groups')
  rows.value = data.data ?? []
})
</script>

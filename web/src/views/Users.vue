<template>
  <n-space vertical size="large">
    <n-page-header title="用户管理" subtitle="查看平台用户与 PJSK 游戏账号绑定关系" />
    <n-card>
      <n-data-table :columns="columns" :data="rows" :bordered="false" />
    </n-card>
  </n-space>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { api } from '../api/client'

interface UserRow {
  id: number
  platform: string
  platform_id: string
  game_id: string
  nickname: string
}

const rows = ref<UserRow[]>([])
const columns: DataTableColumns<UserRow> = [
  { title: '平台', key: 'platform' },
  { title: '用户 ID', key: 'platform_id' },
  { title: '游戏 ID', key: 'game_id' },
  { title: '昵称', key: 'nickname' },
]

onMounted(async () => {
  const { data } = await api.get('/users')
  rows.value = data.data ?? []
})
</script>

<template>
  <UiCard>
    <div class="card-heading">
      <div>
        <h2>最近命令记录</h2>
        <p>从 SQLite 中读取的最近调用记录</p>
      </div>
      <UiButton variant="outline" size="sm" :loading="loading" @click="$emit('refresh')">刷新</UiButton>
    </div>

    <UiAlert v-if="error" variant="destructive" title="加载失败">{{ error }}</UiAlert>
    <div v-else-if="loading" class="table-skeleton">
      <UiSkeleton v-for="item in 5" :key="item" height="42px" />
    </div>
    <div v-else-if="commands.length === 0" class="empty-state">
      <div class="empty-state__icon">⌘</div>
      <p>暂无命令记录，等机器人收到指令后这里会自动显示。</p>
    </div>
    <div v-else class="table-wrap">
      <table class="ui-table">
        <thead>
          <tr>
            <th>指令</th>
            <th>平台</th>
            <th>用户</th>
            <th>群组</th>
            <th>耗时</th>
            <th>时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in commands" :key="row.id">
            <td class="font-medium">{{ row.command }}</td>
            <td>{{ row.platform }}</td>
            <td>{{ row.user_id || '-' }}</td>
            <td>{{ row.group_id || '-' }}</td>
            <td>{{ row.response_ms }} ms</td>
            <td>{{ formatTime(row.created_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <p class="muted-text">{{ message || '展示最近记录，不包含任何密钥信息。' }}</p>
  </UiCard>
</template>

<script setup lang="ts">
import type { RecentCommand } from '../api/types'
import UiAlert from './ui/UiAlert.vue'
import UiButton from './ui/UiButton.vue'
import UiCard from './ui/UiCard.vue'
import UiSkeleton from './ui/UiSkeleton.vue'

defineEmits<{
  refresh: []
}>()

defineProps<{
  commands: RecentCommand[]
  loading?: boolean
  error?: string
  message?: string
}>()

function formatTime(value: string) {
  return value ? new Date(value).toLocaleString() : '-'
}
</script>

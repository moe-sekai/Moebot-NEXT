<template>
  <main class="page-stack">
    <PageHeader eyebrow="Logs" title="日志" subtitle="当前后端暂无日志流接口，先展示最近命令记录作为轻量运行日志。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadRecentCommands">刷新记录</UiButton>
      </template>
    </PageHeader>

    <UiAlert variant="warning" title="日志接口 TODO">
      TODO：后端可新增结构化日志查询或 WebSocket/SSE 日志流接口；当前页面复用 /api/commands/recent。
    </UiAlert>

    <RecentCommands
      :commands="commands"
      :loading="loading"
      :error="error"
      :message="message"
      @refresh="loadRecentCommands"
    />
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getRecentCommands } from '../api/client'
import type { RecentCommand } from '../api/types'
import PageHeader from '../components/PageHeader.vue'
import RecentCommands from '../components/RecentCommands.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'

const commands = ref<RecentCommand[]>([])
const loading = ref(false)
const error = ref('')
const message = ref('')

onMounted(loadRecentCommands)

async function loadRecentCommands() {
  loading.value = true
  error.value = ''
  try {
    const result = await getRecentCommands(30)
    commands.value = result.data ?? []
    message.value = result.message
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载最近命令失败。'
  } finally {
    loading.value = false
  }
}
</script>

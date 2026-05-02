<template>
  <main class="page-stack">
    <PageHeader eyebrow="OneBot" title="Bot" subtitle="查看机器人驱动、命令前缀、昵称与连接说明。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadBot">刷新 Bot 状态</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="Bot 状态加载失败">{{ error }}</UiAlert>

    <div class="dashboard-grid dashboard-grid--main">
      <UiCard>
        <div class="card-heading">
          <div>
            <h2>连接状态</h2>
            <p>ZeroBot / OneBot v11 反向 WebSocket 配置概览。</p>
          </div>
          <UiBadge :variant="status?.bot.ok ? 'success' : 'destructive'">{{ status?.bot.status ?? 'unknown' }}</UiBadge>
        </div>
        <dl class="info-list">
          <div><dt>状态说明</dt><dd>{{ status?.bot.message ?? '-' }}</dd></div>
          <div><dt>驱动类型</dt><dd>{{ status?.bot.driver_type ?? '-' }}</dd></div>
          <div><dt>监听地址</dt><dd>{{ status?.bot.listen ?? '-' }}</dd></div>
          <div><dt>URL 已配置</dt><dd>{{ status?.bot.url_configured ? '是' : '否' }}</dd></div>
          <div><dt>命令前缀</dt><dd>{{ status?.bot.command_prefix ?? '-' }}</dd></div>
          <div><dt>昵称</dt><dd>{{ status?.bot.nicknames?.join(' / ') || '-' }}</dd></div>
        </dl>
      </UiCard>

      <UiCard>
        <div class="card-heading">
          <div>
            <h2>常用命令</h2>
            <p>渲染模板与查询功能对应的入口。</p>
          </div>
          <UiBadge variant="secondary">只读</UiBadge>
        </div>
        <CommandList />
      </UiCard>
    </div>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getStatus } from '../api/client'
import type { RuntimeStatus } from '../api/types'
import CommandList from '../components/CommandList.vue'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'

const status = ref<RuntimeStatus | null>(null)
const loading = ref(false)
const error = ref('')

onMounted(loadBot)

async function loadBot() {
  loading.value = true
  error.value = ''
  try {
    status.value = await getStatus()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载 Bot 状态失败。'
  } finally {
    loading.value = false
  }
}
</script>

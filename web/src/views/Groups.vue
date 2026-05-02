<template>
  <main class="page-stack">
    <PageHeader eyebrow="Groups" title="群组" subtitle="查看 BOT 启用状态和群配置。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadGroups">刷新群组</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="群组加载失败">{{ error }}</UiAlert>
    <UiCard>
      <div v-if="loading" class="table-skeleton">
        <UiSkeleton v-for="item in 6" :key="item" height="42px" />
      </div>
      <div v-else-if="rows.length === 0" class="empty-state">
        <div class="empty-state__icon"><SvgIcon name="groups" :size="22" /></div>
        <p>暂无群组记录，机器人加入或收到群消息后这里会显示。</p>
      </div>
      <div v-else class="table-wrap">
        <table class="ui-table">
          <thead>
            <tr>
              <th>平台</th>
              <th>群号</th>
              <th>群名</th>
              <th>启用</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>{{ row.platform }}</td>
              <td class="font-medium">{{ row.group_id }}</td>
              <td>{{ row.name || '-' }}</td>
              <td><UiBadge :variant="row.enabled ? 'success' : 'warning'">{{ row.enabled ? '是' : '否' }}</UiBadge></td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="muted-text">共 {{ total }} 条记录。当前后端更新群配置接口仍为 TODO，本页保持只读。</p>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getGroups } from '../api/client'
import type { GroupRow } from '../api/types'
import SvgIcon from '../components/icons/SvgIcon.vue'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

const rows = ref<GroupRow[]>([])
const total = ref(0)
const loading = ref(false)
const error = ref('')

onMounted(loadGroups)

async function loadGroups() {
  loading.value = true
  error.value = ''
  try {
    const result = await getGroups(1, 50)
    rows.value = result.data ?? []
    total.value = result.total ?? rows.value.length
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载群组失败。'
  } finally {
    loading.value = false
  }
}
</script>

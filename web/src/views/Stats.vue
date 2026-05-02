<template>
  <main class="page-stack">
    <PageHeader eyebrow="Command Stats" title="指令统计" subtitle="查看最近 7 天指令调用概览。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadStats">刷新统计</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="统计加载失败">{{ error }}</UiAlert>
    <UiCard>
      <div class="card-heading">
        <div>
          <h2>最近 7 天</h2>
          <p>统计起点：{{ sinceLabel }}</p>
        </div>
        <UiBadge variant="secondary">{{ rows.length }} 条指令</UiBadge>
      </div>

      <div v-if="loading" class="table-skeleton">
        <UiSkeleton v-for="item in 6" :key="item" height="42px" />
      </div>
      <div v-else-if="rows.length === 0" class="empty-state">
        <div class="empty-state__icon"><SvgIcon name="stats" :size="22" /></div>
        <p>暂无统计记录，命令被调用后这里会自动显示。</p>
      </div>
      <div v-else class="table-wrap">
        <table class="ui-table">
          <thead>
            <tr>
              <th>指令</th>
              <th>调用次数</th>
              <th>平均耗时</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.command">
              <td class="font-medium">{{ row.command }}</td>
              <td>{{ row.count.toLocaleString() }}</td>
              <td>{{ Math.round(row.avg_ms) }} ms</td>
            </tr>
          </tbody>
        </table>
      </div>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getCommandStats } from '../api/client'
import type { CommandStatRow } from '../api/types'
import SvgIcon from '../components/icons/SvgIcon.vue'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

const rows = ref<CommandStatRow[]>([])
const since = ref('')
const loading = ref(false)
const error = ref('')

const sinceLabel = computed(() => since.value ? new Date(since.value).toLocaleString() : '-')

onMounted(loadStats)

async function loadStats() {
  loading.value = true
  error.value = ''
  try {
    const result = await getCommandStats(7)
    rows.value = result.data ?? []
    since.value = result.since
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载指令统计失败。'
  } finally {
    loading.value = false
  }
}
</script>

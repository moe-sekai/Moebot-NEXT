<template>
  <k-layout>
    <div class="stats-page">
      <h1>指令统计</h1>
      <p class="description">查看各指令的使用情况和平均响应时间。</p>

      <div class="total-card" v-if="stats">
        <span class="total-value">{{ stats.total }}</span>
        <span class="total-label">总调用次数</span>
      </div>

      <table v-if="stats?.commands?.length" class="data-table">
        <thead>
          <tr>
            <th>指令</th>
            <th>调用次数</th>
            <th>平均响应</th>
            <th>占比</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="cmd in stats.commands" :key="cmd.name">
            <td><code>/{{ cmd.name }}</code></td>
            <td>{{ cmd.count }}</td>
            <td>{{ cmd.avgResponseMs }}ms</td>
            <td>
              <div class="bar-container">
                <div class="bar" :style="{ width: (cmd.count / stats.total * 100) + '%' }" />
                <span>{{ (cmd.count / stats.total * 100).toFixed(1) }}%</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else class="empty">暂无统计数据</div>
    </div>
  </k-layout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { send } from '@koishijs/client'

const stats = ref<any>(null)

onMounted(async () => {
  stats.value = await send('moebot/stats')
})
</script>

<style scoped>
.stats-page { padding: 24px; max-width: 960px; margin: 0 auto; }
h1 { margin: 0 0 8px; }
.description { color: #888; margin-bottom: 24px; }
.total-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  background: var(--k-card-bg);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
  border: 1px solid var(--k-color-border);
}
.total-value { font-size: 48px; font-weight: 700; color: var(--k-color-active); }
.total-label { font-size: 14px; color: #888; margin-top: 4px; }
.data-table { width: 100%; border-collapse: collapse; }
.data-table th, .data-table td {
  padding: 10px 12px;
  text-align: left;
  border-bottom: 1px solid var(--k-color-border);
}
.data-table th { color: #888; font-size: 13px; font-weight: 600; }
code { background: var(--k-side-bg); padding: 2px 6px; border-radius: 4px; font-size: 13px; }
.bar-container {
  display: flex;
  align-items: center;
  gap: 8px;
}
.bar {
  height: 6px;
  background: var(--k-color-active);
  border-radius: 3px;
  min-width: 4px;
  max-width: 120px;
}
.bar-container span { font-size: 12px; color: #888; }
.empty { text-align: center; padding: 48px; color: #888; }
</style>

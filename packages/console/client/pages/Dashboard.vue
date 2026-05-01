<template>
  <k-layout>
    <div class="moebot-dashboard">
      <h1>🤖 Moebot NEXT</h1>
      <p class="subtitle">PJSK 查询机器人管理面板</p>

      <div class="stats-grid" v-if="status">
        <div class="stat-card">
          <div class="stat-value">{{ formatUptime(status.uptime) }}</div>
          <div class="stat-label">运行时间</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ status.totalUsers }}</div>
          <div class="stat-label">注册用户</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ status.totalGroups }}</div>
          <div class="stat-label">接入群组</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ status.totalCommands }}</div>
          <div class="stat-label">指令调用</div>
        </div>
      </div>

      <div class="info-section" v-if="status">
        <h3>系统信息</h3>
        <table>
          <tbody>
            <tr><td>Node.js</td><td>{{ status.nodeVersion }}</td></tr>
            <tr><td>平台</td><td>{{ status.platform }}</td></tr>
            <tr>
              <td>内存使用</td>
              <td>{{ formatBytes(status.memoryUsage?.heapUsed) }} / {{ formatBytes(status.memoryUsage?.heapTotal) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="loading" v-else>加载中...</div>
    </div>
  </k-layout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { send } from '@koishijs/client'

const status = ref<any>(null)

onMounted(async () => {
  status.value = await send('moebot/status')
})

function formatUptime(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `${h}h ${m}m`
}

function formatBytes(bytes?: number): string {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  return `${bytes.toFixed(1)} ${units[i]}`
}
</script>

<style scoped>
.moebot-dashboard {
  padding: 24px;
  max-width: 960px;
  margin: 0 auto;
}
h1 { margin: 0 0 4px; font-size: 28px; }
.subtitle { color: #888; margin: 0 0 24px; }
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
  margin-bottom: 32px;
}
.stat-card {
  background: var(--k-card-bg);
  border-radius: 12px;
  padding: 20px;
  text-align: center;
  border: 1px solid var(--k-color-border);
}
.stat-value { font-size: 32px; font-weight: 700; color: var(--k-color-active); }
.stat-label { font-size: 14px; color: #888; margin-top: 4px; }
.info-section { background: var(--k-card-bg); border-radius: 12px; padding: 20px; }
.info-section h3 { margin: 0 0 12px; }
.info-section table { width: 100%; }
.info-section td { padding: 6px 0; }
.info-section td:first-child { color: #888; width: 120px; }
.loading { text-align: center; padding: 48px; color: #888; }
</style>

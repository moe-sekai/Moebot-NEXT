<template>
  <k-layout>
    <div class="groups-page">
      <h1>群组管理</h1>
      <p class="description">管理 BOT 接入的群组，可按群开关功能。</p>

      <div v-if="groups.length === 0" class="empty">暂无群组数据</div>

      <table v-else class="data-table">
        <thead>
          <tr>
            <th>群组 ID</th>
            <th>平台</th>
            <th>名称</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="group in groups" :key="group.id">
            <td>{{ group.groupId }}</td>
            <td>{{ group.platform }}</td>
            <td>{{ group.name || '-' }}</td>
            <td>
              <span class="badge" :class="group.enabled ? 'badge-success' : 'badge-muted'">
                {{ group.enabled ? '启用' : '禁用' }}
              </span>
            </td>
            <td>
              <button class="btn btn-sm" @click="toggleGroup(group)">
                {{ group.enabled ? '禁用' : '启用' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </k-layout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { send } from '@koishijs/client'

const groups = ref<any[]>([])

onMounted(async () => {
  groups.value = await send('moebot/groups') || []
})

async function toggleGroup(group: any) {
  await send('moebot/groups/toggle', { id: group.id, enabled: !group.enabled })
  group.enabled = !group.enabled
}
</script>

<style scoped>
.groups-page { padding: 24px; max-width: 960px; margin: 0 auto; }
h1 { margin: 0 0 8px; }
.description { color: #888; margin-bottom: 24px; }
.empty { text-align: center; padding: 48px; color: #888; }
.data-table { width: 100%; border-collapse: collapse; }
.data-table th, .data-table td {
  padding: 10px 12px;
  text-align: left;
  border-bottom: 1px solid var(--k-color-border);
}
.data-table th { color: #888; font-size: 13px; font-weight: 600; }
.badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}
.badge-success { background: rgba(68, 204, 136, 0.15); color: #44cc88; }
.badge-muted { background: rgba(136, 136, 136, 0.15); color: #888; }
.btn { padding: 6px 12px; border-radius: 6px; border: 1px solid var(--k-color-border); background: var(--k-card-bg); color: inherit; cursor: pointer; font-size: 13px; }
.btn:hover { opacity: 0.85; }
.btn-sm { padding: 4px 10px; font-size: 12px; }
</style>

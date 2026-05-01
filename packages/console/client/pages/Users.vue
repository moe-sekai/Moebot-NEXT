<template>
  <k-layout>
    <div class="users-page">
      <h1>用户管理</h1>
      <p class="description">查看已绑定游戏账号的用户列表。</p>

      <div class="search-bar">
        <input v-model="search" placeholder="搜索用户 ID / 游戏 ID..." />
      </div>

      <div v-if="filteredUsers.length === 0" class="empty">暂无用户数据</div>

      <table v-else class="data-table">
        <thead>
          <tr>
            <th>平台</th>
            <th>用户 ID</th>
            <th>游戏 ID</th>
            <th>区服</th>
            <th>绑定时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in filteredUsers" :key="user.id">
            <td>{{ user.platform }}</td>
            <td>{{ user.platformId }}</td>
            <td>{{ user.gameId || '-' }}</td>
            <td>{{ user.region?.toUpperCase() || 'JP' }}</td>
            <td>{{ formatDate(user.createdAt) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </k-layout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { send } from '@koishijs/client'

const users = ref<any[]>([])
const search = ref('')

onMounted(async () => {
  users.value = await send('moebot/users') || []
})

const filteredUsers = computed(() => {
  if (!search.value) return users.value
  const q = search.value.toLowerCase()
  return users.value.filter(u =>
    u.platformId?.toLowerCase().includes(q) ||
    u.gameId?.toLowerCase().includes(q)
  )
})

function formatDate(d: any): string {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.users-page { padding: 24px; max-width: 960px; margin: 0 auto; }
h1 { margin: 0 0 8px; }
.description { color: #888; margin-bottom: 16px; }
.search-bar { margin-bottom: 16px; }
.search-bar input {
  width: 100%;
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px solid var(--k-color-border);
  background: var(--k-card-bg);
  color: inherit;
  font-size: 14px;
}
.empty { text-align: center; padding: 48px; color: #888; }
.data-table { width: 100%; border-collapse: collapse; }
.data-table th, .data-table td {
  padding: 10px 12px;
  text-align: left;
  border-bottom: 1px solid var(--k-color-border);
}
.data-table th { color: #888; font-size: 13px; font-weight: 600; }
</style>

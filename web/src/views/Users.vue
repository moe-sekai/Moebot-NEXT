<template>
  <main class="page-stack">
    <PageHeader eyebrow="Users" title="用户" subtitle="查看平台用户与 PJSK 游戏账号绑定关系。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadUsers">刷新用户</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="用户加载失败">{{ error }}</UiAlert>
    <UiCard>
      <div v-if="loading" class="table-skeleton">
        <UiSkeleton v-for="item in 6" :key="item" height="42px" />
      </div>
      <div v-else-if="rows.length === 0" class="empty-state">
        <div class="empty-state__icon"><SvgIcon name="users" :size="22" /></div>
        <p>暂无用户记录，用户与机器人交互后这里会显示。</p>
      </div>
      <div v-else class="table-wrap">
        <table class="ui-table">
          <thead>
            <tr>
              <th>平台</th>
              <th>用户 ID</th>
              <th>游戏 ID</th>
              <th>昵称</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>{{ row.platform }}</td>
              <td class="font-medium">{{ row.platform_id }}</td>
              <td>{{ row.game_id || '-' }}</td>
              <td>{{ row.nickname || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="muted-text">共 {{ total }} 条记录。删除用户接口已存在，但本页未提供写操作以避免误触。</p>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getUsers } from '../api/client'
import type { UserRow } from '../api/types'
import SvgIcon from '../components/icons/SvgIcon.vue'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'
import UiSkeleton from '../components/ui/UiSkeleton.vue'

const rows = ref<UserRow[]>([])
const total = ref(0)
const loading = ref(false)
const error = ref('')

onMounted(loadUsers)

async function loadUsers() {
  loading.value = true
  error.value = ''
  try {
    const result = await getUsers(1, 50)
    rows.value = result.data ?? []
    total.value = result.total ?? rows.value.length
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载用户失败。'
  } finally {
    loading.value = false
  }
}
</script>

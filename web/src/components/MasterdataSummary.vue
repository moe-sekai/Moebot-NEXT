<template>
  <UiCard>
    <div class="card-heading">
      <div>
        <h2>Masterdata 数据统计</h2>
        <p>当前已载入的 Project SEKAI 基础数据</p>
      </div>
      <UiBadge :variant="summary?.loaded ? 'success' : 'warning'">{{ summary?.loaded ? '已加载' : '未加载' }}</UiBadge>
    </div>

    <UiAlert v-if="error" variant="destructive" title="加载失败">{{ error }}</UiAlert>
    <div v-else-if="loading" class="summary-grid">
      <UiSkeleton v-for="item in 4" :key="item" height="92px" />
    </div>
    <div v-else class="summary-grid">
      <div v-for="item in items" :key="item.key" class="summary-tile">
        <div class="summary-tile__icon"><SvgIcon :name="item.icon" :size="22" /></div>
        <div>
          <div class="summary-tile__label">{{ item.label }}</div>
          <div class="summary-tile__value">{{ item.value.toLocaleString() }}</div>
        </div>
      </div>
    </div>

    <p class="muted-text">最近加载：{{ loadedAt }}</p>
  </UiCard>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { MasterdataSummary } from '../api/types'
import SvgIcon from './icons/SvgIcon.vue'
import UiAlert from './ui/UiAlert.vue'
import UiBadge from './ui/UiBadge.vue'
import UiCard from './ui/UiCard.vue'
import UiSkeleton from './ui/UiSkeleton.vue'

const props = defineProps<{
  summary: MasterdataSummary | null
  loading?: boolean
  error?: string
}>()

const counts = computed(() => props.summary?.counts ?? { cards: 0, musics: 0, events: 0, gachas: 0 })

const items = computed(() => [
  { key: 'cards', label: '卡牌数量', icon: 'preview' as const, value: counts.value.cards },
  { key: 'musics', label: '曲目数量', icon: 'resources' as const, value: counts.value.musics },
  { key: 'events', label: '活动数量', icon: 'clock' as const, value: counts.value.events },
  { key: 'gachas', label: '卡池数量', icon: 'sparkle' as const, value: counts.value.gachas },
])

const loadedAt = computed(() => {
  if (!props.summary?.loaded_at) return '暂无记录'
  return new Date(props.summary.loaded_at).toLocaleString()
})
</script>

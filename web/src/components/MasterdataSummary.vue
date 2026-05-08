<template>
  <UiCard>
    <div class="card-heading">
      <div>
        <h2>Masterdata 数据统计</h2>
        <p>{{ multiServer ? '各已启用区服分别统计已载入的 Project SEKAI 基础数据。' : '当前已载入的 Project SEKAI 基础数据' }}</p>
      </div>
      <UiBadge v-if="multiServer" :variant="anyLoaded ? 'success' : 'warning'">{{ loadedRegionCount }}/{{ servers!.length }} 已加载</UiBadge>
      <UiBadge v-else :variant="summary?.loaded ? 'success' : 'warning'">{{ summary?.loaded ? '已加载' : '未加载' }}</UiBadge>
    </div>

    <UiAlert v-if="error" variant="destructive" title="加载失败">{{ error }}</UiAlert>
    <div v-else-if="loading" class="summary-grid">
      <UiSkeleton v-for="item in 4" :key="item" height="92px" />
    </div>
    <template v-else-if="multiServer">
      <div v-if="servers!.length === 0" class="muted-text">暂无已启用区服。</div>
      <div v-else class="masterdata-server-stack">
        <section v-for="entry in serverEntries" :key="entry.region" class="masterdata-server-block">
          <header class="masterdata-server-block__head">
            <div>
              <strong>{{ entry.label }} · {{ entry.region.toUpperCase() }}</strong>
              <span v-if="entry.is_default" class="masterdata-server-block__default">默认</span>
            </div>
            <div class="masterdata-server-block__meta">
              <UiBadge :variant="entry.loaded ? 'success' : 'warning'">{{ entry.loaded ? '已加载' : '未加载' }}</UiBadge>
              <span class="muted-text">加载时间：{{ formatTime(entry.loaded_at) }}</span>
            </div>
          </header>
          <div class="summary-grid">
            <div v-for="item in countItems(entry.counts)" :key="item.key" class="summary-tile">
              <div class="summary-tile__icon"><SvgIcon :name="item.icon" :size="22" /></div>
              <div>
                <div class="summary-tile__label">{{ item.label }}</div>
                <div class="summary-tile__value">{{ item.value.toLocaleString() }}</div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </template>
    <div v-else class="summary-grid">
      <div v-for="item in items" :key="item.key" class="summary-tile">
        <div class="summary-tile__icon"><SvgIcon :name="item.icon" :size="22" /></div>
        <div>
          <div class="summary-tile__label">{{ item.label }}</div>
          <div class="summary-tile__value">{{ item.value.toLocaleString() }}</div>
        </div>
      </div>
    </div>

    <p v-if="!multiServer" class="muted-text">最近加载：{{ loadedAt }}</p>
  </UiCard>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { MasterdataCounts, MasterdataSummary, PublicServerProfile } from "../api/types";
import SvgIcon, { type IconName } from "./icons/SvgIcon.vue";
import UiAlert from "./ui/UiAlert.vue";
import UiBadge from "./ui/UiBadge.vue";
import UiCard from "./ui/UiCard.vue";
import UiSkeleton from "./ui/UiSkeleton.vue";

const props = defineProps<{
	summary: MasterdataSummary | null;
	servers?: PublicServerProfile[] | null;
	loading?: boolean;
	error?: string;
}>();

const multiServer = computed(() => Array.isArray(props.servers));

const serverEntries = computed(() =>
	(props.servers ?? []).filter((entry) => entry?.enabled !== false),
);
const loadedRegionCount = computed(
	() => serverEntries.value.filter((entry) => entry.loaded).length,
);
const anyLoaded = computed(() => loadedRegionCount.value > 0);

const counts = computed(
	() =>
		props.summary?.counts ?? {
			cards: 0,
			musics: 0,
			events: 0,
			gachas: 0,
			virtual_lives: 0,
		},
);

function countItems(c: MasterdataCounts | undefined) {
	const v = c ?? { cards: 0, musics: 0, events: 0, gachas: 0, virtual_lives: 0 };
	return [
		{ key: "cards", label: "卡牌数量", icon: "preview" as IconName, value: v.cards ?? 0 },
		{ key: "musics", label: "曲目数量", icon: "resources" as IconName, value: v.musics ?? 0 },
		{ key: "events", label: "活动数量", icon: "clock" as IconName, value: v.events ?? 0 },
		{ key: "gachas", label: "卡池数量", icon: "sparkle" as IconName, value: v.gachas ?? 0 },
		{ key: "virtual_lives", label: "演唱会数量", icon: "clock" as IconName, value: v.virtual_lives ?? 0 },
	];
}

const items = computed(() => countItems(counts.value));

function formatTime(value: string | null | undefined) {
	if (!value) return "暂无记录";
	return new Date(value).toLocaleString();
}

const loadedAt = computed(() => formatTime(props.summary?.loaded_at));
</script>

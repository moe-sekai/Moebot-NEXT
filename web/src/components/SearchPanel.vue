<template>
  <UiCard>
    <div class="card-heading">
      <div>
        <h2>Masterdata 搜索测试</h2>
        <p>通过 Go 后端复用 Store.Search* 的只读接口</p>
      </div>
      <UiBadge variant="secondary">只读 API</UiBadge>
    </div>

    <div class="search-row">
      <select v-model="type" class="ui-select" aria-label="搜索类型">
        <option v-for="option in typeOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
      </select>
      <input
        v-model="keyword"
        class="ui-input"
        placeholder="输入 ID、卡名、曲名、活动名或卡池关键词"
        @keyup.enter="runSearch"
      />
      <UiButton :loading="loading" @click="runSearch">搜索</UiButton>
    </div>

    <UiAlert v-if="error" variant="destructive" title="搜索失败">{{ error }}</UiAlert>
    <div v-else-if="searched && rows.length === 0" class="empty-state compact">
      <div class="empty-state__icon"><SvgIcon name="search" :size="22" /></div>
      <p>{{ message || '没有找到匹配结果。' }}</p>
    </div>
    <div v-else-if="rows.length > 0" class="table-wrap">
      <table class="ui-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>类型</th>
            <th>名称</th>
            <th>摘要</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rows" :key="`${row.type}-${row.id}`">
            <td>{{ row.id }}</td>
            <td><UiBadge variant="outline">{{ typeLabel(row.type) }}</UiBadge></td>
            <td class="font-medium">{{ row.title }}</td>
            <td>{{ row.subtitle }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <UiAlert v-else variant="info">
      可以直接试试「miku」「Tell Your World」「限定」「马拉松」或数字 ID。
    </UiAlert>
  </UiCard>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useMessage } from "naive-ui";
import { searchMasterdata } from "../api/client";
import type { SearchResult, SearchType } from "../api/types";
import SvgIcon from "./icons/SvgIcon.vue";
import UiAlert from "./ui/UiAlert.vue";
import UiBadge from "./ui/UiBadge.vue";
import UiButton from "./ui/UiButton.vue";
import UiCard from "./ui/UiCard.vue";

const messageApi = useMessage();
const type = ref<SearchType>("cards");
const keyword = ref("");
const loading = ref(false);
const error = ref("");
const message = ref("");
const searched = ref(false);
const rows = ref<SearchResult[]>([]);

const typeOptions = [
	{ label: "卡牌", value: "cards" },
	{ label: "曲目", value: "musics" },
	{ label: "活动", value: "events" },
	{ label: "卡池", value: "gachas" },
	{ label: "演唱会", value: "virtual-lives" },
];

async function runSearch() {
	const q = keyword.value.trim();
	searched.value = true;
	error.value = "";
	message.value = "";
	rows.value = [];

	if (!q) {
		messageApi.warning("请先输入搜索关键词");
		return;
	}

	loading.value = true;
	try {
		const result = await searchMasterdata(type.value, q);
		rows.value = result.data ?? [];
		message.value = result.message;
	} catch (err) {
		error.value =
			err instanceof Error ? err.message : "搜索失败，请检查后端 API。";
	} finally {
		loading.value = false;
	}
}

function typeLabel(value: string) {
	const labels: Record<string, string> = {
		card: "卡牌",
		music: "曲目",
		event: "活动",
		gacha: "卡池",
	};
	return labels[value] ?? value;
}
</script>

<template>
  <main class="page-stack">
    <PageHeader eyebrow="Runtime Status" title="运行状态" subtitle="完整检查 Bot、Web、Renderer、Masterdata 与数据库链路。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadStatus">刷新状态</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="状态加载失败">{{ error }}</UiAlert>

    <div v-if="loading" class="status-grid">
      <UiSkeleton v-for="item in 5" :key="item" height="148px" />
    </div>
    <div v-else class="status-grid">
      <StatusCard title="Bot" icon="bot" :ok="status?.bot.ok" :status="status?.bot.status" :message="status?.bot.message" :meta="botMeta" />
      <StatusCard title="Web" icon="web" :ok="status?.web.ok" :status="status?.web.status" :message="status?.web.message" :meta="webMeta" />
      <StatusCard title="Renderer" icon="renderer" :ok="status?.renderer.ok" :status="status?.renderer.status" :message="status?.renderer.message" :meta="rendererMeta" />
      <StatusCard title="Masterdata" icon="masterdata" :ok="status?.masterdata.ok" :status="status?.masterdata.status" :message="status?.masterdata.message" :meta="masterdataMeta" />
      <StatusCard title="Database" icon="database" :ok="status?.database.ok" :status="status?.database.status" :message="status?.database.message" :meta="status?.database.path" />
    </div>

    <div class="dashboard-grid dashboard-grid--main">
      <UiCard>
        <div class="card-heading">
          <div>
            <h2>运行时详情</h2>
            <p>来自 /api/status 的聚合状态。</p>
          </div>
          <UiBadge variant="secondary">v{{ status?.version ?? '0.1.0' }}</UiBadge>
        </div>
        <dl class="info-list">
          <div><dt>当前时间</dt><dd>{{ formatTime(status?.time) }}</dd></div>
          <div><dt>运行时长</dt><dd>{{ status?.uptime ?? '-' }}</dd></div>
          <div><dt>命令前缀</dt><dd>{{ status?.bot.command_prefix ?? '-' }}</dd></div>
          <div><dt>昵称</dt><dd>{{ status?.bot.nicknames?.join(' / ') || '-' }}</dd></div>
          <div><dt>Masterdata 加载</dt><dd>{{ formatTime(status?.masterdata.loaded_at) }}</dd></div>
        </dl>
      </UiCard>

      <UiCard>
        <div class="card-heading">
          <div>
            <h2>Renderer 健康检查</h2>
            <p>来自 /api/renderer/health 的实时探测。</p>
          </div>
          <UiBadge :variant="rendererHealth?.ok ? 'success' : 'destructive'">{{ rendererHealth?.status ?? 'unknown' }}</UiBadge>
        </div>
        <dl class="info-list">
          <div><dt>地址</dt><dd>{{ rendererHealth?.base_url ?? '-' }}</dd></div>
          <div><dt>HTTP 状态</dt><dd>{{ rendererHealth?.status_code ?? '-' }}</dd></div>
          <div><dt>探测耗时</dt><dd>{{ rendererHealth?.latency_ms ?? '-' }} ms</dd></div>
          <div><dt>Renderer 端口</dt><dd>{{ rendererHealth?.renderer_port ?? '-' }}</dd></div>
          <div><dt>面板端口</dt><dd>{{ rendererHealth?.dashboard_port ?? '-' }}</dd></div>
          <div><dt>说明</dt><dd>{{ rendererHealth?.note ?? rendererHealth?.message ?? '-' }}</dd></div>
        </dl>
      </UiCard>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { getRendererHealth, getStatus } from "../api/client";
import type { RendererHealth, RuntimeStatus } from "../api/types";
import PageHeader from "../components/PageHeader.vue";
import StatusCard from "../components/StatusCard.vue";
import UiAlert from "../components/ui/UiAlert.vue";
import UiBadge from "../components/ui/UiBadge.vue";
import UiButton from "../components/ui/UiButton.vue";
import UiCard from "../components/ui/UiCard.vue";
import UiSkeleton from "../components/ui/UiSkeleton.vue";

const status = ref<RuntimeStatus | null>(null);
const rendererHealth = ref<RendererHealth | null>(null);
const loading = ref(false);
const error = ref("");

const botMeta = computed(() =>
	status.value
		? `${status.value.bot.driver_type} · ${status.value.bot.listen || "未配置监听地址"}`
		: "-",
);
const webMeta = computed(() =>
	status.value ? `${status.value.web.host}:${status.value.web.port}` : "-",
);
const rendererMeta = computed(() =>
	status.value
		? `${status.value.renderer.base_url} · ${status.value.renderer.latency_ms} ms`
		: "-",
);
const masterdataMeta = computed(() => {
	const counts = status.value?.masterdata.counts;
	return counts
		? `卡牌 ${counts.cards} / 曲目 ${counts.musics} / 活动 ${counts.events} / 卡池 ${counts.gachas} / 演唱会 ${counts.virtual_lives ?? 0}`
		: "-";
});

onMounted(loadStatus);

async function loadStatus() {
	loading.value = true;
	error.value = "";
	try {
		const [statusData, rendererData] = await Promise.all([
			getStatus(),
			getRendererHealth(),
		]);
		status.value = statusData;
		rendererHealth.value = rendererData;
	} catch (err) {
		error.value = err instanceof Error ? err.message : "加载运行状态失败。";
	} finally {
		loading.value = false;
	}
}

function formatTime(value?: string | null) {
	return value ? new Date(value).toLocaleString() : "-";
}
</script>

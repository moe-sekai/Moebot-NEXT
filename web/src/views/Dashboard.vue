<template>
  <main class="page-stack">
    <PageHeader eyebrow="Dashboard" title="概览" subtitle="服务状态、版本信息与最近活动集中展示。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadOverview">刷新概览</UiButton>
      </template>
    </PageHeader>

    <section class="hero-panel">
      <div>
        <UiBadge variant="secondary">Moebot NEXT Console</UiBadge>
        <h2>柔和、清晰、可扩展的管理控制台</h2>
        <p>Go + ZeroBot + Fiber + Satori Renderer，专注呈现机器人运行、数据加载与渲染链路状态。</p>
        <div class="hero-panel__meta">
          <UiBadge variant="outline">管理面板 {{ webPortLabel }}</UiBadge>
          <UiBadge variant="outline">Renderer {{ rendererUrl }}</UiBadge>
          <UiBadge variant="outline">v{{ status?.version ?? health?.version ?? '0.1.0' }}</UiBadge>
        </div>
      </div>
      <div class="hero-panel__aside">
        <MetricCard label="Renderer" :value="status?.renderer.ok ? 'Ready' : 'Check'" :hint="rendererLatency" icon="renderer" />
        <MetricCard label="Masterdata" :value="status?.masterdata.loaded ? 'Loaded' : 'Pending'" :hint="masterdataCountLabel" icon="masterdata" />
      </div>
    </section>

    <UiAlert v-if="pageError" variant="destructive" title="状态加载失败">{{ pageError }}</UiAlert>

    <div v-if="loading" class="status-grid">
      <UiSkeleton v-for="item in 5" :key="item" height="148px" />
    </div>
    <div v-else class="status-grid">
      <StatusCard title="Bot 状态" icon="bot" :ok="status?.bot.ok" :status="status?.bot.status" :message="status?.bot.message" :meta="botMeta" />
      <StatusCard title="Web 状态" icon="web" :ok="status?.web.ok" :status="status?.web.status" :message="status?.web.message" :meta="webMeta" />
      <StatusCard title="Renderer 状态" icon="renderer" :ok="status?.renderer.ok" :status="status?.renderer.status" :message="status?.renderer.message" :meta="rendererMeta" />
      <StatusCard title="Masterdata 状态" icon="masterdata" :ok="status?.masterdata.ok" :status="status?.masterdata.status" :message="status?.masterdata.message" :meta="masterdataMeta" />
      <StatusCard title="Database 状态" icon="database" :ok="status?.database.ok" :status="status?.database.status" :message="status?.database.message" :meta="status?.database.path" />
    </div>

    <div class="dashboard-grid dashboard-grid--main">
      <MasterdataSummary :summary="summary" :loading="loading" :error="summaryError" />

      <UiCard class-name="renderer-info">
        <div class="card-heading">
          <div>
            <h2>Renderer 信息</h2>
            <p>Satori 渲染服务健康检查</p>
          </div>
          <UiBadge :variant="rendererHealth?.ok ? 'success' : 'destructive'">{{ rendererHealth?.ok ? '可用' : '不可用' }}</UiBadge>
        </div>
        <UiAlert v-if="rendererError" variant="destructive" title="检查失败">{{ rendererError }}</UiAlert>
        <dl v-else class="info-list">
          <div><dt>Renderer 地址</dt><dd>{{ rendererUrl }}</dd></div>
          <div><dt>健康状态</dt><dd>{{ rendererHealth?.message ?? '等待检查' }}</dd></div>
          <div><dt>响应耗时</dt><dd>{{ rendererHealth?.latency_ms ?? status?.renderer.latency_ms ?? 0 }} ms</dd></div>
          <div><dt>服务说明</dt><dd>{{ rendererHealth?.note ?? '渲染服务与管理面板端口相互独立。' }}</dd></div>
        </dl>
      </UiCard>
    </div>

    <div class="dashboard-grid dashboard-grid--bottom">
      <RecentCommands
        :commands="recentCommands"
        :loading="recentLoading"
        :error="recentError"
        :message="recentMessage"
        @refresh="loadRecentCommands"
      />
      <CommandList />
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
	getHealth,
	getMasterdataSummary,
	getPublicConfig,
	getRecentCommands,
	getRendererHealth,
	getStatus,
} from "../api/client";
import type {
	HealthResponse,
	MasterdataSummary as MasterdataSummaryData,
	PublicConfig,
	RecentCommand,
	RendererHealth,
	RuntimeStatus,
} from "../api/types";
import CommandList from "../components/CommandList.vue";
import MasterdataSummary from "../components/MasterdataSummary.vue";
import MetricCard from "../components/MetricCard.vue";
import PageHeader from "../components/PageHeader.vue";
import RecentCommands from "../components/RecentCommands.vue";
import StatusCard from "../components/StatusCard.vue";
import UiAlert from "../components/ui/UiAlert.vue";
import UiBadge from "../components/ui/UiBadge.vue";
import UiButton from "../components/ui/UiButton.vue";
import UiCard from "../components/ui/UiCard.vue";
import UiSkeleton from "../components/ui/UiSkeleton.vue";

const health = ref<HealthResponse | null>(null);
const status = ref<RuntimeStatus | null>(null);
const summary = ref<MasterdataSummaryData | null>(null);
const rendererHealth = ref<RendererHealth | null>(null);
const publicConfig = ref<PublicConfig | null>(null);
const recentCommands = ref<RecentCommand[]>([]);
const recentMessage = ref("");

const loading = ref(false);
const recentLoading = ref(false);
const pageError = ref("");
const summaryError = ref("");
const rendererError = ref("");
const recentError = ref("");

const webPortLabel = computed(() => {
	const port = publicConfig.value?.web.port ?? status.value?.web.port ?? 8080;
	return `:${port}`;
});

const rendererUrl = computed(() => {
	return (
		publicConfig.value?.renderer.base_url ||
		status.value?.renderer.base_url ||
		"http://127.0.0.1:3001"
	);
});

const rendererLatency = computed(() => {
	const latency =
		rendererHealth.value?.latency_ms ?? status.value?.renderer.latency_ms;
	return typeof latency === "number" ? `${latency} ms` : "等待检测";
});

const botMeta = computed(() => {
	if (!status.value) return "OneBot v11 反向 WebSocket 默认监听 :6700";
	return `${status.value.bot.driver_type} · ${status.value.bot.listen || "未配置监听地址"}`;
});

const webMeta = computed(() => {
	if (!status.value) return "Fiber + Vue 3 + Vite";
	return `${status.value.web.host}:${status.value.web.port}`;
});

const rendererMeta = computed(() => {
	if (!status.value) return rendererUrl.value;
	return `${status.value.renderer.base_url} · ${status.value.renderer.latency_ms} ms`;
});

const masterdataMeta = computed(() => masterdataCountLabel.value);

const masterdataCountLabel = computed(() => {
	const counts = status.value?.masterdata.counts ?? summary.value?.counts;
	if (!counts) return "等待数据加载";
	return `卡牌 ${counts.cards} / 曲目 ${counts.musics} / 活动 ${counts.events} / 卡池 ${counts.gachas} / 演唱会 ${counts.virtual_lives ?? 0}`;
});

onMounted(async () => {
	await Promise.all([loadOverview(), loadRecentCommands()]);
});

async function loadOverview() {
	loading.value = true;
	pageError.value = "";
	summaryError.value = "";
	rendererError.value = "";
	try {
		const [healthData, statusData, configData, summaryData, rendererData] =
			await Promise.all([
				getHealth(),
				getStatus(),
				getPublicConfig(),
				getMasterdataSummary(),
				getRendererHealth(),
			]);
		health.value = healthData;
		status.value = statusData;
		publicConfig.value = configData;
		summary.value = summaryData;
		rendererHealth.value = rendererData;
	} catch (err) {
		pageError.value = normalizeError(err, "加载运行状态失败");
	} finally {
		loading.value = false;
	}
}

async function loadRecentCommands() {
	recentLoading.value = true;
	recentError.value = "";
	try {
		const result = await getRecentCommands(8);
		recentCommands.value = result.data ?? [];
		recentMessage.value = result.message;
	} catch (err) {
		recentError.value = normalizeError(err, "加载最近命令失败");
	} finally {
		recentLoading.value = false;
	}
}

function normalizeError(err: unknown, fallback: string) {
	return err instanceof Error ? `${fallback}：${err.message}` : fallback;
}
</script>

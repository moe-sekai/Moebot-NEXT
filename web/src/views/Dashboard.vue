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

    <UiAlert
      v-if="thumbnailPreloadIncomplete"
      variant="warning"
      title="渲染服务预载未完成"
    >
      当前 {{ thumbnailCache?.region_label ?? '默认区服' }} 卡牌缩略图覆盖率仅 {{ thumbnailCacheProgress }}%（{{ thumbnailCache?.cached ?? 0 }}/{{ thumbnailCache?.total ?? 0 }}），可能导致部分渲染缺图。
      <span style="display: block; margin-top: 4px; opacity: 0.75;">超过 99% 后此提醒会自动消失（少量素材无法加载属正常情况）。</span>
      <div class="thumbnail-preload-actions">
        <UiButton size="sm" :loading="preloadStarting" :disabled="preloadStarting" @click="startPreloadFromDashboard">
          {{ preloadStarting ? '正在启动…' : '立即预载缩略图' }}
        </UiButton>
        <UiButton size="sm" variant="outline" @click="goToRendererSettings">前往设置</UiButton>
        <span v-if="preloadHint" class="thumbnail-preload-hint">{{ preloadHint }}</span>
      </div>
    </UiAlert>

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

  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref } from "vue";
import { useRouter } from "vue-router";
import {
	getHealth,
	getPublicConfig,
	getRendererCardThumbnailCacheStatus,
	getRendererHealth,
	getStatus,
	preloadRendererCardThumbnails,
} from "../api/client";
import type {
	HealthResponse,
	PublicConfig,
	RendererCardThumbnailCacheStatus,
	RendererHealth,
	RuntimeStatus,
} from "../api/types";
import MetricCard from "../components/MetricCard.vue";
import PageHeader from "../components/PageHeader.vue";
import StatusCard from "../components/StatusCard.vue";
import UiAlert from "../components/ui/UiAlert.vue";
import UiBadge from "../components/ui/UiBadge.vue";
import UiButton from "../components/ui/UiButton.vue";
import UiCard from "../components/ui/UiCard.vue";
import UiSkeleton from "../components/ui/UiSkeleton.vue";

const health = ref<HealthResponse | null>(null);
const status = ref<RuntimeStatus | null>(null);
const rendererHealth = ref<RendererHealth | null>(null);
const publicConfig = ref<PublicConfig | null>(null);
const thumbnailCache = ref<RendererCardThumbnailCacheStatus | null>(null);

const loading = ref(false);
const pageError = ref("");
const rendererError = ref("");
const preloadStarting = ref(false);
const preloadHint = ref("");
let thumbnailPollTimer: ReturnType<typeof window.setTimeout> | null = null;

const router = useRouter();

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

const thumbnailCacheProgress = computed(() => {
	const status = thumbnailCache.value;
	if (!status) return 0;
	const sourceTotal = status.total ?? 0;
	const compositeTotal = status.composite_total ?? 0;
	const total = sourceTotal + compositeTotal;
	if (total <= 0) return 0;
	const cached = (status.cached ?? 0) + (status.composite_cached ?? 0);
	return Math.round((cached / total) * 100);
});

const thumbnailPreloadIncomplete = computed(() => {
	const status = thumbnailCache.value;
	if (!status) return false;
	if (!status.enabled) return false;
	if (status.running) return false;
	if ((status.total ?? 0) <= 0 && (status.composite_total ?? 0) <= 0) return false;
	return thumbnailCacheProgress.value < 99;
});

const masterdataServerProfiles = computed(() => {
	const servers = publicConfig.value?.servers;
	if (!servers) return null;
	const regions = publicConfig.value?.presets.regions ?? [];
	const order = new Map(regions.map((r, i) => [r.key, i] as const));
	return Object.values(servers)
		.filter((entry) => entry?.enabled !== false)
		.sort((a, b) => (order.get(a.region) ?? 99) - (order.get(b.region) ?? 99));
});

const masterdataCountLabel = computed(() => {
	const profiles = masterdataServerProfiles.value;
	if (profiles && profiles.length > 0) {
		const loaded = profiles.filter((p) => p.loaded).length;
		return `${loaded}/${profiles.length} 区服已加载·点击刷新查看详情`;
	}
	const counts = status.value?.masterdata.counts;
	if (!counts) return "等待数据加载";
	return `卡牌 ${counts.cards} / 曲目 ${counts.musics} / 活动 ${counts.events} / 卡池 ${counts.gachas} / 演唱会 ${counts.virtual_lives ?? 0}`;
});

onMounted(async () => {
	await loadOverview();
});

async function loadOverview() {
	loading.value = true;
	pageError.value = "";
	rendererError.value = "";
	try {
		const [healthData, statusData, configData, rendererData] =
			await Promise.all([
				getHealth(),
				getStatus(),
				getPublicConfig(),
				getRendererHealth(),
			]);
		health.value = healthData;
		status.value = statusData;
		publicConfig.value = configData;
		rendererHealth.value = rendererData;
		void loadThumbnailCacheStatus();
	} catch (err) {
		pageError.value = normalizeError(err, "加载运行状态失败");
	} finally {
		loading.value = false;
	}
}

async function loadThumbnailCacheStatus() {
	try {
		const data = await getRendererCardThumbnailCacheStatus();
		thumbnailCache.value = data;
		if (data?.running) scheduleThumbnailPoll();
	} catch {
		thumbnailCache.value = null;
	}
}

function scheduleThumbnailPoll() {
	clearThumbnailPoll();
	thumbnailPollTimer = window.setTimeout(async () => {
		thumbnailPollTimer = null;
		await loadThumbnailCacheStatus();
	}, 2000);
}

function clearThumbnailPoll() {
	if (thumbnailPollTimer) {
		window.clearTimeout(thumbnailPollTimer);
		thumbnailPollTimer = null;
	}
}

async function startPreloadFromDashboard() {
	if (preloadStarting.value) return;
	preloadStarting.value = true;
	preloadHint.value = "";
	try {
		const data = await preloadRendererCardThumbnails();
		thumbnailCache.value = data;
		preloadHint.value = `已开始预载：${data.cached}/${data.total}，可在「设置」查看进度。`;
		scheduleThumbnailPoll();
	} catch (err) {
		preloadHint.value = normalizeError(err, "启动预载失败");
	} finally {
		preloadStarting.value = false;
	}
}

function goToRendererSettings() {
	router.push({ path: "/plugins/moesekai/advanced", hash: "#renderer-cache" });
}

onBeforeUnmount(() => {
	clearThumbnailPoll();
});

function normalizeError(err: unknown, fallback: string) {
	return err instanceof Error ? `${fallback}：${err.message}` : fallback;
}
</script>

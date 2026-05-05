<template>
  <main class="page-stack">
    <PageHeader eyebrow="Settings" title="设置" subtitle="按功能配置区服、Masterdata 数据源、Assets 资源源与接口；敏感字段不会在控制台暴露。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadConfig">刷新配置</UiButton>
        <UiButton size="sm" :loading="saving" :disabled="!dirty || !canSave" @click="saveSettings">保存设置</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="success" variant="info" title="操作完成">{{ success }}</UiAlert>
    <UiAlert v-if="error" variant="destructive" title="配置操作失败">{{ error }}</UiAlert>

    <div v-if="!loading" class="settings-save-bar">
      <div>
        <strong>{{ dirty ? '有未保存设置' : '设置已同步' }}</strong>
        <span>{{ canSave ? '保存设置会写入当前表单配置。' : '请先修正无效配置。' }}</span>
      </div>
      <div class="settings-save-bar__actions">
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadConfig">刷新配置</UiButton>
        <UiButton size="sm" :loading="saving" :disabled="!dirty || !canSave" @click="saveSettings">保存设置</UiButton>
      </div>
    </div>

    <div v-if="loading" class="settings-function-stack">
      <UiSkeleton v-for="item in 5" :key="item" height="260px" />
    </div>
    <template v-else>
      <div class="settings-function-stack">
        <!-- 区服详细配置 Tabs -->
        <UiCard className="settings-card settings-function-card">
          <div class="settings-card__heading">
            <div class="settings-card__icon"><SvgIcon name="masterdata" :size="22" /></div>
            <div>
              <h2>区服详细配置</h2>
              <p>请选择一个区服标签页进行启用状态、默认区服、Masterdata、Assets 及接口功能的详细配置。</p>
            </div>
          </div>

          <div class="ui-tabs">
            <div class="ui-tabs-list">
              <button
                v-for="entry in serverEntries"
                :key="entry.option.key"
                class="ui-tabs-trigger"
                :data-state="activeTab === entry.option.key ? 'active' : ''"
                @click="activeTab = entry.option.key"
              >
                {{ entry.option.label }} · {{ entry.option.key.toUpperCase() }}
                <span v-if="entry.option.key === form.server.region" style="opacity: 0.7; font-size: 12px; margin-left: 4px;">(默认)</span>
              </button>
            </div>

            <div
              v-for="entry in serverEntries"
              :key="entry.option.key"
              class="ui-tabs-content"
              :data-state="activeTab === entry.option.key ? 'active' : ''"
            >
              <div class="settings-function-stack" style="margin-top: 6px;">
                
                <!-- 1. 基础状态 -->
                <div class="settings-region-row settings-region-row--compact">
                  <div class="settings-row-header">
                    <div>
                      <h3>基础状态</h3>
                      <p>命令前缀 /{{ entry.option.key }}，例如 /{{ entry.option.key }}绑定、/{{ entry.option.key }}查曲。</p>
                    </div>
                    <div class="settings-row-badges">
                      <UiBadge v-if="entry.option.key === form.server.region" variant="default">默认</UiBadge>
                      <UiBadge :variant="entry.form.enabled ? 'success' : 'outline'">{{ entry.form.enabled ? '启用' : '停用' }}</UiBadge>
                      <UiBadge :variant="entry.state?.loaded ? 'success' : 'warning'">{{ entry.state?.loaded ? '已加载' : '未加载' }}</UiBadge>
                      <UiButton
                        variant="outline"
                        size="sm"
                        :disabled="entry.option.key === form.server.region"
                        @click="() => setDefaultRegion(entry.option.key)"
                      >
                        {{ entry.option.key === form.server.region ? '当前默认区服' : '设为默认区服' }}
                      </UiButton>
                    </div>
                  </div>
                  <div class="settings-row-body settings-row-body--inline">
                    <label class="settings-field">
                      <span>启用状态</span>
                      <select v-model="entry.form.enabled" class="ui-select" :disabled="isRegionLocked(entry.option.key)">
                        <option :value="true">启用</option>
                        <option :value="false">停用</option>
                      </select>
                    </label>
                    <div class="settings-row-meta">
                      <div><span>数据量</span><strong>{{ countsText(entry.state?.counts) }}</strong></div>
                      <div><span>加载时间</span><strong>{{ formatTime(entry.state?.loaded_at) }}</strong></div>
                      <div><span>说明</span><strong>{{ isRegionLocked(entry.option.key) ? 'JP / 默认区服固定启用' : '可按需启用或停用' }}</strong></div>
                    </div>
                  </div>
                </div>

                <!-- 2. Masterdata -->
                <div class="settings-region-row">
                  <div class="settings-row-header">
                    <div>
                      <h3>Masterdata 数据源</h3>
                      <p>{{ entry.form.enabled ? '配置卡牌、曲目等基础数据来源。该区服已启用。' : '该区服暂未启用，仍可先配置数据源。' }}</p>
                    </div>
                    <div class="settings-row-badges">
                      <UiBadge :variant="masterdataProfileSupported(entry.form) ? 'success' : 'destructive'">{{ masterdataProfileSupported(entry.form) ? '可用' : '不可用' }}</UiBadge>
                      <UiBadge variant="secondary">{{ sourceLabel(masterdataSourceOptions, entry.form.masterdata.source) }}</UiBadge>
                    </div>
                  </div>

                  <div class="settings-form settings-form--region">
                    <label class="settings-field">
                      <span>数据来源</span>
                      <select v-model="entry.form.masterdata.source" class="ui-select" @change="() => normalizeServerProfile(entry.option.key)">
                        <option v-for="option in masterdataSourceOptions" :key="option.key" :value="option.key" :disabled="!optionAvailable(option, entry.option.key)">
                          {{ option.label }}
                        </option>
                      </select>
                    </label>
                    <label class="settings-field">
                      <span>本地缓存路径</span>
                      <input v-model.trim="entry.form.masterdata.local_path" class="ui-input" placeholder="./data/master/jp" />
                    </label>
                    <label class="settings-field">
                      <span>刷新间隔（秒）</span>
                      <input v-model.number="entry.form.masterdata.refresh_interval" class="ui-input" type="number" min="0" />
                    </label>
                    <div class="settings-field settings-field--readonly">
                      <span>支持区服</span>
                      <strong>{{ sourceSupportText(masterdataSourceOptions, entry.form.masterdata.source) }}</strong>
                    </div>
                    <label v-if="entry.form.masterdata.source === 'custom'" class="settings-field settings-field--full">
                      <span>自定义主 URL</span>
                      <input v-model.trim="entry.form.masterdata.custom_url" class="ui-input" placeholder="https://example.com/master" />
                    </label>
                    <label v-if="entry.form.masterdata.source === 'custom'" class="settings-field settings-field--full">
                      <span>自定义备用 URL</span>
                      <input v-model.trim="entry.form.masterdata.custom_fallback_url" class="ui-input" placeholder="可选" />
                    </label>
                    <UiAlert v-if="!masterdataProfileSupported(entry.form)" variant="warning" title="组合不可用">
                      {{ sourceLabel(masterdataSourceOptions, entry.form.masterdata.source) }} 暂不支持 {{ entry.option.label }}，保存前请更换来源。
                    </UiAlert>
                  </div>

                  <div class="settings-preview">
                    <div><span>加载状态</span><code>{{ entry.state?.loaded ? '已加载' : '未加载' }} · {{ countsText(entry.state?.counts) }}</code></div>
                    <div><span>Master URL</span><code>{{ masterdataPreview(entry, 'primary') }}</code></div>
                    <div><span>备用 URL</span><code>{{ masterdataPreview(entry, 'fallback') }}</code></div>
                    <div><span>缓存路径</span><code>{{ entry.form.masterdata.local_path || '-' }}</code></div>
                  </div>
                  <div class="settings-actions-row">
                    <UiButton variant="outline" size="sm" :loading="reloading === entry.option.key" :disabled="!entry.form.enabled" @click="() => reloadMasterdataNow(entry.option.key)">重载该服 Masterdata</UiButton>
                    <span class="settings-hint">{{ masterdataHint(entry) }}</span>
                  </div>
                </div>

                <!-- 3. Assets -->
                <div class="settings-region-row">
                  <div class="settings-row-header">
                    <div>
                      <h3>Assets 资源源</h3>
                      <p>配置卡面、活动图、谱面相关图片等资源来源。</p>
                    </div>
                    <div class="settings-row-badges">
                      <UiBadge :variant="assetProfileSupported(entry.form) ? 'success' : 'destructive'">{{ assetProfileSupported(entry.form) ? '可用' : '不可用' }}</UiBadge>
                      <UiBadge variant="secondary">{{ sourceLabel(assetSourceOptions, entry.form.assets.source) }}</UiBadge>
                    </div>
                  </div>

                  <div class="settings-form settings-form--region">
                    <label class="settings-field">
                      <span>资源来源</span>
                      <select v-model="entry.form.assets.source" class="ui-select" @change="() => normalizeServerProfile(entry.option.key)">
                        <option v-for="option in assetSourceOptions" :key="option.key" :value="option.key" :disabled="!optionAvailable(option, entry.option.key)">
                          {{ option.label }}
                        </option>
                      </select>
                    </label>
                    <label v-if="entry.form.assets.source === 'moesekai'" class="settings-field">
                      <span>MoeSekai 镜像</span>
                      <select v-model="entry.form.assets.mirror" class="ui-select">
                        <option v-for="option in assetMirrorOptions" :key="option.key" :value="option.key">{{ option.label }}</option>
                      </select>
                    </label>
                    <div v-else class="settings-field settings-field--readonly">
                      <span>镜像</span>
                      <strong>{{ entry.form.assets.source === 'custom' ? '由自定义 URL 决定' : 'sekai.best 自动选择' }}</strong>
                    </div>
                    <label class="settings-field settings-field--full">
                      <span>曲名别名 URL</span>
                      <input v-model.trim="entry.form.assets.music_alias_url" class="ui-input" placeholder="https://.../music_aliases.json" />
                    </label>
                    <label class="settings-field settings-field--full">
                      <span>谱面来源</span>
                      <input v-model.trim="entry.form.assets.chart_source_url" class="ui-input" placeholder="https://charts-new.unipjsk.com/moe/svg/{id}/{difficulty}.svg" />
                    </label>
                    <label class="settings-field">
                      <span>贴纸路径</span>
                      <input v-model.trim="entry.form.assets.sticker_path" class="ui-input" placeholder="./assets/stickers" />
                    </label>
                    <div class="settings-field settings-field--readonly">
                      <span>支持区服</span>
                      <strong>{{ sourceSupportText(assetSourceOptions, entry.form.assets.source) }}</strong>
                    </div>
                    <label v-if="entry.form.assets.source === 'custom'" class="settings-field settings-field--full">
                      <span>自定义 Base URL</span>
                      <input v-model.trim="entry.form.assets.custom_base_url" class="ui-input" placeholder="https://example.com/sekai-jp-assets" />
                    </label>
                    <UiAlert v-if="!assetProfileSupported(entry.form)" variant="warning" title="组合不可用">
                      {{ sourceLabel(assetSourceOptions, entry.form.assets.source) }} 暂不支持 {{ entry.option.label }}，保存前请更换来源。
                    </UiAlert>
                  </div>

                  <div class="settings-preview">
                    <div><span>Base URL</span><code>{{ assetsPreview(entry, 'base') }}</code></div>
                    <div><span>Renderer Source</span><code>{{ assetsPreview(entry, 'renderer') }}</code></div>
                    <div><span>贴纸路径</span><code>{{ entry.form.assets.sticker_path || '-' }}</code></div>
                    <div><span>曲名别名</span><code>{{ entry.form.assets.music_alias_url || '-' }}</code></div>
                    <div><span>谱面来源</span><code>{{ entry.form.assets.chart_source_url || '-' }}</code></div>
                  </div>
                </div>

                <!-- 4. API -->
                <div class="settings-region-row settings-region-row--compact">
                  <div class="settings-row-header">
                    <div>
                      <h3>接口功能</h3>
                      <p>SEKAI / Suite 均通过端点与自定义请求头配置；Suite 使用 Haruki 公开 API；Ranking 自动使用 MoeSekai 公开榜线。</p>
                    </div>
                    <div class="settings-row-badges">
                      <UiBadge :variant="entry.form.sekai_api.enabled ? 'success' : 'outline'">SEKAI API {{ entry.form.sekai_api.enabled ? '启用' : '关闭' }}</UiBadge>
                      <UiBadge variant="secondary">Suite · Haruki 公开 API</UiBadge>
                      <UiBadge variant="secondary">Ranking 自动 · MoeSekai</UiBadge>
                      <UiButton variant="outline" size="sm" :loading="Boolean(sekaiTesting[entry.option.key])" @click="() => testSekaiConnectivity(entry)">测试连通性</UiButton>
                    </div>
                  </div>

                  <div class="settings-form settings-form--region">
                    <label class="settings-field">
                      <span>SEKAI API</span>
                      <select v-model="entry.form.sekai_api.enabled" class="ui-select">
                        <option :value="true">启用</option>
                        <option :value="false">关闭</option>
                      </select>
                    </label>
                    <label class="settings-field settings-field--full">
                      <span>SEKAI API Base URL</span>
                      <input v-model.trim="entry.form.sekai_api.base_url" class="ui-input" placeholder="https://seka-api.exmeaning.com 或 https://example.com/api/{region}" />
                    </label>
                    <label class="settings-field settings-field--full">
                      <span>SEKAI API Headers（JSON）</span>
                      <textarea v-model.trim="entry.form.sekai_api.headers_text" class="ui-input ui-textarea" placeholder='留空保持现有请求头；如需 token 请写请求头，例如 { "x-moe-sekai-token": "..." }（值不加引号也可）；{} 清空'></textarea>
                    </label>
                    <div class="settings-field settings-field--readonly settings-field--full">
                      <span>System 测试 URL</span>
                      <strong>测试 {{ sekaiSystemPreview(entry) }} 的连通性</strong>
                    </div>
                    <UiAlert v-if="sekaiTestResults[entry.option.key]" class="settings-field--full" :variant="sekaiTestResults[entry.option.key]?.ok ? 'info' : 'warning'" :title="sekaiTestResults[entry.option.key]?.ok ? 'SEKAI API 连通正常' : 'SEKAI API 不可用'">
                      {{ sekaiTestResults[entry.option.key]?.message }} · {{ sekaiTestResults[entry.option.key]?.status_code ? `HTTP ${sekaiTestResults[entry.option.key]?.status_code}` : '无 HTTP 状态' }} · {{ sekaiTestResults[entry.option.key]?.duration_ms ?? 0 }}ms
                    </UiAlert>
                    <label class="settings-field">
                      <span>Suite URL（Haruki 公开 API）</span>
                      <input v-model.trim="entry.form.suite_api.url" class="ui-input" placeholder="https://suite-api.haruki.seiunx.com/public/{region}/suite/{uid}" />
                    </label>
                    <label class="settings-field">
                      <span>Suite Headers（JSON）</span>
                      <textarea v-model.trim="entry.form.suite_api.headers_text" class="ui-input ui-textarea" placeholder='留空保持现有请求头；填写 { "Authorization": "Bearer ..." } 更新（值不加引号也可）；{} 清空'></textarea>
                    </label>
                    <div class="settings-field settings-field--readonly">
                      <span>Ranking API</span>
                      <strong>MoeSekai 公开榜线 API，区服自动跟随当前标签：{{ entry.option.label }} · {{ entry.option.key.toUpperCase() }}</strong>
                    </div>
                    <div class="settings-field settings-field--readonly">
                      <span>请求限制</span>
                      <strong>SEKAI {{ entry.form.sekai_api.timeout }}s · {{ entry.form.sekai_api.rate_limit }}/min · Suite {{ entry.form.suite_api.timeout }}s · Ranking {{ entry.form.ranking_api.timeout }}s</strong>
                    </div>
                  </div>
                </div>

              </div>
            </div>
          </div>
        </UiCard>
      </div>

      <div class="settings-grid">
        <ConfigSection title="Bot" description="OneBot 驱动、命令前缀与昵称。" icon="bot" :items="botItems" />
        <UiCard className="settings-card">
          <div class="settings-card__heading">
            <div class="settings-card__icon"><SvgIcon name="renderer" :size="22" /></div>
            <div>
              <h2>Renderer</h2>
              <p>Satori 渲染服务、SVG 转 PNG 精度与缓存配置。</p>
            </div>
          </div>
          <div class="settings-form">
            <label class="settings-field">
              <span>渲染精度</span>
              <input v-model.number="form.renderer.precision" class="ui-input" type="number" min="0.1" step="0.1" @input="normalizeRendererPrecision" />
            </label>
            <label class="settings-field">
              <span>谱面渲染精度</span>
              <input v-model.number="form.renderer.chart_precision" class="ui-input" type="number" min="0.1" step="0.1" @input="normalizeRendererPrecision" />
            </label>
            <div class="settings-field settings-field--readonly">
              <span>说明</span>
              <strong>普通图片约 {{ rendererOutputScaleText }}，谱面约 {{ chartRendererOutputScaleText }}；越高越清晰但图片体积和耗时越大。</strong>
            </div>
          </div>
          <div class="renderer-cache-panel">
            <div class="renderer-cache-panel__header">
              <div>
                <strong>卡牌缩略图预载</strong>
                <span>{{ regionLabel(thumbnailCacheRegion) }} · 统一使用日服缩略图资源 · {{ thumbnailCacheSummary }}</span>
              </div>
              <UiBadge :variant="thumbnailCacheBadgeVariant">{{ thumbnailCacheStatusLabel }}</UiBadge>
            </div>
            <div class="renderer-cache-panel__meter" aria-hidden="true">
              <span :style="{ width: `${thumbnailCacheProgress}%` }" />
            </div>
            <div class="renderer-cache-panel__meta">
              <span>覆盖率 {{ thumbnailCacheProgress }}%</span>
              <span v-if="thumbnailCache?.composite_total">合成 SVG {{ thumbnailCache.composite_cached }}/{{ thumbnailCache.composite_total }}</span>
              <span>缓存目录：{{ thumbnailCache?.cache_dir || config?.renderer.cache.path || '-' }}</span>
            </div>
            <div v-if="thumbnailCache?.errors?.length" class="renderer-cache-panel__errors">
              <span v-for="item in thumbnailCache.errors.slice(0, 3)" :key="item">{{ item }}</span>
            </div>
            <div class="settings-actions-row">
              <UiButton variant="outline" size="sm" :loading="thumbnailCacheLoading" @click="() => refreshThumbnailCacheStatus()">刷新状态</UiButton>
              <UiButton size="sm" :loading="thumbnailPreloadButtonLoading" :disabled="!config?.renderer.cache.enabled" @click="preloadCardThumbnails">预载卡牌缩略图</UiButton>
            </div>
          </div>
          <dl class="settings-list settings-list--compact">
            <div v-for="item in rendererItems" :key="item.label">
              <dt>{{ item.label }}</dt>
              <dd>
                <UiBadge v-if="item.badge" :variant="item.value ? 'success' : 'warning'">{{ item.value ? '是' : '否' }}</UiBadge>
                <template v-else>{{ item.value }}</template>
              </dd>
            </div>
          </dl>
        </UiCard>
        <ConfigSection title="Web" description="Fiber 管理控制台监听配置。" icon="web" :items="webItems" />
        <ConfigSection title="默认区服数据状态" description="当前默认区服生效的数据来源、刷新与本地路径。" icon="masterdata" :items="masterdataItems" />
        <ConfigSection title="默认区服接口状态" description="默认区服玩家资料接口开关、地区与请求头配置状态。" icon="web" :items="sekaiApiItems" />
        <ConfigSection title="默认区服资源状态" description="当前默认区服生效的 CDN、别名与贴纸资源配置。" icon="resources" :items="assetItems" />
      </div>
    </template>
  </main>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onBeforeUnmount, onMounted, ref } from "vue";
import {
	getPublicConfig,
	getRendererCardThumbnailCacheStatus,
	preloadRendererCardThumbnails,
	reloadMasterdata,
	testSekaiSystem,
	updatePublicConfig,
} from "../api/client";
import type {
	ConfigOption,
	MasterdataCounts,
	PublicConfig,
	PublicServerProfile,
	RendererCardThumbnailCacheStatus,
	SekaiSystemTestResponse,
	UpdatePublicConfigPayload,
} from "../api/types";
import SvgIcon, { type IconName } from "../components/icons/SvgIcon.vue";
import PageHeader from "../components/PageHeader.vue";
import UiAlert from "../components/ui/UiAlert.vue";
import UiBadge from "../components/ui/UiBadge.vue";
import UiButton from "../components/ui/UiButton.vue";
import UiCard from "../components/ui/UiCard.vue";
import UiSkeleton from "../components/ui/UiSkeleton.vue";

type BadgeVariant = "default" | "secondary" | "success" | "warning" | "destructive" | "outline";

interface ConfigItem {
	label: string;
	value: string | number | boolean | null | undefined;
	type?: "text" | "boolean";
	badge?: boolean;
}


interface MasterdataForm {
	region: string;
	source: string;
	custom_url: string;
	custom_fallback_url: string;
	local_path: string;
	refresh_interval: number;
}

interface AssetsForm {
	region: string;
	source: string;
	mirror: string;
	custom_base_url: string;
	music_alias_url: string;
	chart_source_url: string;
	sticker_path: string;
}

interface ServerProfileForm {
	enabled: boolean;
	masterdata: MasterdataForm;
	assets: AssetsForm;
	sekai_api: {
		enabled: boolean;
		base_url: string;
		region: string;
		headers_text: string;
		timeout: number;
		rate_limit: number;
	};
	suite_api: {
		enabled: boolean;
		url: string;
		headers_text: string;
		timeout: number;
	};
	ranking_api: {
		timeout: number;
	};
}

interface SettingsForm {
	server: { region: string };
	renderer: { precision: number; chart_precision: number };
	servers: Record<string, ServerProfileForm>;
}

type ServerEntry = {
	option: ConfigOption;
	form: ServerProfileForm;
	state?: PublicServerProfile;
};

const fallbackRegions: ConfigOption[] = [
	{ key: "cn", label: "国服" },
	{ key: "jp", label: "日服" },
	{ key: "tw", label: "台服" },
	{ key: "kr", label: "韩服" },
	{ key: "en", label: "国际服" },
];
const fallbackMasterdataSources: ConfigOption[] = [
	{ key: "moesekai", label: "MoeSekai", regions: ["jp", "cn"] },
	{
		key: "haruki",
		label: "Haruki GitHub",
		regions: ["jp", "cn", "tw", "kr", "en"],
	},
	{ key: "8823", label: "8823 GitHub", regions: ["jp", "cn", "tw"] },
	{ key: "custom", label: "自定义", regions: ["jp", "cn", "tw", "kr", "en"] },
];
const fallbackAssetSources: ConfigOption[] = [
	{ key: "moesekai", label: "MoeSekai", regions: ["jp", "cn"] },
	{
		key: "sekai_best",
		label: "sekai.best",
		regions: ["jp", "cn", "tw", "kr", "en"],
	},
	{ key: "custom", label: "自定义", regions: ["jp", "cn", "tw", "kr", "en"] },
];
const fallbackAssetMirrors: ConfigOption[] = [
	{ key: "main", label: "主镜像" },
	{ key: "backup", label: "备用镜像" },
	{ key: "overseas", label: "海外镜像" },
	{ key: "overseas_backup", label: "海外备用" },
];

const config = ref<PublicConfig | null>(null);
const form = ref<SettingsForm>(createEmptyForm());
const activeTab = ref("jp");
const savedSnapshot = ref("");
const loading = ref(false);
const saving = ref(false);
const reloading = ref("");
const error = ref("");
const success = ref("");
const sekaiTesting = ref<Record<string, boolean>>({});
const sekaiTestResults = ref<Record<string, SekaiSystemTestResponse>>({});
const thumbnailCache = ref<RendererCardThumbnailCacheStatus | null>(null);
const thumbnailCacheLoading = ref(false);
const thumbnailCachePreloading = ref(false);
let thumbnailCachePollTimer: ReturnType<typeof window.setTimeout> | null = null;

const regionOptions = computed(
	() => config.value?.presets.regions ?? fallbackRegions,
);
const masterdataSourceOptions = computed(
	() => config.value?.presets.masterdata_sources ?? fallbackMasterdataSources,
);
const assetSourceOptions = computed(
	() => config.value?.presets.asset_sources ?? fallbackAssetSources,
);
const assetMirrorOptions = computed(
	() => config.value?.presets.asset_mirrors ?? fallbackAssetMirrors,
);

const dirty = computed(() => {
	if (!config.value) return false;
	return JSON.stringify(buildPayload()) !== savedSnapshot.value;
});
const serverEntries = computed<ServerEntry[]>(() =>
	regionOptions.value.map((region) => ({
		option: region,
		form: ensureServerForm(region.key),
		state: config.value?.servers?.[region.key],
	})),
);
const serverProfilesSupported = computed(() =>
	serverEntries.value.every((entry) => serverProfileSupported(entry.form)),
);
const canSave = computed(
	() => serverProfilesSupported.value && headersJSONValid.value && form.value.renderer.precision > 0 && form.value.renderer.chart_precision > 0,
);
const headersJSONValid = computed(() =>
	Object.values(form.value.servers).every(
		(profile) => parseHeadersText(profile.sekai_api.headers_text).ok && parseHeadersText(profile.suite_api.headers_text).ok,
	),
);
const rendererOutputScaleText = computed(
	() => `${formatNumber(form.value.renderer.precision)}x`,
);
const chartRendererOutputScaleText = computed(
	() => `${formatNumber(form.value.renderer.chart_precision)}x`,
);
const thumbnailCacheRegion = computed(
	() => form.value.server.region || config.value?.server.region || "jp",
);
const thumbnailCacheProgress = computed(() => {
	const status = thumbnailCache.value;
	if (!status) return 0;
	const sourceTotal = status.total ?? 0;
	const sourceProgress = status.running
		? Math.max(status.progress ?? 0, sourceTotal ? (status.cached ?? 0) / sourceTotal : 1)
		: sourceTotal ? (status.cached ?? 0) / sourceTotal : 1;
	const compositeTotal = status.composite_total ?? 0;
	const compositeProgress = status.running
		? Math.max(status.composite_progress ?? 0, compositeTotal ? (status.composite_cached ?? 0) / compositeTotal : 1)
		: compositeTotal ? (status.composite_cached ?? 0) / compositeTotal : 1;
	const total = sourceTotal + compositeTotal;
	if (total <= 0) return 0;
	return Math.round(((sourceProgress * sourceTotal + compositeProgress * compositeTotal) / total) * 100);
});
const thumbnailCacheBadgeVariant = computed<BadgeVariant>(() => {
	if (!thumbnailCache.value?.enabled) return "warning";
	if (thumbnailCache.value.running) return "secondary";
	if ((thumbnailCache.value.failed ?? 0) > 0 || (thumbnailCache.value.composite_failed ?? 0) > 0) return "warning";
	if ((thumbnailCache.value.missing ?? 0) === 0 && (thumbnailCache.value.composite_missing ?? 0) === 0 && ((thumbnailCache.value.total ?? 0) > 0 || (thumbnailCache.value.composite_total ?? 0) > 0)) return "success";
	return "outline";
});
const thumbnailCacheStatusLabel = computed(() => {
	const status = thumbnailCache.value;
	if (!status) return "未检查";
	if (!status.enabled) return "缓存关闭";
	if (status.running) return "预载中";
	if (status.total === 0 && !status.composite_total) return "无卡牌";
	if (status.missing === 0 && (status.composite_missing ?? 0) === 0) return "已完成";
	return "待预载";
});
const thumbnailPreloadButtonLoading = computed(
	() => thumbnailCachePreloading.value || Boolean(thumbnailCache.value?.running),
);
const thumbnailCacheSummary = computed(() => {
	const status = thumbnailCache.value;
	if (!status) return "点击刷新后查看当前区服缩略图缓存覆盖率。";
	const sourceTotal = status.total_urls ?? status.total;
	const sourceCached = sourceTotal && status.composite_total ? Math.max(0, status.cached - (status.composite_cached ?? 0)) : status.cached;
	const sourceMissing = Math.max(0, sourceTotal - sourceCached);
	const composite = status.composite_total ? ` · 合成 SVG ${status.composite_cached ?? 0}/${status.composite_total}` : "";
	const render = status.composite_render_ms ? ` · SVG 生成 ${status.composite_render_ms}ms` : "";
	return `原图 ${sourceCached}/${sourceTotal} · 缺失 ${sourceMissing} · 失败 ${status.failed}${composite}${render}`;
});

const webItems = computed<ConfigItem[]>(() => [
	{ label: "Host", value: config.value?.web.host ?? "-" },
	{ label: "Port", value: config.value?.web.port ?? "-" },
]);

const botItems = computed<ConfigItem[]>(() => [
	{ label: "驱动类型", value: config.value?.bot.driver_type ?? "-" },
	{ label: "监听地址", value: config.value?.bot.listen ?? "-" },
	{ label: "命令前缀", value: config.value?.bot.command_prefix ?? "-" },
	{
		label: "自定义关键词",
		value: aliasCount(config.value?.bot.command_aliases),
		badge: true,
	},
	{ label: "昵称", value: config.value?.bot.nickname?.join(" / ") || "-" },
	{
		label: "URL 已配置",
		value: Boolean(config.value?.bot.url_configured),
		badge: true,
	},
	{
		label: "Token 已设置",
		value: Boolean(config.value?.bot.token_set),
		badge: true,
	},
]);

const rendererItems = computed<ConfigItem[]>(() => [
	{ label: "Base URL", value: config.value?.renderer.base_url ?? "-" },
	{ label: "Host", value: config.value?.renderer.host ?? "-" },
	{ label: "Port", value: config.value?.renderer.port ?? "-" },
	{
		label: "当前精度",
		value: `${formatNumber(config.value?.renderer.precision ?? 1.5)}x`,
	},
	{
		label: "谱面精度",
		value: `${formatNumber(config.value?.renderer.chart_precision ?? 4)}x`,
	},
	{
		label: "缓存启用",
		value: Boolean(config.value?.renderer.cache.enabled),
		badge: true,
	},
	{ label: "缓存路径", value: config.value?.renderer.cache.path ?? "-" },
	{
		label: "缓存上限",
		value: formatCacheMaxSize(config.value?.renderer.cache.max_size_mb),
	},
	{
		label: "缓存有效期",
		value: formatCacheTTL(config.value?.renderer.cache.ttl_hours),
	},
]);

const masterdataItems = computed<ConfigItem[]>(() => [
	{
		label: "区服",
		value: `${config.value?.masterdata.region_label ?? "-"} (${config.value?.masterdata.region ?? "-"})`,
	},
	{
		label: "来源",
		value:
			config.value?.masterdata.source_label ||
			config.value?.masterdata.source ||
			"-",
	},
	{ label: "主 URL", value: config.value?.masterdata.url || "-" },
	{ label: "备用 URL", value: config.value?.masterdata.fallback_url || "-" },
	{
		label: "源可用",
		value: Boolean(config.value?.masterdata.supported),
		badge: true,
	},
	{ label: "本地路径", value: config.value?.masterdata.local_path ?? "-" },
	{
		label: "刷新间隔",
		value: `${config.value?.masterdata.refresh_interval ?? "-"} 秒`,
	},
]);

const sekaiApiItems = computed<ConfigItem[]>(() => [
	{
		label: "SEKAI 启用",
		value: Boolean(config.value?.sekai_api.enabled),
		badge: true,
	},
	{
		label: "端点已配置",
		value: Boolean(config.value?.sekai_api.base_url_configured),
		badge: true,
	},
	{
		label: "请求头已配置",
		value: Boolean(config.value?.sekai_api.headers_configured),
		badge: true,
	},
	{
		label: "Suite URL 已配置",
		value: Boolean(config.value?.suite_api?.url_configured),
		badge: true,
	},
	{
		label: "Suite 请求头已配置",
		value: Boolean(config.value?.suite_api?.headers_configured),
		badge: true,
	},
	{ label: "Ranking 来源", value: "MoeSekai 公开榜线" },
	{ label: "Ranking 自动区服", value: config.value?.ranking_api?.region ?? "-" },
]);

function aliasCount(aliases?: Record<string, string[]>) {
	if (!aliases) return false;
	return (
		Object.values(aliases).reduce(
			(total, list) => total + (Array.isArray(list) ? list.length : 0),
			0,
		) > 0
	);
}

const assetItems = computed<ConfigItem[]>(() => [
	{
		label: "区服",
		value: `${config.value?.assets.region_label ?? "-"} (${config.value?.assets.region ?? "-"})`,
	},
	{
		label: "来源",
		value:
			config.value?.assets.source_label || config.value?.assets.source || "-",
	},
	{
		label: "镜像",
		value:
			config.value?.assets.mirror_label || config.value?.assets.mirror || "-",
	},
	{ label: "Base URL", value: config.value?.assets.base_url || "-" },
	{
		label: "Renderer Source",
		value: config.value?.assets.renderer_source || "-",
	},
	{
		label: "曲名别名配置",
		value: Boolean(config.value?.assets.music_alias_configured),
		badge: true,
	},
	{ label: "贴纸路径", value: config.value?.assets.sticker_path ?? "-" },
	{ label: "版本", value: config.value?.version ?? "-" },
]);

const ConfigSection = defineComponent({
	props: {
		title: { type: String, required: true },
		description: { type: String, required: true },
		icon: { type: String as () => IconName, required: true },
		items: { type: Array as () => ConfigItem[], required: true },
	},
	setup(props) {
		return () =>
			h(UiCard, { className: "settings-card" }, () => [
				h("div", { class: "settings-card__heading" }, [
					h("div", { class: "settings-card__icon" }, [
						h(SvgIcon, { name: props.icon, size: 22 }),
					]),
					h("div", null, [h("h2", props.title), h("p", props.description)]),
				]),
				h(
					"dl",
					{ class: "settings-list" },
					props.items.map((item) =>
						h("div", { key: item.label }, [
							h("dt", item.label),
							h(
								"dd",
								item.badge
									? h(
											UiBadge,
											{ variant: item.value ? "success" : "warning" },
											() => (item.value ? "是" : "否"),
										)
									: String(item.value),
							),
						]),
					),
				),
			]);
	},
});

onMounted(loadConfig);
onBeforeUnmount(clearThumbnailCachePoll);

async function loadConfig() {
	loading.value = true;
	error.value = "";
	success.value = "";
	try {
		const data = await getPublicConfig();
		config.value = data;
		applyConfigToForm(data);
		await refreshThumbnailCacheStatus(true);
	} catch (err) {
		error.value = getErrorMessage(err, "加载配置失败。");
	} finally {
		loading.value = false;
	}
}

async function saveSettings() {
	const invalidHeaders = firstInvalidHeadersEntry();
	if (invalidHeaders) {
		error.value = `${regionLabel(invalidHeaders.region)} 的 ${invalidHeaders.api} Headers JSON 无效：${invalidHeaders.message}`;
		success.value = "";
		return;
	}
	saving.value = true;
	error.value = "";
	success.value = "";
	try {
		const response = await updatePublicConfig(buildPayload());
		config.value = response.config;
		applyConfigToForm(response.config);
		success.value = response.message || "设置已保存。";
	} catch (err) {
		error.value = getErrorMessage(err, "保存设置失败。");
	} finally {
		saving.value = false;
	}
}

async function testSekaiConnectivity(entry: ServerEntry) {
	const region = entry.option.key;
	const headers = parseHeadersText(entry.form.sekai_api.headers_text);
	if (!headers.ok) {
		sekaiTestResults.value = {
			...sekaiTestResults.value,
			[region]: {
				ok: false,
				url: sekaiSystemPreview(entry),
				duration_ms: 0,
				message: `Headers JSON 无效：${headers.message || "JSON 解析失败"}`,
			},
		};
		return;
	}
	sekaiTesting.value = { ...sekaiTesting.value, [region]: true };
	try {
		const result = await testSekaiSystem({
			base_url: entry.form.sekai_api.base_url,
			region,
			headers: headers.headers,
			timeout: Number(entry.form.sekai_api.timeout) || 10,
		});
		sekaiTestResults.value = { ...sekaiTestResults.value, [region]: result };
	} catch (err) {
		sekaiTestResults.value = {
			...sekaiTestResults.value,
			[region]: {
				ok: false,
				url: sekaiSystemPreview(entry),
				duration_ms: 0,
				message: getErrorMessage(err, "SEKAI API /system 连通性测试失败。"),
			},
		};
	} finally {
		sekaiTesting.value = { ...sekaiTesting.value, [region]: false };
	}
}

async function reloadMasterdataNow(region = form.value.server.region) {
	reloading.value = region;
	error.value = "";
	success.value = "";
	try {
		const result = await reloadMasterdata(region);
		success.value = `${result.message}：卡牌 ${result.counts.cards} / 曲目 ${result.counts.musics} / 活动 ${result.counts.events} / 卡池 ${result.counts.gachas} / 演唱会 ${result.counts.virtual_lives ?? 0}`;
		await loadConfig();
	} catch (err) {
		error.value = getErrorMessage(err, "重载 Masterdata 失败。");
	} finally {
		reloading.value = "";
	}
}

async function refreshThumbnailCacheStatus(silent = false) {
	if (!config.value) return;
	clearThumbnailCachePoll();
	thumbnailCacheLoading.value = true;
	try {
		const status = await getRendererCardThumbnailCacheStatus(
			thumbnailCacheRegion.value,
		);
		thumbnailCache.value = status;
		if (status.running) scheduleThumbnailCachePoll();
	} catch (err) {
		if (!silent) error.value = getErrorMessage(err, "刷新缩略图缓存状态失败。");
	} finally {
		thumbnailCacheLoading.value = false;
	}
}

async function preloadCardThumbnails() {
	clearThumbnailCachePoll();
	thumbnailCachePreloading.value = true;
	error.value = "";
	success.value = "";
	try {
		const status = await preloadRendererCardThumbnails(thumbnailCacheRegion.value);
		thumbnailCache.value = status;
		success.value = `${status.region_label} 卡牌缩略图预载已启动：${status.cached}/${status.total} 已缓存。`;
		scheduleThumbnailCachePoll();
	} catch (err) {
		error.value = getErrorMessage(err, "启动卡牌缩略图预载失败。");
	} finally {
		thumbnailCachePreloading.value = false;
	}
}

function scheduleThumbnailCachePoll() {
	clearThumbnailCachePoll();
	thumbnailCachePollTimer = window.setTimeout(async () => {
		await refreshThumbnailCacheStatus(true);
		if (thumbnailCache.value?.running) scheduleThumbnailCachePoll();
	}, 1800);
}

function clearThumbnailCachePoll() {
	if (thumbnailCachePollTimer) {
		window.clearTimeout(thumbnailCachePollTimer);
		thumbnailCachePollTimer = null;
	}
}

function createEmptyForm(): SettingsForm {
	return {
		server: { region: "jp" },
		renderer: { precision: 1.5, chart_precision: 4 },
		servers: {},
	};
}

function applyConfigToForm(data: PublicConfig) {
	const defaultRegion = data.server.region || "jp";
	form.value = {
		server: { region: defaultRegion },
		renderer: { precision: data.renderer.precision || 1.5, chart_precision: data.renderer.chart_precision || 4 },
		servers: {},
	};
	for (const option of regionOptions.value) {
		const server =
			data.servers?.[option.key] ??
			(option.key === defaultRegion
				? {
						enabled: true,
						masterdata: data.masterdata,
						assets: data.assets,
						sekai_api: data.sekai_api,
						suite_api: data.suite_api,
						ranking_api: data.ranking_api,
					}
				: undefined);
		form.value.servers[option.key] = createServerForm(option.key, server);
		normalizeServerProfile(option.key, false);
	}
	savedSnapshot.value = JSON.stringify(buildPayload());
}

function buildPayload(): UpdatePublicConfigPayload {
	const defaultProfile = ensureServerForm(form.value.server.region);
	return {
		server: { region: form.value.server.region },
		renderer: {
			precision: Number(form.value.renderer.precision) || 1.5,
			chart_precision: Number(form.value.renderer.chart_precision) || 4,
		},
		masterdata: buildMasterdataPayload(defaultProfile.masterdata),
		assets: buildAssetsPayload(defaultProfile.assets),
		servers: Object.fromEntries(
			regionOptions.value.map((option) => [
				option.key,
				buildServerPayload(option.key),
			]),
		),
	};
}

function buildMasterdataPayload(masterdata: MasterdataForm) {
	return {
		region: masterdata.region,
		source: masterdata.source,
		custom_url: masterdata.custom_url,
		custom_fallback_url: masterdata.custom_fallback_url,
		local_path: masterdata.local_path,
		refresh_interval: Number(masterdata.refresh_interval) || 0,
	};
}

function buildAssetsPayload(assets: AssetsForm) {
	return {
		region: assets.region,
		source: assets.source,
		mirror: assets.mirror,
		custom_base_url: assets.custom_base_url,
		music_alias_url: assets.music_alias_url,
		chart_source_url: assets.chart_source_url,
		sticker_path: assets.sticker_path,
	};
}

function setDefaultRegion(region: string) {
	form.value.server.region = region;
	const defaultProfile = ensureServerForm(region);
	defaultProfile.enabled = true;
	normalizeServerProfile(region);
	thumbnailCache.value = null;
	void refreshThumbnailCacheStatus(true);
	success.value = "";
}

function normalizeRendererPrecision() {
	const precision = Number(form.value.renderer.precision);
	if (!Number.isFinite(precision) || precision <= 0) {
		form.value.renderer.precision = 1.5;
	}
	const chartPrecision = Number(form.value.renderer.chart_precision);
	if (!Number.isFinite(chartPrecision) || chartPrecision <= 0) {
		form.value.renderer.chart_precision = 4;
	}
}

function createServerForm(
	region: string,
	server?: Partial<PublicServerProfile>,
): ServerProfileForm {
	const masterdata = server?.masterdata;
	const assets = server?.assets;
	return {
		enabled: server?.enabled ?? region === "jp",
		masterdata: {
			region: masterdata?.region || region,
			source:
				masterdata?.source ||
				(region === "jp" || region === "cn" ? "moesekai" : "haruki"),
			custom_url:
				masterdata?.custom_url ||
				(masterdata?.source === "custom" ? masterdata?.url || "" : ""),
			custom_fallback_url:
				masterdata?.custom_fallback_url ||
				(masterdata?.source === "custom" ? masterdata?.fallback_url || "" : ""),
			local_path: masterdata?.local_path || `./data/master/${region}`,
			refresh_interval: masterdata?.refresh_interval ?? 3600,
		},
		assets: {
			region: assets?.region || region,
			source:
				assets?.source ||
				(region === "jp" || region === "cn" ? "moesekai" : "sekai_best"),
			mirror: assets?.mirror || "main",
			custom_base_url:
				assets?.custom_base_url ||
				(assets?.source === "custom" ? assets?.base_url || "" : ""),
			music_alias_url: assets?.music_alias_url || "",
			chart_source_url: assets?.chart_source_url || "https://charts-new.unipjsk.com/moe/svg/{id}/{difficulty}.svg",
			sticker_path: assets?.sticker_path || "./assets/stickers",
		},
		sekai_api: {
			enabled: server?.sekai_api?.enabled ?? false,
			base_url: server?.sekai_api?.base_url || "https://seka-api.exmeaning.com",
			region: server?.sekai_api?.region || region,
			headers_text: formatHeadersText(server?.sekai_api?.headers),
			timeout: server?.sekai_api?.timeout ?? 10,
			rate_limit: server?.sekai_api?.rate_limit ?? 30,
		},
		suite_api: {
			enabled: true,
			url: server?.suite_api?.url || "https://suite-api.haruki.seiunx.com/public/{region}/suite/{uid}",
			headers_text: formatHeadersText(server?.suite_api?.headers),
			timeout: server?.suite_api?.timeout ?? 10,
		},
		ranking_api: {
			timeout: server?.ranking_api?.timeout ?? 10,
		},
	};
}

function ensureServerForm(region: string) {
	if (!form.value.servers[region]) {
		form.value.servers[region] = createServerForm(region);
	}
	return form.value.servers[region];
}

function buildServerPayload(region: string) {
	const profile = ensureServerForm(region);
	return {
		enabled: isRegionLocked(region) ? true : profile.enabled,
		masterdata: buildMasterdataPayload(profile.masterdata),
		assets: buildAssetsPayload(profile.assets),
		sekai_api: {
			enabled: profile.sekai_api.enabled,
			base_url: profile.sekai_api.base_url,
			region,
			...(profile.sekai_api.headers_text.trim()
				? { headers: parseHeadersText(profile.sekai_api.headers_text).headers }
				: {}),
			timeout: Number(profile.sekai_api.timeout) || 10,
			rate_limit: Number(profile.sekai_api.rate_limit) || 30,
		},
		suite_api: {
			enabled: true,
			url: profile.suite_api.url,
			...(profile.suite_api.headers_text.trim()
				? { headers: parseHeadersText(profile.suite_api.headers_text).headers }
				: {}),
			timeout: Number(profile.suite_api.timeout) || 10,
		},
		ranking_api: {
			timeout: Number(profile.ranking_api.timeout) || 10,
		},
	};
}

function normalizeServerProfile(region: string, clearSuccess = true) {
	const profile = ensureServerForm(region);
	profile.masterdata.region = region;
	profile.assets.region = region;
	if (
		!optionAvailable(
			findOption(masterdataSourceOptions.value, profile.masterdata.source),
			region,
		)
	) {
		profile.masterdata.source =
			firstAvailableOption(masterdataSourceOptions.value, region)?.key ??
			"custom";
	}
	if (
		!optionAvailable(
			findOption(assetSourceOptions.value, profile.assets.source),
			region,
		)
	) {
		profile.assets.source =
			firstAvailableOption(assetSourceOptions.value, region)?.key ?? "custom";
	}
	if (profile.assets.source !== "moesekai") {
		profile.assets.mirror = "";
	} else if (!profile.assets.mirror) {
		profile.assets.mirror = "main";
	}
	if (clearSuccess) success.value = "";
}

function serverProfileSupported(profile: ServerProfileForm) {
	return masterdataProfileSupported(profile) && assetProfileSupported(profile);
}

function masterdataProfileSupported(profile: ServerProfileForm) {
	return optionAvailable(
		findOption(masterdataSourceOptions.value, profile.masterdata.source),
		profile.masterdata.region,
	);
}

function assetProfileSupported(profile: ServerProfileForm) {
	return optionAvailable(
		findOption(assetSourceOptions.value, profile.assets.source),
		profile.assets.region,
	);
}

function optionAvailable(option: ConfigOption | undefined, region: string) {
	if (!option) return false;
	return !option.regions?.length || option.regions.includes(region);
}

function findOption(options: ConfigOption[], key: string) {
	return options.find((option) => option.key === key);
}

function firstAvailableOption(options: ConfigOption[], region: string) {
	return options.find((option) => optionAvailable(option, region));
}

function regionLabel(region: string) {
	return findOption(regionOptions.value, region)?.label ?? region;
}

function sourceLabel(options: ConfigOption[], key: string) {
	return findOption(options, key)?.label ?? key;
}

function sourceSupportText(options: ConfigOption[], key: string) {
	const regions = findOption(options, key)?.regions;
	if (!regions?.length) return "全部区服";
	return regions.map((region) => regionLabel(region)).join(" / ");
}

function isRegionLocked(region: string) {
	return region === "jp" || region === form.value.server.region;
}

function formatHeadersText(headers?: Record<string, string>) {
	return headers && Object.keys(headers).length > 0 ? JSON.stringify(headers, null, 2) : "";
}

function parseHeadersText(raw: string): { ok: boolean; headers: Record<string, string>; message?: string } {
	const text = raw.trim();
	if (!text) return { ok: true, headers: {} };
	try {
		const parsed = JSON.parse(text) as unknown;
		if (!parsed || Array.isArray(parsed) || typeof parsed !== "object") {
			return { ok: false, headers: {}, message: "必须是对象" };
		}
		return normalizeHeadersRecord(parsed as Record<string, unknown>);
	} catch {
		return parseLooseHeadersText(text);
	}
}

function normalizeHeadersRecord(input: Record<string, unknown>): { ok: boolean; headers: Record<string, string>; message?: string } {
	const headers: Record<string, string> = {};
	for (const [key, value] of Object.entries(input)) {
		const headerKey = key.trim();
		if (!headerKey) continue;
		if (typeof value !== "string" && typeof value !== "number" && typeof value !== "boolean") {
			return { ok: false, headers: {}, message: `${headerKey} 的值必须是字符串、数字或布尔值` };
		}
		const headerValue = String(value).trim();
		if (headerValue) headers[headerKey] = headerValue;
	}
	return { ok: true, headers };
}

function parseLooseHeadersText(text: string): { ok: boolean; headers: Record<string, string>; message?: string } {
	const body = text.replace(/^\s*\{/, "").replace(/\}\s*$/, "").trim();
	if (!body) return { ok: true, headers: {} };
	const entries = body.split(/,?\r?\n/).map((line) => line.trim()).filter(Boolean);
	if (!entries.length) return { ok: true, headers: {} };
	const headers: Record<string, string> = {};
	for (const entry of entries) {
		const line = entry.replace(/,$/, "").trim();
		const colon = line.indexOf(":");
		if (colon <= 0) {
			return { ok: false, headers: {}, message: "每一项都需要写成 header: value" };
		}
		const headerKey = stripHeaderQuotes(line.slice(0, colon).trim());
		const headerValue = stripHeaderQuotes(line.slice(colon + 1).trim());
		if (headerKey && headerValue) headers[headerKey] = headerValue;
	}
	return { ok: true, headers };
}

function stripHeaderQuotes(value: string) {
	const cleaned = value.replace(/,$/, "").trim();
	if ((cleaned.startsWith('"') && cleaned.endsWith('"')) || (cleaned.startsWith("'") && cleaned.endsWith("'"))) {
		return cleaned.slice(1, -1).trim();
	}
	return cleaned;
}

function firstInvalidHeadersEntry() {
	for (const [region, profile] of Object.entries(form.value.servers)) {
		const sekai = parseHeadersText(profile.sekai_api.headers_text);
		if (!sekai.ok) return { region, api: "SEKAI API", message: sekai.message || "JSON 解析失败" };
		const suite = parseHeadersText(profile.suite_api.headers_text);
		if (!suite.ok) return { region, api: "Suite API", message: suite.message || "JSON 解析失败" };
	}
	return null;
}

function countsText(counts?: MasterdataCounts) {
	return `卡 ${counts?.cards ?? 0} / 曲 ${counts?.musics ?? 0} / 活 ${counts?.events ?? 0} / 池 ${counts?.gachas ?? 0} / 演 ${counts?.virtual_lives ?? 0}`;
}

function formatTime(value?: string | null) {
	return value ? new Date(value).toLocaleString() : "-";
}

function formatNumber(value: number) {
	if (!Number.isFinite(value)) return "-";
	return Number.isInteger(value)
		? String(value)
		: value.toFixed(2).replace(/0+$/, "").replace(/\.$/, "");
}

function formatCacheMaxSize(value?: number) {
	if (typeof value !== "number") return "-";
	return value <= 0 ? "不限制" : `${value} MB`;
}

function formatCacheTTL(value?: number) {
	if (typeof value !== "number") return "-";
	return value <= 0 ? "永久有效" : `${value} 小时`;
}

function sekaiSystemPreview(entry: ServerEntry) {
	const baseURL = (entry.form.sekai_api.base_url || "https://seka-api.exmeaning.com").trim().replace(/\/+$/, "");
	const region = entry.option.key;
	const replaced = baseURL.replaceAll("{region}", region);
	if (baseURL.includes("{region}")) return `${replaced}/system`;
	return `${replaced}/api/${region}/system`;
}

function masterdataPreview(entry: ServerEntry, kind: "primary" | "fallback") {
	const masterdata = entry.form.masterdata;
	if (masterdata.source === "custom") {
		return kind === "primary"
			? masterdata.custom_url || "-"
			: masterdata.custom_fallback_url || "-";
	}
	const current = entry.state?.masterdata;
	if (
		!current ||
		current.source !== masterdata.source ||
		current.region !== masterdata.region
	) {
		return "保存后由后端解析";
	}
	return kind === "primary" ? current.url || "-" : current.fallback_url || "-";
}

function assetsPreview(entry: ServerEntry, kind: "base" | "renderer") {
	const assets = entry.form.assets;
	if (assets.source === "custom") {
		return kind === "base"
			? assets.custom_base_url || "-"
			: assets.custom_base_url || "-";
	}
	const current = entry.state?.assets;
	if (
		!current ||
		current.source !== assets.source ||
		current.region !== assets.region ||
		current.mirror !== assets.mirror
	) {
		return "保存后由后端解析";
	}
	return kind === "base"
		? current.base_url || "-"
		: current.renderer_source || "-";
}

function masterdataHint(entry: ServerEntry) {
	if (!entry.form.enabled) return "该区服未启用；启用并保存后即可重载。";
	return (
		entry.state?.masterdata.error ||
		entry.state?.masterdata.load_error ||
		"保存来源后可立即重载；失败时会继续保留本地缓存兜底。"
	);
}

function getErrorMessage(err: unknown, fallback: string) {
	const maybeAxios = err as {
		response?: { data?: { message?: string } };
		message?: string;
	};
	return maybeAxios.response?.data?.message || maybeAxios.message || fallback;
}
</script>

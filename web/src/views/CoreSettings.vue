<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="Core & Bot"
      title="核心设置"
      subtitle="框架核心运行时配置（含 Bot 连接）与已加载插件管理；业务设置由各插件提供其专属设置页。">
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loadingBot" @click="loadBot">刷新 Bot 状态</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="botError" variant="destructive" title="Bot 状态加载失败">{{ botError }}</UiAlert>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>Bot 连接状态</h2>
          <p>ZeroBot / OneBot v11 反向 WebSocket 配置概览。</p>
        </div>
        <UiBadge :variant="status?.bot.ok ? 'success' : 'destructive'">{{ status?.bot.status ?? 'unknown' }}</UiBadge>
      </div>
      <dl class="info-list">
        <div><dt>状态说明</dt><dd>{{ status?.bot.message ?? '-' }}</dd></div>
        <div><dt>驱动类型</dt><dd>{{ status?.bot.driver_type ?? '-' }}</dd></div>
        <div><dt>监听地址</dt><dd>{{ status?.bot.listen ?? '-' }}</dd></div>
        <div><dt>URL 已配置</dt><dd>{{ status?.bot.url_configured ? '是' : '否' }}</dd></div>
        <div><dt>命令前缀</dt><dd>{{ status?.bot.command_prefix ?? '-' }}</dd></div>
        <div><dt>昵称</dt><dd>{{ status?.bot.nicknames?.join(' / ') || '-' }}</dd></div>
      </dl>
    </UiCard>

    <!-- 超级管理员 (ZeroBot SuperUsers) -->
    <UiCard>
      <div class="card-heading">
        <div>
          <h2>超级管理员</h2>
          <p>在这里填入的 QQ 将被设置为 ZeroBot 的 SuperUsers，拥有全局最高权限。多个插件的权限命令仅限超级管理员使用，例如 AutoChat 的 <code>/开启聊天</code>、<code>/关闭聊天</code>、<code>/开启autochat</code>、<code>/关闭autochat</code>。</p>
        </div>
      </div>
      <UiAlert v-if="superSaveError" variant="destructive" title="保存失败">{{ superSaveError }}</UiAlert>
      <UiAlert v-if="superSaveSuccess" variant="info" title="已保存">
        超级管理员名单已写入 <code>data/config.yml</code>。<strong>需要重启 Moebot 进程</strong> 后 ZeroBot 才会重新加载。
      </UiAlert>
      <div class="super-users">
        <label class="super-label">QQ 列表（每行一个）</label>
        <textarea
          v-model="superUsersText"
          rows="4"
          class="ui-textarea"
          placeholder="123456789&#10;987654321"
        />
        <div class="super-foot">
          <span class="hint">当前已配置 {{ parsedSuperUsers.length }} 人</span>
          <UiButton variant="default" size="sm" :loading="superSaving" :disabled="!superDirty" @click="saveSuperUsers">保存</UiButton>
        </div>
      </div>
    </UiCard>

    <!-- 控制台账号 / 修改密码 -->
    <UiCard>
      <div class="card-heading">
        <div>
          <h2>控制台账号</h2>
          <p>账号与昵称在初始化时创建后<strong>无法更改</strong>；昵称会显示在所有 Satori 渲染卡片底部 <code>Moebot NEXT (deployed by 昵称)</code>。</p>
        </div>
      </div>
      <dl class="info-list">
        <div><dt>账号</dt><dd>{{ auth.username || '-' }}</dd></div>
        <div><dt>昵称</dt><dd>{{ auth.nickname || '-' }}</dd></div>
      </dl>
      <div class="pwd-form">
        <h3 class="pwd-title">修改密码</h3>
        <UiAlert v-if="pwdError" variant="destructive" title="修改失败">{{ pwdError }}</UiAlert>
        <UiAlert v-if="pwdSuccess" variant="info" title="已更新">密码修改成功，下次登录请使用新密码。</UiAlert>
        <label class="pwd-row"><span>旧密码</span><input v-model="pwdForm.old_password" type="password" class="ui-textarea" /></label>
        <label class="pwd-row"><span>新密码（至少 8 位）</span><input v-model="pwdForm.new_password" type="password" class="ui-textarea" /></label>
        <label class="pwd-row"><span>确认新密码</span><input v-model="pwdForm.new_password_confirm" type="password" class="ui-textarea" /></label>
        <div class="super-foot">
          <span class="hint">两次密码需要保持一致</span>
          <UiButton variant="default" size="sm" :loading="pwdSaving" :disabled="!pwdReady" @click="onSavePassword">提交</UiButton>
        </div>
      </div>
    </UiCard>

    <!-- Renderer：所有插件共用的 Satori/SVG→PNG 渲染服务 -->
    <UiCard>
      <div class="card-heading">
        <div>
          <h2>Renderer</h2>
          <p>Satori 渲染服务、SVG→PNG 精度、字体与渲染结果缓存。所有依赖渲染的插件共用本配置。</p>
        </div>
        <UiBadge :variant="rendererConfig ? 'success' : 'outline'">
          {{ rendererConfig ? `Bun :${rendererConfig.port}` : '未加载' }}
        </UiBadge>
      </div>
      <UiAlert v-if="rendererError" variant="destructive" title="渲染设置错误">{{ rendererError }}</UiAlert>
      <UiAlert v-if="rendererSuccess" variant="info" title="已保存">{{ rendererSuccess }}</UiAlert>

      <div class="renderer-form">
        <label class="renderer-field">
          <span>渲染精度</span>
          <input v-model.number="rendererForm.precision" class="ui-textarea" type="number" min="0.1" step="0.1" />
        </label>
        <label class="renderer-field">
          <span>谱面渲染精度</span>
          <input v-model.number="rendererForm.chart_precision" class="ui-textarea" type="number" min="0.1" step="0.1" />
        </label>
        <div class="renderer-field renderer-field--readonly">
          <span>说明</span>
          <strong>普通图片约 {{ rendererForm.precision || 1.5 }}x，谱面约 {{ rendererForm.chart_precision || 4 }}x；越高越清晰但图片体积和耗时越大。</strong>
        </div>
      </div>

      <!-- 渲染字体 -->
      <div class="renderer-subpanel">
        <div class="renderer-subpanel__header">
          <div>
            <strong>渲染字体</strong>
            <span>放置字体文件到 <code>renderer/assets/fonts/</code> 目录即可自动加载，支持 .otf / .ttf / .woff。保存后即时生效。</span>
          </div>
          <UiBadge :variant="fontsLoaded ? 'success' : 'warning'">
            {{ fontsLoaded ? `${fontsData?.total ?? 0} 字体` : '未加载' }}
          </UiBadge>
        </div>
        <div v-if="fontsData" class="renderer-form">
          <label class="renderer-field">
            <span>正文字体</span>
            <select v-model="rendererForm.fonts.body_family" class="ui-textarea">
              <option value="">默认（{{ fontsData.config.body }}）</option>
              <option v-for="family in fontsData.families" :key="`body-${family}`" :value="family">{{ family }}</option>
            </select>
          </label>
          <label class="renderer-field">
            <span>PT 得分字体（黑体）</span>
            <select v-model="rendererForm.fonts.score_family" class="ui-textarea">
              <option value="">默认（{{ fontsData.config.score }}）</option>
              <option v-for="family in fontsData.families" :key="`score-${family}`" :value="family">{{ family }}</option>
            </select>
          </label>
          <div class="renderer-field renderer-field--readonly renderer-field--full">
            <span>当前生效</span>
            <strong style="font-size: 12px; line-height: 1.6;">
              正文：{{ fontsData.defaults.body }}<br />
              PT：{{ fontsData.defaults.score }}
            </strong>
          </div>
        </div>
        <div v-else-if="!fontsLoading" class="muted">无法获取字体信息，请检查 Renderer 服务状态。</div>
        <div class="super-foot">
          <span class="hint">字体目录：{{ rendererConfig?.cache.path || '-' }}</span>
          <UiButton variant="outline" size="sm" :loading="fontsLoading" @click="loadFontsInfo">刷新字体</UiButton>
        </div>
      </div>

      <!-- 渲染结果缓存 -->
      <div class="renderer-subpanel">
        <div class="renderer-subpanel__header">
          <div>
            <strong>渲染结果缓存</strong>
            <span>分层 TTL：详情 365d · 列表 10min · 用户 5min · 动态 10s。data 哈希保证正确性。</span>
          </div>
          <UiBadge :variant="renderCache ? 'success' : 'outline'">
            {{ renderCache ? `命中率 ${(renderCache.hitRate * 100).toFixed(1)}%` : '未加载' }}
          </UiBadge>
        </div>
        <div v-if="renderCache" class="renderer-meter" aria-hidden="true">
          <span :style="{ width: `${Math.min(100, renderCache.byteUsageRatio * 100).toFixed(1)}%` }" />
        </div>
        <div v-if="renderCache" class="renderer-meta">
          <span>已用 {{ formatBytes(renderCache.bytes) }} / {{ formatBytes(renderCache.maxBytes) }}（{{ (renderCache.byteUsageRatio * 100).toFixed(1) }}%）</span>
          <span>条数 {{ renderCache.size }} / {{ renderCache.maxEntries }}</span>
          <span>命中 {{ renderCache.hits }} · 未命中 {{ renderCache.misses }} · 淘汰 {{ renderCache.evictions }}</span>
        </div>
        <div v-if="renderCache" class="renderer-form">
          <label class="renderer-field">
            <span>字节预算 (MB)</span>
            <input
              v-model.number="renderCacheForm.maxBytesMB"
              class="ui-textarea"
              type="number"
              :min="Math.ceil(renderCache.limits.minMaxBytes / (1024 * 1024))"
              :max="Math.floor(renderCache.limits.hardMaxBytes / (1024 * 1024))"
              step="1"
            />
          </label>
          <label class="renderer-field">
            <span>最大条数</span>
            <input
              v-model.number="renderCacheForm.maxEntries"
              class="ui-textarea"
              type="number"
              :min="renderCache.limits.minMaxEntries"
              :max="renderCache.limits.hardMaxEntries"
              step="16"
            />
          </label>
          <div class="renderer-field renderer-field--readonly">
            <span>说明</span>
            <strong>
              范围 {{ formatBytes(renderCache.limits.minMaxBytes) }}–{{ formatBytes(renderCache.limits.hardMaxBytes) }} ·
              条数 {{ renderCache.limits.minMaxEntries }}–{{ renderCache.limits.hardMaxEntries }}。
              超限会按 LRU 立即淘汰旧项。
            </strong>
          </div>
        </div>
        <div class="super-foot">
          <span class="hint">{{ renderCache ? `默认 TTL ${(renderCache.defaultTtlMs / 1000).toFixed(0)}s` : '' }}</span>
          <div class="renderer-actions">
            <UiButton variant="outline" size="sm" :loading="renderCacheLoading" @click="() => refreshRenderCacheStats()">刷新</UiButton>
            <UiButton size="sm" :loading="renderCacheSaving" :disabled="!renderCache" @click="saveRenderCacheConfig">保存上限</UiButton>
            <UiButton variant="destructive" size="sm" :loading="renderCacheClearing" :disabled="!renderCache" @click="clearRenderCacheAction">清空缓存</UiButton>
          </div>
        </div>
      </div>

      <!-- 渲染并发预算（防 OOM / 调度多核） -->
      <div class="renderer-subpanel">
        <div class="renderer-subpanel__header">
          <div>
            <strong>渲染并发预算</strong>
            <span>限制同时进行的渲染数和等待队列。低内存机器请保守，超过队列上限直接 503。</span>
          </div>
          <UiBadge :variant="budgetStats ? 'success' : 'outline'">
            {{ budgetStats ? `Worker ${budgetStats.inFlight}/${budgetStats.maxConcurrency} · 排队 ${budgetStats.queued}/${budgetStats.queueLimit}` : '未加载' }}
          </UiBadge>
        </div>
        <div v-if="budgetStats" class="renderer-meta">
          <span>已完成 {{ budgetStats.completed }} · 拒绝 {{ budgetStats.rejected }} · 平均等待 {{ budgetStats.avgWaitMs }}ms</span>
          <span>峰值并发 {{ budgetStats.peakInFlight }} · 峰值队列 {{ budgetStats.peakQueued }} · Worker {{ budgetStats.workerCount ?? budgetStats.maxConcurrency }} 个</span>
        </div>
        <div class="renderer-form">
          <label class="renderer-field">
            <span>渲染 Worker 数</span>
            <input
              v-model.number="budgetForm.maxConcurrency"
              class="ui-textarea"
              type="number"
              min="1"
              :max="budgetStats?.limits.hardMaxConcurrency ?? 64"
              step="1"
            />
          </label>
          <label class="renderer-field">
            <span>排队上限</span>
            <input
              v-model.number="budgetForm.queueLimit"
              class="ui-textarea"
              type="number"
              min="0"
              :max="budgetStats?.limits.hardMaxQueue ?? 1024"
              step="1"
            />
          </label>
          <label class="renderer-field">
            <span>单 Worker 准备并发</span>
            <input
              v-model.number="budgetForm.prepareConcurrency"
              class="ui-textarea"
              type="number"
              min="1"
              :max="budgetStats?.limits.hardMaxPrepareConcurrency ?? 32"
              step="1"
            />
          </label>
          <div class="renderer-field renderer-field--readonly renderer-field--full">
            <span>建议</span>
            <strong>
              16G / 12v 服务器建议 Worker 4、队列 16、单 Worker 准备并发 4；内存稳定后可尝试 Worker 6。
              范围：Worker 1–{{ budgetStats?.limits.hardMaxConcurrency ?? 32 }} · 队列 0–{{ budgetStats?.limits.hardMaxQueue ?? 1024 }} · 准备并发 1–{{ budgetStats?.limits.hardMaxPrepareConcurrency ?? 32 }}
            </strong>
          </div>
        </div>
        <div class="super-foot">
          <span class="hint">保存后即时生效，并写入配置文件。</span>
          <div class="renderer-actions">
            <UiButton variant="outline" size="sm" :loading="budgetLoading" @click="() => refreshBudgetStats()">刷新</UiButton>
            <UiButton size="sm" :loading="budgetSaving" @click="saveBudget">保存并发预算</UiButton>
          </div>
        </div>
      </div>

      <div class="super-foot" style="margin-top: 16px;">
        <span class="hint">保存后立即生效；字体目录变化需重启 Bun 进程。</span>
        <UiButton variant="default" size="sm" :loading="rendererSaving" :disabled="!rendererDirty" @click="saveRendererSettings">保存渲染设置</UiButton>
      </div>
    </UiCard>

    <UiAlert variant="info" title="业务配置已下沉到插件">
      原“区服 / Masterdata / Assets / Sekai API”等设置属于
      <strong>MoeSekai</strong> 插件，已迁移至
      <RouterLink to="/plugins/moesekai">/plugins/moesekai</RouterLink>。
      本页仅保留对所有插件通用的核心运行时配置。
    </UiAlert>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>插件管理</h2>
          <p>启用 / 禁用插件、跳转到各插件设置、浏览插件市场。</p>
        </div>
      </div>
      <div class="quick-links">
        <RouterLink to="/plugins" class="ui-button ui-button--default ui-button--sm">插件列表</RouterLink>
        <RouterLink to="/plugins/market" class="ui-button ui-button--outline ui-button--sm">插件市场</RouterLink>
      </div>
    </UiCard>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>已加载插件</h2>
          <p>下表反映当前进程实际加载的插件，供快速判断状态。</p>
        </div>
        <UiButton variant="outline" size="sm" :loading="loading" @click="load">刷新</UiButton>
      </div>

      <UiAlert v-if="error" variant="destructive" title="加载失败">{{ error }}</UiAlert>

      <table v-if="plugins.length" class="plugins-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>分类</th>
            <th>启用偏好</th>
            <th>运行状态</th>
            <th>设置</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in plugins" :key="p.name">
            <td>
              <div class="plugin-name">{{ p.title }}</div>
              <div class="plugin-id">{{ p.name }} · v{{ p.version }}</div>
            </td>
            <td>{{ categoryLabel(p.category) }}</td>
            <td>{{ p.enabled ? '已启用' : '已禁用' }}</td>
            <td>
              <span :class="['dot', p.loaded ? 'dot--ok' : 'dot--off']" /> {{ p.loaded ? '运行中' : '已停止' }}
            </td>
            <td>
              <RouterLink
                v-if="p.settings_route"
                :to="p.settings_route"
                class="ui-button ui-button--outline ui-button--sm"
              >打开</RouterLink>
              <span v-else class="muted">—</span>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else-if="!loading" class="muted">暂无已注册插件。</div>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import {
  clearRenderCache as apiClearRenderCache,
  getPublicConfig,
  getRenderCacheStats,
  getRendererBudgetStats,
  getRendererFonts,
  getStatus,
  listPlugins,
  updatePublicConfig,
  updateRenderCacheConfig,
  updateRendererBudget,
  type PluginListItem,
} from '../api/client'
import { useAuthStore } from '../stores/auth'
import type { PublicConfig, RenderCacheStats, RendererBudgetStats, RendererFontsResponse, RuntimeStatus } from '../api/types'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'

const plugins = ref<PluginListItem[]>([])
const loading = ref(false)
const error = ref('')

const status = ref<RuntimeStatus | null>(null)
const loadingBot = ref(false)
const botError = ref('')

// 超级管理员状态
const superUsersText = ref('')
const superUsersOriginal = ref('')
const superSaving = ref(false)
const superSaveError = ref('')
const superSaveSuccess = ref(false)

// 解析文本区为 int64[]：允许每行 / 逗号 / 空格分隔；忽略非法项。
const parsedSuperUsers = computed(() => {
  const seen = new Set<number>()
  const out: number[] = []
  for (const tok of superUsersText.value.split(/[\s,]+/)) {
    const t = tok.trim()
    if (!t) continue
    const qq = Number(t)
    if (!Number.isInteger(qq) || qq <= 0) continue
    if (seen.has(qq)) continue
    seen.add(qq)
    out.push(qq)
  }
  return out
})

const superDirty = computed(() => superUsersText.value !== superUsersOriginal.value)

async function loadSuperUsers() {
  try {
    const cfg = await getPublicConfig()
    const list = cfg.bot?.super_users || []
    superUsersText.value = list.join('\n')
    superUsersOriginal.value = superUsersText.value
  } catch (err) {
    // 不阻断其他加载；在超级管理员的保存错误区提示。
    superSaveError.value = err instanceof Error ? err.message : '加载超级管理员名单失败。'
  }
}

async function saveSuperUsers() {
  superSaving.value = true
  superSaveError.value = ''
  superSaveSuccess.value = false
  try {
    const list = parsedSuperUsers.value
    await updatePublicConfig({
      // 仅 patch bot.super_users；不传 server，避免触发区服校验。
      bot: { super_users: list },
    })
    superUsersText.value = list.join('\n')
    superUsersOriginal.value = superUsersText.value
    superSaveSuccess.value = true
  } catch (err) {
    superSaveError.value = err instanceof Error ? err.message : '保存失败。'
  } finally {
    superSaving.value = false
  }
}

// ===== Renderer 设置（精度 / 字体 / 渲染结果缓存） =====
const rendererConfig = ref<PublicConfig['renderer'] | null>(null)
const rendererForm = reactive({
  precision: 1.5,
  chart_precision: 4,
  fonts: { body_family: '', score_family: '' },
})
const rendererFormOriginal = ref(JSON.stringify(rendererForm))
const rendererSaving = ref(false)
const rendererError = ref('')
const rendererSuccess = ref('')
const rendererDirty = computed(() => JSON.stringify(rendererForm) !== rendererFormOriginal.value)

const fontsData = ref<RendererFontsResponse | null>(null)
const fontsLoading = ref(false)
const fontsLoaded = computed(() => fontsData.value?.ok === true && (fontsData.value?.total ?? 0) > 0)

const renderCache = ref<RenderCacheStats | null>(null)
const renderCacheLoading = ref(false)
const renderCacheClearing = ref(false)
const renderCacheSaving = ref(false)
const renderCacheForm = reactive({ maxBytesMB: 256, maxEntries: 1024 })

// 渲染并发预算（防 OOM）：低内存机器把 max_concurrency 调到 1-2，
// queue_limit 是排队上限，超出会直接 503 让 Bot 回错而不是把进程打挂。
const budgetStats = ref<RendererBudgetStats | null>(null)
const budgetLoading = ref(false)
const budgetSaving = ref(false)
const budgetForm = reactive({ maxConcurrency: 2, queueLimit: 8, prepareConcurrency: 4 })

async function loadRendererConfig() {
  try {
    const cfg = await getPublicConfig()
    rendererConfig.value = cfg.renderer
    rendererForm.precision = cfg.renderer.precision || 1.5
    rendererForm.chart_precision = cfg.renderer.chart_precision || 4
    rendererForm.fonts.body_family = cfg.renderer.fonts?.body_family ?? ''
    rendererForm.fonts.score_family = cfg.renderer.fonts?.score_family ?? ''
    if (cfg.renderer.budget) {
      budgetForm.maxConcurrency = cfg.renderer.budget.max_concurrency || 2
      budgetForm.queueLimit = cfg.renderer.budget.queue_limit ?? 8
      budgetForm.prepareConcurrency = cfg.renderer.budget.prepare_concurrency || 4
    }
    rendererFormOriginal.value = JSON.stringify(rendererForm)
  } catch (err) {
    rendererError.value = err instanceof Error ? err.message : '加载渲染配置失败。'
  }
}

async function loadFontsInfo() {
  fontsLoading.value = true
  try {
    fontsData.value = await getRendererFonts()
  } catch (err) {
    rendererError.value = err instanceof Error ? err.message : '获取字体列表失败。'
  } finally {
    fontsLoading.value = false
  }
}

async function saveRendererSettings() {
  rendererSaving.value = true
  rendererError.value = ''
  rendererSuccess.value = ''
  try {
    const precision = Number(rendererForm.precision) || 1.5
    const chartPrecision = Number(rendererForm.chart_precision) || 4
    if (precision <= 0 || chartPrecision <= 0) {
      rendererError.value = '精度必须为大于 0 的数值。'
      return
    }
    await updatePublicConfig({
      renderer: {
        precision,
        chart_precision: chartPrecision,
        fonts: {
          body_family: rendererForm.fonts.body_family,
          score_family: rendererForm.fonts.score_family,
        },
      },
    })
    rendererFormOriginal.value = JSON.stringify(rendererForm)
    rendererSuccess.value = '渲染设置已保存并立即生效。'
    void loadRendererConfig()
  } catch (err) {
    rendererError.value = err instanceof Error ? err.message : '保存失败。'
  } finally {
    rendererSaving.value = false
  }
}

async function refreshRenderCacheStats(silent = false) {
  renderCacheLoading.value = true
  try {
    const stats = await getRenderCacheStats()
    renderCache.value = stats
    renderCacheForm.maxBytesMB = Math.round(stats.maxBytes / (1024 * 1024))
    renderCacheForm.maxEntries = stats.maxEntries
  } catch (err) {
    if (!silent) rendererError.value = err instanceof Error ? err.message : '获取渲染缓存状态失败。'
  } finally {
    renderCacheLoading.value = false
  }
}

async function clearRenderCacheAction() {
  renderCacheClearing.value = true
  rendererError.value = ''
  rendererSuccess.value = ''
  try {
    const resp = await apiClearRenderCache()
    renderCache.value = resp.stats
    rendererSuccess.value = resp.message || '渲染缓存已清空。'
  } catch (err) {
    rendererError.value = err instanceof Error ? err.message : '清空渲染缓存失败。'
  } finally {
    renderCacheClearing.value = false
  }
}

async function refreshBudgetStats(silent = false) {
  budgetLoading.value = true
  try {
    const stats = await getRendererBudgetStats()
    budgetStats.value = stats
    if (stats.maxConcurrency > 0) budgetForm.maxConcurrency = stats.maxConcurrency
    if (stats.queueLimit >= 0) budgetForm.queueLimit = stats.queueLimit
    if (stats.prepareConcurrency > 0) budgetForm.prepareConcurrency = stats.prepareConcurrency
  } catch (err) {
    if (!silent) rendererError.value = err instanceof Error ? err.message : '获取渲染并发预算失败。'
  } finally {
    budgetLoading.value = false
  }
}

async function saveBudget() {
  budgetSaving.value = true
  rendererError.value = ''
  rendererSuccess.value = ''
  try {
    const maxConcurrency = Math.max(1, Math.round(budgetForm.maxConcurrency))
    const queueLimit = Math.max(0, Math.round(budgetForm.queueLimit))
    const prepareConcurrency = Math.max(1, Math.round(budgetForm.prepareConcurrency))
    const resp = await updateRendererBudget({
      max_concurrency: maxConcurrency,
      queue_limit: queueLimit,
      prepare_concurrency: prepareConcurrency,
    })
    budgetStats.value = resp.stats
    budgetForm.maxConcurrency = resp.stats.maxConcurrency
    budgetForm.queueLimit = resp.stats.queueLimit
    budgetForm.prepareConcurrency = resp.stats.prepareConcurrency
    rendererSuccess.value = resp.message || '渲染并发预算已更新并立即生效。'
  } catch (err) {
    rendererError.value = err instanceof Error ? err.message : '更新渲染并发预算失败。'
  } finally {
    budgetSaving.value = false
  }
}

async function saveRenderCacheConfig() {
  if (!renderCache.value) return
  renderCacheSaving.value = true
  rendererError.value = ''
  rendererSuccess.value = ''
  try {
    const maxBytes = Math.max(1, Math.round(renderCacheForm.maxBytesMB)) * 1024 * 1024
    const maxEntries = Math.max(1, Math.round(renderCacheForm.maxEntries))
    const resp = await updateRenderCacheConfig({ maxBytes, maxEntries })
    renderCache.value = resp.stats
    renderCacheForm.maxBytesMB = Math.round(resp.stats.maxBytes / (1024 * 1024))
    renderCacheForm.maxEntries = resp.stats.maxEntries
    rendererSuccess.value = resp.message || '渲染缓存上限已更新。'
  } catch (err) {
    rendererError.value = err instanceof Error ? err.message : '更新渲染缓存配置失败。'
  } finally {
    renderCacheSaving.value = false
  }
}

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit++
  }
  return `${value.toFixed(value >= 100 || unit === 0 ? 0 : 1)} ${units[unit]}`
}

onMounted(() => {
  load()
  loadBot()
  loadSuperUsers()
  void loadRendererConfig()
  void loadFontsInfo()
  void refreshRenderCacheStats(true)
  void refreshBudgetStats(true)
})

async function loadBot() {
  loadingBot.value = true
  botError.value = ''
  try {
    status.value = await getStatus()
  } catch (err) {
    botError.value = err instanceof Error ? err.message : '加载 Bot 状态失败。'
  } finally {
    loadingBot.value = false
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    plugins.value = await listPlugins()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败。'
  } finally {
    loading.value = false
  }
}

// 控制台账号 / 修改密码
const auth = useAuthStore()
const pwdForm = reactive({
  old_password: '',
  new_password: '',
  new_password_confirm: '',
})
const pwdSaving = ref(false)
const pwdError = ref('')
const pwdSuccess = ref(false)
const pwdReady = computed(() =>
  pwdForm.old_password.length > 0 &&
  pwdForm.new_password.length >= 8 &&
  pwdForm.new_password === pwdForm.new_password_confirm,
)

async function onSavePassword() {
  pwdError.value = ''
  pwdSuccess.value = false
  pwdSaving.value = true
  try {
    await auth.changePassword({
      old_password: pwdForm.old_password,
      new_password: pwdForm.new_password,
      new_password_confirm: pwdForm.new_password_confirm,
    })
    pwdForm.old_password = ''
    pwdForm.new_password = ''
    pwdForm.new_password_confirm = ''
    pwdSuccess.value = true
  } catch (err: any) {
    pwdError.value = err?.response?.data?.message || (err instanceof Error ? err.message : '修改失败')
  } finally {
    pwdSaving.value = false
  }
}

function categoryLabel(c: PluginListItem['category']) {
  switch (c) {
    case 'official':
      return '官方'
    case 'market':
      return '市场'
    case 'third':
      return '第三方'
    default:
      return c
  }
}
</script>

<style scoped>
.card-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; margin-bottom: 8px; }
.quick-links { display: flex; gap: 12px; flex-wrap: wrap; }
.quick-links .ui-button { text-decoration: none; }
.plugins-table { width: 100%; border-collapse: collapse; margin-top: 12px; font-size: 13px; }
.plugins-table th, .plugins-table td {
  text-align: left; padding: 10px 12px;
  border-bottom: 1px solid var(--border);
}
.plugins-table th { color: var(--muted-foreground); font-weight: 500; font-size: 12px; }
.plugin-name { font-weight: 500; }
.plugin-id { color: var(--muted-foreground); font-size: 11px; }
.dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; vertical-align: middle; }
.dot--ok { background: #5fd49a; box-shadow: 0 0 6px rgba(95,212,154,0.6); }
.dot--off { background: #6c707a; }
.muted { color: var(--muted-foreground); }

/* ---- 超级管理员 ---- */
.super-users { display: flex; flex-direction: column; gap: 8px; margin-top: 6px; }
.super-label { font-size: 12px; color: var(--muted-foreground); font-weight: 500; }
.ui-textarea {
  width: 100%; box-sizing: border-box;
  border: 1px solid var(--input);
  border-radius: 12px; padding: 10px 12px;
  background: rgba(255,255,255,0.9); color: var(--foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px; resize: vertical; min-height: 80px;
}
.ui-textarea:focus { outline: none; border-color: var(--ring); box-shadow: 0 0 0 4px rgba(147,197,253,.34); }
.super-foot { display: flex; justify-content: space-between; align-items: center; gap: 12px; }
.super-foot .hint { font-size: 12px; color: var(--muted-foreground); }

/* ---- 修改密码表单 ---- */
.pwd-form { display: flex; flex-direction: column; gap: 10px; margin-top: 14px; }
.pwd-title { margin: 0; font-size: 14px; font-weight: 600; }
.pwd-row { display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--muted-foreground); }
.pwd-row input { min-height: 36px; }
.info-list { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 8px 16px; margin: 8px 0 0; padding: 0; }
.info-list dt { font-size: 12px; color: var(--muted-foreground); }
.info-list dd { margin: 0; font-size: 14px; }

/* ---- Renderer 卡片 ---- */
.renderer-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px 16px;
  margin: 12px 0;
}
.renderer-field { display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--muted-foreground); }
.renderer-field input,
.renderer-field select {
  min-height: 36px;
  font-family: inherit;
  font-size: 13px;
  color: var(--foreground);
}
.renderer-field--readonly { color: var(--muted-foreground); }
.renderer-field--readonly strong { color: var(--foreground); font-weight: 400; font-size: 12px; line-height: 1.5; }
.renderer-field--full { grid-column: 1 / -1; }

.renderer-subpanel {
  margin-top: 14px;
  padding: 14px;
  border-radius: 12px;
  background: rgba(255,255,255,0.6);
  border: 1px solid var(--border);
}
.renderer-subpanel__header {
  display: flex; align-items: flex-start; justify-content: space-between; gap: 12px;
}
.renderer-subpanel__header strong { display: block; font-size: 13px; font-weight: 600; }
.renderer-subpanel__header span { font-size: 12px; color: var(--muted-foreground); }

.renderer-meter {
  margin: 10px 0 6px; height: 6px; border-radius: 4px;
  background: rgba(165,180,252,0.18); overflow: hidden;
}
.renderer-meter span {
  display: block; height: 100%;
  background: linear-gradient(90deg, #5fd49a, #4dbf86);
  transition: width 200ms ease;
}
.renderer-meta {
  display: flex; flex-wrap: wrap; gap: 8px 16px;
  font-size: 12px; color: var(--muted-foreground); margin-bottom: 8px;
}
.renderer-actions { display: flex; gap: 8px; }
</style>

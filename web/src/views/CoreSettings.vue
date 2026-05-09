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
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import {
  getPublicConfig,
  getStatus,
  listPlugins,
  updatePublicConfig,
  type PluginListItem,
} from '../api/client'
import type { RuntimeStatus } from '../api/types'
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

onMounted(() => {
  load()
  loadBot()
  loadSuperUsers()
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
  border-bottom: 1px solid var(--border-default, rgba(255,255,255,0.06));
}
.plugins-table th { color: var(--text-muted); font-weight: 500; font-size: 12px; }
.plugin-name { font-weight: 500; }
.plugin-id { color: var(--text-muted); font-size: 11px; }
.dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; vertical-align: middle; }
.dot--ok { background: #5fd49a; box-shadow: 0 0 6px rgba(95,212,154,0.6); }
.dot--off { background: #6c707a; }
.muted { color: var(--text-muted); }

/* ---- 超级管理员 ---- */
.super-users { display: flex; flex-direction: column; gap: 8px; margin-top: 6px; }
.super-label { font-size: 12px; color: var(--text-muted); font-weight: 500; }
.ui-textarea {
  width: 100%; box-sizing: border-box;
  border: 1px solid var(--border-default, rgba(255,255,255,0.08));
  border-radius: 12px; padding: 10px 12px;
  background: rgba(255,255,255,0.04); color: var(--text);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px; resize: vertical; min-height: 80px;
}
.ui-textarea:focus { outline: none; border-color: var(--accent, #5fd49a); }
.super-foot { display: flex; justify-content: space-between; align-items: center; gap: 12px; }
.super-foot .hint { font-size: 12px; color: var(--text-muted); }
</style>

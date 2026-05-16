<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="S3 Backup"
      title="备份恢复"
      subtitle="将运行时数据目录（Docker 内通常为 /app/data）备份到 S3 兼容对象存储，并可从远端归档恢复。"
    >
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="loadAll">刷新</UiButton>
        <UiButton size="sm" :loading="saving" @click="saveConfig">保存配置</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="操作失败">{{ error }}</UiAlert>
    <UiAlert v-if="success" variant="info" title="操作完成">{{ success }}</UiAlert>

    <UiAlert variant="warning" title="请保护好备份桶">
      备份包会包含 <code>data/config.yml</code>、数据库与插件配置；如果配置里保存了 S3 密钥，远端备份也会包含这些密钥。请使用私有桶与最小权限 Access Key。
    </UiAlert>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>S3 兼容存储配置</h2>
          <p>支持 MinIO、Cloudflare R2 以及其他兼容 AWS Signature V4 的对象存储。</p>
        </div>
        <UiBadge :variant="configForm.configured ? 'success' : 'warning'">
          {{ configForm.configured ? '已配置' : '未完成' }}
        </UiBadge>
      </div>

      <div class="settings-form settings-form--region">
        <label class="settings-field">
          <span>Endpoint</span>
          <input v-model.trim="configForm.endpoint" class="ui-input" placeholder="minio.example.com:9000" />
        </label>
        <label class="settings-field">
          <span>Region</span>
          <input v-model.trim="configForm.region" class="ui-input" placeholder="us-east-1 / 可留空" />
        </label>
        <label class="settings-field">
          <span>Bucket</span>
          <input v-model.trim="configForm.bucket" class="ui-input" placeholder="moebot-backups" />
        </label>
        <label class="settings-field">
          <span>Prefix</span>
          <input v-model.trim="configForm.prefix" class="ui-input" placeholder="moebot-next/backups" />
        </label>

        <label class="settings-field">
          <span>Access Key</span>
          <input v-model.trim="secrets.access_key" class="ui-input" :placeholder="configForm.access_key_set ? '已配置；留空不覆盖' : '请输入 Access Key'" />
        </label>
        <label class="settings-field">
          <span>Secret Key</span>
          <input v-model.trim="secrets.secret_key" type="password" class="ui-input" :placeholder="configForm.secret_key_set ? '已配置；留空不覆盖' : '请输入 Secret Key'" />
        </label>
        <label class="settings-field settings-field--full">
          <span>Session Token（可选）</span>
          <input v-model.trim="secrets.session_token" type="password" class="ui-input" :placeholder="configForm.session_token_set ? '已配置；留空不覆盖' : 'STS 临时凭证可填写'" />
        </label>

        <label class="settings-field">
          <span>Use SSL</span>
          <select v-model="configForm.use_ssl" class="ui-select">
            <option :value="true">HTTPS</option>
            <option :value="false">HTTP</option>
          </select>
        </label>
        <label class="settings-field">
          <span>Path Style</span>
          <select v-model="configForm.force_path_style" class="ui-select">
            <option :value="true">强制 path-style（MinIO 推荐）</option>
            <option :value="false">虚拟主机风格</option>
          </select>
        </label>
        <label class="settings-field">
          <span>数据目录</span>
          <input v-model.trim="configForm.data_dir" class="ui-input" placeholder="./data 或 /app/data" />
        </label>
        <label class="settings-field">
          <span>临时目录</span>
          <input v-model.trim="configForm.temp_dir" class="ui-input" placeholder="./data/backups/tmp" />
        </label>
        <label class="settings-field settings-field--full">
          <span>排除规则（每行一个）</span>
          <textarea
            v-model="excludePatternsText"
            class="ui-textarea"
            rows="4"
            placeholder="cache/**&#10;backups/tmp/**&#10;*.tmp&#10;*.restore-backup-*"
          />
        </label>
      </div>

      <div class="settings-preview" style="margin-top: 14px;">
        <div><span>目标</span><code>{{ targetText }}</code></div>
        <div><span>密钥状态</span><code>AK {{ configForm.access_key_set ? '已保存' : '未保存' }} · SK {{ configForm.secret_key_set ? '已保存' : '未保存' }} · STS {{ configForm.session_token_set ? '已保存' : '未保存' }}</code></div>
        <div><span>排除</span><code>{{ parsedExcludePatterns.length ? parsedExcludePatterns.join(' / ') : '不排除额外路径' }}</code></div>
        <div><span>说明</span><code>密钥输入框留空不会覆盖旧值；如需清空请勾选下方清空项后保存。</code></div>
      </div>

      <div class="settings-actions-row" style="margin-top: 14px;">
        <label class="backup-check"><input v-model="clearSecrets.access_key" type="checkbox" /> 清空 Access Key</label>
        <label class="backup-check"><input v-model="clearSecrets.secret_key" type="checkbox" /> 清空 Secret Key</label>
        <label class="backup-check"><input v-model="clearSecrets.session_token" type="checkbox" /> 清空 Session Token</label>
      </div>

      <div class="super-foot" style="margin-top: 16px;">
        <span class="hint">保存会写入 data/config.yml；测试连接会尝试列出备份 prefix。</span>
        <div class="renderer-actions">
          <UiButton variant="outline" size="sm" :loading="testing" @click="testConnection">测试连接</UiButton>
          <UiButton size="sm" :loading="saving" @click="saveConfig">保存配置</UiButton>
        </div>
      </div>
    </UiCard>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>定期备份设置</h2>
          <p>开启后主进程会按固定间隔自动创建 S3 备份；例如间隔 24 小时就是每天一次。</p>
        </div>
        <UiBadge :variant="configForm.schedule_enabled ? 'success' : 'secondary'">
          {{ scheduleStatusText }}
        </UiBadge>
      </div>

      <div class="settings-form settings-form--region">
        <div class="settings-field backup-toggle-field">
          <span>自动备份</span>
          <label class="backup-check backup-check--switch">
            <input v-model="configForm.schedule_enabled" type="checkbox" />
            {{ configForm.schedule_enabled ? '已启用' : '已关闭' }}
          </label>
        </div>
        <label class="settings-field">
          <span>间隔小时</span>
          <input
            v-model.number="configForm.schedule_interval_hours"
            class="ui-input"
            type="number"
            min="1"
            step="1"
            placeholder="24"
          />
        </label>
        <div class="settings-field settings-field--full schedule-presets">
          <span>快捷设置</span>
          <div class="renderer-actions">
            <UiButton variant="outline" size="sm" @click="setScheduleHours(6)">每 6 小时</UiButton>
            <UiButton variant="outline" size="sm" @click="setScheduleHours(12)">每 12 小时</UiButton>
            <UiButton variant="outline" size="sm" @click="setScheduleHours(24)">每天一次</UiButton>
            <UiButton variant="outline" size="sm" @click="setScheduleHours(168)">每周一次</UiButton>
          </div>
        </div>
      </div>

      <div class="settings-preview" style="margin-top: 14px;">
        <div><span>当前计划</span><code>{{ scheduleSummaryText }}</code></div>
        <div><span>保存方式</span><code>点击“保存配置”后立即生效，无需重启。</code></div>
        <div><span>执行要求</span><code>{{ configForm.configured ? 'S3 已配置，自动任务可正常上传。' : 'S3 尚未配置完整，自动执行时会失败并等待下次。' }}</code></div>
      </div>
    </UiCard>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>创建备份</h2>
          <p>会先将数据目录打成 <code>.tar.gz</code>，再上传到 S3 兼容存储。运行中备份会包含 SQLite WAL 文件。</p>
        </div>
        <UiButton size="sm" :loading="creating" :disabled="!configForm.configured" @click="createNow">立即备份</UiButton>
      </div>
      <div class="settings-preview">
        <div><span>数据目录</span><code>{{ configForm.data_dir || '-' }}</code></div>
        <div><span>临时目录</span><code>{{ configForm.temp_dir || '-' }}</code></div>
        <div><span>定期备份</span><code>{{ scheduleSummaryText }}</code></div>
        <div><span>最近结果</span><code>{{ lastResult || '暂无' }}</code></div>
      </div>
    </UiCard>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>远端备份</h2>
          <p>恢复会替换当前数据目录，并把旧目录保留为 <code>.restore-backup-*</code>；恢复后请重启进程/容器。</p>
        </div>
        <UiButton variant="outline" size="sm" :loading="listing" @click="loadObjects">刷新列表</UiButton>
      </div>

      <table v-if="objects.length" class="plugins-table">
        <thead>
          <tr>
            <th>备份</th>
            <th>大小</th>
            <th>最后修改</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="obj in objects" :key="obj.key">
            <td>
              <div class="plugin-name">{{ obj.name }}</div>
              <div class="plugin-id">{{ obj.key }}</div>
            </td>
            <td>{{ formatBytes(obj.size) }}</td>
            <td>{{ formatTime(obj.last_modified) }}</td>
            <td>
              <div class="renderer-actions">
                <UiButton variant="outline" size="sm" :loading="restoringKey === obj.key" @click="restore(obj.key)">恢复</UiButton>
                <UiButton variant="destructive" size="sm" :loading="deletingKey === obj.key" @click="remove(obj.key)">删除</UiButton>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else-if="!listing" class="muted">暂无远端备份，保存配置并创建一次备份后会显示在这里。</div>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  createBackup,
  deleteBackup,
  getBackupConfig,
  listBackups,
  restoreBackup,
  testBackupConnection,
  updateBackupConfig,
} from '../api/client'
import type { BackupObject, BackupPublicConfig } from '../api/types'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const listing = ref(false)
const creating = ref(false)
const restoringKey = ref('')
const deletingKey = ref('')
const error = ref('')
const success = ref('')
const lastResult = ref('')
const objects = ref<BackupObject[]>([])

const configForm = reactive<BackupPublicConfig>({
  data_dir: './data',
  temp_dir: './data/backups/tmp',
  exclude_patterns: ['cache/**', 'backups/tmp/**', '*.tmp', '*.restore-backup-*'],
  endpoint: '',
  region: '',
  bucket: '',
  prefix: 'moebot-next/backups',
  use_ssl: true,
  force_path_style: true,
  schedule_enabled: false,
  schedule_interval_hours: 24,
  access_key_set: false,
  secret_key_set: false,
  session_token_set: false,
  configured: false,
})

const secrets = reactive({ access_key: '', secret_key: '', session_token: '' })
const clearSecrets = reactive({ access_key: false, secret_key: false, session_token: false })
const excludePatternsText = ref('cache/**\nbackups/tmp/**\n*.tmp\n*.restore-backup-*')
const parsedExcludePatterns = computed(() => {
  const seen = new Set<string>()
  const out: string[] = []
  for (const raw of excludePatternsText.value.split(/[\r\n]+/)) {
    const item = raw.trim().replace(/^\/+|\/+$/g, '')
    if (!item || seen.has(item)) continue
    seen.add(item)
    out.push(item)
  }
  return out
})

const targetText = computed(() => {
  const scheme = configForm.use_ssl ? 'https' : 'http'
  const endpoint = configForm.endpoint || '<endpoint>'
  const bucket = configForm.bucket || '<bucket>'
  const prefix = configForm.prefix || ''
  return `${scheme}://${endpoint}/${bucket}/${prefix}`
})

const normalizedScheduleHours = computed(() => normalizeScheduleHours(configForm.schedule_interval_hours))
const scheduleStatusText = computed(() => (configForm.schedule_enabled ? '自动备份已启用' : '自动备份关闭'))
const scheduleSummaryText = computed(() => {
  if (!configForm.schedule_enabled) return '已关闭'
  return `每 ${normalizedScheduleHours.value} 小时自动备份一次${normalizedScheduleHours.value === 24 ? '（每天一次）' : ''}`
})

onMounted(() => {
  void loadAll()
})

async function loadAll() {
  loading.value = true
  error.value = ''
  try {
    await loadConfig()
    await loadObjects(true)
  } finally {
    loading.value = false
  }
}

async function loadConfig() {
  try {
    const cfg = await getBackupConfig()
    Object.assign(configForm, cfg)
    configForm.schedule_interval_hours = normalizeScheduleHours(configForm.schedule_interval_hours)
    excludePatternsText.value = (cfg.exclude_patterns ?? []).join('\n')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载备份配置失败。'
  }
}

async function saveConfig() {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    const resp = await updateBackupConfig({
      data_dir: configForm.data_dir,
      temp_dir: configForm.temp_dir,
      exclude_patterns: parsedExcludePatterns.value,
      endpoint: configForm.endpoint,
      region: configForm.region,
      bucket: configForm.bucket,
      prefix: configForm.prefix,
      use_ssl: configForm.use_ssl,
      force_path_style: configForm.force_path_style,
      schedule_enabled: configForm.schedule_enabled,
      schedule_interval_hours: normalizedScheduleHours.value,
      access_key: secrets.access_key,
      secret_key: secrets.secret_key,
      session_token: secrets.session_token,
      clear_access_key: clearSecrets.access_key,
      clear_secret_key: clearSecrets.secret_key,
      clear_session_token: clearSecrets.session_token,
    })
    if (resp.config) Object.assign(configForm, resp.config)
    configForm.schedule_interval_hours = normalizeScheduleHours(configForm.schedule_interval_hours)
    secrets.access_key = ''
    secrets.secret_key = ''
    secrets.session_token = ''
    clearSecrets.access_key = false
    clearSecrets.secret_key = false
    clearSecrets.session_token = false
    success.value = resp.message || '备份配置已保存。'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存备份配置失败。'
  } finally {
    saving.value = false
  }
}

async function testConnection() {
  testing.value = true
  error.value = ''
  success.value = ''
  try {
    const resp = await testBackupConnection()
    success.value = resp.message || 'S3 连接测试成功。'
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'S3 连接测试失败。'
  } finally {
    testing.value = false
  }
}

async function loadObjects(silent = false) {
  listing.value = true
  if (!silent) error.value = ''
  try {
    const resp = await listBackups()
    objects.value = resp.data ?? []
  } catch (err) {
    if (!silent) error.value = err instanceof Error ? err.message : '加载远端备份失败。'
  } finally {
    listing.value = false
  }
}

async function createNow() {
  creating.value = true
  error.value = ''
  success.value = ''
  try {
    const resp = await createBackup()
    const obj = resp.result?.object
    lastResult.value = obj ? `${obj.name} · ${formatBytes(obj.size)} · ${resp.result?.duration_ms ?? 0}ms` : resp.message
    success.value = resp.message || '备份已创建。'
    await loadObjects(true)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '创建备份失败。'
  } finally {
    creating.value = false
  }
}

async function restore(key: string) {
  const ok = window.confirm(`恢复会替换当前数据目录，并要求重启 Moebot。确定恢复？\n\n${key}`)
  if (!ok) return
  restoringKey.value = key
  error.value = ''
  success.value = ''
  try {
    const resp = await restoreBackup(key)
    const result = resp.result
    success.value = result
      ? `${resp.message}；旧目录保留在 ${result.backup_data_dir}`
      : resp.message || '恢复完成，请重启。'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '恢复失败。'
  } finally {
    restoringKey.value = ''
  }
}

async function remove(key: string) {
  const ok = window.confirm(`确定删除远端备份？此操作不可撤销。\n\n${key}`)
  if (!ok) return
  deletingKey.value = key
  error.value = ''
  success.value = ''
  try {
    const resp = await deleteBackup(key)
    success.value = resp.message || '远端备份已删除。'
    await loadObjects(true)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除失败。'
  } finally {
    deletingKey.value = ''
  }
}

function normalizeScheduleHours(value: number): number {
  if (!Number.isFinite(value)) return 24
  return Math.max(1, Math.floor(value))
}

function setScheduleHours(hours: number) {
  configForm.schedule_interval_hours = normalizeScheduleHours(hours)
  configForm.schedule_enabled = true
}

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit++
  }
  return `${value.toFixed(value >= 100 || unit === 0 ? 0 : 1)} ${units[unit]}`
}

function formatTime(value: string): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}
</script>

<style scoped>
.backup-check {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--muted-foreground);
  font-size: 12px;
  font-weight: 850;
}

.backup-check--switch {
  min-height: 36px;
  color: var(--foreground);
}

.backup-toggle-field {
  justify-content: flex-start;
}

.schedule-presets > span {
  display: block;
  margin-bottom: 8px;
  color: var(--muted-foreground);
  font-size: 12px;
  font-weight: 850;
}
</style>

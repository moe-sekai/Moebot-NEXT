<template>
  <k-layout>
    <div class="sekai-api-config">
      <h1>SEKAI API 配置</h1>
      <p class="description">
        配置 SEKAI 游戏 API 端点，用于获取玩家数据、实时排行等功能。
        <strong>不配置也不影响基础查询功能。</strong>
      </p>

      <!-- Existing endpoints list -->
      <div class="endpoints-list">
        <div v-for="(ep, idx) in config.endpoints" :key="idx" class="endpoint-card">
          <div class="endpoint-header">
            <span class="endpoint-name">
              <span class="status-dot" :class="ep.enabled ? 'active' : 'inactive'" />
              {{ ep.name || '未命名端点' }}
            </span>
            <span class="endpoint-region">{{ ep.region?.toUpperCase() }}</span>
            <div class="endpoint-actions">
              <button class="btn btn-sm" @click="testEndpoint(idx)">
                {{ testing === idx ? '测试中...' : '测试连接' }}
              </button>
              <button class="btn btn-sm btn-danger" @click="removeEndpoint(idx)">删除</button>
            </div>
          </div>

          <div class="endpoint-body">
            <div class="form-row">
              <label>启用</label>
              <input type="checkbox" v-model="ep.enabled" />
            </div>
            <div class="form-row">
              <label>名称</label>
              <input v-model="ep.name" placeholder="如: JP Server" />
            </div>
            <div class="form-row">
              <label>区服</label>
              <select v-model="ep.region">
                <option value="jp">日服 (JP)</option>
                <option value="en">国际服 (EN)</option>
                <option value="tw">台服 (TW)</option>
                <option value="kr">韩服 (KR)</option>
              </select>
            </div>
            <div class="form-row">
              <label>API 地址</label>
              <input v-model="ep.baseUrl" placeholder="https://api.example.com" />
            </div>
            <div class="form-row">
              <label>超时 (ms)</label>
              <input type="number" v-model.number="ep.timeout" />
            </div>
            <div class="form-row">
              <label>频率限制 (次/分)</label>
              <input type="number" v-model.number="ep.rateLimit" />
            </div>
            <div class="form-row">
              <label>代理地址</label>
              <input v-model="ep.proxy" placeholder="http://proxy:8080 (可选)" />
            </div>

            <!-- Headers (key-value pairs) -->
            <div class="form-section">
              <label>请求头 (Headers)</label>
              <div v-for="(_, hIdx) in ep.headerEntries" :key="hIdx" class="header-row">
                <input v-model="ep.headerEntries[hIdx].key" placeholder="Header-Name" />
                <input v-model="ep.headerEntries[hIdx].value" placeholder="Header-Value" type="password" />
                <button class="btn btn-sm btn-danger" @click="removeHeader(idx, hIdx)">×</button>
              </div>
              <button class="btn btn-sm" @click="addHeader(idx)">+ 添加请求头</button>
            </div>

            <!-- Test result -->
            <div v-if="testResults[idx]" class="test-result" :class="testResults[idx].success ? 'success' : 'error'">
              <template v-if="testResults[idx].success">
                ✅ 连接成功 (HTTP {{ testResults[idx].status }})
              </template>
              <template v-else>
                ❌ 连接失败: {{ testResults[idx].error || `HTTP ${testResults[idx].status}` }}
              </template>
            </div>
          </div>
        </div>
      </div>

      <!-- Add new endpoint button -->
      <button class="btn btn-primary" @click="addEndpoint" style="margin-top: 16px">
        + 添加 API 端点
      </button>

      <!-- Save button -->
      <div class="save-bar">
        <span v-if="saveStatus" :class="saveStatus === 'success' ? 'text-success' : 'text-error'">
          {{ saveStatus === 'success' ? '✅ 已保存' : '❌ 保存失败' }}
        </span>
        <button class="btn btn-primary" @click="saveConfig" :disabled="saving">
          {{ saving ? '保存中...' : '保存配置' }}
        </button>
      </div>
    </div>
  </k-layout>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { send } from '@koishijs/client'

interface HeaderEntry {
  key: string
  value: string
}

interface EndpointConfig {
  enabled: boolean
  name: string
  region: 'jp' | 'en' | 'tw' | 'kr'
  baseUrl: string
  timeout: number
  rateLimit: number
  proxy: string
  headers: Record<string, string>
  headerEntries: HeaderEntry[]
}

const config = reactive<{ endpoints: EndpointConfig[] }>({ endpoints: [] })
const testing = ref<number | null>(null)
const testResults = reactive<Record<number, any>>({})
const saving = ref(false)
const saveStatus = ref<'success' | 'error' | null>(null)

onMounted(async () => {
  const data = await send('moebot/sekai-api/config')
  if (data?.endpoints) {
    config.endpoints = data.endpoints.map((ep: any) => ({
      ...ep,
      headerEntries: Object.entries(ep.headers || {}).map(([key, value]) => ({ key, value: value as string })),
    }))
  }
})

function addEndpoint() {
  config.endpoints.push({
    enabled: true,
    name: '',
    region: 'jp',
    baseUrl: '',
    timeout: 10000,
    rateLimit: 30,
    proxy: '',
    headers: {},
    headerEntries: [],
  })
}

function removeEndpoint(idx: number) {
  config.endpoints.splice(idx, 1)
}

function addHeader(epIdx: number) {
  config.endpoints[epIdx].headerEntries.push({ key: '', value: '' })
}

function removeHeader(epIdx: number, hIdx: number) {
  config.endpoints[epIdx].headerEntries.splice(hIdx, 1)
}

async function testEndpoint(idx: number) {
  testing.value = idx
  delete testResults[idx]

  const ep = config.endpoints[idx]
  // Convert header entries to object
  const headers: Record<string, string> = {}
  for (const h of ep.headerEntries) {
    if (h.key) headers[h.key] = h.value
  }

  const result = await send('moebot/sekai-api/test', {
    ...ep,
    headers,
  })
  testResults[idx] = result
  testing.value = null
}

async function saveConfig() {
  saving.value = true
  saveStatus.value = null

  // Convert headerEntries back to headers objects
  const payload = {
    endpoints: config.endpoints.map(ep => {
      const headers: Record<string, string> = {}
      for (const h of ep.headerEntries) {
        if (h.key) headers[h.key] = h.value
      }
      return {
        enabled: ep.enabled,
        name: ep.name,
        region: ep.region,
        baseUrl: ep.baseUrl,
        timeout: ep.timeout,
        rateLimit: ep.rateLimit,
        proxy: ep.proxy || undefined,
        headers,
      }
    }),
  }

  const result = await send('moebot/sekai-api/save', payload)
  saveStatus.value = result?.success ? 'success' : 'error'
  saving.value = false

  // Clear status after 3s
  setTimeout(() => { saveStatus.value = null }, 3000)
}
</script>

<style scoped>
.sekai-api-config {
  padding: 24px;
  max-width: 800px;
  margin: 0 auto;
}
h1 { margin: 0 0 8px; }
.description { color: #888; margin-bottom: 24px; line-height: 1.6; }
.description strong { color: var(--k-color-active); }

.endpoint-card {
  background: var(--k-card-bg);
  border-radius: 12px;
  margin-bottom: 16px;
  border: 1px solid var(--k-color-border);
  overflow: hidden;
}
.endpoint-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--k-side-bg);
  border-bottom: 1px solid var(--k-color-border);
}
.endpoint-name { font-weight: 600; flex: 1; }
.endpoint-region {
  background: var(--k-color-active);
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
}
.endpoint-actions { display: flex; gap: 8px; }
.endpoint-body { padding: 16px; }

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 8px;
}
.status-dot.active { background: #44cc88; }
.status-dot.inactive { background: #888; }

.form-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}
.form-row label {
  width: 120px;
  flex-shrink: 0;
  color: #888;
  font-size: 14px;
}
.form-row input[type="text"],
.form-row input[type="number"],
.form-row input[type="password"],
.form-row input:not([type]),
.form-row select {
  flex: 1;
  padding: 8px 12px;
  border-radius: 6px;
  border: 1px solid var(--k-color-border);
  background: var(--k-side-bg);
  color: inherit;
  font-size: 14px;
}

.form-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--k-color-border);
}
.form-section > label {
  display: block;
  color: #888;
  font-size: 14px;
  margin-bottom: 8px;
}
.header-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}
.header-row input {
  flex: 1;
  padding: 6px 10px;
  border-radius: 6px;
  border: 1px solid var(--k-color-border);
  background: var(--k-side-bg);
  color: inherit;
  font-size: 13px;
}

.test-result {
  margin-top: 12px;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 14px;
}
.test-result.success { background: rgba(68, 204, 136, 0.15); color: #44cc88; }
.test-result.error { background: rgba(255, 68, 102, 0.15); color: #ff4466; }

.btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid var(--k-color-border);
  background: var(--k-card-bg);
  color: inherit;
  cursor: pointer;
  font-size: 14px;
}
.btn:hover { opacity: 0.85; }
.btn-primary { background: var(--k-color-active); color: white; border-color: transparent; }
.btn-danger { color: #ff4466; border-color: #ff4466; }
.btn-sm { padding: 4px 10px; font-size: 12px; }

.save-bar {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--k-color-border);
}
.text-success { color: #44cc88; }
.text-error { color: #ff4466; }
</style>

<template>
  <UiCard class="conn-guide">
    <header class="conn-head">
      <div>
        <h2>连接指南</h2>
        <p>把网关地址配到 OneBot 客户端；下游 bot 应用的地址用于让 Moebot 主动连接它们。</p>
      </div>
      <UiBadge :variant="status?.running ? 'success' : 'secondary'">
        {{ status?.running ? '网关运行中' : '网关未启动' }}
      </UiBadge>
    </header>

    <div class="conn-grid">
      <!-- 对外 -->
      <section class="conn-block conn-block--ext">
        <div class="conn-block-head">
          <span class="conn-tag" aria-hidden="true">↘</span>
          <div>
            <h3>对外暴露地址</h3>
            <p>OneBot 客户端（NapCat / Lagrange / LLOneBot 等）请把「反向 WS」指向这里。</p>
          </div>
        </div>
        <div class="conn-addr">
          <code>{{ externalAddress || '（请先启用并保存网关）' }}</code>
          <UiButton
            size="sm"
            variant="outline"
            :disabled="!externalAddress"
            @click="copy(externalAddress, 'ext')"
          >
            {{ copied === 'ext' ? '已复制' : '复制' }}
          </UiButton>
        </div>
        <div class="conn-meta">
          <UiBadge variant="secondary">已连接 × {{ status?.upstreams?.length ?? 0 }}</UiBadge>
          <span v-if="status?.upstream_up && status?.upstreams?.length" class="conn-meta-dim">
            {{ upstreamSelfIDs }}
          </span>
        </div>
      </section>

      <!-- 对内 -->
      <section class="conn-block conn-block--int">
        <div class="conn-block-head">
          <span class="conn-tag" aria-hidden="true">↗</span>
          <div>
            <h3>对内暴露地址</h3>
            <p>下游 bot 应用要监听以下地址；Moebot 会作为 OneBot 客户端主动连入。</p>
          </div>
        </div>
        <ul v-if="apps.length" class="conn-list">
          <li v-for="a in apps" :key="a.name" class="conn-row">
            <UiBadge :variant="a.connected ? 'success' : 'secondary'">
              {{ a.connected ? '已连接' : '未连接' }}
            </UiBadge>
            <span class="conn-name">
              {{ a.name }}<UiBadge v-if="a.builtin" variant="secondary">内置</UiBadge>
            </span>
            <code class="conn-uri">{{ a.uri || '—' }}</code>
            <UiButton
              size="sm"
              variant="ghost"
              :disabled="!a.uri"
              @click="copy(a.uri, `app-${a.name}`)"
            >
              {{ copied === `app-${a.name}` ? '已复制' : '复制' }}
            </UiButton>
          </li>
        </ul>
        <p v-else class="conn-empty">暂无下游应用。</p>
      </section>
    </div>
  </UiCard>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import UiBadge from '../ui/UiBadge.vue'
import UiButton from '../ui/UiButton.vue'
import UiCard from '../ui/UiCard.vue'
import type { FilterClientStatus, FilterStatus } from '../../api/types'

const props = defineProps<{
  status: FilterStatus | null
  /** Host configured on the gateway (may be 0.0.0.0 / :: / a specific address). */
  host?: string
  /** Listening port from gateway settings; falls back to status.listen parsing. */
  port?: number | string
  suffix?: string
  apps: FilterClientStatus[]
}>()

const externalAddress = computed(() => {
  const host = props.host || hostFromListen(props.status?.listen) || '0.0.0.0'
  const port = props.port || portFromListen(props.status?.listen) || ''
  if (!port) return ''
  const path = props.suffix || props.status?.suffix || '/ws'
  return `ws://${host}:${port}${path}`
})

const upstreamSelfIDs = computed(() => {
  const ids = (props.status?.upstreams || []).map((u) => u.self_id).filter(Boolean)
  return ids.length ? `self-id: ${ids.join(', ')}` : ''
})

const copied = ref<string | null>(null)

async function copy(text: string | undefined, key: string) {
  if (!text) return
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
    } else {
      // Fallback for non-secure contexts.
      const ta = document.createElement('textarea')
      ta.value = text
      ta.style.position = 'fixed'
      ta.style.opacity = '0'
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
    }
    copied.value = key
    setTimeout(() => {
      if (copied.value === key) copied.value = null
    }, 1500)
  } catch {
    /* ignore */
  }
}

function hostFromListen(listen?: string): string {
  if (!listen) return ''
  const idx = listen.lastIndexOf(':')
  if (idx === -1) return listen
  return listen.slice(0, idx)
}

function portFromListen(listen?: string): string {
  if (!listen) return ''
  const idx = listen.lastIndexOf(':')
  if (idx === -1) return ''
  return listen.slice(idx + 1)
}
</script>

<style scoped>
.conn-guide { padding: 0; overflow: hidden; }
.conn-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  padding: 18px 20px 12px;
  border-bottom: 1px solid var(--border, #e4e4e7);
}
.conn-head h2 { margin: 0; font-size: 16px; }
.conn-head p { margin: 4px 0 0; color: var(--muted-foreground, #71717a); font-size: 13px; }

.conn-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0;
}
@media (max-width: 880px) {
  .conn-grid { grid-template-columns: 1fr; }
}

.conn-block { padding: 18px 20px; display: flex; flex-direction: column; gap: 12px; }
.conn-block--ext { background: linear-gradient(135deg, rgba(59,130,246,0.06), transparent 60%); }
.conn-block--int { background: linear-gradient(135deg, rgba(34,197,94,0.06), transparent 60%); border-left: 1px solid var(--border, #e4e4e7); }
@media (max-width: 880px) {
  .conn-block--int { border-left: none; border-top: 1px solid var(--border, #e4e4e7); }
}

.conn-block-head { display: flex; gap: 10px; align-items: flex-start; }
.conn-block-head h3 { margin: 0; font-size: 14px; }
.conn-block-head p { margin: 2px 0 0; font-size: 12px; color: var(--muted-foreground, #71717a); }
.conn-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 700;
  background: var(--muted, #f4f4f5);
  color: var(--foreground, #18181b);
  flex-shrink: 0;
}
.conn-block--ext .conn-tag { background: rgba(59,130,246,0.15); color: #1d4ed8; }
.conn-block--int .conn-tag { background: rgba(34,197,94,0.15); color: #15803d; }

.conn-addr {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--muted, #f4f4f5);
  padding: 12px 14px;
  border-radius: 10px;
  border: 1px dashed var(--border, #e4e4e7);
}
.conn-addr code {
  flex: 1 1 auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
  word-break: break-all;
  color: var(--foreground, #18181b);
}

.conn-meta { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
.conn-meta-dim { font-size: 12px; color: var(--muted-foreground, #71717a); }

.conn-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 6px; }
.conn-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  background: var(--background, #fff);
  border: 1px solid var(--border, #e4e4e7);
  border-radius: 8px;
  flex-wrap: wrap;
}
.conn-name { display: inline-flex; gap: 4px; align-items: center; font-weight: 600; font-size: 13px; }
.conn-uri {
  flex: 1 1 200px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  color: var(--muted-foreground, #71717a);
  word-break: break-all;
}
.conn-empty { font-size: 12px; color: var(--muted-foreground, #71717a); margin: 0; font-style: italic; }
</style>

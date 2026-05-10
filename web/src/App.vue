<template>
  <n-config-provider>
    <n-loading-bar-provider>
      <n-notification-provider>
        <n-dialog-provider>
          <n-message-provider>
            <!-- 鉴权 / 首启页：跳过完整控制台 shell，由视图自身全屏渲染。 -->
            <RouterView v-if="isAuthRoute" />
            <div v-else class="app-shell" :class="{ 'app-shell--collapsed': collapsed }">
        <aside class="app-sidebar" aria-label="Moebot NEXT 控制台侧边栏">
          <button
            class="sidebar-toggle"
            type="button"
            :title="collapsed ? '展开侧栏' : '收起侧栏'"
            :aria-label="collapsed ? '展开侧栏' : '收起侧栏'"
            @click="toggleSidebar"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="15 18 9 12 15 6" />
            </svg>
          </button>
          <RouterLink to="/" class="brand" aria-label="返回概览">
            <MoebotLogo color="var(--accent-pink)" :height="collapsed ? 36 : 60" />
          </RouterLink>
          <ConsoleNav />
        </aside>

        <div class="app-main">
          <header class="topbar">
            <div>
              <div class="topbar__eyebrow">{{ page.eyebrow }}</div>
              <div class="topbar__title">{{ page.title }}</div>
            </div>
            <div class="topbar__status">
              <span class="pulse-dot" aria-hidden="true" />
              <span>Console Ready</span>
            </div>
          </header>

          <div class="mobile-nav-wrap">
            <ConsoleNav />
          </div>

          <main class="content">
            <RouterView />
          </main>
          <footer class="app-footer">
            <span>Moebot NEXT<span v-if="auth.nickname"> (deployed by {{ auth.nickname }})</span></span>
            <button v-if="auth.isLoggedIn" class="footer-logout" type="button" @click="onLogout">退出登录</button>
          </footer>
        </div>
            </div>
          </n-message-provider>
        </n-dialog-provider>
      </n-notification-provider>
    </n-loading-bar-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { NConfigProvider, NMessageProvider, NDialogProvider, NNotificationProvider, NLoadingBarProvider } from 'naive-ui'
import ConsoleNav from './components/layout/ConsoleNav.vue'
import MoebotLogo from './components/MoebotLogo.vue'
import { getPageDescription } from './navigation'
import { useAuthStore } from './stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const page = computed(() => getPageDescription(route.name))
// 鉴权 / 首启视图自带全屏布局，App 层不再叠加侧边栏。
const isAuthRoute = computed(() => route.path === '/login' || route.path === '/setup')

async function onLogout() {
  auth.logout()
  await router.replace('/login')
}

// 侧栏折叠状态：持久化到 localStorage，刷新后保持。
const SIDEBAR_KEY = 'moebot.sidebar.collapsed'
const collapsed = ref(typeof localStorage !== 'undefined' && localStorage.getItem(SIDEBAR_KEY) === '1')
function toggleSidebar() { collapsed.value = !collapsed.value }
watch(collapsed, value => {
  try { localStorage.setItem(SIDEBAR_KEY, value ? '1' : '0') } catch {}
})
</script>

<style scoped>
.app-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 24px;
  font-size: 12px;
  color: var(--text-muted, #8a8f99);
  border-top: 1px solid var(--border, #2a2f38);
  background: var(--surface, #181b22);
}
.footer-logout {
  background: transparent;
  color: inherit;
  border: 1px solid var(--border, #2a2f38);
  border-radius: 6px;
  padding: 4px 10px;
  font-size: 12px;
  cursor: pointer;
}
.footer-logout:hover {
  border-color: var(--accent-pink, #ff66b2);
  color: var(--accent-pink, #ff66b2);
}
</style>

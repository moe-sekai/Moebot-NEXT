<template>
  <n-config-provider>
    <n-loading-bar-provider>
      <n-notification-provider>
        <n-dialog-provider>
          <n-message-provider>
            <div class="app-shell" :class="{ 'app-shell--collapsed': collapsed }">
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
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { NConfigProvider, NMessageProvider, NDialogProvider, NNotificationProvider, NLoadingBarProvider } from 'naive-ui'
import ConsoleNav from './components/layout/ConsoleNav.vue'
import MoebotLogo from './components/MoebotLogo.vue'
import { getPageDescription } from './navigation'

const route = useRoute()
const page = computed(() => getPageDescription(route.name))

// 侧栏折叠状态：持久化到 localStorage，刷新后保持。
const SIDEBAR_KEY = 'moebot.sidebar.collapsed'
const collapsed = ref(typeof localStorage !== 'undefined' && localStorage.getItem(SIDEBAR_KEY) === '1')
function toggleSidebar() { collapsed.value = !collapsed.value }
watch(collapsed, value => {
  try { localStorage.setItem(SIDEBAR_KEY, value ? '1' : '0') } catch {}
})
</script>

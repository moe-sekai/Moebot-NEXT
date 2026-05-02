<template>
  <n-config-provider>
    <n-message-provider>
      <div class="app-shell">
        <aside class="app-sidebar" aria-label="Moebot NEXT 控制台侧边栏">
          <RouterLink to="/" class="brand" aria-label="返回概览">
            <MoebotLogo color="var(--accent-pink)" :height="40" />
            <small>Go Console</small>
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
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import ConsoleNav from './components/layout/ConsoleNav.vue'
import MoebotLogo from './components/MoebotLogo.vue'
import { getPageDescription } from './navigation'

const route = useRoute()
const page = computed(() => getPageDescription(route.name))
</script>

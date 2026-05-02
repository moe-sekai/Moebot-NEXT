<template>
  <n-config-provider>
    <n-message-provider>
      <n-layout class="app-shell" has-sider>
        <n-layout-sider class="app-sider" bordered :width="248" collapse-mode="width" :collapsed-width="72" show-trigger="bar">
          <div class="brand"><span class="brand-mark">M</span><span>Moebot NEXT</span></div>
          <n-menu :options="menuOptions" :value="route.name as string" @update:value="go" />
        </n-layout-sider>
        <n-layout class="main-layout">
          <n-layout-header bordered class="header">
            <div>
              <div class="header__title">Moebot NEXT</div>
              <div class="header__subtitle">8080 管理首页 · 6700 OneBot 反向 WS · 3001 Satori Renderer</div>
            </div>
          </n-layout-header>
          <n-layout-content class="content">
            <router-view />
          </n-layout-content>
        </n-layout>
      </n-layout>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { h } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import type { MenuOption } from 'naive-ui'

const router = useRouter()
const route = useRoute()

const menuOptions: MenuOption[] = [
  { label: () => h(RouterLink, { to: '/' }, { default: () => '仪表盘' }), key: 'dashboard' },
  { label: () => h(RouterLink, { to: '/groups' }, { default: () => '群组管理' }), key: 'groups' },
  { label: () => h(RouterLink, { to: '/users' }, { default: () => '用户管理' }), key: 'users' },
  { label: () => h(RouterLink, { to: '/stats' }, { default: () => '指令统计' }), key: 'stats' },
]

function go(key: string) {
  const routes: Record<string, string> = {
    dashboard: '/',
    groups: '/groups',
    users: '/users',
    stats: '/stats',
  }
  router.push(routes[key] ?? '/')
}
</script>

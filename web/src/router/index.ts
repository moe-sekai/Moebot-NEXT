import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import Dashboard from '../views/Dashboard.vue'
import Login from '../views/Login.vue'
import Setup from '../views/Setup.vue'
import Groups from '../views/Groups.vue'
import Users from '../views/Users.vue'
import Settings from '../views/Settings.vue'
import CommandParser from '../views/CommandParser.vue'
import About from '../views/About.vue'
import Status from '../views/Status.vue'
import Filter from '../views/Filter.vue'
import Logs from '../views/Logs.vue'
import Plugins from '../views/Plugins.vue'
import PluginMarket from '../views/PluginMarket.vue'
import MoesekaiSettings from '../views/MoesekaiSettings.vue'
import AutochatOverview from '../views/AutochatOverview.vue'
import AutochatSettings from '../views/AutochatSettings.vue'
import AutochatMemory from '../views/AutochatMemory.vue'
import GalleryOverview from '../views/GalleryOverview.vue'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    if (to.hash) {
      return { el: to.hash, top: 80, behavior: 'smooth' }
    }
    return { left: 0, top: 0 }
  },
  routes: [
    // 鉴权 / 首启
    { path: '/login', name: 'login', component: Login, meta: { public: true } },
    { path: '/setup', name: 'setup', component: Setup, meta: { public: true } },

    // 核心控制台
    { path: '/', name: 'dashboard', component: Dashboard },
    { path: '/status', name: 'status', component: Status },
    { path: '/settings', name: 'settings', component: () => import('../views/CoreSettings.vue') },
    { path: '/filter', name: 'filter', component: Filter },
    { path: '/groups', name: 'groups', component: Groups },
    { path: '/users', name: 'users', component: Users },
    { path: '/logs', name: 'logs', component: Logs },
    { path: '/about', name: 'about', component: About },

    // 插件框架
    { path: '/plugins', name: 'plugins', component: Plugins },
    { path: '/plugins/market', name: 'plugins-market', component: PluginMarket },

    // MoeSekai 插件命名空间
    { path: '/plugins/moesekai', name: 'plugins-moesekai', component: MoesekaiSettings },
    { path: '/plugins/moesekai/advanced', name: 'plugins-moesekai-advanced', component: Settings },
    { path: '/plugins/moesekai/commands', name: 'plugins-moesekai-commands', component: CommandParser },

    // AutoChat 插件命名空间
    { path: '/plugins/autochat', name: 'plugins-autochat', component: AutochatOverview },
    { path: '/plugins/autochat/settings', name: 'plugins-autochat-settings', component: AutochatSettings },
    { path: '/plugins/autochat/memory', name: 'plugins-autochat-memory', component: AutochatMemory },

    // Gallery 插件命名空间
    { path: '/plugins/gallery', name: 'plugins-gallery', component: GalleryOverview },

    // 旧路径重定向（保持外链兼容）
    { path: '/bot', redirect: '/settings' },
    { path: '/commands', redirect: '/plugins/moesekai/commands' },
    { path: '/masterdata', redirect: '/plugins/moesekai' },
    { path: '/stats', redirect: '/plugins/moesekai' },
    { path: '/plugins/moesekai/masterdata', redirect: '/plugins/moesekai' },
    { path: '/plugins/moesekai/stats', redirect: '/plugins/moesekai' },

    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

// 全局守卫：
// 1) 拉取一次 /api/auth/status，根据 initialized 引导首启或登录。
// 2) 已登录访问 /login 或 /setup 自动重定向回 /。
//
// 注意：失败时（例如后端宕机）我们 silently 放行，让页面自身的 axios 拦截器
// 兜底处理 401 / 网络错误。
router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (auth.initialized === null) {
    try {
      await auth.refreshStatus()
    } catch {
      // 后端不可达；放行，让用户至少能看到错误页
      return true
    }
  }
  const isPublic = Boolean(to.meta?.public)
  // 未初始化：所有路径都引导到 /setup（除 /setup 本身）。
  if (auth.initialized === false) {
    if (to.path !== '/setup') return { path: '/setup' }
    return true
  }
  // 已初始化：/setup 已不再可用，未登录访问应跳到 /login，已登录则回 /。
  if (auth.initialized === true && to.path === '/setup') {
    return auth.isLoggedIn ? { path: '/' } : { path: '/login' }
  }
  if (!auth.isLoggedIn) {
    if (isPublic) return true
    return { path: '/login', query: to.fullPath !== '/' ? { redirect: to.fullPath } : undefined }
  }
  if (auth.isLoggedIn && to.path === '/login') {
    return { path: '/' }
  }
  return true
})

export default router

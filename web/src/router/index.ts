import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
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
import AutochatSettings from '../views/AutochatSettings.vue'

export default createRouter({
  history: createWebHistory(),
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    if (to.hash) {
      return { el: to.hash, top: 80, behavior: 'smooth' }
    }
    return { left: 0, top: 0 }
  },
  routes: [
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
    { path: '/plugins/autochat', name: 'plugins-autochat', component: AutochatSettings },

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

import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Groups from '../views/Groups.vue'
import Users from '../views/Users.vue'
import Stats from '../views/Stats.vue'
import Settings from '../views/Settings.vue'
import CommandParser from '../views/CommandParser.vue'
import About from '../views/About.vue'
import Status from '../views/Status.vue'
import Masterdata from '../views/Masterdata.vue'
import Bot from '../views/Bot.vue'
import Filter from '../views/Filter.vue'
import Logs from '../views/Logs.vue'
import Plugins from '../views/Plugins.vue'
import PluginMarket from '../views/PluginMarket.vue'
import MoesekaiSettings from '../views/MoesekaiSettings.vue'

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
    { path: '/', name: 'dashboard', component: Dashboard },
    { path: '/status', name: 'status', component: Status },
    { path: '/commands', name: 'commands', component: CommandParser },
    { path: '/masterdata', name: 'masterdata', component: Masterdata },
    { path: '/settings', name: 'settings', component: Settings },
    { path: '/bot', name: 'bot', component: Bot },
    { path: '/filter', name: 'filter', component: Filter },
    { path: '/groups', name: 'groups', component: Groups },
    { path: '/users', name: 'users', component: Users },
    { path: '/stats', name: 'stats', component: Stats },
    { path: '/logs', name: 'logs', component: Logs },
    { path: '/plugins', name: 'plugins', component: Plugins },
    { path: '/plugins/market', name: 'plugins-market', component: PluginMarket },
    { path: '/plugins/moesekai', name: 'plugins-moesekai', component: MoesekaiSettings },
    { path: '/about', name: 'about', component: About },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

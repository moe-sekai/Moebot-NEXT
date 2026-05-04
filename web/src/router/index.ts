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
import Logs from '../views/Logs.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: Dashboard },
    { path: '/status', name: 'status', component: Status },
    { path: '/commands', name: 'commands', component: CommandParser },
    { path: '/masterdata', name: 'masterdata', component: Masterdata },
    { path: '/settings', name: 'settings', component: Settings },
    { path: '/bot', name: 'bot', component: Bot },
    { path: '/groups', name: 'groups', component: Groups },
    { path: '/users', name: 'users', component: Users },
    { path: '/stats', name: 'stats', component: Stats },
    { path: '/logs', name: 'logs', component: Logs },
    { path: '/about', name: 'about', component: About },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

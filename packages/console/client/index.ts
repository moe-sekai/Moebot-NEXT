import { Context } from '@koishijs/client'
import Dashboard from './pages/Dashboard.vue'
import SekaiApi from './pages/SekaiApi.vue'
import Groups from './pages/Groups.vue'
import Users from './pages/Users.vue'
import CommandStats from './pages/CommandStats.vue'

export default (ctx: Context) => {
  ctx.page({
    name: 'Moebot 总览',
    path: '/moebot',
    icon: 'activity:default',
    order: 1000,
    component: Dashboard,
  })

  ctx.page({
    name: 'SEKAI API',
    path: '/moebot/sekai-api',
    icon: 'activity:default',
    order: 999,
    component: SekaiApi,
  })

  ctx.page({
    name: '群组管理',
    path: '/moebot/groups',
    icon: 'activity:default',
    order: 998,
    component: Groups,
  })

  ctx.page({
    name: '用户管理',
    path: '/moebot/users',
    icon: 'activity:default',
    order: 997,
    component: Users,
  })

  ctx.page({
    name: '指令统计',
    path: '/moebot/stats',
    icon: 'activity:default',
    order: 996,
    component: CommandStats,
  })
}

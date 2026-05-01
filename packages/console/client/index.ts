import { Context } from '@koishijs/client'
import Dashboard from './pages/Dashboard.vue'
import SekaiApi from './pages/SekaiApi.vue'
import Groups from './pages/Groups.vue'
import Users from './pages/Users.vue'
import CommandStats from './pages/CommandStats.vue'
import RendererPreview from './pages/RendererPreview.vue'

export default (ctx: Context) => {
  ctx.page({
    id: 'moebot-dashboard',
    name: 'Moebot 总览',
    path: '/moebot',
    icon: 'activity:default',
    order: 1000,
    component: Dashboard,
  })

  ctx.page({
    id: 'moebot-sekai-api',
    name: 'SEKAI API',
    path: '/moebot/sekai-api',
    icon: 'activity:default',
    order: 999,
    component: SekaiApi,
  })

  ctx.page({
    id: 'moebot-groups',
    name: '群组管理',
    path: '/moebot/groups',
    icon: 'activity:default',
    order: 998,
    component: Groups,
  })

  ctx.page({
    id: 'moebot-users',
    name: '用户管理',
    path: '/moebot/users',
    icon: 'activity:default',
    order: 997,
    component: Users,
  })

  ctx.page({
    id: 'moebot-renderer-preview',
    name: '渲染预览',
    path: '/moebot/renderer',
    icon: 'activity:default',
    order: 996,
    component: RendererPreview,
  })

  ctx.page({
    id: 'moebot-stats',
    name: '指令统计',
    path: '/moebot/stats',
    icon: 'activity:default',
    order: 995,
    component: CommandStats,
  })
}

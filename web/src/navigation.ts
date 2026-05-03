import type { IconName } from './components/icons/SvgIcon.vue'

export type ConsoleIconName = IconName

export interface ConsoleNavItem {
  path: string
  name: string
  label: string
  subtitle: string
  icon: ConsoleIconName
  section: 'main' | 'manage' | 'system'
}

export const consoleNavItems: ConsoleNavItem[] = [
  { path: '/', name: 'dashboard', label: '概览', subtitle: 'Dashboard', icon: 'dashboard', section: 'main' },
  { path: '/status', name: 'status', label: '状态', subtitle: 'Runtime', icon: 'status', section: 'main' },
  { path: '/preview', name: 'preview', label: '渲染预览', subtitle: 'Renderer', icon: 'preview', section: 'main' },
  { path: '/masterdata', name: 'masterdata', label: 'Masterdata', subtitle: 'Search', icon: 'masterdata', section: 'main' },
  { path: '/settings', name: 'settings', label: '设置', subtitle: 'Config', icon: 'settings', section: 'main' },
  { path: '/bot', name: 'bot', label: 'Bot', subtitle: 'OneBot', icon: 'bot', section: 'manage' },
  { path: '/groups', name: 'groups', label: '群组', subtitle: 'Groups', icon: 'groups', section: 'manage' },
  { path: '/users', name: 'users', label: '用户', subtitle: 'Users', icon: 'users', section: 'manage' },
  { path: '/stats', name: 'stats', label: '统计', subtitle: 'Stats', icon: 'stats', section: 'manage' },
  { path: '/logs', name: 'logs', label: '日志', subtitle: 'Logs', icon: 'logs', section: 'system' },
  { path: '/about', name: 'about', label: '关于', subtitle: 'About', icon: 'about', section: 'system' },
]

export const navSectionLabels: Record<ConsoleNavItem['section'], string> = {
  main: '控制台',
  manage: '管理',
  system: '系统',
}

export const pageDescriptions: Record<string, { title: string; subtitle: string; eyebrow: string }> = {
  dashboard: {
    title: '概览',
    subtitle: '服务状态、版本与最近活动一屏掌握。',
    eyebrow: 'Dashboard',
  },
  status: {
    title: '运行状态',
    subtitle: '检查 Bot、Web、Renderer、Masterdata 与数据库链路。',
    eyebrow: 'Runtime Status',
  },
  preview: {
    title: '渲染预览',
    subtitle: '查看 Satori 模板输出、渲染时间与各步骤耗时。',
    eyebrow: 'Renderer Preview',
  },
  masterdata: {
    title: 'Masterdata',
    subtitle: '确认基础数据加载情况，并执行只读搜索测试。',
    eyebrow: 'Data Console',
  },
  settings: {
    title: '设置',
    subtitle: '按区服、数据源、资源源与接口功能管理配置。',
    eyebrow: 'Settings',
  },
  bot: {
    title: 'Bot',
    subtitle: '查看 OneBot 驱动、命令前缀、昵称与连接说明。',
    eyebrow: 'OneBot',
  },
  groups: {
    title: '群组',
    subtitle: '查看 BOT 群配置与启用状态。',
    eyebrow: 'Groups',
  },
  users: {
    title: '用户',
    subtitle: '查看平台用户与 PJSK 游戏账号绑定关系。',
    eyebrow: 'Users',
  },
  stats: {
    title: '指令统计',
    subtitle: '查看最近指令调用数量与平均响应时间。',
    eyebrow: 'Command Stats',
  },
  logs: {
    title: '日志',
    subtitle: '当前以最近命令记录作为轻量日志视图。',
    eyebrow: 'Logs',
  },
  about: {
    title: '关于',
    subtitle: '了解 Moebot NEXT-Go、运行时与控制台依赖。',
    eyebrow: 'About',
  },
}

export function getPageDescription(routeName: unknown) {
  return pageDescriptions[String(routeName || 'dashboard')] ?? pageDescriptions.dashboard
}

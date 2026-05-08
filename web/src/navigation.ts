import type { IconName } from './components/icons/SvgIcon.vue'

export type ConsoleIconName = IconName

export interface ConsoleNavItem {
  path: string
  name: string
  label: string
  subtitle: string
  icon: ConsoleIconName
  section: 'main' | 'plugins' | 'moesekai' | 'manage' | 'system'
  // 若设置，则仅当对应插件 loaded=true 时该项才显示。
  requiresPlugin?: string
}

export const consoleNavItems: ConsoleNavItem[] = [
  // 控制台核心
  { path: '/', name: 'dashboard', label: '概览', subtitle: 'Dashboard', icon: 'dashboard', section: 'main' },
  { path: '/status', name: 'status', label: '状态', subtitle: 'Runtime', icon: 'status', section: 'main' },
  { path: '/settings', name: 'settings', label: '核心设置', subtitle: 'Core Config', icon: 'settings', section: 'main' },
  { path: '/bot', name: 'bot', label: 'Bot', subtitle: 'OneBot', icon: 'bot', section: 'main' },
  { path: '/filter', name: 'filter', label: 'Filter', subtitle: 'Gateway', icon: 'filter', section: 'main' },

  // 插件框架
  { path: '/plugins', name: 'plugins', label: '插件管理', subtitle: 'Plugins', icon: 'plugin', section: 'plugins' },
  { path: '/plugins/market', name: 'plugins-market', label: '插件市场', subtitle: 'Marketplace', icon: 'market', section: 'plugins' },

  // MoeSekai 插件命名空间（仅当 moesekai 启用时显示；段标题已表明归属，无需重复前缀）
  { path: '/plugins/moesekai', name: 'plugins-moesekai', label: '概览', subtitle: 'Overview', icon: 'sparkle', section: 'moesekai', requiresPlugin: 'moesekai' },
  { path: '/plugins/moesekai/advanced', name: 'plugins-moesekai-advanced', label: '高级配置', subtitle: 'Region/API/Assets', icon: 'settings', section: 'moesekai', requiresPlugin: 'moesekai' },
  { path: '/plugins/moesekai/commands', name: 'plugins-moesekai-commands', label: '指令解析', subtitle: 'Parser', icon: 'command', section: 'moesekai', requiresPlugin: 'moesekai' },
  { path: '/plugins/moesekai/masterdata', name: 'plugins-moesekai-masterdata', label: 'Masterdata', subtitle: 'Search', icon: 'masterdata', section: 'moesekai', requiresPlugin: 'moesekai' },
  { path: '/plugins/moesekai/stats', name: 'plugins-moesekai-stats', label: '指令统计', subtitle: 'Stats', icon: 'stats', section: 'moesekai', requiresPlugin: 'moesekai' },

  // 通用管理
  { path: '/groups', name: 'groups', label: '群组', subtitle: 'Groups', icon: 'groups', section: 'manage' },
  { path: '/users', name: 'users', label: '用户', subtitle: 'Users', icon: 'users', section: 'manage' },

  // 系统
  { path: '/logs', name: 'logs', label: '日志', subtitle: 'Logs', icon: 'logs', section: 'system' },
  { path: '/about', name: 'about', label: '关于', subtitle: 'About', icon: 'about', section: 'system' },
]

export const navSectionLabels: Record<ConsoleNavItem['section'], string> = {
  main: '控制台',
  plugins: '插件框架',
  moesekai: 'MoeSekai 插件',
  manage: '通用管理',
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
  commands: {
    title: '指令解析',
    subtitle: '测试聊天指令解析、别名触发与 Satori 渲染预览。',
    eyebrow: 'Command Parser',
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
  filter: {
    title: 'Filter 网关',
    subtitle: '管理 OneBot 反向 WS 网关、下游 Bot 应用与过滤规则。',
    eyebrow: 'OneBot Gateway',
  },
  plugins: {
    title: '插件',
    subtitle: '管理已编译进当前进程的插件，启用/禁用并跳转到各插件设置。',
    eyebrow: 'Plugins',
  },
  'plugins-market': {
    title: '插件市场',
    subtitle: '浏览官方与第三方（FloatTech ZeroBot-Plugin 上游）插件清单。',
    eyebrow: 'Marketplace',
  },
  'plugins-moesekai': {
    title: 'MoeSekai · 概览',
    subtitle: '管理 Project Sekai 业务相关的 masterdata、资源、API 与多区服配置。',
    eyebrow: 'Plugin · MoeSekai',
  },
  'plugins-moesekai-advanced': {
    title: 'MoeSekai · 高级配置',
    subtitle: '区服 / Masterdata / Assets / Sekai-Suite-Ranking API 详细参数。',
    eyebrow: 'Plugin · MoeSekai',
  },
  'plugins-moesekai-commands': {
    title: 'MoeSekai · 指令解析',
    subtitle: '测试 Project Sekai 业务指令的解析、别名触发与渲染预览。',
    eyebrow: 'Plugin · MoeSekai',
  },
  'plugins-moesekai-masterdata': {
    title: 'MoeSekai · Masterdata',
    subtitle: '查看 PJSK Masterdata 加载情况并执行只读搜索测试。',
    eyebrow: 'Plugin · MoeSekai',
  },
  'plugins-moesekai-stats': {
    title: 'MoeSekai · 指令统计',
    subtitle: '查看 PJSK 指令调用数量与平均响应时间。',
    eyebrow: 'Plugin · MoeSekai',
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
    subtitle: '查看运行时日志缓冲并按等级或关键字过滤。',
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

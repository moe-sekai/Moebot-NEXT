import type { IconName } from './components/icons/SvgIcon.vue'

export type ConsoleIconName = IconName

export interface ConsoleNavItem {
  path: string
  name: string
  label: string
  subtitle: string
  icon: ConsoleIconName
  section: 'main' | 'plugins' | 'moesekai' | 'autochat' | 'gallery' | 'manage' | 'system'
  // 若设置，则仅当对应插件 loaded=true 时该项才显示。
  requiresPlugin?: string
}

export const consoleNavItems: ConsoleNavItem[] = [
  // 控制台核心
  { path: '/', name: 'dashboard', label: '概览', subtitle: 'Dashboard', icon: 'dashboard', section: 'main' },
  { path: '/status', name: 'status', label: '状态', subtitle: 'Runtime', icon: 'status', section: 'main' },
  { path: '/settings', name: 'settings', label: '核心设置', subtitle: 'Core & Bot', icon: 'settings', section: 'main' },
  { path: '/filter', name: 'filter', label: 'Filter', subtitle: 'Gateway', icon: 'filter', section: 'main' },

  // 插件框架
  { path: '/plugins', name: 'plugins', label: '插件管理', subtitle: 'Plugins', icon: 'plugin', section: 'plugins' },
  { path: '/plugins/market', name: 'plugins-market', label: '插件市场', subtitle: 'Marketplace', icon: 'market', section: 'plugins' },

  // MoeSekai 插件命名空间（仅当 moesekai 启用时显示；段标题已表明归属，无需重复前缀）
  { path: '/plugins/moesekai', name: 'plugins-moesekai', label: '概览', subtitle: 'Overview', icon: 'sparkle', section: 'moesekai', requiresPlugin: 'moesekai' },
  { path: '/plugins/moesekai/advanced', name: 'plugins-moesekai-advanced', label: '高级配置', subtitle: 'Region/API/Assets', icon: 'settings', section: 'moesekai', requiresPlugin: 'moesekai' },
  { path: '/plugins/moesekai/commands', name: 'plugins-moesekai-commands', label: '指令解析', subtitle: 'Parser', icon: 'command', section: 'moesekai', requiresPlugin: 'moesekai' },

  // AutoChat 插件命名空间（仅当 autochat 启用时显示）
  { path: '/plugins/autochat', name: 'plugins-autochat', label: '概览', subtitle: 'Overview & Providers', icon: 'sparkle', section: 'autochat', requiresPlugin: 'autochat' },
  { path: '/plugins/autochat/settings', name: 'plugins-autochat-settings', label: '设置', subtitle: 'Persona / Triggers / Groups', icon: 'settings', section: 'autochat', requiresPlugin: 'autochat' },
  { path: '/plugins/autochat/memory', name: 'plugins-autochat-memory', label: '记忆管理', subtitle: 'Memory', icon: 'logs', section: 'autochat', requiresPlugin: 'autochat' },

  // Gallery 插件命名空间
  { path: '/plugins/gallery', name: 'plugins-gallery', label: '画廊管理', subtitle: 'Galleries & Pics', icon: 'gallery', section: 'gallery', requiresPlugin: 'gallery' },

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
  autochat: 'AutoChat 插件',
  gallery: 'Gallery 插件',
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
    subtitle: '检查 Bot、Web、Renderer 与数据库链路。',
    eyebrow: 'Runtime Status',
  },
  commands: {
    title: '指令解析',
    subtitle: '测试聊天指令解析、别名触发与 Satori 渲染预览。',
    eyebrow: 'Command Parser',
  },
  settings: {
    title: '核心设置',
    subtitle: '框架核心运行时配置（含 Bot 连接状态）与已加载插件管理。',
    eyebrow: 'Core & Bot',
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
  'plugins-autochat': {
    title: 'AutoChat · 概览',
    subtitle: '查看插件状态、Token 用量并集中配置 LLM / 视觉 / 嵌入 / 重排 / 向量 等提供商。',
    eyebrow: 'Plugin · AutoChat',
  },
  'plugins-autochat-settings': {
    title: 'AutoChat · 设置',
    subtitle: '人设 / 触发与阈值 / 单群覆盖 / YAML 高级编辑。',
    eyebrow: 'Plugin · AutoChat',
  },
  'plugins-autochat-memory': {
    title: 'AutoChat · 记忆管理',
    subtitle: '检索向量库中的用户画像与对话总结，支持语义搜索与单条删除。',
    eyebrow: 'Plugin · AutoChat',
  },
  'plugins-gallery': {
    title: 'Gallery · 画廊管理',
    subtitle: '创建/管理图片画廊，浏览缩略图、上传与删除图片。',
    eyebrow: 'Plugin · Gallery',
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
  logs: {
    title: '日志',
    subtitle: '查看运行时日志缓冲并按等级或关键字过滤。',
    eyebrow: 'Logs',
  },
  about: {
    title: '关于',
    subtitle: '了解 Moebot NEXT、运行时与控制台依赖。',
    eyebrow: 'About',
  },
}

export function getPageDescription(routeName: unknown) {
  return pageDescriptions[String(routeName || 'dashboard')] ?? pageDescriptions.dashboard
}

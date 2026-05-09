<template>
  <nav class="console-nav" aria-label="控制台导航">
    <div v-for="section in sections" :key="section.key" class="console-nav__section">
      <div class="console-nav__section-title">{{ section.label }}</div>
      <RouterLink
        v-for="item in section.items"
        :key="item.name"
        :to="item.path"
        class="console-nav__item"
        :class="{ 'console-nav__item--active': route.name === item.name }"
      >
        <span class="console-nav__icon"><SvgIcon :name="item.icon" :size="18" /></span>
        <span class="console-nav__text">
          <span class="console-nav__label">{{ item.label }}</span>
          <span class="console-nav__subtitle">{{ item.subtitle }}</span>
        </span>
      </RouterLink>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { consoleNavItems, navSectionLabels, type ConsoleNavItem } from '../../navigation'
import { usePluginsStore } from '../../stores/plugins'
import SvgIcon from '../icons/SvgIcon.vue'

const route = useRoute()
const plugins = usePluginsStore()

onMounted(() => plugins.fetch())

function isVisible(item: ConsoleNavItem) {
  // 在 plugins store 完成首次加载前默认显示，避免侧栏闪烁；
  // 加载完成后按 enabled 状态过滤（即使尚未重启 loaded=false，
  // 用户也应能进入子页查看 / 配置该插件）。
  if (!item.requiresPlugin) return true
  if (!plugins.loaded) return true
  return plugins.isEnabled(item.requiresPlugin)
}

const sections = computed(() => {
  return (['main', 'plugins', 'moesekai', 'autochat', 'manage', 'system'] as const)
    .map(section => ({
      key: section,
      label: navSectionLabels[section],
      items: consoleNavItems.filter(item => item.section === section && isVisible(item)),
    }))
    .filter(s => s.items.length > 0)
})
</script>

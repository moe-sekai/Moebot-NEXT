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
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { consoleNavItems, navSectionLabels } from '../../navigation'
import SvgIcon from '../icons/SvgIcon.vue'

const route = useRoute()

const sections = computed(() => {
  return (['main', 'manage', 'system'] as const).map(section => ({
    key: section,
    label: navSectionLabels[section],
    items: consoleNavItems.filter(item => item.section === section),
  }))
})
</script>

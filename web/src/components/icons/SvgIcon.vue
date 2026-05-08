<template>
  <svg
    class="svg-icon"
    :class="className"
    :width="size"
    :height="size"
    :viewBox="definition.viewBox"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true"
  >
    <path
      v-for="(path, index) in definition.paths"
      :key="`path-${index}`"
      :d="path"
      stroke="currentColor"
      stroke-width="1.9"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
    <circle
      v-for="(circle, index) in definition.circles"
      :key="`circle-${index}`"
      :cx="circle.cx"
      :cy="circle.cy"
      :r="circle.r"
      stroke="currentColor"
      stroke-width="1.9"
    />
    <rect
      v-for="(rect, index) in definition.rects"
      :key="`rect-${index}`"
      :x="rect.x"
      :y="rect.y"
      :width="rect.width"
      :height="rect.height"
      :rx="rect.rx"
      stroke="currentColor"
      stroke-width="1.9"
    />
  </svg>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type CircleDef = { cx: number; cy: number; r: number }
type RectDef = { x: number; y: number; width: number; height: number; rx: number }
type IconDef = { viewBox: string; paths: string[]; circles?: CircleDef[]; rects?: RectDef[] }

const icons = {
  dashboard: {
    viewBox: '0 0 24 24',
    paths: ['M4 13h6v7H4z', 'M14 4h6v16h-6z', 'M4 4h6v5H4z'],
  },
  status: {
    viewBox: '0 0 24 24',
    paths: ['M4 13.5 8.5 9l4 4L20 5.5', 'M5 20h14'],
  },
  settings: {
    viewBox: '0 0 24 24',
    paths: ['M12 8.5a3.5 3.5 0 1 0 0 7 3.5 3.5 0 0 0 0-7Z', 'M19.4 15a1.7 1.7 0 0 0 .34 1.87l.05.05a2.05 2.05 0 0 1-2.9 2.9l-.05-.05A1.7 1.7 0 0 0 15 19.4a1.7 1.7 0 0 0-1 .6 1.7 1.7 0 0 0-.4 1.1V21a2.05 2.05 0 0 1-4.1 0v-.08A1.7 1.7 0 0 0 8.4 19.4a1.7 1.7 0 0 0-1.87.34l-.05.05a2.05 2.05 0 1 1-2.9-2.9l.05-.05A1.7 1.7 0 0 0 4 15a1.7 1.7 0 0 0-.6-1 1.7 1.7 0 0 0-1.1-.4H2a2.05 2.05 0 1 1 0-4.1h.08A1.7 1.7 0 0 0 3.6 8.4 1.7 1.7 0 0 0 3.26 6.53l-.05-.05a2.05 2.05 0 1 1 2.9-2.9l.05.05A1.7 1.7 0 0 0 8 4a1.7 1.7 0 0 0 1-.6 1.7 1.7 0 0 0 .4-1.1V2a2.05 2.05 0 1 1 4.1 0v.08A1.7 1.7 0 0 0 15 4a1.7 1.7 0 0 0 1.87-.34l.05-.05a2.05 2.05 0 1 1 2.9 2.9l-.05.05A1.7 1.7 0 0 0 19.4 8c.22.37.58.62 1 .7.36.06.72.08 1.1.08H22a2.05 2.05 0 1 1 0 4.1h-.08A1.7 1.7 0 0 0 19.4 15Z'],
  },
  preview: {
    viewBox: '0 0 24 24',
    paths: ['M4 6.5A2.5 2.5 0 0 1 6.5 4h11A2.5 2.5 0 0 1 20 6.5v11a2.5 2.5 0 0 1-2.5 2.5h-11A2.5 2.5 0 0 1 4 17.5z', 'M7 16l3.2-3.2a1.2 1.2 0 0 1 1.7 0l1.1 1.1 1.9-1.9a1.2 1.2 0 0 1 1.7 0L20 15.4', 'M8.5 8.5h.01'],
  },
  about: {
    viewBox: '0 0 24 24',
    paths: ['M12 11v6', 'M12 7h.01'],
    circles: [{ cx: 12, cy: 12, r: 9 }],
  },
  masterdata: {
    viewBox: '0 0 24 24',
    paths: ['M5 5.5A2.5 2.5 0 0 1 7.5 3H19v16H7.5A2.5 2.5 0 0 0 5 21z', 'M5 5.5v13', 'M9 7h6', 'M9 11h7'],
  },
  bot: {
    viewBox: '0 0 24 24',
    paths: ['M12 4V2', 'M7 4h10a3 3 0 0 1 3 3v8a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4V7a3 3 0 0 1 3-3Z', 'M9 10h.01', 'M15 10h.01', 'M9 15h6'],
  },
  logs: {
    viewBox: '0 0 24 24',
    paths: ['M6 4h12a2 2 0 0 1 2 2v14l-4-2-4 2-4-2-4 2V6a2 2 0 0 1 2-2Z', 'M8 8h8', 'M8 12h8', 'M8 16h5'],
  },
  groups: {
    viewBox: '0 0 24 24',
    paths: ['M9 11a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z', 'M17 10a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5Z', 'M3.5 20a5.5 5.5 0 0 1 11 0', 'M14.5 18a4.5 4.5 0 0 1 6 2'],
  },
  users: {
    viewBox: '0 0 24 24',
    paths: ['M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Z', 'M4 21a8 8 0 0 1 16 0'],
  },
  stats: {
    viewBox: '0 0 24 24',
    paths: ['M4 19V5', 'M4 19h16', 'M8 16v-5', 'M12 16V8', 'M16 16v-7'],
  },
  database: {
    viewBox: '0 0 24 24',
    paths: ['M5 6c0-1.66 3.13-3 7-3s7 1.34 7 3-3.13 3-7 3-7-1.34-7-3Z', 'M5 6v6c0 1.66 3.13 3 7 3s7-1.34 7-3V6', 'M5 12v6c0 1.66 3.13 3 7 3s7-1.34 7-3v-6'],
  },
  renderer: {
    viewBox: '0 0 24 24',
    paths: ['M4 6.5A2.5 2.5 0 0 1 6.5 4h11A2.5 2.5 0 0 1 20 6.5v7A2.5 2.5 0 0 1 17.5 16h-11A2.5 2.5 0 0 1 4 13.5z', 'M8 20h8', 'M12 16v4', 'M8.5 8.5h7', 'M8.5 11.5h4'],
  },
  web: {
    viewBox: '0 0 24 24',
    paths: ['M4 5h16v14H4z', 'M4 9h16', 'M8 5v4'],
  },
  command: {
    viewBox: '0 0 24 24',
    paths: ['M7 8l-4 4 4 4', 'M17 8l4 4-4 4', 'M14 4l-4 16'],
  },
  search: {
    viewBox: '0 0 24 24',
    paths: ['m16.5 16.5 4 4'],
    circles: [{ cx: 10.5, cy: 10.5, r: 6.5 }],
  },
  clock: {
    viewBox: '0 0 24 24',
    paths: ['M12 7v5l3 2'],
    circles: [{ cx: 12, cy: 12, r: 9 }],
  },
  check: {
    viewBox: '0 0 24 24',
    paths: ['M5 12.5 9.5 17 19 7'],
  },
  warning: {
    viewBox: '0 0 24 24',
    paths: ['M12 8v5', 'M12 17h.01', 'M10.3 4.5 2.8 18a2 2 0 0 0 1.75 3h14.9a2 2 0 0 0 1.75-3L13.7 4.5a2 2 0 0 0-3.4 0Z'],
  },
  sparkle: {
    viewBox: '0 0 24 24',
    paths: ['M12 3l1.8 5.2L19 10l-5.2 1.8L12 17l-1.8-5.2L5 10l5.2-1.8Z', 'M18 16l.8 2.2L21 19l-2.2.8L18 22l-.8-2.2L15 19l2.2-.8Z'],
  },
  resources: {
    viewBox: '0 0 24 24',
    paths: ['M4 7.5 12 3l8 4.5-8 4.5z', 'M4 12l8 4.5 8-4.5', 'M4 16.5 12 21l8-4.5'],
  },
  filter: {
    viewBox: '0 0 24 24',
    paths: ['M4 5h16l-6 8v6l-4-2v-4Z'],
  },
  plugin: {
    viewBox: '0 0 24 24',
    paths: ['M9 3v3', 'M15 3v3', 'M5 7h14a1 1 0 0 1 1 1v3a3 3 0 0 1-3 3v3a3 3 0 0 1-3 3h-4a3 3 0 0 1-3-3v-3a3 3 0 0 1-3-3V8a1 1 0 0 1 1-1Z'],
  },
  market: {
    viewBox: '0 0 24 24',
    paths: ['M3 7h18l-1.5 4H4.5z', 'M5 11v9h14v-9', 'M9 16h6'],
  },
} as const satisfies Record<string, IconDef>

export type IconName = keyof typeof icons

const props = withDefaults(
  defineProps<{
    name: IconName
    size?: number | string
    className?: string
  }>(),
  {
    size: 20,
  },
)

const definition = computed<Required<IconDef>>(() => {
  const icon = (icons[props.name] ?? icons.dashboard) as IconDef
  return {
    viewBox: icon.viewBox,
    paths: [...icon.paths],
    circles: [...(icon.circles ?? [])],
    rects: [...(icon.rects ?? [])],
  }
})
</script>

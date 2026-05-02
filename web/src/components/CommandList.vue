<template>
  <UiCard>
    <div class="card-heading">
      <div>
        <h2>常用命令</h2>
        <p>当前管理首页收录的高频 OneBot 指令</p>
      </div>
      <UiBadge variant="secondary">OneBot v11</UiBadge>
    </div>

    <div class="command-grid">
      <div v-for="command in commands" :key="command.name" class="command-card">
        <div class="command-card__name">{{ command.name }}</div>
        <div class="command-card__desc">{{ command.description }}</div>
      </div>
    </div>
  </UiCard>
</template>

<script setup lang="ts">
import UiBadge from './ui/UiBadge.vue'
import UiCard from './ui/UiCard.vue'

const commands = [
  { name: '/帮助', description: '查看 Moebot NEXT 指令菜单' },
  { name: '/查卡', description: '搜索并渲染卡牌详情' },
  { name: '/查曲', description: '查询曲目信息与谱面难度' },
  { name: '/查活动', description: '查看活动时间与加成信息' },
  { name: '/查卡池', description: '查看卡池与 Pickup 卡牌' },
  { name: '/绑定', description: '绑定 PJSK 游戏账号' },
  { name: '/解绑', description: '解除当前账号绑定' },
  { name: '/个人信息', description: '查看绑定资料与概况' },
  { name: '/生日', description: '查看角色生日提醒' },
]
</script>

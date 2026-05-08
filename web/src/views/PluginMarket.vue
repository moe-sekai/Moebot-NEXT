<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="Marketplace"
      title="插件市场"
      subtitle="浏览官方与第三方（FloatTech ZeroBot-Plugin 上游）插件清单。"
    />

    <UiAlert variant="info" title="安装方式">
      <p>
        Moebot NEXT 采用 <strong>编译期选入 + 运行期开关</strong> 模型：
        市场列出的第三方插件需在源码 <code>main.go</code> 中以
        <code>_ "github.com/.../plugin/xxx"</code> 形式引入并重新编译，之后
        在 <RouterLink to="/plugins">插件</RouterLink> 页面切换启用。
      </p>
      <p>
        二期会接入 FloatTech/zbputils/control，使该仓库的插件
        无需任何代码改动即可加载。
      </p>
    </UiAlert>

    <div class="dashboard-grid dashboard-grid--main">
      <UiCard v-for="entry in catalog" :key="entry.name">
        <div class="card-heading">
          <div>
            <h2>{{ entry.title }}</h2>
            <p>{{ entry.description }}</p>
          </div>
          <UiBadge :variant="entry.category === 'official' ? 'success' : 'secondary'">
            {{ entry.category === 'official' ? '官方' : '市场' }}
          </UiBadge>
        </div>
        <dl class="info-list">
          <div><dt>包路径</dt><dd><code>{{ entry.import_path }}</code></dd></div>
          <div v-if="entry.repo">
            <dt>仓库</dt>
            <dd><a :href="entry.repo" target="_blank" rel="noopener">{{ entry.repo }}</a></dd>
          </div>
        </dl>
      </UiCard>
    </div>
  </main>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiBadge from '../components/ui/UiBadge.vue'
import UiCard from '../components/ui/UiCard.vue'

interface MarketEntry {
  name: string
  title: string
  description: string
  category: 'official' | 'market'
  import_path: string
  repo?: string
}

// 静态目录：官方插件 + 上游 ZeroBot-Plugin 仓库精选示例。
// 二期可改为从 GitHub Releases / 自托管 manifest 拉取。
const catalog: MarketEntry[] = [
  {
    name: 'moesekai',
    title: 'MoeSekai (Project Sekai)',
    description: 'Project Sekai 全部业务：查卡 / 查曲 / 抽卡 / Suite / 组卡推荐 / 榜线 / B30 等。',
    category: 'official',
    import_path: 'moebot-next/internal/plugins/moesekai',
    repo: 'https://github.com/moe-sekai/Moebot-NEXT',
  },
  {
    name: 'chouxianghua',
    title: '抽象话生成器',
    description: '上游 ZeroBot-Plugin 示例：把普通话转为抽象话，支持双向。',
    category: 'market',
    import_path: 'github.com/FloatTech/ZeroBot-Plugin/plugin/chouxianghua',
    repo: 'https://github.com/FloatTech/ZeroBot-Plugin',
  },
  {
    name: 'fortune',
    title: '今日运势',
    description: '上游 ZeroBot-Plugin 示例：今日运势抽签。',
    category: 'market',
    import_path: 'github.com/FloatTech/ZeroBot-Plugin/plugin/fortune',
    repo: 'https://github.com/FloatTech/ZeroBot-Plugin',
  },
]
</script>

<style scoped>
code {
  background: var(--surface-soft, rgba(255, 255, 255, 0.04));
  padding: 0 4px;
  border-radius: 4px;
}
</style>

<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="Plugin · MoeSekai"
      title="MoeSekai 插件设置"
      subtitle="管理 Project Sekai 业务相关的 masterdata、资源、API 与多区服配置。"
    >
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="load">重新加载</UiButton>
        <UiButton variant="default" size="sm" :loading="saving" @click="save">保存</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="加载/保存失败">{{ error }}</UiAlert>
    <UiAlert v-if="restartNotice" variant="info" title="需要重启">
      已写入 <code>{{ path }}</code>，重启 moebot 后 PJSK 业务才会以新配置启动。
    </UiAlert>
    <UiAlert variant="info" title="说明">
      此页直接编辑插件子配置文件 <code>data/plugins/moesekai.yml</code>。
      富表单 UI 仍在迁移中（旧版位于 <RouterLink to="/settings">设置</RouterLink>），
      期间可在此处使用 YAML 直接编辑。
    </UiAlert>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>moesekai.yml</h2>
          <p>路径：<code>{{ path || '(尚未生成)' }}</code></p>
        </div>
      </div>
      <textarea v-model="yamlText" class="yaml-editor" spellcheck="false" />
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getPluginConfig, updatePluginConfig } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import UiAlert from '../components/ui/UiAlert.vue'
import UiButton from '../components/ui/UiButton.vue'
import UiCard from '../components/ui/UiCard.vue'

const yamlText = ref('')
const path = ref('')
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const restartNotice = ref(false)

onMounted(load)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await getPluginConfig('moesekai')
    yamlText.value = data.yaml
    path.value = data.path
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败。'
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  error.value = ''
  try {
    const data = await updatePluginConfig('moesekai', yamlText.value)
    path.value = data.path
    if (data.requires_restart) restartNotice.value = true
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存失败。'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.yaml-editor {
  width: 100%;
  min-height: 480px;
  font-family: 'JetBrains Mono', Consolas, ui-monospace, monospace;
  font-size: 13px;
  line-height: 1.55;
  background: var(--surface-soft, #0c0d12);
  color: var(--text-primary, #e8e8f0);
  border: 1px solid var(--border-default, rgba(255, 255, 255, 0.08));
  border-radius: 8px;
  padding: 12px 14px;
  resize: vertical;
}
code {
  background: var(--surface-soft, rgba(255, 255, 255, 0.04));
  padding: 0 4px;
  border-radius: 4px;
}
</style>

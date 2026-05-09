<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="Plugin · AutoChat"
      title="AutoChat 插件设置"
      subtitle="基于 OpenAI 兼容 / Anthropic 的群聊 LLM 对话，配合 SQLite + sqlite-vec 做记忆 RAG。"
    >
      <template #actions>
        <UiButton variant="outline" size="sm" :loading="loading" @click="load">重新加载</UiButton>
        <UiButton variant="default" size="sm" :loading="saving" @click="save">保存 YAML</UiButton>
      </template>
    </PageHeader>

    <UiAlert v-if="error" variant="destructive" title="加载/保存失败">{{ error }}</UiAlert>
    <UiAlert v-if="restartNotice" variant="info" title="已写入">
      已写入 <code>{{ path }}</code>。常用项即时生效；其他高级字段（人设字典、关键词等）需在 YAML 中编辑。
    </UiAlert>

    <UiAlert variant="info" title="说明">
      下方「常用设置」由插件通过 <code>plugin.Configurable</code> 暴露，保存后即时生效。
      首次使用请先填好 API Key，并在控制台「插件管理」中启用 autochat。
    </UiAlert>

    <PluginSettingsForm plugin-name="autochat" title="AutoChat 常用设置" />

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>autochat.yml</h2>
          <p>路径：<code>{{ path || '(尚未生成)' }}</code></p>
          <p>包含 LLM provider 列表、模型列表、Embedding/Rerank、向量库、消息阈值、群人设等高级字段。</p>
        </div>
      </div>
      <textarea v-model="yamlText" class="yaml-editor" spellcheck="false" />
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getPluginConfig, updatePluginConfig } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import PluginSettingsForm from '../components/PluginSettingsForm.vue'
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
    const data = await getPluginConfig('autochat')
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
    const data = await updatePluginConfig('autochat', yamlText.value)
    path.value = data.path
    restartNotice.value = true
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

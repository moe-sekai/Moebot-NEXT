<template>
  <div class="auth-shell">
    <div class="auth-card">
      <div class="auth-brand">
        <MoebotLogo color="var(--accent-pink)" :height="56" />
        <h1>欢迎初始化 Moebot NEXT</h1>
        <p class="auth-subtitle">检测到尚未创建管理员，请创建账号、昵称与密码</p>
      </div>

      <n-alert type="warning" :show-icon="true" style="margin-bottom: 18px">
        <strong>账号与昵称一经创建即<u>无法更改</u></strong>，请谨慎填写。
        昵称将显示在控制台底部以及 Moebot NEXT 所有渲染卡片底部
        （<code>Moebot NEXT (deployed by 昵称)</code>）。
      </n-alert>

      <n-form ref="formRef" :model="form" :rules="rules" label-placement="top" @submit.prevent="onSubmit">
        <n-form-item label="账号（登录用，仅英文/数字/下划线，3–32 位）" path="username">
          <n-input v-model:value="form.username" placeholder="例如 admin" :disabled="loading" autofocus />
        </n-form-item>
        <n-form-item label="昵称（显示用，可中文/空格，最长 32）" path="nickname">
          <n-input v-model:value="form.nickname" placeholder="例如 喵喵酱" :disabled="loading" />
        </n-form-item>
        <n-form-item label="密码（至少 8 位）" path="password">
          <n-input v-model:value="form.password" type="password" show-password-on="click" :disabled="loading" />
        </n-form-item>
        <n-form-item label="确认密码" path="password_confirm">
          <n-input v-model:value="form.password_confirm" type="password" show-password-on="click" :disabled="loading" @keyup.enter="onSubmit" />
        </n-form-item>
        <n-button type="primary" block :loading="loading" @click="onSubmit">创建管理员并进入控制台</n-button>
        <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
      </n-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, NAlert, type FormInst, type FormRules } from 'naive-ui'
import MoebotLogo from '../components/MoebotLogo.vue'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const formRef = ref<FormInst | null>(null)
const form = reactive({
  username: '',
  nickname: '',
  password: '',
  password_confirm: '',
})
const loading = ref(false)
const errorMsg = ref('')

const rules: FormRules = {
  username: [
    { required: true, message: '请输入账号', trigger: 'blur' },
    {
      validator: (_r, v) => /^[A-Za-z0-9_]{3,32}$/.test(String(v ?? '')),
      message: '账号需为 3–32 位英文/数字/下划线',
      trigger: 'blur',
    },
  ],
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    {
      validator: (_r, v) => {
        const s = String(v ?? '').trim()
        if (!s) return false
        return [...s].length <= 32
      },
      message: '昵称不能为空且不超过 32 字符',
      trigger: 'blur',
    },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, max: 128, message: '密码至少 8 位', trigger: 'blur' },
  ],
  password_confirm: [
    { required: true, message: '请再次输入密码', trigger: 'blur' },
    {
      validator: (_r, v) => v === form.password,
      message: '两次输入的密码不一致',
      trigger: ['blur', 'input'],
    },
  ],
}

async function onSubmit() {
  errorMsg.value = ''
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  loading.value = true
  try {
    await auth.setup({
      username: form.username.trim(),
      nickname: form.nickname.trim(),
      password: form.password,
      password_confirm: form.password_confirm,
    })
    await router.replace('/')
  } catch (err: any) {
    errorMsg.value = err?.response?.data?.message || err?.message || '初始化失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-shell {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: radial-gradient(ellipse at top, rgba(255, 102, 178, 0.08), transparent 60%), var(--bg-primary, #0f1115);
}
.auth-card {
  width: 100%;
  max-width: 480px;
  padding: 36px;
  border-radius: 16px;
  background: var(--surface, #181b22);
  border: 1px solid var(--border, #2a2f38);
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.35);
}
.auth-brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  margin-bottom: 18px;
  text-align: center;
}
.auth-brand h1 {
  margin: 8px 0 0;
  font-size: 20px;
  font-weight: 700;
}
.auth-subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-muted, #8a8f99);
}
.auth-error {
  margin: 14px 0 0;
  padding: 10px 12px;
  border-radius: 8px;
  background: rgba(255, 80, 110, 0.12);
  border: 1px solid rgba(255, 80, 110, 0.4);
  color: #ff6e8a;
  font-size: 13px;
}
code {
  background: rgba(255, 255, 255, 0.06);
  padding: 1px 6px;
  border-radius: 4px;
}
</style>

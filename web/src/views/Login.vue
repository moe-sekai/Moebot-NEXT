<template>
  <div class="auth-shell">
    <div class="auth-card">
      <div class="auth-brand">
        <MoebotLogo color="var(--accent-pink)" :height="56" />
        <h1>Moebot NEXT 控制台</h1>
        <p class="auth-subtitle">请输入管理员账号与密码</p>
      </div>
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="top" @submit.prevent="onSubmit">
        <n-form-item label="账号" path="username">
          <n-input v-model:value="form.username" placeholder="管理员账号" :disabled="loading" autofocus />
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input v-model:value="form.password" type="password" show-password-on="click" placeholder="密码" :disabled="loading" @keyup.enter="onSubmit" />
        </n-form-item>
        <n-button type="primary" block :loading="loading" @click="onSubmit">登录</n-button>
        <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
      </n-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, type FormInst, type FormRules } from 'naive-ui'
import MoebotLogo from '../components/MoebotLogo.vue'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const formRef = ref<FormInst | null>(null)
const form = reactive({ username: '', password: '' })
const loading = ref(false)
const errorMsg = ref('')

const rules: FormRules = {
  username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
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
    await auth.login(form.username.trim(), form.password)
    const redirect = (route.query.redirect as string | undefined) || '/'
    await router.replace(redirect)
  } catch (err: any) {
    errorMsg.value = err?.response?.data?.message || err?.message || '登录失败'
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
  max-width: 420px;
  padding: 40px 36px 32px;
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
  margin-bottom: 24px;
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
</style>

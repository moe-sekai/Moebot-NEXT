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
  background:
    radial-gradient(circle at 8% 12%, rgba(255, 120, 183, .18), transparent 38%),
    radial-gradient(circle at 92% 18%, rgba(53, 199, 212, .16), transparent 38%),
    linear-gradient(180deg, #fef7fb 0%, #f5f7ff 100%);
  color: #1f2230;
}
.auth-card {
  width: 100%;
  max-width: 440px;
  padding: 40px 36px 32px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(165, 180, 252, 0.3);
  box-shadow: 0 24px 60px rgba(142, 124, 195, 0.22);
  backdrop-filter: blur(18px);
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
  font-weight: 800;
  color: #1f2230;
  letter-spacing: -0.01em;
}
.auth-subtitle {
  margin: 0;
  font-size: 13px;
  color: #5b6270;
}
.auth-error {
  margin: 14px 0 0;
  padding: 10px 12px;
  border-radius: 10px;
  background: #fff1f5;
  border: 1px solid #fecdd3;
  color: #b91c3c;
  font-size: 13px;
}
:deep(.n-form-item .n-form-item-label) { color: #1f2230; font-weight: 600; }
</style>

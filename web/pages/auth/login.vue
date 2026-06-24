<template>
  <div class="login-page">
    <div class="login-card">
      <h1>登录 / 注册</h1>
      <p class="subtitle">使用手机号登录或注册</p>

      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" maxlength="11" />
        </el-form-item>

        <el-form-item label="验证码" prop="code">
          <div class="code-row">
            <el-input v-model="form.code" placeholder="请输入6位验证码" maxlength="6" />
            <el-button
              :disabled="countdown > 0"
              @click="sendCode"
            >
              {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleLogin" style="width: 100%">
            登录 / 注册
          </el-button>
        </el-form-item>
      </el-form>

      <div class="divider">
        <span>其他登录方式</span>
      </div>

      <el-button @click="handleWechatLogin" style="width: 100%">
        微信登录
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'

const formRef = ref<FormInstance>()
const loading = ref(false)
const countdown = ref(0)

const form = reactive({
  phone: '',
  code: '',
})

const rules: FormRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { pattern: /^\d{6}$/, message: '验证码为6位数字', trigger: 'blur' },
  ],
}

let timer: ReturnType<typeof setInterval> | null = null

function sendCode() {
  // TODO: call POST /api/v1/auth/sms-code
  countdown.value = 60
  timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0 && timer) {
      clearInterval(timer)
      timer = null
    }
  }, 1000)
}

function handleLogin() {
  // TODO: call POST /api/v1/auth/login
}

function handleWechatLogin() {
  // TODO: redirect to WeChat OAuth
}

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: var(--color-bg-base);
}
.login-card {
  width: 400px;
  padding: var(--space-xl);
  background: var(--color-bg-white);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
}
.login-card h1 {
  text-align: center;
  margin-bottom: var(--space-xs);
}
.subtitle {
  text-align: center;
  color: var(--color-text-secondary);
  margin-bottom: var(--space-lg);
}
.code-row {
  display: flex;
  gap: var(--space-sm);
}
.code-row .el-input {
  flex: 1;
}
.divider {
  display: flex;
  align-items: center;
  margin: var(--space-lg) 0;
  color: var(--color-text-secondary);
}
.divider::before,
.divider::after {
  content: '';
  flex: 1;
  border-top: 1px solid var(--color-border);
}
.divider span {
  padding: 0 var(--space-sm);
}
</style>

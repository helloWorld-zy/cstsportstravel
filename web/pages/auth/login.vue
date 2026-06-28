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
              :disabled="countdown > 0 || sendingCode"
              :loading="sendingCode"
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
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const { sendSmsCode, login } = useAuth()

const formRef = ref<FormInstance>()
const loading = ref(false)
const sendingCode = ref(false)
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

async function sendCode() {
  if (!formRef.value) return
  try {
    await formRef.value.validateField('phone')
  } catch {
    return
  }

  sendingCode.value = true
  try {
    const result = await sendSmsCode(form.phone)
    if (result.code) {
      // Dev/test mode: code returned directly
      ElMessage({ message: `验证码: ${result.code}`, type: 'success', duration: 10000 })
      form.code = result.code
    } else {
      ElMessage.success('验证码已发送')
    }
    countdown.value = 60
    timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0 && timer) {
        clearInterval(timer)
        timer = null
      }
    }, 1000)
  } catch (err: any) {
    ElMessage.error(err.message || '发送验证码失败')
  } finally {
    sendingCode.value = false
  }
}

async function handleLogin() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    const result = await login(form.phone, form.code)
    if (result.is_new_user) {
      ElMessage.success('注册成功，欢迎加入！')
    } else {
      ElMessage.success('登录成功')
    }
    navigateTo('/')
  } catch (err: any) {
    ElMessage.error(err.message || '登录失败')
  } finally {
    loading.value = false
  }
}

function handleWechatLogin() {
  // WeChat OAuth 2.0 redirect flow (FR-002)
  // The appid and redirect_uri should come from runtime config
  const appid = useRuntimeConfig().public.wechatAppId || ''
  if (!appid) {
    ElMessage.info('微信登录功能即将上线')
    return
  }

  const redirectUri = encodeURIComponent(`${window.location.origin}/auth/wechat-callback`)
  const state = Math.random().toString(36).substring(2, 15)

  // Store state for CSRF verification
  sessionStorage.setItem('wechat_oauth_state', state)

  // Redirect to WeChat OAuth authorization page
  const oauthUrl = `https://open.weixin.qq.com/connect/oauth2/authorize?appid=${appid}&redirect_uri=${redirectUri}&response_type=code&scope=snsapi_userinfo&state=${state}#wechat_redirect`
  window.location.href = oauthUrl
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

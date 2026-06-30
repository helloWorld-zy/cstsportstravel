<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2>旅行预订 - 管理后台</h2>
      <el-form :model="form" :rules="rules" ref="formRef" @submit.prevent="handleLogin">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item>
          <div class="login-options">
            <el-checkbox v-model="rememberMe">记住我</el-checkbox>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleLogin" style="width: 100%">
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const rememberMe = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const handleLogin = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/products'
    router.push(redirect)
  } catch (err: any) {
    ElMessage.error(err.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #0f172a; /* Slate 900 background for a premium dark look */
  background-image: radial-gradient(circle at 0% 0%, rgba(59, 130, 246, 0.15) 0%, transparent 50%),
                    radial-gradient(circle at 100% 100%, rgba(96, 165, 250, 0.1) 0%, transparent 50%);
  position: relative;
  overflow: hidden;
}

.login-card {
  width: 420px;
  padding: 16px 8px;
  border-radius: 20px !important;
  border: 1px solid rgba(255, 255, 255, 0.08) !important;
  background: rgba(15, 23, 42, 0.75) !important;
  backdrop-filter: blur(16px);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5) !important;
}

.login-card h2 {
  font-size: 24px;
  font-weight: 800;
  margin-top: 8px;
  margin-bottom: 32px;
  text-align: center;
  background: linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: 0.5px;
}

.login-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.login-card :deep(.el-input__wrapper) {
  background-color: rgba(30, 41, 59, 0.6) !important;
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.1) inset !important;
}

.login-card :deep(.el-input__inner) {
  color: #fff !important;
}

.login-card :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #3b82f6 inset, 0 0 0 3px rgba(59, 130, 246, 0.25) !important;
}

.login-card :deep(.el-checkbox__label) {
  color: #94a3b8 !important;
  font-weight: 500;
}

.login-card :deep(.el-checkbox__inner) {
  background-color: rgba(30, 41, 59, 0.6) !important;
  border-color: rgba(255, 255, 255, 0.1) !important;
}

.login-card :deep(.el-button) {
  height: 42px !important;
  border-radius: 8px !important;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%) !important;
  border: none !important;
  box-shadow: 0 4px 14px rgba(59, 130, 246, 0.3) !important;
  transition: all 0.2s !important;
  font-size: 15px;
}

.login-card :deep(.el-button):hover {
  opacity: 0.95;
  transform: translateY(-0.5px);
  box-shadow: 0 6px 20px rgba(59, 130, 246, 0.4) !important;
}
</style>

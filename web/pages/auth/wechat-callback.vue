<template>
  <div class="callback-page">
    <div class="loading-card">
      <el-icon class="loading-icon" :size="48"><Loading /></el-icon>
      <p>{{ statusText }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Loading } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

definePageMeta({
  layout: 'default',
})

const route = useRoute()
const router = useRouter()
const { loginWithWechat } = useAuth()

const statusText = ref('正在处理微信登录...')

onMounted(async () => {
  const code = route.query.code as string
  const state = route.query.state as string

  // Verify state parameter for CSRF protection
  const storedState = sessionStorage.getItem('wechat_oauth_state')
  sessionStorage.removeItem('wechat_oauth_state')

  if (!code) {
    ElMessage.error('微信授权失败：未收到授权码')
    router.replace('/auth/login')
    return
  }

  if (state && storedState && state !== storedState) {
    ElMessage.error('微信授权失败：状态验证不通过')
    router.replace('/auth/login')
    return
  }

  try {
    // Exchange code for login via backend
    const result = await loginWithWechat(code)
    if (result.is_new_user) {
      ElMessage.success('微信登录成功，欢迎加入！')
    } else {
      ElMessage.success('微信登录成功')
    }
    router.replace('/')
  } catch (err: any) {
    ElMessage.error(err.message || '微信登录失败')
    router.replace('/auth/login')
  }
})
</script>

<style scoped>
.callback-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
}

.loading-card {
  text-align: center;
  padding: 40px;
}

.loading-icon {
  animation: spin 1s linear infinite;
  color: var(--color-primary);
  margin-bottom: 16px;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

p {
  color: var(--color-text-secondary);
  font-size: 16px;
}
</style>

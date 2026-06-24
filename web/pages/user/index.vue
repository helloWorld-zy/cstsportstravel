<template>
  <div class="personal-center">
    <!-- User Card -->
    <div class="user-card">
      <el-avatar :size="64" :src="user?.avatar_url" />
      <div class="user-info">
        <h2>{{ user?.nickname || '未登录' }}</h2>
        <p class="user-phone">{{ user?.phone }}</p>
        <div class="user-badges">
          <el-tag :type="realNameTagType" size="small">{{ realNameLabel }}</el-tag>
          <el-tag type="info" size="small">Lv.{{ user?.member_level || 1 }}</el-tag>
        </div>
      </div>
    </div>

    <!-- Menu Groups -->
    <div class="menu-group">
      <h3 class="group-title">账号管理</h3>
      <div class="menu-list">
        <div class="menu-item" @click="navigateTo('/user/real-name')">
          <span>实名认证</span>
          <div>
            <el-tag :type="realNameTagType" size="small">{{ realNameLabel }}</el-tag>
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>
        <div class="menu-item" @click="navigateTo('/user/travellers')">
          <span>常用出游人</span>
          <el-icon><ArrowRight /></el-icon>
        </div>
      </div>
    </div>

    <div class="menu-group">
      <h3 class="group-title">订单管理</h3>
      <div class="menu-list">
        <div class="menu-item" @click="navigateTo('/user/orders')">
          <span>我的订单</span>
          <el-icon><ArrowRight /></el-icon>
        </div>
      </div>
    </div>

    <div class="menu-group">
      <h3 class="group-title">服务</h3>
      <div class="menu-list">
        <div class="menu-item" @click="handleLogout">
          <span class="logout-text">退出登录</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ArrowRight } from '@element-plus/icons-vue'

const { user, logout, init } = useAuth()

const realNameLabel = computed(() => {
  switch (user.value?.real_name_status) {
    case 'verified': return '已实名'
    case 'pending': return '审核中'
    case 'rejected': return '已驳回'
    default: return '未认证'
  }
})

const realNameTagType = computed(() => {
  switch (user.value?.real_name_status) {
    case 'verified': return 'success'
    case 'pending': return 'warning'
    case 'rejected': return 'danger'
    default: return 'info'
  }
})

onMounted(async () => {
  await init()
  if (!user.value) {
    navigateTo('/auth/login')
  }
})

function handleLogout() {
  logout()
}
</script>

<style scoped>
.personal-center {
  max-width: 600px;
  margin: 0 auto;
  padding: var(--space-lg);
}
.user-card {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding: var(--space-lg);
  background: var(--color-bg-white);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-lg);
}
.user-info h2 {
  margin: 0 0 var(--space-xs);
  font-size: 18px;
}
.user-phone {
  color: var(--color-text-secondary);
  margin: 0 0 var(--space-xs);
}
.user-badges {
  display: flex;
  gap: var(--space-xs);
}
.menu-group {
  background: var(--color-bg-white);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-md);
  overflow: hidden;
}
.group-title {
  padding: var(--space-md) var(--space-md) var(--space-xs);
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0;
}
.menu-list {
  padding: 0 var(--space-md);
}
.menu-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md) 0;
  border-bottom: 1px solid var(--color-border-light);
  cursor: pointer;
}
.menu-item:last-child {
  border-bottom: none;
}
.menu-item:hover {
  color: var(--color-primary);
}
.logout-text {
  color: var(--color-danger);
  width: 100%;
  text-align: center;
}
</style>

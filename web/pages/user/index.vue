<template>
  <div class="personal-center">
    <!-- User Card -->
    <div class="user-card">
      <el-avatar :size="64" :src="user?.avatar_url" />
      <div class="user-info">
        <h2>{{ user?.nickname || '未登录' }}</h2>
        <p class="user-phone">{{ maskedPhone }}</p>
        <div class="user-badges">
          <el-tag :type="realNameTagType" size="small">{{ realNameLabel }}</el-tag>
          <el-tag type="warning" size="small" effect="plain">
            <span class="level-icon">👑</span> Lv.{{ user?.member_level || 1 }}
          </el-tag>
        </div>
      </div>
    </div>

    <!-- Order Stats Quick View -->
    <div class="order-stats-card">
      <div class="stats-header" @click="navigateTo('/user/orders')">
        <span class="stats-title">我的订单</span>
        <span class="stats-link">全部订单 <el-icon><ArrowRight /></el-icon></span>
      </div>
      <div class="stats-grid">
        <div class="stat-item" @click="navigateTo('/user/orders?status=pending_pay')">
          <div class="stat-icon-wrapper">
            <span class="stat-icon">💳</span>
            <el-badge v-if="orderStats?.pending_pay" :value="orderStats.pending_pay" :max="99" class="stat-badge" />
          </div>
          <span class="stat-label">待付款</span>
        </div>
        <div class="stat-item" @click="navigateTo('/user/orders?status=pending_travel')">
          <div class="stat-icon-wrapper">
            <span class="stat-icon">✈️</span>
            <el-badge v-if="orderStats?.pending_travel" :value="orderStats.pending_travel" :max="99" class="stat-badge" />
          </div>
          <span class="stat-label">待出行</span>
        </div>
        <div class="stat-item" @click="navigateTo('/user/orders?status=refunding')">
          <div class="stat-icon-wrapper">
            <span class="stat-icon">💰</span>
            <el-badge v-if="orderStats?.refunding" :value="orderStats.refunding" :max="99" class="stat-badge" />
          </div>
          <span class="stat-label">退款中</span>
        </div>
        <div class="stat-item" @click="navigateTo('/user/orders?status=completed')">
          <div class="stat-icon-wrapper">
            <span class="stat-icon">✅</span>
          </div>
          <span class="stat-label">已完成</span>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="quick-actions-card">
      <h3 class="card-title">常用功能</h3>
      <div class="actions-grid">
        <div class="action-item" @click="navigateTo('/user/orders')">
          <span class="action-icon">📋</span>
          <span class="action-label">我的订单</span>
        </div>
        <div class="action-item" @click="navigateTo('/user/travellers')">
          <span class="action-icon">👥</span>
          <span class="action-label">常用出游人</span>
        </div>
        <div class="action-item" @click="navigateTo('/user/real-name')">
          <span class="action-icon">🪪</span>
          <span class="action-label">实名认证</span>
        </div>
        <div class="action-item" @click="navigateTo('/products')">
          <span class="action-icon">🔍</span>
          <span class="action-label">发现好游</span>
        </div>
      </div>
    </div>

    <!-- Menu Groups -->
    <div class="menu-group">
      <h3 class="group-title">账号管理</h3>
      <div class="menu-list">
        <div class="menu-item" @click="navigateTo('/user/real-name')">
          <div class="menu-left">
            <span class="menu-icon">🪪</span>
            <span>实名认证</span>
          </div>
          <div>
            <el-tag :type="realNameTagType" size="small">{{ realNameLabel }}</el-tag>
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>
        <div class="menu-item" @click="navigateTo('/user/travellers')">
          <div class="menu-left">
            <span class="menu-icon">👥</span>
            <span>常用出游人</span>
          </div>
          <el-icon><ArrowRight /></el-icon>
        </div>
      </div>
    </div>

    <div class="menu-group">
      <h3 class="group-title">订单管理</h3>
      <div class="menu-list">
        <div class="menu-item" @click="navigateTo('/user/orders')">
          <div class="menu-left">
            <span class="menu-icon">📋</span>
            <span>我的订单</span>
          </div>
          <div>
            <el-badge v-if="orderStats?.total" :value="orderStats.total" :max="99" />
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <div class="menu-group">
      <h3 class="group-title">服务</h3>
      <div class="menu-list">
        <div class="menu-item" @click="handleLogout">
          <div class="menu-left">
            <span class="menu-icon">🚪</span>
            <span class="logout-text">退出登录</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ArrowRight } from '@element-plus/icons-vue'
import { useApi } from '~/composables/useApi'

const { user, logout, init } = useAuth()
const api = useApi()

// Order stats
interface OrderStats {
  pending_pay: number
  paid_full: number
  pending_travel: number
  refunding: number
  completed: number
  cancelled: number
  total: number
}

const orderStats = ref<OrderStats | null>(null)

const maskedPhone = computed(() => {
  const phone = user.value?.phone
  if (!phone || phone.length < 7) return phone || ''
  return phone.slice(0, 3) + '****' + phone.slice(7)
})

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
    return
  }

  // Fetch order stats
  try {
    orderStats.value = await api.get<OrderStats>('/orders/stats')
  } catch {
    // Silently fail - stats are optional
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
  background: #f5f5f5;
  min-height: 100vh;
}

/* User Card */
.user-card {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding: var(--space-lg);
  background: linear-gradient(135deg, #ff5722 0%, #ff7043 100%);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-md);
  color: #fff;
}

.user-info h2 {
  margin: 0 0 var(--space-xs);
  font-size: 18px;
  color: #fff;
}

.user-phone {
  color: rgba(255, 255, 255, 0.8);
  margin: 0 0 var(--space-xs);
  font-size: 14px;
}

.user-badges {
  display: flex;
  gap: var(--space-xs);
}

.level-icon {
  font-size: 12px;
}

/* Order Stats Card */
.order-stats-card {
  background: var(--color-bg-white);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-md);
  overflow: hidden;
}

.stats-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md);
  cursor: pointer;
}

.stats-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.stats-link {
  font-size: 13px;
  color: #999;
  display: flex;
  align-items: center;
  gap: 4px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  padding: 0 var(--space-md) var(--space-md);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  padding: 8px 0;
}

.stat-item:hover {
  opacity: 0.8;
}

.stat-icon-wrapper {
  position: relative;
}

.stat-icon {
  font-size: 24px;
}

.stat-badge {
  position: absolute;
  top: -8px;
  right: -12px;
}

.stat-label {
  font-size: 12px;
  color: #666;
}

/* Quick Actions */
.quick-actions-card {
  background: var(--color-bg-white);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-md);
  padding: var(--space-md);
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin: 0 0 var(--space-md);
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 12px 8px;
  border-radius: 12px;
  background: #f8f8f8;
  cursor: pointer;
  transition: all 0.2s;
}

.action-item:hover {
  background: #fff3e0;
  transform: translateY(-2px);
}

.action-icon {
  font-size: 28px;
}

.action-label {
  font-size: 12px;
  color: #333;
}

/* Menu Groups */
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

.menu-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.menu-icon {
  font-size: 18px;
}

.logout-text {
  color: var(--color-danger);
}
</style>

<template>
  <div class="personal-center">
    <!-- Desktop Premium Dashboard -->
    <div class="desktop-dashboard">
      <h2 class="welcome-title">您好，{{ user?.nickname || '未登录' }}！</h2>
      
      <!-- Gold Member Card -->
      <div class="member-gold-card">
        <div class="card-left">
          <div class="gold-badge">👑 GOLD MEMBER</div>
          <div class="gold-level">山海体育旅游 VIP.{{ user?.member_level || 1 }} 会员</div>
          <div class="gold-number">会员账号：{{ maskedPhone }}</div>
        </div>
        <div class="card-right">
          <div class="real-name-status" @click="navigateTo('/user/real-name')">
            <span class="status-dot" :class="user?.real_name_status"></span>
            {{ realNameLabel }}
          </div>
        </div>
      </div>

      <!-- Order Statistics -->
      <div class="stats-row">
        <div class="stat-card-box" @click="navigateTo('/user/orders?status=pending_pay')">
          <div class="stat-num color-pending">{{ orderStats?.pending_pay || 0 }}</div>
          <div class="stat-txt">待付款</div>
        </div>
        <div class="stat-card-box" @click="navigateTo('/user/orders?status=pending_travel')">
          <div class="stat-num color-travel">{{ orderStats?.pending_travel || 0 }}</div>
          <div class="stat-txt">待出行</div>
        </div>
        <div class="stat-card-box" @click="navigateTo('/user/orders?status=refunding')">
          <div class="stat-num color-refunding">{{ orderStats?.refunding || 0 }}</div>
          <div class="stat-txt">退款中</div>
        </div>
        <div class="stat-card-box" @click="navigateTo('/user/orders')">
          <div class="stat-num color-total">{{ orderStats?.total || 0 }}</div>
          <div class="stat-txt">全部订单</div>
        </div>
      </div>

      <!-- Recent Orders List -->
      <div class="recent-orders-section">
        <div class="section-title-bar">
          <h3>最近预订</h3>
          <NuxtLink to="/user/orders" class="more-link">查看全部订单 →</NuxtLink>
        </div>

        <div v-if="recentOrders.length === 0" class="empty-recent">
          <p>您最近暂无订单。开始探索世界，开启您的下一段山海之旅吧！</p>
          <el-button type="primary" size="default" @click="navigateTo('/products')">浏览精彩行程</el-button>
        </div>

        <div v-else class="recent-list">
          <div v-for="order in recentOrders" :key="order.id" class="recent-order-item" @click="navigateTo(`/user/order-${order.id}`)">
            <el-image :src="order.cover_image" class="recent-img" fit="cover">
              <template #error>
                <div class="img-placeholder">🏔️</div>
              </template>
            </el-image>
            <div class="recent-details">
              <h4>{{ order.product_name || '旅游产品' }}</h4>
              <p class="recent-meta">
                <span>订单号：{{ order.order_no }}</span>
                <span class="meta-dot">·</span>
                <span>下单时间：{{ formatDate(order.created_at) }}</span>
              </p>
            </div>
            <div class="recent-price">
              <el-tag :type="getStatusTagType(order.order_status)" size="small">
                {{ getStatusLabel(order.order_status) }}
              </el-tag>
              <div class="price-val">¥{{ (order.payable_amount / 100).toFixed(2) }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Mobile Menu Hub (Only displayed on mobile screen width) -->
    <div class="mobile-menu-hub">
      <!-- User Info Card -->
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

      <!-- Quick Order Stats -->
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
  </div>
</template>

<script setup lang="ts">
import { ArrowRight } from '@element-plus/icons-vue'
import { useApi } from '~/composables/useApi'

definePageMeta({
  layout: 'user',
})

const { user, logout, init } = useAuth()
const api = useApi()

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
const recentOrders = ref<any[]>([])

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

const statusLabels: Record<string, string> = {
  pending_pay: '待付款',
  paid_full: '待出行',
  pending_travel: '待出行',
  in_travel: '出行中',
  completed: '已完成',
  cancelled: '已取消',
  refunding: '退款中',
  refunded: '已退款',
  closed: '已关闭',
}

function getStatusLabel(status: string): string {
  return statusLabels[status] || status
}

function getStatusTagType(status: string): string {
  const map: Record<string, string> = {
    pending_pay: 'warning',
    paid_full: 'success',
    pending_travel: 'success',
    in_travel: 'primary',
    completed: 'info',
    cancelled: 'info',
    refunding: 'danger',
    refunded: 'danger',
    closed: 'info',
  }
  return map[status] || ''
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

onMounted(async () => {
  await init()
  if (!user.value) {
    navigateTo('/auth/login')
    return
  }

  // Fetch order stats
  try {
    orderStats.value = await api.get<OrderStats>('/orders/stats')
  } catch {}

  // Fetch 3 most recent orders
  try {
    const ordersRes = await api.get<{ items: any[]; total: number }>('/orders', {
      params: { page: 1, page_size: 3 }
    })
    recentOrders.value = ordersRes.items || []
  } catch {}
})

function handleLogout() {
  logout()
}
</script>

<style scoped>
.personal-center {
  background: transparent;
}

/* Desktop Premium Dashboard Styles */
.welcome-title {
  margin: 0 0 24px;
  font-size: 22px;
  font-weight: 700;
  color: #0f172a;
}

.member-gold-card {
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
  border-radius: 16px;
  padding: 30px;
  color: #fff;
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  overflow: hidden;
  box-shadow: 0 10px 20px -5px rgba(15, 23, 42, 0.15);
  margin-bottom: 30px;
}

.member-gold-card::after {
  content: '';
  position: absolute;
  right: -50px;
  bottom: -50px;
  width: 200px;
  height: 200px;
  background: radial-gradient(circle, rgba(245, 158, 11, 0.08) 0%, transparent 70%);
  pointer-events: none;
}

.gold-badge {
  font-size: 11px;
  font-weight: 800;
  color: #f59e0b;
  letter-spacing: 2px;
  margin-bottom: 8px;
  background: rgba(245, 158, 11, 0.15);
  padding: 3px 10px;
  border-radius: 20px;
  display: inline-block;
}

.gold-level {
  font-size: 20px;
  font-weight: 700;
  color: #fef08a; /* Soft Gold */
  margin-bottom: 8px;
}

.gold-number {
  font-size: 13px;
  color: #94a3b8;
}

.real-name-status {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 24px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  color: #f1f5f9;
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  transition: all 0.25s;
}

.real-name-status:hover {
  background: rgba(255, 255, 255, 0.15);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #94a3b8;
}

.status-dot.verified {
  background-color: #10b981;
}

.status-dot.pending {
  background-color: #f59e0b;
}

.status-dot.rejected {
  background-color: #ef4444;
}

/* Stats boxes row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 40px;
}

.stat-card-box {
  background: #f8fafc;
  border: 1px solid #f1f5f9;
  border-radius: 12px;
  padding: 20px;
  text-align: center;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.stat-card-box:hover {
  transform: translateY(-2px);
  background: #fff;
  border-color: #e2e8f0;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.03);
}

.stat-num {
  font-size: 28px;
  font-weight: 800;
  margin-bottom: 6px;
}

.stat-txt {
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
}

.color-pending { color: #f59e0b; }
.color-travel { color: #3b82f6; }
.color-refunding { color: #ef4444; }
.color-total { color: #0f172a; }

/* Recent bookings */
.recent-orders-section {
  border-top: 1px solid #f1f5f9;
  padding-top: 30px;
}

.section-title-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.section-title-bar h3 {
  margin: 0;
  font-size: 17px;
  font-weight: 700;
  color: #0f172a;
}

.more-link {
  font-size: 13px;
  font-weight: 600;
  color: #2563eb;
  text-decoration: none;
}

.more-link:hover {
  text-decoration: underline;
}

.empty-recent {
  text-align: center;
  padding: 40px 20px;
  background: #f8fafc;
  border-radius: 12px;
  border: 1px dashed #e2e8f0;
  color: #64748b;
}

.empty-recent p {
  margin: 0 0 16px;
  font-size: 14px;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.recent-order-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  border-radius: 12px;
  border: 1px solid #f1f5f9;
  background-color: #fff;
  transition: all 0.2s;
  cursor: pointer;
}

.recent-order-item:hover {
  border-color: #cbd5e1;
  box-shadow: 0 4px 12px rgba(0,0,0,0.02);
}

.recent-img {
  width: 72px;
  height: 54px;
  border-radius: 8px;
  flex-shrink: 0;
}

.img-placeholder {
  width: 100%;
  height: 100%;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.recent-details {
  flex: 1;
}

.recent-details h4 {
  margin: 0 0 4px;
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.recent-meta {
  margin: 0;
  font-size: 12px;
  color: #64748b;
  display: flex;
  align-items: center;
  gap: 6px;
}

.meta-dot {
  font-weight: 800;
}

.recent-price {
  text-align: right;
  flex-shrink: 0;
}

.price-val {
  font-size: 16px;
  font-weight: 700;
  color: #ef4444;
  margin-top: 4px;
}

/* Mobile Fallback styles */
.mobile-menu-hub {
  display: none;
}

.user-card {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding: var(--space-lg);
  background: linear-gradient(135deg, #1d4ed8 0%, #3b82f6 100%);
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

.order-stats-card {
  background: var(--color-bg-white);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-md);
  overflow: hidden;
  border: 1px solid var(--color-border-light);
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

.menu-group {
  background: var(--color-bg-white);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-md);
  overflow: hidden;
  border: 1px solid var(--color-border-light);
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

/* Media Queries for Hub Switching */
@media (min-width: 969px) {
  .mobile-menu-hub {
    display: none !important;
  }
  .desktop-dashboard {
    display: block;
  }
}

@media (max-width: 968px) {
  .desktop-dashboard {
    display: none !important;
  }
  .mobile-menu-hub {
    display: block;
  }
}
</style>

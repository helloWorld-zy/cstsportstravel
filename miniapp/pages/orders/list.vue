<template>
  <view class="orders-page">
    <!-- Status Tabs -->
    <view class="tab-bar">
      <scroll-view scroll-x class="tab-scroll">
        <view
          v-for="tab in statusTabs"
          :key="tab.value"
          :class="['tab-item', { active: activeTab === tab.value }]"
          @tap="switchTab(tab.value)"
        >
          {{ tab.label }}
        </view>
      </scroll-view>
    </view>

    <!-- Order List -->
    <view class="order-list">
      <view v-if="loading && orders.length === 0" class="loading-state">
        <uni-load-more status="loading" />
      </view>

      <view v-else-if="orders.length === 0" class="empty-state">
        <uni-empty text="暂无订单" />
        <button class="btn-primary" @tap="goToProducts">去看看产品</button>
      </view>

      <view
        v-for="order in orders"
        :key="order.id"
        class="order-card"
        @tap="goToDetail(order.id)"
      >
        <!-- Header -->
        <view class="card-header">
          <text class="order-no">{{ order.order_no }}</text>
          <text :class="['status-tag', `status-${order.order_status}`]">
            {{ getStatusLabel(order.order_status) }}
          </text>
        </view>

        <!-- Body -->
        <view class="card-body">
          <image
            v-if="order.cover_image"
            :src="order.cover_image"
            class="product-image"
            mode="aspectFill"
          />
          <view class="product-info">
            <text class="product-name">{{ order.product_name || '旅游产品' }}</text>
            <text class="product-meta">
              {{ order.days ? order.days + '天游 · ' : '' }}{{ order.adult_count }}成人
              {{ order.child_count ? '· ' + order.child_count + '儿童' : '' }}
            </text>
            <text class="order-date">{{ formatDate(order.created_at) }}</text>
          </view>
          <view class="amount-section">
            <text class="amount-label">应付</text>
            <text class="amount-value">¥{{ formatAmount(order.payable_amount) }}</text>
          </view>
        </view>

        <!-- Footer Actions -->
        <view class="card-footer" @tap.stop>
          <button
            v-if="order.order_status === 'pending_pay'"
            class="btn-action btn-primary-sm"
            @tap.stop="goToPayment(order.id)"
          >
            立即支付
          </button>
          <button
            v-if="order.order_status === 'pending_pay'"
            class="btn-action btn-default-sm"
            @tap.stop="cancelOrder(order.id)"
          >
            取消订单
          </button>
          <button
            v-if="order.order_status === 'paid_full' || order.order_status === 'pending_travel'"
            class="btn-action btn-danger-sm"
            @tap.stop="requestRefund(order.id)"
          >
            申请退款
          </button>
          <button
            v-if="order.order_status === 'completed'"
            class="btn-action btn-primary-sm"
            @tap.stop="goToReview(order)"
          >
            去评价
          </button>
          <button
            v-if="order.order_status === 'refunding'"
            class="btn-action btn-default-sm"
            @tap.stop="goToDetail(order.id)"
          >
            退款进度
          </button>
        </view>
      </view>
    </view>

    <!-- Load More -->
    <view v-if="orders.length > 0 && hasMore" class="load-more">
      <uni-load-more :status="loading ? 'loading' : 'more'" @tap="loadMore" />
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface OrderItem {
  id: number
  order_no: string
  order_status: string
  product_name: string
  cover_image: string
  days: number
  adult_count: number
  child_count: number
  payable_amount: number
  created_at: string
}

const loading = ref(false)
const orders = ref<OrderItem[]>([])
const activeTab = ref('all')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const hasMore = ref(true)

const statusTabs = [
  { label: '全部', value: 'all' },
  { label: '待付款', value: 'pending_pay' },
  { label: '待出行', value: 'pending_travel' },
  { label: '退款中', value: 'refunding' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' },
]

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

function formatAmount(cents: number): string {
  return (cents / 100).toFixed(2)
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

async function fetchOrders(append = false) {
  loading.value = true
  try {
    const token = uni.getStorageSync('access_token')
    const res = await uni.request({
      url: `${getApp().globalData?.apiBase || ''}/api/v1/orders`,
      method: 'GET',
      data: {
        status: activeTab.value,
        page: page.value,
        page_size: pageSize.value,
      },
      header: {
        Authorization: `Bearer ${token}`,
      },
    })

    const data = res.data as any
    if (data?.code === 0) {
      const items = data.data?.items || []
      total.value = data.data?.total || 0
      if (append) {
        orders.value = [...orders.value, ...items]
      } else {
        orders.value = items
      }
      hasMore.value = orders.value.length < total.value
    }
  } catch (error) {
    console.error('Failed to fetch orders:', error)
    uni.showToast({ title: '获取订单失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function switchTab(tab: string) {
  activeTab.value = tab
  page.value = 1
  orders.value = []
  fetchOrders()
}

function loadMore() {
  if (!hasMore.value || loading.value) return
  page.value++
  fetchOrders(true)
}

function goToDetail(orderId: number) {
  uni.navigateTo({ url: `/pages/orders/detail?id=${orderId}` })
}

function goToPayment(orderId: number) {
  uni.navigateTo({ url: `/pages/payment/index?orderId=${orderId}` })
}

function goToProducts() {
  uni.switchTab({ url: '/pages/index/index' })
}

async function cancelOrder(orderId: number) {
  uni.showModal({
    title: '取消订单',
    content: '确定要取消此订单吗？',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const token = uni.getStorageSync('access_token')
        await uni.request({
          url: `${getApp().globalData?.apiBase || ''}/api/v1/orders/${orderId}/cancel`,
          method: 'POST',
          data: { reason: '用户主动取消' },
          header: { Authorization: `Bearer ${token}` },
        })
        uni.showToast({ title: '订单已取消', icon: 'success' })
        fetchOrders()
      } catch {
        uni.showToast({ title: '取消失败', icon: 'none' })
      }
    },
  })
}

function requestRefund(orderId: number) {
  uni.navigateTo({ url: `/pages/orders/detail?id=${orderId}&action=refund` })
}

function goToReview(order: OrderItem) {
  uni.navigateTo({ url: `/pages/orders/detail?id=${order.id}&action=review` })
}

onMounted(() => {
  fetchOrders()
})
</script>

<style scoped>
.orders-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.tab-bar {
  background: #fff;
  border-bottom: 1px solid #eee;
  position: sticky;
  top: 0;
  z-index: 10;
}

.tab-scroll {
  white-space: nowrap;
  padding: 12px 16px;
}

.tab-item {
  display: inline-block;
  padding: 6px 16px;
  margin-right: 8px;
  font-size: 14px;
  color: #666;
  border-radius: 16px;
  background: #f5f5f5;
}

.tab-item.active {
  color: #fff;
  background: #409eff;
}

.order-list {
  padding: 12px 16px;
}

.order-card {
  background: #fff;
  border-radius: 8px;
  margin-bottom: 12px;
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.order-no {
  font-size: 12px;
  color: #999;
}

.status-tag {
  font-size: 12px;
  font-weight: 500;
}

.status-pending_pay { color: #e6a23c; }
.status-paid_full, .status-pending_travel { color: #67c23a; }
.status-completed { color: #909399; }
.status-cancelled { color: #909399; }
.status-refunding { color: #f56c6c; }

.card-body {
  display: flex;
  padding: 12px 16px;
  gap: 12px;
}

.product-image {
  width: 80px;
  height: 60px;
  border-radius: 4px;
  flex-shrink: 0;
}

.product-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.product-name {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-meta {
  font-size: 12px;
  color: #666;
}

.order-date {
  font-size: 11px;
  color: #999;
}

.amount-section {
  text-align: right;
  flex-shrink: 0;
}

.amount-label {
  display: block;
  font-size: 11px;
  color: #999;
}

.amount-value {
  font-size: 16px;
  font-weight: 600;
  color: #f56c6c;
}

.card-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 8px 16px 12px;
  border-top: 1px solid #f0f0f0;
}

.btn-action {
  font-size: 12px;
  padding: 4px 12px;
  border-radius: 4px;
  line-height: 1.5;
  margin: 0;
}

.btn-primary-sm {
  color: #fff;
  background: #409eff;
  border: none;
}

.btn-danger-sm {
  color: #f56c6c;
  background: #fff;
  border: 1px solid #f56c6c;
}

.btn-default-sm {
  color: #666;
  background: #fff;
  border: 1px solid #ddd;
}

.btn-primary {
  color: #fff;
  background: #409eff;
  border: none;
  border-radius: 8px;
  padding: 10px 24px;
  font-size: 14px;
  margin-top: 16px;
}

.loading-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
}

.load-more {
  padding: 16px 0;
}
</style>

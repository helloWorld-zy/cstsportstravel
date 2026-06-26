<template>
  <div class="orders-page">
    <div class="page-header">
      <h1>我的订单</h1>
      <div class="search-bar">
        <el-input
          v-model="searchQuery"
          placeholder="搜索产品名称或订单号"
          prefix-icon="Search"
          clearable
          @input="onSearch"
        />
      </div>
    </div>

    <!-- Status Tabs -->
    <el-tabs v-model="activeTab" class="order-tabs" @tab-change="onTabChange">
      <el-tab-pane
        v-for="tab in statusTabs"
        :key="tab.value"
        :label="tab.label"
        :name="tab.value"
      />
    </el-tabs>

    <!-- Order List -->
    <div v-loading="loading" class="order-list">
      <div v-if="orders.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无订单">
          <el-button type="primary" @click="$router.push('/products')">去看看产品</el-button>
        </el-empty>
      </div>

      <div
        v-for="order in orders"
        :key="order.id"
        class="order-card"
        @click="$router.push(`/user/order-${order.id}`)"
      >
        <div class="order-card-header">
          <span class="order-no">{{ order.order_no }}</span>
          <el-tag :type="getStatusTagType(order.order_status)" size="small">
            {{ getStatusLabel(order.order_status) }}
          </el-tag>
        </div>

        <div class="order-card-body">
          <div class="product-info">
            <el-image
              :src="order.cover_image"
              class="product-image"
              fit="cover"
            >
              <template #error>
                <div class="image-placeholder">
                  <el-icon><Picture /></el-icon>
                </div>
              </template>
            </el-image>
            <div class="product-detail">
              <h3 class="product-name">{{ order.product_name || '旅游产品' }}</h3>
              <p class="product-meta">
                <span v-if="order.days">{{ order.days }}天游</span>
                <span class="separator" v-if="order.days">·</span>
                <span>{{ order.adult_count }}成人</span>
                <span v-if="order.child_count">·{{ order.child_count }}儿童</span>
                <span v-if="order.infant_count">·{{ order.infant_count }}婴儿</span>
              </p>
              <p class="order-date">下单时间：{{ formatDate(order.created_at) }}</p>
            </div>
          </div>

          <div class="order-amount">
            <span class="amount-label">应付金额</span>
            <span class="amount-value">¥{{ formatAmount(order.payable_amount) }}</span>
          </div>
        </div>

        <div class="order-card-footer">
          <div class="action-buttons" @click.stop>
            <el-button
              v-if="order.order_status === 'pending_pay'"
              type="primary"
              size="small"
              @click.stop="goToPayment(order.id)"
            >
              立即支付
            </el-button>
            <el-button
              v-if="order.order_status === 'pending_pay'"
              size="small"
              @click.stop="cancelOrder(order.id)"
            >
              取消订单
            </el-button>
            <el-button
              v-if="order.order_status === 'paid_full' || order.order_status === 'pending_travel'"
              type="danger"
              size="small"
              plain
              @click.stop="requestRefund(order.id)"
            >
              申请退款
            </el-button>
            <el-button
              v-if="order.order_status === 'completed'"
              type="primary"
              size="small"
              plain
              @click.stop="goToReview(order)"
            >
              去评价
            </el-button>
            <el-button
              v-if="order.order_status === 'refunding'"
              size="small"
              @click.stop="$router.push(`/user/order-${order.id}`)"
            >
              查看退款进度
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="total > pageSize" class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        layout="prev, pager, next"
        @current-change="fetchOrders"
      />
    </div>

    <!-- Refund Dialog -->
    <RefundRequest
      v-if="refundOrderId"
      :order-id="refundOrderId"
      @close="refundOrderId = null"
      @success="onRefundSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Picture } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import RefundRequest from '~/components/RefundRequest.vue'

interface OrderItem {
  id: number
  order_no: string
  order_status: string
  product_id: number
  product_name: string
  cover_image: string
  days: number
  adult_count: number
  child_count: number
  infant_count: number
  payable_amount: number
  created_at: string
}

const router = useRouter()
const loading = ref(false)
const orders = ref<OrderItem[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const activeTab = ref('all')
const searchQuery = ref('')
const refundOrderId = ref<number | null>(null)

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

function formatAmount(cents: number): string {
  return (cents / 100).toFixed(2)
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

async function fetchOrders() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    params.set('status', activeTab.value)
    params.set('page', String(currentPage.value))
    params.set('page_size', String(pageSize.value))
    if (searchQuery.value) {
      params.set('keyword', searchQuery.value)
    }

    const response = await $fetch<{ data: { items: OrderItem[]; total: number } }>(
      `/api/v1/orders?${params.toString()}`
    )
    orders.value = response.data?.items || []
    total.value = response.data?.total || 0
  } catch (error) {
    console.error('Failed to fetch orders:', error)
    ElMessage.error('获取订单列表失败')
  } finally {
    loading.value = false
  }
}

function onTabChange() {
  currentPage.value = 1
  fetchOrders()
}

function onSearch() {
  currentPage.value = 1
  fetchOrders()
}

function goToPayment(orderId: number) {
  router.push(`/payment/${orderId}`)
}

async function cancelOrder(orderId: number) {
  try {
    await ElMessageBox.confirm('确定要取消此订单吗？', '取消订单', {
      confirmButtonText: '确定取消',
      cancelButtonText: '暂不取消',
      type: 'warning',
    })

    await $fetch(`/api/v1/orders/${orderId}/cancel`, {
      method: 'POST',
      body: { reason: '用户主动取消' },
    })

    ElMessage.success('订单已取消')
    fetchOrders()
  } catch (error: any) {
    if (error?.message !== 'cancel') {
      ElMessage.error('取消订单失败')
    }
  }
}

function requestRefund(orderId: number) {
  refundOrderId.value = orderId
}

function onRefundSuccess() {
  refundOrderId.value = null
  fetchOrders()
}

function goToReview(order: OrderItem) {
  router.push(`/user/order-${order.id}`)
}

onMounted(() => {
  fetchOrders()
})
</script>

<style scoped>
.orders-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.search-bar {
  width: 300px;
}

.order-tabs {
  margin-bottom: 20px;
}

.order-card {
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  margin-bottom: 16px;
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.order-card:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.order-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.order-no {
  font-size: 13px;
  color: #909399;
}

.order-card-body {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
}

.product-info {
  display: flex;
  gap: 12px;
  flex: 1;
}

.product-image {
  width: 80px;
  height: 60px;
  border-radius: 4px;
  flex-shrink: 0;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  color: #c0c4cc;
}

.product-detail {
  flex: 1;
}

.product-name {
  margin: 0 0 4px;
  font-size: 15px;
  font-weight: 500;
  color: #303133;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.product-meta {
  margin: 0 0 4px;
  font-size: 13px;
  color: #606266;
}

.separator {
  margin: 0 4px;
}

.order-date {
  margin: 0;
  font-size: 12px;
  color: #909399;
}

.order-amount {
  text-align: right;
  flex-shrink: 0;
  margin-left: 16px;
}

.amount-label {
  display: block;
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.amount-value {
  font-size: 18px;
  font-weight: 600;
  color: #f56c6c;
}

.order-card-footer {
  display: flex;
  justify-content: flex-end;
  padding: 12px 16px;
  border-top: 1px solid #f0f0f0;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.empty-state {
  padding: 60px 0;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>

<template>
  <div class="order-detail-page" v-loading="loading">
    <div v-if="order" class="order-detail">
      <!-- Header -->
      <div class="detail-header">
        <el-page-header @back="$router.push('/user/orders')">
          <template #content>
            <span class="page-title">订单详情</span>
          </template>
        </el-page-header>
        <el-tag :type="getStatusTagType(order.order_status)" size="large">
          {{ getStatusLabel(order.order_status) }}
        </el-tag>
      </div>

      <!-- Product Info -->
      <el-card class="section-card" shadow="never">
        <template #header>
          <span class="section-title">产品信息</span>
        </template>
        <div class="product-section">
          <el-image :src="order.cover_image" class="product-cover" fit="cover">
            <template #error>
              <div class="image-placeholder"><el-icon><Picture /></el-icon></div>
            </template>
          </el-image>
          <div class="product-info">
            <h2 class="product-name">{{ order.product_name }}</h2>
            <p class="product-meta" v-if="order.departure_date">
              出发日期：{{ order.departure_date }}
              <span v-if="order.return_date"> — {{ order.return_date }}</span>
            </p>
            <p class="product-meta" v-if="order.days">{{ order.days }}天{{ order.nights }}晚</p>
          </div>
        </div>
      </el-card>

      <!-- Travellers -->
      <el-card class="section-card" shadow="never">
        <template #header>
          <span class="section-title">出游人信息</span>
        </template>
        <el-table :data="order.travellers" stripe size="small">
          <el-table-column label="姓名" prop="real_name" />
          <el-table-column label="身份证号" prop="id_card_no" />
          <el-table-column label="手机号" prop="phone" />
          <el-table-column label="类型" width="80">
            <template #default="{ row }">
              <el-tag v-if="row.is_child" type="warning" size="small">儿童</el-tag>
              <el-tag v-else-if="row.is_infant" type="info" size="small">婴儿</el-tag>
              <el-tag v-else type="success" size="small">成人</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- Fee Breakdown -->
      <el-card class="section-card" shadow="never">
        <template #header>
          <span class="section-title">费用明细</span>
        </template>
        <div class="fee-list">
          <div class="fee-item">
            <span>产品费用（{{ order.adult_count }}成人<span v-if="order.child_count">+{{ order.child_count }}儿童</span><span v-if="order.infant_count">+{{ order.infant_count }}婴儿</span>）</span>
            <span>¥{{ formatAmount(order.total_amount) }}</span>
          </div>
          <div v-if="order.single_supplement_amount > 0" class="fee-item">
            <span>单房差</span>
            <span>¥{{ formatAmount(order.single_supplement_amount) }}</span>
          </div>
          <div v-if="order.addon_amount > 0" class="fee-item">
            <span>附加服务</span>
            <span>¥{{ formatAmount(order.addon_amount) }}</span>
          </div>
          <div v-if="order.discount_amount > 0" class="fee-item discount">
            <span>优惠</span>
            <span>-¥{{ formatAmount(order.discount_amount) }}</span>
          </div>
          <el-divider />
          <div class="fee-item total">
            <span>应付金额</span>
            <span class="total-amount">¥{{ formatAmount(order.payable_amount) }}</span>
          </div>
        </div>
      </el-card>

      <!-- Payment Records -->
      <el-card v-if="order.payment_records?.length" class="section-card" shadow="never">
        <template #header>
          <span class="section-title">支付记录</span>
        </template>
        <el-table :data="order.payment_records" stripe size="small">
          <el-table-column label="支付单号" prop="payment_no" />
          <el-table-column label="支付渠道" prop="channel" />
          <el-table-column label="金额" width="120">
            <template #default="{ row }">¥{{ formatAmount(row.amount) }}</template>
          </el-table-column>
          <el-table-column label="状态" prop="status" width="80" />
          <el-table-column label="支付时间" width="160">
            <template #default="{ row }">{{ row.paid_at ? formatDate(row.paid_at) : '-' }}</template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- Cancellation Rules -->
      <el-card v-if="order.cancellation_rules?.length" class="section-card" shadow="never">
        <template #header>
          <span class="section-title">退改政策</span>
        </template>
        <el-table :data="order.cancellation_rules" stripe size="small">
          <el-table-column label="距出发日期" min-width="150">
            <template #default="{ row }">
              <span v-if="row.days_before_max">{{ row.days_before_min }}-{{ row.days_before_max }}天</span>
              <span v-else>≥{{ row.days_before_min }}天</span>
            </template>
          </el-table-column>
          <el-table-column label="退款比例" width="100">
            <template #default="{ row }">{{ row.refund_percentage }}%</template>
          </el-table-column>
          <el-table-column label="说明" prop="description" />
        </el-table>
      </el-card>

      <!-- Status Log -->
      <el-card v-if="order.status_logs?.length" class="section-card" shadow="never">
        <template #header>
          <span class="section-title">订单状态记录</span>
        </template>
        <el-timeline>
          <el-timeline-item
            v-for="(log, index) in order.status_logs"
            :key="index"
            :timestamp="formatDate(log.created_at)"
            placement="top"
          >
            <p>{{ getStatusLabel(log.from_status) }} → {{ getStatusLabel(log.to_status) }}</p>
            <p v-if="log.reason" class="log-reason">{{ log.reason }}</p>
          </el-timeline-item>
        </el-timeline>
      </el-card>

      <!-- Action Buttons -->
      <div class="action-bar">
        <el-button
          v-if="order.order_status === 'pending_pay'"
          type="primary"
          size="large"
          @click="$router.push(`/payment/${order.id}`)"
        >
          立即支付
        </el-button>
        <el-button
          v-if="order.order_status === 'pending_pay'"
          size="large"
          @click="cancelOrder"
        >
          取消订单
        </el-button>
        <el-button
          v-if="order.order_status === 'paid_full' || order.order_status === 'pending_travel'"
          type="danger"
          size="large"
          plain
          @click="showRefund = true"
        >
          申请退款
        </el-button>
        <el-button
          v-if="order.order_status === 'completed'"
          type="primary"
          size="large"
          plain
          @click="showReview = true"
        >
          去评价
        </el-button>
      </div>
    </div>

    <!-- Refund Dialog -->
    <RefundRequest
      v-if="showRefund && order"
      :order-id="order.id"
      @close="showRefund = false"
      @success="onRefundSuccess"
    />

    <!-- Review Dialog -->
    <ReviewForm
      v-if="showReview && order"
      :product-id="order.product_id"
      :order-id="order.id"
      @close="showReview = false"
      @success="onReviewSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Picture } from '@element-plus/icons-vue'

definePageMeta({
  layout: 'user',
})
import { ElMessage, ElMessageBox } from 'element-plus'
import RefundRequest from '~/components/RefundRequest.vue'
import ReviewForm from '~/components/ReviewForm.vue'

interface OrderDetail {
  id: number
  order_no: string
  order_status: string
  payment_status: string
  product_id: number
  product_name: string
  cover_image: string
  days: number
  nights: number
  departure_date: string
  return_date: string
  adult_count: number
  child_count: number
  infant_count: number
  total_amount: number
  discount_amount: number
  payable_amount: number
  single_supplement_amount: number
  addon_amount: number
  contact_name: string
  contact_phone: string
  travellers: any[]
  payment_records: any[]
  cancellation_rules: any[]
  status_logs: any[]
  created_at: string
  paid_at?: string
  cancel_reason?: string
}

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const order = ref<OrderDetail | null>(null)
const showRefund = ref(false)
const showReview = ref(false)

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

async function fetchOrder() {
  const id = route.params.id
  loading.value = true
  try {
    const response = await $fetch<{ data: OrderDetail }>(`/api/v1/orders/${id}`)
    order.value = response.data
  } catch (error) {
    console.error('Failed to fetch order:', error)
    ElMessage.error('获取订单详情失败')
  } finally {
    loading.value = false
  }
}

async function cancelOrder() {
  if (!order.value) return
  try {
    await ElMessageBox.confirm('确定要取消此订单吗？', '取消订单', {
      confirmButtonText: '确定取消',
      cancelButtonText: '暂不取消',
      type: 'warning',
    })

    await $fetch(`/api/v1/orders/${order.value.id}/cancel`, {
      method: 'POST',
      body: { reason: '用户主动取消' },
    })

    ElMessage.success('订单已取消')
    fetchOrder()
  } catch (error: any) {
    if (error?.message !== 'cancel') {
      ElMessage.error('取消订单失败')
    }
  }
}

function onRefundSuccess() {
  showRefund.value = false
  fetchOrder()
}

function onReviewSuccess() {
  showReview.value = false
  ElMessage.success('评价提交成功')
  fetchOrder()
}

onMounted(() => {
  fetchOrder()
})
</script>

<style scoped>
.order-detail-page {
  max-width: 100%;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
}

.section-card {
  margin-bottom: 16px;
}

.section-card :deep(.el-card__header) {
  padding: 12px 16px;
  background: #fafafa;
}

.section-title {
  font-weight: 600;
  font-size: 15px;
}

.product-section {
  display: flex;
  gap: 16px;
}

.product-cover {
  width: 120px;
  height: 90px;
  border-radius: 6px;
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

.product-name {
  margin: 0 0 8px;
  font-size: 17px;
  font-weight: 600;
}

.product-meta {
  margin: 0 0 4px;
  font-size: 14px;
  color: #606266;
}

.fee-list {
  padding: 0 8px;
}

.fee-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 14px;
  color: #606266;
}

.fee-item.discount {
  color: #f56c6c;
}

.fee-item.total {
  font-weight: 600;
  font-size: 16px;
  color: #303133;
}

.total-amount {
  color: #f56c6c;
  font-size: 20px;
}

.log-reason {
  font-size: 13px;
  color: #909399;
  margin: 4px 0 0;
}

.action-bar {
  display: flex;
  justify-content: center;
  gap: 16px;
  padding: 24px 0;
}
</style>

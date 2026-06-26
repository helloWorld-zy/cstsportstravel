<template>
  <view class="order-detail-page">
    <view v-if="loading" class="loading-state">
      <uni-load-more status="loading" />
    </view>

    <view v-else-if="order" class="order-content">
      <!-- Status Banner -->
      <view :class="['status-banner', `status-${order.order_status}`]">
        <text class="status-text">{{ getStatusLabel(order.order_status) }}</text>
        <text class="status-desc">{{ getStatusDesc(order.order_status) }}</text>
      </view>

      <!-- Product Info -->
      <view class="section">
        <view class="section-title">产品信息</view>
        <view class="product-row">
          <image
            v-if="order.cover_image"
            :src="order.cover_image"
            class="product-image"
            mode="aspectFill"
          />
          <view class="product-info">
            <text class="product-name">{{ order.product_name }}</text>
            <text class="product-meta" v-if="order.departure_date">
              {{ order.departure_date }} 出发
            </text>
            <text class="product-meta">{{ order.days }}天{{ order.nights }}晚</text>
          </view>
        </view>
      </view>

      <!-- Order Info -->
      <view class="section">
        <view class="section-title">订单信息</view>
        <view class="info-list">
          <view class="info-item">
            <text class="info-label">订单编号</text>
            <text class="info-value">{{ order.order_no }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">下单时间</text>
            <text class="info-value">{{ formatDate(order.created_at) }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">出游人数</text>
            <text class="info-value">
              {{ order.adult_count }}成人
              {{ order.child_count ? '· ' + order.child_count + '儿童' : '' }}
              {{ order.infant_count ? '· ' + order.infant_count + '婴儿' : '' }}
            </text>
          </view>
          <view class="info-item">
            <text class="info-label">联系人</text>
            <text class="info-value">{{ order.contact_name }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">联系电话</text>
            <text class="info-value">{{ order.contact_phone }}</text>
          </view>
        </view>
      </view>

      <!-- Travellers -->
      <view class="section" v-if="order.travellers?.length">
        <view class="section-title">出游人信息</view>
        <view
          v-for="(traveller, index) in order.travellers"
          :key="index"
          class="traveller-card"
        >
          <view class="traveller-row">
            <text class="traveller-name">{{ traveller.real_name }}</text>
            <text :class="['traveller-type', traveller.is_child ? 'child' : traveller.is_infant ? 'infant' : 'adult']">
              {{ traveller.is_child ? '儿童' : traveller.is_infant ? '婴儿' : '成人' }}
            </text>
          </view>
          <text class="traveller-id">{{ traveller.id_card_no }}</text>
          <text v-if="traveller.phone" class="traveller-phone">{{ traveller.phone }}</text>
        </view>
      </view>

      <!-- Fee Breakdown -->
      <view class="section">
        <view class="section-title">费用明细</view>
        <view class="fee-list">
          <view class="fee-item">
            <text>产品费用</text>
            <text>¥{{ formatAmount(order.total_amount) }}</text>
          </view>
          <view v-if="order.single_supplement_amount > 0" class="fee-item">
            <text>单房差</text>
            <text>¥{{ formatAmount(order.single_supplement_amount) }}</text>
          </view>
          <view v-if="order.discount_amount > 0" class="fee-item discount">
            <text>优惠</text>
            <text>-¥{{ formatAmount(order.discount_amount) }}</text>
          </view>
          <view class="fee-divider" />
          <view class="fee-item total">
            <text>应付金额</text>
            <text class="total-amount">¥{{ formatAmount(order.payable_amount) }}</text>
          </view>
        </view>
      </view>

      <!-- Status Log -->
      <view class="section" v-if="order.status_logs?.length">
        <view class="section-title">状态记录</view>
        <view class="timeline">
          <view
            v-for="(log, index) in order.status_logs"
            :key="index"
            class="timeline-item"
          >
            <view class="timeline-dot" :class="{ active: index === 0 }" />
            <view class="timeline-content">
              <text class="timeline-status">
                {{ getStatusLabel(log.from_status) }} → {{ getStatusLabel(log.to_status) }}
              </text>
              <text v-if="log.reason" class="timeline-reason">{{ log.reason }}</text>
              <text class="timeline-time">{{ formatDate(log.created_at) }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- Action Buttons -->
      <view class="action-bar">
        <button
          v-if="order.order_status === 'pending_pay'"
          class="btn-action btn-primary"
          @tap="goToPayment"
        >
          立即支付
        </button>
        <button
          v-if="order.order_status === 'pending_pay'"
          class="btn-action btn-default"
          @tap="cancelOrder"
        >
          取消订单
        </button>
        <button
          v-if="order.order_status === 'paid_full' || order.order_status === 'pending_travel'"
          class="btn-action btn-danger"
          @tap="requestRefund"
        >
          申请退款
        </button>
        <button
          v-if="order.order_status === 'completed'"
          class="btn-action btn-primary"
          @tap="goToReview"
        >
          去评价
        </button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

interface OrderDetail {
  id: number
  order_no: string
  order_status: string
  product_name: string
  cover_image: string
  days: number
  nights: number
  departure_date: string
  adult_count: number
  child_count: number
  infant_count: number
  total_amount: number
  discount_amount: number
  payable_amount: number
  single_supplement_amount: number
  contact_name: string
  contact_phone: string
  travellers: any[]
  status_logs: any[]
  created_at: string
}

const loading = ref(false)
const order = ref<OrderDetail | null>(null)
let orderId = 0

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

const statusDescs: Record<string, string> = {
  pending_pay: '请在30分钟内完成支付',
  paid_full: '等待出发日期到来',
  pending_travel: '即将出发，请做好准备',
  in_travel: '祝您旅途愉快',
  completed: '行程已结束',
  cancelled: '订单已取消',
  refunding: '退款申请审核中',
  refunded: '退款已完成',
}

function getStatusLabel(status: string): string {
  return statusLabels[status] || status
}

function getStatusDesc(status: string): string {
  return statusDescs[status] || ''
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
  loading.value = true
  try {
    const token = uni.getStorageSync('access_token')
    const res = await uni.request({
      url: `${getApp().globalData?.apiBase || ''}/api/v1/orders/${orderId}`,
      method: 'GET',
      header: { Authorization: `Bearer ${token}` },
    })

    const data = res.data as any
    if (data?.code === 0) {
      order.value = data.data
    }
  } catch (error) {
    console.error('Failed to fetch order:', error)
    uni.showToast({ title: '获取订单失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function goToPayment() {
  uni.navigateTo({ url: `/pages/payment/index?orderId=${orderId}` })
}

function cancelOrder() {
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
        fetchOrder()
      } catch {
        uni.showToast({ title: '取消失败', icon: 'none' })
      }
    },
  })
}

function requestRefund() {
  // Navigate to refund form or show modal
  uni.showToast({ title: '请在Web端申请退款', icon: 'none' })
}

function goToReview() {
  uni.showToast({ title: '请在Web端提交评价', icon: 'none' })
}

onLoad((options) => {
  if (options?.id) {
    orderId = Number(options.id)
    fetchOrder()
  }
})
</script>

<style scoped>
.order-detail-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.status-banner {
  padding: 24px 16px;
  color: #fff;
}

.status-pending_pay { background: linear-gradient(135deg, #e6a23c, #f0c78a); }
.status-paid_full, .status-pending_travel { background: linear-gradient(135deg, #67c23a, #95d475); }
.status-in_travel { background: linear-gradient(135deg, #409eff, #79bbff); }
.status-completed { background: linear-gradient(135deg, #909399, #b1b3b8); }
.status-cancelled { background: linear-gradient(135deg, #909399, #b1b3b8); }
.status-refunding { background: linear-gradient(135deg, #f56c6c, #f89898); }

.status-text {
  font-size: 20px;
  font-weight: 600;
  display: block;
}

.status-desc {
  font-size: 13px;
  opacity: 0.9;
  margin-top: 4px;
  display: block;
}

.section {
  background: #fff;
  margin: 12px 16px;
  border-radius: 8px;
  padding: 16px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #333;
  margin-bottom: 12px;
}

.product-row {
  display: flex;
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
  font-size: 15px;
  font-weight: 500;
  color: #333;
}

.product-meta {
  font-size: 13px;
  color: #666;
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-item {
  display: flex;
  justify-content: space-between;
}

.info-label {
  font-size: 14px;
  color: #999;
}

.info-value {
  font-size: 14px;
  color: #333;
}

.traveller-card {
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.traveller-card:last-child {
  border-bottom: none;
}

.traveller-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.traveller-name {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.traveller-type {
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 3px;
}

.traveller-type.adult { color: #67c23a; background: #f0f9eb; }
.traveller-type.child { color: #e6a23c; background: #fdf6ec; }
.traveller-type.infant { color: #909399; background: #f4f4f5; }

.traveller-id, .traveller-phone {
  font-size: 12px;
  color: #999;
  display: block;
}

.fee-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.fee-item {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  color: #666;
}

.fee-item.discount { color: #f56c6c; }
.fee-item.total { font-weight: 600; color: #333; }

.fee-divider {
  height: 1px;
  background: #f0f0f0;
  margin: 4px 0;
}

.total-amount {
  color: #f56c6c;
  font-size: 18px;
}

.timeline {
  padding-left: 16px;
}

.timeline-item {
  display: flex;
  gap: 12px;
  padding-bottom: 16px;
  position: relative;
}

.timeline-item::before {
  content: '';
  position: absolute;
  left: 4px;
  top: 12px;
  bottom: 0;
  width: 1px;
  background: #e0e0e0;
}

.timeline-item:last-child::before {
  display: none;
}

.timeline-dot {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: #ddd;
  flex-shrink: 0;
  margin-top: 4px;
}

.timeline-dot.active {
  background: #409eff;
}

.timeline-content {
  flex: 1;
}

.timeline-status {
  font-size: 13px;
  color: #333;
  display: block;
}

.timeline-reason {
  font-size: 12px;
  color: #999;
  display: block;
  margin-top: 2px;
}

.timeline-time {
  font-size: 11px;
  color: #bbb;
  display: block;
  margin-top: 2px;
}

.action-bar {
  display: flex;
  gap: 12px;
  padding: 16px;
}

.btn-action {
  flex: 1;
  font-size: 14px;
  padding: 10px 0;
  border-radius: 8px;
  text-align: center;
}

.btn-primary {
  color: #fff;
  background: #409eff;
  border: none;
}

.btn-danger {
  color: #f56c6c;
  background: #fff;
  border: 1px solid #f56c6c;
}

.btn-default {
  color: #666;
  background: #fff;
  border: 1px solid #ddd;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
}
</style>

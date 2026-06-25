<template>
  <div class="payment-page">
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">
      <el-result icon="error" :title="error" />
    </div>
    <div v-else class="payment-content">
      <!-- Countdown -->
      <PaymentCountdown :expire-at="order.expire_at" @expired="onExpired" />

      <!-- Order summary -->
      <div class="order-summary">
        <h3>订单信息</h3>
        <div class="info-row">
          <span class="label">订单号</span>
          <span class="value">{{ order.order_no }}</span>
        </div>
        <div class="info-row">
          <span class="label">应付金额</span>
          <span class="value price">¥{{ formatAmount(order.payable_amount) }}</span>
        </div>
      </div>

      <!-- Payment method selection -->
      <div class="payment-methods">
        <h3>选择支付方式</h3>
        <div
          class="method-card"
          :class="{ selected: selectedChannel === 'alipay' }"
          @click="selectedChannel = 'alipay'"
        >
          <div class="method-icon alipay-icon">支付宝</div>
          <div class="method-name">支付宝支付</div>
        </div>
        <div
          class="method-card"
          :class="{ selected: selectedChannel === 'wechat' }"
          @click="selectedChannel = 'wechat'"
        >
          <div class="method-icon wechat-icon">微信</div>
          <div class="method-name">微信支付</div>
        </div>
      </div>

      <!-- Pay button -->
      <div class="pay-actions">
        <el-button
          type="primary"
          size="large"
          :loading="paying"
          :disabled="!selectedChannel || order.order_status !== 'pending_pay'"
          @click="handlePay"
        >
          立即支付 ¥{{ formatAmount(order.payable_amount) }}
        </el-button>
      </div>

      <!-- Status display -->
      <div v-if="order.order_status !== 'pending_pay'" class="status-display">
        <el-result
          :icon="order.order_status === 'paid_full' ? 'success' : 'warning'"
          :title="statusText"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { formatAmount } from '~/shared/utils/amount'

definePageMeta({
  layout: 'default',
  middleware: ['auth'],
})

const route = useRoute()
const router = useRouter()
const api = useApi()

const orderId = Number(route.params.orderId)
const loading = ref(true)
const error = ref('')
const order = ref<any>({})
const selectedChannel = ref('alipay')
const paying = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const statusText = computed(() => {
  const statusMap: Record<string, string> = {
    paid_full: '支付成功',
    pending_travel: '支付成功',
    cancelled: '订单已取消',
    refunding: '退款处理中',
  }
  return statusMap[order.value.order_status] || order.value.order_status
})

onMounted(async () => {
  try {
    order.value = await api.get<any>(`/orders/${orderId}`)
    if (order.value.order_status === 'pending_pay') {
      startPolling()
    }
  } catch (err: any) {
    error.value = err.message || '加载订单失败'
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

function startPolling() {
  pollTimer = setInterval(async () => {
    try {
      const updated = await api.get<any>(`/orders/${orderId}`)
      order.value = updated
      if (updated.order_status !== 'pending_pay') {
        if (pollTimer) clearInterval(pollTimer)
        if (updated.order_status === 'paid_full' || updated.order_status === 'pending_travel') {
          ElMessage.success('支付成功！')
        }
      }
    } catch {
      // Ignore polling errors
    }
  }, 3000)
}

async function handlePay() {
  if (!selectedChannel.value) return
  paying.value = true

  try {
    const result = await api.post<any>('/payments/create', {
      order_id: orderId,
      channel: selectedChannel.value,
    })

    // In production, redirect to payment URL or show QR code
    if (result.pay_params) {
      if (selectedChannel.value === 'alipay' && result.pay_params.pay_url) {
        // Redirect to Alipay
        window.location.href = result.pay_params.pay_url
      } else if (selectedChannel.value === 'wechat' && result.pay_params.code_url) {
        // Show WeChat QR code (would need a QR code component)
        ElMessage.info('请使用微信扫描二维码支付')
      }
    }

    // Start polling for payment result
    startPolling()
  } catch (err: any) {
    ElMessage.error(err.message || '创建支付失败')
  } finally {
    paying.value = false
  }
}

function onExpired() {
  ElMessage.warning('支付已超时，订单已自动取消')
  order.value.order_status = 'cancelled'
}
</script>

<style scoped>
.payment-page {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

.loading, .error {
  text-align: center;
  padding: 40px;
}

.order-summary {
  background: #fafafa;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 24px;
}

.order-summary h3 {
  margin-bottom: 12px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.info-row .label {
  color: #666;
}

.info-row .price {
  color: #ff4d4f;
  font-size: 20px;
  font-weight: bold;
}

.payment-methods {
  margin-bottom: 24px;
}

.payment-methods h3 {
  margin-bottom: 12px;
}

.method-card {
  display: flex;
  align-items: center;
  gap: 12px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.method-card:hover {
  border-color: #409eff;
}

.method-card.selected {
  border-color: #409eff;
  background: #ecf5ff;
}

.method-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: bold;
  color: #fff;
}

.alipay-icon {
  background: #1677ff;
}

.wechat-icon {
  background: #07c160;
}

.method-name {
  font-weight: 500;
}

.pay-actions {
  text-align: center;
  margin-top: 24px;
}

.pay-actions .el-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
}

.status-display {
  margin-top: 24px;
}
</style>

<template>
  <view class="payment-page">
    <view v-if="loading" class="loading">加载中...</view>
    <view v-else>
      <!-- Countdown -->
      <view class="countdown" :class="{ warning: isWarning, expired: isExpired }">
        <text v-if="isExpired">支付已超时</text>
        <text v-else>支付剩余时间：{{ minutes }}:{{ seconds }}</text>
      </view>

      <!-- Order info -->
      <view class="order-card">
        <text class="order-no">订单号：{{ order.order_no }}</text>
        <text class="amount">¥{{ (order.payable_amount / 100).toFixed(2) }}</text>
      </view>

      <!-- Pay button -->
      <button
        v-if="order.order_status === 'pending_pay'"
        type="primary"
        class="pay-btn"
        :loading="paying"
        @tap="handlePay"
      >
        微信支付 ¥{{ (order.payable_amount / 100).toFixed(2) }}
      </button>

      <!-- Status -->
      <view v-else class="status-display">
        <text v-if="order.order_status === 'paid_full' || order.order_status === 'pending_travel'" class="status-success">支付成功</text>
        <text v-else-if="order.order_status === 'cancelled'" class="status-cancelled">订单已取消</text>
        <text v-else class="status-other">{{ order.order_status }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { request } from '@/shared/api/request'

const orderId = ref(0)
const loading = ref(true)
const order = ref<any>({})
const paying = ref(false)
const remaining = ref(0)
let timer: any = null

const isWarning = computed(() => remaining.value > 0 && remaining.value <= 300)
const isExpired = computed(() => remaining.value <= 0)
const minutes = computed(() => String(Math.floor(remaining.value / 60)).padStart(2, '0'))
const seconds = computed(() => String(remaining.value % 60).padStart(2, '0'))

onMounted(async () => {
  const pages = getCurrentPages()
  const page = pages[pages.length - 1]
  orderId.value = Number(page?.options?.orderId || 0)

  try {
    const res = await request({ url: `/orders/${orderId.value}`, method: 'GET' })
    order.value = res.data
    updateRemaining()
    timer = setInterval(updateRemaining, 1000)
  } catch (e) {
    uni.showToast({ title: '加载订单失败', icon: 'none' })
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function updateRemaining() {
  if (!order.value.expire_at) return
  const expire = new Date(order.value.expire_at).getTime()
  const diff = Math.floor((expire - Date.now()) / 1000)
  remaining.value = Math.max(0, diff)

  if (remaining.value <= 0 && timer) {
    clearInterval(timer)
    timer = null
  }
}

async function handlePay() {
  paying.value = true
  try {
    // Create payment
    const payRes = await request({
      url: '/payments/create',
      method: 'POST',
      data: {
        order_id: orderId.value,
        channel: 'wechat',
        method: 'miniapp',
      },
    })

    const payParams = payRes.data.pay_params

    // #ifdef MP-WEIXIN
    // Call wx.requestPayment for WeChat mini program
    uni.requestPayment({
      provider: 'wxpay',
      timeStamp: payParams.timestamp || String(Math.floor(Date.now() / 1000)),
      nonceStr: payParams.nonce_str || '',
      package: `prepay_id=${payParams.prepay_id || ''}`,
      signType: payParams.sign_type || 'RSA',
      paySign: payParams.pay_sign || '',
      success: () => {
        uni.showToast({ title: '支付成功', icon: 'success' })
        // Refresh order status
        refreshOrder()
      },
      fail: (err) => {
        if (err.errMsg !== 'requestPayment:fail cancel') {
          uni.showToast({ title: '支付失败', icon: 'none' })
        }
      },
    })
    // #endif

    // #ifndef MP-WEIXIN
    uni.showToast({ title: '请在微信小程序中使用微信支付', icon: 'none' })
    // #endif
  } catch (e: any) {
    uni.showToast({ title: e.message || '创建支付失败', icon: 'none' })
  } finally {
    paying.value = false
  }
}

async function refreshOrder() {
  try {
    const res = await request({ url: `/orders/${orderId.value}`, method: 'GET' })
    order.value = res.data
  } catch {
    // ignore
  }
}
</script>

<style scoped>
.payment-page { padding: 30rpx; }
.loading { text-align: center; padding: 100rpx; }
.countdown { text-align: center; font-size: 32rpx; padding: 30rpx; margin-bottom: 30rpx; }
.countdown.warning text { color: #faad14; }
.countdown.expired text { color: #ff4d4f; font-weight: bold; }
.order-card { background: #fafafa; border-radius: 16rpx; padding: 30rpx; margin-bottom: 30rpx; text-align: center; }
.order-no { display: block; color: #666; margin-bottom: 16rpx; }
.amount { font-size: 48rpx; color: #ff4d4f; font-weight: bold; }
.pay-btn { width: 100%; height: 88rpx; line-height: 88rpx; font-size: 32rpx; margin-top: 30rpx; }
.status-display { text-align: center; padding: 60rpx; }
.status-success { color: #67c23a; font-size: 36rpx; font-weight: bold; }
.status-cancelled { color: #999; font-size: 36rpx; }
.status-other { color: #333; font-size: 36rpx; }
</style>

<template>
  <div class="balance-reminder">
    <div class="reminder-header">
      <h2>待付尾款</h2>
      <div class="status-badge" :class="statusClass">
        {{ statusText }}
      </div>
    </div>

    <!-- Order Info -->
    <div class="order-info">
      <div class="info-row">
        <span class="label">订单编号：</span>
        <span class="value">{{ orderNo }}</span>
      </div>
      <div class="info-row">
        <span class="label">产品名称：</span>
        <span class="value">{{ productName }}</span>
      </div>
    </div>

    <!-- Payment Summary -->
    <div class="payment-summary">
      <div class="summary-item">
        <span class="summary-label">订单总额</span>
        <span class="summary-amount">¥{{ formatAmount(totalAmount) }}</span>
      </div>
      <div class="summary-item paid">
        <span class="summary-label">已付定金</span>
        <span class="summary-amount">¥{{ formatAmount(depositAmount) }}</span>
        <span class="summary-time">{{ formatDate(depositPaidAt) }}</span>
      </div>
      <div class="summary-item due">
        <span class="summary-label">应付尾款</span>
        <span class="summary-amount highlight">¥{{ formatAmount(balanceAmount) }}</span>
      </div>
    </div>

    <!-- Countdown -->
    <div class="countdown-section" v-if="!isOverdue && !isPaid">
      <div class="countdown-label">尾款支付截止倒计时</div>
      <div class="countdown-timer">
        <div class="time-unit">
          <span class="time-value">{{ countdown.days }}</span>
          <span class="time-label">天</span>
        </div>
        <div class="time-separator">:</div>
        <div class="time-unit">
          <span class="time-value">{{ countdown.hours }}</span>
          <span class="time-label">时</span>
        </div>
        <div class="time-separator">:</div>
        <div class="time-unit">
          <span class="time-value">{{ countdown.minutes }}</span>
          <span class="time-label">分</span>
        </div>
        <div class="time-separator">:</div>
        <div class="time-unit">
          <span class="time-value">{{ countdown.seconds }}</span>
          <span class="time-label">秒</span>
        </div>
      </div>
      <div class="deadline-info">
        截止时间：{{ formatDate(balanceDeadline) }}
      </div>
    </div>

    <!-- Overdue Warning -->
    <div v-if="isOverdue" class="overdue-warning">
      <div class="warning-icon">⚠️</div>
      <div class="warning-text">
        <h3>尾款支付已逾期</h3>
        <p>截止日后24小时宽限期已结束，订单将自动取消</p>
        <p>定金将按退改规则退还</p>
      </div>
    </div>

    <!-- Paid Status -->
    <div v-if="isPaid" class="paid-status">
      <div class="paid-icon">✅</div>
      <div class="paid-text">
        <h3>尾款已支付</h3>
        <p>支付时间：{{ formatDate(balancePaidAt) }}</p>
      </div>
    </div>

    <!-- Actions -->
    <div class="actions" v-if="!isPaid">
      <button
        v-if="!isOverdue"
        class="pay-btn"
        @click="onPayBalance"
      >
        立即支付尾款 ¥{{ formatAmount(balanceAmount) }}
      </button>
      <button
        v-if="isOverdue"
        class="cancel-btn"
        @click="onCancelOrder"
      >
        取消订单
      </button>
    </div>

    <!-- Overdue Policy -->
    <div class="policy-section">
      <h4>逾期处理说明</h4>
      <ul>
        <li>尾款支付截止时间为出发前{{ deadlineDays }}天</li>
        <li>截止日后有24小时宽限期</li>
        <li>宽限期内未付尾款，订单自动取消</li>
        <li>定金将按退改规则退还</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface Props {
  orderNo: string
  productName: string
  totalAmount: number        // cents
  depositAmount: number      // cents
  balanceAmount: number      // cents
  depositPaidAt?: string
  balanceDeadline?: string
  balancePaidAt?: string
  deadlineDays?: number
}

const props = withDefaults(defineProps<Props>(), {
  deadlineDays: 30,
})

const emit = defineEmits<{
  (e: 'pay-balance'): void
  (e: 'cancel-order'): void
}>()

// State
const countdown = ref({ days: 0, hours: 0, minutes: 0, seconds: 0 })
let timer: ReturnType<typeof setInterval> | null = null

// Computed
const isPaid = computed(() => !!props.balancePaidAt)

const isOverdue = computed(() => {
  if (!props.balanceDeadline || isPaid.value) return false
  const deadline = new Date(props.balanceDeadline)
  const graceDeadline = new Date(deadline.getTime() + 24 * 60 * 60 * 1000)
  return new Date() > graceDeadline
})

const statusClass = computed(() => {
  if (isPaid.value) return 'paid'
  if (isOverdue.value) return 'overdue'
  return 'pending'
})

const statusText = computed(() => {
  if (isPaid.value) return '已付全款'
  if (isOverdue.value) return '已逾期'
  return '待付尾款'
})

// Methods
const updateCountdown = () => {
  if (!props.balanceDeadline || isPaid.value) return

  const now = new Date().getTime()
  const deadline = new Date(props.balanceDeadline).getTime()
  const diff = deadline - now

  if (diff <= 0) {
    countdown.value = { days: 0, hours: 0, minutes: 0, seconds: 0 }
    return
  }

  countdown.value = {
    days: Math.floor(diff / (1000 * 60 * 60 * 24)),
    hours: Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)),
    minutes: Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60)),
    seconds: Math.floor((diff % (1000 * 60)) / 1000),
  }
}

const formatAmount = (cents: number): string => {
  return (cents / 100).toFixed(2)
}

const formatDate = (dateStr?: string): string => {
  if (!dateStr) return '未设置'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const onPayBalance = () => {
  emit('pay-balance')
}

const onCancelOrder = () => {
  emit('cancel-order')
}

// Lifecycle
onMounted(() => {
  updateCountdown()
  timer = setInterval(updateCountdown, 1000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style scoped>
.balance-reminder {
  max-width: 600px;
  margin: 0 auto;
  padding: 24px;
}

.reminder-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 500;
}

.status-badge.pending {
  background: #fff7e6;
  color: #fa8c16;
  border: 1px solid #ffd591;
}

.status-badge.paid {
  background: #f6ffed;
  color: #52c41a;
  border: 1px solid #b7eb8f;
}

.status-badge.overdue {
  background: #fff2f0;
  color: #ff4d4f;
  border: 1px solid #ffccc7;
}

.payment-summary {
  background: #fafafa;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.summary-label {
  color: #666;
}

.summary-amount {
  font-size: 16px;
  font-weight: 600;
}

.summary-amount.highlight {
  color: #ff4d4f;
  font-size: 20px;
}

.summary-time {
  font-size: 12px;
  color: #999;
}

.countdown-section {
  text-align: center;
  margin-bottom: 24px;
}

.countdown-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 12px;
}

.countdown-timer {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.time-unit {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.time-value {
  font-size: 32px;
  font-weight: 700;
  color: #1890ff;
  background: #f0f7ff;
  padding: 8px 16px;
  border-radius: 8px;
  min-width: 60px;
}

.time-label {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.time-separator {
  font-size: 32px;
  font-weight: 700;
  color: #1890ff;
  align-self: flex-start;
  padding-top: 8px;
}

.deadline-info {
  margin-top: 12px;
  font-size: 14px;
  color: #999;
}

.overdue-warning {
  background: #fff2f0;
  border: 1px solid #ffccc7;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
  display: flex;
  gap: 12px;
}

.warning-text h3 {
  color: #ff4d4f;
  margin: 0 0 8px 0;
}

.warning-text p {
  color: #666;
  margin: 4px 0;
  font-size: 14px;
}

.paid-status {
  background: #f6ffed;
  border: 1px solid #b7eb8f;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
  display: flex;
  gap: 12px;
}

.paid-text h3 {
  color: #52c41a;
  margin: 0 0 8px 0;
}

.actions {
  margin-bottom: 24px;
}

.pay-btn {
  width: 100%;
  padding: 14px;
  background: #1890ff;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
}

.pay-btn:hover {
  background: #40a9ff;
}

.cancel-btn {
  width: 100%;
  padding: 14px;
  background: #ff4d4f;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
}

.policy-section {
  background: #fafafa;
  border-radius: 8px;
  padding: 16px;
}

.policy-section h4 {
  margin: 0 0 12px 0;
  color: #333;
}

.policy-section ul {
  margin: 0;
  padding-left: 20px;
  color: #666;
  font-size: 14px;
}

.policy-section li {
  margin-bottom: 4px;
}
</style>

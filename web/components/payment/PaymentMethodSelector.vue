<template>
  <div class="payment-method-selector">
    <!-- Payment Mode Selection (Full vs Deposit+Balance) -->
    <div v-if="showDepositOption" class="payment-mode-section">
      <h3 class="section-title">支付方式</h3>
      <div class="mode-options">
        <label
          class="mode-option"
          :class="{ active: selectedMode === 'full' }"
        >
          <input
            type="radio"
            v-model="selectedMode"
            value="full"
            @change="onModeChange"
          />
          <div class="mode-content">
            <span class="mode-name">全额支付</span>
            <span class="mode-amount">¥{{ formatAmount(totalAmount) }}</span>
          </div>
        </label>

        <label
          class="mode-option"
          :class="{ active: selectedMode === 'deposit' }"
        >
          <input
            type="radio"
            v-model="selectedMode"
            value="deposit"
            @change="onModeChange"
          />
          <div class="mode-content">
            <span class="mode-name">定金+尾款</span>
            <span class="mode-detail">
              定金 ¥{{ formatAmount(depositAmount) }} + 尾款 ¥{{ formatAmount(balanceAmount) }}
            </span>
          </div>
        </label>
      </div>

      <!-- Deposit Info -->
      <div v-if="selectedMode === 'deposit'" class="deposit-info">
        <div class="info-item">
          <span class="label">定金金额：</span>
          <span class="value">¥{{ formatAmount(depositAmount) }}</span>
        </div>
        <div class="info-item">
          <span class="label">尾款金额：</span>
          <span class="value">¥{{ formatAmount(balanceAmount) }}</span>
        </div>
        <div class="info-item">
          <span class="label">尾款截止：</span>
          <span class="value">{{ formatDate(balanceDeadline) }}</span>
        </div>
        <div class="info-item warning">
          <span class="label">逾期说明：</span>
          <span class="value">截止日后24小时宽限期内未付尾款，订单自动取消</span>
        </div>
      </div>
    </div>

    <!-- Payment Channel Selection -->
    <div class="channel-section">
      <h3 class="section-title">选择支付渠道</h3>
      <div class="channel-options">
        <label
          v-for="channel in availableChannels"
          :key="channel.id"
          class="channel-option"
          :class="{ active: selectedChannel === channel.id }"
        >
          <input
            type="radio"
            v-model="selectedChannel"
            :value="channel.id"
            @change="onChannelChange"
          />
          <div class="channel-content">
            <img :src="channel.icon" :alt="channel.name" class="channel-icon" />
            <span class="channel-name">{{ channel.name }}</span>
          </div>
        </label>
      </div>
    </div>

    <!-- UnionPay Sub-methods -->
    <div v-if="selectedChannel === 'unionpay'" class="unionpay-methods">
      <h3 class="section-title">银联支付方式</h3>
      <div class="method-options">
        <label
          class="method-option"
          :class="{ active: selectedMethod === 'gateway' }"
        >
          <input
            type="radio"
            v-model="selectedMethod"
            value="gateway"
          />
          <div class="method-content">
            <span class="method-name">网关支付</span>
            <span class="method-desc">PC端浏览器支付</span>
          </div>
        </label>

        <label
          class="method-option"
          :class="{ active: selectedMethod === 'wap' }"
        >
          <input
            type="radio"
            v-model="selectedMethod"
            value="wap"
          />
          <div class="method-content">
            <span class="method-name">手机WAP支付</span>
            <span class="method-desc">移动端浏览器支付</span>
          </div>
        </label>
      </div>
    </div>

    <!-- Submit Button -->
    <div class="submit-section">
      <button
        class="submit-btn"
        :disabled="!isValid"
        @click="onSubmit"
      >
        确认支付 ¥{{ formatAmount(currentPayAmount) }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  totalAmount: number        // 总金额 (cents)
  depositRatio?: number      // 定金比例 (0.10-0.50)
  balanceDeadline?: string   // 尾款截止日期
  showDepositOption?: boolean // 是否显示定金+尾款选项
  productType?: string       // 产品类型 (outbound_group, cruise)
}

const props = withDefaults(defineProps<Props>(), {
  depositRatio: 0.30,
  showDepositOption: false,
  productType: '',
})

const emit = defineEmits<{
  (e: 'submit', data: PaymentSubmitData): void
}>()

interface PaymentSubmitData {
  mode: 'full' | 'deposit'
  channel: string
  method: string
  amount: number
}

// State
const selectedMode = ref<'full' | 'deposit'>('full')
const selectedChannel = ref('alipay')
const selectedMethod = ref('pc')

// Computed
const depositAmount = computed(() => {
  return Math.ceil(props.totalAmount * props.depositRatio)
})

const balanceAmount = computed(() => {
  return props.totalAmount - depositAmount.value
})

const currentPayAmount = computed(() => {
  if (selectedMode.value === 'deposit') {
    return depositAmount.value
  }
  return props.totalAmount
})

const isValid = computed(() => {
  return selectedChannel.value && selectedMethod.value
})

// Available channels
const availableChannels = [
  { id: 'alipay', name: '支付宝', icon: '/icons/alipay.svg' },
  { id: 'wechat', name: '微信支付', icon: '/icons/wechat.svg' },
  { id: 'unionpay', name: '银联支付', icon: '/icons/unionpay.svg' },
]

// Methods
const onModeChange = () => {
  // Reset method when switching modes
  if (selectedChannel.value !== 'unionpay') {
    selectedMethod.value = 'pc'
  }
}

const onChannelChange = () => {
  // Set default method based on channel
  if (selectedChannel.value === 'unionpay') {
    selectedMethod.value = 'gateway'
  } else {
    selectedMethod.value = 'pc'
  }
}

const onSubmit = () => {
  emit('submit', {
    mode: selectedMode.value,
    channel: selectedChannel.value,
    method: selectedMethod.value,
    amount: currentPayAmount.value,
  })
}

const formatAmount = (cents: number): string => {
  return (cents / 100).toFixed(2)
}

const formatDate = (dateStr?: string): string => {
  if (!dateStr) return '未设置'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.payment-method-selector {
  padding: 20px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #333;
}

.mode-options,
.channel-options,
.method-options {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.mode-option,
.channel-option,
.method-option {
  flex: 1;
  min-width: 200px;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.mode-option.active,
.channel-option.active,
.method-option.active {
  border-color: #1890ff;
  background: #f0f7ff;
}

.mode-option input,
.channel-option input,
.method-option input {
  display: none;
}

.mode-content,
.channel-content,
.method-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mode-name,
.channel-name,
.method-name {
  font-size: 14px;
  font-weight: 500;
}

.mode-amount {
  font-size: 18px;
  font-weight: 600;
  color: #ff4d4f;
}

.mode-detail {
  font-size: 12px;
  color: #666;
}

.channel-icon {
  width: 24px;
  height: 24px;
}

.method-desc {
  font-size: 12px;
  color: #999;
}

.deposit-info {
  margin-top: 16px;
  padding: 12px;
  background: #fffbe6;
  border-radius: 8px;
  border: 1px solid #ffe58f;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  font-size: 14px;
}

.info-item.warning {
  color: #faad14;
}

.submit-section {
  margin-top: 24px;
}

.submit-btn {
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

.submit-btn:disabled {
  background: #d9d9d9;
  cursor: not-allowed;
}
</style>

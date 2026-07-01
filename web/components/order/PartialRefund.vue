<template>
  <div class="partial-refund">
    <div class="refund-header">
      <h2>申请退款</h2>
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
      <div class="info-row">
        <span class="label">实付金额：</span>
        <span class="value highlight">¥{{ formatAmount(totalPaid) }}</span>
      </div>
      <div v-if="alreadyRefunded > 0" class="info-row">
        <span class="label">已退金额：</span>
        <span class="value">¥{{ formatAmount(alreadyRefunded) }}</span>
      </div>
      <div class="info-row">
        <span class="label">可退金额：</span>
        <span class="value highlight">¥{{ formatAmount(maxRefundable) }}</span>
      </div>
    </div>

    <!-- Refund Amount -->
    <div class="refund-amount-section">
      <h3>退款金额</h3>
      <div class="amount-input">
        <span class="currency">¥</span>
        <input
          type="number"
          v-model.number="refundAmountYuan"
          :min="0.01"
          :max="maxRefundableYuan"
          step="0.01"
          placeholder="请输入退款金额"
          @input="onAmountChange"
        />
        <button class="max-btn" @click="setMaxAmount">最大金额</button>
      </div>
      <div class="amount-error" v-if="amountError">{{ amountError }}</div>
      <div class="amount-preview">
        <span>退款金额：</span>
        <span class="preview-value">¥{{ refundAmountYuan.toFixed(2) }}</span>
        <span v-if="refundPercentage > 0" class="percentage">
          ({{ refundPercentage.toFixed(0) }}%)
        </span>
      </div>
    </div>

    <!-- Refund Reason -->
    <div class="reason-section">
      <h3>退款原因</h3>
      <div class="reason-options">
        <label
          v-for="reason in reasonOptions"
          :key="reason.value"
          class="reason-option"
          :class="{ active: selectedReason === reason.value }"
        >
          <input
            type="radio"
            v-model="selectedReason"
            :value="reason.value"
          />
          <span>{{ reason.label }}</span>
        </label>
      </div>
    </div>

    <!-- Description -->
    <div class="description-section">
      <h3>补充说明</h3>
      <textarea
        v-model="description"
        placeholder="请详细说明退款原因（选填）"
        maxlength="500"
        rows="4"
      ></textarea>
      <div class="char-count">{{ description.length }}/500</div>
    </div>

    <!-- Attachments -->
    <div class="attachment-section">
      <h3>凭证上传</h3>
      <div class="upload-area">
        <input
          type="file"
          ref="fileInput"
          multiple
          accept="image/*,.pdf"
          @change="onFileSelect"
          style="display: none"
        />
        <button class="upload-btn" @click="triggerUpload">
          📎 上传凭证
        </button>
        <div class="file-list" v-if="attachments.length > 0">
          <div
            v-for="(file, index) in attachments"
            :key="index"
            class="file-item"
          >
            <span class="file-name">{{ file.name }}</span>
            <button class="remove-btn" @click="removeFile(index)">×</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Refund Rules Preview -->
    <div class="rules-preview">
      <h3>退改规则预览</h3>
      <div class="rule-info">
        <div class="rule-item">
          <span class="rule-label">距出发天数：</span>
          <span class="rule-value">{{ daysBeforeDeparture }}天</span>
        </div>
        <div class="rule-item">
          <span class="rule-label">匹配规则：</span>
          <span class="rule-value">{{ matchingRule || '无匹配规则' }}</span>
        </div>
        <div class="rule-item">
          <span class="rule-label">退款比例：</span>
          <span class="rule-value">{{ refundPercentage.toFixed(0) }}%</span>
        </div>
      </div>
    </div>

    <!-- Submit -->
    <div class="submit-section">
      <button
        class="submit-btn"
        :disabled="!isValid"
        @click="onSubmit"
      >
        提交退款申请
      </button>
      <div class="submit-note">
        提交后将由运营人员审核，审核通过后退款将原路退回
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  orderNo: string
  productName: string
  totalPaid: number          // 实付金额 (cents)
  alreadyRefunded: number    // 已退金额 (cents)
  daysBeforeDeparture: number
  matchingRule?: string
  refundPercentage?: number  // 0-100
}

const props = withDefaults(defineProps<Props>(), {
  matchingRule: '',
  refundPercentage: 100,
})

const emit = defineEmits<{
  (e: 'submit', data: RefundSubmitData): void
}>()

interface RefundSubmitData {
  orderNo: string
  refundAmount: number       // cents
  reason: string
  reasonCategory: string
  description: string
  attachments: string[]
}

// State
const refundAmountYuan = ref(0)
const selectedReason = ref('')
const description = ref('')
const attachments = ref<{ name: string; url: string }[]>([])
const fileInput = ref<HTMLInputElement | null>(null)

// Reason options
const reasonOptions = [
  { value: 'user_request', label: '个人原因' },
  { value: 'visa_rejected', label: '签证被拒' },
  { value: 'force_majeure', label: '不可抗力' },
  { value: 'supplier_issue', label: '供应商问题' },
]

// Computed
const maxRefundable = computed(() => {
  return props.totalPaid - props.alreadyRefunded
})

const maxRefundableYuan = computed(() => {
  return maxRefundable.value / 100
})

const refundAmount = computed(() => {
  return Math.round(refundAmountYuan.value * 100)
})

const amountError = computed(() => {
  if (refundAmountYuan.value <= 0) return '请输入退款金额'
  if (refundAmount.value > maxRefundable.value) return '退款金额不能超过可退金额'
  return ''
})

const isValid = computed(() => {
  return (
    refundAmount.value > 0 &&
    refundAmount.value <= maxRefundable.value &&
    selectedReason.value &&
    !amountError.value
  )
})

// Methods
const onAmountChange = () => {
  // Validate on input
}

const setMaxAmount = () => {
  refundAmountYuan.value = maxRefundableYuan.value
}

const formatAmount = (cents: number): string => {
  return (cents / 100).toFixed(2)
}

const triggerUpload = () => {
  fileInput.value?.click()
}

const onFileSelect = (event: Event) => {
  const input = event.target as HTMLInputElement
  if (input.files) {
    for (const file of Array.from(input.files)) {
      attachments.value.push({
        name: file.name,
        url: URL.createObjectURL(file),
      })
    }
  }
}

const removeFile = (index: number) => {
  attachments.value.splice(index, 1)
}

const onSubmit = () => {
  emit('submit', {
    orderNo: props.orderNo,
    refundAmount: refundAmount.value,
    reason: selectedReason.value,
    reasonCategory: selectedReason.value,
    description: description.value,
    attachments: attachments.value.map(a => a.url),
  })
}
</script>

<style scoped>
.partial-refund {
  max-width: 600px;
  margin: 0 auto;
  padding: 24px;
}

.refund-header {
  margin-bottom: 24px;
}

.order-info {
  background: #fafafa;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
}

.highlight {
  color: #ff4d4f;
  font-weight: 600;
}

.refund-amount-section {
  margin-bottom: 24px;
}

.amount-input {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.currency {
  font-size: 20px;
  font-weight: 600;
}

.amount-input input {
  flex: 1;
  padding: 12px;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  font-size: 16px;
}

.max-btn {
  padding: 12px 16px;
  background: #f0f0f0;
  border: none;
  border-radius: 8px;
  cursor: pointer;
}

.amount-error {
  color: #ff4d4f;
  font-size: 14px;
  margin-bottom: 8px;
}

.amount-preview {
  padding: 8px;
  background: #f0f7ff;
  border-radius: 4px;
}

.preview-value {
  font-weight: 600;
  color: #1890ff;
}

.percentage {
  color: #999;
  font-size: 14px;
}

.reason-section {
  margin-bottom: 24px;
}

.reason-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.reason-option {
  padding: 8px 16px;
  border: 2px solid #e8e8e8;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.2s;
}

.reason-option.active {
  border-color: #1890ff;
  background: #f0f7ff;
}

.reason-option input {
  display: none;
}

.description-section {
  margin-bottom: 24px;
}

textarea {
  width: 100%;
  padding: 12px;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  resize: vertical;
  font-family: inherit;
}

.char-count {
  text-align: right;
  color: #999;
  font-size: 12px;
}

.attachment-section {
  margin-bottom: 24px;
}

.upload-btn {
  padding: 12px 24px;
  background: #f0f0f0;
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  cursor: pointer;
  width: 100%;
}

.file-list {
  margin-top: 8px;
}

.file-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  background: #fafafa;
  border-radius: 4px;
  margin-bottom: 4px;
}

.remove-btn {
  background: #ff4d4f;
  color: white;
  border: none;
  border-radius: 50%;
  width: 24px;
  height: 24px;
  cursor: pointer;
}

.rules-preview {
  background: #fffbe6;
  border: 1px solid #ffe58f;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}

.rule-item {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
}

.submit-section {
  text-align: center;
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

.submit-note {
  margin-top: 8px;
  font-size: 12px;
  color: #999;
}
</style>

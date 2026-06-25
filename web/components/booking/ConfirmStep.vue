<template>
  <div class="confirm-step">
    <h3>确认订单</h3>

    <!-- Product info -->
    <div class="section">
      <h4>产品信息</h4>
      <div class="info-card">
        <div class="info-row">
          <span class="label">产品名称</span>
          <span class="value">{{ bookingData.productName || '产品' }}</span>
        </div>
        <div class="info-row">
          <span class="label">出发日期</span>
          <span class="value">{{ bookingData.departureDate }}</span>
        </div>
        <div class="info-row">
          <span class="label">人数</span>
          <span class="value">
            成人 {{ bookingData.adultCount }}
            <template v-if="bookingData.childCount > 0"> / 儿童 {{ bookingData.childCount }}</template>
            <template v-if="bookingData.infantCount > 0"> / 婴儿 {{ bookingData.infantCount }}</template>
          </span>
        </div>
      </div>
    </div>

    <!-- Traveller list -->
    <div class="section">
      <h4>出游人信息</h4>
      <div v-for="(t, i) in bookingData.travellers" :key="i" class="traveller-card">
        <span>{{ t.real_name }}</span>
        <span class="type-tag">{{ t.is_child ? '儿童' : t.is_infant ? '婴儿' : '成人' }}</span>
      </div>
    </div>

    <!-- Fee breakdown -->
    <div class="section">
      <h4>费用明细</h4>
      <div class="fee-card">
        <div class="fee-row">
          <span>成人 {{ bookingData.adultCount }} × ¥{{ formatAmount(bookingData.adultPrice) }}</span>
          <span>¥{{ formatAmount(bookingData.adultCount * bookingData.adultPrice) }}</span>
        </div>
        <div v-if="bookingData.childCount > 0" class="fee-row">
          <span>儿童 {{ bookingData.childCount }} × ¥{{ formatAmount(bookingData.childPrice) }}</span>
          <span>¥{{ formatAmount(bookingData.childCount * bookingData.childPrice) }}</span>
        </div>
        <div v-if="bookingData.infantCount > 0" class="fee-row">
          <span>婴儿 {{ bookingData.infantCount }} × ¥{{ formatAmount(bookingData.infantPrice) }}</span>
          <span>¥{{ formatAmount(bookingData.infantCount * bookingData.infantPrice) }}</span>
        </div>
        <div v-if="supplementAmount > 0" class="fee-row supplement">
          <span>单房差（成人数为奇数自动附加）</span>
          <span>¥{{ formatAmount(supplementAmount) }}</span>
        </div>
        <div v-for="addon in bookingData.addons" :key="addon.id" class="fee-row">
          <span>{{ addon.name }}</span>
          <span>¥{{ formatAmount(addon.price * (addon.quantity || 1)) }}</span>
        </div>
        <div class="fee-row total">
          <span>应付总额</span>
          <span class="total-price">¥{{ formatAmount(totalAmount) }}</span>
        </div>
      </div>
    </div>

    <!-- Cancellation policy -->
    <div class="section">
      <h4>退改政策</h4>
      <div class="policy-card">
        <p>请在下单前仔细阅读退改政策。订单支付后如需退改，将按照产品的退改规则执行。</p>
      </div>
    </div>

    <!-- Agreement -->
    <div class="section">
      <el-checkbox v-model="agreed">
        我已阅读并同意《退改政策》和《预订须知》
      </el-checkbox>
    </div>

    <div class="actions">
      <el-button @click="emit('back')">上一步</el-button>
      <el-button type="primary" :disabled="!agreed" :loading="submitting" @click="handleSubmit">
        提交订单
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { formatAmount } from '~/shared/utils/amount'

const props = defineProps<{
  bookingData: any
}>()

const emit = defineEmits<{
  submit: []
  back: []
}>()

const agreed = ref(false)
const submitting = ref(false)

const supplementAmount = computed(() => {
  const d = props.bookingData
  if (d.adultCount > 0 && d.adultCount % 2 !== 0) {
    return d.singleSupplement
  }
  return 0
})

const totalAmount = computed(() => {
  const d = props.bookingData
  const adultTotal = d.adultCount * d.adultPrice
  const childTotal = d.childCount * d.childPrice
  const infantTotal = d.infantCount * d.infantPrice
  const addonTotal = (d.addons || []).reduce((sum: number, a: any) => sum + (a.price * (a.quantity || 1)), 0)
  return adultTotal + childTotal + infantTotal + supplementAmount.value + addonTotal
})

async function handleSubmit() {
  if (!agreed.value) return
  submitting.value = true
  try {
    emit('submit')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.confirm-step h3 {
  margin-bottom: 20px;
}

.section {
  margin-bottom: 24px;
}

.section h4 {
  margin-bottom: 12px;
  font-size: 16px;
}

.info-card, .fee-card, .policy-card {
  background: #fafafa;
  padding: 16px;
  border-radius: 8px;
}

.info-row, .fee-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.info-row .label {
  color: #666;
}

.fee-row.supplement {
  color: #faad14;
}

.fee-row.total {
  border-top: 1px solid #e8e8e8;
  padding-top: 8px;
  margin-top: 8px;
  font-weight: bold;
}

.total-price {
  color: #ff4d4f;
  font-size: 18px;
}

.traveller-card {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  padding: 8px 12px;
  margin: 0 8px 8px 0;
}

.type-tag {
  font-size: 12px;
  background: #ecf5ff;
  color: #409eff;
  padding: 2px 6px;
  border-radius: 4px;
}

.policy-card p {
  color: #666;
  font-size: 14px;
  line-height: 1.6;
}

.actions {
  display: flex;
  justify-content: space-between;
  margin-top: 24px;
}
</style>

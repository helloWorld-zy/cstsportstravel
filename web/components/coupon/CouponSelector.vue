<template>
  <div class="coupon-selector">
    <div class="selector-header" @click="toggleExpand">
      <span class="label">优惠券</span>
      <span class="selected-info" v-if="selectedCoupon">
        已选: {{ selectedCoupon.coupon_name }} (省¥{{ selectedCoupon.discount_amount }})
      </span>
      <span class="selected-info no-coupon" v-else-if="availableCoupons.length === 0">
        暂无可用优惠券
      </span>
      <span class="selected-info" v-else>
        {{ availableCoupons.length }}张可用
      </span>
      <span class="expand-icon" :class="{ expanded }">▼</span>
    </div>

    <div class="coupon-dropdown" v-show="expanded">
      <div class="coupon-option no-coupon-option" @click="selectNone">
        <span>不使用优惠券</span>
        <span class="check" v-if="!selectedCoupon">✓</span>
      </div>
      <div
        v-for="coupon in availableCoupons"
        :key="coupon.claim_id"
        class="coupon-option"
        :class="{ selected: selectedCoupon?.claim_id === coupon.claim_id }"
        @click="selectCoupon(coupon)"
      >
        <div class="option-left">
          <div class="coupon-type-badge" :class="coupon.coupon_type">
            {{ couponTypeLabel(coupon.coupon_type) }}
          </div>
          <div class="coupon-info">
            <div class="coupon-name">{{ coupon.coupon_name }}</div>
            <div class="coupon-condition" v-if="coupon.min_consumption > 0">
              满{{ coupon.min_consumption }}减{{ coupon.discount_amount }}
            </div>
            <div class="coupon-validity" v-if="coupon.valid_to">
              {{ formatDate(coupon.valid_to) }}到期
            </div>
          </div>
        </div>
        <div class="option-right">
          <span class="discount-amount">-¥{{ coupon.discount_amount.toFixed(2) }}</span>
          <span class="check" v-if="selectedCoupon?.claim_id === coupon.claim_id">✓</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

interface AvailableCoupon {
  claim_id: number
  coupon_id: number
  coupon_name: string
  coupon_type: string
  discount_amount: number
  min_consumption: number
  valid_to?: string
}

const props = defineProps<{
  productId?: number
  orderAmount: number
}>()

const emit = defineEmits<{
  (e: 'select', coupon: AvailableCoupon | null): void
  (e: 'discount-change', amount: number): void
}>()

const expanded = ref(false)
const availableCoupons = ref<AvailableCoupon[]>([])
const selectedCoupon = ref<AvailableCoupon | null>(null)

const couponTypeLabel = (type: string): string => {
  const labels: Record<string, string> = {
    full_reduction: '满减',
    discount: '折扣',
    cash: '现金',
    exchange: '兑换',
  }
  return labels[type] || type
}

const formatDate = (dateStr?: string): string => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

const toggleExpand = () => {
  expanded.value = !expanded.value
}

const selectCoupon = (coupon: AvailableCoupon) => {
  selectedCoupon.value = coupon
  expanded.value = false
  emit('select', coupon)
  emit('discount-change', coupon.discount_amount)
}

const selectNone = () => {
  selectedCoupon.value = null
  expanded.value = false
  emit('select', null)
  emit('discount-change', 0)
}

const loadAvailableCoupons = async () => {
  try {
    const res = await $fetch('/api/v2/coupons/available', {
      params: {
        productId: props.productId,
        orderAmount: props.orderAmount,
      },
    })
    if (res.code === 200) {
      availableCoupons.value = res.data
      // Auto-select best coupon
      if (availableCoupons.value.length > 0 && !selectedCoupon.value) {
        selectCoupon(availableCoupons.value[0])
      }
    }
  } catch (err) {
    console.error('Failed to load available coupons:', err)
  }
}

watch(() => props.orderAmount, () => {
  loadAvailableCoupons()
})

onMounted(() => {
  if (props.orderAmount > 0) {
    loadAvailableCoupons()
  }
})
</script>

<style scoped>
.coupon-selector {
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  overflow: hidden;
}

.selector-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  background: #fafafa;
}

.selector-header .label {
  font-weight: 500;
  margin-right: 12px;
}

.selected-info {
  flex: 1;
  color: #ff6b6b;
  font-size: 14px;
}

.selected-info.no-coupon {
  color: #999;
}

.expand-icon {
  font-size: 12px;
  transition: transform 0.2s;
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

.coupon-dropdown {
  border-top: 1px solid #e8e8e8;
  max-height: 300px;
  overflow-y: auto;
}

.coupon-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f5f5f5;
  transition: background 0.2s;
}

.coupon-option:hover {
  background: #fff8f8;
}

.coupon-option.selected {
  background: #fff3f3;
}

.no-coupon-option {
  color: #999;
}

.option-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.coupon-type-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: #fff;
}

.coupon-type-badge.full_reduction {
  background: #ff6b6b;
}

.coupon-type-badge.discount {
  background: #ffa726;
}

.coupon-type-badge.cash {
  background: #66bb6a;
}

.coupon-type-badge.exchange {
  background: #42a5f5;
}

.coupon-info .coupon-name {
  font-size: 14px;
  font-weight: 500;
}

.coupon-info .coupon-condition,
.coupon-info .coupon-validity {
  font-size: 12px;
  color: #999;
}

.option-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.discount-amount {
  color: #ff6b6b;
  font-weight: 500;
  font-size: 15px;
}

.check {
  color: #ff6b6b;
  font-weight: bold;
}
</style>

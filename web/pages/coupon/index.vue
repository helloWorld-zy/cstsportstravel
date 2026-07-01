<template>
  <div class="coupon-center-page">
    <!-- Header -->
    <div class="page-header">
      <h1>领券中心</h1>
      <p>精选优惠券，领取后下单更划算</p>
    </div>

    <!-- Coupon List -->
    <div class="coupon-list" v-if="coupons.length > 0">
      <div
        v-for="coupon in coupons"
        :key="coupon.id"
        class="coupon-card"
        :class="{ claimed: claimedIds.has(coupon.id) }"
      >
        <div class="coupon-left">
          <div class="coupon-value">
            <template v-if="coupon.coupon_type === 'full_reduction'">
              <span class="currency">¥</span>
              <span class="amount">{{ coupon.discount_amount }}</span>
            </template>
            <template v-else-if="coupon.coupon_type === 'discount'">
              <span class="amount">{{ coupon.discount_rate / 10 }}</span>
              <span class="unit">折</span>
            </template>
            <template v-else-if="coupon.coupon_type === 'cash'">
              <span class="currency">¥</span>
              <span class="amount">{{ coupon.discount_amount }}</span>
            </template>
            <template v-else-if="coupon.coupon_type === 'exchange'">
              <span class="amount">兑换</span>
            </template>
          </div>
          <div class="coupon-condition" v-if="coupon.min_consumption > 0">
            满{{ coupon.min_consumption }}可用
          </div>
          <div class="coupon-condition" v-else>
            无门槛
          </div>
        </div>

        <div class="coupon-right">
          <div class="coupon-name">{{ coupon.coupon_name }}</div>
          <div class="coupon-type-tag">{{ couponTypeLabel(coupon.coupon_type) }}</div>
          <div class="coupon-validity">
            <template v-if="coupon.validity_type === 'fixed' && coupon.valid_to">
              有效期至 {{ formatDate(coupon.valid_to) }}
            </template>
            <template v-else-if="coupon.validity_type === 'relative' && coupon.valid_days">
              领取后{{ coupon.valid_days }}天内有效
            </template>
          </div>
          <div class="coupon-stock">
            剩余 {{ coupon.total_stock - coupon.claimed_count }} 张
          </div>
          <button
            class="claim-btn"
            :disabled="claimedIds.has(coupon.id) || claiming"
            @click="claimCoupon(coupon.id)"
          >
            {{ claimedIds.has(coupon.id) ? '已领取' : '立即领取' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="!loading" class="empty-state">
      <p>暂无可领取的优惠券</p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      加载中...
    </div>

    <!-- Pagination -->
    <div class="pagination" v-if="total > pageSize">
      <button :disabled="page <= 1" @click="changePage(page - 1)">上一页</button>
      <span>{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
      <button :disabled="page >= Math.ceil(total / pageSize)" @click="changePage(page + 1)">下一页</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Coupon {
  id: number
  coupon_name: string
  coupon_type: string
  discount_amount: number
  discount_rate: number
  min_consumption: number
  total_stock: number
  claimed_count: number
  validity_type: string
  valid_to?: string
  valid_days?: number
}

const coupons = ref<Coupon[]>([])
const claimedIds = ref<Set<number>>(new Set())
const loading = ref(false)
const claiming = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const couponTypeLabel = (type: string): string => {
  const labels: Record<string, string> = {
    full_reduction: '满减券',
    discount: '折扣券',
    cash: '现金券',
    exchange: '兑换券',
  }
  return labels[type] || type
}

const formatDate = (dateStr?: string): string => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const loadCoupons = async () => {
  loading.value = true
  try {
    const res = await $fetch('/api/v2/coupons/center', {
      params: { page: page.value, pageSize: pageSize.value },
    })
    if (res.code === 200) {
      coupons.value = res.data.list
      total.value = res.data.total
    }
  } catch (err) {
    console.error('Failed to load coupons:', err)
  } finally {
    loading.value = false
  }
}

const claimCoupon = async (couponId: number) => {
  if (claiming.value || claimedIds.value.has(couponId)) return
  claiming.value = true
  try {
    const res = await $fetch(`/api/v2/coupons/${couponId}/claim`, { method: 'POST' })
    if (res.code === 200) {
      claimedIds.value.add(couponId)
      // Update stock display
      const coupon = coupons.value.find(c => c.id === couponId)
      if (coupon) {
        coupon.claimed_count++
      }
    }
  } catch (err: any) {
    const message = err?.data?.message || '领取失败'
    alert(message)
  } finally {
    claiming.value = false
  }
}

const changePage = (newPage: number) => {
  page.value = newPage
  loadCoupons()
}

onMounted(() => {
  loadCoupons()
})
</script>

<style scoped>
.coupon-center-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 24px 16px;
}

.page-header {
  text-align: center;
  margin-bottom: 32px;
}

.page-header h1 {
  font-size: 28px;
  margin-bottom: 8px;
}

.page-header p {
  color: #666;
}

.coupon-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.coupon-card {
  display: flex;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
  transition: box-shadow 0.2s;
}

.coupon-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.coupon-card.claimed {
  opacity: 0.6;
}

.coupon-left {
  width: 140px;
  background: linear-gradient(135deg, #ff6b6b, #ee5a24);
  color: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.coupon-value {
  display: flex;
  align-items: baseline;
}

.coupon-value .currency {
  font-size: 16px;
}

.coupon-value .amount {
  font-size: 32px;
  font-weight: bold;
}

.coupon-value .unit {
  font-size: 16px;
  margin-left: 2px;
}

.coupon-condition {
  font-size: 12px;
  margin-top: 4px;
  opacity: 0.9;
}

.coupon-right {
  flex: 1;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.coupon-name {
  font-size: 16px;
  font-weight: 500;
}

.coupon-type-tag {
  display: inline-block;
  background: #fff3e0;
  color: #e65100;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  width: fit-content;
}

.coupon-validity,
.coupon-stock {
  font-size: 13px;
  color: #999;
}

.claim-btn {
  align-self: flex-end;
  padding: 8px 24px;
  background: #ff6b6b;
  color: #fff;
  border: none;
  border-radius: 20px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s;
}

.claim-btn:hover:not(:disabled) {
  background: #ee5a24;
}

.claim-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.empty-state,
.loading-state {
  text-align: center;
  padding: 48px;
  color: #999;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-top: 24px;
}

.pagination button {
  padding: 8px 16px;
  border: 1px solid #ddd;
  background: #fff;
  border-radius: 4px;
  cursor: pointer;
}

.pagination button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>

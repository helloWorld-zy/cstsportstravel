<template>
  <div class="my-coupons-page">
    <div class="page-header">
      <h1>我的优惠券</h1>
    </div>

    <!-- Status Tabs -->
    <div class="status-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        class="tab-btn"
        :class="{ active: activeTab === tab.value }"
        @click="switchTab(tab.value)"
      >
        {{ tab.label }}
        <span v-if="tabCounts[tab.value]" class="tab-count">{{ tabCounts[tab.value] }}</span>
      </button>
    </div>

    <!-- Coupon List -->
    <div class="coupon-list" v-if="coupons.length > 0">
      <div
        v-for="claim in coupons"
        :key="claim.id"
        class="coupon-card"
        :class="{ expired: claim.status === 'expired', used: claim.status === 'used' }"
      >
        <div class="coupon-left">
          <div class="coupon-value">
            <span class="currency">¥</span>
            <span class="amount">--</span>
          </div>
          <div class="coupon-condition">
            {{ claim.status === 'available' ? '可使用' : statusLabel(claim.status) }}
          </div>
        </div>
        <div class="coupon-right">
          <div class="coupon-name">优惠券 #{{ claim.coupon_id }}</div>
          <div class="coupon-time">
            领取时间: {{ formatDate(claim.claimed_at) }}
          </div>
          <div class="coupon-time" v-if="claim.used_at">
            使用时间: {{ formatDate(claim.used_at) }}
          </div>
          <div class="coupon-time" v-if="claim.expired_at">
            过期时间: {{ formatDate(claim.expired_at) }}
          </div>
          <div class="coupon-status-badge" :class="claim.status">
            {{ statusLabel(claim.status) }}
          </div>
          <button
            v-if="claim.status === 'available'"
            class="use-btn"
            @click="goToOrder(claim)"
          >
            去使用
          </button>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="!loading" class="empty-state">
      <p>{{ emptyText }}</p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">加载中...</div>

    <!-- Pagination -->
    <div class="pagination" v-if="total > pageSize">
      <button :disabled="page <= 1" @click="changePage(page - 1)">上一页</button>
      <span>{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
      <button :disabled="page >= Math.ceil(total / pageSize)" @click="changePage(page + 1)">下一页</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

interface CouponClaim {
  id: number
  coupon_id: number
  status: string
  claimed_at: string
  used_at?: string
  expired_at?: string
  returned_at?: string
}

const router = useRouter()
const coupons = ref<CouponClaim[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const activeTab = ref('available')

const tabs = [
  { label: '待使用', value: 'available' },
  { label: '已使用', value: 'used' },
  { label: '已过期', value: 'expired' },
]

const tabCounts = ref<Record<string, number>>({
  available: 0,
  used: 0,
  expired: 0,
})

const emptyText = computed(() => {
  const texts: Record<string, string> = {
    available: '暂无待使用的优惠券',
    used: '暂无已使用的优惠券',
    expired: '暂无已过期的优惠券',
  }
  return texts[activeTab.value] || '暂无优惠券'
})

const statusLabel = (status: string): string => {
  const labels: Record<string, string> = {
    available: '待使用',
    occupied: '已占用',
    used: '已使用',
    expired: '已过期',
    returned: '已退还',
    voided: '已作废',
  }
  return labels[status] || status
}

const formatDate = (dateStr?: string): string => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const loadCoupons = async () => {
  loading.value = true
  try {
    const res = await $fetch('/api/v2/coupons/mine', {
      params: {
        status: activeTab.value,
        page: page.value,
        pageSize: pageSize.value,
      },
    })
    if (res.code === 200) {
      coupons.value = res.data.list
      total.value = res.data.total
      tabCounts.value[activeTab.value] = res.data.total
    }
  } catch (err) {
    console.error('Failed to load coupons:', err)
  } finally {
    loading.value = false
  }
}

const switchTab = (tab: string) => {
  activeTab.value = tab
  page.value = 1
  loadCoupons()
}

const changePage = (newPage: number) => {
  page.value = newPage
  loadCoupons()
}

const goToOrder = (claim: CouponClaim) => {
  router.push(`/products?couponId=${claim.id}`)
}

onMounted(() => {
  loadCoupons()
})
</script>

<style scoped>
.my-coupons-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 24px 16px;
}

.page-header h1 {
  font-size: 24px;
  margin-bottom: 24px;
}

.status-tabs {
  display: flex;
  gap: 0;
  border-bottom: 2px solid #e8e8e8;
  margin-bottom: 24px;
}

.tab-btn {
  padding: 12px 24px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  cursor: pointer;
  font-size: 15px;
  color: #666;
  transition: all 0.2s;
}

.tab-btn.active {
  color: #ff6b6b;
  border-bottom-color: #ff6b6b;
}

.tab-count {
  display: inline-block;
  background: #ff6b6b;
  color: #fff;
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 10px;
  margin-left: 4px;
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
}

.coupon-card.expired,
.coupon-card.used {
  opacity: 0.5;
}

.coupon-left {
  width: 120px;
  background: linear-gradient(135deg, #ff6b6b, #ee5a24);
  color: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.coupon-card.expired .coupon-left,
.coupon-card.used .coupon-left {
  background: #ccc;
}

.coupon-value {
  display: flex;
  align-items: baseline;
}

.coupon-value .currency {
  font-size: 14px;
}

.coupon-value .amount {
  font-size: 28px;
  font-weight: bold;
}

.coupon-condition {
  font-size: 12px;
  margin-top: 4px;
}

.coupon-right {
  flex: 1;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.coupon-name {
  font-size: 15px;
  font-weight: 500;
}

.coupon-time {
  font-size: 13px;
  color: #999;
}

.coupon-status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  width: fit-content;
}

.coupon-status-badge.available {
  background: #e8f5e9;
  color: #2e7d32;
}

.coupon-status-badge.used {
  background: #e0e0e0;
  color: #616161;
}

.coupon-status-badge.expired {
  background: #ffebee;
  color: #c62828;
}

.coupon-status-badge.returned {
  background: #fff3e0;
  color: #e65100;
}

.use-btn {
  align-self: flex-end;
  padding: 6px 20px;
  background: #ff6b6b;
  color: #fff;
  border: none;
  border-radius: 20px;
  cursor: pointer;
  font-size: 13px;
}

.use-btn:hover {
  background: #ee5a24;
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

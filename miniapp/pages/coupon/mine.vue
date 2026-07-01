<template>
  <view class="my-coupons">
    <!-- Tabs -->
    <view class="tabs">
      <view
        v-for="tab in tabs"
        :key="tab.value"
        class="tab"
        :class="{ active: activeTab === tab.value }"
        @tap="switchTab(tab.value)"
      >
        <text>{{ tab.label }}</text>
      </view>
    </view>

    <!-- Coupon List -->
    <view class="coupon-list" v-if="coupons.length > 0">
      <view
        v-for="claim in coupons"
        :key="claim.id"
        class="coupon-card"
        :class="claim.status"
      >
        <view class="coupon-left">
          <text class="coupon-id">#{{ claim.coupon_id }}</text>
          <text class="status-text">{{ statusLabel(claim.status) }}</text>
        </view>
        <view class="coupon-right">
          <text class="time">领取: {{ formatDate(claim.claimed_at) }}</text>
          <text class="time" v-if="claim.used_at">使用: {{ formatDate(claim.used_at) }}</text>
          <text class="time" v-if="claim.expired_at">过期: {{ formatDate(claim.expired_at) }}</text>
          <view class="status-badge" :class="claim.status">
            <text>{{ statusLabel(claim.status) }}</text>
          </view>
        </view>
      </view>
    </view>

    <view v-else class="empty-state">
      <text>{{ emptyText }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const coupons = ref<any[]>([])
const activeTab = ref('available')

const tabs = [
  { label: '待使用', value: 'available' },
  { label: '已使用', value: 'used' },
  { label: '已过期', value: 'expired' },
]

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

const formatDate = (dateStr: string): string => {
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const loadCoupons = async () => {
  try {
    const res = await uni.request({
      url: `/api/v2/coupons/mine?status=${activeTab.value}`,
      method: 'GET',
    })
    if (res.data?.code === 200) {
      coupons.value = res.data.data.list
    }
  } catch (err) {
    console.error('Failed to load coupons:', err)
  }
}

const switchTab = (tab: string) => {
  activeTab.value = tab
  loadCoupons()
}

onMounted(() => {
  loadCoupons()
})
</script>

<style scoped>
.my-coupons {
  padding: 24rpx;
}

.tabs {
  display: flex;
  border-bottom: 4rpx solid #e8e8e8;
  margin-bottom: 24rpx;
}

.tab {
  flex: 1;
  text-align: center;
  padding: 20rpx 0;
  font-size: 28rpx;
  color: #666;
  border-bottom: 4rpx solid transparent;
}

.tab.active {
  color: #ff6b6b;
  border-bottom-color: #ff6b6b;
}

.coupon-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.coupon-card {
  display: flex;
  border-radius: 16rpx;
  overflow: hidden;
  background: #fff;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.06);
}

.coupon-card.expired,
.coupon-card.used {
  opacity: 0.5;
}

.coupon-left {
  width: 180rpx;
  background: linear-gradient(135deg, #ff6b6b, #ee5a24);
  color: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24rpx;
}

.coupon-card.expired .coupon-left,
.coupon-card.used .coupon-left {
  background: #ccc;
}

.coupon-id {
  font-size: 28rpx;
  font-weight: bold;
}

.status-text {
  font-size: 22rpx;
  margin-top: 8rpx;
}

.coupon-right {
  flex: 1;
  padding: 24rpx;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.time {
  font-size: 24rpx;
  color: #999;
}

.status-badge {
  display: inline-block;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  width: fit-content;
  margin-top: 8rpx;
}

.status-badge.available {
  background: #e8f5e9;
  color: #2e7d32;
}

.status-badge.used {
  background: #e0e0e0;
  color: #616161;
}

.status-badge.expired {
  background: #ffebee;
  color: #c62828;
}

.empty-state {
  text-align: center;
  padding: 80rpx;
  color: #999;
}
</style>

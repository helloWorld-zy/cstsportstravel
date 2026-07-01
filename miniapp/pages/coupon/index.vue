<template>
  <view class="coupon-center">
    <view class="header">
      <text class="title">领券中心</text>
    </view>

    <view class="coupon-list" v-if="coupons.length > 0">
      <view
        v-for="coupon in coupons"
        :key="coupon.id"
        class="coupon-card"
        :class="{ claimed: claimedIds.includes(coupon.id) }"
      >
        <view class="coupon-left">
          <view class="coupon-value">
            <text v-if="coupon.coupon_type === 'full_reduction'" class="amount">
              ¥{{ coupon.discount_amount }}
            </text>
            <text v-else-if="coupon.coupon_type === 'discount'" class="amount">
              {{ coupon.discount_rate / 10 }}折
            </text>
            <text v-else-if="coupon.coupon_type === 'cash'" class="amount">
              ¥{{ coupon.discount_amount }}
            </text>
            <text v-else class="amount">兑换</text>
          </view>
          <text class="condition" v-if="coupon.min_consumption > 0">
            满{{ coupon.min_consumption }}可用
          </text>
          <text class="condition" v-else>无门槛</text>
        </view>

        <view class="coupon-right">
          <text class="coupon-name">{{ coupon.coupon_name }}</text>
          <text class="validity" v-if="coupon.validity_type === 'fixed' && coupon.valid_to">
            有效期至 {{ formatDate(coupon.valid_to) }}
          </text>
          <text class="validity" v-else-if="coupon.valid_days">
            领取后{{ coupon.valid_days }}天内有效
          </text>
          <text class="stock">剩余 {{ coupon.total_stock - coupon.claimed_count }} 张</text>
          <view
            class="claim-btn"
            :class="{ disabled: claimedIds.includes(coupon.id) }"
            @tap="claimCoupon(coupon)"
          >
            <text>{{ claimedIds.includes(coupon.id) ? '已领取' : '立即领取' }}</text>
          </view>
        </view>
      </view>
    </view>

    <view v-else class="empty-state">
      <text>暂无可领取的优惠券</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const coupons = ref<any[]>([])
const claimedIds = ref<number[]>([])

const formatDate = (dateStr: string): string => {
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

const loadCoupons = async () => {
  try {
    const res = await uni.request({
      url: '/api/v2/coupons/center',
      method: 'GET',
    })
    if (res.data?.code === 200) {
      coupons.value = res.data.data.list
    }
  } catch (err) {
    console.error('Failed to load coupons:', err)
  }
}

const claimCoupon = async (coupon: any) => {
  if (claimedIds.value.includes(coupon.id)) return
  try {
    const res = await uni.request({
      url: `/api/v2/coupons/${coupon.id}/claim`,
      method: 'POST',
    })
    if (res.data?.code === 200) {
      claimedIds.value.push(coupon.id)
      uni.showToast({ title: '领取成功', icon: 'success' })
    } else {
      uni.showToast({ title: res.data?.message || '领取失败', icon: 'none' })
    }
  } catch (err: any) {
    uni.showToast({ title: err?.data?.message || '领取失败', icon: 'none' })
  }
}

onMounted(() => {
  loadCoupons()
})
</script>

<style scoped>
.coupon-center {
  padding: 24rpx;
}

.header {
  text-align: center;
  margin-bottom: 32rpx;
}

.title {
  font-size: 36rpx;
  font-weight: bold;
}

.coupon-list {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.coupon-card {
  display: flex;
  border-radius: 16rpx;
  overflow: hidden;
  background: #fff;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.08);
}

.coupon-card.claimed {
  opacity: 0.6;
}

.coupon-left {
  width: 220rpx;
  background: linear-gradient(135deg, #ff6b6b, #ee5a24);
  color: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24rpx;
}

.amount {
  font-size: 48rpx;
  font-weight: bold;
}

.condition {
  font-size: 22rpx;
  margin-top: 8rpx;
  opacity: 0.9;
}

.coupon-right {
  flex: 1;
  padding: 24rpx;
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.coupon-name {
  font-size: 28rpx;
  font-weight: 500;
}

.validity,
.stock {
  font-size: 24rpx;
  color: #999;
}

.claim-btn {
  align-self: flex-end;
  padding: 12rpx 32rpx;
  background: #ff6b6b;
  color: #fff;
  border-radius: 32rpx;
  font-size: 24rpx;
}

.claim-btn.disabled {
  background: #ccc;
}

.empty-state {
  text-align: center;
  padding: 80rpx;
  color: #999;
}
</style>

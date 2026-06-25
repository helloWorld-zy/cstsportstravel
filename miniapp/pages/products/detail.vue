<template>
  <view class="product-detail">
    <!-- Loading -->
    <view v-if="isLoading" class="loading">
      <text>加载中...</text>
    </view>

    <template v-else-if="product">
      <!-- Image Swiper -->
      <swiper class="image-swiper" indicator-dots autoplay circular>
        <swiper-item v-for="(img, idx) in productImages" :key="idx">
          <image :src="img" mode="aspectFill" class="swiper-img" />
        </swiper-item>
      </swiper>

      <!-- Product Info -->
      <view class="info-section">
        <text class="product-name">{{ product.product_name }}</text>
        <view class="product-meta">
          <text v-if="product.product_grade" class="meta-tag">{{ gradeLabel }}</text>
          <text class="meta-tag">{{ product.days }}天{{ product.nights }}晚</text>
        </view>
        <view class="price-row">
          <text class="price-label">起</text>
          <text class="price-value">¥{{ product.min_price }}</text>
          <text class="price-unit">/人</text>
        </view>
        <text v-if="product.summary" class="summary">{{ product.summary }}</text>
      </view>

      <!-- Itinerary -->
      <view v-if="product.itinerary?.length" class="section">
        <text class="section-title">行程详情</text>
        <view v-for="day in product.itinerary" :key="day.day_no" class="itinerary-day">
          <view class="day-header">
            <text class="day-badge">Day {{ day.day_no }}</text>
            <text class="day-title">{{ day.title }}</text>
          </view>
          <text v-if="day.description" class="day-desc">{{ day.description }}</text>
          <view v-if="day.hotel" class="day-detail">
            <text class="detail-icon">🏨</text>
            <text>{{ day.hotel }}</text>
          </view>
        </view>
      </view>

      <!-- Fees -->
      <view class="section">
        <text class="section-title">费用说明</text>
        <view v-if="product.fee_included" class="fee-block">
          <text class="fee-label included">费用包含</text>
          <text class="fee-content">{{ product.fee_included }}</text>
        </view>
        <view v-if="product.fee_excluded" class="fee-block">
          <text class="fee-label excluded">费用不含</text>
          <text class="fee-content">{{ product.fee_excluded }}</text>
        </view>
      </view>

      <!-- Cancellation Policy -->
      <view v-if="product.cancellation_rules?.length" class="section">
        <text class="section-title">退改政策</text>
        <view v-for="rule in product.cancellation_rules" :key="rule.id" class="rule-row">
          <text class="rule-name">{{ rule.rule_name }}</text>
          <text class="rule-pct">退{{ rule.refund_percentage }}%</text>
        </view>
      </view>

      <!-- Reviews Summary -->
      <view v-if="product.review_summary" class="section">
        <text class="section-title">用户评价</text>
        <view class="review-summary">
          <text class="rating-num">{{ product.review_summary.average_rating.toFixed(1) }}</text>
          <text class="rating-label">分 · {{ product.review_summary.total_count }}条评价</text>
        </view>
      </view>

      <!-- Bottom Bar -->
      <view class="bottom-bar">
        <view class="bar-price">
          <text class="bar-price-label">起</text>
          <text class="bar-price-value">¥{{ product.min_price }}</text>
        </view>
        <button class="book-btn" @click="goToBooking">立即预订</button>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/shared/api/request'

interface ItineraryDay {
  day_no: number
  title: string
  description?: string
  hotel?: string
}

interface CancellationRule {
  id: number
  rule_name: string
  refund_percentage: number
}

interface ReviewSummary {
  total_count: number
  average_rating: number
}

interface ProductDetail {
  id: number
  product_name: string
  cover_image?: string
  images?: string[]
  days: number
  nights: number
  min_price: number
  product_grade?: string
  summary?: string
  fee_included?: string
  fee_excluded?: string
  itinerary?: ItineraryDay[]
  cancellation_rules?: CancellationRule[]
  review_summary?: ReviewSummary
}

const product = ref<ProductDetail | null>(null)
const isLoading = ref(true)
const productId = ref('')

const productImages = computed(() => {
  if (!product.value) return []
  const imgs = product.value.images || []
  if (product.value.cover_image) return [product.value.cover_image, ...imgs]
  return imgs.length ? imgs : ['/static/images/default-product.png']
})

const gradeLabel = computed(() => {
  const map: Record<string, string> = { standard: '经济', comfort: '舒适', luxury: '豪华' }
  return map[product.value?.product_grade || ''] || ''
})

function goToBooking() {
  uni.navigateTo({ url: `/pages/booking/index?productId=${productId.value}` })
}

onMounted(async () => {
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1]
  productId.value = currentPage?.options?.id || ''

  if (!productId.value) {
    isLoading.value = false
    return
  }

  try {
    product.value = await api.get<ProductDetail>(`/products/${productId.value}`)
  } catch (e) {
    console.error('Failed to load product', e)
  } finally {
    isLoading.value = false
  }
})
</script>

<style scoped>
.product-detail {
  background: #f5f5f5;
  min-height: 100vh;
  padding-bottom: 120rpx;
}

.image-swiper {
  height: 400rpx;
}

.swiper-img {
  width: 100%;
  height: 400rpx;
}

.info-section {
  background: #fff;
  padding: 24rpx;
}

.product-name {
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 12rpx;
}

.product-meta {
  display: flex;
  gap: 12rpx;
  margin-bottom: 12rpx;
}

.meta-tag {
  font-size: 22rpx;
  color: #666;
  padding: 4rpx 12rpx;
  background: #f5f5f5;
  border-radius: 4rpx;
}

.price-row {
  display: flex;
  align-items: baseline;
  gap: 4rpx;
  margin-bottom: 12rpx;
}

.price-label {
  font-size: 24rpx;
  color: #999;
}

.price-value {
  font-size: 48rpx;
  font-weight: bold;
  color: #ff5722;
}

.price-unit {
  font-size: 24rpx;
  color: #999;
}

.summary {
  font-size: 28rpx;
  color: #666;
  line-height: 1.6;
}

.section {
  background: #fff;
  margin-top: 16rpx;
  padding: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 16rpx;
  display: block;
}

.itinerary-day {
  margin-bottom: 24rpx;
  padding-bottom: 24rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.day-header {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 8rpx;
}

.day-badge {
  background: #ff5722;
  color: #fff;
  padding: 4rpx 16rpx;
  border-radius: 16rpx;
  font-size: 22rpx;
  font-weight: 600;
}

.day-title {
  font-size: 28rpx;
  font-weight: 500;
  color: #333;
}

.day-desc {
  font-size: 26rpx;
  color: #666;
  line-height: 1.6;
  margin-bottom: 8rpx;
}

.day-detail {
  display: flex;
  align-items: center;
  gap: 8rpx;
  font-size: 24rpx;
  color: #666;
}

.detail-icon {
  font-size: 28rpx;
}

.fee-block {
  margin-bottom: 16rpx;
}

.fee-label {
  font-size: 28rpx;
  font-weight: 500;
  display: block;
  margin-bottom: 8rpx;
}

.fee-label.included {
  color: #4caf50;
}

.fee-label.excluded {
  color: #ff5722;
}

.fee-content {
  font-size: 26rpx;
  color: #666;
  line-height: 1.6;
}

.rule-row {
  display: flex;
  justify-content: space-between;
  padding: 12rpx 0;
  border-bottom: 1rpx solid #f5f5f5;
}

.rule-name {
  font-size: 26rpx;
  color: #333;
}

.rule-pct {
  font-size: 26rpx;
  color: #ff5722;
  font-weight: 500;
}

.review-summary {
  display: flex;
  align-items: baseline;
  gap: 8rpx;
}

.rating-num {
  font-size: 48rpx;
  font-weight: bold;
  color: #ff9800;
}

.rating-label {
  font-size: 26rpx;
  color: #666;
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 24rpx;
  background: #fff;
  box-shadow: 0 -4rpx 16rpx rgba(0, 0, 0, 0.1);
}

.bar-price-value {
  font-size: 40rpx;
  font-weight: bold;
  color: #ff5722;
}

.bar-price-label {
  font-size: 24rpx;
  color: #999;
}

.book-btn {
  padding: 16rpx 64rpx;
  background: #ff5722;
  color: #fff;
  border: none;
  border-radius: 40rpx;
  font-size: 32rpx;
  font-weight: 600;
}

.loading {
  text-align: center;
  padding: 100rpx;
  color: #999;
  font-size: 28rpx;
}
</style>

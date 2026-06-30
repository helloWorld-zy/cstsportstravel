<template>
  <div class="product-detail-page">
    <!-- Loading State -->
    <div v-if="isLoading" class="loading-state">
      <div class="skeleton-banner" />
      <div class="skeleton-content">
        <div class="skeleton-line w80" />
        <div class="skeleton-line w60" />
        <div class="skeleton-line w40" />
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-state">
      <p>产品信息加载失败</p>
      <button @click="refetch">重试</button>
    </div>

    <!-- Product Content -->
    <template v-else-if="product">
      <!-- Image Carousel -->
      <div class="image-carousel">
        <div class="carousel-track">
          <img
            v-for="(img, idx) in productImages"
            :key="idx"
            :src="img"
            :alt="product.product_name"
            :class="{ active: currentImage === idx }"
            loading="lazy"
          />
        </div>
        <div class="carousel-dots">
          <span
            v-for="(_, idx) in productImages"
            :key="idx"
            class="dot"
            :class="{ active: currentImage === idx }"
            @click="currentImage = idx"
          />
        </div>
      </div>

      <!-- Product Info -->
      <div class="product-info section">
        <h1 class="product-name">{{ product.product_name }}</h1>
        <div class="product-meta">
          <span class="grade" v-if="product.product_grade">{{ gradeLabel }}</span>
          <span class="days">{{ product.days }}天{{ product.nights }}晚</span>
          <span class="transport" v-if="product.transport_mode">{{ transportLabel }}</span>
        </div>
        <div class="product-price">
          <span class="price-label">起</span>
          <span class="price-value">¥{{ product.min_price }}</span>
          <span class="price-unit">/人</span>
        </div>
        <p class="product-summary" v-if="product.summary">{{ product.summary }}</p>
        <div class="product-tags" v-if="product.tags?.length">
          <span v-for="tag in product.tags" :key="tag" class="tag">{{ tag }}</span>
        </div>
      </div>

      <!-- Itinerary -->
      <div class="itinerary-section section" v-if="product.itinerary?.length">
        <h2 class="section-title">行程详情</h2>
        <div class="itinerary-timeline">
          <div v-for="day in product.itinerary" :key="day.day_no" class="itinerary-day">
            <div class="day-marker">
              <span class="day-num">Day {{ day.day_no }}</span>
            </div>
            <div class="day-content">
              <h3 class="day-title">{{ day.title }}</h3>
              <p v-if="day.description" class="day-desc">{{ day.description }}</p>
              <div class="day-details">
                <div v-if="day.transport" class="detail-item">
                  <span class="detail-icon">🚗</span>
                  <span>{{ day.transport }}</span>
                </div>
                <div v-if="day.hotel" class="detail-item">
                  <span class="detail-icon">🏨</span>
                  <span>{{ day.hotel }}</span>
                </div>
                <div v-if="day.meals" class="detail-item">
                  <span class="detail-icon">🍽️</span>
                  <span>{{ mealsText(day.meals) }}</span>
                </div>
              </div>
              <div v-if="day.spots?.length" class="day-spots">
                <div v-for="spot in day.spots" :key="spot.name" class="spot-item">
                  <span class="spot-name">{{ spot.name }}</span>
                  <span v-if="spot.duration" class="spot-duration">{{ spot.duration }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Fees -->
      <div class="fees-section section">
        <h2 class="section-title">费用说明</h2>
        <div class="fee-block" v-if="product.fee_included">
          <h3 class="fee-title included">费用包含</h3>
          <div class="fee-content" v-html="product.fee_included" />
        </div>
        <div class="fee-block" v-if="product.fee_excluded">
          <h3 class="fee-title excluded">费用不含</h3>
          <div class="fee-content" v-html="product.fee_excluded" />
        </div>
      </div>

      <!-- Pricing by Type -->
      <div class="pricing-section section">
        <h2 class="section-title">人群价格</h2>
        <div class="pricing-grid">
          <div class="pricing-item">
            <span class="pricing-label">成人价</span>
            <span class="pricing-value">¥{{ product.min_price }}/人</span>
          </div>
          <div class="pricing-item">
            <span class="pricing-label">儿童价</span>
            <span class="pricing-value">¥{{ childPrice }}/人</span>
          </div>
          <div class="pricing-item">
            <span class="pricing-label">婴儿价</span>
            <span class="pricing-value">¥{{ infantPrice }}/人</span>
          </div>
          <div class="pricing-item" v-if="singleSupplement > 0">
            <span class="pricing-label">单房差</span>
            <span class="pricing-value">¥{{ singleSupplement }}/间</span>
          </div>
        </div>
      </div>

      <!-- Cancellation Policy (always visible) -->
      <div class="cancellation-section section" v-if="product.cancellation_rules?.length">
        <h2 class="section-title">退改政策</h2>
        <div class="cancellation-rules">
          <div v-for="rule in product.cancellation_rules" :key="rule.id" class="rule-item">
            <span class="rule-name">{{ rule.rule_name }}</span>
            <span class="rule-desc">{{ rule.description }}</span>
            <span class="rule-pct">退还 {{ rule.refund_percentage }}%</span>
          </div>
        </div>
      </div>

      <!-- Departure Calendar -->
      <div class="calendar-section section">
        <h2 class="section-title">团期日历</h2>
        <DepartureCalendar
          :departures="departures || []"
          :selected-date="selectedDeparture?.departure_date"
          @select="onSelectDeparture"
        />
      </div>

      <!-- Reviews -->
      <div class="reviews-section section" v-if="product.review_summary">
        <h2 class="section-title">用户评价</h2>
        <div class="review-summary">
          <div class="rating-big">
            <span class="rating-num">{{ product.review_summary.average_rating.toFixed(1) }}</span>
            <span class="rating-label">分</span>
          </div>
          <div class="rating-dist">
            <div v-for="star in [5, 4, 3, 2, 1]" :key="star" class="dist-row">
              <span class="dist-label">{{ star }}星</span>
              <div class="dist-bar">
                <div class="dist-fill" :style="{ width: ratingPct(star) + '%' }" />
              </div>
              <span class="dist-count">{{ product.review_summary.rating_distribution[String(star)] || 0 }}</span>
            </div>
          </div>
        </div>

        <div class="review-filters">
          <button
            v-for="opt in reviewFilterOptions"
            :key="opt.value"
            :class="{ active: reviewRatingFilter === opt.value }"
            @click="reviewRatingFilter = opt.value"
          >
            {{ opt.label }}
          </button>
        </div>

        <div v-if="reviewsLoading" class="loading-text">加载中...</div>
        <div v-else-if="reviews?.items?.length" class="review-list">
          <div v-for="review in reviews.items" :key="review.id" class="review-item">
            <div class="review-header">
              <span class="review-user">{{ review.is_anonymous ? '匿名用户' : `用户${review.user_id}` }}</span>
              <span class="review-rating">{{ '★'.repeat(review.rating) }}{{ '☆'.repeat(5 - review.rating) }}</span>
              <span class="review-date">{{ formatDate(review.created_at) }}</span>
            </div>
            <p class="review-content" v-if="review.content">{{ review.content }}</p>
            <div v-if="review.images?.length" class="review-images">
              <img v-for="(img, idx) in review.images" :key="idx" :src="img" alt="评价图片" />
            </div>
          </div>
        </div>
        <div v-else class="empty-reviews">暂无评价</div>
      </div>

      <!-- Bottom Booking Bar -->
      <div class="booking-bar">
        <div class="booking-price">
          <span class="price-label">起</span>
          <span class="price-value">¥{{ product.min_price }}</span>
          <span class="price-unit">/人</span>
        </div>
        <button class="booking-btn" @click="goToBooking">
          立即预订
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { useApi } from '~/composables/useApi'
import type { DepartureDate, ProductDetail, ReviewListResponse } from '~/composables/useProduct'

useHead({
  title: computed(() => product.value?.product_name || '产品详情'),
})

const route = useRoute()
const router = useRouter()
const api = useApi()

const productId = computed(() => route.params.id as string)

// Fetch product detail
const { data: product, isLoading, error, refetch } = useQuery({
  queryKey: ['product', productId],
  queryFn: () => api.get<ProductDetail>(`/products/${productId.value}`),
  enabled: computed(() => !!productId.value),
})

// Fetch departures
const currentMonth = new Date()
const monthStr = `${currentMonth.getFullYear()}-${String(currentMonth.getMonth() + 1).padStart(2, '0')}`
const { data: departures } = useQuery({
  queryKey: ['departures', productId, monthStr],
  queryFn: () => api.get<DepartureDate[]>(`/products/${productId.value}/departures`, { params: { month: monthStr, months: 3 } }),
  enabled: computed(() => !!productId.value),
})

// Reviews
const reviewRatingFilter = ref<number | undefined>(undefined)
const { data: reviews, isLoading: reviewsLoading } = useQuery({
  queryKey: ['reviews', productId, reviewRatingFilter],
  queryFn: () => api.get<ReviewListResponse>(`/products/${productId.value}/reviews`, {
    params: { rating: reviewRatingFilter.value, page: 1, page_size: 20 },
  }),
  enabled: computed(() => !!productId.value),
})

const reviewFilterOptions = [
  { label: '全部', value: undefined },
  { label: '好评', value: 5 },
  { label: '中评', value: 3 },
  { label: '差评', value: 1 },
]

// Image carousel
const currentImage = ref(0)
const productImages = computed(() => {
  if (!product.value) return []
  const imgs = product.value.images || []
  if (product.value.cover_image) return [product.value.cover_image, ...imgs]
  return imgs.length ? imgs : ['/static/images/default-product.png']
})

// Pricing
const childPrice = computed(() => {
  const dep = departures.value?.[0]
  return dep ? dep.child_price : 0
})
const infantPrice = computed(() => {
  const dep = departures.value?.[0]
  return dep ? dep.infant_price : 0
})
const singleSupplement = computed(() => {
  const dep = departures.value?.[0]
  return dep ? dep.single_supplement : 0
})

// Labels
const gradeLabel = computed(() => {
  const map: Record<string, string> = { standard: '经济', comfort: '舒适', luxury: '豪华' }
  return map[product.value?.product_grade || ''] || ''
})
const transportLabel = computed(() => {
  const map: Record<string, string> = { flight: '飞机', train: '火车', bus: '大巴' }
  return map[product.value?.transport_mode || ''] || ''
})

// Departure selection
const selectedDeparture = ref<DepartureDate | null>(null)
function onSelectDeparture(dep: DepartureDate) {
  selectedDeparture.value = dep
}

function mealsText(meals: { breakfast?: boolean; lunch?: boolean; dinner?: boolean }): string {
  const parts: string[] = []
  if (meals.breakfast) parts.push('早')
  if (meals.lunch) parts.push('中')
  if (meals.dinner) parts.push('晚')
  return parts.length ? `${parts.join('')}餐` : '自理'
}

function ratingPct(star: number): number {
  const total = product.value?.review_summary?.total_count || 1
  const count = product.value?.review_summary?.rating_distribution[String(star)] || 0
  return Math.round((count / total) * 100)
}

function formatDate(dateStr: string): string {
  return dateStr.substring(0, 10)
}

function goToBooking() {
  const depId = selectedDeparture.value?.id
  router.push(`/booking/${productId.value}${depId ? `?departure=${depId}` : ''}`)
}
</script>

<style scoped>
.product-detail-page {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px 16px 90px;
  background: transparent;
  min-height: calc(100vh - 70px);
}

.section {
  background: #fff;
  margin-top: 16px;
  padding: 24px;
  border-radius: 16px;
  border: 1px solid rgba(226, 232, 240, 0.8);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px -2px rgba(0, 0, 0, 0.02);
}

.section-title {
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f1f5f9;
  position: relative;
}

/* Image Carousel */
.image-carousel {
  position: relative;
  background: #0f172a;
  border-radius: 16px;
  overflow: hidden;
}

.carousel-track img {
  display: none;
  width: 100%;
  height: 240px;
  object-fit: cover;
}

.carousel-track img.active {
  display: block;
}

.carousel-dots {
  position: absolute;
  bottom: 16px;
  right: 24px;
  display: flex;
  gap: 8px;
  z-index: 5;
}

.carousel-dots .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.4);
  cursor: pointer;
  transition: all 0.2s;
}

.carousel-dots .dot.active {
  background: #fff;
  width: 18px;
  border-radius: 4px;
}

/* Product Info */
.product-name {
  font-size: 22px;
  font-weight: 800;
  color: #0f172a;
  margin: 0 0 12px;
  line-height: 1.3;
}

.product-meta {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.product-meta span {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  padding: 4px 10px;
  background: #f1f5f9;
  border-radius: 6px;
}

.product-price {
  display: flex;
  align-items: baseline;
  gap: 4px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f1f5f9;
}

.price-label {
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

.price-value {
  font-size: 28px;
  font-weight: 800;
  color: #2563eb;
}

.price-unit {
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

.product-summary {
  font-size: 14px;
  color: #475569;
  line-height: 1.6;
  margin: 0 0 16px;
}

.product-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.tag {
  padding: 4px 10px;
  background: #eff6ff;
  color: #2563eb;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
}

/* Itinerary */
.itinerary-timeline {
  position: relative;
  padding-left: 24px;
}

.itinerary-timeline::before {
  content: '';
  position: absolute;
  left: 8px;
  top: 8px;
  bottom: 8px;
  width: 2px;
  background: #e2e8f0;
}

.itinerary-day {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  position: relative;
}

.itinerary-day:last-child {
  margin-bottom: 0;
}

.day-marker {
  position: relative;
  z-index: 1;
}

.day-num {
  display: inline-block;
  background: linear-gradient(135deg, #1d4ed8 0%, #3b82f6 100%);
  color: #fff;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
  box-shadow: 0 4px 10px rgba(37, 99, 235, 0.2);
}

.day-content {
  flex: 1;
  padding-top: 2px;
}

.day-title {
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 8px;
}

.day-desc {
  font-size: 13px;
  color: #475569;
  margin: 0 0 12px;
  line-height: 1.6;
}

.day-details {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 12px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #64748b;
  font-weight: 500;
}

.day-spots {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.spot-item {
  padding: 6px 12px;
  background: #f8fafc;
  border-radius: 8px;
  font-size: 12px;
  border: 1px solid #f1f5f9;
  font-weight: 500;
}

.spot-name {
  color: #334155;
}

.spot-duration {
  color: #94a3b8;
  margin-left: 6px;
}

/* Fees */
.fee-block {
  margin-bottom: 20px;
}

.fee-block:last-child {
  margin-bottom: 0;
}

.fee-title {
  font-size: 15px;
  font-weight: 700;
  margin: 0 0 10px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.fee-title.included {
  color: #10b981;
}

.fee-title.excluded {
  color: #f59e0b;
}

.fee-content {
  font-size: 13px;
  color: #475569;
  line-height: 1.7;
}

/* Pricing */
.pricing-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.pricing-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f8fafc;
  border-radius: 12px;
  border: 1px solid #f1f5f9;
}

.pricing-label {
  font-size: 13px;
  color: #64748b;
  font-weight: 500;
}

.pricing-value {
  font-size: 14px;
  font-weight: 700;
  color: #2563eb;
}

/* Cancellation */
.cancellation-rules {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.rule-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f8fafc;
  border-radius: 12px;
  font-size: 13px;
  border: 1px solid #f1f5f9;
}

.rule-name {
  color: #0f172a;
  font-weight: 700;
}

.rule-desc {
  color: #475569;
  flex: 1;
  margin: 0 16px;
  font-weight: 500;
}

.rule-pct {
  color: #2563eb;
  font-weight: 700;
}

/* Reviews */
.review-summary {
  display: flex;
  gap: 32px;
  margin-bottom: 24px;
  align-items: center;
}

.rating-big {
  text-align: center;
  background: #fffbeb;
  padding: 16px;
  border-radius: 16px;
  border: 1px solid #fef3c7;
  min-width: 100px;
}

.rating-num {
  font-size: 42px;
  font-weight: 800;
  color: #d97706;
}

.rating-label {
  font-size: 12px;
  color: #b45309;
  font-weight: 600;
  display: block;
  margin-top: -2px;
}

.rating-dist {
  flex: 1;
}

.dist-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.dist-label {
  font-size: 12px;
  color: #64748b;
  width: 32px;
  font-weight: 500;
}

.dist-bar {
  flex: 1;
  height: 8px;
  background: #f1f5f9;
  border-radius: 4px;
  overflow: hidden;
}

.dist-fill {
  height: 100%;
  background: #f59e0b;
  border-radius: 4px;
}

.dist-count {
  font-size: 12px;
  color: #94a3b8;
  width: 30px;
  text-align: right;
  font-weight: 500;
}

.review-filters {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
}

.review-filters button {
  padding: 6px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  background: #fff;
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s;
}

.review-filters button:hover {
  color: #2563eb;
  border-color: rgba(37, 99, 235, 0.3);
}

.review-filters button.active {
  border-color: #2563eb;
  color: #2563eb;
  background: #eff6ff;
}

.review-item {
  padding: 20px 0;
  border-bottom: 1px solid #f1f5f9;
}

.review-item:last-child {
  border-bottom: none;
}

.review-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.review-user {
  font-size: 14px;
  font-weight: 700;
  color: #334155;
}

.review-rating {
  color: #f59e0b;
  font-size: 12px;
  letter-spacing: 2px;
}

.review-date {
  font-size: 12px;
  color: #94a3b8;
  margin-left: auto;
  font-weight: 500;
}

.review-content {
  font-size: 13px;
  color: #334155;
  line-height: 1.6;
  margin: 0;
  font-weight: 500;
}

.review-images {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.review-images img {
  width: 70px;
  height: 70px;
  object-fit: cover;
  border-radius: 8px;
  border: 1px solid #f1f5f9;
  cursor: zoom-in;
  transition: opacity 0.2s;
}

.review-images img:hover {
  opacity: 0.9;
}

.empty-reviews {
  text-align: center;
  padding: 32px;
  color: #94a3b8;
  font-weight: 500;
}

/* Booking Bar */
.booking-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 24px;
  background: #fff;
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.05);
  border-top: 1px solid #e2e8f0;
  z-index: 100;
}

.booking-bar .price-value {
  font-size: 24px;
}

.booking-btn {
  padding: 12px 36px;
  background: linear-gradient(135deg, #1d4ed8 0%, #3b82f6 100%);
  color: #fff;
  border: none;
  border-radius: 24px;
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.2);
  transition: all 0.2s;
}

.booking-btn:hover {
  opacity: 0.95;
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(37, 99, 235, 0.3);
}

/* Loading/Error/Skeleton */
.loading-state {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.skeleton-banner {
  height: 360px;
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 16px;
}

.skeleton-content {
  padding: 24px 0;
}

.skeleton-line {
  height: 18px;
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  margin-bottom: 16px;
  border-radius: 6px;
}

.w80 { width: 80%; }
.w60 { width: 60%; }
.w40 { width: 40%; }

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.error-state {
  text-align: center;
  padding: 80px 24px;
  color: #64748b;
  font-weight: 500;
}

.error-state button {
  margin-top: 16px;
  padding: 10px 28px;
  border: 1px solid #2563eb;
  border-radius: 8px;
  background: #fff;
  color: #2563eb;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.error-state button:hover {
  background: #eff6ff;
}

.loading-text {
  text-align: center;
  padding: 24px;
  color: #94a3b8;
  font-weight: 500;
}

/* Responsive desktop enhancements */
@media (min-width: 1024px) {
  .product-detail-page {
    display: grid;
    grid-template-columns: 1fr 380px;
    gap: 24px;
    padding: 24px 24px 100px;
  }

  .image-carousel {
    grid-column: span 2;
    margin-bottom: 8px;
  }
  
  .carousel-track img {
    height: 400px;
  }

  .product-info,
  .itinerary-section,
  .fees-section,
  .pricing-section,
  .cancellation-section,
  .reviews-section {
    grid-column: 1;
  }

  .calendar-section {
    grid-column: 2;
    grid-row: 3 / span 5; /* spans along the left content columns */
    align-self: start;
    position: sticky;
    top: 94px; /* sticky position below header */
    margin-top: 16px;
    z-index: 10;
  }

  .booking-bar {
    position: fixed;
    bottom: 24px;
    left: 50%;
    transform: translateX(-50%);
    width: calc(100% - 48px);
    max-width: 1200px;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(8px);
    border-radius: 16px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.12);
    border: 1px solid rgba(226, 232, 240, 0.8);
    padding: 16px 32px;
    z-index: 100;
  }
}
</style>

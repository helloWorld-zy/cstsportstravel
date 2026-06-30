<template>
  <div class="outbound-detail-page">
    <!-- Loading State -->
    <div v-if="isLoading" class="loading-state">
      <div class="skeleton-image" />
      <div class="skeleton-content" />
    </div>

    <!-- Product Detail -->
    <div v-else-if="product" class="product-detail">
      <!-- Image Gallery -->
      <div class="image-gallery">
        <img :src="product.cover_image || '/images/default-product.jpg'" :alt="product.product_name" />
        <!-- Visa Badge -->
        <span v-if="product.destination_country" class="visa-badge" :class="getVisaBadgeClass()">
          {{ getVisaBadgeText() }}
        </span>
      </div>

      <!-- Product Info -->
      <div class="product-info">
        <h1>{{ product.product_name }}</h1>
        <div class="info-meta">
          <span class="days">{{ product.days }}天{{ product.nights }}晚</span>
          <span class="origin">{{ product.origin_city }}出发</span>
          <span v-if="product.destination_country" class="destination">
            目的地：{{ product.destination_country.name_cn }}
          </span>
        </div>
        <div class="info-price">
          <span class="price-value">¥{{ getMinPrice() }}</span>
          <span class="price-unit">起/人</span>
        </div>
      </div>

      <!-- Visa Info Card (PRD §4.3.2) -->
      <div v-if="product.visa_info_parsed" class="visa-info-card">
        <h3 class="card-title">
          <span class="icon">🛂</span>
          签证信息
        </h3>
        <div class="visa-details">
          <div class="visa-row">
            <span class="label">签证类型</span>
            <span class="value">{{ product.visa_info_parsed.visa_type }}</span>
          </div>
          <div v-if="product.visa_info_parsed.processing_days" class="visa-row">
            <span class="label">办理周期</span>
            <span class="value">{{ product.visa_info_parsed.processing_days }}个工作日</span>
          </div>
          <div v-if="product.visa_info_parsed.fee" class="visa-row">
            <span class="label">签证费用</span>
            <span class="value">¥{{ (product.visa_info_parsed.fee / 100).toFixed(0) }}</span>
          </div>
          <div v-if="product.visa_info_parsed.consular_district" class="visa-row">
            <span class="label">领区</span>
            <span class="value">{{ product.visa_info_parsed.consular_district }}</span>
          </div>
        </div>

        <!-- Material Preview by Occupation -->
        <div v-if="product.material_preview && product.material_preview.length > 0" class="material-preview">
          <h4>材料清单预览</h4>
          <div class="occupation-tabs">
            <button
              v-for="occ in product.material_preview"
              :key="occ.occupation_type"
              class="occ-tab"
              :class="{ active: selectedOccupation === occ.occupation_type }"
              @click="selectedOccupation = occ.occupation_type"
            >
              {{ occ.occupation_name }}
            </button>
          </div>
          <div class="material-list">
            <div v-for="mat in currentMaterials" :key="mat.id" class="material-item">
              <span class="material-name">{{ mat.material_name }}</span>
              <span v-if="mat.is_required" class="required">必填</span>
              <span v-else class="optional">选填</span>
            </div>
          </div>
          <button class="view-full-btn" @click="showFullMaterials = true">
            查看完整材料清单
          </button>
        </div>

        <!-- Reject Refund Policy -->
        <div v-if="product.visa_info_parsed.reject_refund_policy" class="refund-policy">
          <span class="icon">🛡️</span>
          <span>{{ product.visa_info_parsed.reject_refund_policy }}</span>
        </div>
      </div>

      <!-- International Flight Info -->
      <div v-if="product.flight_info_parsed" class="flight-info-card">
        <h3 class="card-title">
          <span class="icon">✈️</span>
          国际航班信息
        </h3>
        <div class="flight-details">
          <div class="flight-row">
            <span class="label">航空公司</span>
            <span class="value">{{ product.flight_info_parsed.airline }}</span>
          </div>
          <div class="flight-row">
            <span class="label">航班号</span>
            <span class="value">{{ product.flight_info_parsed.flight_no }}</span>
          </div>
          <div class="flight-row">
            <span class="label">出发城市</span>
            <span class="value">{{ product.flight_info_parsed.depart_city }}</span>
          </div>
          <div class="flight-row">
            <span class="label">到达城市</span>
            <span class="value">{{ product.flight_info_parsed.arrive_city }}</span>
          </div>
          <div v-if="product.flight_info_parsed.stops !== undefined" class="flight-row">
            <span class="label">经停</span>
            <span class="value">{{ product.flight_info_parsed.stops === 0 ? '直飞' : product.flight_info_parsed.stops + '次经停' }}</span>
          </div>
        </div>
      </div>

      <!-- Insurance Requirements -->
      <div v-if="product.insurance_requirements_parsed" class="insurance-card">
        <h3 class="card-title">
          <span class="icon">🏥</span>
          保险要求
        </h3>
        <div class="insurance-details">
          <p v-if="product.insurance_requirements_parsed.schengen" class="schengen-notice">
            ⚠️ 目的地为申根国家，需购买符合申根签证要求的旅行保险（医疗保额≥3万欧元）
          </p>
          <p v-if="product.insurance_requirements_parsed.description">
            {{ product.insurance_requirements_parsed.description }}
          </p>
        </div>
      </div>

      <!-- Itinerary -->
      <div v-if="product.itineraries && product.itineraries.length > 0" class="itinerary-card">
        <h3 class="card-title">行程安排</h3>
        <div v-for="day in product.itineraries" :key="day.day_no" class="itinerary-day">
          <div class="day-header">第{{ day.day_no }}天</div>
          <div class="day-content">
            <h4>{{ day.title }}</h4>
            <p v-if="day.description">{{ day.description }}</p>
          </div>
        </div>
      </div>

      <!-- Outbound FAQ -->
      <div class="faq-card">
        <h3 class="card-title">常见问题</h3>
        <div class="faq-list">
          <div class="faq-item">
            <div class="faq-q">签证需要多长时间办理？</div>
            <div class="faq-a">一般需要7-15个工作日，具体时间以目的地国家使领馆为准。建议提前30天以上准备材料。</div>
          </div>
          <div class="faq-item">
            <div class="faq-q">护照有效期有什么要求？</div>
            <div class="faq-a">大多数国家要求护照有效期覆盖回程日期后至少6个月。请在预订前确认护照有效期。</div>
          </div>
          <div class="faq-item">
            <div class="faq-q">需要购买什么保险？</div>
            <div class="faq-a">出境游建议购买包含医疗、意外、行李丢失等保障的旅行保险。申根国家要求医疗保额≥3万欧元。</div>
          </div>
          <div class="faq-item">
            <div class="faq-q">外币如何兑换？</div>
            <div class="faq-a">建议在国内银行提前兑换部分当地货币，也可携带银联卡/Visa卡在当地ATM取现。</div>
          </div>
        </div>
      </div>

      <!-- Booking Bar -->
      <div class="booking-bar">
        <div class="bar-price">
          <span class="price-value">¥{{ getMinPrice() }}</span>
          <span class="price-unit">起/人</span>
        </div>
        <button class="book-btn" @click="goToBooking">立即预订</button>
      </div>
    </div>

    <!-- Full Materials Modal -->
    <div v-if="showFullMaterials" class="modal-overlay" @click.self="showFullMaterials = false">
      <div class="modal-content">
        <h3>完整材料清单</h3>
        <div v-if="product?.material_preview" class="full-materials">
          <div v-for="occ in product.material_preview" :key="occ.occupation_type" class="occ-section">
            <h4>{{ occ.occupation_name }}</h4>
            <div v-for="mat in occ.materials" :key="mat.id" class="material-item">
              <span class="material-name">{{ mat.material_name }}</span>
              <span v-if="mat.is_required" class="required">必填</span>
              <span v-else class="optional">选填</span>
              <p v-if="mat.description" class="material-desc">{{ mat.description }}</p>
            </div>
          </div>
        </div>
        <button class="close-btn" @click="showFullMaterials = false">关闭</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const route = useRoute()
const product = ref<any>(null)
const isLoading = ref(true)
const selectedOccupation = ref('')
const showFullMaterials = ref(false)

// Load product detail
const loadProduct = async () => {
  isLoading.value = true
  try {
    const id = route.params.id
    const data = await $fetch(`/api/v2/products/outbound/${id}`)
    product.value = data
    // Set default occupation tab
    if (product.value?.material_preview?.length > 0) {
      selectedOccupation.value = product.value.material_preview[0].occupation_type
    }
  } catch (error) {
    console.error('Failed to load product:', error)
  } finally {
    isLoading.value = false
  }
}

// Computed
const currentMaterials = computed(() => {
  if (!product.value?.material_preview) return []
  const occ = product.value.material_preview.find(
    (o: any) => o.occupation_type === selectedOccupation.value
  )
  return occ?.materials || []
})

// Methods
const getVisaBadgeClass = () => {
  if (!product.value?.destination_country) return ''
  switch (product.value.destination_country.visa_type) {
    case 'visa_free': return 'badge-free'
    case 'visa_on_arrival': return 'badge-arrival'
    case 'e_visa': return 'badge-evisa'
    case 'visa_required': return 'badge-required'
    default: return ''
  }
}

const getVisaBadgeText = () => {
  if (!product.value?.destination_country) return ''
  switch (product.value.destination_country.visa_type) {
    case 'visa_free': return '免签直飞'
    case 'visa_on_arrival': return '落地签'
    case 'e_visa': return '电子签'
    case 'visa_required': return '含签证代办'
    default: return ''
  }
}

const getMinPrice = () => {
  if (!product.value?.departure_dates?.length) return '--'
  const minPrice = Math.min(...product.value.departure_dates.map((d: any) => d.adult_price))
  return (minPrice / 100).toFixed(0)
}

const goToBooking = () => {
  navigateTo(`/outbound/booking?product_id=${product.value.id}`)
}

// Lifecycle
onMounted(() => {
  loadProduct()
})
</script>

<style scoped>
.outbound-detail-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  padding-bottom: 100px;
}

.image-gallery {
  position: relative;
  height: 400px;
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 24px;
}

.image-gallery img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.visa-badge {
  position: absolute;
  top: 16px;
  left: 16px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  color: white;
}

.badge-free { background: #52c41a; }
.badge-arrival { background: #1890ff; }
.badge-evisa { background: #722ed1; }
.badge-required { background: #fa8c16; }

.product-info {
  margin-bottom: 24px;
}

.product-info h1 {
  font-size: 28px;
  color: #1a1a1a;
  margin-bottom: 12px;
}

.info-meta {
  display: flex;
  gap: 16px;
  color: #666;
  margin-bottom: 12px;
}

.info-price {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.price-value {
  font-size: 32px;
  color: #ff4d4f;
  font-weight: 600;
}

.price-unit {
  font-size: 14px;
  color: #999;
}

/* Visa Info Card */
.visa-info-card, .flight-info-card, .insurance-card, .itinerary-card, .faq-card {
  background: white;
  border: 1px solid #eee;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
}

.card-title {
  font-size: 18px;
  color: #1a1a1a;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.visa-row, .flight-row {
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.visa-row:last-child, .flight-row:last-child {
  border-bottom: none;
}

.label {
  color: #666;
}

.value {
  color: #1a1a1a;
  font-weight: 500;
}

/* Material Preview */
.material-preview {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #eee;
}

.material-preview h4 {
  font-size: 16px;
  margin-bottom: 12px;
}

.occupation-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.occ-tab {
  padding: 8px 16px;
  border: 1px solid #ddd;
  border-radius: 20px;
  background: white;
  cursor: pointer;
  font-size: 14px;
}

.occ-tab.active {
  background: #e6f7ff;
  border-color: #1890ff;
  color: #1890ff;
}

.material-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}

.material-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: #fafafa;
  border-radius: 8px;
}

.material-name {
  flex: 1;
}

.required {
  color: #ff4d4f;
  font-size: 12px;
}

.optional {
  color: #999;
  font-size: 12px;
}

.view-full-btn {
  width: 100%;
  padding: 12px;
  background: #f5f5f5;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  color: #1890ff;
}

.refund-policy {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  padding: 12px;
  background: #fff7e6;
  border-radius: 8px;
  color: #fa8c16;
}

.schengen-notice {
  color: #fa8c16;
  font-weight: 500;
  margin-bottom: 8px;
}

/* Itinerary */
.itinerary-day {
  display: flex;
  gap: 16px;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}

.day-header {
  width: 60px;
  height: 60px;
  background: #1890ff;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  flex-shrink: 0;
}

.day-content h4 {
  margin-bottom: 8px;
}

.day-content p {
  color: #666;
  line-height: 1.6;
}

/* FAQ */
.faq-item {
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}

.faq-item:last-child {
  border-bottom: none;
}

.faq-q {
  font-weight: 500;
  margin-bottom: 8px;
  color: #1a1a1a;
}

.faq-a {
  color: #666;
  line-height: 1.6;
}

/* Booking Bar */
.booking-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  padding: 16px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.1);
  z-index: 100;
}

.bar-price .price-value {
  font-size: 28px;
}

.book-btn {
  padding: 12px 48px;
  background: #ff4d4f;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 24px;
  max-width: 800px;
  max-height: 80vh;
  overflow-y: auto;
  width: 90%;
}

.modal-content h3 {
  margin-bottom: 20px;
}

.occ-section {
  margin-bottom: 24px;
}

.occ-section h4 {
  margin-bottom: 12px;
  color: #1890ff;
}

.material-desc {
  font-size: 13px;
  color: #999;
  margin-top: 4px;
}

.close-btn {
  width: 100%;
  padding: 12px;
  background: #1890ff;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  margin-top: 20px;
}

/* Loading */
.loading-state {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.skeleton-image {
  height: 400px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: loading 1.5s infinite;
  border-radius: 12px;
}

.skeleton-content {
  height: 200px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: loading 1.5s infinite;
  border-radius: 12px;
}

@keyframes loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>

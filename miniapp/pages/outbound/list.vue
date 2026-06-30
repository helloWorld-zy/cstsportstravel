<template>
  <view class="outbound-list">
    <!-- Header -->
    <view class="header">
      <text class="title">出境游</text>
      <text class="subtitle">探索世界，从这里开始</text>
    </view>

    <!-- Continent Tabs -->
    <scroll-view scroll-x class="continent-tabs">
      <view
        v-for="continent in continents"
        :key="continent.code"
        class="tab-item"
        :class="{ active: selectedContinent === continent.code }"
        @tap="selectContinent(continent.code)"
      >
        <text>{{ continent.name }}</text>
      </view>
    </scroll-view>

    <!-- Filter Tags -->
    <scroll-view scroll-x class="filter-tags">
      <view
        v-for="vt in visaTypes"
        :key="vt.value"
        class="tag-item"
        :class="{ active: filters.visa_type === vt.value }"
        @tap="toggleVisaType(vt.value)"
      >
        <text>{{ vt.label }}</text>
      </view>
    </scroll-view>

    <!-- Product List -->
    <view v-if="isLoading" class="loading">
      <view v-for="i in 3" :key="i" class="skeleton-card" />
    </view>

    <view v-else-if="products.length === 0" class="empty">
      <text>暂无符合条件的产品</text>
    </view>

    <view v-else class="product-list">
      <view
        v-for="product in products"
        :key="product.id"
        class="product-card"
        @tap="goToDetail(product.id)"
      >
        <image
          class="card-image"
          :src="product.cover_image || '/static/images/default-product.jpg'"
          mode="aspectFill"
        />
        <view class="visa-badge" :class="getVisaBadgeClass(product)">
          <text>{{ getVisaBadgeText(product) }}</text>
        </view>
        <view class="card-content">
          <text class="card-title">{{ product.product_name }}</text>
          <view class="card-meta">
            <text class="days">{{ product.days }}天{{ product.nights }}晚</text>
            <text class="origin">{{ product.origin_city }}出发</text>
          </view>
          <view class="card-price">
            <text class="price-symbol">¥</text>
            <text class="price-value">{{ getMinPrice(product) }}</text>
            <text class="price-unit">起/人</text>
          </view>
        </view>
      </view>
    </view>

    <!-- Load More -->
    <view v-if="hasMore && !isLoading" class="load-more" @tap="loadMore">
      <text>加载更多</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'

const isLoading = ref(false)
const products = ref<any[]>([])
const currentPage = ref(1)
const hasMore = ref(true)
const selectedContinent = ref('')

const filters = reactive({
  visa_type: '',
})

const continents = [
  { code: '', name: '全部' },
  { code: 'asia', name: '亚洲' },
  { code: 'europe', name: '欧洲' },
  { code: 'north_america', name: '北美洲' },
  { code: 'oceania', name: '大洋洲' },
]

const visaTypes = [
  { value: '', label: '全部' },
  { value: 'visa_free', label: '免签' },
  { value: 'visa_on_arrival', label: '落地签' },
  { value: 'e_visa', label: '电子签' },
  { value: 'visa_required', label: '需办签' },
]

const loadProducts = async (reset = false) => {
  if (reset) {
    currentPage.value = 1
    products.value = []
    hasMore.value = true
  }

  isLoading.value = true
  try {
    const params: Record<string, any> = {
      page: currentPage.value,
      page_size: 10,
    }
    if (selectedContinent.value) params.continent = selectedContinent.value
    if (filters.visa_type) params.visa_type = filters.visa_type

    const res = await uni.request({
      url: '/api/v2/products/outbound',
      method: 'GET',
      data: params,
    })

    const result = res.data as any
    if (result.items) {
      products.value = [...products.value, ...result.items]
      hasMore.value = result.items.length === 10
    }
  } catch (error) {
    console.error('Failed to load products:', error)
  } finally {
    isLoading.value = false
  }
}

const selectContinent = (code: string) => {
  selectedContinent.value = code
  loadProducts(true)
}

const toggleVisaType = (type: string) => {
  filters.visa_type = filters.visa_type === type ? '' : type
  loadProducts(true)
}

const loadMore = () => {
  currentPage.value++
  loadProducts()
}

const goToDetail = (id: number) => {
  uni.navigateTo({ url: `/pages/outbound/detail?id=${id}` })
}

const getVisaBadgeClass = (product: any) => {
  if (!product.destination_country) return ''
  switch (product.destination_country.visa_type) {
    case 'visa_free': return 'badge-free'
    case 'visa_on_arrival': return 'badge-arrival'
    case 'e_visa': return 'badge-evisa'
    case 'visa_required': return 'badge-required'
    default: return ''
  }
}

const getVisaBadgeText = (product: any) => {
  if (!product.destination_country) return ''
  switch (product.destination_country.visa_type) {
    case 'visa_free': return '免签'
    case 'visa_on_arrival': return '落地签'
    case 'e_visa': return '电子签'
    case 'visa_required': return '需办签'
    default: return ''
  }
}

const getMinPrice = (product: any) => {
  if (!product.departure_dates?.length) return '--'
  const minPrice = Math.min(...product.departure_dates.map((d: any) => d.adult_price))
  return (minPrice / 100).toFixed(0)
}

onMounted(() => {
  loadProducts(true)
})
</script>

<style scoped>
.outbound-list {
  padding: 20rpx;
  background: #f5f5f5;
  min-height: 100vh;
}

.header {
  text-align: center;
  padding: 40rpx 0;
}

.title {
  font-size: 48rpx;
  font-weight: 600;
  color: #1a1a1a;
  display: block;
}

.subtitle {
  font-size: 28rpx;
  color: #666;
  margin-top: 8rpx;
}

.continent-tabs {
  white-space: nowrap;
  margin-bottom: 20rpx;
}

.tab-item {
  display: inline-block;
  padding: 16rpx 32rpx;
  margin-right: 16rpx;
  background: white;
  border-radius: 32rpx;
  font-size: 28rpx;
}

.tab-item.active {
  background: #1890ff;
  color: white;
}

.filter-tags {
  white-space: nowrap;
  margin-bottom: 20rpx;
}

.tag-item {
  display: inline-block;
  padding: 12rpx 24rpx;
  margin-right: 12rpx;
  background: white;
  border-radius: 8rpx;
  font-size: 26rpx;
}

.tag-item.active {
  background: #e6f7ff;
  color: #1890ff;
}

.product-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.product-card {
  background: white;
  border-radius: 16rpx;
  overflow: hidden;
  position: relative;
}

.card-image {
  width: 100%;
  height: 300rpx;
}

.visa-badge {
  position: absolute;
  top: 20rpx;
  left: 20rpx;
  padding: 8rpx 16rpx;
  border-radius: 8rpx;
  font-size: 24rpx;
  color: white;
}

.badge-free { background: #52c41a; }
.badge-arrival { background: #1890ff; }
.badge-evisa { background: #722ed1; }
.badge-required { background: #fa8c16; }

.card-content {
  padding: 20rpx;
}

.card-title {
  font-size: 30rpx;
  color: #1a1a1a;
  font-weight: 500;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-meta {
  display: flex;
  gap: 16rpx;
  margin-top: 12rpx;
  font-size: 26rpx;
  color: #999;
}

.card-price {
  display: flex;
  align-items: baseline;
  gap: 4rpx;
  margin-top: 16rpx;
}

.price-symbol {
  font-size: 26rpx;
  color: #ff4d4f;
}

.price-value {
  font-size: 40rpx;
  color: #ff4d4f;
  font-weight: 600;
}

.price-unit {
  font-size: 24rpx;
  color: #999;
}

.loading {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.skeleton-card {
  height: 400rpx;
  background: #e0e0e0;
  border-radius: 16rpx;
}

.empty {
  text-align: center;
  padding: 100rpx 0;
  color: #999;
}

.load-more {
  text-align: center;
  padding: 30rpx 0;
  color: #1890ff;
}
</style>

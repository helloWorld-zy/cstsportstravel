<template>
  <view class="product-list-page">
    <!-- Search Bar -->
    <view class="search-bar">
      <input
        v-model="keyword"
        type="text"
        placeholder="搜索目的地、产品名称"
        confirm-type="search"
        @confirm="onSearch"
      />
    </view>

    <!-- Filter Tabs -->
    <view class="filter-tabs">
      <scroll-view scroll-x class="filter-scroll">
        <view
          v-for="tab in filterTabs"
          :key="tab.key"
          class="filter-tab"
          :class="{ active: activeFilter === tab.key }"
          @click="openFilter(tab.key)"
        >
          <text>{{ tab.label }}</text>
        </view>
      </scroll-view>
    </view>

    <!-- Sort Bar -->
    <view class="sort-bar">
      <view
        v-for="opt in sortOptions"
        :key="opt.value"
        class="sort-item"
        :class="{ active: sort === opt.value }"
        @click="sort = opt.value; page = 1; loadProducts()"
      >
        <text>{{ opt.label }}</text>
      </view>
    </view>

    <!-- Product List -->
    <scroll-view
      scroll-y
      class="product-scroll"
      :style="{ height: scrollHeight }"
      @scrolltolower="loadMore"
      refresher-enabled
      :refresher-triggered="isRefreshing"
      @refresherrefresh="onRefresh"
    >
      <view v-if="isLoading && products.length === 0" class="loading">
        <text>加载中...</text>
      </view>

      <view v-else-if="products.length === 0" class="empty">
        <text>暂无符合条件的产品</text>
      </view>

      <view v-else class="product-cards">
        <view
          v-for="product in products"
          :key="product.id"
          class="product-card"
          @click="goToDetail(product.id)"
        >
          <image :src="product.cover_image || '/static/images/default-product.png'" mode="aspectFill" class="card-img" />
          <view class="card-body">
            <text class="card-title">{{ product.product_name }}</text>
            <text class="card-dest">{{ product.destination_cities.join('·') }}</text>
            <text class="card-days">{{ product.days }}天{{ product.nights }}晚 · {{ product.origin_city }}出发</text>
            <view class="card-footer">
              <text class="card-price">¥{{ product.min_price }}起</text>
              <text v-if="product.satisfaction_rate" class="card-rating">{{ product.satisfaction_rate.toFixed(1) }}分</text>
            </view>
          </view>
        </view>
      </view>

      <view v-if="!hasMore && products.length > 0" class="no-more">
        <text>没有更多了</text>
      </view>
    </scroll-view>

    <!-- Filter Drawer -->
    <view v-if="showDrawer" class="drawer-mask" @click="showDrawer = false">
      <view class="drawer" @click.stop>
        <view class="drawer-header">
          <text class="drawer-title">筛选</text>
          <text class="drawer-close" @click="showDrawer = false">✕</text>
        </view>
        <scroll-view scroll-y class="drawer-body">
          <view class="filter-group">
            <text class="group-title">目的地</text>
            <view class="group-options">
              <text
                v-for="dest in destinationOptions"
                :key="dest"
                class="option-item"
                :class="{ selected: destination === dest }"
                @click="destination = destination === dest ? '' : dest"
              >{{ dest }}</text>
            </view>
          </view>
          <view class="filter-group">
            <text class="group-title">行程天数</text>
            <view class="group-options">
              <text
                v-for="range in daysOptions"
                :key="range.label"
                class="option-item"
                :class="{ selected: daysMin === range.min && daysMax === range.max }"
                @click="daysMin = range.min; daysMax = range.max"
              >{{ range.label }}</text>
            </view>
          </view>
          <view class="filter-group">
            <text class="group-title">价格区间</text>
            <view class="group-options">
              <text
                v-for="range in priceOptions"
                :key="range.label"
                class="option-item"
                :class="{ selected: priceMin === range.min && priceMax === range.max }"
                @click="priceMin = range.min; priceMax = range.max"
              >{{ range.label }}</text>
            </view>
          </view>
        </scroll-view>
        <view class="drawer-footer">
          <button class="btn-reset" @click="resetFilters">重置</button>
          <button class="btn-apply" @click="applyFilters">确定</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/shared/api/request'

interface ProductSummary {
  id: number
  product_name: string
  cover_image?: string
  origin_city: string
  destination_cities: string[]
  days: number
  nights: number
  min_price: number
  satisfaction_rate?: number
  order_count: number
}

const keyword = ref('')
const destination = ref('')
const daysMin = ref<number | undefined>(undefined)
const daysMax = ref<number | undefined>(undefined)
const priceMin = ref<number | undefined>(undefined)
const priceMax = ref<number | undefined>(undefined)
const sort = ref('recommended')
const page = ref(1)
const pageSize = 20
const products = ref<ProductSummary[]>([])
const total = ref(0)
const isLoading = ref(false)
const isRefreshing = ref(false)
const showDrawer = ref(false)
const activeFilter = ref('')
const scrollHeight = ref('600px')

const hasMore = computed(() => products.value.length < total.value)

const filterTabs = [
  { key: 'destination', label: '目的地' },
  { key: 'days', label: '天数' },
  { key: 'price', label: '价格' },
]

const sortOptions = [
  { label: '推荐', value: 'recommended' },
  { label: '价格↑', value: 'price_asc' },
  { label: '价格↓', value: 'price_desc' },
  { label: '满意度', value: 'satisfaction' },
]

const destinationOptions = ['云南', '海南', '北京', '四川', '广西', '西安', '厦门']
const daysOptions = [
  { label: '不限', min: undefined, max: undefined },
  { label: '1-3天', min: 1, max: 3 },
  { label: '4-6天', min: 4, max: 6 },
  { label: '7天以上', min: 7, max: undefined },
]
const priceOptions = [
  { label: '不限', min: undefined, max: undefined },
  { label: '2000以下', min: undefined, max: 2000 },
  { label: '2000-4000', min: 2000, max: 4000 },
  { label: '4000以上', min: 4000, max: undefined },
]

async function loadProducts() {
  isLoading.value = true
  try {
    const params: Record<string, any> = {
      page: page.value,
      page_size: pageSize,
      sort: sort.value,
    }
    if (destination.value) params.destination = destination.value
    if (daysMin.value) params.days_min = daysMin.value
    if (daysMax.value) params.days_max = daysMax.value
    if (priceMin.value) params.price_min = priceMin.value
    if (priceMax.value) params.price_max = priceMax.value
    if (keyword.value) params.keyword = keyword.value

    const data = await api.get<{ items: ProductSummary[]; total: number }>('/products', params)
    if (page.value === 1) {
      products.value = data.items || []
    } else {
      products.value = [...products.value, ...(data.items || [])]
    }
    total.value = data.total || 0
  } catch (e) {
    console.error('Failed to load products', e)
  } finally {
    isLoading.value = false
    isRefreshing.value = false
  }
}

function loadMore() {
  if (!hasMore.value || isLoading.value) return
  page.value++
  loadProducts()
}

function onRefresh() {
  isRefreshing.value = true
  page.value = 1
  loadProducts()
}

function onSearch() {
  page.value = 1
  loadProducts()
}

function openFilter(key: string) {
  activeFilter.value = key
  showDrawer.value = true
}

function resetFilters() {
  destination.value = ''
  daysMin.value = undefined
  daysMax.value = undefined
  priceMin.value = undefined
  priceMax.value = undefined
}

function applyFilters() {
  showDrawer.value = false
  page.value = 1
  loadProducts()
}

function goToDetail(id: number) {
  uni.navigateTo({ url: `/pages/products/detail?id=${id}` })
}

onMounted(() => {
  // Get system info for scroll height
  const sysInfo = uni.getSystemInfoSync()
  scrollHeight.value = `${sysInfo.windowHeight - 200}px`

  // Check for query params
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1]
  if (currentPage?.options?.destination) {
    destination.value = currentPage.options.destination
  }
  if (currentPage?.options?.keyword) {
    keyword.value = currentPage.options.keyword
  }

  loadProducts()
})
</script>

<style scoped>
.product-list-page {
  background: #f5f5f5;
  min-height: 100vh;
}

.search-bar {
  padding: 16rpx 24rpx;
  background: #fff;
}

.search-bar input {
  background: #f5f5f5;
  padding: 12rpx 24rpx;
  border-radius: 32rpx;
  font-size: 28rpx;
}

.filter-tabs {
  background: #fff;
  border-bottom: 1rpx solid #f0f0f0;
}

.filter-scroll {
  white-space: nowrap;
  padding: 12rpx 24rpx;
}

.filter-tab {
  display: inline-block;
  padding: 8rpx 24rpx;
  margin-right: 16rpx;
  border: 1rpx solid #ddd;
  border-radius: 24rpx;
  font-size: 24rpx;
}

.filter-tab.active {
  border-color: #ff5722;
  color: #ff5722;
  background: #fff3e0;
}

.sort-bar {
  display: flex;
  padding: 12rpx 24rpx;
  background: #fff;
  border-bottom: 1rpx solid #f0f0f0;
}

.sort-item {
  margin-right: 32rpx;
  font-size: 24rpx;
  color: #666;
}

.sort-item.active {
  color: #ff5722;
  font-weight: 600;
}

.product-scroll {
  padding: 16rpx;
}

.product-cards {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.product-card {
  display: flex;
  gap: 16rpx;
  background: #fff;
  border-radius: 12rpx;
  overflow: hidden;
  padding: 16rpx;
}

.card-img {
  width: 220rpx;
  height: 180rpx;
  border-radius: 8rpx;
}

.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.card-title {
  font-size: 28rpx;
  font-weight: 500;
  color: #333;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-dest {
  font-size: 24rpx;
  color: #666;
}

.card-days {
  font-size: 22rpx;
  color: #999;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.card-price {
  font-size: 32rpx;
  font-weight: bold;
  color: #ff5722;
}

.card-rating {
  font-size: 22rpx;
  color: #ff9800;
}

.loading, .empty, .no-more {
  text-align: center;
  padding: 40rpx;
  color: #999;
  font-size: 28rpx;
}

/* Drawer */
.drawer-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 100;
}

.drawer {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 560rpx;
  background: #fff;
  display: flex;
  flex-direction: column;
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx;
  border-bottom: 1rpx solid #eee;
}

.drawer-title {
  font-size: 32rpx;
  font-weight: 600;
}

.drawer-close {
  font-size: 32rpx;
  color: #999;
}

.drawer-body {
  flex: 1;
  padding: 24rpx;
}

.filter-group {
  margin-bottom: 32rpx;
}

.group-title {
  font-size: 28rpx;
  font-weight: 500;
  margin-bottom: 12rpx;
  display: block;
}

.group-options {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.option-item {
  padding: 8rpx 24rpx;
  border: 1rpx solid #eee;
  border-radius: 8rpx;
  font-size: 24rpx;
}

.option-item.selected {
  border-color: #ff5722;
  color: #ff5722;
  background: #fff3e0;
}

.drawer-footer {
  display: flex;
  gap: 16rpx;
  padding: 24rpx;
  border-top: 1rpx solid #eee;
}

.drawer-footer button {
  flex: 1;
  padding: 16rpx;
  border-radius: 8rpx;
  font-size: 28rpx;
}

.btn-reset {
  background: #fff;
  border: 1rpx solid #ddd;
  color: #333;
}

.btn-apply {
  background: #ff5722;
  color: #fff;
  border: none;
}
</style>

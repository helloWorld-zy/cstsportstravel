<template>
  <view class="home">
    <!-- Search Bar -->
    <view class="search-bar" @click="goToSearch">
      <text class="search-icon">🔍</text>
      <text class="search-placeholder">搜索目的地、产品</text>
    </view>

    <!-- Banner -->
    <view class="banner">
      <swiper indicator-dots autoplay circular :interval="4000">
        <swiper-item v-for="banner in homepageData?.banners || defaultBanners" :key="banner.id">
          <view class="banner-item">
            <image :src="banner.image_url" mode="aspectFill" class="banner-img" />
            <text class="banner-title">{{ banner.title }}</text>
          </view>
        </swiper-item>
      </swiper>
    </view>

    <!-- 金刚区 -->
    <view class="icon-grid">
      <view v-for="item in iconGridItems" :key="item.label" class="icon-item" @click="navigateTo(item.link)">
        <text class="icon-emoji">{{ item.icon }}</text>
        <text class="icon-label">{{ item.label }}</text>
      </view>
    </view>

    <!-- Popular Destinations -->
    <view class="section">
      <view class="section-header">
        <text class="section-title">热门目的地</text>
      </view>
      <scroll-view scroll-x class="dest-scroll">
        <view
          v-for="dest in popularDestinations"
          :key="dest.name"
          class="dest-card"
          @click="goToProducts(dest.name)"
        >
          <text class="dest-name">{{ dest.name }}</text>
          <text class="dest-info">{{ dest.product_count }}条线路 ¥{{ dest.min_price }}起</text>
        </view>
      </scroll-view>
    </view>

    <!-- Recommended Products -->
    <view class="section">
      <view class="section-header">
        <text class="section-title">猜你喜欢</text>
      </view>
      <view v-if="isLoading" class="loading">
        <text>加载中...</text>
      </view>
      <view v-else class="product-list">
        <view
          v-for="product in recommendedProducts"
          :key="product.id"
          class="product-card"
          @click="goToDetail(product.id)"
        >
          <image :src="product.cover_image || '/static/images/default-product.png'" mode="aspectFill" class="product-img" />
          <view class="product-info">
            <text class="product-name">{{ product.product_name }}</text>
            <text class="product-dest">{{ product.destination_cities.join('·') }}</text>
            <text class="product-days">{{ product.days }}天{{ product.nights }}晚</text>
            <view class="product-bottom">
              <text class="product-price">¥{{ product.min_price }}起</text>
              <text v-if="product.order_count > 0" class="product-sales">已售{{ product.order_count }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/shared/api/request'

interface Banner {
  id: number
  image_url: string
  title: string
  link: string
}

interface HomepageData {
  banners: Banner[]
  popular_destinations: Array<{ name: string; product_count: number; min_price: number }>
  recommended_products: Array<{
    id: number
    product_name: string
    cover_image?: string
    destination_cities: string[]
    days: number
    nights: number
    min_price: number
    order_count: number
  }>
}

const homepageData = ref<HomepageData | null>(null)
const isLoading = ref(true)

const defaultBanners: Banner[] = [
  { id: 1, image_url: '/static/images/banner1.jpg', title: '暑期特惠·云南6日游', link: '' },
  { id: 2, image_url: '/static/images/banner2.jpg', title: '亲子游·北京5日研学之旅', link: '' },
  { id: 3, image_url: '/static/images/banner3.jpg', title: '海岛度假·海南三亚4日游', link: '' },
]

const iconGridItems = [
  { icon: '🏔️', label: '境内跟团游', link: '/pages/products/list' },
  { icon: '✈️', label: '出境跟团游', link: '' },
  { icon: '🚢', label: '邮轮游', link: '' },
  { icon: '🎒', label: '自由行', link: '' },
  { icon: '🏨', label: '酒店+景点', link: '' },
  { icon: '🎫', label: '门票', link: '' },
  { icon: '🚗', label: '当地玩乐', link: '' },
  { icon: '📋', label: '签证', link: '' },
]

const popularDestinations = ref([
  { name: '云南', product_count: 25, min_price: 2999 },
  { name: '海南', product_count: 18, min_price: 1999 },
  { name: '北京', product_count: 20, min_price: 2599 },
  { name: '四川', product_count: 15, min_price: 3299 },
  { name: '广西', product_count: 12, min_price: 2199 },
])

const recommendedProducts = computed(() => homepageData.value?.recommended_products || [])

onMounted(async () => {
  try {
    const data = await api.get<HomepageData>('/homepage')
    homepageData.value = data
  } catch (e) {
    console.error('Failed to load homepage', e)
  } finally {
    isLoading.value = false
  }
})

const goToSearch = () => {
  uni.navigateTo({ url: '/pages/products/list' })
}

const navigateTo = (link: string) => {
  if (link) uni.navigateTo({ url: link })
}

const goToProducts = (dest: string) => {
  uni.navigateTo({ url: `/pages/products/list?destination=${dest}` })
}

const goToDetail = (id: number) => {
  uni.navigateTo({ url: `/pages/products/detail?id=${id}` })
}
</script>

<style scoped>
.home {
  background: #f5f5f5;
  min-height: 100vh;
}

.search-bar {
  display: flex;
  align-items: center;
  gap: 8rpx;
  padding: 16rpx 24rpx;
  background: #ff5722;
}

.search-icon {
  font-size: 28rpx;
}

.search-placeholder {
  color: rgba(255, 255, 255, 0.8);
  font-size: 28rpx;
}

.banner {
  background: #fff;
}

.banner-item {
  position: relative;
  height: 300rpx;
}

.banner-img {
  width: 100%;
  height: 300rpx;
}

.banner-title {
  position: absolute;
  bottom: 20rpx;
  left: 20rpx;
  color: #fff;
  font-size: 32rpx;
  font-weight: bold;
  text-shadow: 0 2rpx 4rpx rgba(0, 0, 0, 0.5);
}

.icon-grid {
  display: flex;
  flex-wrap: wrap;
  padding: 24rpx 16rpx;
  background: #fff;
}

.icon-item {
  width: 25%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
  margin-bottom: 16rpx;
}

.icon-emoji {
  font-size: 48rpx;
}

.icon-label {
  font-size: 24rpx;
  color: #333;
}

.section {
  margin-top: 16rpx;
  background: #fff;
  padding: 24rpx;
}

.section-header {
  margin-bottom: 16rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.dest-scroll {
  white-space: nowrap;
}

.dest-card {
  display: inline-block;
  width: 240rpx;
  padding: 20rpx;
  margin-right: 16rpx;
  background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%);
  border-radius: 12rpx;
}

.dest-name {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  display: block;
}

.dest-info {
  font-size: 22rpx;
  color: #666;
  display: block;
  margin-top: 8rpx;
}

.product-list {
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
  padding-bottom: 16rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.product-img {
  width: 240rpx;
  height: 180rpx;
  border-radius: 8rpx;
}

.product-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.product-name {
  font-size: 28rpx;
  font-weight: 500;
  color: #333;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.product-dest {
  font-size: 24rpx;
  color: #666;
}

.product-days {
  font-size: 24rpx;
  color: #999;
}

.product-bottom {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.product-price {
  font-size: 32rpx;
  font-weight: bold;
  color: #ff5722;
}

.product-sales {
  font-size: 22rpx;
  color: #999;
}

.loading {
  text-align: center;
  padding: 40rpx;
  color: #999;
}
</style>

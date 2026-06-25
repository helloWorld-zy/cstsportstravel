<template>
  <div class="home-page">
    <!-- Search Bar -->
    <div class="search-section">
      <div class="search-box" @click="goToSearch">
        <span class="search-icon">🔍</span>
        <span class="search-placeholder">搜索目的地、产品名称</span>
      </div>
    </div>

    <!-- Banner Carousel -->
    <div class="banner-section">
      <div class="banner-carousel">
        <div
          v-for="(banner, idx) in homepageData?.banners || defaultBanners"
          :key="banner.id"
          class="banner-slide"
          :class="{ active: currentBanner === idx }"
        >
          <a :href="banner.link" class="banner-link">
            <div class="banner-image" :style="{ backgroundImage: `url(${banner.image_url})` }">
              <span class="banner-title">{{ banner.title }}</span>
            </div>
          </a>
        </div>
        <div class="banner-dots">
          <span
            v-for="(_, idx) in homepageData?.banners || defaultBanners"
            :key="idx"
            class="dot"
            :class="{ active: currentBanner === idx }"
            @click="currentBanner = idx"
          />
        </div>
      </div>
    </div>

    <!-- 金刚区 Icon Grid -->
    <div class="icon-grid-section">
      <div class="icon-grid">
        <a v-for="item in iconGridItems" :key="item.label" :href="item.link" class="icon-item">
          <span class="icon-emoji">{{ item.icon }}</span>
          <span class="icon-label">{{ item.label }}</span>
        </a>
      </div>
    </div>

    <!-- Popular Destinations -->
    <div class="section popular-section">
      <div class="section-header">
        <h2 class="section-title">热门目的地</h2>
      </div>
      <div class="dest-tabs">
        <span
          v-for="dest in popularDestinations"
          :key="dest.name"
          class="dest-tab"
          :class="{ active: activeDestTab === dest.name }"
          @click="activeDestTab = dest.name"
        >
          {{ dest.name }}
        </span>
      </div>
      <div class="dest-content">
        <div v-if="activeDestData" class="dest-card">
          <div class="dest-info">
            <h3>{{ activeDestData.name }}</h3>
            <p>{{ activeDestData.product_count }}条线路 · ¥{{ activeDestData.min_price }}起</p>
            <a :href="`/products?destination=${activeDestData.name}`" class="dest-link">去看看 →</a>
          </div>
        </div>
      </div>
    </div>

    <!-- Recommended Products -->
    <div class="section recommend-section">
      <div class="section-header">
        <h2 class="section-title">猜你喜欢</h2>
      </div>
      <div v-if="isLoading" class="loading-skeleton">
        <div v-for="i in 6" :key="i" class="skeleton-card" />
      </div>
      <div v-else-if="recommendedProducts.length" class="product-grid">
        <ProductCard
          v-for="product in recommendedProducts"
          :key="product.id"
          :product="product"
        />
      </div>
      <div v-else class="empty-state">
        <p>暂无推荐产品</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { useApi } from '~/composables/useApi'
import type { ProductSummary } from '~/composables/useProduct'

useHead({
  title: '首页 - 境内跟团游',
})

const router = useRouter()
const api = useApi()

// Homepage data
interface HomepageData {
  banners: Array<{ id: number; image_url: string; title: string; link: string }>
  categories: Array<{ id: number; name: string; icon_url?: string }>
  popular_destinations: Array<{ name: string; product_count: number; min_price: number }>
  recommended_products: ProductSummary[]
}

const { data: homepageData, isLoading } = useQuery({
  queryKey: ['homepage'],
  queryFn: () => api.get<HomepageData>('/homepage'),
  staleTime: 5 * 60 * 1000,
})

// Banner carousel
const currentBanner = ref(0)
let bannerTimer: ReturnType<typeof setInterval> | null = null

const defaultBanners = [
  { id: 1, image_url: '/static/images/banner1.jpg', title: '暑期特惠·云南6日游', link: '/products?destination=云南' },
  { id: 2, image_url: '/static/images/banner2.jpg', title: '亲子游·北京5日研学之旅', link: '/products?destination=北京' },
  { id: 3, image_url: '/static/images/banner3.jpg', title: '海岛度假·海南三亚4日游', link: '/products?destination=海南' },
]

onMounted(() => {
  bannerTimer = setInterval(() => {
    const total = homepageData.value?.banners?.length || defaultBanners.length
    currentBanner.value = (currentBanner.value + 1) % total
  }, 4000)
})

onUnmounted(() => {
  if (bannerTimer) clearInterval(bannerTimer)
})

// 金刚区
const iconGridItems = [
  { icon: '🏔️', label: '境内跟团游', link: '/products' },
  { icon: '✈️', label: '出境跟团游', link: '#' },
  { icon: '🚢', label: '邮轮游', link: '#' },
  { icon: '🎒', label: '自由行', link: '#' },
  { icon: '🏨', label: '酒店+景点', link: '#' },
  { icon: '🎫', label: '门票', link: '#' },
  { icon: '🚗', label: '当地玩乐', link: '#' },
  { icon: '📋', label: '签证', link: '#' },
]

// Popular destinations
const activeDestTab = ref('云南')

const popularDestinations = computed(() =>
  homepageData.value?.popular_destinations || [
    { name: '云南', product_count: 25, min_price: 2999 },
    { name: '海南', product_count: 18, min_price: 1999 },
    { name: '北京', product_count: 20, min_price: 2599 },
    { name: '四川', product_count: 15, min_price: 3299 },
    { name: '广西', product_count: 12, min_price: 2199 },
  ]
)

const activeDestData = computed(() =>
  popularDestinations.value.find(d => d.name === activeDestTab.value)
)

// Recommended products
const recommendedProducts = computed(() => homepageData.value?.recommended_products || [])

const goToSearch = () => {
  router.push('/products')
}
</script>

<style scoped>
.home-page {
  max-width: 768px;
  margin: 0 auto;
  background: #f5f5f5;
  min-height: 100vh;
}

.search-section {
  padding: 12px 16px;
  background: #ff5722;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fff;
  border-radius: 20px;
  padding: 8px 16px;
  cursor: pointer;
}

.search-icon {
  font-size: 16px;
}

.search-placeholder {
  color: #999;
  font-size: 14px;
}

.banner-section {
  position: relative;
}

.banner-carousel {
  position: relative;
  overflow: hidden;
}

.banner-slide {
  display: none;
}

.banner-slide.active {
  display: block;
}

.banner-link {
  display: block;
  text-decoration: none;
}

.banner-image {
  height: 180px;
  background-size: cover;
  background-position: center;
  background-color: #e0e0e0;
  display: flex;
  align-items: flex-end;
  padding: 16px;
}

.banner-title {
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.5);
}

.banner-dots {
  position: absolute;
  bottom: 12px;
  right: 16px;
  display: flex;
  gap: 6px;
}

.banner-dots .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.5);
  cursor: pointer;
}

.banner-dots .dot.active {
  background: #fff;
}

.icon-grid-section {
  background: #fff;
  padding: 16px;
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  text-decoration: none;
  color: #333;
}

.icon-emoji {
  font-size: 28px;
}

.icon-label {
  font-size: 12px;
}

.section {
  margin-top: 8px;
  background: #fff;
  padding: 16px;
}

.section-header {
  margin-bottom: 12px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.dest-tabs {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding-bottom: 8px;
  margin-bottom: 12px;
}

.dest-tab {
  white-space: nowrap;
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 13px;
  color: #666;
  cursor: pointer;
  background: #f5f5f5;
}

.dest-tab.active {
  background: #ff5722;
  color: #fff;
}

.dest-card {
  background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%);
  border-radius: 8px;
  padding: 20px;
}

.dest-info h3 {
  font-size: 18px;
  margin: 0 0 4px;
  color: #333;
}

.dest-info p {
  font-size: 13px;
  color: #666;
  margin: 0 0 8px;
}

.dest-link {
  font-size: 13px;
  color: #ff5722;
  text-decoration: none;
  font-weight: 500;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.loading-skeleton {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.skeleton-card {
  height: 200px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 8px;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.empty-state {
  text-align: center;
  padding: 32px;
  color: #999;
}
</style>

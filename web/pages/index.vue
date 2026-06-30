<template>
  <div class="home-page">
    <div class="hero-search-container">
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
                <span class="banner-title" v-if="!searchFocused">{{ banner.title }}</span>
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

      <!-- Search Bar with Autocomplete -->
      <div class="search-section">
        <h2 class="search-title hide-on-mobile">开启您的完美旅程</h2>
        <p class="search-subtitle hide-on-mobile">高品质户外跟团游 · 体育旅行探索</p>
        <div class="search-box" :class="{ focused: searchFocused }">
          <span class="search-icon">🔍</span>
          <input
            v-model="searchQuery"
            type="text"
            class="search-input"
            placeholder="搜索目的地、产品名称..."
            @focus="searchFocused = true"
            @blur="handleSearchBlur"
            @input="handleSearchInput"
          />
          <span v-if="searchQuery" class="search-clear" @click="clearSearch">✕</span>
        </div>

        <!-- Autocomplete Dropdown -->
        <div v-if="showSuggestions && suggestions.length" class="suggestions-dropdown">
          <div
            v-for="(item, idx) in suggestions"
            :key="idx"
            class="suggestion-item"
            @mousedown.prevent="selectSuggestion(item)"
          >
            <span class="suggestion-icon">{{ item.type === 'destination' ? '📍' : item.type === 'product' ? '🧳' : '🏔️' }}</span>
            <span class="suggestion-text">{{ item.text }}</span>
            <span v-if="item.type" class="suggestion-type">{{ typeLabels[item.type] || '' }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 金刚区 Icon Grid -->
    <div class="icon-grid-section">
      <div class="icon-grid">
        <a v-for="item in iconGridItems" :key="item.label" :href="item.link" class="icon-item">
          <div class="icon-circle" :style="{ backgroundColor: item.bgColor, color: item.color }">
            <span class="icon-svg-wrapper" v-html="item.icon"></span>
          </div>
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
          <div v-if="activeDestData.cover_image" class="dest-cover" :style="{ backgroundImage: `url(${activeDestData.cover_image})` }" />
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
  popular_destinations: Array<{ id?: number; name: string; cover_image?: string; product_count: number; min_price: number }>
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
  {
    icon: `<svg viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"><path d="m8 3 4 8 5-5 5 15H2L8 3z"/></svg>`,
    label: '境内跟团游',
    link: '/products',
    color: '#2563eb', // Brand Blue
    bgColor: '#eff6ff'
  },
  {
    icon: `<svg viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-2-2h-3L9 1H7v5H4a2 2 0 0 0-2 2v1a2 2 0 0 0 2 2h3v5l7 5h2v-5h3a2 2 0 0 0 2-2z"/></svg>`,
    label: '出境跟团游',
    link: '#',
    color: '#3b82f6', // Premium blue
    bgColor: '#eff6ff'
  },
  {
    icon: `<svg viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"><path d="M2 21h20M19.3 14.8C21.1 13.5 22 11.7 22 10V4h-3v3H5V4H2v6c0 1.7.9 3.5 2.7 4.8L2 19h20l-2.7-4.2z"/></svg>`,
    label: '邮轮游',
    link: '#',
    color: '#06b6d4', // Cyan
    bgColor: '#ecfeff'
  },
  {
    icon: `<svg viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"/></svg>`,
    label: '自由行',
    link: '#',
    color: '#f59e0b', // Amber/gold
    bgColor: '#fffbeb'
  },
  {
    icon: `<svg viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"><path d="M2 22V4a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v18M6 12H4M10 12H8M16 12h-2M20 12h-2M6 16H4M10 16H8M16 16h-2M20 16h-2M6 8H4M10 8H8M16 8h-2M20 8h-2"/></svg>`,
    label: '酒店+景点',
    link: '#',
    color: '#ec4899', // Pink
    bgColor: '#fdf2f8'
  },
  {
    icon: `<svg viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"><path d="M2 9a3 3 0 0 1 0 6v3a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-3a3 3 0 0 1 0-6V6a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2v3z"/></svg>`,
    label: '门票',
    link: '#',
    color: '#8b5cf6', // Purple
    bgColor: '#f5f3ff'
  },
  {
    icon: `<svg viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"><polygon points="3 6 9 3 15 6 21 3 21 18 15 21 9 18 3 21"/><line x1="9" y1="3" x2="9" y2="18"/><line x1="15" y1="6" x2="15" y2="21"/></svg>`,
    label: '当地玩乐',
    link: '#',
    color: '#10b981', // Green
    bgColor: '#ecfdf5'
  },
  {
    icon: `<svg viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><rect x="8" y="2" width="8" height="4" rx="1" ry="1"/></svg>`,
    label: '签证',
    link: '#',
    color: '#64748b', // Slate
    bgColor: '#f1f5f9'
  },
]

// Popular destinations
const activeDestTab = ref('')

const popularDestinations = computed(() =>
  homepageData.value?.popular_destinations || [
    { name: '云南', product_count: 25, min_price: 2999 },
    { name: '海南', product_count: 18, min_price: 1999 },
    { name: '北京', product_count: 20, min_price: 2599 },
    { name: '四川', product_count: 15, min_price: 3299 },
    { name: '广西', product_count: 12, min_price: 2199 },
  ]
)

// Set initial active tab when data loads
watch(popularDestinations, (dests) => {
  if (!activeDestTab.value && dests.length) {
    activeDestTab.value = dests[0].name
  }
}, { immediate: true })

const activeDestData = computed(() =>
  popularDestinations.value.find(d => d.name === activeDestTab.value)
)

// Recommended products
const recommendedProducts = computed(() => homepageData.value?.recommended_products || [])

// Search autocomplete
interface SuggestItem {
  text: string
  type: 'destination' | 'product' | 'spot'
}

const searchQuery = ref('')
const searchFocused = ref(false)
const suggestions = ref<SuggestItem[]>([])
const showSuggestions = ref(false)
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const typeLabels: Record<string, string> = {
  destination: '目的地',
  product: '产品',
  spot: '景点',
}

function handleSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)

  if (!searchQuery.value.trim()) {
    suggestions.value = []
    showSuggestions.value = false
    return
  }

  debounceTimer = setTimeout(async () => {
    try {
      const result = await api.get<SuggestItem[]>('/search/autocomplete', {
        params: { q: searchQuery.value.trim(), limit: 8 },
      })
      suggestions.value = result || []
      showSuggestions.value = suggestions.value.length > 0
    } catch {
      suggestions.value = []
    }
  }, 300) // 300ms debounce
}

function handleSearchBlur() {
  // Delay hiding to allow click on suggestion
  setTimeout(() => {
    searchFocused.value = false
    showSuggestions.value = false
  }, 200)
}

function selectSuggestion(item: SuggestItem) {
  searchQuery.value = item.text
  showSuggestions.value = false
  router.push(`/products?keyword=${encodeURIComponent(item.text)}`)
}

function clearSearch() {
  searchQuery.value = ''
  suggestions.value = []
  showSuggestions.value = false
}
</script>

<style scoped>
.home-page {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 16px 40px;
  background: transparent;
  min-height: calc(100vh - 70px);
}

/* Hero & Search overlay container */
.hero-search-container {
  position: relative;
  width: 100%;
  margin-top: 16px;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05), 0 4px 6px -4px rgba(0, 0, 0, 0.05);
}

/* Banner Carousel */
.banner-section {
  position: relative;
  z-index: 1;
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
  height: 200px;
  background-size: cover;
  background-position: center;
  background-color: #e2e8f0;
  display: flex;
  align-items: flex-end;
  padding: 24px;
  position: relative;
}

.banner-image::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.1) 0%, rgba(15, 23, 42, 0.6) 100%);
  pointer-events: none;
}

.banner-title {
  color: #fff;
  font-size: 20px;
  font-weight: 700;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  z-index: 2;
  position: relative;
}

.banner-dots {
  position: absolute;
  bottom: 16px;
  right: 24px;
  display: flex;
  gap: 8px;
  z-index: 3;
}

.banner-dots .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.4);
  cursor: pointer;
  transition: all 0.2s;
}

.banner-dots .dot.active {
  background: #fff;
  width: 18px;
  border-radius: 4px;
}

/* Search Section */
.search-section {
  position: relative;
  padding: 16px;
  background: #2563eb;
  z-index: 2;
}

.search-title {
  font-size: 24px;
  font-weight: 800;
  color: #0f172a;
  margin: 0 0 6px 0;
  letter-spacing: -0.5px;
}

.search-subtitle {
  font-size: 14px;
  color: #475569;
  margin: 0 0 24px 0;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #fff;
  border-radius: 12px;
  padding: 10px 16px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 2px solid transparent;
}

.search-box.focused {
  border-color: #2563eb;
  box-shadow: 0 10px 15px -3px rgba(37, 99, 235, 0.1), 0 4px 6px -4px rgba(37, 99, 235, 0.1);
}

.search-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 15px;
  color: #1e293b;
  background: transparent;
  font-weight: 500;
}

.search-input::placeholder {
  color: #94a3b8;
}

.search-clear {
  cursor: pointer;
  color: #94a3b8;
  font-size: 14px;
  padding: 4px;
  transition: color 0.2s;
}

.search-clear:hover {
  color: #475569;
}

/* Suggestions Dropdown */
.suggestions-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  left: 16px;
  right: 16px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
  z-index: 100;
  max-height: 320px;
  overflow-y: auto;
  border: 1px solid #f1f5f9;
  padding: 6px;
}

.suggestion-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  cursor: pointer;
  border-radius: 8px;
  transition: background 0.2s;
}

.suggestion-item:hover {
  background: #f8fafc;
}

.suggestion-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.suggestion-text {
  flex: 1;
  font-size: 14px;
  color: #334155;
  font-weight: 500;
}

.suggestion-type {
  font-size: 11px;
  color: #2563eb;
  background: #eff6ff;
  padding: 2px 8px;
  border-radius: 6px;
  font-weight: 600;
}

/* Icon Grid - Navigation Categories */
.icon-grid-section {
  background: transparent;
  padding: 24px 0;
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  text-decoration: none;
  color: #1e293b;
  background: #fff;
  padding: 16px 8px;
  border-radius: 16px;
  border: 1px solid rgba(226, 232, 240, 0.8);
  box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02), 0 2px 4px -2px rgba(0,0,0,0.02);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
}

.icon-item:hover {
  transform: translateY(-4px);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05), 0 4px 6px -4px rgba(0, 0, 0, 0.05);
  border-color: #2563eb;
  color: #2563eb;
}

.icon-circle {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.03);
}

.icon-item:hover .icon-circle {
  transform: scale(1.1);
  box-shadow: 0 8px 12px -3px rgba(0, 0, 0, 0.08);
}

.icon-svg-wrapper {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-svg-wrapper :deep(svg) {
  width: 100%;
  height: 100%;
}

.icon-label {
  font-size: 13px;
  font-weight: 600;
}

/* Sections */
.section {
  margin-top: 32px;
  background: transparent;
  padding: 0;
}

.section-header {
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.section-title {
  font-size: 22px;
  font-weight: 800;
  color: #0f172a;
  letter-spacing: -0.5px;
  position: relative;
  padding-left: 12px;
}

.section-title::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 4px;
  height: 20px;
  background: linear-gradient(180deg, #1d4ed8 0%, #3b82f6 100%);
  border-radius: 2px;
}

/* Popular Destinations */
.dest-tabs {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 12px;
  margin-bottom: 16px;
  scrollbar-width: none; /* Firefox */
  -webkit-overflow-scrolling: touch;
}

.dest-tabs::-webkit-scrollbar {
  display: none; /* Safari and Chrome */
}

.dest-tab {
  white-space: nowrap;
  padding: 8px 18px;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  background: #fff;
  border: 1px solid rgba(226, 232, 240, 0.8);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.01);
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.dest-tab:hover {
  color: #2563eb;
  background: #eff6ff;
  border-color: rgba(37, 99, 235, 0.3);
}

.dest-tab.active {
  background: #2563eb;
  color: #fff;
  border-color: #2563eb;
  box-shadow: 0 4px 10px rgba(37, 99, 235, 0.2);
}

.dest-card {
  background: #fff;
  border-radius: 20px;
  overflow: hidden;
  border: 1px solid rgba(226, 232, 240, 0.8);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.02), 0 4px 6px -4px rgba(0, 0, 0, 0.02);
  transition: transform 0.3s ease;
}

.dest-cover {
  height: 180px;
  background-size: cover;
  background-position: center;
}

.dest-info {
  padding: 24px;
  background: #fff;
}

.dest-info h3 {
  font-size: 20px;
  font-weight: 800;
  margin: 0 0 6px;
  color: #0f172a;
}

.dest-info p {
  font-size: 14px;
  color: #64748b;
  margin: 0 0 16px;
  font-weight: 500;
}

.dest-link {
  font-size: 14px;
  color: #2563eb;
  text-decoration: none;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  transition: transform 0.2s;
}

.dest-link:hover {
  transform: translateX(4px);
}

/* Products */
.product-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

/* Skeletons */
.loading-skeleton {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.skeleton-card {
  height: 260px;
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 16px;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.empty-state {
  text-align: center;
  padding: 48px;
  background: #fff;
  border-radius: 20px;
  color: #64748b;
  font-weight: 500;
  border: 1px solid rgba(226, 232, 240, 0.8);
}

/* Responsive adjustments */
@media (min-width: 640px) {
  .product-grid, .loading-skeleton {
    grid-template-columns: repeat(2, 1fr);
  }

  .banner-image {
    height: 280px;
  }

  .icon-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (min-width: 1024px) {
  .home-page {
    padding: 24px 24px 60px;
  }

  /* Desktop Hero & Search Layout */
  .hero-search-container {
    height: 400px;
    margin-top: 8px;
  }

  .banner-section {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: 1;
  }

  .banner-image {
    height: 400px;
  }

  .banner-title {
    font-size: 28px;
  }

  .search-section {
    position: absolute;
    top: 50%;
    left: 48px;
    transform: translateY(-50%);
    width: 380px;
    padding: 24px;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(8px);
    border-radius: 16px;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.6);
    z-index: 5;
  }

  .suggestions-dropdown {
    left: 0;
    right: 0;
    width: 100%;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.08);
  }

  .icon-grid-section {
    padding: 32px 0 16px;
  }

  .icon-grid {
    grid-template-columns: repeat(8, 1fr);
    gap: 16px;
  }

  /* Desktop Popular Destinations */
  .dest-card {
    display: flex;
    flex-direction: row;
    height: 280px;
  }

  .dest-cover {
    width: 60%;
    height: 100%;
  }

  .dest-info {
    width: 40%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: 40px;
  }

  .dest-info h3 {
    font-size: 26px;
    margin-bottom: 8px;
  }

  .dest-info p {
    font-size: 16px;
    margin-bottom: 24px;
  }

  /* Desktop Products list */
  .product-grid, .loading-skeleton {
    grid-template-columns: repeat(4, 1fr);
    gap: 20px;
  }
}
</style>

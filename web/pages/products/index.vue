<template>
  <div class="product-list-page">
    <!-- Filter Bar -->
    <div class="filter-bar">
      <div class="filter-tags">
        <button
          v-for="tag in filterTags"
          :key="tag.key"
          class="filter-tag"
          :class="{ active: isFilterActive(tag.key) }"
          @click="openFilter(tag.key)"
        >
          {{ getFilterLabel(tag) }}
        </button>
      </div>
      <button class="filter-more-btn" @click="showDrawer = true">
        筛选
      </button>
    </div>

    <!-- Sort Bar -->
    <div class="sort-bar">
      <button
        v-for="opt in sortOptions"
        :key="opt.value"
        class="sort-btn"
        :class="{ active: params.sort === opt.value }"
        @click="params.sort = opt.value; params.page = 1"
      >
        {{ opt.label }}
      </button>
      <div class="view-toggle">
        <button :class="{ active: viewMode === 'grid' }" @click="viewMode = 'grid'">网格</button>
        <button :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'">列表</button>
      </div>
    </div>

    <!-- Product List -->
    <div v-if="isLoading" class="loading-state">
      <div v-for="i in 6" :key="i" class="skeleton-card" />
    </div>

    <div v-else-if="products.length === 0" class="empty-state">
      <p>暂无符合条件的产品</p>
      <button class="reset-btn" @click="resetFilters">重置筛选条件</button>
    </div>

    <div v-else :class="['product-grid', viewMode]">
      <ProductCard v-for="product in products" :key="product.id" :product="product" />
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="pagination">
      <button :disabled="params.page <= 1" @click="params.page--">上一页</button>
      <span>第 {{ params.page }} / {{ totalPages }} 页</span>
      <button :disabled="params.page >= totalPages" @click="params.page++">下一页</button>
    </div>

    <!-- Filter Drawer -->
    <div v-if="showDrawer" class="drawer-overlay" @click.self="showDrawer = false">
      <div class="drawer">
        <div class="drawer-header">
          <h3>筛选</h3>
          <button @click="showDrawer = false">✕</button>
        </div>
        <div class="drawer-body">
          <div class="filter-group">
            <h4>目的地</h4>
            <div class="filter-options">
              <span
                v-for="dest in destinationOptions"
                :key="dest"
                class="filter-option"
                :class="{ selected: params.destination === dest }"
                @click="params.destination = params.destination === dest ? '' : dest"
              >
                {{ dest }}
              </span>
            </div>
          </div>
          <div class="filter-group">
            <h4>出发城市</h4>
            <div class="filter-options">
              <span
                v-for="city in originOptions"
                :key="city"
                class="filter-option"
                :class="{ selected: params.origin === city }"
                @click="params.origin = params.origin === city ? '' : city"
              >
                {{ city }}
              </span>
            </div>
          </div>
          <div class="filter-group">
            <h4>行程天数</h4>
            <div class="filter-options">
              <span
                v-for="range in daysOptions"
                :key="range.label"
                class="filter-option"
                :class="{ selected: params.days_min === range.min && params.days_max === range.max }"
                @click="setDaysFilter(range)"
              >
                {{ range.label }}
              </span>
            </div>
          </div>
          <div class="filter-group">
            <h4>价格区间</h4>
            <div class="filter-options">
              <span
                v-for="range in priceOptions"
                :key="range.label"
                class="filter-option"
                :class="{ selected: params.price_min === range.min && params.price_max === range.max }"
                @click="setPriceFilter(range)"
              >
                {{ range.label }}
              </span>
            </div>
          </div>
          <div class="filter-group">
            <h4>产品等级</h4>
            <div class="filter-options">
              <span
                v-for="grade in gradeOptions"
                :key="grade.value"
                class="filter-option"
                :class="{ selected: params.product_grade === grade.value }"
                @click="params.product_grade = params.product_grade === grade.value ? '' : grade.value"
              >
                {{ grade.label }}
              </span>
            </div>
          </div>
          <!-- CHK011: Additional filter groups (PRD F-I-L06, F-I-L07, F-I-L09) -->
          <div class="filter-group">
            <h4>住宿标准</h4>
            <div class="filter-options">
              <span
                v-for="opt in accommodationOptions"
                :key="opt.value"
                class="filter-option"
                :class="{ selected: params.accommodation_standard === opt.value }"
                @click="params.accommodation_standard = params.accommodation_standard === opt.value ? '' : opt.value"
              >
                {{ opt.label }}
              </span>
            </div>
          </div>
          <div class="filter-group">
            <h4>主题标签</h4>
            <div class="filter-options">
              <span
                v-for="tag in themeOptions"
                :key="tag.value"
                class="filter-option"
                :class="{ selected: params.theme_tags === tag.value }"
                @click="params.theme_tags = params.theme_tags === tag.value ? '' : tag.value"
              >
                {{ tag.label }}
              </span>
            </div>
          </div>
          <div class="filter-group">
            <h4>交通工具</h4>
            <div class="filter-options">
              <span
                v-for="opt in transportOptions"
                :key="opt.value"
                class="filter-option"
                :class="{ selected: params.transport_mode === opt.value }"
                @click="params.transport_mode = params.transport_mode === opt.value ? '' : opt.value"
              >
                {{ opt.label }}
              </span>
            </div>
          </div>
        </div>
        <div class="drawer-footer">
          <button class="reset-btn" @click="resetFilters">重置</button>
          <button class="apply-btn" @click="applyFilters">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { useApi } from '~/composables/useApi'
import type { PaginatedProducts, ProductListParams, ProductSummary } from '~/composables/useProduct'

useHead({
  title: '产品列表 - 境内跟团游',
})

const route = useRoute()
const api = useApi()

// State
const viewMode = ref<'grid' | 'list'>('grid')
const showDrawer = ref(false)

const params = reactive<ProductListParams>({
  destination: (route.query.destination as string) || '',
  origin: (route.query.origin as string) || '',
  days_min: route.query.days_min ? Number(route.query.days_min) : undefined,
  days_max: route.query.days_max ? Number(route.query.days_max) : undefined,
  price_min: route.query.price_min ? Number(route.query.price_min) : undefined,
  price_max: route.query.price_max ? Number(route.query.price_max) : undefined,
  category_id: route.query.category_id ? Number(route.query.category_id) : undefined,
  product_grade: (route.query.product_grade as string) || '',
  keyword: (route.query.keyword as string) || '',
  sort: (route.query.sort as string) || 'recommended',
  page: 1,
  page_size: 20,
  // CHK011: Additional filter fields (PRD F-I-L06, F-I-L07, F-I-L09)
  accommodation_standard: (route.query.accommodation_standard as string) || '',
  theme_tags: (route.query.theme_tags as string) || '',
  transport_mode: (route.query.transport_mode as string) || '',
})

// Sort options
const sortOptions = [
  { label: '推荐', value: 'recommended' },
  { label: '价格↑', value: 'price_asc' },
  { label: '价格↓', value: 'price_desc' },
  { label: '满意度', value: 'satisfaction' },
  { label: '销量', value: 'sales' },
  { label: '天数↑', value: 'days_asc' },
]

// Filter options
const destinationOptions = ['云南', '海南', '北京', '四川', '广西', '西安', '厦门', '张家界']
const originOptions = ['上海', '北京', '广州', '深圳', '杭州', '成都', '南京']
const daysOptions = [
  { label: '不限', min: undefined, max: undefined },
  { label: '1-3天', min: 1, max: 3 },
  { label: '4-6天', min: 4, max: 6 },
  { label: '7-9天', min: 7, max: 9 },
  { label: '10天以上', min: 10, max: undefined },
]
const priceOptions = [
  { label: '不限', min: undefined, max: undefined },
  { label: '2000以下', min: undefined, max: 2000 },
  { label: '2000-4000', min: 2000, max: 4000 },
  { label: '4000-6000', min: 4000, max: 6000 },
  { label: '6000以上', min: 6000, max: undefined },
]
// CHK011: Additional filter options (PRD F-I-L06, F-I-L07, F-I-L09)
const accommodationOptions = [
  { label: '不限', value: '' },
  { label: '经济型', value: 'economy' },
  { label: '舒适型', value: 'comfort' },
  { label: '豪华型', value: 'luxury' },
  { label: '五星', value: 'five_star' },
]
const themeOptions = [
  { label: '亲子', value: 'family' },
  { label: '蜜月', value: 'honeymoon' },
  { label: '摄影', value: 'photography' },
  { label: '美食', value: 'food' },
  { label: '购物', value: 'shopping' },
  { label: '探险', value: 'adventure' },
  { label: '红色旅游', value: 'red_tourism' },
  { label: '康养', value: 'health' },
]
const transportOptions = [
  { label: '不限', value: '' },
  { label: '飞机', value: 'flight' },
  { label: '高铁', value: 'train' },
  { label: '大巴', value: 'bus' },
]
const gradeOptions = [
  { label: '不限', value: '' },
  { label: '经济', value: 'standard' },
  { label: '舒适', value: 'comfort' },
  { label: '豪华', value: 'luxury' },
]

// Filter tags
const filterTags = [
  { key: 'destination', label: '目的地' },
  { key: 'origin', label: '出发城市' },
  { key: 'days', label: '天数' },
  { key: 'price', label: '价格' },
]

function isFilterActive(key: string): boolean {
  switch (key) {
    case 'destination': return !!params.destination
    case 'origin': return !!params.origin
    case 'days': return !!params.days_min || !!params.days_max
    case 'price': return !!params.price_min || !!params.price_max
    default: return false
  }
}

function getFilterLabel(tag: { key: string; label: string }): string {
  switch (tag.key) {
    case 'destination': return params.destination || tag.label
    case 'origin': return params.origin || tag.label
    case 'days':
      if (params.days_min && params.days_max) return `${params.days_min}-${params.days_max}天`
      if (params.days_min) return `${params.days_min}天起`
      if (params.days_max) return `${params.days_max}天内`
      return tag.label
    case 'price':
      if (params.price_min && params.price_max) return `¥${params.price_min}-${params.price_max}`
      if (params.price_min) return `¥${params.price_min}起`
      if (params.price_max) return `¥${params.price_max}内`
      return tag.label
    default: return tag.label
  }
}

function openFilter(key: string) {
  showDrawer.value = true
}

function setDaysFilter(range: { min?: number; max?: number }) {
  params.days_min = range.min
  params.days_max = range.max
}

function setPriceFilter(range: { min?: number; max?: number }) {
  params.price_min = range.min
  params.price_max = range.max
}

function resetFilters() {
  params.destination = ''
  params.origin = ''
  params.days_min = undefined
  params.days_max = undefined
  params.price_min = undefined
  params.price_max = undefined
  params.product_grade = ''
  // CHK011: Reset additional filters
  params.accommodation_standard = ''
  params.theme_tags = ''
  params.transport_mode = ''
  params.page = 1
}

function applyFilters() {
  params.page = 1
  showDrawer.value = false
}

// Fetch products
const { data, isLoading } = useQuery({
  queryKey: ['products', computed(() => ({ ...params }))],
  queryFn: () => api.get<PaginatedProducts>('/products', { params: { ...params } }),
  staleTime: 60 * 1000,
})

const products = computed<ProductSummary[]>(() => data.value?.items || [])
const total = computed(() => data.value?.total || 0)
const totalPages = computed(() => Math.ceil(total.value / (params.page_size || 20)))

onMounted(() => {
  if (typeof window !== 'undefined' && window.innerWidth >= 1024) {
    showDrawer.value = true
  }
})
</script>

<style scoped>
.product-list-page {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px 16px 40px;
  background: transparent;
  min-height: calc(100vh - 70px);
}

/* Filter Bar */
.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f1f5f9;
}

.filter-tags {
  display: flex;
  gap: 8px;
  flex: 1;
  overflow-x: auto;
  scrollbar-width: none;
}

.filter-tags::-webkit-scrollbar {
  display: none;
}

.filter-tag {
  white-space: nowrap;
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

.filter-tag:hover {
  color: #2563eb;
  border-color: rgba(37, 99, 235, 0.3);
}

.filter-tag.active {
  border-color: #2563eb;
  color: #2563eb;
  background: #eff6ff;
}

.filter-more-btn {
  padding: 6px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  background: #fff;
  font-size: 13px;
  font-weight: 600;
  color: #2563eb;
  border-color: rgba(37, 99, 235, 0.4);
  cursor: pointer;
  transition: all 0.2s;
}

.filter-more-btn:hover {
  background: #eff6ff;
}

/* Sort Bar */
.sort-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
}

.sort-btn {
  padding: 6px 12px;
  background: none;
  border: none;
  font-size: 13px;
  color: #64748b;
  font-weight: 600;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
}

.sort-btn:hover {
  color: #2563eb;
}

.sort-btn.active {
  color: #2563eb;
  background: #eff6ff;
}

.view-toggle {
  margin-left: auto;
  display: flex;
  gap: 4px;
  background: #f1f5f9;
  padding: 3px;
  border-radius: 8px;
}

.view-toggle button {
  padding: 4px 10px;
  border: none;
  background: transparent;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
}

.view-toggle button.active {
  background: #fff;
  color: #2563eb;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

/* Product list area */
.product-grid {
  display: grid;
  gap: 16px;
  padding: 16px 0;
}

.product-grid.grid {
  grid-template-columns: repeat(2, 1fr);
}

.product-grid.list {
  grid-template-columns: 1fr;
}

.product-grid.list :deep(.product-card) {
  display: flex;
  flex-direction: row;
  height: 160px;
}

.product-grid.list :deep(.card-image) {
  width: 240px;
  min-width: 240px;
  padding-top: 0;
  height: 100%;
}

.product-grid.list :deep(.card-body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 16px 20px;
}

.product-grid.list :deep(.card-title) {
  height: auto;
  margin-bottom: 6px;
}

.loading-state {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  padding: 16px 0;
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
  padding: 60px 24px;
  background: #fff;
  border-radius: 16px;
  color: #64748b;
  border: 1px solid rgba(226, 232, 240, 0.8);
}

.reset-btn {
  margin-top: 16px;
  padding: 10px 28px;
  border: 1px solid #2563eb;
  border-radius: 20px;
  background: #fff;
  color: #2563eb;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}

.reset-btn:hover {
  background: #eff6ff;
}

/* Pagination */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  padding: 32px 0;
}

.pagination button {
  padding: 8px 18px;
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  background: #fff;
  font-weight: 600;
  color: #475569;
  cursor: pointer;
  transition: all 0.2s;
}

.pagination button:hover:not(:disabled) {
  color: #2563eb;
  border-color: rgba(37, 99, 235, 0.3);
}

.pagination button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.pagination span {
  font-size: 14px;
  color: #64748b;
  font-weight: 500;
}

/* Drawer */
.drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(2px);
  z-index: 1100;
  display: flex;
  justify-content: flex-end;
}

.drawer {
  width: 320px;
  background: #fff;
  display: flex;
  flex-direction: column;
  box-shadow: -10px 0 25px rgba(0, 0, 0, 0.1);
  height: 100%;
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #f1f5f9;
}

.drawer-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 800;
  color: #0f172a;
}

.drawer-header button {
  background: none;
  border: none;
  font-size: 20px;
  color: #94a3b8;
  cursor: pointer;
  transition: color 0.2s;
}

.drawer-header button:hover {
  color: #475569;
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.filter-group {
  margin-bottom: 24px;
}

.filter-group h4 {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

.filter-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-option {
  padding: 6px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-option:hover {
  border-color: rgba(37, 99, 235, 0.3);
  color: #2563eb;
}

.filter-option.selected {
  border-color: #2563eb;
  color: #2563eb;
  background: #eff6ff;
  font-weight: 600;
}

.drawer-footer {
  display: flex;
  gap: 12px;
  padding: 20px;
  border-top: 1px solid #f1f5f9;
  background: #fff;
}

.drawer-footer button {
  flex: 1;
  padding: 12px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.drawer-footer .reset-btn {
  margin-top: 0;
  border: 1px solid #e2e8f0;
  color: #475569;
}

.drawer-footer .reset-btn:hover {
  background: #f8fafc;
}

.apply-btn {
  background: linear-gradient(135deg, #1d4ed8 0%, #3b82f6 100%);
  color: #fff;
  border: none;
  box-shadow: 0 4px 10px rgba(37, 99, 235, 0.2);
}

.apply-btn:hover {
  opacity: 0.95;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .product-grid.grid {
    grid-template-columns: 1fr;
  }
}

@media (min-width: 1024px) {
  .product-list-page {
    display: grid;
    grid-template-columns: 260px 1fr;
    gap: 24px;
    padding: 24px 24px 60px;
  }

  .filter-bar {
    grid-column: 2;
    grid-row: 1;
    border: none;
    padding: 0;
    margin-bottom: 8px;
  }

  .sort-bar {
    grid-column: 2;
    grid-row: 2;
    padding: 12px 20px;
    background: #fff;
    border-radius: 16px;
    border: 1px solid rgba(226, 232, 240, 0.8);
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02);
  }

  .product-grid {
    grid-column: 2;
    grid-row: 3;
    padding: 0;
  }

  .product-grid.grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
  }

  .loading-state {
    grid-column: 2;
    grid-row: 3;
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
    padding: 0;
  }

  .empty-state {
    grid-column: 2;
    grid-row: 3;
  }

  .pagination {
    grid-column: 2;
    grid-row: 4;
  }

  /* Static Left Sidebar Panel */
  .drawer-overlay {
    grid-column: 1;
    grid-row: 1 / span 4;
    display: block !important;
    position: sticky;
    top: 94px;
    height: calc(100vh - 120px);
    background: transparent;
    backdrop-filter: none;
    z-index: 10;
    width: 100%;
  }

  .drawer {
    width: 100% !important;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px -2px rgba(0, 0, 0, 0.02);
    border: 1px solid rgba(226, 232, 240, 0.8);
    border-radius: 16px;
    height: 100%;
  }

  .drawer-header button {
    display: none; /* No close button on desktop */
  }

  .drawer-body {
    padding: 20px;
  }
}
</style>

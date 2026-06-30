<template>
  <div class="outbound-list-page">
    <!-- Header -->
    <div class="page-header">
      <h1>出境游</h1>
      <p>探索世界，从这里开始</p>
    </div>

    <!-- Continent Tabs -->
    <div class="continent-tabs">
      <button
        v-for="continent in continents"
        :key="continent.code"
        class="continent-tab"
        :class="{ active: selectedContinent === continent.code }"
        @click="selectContinent(continent.code)"
      >
        {{ continent.name }}
      </button>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <!-- Country Filter -->
      <div class="filter-group">
        <label>目的地</label>
        <select v-model="filters.country_id" @change="loadProducts">
          <option value="">全部国家</option>
          <option v-for="country in countries" :key="country.id" :value="country.id">
            {{ country.name_cn }}
          </option>
        </select>
      </div>

      <!-- Visa Type Filter -->
      <div class="filter-group">
        <label>签证类型</label>
        <div class="visa-type-tags">
          <button
            v-for="vt in visaTypes"
            :key="vt.value"
            class="visa-tag"
            :class="{ active: filters.visa_type === vt.value }"
            @click="toggleVisaType(vt.value)"
          >
            {{ vt.label }}
          </button>
        </div>
      </div>

      <!-- Days Filter -->
      <div class="filter-group">
        <label>行程天数</label>
        <select v-model="filters.days_range" @change="loadProducts">
          <option value="">不限</option>
          <option value="1-3">1-3天</option>
          <option value="4-6">4-6天</option>
          <option value="7-10">7-10天</option>
          <option value="11-14">11-14天</option>
          <option value="15+">15天以上</option>
        </select>
      </div>

      <!-- Origin City Filter -->
      <div class="filter-group">
        <label>出发城市</label>
        <select v-model="filters.origin_city" @change="loadProducts">
          <option value="">不限</option>
          <option value="北京">北京</option>
          <option value="上海">上海</option>
          <option value="广州">广州</option>
          <option value="深圳">深圳</option>
          <option value="成都">成都</option>
        </select>
      </div>
    </div>

    <!-- Sort Bar -->
    <div class="sort-bar">
      <button
        v-for="opt in sortOptions"
        :key="opt.value"
        class="sort-btn"
        :class="{ active: filters.sort === opt.value }"
        @click="filters.sort = opt.value; loadProducts()"
      >
        {{ opt.label }}
      </button>
    </div>

    <!-- Product List -->
    <div v-if="isLoading" class="loading-state">
      <div v-for="i in 6" :key="i" class="skeleton-card" />
    </div>

    <div v-else-if="products.length === 0" class="empty-state">
      <p>暂无符合条件的出境游产品</p>
      <button class="reset-btn" @click="resetFilters">重置筛选条件</button>
    </div>

    <div v-else class="product-grid">
      <div v-for="product in products" :key="product.id" class="product-card" @click="goToDetail(product.id)">
        <div class="card-image">
          <img :src="product.cover_image || '/images/default-product.jpg'" :alt="product.product_name" />
          <!-- Visa Type Badge -->
          <span class="visa-badge" :class="getVisaBadgeClass(product)">
            {{ getVisaBadgeText(product) }}
          </span>
        </div>
        <div class="card-content">
          <h3 class="card-title">{{ product.product_name }}</h3>
          <div class="card-meta">
            <span class="days">{{ product.days }}天{{ product.nights }}晚</span>
            <span class="origin">{{ product.origin_city }}出发</span>
          </div>
          <div class="card-tags">
            <span v-if="product.destination_country" class="country-tag">
              {{ product.destination_country.name_cn }}
            </span>
            <span v-if="product.visa_info_parsed" class="visa-type-tag">
              {{ product.visa_info_parsed.visa_type }}
            </span>
          </div>
          <div class="card-price">
            <span class="price-label">起</span>
            <span class="price-value">¥{{ getMinPrice(product) }}</span>
            <span class="price-unit">/人</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="pagination">
      <button :disabled="currentPage <= 1" @click="currentPage--; loadProducts()">上一页</button>
      <span>第 {{ currentPage }} / {{ totalPages }} 页</span>
      <button :disabled="currentPage >= totalPages" @click="currentPage++; loadProducts()">下一页</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'

// Types
interface Country {
  id: number
  name_cn: string
  continent: string
  visa_type: string
}

interface ContinentTree {
  continent: string
  continent_name: string
  countries: Country[]
}

interface Product {
  id: number
  product_name: string
  cover_image: string
  days: number
  nights: number
  origin_city: string
  destination_country?: Country
  visa_info_parsed?: {
    visa_type: string
    processing_days: number
    fee: number
  }
  departure_dates?: Array<{
    adult_price: number
  }>
}

// State
const isLoading = ref(false)
const products = ref<Product[]>([])
const continentTree = ref<ContinentTree[]>([])
const countries = ref<Country[]>([])
const currentPage = ref(1)
const pageSize = 20
const total = ref(0)

const filters = reactive({
  continent: '',
  country_id: '',
  visa_type: '',
  origin_city: '',
  days_range: '',
  sort: 'recommended',
})

// Continent definitions
const continents = [
  { code: '', name: '全部' },
  { code: 'asia', name: '亚洲' },
  { code: 'europe', name: '欧洲' },
  { code: 'north_america', name: '北美洲' },
  { code: 'south_america', name: '南美洲' },
  { code: 'oceania', name: '大洋洲' },
  { code: 'africa', name: '非洲' },
]

// Visa type options
const visaTypes = [
  { value: '', label: '全部' },
  { value: 'visa_free', label: '免签' },
  { value: 'visa_on_arrival', label: '落地签' },
  { value: 'e_visa', label: '电子签' },
  { value: 'visa_required', label: '需提前办签' },
]

// Sort options
const sortOptions = [
  { value: 'recommended', label: '推荐' },
  { value: 'price_asc', label: '价格低→高' },
  { value: 'price_desc', label: '价格高→低' },
  { value: 'days_asc', label: '天数少→多' },
  { value: 'days_desc', label: '天数多→少' },
]

// Computed
const totalPages = computed(() => Math.ceil(total.value / pageSize))

// Methods
const loadContinentTree = async () => {
  try {
    const data = await $fetch('/api/v2/products/outbound/continents')
    continentTree.value = data as ContinentTree[]
  } catch (error) {
    console.error('Failed to load continent tree:', error)
  }
}

const loadProducts = async () => {
  isLoading.value = true
  try {
    const params: Record<string, any> = {
      page: currentPage.value,
      page_size: pageSize,
      sort: filters.sort,
    }
    if (filters.continent) params.continent = filters.continent
    if (filters.country_id) params.country_id = filters.country_id
    if (filters.visa_type) params.visa_type = filters.visa_type
    if (filters.origin_city) params.origin_city = filters.origin_city

    // Parse days range
    if (filters.days_range) {
      if (filters.days_range === '15+') {
        params.days_min = 15
      } else {
        const [min, max] = filters.days_range.split('-').map(Number)
        params.days_min = min
        params.days_max = max
      }
    }

    const data = await $fetch('/api/v2/products/outbound', { params })
    const result = data as any
    products.value = result.items || []
    total.value = result.total || 0
  } catch (error) {
    console.error('Failed to load products:', error)
  } finally {
    isLoading.value = false
  }
}

const selectContinent = (code: string) => {
  filters.continent = code
  filters.country_id = ''
  // Update country list based on continent
  if (code) {
    const tree = continentTree.value.find(t => t.continent === code)
    countries.value = tree?.countries || []
  } else {
    countries.value = continentTree.value.flatMap(t => t.countries)
  }
  currentPage.value = 1
  loadProducts()
}

const toggleVisaType = (type: string) => {
  filters.visa_type = filters.visa_type === type ? '' : type
  currentPage.value = 1
  loadProducts()
}

const resetFilters = () => {
  filters.continent = ''
  filters.country_id = ''
  filters.visa_type = ''
  filters.origin_city = ''
  filters.days_range = ''
  filters.sort = 'recommended'
  currentPage.value = 1
  loadProducts()
}

const goToDetail = (id: number) => {
  navigateTo(`/outbound/${id}`)
}

const getVisaBadgeClass = (product: Product) => {
  if (!product.destination_country) return ''
  switch (product.destination_country.visa_type) {
    case 'visa_free': return 'badge-free'
    case 'visa_on_arrival': return 'badge-arrival'
    case 'e_visa': return 'badge-evisa'
    case 'visa_required': return 'badge-required'
    default: return ''
  }
}

const getVisaBadgeText = (product: Product) => {
  if (!product.destination_country) return ''
  switch (product.destination_country.visa_type) {
    case 'visa_free': return '免签'
    case 'visa_on_arrival': return '落地签'
    case 'e_visa': return '电子签'
    case 'visa_required': return '需办签'
    default: return ''
  }
}

const getMinPrice = (product: Product) => {
  if (!product.departure_dates || product.departure_dates.length === 0) return '--'
  const minPrice = Math.min(...product.departure_dates.map(d => d.adult_price))
  return (minPrice / 100).toFixed(0)
}

// Lifecycle
onMounted(() => {
  loadContinentTree()
  loadProducts()
})
</script>

<style scoped>
.outbound-list-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.page-header {
  text-align: center;
  margin-bottom: 30px;
}

.page-header h1 {
  font-size: 32px;
  color: #1a1a1a;
  margin-bottom: 8px;
}

.page-header p {
  color: #666;
  font-size: 16px;
}

.continent-tabs {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  overflow-x: auto;
  padding-bottom: 8px;
}

.continent-tab {
  padding: 8px 20px;
  border: 1px solid #ddd;
  border-radius: 20px;
  background: white;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
}

.continent-tab.active {
  background: #1890ff;
  color: white;
  border-color: #1890ff;
}

.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  margin-bottom: 20px;
  padding: 16px;
  background: #f5f5f5;
  border-radius: 8px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-group label {
  font-size: 14px;
  color: #666;
}

.filter-group select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
}

.visa-type-tags {
  display: flex;
  gap: 8px;
}

.visa-tag {
  padding: 6px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  font-size: 13px;
}

.visa-tag.active {
  background: #e6f7ff;
  border-color: #1890ff;
  color: #1890ff;
}

.sort-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.sort-btn {
  padding: 8px 16px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: #666;
}

.sort-btn.active {
  color: #1890ff;
  font-weight: 500;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 30px;
}

.product-card {
  border: 1px solid #eee;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.product-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-image {
  position: relative;
  height: 200px;
}

.card-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.visa-badge {
  position: absolute;
  top: 12px;
  left: 12px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: white;
}

.badge-free { background: #52c41a; }
.badge-arrival { background: #1890ff; }
.badge-evisa { background: #722ed1; }
.badge-required { background: #fa8c16; }

.card-content {
  padding: 16px;
}

.card-title {
  font-size: 16px;
  color: #1a1a1a;
  margin-bottom: 8px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-meta {
  display: flex;
  gap: 12px;
  font-size: 13px;
  color: #999;
  margin-bottom: 8px;
}

.card-tags {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.country-tag, .visa-type-tag {
  padding: 2px 8px;
  background: #f0f0f0;
  border-radius: 4px;
  font-size: 12px;
  color: #666;
}

.card-price {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.price-label {
  font-size: 12px;
  color: #999;
}

.price-value {
  font-size: 24px;
  color: #ff4d4f;
  font-weight: 600;
}

.price-unit {
  font-size: 12px;
  color: #999;
}

.loading-state {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.skeleton-card {
  height: 300px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: loading 1.5s infinite;
  border-radius: 8px;
}

@keyframes loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-state p {
  color: #999;
  margin-bottom: 16px;
}

.reset-btn {
  padding: 8px 24px;
  background: #1890ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  padding: 20px;
}

.pagination button {
  padding: 8px 16px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  cursor: pointer;
}

.pagination button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>

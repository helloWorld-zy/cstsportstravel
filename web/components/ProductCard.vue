<template>
  <div class="product-card" :class="{ 'sold-out': isSoldOut }" @click="goToDetail">
    <div class="card-image">
      <img
        :src="product.cover_image || '/static/images/default-product.png'"
        :alt="product.product_name"
        loading="lazy"
      />
      <div v-if="isSoldOut" class="sold-out-overlay">已售罄</div>
      <div v-if="product.tags && product.tags.length" class="card-tags">
        <span v-for="tag in product.tags.slice(0, 3)" :key="tag" class="tag">{{ tag }}</span>
      </div>
    </div>
    <div class="card-body">
      <h3 class="card-title">{{ product.product_name }}</h3>
      <div class="card-info">
        <span class="destinations">{{ product.destination_cities.join('·') }}</span>
        <span class="days">{{ product.days }}天{{ product.nights }}晚</span>
      </div>
      <div class="card-meta">
        <span class="origin">出发: {{ product.origin_city }}</span>
        <span v-if="product.satisfaction_rate" class="rating">
          {{ product.satisfaction_rate.toFixed(1) }}分
        </span>
      </div>
      <div class="card-footer">
        <div class="price">
          <span class="price-label">起</span>
          <span class="price-value">¥{{ product.min_price }}</span>
        </div>
        <span v-if="product.order_count > 0" class="sales">
          已售{{ product.order_count }}件
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProductSummary } from '~/composables/useProduct'

const props = defineProps<{
  product: ProductSummary
}>()

const router = useRouter()

const isSoldOut = computed(() => props.product.min_price === 0)

const goToDetail = () => {
  router.push(`/products/${props.product.id}`)
}
</script>

<style scoped>
.product-card {
  background: #fff;
  border-radius: 16px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgba(226, 232, 240, 0.8);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px -2px rgba(0, 0, 0, 0.02);
}

.product-card:hover {
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.08), 0 10px 10px -5px rgba(0, 0, 0, 0.02);
  transform: translateY(-4px);
  border-color: rgba(255, 87, 34, 0.2);
}

.product-card.sold-out {
  opacity: 0.7;
}

.card-image {
  position: relative;
  width: 100%;
  padding-top: 66.67%; /* 3:2 aspect ratio */
  overflow: hidden;
}

.card-image img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.product-card:hover .card-image img {
  transform: scale(1.08);
}

.sold-out-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 1px;
}

.card-tags {
  position: absolute;
  top: 12px;
  left: 12px;
  display: flex;
  gap: 6px;
  z-index: 2;
}

.tag {
  background: rgba(15, 23, 42, 0.75);
  backdrop-filter: blur(4px);
  color: #fff;
  padding: 3px 8px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 600;
}

.card-body {
  padding: 16px;
}

.card-title {
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 0 0 10px;
  height: 42px; /* Fixed height to keep grids aligned */
}

.card-info {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #64748b;
  margin-bottom: 6px;
  font-weight: 500;
}

.card-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 12px;
  font-weight: 500;
}

.rating {
  color: #f59e0b;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 2px;
}

.rating::before {
  content: '★';
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  padding-top: 10px;
  border-top: 1px solid #f1f5f9;
}

.price {
  display: flex;
  align-items: baseline;
  gap: 2px;
}

.price-label {
  font-size: 11px;
  color: #64748b;
  font-weight: 500;
}

.price-value {
  font-size: 18px;
  font-weight: 800;
  color: #2563eb;
}

.sales {
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}
</style>

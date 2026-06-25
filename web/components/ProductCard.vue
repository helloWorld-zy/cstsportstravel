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
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.2s, transform 0.2s;
}

.product-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
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
}

.sold-out-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
}

.card-tags {
  position: absolute;
  top: 8px;
  left: 8px;
  display: flex;
  gap: 4px;
}

.tag {
  background: rgba(255, 87, 34, 0.9);
  color: #fff;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
}

.card-body {
  padding: 12px;
}

.card-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 0 0 8px;
}

.card-info {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
}

.card-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #999;
  margin-bottom: 8px;
}

.rating {
  color: #ff9800;
  font-weight: 500;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.price {
  display: flex;
  align-items: baseline;
  gap: 2px;
}

.price-label {
  font-size: 12px;
  color: #999;
}

.price-value {
  font-size: 18px;
  font-weight: bold;
  color: #ff5722;
}

.sales {
  font-size: 12px;
  color: #999;
}
</style>

<template>
  <div class="departure-step">
    <h3>选择团期与人数</h3>

    <!-- Departure calendar -->
    <div class="section">
      <h4>选择出发日期</h4>
      <div v-if="loading" class="loading">加载中...</div>
      <div v-else class="departure-grid">
        <div
          v-for="dep in departures"
          :key="dep.id"
          class="departure-card"
          :class="{
            selected: selectedDeparture?.id === dep.id,
            disabled: dep.status !== 'open' || dep.available_stock <= 0,
          }"
          @click="selectDeparture(dep)"
        >
          <div class="date">{{ formatDate(dep.departure_date) }}</div>
          <div class="price">¥{{ formatAmount(dep.adult_price) }}/人</div>
          <div class="stock" :class="stockClass(dep)">
            {{ stockText(dep) }}
          </div>
        </div>
      </div>
    </div>

    <!-- Passenger count -->
    <div class="section">
      <h4>选择人数</h4>
      <div class="counter-row">
        <span class="label">成人</span>
        <el-input-number
          v-model="adultCount"
          :min="1"
          :max="20"
          @change="updatePrice"
        />
        <span class="hint">≥12周岁</span>
      </div>
      <div class="counter-row">
        <span class="label">儿童</span>
        <el-input-number
          v-model="childCount"
          :min="0"
          :max="10"
          @change="updatePrice"
        />
        <span class="hint">2-12周岁，不占床</span>
      </div>
      <div class="counter-row">
        <span class="label">婴儿</span>
        <el-input-number
          v-model="infantCount"
          :min="0"
          :max="adultCount"
          @change="updatePrice"
        />
        <span class="hint">&lt;2周岁，每成人最多1名</span>
      </div>
    </div>

    <!-- Price preview -->
    <div v-if="selectedDeparture" class="section price-preview">
      <h4>费用预览</h4>
      <div class="price-row">
        <span>成人 {{ adultCount }} × ¥{{ formatAmount(selectedDeparture.adult_price) }}</span>
        <span>¥{{ formatAmount(adultCount * selectedDeparture.adult_price) }}</span>
      </div>
      <div v-if="childCount > 0" class="price-row">
        <span>儿童 {{ childCount }} × ¥{{ formatAmount(selectedDeparture.child_price) }}</span>
        <span>¥{{ formatAmount(childCount * selectedDeparture.child_price) }}</span>
      </div>
      <div v-if="infantCount > 0" class="price-row">
        <span>婴儿 {{ infantCount }} × ¥{{ formatAmount(selectedDeparture.infant_price) }}</span>
        <span>¥{{ formatAmount(infantCount * selectedDeparture.infant_price) }}</span>
      </div>
      <div v-if="supplementAmount > 0" class="price-row supplement">
        <span>单房差（成人数为奇数自动附加）</span>
        <span>¥{{ formatAmount(supplementAmount) }}</span>
      </div>
      <div class="price-row total">
        <span>合计</span>
        <span class="total-price">¥{{ formatAmount(totalPrice) }}</span>
      </div>
    </div>

    <!-- Group size hint -->
    <div v-if="selectedDeparture" class="hint-text">
      当前余位：{{ selectedDeparture.available_stock ?? (selectedDeparture.total_stock - selectedDeparture.sold_count - selectedDeparture.locked_count) }} 位
    </div>

    <div class="actions">
      <el-button
        type="primary"
        :disabled="!selectedDeparture"
        @click="handleNext"
      >
        下一步
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { formatAmount } from '~/shared/utils/amount'

const props = defineProps<{
  productId: number
}>()

const emit = defineEmits<{
  update: [data: any]
  next: []
}>()

const api = useApi()
const loading = ref(true)
const departures = ref<any[]>([])
const selectedDeparture = ref<any>(null)
const adultCount = ref(1)
const childCount = ref(0)
const infantCount = ref(0)

onMounted(async () => {
  try {
    const now = new Date()
    const month = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
    departures.value = await api.get<any[]>(`/products/${props.productId}/departures`, {
      params: { month },
    })
  } catch (error) {
    console.error('Failed to load departures:', error)
  } finally {
    loading.value = false
  }
})

function selectDeparture(dep: any) {
  if (dep.status !== 'open') return
  const available = dep.available_stock ?? (dep.total_stock - dep.sold_count - dep.locked_count)
  if (available <= 0) return
  selectedDeparture.value = dep
  updatePrice()
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

function stockClass(dep: any) {
  const available = dep.available_stock ?? (dep.total_stock - dep.sold_count - dep.locked_count)
  if (available <= 0) return 'sold-out'
  if (available <= 5) return 'tight'
  return 'adequate'
}

function stockText(dep: any) {
  const available = dep.available_stock ?? (dep.total_stock - dep.sold_count - dep.locked_count)
  if (available <= 0) return '已售罄'
  if (available <= 5) return `仅剩${available}位`
  return '充足'
}

const supplementAmount = computed(() => {
  if (!selectedDeparture.value) return 0
  if (adultCount.value > 0 && adultCount.value % 2 !== 0) {
    return selectedDeparture.value.single_supplement
  }
  return 0
})

const totalPrice = computed(() => {
  if (!selectedDeparture.value) return 0
  const dep = selectedDeparture.value
  return (
    adultCount.value * dep.adult_price +
    childCount.value * dep.child_price +
    infantCount.value * dep.infant_price +
    supplementAmount.value
  )
})

function updatePrice() {
  // Reactive computed handles this
}

function handleNext() {
  if (!selectedDeparture.value) return

  emit('update', {
    departureId: selectedDeparture.value.id,
    adultCount: adultCount.value,
    childCount: childCount.value,
    infantCount: infantCount.value,
    adultPrice: selectedDeparture.value.adult_price,
    childPrice: selectedDeparture.value.child_price,
    infantPrice: selectedDeparture.value.infant_price,
    singleSupplement: selectedDeparture.value.single_supplement,
    departureDate: selectedDeparture.value.departure_date,
  })
  emit('next')
}
</script>

<style scoped>
.departure-step h3 {
  margin-bottom: 20px;
}

.section {
  margin-bottom: 24px;
}

.section h4 {
  margin-bottom: 12px;
  font-size: 16px;
}

.departure-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
}

.departure-card {
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  text-align: center;
  transition: all 0.2s;
}

.departure-card:hover:not(.disabled) {
  border-color: #409eff;
}

.departure-card.selected {
  border-color: #409eff;
  background: #ecf5ff;
}

.departure-card.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.date {
  font-weight: bold;
  margin-bottom: 4px;
}

.price {
  color: #ff4d4f;
  font-size: 14px;
}

.stock {
  font-size: 12px;
  margin-top: 4px;
}

.stock.adequate { color: #52c41a; }
.stock.tight { color: #faad14; }
.stock.sold-out { color: #999; }

.counter-row {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 12px;
}

.counter-row .label {
  width: 40px;
  font-weight: 500;
}

.counter-row .hint {
  color: #999;
  font-size: 12px;
}

.price-preview {
  background: #fafafa;
  padding: 16px;
  border-radius: 8px;
}

.price-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
}

.price-row.supplement {
  color: #faad14;
}

.price-row.total {
  border-top: 1px solid #e8e8e8;
  padding-top: 8px;
  margin-top: 8px;
  font-weight: bold;
}

.total-price {
  color: #ff4d4f;
  font-size: 18px;
}

.hint-text {
  color: #999;
  font-size: 13px;
  margin-bottom: 16px;
}

.actions {
  text-align: right;
  margin-top: 20px;
}
</style>

<template>
  <div class="addon-step">
    <h3>选择附加服务</h3>
    <p class="subtitle">以下服务为可选项，不选择也可以继续预订</p>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="addons.length === 0" class="empty">暂无可选附加服务</div>
    <div v-else class="addon-list">
      <div v-for="addon in addons" :key="addon.id" class="addon-item">
        <el-checkbox
          :model-value="isSelected(addon.id)"
          @change="(val: boolean) => toggleAddon(addon, val)"
        >
          <div class="addon-info">
            <div class="name">{{ addon.name }}</div>
            <div class="desc">{{ addon.description }}</div>
            <div class="price">
              <span v-if="addon.original_price" class="original-price">¥{{ formatAmount(addon.original_price) }}</span>
              <span class="current-price">¥{{ formatAmount(addon.price) }}</span>
            </div>
          </div>
        </el-checkbox>
      </div>
    </div>

    <div class="actions">
      <el-button @click="emit('back')">上一步</el-button>
      <el-button type="primary" @click="handleNext">下一步</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { formatAmount } from '~/shared/utils/amount'

interface Addon {
  id: number
  name: string
  description: string
  price: number
  original_price?: number
}

const props = defineProps<{
  productId?: number
  initialAddons?: Addon[]
}>()

const emit = defineEmits<{
  update: [data: any]
  next: []
  back: []
}>()

const api = useApi()
const loading = ref(false)
const addons = ref<Addon[]>([])
const selectedAddons = ref<Map<number, Addon & { quantity: number }>>(new Map())

onMounted(async () => {
  // Use initial addons from parent if provided
  if (props.initialAddons && props.initialAddons.length > 0) {
    addons.value = props.initialAddons
    return
  }

  // Otherwise, fetch from API
  if (props.productId) {
    loading.value = true
    try {
      const data = await api.get(`/products/${props.productId}/addons`)
      addons.value = data?.items || data || []
    } catch {
      // Addons endpoint may not exist yet — show empty state
      addons.value = []
    } finally {
      loading.value = false
    }
  }
})

function isSelected(addonId: number): boolean {
  return selectedAddons.value.has(addonId)
}

function toggleAddon(addon: any, selected: boolean) {
  if (selected) {
    selectedAddons.value.set(addon.id, { ...addon, quantity: 1 })
  } else {
    selectedAddons.value.delete(addon.id)
  }
}

function handleNext() {
  emit('update', {
    addons: Array.from(selectedAddons.value.values()),
  })
  emit('next')
}
</script>

<style scoped>
.addon-step h3 {
  margin-bottom: 8px;
}

.subtitle {
  color: #999;
  margin-bottom: 20px;
}

.addon-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.addon-item {
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 16px;
}

.addon-info {
  display: inline-block;
  vertical-align: top;
  margin-left: 8px;
}

.name {
  font-weight: 500;
  margin-bottom: 4px;
}

.desc {
  color: #666;
  font-size: 13px;
  margin-bottom: 4px;
}

.original-price {
  text-decoration: line-through;
  color: #999;
  margin-right: 8px;
  font-size: 13px;
}

.current-price {
  color: #ff4d4f;
  font-weight: bold;
}

.actions {
  display: flex;
  justify-content: space-between;
  margin-top: 24px;
}

.empty, .loading {
  text-align: center;
  color: #999;
  padding: 40px;
}
</style>

<template>
  <view class="booking-page">
    <!-- Step indicator -->
    <view class="step-bar">
      <view
        v-for="(step, i) in steps"
        :key="i"
        class="step-item"
        :class="{ active: currentStep === i, done: currentStep > i }"
      >
        <text class="step-num">{{ i + 1 }}</text>
        <text class="step-label">{{ step }}</text>
      </view>
    </view>

    <!-- Step 0: Departure selection -->
    <view v-if="currentStep === 0" class="step-content">
      <view class="section-title">选择出发日期</view>
      <view v-for="dep in departures" :key="dep.id" class="dep-card" :class="{ selected: selectedDep?.id === dep.id }" @tap="selectDeparture(dep)">
        <text class="dep-date">{{ formatDate(dep.departure_date) }}</text>
        <text class="dep-price">¥{{ (dep.adult_price / 100).toFixed(2) }}/人</text>
        <text class="dep-stock" :class="dep.status === 'open' ? 'stock-ok' : 'stock-out'">
          {{ dep.status === 'open' ? '可预订' : '已售罄' }}
        </text>
      </view>

      <view class="section-title">选择人数</view>
      <view class="counter-row">
        <text>成人</text>
        <view class="counter">
          <view class="btn" @tap="adultCount = Math.max(1, adultCount - 1)">-</view>
          <text>{{ adultCount }}</text>
          <view class="btn" @tap="adultCount++">+</view>
        </view>
      </view>
      <view class="counter-row">
        <text>儿童</text>
        <view class="counter">
          <view class="btn" @tap="childCount = Math.max(0, childCount - 1)">-</view>
          <text>{{ childCount }}</text>
          <view class="btn" @tap="childCount++">+</view>
        </view>
      </view>

      <view v-if="selectedDep" class="price-preview">
        <text>合计：¥{{ (totalPrice / 100).toFixed(2) }}</text>
        <text v-if="supplement > 0" class="supplement-hint">含单房差 ¥{{ (supplement / 100).toFixed(2) }}</text>
      </view>

      <button type="primary" :disabled="!selectedDep" @tap="currentStep = 1">下一步</button>
    </view>

    <!-- Step 1: Traveller info -->
    <view v-if="currentStep === 1" class="step-content">
      <view class="section-title">出游人信息</view>
      <view v-for="(t, i) in travellers" :key="i" class="traveller-form">
        <text class="form-label">{{ i < adultCount ? '成人' : '儿童' }} {{ i + 1 }}</text>
        <input v-model="t.real_name" placeholder="姓名" class="input" />
        <input v-model="t.id_card_no" placeholder="身份证号" class="input" maxlength="18" />
        <input v-model="t.phone" placeholder="手机号" class="input" maxlength="11" />
      </view>

      <view class="section-title">联系人</view>
      <input v-model="contactName" placeholder="联系人姓名" class="input" />
      <input v-model="contactPhone" placeholder="联系人手机号" class="input" maxlength="11" />

      <view class="btn-row">
        <button @tap="currentStep = 0">上一步</button>
        <button type="primary" @tap="currentStep = 2">下一步</button>
      </view>
    </view>

    <!-- Step 2: Confirm -->
    <view v-if="currentStep === 2" class="step-content">
      <view class="section-title">确认订单</view>
      <view class="confirm-card">
        <text>出发日期：{{ formatDate(selectedDep?.departure_date) }}</text>
        <text>人数：成人{{ adultCount }} 儿童{{ childCount }}</text>
        <text>应付：¥{{ (totalPrice / 100).toFixed(2) }}</text>
      </view>

      <view class="policy-check">
        <checkbox :checked="agreed" @change="agreed = $event.detail.value.length > 0" />
        <text>我已阅读并同意退改政策</text>
      </view>

      <view class="btn-row">
        <button @tap="currentStep = 1">上一步</button>
        <button type="primary" :disabled="!agreed" @tap="submitOrder">提交订单</button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { request } from '@/shared/api/request'

const steps = ['选团期', '填信息', '确认']
const currentStep = ref(0)
const productId = ref(0)
const departures = ref<any[]>([])
const selectedDep = ref<any>(null)
const adultCount = ref(1)
const childCount = ref(0)
const travellers = ref<any[]>([])
const contactName = ref('')
const contactPhone = ref('')
const agreed = ref(false)

const supplement = computed(() => {
  if (!selectedDep.value) return 0
  return adultCount.value > 0 && adultCount.value % 2 !== 0 ? selectedDep.value.single_supplement : 0
})

const totalPrice = computed(() => {
  if (!selectedDep.value) return 0
  const d = selectedDep.value
  return adultCount.value * d.adult_price + childCount.value * d.child_price + supplement.value
})

onMounted(() => {
  const pages = getCurrentPages()
  const page = pages[pages.length - 1]
  productId.value = Number(page?.options?.productId || page?.options?.productid || 0)
  loadDepartures()
  initTravellers()
})

async function loadDepartures() {
  try {
    const res = await request({ url: `/products/${productId.value}/departures`, method: 'GET' })
    departures.value = res.data || []
  } catch (e) {
    console.error('load departures failed', e)
  }
}

function initTravellers() {
  const forms = []
  for (let i = 0; i < adultCount.value + childCount.value; i++) {
    forms.push({ real_name: '', id_card_no: '', phone: '', is_child: i >= adultCount.value })
  }
  travellers.value = forms
}

function selectDeparture(dep: any) {
  if (dep.status !== 'open') return
  selectedDep.value = dep
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

async function submitOrder() {
  try {
    const res = await request({
      url: '/orders',
      method: 'POST',
      data: {
        product_id: productId.value,
        departure_id: selectedDep.value.id,
        adult_count: adultCount.value,
        child_count: childCount.value,
        infant_count: 0,
        travellers: travellers.value.map((t: any, i: number) => ({
          real_name: t.real_name,
          id_card_no: t.id_card_no,
          phone: t.phone,
          is_child: t.is_child,
        })),
        contact_name: contactName.value,
        contact_phone: contactPhone.value,
      },
    })

    const orderId = res.data.order_id
    uni.navigateTo({ url: `/pages/payment/index?orderId=${orderId}` })
  } catch (e: any) {
    uni.showToast({ title: e.message || '下单失败', icon: 'none' })
  }
}
</script>

<style scoped>
.booking-page { padding: 20rpx; }
.step-bar { display: flex; justify-content: space-around; margin-bottom: 30rpx; }
.step-item { text-align: center; }
.step-num { display: block; width: 40rpx; height: 40rpx; line-height: 40rpx; border-radius: 50%; background: #ddd; color: #fff; margin: 0 auto 8rpx; }
.step-item.active .step-num { background: #409eff; }
.step-item.done .step-num { background: #67c23a; }
.step-label { font-size: 24rpx; color: #999; }
.step-item.active .step-label { color: #333; }
.section-title { font-size: 30rpx; font-weight: bold; margin: 20rpx 0 10rpx; }
.dep-card { border: 1rpx solid #eee; border-radius: 12rpx; padding: 20rpx; margin-bottom: 16rpx; }
.dep-card.selected { border-color: #409eff; background: #ecf5ff; }
.dep-date { font-weight: bold; }
.dep-price { color: #ff4d4f; margin-left: 20rpx; }
.dep-stock { float: right; font-size: 24rpx; }
.stock-ok { color: #67c23a; }
.stock-out { color: #999; }
.counter-row { display: flex; justify-content: space-between; align-items: center; padding: 16rpx 0; }
.counter { display: flex; align-items: center; gap: 20rpx; }
.btn { width: 50rpx; height: 50rpx; line-height: 50rpx; text-align: center; border: 1rpx solid #ddd; border-radius: 8rpx; }
.price-preview { background: #fafafa; padding: 20rpx; border-radius: 12rpx; margin: 20rpx 0; text-align: center; }
.supplement-hint { display: block; font-size: 24rpx; color: #faad14; }
.traveller-form { border: 1rpx solid #eee; border-radius: 12rpx; padding: 20rpx; margin-bottom: 16rpx; }
.form-label { font-weight: bold; margin-bottom: 10rpx; display: block; }
.input { border: 1rpx solid #ddd; border-radius: 8rpx; padding: 16rpx; margin-bottom: 12rpx; width: 100%; box-sizing: border-box; }
.confirm-card { background: #fafafa; padding: 20rpx; border-radius: 12rpx; margin-bottom: 20rpx; }
.confirm-card text { display: block; margin-bottom: 8rpx; }
.policy-check { display: flex; align-items: center; gap: 10rpx; margin-bottom: 20rpx; }
.btn-row { display: flex; gap: 20rpx; margin-top: 20rpx; }
.btn-row button { flex: 1; }
</style>

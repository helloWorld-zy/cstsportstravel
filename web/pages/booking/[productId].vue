<template>
  <div class="booking-page">
    <!-- Step progress bar -->
    <div class="step-progress">
      <el-steps :active="currentStep" finish-status="success" align-center>
        <el-step title="选择团期" />
        <el-step title="填写出游人" />
        <el-step title="附加服务" />
        <el-step title="确认支付" />
      </el-steps>
    </div>

    <!-- Step content -->
    <div class="step-content">
      <DepartureStep
        v-if="currentStep === 0"
        :product-id="productId"
        @update="onDepartureUpdate"
        @next="nextStep"
      />
      <TravellerStep
        v-if="currentStep === 1"
        :adult-count="bookingData.adultCount"
        :child-count="bookingData.childCount"
        :infant-count="bookingData.infantCount"
        @update="onTravellerUpdate"
        @next="nextStep"
        @back="prevStep"
      />
      <AddonStep
        v-if="currentStep === 2"
        :product-id="productId"
        @update="onAddonUpdate"
        @next="nextStep"
        @back="prevStep"
      />
      <ConfirmStep
        v-if="currentStep === 3"
        :booking-data="bookingData"
        @submit="onSubmitOrder"
        @back="prevStep"
      />
    </div>

    <!-- Price summary footer -->
    <div class="price-footer">
      <div class="price-summary">
        <span class="label">应付总额：</span>
        <span class="price">{{ formatPrice(totalAmount) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { formatPrice } from '~/shared/utils/amount'

definePageMeta({
  layout: 'default',
  middleware: ['auth'],
})

const route = useRoute()
const router = useRouter()
const api = useApi()

const productId = Number(route.params.productId)
const currentStep = ref(0)

interface BookingData {
  departureId: number | null
  adultCount: number
  childCount: number
  infantCount: number
  adultPrice: number
  childPrice: number
  infantPrice: number
  singleSupplement: number
  travellers: any[]
  addons: any[]
  contactName: string
  contactPhone: string
  productId: number
  productName: string
  departureDate: string
}

const bookingData = ref<BookingData>({
  departureId: null,
  adultCount: 1,
  childCount: 0,
  infantCount: 0,
  adultPrice: 0,
  childPrice: 0,
  infantPrice: 0,
  singleSupplement: 0,
  travellers: [],
  addons: [],
  contactName: '',
  contactPhone: '',
  productId,
  productName: '',
  departureDate: '',
})

const totalAmount = computed(() => {
  const d = bookingData.value
  const adultTotal = d.adultCount * d.adultPrice
  const childTotal = d.childCount * d.childPrice
  const infantTotal = d.infantCount * d.infantPrice
  // Single room supplement: auto-add when adult count is odd
  const supplement = d.adultCount > 0 && d.adultCount % 2 !== 0 ? d.singleSupplement : 0
  const addonTotal = d.addons.reduce((sum: number, a: any) => sum + (a.price * (a.quantity || 1)), 0)
  return adultTotal + childTotal + infantTotal + supplement + addonTotal
})

function onDepartureUpdate(data: any) {
  Object.assign(bookingData.value, data)
}

function onTravellerUpdate(data: any) {
  bookingData.value.travellers = data.travellers
  bookingData.value.contactName = data.contactName
  bookingData.value.contactPhone = data.contactPhone
}

function onAddonUpdate(data: any) {
  bookingData.value.addons = data.addons
}

function nextStep() {
  if (currentStep.value < 3) {
    currentStep.value++
  }
}

function prevStep() {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

async function onSubmitOrder() {
  try {
    const result = await api.post<any>('/orders', {
      product_id: bookingData.value.productId,
      departure_id: bookingData.value.departureId,
      adult_count: bookingData.value.adultCount,
      child_count: bookingData.value.childCount,
      infant_count: bookingData.value.infantCount,
      travellers: bookingData.value.travellers.map((t: any) => ({
        real_name: t.real_name,
        id_card_no: t.id_card_no,
        phone: t.phone,
        birth_date: t.birth_date,
        gender: t.gender,
        is_child: t.is_child || false,
        is_infant: t.is_infant || false,
        linked_adult_traveller_index: t.linked_adult_traveller_index,
      })),
      addons: bookingData.value.addons.map((a: any) => ({
        addon_id: a.id,
        quantity: a.quantity || 1,
      })),
      contact_name: bookingData.value.contactName,
      contact_phone: bookingData.value.contactPhone,
    })

    // Navigate to payment page
    router.push(`/payment/${result.order_id}`)
  } catch (error: any) {
    ElMessage.error(error.message || '下单失败，请重试')
  }
}
</script>

<style scoped>
.booking-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 20px;
  padding-bottom: 80px;
}

.step-progress {
  margin-bottom: 30px;
}

.step-content {
  min-height: 400px;
}

.price-footer {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: #fff;
  border-top: 1px solid #eee;
  padding: 12px 20px;
  display: flex;
  justify-content: center;
  z-index: 100;
}

.price-summary {
  display: flex;
  align-items: center;
  gap: 8px;
}

.price-summary .label {
  font-size: 14px;
  color: #666;
}

.price-summary .price {
  font-size: 24px;
  font-weight: bold;
  color: #ff4d4f;
}
</style>

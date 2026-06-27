<template>
  <div class="product-form">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑产品' : '发布产品' }}</span>
          <el-button @click="goBack">返回列表</el-button>
        </div>
      </template>

      <!-- Step progress -->
      <el-steps :active="currentStep" finish-status="success" align-center style="margin-bottom: 32px">
        <el-step title="基础信息" />
        <el-step title="行程编辑" />
        <el-step title="价格配置" />
        <el-step title="退改规则" />
        <el-step title="库存设置" />
      </el-steps>

      <!-- Step 1: Basic Info -->
      <div v-show="currentStep === 0">
        <el-form ref="basicFormRef" :model="form" :rules="basicRules" label-width="120px" style="max-width: 800px">
          <el-form-item label="产品名称" prop="product_name">
            <el-input v-model="form.product_name" maxlength="200" show-word-limit />
          </el-form-item>
          <el-form-item label="产品分类" prop="category_id">
            <el-select v-model="form.category_id" placeholder="选择分类">
              <el-option
                v-for="cat in categories"
                :key="cat.id"
                :label="cat.name"
                :value="cat.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="出发城市" prop="origin_city">
            <el-input v-model="form.origin_city" />
          </el-form-item>
          <el-form-item label="目的地城市" prop="destination_cities">
            <el-select
              v-model="form.destination_cities"
              multiple
              filterable
              allow-create
              default-first-option
              placeholder="输入目的地城市后回车"
            />
          </el-form-item>
          <el-row :gutter="16">
            <el-col :span="8">
              <el-form-item label="行程天数" prop="days">
                <el-input-number v-model="form.days" :min="1" :max="30" />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="住宿晚数" prop="nights">
                <el-input-number v-model="form.nights" :min="0" :max="29" />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="交通方式">
                <el-select v-model="form.transport_mode" placeholder="选择">
                  <el-option label="飞机" value="flight" />
                  <el-option label="火车" value="train" />
                  <el-option label="大巴" value="bus" />
                  <el-option label="自驾" value="self_drive" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="8">
              <el-form-item label="产品等级">
                <el-select v-model="form.product_grade" placeholder="选择">
                  <el-option label="经济" value="standard" />
                  <el-option label="舒适" value="comfort" />
                  <el-option label="豪华" value="luxury" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="最小成团">
                <el-input-number v-model="form.min_group_size" :min="1" />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="最大团量">
                <el-input-number v-model="form.max_group_size" :min="1" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="封面图片">
            <el-input v-model="form.cover_image" placeholder="图片URL" />
          </el-form-item>
          <el-form-item label="产品简介" prop="summary">
            <el-input v-model="form.summary" type="textarea" :rows="2" maxlength="500" show-word-limit />
          </el-form-item>
          <el-form-item label="产品描述">
            <el-input v-model="form.description" type="textarea" :rows="4" />
          </el-form-item>
          <el-form-item label="费用包含">
            <el-input v-model="form.fee_included" type="textarea" :rows="3" placeholder="往返机票、酒店住宿、景点门票、导游服务" />
          </el-form-item>
          <el-form-item label="费用不含">
            <el-input v-model="form.fee_excluded" type="textarea" :rows="3" placeholder="个人消费、自费项目" />
          </el-form-item>
          <el-form-item label="预订须知">
            <el-input v-model="form.booking_notes" type="textarea" :rows="3" />
          </el-form-item>
        </el-form>
      </div>

      <!-- Step 2: Itinerary -->
      <div v-show="currentStep === 1">
        <ItineraryEditor
          v-model="itineraries"
          :total-days="form.days"
        />
      </div>

      <!-- Step 3: Price Calendar -->
      <div v-show="currentStep === 2">
        <PriceCalendar
          v-if="productID > 0"
          :product-id="productID"
        />
        <el-alert
          v-else
          type="warning"
          :closable="false"
          description="请先保存基础信息后再配置价格日历"
        />
      </div>

      <!-- Step 4: Cancellation Rules -->
      <div v-show="currentStep === 3">
        <div class="refund-rules">
          <div class="rules-header">
            <span>退改阶梯费率</span>
            <el-button size="small" @click="addRefundRule">+ 添加规则</el-button>
          </div>
          <el-table :data="refundRules" style="width: 100%">
            <el-table-column label="规则名称" min-width="150">
              <template #default="{ row }">
                <el-input v-model="row.rule_name" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="距出发天数（最小）" width="150">
              <template #default="{ row }">
                <el-input-number v-model="row.days_before_min" :min="0" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="距出发天数（最大）" width="150">
              <template #default="{ row }">
                <el-input-number v-model="row.days_before_max" :min="0" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="退款比例(%)" width="130">
              <template #default="{ row }">
                <el-input-number v-model="row.refund_percentage" :min="0" :max="100" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="说明" min-width="200">
              <template #default="{ row }">
                <el-input v-model="row.description" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="{ $index }">
                <el-button size="small" type="danger" :icon="Delete" circle @click="refundRules.splice($index, 1)" />
              </template>
            </el-table-column>
          </el-table>

          <div style="margin-top: 16px">
            <el-button size="small" @click="loadTemplateRules">从模板加载</el-button>
          </div>
        </div>
      </div>

      <!-- Step 5: Departure/Stock -->
      <div v-show="currentStep === 4">
        <PriceCalendar
          v-if="productID > 0"
          :product-id="productID"
        />
        <el-alert
          v-else
          type="warning"
          :closable="false"
          description="请先保存基础信息后再设置团期库存"
        />
      </div>

      <!-- Navigation buttons -->
      <div class="step-actions">
        <el-button v-if="currentStep > 0" @click="currentStep--">上一步</el-button>
        <el-button
          v-if="currentStep < 4"
          type="primary"
          @click="handleNext"
        >
          下一步
        </el-button>
        <el-button
          v-if="currentStep === 4 || (isEdit && currentStep >= 0)"
          type="success"
          :loading="saving"
          @click="saveDraft"
        >
          保存草稿
        </el-button>
        <el-button
          v-if="productID > 0"
          type="warning"
          :loading="submitting"
          @click="submitForReview"
        >
          提交审核
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import { adminApi } from '@/api/request'
import ItineraryEditor from '@/components/ItineraryEditor.vue'
import PriceCalendar from '@/components/PriceCalendar.vue'

interface Category {
  id: number
  name: string
}

interface ItineraryDay {
  day_no: number
  title: string
  description: string
  meals: { breakfast: boolean; lunch: boolean; dinner: boolean }
  hotel: string
  transport: string
  spots: { name: string; description: string; duration: string; image: string }[]
  images: string[]
}

interface RefundRule {
  rule_name: string
  days_before_min: number
  days_before_max: number | null
  refund_percentage: number
  description: string
}

const router = useRouter()
const route = useRoute()
const isEdit = computed(() => !!route.params.id)
const productID = ref(Number(route.params.id) || 0)
const currentStep = ref(0)
const saving = ref(false)
const submitting = ref(false)

const basicFormRef = ref<FormInstance>()
const categories = ref<Category[]>([])

const form = reactive({
  product_name: '',
  category_id: null as number | null,
  origin_city: '',
  destination_cities: [] as string[],
  days: 3,
  nights: 2,
  transport_mode: '',
  product_grade: '',
  min_group_size: 2,
  max_group_size: 50,
  cover_image: '',
  summary: '',
  description: '',
  fee_included: '',
  fee_excluded: '',
  booking_notes: '',
})

const basicRules: FormRules = {
  product_name: [{ required: true, message: '请输入产品名称', trigger: 'blur' }],
  category_id: [{ required: true, message: '请选择分类', trigger: 'change' }],
  origin_city: [{ required: true, message: '请输入出发城市', trigger: 'blur' }],
  destination_cities: [{ required: true, type: 'array', min: 1, message: '请添加目的地', trigger: 'change' }],
  days: [{ required: true, type: 'number', min: 1, message: '天数至少为1', trigger: 'change' }],
}

const itineraries = ref<ItineraryDay[]>([])
const refundRules = ref<RefundRule[]>([])

async function loadCategories() {
  try {
    // Use a simple category list - in production this comes from the category API
    categories.value = [
      { id: 1, name: '境内跟团游' },
      { id: 2, name: '境内自由行' },
      { id: 3, name: '出境跟团游' },
    ]
  } catch {
    // ignore
  }
}

async function loadProduct() {
  if (!productID.value) return
  try {
    const data = await adminApi.get<any>(`/admin/products/${productID.value}`)
    form.product_name = data.product_name
    form.category_id = data.category_id
    form.origin_city = data.origin_city
    form.destination_cities = data.destination_cities || []
    form.days = data.days
    form.nights = data.nights
    form.transport_mode = data.transport_mode
    form.product_grade = data.product_grade
    form.min_group_size = data.min_group_size
    form.max_group_size = data.max_group_size
    form.cover_image = data.cover_image
    form.summary = data.summary
    form.description = data.description
    form.fee_included = data.fee_included
    form.fee_excluded = data.fee_excluded
    form.booking_notes = data.booking_notes
  } catch (e: any) {
    ElMessage.error(e.message || '加载产品失败')
  }
}

async function loadItinerary() {
  if (!productID.value) return
  try {
    const data = await adminApi.get<any>(`/admin/products/${productID.value}/itinerary`)
    if (data?.itineraries) {
      itineraries.value = data.itineraries
    }
  } catch {
    // ignore - product may not have itinerary yet
  }
}

async function handleNext() {
  if (currentStep.value === 0) {
    // Validate basic form before proceeding
    if (!basicFormRef.value) return
    try {
      await basicFormRef.value.validate()
    } catch {
      ElMessage.warning('请填写所有必填项')
      return
    }
    // Auto-save draft when moving from step 0
    if (!productID.value) {
      await saveDraft()
    }
  }
  currentStep.value++
}

async function saveDraft() {
  saving.value = true
  try {
    const payload = {
      product_name: form.product_name,
      category_id: form.category_id,
      origin_city: form.origin_city,
      destination_cities: form.destination_cities,
      days: form.days,
      nights: form.nights,
      transport_mode: form.transport_mode,
      product_grade: form.product_grade,
      min_group_size: form.min_group_size,
      max_group_size: form.max_group_size,
      cover_image: form.cover_image,
      summary: form.summary,
      description: form.description,
      fee_included: form.fee_included,
      fee_excluded: form.fee_excluded,
      booking_notes: form.booking_notes,
    }

    if (productID.value) {
      await adminApi.put(`/admin/products/${productID.value}`, payload)
      ElMessage.success('保存成功')
    } else {
      const data = await adminApi.post<any>('/admin/products', payload)
      productID.value = data.id
      ElMessage.success('创建成功')
      // Update URL to edit mode
      router.replace(`/products/edit/${data.id}`)
    }

    // Save itinerary if we have data
    if (itineraries.value.length > 0 && productID.value) {
      await adminApi.post(`/admin/products/${productID.value}/itinerary`, {
        itineraries: itineraries.value,
      })
    }
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function submitForReview() {
  if (!productID.value) return
  submitting.value = true
  try {
    // Save first
    await saveDraft()
    // Then submit
    await adminApi.post(`/admin/products/${productID.value}/submit-review`)
    ElMessage.success('已提交审核')
    router.push('/products')
  } catch (e: any) {
    ElMessage.error(e.message || '提交失败')
  } finally {
    submitting.value = false
  }
}

function addRefundRule() {
  refundRules.value.push({
    rule_name: '',
    days_before_min: 0,
    days_before_max: null,
    refund_percentage: 100,
    description: '',
  })
}

async function loadTemplateRules() {
  try {
    const data = await adminApi.get<any[]>('/admin/cancellation-rules')
    if (data && data.length > 0) {
      refundRules.value = data.map((r: any) => ({
        rule_name: r.rule_name,
        days_before_min: r.days_before_min,
        days_before_max: r.days_before_max,
        refund_percentage: r.refund_percentage,
        description: r.description || '',
      }))
      ElMessage.success('已加载模板')
    } else {
      ElMessage.info('暂无可用模板')
    }
  } catch {
    ElMessage.info('暂无可用模板')
  }
}

function goBack() {
  router.push('/products')
}

onMounted(() => {
  loadCategories()
  if (isEdit.value) {
    loadProduct()
    loadItinerary()
  }
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.step-actions {
  margin-top: 32px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
  display: flex;
  gap: 12px;
  justify-content: center;
}
.refund-rules {
  max-width: 1000px;
}
.rules-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
</style>

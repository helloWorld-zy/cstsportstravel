<template>
  <el-dialog
    v-model="visible"
    title="评价订单"
    width="560px"
    :close-on-click-modal="false"
    @close="$emit('close')"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
    >
      <!-- Overall Rating -->
      <el-form-item label="总体评分" prop="rating">
        <div class="rating-section">
          <el-rate
            v-model="form.rating"
            :colors="['#99A9BF', '#F7BA2A', '#FF9900']"
            show-text
            :texts="['很差', '较差', '一般', '满意', '非常满意']"
            :score-template="ratingText"
            size="large"
          />
        </div>
      </el-form-item>

      <!-- Dimension Ratings -->
      <el-form-item label="分项评分">
        <div class="dimension-ratings">
          <div v-for="dim in dimensions" :key="dim.key" class="dimension-item">
            <span class="dim-label">{{ dim.label }}</span>
            <el-rate
              v-model="form.dimensions[dim.key]"
              :colors="['#99A9BF', '#F7BA2A', '#FF9900']"
              allow-half
            />
          </div>
        </div>
      </el-form-item>

      <!-- Review Content -->
      <el-form-item label="评价内容" prop="content">
        <el-input
          v-model="form.content"
          type="textarea"
          :rows="4"
          placeholder="请分享您的旅行体验，帮助其他用户做出选择（至少10个字）"
          maxlength="1000"
          show-word-limit
        />
      </el-form-item>

      <!-- Image Upload -->
      <el-form-item label="上传图片（最多5张）">
        <el-upload
          v-model:file-list="fileList"
          action="/api/v1/upload"
          list-type="picture-card"
          :limit="5"
          :on-success="onUploadSuccess"
          :on-remove="onRemove"
          :before-upload="beforeUpload"
          accept="image/jpeg,image/png"
        >
          <el-icon><Plus /></el-icon>
        </el-upload>
      </el-form-item>

      <!-- Anonymous Option -->
      <el-form-item>
        <el-checkbox v-model="form.is_anonymous">匿名评价</el-checkbox>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="$emit('close')">取消</el-button>
      <el-button
        type="primary"
        :loading="submitting"
        :disabled="form.rating === 0"
        @click="submitReview"
      >
        提交评价
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import type { FormInstance, FormRules, UploadFile } from 'element-plus'
import { ElMessage } from 'element-plus'

interface Props {
  productId: number
  orderId: number
}

const props = defineProps<Props>()
const emit = defineEmits(['close', 'success'])

const visible = ref(true)
const formRef = ref<FormInstance>()
const submitting = ref(false)
const fileList = ref<UploadFile[]>([])

const dimensions = [
  { key: 'guide', label: '导游服务' },
  { key: 'itinerary', label: '行程安排' },
  { key: 'hotel', label: '住宿条件' },
  { key: 'transport', label: '交通出行' },
  { key: 'food', label: '餐饮质量' },
]

const form = reactive({
  rating: 0,
  content: '',
  is_anonymous: false,
  dimensions: {
    guide: 0,
    itinerary: 0,
    hotel: 0,
    transport: 0,
    food: 0,
  } as Record<string, number>,
  images: [] as string[],
})

const ratingText = computed(() => {
  const texts = ['', '很差', '较差', '一般', '满意', '非常满意']
  return texts[form.rating] || ''
})

const rules: FormRules = {
  rating: [
    { required: true, message: '请选择评分', trigger: 'change' },
    {
      validator: (_rule: any, value: number, callback: Function) => {
        if (value < 1 || value > 5) {
          callback(new Error('评分必须在1-5之间'))
        } else {
          callback()
        }
      },
      trigger: 'change',
    },
  ],
  content: [
    { required: true, message: '请输入评价内容', trigger: 'blur' },
    { min: 10, message: '评价内容至少10个字', trigger: 'blur' },
  ],
}

function beforeUpload(file: File) {
  const isImage = ['image/jpeg', 'image/png'].includes(file.type)
  const isLt5M = file.size / 1024 / 1024 < 5

  if (!isImage) {
    ElMessage.error('只能上传 JPG/PNG 格式图片')
    return false
  }
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过 5MB')
    return false
  }
  return true
}

function onUploadSuccess(response: any, file: UploadFile) {
  if (response?.data?.url) {
    form.images.push(response.data.url)
  }
}

function onRemove(file: UploadFile) {
  const url = file.response?.data?.url || file.url
  if (url) {
    form.images = form.images.filter(img => img !== url)
  }
}

async function submitReview() {
  if (!formRef.value) return

  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    await $fetch(`/api/v1/products/${props.productId}/reviews?order_id=${props.orderId}`, {
      method: 'POST',
      body: {
        rating: form.rating,
        content: form.content,
        images: form.images,
        is_anonymous: form.is_anonymous,
      },
    })

    ElMessage.success('评价提交成功')
    emit('success')
  } catch (error: any) {
    const msg = error?.data?.message || '评价提交失败'
    ElMessage.error(msg)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.rating-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dimension-ratings {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 24px;
}

.dimension-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dim-label {
  font-size: 14px;
  color: #606266;
  min-width: 60px;
}
</style>

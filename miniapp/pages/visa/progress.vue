<template>
  <view class="visa-progress-page">
    <!-- Loading -->
    <view v-if="isLoading" class="loading">
      <text>加载中...</text>
    </view>

    <!-- Content -->
    <view v-else-if="visaOrder" class="content">
      <!-- Status Card -->
      <view class="status-card">
        <view class="status-badge" :class="getStatusClass()">
          <text>{{ visaOrder.status_name }}</text>
        </view>
        <view v-if="visaOrder.estimated_completion_date" class="estimated">
          <text>预计完成：{{ visaOrder.estimated_completion_date }}</text>
        </view>
      </view>

      <!-- Progress Steps -->
      <view class="progress-steps">
        <view
          v-for="(step, index) in steps"
          :key="step.status"
          class="step-item"
          :class="getStepClass(step, index)"
        >
          <view class="step-dot">
            <text v-if="isCompleted(step)">✓</text>
            <text v-else>{{ index + 1 }}</text>
          </view>
          <view v-if="index < steps.length - 1" class="step-line" :class="{ completed: isLineCompleted(index) }" />
          <text class="step-label">{{ step.label }}</text>
        </view>
      </view>

      <!-- Materials Section -->
      <view class="section">
        <text class="section-title">签证材料</text>
        <view class="material-list">
          <view
            v-for="material in materials"
            :key="material.id"
            class="material-item"
          >
            <view class="material-info">
              <text class="material-name">{{ material.material_name }}</text>
              <text v-if="material.is_required" class="required">必填</text>
            </view>
            <view class="material-status" :class="getMaterialStatusClass(material.status)">
              <text>{{ getMaterialStatusName(material.status) }}</text>
            </view>
            <view v-if="material.status === 'pending' || material.status === 'rejected'" class="material-action">
              <button size="mini" @tap="uploadMaterial(material)">上传</button>
            </view>
          </view>
        </view>
      </view>

      <!-- Submit Button -->
      <view v-if="canSubmit" class="submit-section">
        <button class="submit-btn" @tap="submitMaterials">提交全部材料</button>
      </view>

      <!-- Timeline -->
      <view v-if="progressList.length > 0" class="section">
        <text class="section-title">办理记录</text>
        <view class="timeline">
          <view v-for="item in progressList" :key="item.id" class="timeline-item">
            <view class="timeline-dot" />
            <view class="timeline-content">
              <text class="timeline-status">{{ item.status_name }}</text>
              <text class="timeline-time">{{ item.created_at }}</text>
              <text v-if="item.comment" class="timeline-comment">{{ item.comment }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- Tracking Info -->
      <view v-if="visaOrder.tracking_company" class="section">
        <text class="section-title">物流信息</text>
        <view class="tracking-card">
          <view class="tracking-row">
            <text class="label">快递公司</text>
            <text class="value">{{ visaOrder.tracking_company }}</text>
          </view>
          <view class="tracking-row">
            <text class="label">运单号</text>
            <text class="value">{{ visaOrder.tracking_number }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const isLoading = ref(true)
const visaOrder = ref<any>(null)
const materials = ref<any[]>([])
const progressList = ref<any[]>([])

const steps = [
  { status: 'pending_submit', label: '待提交' },
  { status: 'reviewing', label: '审核中' },
  { status: 'submitted', label: '已送签' },
  { status: 'approved', label: '已出签' },
]

const statusOrder = ['pending_submit', 'reviewing', 'submitted', 'approved', 'rejected']

const getStatusIndex = (status: string) => statusOrder.indexOf(status)
const currentIndex = computed(() => visaOrder.value ? getStatusIndex(visaOrder.value.status) : -1)

const isCompleted = (step: any) => {
  const idx = getStatusIndex(step.status)
  return idx < currentIndex.value && visaOrder.value?.status !== 'rejected'
}

const isLineCompleted = (index: number) => {
  return index < currentIndex.value && visaOrder.value?.status !== 'rejected'
}

const getStepClass = (step: any, index: number) => {
  if (visaOrder.value?.status === 'rejected' && step.status === 'approved') return 'rejected'
  if (isCompleted(step)) return 'completed'
  if (step.status === visaOrder.value?.status) return 'current'
  return 'pending'
}

const getStatusClass = () => {
  if (!visaOrder.value) return ''
  switch (visaOrder.value.status) {
    case 'pending_submit': return 'status-pending'
    case 'reviewing': return 'status-reviewing'
    case 'submitted': return 'status-submitted'
    case 'approved': return 'status-approved'
    case 'rejected': return 'status-rejected'
    default: return ''
  }
}

const getMaterialStatusClass = (status: string) => {
  switch (status) {
    case 'pending': return 'mat-pending'
    case 'submitted': return 'mat-submitted'
    case 'approved': return 'mat-approved'
    case 'rejected': return 'mat-rejected'
    default: return ''
  }
}

const getMaterialStatusName = (status: string) => {
  switch (status) {
    case 'pending': return '待上传'
    case 'submitted': return '已上传'
    case 'approved': return '审核通过'
    case 'rejected': return '需修改'
    case 'supplement': return '需补充'
    default: return status
  }
}

const canSubmit = computed(() => {
  return visaOrder.value?.status === 'pending_submit' &&
    materials.value.some(m => m.status === 'pending')
})

const loadData = async () => {
  try {
    const pages = getCurrentPages()
    const page = pages[pages.length - 1]
    const visaOrderId = page.options?.id

    if (!visaOrderId) return

    // Load visa order
    const orderRes = await uni.request({
      url: `/api/v2/visa-orders/${visaOrderId}`,
      method: 'GET',
    })
    visaOrder.value = orderRes.data

    // Load materials
    const matRes = await uni.request({
      url: `/api/v2/visa-orders/${visaOrderId}/materials`,
      method: 'GET',
    })
    materials.value = matRes.data?.materials || []

    // Load progress
    const progRes = await uni.request({
      url: `/api/v2/visa-orders/${visaOrderId}/progress`,
      method: 'GET',
    })
    visaOrder.value = { ...visaOrder.value, ...progRes.data }
    progressList.value = progRes.data?.timeline || []
  } catch (error) {
    console.error('Failed to load data:', error)
    uni.showToast({ title: '加载失败', icon: 'error' })
  } finally {
    isLoading.value = false
  }
}

const uploadMaterial = (material: any) => {
  uni.chooseImage({
    count: 1,
    success: async (res) => {
      const filePath = res.tempFilePaths[0]
      try {
        await uni.uploadFile({
          url: `/api/v2/visa-orders/${visaOrder.value.id}/materials/${material.id}/upload`,
          filePath,
          name: 'file',
        })
        uni.showToast({ title: '上传成功', icon: 'success' })
        loadData()
      } catch (error) {
        uni.showToast({ title: '上传失败', icon: 'error' })
      }
    },
  })
}

const submitMaterials = async () => {
  try {
    await uni.request({
      url: `/api/v2/visa-orders/${visaOrder.value.id}/materials/submit`,
      method: 'POST',
    })
    uni.showToast({ title: '提交成功', icon: 'success' })
    loadData()
  } catch (error) {
    uni.showToast({ title: '提交失败', icon: 'error' })
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.visa-progress-page {
  padding: 20rpx;
  background: #f5f5f5;
  min-height: 100vh;
}

.loading {
  text-align: center;
  padding: 100rpx 0;
  color: #999;
}

.status-card {
  background: white;
  border-radius: 16rpx;
  padding: 32rpx;
  text-align: center;
  margin-bottom: 20rpx;
}

.status-badge {
  display: inline-block;
  padding: 12rpx 32rpx;
  border-radius: 32rpx;
  font-size: 32rpx;
  font-weight: 500;
}

.status-pending { background: #f0f0f0; color: #666; }
.status-reviewing { background: #e6f7ff; color: #1890ff; }
.status-submitted { background: #fff7e6; color: #fa8c16; }
.status-approved { background: #f6ffed; color: #52c41a; }
.status-rejected { background: #fff2f0; color: #ff4d4f; }

.estimated {
  margin-top: 16rpx;
  font-size: 26rpx;
  color: #666;
}

.progress-steps {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  background: white;
  border-radius: 16rpx;
  padding: 32rpx;
  margin-bottom: 20rpx;
  position: relative;
}

.step-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  flex: 1;
}

.step-dot {
  width: 48rpx;
  height: 48rpx;
  border-radius: 50%;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  color: #999;
  z-index: 1;
}

.step-item.completed .step-dot {
  background: #52c41a;
  color: white;
}

.step-item.current .step-dot {
  background: #1890ff;
  color: white;
}

.step-line {
  position: absolute;
  top: 24rpx;
  left: 50%;
  width: 100%;
  height: 4rpx;
  background: #f0f0f0;
}

.step-line.completed {
  background: #52c41a;
}

.step-label {
  margin-top: 12rpx;
  font-size: 24rpx;
  color: #666;
}

.section {
  background: white;
  border-radius: 16rpx;
  padding: 32rpx;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 500;
  color: #1a1a1a;
  margin-bottom: 24rpx;
  display: block;
}

.material-item {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.material-item:last-child {
  border-bottom: none;
}

.material-info {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.material-name {
  font-size: 28rpx;
  color: #1a1a1a;
}

.required {
  font-size: 22rpx;
  color: #ff4d4f;
}

.material-status {
  padding: 8rpx 16rpx;
  border-radius: 8rpx;
  font-size: 24rpx;
  margin-right: 16rpx;
}

.mat-pending { background: #f0f0f0; color: #666; }
.mat-submitted { background: #e6f7ff; color: #1890ff; }
.mat-approved { background: #f6ffed; color: #52c41a; }
.mat-rejected { background: #fff2f0; color: #ff4d4f; }

.material-action button {
  font-size: 24rpx;
  padding: 8rpx 16rpx;
}

.submit-section {
  margin-bottom: 20rpx;
}

.submit-btn {
  width: 100%;
  background: #1890ff;
  color: white;
  border-radius: 16rpx;
  font-size: 32rpx;
  padding: 24rpx;
}

.timeline-item {
  display: flex;
  gap: 20rpx;
  padding: 20rpx 0;
  border-left: 4rpx solid #f0f0f0;
  margin-left: 12rpx;
  padding-left: 28rpx;
  position: relative;
}

.timeline-dot {
  position: absolute;
  left: -12rpx;
  top: 28rpx;
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  background: #1890ff;
}

.timeline-status {
  font-size: 28rpx;
  font-weight: 500;
  color: #1a1a1a;
  display: block;
}

.timeline-time {
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
  display: block;
}

.timeline-comment {
  font-size: 26rpx;
  color: #666;
  margin-top: 8rpx;
  display: block;
}

.tracking-card {
  background: #fafafa;
  border-radius: 12rpx;
  padding: 20rpx;
}

.tracking-row {
  display: flex;
  justify-content: space-between;
  padding: 12rpx 0;
}

.label {
  color: #666;
  font-size: 28rpx;
}

.value {
  color: #1a1a1a;
  font-size: 28rpx;
}
</style>

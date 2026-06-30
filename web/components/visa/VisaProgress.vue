<template>
  <div class="visa-progress">
    <!-- Progress Bar -->
    <div class="progress-bar">
      <div
        v-for="(step, index) in steps"
        :key="step.status"
        class="progress-step"
        :class="getStepClass(step, index)"
      >
        <div class="step-icon">
          <span v-if="isCompleted(step)">✓</span>
          <span v-else-if="isCurrent(step)">{{ index + 1 }}</span>
          <span v-else>{{ index + 1 }}</span>
        </div>
        <div class="step-label">{{ step.label }}</div>
        <div v-if="index < steps.length - 1" class="step-line" :class="{ completed: isLineCompleted(index) }" />
      </div>
    </div>

    <!-- Current Status -->
    <div class="current-status">
      <div class="status-badge" :class="getStatusBadgeClass()">
        {{ currentStatusName }}
      </div>
      <div v-if="estimatedDate" class="estimated-date">
        预计完成时间：{{ formatDate(estimatedDate) }}
      </div>
    </div>

    <!-- Timeline -->
    <div v-if="timeline && timeline.length > 0" class="timeline">
      <h4>办理进度</h4>
      <div v-for="item in timeline" :key="item.status" class="timeline-item">
        <div class="timeline-dot" :class="{ active: item.is_current, completed: item.is_completed }" />
        <div class="timeline-content">
          <div class="timeline-header">
            <span class="timeline-status">{{ item.status_name }}</span>
            <span class="timeline-time">{{ formatDate(item.timestamp) }}</span>
          </div>
          <div v-if="item.comment" class="timeline-comment">{{ item.comment }}</div>
        </div>
      </div>
    </div>

    <!-- Tracking Info -->
    <div v-if="trackingCompany && trackingNumber" class="tracking-info">
      <h4>物流信息</h4>
      <div class="tracking-details">
        <div class="tracking-row">
          <span class="label">快递公司</span>
          <span class="value">{{ trackingCompany }}</span>
        </div>
        <div class="tracking-row">
          <span class="label">运单号</span>
          <span class="value">{{ trackingNumber }}</span>
        </div>
      </div>
      <div v-if="trackingTimeline && trackingTimeline.length > 0" class="tracking-timeline">
        <div v-for="(event, index) in trackingTimeline" :key="index" class="tracking-event">
          <div class="event-dot" :class="{ first: index === 0 }" />
          <div class="event-content">
            <div class="event-time">{{ formatDate(event.time) }}</div>
            <div class="event-location">{{ event.location }}</div>
            <div class="event-action">{{ event.action }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Material Upload Section -->
    <div v-if="showMaterials" class="materials-section">
      <h4>签证材料</h4>
      <div class="material-list">
        <div v-for="material in materials" :key="material.id" class="material-item">
          <div class="material-info">
            <span class="material-name">{{ material.material_name }}</span>
            <span v-if="material.is_required" class="required">必填</span>
            <span v-else class="optional">选填</span>
          </div>
          <div class="material-status" :class="getMaterialStatusClass(material.status)">
            {{ getMaterialStatusName(material.status) }}
          </div>
          <div v-if="material.review_comment" class="material-comment">
            {{ material.review_comment }}
          </div>
          <div class="material-actions">
            <button
              v-if="material.status === 'pending' || material.status === 'rejected' || material.status === 'supplement'"
              class="upload-btn"
              @click="uploadMaterial(material)"
            >
              {{ material.status === 'pending' ? '上传' : '重新上传' }}
            </button>
            <button v-if="material.file_url" class="view-btn" @click="viewMaterial(material)">
              查看
            </button>
          </div>
        </div>
      </div>

      <!-- Submit Button -->
      <div v-if="canSubmit" class="submit-section">
        <button class="submit-btn" @click="submitMaterials">
          提交全部材料
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

// Props
const props = defineProps<{
  currentStatus: string
  currentStatusName: string
  estimatedDate?: string
  timeline?: Array<{
    status: string
    status_name: string
    timestamp: string
    comment?: string
    is_current: boolean
    is_completed: boolean
  }>
  trackingCompany?: string
  trackingNumber?: string
  trackingTimeline?: Array<{
    time: string
    location: string
    action: string
  }>
  materials?: Array<{
    id: number
    material_name: string
    is_required: boolean
    status: string
    file_url?: string
    review_comment?: string
  }>
  showMaterials?: boolean
}>()

const emit = defineEmits<{
  (e: 'upload', material: any): void
  (e: 'view', material: any): void
  (e: 'submit'): void
}>()

// Steps definition
const steps = [
  { status: 'pending_submit', label: '待提交' },
  { status: 'reviewing', label: '审核中' },
  { status: 'submitted', label: '已送签' },
  { status: 'approved', label: '已出签' },
]

// Status order for comparison
const statusOrder = ['pending_submit', 'reviewing', 'submitted', 'approved', 'rejected']

// Methods
const getStatusIndex = (status: string) => statusOrder.indexOf(status)
const currentIndex = computed(() => getStatusIndex(props.currentStatus))

const isCompleted = (step: any) => {
  const stepIndex = getStatusIndex(step.status)
  return stepIndex < currentIndex.value && props.currentStatus !== 'rejected'
}

const isCurrent = (step: any) => {
  return step.status === props.currentStatus
}

const isLineCompleted = (index: number) => {
  return index < currentIndex.value && props.currentStatus !== 'rejected'
}

const getStepClass = (step: any, index: number) => {
  if (props.currentStatus === 'rejected' && step.status === 'approved') {
    return 'rejected'
  }
  if (isCompleted(step)) return 'completed'
  if (isCurrent(step)) return 'current'
  return 'pending'
}

const getStatusBadgeClass = () => {
  switch (props.currentStatus) {
    case 'pending_submit': return 'badge-pending'
    case 'reviewing': return 'badge-reviewing'
    case 'submitted': return 'badge-submitted'
    case 'approved': return 'badge-approved'
    case 'rejected': return 'badge-rejected'
    default: return ''
  }
}

const getMaterialStatusClass = (status: string) => {
  switch (status) {
    case 'pending': return 'status-pending'
    case 'submitted': return 'status-submitted'
    case 'approved': return 'status-approved'
    case 'rejected': return 'status-rejected'
    case 'supplement': return 'status-supplement'
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
  if (!props.materials) return false
  return props.materials.some(m => m.status === 'pending') &&
    props.currentStatus === 'pending_submit'
})

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

const uploadMaterial = (material: any) => {
  emit('upload', material)
}

const viewMaterial = (material: any) => {
  emit('view', material)
}

const submitMaterials = () => {
  emit('submit')
}
</script>

<style scoped>
.visa-progress {
  background: white;
  border-radius: 12px;
  padding: 24px;
}

/* Progress Bar */
.progress-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  position: relative;
}

.progress-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  flex: 1;
}

.step-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 16px;
  background: #f0f0f0;
  color: #999;
  z-index: 1;
}

.progress-step.completed .step-icon {
  background: #52c41a;
  color: white;
}

.progress-step.current .step-icon {
  background: #1890ff;
  color: white;
}

.progress-step.rejected .step-icon {
  background: #ff4d4f;
  color: white;
}

.step-label {
  margin-top: 8px;
  font-size: 13px;
  color: #666;
}

.progress-step.completed .step-label,
.progress-step.current .step-label {
  color: #1a1a1a;
  font-weight: 500;
}

.step-line {
  position: absolute;
  top: 20px;
  left: 50%;
  width: 100%;
  height: 2px;
  background: #f0f0f0;
  z-index: 0;
}

.step-line.completed {
  background: #52c41a;
}

/* Current Status */
.current-status {
  text-align: center;
  margin-bottom: 24px;
}

.status-badge {
  display: inline-block;
  padding: 8px 24px;
  border-radius: 20px;
  font-size: 16px;
  font-weight: 500;
}

.badge-pending { background: #f0f0f0; color: #666; }
.badge-reviewing { background: #e6f7ff; color: #1890ff; }
.badge-submitted { background: #fff7e6; color: #fa8c16; }
.badge-approved { background: #f6ffed; color: #52c41a; }
.badge-rejected { background: #fff2f0; color: #ff4d4f; }

.estimated-date {
  margin-top: 8px;
  color: #666;
  font-size: 14px;
}

/* Timeline */
.timeline {
  margin-bottom: 24px;
}

.timeline h4 {
  margin-bottom: 16px;
  color: #1a1a1a;
}

.timeline-item {
  display: flex;
  gap: 16px;
  padding: 16px 0;
  border-left: 2px solid #f0f0f0;
  margin-left: 8px;
  padding-left: 24px;
  position: relative;
}

.timeline-dot {
  position: absolute;
  left: -9px;
  top: 20px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #f0f0f0;
  border: 2px solid white;
}

.timeline-dot.completed {
  background: #52c41a;
}

.timeline-dot.active {
  background: #1890ff;
}

.timeline-content {
  flex: 1;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}

.timeline-status {
  font-weight: 500;
  color: #1a1a1a;
}

.timeline-time {
  color: #999;
  font-size: 13px;
}

.timeline-comment {
  color: #666;
  font-size: 14px;
}

/* Tracking Info */
.tracking-info {
  margin-bottom: 24px;
}

.tracking-info h4 {
  margin-bottom: 16px;
}

.tracking-details {
  background: #fafafa;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.tracking-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
}

.tracking-event {
  display: flex;
  gap: 16px;
  padding: 12px 0;
  border-left: 2px solid #f0f0f0;
  margin-left: 8px;
  padding-left: 24px;
  position: relative;
}

.event-dot {
  position: absolute;
  left: -7px;
  top: 16px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #f0f0f0;
}

.event-dot.first {
  background: #1890ff;
}

.event-time {
  font-size: 13px;
  color: #999;
}

.event-location {
  font-weight: 500;
  color: #1a1a1a;
}

.event-action {
  color: #666;
}

/* Materials Section */
.materials-section h4 {
  margin-bottom: 16px;
}

.material-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
  margin-bottom: 12px;
}

.material-info {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.material-name {
  font-weight: 500;
}

.required {
  color: #ff4d4f;
  font-size: 12px;
}

.optional {
  color: #999;
  font-size: 12px;
}

.material-status {
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 13px;
}

.status-pending { background: #f0f0f0; color: #666; }
.status-submitted { background: #e6f7ff; color: #1890ff; }
.status-approved { background: #f6ffed; color: #52c41a; }
.status-rejected { background: #fff2f0; color: #ff4d4f; }
.status-supplement { background: #fff7e6; color: #fa8c16; }

.material-comment {
  font-size: 13px;
  color: #ff4d4f;
  margin-top: 4px;
}

.material-actions {
  display: flex;
  gap: 8px;
}

.upload-btn, .view-btn {
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
}

.upload-btn {
  background: #1890ff;
  color: white;
  border: none;
}

.view-btn {
  background: white;
  border: 1px solid #ddd;
  color: #666;
}

.submit-section {
  margin-top: 24px;
  text-align: center;
}

.submit-btn {
  padding: 12px 48px;
  background: #1890ff;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  cursor: pointer;
}
</style>

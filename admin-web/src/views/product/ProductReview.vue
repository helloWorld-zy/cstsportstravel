<template>
  <div class="product-review">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>产品审核</span>
          <el-button @click="loadReviewQueue">刷新</el-button>
        </div>
      </template>

      <!-- Filters -->
      <div class="filter-bar">
        <el-input
          v-model="filters.keyword"
          placeholder="搜索产品名称/编号"
          clearable
          style="width: 200px"
          @clear="loadReviewQueue"
          @keyup.enter="loadReviewQueue"
        />
        <el-select v-model="filters.status" placeholder="审核状态" clearable style="width: 160px" @change="loadReviewQueue">
          <el-option label="全部待审" value="" />
          <el-option label="待审核" value="pending_review" />
          <el-option label="变更待审" value="change_pending_review" />
        </el-select>
        <el-button type="primary" @click="loadReviewQueue">搜索</el-button>
      </div>

      <!-- Review Queue Table -->
      <el-table :data="reviewItems" style="width: 100%" v-loading="isLoading">
        <el-table-column prop="product_no" label="产品编号" width="180" />
        <el-table-column prop="product_name" label="产品名称" min-width="200" />
        <el-table-column label="目的地" width="150">
          <template #default="{ row }">
            {{ (row.destination_cities || []).join('·') }}
          </template>
        </el-table-column>
        <el-table-column prop="days" label="天数" width="80" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.status === 'pending_review' ? 'warning' : 'danger'">
              {{ row.status === 'pending_review' ? '待审核' : '变更待审' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="提交时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="info" @click="handlePreview(row)">查看详情</el-button>
            <el-button size="small" type="success" @click="handleApprove(row)">通过</el-button>
            <el-button size="small" type="danger" @click="handleReject(row)">驳回</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @current-change="loadReviewQueue"
          @size-change="loadReviewQueue"
        />
      </div>
    </el-card>

    <!-- Product Preview Dialog -->
    <el-dialog v-model="showPreview" title="产品详情预览" width="80%" top="5vh">
      <div v-if="previewProduct" class="preview-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="产品编号">{{ previewProduct.product_no }}</el-descriptions-item>
          <el-descriptions-item label="产品名称">{{ previewProduct.product_name }}</el-descriptions-item>
          <el-descriptions-item label="出发城市">{{ previewProduct.origin_city }}</el-descriptions-item>
          <el-descriptions-item label="目的地">{{ (previewProduct.destination_cities || []).join('、') }}</el-descriptions-item>
          <el-descriptions-item label="行程天数">{{ previewProduct.days }}天{{ previewProduct.nights }}晚</el-descriptions-item>
          <el-descriptions-item label="交通方式">{{ previewProduct.transport_mode }}</el-descriptions-item>
          <el-descriptions-item label="产品等级">{{ previewProduct.product_grade }}</el-descriptions-item>
          <el-descriptions-item label="成团设置">{{ previewProduct.min_group_size }}-{{ previewProduct.max_group_size }}人</el-descriptions-item>
        </el-descriptions>

        <el-divider />

        <h4>产品简介</h4>
        <p>{{ previewProduct.summary || '无' }}</p>

        <h4>费用包含</h4>
        <p>{{ previewProduct.fee_included || '无' }}</p>

        <h4>费用不含</h4>
        <p>{{ previewProduct.fee_excluded || '无' }}</p>

        <h4>预订须知</h4>
        <p>{{ previewProduct.booking_notes || '无' }}</p>

        <el-divider />

        <h4>行程安排</h4>
        <div v-if="previewItinerary.length > 0">
          <div v-for="day in previewItinerary" :key="day.day_no" class="preview-day">
            <el-tag type="primary" size="small">Day {{ day.day_no }}</el-tag>
            <strong style="margin-left: 8px">{{ day.title }}</strong>
            <p v-if="day.description" style="margin: 4px 0 0 32px; color: #606266">{{ day.description }}</p>
            <div v-if="day.hotel" style="margin-left: 32px; color: #909399; font-size: 13px">
              住宿：{{ day.hotel }}
            </div>
          </div>
        </div>
        <el-empty v-else description="暂无行程" />
      </div>
      <template #footer>
        <el-button @click="showPreview = false">关闭</el-button>
        <el-button type="success" @click="handleApproveById">通过</el-button>
        <el-button type="danger" @click="handleRejectById">驳回</el-button>
      </template>
    </el-dialog>

    <!-- Reject Reason Dialog -->
    <el-dialog v-model="showRejectDialog" title="驳回原因" width="450px">
      <el-form :model="rejectForm" label-width="80px">
        <el-form-item label="驳回原因" required>
          <el-input
            v-model="rejectForm.reason"
            type="textarea"
            :rows="4"
            maxlength="500"
            show-word-limit
            placeholder="请填写驳回原因，将反馈给供应商"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRejectDialog = false">取消</el-button>
        <el-button type="danger" :loading="rejecting" @click="confirmReject">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/request'
import dayjs from 'dayjs'

interface ReviewProduct {
  id: number
  product_no: string
  product_name: string
  destination_cities: string[]
  days: number
  nights: number
  origin_city: string
  transport_mode: string
  product_grade: string
  min_group_size: number
  max_group_size: number
  summary: string
  fee_included: string
  fee_excluded: string
  booking_notes: string
  status: string
  updated_at: string
}

interface ItineraryDay {
  day_no: number
  title: string
  description: string
  hotel: string
}

const reviewItems = ref<ReviewProduct[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const isLoading = ref(false)
const showPreview = ref(false)
const showRejectDialog = ref(false)
const rejecting = ref(false)

const previewProduct = ref<ReviewProduct | null>(null)
const previewItinerary = ref<ItineraryDay[]>([])
const rejectTargetID = ref(0)

const filters = reactive({
  keyword: '',
  status: '',
})

const rejectForm = reactive({
  reason: '',
})

function formatTime(t: string) {
  return t ? dayjs(t).format('YYYY-MM-DD HH:mm') : ''
}

async function loadReviewQueue() {
  isLoading.value = true
  try {
    const params: Record<string, any> = {
      page: page.value,
      page_size: pageSize.value,
    }
    if (filters.keyword) params.keyword = filters.keyword
    if (filters.status) params.status = filters.status

    const data = await adminApi.get<{ items: ReviewProduct[]; total: number }>(
      '/admin/products/review-queue',
      { params },
    )
    reviewItems.value = data.items || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    isLoading.value = false
  }
}

async function handlePreview(row: ReviewProduct) {
  try {
    const product = await adminApi.get<ReviewProduct>(`/admin/products/${row.id}`)
    previewProduct.value = product

    try {
      const itin = await adminApi.get<any>(`/admin/products/${row.id}/itinerary`)
      previewItinerary.value = itin?.itineraries || []
    } catch {
      previewItinerary.value = []
    }

    showPreview.value = true
  } catch (e: any) {
    ElMessage.error(e.message || '加载详情失败')
  }
}

async function handleApprove(row: ReviewProduct) {
  try {
    await ElMessageBox.confirm(`确定审核通过产品「${row.product_name}」？`, '确认审核')
    await adminApi.put(`/admin/products/${row.id}/approve`, { note: '审核通过' })
    ElMessage.success('审核通过')
    loadReviewQueue()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败')
  }
}

function handleApproveById() {
  if (previewProduct.value) {
    handleApprove(previewProduct.value)
    showPreview.value = false
  }
}

function handleReject(row: ReviewProduct) {
  rejectTargetID.value = row.id
  rejectForm.reason = ''
  showRejectDialog.value = true
}

function handleRejectById() {
  if (previewProduct.value) {
    rejectTargetID.value = previewProduct.value.id
    rejectForm.reason = ''
    showRejectDialog.value = false
    showRejectDialog.value = true
  }
}

async function confirmReject() {
  if (!rejectForm.reason.trim()) {
    ElMessage.warning('请填写驳回原因')
    return
  }
  rejecting.value = true
  try {
    await adminApi.put(`/admin/products/${rejectTargetID.value}/reject`, {
      reason: rejectForm.reason,
    })
    ElMessage.success('已驳回')
    showRejectDialog.value = false
    loadReviewQueue()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    rejecting.value = false
  }
}

onMounted(() => {
  loadReviewQueue()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.preview-content h4 {
  margin: 12px 0 8px;
  color: #303133;
}
.preview-content p {
  color: #606266;
  line-height: 1.6;
}
.preview-day {
  margin-bottom: 12px;
  padding: 8px 0;
  border-bottom: 1px dashed #ebeef5;
}
</style>

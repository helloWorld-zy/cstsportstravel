<template>
  <div class="visa-order-list">
    <div class="page-header">
      <h2>签证订单管理</h2>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <el-select v-model="filters.status" placeholder="状态筛选" clearable @change="loadOrders">
        <el-option label="全部" value="" />
        <el-option label="待提交" value="pending_submit" />
        <el-option label="审核中" value="reviewing" />
        <el-option label="已送签" value="submitted" />
        <el-option label="已出签" value="approved" />
        <el-option label="已拒签" value="rejected" />
      </el-select>
      <el-input v-model="filters.keyword" placeholder="签证订单号/用户ID" clearable @keyup.enter="loadOrders" />
      <el-button type="primary" @click="loadOrders">搜索</el-button>
    </div>

    <!-- Table -->
    <el-table :data="orders" v-loading="isLoading" stripe>
      <el-table-column prop="visa_order_no" label="签证订单号" width="180" />
      <el-table-column prop="main_order_id" label="主订单ID" width="120" />
      <el-table-column prop="user_id" label="用户ID" width="100" />
      <el-table-column prop="visa_type" label="签证类型" width="120" />
      <el-table-column prop="status" label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">{{ getStatusName(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="visa_fee" label="签证费用" width="120">
        <template #default="{ row }">
          ¥{{ (row.visa_fee / 100).toFixed(2) }}
        </template>
      </el-table-column>
      <el-table-column prop="estimated_completion_date" label="预计完成" width="120" />
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDateTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewDetail(row)">详情</el-button>
          <el-button
            v-if="row.status === 'reviewing'"
            size="small"
            type="primary"
            @click="reviewOrder(row)"
          >
            审核
          </el-button>
          <el-button
            v-if="row.status === 'reviewing'"
            size="small"
            type="success"
            @click="updateStatus(row, 'submitted')"
          >
            送签
          </el-button>
          <el-button
            v-if="row.status === 'submitted'"
            size="small"
            type="success"
            @click="updateStatus(row, 'approved')"
          >
            出签
          </el-button>
          <el-button
            v-if="row.status === 'submitted'"
            size="small"
            type="danger"
            @click="showRejectDialog(row)"
          >
            拒签
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadOrders"
        @size-change="loadOrders"
      />
    </div>

    <!-- Review Dialog -->
    <el-dialog v-model="reviewDialogVisible" title="材料审核" width="800px">
      <div v-if="currentOrder" class="review-content">
        <h4>签证订单：{{ currentOrder.visa_order_no }}</h4>
        <div class="materials-list">
          <div v-for="material in currentOrder.materials" :key="material.id" class="material-item">
            <div class="material-info">
              <span class="name">{{ material.material_name }}</span>
              <el-tag size="small" :type="getMaterialStatusType(material.status)">
                {{ getMaterialStatusName(material.status) }}
              </el-tag>
            </div>
            <div v-if="material.file_url" class="material-file">
              <el-button size="small" @click="previewFile(material.file_url)">查看文件</el-button>
            </div>
            <div class="material-review">
              <el-input v-model="material.review_comment" placeholder="审核意见" />
              <div class="review-actions">
                <el-button size="small" type="success" @click="reviewMaterial(material, 'approved')">通过</el-button>
                <el-button size="small" type="danger" @click="reviewMaterial(material, 'rejected')">不通过</el-button>
                <el-button size="small" type="warning" @click="reviewMaterial(material, 'supplement')">需补充</el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- Reject Dialog -->
    <el-dialog v-model="rejectDialogVisible" title="拒签确认" width="500px">
      <el-form :model="rejectForm">
        <el-form-item label="拒签原因" required>
          <el-input v-model="rejectForm.reason" type="textarea" rows="4" placeholder="请输入拒签原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmReject">确认拒签</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

// State
const isLoading = ref(false)
const orders = ref<any[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const reviewDialogVisible = ref(false)
const rejectDialogVisible = ref(false)
const currentOrder = ref<any>(null)

const filters = reactive({
  status: '',
  keyword: '',
})

const rejectForm = reactive({
  reason: '',
})

// Methods
const loadOrders = async () => {
  isLoading.value = true
  try {
    const params: Record<string, any> = {
      page: currentPage.value,
      page_size: pageSize.value,
    }
    if (filters.status) params.status = filters.status

    const data = await fetch('/api/v2/admin/visa-orders?' + new URLSearchParams(params))
    const result = await data.json()
    orders.value = result.items || []
    total.value = result.total || 0
  } catch (error) {
    console.error('Failed to load orders:', error)
  } finally {
    isLoading.value = false
  }
}

const viewDetail = (order: any) => {
  currentOrder.value = order
  reviewDialogVisible.value = true
}

const reviewOrder = (order: any) => {
  currentOrder.value = order
  reviewDialogVisible.value = true
}

const reviewMaterial = async (material: any, action: string) => {
  try {
    await fetch(`/api/v2/admin/visa-orders/${currentOrder.value.id}/review`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        action,
        comments: [{ materialId: material.id, comment: material.review_comment }],
      }),
    })
    ElMessage.success('审核完成')
    loadOrders()
  } catch (error) {
    ElMessage.error('审核失败')
  }
}

const updateStatus = async (order: any, status: string) => {
  try {
    await ElMessageBox.confirm(`确认将状态更新为${getStatusName(status)}？`, '确认操作')
    await fetch(`/api/v2/admin/visa-orders/${order.id}/status`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status }),
    })
    ElMessage.success('状态更新成功')
    loadOrders()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('状态更新失败')
    }
  }
}

const showRejectDialog = (order: any) => {
  currentOrder.value = order
  rejectForm.reason = ''
  rejectDialogVisible.value = true
}

const confirmReject = async () => {
  if (!rejectForm.reason) {
    ElMessage.warning('请输入拒签原因')
    return
  }
  try {
    await fetch(`/api/v2/admin/visa-orders/${currentOrder.value.id}/status`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        status: 'rejected',
        reject_reason: rejectForm.reason,
      }),
    })
    ElMessage.success('拒签操作完成')
    rejectDialogVisible.value = false
    loadOrders()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const previewFile = (url: string) => {
  window.open(url, '_blank')
}

// Helpers
const getStatusType = (status: string) => {
  switch (status) {
    case 'pending_submit': return 'info'
    case 'reviewing': return 'warning'
    case 'submitted': return ''
    case 'approved': return 'success'
    case 'rejected': return 'danger'
    default: return 'info'
  }
}

const getStatusName = (status: string) => {
  switch (status) {
    case 'pending_submit': return '待提交'
    case 'reviewing': return '审核中'
    case 'submitted': return '已送签'
    case 'approved': return '已出签'
    case 'rejected': return '已拒签'
    default: return status
  }
}

const getMaterialStatusType = (status: string) => {
  switch (status) {
    case 'pending': return 'info'
    case 'submitted': return 'warning'
    case 'approved': return 'success'
    case 'rejected': return 'danger'
    case 'supplement': return 'warning'
    default: return 'info'
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

const formatDateTime = (dateStr: string) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('zh-CN')
}

// Lifecycle
onMounted(() => {
  loadOrders()
})
</script>

<style scoped>
.visa-order-list {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.review-content h4 {
  margin-bottom: 16px;
}

.material-item {
  padding: 16px;
  border: 1px solid #eee;
  border-radius: 8px;
  margin-bottom: 12px;
}

.material-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.material-info .name {
  font-weight: 500;
}

.material-file {
  margin-bottom: 12px;
}

.review-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}
</style>

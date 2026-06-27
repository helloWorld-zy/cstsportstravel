<template>
  <div class="refund-review">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>退款审核</span>
          <el-button @click="loadRefunds">刷新</el-button>
        </div>
      </template>

      <!-- Filters -->
      <div class="filter-bar">
        <el-input
          v-model="filters.order_no"
          placeholder="订单号"
          clearable
          style="width: 180px"
          @clear="loadRefunds"
          @keyup.enter="loadRefunds"
        />
        <el-select v-model="filters.status" placeholder="退款状态" clearable style="width: 140px" @change="loadRefunds">
          <el-option label="全部" value="" />
          <el-option label="待审核" value="pending" />
          <el-option label="已批准" value="approved" />
          <el-option label="已驳回" value="failed" />
          <el-option label="已退款" value="success" />
        </el-select>
        <el-button type="primary" @click="loadRefunds">搜索</el-button>
      </div>

      <!-- Refund Table -->
      <el-table :data="refunds" style="width: 100%" v-loading="isLoading">
        <el-table-column prop="refund_no" label="退款单号" width="200" />
        <el-table-column prop="order_no" label="订单号" width="200" />
        <el-table-column prop="product_name" label="产品名称" min-width="180" />
        <el-table-column prop="user_phone" label="用户手机" width="130" />
        <el-table-column label="退款金额" width="120">
          <template #default="{ row }">
            <span class="amount">¥{{ (row.refund_amount / 100).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="approval_level_cn" label="审批级别" width="130">
          <template #default="{ row }">
            <el-tag :type="levelType(row.approval_level)" size="small">
              {{ row.approval_level_cn }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="申请时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="info" @click="handleDetail(row)">详情</el-button>
            <template v-if="row.status === 'pending'">
              <el-button size="small" type="success" @click="handleApprove(row)">通过</el-button>
              <el-button size="small" type="danger" @click="handleReject(row)">驳回</el-button>
            </template>
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
          @current-change="loadRefunds"
          @size-change="loadRefunds"
        />
      </div>
    </el-card>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetail" title="退款详情" width="700px">
      <template v-if="detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="退款单号">{{ detail.refund_no }}</el-descriptions-item>
          <el-descriptions-item label="订单号">{{ detail.order_no }}</el-descriptions-item>
          <el-descriptions-item label="产品名称">{{ detail.product_name }}</el-descriptions-item>
          <el-descriptions-item label="退款类型">{{ detail.refund_type === 'full' ? '全额退款' : '部分退款' }}</el-descriptions-item>
          <el-descriptions-item label="应付金额">
            ¥{{ (detail.payable_amount / 100).toFixed(2) }}
          </el-descriptions-item>
          <el-descriptions-item label="退款金额">
            <span class="amount">¥{{ (detail.refund_amount / 100).toFixed(2) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="距出发天数">{{ detail.days_before_departure }}天</el-descriptions-item>
          <el-descriptions-item label="退款比例">{{ detail.refund_percentage }}%</el-descriptions-item>
          <el-descriptions-item label="匹配规则" :span="2">{{ detail.matching_rule }}</el-descriptions-item>
          <el-descriptions-item label="审批级别">
            <el-tag :type="levelType(detail.approval_level)">{{ detail.approval_level_cn }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="退款原因" :span="2">{{ detail.refund_reason }}</el-descriptions-item>
        </el-descriptions>
      </template>
      <template #footer>
        <el-button @click="showDetail = false">关闭</el-button>
        <template v-if="detail && detail.status === 'pending'">
          <el-button type="success" @click="handleApproveById">通过</el-button>
          <el-button type="danger" @click="handleRejectById">驳回</el-button>
        </template>
      </template>
    </el-dialog>

    <!-- Reject Dialog -->
    <el-dialog v-model="showRejectDialog" title="驳回退款" width="450px">
      <el-form label-width="80px">
        <el-form-item label="驳回原因" required>
          <el-input
            v-model="rejectReason"
            type="textarea"
            :rows="4"
            maxlength="500"
            show-word-limit
            placeholder="请填写驳回原因"
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

interface RefundItem {
  id: number
  order_no: string
  product_name: string
  user_phone: string
  refund_no: string
  refund_amount: number
  refund_reason: string
  refund_type: string
  status: string
  approval_level: string
  approval_level_cn: string
  created_at: string
}

interface RefundDetail extends RefundItem {
  payable_amount: number
  days_before_departure: number
  matching_rule: string
  refund_percentage: number
  approved_by: number | null
  approved_at: string | null
  completed_at: string | null
}

const refunds = ref<RefundItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const isLoading = ref(false)
const showDetail = ref(false)
const showRejectDialog = ref(false)
const rejecting = ref(false)
const detail = ref<RefundDetail | null>(null)
const rejectTargetID = ref(0)
const rejectReason = ref('')

const filters = reactive({
  order_no: '',
  status: '',
})

function formatTime(t: string) {
  return t ? dayjs(t).format('YYYY-MM-DD HH:mm') : ''
}

const statusType = (status: string) => {
  const map: Record<string, string> = {
    pending: 'warning',
    approved: 'success',
    processing: 'primary',
    success: 'success',
    failed: 'danger',
  }
  return (map[status] || 'info') as any
}

const statusLabel = (status: string) => {
  const map: Record<string, string> = {
    pending: '待审核',
    approved: '已批准',
    processing: '处理中',
    success: '已退款',
    failed: '已驳回',
  }
  return map[status] || status
}

const levelType = (level: string) => {
  const map: Record<string, string> = {
    operator: 'info',
    finance_director: 'warning',
    director: 'danger',
  }
  return (map[level] || 'info') as any
}

async function loadRefunds() {
  isLoading.value = true
  try {
    const params: Record<string, any> = {
      page: page.value,
      page_size: pageSize.value,
    }
    if (filters.order_no) params.order_no = filters.order_no
    if (filters.status) params.status = filters.status

    const data = await adminApi.get<{ items: RefundItem[]; total: number }>('/admin/refunds', { params })
    refunds.value = data.items || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    isLoading.value = false
  }
}

async function handleDetail(row: RefundItem) {
  try {
    const data = await adminApi.get<RefundDetail>(`/admin/refunds/${row.id}`)
    detail.value = data
    showDetail.value = true
  } catch (e: any) {
    ElMessage.error(e.message || '加载详情失败')
  }
}

async function handleApprove(row: RefundItem) {
  try {
    await ElMessageBox.confirm(
      `确定审核通过退款 ¥${(row.refund_amount / 100).toFixed(2)}？`,
      '确认审核',
    )
    await adminApi.put(`/admin/refunds/${row.id}/approve`, { note: '审核通过' })
    ElMessage.success('已批准')
    loadRefunds()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败')
  }
}

function handleApproveById() {
  if (detail.value) {
    handleApprove(detail.value)
    showDetail.value = false
  }
}

function handleReject(row: RefundItem) {
  rejectTargetID.value = row.id
  rejectReason.value = ''
  showRejectDialog.value = true
}

function handleRejectById() {
  if (detail.value) {
    rejectTargetID.value = detail.value.id
    rejectReason.value = ''
    showDetail.value = false
    showRejectDialog.value = true
  }
}

async function confirmReject() {
  if (!rejectReason.value.trim()) {
    ElMessage.warning('请填写驳回原因')
    return
  }
  rejecting.value = true
  try {
    await adminApi.put(`/admin/refunds/${rejectTargetID.value}/reject`, {
      reason: rejectReason.value,
    })
    ElMessage.success('已驳回')
    showRejectDialog.value = false
    loadRefunds()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    rejecting.value = false
  }
}

onMounted(() => {
  loadRefunds()
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
.amount {
  color: #f56c6c;
  font-weight: 600;
}
</style>

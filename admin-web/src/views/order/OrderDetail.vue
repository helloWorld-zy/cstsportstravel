<template>
  <div class="order-detail">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>订单详情</span>
          <el-button @click="router.push('/orders')">返回列表</el-button>
        </div>
      </template>

      <div v-loading="isLoading">
        <template v-if="order">
          <!-- Basic Info -->
          <el-descriptions title="基本信息" :column="3" border>
            <el-descriptions-item label="订单号">{{ order.order_no }}</el-descriptions-item>
            <el-descriptions-item label="产品名称">{{ order.product_name }}</el-descriptions-item>
            <el-descriptions-item label="订单状态">
              <el-tag :type="statusType(order.order_status)">{{ statusLabel(order.order_status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="联系人">{{ order.contact_name }}</el-descriptions-item>
            <el-descriptions-item label="联系电话">{{ order.contact_phone }}</el-descriptions-item>
            <el-descriptions-item label="下单渠道">{{ order.channel }}</el-descriptions-item>
            <el-descriptions-item label="下单时间">{{ formatTime(order.created_at) }}</el-descriptions-item>
            <el-descriptions-item label="支付时间">{{ formatTime(order.paid_at) || '-' }}</el-descriptions-item>
            <el-descriptions-item label="取消时间">{{ formatTime(order.cancelled_at) || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Fee Breakdown -->
          <el-descriptions title="费用明细" :column="2" border style="margin-top: 24px">
            <el-descriptions-item label="产品总价">
              <span class="amount">¥{{ (order.total_amount / 100).toFixed(2) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="单房差">
              ¥{{ (order.single_supplement_amount / 100).toFixed(2) }}
            </el-descriptions-item>
            <el-descriptions-item label="附加服务">
              ¥{{ (order.addon_amount / 100).toFixed(2) }}
            </el-descriptions-item>
            <el-descriptions-item label="优惠金额">
              -¥{{ (order.discount_amount / 100).toFixed(2) }}
            </el-descriptions-item>
            <el-descriptions-item label="应付金额">
              <span class="amount highlight">¥{{ (order.payable_amount / 100).toFixed(2) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="支付状态">{{ order.payment_status }}</el-descriptions-item>
          </el-descriptions>

          <!-- Travellers -->
          <el-divider />
          <h4>出游人信息</h4>
          <el-table :data="order.travellers" style="width: 100%">
            <el-table-column label="类型" width="100">
              <template #default="{ row }">
                <el-tag v-if="row.is_infant" type="info" size="small">婴儿</el-tag>
                <el-tag v-else-if="row.is_child" type="warning" size="small">儿童</el-tag>
                <el-tag v-else type="primary" size="small">成人</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="phone" label="手机号" width="130" />
            <el-table-column prop="gender" label="性别" width="80" />
            <el-table-column label="出生日期" width="120">
              <template #default="{ row }">{{ row.birth_date || '-' }}</template>
            </el-table-column>
          </el-table>

          <!-- Payment Records -->
          <el-divider />
          <h4>支付记录</h4>
          <el-table :data="order.payments" style="width: 100%">
            <el-table-column prop="payment_no" label="支付单号" width="200" />
            <el-table-column prop="channel" label="渠道" width="100" />
            <el-table-column prop="method" label="方式" width="100" />
            <el-table-column label="金额" width="120">
              <template #default="{ row }">¥{{ (row.amount / 100).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100" />
            <el-table-column prop="channel_trade_no" label="渠道交易号" width="200" />
            <el-table-column label="支付时间" width="170">
              <template #default="{ row }">{{ formatTime(row.paid_at) || '-' }}</template>
            </el-table-column>
          </el-table>

          <!-- Refund Records -->
          <template v-if="order.refunds && order.refunds.length > 0">
            <el-divider />
            <h4>退款记录</h4>
            <el-table :data="order.refunds" style="width: 100%">
              <el-table-column prop="refund_no" label="退款单号" width="200" />
              <el-table-column label="退款金额" width="120">
                <template #default="{ row }">
                  <span class="amount">¥{{ (row.refund_amount / 100).toFixed(2) }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="refund_reason" label="退款原因" min-width="200" />
              <el-table-column prop="approval_level" label="审批级别" width="120">
                <template #default="{ row }">{{ approvalLevelCN(row.approval_level) }}</template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="refundStatusType(row.status)">{{ refundStatusLabel(row.status) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="创建时间" width="170">
                <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
              </el-table-column>
            </el-table>
          </template>

          <!-- Status Logs -->
          <el-divider />
          <h4>操作日志</h4>
          <el-timeline>
            <el-timeline-item
              v-for="log in order.status_logs"
              :key="log.id"
              :timestamp="formatTime(log.created_at)"
              placement="top"
            >
              <p>
                <el-tag size="small">{{ statusLabel(log.from_status) }}</el-tag>
                →
                <el-tag size="small" type="success">{{ statusLabel(log.to_status) }}</el-tag>
                <span style="margin-left: 8px; color: #909399">({{ log.operator_type }})</span>
              </p>
              <p v-if="log.reason" style="color: #606266; margin-top: 4px">{{ log.reason }}</p>
            </el-timeline-item>
          </el-timeline>
        </template>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/request'
import dayjs from 'dayjs'

interface OrderDetail {
  id: number
  order_no: string
  product_name: string
  order_status: string
  payment_status: string
  contact_name: string
  contact_phone: string
  channel: string
  total_amount: number
  discount_amount: number
  payable_amount: number
  single_supplement_amount: number
  addon_amount: number
  cancel_reason: string
  created_at: string
  paid_at: string | null
  cancelled_at: string | null
  completed_at: string | null
  travellers: any[]
  payments: any[]
  refunds: any[]
  status_logs: any[]
}

const router = useRouter()
const route = useRoute()
const order = ref<OrderDetail | null>(null)
const isLoading = ref(false)

function formatTime(t: string | null | undefined) {
  return t ? dayjs(t).format('YYYY-MM-DD HH:mm:ss') : ''
}

const statusType = (status: string) => {
  const map: Record<string, string> = {
    pending_pay: 'warning',
    paid_full: 'primary',
    pending_travel: 'primary',
    in_travel: 'success',
    completed: 'success',
    cancelled: 'info',
    refunding: 'danger',
    refunded: 'danger',
    closed: 'info',
  }
  return (map[status] || 'info') as any
}

const statusLabel = (status: string) => {
  const map: Record<string, string> = {
    pending_pay: '待付款',
    paid_full: '待出行',
    pending_travel: '待出行',
    in_travel: '出行中',
    completed: '已完成',
    cancelled: '已取消',
    refunding: '退款中',
    refunded: '已退款',
    closed: '已关闭',
  }
  return map[status] || status
}

const refundStatusType = (status: string) => {
  const map: Record<string, string> = {
    pending: 'warning',
    approved: 'success',
    processing: 'primary',
    success: 'success',
    failed: 'danger',
  }
  return (map[status] || 'info') as any
}

const refundStatusLabel = (status: string) => {
  const map: Record<string, string> = {
    pending: '待审核',
    approved: '已批准',
    processing: '处理中',
    success: '已退款',
    failed: '已驳回',
  }
  return map[status] || status
}

function approvalLevelCN(level: string) {
  const map: Record<string, string> = {
    operator: '运营审批',
    finance_director: '财务主管审批',
    director: '总监审批',
  }
  return map[level] || level
}

async function loadOrder() {
  const id = route.params.id
  isLoading.value = true
  try {
    const data = await adminApi.get<OrderDetail>(`/admin/orders/${id}`)
    order.value = data
  } catch (e: any) {
    ElMessage.error(e.message || '加载订单失败')
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadOrder()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.amount {
  color: #f56c6c;
  font-weight: 600;
}
.amount.highlight {
  font-size: 16px;
}
h4 {
  margin: 16px 0 12px;
  color: #303133;
}
</style>

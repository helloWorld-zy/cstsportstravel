<template>
  <div class="order-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>订单管理</span>
        </div>
      </template>

      <!-- Filters -->
      <div class="filter-bar">
        <el-input
          v-model="filters.order_no"
          placeholder="订单号"
          clearable
          style="width: 180px"
          @clear="loadOrders"
          @keyup.enter="loadOrders"
        />
        <el-input
          v-model="filters.user_phone"
          placeholder="手机号"
          clearable
          style="width: 140px"
          @clear="loadOrders"
          @keyup.enter="loadOrders"
        />
        <el-select v-model="filters.status" placeholder="订单状态" clearable style="width: 140px" @change="loadOrders">
          <el-option label="全部" value="" />
          <el-option label="待付款" value="pending_pay" />
          <el-option label="待出行" value="paid_full" />
          <el-option label="出行中" value="in_travel" />
          <el-option label="已完成" value="completed" />
          <el-option label="已取消" value="cancelled" />
          <el-option label="退款中" value="refunding" />
          <el-option label="已退款" value="refunded" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 260px"
          @change="loadOrders"
        />
        <el-button type="primary" @click="loadOrders">搜索</el-button>
      </div>

      <!-- Table -->
      <el-table :data="orders" style="width: 100%" v-loading="isLoading">
        <el-table-column prop="order_no" label="订单号" width="200" />
        <el-table-column prop="product_name" label="产品名称" min-width="200" />
        <el-table-column prop="contact_name" label="联系人" width="100" />
        <el-table-column prop="user_phone" label="手机号" width="130" />
        <el-table-column label="人数" width="120">
          <template #default="{ row }">
            成人{{ row.adult_count }}
            <template v-if="row.child_count">/儿童{{ row.child_count }}</template>
            <template v-if="row.infant_count">/婴儿{{ row.infant_count }}</template>
          </template>
        </el-table-column>
        <el-table-column label="应付金额" width="120">
          <template #default="{ row }">
            <span class="amount">¥{{ (row.payable_amount / 100).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="order_status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.order_status)">{{ statusLabel(row.order_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="channel" label="渠道" width="80" />
        <el-table-column prop="created_at" label="下单时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="info" @click="handleDetail(row)">详情</el-button>
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
          @current-change="loadOrders"
          @size-change="loadOrders"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/request'
import dayjs from 'dayjs'

interface Order {
  id: number
  order_no: string
  product_name: string
  contact_name: string
  user_phone: string
  adult_count: number
  child_count: number
  infant_count: number
  payable_amount: number
  order_status: string
  channel: string
  created_at: string
}

const router = useRouter()
const orders = ref<Order[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const isLoading = ref(false)
const dateRange = ref<string[] | null>(null)

const filters = reactive({
  order_no: '',
  user_phone: '',
  status: '',
})

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

function formatTime(t: string) {
  return t ? dayjs(t).format('YYYY-MM-DD HH:mm') : ''
}

async function loadOrders() {
  isLoading.value = true
  try {
    const params: Record<string, any> = {
      page: page.value,
      page_size: pageSize.value,
    }
    if (filters.order_no) params.order_no = filters.order_no
    if (filters.user_phone) params.user_phone = filters.user_phone
    if (filters.status) params.status = filters.status
    if (dateRange.value && dateRange.value[0]) {
      params.date_from = dateRange.value[0]
      params.date_to = dateRange.value[1]
    }

    const data = await adminApi.get<{ items: Order[]; total: number }>('/admin/orders', { params })
    orders.value = data.items || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    isLoading.value = false
  }
}

function handleDetail(row: Order) {
  router.push(`/orders/${row.id}`)
}

onMounted(() => {
  loadOrders()
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
  flex-wrap: wrap;
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

<template>
  <div class="supplier-list">
    <div class="page-header">
      <h2>供应商管理</h2>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <el-select v-model="filters.status" placeholder="状态筛选" clearable @change="loadSuppliers">
        <el-option label="全部" value="all" />
        <el-option label="待审核" value="pending" />
        <el-option label="审核中" value="reviewing" />
        <el-option label="正常" value="active" />
        <el-option label="已暂停" value="suspended" />
        <el-option label="已终止" value="terminated" />
      </el-select>
      <el-input v-model="filters.keyword" placeholder="供应商名称/编号/信用代码" clearable @keyup.enter="loadSuppliers" />
      <el-button type="primary" @click="loadSuppliers">搜索</el-button>
    </div>

    <!-- Table -->
    <el-table :data="suppliers" v-loading="isLoading" stripe>
      <el-table-column prop="supplier_no" label="供应商编号" width="180" />
      <el-table-column prop="company_name" label="企业名称" min-width="200" />
      <el-table-column prop="unified_social_credit_code" label="信用代码" width="200" />
      <el-table-column prop="contact_name" label="联系人" width="120" />
      <el-table-column prop="settlement_cycle" label="结算周期" width="100">
        <template #default="{ row }">
          {{ { daily: '日结', weekly: '周结', monthly: '月结' }[row.settlement_cycle] || row.settlement_cycle }}
        </template>
      </el-table-column>
      <el-table-column prop="commission_rate" label="佣金比例" width="100">
        <template #default="{ row }">{{ row.commission_rate ? row.commission_rate + '%' : '-' }}</template>
      </el-table-column>
      <el-table-column prop="rating_score" label="评分" width="80" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">{{ getStatusName(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="入驻时间" width="180">
        <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewDetail(row)">详情</el-button>
          <el-button
            v-if="row.status === 'active'"
            size="small"
            type="warning"
            @click="toggleStatus(row, 'suspended')"
          >
            暂停
          </el-button>
          <el-button
            v-if="row.status === 'suspended'"
            size="small"
            type="success"
            @click="toggleStatus(row, 'active')"
          >
            恢复
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <el-pagination
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.pageSize"
      :total="pagination.total"
      layout="total, prev, pager, next"
      @current-change="loadSuppliers"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'

const router = useRouter()
const isLoading = ref(false)
const suppliers = ref<any[]>([])

const filters = reactive({
  status: 'all',
  keyword: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const loadSuppliers = async () => {
  isLoading.value = true
  try {
    const { data } = await useFetch('/api/v2/admin/suppliers', {
      params: {
        status: filters.status,
        keyword: filters.keyword,
        page: pagination.page,
        pageSize: pagination.pageSize,
      },
    })
    if (data.value?.code === 0) {
      suppliers.value = data.value.data.items || []
      pagination.total = data.value.data.total || 0
    }
  } finally {
    isLoading.value = false
  }
}

const viewDetail = (row: any) => {
  router.push(`/supplier/${row.id}`)
}

const toggleStatus = async (row: any, newStatus: string) => {
  const action = newStatus === 'suspended' ? '暂停' : '恢复'
  await ElMessageBox.confirm(`确认${action}供应商「${row.company_name}」？`, '确认操作')
  const { data } = await useFetch(`/api/v2/admin/suppliers/${row.id}/status`, {
    method: 'PUT',
    body: { status: newStatus },
  })
  if (data.value?.code === 0) {
    ElMessage.success(`已${action}`)
    loadSuppliers()
  }
}

const getStatusType = (status: string) => {
  const map: Record<string, string> = { pending: 'warning', reviewing: '', active: 'success', suspended: 'danger', terminated: 'info' }
  return map[status] || 'info'
}

const getStatusName = (status: string) => {
  const map: Record<string, string> = { pending: '待审核', reviewing: '审核中', active: '正常', suspended: '已暂停', terminated: '已终止' }
  return map[status] || status
}

const formatDateTime = (s: string) => s ? new Date(s).toLocaleString() : ''

onMounted(loadSuppliers)
</script>

<style scoped>
.supplier-list { padding: 16px; }
.page-header { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; }
</style>

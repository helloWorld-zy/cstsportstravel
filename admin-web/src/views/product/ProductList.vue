<template>
  <div class="product-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>产品管理</span>
          <el-button type="primary" @click="handleCreate">新增产品</el-button>
        </div>
      </template>

      <!-- Filters -->
      <div class="filter-bar">
        <el-input
          v-model="filters.keyword"
          placeholder="搜索产品名称"
          clearable
          style="width: 200px"
          @clear="loadProducts"
          @keyup.enter="loadProducts"
        />
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 140px" @change="loadProducts">
          <el-option label="全部" value="" />
          <el-option label="草稿" value="draft" />
          <el-option label="待审核" value="pending_review" />
          <el-option label="已上架" value="approved" />
          <el-option label="已下架" value="suspended" />
        </el-select>
        <el-input
          v-model="filters.destination"
          placeholder="目的地"
          clearable
          style="width: 140px"
          @clear="loadProducts"
          @keyup.enter="loadProducts"
        />
        <el-button type="primary" @click="loadProducts">搜索</el-button>
      </div>

      <!-- Table -->
      <el-table :data="products" style="width: 100%" v-loading="isLoading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="product_no" label="产品编号" width="180" />
        <el-table-column prop="product_name" label="产品名称" min-width="200" />
        <el-table-column prop="origin_city" label="出发城市" width="100" />
        <el-table-column label="目的地" width="150">
          <template #default="{ row }">
            {{ (row.destination_cities || []).join('·') }}
          </template>
        </el-table-column>
        <el-table-column prop="days" label="天数" width="80" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="order_count" label="订单数" width="80" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button
              v-if="row.status === 'draft'"
              size="small"
              type="success"
              @click="handleSubmitReview(row)"
            >提交审核</el-button>
            <el-button
              v-if="row.status === 'approved'"
              size="small"
              type="warning"
              @click="handleSuspend(row)"
            >下架</el-button>
            <el-button size="small" type="info" @click="handleViewDepartures(row)">团期</el-button>
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
          @current-change="loadProducts"
          @size-change="loadProducts"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/request'

interface Product {
  id: number
  product_no: string
  product_name: string
  origin_city: string
  destination_cities: string[]
  days: number
  status: string
  order_count: number
}

const products = ref<Product[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const isLoading = ref(false)

const filters = reactive({
  keyword: '',
  status: '',
  destination: '',
})

const statusType = (status: string) => {
  const map: Record<string, string> = {
    draft: 'info',
    pending_review: 'warning',
    approved: 'success',
    suspended: 'danger',
    change_pending_review: 'warning',
  }
  return map[status] || 'info'
}

const statusLabel = (status: string) => {
  const map: Record<string, string> = {
    draft: '草稿',
    pending_review: '待审核',
    approved: '已上架',
    suspended: '已下架',
    change_pending_review: '变更待审',
  }
  return map[status] || status
}

async function loadProducts() {
  isLoading.value = true
  try {
    const params: Record<string, any> = {
      page: page.value,
      page_size: pageSize.value,
    }
    if (filters.keyword) params.keyword = filters.keyword
    if (filters.status) params.status = filters.status
    if (filters.destination) params.destination = filters.destination

    const data = await adminApi.get<{ items: Product[]; total: number }>('/admin/products', { params })
    products.value = data.items || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    isLoading.value = false
  }
}

function handleCreate() {
  // TODO: navigate to product form
  ElMessage.info('产品创建功能开发中')
}

function handleEdit(row: Product) {
  // TODO: navigate to product edit form
  ElMessage.info(`编辑产品 ${row.product_no}`)
}

async function handleSubmitReview(row: Product) {
  try {
    await ElMessageBox.confirm('确定提交审核？', '确认')
    await adminApi.post(`/admin/products/${row.id}/submit-review`)
    ElMessage.success('已提交审核')
    loadProducts()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败')
  }
}

async function handleSuspend(row: Product) {
  try {
    await ElMessageBox.confirm('确定下架该产品？', '确认')
    await adminApi.put(`/admin/products/${row.id}/suspend`)
    ElMessage.success('已下架')
    loadProducts()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败')
  }
}

function handleViewDepartures(row: Product) {
  ElMessage.info(`查看团期: ${row.product_no}`)
}

onMounted(() => {
  loadProducts()
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
</style>

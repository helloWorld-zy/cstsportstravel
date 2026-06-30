<template>
  <div class="visa-material-audit">
    <div class="page-header">
      <h2>签证材料审核</h2>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <el-select v-model="filters.status" placeholder="审核状态" clearable @change="loadMaterials">
        <el-option label="全部" value="" />
        <el-option label="待审核" value="submitted" />
        <el-option label="需修改" value="rejected" />
        <el-option label="需补充" value="supplement" />
      </el-select>
      <el-button type="primary" @click="loadMaterials">刷新</el-button>
    </div>

    <!-- Material List -->
    <div class="material-grid">
      <el-card v-for="item in materials" :key="item.id" class="material-card">
        <template #header>
          <div class="card-header">
            <span>{{ item.visa_order_no }}</span>
            <el-tag :type="getStatusType(item.status)">{{ getStatusName(item.status) }}</el-tag>
          </div>
        </template>
        <div class="card-content">
          <div class="info-row">
            <span class="label">材料类型</span>
            <span class="value">{{ item.material_name }}</span>
          </div>
          <div class="info-row">
            <span class="label">用户ID</span>
            <span class="value">{{ item.user_id }}</span>
          </div>
          <div class="info-row">
            <span class="label">上传时间</span>
            <span class="value">{{ formatDateTime(item.created_at) }}</span>
          </div>
          <div v-if="item.file_url" class="file-preview">
            <el-button size="small" @click="previewFile(item.file_url)">查看文件</el-button>
          </div>
          <div class="review-section">
            <el-input
              v-model="item.review_comment"
              type="textarea"
              rows="2"
              placeholder="审核意见"
            />
            <div class="review-actions">
              <el-button size="small" type="success" @click="auditMaterial(item, 'approved')">
                通过
              </el-button>
              <el-button size="small" type="danger" @click="auditMaterial(item, 'rejected')">
                不通过
              </el-button>
              <el-button size="small" type="warning" @click="auditMaterial(item, 'supplement')">
                需补充
              </el-button>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- Empty State -->
    <el-empty v-if="materials.length === 0 && !isLoading" description="暂无待审核材料" />

    <!-- Pagination -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="loadMaterials"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

// State
const isLoading = ref(false)
const materials = ref<any[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const filters = reactive({
  status: 'submitted',
})

// Methods
const loadMaterials = async () => {
  isLoading.value = true
  try {
    const params: Record<string, any> = {
      page: currentPage.value,
      page_size: pageSize.value,
    }
    if (filters.status) params.status = filters.status

    const data = await fetch('/api/v2/admin/visa-materials?' + new URLSearchParams(params))
    const result = await data.json()
    materials.value = result.items || []
    total.value = result.total || 0
  } catch (error) {
    console.error('Failed to load materials:', error)
  } finally {
    isLoading.value = false
  }
}

const auditMaterial = async (material: any, action: string) => {
  try {
    await fetch(`/api/v2/admin/visa-materials/${material.id}/audit`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        action,
        comment: material.review_comment,
      }),
    })
    ElMessage.success('审核完成')
    loadMaterials()
  } catch (error) {
    ElMessage.error('审核失败')
  }
}

const previewFile = (url: string) => {
  window.open(url, '_blank')
}

// Helpers
const getStatusType = (status: string) => {
  switch (status) {
    case 'submitted': return 'warning'
    case 'approved': return 'success'
    case 'rejected': return 'danger'
    case 'supplement': return 'warning'
    default: return 'info'
  }
}

const getStatusName = (status: string) => {
  switch (status) {
    case 'pending': return '待上传'
    case 'submitted': return '待审核'
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
  loadMaterials()
})
</script>

<style scoped>
.visa-material-audit {
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

.material-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.material-card {
  height: fit-content;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.info-row:last-child {
  border-bottom: none;
}

.label {
  color: #666;
}

.value {
  color: #1a1a1a;
}

.file-preview {
  margin: 12px 0;
}

.review-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #eee;
}

.review-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
}
</style>

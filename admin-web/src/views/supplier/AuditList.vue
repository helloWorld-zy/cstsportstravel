<template>
  <div class="supplier-audit-list">
    <div class="page-header">
      <h2>供应商入驻审核</h2>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <el-select v-model="filters.status" placeholder="审核状态" clearable @change="loadApplications">
        <el-option label="全部" value="all" />
        <el-option label="待初审" value="pending_first_review" />
        <el-option label="待复审" value="pending_second_review" />
      </el-select>
      <el-date-picker
        v-model="filters.dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        @change="loadApplications"
      />
      <el-button type="primary" @click="loadApplications">搜索</el-button>
    </div>

    <!-- Table -->
    <el-table :data="applications" v-loading="isLoading" stripe>
      <el-table-column prop="application_no" label="申请编号" width="180" />
      <el-table-column prop="company_name" label="企业名称" min-width="200" />
      <el-table-column prop="unified_social_credit_code" label="信用代码" width="200" />
      <el-table-column prop="contact_name" label="联系人" width="120" />
      <el-table-column prop="contact_phone" label="联系电话" width="140" />
      <el-table-column prop="status" label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">{{ getStatusName(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="applied_at" label="申请时间" width="180">
        <template #default="{ row }">
          {{ formatDateTime(row.applied_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewDetail(row)">查看资料</el-button>
          <el-button
            v-if="row.status === 'pending'"
            size="small"
            type="success"
            @click="handleAudit(row, 'approve')"
          >
            初审通过
          </el-button>
          <el-button
            v-if="row.status === 'reviewing'"
            size="small"
            type="success"
            @click="handleAudit(row, 'approve')"
          >
            复审通过
          </el-button>
          <el-button
            v-if="row.status === 'pending' || row.status === 'reviewing'"
            size="small"
            type="danger"
            @click="handleAudit(row, 'reject')"
          >
            拒绝
          </el-button>
          <el-button
            v-if="row.status === 'pending' || row.status === 'reviewing'"
            size="small"
            type="warning"
            @click="handleAudit(row, 'return_for_revision')"
          >
            退回修改
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
      @current-change="loadApplications"
    />

    <!-- Detail Dialog -->
    <el-dialog v-model="detailVisible" title="供应商申请详情" width="800px">
      <el-descriptions :column="2" border v-if="selectedApp">
        <el-descriptions-item label="申请编号">{{ selectedApp.application_no }}</el-descriptions-item>
        <el-descriptions-item label="企业名称">{{ selectedApp.company_name }}</el-descriptions-item>
        <el-descriptions-item label="统一社会信用代码">{{ selectedApp.unified_social_credit_code }}</el-descriptions-item>
        <el-descriptions-item label="注册地址">{{ selectedApp.registered_address }}</el-descriptions-item>
        <el-descriptions-item label="法人姓名">{{ selectedApp.legal_person_name }}</el-descriptions-item>
        <el-descriptions-item label="经营范围">{{ selectedApp.business_scope }}</el-descriptions-item>
        <el-descriptions-item label="联系人">{{ selectedApp.contact_name }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ selectedApp.contact_phone }}</el-descriptions-item>
      </el-descriptions>

      <h4 style="margin: 16px 0 8px">资质文件</h4>
      <el-table :data="qualifications" size="small">
        <el-table-column prop="qualification_type" label="文件类型" width="150">
          <template #default="{ row }">{{ getQualTypeName(row.qualification_type) }}</template>
        </el-table-column>
        <el-table-column prop="file_name" label="文件名" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'approved' ? 'success' : row.status === 'rejected' ? 'danger' : 'warning'">
              {{ row.status === 'approved' ? '已通过' : row.status === 'rejected' ? '已拒绝' : '待审核' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="预览" width="100">
          <template #default="{ row }">
            <el-button size="small" @click="previewFile(row)">预览</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- Audit Reason Dialog -->
    <el-dialog v-model="reasonDialogVisible" :title="auditActionTitle" width="500px">
      <el-form>
        <el-form-item label="审核原因" :required="auditAction !== 'approve'">
          <el-input v-model="auditReason" type="textarea" :rows="3" placeholder="请输入审核原因" />
        </el-form-item>
        <el-form-item v-if="auditAction === 'return_for_revision'" label="需修改材料">
          <el-checkbox-group v-model="itemsToRevise">
            <el-checkbox label="business_license">营业执照</el-checkbox>
            <el-checkbox label="id_card">法人身份证</el-checkbox>
            <el-checkbox label="travel_license">旅行社许可证</el-checkbox>
            <el-checkbox label="contact_info">联系人信息</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reasonDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="auditing" @click="confirmAudit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const isLoading = ref(false)
const applications = ref<any[]>([])
const detailVisible = ref(false)
const selectedApp = ref<any>(null)
const qualifications = ref<any[]>([])
const reasonDialogVisible = ref(false)
const auditAction = ref('')
const auditReason = ref('')
const itemsToRevise = ref<string[]>([])
const auditing = ref(false)
let auditTargetId = 0

const filters = reactive({
  status: 'all',
  dateRange: null,
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const loadApplications = async () => {
  isLoading.value = true
  try {
    const { data } = await useFetch('/api/v2/admin/suppliers/applications', {
      params: {
        status: filters.status,
        page: pagination.page,
        pageSize: pagination.pageSize,
      },
    })
    if (data.value?.code === 0) {
      applications.value = data.value.data.items || []
      pagination.total = data.value.data.total || 0
    }
  } finally {
    isLoading.value = false
  }
}

const viewDetail = async (row: any) => {
  selectedApp.value = row
  detailVisible.value = true
  const { data } = await useFetch(`/api/v2/admin/suppliers/applications/${row.id}`)
  if (data.value?.code === 0) {
    qualifications.value = data.value.data.qualifications || []
  }
}

const handleAudit = (row: any, action: string) => {
  auditTargetId = row.id
  auditAction.value = action
  auditReason.value = ''
  itemsToRevise.value = []
  reasonDialogVisible.value = true
}

const auditActionTitle = ref('')
watchEffect(() => {
  const titles: Record<string, string> = {
    approve: '审核通过',
    reject: '拒绝申请',
    return_for_revision: '退回修改',
  }
  auditActionTitle.value = titles[auditAction.value] || '审核'
})

const confirmAudit = async () => {
  if (auditAction.value !== 'approve' && !auditReason.value) {
    ElMessage.warning('请输入审核原因')
    return
  }
  auditing.value = true
  try {
    const { data } = await useFetch(`/api/v2/admin/suppliers/applications/${auditTargetId}/audit`, {
      method: 'POST',
      body: {
        action: auditAction.value,
        reason: auditReason.value,
        itemsToRevise: itemsToRevise.value,
      },
    })
    if (data.value?.code === 0) {
      ElMessage.success('操作成功')
      reasonDialogVisible.value = false
      loadApplications()
    } else {
      ElMessage.error(data.value?.message || '操作失败')
    }
  } finally {
    auditing.value = false
  }
}

const previewFile = (row: any) => {
  window.open(row.file_url, '_blank')
}

const getStatusType = (status: string) => {
  const map: Record<string, string> = { pending: 'warning', reviewing: '', active: 'success', terminated: 'info' }
  return map[status] || 'info'
}

const getStatusName = (status: string) => {
  const map: Record<string, string> = { pending: '待初审', reviewing: '待复审', active: '已通过', terminated: '已拒绝' }
  return map[status] || status
}

const getQualTypeName = (type: string) => {
  const map: Record<string, string> = {
    business_license: '营业执照',
    travel_license: '旅行社许可证',
    id_card_front: '法人身份证正面',
    id_card_back: '法人身份证背面',
    other: '其他',
  }
  return map[type] || type
}

const formatDateTime = (s: string) => s ? new Date(s).toLocaleString() : ''

onMounted(loadApplications)
</script>

<style scoped>
.supplier-audit-list { padding: 16px; }
.page-header { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; }
</style>

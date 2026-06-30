<template>
  <div class="supplier-detail" v-loading="isLoading">
    <div class="page-header">
      <el-button @click="$router.back()" size="small">返回列表</el-button>
      <h2>供应商详情</h2>
    </div>

    <template v-if="supplier">
      <!-- Basic Info -->
      <el-card class="section-card">
        <template #header><span>企业基本信息</span></template>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="供应商编号">{{ supplier.supplier_no }}</el-descriptions-item>
          <el-descriptions-item label="企业名称">{{ supplier.company_name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(supplier.status)">{{ getStatusName(supplier.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="统一社会信用代码">{{ supplier.unified_social_credit_code }}</el-descriptions-item>
          <el-descriptions-item label="注册地址" :span="2">{{ supplier.registered_address }}</el-descriptions-item>
          <el-descriptions-item label="注册资本">
            {{ supplier.registered_capital ? (supplier.registered_capital / 10000).toFixed(2) + ' 万元' : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="成立日期">{{ supplier.establishment_date || '-' }}</el-descriptions-item>
          <el-descriptions-item label="经营范围" :span="3">{{ supplier.business_scope }}</el-descriptions-item>
          <el-descriptions-item label="旅行社许可证号">{{ supplier.travel_license_no || '-' }}</el-descriptions-item>
          <el-descriptions-item label="申请编号">{{ supplier.application_no }}</el-descriptions-item>
          <el-descriptions-item label="申请时间">{{ formatDateTime(supplier.applied_at) }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Contact Info -->
      <el-card class="section-card">
        <template #header><span>联系人信息</span></template>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="法人姓名">{{ supplier.legal_person_name }}</el-descriptions-item>
          <el-descriptions-item label="业务联系人">{{ supplier.contact_name }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ supplier.contact_phone }}</el-descriptions-item>
          <el-descriptions-item label="联系邮箱">{{ supplier.contact_email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="财务联系人">{{ supplier.finance_contact_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="财务手机">{{ supplier.finance_contact_phone || '-' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Bank Info -->
      <el-card class="section-card">
        <template #header><span>银行账户信息</span></template>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="开户行">{{ supplier.bank_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="账户名">{{ supplier.bank_account_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="银行账号">{{ maskBankAccount(supplier.bank_account_number) }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Settlement Config -->
      <el-card class="section-card">
        <template #header>
          <span>结算配置</span>
          <el-button size="small" type="primary" style="float:right" @click="showCommissionDialog = true">
            配置佣金规则
          </el-button>
        </template>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="结算周期">
            {{ { daily: '日结', weekly: '周结', monthly: '月结' }[supplier.settlement_cycle] }}
          </el-descriptions-item>
          <el-descriptions-item label="默认佣金比例">
            {{ supplier.commission_rate ? supplier.commission_rate + '%' : '未配置' }}
          </el-descriptions-item>
          <el-descriptions-item label="评分">{{ supplier.rating_score || '-' }}</el-descriptions-item>
          <el-descriptions-item label="审核通过时间">{{ formatDateTime(supplier.approved_at) }}</el-descriptions-item>
          <el-descriptions-item label="合同签署时间">{{ formatDateTime(supplier.contract_signed_at) }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Qualifications -->
      <el-card class="section-card">
        <template #header><span>资质文件</span></template>
        <el-table :data="qualifications" size="small">
          <el-table-column prop="qualification_type" label="文件类型" width="150">
            <template #default="{ row }">{{ getQualTypeName(row.qualification_type) }}</template>
          </el-table-column>
          <el-table-column prop="file_name" label="文件名" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'approved' ? 'success' : 'warning'">
                {{ row.status === 'approved' ? '已通过' : '待审核' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button size="small" @click="previewQualification(row)">预览</el-button>
              <el-button size="small" type="success" @click="approveQualification(row)">通过</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- Status Actions -->
      <el-card class="section-card" v-if="supplier.status === 'active' || supplier.status === 'suspended'">
        <template #header><span>状态管理</span></template>
        <el-space>
          <el-button
            v-if="supplier.status === 'active'"
            type="warning"
            @click="updateStatus('suspended')"
          >
            暂停供应商
          </el-button>
          <el-button
            v-if="supplier.status === 'suspended'"
            type="success"
            @click="updateStatus('active')"
          >
            恢复供应商
          </el-button>
          <el-button
            type="danger"
            @click="updateStatus('terminated')"
          >
            终止合作
          </el-button>
        </el-space>
      </el-card>
    </template>

    <!-- Commission Config Dialog -->
    <el-dialog v-model="showCommissionDialog" title="佣金规则配置" width="500px">
      <el-form label-width="120px">
        <el-form-item label="佣金比例(%)">
          <el-input-number v-model="commissionForm.rate" :min="0.1" :max="50" :precision="2" />
        </el-form-item>
        <el-form-item label="结算周期">
          <el-select v-model="commissionForm.settlementCycle">
            <el-option label="日结" value="daily" />
            <el-option label="周结" value="weekly" />
            <el-option label="月结" value="monthly" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCommissionDialog = false">取消</el-button>
        <el-button type="primary" @click="saveCommission">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const isLoading = ref(false)
const supplier = ref<any>(null)
const qualifications = ref<any[]>([])
const showCommissionDialog = ref(false)

const commissionForm = reactive({
  rate: 15,
  settlementCycle: 'monthly',
})

const loadSupplier = async () => {
  isLoading.value = true
  try {
    const id = route.params.id
    const { data } = await useFetch(`/api/v2/admin/suppliers/${id}`)
    if (data.value?.code === 0) {
      supplier.value = data.value.data.supplier
      qualifications.value = data.value.data.qualifications || []
      if (supplier.value?.commission_rate) commissionForm.rate = supplier.value.commission_rate
      if (supplier.value?.settlement_cycle) commissionForm.settlementCycle = supplier.value.settlement_cycle
    }
  } finally {
    isLoading.value = false
  }
}

const updateStatus = async (newStatus: string) => {
  const names: Record<string, string> = { suspended: '暂停', active: '恢复', terminated: '终止' }
  await ElMessageBox.confirm(`确认${names[newStatus]}供应商？`, '确认操作')
  const { data } = await useFetch(`/api/v2/admin/suppliers/${supplier.value.id}/status`, {
    method: 'PUT',
    body: { status: newStatus },
  })
  if (data.value?.code === 0) {
    ElMessage.success('操作成功')
    loadSupplier()
  }
}

const previewQualification = (row: any) => {
  window.open(row.file_url, '_blank')
}

const approveQualification = async (row: any) => {
  const { data } = await useFetch(`/api/v2/admin/suppliers/qualifications/${row.id}/approve`, { method: 'POST' })
  if (data.value?.code === 0) {
    ElMessage.success('已通过')
    loadSupplier()
  }
}

const saveCommission = async () => {
  const { data } = await useFetch(`/api/v2/admin/suppliers/${supplier.value.id}/commission`, {
    method: 'PUT',
    body: commissionForm,
  })
  if (data.value?.code === 0) {
    ElMessage.success('佣金配置已保存')
    showCommissionDialog.value = false
    loadSupplier()
  }
}

const maskBankAccount = (s: string) => {
  if (!s || s.length < 8) return s || '-'
  return s.slice(0, 4) + '****' + s.slice(-4)
}

const getStatusType = (status: string) => {
  const map: Record<string, string> = { pending: 'warning', reviewing: '', active: 'success', suspended: 'danger', terminated: 'info' }
  return map[status] || 'info'
}

const getStatusName = (status: string) => {
  const map: Record<string, string> = { pending: '待审核', reviewing: '审核中', active: '正常', suspended: '已暂停', terminated: '已终止' }
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

const formatDateTime = (s: string) => s ? new Date(s).toLocaleString() : '-'

onMounted(loadSupplier)
</script>

<style scoped>
.supplier-detail { padding: 16px; }
.page-header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.section-card { margin-bottom: 16px; }
</style>

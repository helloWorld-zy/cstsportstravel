<template>
  <div class="supplier-apply">
    <div class="page-header">
      <h1>供应商入驻申请</h1>
      <p>填写企业信息，提交入驻申请，审核通过后即可开通工作台</p>
    </div>

    <!-- Progress indicator -->
    <el-steps :active="currentStep" finish-status="success" align-center class="steps-bar">
      <el-step title="企业信息" />
      <el-step title="资质上传" />
      <el-step title="联系人信息" />
      <el-step title="银行账户" />
    </el-steps>

    <!-- Step 1: Company Info -->
    <el-card v-if="currentStep === 0" class="step-card">
      <template #header><span>企业基本信息</span></template>
      <el-form :model="form" :rules="rules.company" ref="companyFormRef" label-width="140px">
        <el-form-item label="企业全称" prop="companyName">
          <el-input v-model="form.companyName" placeholder="请输入企业全称" />
        </el-form-item>
        <el-form-item label="统一社会信用代码" prop="creditCode">
          <el-input v-model="form.creditCode" placeholder="18位统一社会信用代码" maxlength="18" />
        </el-form-item>
        <el-form-item label="注册地址" prop="registeredAddress">
          <el-input v-model="form.registeredAddress" placeholder="请输入注册地址" />
        </el-form-item>
        <el-form-item label="注册资本（万元）">
          <el-input-number v-model="form.registeredCapital" :min="0" :precision="2" />
        </el-form-item>
        <el-form-item label="成立日期">
          <el-date-picker v-model="form.establishmentDate" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="经营范围" prop="businessScope">
          <el-input v-model="form.businessScope" type="textarea" :rows="3" placeholder="请输入经营范围" />
        </el-form-item>
        <el-form-item label="旅行社经营许可证号">
          <el-input v-model="form.travelLicenseNo" placeholder="如有请输入" />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Step 2: Qualification Upload -->
    <el-card v-if="currentStep === 1" class="step-card">
      <template #header><span>资质文件上传</span></template>
      <el-form :model="form" label-width="140px">
        <el-form-item label="营业执照扫描件" required>
          <el-upload
            :auto-upload="false"
            :limit="1"
            accept=".pdf,.jpg,.jpeg,.png"
            :on-change="(f) => handleFileChange(f, 'businessLicense')"
            :file-list="files.businessLicense"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip><div class="el-upload__tip">支持 PDF/JPG/PNG，最大 10MB</div></template>
          </el-upload>
        </el-form-item>
        <el-form-item label="法人身份证正面">
          <el-upload
            :auto-upload="false"
            :limit="1"
            accept=".jpg,.jpeg,.png"
            :on-change="(f) => handleFileChange(f, 'idCardFront')"
            :file-list="files.idCardFront"
          >
            <el-button>选择文件</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item label="法人身份证背面">
          <el-upload
            :auto-upload="false"
            :limit="1"
            accept=".jpg,.jpeg,.png"
            :on-change="(f) => handleFileChange(f, 'idCardBack')"
            :file-list="files.idCardBack"
          >
            <el-button>选择文件</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item label="旅行社经营许可证" v-if="form.travelLicenseNo">
          <el-upload
            :auto-upload="false"
            :limit="1"
            accept=".pdf,.jpg,.jpeg,.png"
            :on-change="(f) => handleFileChange(f, 'travelLicense')"
            :file-list="files.travelLicense"
          >
            <el-button>选择文件</el-button>
          </el-upload>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Step 3: Contact Info -->
    <el-card v-if="currentStep === 2" class="step-card">
      <template #header><span>联系人信息</span></template>
      <el-form :model="form" :rules="rules.contact" ref="contactFormRef" label-width="140px">
        <el-form-item label="法人姓名" prop="legalPersonName">
          <el-input v-model="form.legalPersonName" />
        </el-form-item>
        <el-form-item label="法人身份证号" prop="legalPersonIdCard">
          <el-input v-model="form.legalPersonIdCard" maxlength="18" />
        </el-form-item>
        <el-form-item label="业务联系人" prop="contactName">
          <el-input v-model="form.contactName" />
        </el-form-item>
        <el-form-item label="联系人手机号" prop="contactPhone">
          <el-input v-model="form.contactPhone" maxlength="11" />
        </el-form-item>
        <el-form-item label="联系人邮箱">
          <el-input v-model="form.contactEmail" />
        </el-form-item>
        <el-form-item label="财务联系人">
          <el-input v-model="form.financeContactName" />
        </el-form-item>
        <el-form-item label="财务联系人手机">
          <el-input v-model="form.financeContactPhone" maxlength="11" />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Step 4: Bank Account -->
    <el-card v-if="currentStep === 3" class="step-card">
      <template #header><span>银行账户信息</span></template>
      <el-form :model="form" label-width="140px">
        <el-form-item label="开户行">
          <el-input v-model="form.bankName" placeholder="例：中国工商银行北京分行" />
        </el-form-item>
        <el-form-item label="账户名">
          <el-input v-model="form.bankAccountName" />
        </el-form-item>
        <el-form-item label="银行账号">
          <el-input v-model="form.bankAccountNumber" />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Navigation buttons -->
    <div class="step-actions">
      <el-button v-if="currentStep > 0" @click="prevStep">上一步</el-button>
      <el-button v-if="currentStep < 3" type="primary" @click="nextStep">下一步</el-button>
      <el-button v-if="currentStep === 3" type="success" :loading="submitting" @click="submitApplication">
        提交申请
      </el-button>
      <el-button @click="saveDraft" size="small">保存草稿</el-button>
    </div>

    <!-- Application status (after submission) -->
    <el-card v-if="applicationNo" class="status-card">
      <template #header><span>申请状态</span></template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="申请编号">{{ applicationNo }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(applicationStatus)">{{ getStatusName(applicationStatus) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ applicationTime }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'

const currentStep = ref(0)
const submitting = ref(false)
const applicationNo = ref('')
const applicationStatus = ref('')
const applicationTime = ref('')

const form = reactive({
  companyName: '',
  creditCode: '',
  registeredAddress: '',
  registeredCapital: 0,
  establishmentDate: '',
  businessScope: '',
  travelLicenseNo: '',
  legalPersonName: '',
  legalPersonIdCard: '',
  contactName: '',
  contactPhone: '',
  contactEmail: '',
  financeContactName: '',
  financeContactPhone: '',
  bankName: '',
  bankAccountName: '',
  bankAccountNumber: '',
})

const files = reactive<Record<string, any[]>>({
  businessLicense: [],
  idCardFront: [],
  idCardBack: [],
  travelLicense: [],
})

const rules = {
  company: {
    companyName: [{ required: true, message: '请输入企业全称', trigger: 'blur' }],
    creditCode: [
      { required: true, message: '请输入统一社会信用代码', trigger: 'blur' },
      { len: 18, message: '统一社会信用代码为18位', trigger: 'blur' },
    ],
    businessScope: [{ required: true, message: '请输入经营范围', trigger: 'blur' }],
  },
  contact: {
    legalPersonName: [{ required: true, message: '请输入法人姓名', trigger: 'blur' }],
    legalPersonIdCard: [{ required: true, message: '请输入法人身份证号', trigger: 'blur' }],
    contactName: [{ required: true, message: '请输入业务联系人', trigger: 'blur' }],
    contactPhone: [{ required: true, message: '请输入联系人手机号', trigger: 'blur' }],
  },
}

const handleFileChange = (file: any, key: string) => {
  files[key] = [file]
}

const nextStep = () => {
  if (currentStep.value < 3) currentStep.value++
}

const prevStep = () => {
  if (currentStep.value > 0) currentStep.value--
}

const saveDraft = () => {
  localStorage.setItem('supplier_apply_draft', JSON.stringify(form))
  ElMessage.success('草稿已保存')
}

const submitApplication = async () => {
  submitting.value = true
  try {
    const formData = new FormData()
    Object.entries(form).forEach(([key, value]) => {
      if (value) formData.append(key, String(value))
    })
    if (files.businessLicense[0]) formData.append('businessLicense', files.businessLicense[0].raw)
    if (files.idCardFront[0]) formData.append('legalPersonIdCardFront', files.idCardFront[0].raw)
    if (files.idCardBack[0]) formData.append('legalPersonIdCardBack', files.idCardBack[0].raw)
    if (files.travelLicense[0]) formData.append('travelLicense', files.travelLicense[0].raw)

    const { data } = await useFetch('/api/v2/suppliers/apply', {
      method: 'POST',
      body: formData,
    })

    if (data.value?.code === 0) {
      applicationNo.value = data.value.data.applicationNo
      applicationStatus.value = data.value.data.status
      applicationTime.value = new Date().toLocaleString()
      ElMessage.success('申请提交成功')
      localStorage.removeItem('supplier_apply_draft')
    } else {
      ElMessage.error(data.value?.message || '提交失败')
    }
  } catch (e) {
    ElMessage.error('网络错误，请重试')
  } finally {
    submitting.value = false
  }
}

const getStatusType = (status: string) => {
  const map: Record<string, string> = {
    pending: 'warning',
    reviewing: '',
    active: 'success',
    suspended: 'danger',
    terminated: 'info',
  }
  return map[status] || 'info'
}

const getStatusName = (status: string) => {
  const map: Record<string, string> = {
    pending: '待初审',
    reviewing: '待复审',
    active: '已通过',
    suspended: '已暂停',
    terminated: '已终止',
  }
  return map[status] || status
}
</script>

<style scoped>
.supplier-apply {
  max-width: 800px;
  margin: 0 auto;
  padding: 24px;
}
.page-header {
  text-align: center;
  margin-bottom: 24px;
}
.steps-bar {
  margin-bottom: 24px;
}
.step-card {
  margin-bottom: 16px;
}
.step-actions {
  text-align: center;
  margin: 24px 0;
}
.status-card {
  margin-top: 24px;
}
</style>

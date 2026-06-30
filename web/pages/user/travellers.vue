<template>
  <div class="travellers-page">
    <div class="page-header">
      <h1>常用出游人</h1>
      <el-button type="primary" @click="showAddDialog" :disabled="travellers.length >= 20">
        添加出游人
      </el-button>
    </div>

    <p class="hint">最多可添加20位常用出游人，当前 {{ travellers.length }}/20</p>

    <!-- Traveller List -->
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>

    <div v-else-if="travellers.length === 0" class="empty-state">
      <el-empty description="暂无常用出游人">
        <el-button type="primary" @click="showAddDialog">添加第一位出游人</el-button>
      </el-empty>
    </div>

    <div v-else class="traveller-list">
      <div v-for="t in travellers" :key="t.id" class="traveller-card">
        <div class="traveller-info">
          <div class="traveller-name">
            {{ t.real_name }}
            <el-tag v-if="t.is_default" type="success" size="small">默认</el-tag>
          </div>
          <div class="traveller-detail">身份证：{{ t.id_card_no }}</div>
          <div v-if="t.phone" class="traveller-detail">手机：{{ t.phone }}</div>
          <div v-if="t.birth_date" class="traveller-detail">出生日期：{{ t.birth_date }}</div>
        </div>
        <div class="traveller-actions">
          <el-button text type="primary" @click="showEditDialog(t)">编辑</el-button>
          <el-button text type="danger" @click="handleDelete(t.id)">删除</el-button>
        </div>
      </div>
    </div>

    <!-- Add/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑出游人' : '添加出游人'"
      width="500px"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item label="真实姓名" prop="real_name">
          <el-input v-model="form.real_name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="身份证号" prop="id_card_no">
          <el-input v-model="form.id_card_no" placeholder="请输入18位身份证号" maxlength="18" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号（选填）" maxlength="11" />
        </el-form-item>
        <el-form-item label="出生日期" prop="birth_date">
          <el-date-picker v-model="form.birth_date" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="性别" prop="gender">
          <el-radio-group v-model="form.gender">
            <el-radio value="male">男</el-radio>
            <el-radio value="female">女</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="form.is_default">设为默认出游人</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ isEditing ? '保存' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useApi } from '~/composables/useApi'

definePageMeta({
  layout: 'user',
})

interface Traveller {
  id: number
  real_name: string
  id_card_no: string
  phone: string
  birth_date: string
  gender: string
  is_default: boolean
  created_at: string
}

const { user, init } = useAuth()
const api = useApi()

const travellers = ref<Traveller[]>([])
const loading = ref(true)
const dialogVisible = ref(false)
const submitting = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)

const formRef = ref<FormInstance>()
const form = reactive({
  real_name: '',
  id_card_no: '',
  phone: '',
  birth_date: '',
  gender: '',
  is_default: false,
})

function validateIDCard(rule: any, value: string, callback: any) {
  if (!value) {
    callback(new Error('请输入身份证号'))
    return
  }
  if (value.length !== 18) {
    callback(new Error('身份证号应为18位'))
    return
  }
  const weights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2]
  const checkCodes = ['1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2']
  let sum = 0
  for (let i = 0; i < 17; i++) {
    sum += parseInt(value[i]) * weights[i]
  }
  const expected = checkCodes[sum % 11]
  if (value[17].toUpperCase() !== expected) {
    callback(new Error('身份证号校验码不正确'))
    return
  }
  callback()
}

const rules: FormRules = {
  real_name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
  ],
  id_card_no: [
    { required: true, validator: validateIDCard, trigger: 'blur' },
  ],
}

onMounted(async () => {
  await init()
  if (!user.value) {
    navigateTo('/auth/login')
    return
  }
  await loadTravellers()
})

async function loadTravellers() {
  loading.value = true
  try {
    travellers.value = await api.get<Traveller[]>('/users/me/travellers')
  } catch (err: any) {
    ElMessage.error(err.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function showAddDialog() {
  isEditing.value = false
  editingId.value = null
  form.real_name = ''
  form.id_card_no = ''
  form.phone = ''
  form.birth_date = ''
  form.gender = ''
  form.is_default = false
  dialogVisible.value = true
}

function showEditDialog(t: Traveller) {
  isEditing.value = true
  editingId.value = t.id
  form.real_name = '' // Can't decrypt masked name, user needs to re-enter
  form.id_card_no = '' // Same
  form.phone = t.phone
  form.birth_date = t.birth_date
  form.gender = t.gender
  form.is_default = t.is_default
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitting.value = true
  try {
    if (isEditing.value && editingId.value) {
      await api.put(`/users/me/travellers/${editingId.value}`, { ...form })
      ElMessage.success('更新成功')
    } else {
      await api.post('/users/me/travellers', { ...form })
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    await loadTravellers()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除此出游人吗？', '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await api.del(`/users/me/travellers/${id}`)
    ElMessage.success('删除成功')
    await loadTravellers()
  } catch (err: any) {
    if (err !== 'cancel') {
      ElMessage.error(err.message || '删除失败')
    }
  }
}
</script>

<style scoped>
.travellers-page {
  max-width: 100%;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-sm);
}
.page-header h1 {
  margin: 0;
}
.hint {
  color: var(--color-text-secondary);
  font-size: 13px;
  margin-bottom: var(--space-lg);
}
.traveller-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}
.traveller-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md);
  background: var(--color-bg-white);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
}
.traveller-name {
  font-weight: 500;
  margin-bottom: var(--space-xs);
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}
.traveller-detail {
  font-size: 13px;
  color: var(--color-text-secondary);
}
.traveller-actions {
  display: flex;
  gap: var(--space-xs);
}
.loading-container {
  padding: var(--space-lg);
  background: var(--color-bg-white);
  border-radius: var(--radius-md);
}
.empty-state {
  padding: var(--space-xl);
  background: var(--color-bg-white);
  border-radius: var(--radius-md);
}
</style>

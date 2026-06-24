<template>
  <div class="real-name-form">
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <el-form-item label="真实姓名" prop="real_name">
        <el-input v-model="form.real_name" placeholder="请输入身份证上的姓名" maxlength="20" />
      </el-form-item>

      <el-form-item label="身份证号" prop="id_card_no">
        <el-input v-model="form.id_card_no" placeholder="请输入18位身份证号" maxlength="18" />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="loading" @click="handleSubmit" style="width: 100%">
          提交认证
        </el-button>
      </el-form-item>
    </el-form>

    <div class="tips">
      <p>提示：</p>
      <ul>
        <li>请确保姓名和身份证号与身份证上一致</li>
        <li>每天最多可提交3次认证</li>
        <li>实名认证通过后可预订境内游产品</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useApi } from '~/composables/useApi'

const emit = defineEmits<{
  success: [status: string]
}>()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  real_name: '',
  id_card_no: '',
})

// ID card validation (ISO 7064:1983.MOD 11-2)
function validateIDCard(rule: any, value: string, callback: any) {
  if (!value) {
    callback(new Error('请输入身份证号'))
    return
  }
  if (value.length !== 18) {
    callback(new Error('身份证号应为18位'))
    return
  }

  // Check first 17 digits are numeric
  for (let i = 0; i < 17; i++) {
    if (value[i] < '0' || value[i] > '9') {
      callback(new Error('身份证号格式不正确'))
      return
    }
  }

  // Check last digit
  const last = value[17]
  if (last !== 'X' && last !== 'x' && (last < '0' || last > '9')) {
    callback(new Error('身份证号格式不正确'))
    return
  }

  // Verify checksum
  const weights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2]
  const checkCodes = ['1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2']

  let sum = 0
  for (let i = 0; i < 17; i++) {
    sum += parseInt(value[i]) * weights[i]
  }
  const expected = checkCodes[sum % 11]

  if (last.toUpperCase() !== expected) {
    callback(new Error('身份证号校验码不正确'))
    return
  }

  callback()
}

const rules: FormRules = {
  real_name: [
    { required: true, message: '请输入真实姓名', trigger: 'blur' },
    { min: 2, max: 20, message: '姓名长度为2-20个字符', trigger: 'blur' },
  ],
  id_card_no: [
    { required: true, validator: validateIDCard, trigger: 'blur' },
  ],
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    const api = useApi()
    const data = await api.post<{ status: string }>('/users/me/real-name', {
      real_name: form.real_name,
      id_card_no: form.id_card_no,
    })
    ElMessage.success('实名认证提交成功')
    emit('success', data.status)
  } catch (err: any) {
    ElMessage.error(err.message || '认证提交失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.real-name-form {
  max-width: 400px;
}
.tips {
  margin-top: var(--space-lg);
  padding: var(--space-md);
  background: var(--color-bg-base);
  border-radius: var(--radius-md);
  font-size: 13px;
  color: var(--color-text-secondary);
}
.tips p {
  margin: 0 0 var(--space-xs);
  font-weight: 500;
}
.tips ul {
  margin: 0;
  padding-left: var(--space-md);
}
.tips li {
  margin-bottom: var(--space-xs);
}
</style>

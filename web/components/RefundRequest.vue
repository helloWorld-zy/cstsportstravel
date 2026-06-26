<template>
  <el-dialog
    v-model="visible"
    title="申请退款"
    width="500px"
    :close-on-click-modal="false"
    @close="$emit('close')"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      label-position="top"
    >
      <el-form-item label="退款原因" prop="reason">
        <el-select v-model="form.reason" placeholder="请选择退款原因" style="width: 100%">
          <el-option label="行程变更" value="行程变更" />
          <el-option label="健康原因" value="健康原因" />
          <el-option label="不可抗力" value="不可抗力" />
          <el-option label="其他" value="其他" />
        </el-select>
      </el-form-item>

      <el-form-item label="详细说明" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="3"
          placeholder="请描述退款原因（可选）"
          maxlength="500"
          show-word-limit
        />
      </el-form-item>

      <!-- Refund Calculation Preview -->
      <div v-if="calculation" class="refund-preview">
        <el-divider content-position="left">退款金额预览</el-divider>

        <div class="calc-item">
          <span class="calc-label">订单实付金额</span>
          <span class="calc-value">¥{{ formatAmount(calculation.paid_amount) }}</span>
        </div>

        <div v-if="calculation.matching_rule" class="calc-item">
          <span class="calc-label">匹配退改规则</span>
          <span class="calc-value rule">{{ calculation.matching_rule.rule_name }}</span>
        </div>

        <div class="calc-item">
          <span class="calc-label">退款比例</span>
          <span class="calc-value">{{ calculation.refund_percentage }}%</span>
        </div>

        <div v-if="calculation.cancellation_fee > 0" class="calc-item">
          <span class="calc-label">退改手续费</span>
          <span class="calc-value fee">-¥{{ formatAmount(calculation.cancellation_fee) }}</span>
        </div>

        <el-divider />

        <div class="calc-item total">
          <span class="calc-label">预计退款金额</span>
          <span class="calc-value refund-amount">¥{{ formatAmount(calculation.refund_amount) }}</span>
        </div>

        <p class="refund-note">
          * 实际退款金额以审核结果为准，退款将原路退回您的支付账户
        </p>
      </div>
    </el-form>

    <template #footer>
      <el-button @click="$emit('close')">取消</el-button>
      <el-button
        type="primary"
        :loading="submitting"
        @click="submitRefund"
      >
        提交退款申请
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'

interface Props {
  orderId: number
}

interface Calculation {
  paid_amount: number
  refund_percentage: number
  cancellation_fee: number
  refund_amount: number
  matching_rule: {
    rule_id: number
    rule_name: string
    days_before_min: number
    refund_percentage: number
    description: string
  } | null
  days_before: number
}

const props = defineProps<Props>()
const emit = defineEmits(['close', 'success'])

const visible = ref(true)
const formRef = ref<FormInstance>()
const submitting = ref(false)
const calculation = ref<Calculation | null>(null)

const form = reactive({
  reason: '',
  description: '',
})

const rules: FormRules = {
  reason: [{ required: true, message: '请选择退款原因', trigger: 'change' }],
}

function formatAmount(cents: number): string {
  return (cents / 100).toFixed(2)
}

async function fetchRefundPreview() {
  try {
    const response = await $fetch<{ data: any }>(
      `/api/v1/orders/${props.orderId}/refund-status`
    )
    // If there's already a refund, close
    if (response.data) {
      ElMessage.warning('该订单已有退款申请')
      emit('close')
      return
    }
  } catch {
    // No existing refund, which is expected
  }
}

async function submitRefund() {
  if (!formRef.value) return

  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  try {
    await ElMessageBox.confirm(
      '确定要提交退款申请吗？提交后将进入审核流程。',
      '确认退款',
      {
        confirmButtonText: '确定提交',
        cancelButtonText: '再想想',
        type: 'warning',
      }
    )
  } catch {
    return // user cancelled
  }

  submitting.value = true
  try {
    const response = await $fetch<{ data: any }>(
      `/api/v1/orders/${props.orderId}/refund`,
      {
        method: 'POST',
        body: {
          reason: form.reason,
          description: form.description,
        },
      }
    )

    calculation.value = response.data?.calculation || null

    ElMessage.success('退款申请已提交，请等待审核')
    emit('success')
  } catch (error: any) {
    const msg = error?.data?.message || '退款申请提交失败'
    ElMessage.error(msg)
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchRefundPreview()
})
</script>

<style scoped>
.refund-preview {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 16px;
  margin-top: 8px;
}

.calc-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
}

.calc-label {
  font-size: 14px;
  color: #606266;
}

.calc-value {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.calc-value.rule {
  color: #409eff;
}

.calc-value.fee {
  color: #f56c6c;
}

.calc-item.total {
  padding: 8px 0;
}

.calc-item.total .calc-label {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.refund-amount {
  font-size: 20px;
  font-weight: 700;
  color: #f56c6c;
}

.refund-note {
  font-size: 12px;
  color: #909399;
  margin: 8px 0 0;
  line-height: 1.5;
}
</style>

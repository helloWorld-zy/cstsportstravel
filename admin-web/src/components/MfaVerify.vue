<template>
  <el-dialog
    v-model="visible"
    title="安全验证"
    width="400px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
    @close="handleClose"
  >
    <div class="mfa-verify">
      <el-alert type="warning" :closable="false" show-icon>
        <template #title>
          此操作需要多因素认证验证
        </template>
      </el-alert>

      <div class="verify-content">
        <p class="verify-hint">请输入认证器应用中的6位动态验证码：</p>

        <div class="code-input-wrapper">
          <el-input
            ref="codeInputRef"
            v-model="totpCode"
            placeholder="000000"
            maxlength="6"
            size="large"
            class="code-input"
            :class="{ 'is-error': !!errorMessage }"
            @keyup.enter="handleVerify"
            @input="clearError"
          />
        </div>

        <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>

        <div class="verify-actions">
          <el-button @click="handleCancel">取消</el-button>
          <el-button type="primary" :loading="verifying" @click="handleVerify">
            确认验证
          </el-button>
        </div>

        <div class="verify-footer">
          <p class="footer-hint">
            没有认证器？
            <el-button type="primary" link @click="showSmsFallback = true">
              使用短信验证码
            </el-button>
          </p>
        </div>

        <!-- SMS Fallback -->
        <div v-if="showSmsFallback" class="sms-fallback">
          <el-divider>短信验证码</el-divider>
          <p class="sms-hint">短信验证码功能开发中...</p>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { mfaVerify } from '@/api/rbac'

const props = defineProps<{
  /** Controls dialog visibility. */
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  /** Emitted when verification succeeds. */
  (e: 'verified'): void
  /** Emitted when user cancels. */
  (e: 'cancel'): void
}>()

const visible = ref(props.modelValue)
const totpCode = ref('')
const errorMessage = ref('')
const verifying = ref(false)
const showSmsFallback = ref(false)
const codeInputRef = ref()

// Sync visibility with prop
watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    totpCode.value = ''
    errorMessage.value = ''
    showSmsFallback.value = false
    nextTick(() => codeInputRef.value?.focus())
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

function clearError() {
  errorMessage.value = ''
}

async function handleVerify() {
  if (totpCode.value.length !== 6) {
    errorMessage.value = '请输入6位数字验证码'
    return
  }

  verifying.value = true
  errorMessage.value = ''

  try {
    const res = await mfaVerify(totpCode.value)
    if (res.verified) {
      visible.value = false
      emit('verified')
    } else {
      errorMessage.value = '验证码错误，请重试'
      totpCode.value = ''
    }
  } catch (err: any) {
    errorMessage.value = err.message || '验证失败，请重试'
    totpCode.value = ''
  } finally {
    verifying.value = false
  }
}

function handleCancel() {
  visible.value = false
  emit('cancel')
}

function handleClose() {
  emit('cancel')
}
</script>

<style scoped>
.mfa-verify {
  text-align: center;
}

.verify-content {
  margin-top: 16px;
}

.verify-hint {
  font-size: 14px;
  color: #606266;
  margin-bottom: 16px;
}

.code-input-wrapper {
  display: flex;
  justify-content: center;
  margin: 16px 0;
}

.code-input {
  width: 200px;
}

.code-input :deep(.el-input__inner) {
  font-size: 24px;
  text-align: center;
  letter-spacing: 8px;
}

.code-input.is-error :deep(.el-input__inner) {
  border-color: #f56c6c;
}

.error-text {
  color: #f56c6c;
  font-size: 13px;
  margin: 8px 0;
}

.verify-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 16px;
}

.verify-footer {
  margin-top: 16px;
}

.footer-hint {
  font-size: 13px;
  color: #909399;
}

.sms-fallback {
  margin-top: 8px;
}

.sms-hint {
  font-size: 13px;
  color: #909399;
}
</style>

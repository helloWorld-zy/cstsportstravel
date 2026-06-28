<template>
  <div class="mfa-setup">
    <el-steps :active="currentStep" finish-status="success" align-center>
      <el-step title="扫描二维码" description="使用认证器扫描" />
      <el-step title="输入验证码" description="输入6位动态码" />
      <el-step title="完成" description="MFA已启用" />
    </el-steps>

    <!-- Step 1: Show QR Code -->
    <div v-if="currentStep === 0" class="step-content">
      <div class="qr-section">
        <p class="instructions">
          请使用 Google Authenticator、Microsoft Authenticator 或其他 TOTP 认证器扫描以下二维码：
        </p>
        <div class="qr-code-wrapper">
          <img v-if="qrCodeUrl" :src="qrCodeImageUrl" alt="MFA QR Code" class="qr-code" />
          <div v-else class="qr-placeholder">
            <el-icon :size="48"><Loading /></el-icon>
            <p>正在生成二维码...</p>
          </div>
        </div>
        <div class="manual-entry" v-if="secret">
          <p>无法扫描？手动输入密钥：</p>
          <el-input :model-value="secret" readonly size="small">
            <template #append>
              <el-button @click="copySecret">复制</el-button>
            </template>
          </el-input>
        </div>
      </div>
      <div class="step-actions">
        <el-button type="primary" @click="currentStep = 1">下一步</el-button>
      </div>
    </div>

    <!-- Step 2: Verify Code -->
    <div v-if="currentStep === 1" class="step-content">
      <div class="verify-section">
        <p class="instructions">请输入认证器上显示的6位动态验证码：</p>
        <div class="code-input-wrapper">
          <el-input
            v-model="totpCode"
            placeholder="000000"
            maxlength="6"
            size="large"
            class="code-input"
            @keyup.enter="handleVerify"
          />
        </div>
        <p v-if="verifyError" class="error-text">{{ verifyError }}</p>
      </div>
      <div class="step-actions">
        <el-button @click="currentStep = 0">上一步</el-button>
        <el-button type="primary" :loading="verifying" @click="handleVerify">验证</el-button>
      </div>
    </div>

    <!-- Step 3: Complete -->
    <div v-if="currentStep === 2" class="step-content">
      <el-result icon="success" title="MFA 已启用" sub-title="多因素认证已成功启用，敏感操作将需要动态验证码确认。">
        <template #extra>
          <el-button type="primary" @click="emit('complete')">完成</el-button>
        </template>
      </el-result>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { mfaSetup, mfaVerify } from '@/api/rbac'

const emit = defineEmits<{
  (e: 'complete'): void
}>()

const currentStep = ref(0)
const secret = ref('')
const qrCodeUrl = ref('')
const totpCode = ref('')
const verifyError = ref('')
const verifying = ref(false)

// Generate QR code image URL from otpauth:// URL
const qrCodeImageUrl = computed(() => {
  if (!qrCodeUrl.value) return ''
  // Use a QR code generation API (in production, use a local library like qrcode)
  return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(qrCodeUrl.value)}`
})

// Initialize MFA setup
async function initSetup() {
  try {
    const res = await mfaSetup()
    // In a real implementation, the response would include the secret and QR URL
    // For now, we use placeholder values from the API response
    if (res && typeof res === 'object') {
      const data = res as any
      secret.value = data.secret || ''
      qrCodeUrl.value = data.qr_code_url || ''
    }
  } catch (err: any) {
    ElMessage.error(err.message || 'MFA 初始化失败')
  }
}

// Verify TOTP code
async function handleVerify() {
  if (totpCode.value.length !== 6) {
    verifyError.value = '请输入6位数字验证码'
    return
  }

  verifying.value = true
  verifyError.value = ''

  try {
    const res = await mfaVerify(totpCode.value)
    if (res.verified) {
      currentStep.value = 2
      ElMessage.success('MFA 验证成功')
    } else {
      verifyError.value = '验证码错误，请重试'
    }
  } catch (err: any) {
    verifyError.value = err.message || '验证失败，请重试'
  } finally {
    verifying.value = false
  }
}

// Copy secret to clipboard
function copySecret() {
  navigator.clipboard.writeText(secret.value).then(() => {
    ElMessage.success('密钥已复制到剪贴板')
  })
}

// Initialize on mount
initSetup()
</script>

<style scoped>
.mfa-setup {
  max-width: 500px;
  margin: 0 auto;
  padding: 20px;
}

.step-content {
  margin-top: 24px;
  text-align: center;
}

.instructions {
  font-size: 14px;
  color: #606266;
  margin-bottom: 16px;
}

.qr-code-wrapper {
  display: flex;
  justify-content: center;
  margin: 16px 0;
}

.qr-code {
  width: 200px;
  height: 200px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
}

.qr-placeholder {
  width: 200px;
  height: 200px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  border-radius: 4px;
  color: #909399;
}

.manual-entry {
  margin-top: 16px;
  text-align: left;
}

.manual-entry p {
  font-size: 13px;
  color: #909399;
  margin-bottom: 8px;
}

.code-input-wrapper {
  display: flex;
  justify-content: center;
  margin: 16px 0;
}

.code-input {
  width: 200px;
  font-size: 24px;
  text-align: center;
  letter-spacing: 8px;
}

.error-text {
  color: #f56c6c;
  font-size: 13px;
  margin-top: 8px;
}

.step-actions {
  margin-top: 24px;
  display: flex;
  justify-content: center;
  gap: 12px;
}
</style>

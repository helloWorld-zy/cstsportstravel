<template>
  <view class="login-page">
    <view class="logo-section">
      <text class="title">欢迎使用</text>
      <text class="subtitle">旅游预订平台</text>
    </view>

    <!-- 抖音一键登录 -->
    <view class="login-section">
      <button
        class="douyin-btn"
        :loading="douyinLoading"
        @click="handleDouyinLogin"
      >
        抖音一键登录
      </button>
    </view>

    <view class="divider">
      <view class="line"></view>
      <text class="divider-text">或</text>
      <view class="line"></view>
    </view>

    <!-- 手机号登录表单 -->
    <view class="form-section">
      <view class="form-item">
        <input
          v-model="phone"
          type="number"
          placeholder="请输入手机号"
          maxlength="11"
          class="input"
        />
      </view>

      <view class="form-item code-row">
        <input
          v-model="code"
          type="number"
          placeholder="请输入验证码"
          maxlength="6"
          class="input code-input"
        />
        <button
          class="code-btn"
          :disabled="countdown > 0 || sendingCode"
          @click="handleSendCode"
        >
          {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
        </button>
      </view>

      <button
        class="login-btn"
        :loading="loginLoading"
        @click="handlePhoneLogin"
      >
        登录 / 注册
      </button>
    </view>

    <!-- 手机号绑定区域（抖音首次登录用户） -->
    <view v-if="showBindPhone" class="bind-section">
      <view class="bind-header">
        <text class="bind-title">绑定手机号</text>
        <text class="bind-desc">首次使用抖音登录，请绑定手机号以同步订单数据</text>
      </view>

      <view class="form-item">
        <input
          v-model="bindPhone"
          type="number"
          placeholder="请输入手机号"
          maxlength="11"
          class="input"
        />
      </view>

      <view class="form-item code-row">
        <input
          v-model="bindCode"
          type="number"
          placeholder="请输入验证码"
          maxlength="6"
          class="input code-input"
        />
        <button
          class="code-btn"
          :disabled="bindCountdown > 0"
          @click="handleSendBindCode"
        >
          {{ bindCountdown > 0 ? `${bindCountdown}s` : '获取验证码' }}
        </button>
      </view>

      <button
        class="login-btn"
        :loading="bindLoading"
        @click="handleBindPhone"
      >
        绑定并登录
      </button>
    </view>

    <!-- 用户协议 -->
    <view class="agreement">
      <text class="agreement-text">
        登录即表示同意
        <text class="link" @click="openAgreement('user')">《用户协议》</text>
        和
        <text class="link" @click="openAgreement('privacy')">《隐私政策》</text>
      </text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { api, setTokens } from '../shared/api/request'
import { platformLogin, getUserProfile, PlatformType } from '../utils/platform'

interface DouyinLoginResponse {
  user?: {
    id: number
    phone: string
    nickname: string
    avatar_url: string
  }
  access_token?: string
  refresh_token?: string
  need_bindphone: boolean
  is_new_user: boolean
}

const phone = ref('')
const code = ref('')
const sendingCode = ref(false)
const loginLoading = ref(false)
const douyinLoading = ref(false)
const countdown = ref(0)

// 手机号绑定状态
const showBindPhone = ref(false)
const bindPhone = ref('')
const bindCode = ref('')
const bindLoading = ref(false)
const bindCountdown = ref(0)

// 抖音登录临时 code（用于绑定手机号时再次提交）
let douyinCode = ''

let timer: ReturnType<typeof setInterval> | null = null
let bindTimer: ReturnType<typeof setInterval> | null = null

/**
 * 发送短信验证码
 */
async function handleSendCode() {
  if (!phone.value || phone.value.length !== 11) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return
  }

  sendingCode.value = true
  try {
    await api.post('/auth/sms-code', { phone: phone.value })
    uni.showToast({ title: '验证码已发送', icon: 'success' })
    countdown.value = 60
    timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0 && timer) {
        clearInterval(timer)
        timer = null
      }
    }, 1000)
  } catch (err: any) {
    uni.showToast({ title: err.message || '发送失败', icon: 'none' })
  } finally {
    sendingCode.value = false
  }
}

/**
 * 手机号+验证码登录
 */
async function handlePhoneLogin() {
  if (!phone.value || phone.value.length !== 11) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return
  }
  if (!code.value || code.value.length !== 6) {
    uni.showToast({ title: '请输入6位验证码', icon: 'none' })
    return
  }

  loginLoading.value = true
  try {
    const data = await api.post<DouyinLoginResponse>('/auth/login', {
      phone: phone.value,
      code: code.value,
      platform: 'douyin',
    })
    setTokens(data.access_token!, data.refresh_token!)
    uni.showToast({
      title: data.is_new_user ? '注册成功' : '登录成功',
      icon: 'success',
    })
    setTimeout(() => {
      uni.switchTab({ url: '/pages/index/index' })
    }, 1000)
  } catch (err: any) {
    uni.showToast({ title: err.message || '登录失败', icon: 'none' })
  } finally {
    loginLoading.value = false
  }
}

/**
 * 抖音一键登录流程：
 * 1. 调用 tt.login() 获取 code
 * 2. 将 code 发送到后端换取 OpenID
 * 3. 后端判断是否为新用户
 * 4. 新用户需要绑定手机号（同一手机号关联同一平台账号）
 * 5. 老用户直接登录
 *
 * FR-183: 抖音登录获取 OpenID，首次登录自动创建平台账号并绑定
 * 关键约束: 同一手机号在微信和抖音登录应关联到同一平台账号
 */
async function handleDouyinLogin() {
  douyinLoading.value = true
  try {
    // Step 1: 调用抖音登录 API 获取 code
    const loginResult = await platformLogin()

    if (loginResult.platform !== PlatformType.TOUTIAO) {
      throw new Error('当前环境非抖音小程序')
    }

    douyinCode = loginResult.code

    // Step 2: 尝试获取用户信息（可选，用于完善资料）
    let userInfo = null
    try {
      userInfo = await getUserProfile()
    } catch {
      // 用户拒绝授权，不影响登录流程
      console.log('[DouyinLogin] User denied profile access, continue with basic login')
    }

    // Step 3: 将 code 发送到后端
    const data = await api.post<DouyinLoginResponse>('/auth/douyin', {
      code: loginResult.code,
      anonymous_code: loginResult.anonymousCode,
      nickname: userInfo?.nickName,
      avatar_url: userInfo?.avatarUrl,
    })

    if (data.need_bindphone) {
      // Step 4: 新用户需要绑定手机号
      showBindPhone.value = true
      uni.showToast({ title: '请绑定手机号', icon: 'none' })
    } else {
      // Step 5: 老用户直接登录
      setTokens(data.access_token!, data.refresh_token!)
      uni.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => {
        uni.switchTab({ url: '/pages/index/index' })
      }, 1000)
    }
  } catch (err: any) {
    uni.showToast({ title: err.message || '抖音登录失败', icon: 'none' })
  } finally {
    douyinLoading.value = false
  }
}

/**
 * 发送绑定手机验证码
 */
async function handleSendBindCode() {
  if (!bindPhone.value || bindPhone.value.length !== 11) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return
  }
  try {
    await api.post('/auth/sms-code', { phone: bindPhone.value })
    uni.showToast({ title: '验证码已发送', icon: 'success' })
    bindCountdown.value = 60
    bindTimer = setInterval(() => {
      bindCountdown.value--
      if (bindCountdown.value <= 0 && bindTimer) {
        clearInterval(bindTimer)
        bindTimer = null
      }
    }, 1000)
  } catch (err: any) {
    uni.showToast({ title: err.message || '发送失败', icon: 'none' })
  }
}

/**
 * 绑定手机号并完成登录
 *
 * 关键逻辑：同一手机号在微信和抖音登录应关联到同一平台账号
 * 后端通过手机号查找已有账号，若存在则绑定抖音 OpenID 到该账号
 * 若不存在则创建新账号并绑定抖音 OpenID
 */
async function handleBindPhone() {
  if (!bindPhone.value || bindPhone.value.length !== 11) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return
  }
  if (!bindCode.value || bindCode.value.length !== 6) {
    uni.showToast({ title: '请输入6位验证码', icon: 'none' })
    return
  }

  bindLoading.value = true
  try {
    const data = await api.post<DouyinLoginResponse>('/auth/douyin/bindphone', {
      code: douyinCode,
      phone: bindPhone.value,
      sms_code: bindCode.value,
    })

    if (data.access_token) {
      setTokens(data.access_token, data.refresh_token!)
      uni.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => {
        uni.switchTab({ url: '/pages/index/index' })
      }, 1000)
    } else {
      throw new Error('绑定失败')
    }
  } catch (err: any) {
    uni.showToast({ title: err.message || '绑定失败', icon: 'none' })
  } finally {
    bindLoading.value = false
  }
}

/**
 * 打开用户协议/隐私政策
 */
function openAgreement(type: string) {
  const urls: Record<string, string> = {
    user: '/pages/webview?url=https://example.com/agreement',
    privacy: '/pages/webview?url=https://example.com/privacy',
  }
  uni.navigateTo({ url: urls[type] || urls.user })
}

onUnmounted(() => {
  if (timer) clearInterval(timer)
  if (bindTimer) clearInterval(bindTimer)
})
</script>

<style scoped>
.login-page {
  padding: 60rpx 40rpx;
  min-height: 100vh;
  background: #f5f5f5;
}
.logo-section {
  text-align: center;
  margin-bottom: 60rpx;
}
.title {
  display: block;
  font-size: 48rpx;
  font-weight: bold;
  color: #333;
}
.subtitle {
  display: block;
  font-size: 28rpx;
  color: #999;
  margin-top: 10rpx;
}
.login-section {
  margin-bottom: 40rpx;
}
.douyin-btn {
  background: linear-gradient(135deg, #fe2c55, #ff6b81);
  color: #fff;
  border-radius: 12rpx;
  font-size: 32rpx;
  height: 88rpx;
  line-height: 88rpx;
  border: none;
}
.douyin-btn::after {
  border: none;
}
.divider {
  display: flex;
  align-items: center;
  margin: 40rpx 0;
}
.line {
  flex: 1;
  height: 1rpx;
  background: #ddd;
}
.divider-text {
  padding: 0 20rpx;
  color: #999;
  font-size: 24rpx;
}
.form-section {
  background: #fff;
  border-radius: 16rpx;
  padding: 40rpx;
}
.form-item {
  margin-bottom: 30rpx;
}
.input {
  width: 100%;
  height: 88rpx;
  border: 1rpx solid #ddd;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
  box-sizing: border-box;
}
.code-row {
  display: flex;
  gap: 20rpx;
}
.code-input {
  flex: 1;
}
.code-btn {
  width: 220rpx;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  border: 1rpx solid #fe2c55;
  color: #fe2c55;
  border-radius: 12rpx;
  font-size: 26rpx;
  background: #fff;
}
.code-btn[disabled] {
  color: #999;
  border-color: #ddd;
}
.login-btn {
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  background: #fe2c55;
  color: #fff;
  border-radius: 12rpx;
  font-size: 32rpx;
  margin-top: 20rpx;
  border: none;
}
.login-btn::after {
  border: none;
}
.bind-section {
  background: #fff;
  border-radius: 16rpx;
  padding: 40rpx;
  margin-top: 40rpx;
}
.bind-header {
  margin-bottom: 30rpx;
}
.bind-title {
  display: block;
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
}
.bind-desc {
  display: block;
  font-size: 26rpx;
  color: #999;
  margin-top: 10rpx;
}
.agreement {
  margin-top: 60rpx;
  text-align: center;
}
.agreement-text {
  font-size: 24rpx;
  color: #999;
}
.link {
  color: #fe2c55;
}
</style>

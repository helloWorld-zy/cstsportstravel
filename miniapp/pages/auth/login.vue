<template>
  <view class="login-page">
    <view class="logo-section">
      <text class="title">欢迎使用</text>
      <text class="subtitle">旅游预订平台</text>
    </view>

    <!-- WeChat Quick Login -->
    <!-- #ifdef MP-WEIXIN -->
    <view class="login-section">
      <button
        class="wechat-btn"
        :loading="wechatLoading"
        @click="handleWechatLogin"
      >
        微信一键登录
      </button>
    </view>

    <view class="divider">
      <view class="line"></view>
      <text class="divider-text">或</text>
      <view class="line"></view>
    </view>
    <!-- #endif -->

    <!-- Phone Login Form -->
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

    <!-- Phone Binding Section (for WeChat new users) -->
    <view v-if="showBindPhone" class="bind-section">
      <view class="bind-header">
        <text class="bind-title">绑定手机号</text>
        <text class="bind-desc">首次使用微信登录，请绑定手机号</text>
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
  </view>
</template>

<script setup lang="ts">
import { useAuth } from '../../shared/composables/useAuth'

const { sendSmsCode, loginWithPhone, loginWithWechat, bindWechatPhone } = useAuth()

const phone = ref('')
const code = ref('')
const sendingCode = ref(false)
const loginLoading = ref(false)
const wechatLoading = ref(false)
const countdown = ref(0)

// Phone binding state
const showBindPhone = ref(false)
const bindPhone = ref('')
const bindCode = ref('')
const bindLoading = ref(false)
const bindCountdown = ref(0)

let timer: ReturnType<typeof setInterval> | null = null
let bindTimer: ReturnType<typeof setInterval> | null = null

async function handleSendCode() {
  if (!phone.value || phone.value.length !== 11) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return
  }

  sendingCode.value = true
  try {
    await sendSmsCode(phone.value)
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
    const result = await loginWithPhone(phone.value, code.value)
    uni.showToast({
      title: result.is_new_user ? '注册成功' : '登录成功',
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

async function handleWechatLogin() {
  wechatLoading.value = true
  try {
    const result = await loginWithWechat()
    if (result.need_bindphone) {
      showBindPhone.value = true
    } else {
      uni.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => {
        uni.switchTab({ url: '/pages/index/index' })
      }, 1000)
    }
  } catch (err: any) {
    uni.showToast({ title: err.message || '微信登录失败', icon: 'none' })
  } finally {
    wechatLoading.value = false
  }
}

async function handleSendBindCode() {
  if (!bindPhone.value || bindPhone.value.length !== 11) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return
  }
  try {
    await sendSmsCode(bindPhone.value)
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
    await bindWechatPhone(bindPhone.value, bindCode.value)
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(() => {
      uni.switchTab({ url: '/pages/index/index' })
    }, 1000)
  } catch (err: any) {
    uni.showToast({ title: err.message || '绑定失败', icon: 'none' })
  } finally {
    bindLoading.value = false
  }
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
.wechat-btn {
  background: #07c160;
  color: #fff;
  border-radius: 12rpx;
  font-size: 32rpx;
  height: 88rpx;
  line-height: 88rpx;
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
  border: 1rpx solid #007aff;
  color: #007aff;
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
  background: #007aff;
  color: #fff;
  border-radius: 12rpx;
  font-size: 32rpx;
  margin-top: 20rpx;
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
</style>

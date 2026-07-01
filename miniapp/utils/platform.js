/**
 * 平台适配层 — 条件编译配置
 *
 * 统一抽象微信/抖音小程序平台差异，仅在平台 API 层做条件编译，
 * 业务逻辑层共享代码基（FR-185）。
 *
 * 支持平台：
 * - MP-WEIXIN: 微信小程序
 * - MP-TOUTIAO: 抖音小程序
 * - MP-ALIPAY: 支付宝小程序（预留）
 * - H5: Web H5
 *
 * T145: Uni-App 条件编译配置（#ifdef MP-TOUTIAO）
 */

/**
 * 平台类型枚举
 */
export const PlatformType = {
  WEIXIN: 'weixin',
  TOUTIAO: 'toutiao',
  ALIPAY: 'alipay',
  H5: 'h5',
  UNKNOWN: 'unknown',
}

/**
 * 获取当前运行平台
 * @returns {string} 平台类型
 */
export function getPlatform() {
  // #ifdef MP-WEIXIN
  return PlatformType.WEIXIN
  // #endif

  // #ifdef MP-TOUTIAO
  return PlatformType.TOUTIAO
  // #endif

  // #ifdef MP-ALIPAY
  return PlatformType.ALIPAY
  // #endif

  // #ifdef H5
  return PlatformType.H5
  // #endif

  return PlatformType.UNKNOWN
}

/**
 * 是否为微信小程序
 * @returns {boolean}
 */
export function isWeixin() {
  // #ifdef MP-WEIXIN
  return true
  // #endif
  // #ifndef MP-WEIXIN
  return false
  // #endif
}

/**
 * 是否为抖音小程序
 * @returns {boolean}
 */
export function isToutiao() {
  // #ifdef MP-TOUTIAO
  return true
  // #endif
  // #ifndef MP-TOUTIAO
  return false
  // #endif
}

/**
 * 是否为小程序环境（微信/抖音/支付宝）
 * @returns {boolean}
 */
export function isMiniProgram() {
  // #ifdef MP
  return true
  // #endif
  // #ifndef MP
  return false
  // #endif
}

// ============================================================
// 登录 API 适配
// ============================================================

/**
 * 平台登录适配器
 * 微信: wx.login → code
 * 抖音: tt.login → code (匿名登录/授权登录)
 *
 * @returns {Promise<{code: string, platform: string}>}
 */
export function platformLogin() {
  return new Promise((resolve, reject) => {
    // #ifdef MP-WEIXIN
    uni.login({
      provider: 'weixin',
      success: (res) => {
        resolve({ code: res.code, platform: PlatformType.WEIXIN })
      },
      fail: (err) => {
        reject(new Error(err.errMsg || '微信登录失败'))
      },
    })
    // #endif

    // #ifdef MP-TOUTIAO
    // 抖音小程序登录：tt.login 获取 code
    // 抖音有两种登录模式：
    // 1. 匿名登录（不需要用户授权）- 获取匿名 code
    // 2. 授权登录（需要用户点击授权按钮）- 获取授权 code
    tt.login({
      force: false, // false=匿名登录, true=强制授权登录
      success: (res) => {
        // res.code 用于后端换取 openid
        // res.anonymous_code 用于匿名登录场景
        resolve({
          code: res.code,
          anonymousCode: res.anonymous_code,
          platform: PlatformType.TOUTIAO,
        })
      },
      fail: (err) => {
        reject(new Error(err.errMsg || '抖音登录失败'))
      },
    })
    // #endif

    // #ifdef MP-ALIPAY
    my.getAuthCode({
      scopes: 'auth_base',
      success: (res) => {
        resolve({ code: res.authCode, platform: PlatformType.ALIPAY })
      },
      fail: (err) => {
        reject(new Error(err.errorMessage || '支付宝登录失败'))
      },
    })
    // #endif
  })
}

// ============================================================
// 用户信息获取适配
// ============================================================

/**
 * 获取用户信息（需要用户授权）
 * 微信: wx.getUserProfile
 * 抖音: tt.getUserInfo
 *
 * @returns {Promise<Object>} 用户信息
 */
export function getUserProfile() {
  return new Promise((resolve, reject) => {
    // #ifdef MP-WEIXIN
    uni.getUserProfile({
      desc: '用于完善用户资料',
      success: (res) => {
        resolve({
          nickName: res.userInfo.nickName,
          avatarUrl: res.userInfo.avatarUrl,
          gender: res.userInfo.gender,
          platform: PlatformType.WEIXIN,
          rawData: res.rawData,
          signature: res.signature,
          encryptedData: res.encryptedData,
          iv: res.iv,
        })
      },
      fail: (err) => {
        reject(new Error(err.errMsg || '获取用户信息失败'))
      },
    })
    // #endif

    // #ifdef MP-TOUTIAO
    tt.getUserInfo({
      withCredentials: true,
      success: (res) => {
        resolve({
          nickName: res.userInfo.nickName,
          avatarUrl: res.userInfo.avatarUrl,
          gender: res.userInfo.gender,
          platform: PlatformType.TOUTIAO,
          rawData: res.rawData,
          signature: res.signature,
          encryptedData: res.encryptedData,
          iv: res.iv,
        })
      },
      fail: (err) => {
        reject(new Error(err.errMsg || '获取用户信息失败'))
      },
    })
    // #endif
  })
}

// ============================================================
// 手机号获取适配
// ============================================================

/**
 * 获取手机号（需要用户点击授权按钮）
 * 微信: button open-type="getPhoneNumber" → getphonenumber 事件
 * 抖音: button open-type="getPhoneNumber" → getphonenumber 事件
 *
 * 注意：此函数用于处理授权回调的加密数据，实际获取需要配合模板中的 button 组件
 *
 * @param {Object} event - 按钮回调事件对象
 * @returns {Promise<{encryptedData: string, iv: string, cloudID?: string}>}
 */
export function parsePhoneNumberEvent(event) {
  return new Promise((resolve, reject) => {
    const detail = event.detail

    if (detail.errMsg && detail.errMsg.includes('fail')) {
      reject(new Error(detail.errMsg || '获取手机号失败'))
      return
    }

    if (detail.encryptedData) {
      resolve({
        encryptedData: detail.encryptedData,
        iv: detail.iv,
        cloudID: detail.cloudID, // 微信云开发场景
      })
    } else {
      reject(new Error('未获取到手机号授权'))
    }
  })
}

// ============================================================
// 分享适配
// ============================================================

/**
 * 触发分享
 * 微信: wx.shareAppMessage
 * 抖音: tt.shareAppMessage
 *
 * @param {Object} options - 分享参数
 * @param {string} options.title - 分享标题
 * @param {string} options.path - 分享路径
 * @param {string} options.imageUrl - 分享图片
 */
export function shareAppMessage(options) {
  // #ifdef MP-WEIXIN
  uni.shareAppMessage({
    title: options.title,
    path: options.path,
    imageUrl: options.imageUrl,
  })
  // #endif

  // #ifdef MP-TOUTIAO
  tt.shareAppMessage({
    title: options.title,
    path: options.path,
    imageUrl: options.imageUrl,
    channel: options.channel || 'video', // 抖音支持 video 带货分享
  })
  // #endif
}

// ============================================================
// 支付适配（统一入口）
// ============================================================

/**
 * 统一支付入口
 * 根据平台调用对应的支付 API
 *
 * @param {Object} paymentData - 支付参数（由后端生成）
 * @returns {Promise<Object>} 支付结果
 */
export function platformPay(paymentData) {
  return new Promise((resolve, reject) => {
    // #ifdef MP-WEIXIN
    uni.requestPayment({
      provider: 'wxpay',
      timeStamp: paymentData.timestamp,
      nonceStr: paymentData.nonce_str,
      package: paymentData.package || `prepay_id=${paymentData.prepay_id}`,
      signType: paymentData.sign_type || 'RSA',
      paySign: paymentData.pay_sign,
      success: (res) => {
        resolve({ success: true, platform: 'wechat', data: res })
      },
      fail: (err) => {
        if (err.errMsg === 'requestPayment:fail cancel') {
          resolve({ success: false, cancelled: true, platform: 'wechat' })
        } else {
          reject({ success: false, message: err.errMsg, platform: 'wechat' })
        }
      },
    })
    // #endif

    // #ifdef MP-TOUTIAO
    tt.pay({
      orderInfo: {
        order_id: paymentData.order_id,
        order_amount: paymentData.order_amount,
        order_title: paymentData.order_title || '旅游订单',
        sign: paymentData.sign,
        trade_no: paymentData.trade_no,
      },
      service: paymentData.service || 1, // 1=抖音支付
      success: (res) => {
        const code = res.code
        if (code === 0) {
          resolve({ success: true, platform: 'douyin', data: res })
        } else if (code === 3) {
          resolve({ success: false, cancelled: true, platform: 'douyin' })
        } else if (code === 9) {
          resolve({ success: false, pending: true, platform: 'douyin', message: '支付处理中' })
        } else {
          reject({ success: false, platform: 'douyin', message: `支付失败: code=${code}` })
        }
      },
      fail: (err) => {
        reject({ success: false, platform: 'douyin', message: err.errMsg || '支付调用失败' })
      },
    })
    // #endif
  })
}

// ============================================================
// 平台特有功能检测
// ============================================================

/**
 * 平台能力检测
 * @returns {Object} 平台能力支持情况
 */
export function getPlatformCapabilities() {
  const platform = getPlatform()

  return {
    platform,
    // 支付能力
    payment: {
      wechat: platform === PlatformType.WEIXIN,
      douyin: platform === PlatformType.TOUTIAO,
      alipay: platform === PlatformType.ALIPAY,
    },
    // 登录能力
    login: {
      wechat: platform === PlatformType.WEIXIN,
      douyin: platform === PlatformType.TOUTIAO,
      phone: true, // 所有平台都支持手机号登录
    },
    // 分享能力
    share: {
      appMessage: true,
      timeline: platform === PlatformType.WEIXIN, // 仅微信支持朋友圈
      douyinVideo: platform === PlatformType.TOUTIAO, // 抖音支持视频带货分享
    },
    // 直播能力
    live: {
      wechat: platform === PlatformType.WEIXIN,
      douyin: platform === PlatformType.TOUTIAO,
    },
    // 定位能力
    location: true,
    // 扫码能力
    scanCode: true,
    // 相册/相机
    album: true,
    camera: true,
  }
}

// ============================================================
// 存储适配（统一 key 前缀避免跨平台冲突）
// ============================================================

const STORAGE_PREFIX = 'cs_travel_'

/**
 * 平台存储适配 — 设置
 * @param {string} key
 * @param {*} value
 */
export function platformSetStorage(key, value) {
  uni.setStorageSync(`${STORAGE_PREFIX}${key}`, value)
}

/**
 * 平台存储适配 — 获取
 * @param {string} key
 * @param {*} defaultValue
 * @returns {*}
 */
export function platformGetStorage(key, defaultValue = null) {
  try {
    const value = uni.getStorageSync(`${STORAGE_PREFIX}${key}`)
    return value !== '' && value !== undefined ? value : defaultValue
  } catch {
    return defaultValue
  }
}

/**
 * 平台存储适配 — 删除
 * @param {string} key
 */
export function platformRemoveStorage(key) {
  uni.removeStorageSync(`${STORAGE_PREFIX}${key}`)
}

// ============================================================
// 导航适配
// ============================================================

/**
 * 跳转到登录页（根据平台选择不同的登录页）
 */
export function navigateToLogin() {
  // #ifdef MP-TOUTIAO
  uni.navigateTo({ url: '/pages-douyin/login' })
  // #endif

  // #ifndef MP-TOUTIAO
  uni.navigateTo({ url: '/pages/auth/login' })
  // #endif
}

/**
 * 平台首页路径
 * @returns {string}
 */
export function getHomePage() {
  return '/pages/index/index'
}

export default {
  PlatformType,
  getPlatform,
  isWeixin,
  isToutiao,
  isMiniProgram,
  platformLogin,
  getUserProfile,
  parsePhoneNumberEvent,
  shareAppMessage,
  platformPay,
  getPlatformCapabilities,
  platformSetStorage,
  platformGetStorage,
  platformRemoveStorage,
  navigateToLogin,
  getHomePage,
}

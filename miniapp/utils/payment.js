/**
 * 支付工具模块
 *
 * 支持多平台支付：微信小程序、抖音小程序
 * 使用条件编译适配不同平台 API
 *
 * FR-161: 银联支付支持
 * T122: 抖音支付适配 (tt.pay API, MP-TOUTIAO)
 */

// #ifdef MP-WEIXIN
/**
 * 微信小程序支付
 * @param {Object} paymentData - 支付参数
 * @param {string} paymentData.prepay_id - 预支付ID
 * @param {string} paymentData.nonce_str - 随机字符串
 * @param {string} paymentData.timestamp - 时间戳
 * @param {string} paymentData.sign_type - 签名类型
 * @param {string} paymentData.pay_sign - 签名
 * @returns {Promise<Object>} 支付结果
 */
export function wxPay(paymentData) {
  return new Promise((resolve, reject) => {
    uni.requestPayment({
      provider: 'wxpay',
      timeStamp: paymentData.timestamp,
      nonceStr: paymentData.nonce_str,
      package: paymentData.package || `prepay_id=${paymentData.prepay_id}`,
      signType: paymentData.sign_type || 'RSA',
      paySign: paymentData.pay_sign,
      success: (res) => {
        resolve({
          success: true,
          platform: 'wechat',
          data: res,
        })
      },
      fail: (err) => {
        if (err.errMsg === 'requestPayment:fail cancel') {
          resolve({
            success: false,
            platform: 'wechat',
            cancelled: true,
            message: '用户取消支付',
          })
        } else {
          reject({
            success: false,
            platform: 'wechat',
            message: err.errMsg || '支付失败',
            error: err,
          })
        }
      },
    })
  })
}
// #endif

// #ifdef MP-TOUTIAO
/**
 * 抖音小程序支付 (tt.pay API)
 * @param {Object} paymentData - 支付参数
 * @param {string} paymentData.order_id - 订单ID
 * @param {number} paymentData.order_amount - 订单金额（分）
 * @param {string} paymentData.order_title - 订单标题
 * @param {string} paymentData.sign - 签名
 * @param {string} paymentData.trade_no - 交易号
 * @returns {Promise<Object>} 支付结果
 */
export function douyinPay(paymentData) {
  return new Promise((resolve, reject) => {
    // 抖音支付使用 tt.pay API
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
        // 抖音支付成功回调
        // res.code: 0=成功, 1=超时, 2=失败, 3=取消, 9=支付中
        const code = res.code
        if (code === 0) {
          resolve({
            success: true,
            platform: 'douyin',
            data: res,
            message: '支付成功',
          })
        } else if (code === 3) {
          resolve({
            success: false,
            platform: 'douyin',
            cancelled: true,
            message: '用户取消支付',
          })
        } else if (code === 9) {
          resolve({
            success: false,
            platform: 'douyin',
            pending: true,
            message: '支付处理中',
          })
        } else {
          reject({
            success: false,
            platform: 'douyin',
            message: `支付失败: code=${code}`,
            error: res,
          })
        }
      },
      fail: (err) => {
        reject({
          success: false,
          platform: 'douyin',
          message: err.errMsg || '支付调用失败',
          error: err,
        })
      },
    })
  })
}
// #endif

// #ifdef MP-ALIPAY
/**
 * 支付宝小程序支付
 * @param {Object} paymentData - 支付参数
 * @returns {Promise<Object>} 支付结果
 */
export function alipayPay(paymentData) {
  return new Promise((resolve, reject) => {
    my.tradePay({
      tradeNO: paymentData.trade_no,
      success: (res) => {
        if (res.resultCode === '9000') {
          resolve({
            success: true,
            platform: 'alipay',
            data: res,
            message: '支付成功',
          })
        } else if (res.resultCode === '6001') {
          resolve({
            success: false,
            platform: 'alipay',
            cancelled: true,
            message: '用户取消支付',
          })
        } else {
          reject({
            success: false,
            platform: 'alipay',
            message: `支付失败: ${res.memo}`,
            error: res,
          })
        }
      },
      fail: (err) => {
        reject({
          success: false,
          platform: 'alipay',
          message: err.errorMessage || '支付调用失败',
          error: err,
        })
      },
    })
  })
}
// #endif

/**
 * 统一支付入口
 * 根据平台和渠道自动选择支付方式
 *
 * @param {Object} options
 * @param {string} options.channel - 支付渠道: alipay/wechat/unionpay/douyin
 * @param {string} options.method - 支付方式: native/jsapi/wap/gateway
 * @param {Object} options.paymentData - 渠道支付参数
 * @param {string} options.paymentType - 款项类型: full/deposit/balance
 * @returns {Promise<Object>} 支付结果
 */
export async function unifiedPay(options) {
  const { channel, method, paymentData, paymentType = 'full' } = options

  console.log(`[Payment] Unified pay: channel=${channel}, method=${method}, type=${paymentType}`)

  // #ifdef MP-TOUTIAO
  // 抖音小程序：使用抖音支付
  if (channel === 'douyin' || channel === 'toutiao') {
    return douyinPay(paymentData)
  }
  // #endif

  // #ifdef MP-WEIXIN
  // 微信小程序：使用微信支付
  if (channel === 'wechat') {
    return wxPay(paymentData)
  }
  // #endif

  // #ifdef MP-ALIPAY
  // 支付宝小程序：使用支付宝支付
  if (channel === 'alipay') {
    return alipayPay(paymentData)
  }
  // #endif

  // 银联支付：小程序内跳转H5支付
  if (channel === 'unionpay') {
    // 小程序内无法直接使用银联SDK，跳转到H5支付页面
    return {
      success: false,
      platform: 'unionpay',
      needRedirect: true,
      redirectUrl: paymentData.pay_url,
      message: '银联支付请在浏览器中完成',
    }
  }

  throw new Error(`不支持的支付渠道: ${channel}`)
}

/**
 * 支付结果处理
 * 统一处理支付成功/失败的后续逻辑
 *
 * @param {Object} result - 支付结果
 * @param {Object} orderInfo - 订单信息
 */
export function handlePaymentResult(result, orderInfo) {
  if (result.success) {
    uni.showToast({
      title: '支付成功',
      icon: 'success',
    })

    // 跳转到订单详情页
    setTimeout(() => {
      uni.redirectTo({
        url: `/pages/order/detail?id=${orderInfo.orderId}`,
      })
    }, 1500)
  } else if (result.cancelled) {
    // 用户取消，不提示
    console.log('[Payment] User cancelled payment')
  } else if (result.pending) {
    uni.showToast({
      title: '支付处理中，请稍后查看',
      icon: 'none',
      duration: 3000,
    })
  } else if (result.needRedirect) {
    // 银联支付需要跳转
    // #ifdef H5
    window.location.href = result.redirectUrl
    // #endif

    // #ifdef MP
    uni.navigateTo({
      url: `/pages/webview?url=${encodeURIComponent(result.redirectUrl)}`,
    })
    // #endif
  } else {
    uni.showToast({
      title: result.message || '支付失败',
      icon: 'none',
      duration: 3000,
    })
  }
}

/**
 * 支付类型显示文本
 * @param {string} paymentType - full/deposit/balance
 * @returns {string} 显示文本
 */
export function getPaymentTypeText(paymentType) {
  const texts = {
    full: '全额支付',
    deposit: '定金支付',
    balance: '尾款支付',
  }
  return texts[paymentType] || '全额支付'
}

/**
 * 支付渠道显示文本
 * @param {string} channel - 支付渠道
 * @returns {string} 显示文本
 */
export function getChannelText(channel) {
  const texts = {
    alipay: '支付宝',
    wechat: '微信支付',
    unionpay: '银联支付',
    douyin: '抖音支付',
    toutiao: '抖音支付',
  }
  return texts[channel] || channel
}

export default {
  unifiedPay,
  handlePaymentResult,
  getPaymentTypeText,
  getChannelText,
  // #ifdef MP-WEIXIN
  wxPay,
  // #endif
  // #ifdef MP-TOUTIAO
  douyinPay,
  // #endif
  // #ifdef MP-ALIPAY
  alipayPay,
  // #endif
}

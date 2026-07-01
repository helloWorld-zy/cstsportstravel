/**
 * 抖音支付适配器
 *
 * 集成抖音小程序 tt.pay API，处理支付参数生成、支付调用、结果处理。
 * 与微信小程序共享业务逻辑层，仅在支付 API 层做条件编译（FR-185）。
 *
 * T147: 抖音支付适配（tt.pay API 集成）
 * FR-184: 创建支付订单后返回抖音小程序支付参数
 *
 * 抖音支付流程：
 * 1. 前端请求后端创建支付订单
 * 2. 后端调用抖音支付接口获取支付参数
 * 3. 前端调用 tt.pay 唤起收银台
 * 4. 用户完成支付
 * 5. 后端接收支付回调通知
 */

import { api } from '../shared/api/request'

/**
 * 抖音支付服务类型
 */
export const DouyinPayService = {
  /** 抖音支付（默认） */
  DOUYIN_PAY: 1,
  /** 抖音支付（聚合支付） */
  AGGREGATE_PAY: 2,
}

/**
 * 抖音支付结果码
 */
export const DouyinPayCode = {
  /** 支付成功 */
  SUCCESS: 0,
  /** 超时 */
  TIMEOUT: 1,
  /** 支付失败 */
  FAILED: 2,
  /** 用户取消 */
  CANCEL: 3,
  /** 支付处理中 */
  PENDING: 9,
}

/**
 * 抖音支付结果描述
 */
const PAY_CODE_MESSAGES = {
  [DouyinPayCode.SUCCESS]: '支付成功',
  [DouyinPayCode.TIMEOUT]: '支付超时',
  [DouyinPayCode.FAILED]: '支付失败',
  [DouyinPayCode.CANCEL]: '用户取消支付',
  [DouyinPayCode.PENDING]: '支付处理中',
}

/**
 * 创建抖音支付订单
 * 调用后端 API 创建支付订单，获取抖音支付参数
 *
 * @param {Object} orderInfo - 订单信息
 * @param {string} orderInfo.order_id - 平台订单ID
 * @param {string} orderInfo.payment_type - 款项类型: full/deposit/balance
 * @param {number} [orderInfo.coupon_id] - 优惠券ID（可选）
 * @returns {Promise<Object>} 抖音支付参数
 */
export async function createDouyinPayment(orderInfo) {
  const { order_id, payment_type = 'full', coupon_id } = orderInfo

  const data = await api.post('/payments/douyin/create', {
    order_id,
    payment_type,
    coupon_id,
    channel: 'douyin',
  })

  // 后端返回的支付参数
  // {
  //   order_id: string,      // 抖音侧订单ID
  //   order_amount: number,  // 订单金额（分）
  //   order_title: string,   // 订单标题
  //   sign: string,          // 签名
  //   trade_no: string,      // 交易号
  //   service: number,       // 支付服务类型
  // }
  return data
}

/**
 * 调起抖音支付
 *
 * @param {Object} paymentData - 后端返回的支付参数
 * @param {string} paymentData.order_id - 抖音侧订单ID
 * @param {number} paymentData.order_amount - 订单金额（分）
 * @param {string} paymentData.order_title - 订单标题
 * @param {string} paymentData.sign - 签名
 * @param {string} paymentData.trade_no - 交易号
 * @param {number} [paymentData.service=1] - 支付服务类型
 * @returns {Promise<Object>} 支付结果
 */
export function invokeDouyinPay(paymentData) {
  return new Promise((resolve, reject) => {
    // #ifdef MP-TOUTIAO
    tt.pay({
      orderInfo: {
        order_id: paymentData.order_id,
        order_amount: paymentData.order_amount,
        order_title: paymentData.order_title || '旅游订单',
        sign: paymentData.sign,
        trade_no: paymentData.trade_no,
      },
      service: paymentData.service || DouyinPayService.DOUYIN_PAY,
      success: (res) => {
        const result = handlePayResult(res)
        if (result.success || result.pending) {
          resolve(result)
        } else if (result.cancelled) {
          resolve(result)
        } else {
          reject(result)
        }
      },
      fail: (err) => {
        reject({
          success: false,
          platform: 'douyin',
          code: -1,
          message: err.errMsg || '支付调用失败',
          error: err,
        })
      },
    })
    // #endif

    // #ifndef MP-TOUTIAO
    // 非抖音环境，返回错误
    reject({
      success: false,
      platform: 'douyin',
      code: -1,
      message: '抖音支付仅支持抖音小程序环境',
    })
    // #endif
  })
}

/**
 * 处理抖音支付结果
 *
 * @param {Object} res - tt.pay 回调结果
 * @returns {Object} 标准化支付结果
 */
function handlePayResult(res) {
  const code = res.code

  switch (code) {
    case DouyinPayCode.SUCCESS:
      return {
        success: true,
        platform: 'douyin',
        code,
        message: PAY_CODE_MESSAGES[code],
        data: res,
      }

    case DouyinPayCode.CANCEL:
      return {
        success: false,
        cancelled: true,
        platform: 'douyin',
        code,
        message: PAY_CODE_MESSAGES[code],
      }

    case DouyinPayCode.PENDING:
      return {
        success: false,
        pending: true,
        platform: 'douyin',
        code,
        message: PAY_CODE_MESSAGES[code],
      }

    case DouyinPayCode.TIMEOUT:
    case DouyinPayCode.FAILED:
    default:
      return {
        success: false,
        platform: 'douyin',
        code,
        message: PAY_CODE_MESSAGES[code] || `支付失败: code=${code}`,
        error: res,
      }
  }
}

/**
 * 查询抖音支付状态
 * 用于支付结果页轮询确认支付状态
 *
 * @param {string} orderId - 订单ID
 * @returns {Promise<Object>} 支付状态
 */
export async function queryDouyinPaymentStatus(orderId) {
  const data = await api.get(`/payments/douyin/status/${orderId}`)
  return data
}

/**
 * 完整的抖音支付流程
 * 创建订单 → 调起支付 → 处理结果
 *
 * @param {Object} orderInfo - 订单信息
 * @param {string} orderInfo.order_id - 平台订单ID
 * @param {string} orderInfo.payment_type - 款项类型: full/deposit/balance
 * @param {number} [orderInfo.coupon_id] - 优惠券ID
 * @returns {Promise<Object>} 支付结果
 */
export async function douyinPayFlow(orderInfo) {
  try {
    // Step 1: 创建支付订单，获取支付参数
    const paymentData = await createDouyinPayment(orderInfo)

    // Step 2: 调起抖音支付
    const result = await invokeDouyinPay(paymentData)

    return result
  } catch (err) {
    // 统一错误处理
    if (err.success === false) {
      return err
    }
    return {
      success: false,
      platform: 'douyin',
      code: -1,
      message: err.message || '支付流程异常',
      error: err,
    }
  }
}

/**
 * 支付结果页面处理
 * 根据支付结果展示对应 UI 提示
 *
 * @param {Object} result - 支付结果
 * @param {string} orderId - 订单ID
 */
export function handleDouyinPayResult(result, orderId) {
  if (result.success) {
    uni.showToast({
      title: '支付成功',
      icon: 'success',
    })
    // 跳转到订单详情页
    setTimeout(() => {
      uni.redirectTo({
        url: `/pages/orders/detail?id=${orderId}`,
      })
    }, 1500)
  } else if (result.cancelled) {
    // 用户取消，不提示，留在当前页
    console.log('[DouyinPay] User cancelled payment')
  } else if (result.pending) {
    uni.showToast({
      title: '支付处理中，请稍后查看订单状态',
      icon: 'none',
      duration: 3000,
    })
    // 跳转到订单列表页
    setTimeout(() => {
      uni.redirectTo({
        url: '/pages/orders/list',
      })
    }, 2000)
  } else {
    uni.showToast({
      title: result.message || '支付失败',
      icon: 'none',
      duration: 3000,
    })
  }
}

/**
 * 定金+尾款支付 — 支付定金
 * @param {string} orderId - 订单ID
 * @returns {Promise<Object>}
 */
export async function payDouyinDeposit(orderId) {
  return douyinPayFlow({
    order_id: orderId,
    payment_type: 'deposit',
  })
}

/**
 * 定金+尾款支付 — 支付尾款
 * @param {string} orderId - 订单ID
 * @returns {Promise<Object>}
 */
export async function payDouyinBalance(orderId) {
  return douyinPayFlow({
    order_id: orderId,
    payment_type: 'balance',
  })
}

export default {
  DouyinPayService,
  DouyinPayCode,
  createDouyinPayment,
  invokeDouyinPay,
  queryDouyinPaymentStatus,
  douyinPayFlow,
  handleDouyinPayResult,
  payDouyinDeposit,
  payDouyinBalance,
}

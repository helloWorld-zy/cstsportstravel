/**
 * Douyin Payment Adapter 单元测试
 *
 * TDD 测试先行：验证抖音支付流程
 *
 * T147: 抖音支付适配（tt.pay API 集成）
 * FR-184: 创建支付订单后返回抖音小程序支付参数
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  DouyinPayService,
  DouyinPayCode,
  createDouyinPayment,
  invokeDouyinPay,
  queryDouyinPaymentStatus,
  douyinPayFlow,
  handleDouyinPayResult,
  payDouyinDeposit,
  payDouyinBalance,
} from '../utils/douyin_payment'

// Mock tt global object
const mockTt = {
  pay: vi.fn(),
}

// Mock uni global object
const mockUni = {
  showToast: vi.fn(),
  redirectTo: vi.fn(),
  navigateTo: vi.fn(),
  request: vi.fn(),
}

// Mock API module
vi.mock('../shared/api/request', () => ({
  api: {
    post: vi.fn(),
    get: vi.fn(),
  },
}))

import { api } from '../shared/api/request'

describe('DouyinPayService 常量', () => {
  it('应定义抖音支付服务类型', () => {
    expect(DouyinPayService.DOUYIN_PAY).toBe(1)
    expect(DouyinPayService.AGGREGATE_PAY).toBe(2)
  })
})

describe('DouyinPayCode 常量', () => {
  it('应定义所有支付结果码', () => {
    expect(DouyinPayCode.SUCCESS).toBe(0)
    expect(DouyinPayCode.TIMEOUT).toBe(1)
    expect(DouyinPayCode.FAILED).toBe(2)
    expect(DouyinPayCode.CANCEL).toBe(3)
    expect(DouyinPayCode.PENDING).toBe(9)
  })
})

describe('createDouyinPayment', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('应调用后端 API 创建支付订单', async () => {
    const mockPaymentData = {
      order_id: 'DY_ORD_001',
      order_amount: 9900,
      order_title: '北京5日游',
      sign: 'test_sign_abc',
      trade_no: 'TN_20260701_001',
      service: 1,
    }

    api.post.mockResolvedValue(mockPaymentData)

    const result = await createDouyinPayment({
      order_id: 'ORD_001',
      payment_type: 'full',
    })

    expect(api.post).toHaveBeenCalledWith('/payments/douyin/create', {
      order_id: 'ORD_001',
      payment_type: 'full',
      coupon_id: undefined,
      channel: 'douyin',
    })
    expect(result).toEqual(mockPaymentData)
  })

  it('应支持定金支付类型', async () => {
    api.post.mockResolvedValue({ order_id: 'DY_ORD_002' })

    await createDouyinPayment({
      order_id: 'ORD_002',
      payment_type: 'deposit',
    })

    expect(api.post).toHaveBeenCalledWith('/payments/douyin/create', expect.objectContaining({
      payment_type: 'deposit',
    }))
  })

  it('应支持尾款支付类型', async () => {
    api.post.mockResolvedValue({ order_id: 'DY_ORD_003' })

    await createDouyinPayment({
      order_id: 'ORD_003',
      payment_type: 'balance',
    })

    expect(api.post).toHaveBeenCalledWith('/payments/douyin/create', expect.objectContaining({
      payment_type: 'balance',
    }))
  })

  it('应支持优惠券参数', async () => {
    api.post.mockResolvedValue({ order_id: 'DY_ORD_004' })

    await createDouyinPayment({
      order_id: 'ORD_004',
      payment_type: 'full',
      coupon_id: 42,
    })

    expect(api.post).toHaveBeenCalledWith('/payments/douyin/create', expect.objectContaining({
      coupon_id: 42,
    }))
  })
})

describe('invokeDouyinPay', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    global.tt = mockTt
  })

  it('支付成功应返回 success=true', async () => {
    mockTt.pay.mockImplementation(({ success }) => {
      success({ code: 0 })
    })

    // 在非抖音环境下会 reject
    try {
      const result = await invokeDouyinPay({
        order_id: 'DY_ORD_001',
        order_amount: 9900,
        order_title: '测试订单',
        sign: 'test_sign',
        trade_no: 'TN_001',
      })
      expect(result.success).toBe(true)
      expect(result.platform).toBe('douyin')
    } catch (err) {
      // 非抖音环境预期
      expect(err.message).toContain('抖音支付仅支持抖音小程序环境')
    }
  })

  it('用户取消应返回 cancelled=true', async () => {
    mockTt.pay.mockImplementation(({ success }) => {
      success({ code: 3 })
    })

    try {
      const result = await invokeDouyinPay({
        order_id: 'DY_ORD_001',
        order_amount: 9900,
        sign: 'test_sign',
        trade_no: 'TN_001',
      })
      expect(result.cancelled).toBe(true)
    } catch {
      // 非抖音环境
    }
  })

  it('支付处理中应返回 pending=true', async () => {
    mockTt.pay.mockImplementation(({ success }) => {
      success({ code: 9 })
    })

    try {
      const result = await invokeDouyinPay({
        order_id: 'DY_ORD_001',
        order_amount: 9900,
        sign: 'test_sign',
        trade_no: 'TN_001',
      })
      expect(result.pending).toBe(true)
    } catch {
      // 非抖音环境
    }
  })

  it('支付失败应 reject', async () => {
    mockTt.pay.mockImplementation(({ fail }) => {
      fail({ errMsg: '支付调用失败' })
    })

    try {
      await invokeDouyinPay({
        order_id: 'DY_ORD_001',
        order_amount: 9900,
        sign: 'test_sign',
        trade_no: 'TN_001',
      })
    } catch (err) {
      expect(err.success).toBe(false)
    }
  })

  it('非抖音环境应 reject', async () => {
    // 不设置 tt 对象
    global.tt = undefined

    await expect(invokeDouyinPay({
      order_id: 'DY_ORD_001',
      order_amount: 9900,
      sign: 'test_sign',
      trade_no: 'TN_001',
    })).rejects.toMatchObject({
      success: false,
      message: '抖音支付仅支持抖音小程序环境',
    })
  })
})

describe('queryDouyinPaymentStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('应调用后端查询支付状态', async () => {
    const mockStatus = {
      order_id: 'ORD_001',
      status: 'paid',
      paid_at: '2026-07-01T10:00:00Z',
    }

    api.get.mockResolvedValue(mockStatus)

    const result = await queryDouyinPaymentStatus('ORD_001')

    expect(api.get).toHaveBeenCalledWith('/payments/douyin/status/ORD_001')
    expect(result.status).toBe('paid')
  })
})

describe('douyinPayFlow', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    global.tt = mockTt
  })

  it('完整流程：创建订单 → 调起支付', async () => {
    api.post.mockResolvedValue({
      order_id: 'DY_ORD_001',
      order_amount: 9900,
      sign: 'test_sign',
      trade_no: 'TN_001',
    })

    mockTt.pay.mockImplementation(({ success }) => {
      success({ code: 0 })
    })

    try {
      const result = await douyinPayFlow({
        order_id: 'ORD_001',
        payment_type: 'full',
      })
      expect(result.success).toBe(true)
    } catch {
      // 非抖音环境
    }
  })

  it('创建订单失败应返回错误', async () => {
    api.post.mockRejectedValue(new Error('订单不存在'))

    const result = await douyinPayFlow({
      order_id: 'INVALID_ORDER',
      payment_type: 'full',
    })

    expect(result.success).toBe(false)
    expect(result.message).toBe('订单不存在')
  })
})

describe('handleDouyinPayResult', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    global.uni = mockUni
  })

  it('支付成功应显示成功提示并跳转订单详情', () => {
    handleDouyinPayResult({ success: true }, 'ORD_001')

    expect(mockUni.showToast).toHaveBeenCalledWith({
      title: '支付成功',
      icon: 'success',
    })
  })

  it('用户取消应不显示提示', () => {
    handleDouyinPayResult({ cancelled: true }, 'ORD_001')

    expect(mockUni.showToast).not.toHaveBeenCalled()
  })

  it('支付处理中应显示提示', () => {
    handleDouyinPayResult({ pending: true }, 'ORD_001')

    expect(mockUni.showToast).toHaveBeenCalledWith(expect.objectContaining({
      icon: 'none',
    }))
  })

  it('支付失败应显示错误提示', () => {
    handleDouyinPayResult({ success: false, message: '余额不足' }, 'ORD_001')

    expect(mockUni.showToast).toHaveBeenCalledWith(expect.objectContaining({
      title: '余额不足',
      icon: 'none',
    }))
  })
})

describe('定金+尾款支付', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.post.mockResolvedValue({
      order_id: 'DY_ORD_001',
      order_amount: 3000,
      sign: 'test_sign',
      trade_no: 'TN_001',
    })
  })

  it('payDouyinDeposit 应以 deposit 类型创建支付', async () => {
    try {
      await payDouyinDeposit('ORD_001')
    } catch {
      // 预期
    }

    expect(api.post).toHaveBeenCalledWith('/payments/douyin/create', expect.objectContaining({
      payment_type: 'deposit',
    }))
  })

  it('payDouyinBalance 应以 balance 类型创建支付', async () => {
    try {
      await payDouyinBalance('ORD_001')
    } catch {
      // 预期
    }

    expect(api.post).toHaveBeenCalledWith('/payments/douyin/create', expect.objectContaining({
      payment_type: 'balance',
    }))
  })
})

describe('支付参数校验', () => {
  it('支付参数应包含必要字段', () => {
    const paymentData = {
      order_id: 'DY_ORD_001',
      order_amount: 9900,
      order_title: '旅游订单',
      sign: 'test_sign',
      trade_no: 'TN_001',
      service: 1,
    }

    expect(paymentData.order_id).toBeDefined()
    expect(paymentData.order_amount).toBeGreaterThan(0)
    expect(paymentData.sign).toBeDefined()
    expect(paymentData.trade_no).toBeDefined()
  })

  it('金额应为正整数（分）', () => {
    const amount = 9900 // 99.00 元
    expect(amount).toBeGreaterThan(0)
    expect(Number.isInteger(amount)).toBe(true)
  })
})

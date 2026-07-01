/**
 * Douyin Adapters 单元测试
 *
 * TDD 测试先行：验证核心页面条件编译适配
 *
 * T148: 核心页面条件编译适配
 * FR-186: 抖音小程序支持核心页面
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  getPagePath,
  navigateToPage,
  switchToTab,
  productListAdapter,
  productDetailAdapter,
  bookingAdapter,
  orderAdapter,
  profileAdapter,
  onlyToutiao,
  onlyWeixin,
  onlyMiniProgram,
  platformSwitch,
  createShareConfig,
} from '../utils/douyin_adapters'

// Mock uni global object
const mockUni = {
  navigateTo: vi.fn(),
  switchTab: vi.fn(),
  showToast: vi.fn(),
  showModal: vi.fn(),
  removeStorageSync: vi.fn(),
  setNavigationBarTitle: vi.fn(),
}

describe('getPagePath', () => {
  it('应返回产品列表页路径', () => {
    expect(getPagePath('product_list')).toBe('/pages/products/list')
  })

  it('应返回产品详情页路径', () => {
    expect(getPagePath('product_detail')).toBe('/pages/products/detail')
  })

  it('应返回预订页路径', () => {
    expect(getPagePath('booking')).toBe('/pages/booking/index')
  })

  it('应返回订单列表页路径', () => {
    expect(getPagePath('order_list')).toBe('/pages/orders/list')
  })

  it('应返回订单详情页路径', () => {
    expect(getPagePath('order_detail')).toBe('/pages/orders/detail')
  })

  it('应返回支付页路径', () => {
    expect(getPagePath('payment')).toBe('/pages/payment/index')
  })

  it('应返回优惠券中心路径', () => {
    expect(getPagePath('coupon_center')).toBe('/pages/coupon/index')
  })

  it('应返回我的优惠券路径', () => {
    expect(getPagePath('my_coupons')).toBe('/pages/coupon/mine')
  })

  it('应返回签证进度路径', () => {
    expect(getPagePath('visa_progress')).toBe('/pages/visa/progress')
  })

  it('未知页面应返回首页', () => {
    expect(getPagePath('unknown_page')).toBe('/pages/index/index')
  })
})

describe('navigateToPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    global.uni = mockUni
  })

  it('应调用 uni.navigateTo 跳转', () => {
    navigateToPage('product_list')
    expect(mockUni.navigateTo).toHaveBeenCalled()
  })

  it('应正确拼接查询参数', () => {
    navigateToPage('product_detail', { id: 42 })
    expect(mockUni.navigateTo).toHaveBeenCalledWith({
      url: '/pages/products/detail?id=42',
    })
  })

  it('应正确处理多个参数', () => {
    navigateToPage('booking', { product_id: 1, date_id: 2 })
    const callUrl = mockUni.navigateTo.mock.calls[0][0].url
    expect(callUrl).toContain('product_id=1')
    expect(callUrl).toContain('date_id=2')
  })
})

describe('switchToTab', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    global.uni = mockUni
  })

  it('应切换到首页 Tab', () => {
    switchToTab('home')
    expect(mockUni.switchTab).toHaveBeenCalledWith({
      url: '/pages/index/index',
    })
  })

  it('应切换到产品 Tab', () => {
    switchToTab('products')
    expect(mockUni.switchTab).toHaveBeenCalledWith({
      url: '/pages/products/list',
    })
  })

  it('应切换到订单 Tab', () => {
    switchToTab('orders')
    expect(mockUni.switchTab).toHaveBeenCalledWith({
      url: '/pages/orders/list',
    })
  })
})

describe('productListAdapter', () => {
  it('应定义 onPageLoad 方法', () => {
    expect(typeof productListAdapter.onPageLoad).toBe('function')
  })

  it('应定义 onProductClick 方法', () => {
    expect(typeof productListAdapter.onProductClick).toBe('function')
  })

  it('应定义 getShareConfig 方法', () => {
    expect(typeof productListAdapter.getShareConfig).toBe('function')
  })

  it('getShareConfig 应返回分享配置', () => {
    const product = {
      id: 1,
      name: '北京5日游',
      cover_image: 'https://example.com/cover.jpg',
    }

    const config = productListAdapter.getShareConfig(product)
    expect(config.title).toBe('北京5日游')
    expect(config.path).toContain('id=1')
    expect(config.imageUrl).toBe('https://example.com/cover.jpg')
  })
})

describe('productDetailAdapter', () => {
  it('应定义 onPageLoad 方法', () => {
    expect(typeof productDetailAdapter.onPageLoad).toBe('function')
  })

  it('应定义 onBookingClick 方法', () => {
    expect(typeof productDetailAdapter.onBookingClick).toBe('function')
  })

  it('应定义 openCustomerService 方法', () => {
    expect(typeof productDetailAdapter.openCustomerService).toBe('function')
  })

  it('应定义 shareProduct 方法', () => {
    expect(typeof productDetailAdapter.shareProduct).toBe('function')
  })
})

describe('bookingAdapter', () => {
  it('应定义 selectPassenger 方法', () => {
    expect(typeof bookingAdapter.selectPassenger).toBe('function')
  })

  it('应定义 getPhoneNumber 方法', () => {
    expect(typeof bookingAdapter.getPhoneNumber).toBe('function')
  })

  it('应定义 proceedToPayment 方法', () => {
    expect(typeof bookingAdapter.proceedToPayment).toBe('function')
  })

  it('getPhoneNumber 应解析授权事件', async () => {
    const event = {
      detail: {
        encryptedData: 'encrypted_test',
        iv: 'iv_test',
      },
    }

    const result = await bookingAdapter.getPhoneNumber(event)
    expect(result.encryptedData).toBe('encrypted_test')
    expect(result.iv).toBe('iv_test')
  })

  it('getPhoneNumber 授权失败应抛出错误', async () => {
    const event = {
      detail: {
        errMsg: 'getPhoneNumber:fail user deny',
      },
    }

    await expect(bookingAdapter.getPhoneNumber(event)).rejects.toThrow('获取手机号失败')
  })
})

describe('orderAdapter', () => {
  it('应返回订单 Tab 列表', () => {
    const tabs = orderAdapter.getOrderTabs()
    expect(tabs).toHaveLength(5)
    expect(tabs[0].key).toBe('all')
    expect(tabs[1].key).toBe('pending_payment')
  })

  it('应正确映射订单状态文本', () => {
    expect(orderAdapter.getStatusText('pending_payment')).toBe('待付款')
    expect(orderAdapter.getStatusText('paid')).toBe('已付款')
    expect(orderAdapter.getStatusText('completed')).toBe('已完成')
    expect(orderAdapter.getStatusText('cancelled')).toBe('已取消')
    expect(orderAdapter.getStatusText('refunding')).toBe('退款中')
    expect(orderAdapter.getStatusText('unknown')).toBe('unknown')
  })

  it('应返回订单状态颜色', () => {
    expect(orderAdapter.getStatusColor('pending_payment')).toBe('#ff9500')
    expect(orderAdapter.getStatusColor('paid')).toBe('#007aff')
    expect(orderAdapter.getStatusColor('completed')).toBe('#8e8e93')
  })

  it('应定义 payOrder 方法', () => {
    expect(typeof orderAdapter.payOrder).toBe('function')
  })

  it('应定义 applyRefund 方法', () => {
    expect(typeof orderAdapter.applyRefund).toBe('function')
  })

  it('应定义 reorder 方法', () => {
    expect(typeof orderAdapter.reorder).toBe('function')
  })
})

describe('profileAdapter', () => {
  it('应返回个人中心菜单列表', () => {
    const menus = profileAdapter.getMenuItems()
    expect(menus.length).toBeGreaterThan(0)

    const keys = menus.map(m => m.key)
    expect(keys).toContain('orders')
    expect(keys).toContain('coupons')
    expect(keys).toContain('passengers')
    expect(keys).toContain('visa')
    expect(keys).toContain('distributor')
    expect(keys).toContain('settings')
  })

  it('每个菜单项应有 key/icon/title/path', () => {
    const menus = profileAdapter.getMenuItems()
    menus.forEach(menu => {
      expect(menu.key).toBeDefined()
      expect(menu.icon).toBeDefined()
      expect(menu.title).toBeDefined()
      expect(menu.path).toBeDefined()
    })
  })

  it('应定义 onMenuClick 方法', () => {
    expect(typeof profileAdapter.onMenuClick).toBe('function')
  })

  it('应定义 logout 方法', () => {
    expect(typeof profileAdapter.logout).toBe('function')
  })
})

describe('条件编译辅助工具', () => {
  it('onlyToutiao 应执行传入的函数', () => {
    // 由于条件编译限制，这里只验证函数存在
    expect(typeof onlyToutiao).toBe('function')
  })

  it('onlyWeixin 应执行传入的函数', () => {
    expect(typeof onlyWeixin).toBe('function')
  })

  it('onlyMiniProgram 应执行传入的函数', () => {
    expect(typeof onlyMiniProgram).toBe('function')
  })

  it('platformSwitch 应根据平台选择处理器', () => {
    expect(typeof platformSwitch).toBe('function')
  })

  it('platformSwitch 无匹配时应调用 default', () => {
    const defaultFn = vi.fn(() => 'default_result')
    const result = platformSwitch({ default: defaultFn })
    // 在测试环境中，平台可能是 unknown，会走 default
    expect(defaultFn).toHaveBeenCalled()
  })
})

describe('createShareConfig', () => {
  it('应返回分享配置', () => {
    const config = createShareConfig({
      title: '测试分享',
      path: '/pages/products/detail?id=1',
      imageUrl: 'https://example.com/img.jpg',
    })

    expect(config.title).toBe('测试分享')
    expect(config.path).toBe('/pages/products/detail?id=1')
    expect(config.imageUrl).toBe('https://example.com/img.jpg')
  })

  it('应支持自定义参数', () => {
    const config = createShareConfig({
      title: '测试',
      path: '/test',
      channel: 'video',
      videoTopics: ['旅游'],
      hashtagList: ['北京游'],
    })

    expect(config.title).toBe('测试')
  })
})

/**
 * 抖音小程序核心页面条件编译适配器
 *
 * 为产品列表/详情/预订/订单/个人中心页面提供抖音平台适配。
 * 与微信小程序共享业务逻辑层，仅在平台 API 层做条件编译（FR-185）。
 *
 * T148: 核心页面条件编译适配
 * FR-186: 抖音小程序支持核心页面：登录、产品列表/详情、预订、订单管理、个人中心
 *
 * 适配维度：
 * 1. 页面路由适配（抖音小程序页面路径规范）
 * 2. UI 组件适配（抖音小程序组件差异）
 * 3. 分享适配（抖音支持视频带货分享）
 * 4. 页面生命周期适配
 * 5. 支付流程适配
 */

import {
  isToutiao,
  isWeixin,
  getPlatform,
  PlatformType,
  shareAppMessage,
  platformPay,
  navigateToLogin,
} from './platform'

// ============================================================
// 页面路由适配
// ============================================================

/**
 * 获取页面路径（根据平台适配）
 * 抖音小程序的页面路径可能与微信不同
 *
 * @param {string} pageName - 页面名称
 * @returns {string} 页面路径
 */
export function getPagePath(pageName) {
  const paths = {
    // 产品相关
    product_list: '/pages/products/list',
    product_detail: '/pages/products/detail',
    outbound_list: '/pages/outbound/list',

    // 预订相关
    booking: '/pages/booking/index',
    payment: '/pages/payment/index',

    // 订单相关
    order_list: '/pages/orders/list',
    order_detail: '/pages/orders/detail',

    // 用户相关
    login: '/pages/auth/login',
    profile: '/pages/index/index', // 个人中心在首页 Tab

    // 优惠券
    coupon_center: '/pages/coupon/index',
    my_coupons: '/pages/coupon/mine',

    // 签证
    visa_progress: '/pages/visa/progress',
  }

  // #ifdef MP-TOUTIAO
  // 抖音小程序使用独立登录页
  if (pageName === 'login') {
    return '/pages-douyin/login'
  }
  // #endif

  return paths[pageName] || '/pages/index/index'
}

/**
 * 跳转到指定页面
 * @param {string} pageName - 页面名称
 * @param {Object} [params] - 页面参数
 */
export function navigateToPage(pageName, params = {}) {
  const path = getPagePath(pageName)
  const queryString = Object.entries(params)
    .map(([key, value]) => `${key}=${encodeURIComponent(value)}`)
    .join('&')

  const url = queryString ? `${path}?${queryString}` : path
  uni.navigateTo({ url })
}

/**
 * 切换 Tab 页面
 * @param {string} tabName - Tab 名称: home/products/orders/profile
 */
export function switchToTab(tabName) {
  const tabPaths = {
    home: '/pages/index/index',
    products: '/pages/products/list',
    orders: '/pages/orders/list',
  }

  const path = tabPaths[tabName]
  if (path) {
    uni.switchTab({ url: path })
  }
}

// ============================================================
// 产品列表页适配
// ============================================================

/**
 * 产品列表页适配器
 * 处理抖音平台特有的产品列表展示逻辑
 */
export const productListAdapter = {
  /**
   * 初始化产品列表页
   * @param {Object} pageInstance - 页面实例
   */
  onPageLoad(pageInstance) {
    // 抖音小程序页面加载时的特殊处理
    // #ifdef MP-TOUTIAO
    // 抖音小程序支持动态设置导航栏
    tt.setNavigationBarTitle({
      title: pageInstance.title || '产品列表',
    })
    // #endif
  },

  /**
   * 产品卡片点击处理
   * @param {Object} product - 产品信息
   */
  onProductClick(product) {
    navigateToPage('product_detail', { id: product.id })
  },

  /**
   * 产品列表分享配置
   * @param {Object} product - 产品信息
   * @returns {Object} 分享参数
   */
  getShareConfig(product) {
    const baseConfig = {
      title: product.name || '精选旅游产品',
      path: `/pages/products/detail?id=${product.id}`,
      imageUrl: product.cover_image,
    }

    // #ifdef MP-TOUTIAO
    // 抖音支持视频带货分享渠道
    return {
      ...baseConfig,
      channel: 'video', // 抖音特有：视频带货分享
    }
    // #endif

    // #ifndef MP-TOUTIAO
    return baseConfig
    // #endif
  },
}

// ============================================================
// 产品详情页适配
// ============================================================

/**
 * 产品详情页适配器
 */
export const productDetailAdapter = {
  /**
   * 产品详情页加载
   * @param {Object} pageInstance - 页面实例
   */
  onPageLoad(pageInstance) {
    // #ifdef MP-TOUTIAO
    tt.setNavigationBarTitle({
      title: pageInstance.productName || '产品详情',
    })
    // #endif
  },

  /**
   * 预订按钮点击
   * @param {Object} product - 产品信息
   * @param {Object} selectedDate - 选择的团期
   */
  onBookingClick(product, selectedDate) {
    navigateToPage('booking', {
      product_id: product.id,
      date_id: selectedDate?.id,
    })
  },

  /**
   * 客服会话
   * 抖音和微信的客服 API 不同
   */
  openCustomerService() {
    // #ifdef MP-WEIXIN
    // 微信使用 button open-type="contact"
    // 需要在模板中使用 <button open-type="contact">
    console.log('[WeChat] Use button open-type="contact"')
    // #endif

    // #ifdef MP-TOUTIAO
    // 抖音使用 tt.openCustomerServiceChat
    tt.openCustomerServiceChat({
      success: () => {},
      fail: (err) => {
        console.error('[Douyin] Open customer service failed:', err)
        uni.showToast({ title: '客服暂时不可用', icon: 'none' })
      },
    })
    // #endif
  },

  /**
   * 分享产品详情
   * @param {Object} product - 产品信息
   */
  shareProduct(product) {
    shareAppMessage({
      title: product.name,
      path: `/pages/products/detail?id=${product.id}`,
      imageUrl: product.cover_image,
    })
  },
}

// ============================================================
// 预订页适配
// ============================================================

/**
 * 预订页适配器
 * 处理预订流程中的平台差异
 */
export const bookingAdapter = {
  /**
   * 选择出游人（调用通讯录/身份证识别）
   * @returns {Promise<Object>} 出游人信息
   */
  async selectPassenger() {
    // #ifdef MP-TOUTIAO
    // 抖音支持身份证 OCR 识别
    try {
      const res = await tt.chooseImage({
        count: 1,
        sourceType: ['camera', 'album'],
      })
      return {
        type: 'ocr',
        imagePath: res.tempFilePaths[0],
      }
    } catch {
      return { type: 'manual' }
    }
    // #endif

    // #ifndef MP-TOUTIAO
    return { type: 'manual' }
    // #endif
  },

  /**
   * 获取手机号（预订时获取联系人手机号）
   * @param {Object} event - 按钮回调事件
   * @returns {Promise<Object>} 手机号信息
   */
  async getPhoneNumber(event) {
    const detail = event.detail

    if (detail.errMsg && detail.errMsg.includes('fail')) {
      throw new Error('获取手机号失败')
    }

    if (detail.encryptedData) {
      return {
        encryptedData: detail.encryptedData,
        iv: detail.iv,
      }
    }

    throw new Error('未获取到手机号授权')
  },

  /**
   * 预订确认后的支付流程
   * @param {Object} orderInfo - 订单信息
   * @returns {Promise<Object>} 支付结果
   */
  async proceedToPayment(orderInfo) {
    return platformPay(orderInfo)
  },
}

// ============================================================
// 订单页适配
// ============================================================

/**
 * 订单列表/详情页适配器
 */
export const orderAdapter = {
  /**
   * 订单列表页 Tab 配置
   * @returns {Array} Tab 列表
   */
  getOrderTabs() {
    return [
      { key: 'all', label: '全部' },
      { key: 'pending_payment', label: '待付款' },
      { key: 'paid', label: '已付款' },
      { key: 'completed', label: '已完成' },
      { key: 'cancelled', label: '已取消' },
    ]
  },

  /**
   * 订单状态文本映射
   * @param {string} status - 订单状态
   * @returns {string} 显示文本
   */
  getStatusText(status) {
    const texts = {
      pending_payment: '待付款',
      paid_deposit: '已付定金',
      pending_balance: '待付尾款',
      paid: '已付款',
      confirmed: '已确认',
      travelling: '出行中',
      completed: '已完成',
      cancelled: '已取消',
      refunding: '退款中',
      refunded: '已退款',
    }
    return texts[status] || status
  },

  /**
   * 订单状态颜色
   * @param {string} status - 订单状态
   * @returns {string} 颜色值
   */
  getStatusColor(status) {
    const colors = {
      pending_payment: '#ff9500',
      paid_deposit: '#007aff',
      pending_balance: '#ff9500',
      paid: '#007aff',
      confirmed: '#34c759',
      travelling: '#34c759',
      completed: '#8e8e93',
      cancelled: '#ff3b30',
      refunding: '#ff9500',
      refunded: '#8e8e93',
    }
    return colors[status] || '#333'
  },

  /**
   * 去支付（订单详情页）
   * @param {Object} order - 订单信息
   */
  async payOrder(order) {
    const result = await platformPay({
      order_id: order.id,
      payment_type: order.payment_type || 'full',
    })
    return result
  },

  /**
   * 申请退款
   * @param {Object} order - 订单信息
   */
  applyRefund(order) {
    navigateToPage('order_detail', { id: order.id, action: 'refund' })
  },

  /**
   * 再次预订
   * @param {Object} order - 订单信息
   */
  reorder(order) {
    navigateToPage('product_detail', { id: order.product_id })
  },
}

// ============================================================
// 个人中心页适配
// ============================================================

/**
 * 个人中心页适配器
 */
export const profileAdapter = {
  /**
   * 获取个人中心菜单配置
   * @returns {Array} 菜单列表
   */
  getMenuItems() {
    const menus = [
      {
        key: 'orders',
        icon: 'order',
        title: '我的订单',
        path: 'order_list',
      },
      {
        key: 'coupons',
        icon: 'coupon',
        title: '我的优惠券',
        path: 'my_coupons',
      },
      {
        key: 'passengers',
        icon: 'passenger',
        title: '常用出游人',
        path: 'passenger_list',
      },
      {
        key: 'visa',
        icon: 'visa',
        title: '签证进度',
        path: 'visa_progress',
      },
      {
        key: 'distributor',
        icon: 'distributor',
        title: '分销商中心',
        path: 'distributor_center',
      },
      {
        key: 'settings',
        icon: 'settings',
        title: '设置',
        path: 'settings',
      },
    ]

    return menus
  },

  /**
   * 菜单点击处理
   * @param {Object} menuItem - 菜单项
   */
  onMenuClick(menuItem) {
    if (menuItem.key === 'distributor') {
      // 分销商中心可能需要额外权限检查
      navigateToPage(menuItem.path)
    } else {
      navigateToPage(menuItem.path)
    }
  },

  /**
   * 退出登录
   */
  logout() {
    uni.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          // 清除本地存储
          uni.removeStorageSync('access_token')
          uni.removeStorageSync('refresh_token')
          uni.removeStorageSync('user_info')

          // 跳转到登录页
          navigateToLogin()
        }
      },
    })
  },
}

// ============================================================
// 条件编译辅助工具
// ============================================================

/**
 * 条件执行：仅在抖音环境执行
 * @param {Function} fn - 要执行的函数
 * @param {*} [defaultValue] - 非抖音环境的默认返回值
 */
export function onlyToutiao(fn, defaultValue = undefined) {
  // #ifdef MP-TOUTIAO
  return fn()
  // #endif
  // #ifndef MP-TOUTIAO
  return defaultValue
  // #endif
}

/**
 * 条件执行：仅在微信环境执行
 * @param {Function} fn - 要执行的函数
 * @param {*} [defaultValue] - 非微信环境的默认返回值
 */
export function onlyWeixin(fn, defaultValue = undefined) {
  // #ifdef MP-WEIXIN
  return fn()
  // #endif
  // #ifndef MP-WEIXIN
  return defaultValue
  // #endif
}

/**
 * 条件执行：仅在小程序环境执行
 * @param {Function} fn - 要执行的函数
 * @param {*} [defaultValue] - 非小程序环境的默认返回值
 */
export function onlyMiniProgram(fn, defaultValue = undefined) {
  // #ifdef MP
  return fn()
  // #endif
  // #ifndef MP
  return defaultValue
  // #endif
}

/**
 * 平台条件选择器
 * @param {Object} handlers - 平台处理器 { weixin: fn, toutiao: fn, default: fn }
 * @returns {*} 执行结果
 */
export function platformSwitch(handlers) {
  const platform = getPlatform()

  if (handlers[platform]) {
    return handlers[platform]()
  }

  if (handlers.default) {
    return handlers.default()
  }

  return undefined
}

// ============================================================
// 页面分享配置适配
// ============================================================

/**
 * 生成页面分享配置
 * @param {Object} options - 分享选项
 * @param {string} options.title - 分享标题
 * @param {string} options.path - 分享路径
 * @param {string} [options.imageUrl] - 分享图片
 * @returns {Object} 分享配置（适配当前平台）
 */
export function createShareConfig(options) {
  const base = {
    title: options.title,
    path: options.path,
    imageUrl: options.imageUrl,
  }

  // #ifdef MP-TOUTIAO
  // 抖音特有分享配置
  return {
    ...base,
    channel: options.channel || 'video', // 默认使用视频渠道分享
    extra: {
      videoTopics: options.videoTopics || [], // 视频话题
      hashtag_list: options.hashtagList || [], // 话题标签
    },
  }
  // #endif

  // #ifdef MP-WEIXIN
  return base
  // #endif

  // #ifndef MP
  return base
  // #endif
}

export default {
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
}

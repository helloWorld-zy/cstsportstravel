/**
 * Mini-program auth composable.
 * Handles wx.login, phone binding, token management, and user state.
 */
import { api, setTokens, clearTokens } from '../api/request'

export interface User {
  id: number
  phone: string
  nickname: string
  avatar_url: string
  real_name_status: string
  member_level: number
  status: string
  created_at: string
}

interface LoginResponse {
  user: User
  access_token: string
  refresh_token: string
  is_new_user: boolean
}

interface WechatLoginResponse {
  user?: User
  access_token?: string
  refresh_token?: string
  need_bindphone: boolean
}

// Global reactive state (shared across components)
let _user: User | null = null
const _listeners: Array<(user: User | null) => void> = []

function notifyListeners() {
  _listeners.forEach(fn => fn(_user))
}

export function useAuth() {
  function getUser(): User | null {
    return _user
  }

  function setUser(u: User | null) {
    _user = u
    notifyListeners()
  }

  function onUserChange(fn: (user: User | null) => void) {
    _listeners.push(fn)
  }

  async function sendSmsCode(phone: string): Promise<{ expires_in: number }> {
    return api.post('/auth/sms-code', { phone })
  }

  async function loginWithPhone(phone: string, code: string): Promise<LoginResponse> {
    const data = await api.post<LoginResponse>('/auth/login', { phone, code })
    setTokens(data.access_token, data.refresh_token)
    _user = data.user
    notifyListeners()
    return data
  }

  /**
   * WeChat wx.login flow:
   * 1. Call wx.login() to get a temporary code
   * 2. Send code to backend to exchange for OpenID
   * 3. If need_bindphone, prompt user to bind phone
   */
  async function loginWithWechat(): Promise<WechatLoginResponse> {
    return new Promise((resolve, reject) => {
      // #ifdef MP-WEIXIN
      uni.login({
        provider: 'weixin',
        success: async (loginRes) => {
          try {
            const data = await api.post<WechatLoginResponse>('/auth/wechat', {
              code: loginRes.code,
            })
            if (!data.need_bindphone && data.access_token) {
              setTokens(data.access_token, data.refresh_token!)
              _user = data.user!
              notifyListeners()
            }
            resolve(data)
          } catch (err) {
            reject(err)
          }
        },
        fail: (err) => {
          reject(new Error(err.errMsg || '微信登录失败'))
        },
      })
      // #endif
      // #ifndef MP-WEIXIN
      reject(new Error('微信登录仅支持小程序环境'))
      // #endif
    })
  }

  /**
   * Bind phone number to WeChat account (second step).
   */
  async function bindWechatPhone(phone: string, code: string): Promise<LoginResponse> {
    // Get wx.login code again for binding
    return new Promise((resolve, reject) => {
      // #ifdef MP-WEIXIN
      uni.login({
        provider: 'weixin',
        success: async (loginRes) => {
          try {
            const data = await api.post<WechatLoginResponse>('/auth/wechat', {
              code: loginRes.code,
              bind_phone: phone,
              bind_code: code,
            })
            if (data.access_token) {
              setTokens(data.access_token, data.refresh_token!)
              _user = data.user!
              notifyListeners()
              resolve({
                user: data.user!,
                access_token: data.access_token,
                refresh_token: data.refresh_token!,
                is_new_user: true,
              })
            } else {
              reject(new Error('绑定失败'))
            }
          } catch (err) {
            reject(err)
          }
        },
        fail: (err) => {
          reject(new Error(err.errMsg || '微信登录失败'))
        },
      })
      // #endif
    })
  }

  async function fetchProfile(): Promise<User> {
    const data = await api.get<User>('/users/me')
    _user = data
    notifyListeners()
    return data
  }

  function logout() {
    clearTokens()
    _user = null
    notifyListeners()
    uni.navigateTo({ url: '/pages/auth/login' })
  }

  async function init() {
    const token = uni.getStorageSync('access_token')
    if (token && !_user) {
      try {
        await fetchProfile()
      } catch {
        clearTokens()
      }
    }
  }

  return {
    getUser,
    setUser,
    onUserChange,
    sendSmsCode,
    loginWithPhone,
    loginWithWechat,
    bindWechatPhone,
    fetchProfile,
    logout,
    init,
  }
}

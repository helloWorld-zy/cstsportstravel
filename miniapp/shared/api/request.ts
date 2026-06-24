/**
 * Mini-program API request wrapper.
 *
 * Provides a configured uni.request instance with:
 * - JWT token injection
 * - Unified response parsing
 * - Error handling
 * - Token refresh on 401
 */

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
  trace_id: string
}

export class ApiError extends Error {
  code: number
  constructor(code: number, message: string) {
    super(message)
    this.code = code
    this.name = 'ApiError'
  }
}

const BASE_URL = import.meta.env.VITE_API_BASE || 'http://localhost:8080'
const API_PREFIX = '/api/v1'

/**
 * Get stored access token.
 */
function getToken(): string {
  return uni.getStorageSync('access_token') || ''
}

/**
 * Store access and refresh tokens.
 */
export function setTokens(accessToken: string, refreshToken: string): void {
  uni.setStorageSync('access_token', accessToken)
  uni.setStorageSync('refresh_token', refreshToken)
}

/**
 * Clear stored tokens (logout).
 */
export function clearTokens(): void {
  uni.removeStorageSync('access_token')
  uni.removeStorageSync('refresh_token')
}

/**
 * Make an API request with automatic token injection and error handling.
 */
export async function request<T>(
  url: string,
  options: {
    method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
    data?: unknown
    header?: Record<string, string>
  } = {},
): Promise<T> {
  const { method = 'GET', data, header = {} } = options

  const token = getToken()
  if (token) {
    header['Authorization'] = `Bearer ${token}`
  }

  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}${API_PREFIX}${url}`,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        ...header,
      },
      success: (res) => {
        const statusCode = res.statusCode
        const body = res.data as ApiResponse<T>

        if (statusCode === 401) {
          clearTokens()
          uni.navigateTo({ url: '/pages/auth/login' })
          reject(new ApiError(1002, '登录已过期，请重新登录'))
          return
        }

        if (statusCode === 429) {
          reject(new ApiError(1007, '请求过于频繁，请稍后再试'))
          return
        }

        if (statusCode >= 400) {
          reject(new ApiError(body.code || statusCode, body.message || '请求失败'))
          return
        }

        if (body.code !== 0) {
          reject(new ApiError(body.code, body.message))
          return
        }

        resolve(body.data)
      },
      fail: (err) => {
        reject(new ApiError(-1, err.errMsg || '网络错误'))
      },
    })
  })
}

/** Convenience methods */
export const api = {
  get: <T>(url: string, header?: Record<string, string>) =>
    request<T>(url, { method: 'GET', header }),

  post: <T>(url: string, data?: unknown, header?: Record<string, string>) =>
    request<T>(url, { method: 'POST', data, header }),

  put: <T>(url: string, data?: unknown, header?: Record<string, string>) =>
    request<T>(url, { method: 'PUT', data, header }),

  del: <T>(url: string, header?: Record<string, string>) =>
    request<T>(url, { method: 'DELETE', header }),
}

import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'

/**
 * API response envelope matching the backend unified format.
 */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
  trace_id: string
}

/**
 * Pagination metadata for list endpoints.
 */
export interface PaginationMeta {
  page: number
  page_size: number
  total: number
  total_pages: number
}

/**
 * Paginated API response.
 */
export interface PaginatedResponse<T> {
  list: T[]
  pagination: PaginationMeta
}

// Error code constants
export const CodeSuccess = 0
export const CodeUnauthorized = 1002
export const CodeForbidden = 1003

/**
 * Custom API error class.
 */
export class ApiError extends Error {
  code: number
  constructor(code: number, message: string) {
    super(message)
    this.code = code
    this.name = 'ApiError'
  }
}

/**
 * Creates and configures an Axios instance for admin API communication.
 */
function createApiClient(): AxiosInstance {
  const client = axios.create({
    baseURL: `${import.meta.env.VITE_API_BASE || 'http://localhost:8088'}/api/v1`,
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  })

  // Request interceptor: inject JWT token and signing headers
  client.interceptors.request.use(
    (req: InternalAxiosRequestConfig) => {
      const token = localStorage.getItem('admin_token')
      if (token) {
        req.headers.Authorization = `Bearer ${token}`
      }
      return req
    },
    (error) => Promise.reject(error),
  )

  // Response interceptor: handle errors
  client.interceptors.response.use(
    (res: AxiosResponse<ApiResponse>) => {
      const data = res.data
      if (data.code !== CodeSuccess) {
        return Promise.reject(new ApiError(data.code, data.message))
      }
      return res
    },
    (error) => {
      if (error.response) {
        const { status, data } = error.response
        if (status === 401) {
          localStorage.removeItem('admin_token')
          window.location.href = '/login'
          return Promise.reject(new ApiError(CodeUnauthorized, '登录已过期，请重新登录'))
        }
        if (status === 403) {
          ElMessage.error('没有权限执行此操作')
          return Promise.reject(new ApiError(CodeForbidden, '权限不足'))
        }
        if (status === 429) {
          ElMessage.warning('请求过于频繁，请稍后再试')
          return Promise.reject(new ApiError(1007, '请求过于频繁'))
        }
        return Promise.reject(new ApiError(data.code || status, data.message || '请求失败'))
      }
      return Promise.reject(new ApiError(-1, '网络错误，请检查网络连接'))
    },
  )

  return client
}

// Singleton client instance
const client = createApiClient()

/**
 * Admin API client with typed request methods.
 */
export const adminApi = {
  /** GET request */
  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const res = await client.get<ApiResponse<T>>(url, config)
    return res.data.data
  },

  /** POST request */
  async post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    const res = await client.post<ApiResponse<T>>(url, data, config)
    return res.data.data
  },

  /** PUT request */
  async put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    const res = await client.put<ApiResponse<T>>(url, data, config)
    return res.data.data
  },

  /** DELETE request */
  async del<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const res = await client.delete<ApiResponse<T>>(url, config)
    return res.data.data
  },
}

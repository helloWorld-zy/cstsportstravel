import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'

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

// Error code constants matching backend
export const CodeSuccess = 0
export const CodeUnauthorized = 1002
export const CodeForbidden = 1003

/**
 * Creates and configures an Axios instance for API communication.
 */
function createApiClient(): AxiosInstance {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase as string

  const client = axios.create({
    baseURL: `${baseURL}/api/v1`,
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  })

  // Request interceptor: inject JWT token
  client.interceptors.request.use(
    (req: InternalAxiosRequestConfig) => {
      const token = useCookie('access_token').value
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
          // Token expired — clear and redirect to login
          const token = useCookie('access_token')
          token.value = null
          navigateTo('/auth/login')
          return Promise.reject(new ApiError(CodeUnauthorized, '登录已过期，请重新登录'))
        }
        if (status === 429) {
          return Promise.reject(new ApiError(1007, '请求过于频繁，请稍后再试'))
        }
        return Promise.reject(new ApiError(data.code || status, data.message || '请求失败'))
      }
      return Promise.reject(new ApiError(-1, '网络错误，请检查网络连接'))
    },
  )

  return client
}

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
 * Composable providing the API client instance.
 * The client is shared across the application via Nuxt's payload.
 */
export function useApi() {
  const client = createApiClient()

  return {
    client,

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
}

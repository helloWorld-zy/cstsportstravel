/**
 * Auth composable for user authentication state management.
 * Handles login, logout, token storage, and user state.
 */
import { useApi } from './useApi'

export interface User {
  id: number
  phone: string
  nickname: string
  avatar_url: string
  real_name_status: 'unverified' | 'pending' | 'verified' | 'rejected'
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

export function useAuth() {
  const user = useState<User | null>('auth_user', () => null)
  const token = useCookie('access_token', { maxAge: 60 * 15 }) // 15 min
  const refreshToken = useCookie('refresh_token', { maxAge: 60 * 60 * 24 * 7 }) // 7 days

  const isLoggedIn = computed(() => !!token.value && !!user.value)

  async function sendSmsCode(phone: string): Promise<{ expires_in: number; code?: string }> {
    const api = useApi()
    return api.post('/auth/sms-code', { phone })
  }

  async function login(phone: string, code: string): Promise<LoginResponse> {
    const api = useApi()
    const data = await api.post<LoginResponse>('/auth/login', { phone, code })
    token.value = data.access_token
    refreshToken.value = data.refresh_token
    user.value = data.user
    return data
  }

  async function fetchProfile(): Promise<User> {
    const api = useApi()
    const data = await api.get<User>('/users/me')
    user.value = data
    return data
  }

  function logout() {
    token.value = null
    refreshToken.value = null
    user.value = null
    navigateTo('/auth/login')
  }

  async function refreshAccessToken(): Promise<boolean> {
    if (!refreshToken.value) return false
    try {
      const api = useApi()
      const data = await api.post<{ access_token: string; refresh_token: string }>('/auth/refresh-token', {
        refresh_token: refreshToken.value,
      })
      token.value = data.access_token
      refreshToken.value = data.refresh_token
      return true
    } catch {
      logout()
      return false
    }
  }

  // Initialize user state from token on client side
  async function init() {
    if (token.value && !user.value) {
      try {
        await fetchProfile()
      } catch {
        // Token might be expired, try refresh
        const refreshed = await refreshAccessToken()
        if (refreshed) {
          await fetchProfile()
        }
      }
    }
  }

  async function loginWithWechat(code: string): Promise<LoginResponse> {
    const api = useApi()
    const data = await api.post<LoginResponse>('/auth/wechat', { code })
    token.value = data.access_token
    refreshToken.value = data.refresh_token
    user.value = data.user
    return data
  }

  return {
    user,
    token,
    isLoggedIn,
    sendSmsCode,
    login,
    loginWithWechat,
    logout,
    fetchProfile,
    refreshAccessToken,
    init,
  }
}

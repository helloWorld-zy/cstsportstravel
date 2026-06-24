import { defineStore } from 'pinia'
import { adminApi } from '@/api/request'
import { loadDynamicRoutes } from '@/router'

interface UserInfo {
  id: number
  username: string
  real_name: string
  must_change_password: boolean
}

interface LoginResponse {
  user: UserInfo
  access_token: string
  permissions: string[]
  menus: unknown[]
}

interface AdminMeResponse {
  user: UserInfo
  roles: string[]
  permissions: string[]
  last_login: string
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('admin_token') || '',
    userInfo: null as UserInfo | null,
    permissions: [] as string[],
    roles: [] as string[],
    menus: [] as unknown[],
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    userRoles: (state) => state.roles,
    userPermissions: (state) => state.permissions,
    mustChangePassword: (state) => state.userInfo?.must_change_password || false,
  },

  actions: {
    /**
     * Admin login with username and password.
     */
    async login(username: string, password: string) {
      const res = await adminApi.post<LoginResponse>(
        '/auth/admin/login',
        { username, password },
      )

      this.token = res.access_token
      localStorage.setItem('admin_token', res.access_token)

      this.userInfo = res.user
      this.permissions = res.permissions || []
      this.menus = res.menus || []

      // Load dynamic routes based on permissions
      loadDynamicRoutes(this.permissions)
    },

    /**
     * Logout and clear state.
     */
    logout() {
      this.token = ''
      this.userInfo = null
      this.permissions = []
      this.roles = []
      this.menus = []
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_refresh_token')
    },

    /**
     * Fetch current user info and permissions.
     */
    async fetchUserInfo() {
      try {
        const res = await adminApi.get<AdminMeResponse>('/admin/users/me')
        this.userInfo = res.user
        this.roles = res.roles || []
        this.permissions = res.permissions || []
        loadDynamicRoutes(this.permissions)
      } catch {
        this.logout()
      }
    },
  },
})

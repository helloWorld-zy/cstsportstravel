import { defineStore } from 'pinia'
import { adminApi } from '@/api/request'
import { loadDynamicRoutes } from '@/router'

interface UserInfo {
  id: number
  username: string
  real_name: string
  roles: string[]
  permissions: string[]
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('admin_token') || '',
    userInfo: null as UserInfo | null,
    permissions: [] as string[],
    menus: [] as unknown[],
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    userRoles: (state) => state.userInfo?.roles || [],
    userPermissions: (state) => state.permissions,
  },

  actions: {
    /**
     * Admin login with username and password.
     */
    async login(username: string, password: string) {
      const res = await adminApi.post<{ access_token: string; refresh_token: string; user: UserInfo }>(
        '/auth/admin/login',
        { username, password },
      )

      this.token = res.access_token
      localStorage.setItem('admin_token', res.access_token)
      localStorage.setItem('admin_refresh_token', res.refresh_token)

      this.userInfo = res.user
      this.permissions = res.user.permissions || []

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
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_refresh_token')
    },

    /**
     * Fetch current user info and permissions.
     */
    async fetchUserInfo() {
      try {
        const user = await adminApi.get<UserInfo>('/admin/users/me')
        this.userInfo = user
        this.permissions = user.permissions || []
        loadDynamicRoutes(this.permissions)
      } catch {
        this.logout()
      }
    },
  },
})

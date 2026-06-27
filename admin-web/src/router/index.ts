import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

/**
 * Static routes — always available regardless of permissions.
 */
const staticRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login.vue'),
    meta: { requiresAuth: false },
  },
]

/**
 * Dynamic routes — filtered by RBAC menu permissions.
 * In production, these are loaded from the admin menu API and
 * dynamically added via router.addRoute().
 */
const dynamicRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/products',
    children: [
      {
        path: 'products',
        name: 'ProductList',
        component: () => import('@/views/product/ProductList.vue'),
        meta: { title: '产品管理', permission: 'product:list' },
      },
      {
        path: 'products/create',
        name: 'ProductCreate',
        component: () => import('@/views/product/ProductForm.vue'),
        meta: { title: '发布产品', permission: 'product:create' },
      },
      {
        path: 'products/edit/:id',
        name: 'ProductEdit',
        component: () => import('@/views/product/ProductForm.vue'),
        meta: { title: '编辑产品', permission: 'product:update' },
      },
      {
        path: 'products/review',
        name: 'ProductReview',
        component: () => import('@/views/product/ProductReview.vue'),
        meta: { title: '产品审核', permission: 'product:approve' },
      },
      {
        path: 'orders',
        name: 'OrderList',
        component: () => import('@/views/order/OrderList.vue'),
        meta: { title: '订单管理', permission: 'order:list' },
      },
      {
        path: 'system/users',
        name: 'UserManage',
        component: () => import('@/views/system/UserManage.vue'),
        meta: { title: '用户管理', permission: 'user:manage' },
      },
      {
        path: 'system/roles',
        name: 'RoleManage',
        component: () => import('@/views/system/RoleManage.vue'),
        meta: { title: '角色管理', permission: 'role:manage' },
      },
      {
        path: 'config/homepage',
        name: 'HomepageConfig',
        component: () => import('@/views/config/HomepageConfig.vue'),
        meta: { title: '首页配置', permission: 'config:manage' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes: [...staticRoutes, ...dynamicRoutes],
})

/**
 * Auth guard — checks for admin token and redirects to login if missing.
 */
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('admin_token')

  if (to.meta.requiresAuth !== false && !token) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
    return
  }

  // Skip auth check for login page when already authenticated
  if (to.name === 'Login' && token) {
    next({ path: '/' })
    return
  }

  next()
})

/**
 * Load dynamic routes from RBAC menu permissions.
 * Called after successful admin login.
 *
 * In production, this fetches the menu tree from GET /api/v1/admin/menus
 * and registers routes dynamically based on the user's role permissions.
 */
export function loadDynamicRoutes(_permissions: string[]): void {
  // For MVP, all dynamic routes are registered statically.
  // In production, filter dynamicRoutes by permission and call router.addRoute()
  // for each accessible route.
}

export default router

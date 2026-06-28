import { adminApi } from './request'

/** Admin user as returned by the API. */
export interface AdminUser {
  id: number
  username: string
  real_name: string
  phone: string
  email: string
  status: 'active' | 'locked' | 'disabled'
  roles: string[]
  supplier_id: number | null
  must_change_password: boolean
  last_login_at: string | null
  created_at: string
}

/** Role as returned by the API. */
export interface Role {
  id: number
  role_name: string
  role_code: string
  description: string
  is_system: boolean
  status: string
  permission_ids: number[]
  menu_ids: number[]
}

/** Permission tree node. */
export interface PermissionNode {
  id: number
  permission_name: string
  permission_code: string
  permission_type: 'menu' | 'button' | 'api' | 'data'
  description: string
  children: PermissionNode[]
}

/** Menu tree node. */
export interface MenuNode {
  id: number
  menu_name: string
  menu_path: string
  component_name: string
  icon: string
  sort_order: number
  permission_code: string
  children: MenuNode[]
}

/** Paginated user list response. */
export interface UserListResponse {
  items: AdminUser[]
  total: number
  page: number
}

/** Create user input. */
export interface CreateUserInput {
  username: string
  real_name: string
  phone?: string
  email?: string
  role_ids: number[]
  supplier_id?: number
  initial_password?: string
}

/** Create role input. */
export interface CreateRoleInput {
  role_name: string
  role_code: string
  description?: string
  permission_ids?: number[]
  menu_ids?: number[]
}

/** Update role input. */
export interface UpdateRoleInput {
  role_name?: string
  description?: string
  permission_ids?: number[]
  menu_ids?: number[]
  status?: string
}

// ---------- User API ----------

/** List admin users with optional filters. */
export function listUsers(params?: {
  keyword?: string
  role_code?: string
  status?: string
  page?: number
  page_size?: number
}): Promise<UserListResponse> {
  return adminApi.get<UserListResponse>('/admin/users', { params })
}

/** Create a new admin user. */
export function createUser(data: CreateUserInput): Promise<{
  id: number
  username: string
  initial_password: string
  must_change_password: boolean
}> {
  return adminApi.post('/admin/users', data)
}

/** Update user status (freeze/activate). */
export function updateUserStatus(
  userId: number,
  status: 'active' | 'locked' | 'disabled',
): Promise<void> {
  return adminApi.put(`/admin/users/${userId}/status`, { status })
}

/** Update user roles. */
export function updateUserRoles(
  userId: number,
  roleIds: number[],
): Promise<void> {
  return adminApi.put(`/admin/users/${userId}/roles`, { role_ids: roleIds })
}

// ---------- Role API ----------

/** List all roles. */
export function listRoles(): Promise<Role[]> {
  return adminApi.get<Role[]>('/admin/roles')
}

/** Create a new role. */
export function createRole(data: CreateRoleInput): Promise<{ id: number; role_name: string; role_code: string }> {
  return adminApi.post('/admin/roles', data)
}

/** Update a role. */
export function updateRole(
  roleId: number,
  data: UpdateRoleInput,
): Promise<{ id: number; role_name: string; role_code: string; status: string }> {
  return adminApi.put(`/admin/roles/${roleId}`, data)
}

// ---------- Permission & Menu API ----------

/** Get the permission tree. */
export function getPermissionTree(): Promise<PermissionNode[]> {
  return adminApi.get<PermissionNode[]>('/admin/permissions')
}

/** Get the menu tree (filtered by current user's role). */
export function getMenuTree(): Promise<MenuNode[]> {
  return adminApi.get<MenuNode[]>('/admin/menus')
}

// ---------- MFA API ----------

/** Initiate MFA setup (returns QR code URL). */
export function mfaSetup(): Promise<{ message: string }> {
  return adminApi.post('/admin/mfa/setup')
}

/** Verify TOTP code for MFA enrollment or sensitive operation. */
export function mfaVerify(totpCode: string): Promise<{ verified: boolean; message: string }> {
  return adminApi.post('/admin/mfa/verify', { totp_code: totpCode })
}

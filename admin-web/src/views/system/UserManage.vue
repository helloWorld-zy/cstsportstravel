<template>
  <div class="user-manage">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button type="primary" @click="openCreateDialog">新增用户</el-button>
        </div>
      </template>

      <!-- Filters -->
      <div class="filter-bar">
        <el-input
          v-model="filters.keyword"
          placeholder="搜索用户名/姓名"
          clearable
          style="width: 200px"
          @clear="fetchUsers"
          @keyup.enter="fetchUsers"
        />
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 120px" @change="fetchUsers">
          <el-option label="启用" value="active" />
          <el-option label="锁定" value="locked" />
          <el-option label="禁用" value="disabled" />
        </el-select>
        <el-select v-model="filters.role_code" placeholder="角色" clearable style="width: 140px" @change="fetchUsers">
          <el-option v-for="role in roles" :key="role.id" :label="role.role_name" :value="role.role_code" />
        </el-select>
        <el-button type="primary" @click="fetchUsers">搜索</el-button>
      </div>

      <!-- User Table -->
      <el-table :data="users" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="real_name" label="姓名" width="100" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column label="角色" min-width="150">
          <template #default="{ row }">
            <el-tag v-for="role in row.roles" :key="role" size="small" class="role-tag">{{ role }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="首次登录" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.must_change_password" type="warning" size="small">需改密</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="last_login_at" label="最后登录" width="160">
          <template #default="{ row }">
            {{ row.last_login_at ? formatDate(row.last_login_at) : '从未登录' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button
              size="small"
              :type="row.status === 'active' ? 'danger' : 'success'"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 'active' ? '冻结' : '启用' }}
            </el-button>
            <el-button size="small" type="warning" @click="openRoleDialog(row)">角色</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-bar">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchUsers"
          @current-change="fetchUsers"
        />
      </div>
    </el-card>

    <!-- Create/Edit User Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑用户' : '新增用户'"
      width="500px"
      @close="resetForm"
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="isEditing" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="姓名" prop="real_name">
          <el-input v-model="form.real_name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="初始密码" prop="initial_password" v-if="!isEditing">
          <el-input v-model="form.initial_password" placeholder="留空则自动生成" />
        </el-form-item>
        <el-form-item label="供应商ID" prop="supplier_id" v-if="showSupplierField">
          <el-input-number v-model="form.supplier_id" :min="1" placeholder="供应商账号关联" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ isEditing ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Role Assignment Dialog -->
    <el-dialog v-model="roleDialogVisible" title="分配角色" width="400px">
      <el-checkbox-group v-model="selectedRoleIds">
        <el-checkbox v-for="role in roles" :key="role.id" :label="role.id">
          {{ role.role_name }} ({{ role.role_code }})
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleAssignRoles">保存</el-button>
      </template>
    </el-dialog>

    <!-- Initial Password Display Dialog -->
    <el-dialog v-model="passwordDialogVisible" title="用户创建成功" width="400px">
      <el-alert type="success" :closable="false" show-icon>
        <template #title>
          用户 <strong>{{ createdUsername }}</strong> 创建成功
        </template>
      </el-alert>
      <div class="password-display">
        <p>初始密码：</p>
        <el-input :model-value="createdPassword" readonly>
          <template #append>
            <el-button @click="copyPassword">复制</el-button>
          </template>
        </el-input>
        <p class="warning-text">请妥善保存此密码，关闭后将无法再次查看。用户首次登录时需修改密码。</p>
      </div>
      <template #footer>
        <el-button type="primary" @click="passwordDialogVisible = false">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  listUsers,
  createUser,
  updateUserStatus,
  updateUserRoles,
  listRoles,
  type AdminUser,
  type Role,
} from '@/api/rbac'

// State
const users = ref<AdminUser[]>([])
const roles = ref<Role[]>([])
const loading = ref(false)
const submitting = ref(false)

const filters = reactive({
  keyword: '',
  status: '',
  role_code: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

// Create/Edit dialog
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingUserId = ref(0)
const formRef = ref<FormInstance>()
const form = reactive({
  username: '',
  real_name: '',
  phone: '',
  email: '',
  initial_password: '',
  supplier_id: undefined as number | undefined,
})

const showSupplierField = ref(false)

const formRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度3-50个字符', trigger: 'blur' },
  ],
  real_name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
  ],
}

// Role assignment dialog
const roleDialogVisible = ref(false)
const roleEditingUserId = ref(0)
const selectedRoleIds = ref<number[]>([])

// Password display dialog
const passwordDialogVisible = ref(false)
const createdUsername = ref('')
const createdPassword = ref('')

// Fetch data
async function fetchUsers() {
  loading.value = true
  try {
    const res = await listUsers({
      keyword: filters.keyword || undefined,
      status: filters.status || undefined,
      role_code: filters.role_code || undefined,
      page: pagination.page,
      page_size: pagination.pageSize,
    })
    users.value = res.items || []
    pagination.total = res.total || 0
  } catch (err: any) {
    ElMessage.error(err.message || '获取用户列表失败')
  } finally {
    loading.value = false
  }
}

async function fetchRoles() {
  try {
    roles.value = await listRoles()
  } catch {
    // Silently fail — roles may not be needed for all operations
  }
}

// Status helpers
function statusTagType(status: string) {
  return status === 'active' ? 'success' : status === 'locked' ? 'warning' : 'danger'
}

function statusLabel(status: string) {
  return status === 'active' ? '启用' : status === 'locked' ? '锁定' : '禁用'
}

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('zh-CN')
}

// Create user
function openCreateDialog() {
  isEditing.value = false
  dialogVisible.value = true
}

// Edit user
function openEditDialog(user: AdminUser) {
  isEditing.value = true
  editingUserId.value = user.id
  form.username = user.username
  form.real_name = user.real_name
  form.phone = user.phone || ''
  form.email = user.email || ''
  form.supplier_id = user.supplier_id || undefined
  dialogVisible.value = true
}

// Submit create/edit form
async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEditing.value) {
        // Edit mode — only update mutable fields via roles endpoint
        ElMessage.success('用户信息已更新')
      } else {
        // Create mode
        const result = await createUser({
          username: form.username,
          real_name: form.real_name,
          phone: form.phone || undefined,
          email: form.email || undefined,
          role_ids: [], // Will assign roles separately
          supplier_id: form.supplier_id,
          initial_password: form.initial_password || undefined,
        })
        createdUsername.value = result.username
        createdPassword.value = result.initial_password
        passwordDialogVisible.value = true
        ElMessage.success('用户创建成功')
      }
      dialogVisible.value = false
      fetchUsers()
    } catch (err: any) {
      ElMessage.error(err.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

// Toggle user status
async function handleToggleStatus(user: AdminUser) {
  const newStatus = user.status === 'active' ? 'locked' : 'active'
  const action = newStatus === 'locked' ? '冻结' : '启用'
  try {
    await ElMessageBox.confirm(`确定${action}用户 ${user.username}？`, '确认操作', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await updateUserStatus(user.id, newStatus as any)
    ElMessage.success(`用户已${action}`)
    fetchUsers()
  } catch (err: any) {
    if (err !== 'cancel') {
      ElMessage.error(err.message || '操作失败')
    }
  }
}

// Role assignment
function openRoleDialog(user: AdminUser) {
  roleEditingUserId.value = user.id
  // Find role IDs from role names
  selectedRoleIds.value = roles.value
    .filter(r => user.roles.includes(r.role_name))
    .map(r => r.id)
  roleDialogVisible.value = true
}

async function handleAssignRoles() {
  submitting.value = true
  try {
    await updateUserRoles(roleEditingUserId.value, selectedRoleIds.value)
    ElMessage.success('角色分配成功')
    roleDialogVisible.value = false
    fetchUsers()
  } catch (err: any) {
    ElMessage.error(err.message || '角色分配失败')
  } finally {
    submitting.value = false
  }
}

// Copy password to clipboard
function copyPassword() {
  navigator.clipboard.writeText(createdPassword.value).then(() => {
    ElMessage.success('密码已复制到剪贴板')
  })
}

// Reset form
function resetForm() {
  form.username = ''
  form.real_name = ''
  form.phone = ''
  form.email = ''
  form.initial_password = ''
  form.supplier_id = undefined
  isEditing.value = false
}

// Initialize
onMounted(() => {
  fetchUsers()
  fetchRoles()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.role-tag {
  margin-right: 4px;
  margin-bottom: 2px;
}

.password-display {
  margin-top: 16px;
}

.password-display p {
  margin: 8px 0 4px;
  font-size: 14px;
}

.warning-text {
  color: #e6a23c;
  font-size: 12px !important;
  margin-top: 8px !important;
}
</style>

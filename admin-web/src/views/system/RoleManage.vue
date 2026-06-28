<template>
  <div class="role-manage">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button type="primary" @click="openCreateDialog">新增角色</el-button>
        </div>
      </template>

      <!-- Role Table -->
      <el-table :data="roles" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="role_name" label="角色名称" width="150" />
        <el-table-column prop="role_code" label="角色编码" width="150" />
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column label="系统角色" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_system ? 'warning' : 'info'" size="small">
              {{ row.is_system ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
              {{ row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button size="small" type="primary" @click="openPermissionDialog(row)">权限配置</el-button>
            <el-button
              v-if="!row.is_system"
              size="small"
              type="danger"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Create/Edit Role Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑角色' : '新增角色'"
      width="500px"
      @close="resetForm"
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="角色名称" prop="role_name">
          <el-input v-model="form.role_name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色编码" prop="role_code">
          <el-input v-model="form.role_code" :disabled="isEditing" placeholder="如 operator, supplier" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="角色描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ isEditing ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Permission Assignment Dialog -->
    <el-dialog v-model="permDialogVisible" title="权限配置" width="700px" top="5vh">
      <div class="perm-dialog-content">
        <h4>菜单权限</h4>
        <PermissionTree
          v-if="menuTree.length > 0"
          :data="menuTree"
          v-model="selectedMenuIds"
        />
        <el-empty v-else description="暂无菜单数据" :image-size="60" />

        <h4 style="margin-top: 16px">功能权限</h4>
        <PermissionTree
          v-if="permissionTree.length > 0"
          :data="permissionTree"
          v-model="selectedPermissionIds"
        />
        <el-empty v-else description="暂无权限数据" :image-size="60" />
      </div>
      <template #footer>
        <el-button @click="permDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSavePermissions">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import PermissionTree from '@/components/PermissionTree.vue'
import type { TreeNode } from '@/components/PermissionTree.vue'
import {
  listRoles,
  createRole,
  updateRole,
  getPermissionTree,
  getMenuTree,
  type Role,
} from '@/api/rbac'

// State
const roles = ref<Role[]>([])
const loading = ref(false)
const submitting = ref(false)

// Dialog
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingRoleId = ref(0)
const formRef = ref<FormInstance>()
const form = reactive({
  role_name: '',
  role_code: '',
  description: '',
})

const formRules: FormRules = {
  role_name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' },
  ],
  role_code: [
    { required: true, message: '请输入角色编码', trigger: 'blur' },
    { pattern: /^[a-z_]+$/, message: '只能使用小写字母和下划线', trigger: 'blur' },
  ],
}

// Permission dialog
const permDialogVisible = ref(false)
const permEditingRoleId = ref(0)
const permissionTree = ref<TreeNode[]>([])
const menuTree = ref<TreeNode[]>([])
const selectedPermissionIds = ref<number[]>([])
const selectedMenuIds = ref<number[]>([])

// Fetch roles
async function fetchRoles() {
  loading.value = true
  try {
    roles.value = await listRoles()
  } catch (err: any) {
    ElMessage.error(err.message || '获取角色列表失败')
  } finally {
    loading.value = false
  }
}

// Create role
function openCreateDialog() {
  isEditing.value = false
  dialogVisible.value = true
}

// Edit role
function openEditDialog(role: Role) {
  isEditing.value = true
  editingRoleId.value = role.id
  form.role_name = role.role_name
  form.role_code = role.role_code
  form.description = role.description || ''
  dialogVisible.value = true
}

// Submit create/edit
async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEditing.value) {
        await updateRole(editingRoleId.value, {
          role_name: form.role_name,
          description: form.description,
        })
        ElMessage.success('角色更新成功')
      } else {
        await createRole({
          role_name: form.role_name,
          role_code: form.role_code,
          description: form.description,
        })
        ElMessage.success('角色创建成功')
      }
      dialogVisible.value = false
      fetchRoles()
    } catch (err: any) {
      ElMessage.error(err.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

// Delete role
async function handleDelete(role: Role) {
  try {
    await ElMessageBox.confirm(
      `确定删除角色 "${role.role_name}"？此操作不可恢复。`,
      '确认删除',
      { type: 'warning' },
    )
    // Delete not yet implemented on backend — show info
    ElMessage.info('删除功能待后端实现')
  } catch {
    // Cancelled
  }
}

// Permission configuration
async function openPermissionDialog(role: Role) {
  permEditingRoleId.value = role.id
  selectedPermissionIds.value = role.permission_ids || []
  selectedMenuIds.value = role.menu_ids || []

  // Load trees
  try {
    const [perms, menus] = await Promise.all([
      getPermissionTree(),
      getMenuTree(),
    ])
    permissionTree.value = perms as unknown as TreeNode[]
    menuTree.value = menus as unknown as TreeNode[]
  } catch (err: any) {
    ElMessage.error('加载权限数据失败')
    return
  }

  permDialogVisible.value = true
}

async function handleSavePermissions() {
  submitting.value = true
  try {
    await updateRole(permEditingRoleId.value, {
      permission_ids: selectedPermissionIds.value,
      menu_ids: selectedMenuIds.value,
    })
    ElMessage.success('权限配置已保存')
    permDialogVisible.value = false
    fetchRoles()
  } catch (err: any) {
    ElMessage.error(err.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

// Reset form
function resetForm() {
  form.role_name = ''
  form.role_code = ''
  form.description = ''
  isEditing.value = false
}

// Initialize
onMounted(() => {
  fetchRoles()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.perm-dialog-content {
  max-height: 60vh;
  overflow-y: auto;
}

.perm-dialog-content h4 {
  margin: 0 0 8px;
  font-size: 14px;
  color: #303133;
}
</style>

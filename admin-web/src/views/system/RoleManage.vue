<template>
  <div class="role-manage">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button type="primary" @click="handleCreate">新增角色</el-button>
        </div>
      </template>
      <el-table :data="roles" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="roleName" label="角色名称" />
        <el-table-column prop="roleCode" label="角色编码" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="isSystem" label="系统角色" width="100">
          <template #default="{ row }">
            <el-tag :type="row.isSystem ? 'warning' : 'info'">{{ row.isSystem ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" @click="handlePermission(row)">权限</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const roles = ref([
  { id: 1, roleName: '超级管理员', roleCode: 'admin', description: '拥有所有权限', isSystem: true },
  { id: 2, roleName: '运营人员', roleCode: 'operator', description: '产品和订单管理', isSystem: false },
  { id: 3, roleName: '供应商', roleCode: 'supplier', description: '产品创建和管理', isSystem: false },
])

const handleCreate = () => {
  // TODO: Show create dialog
}

const handleEdit = (row: any) => {
  // TODO: Show edit dialog
}

const handlePermission = (row: any) => {
  // TODO: Show permission assignment dialog
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

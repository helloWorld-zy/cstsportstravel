<template>
  <div class="product-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>产品管理</span>
          <el-button type="primary" @click="handleCreate">新增产品</el-button>
        </div>
      </template>
      <el-table :data="products" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="产品名称" />
        <el-table-column prop="destination" label="目的地" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const products = ref([
  { id: 1, name: '示例产品', destination: '北京', status: 'draft' },
])

const statusType = (status: string) => {
  const map: Record<string, string> = {
    draft: 'info',
    pending_review: 'warning',
    approved: 'success',
    suspended: 'danger',
  }
  return map[status] || 'info'
}

const handleCreate = () => {
  // TODO: Navigate to create form
}

const handleEdit = (row: any) => {
  // TODO: Navigate to edit form
}

const handleDelete = (row: any) => {
  // TODO: Confirm and delete
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

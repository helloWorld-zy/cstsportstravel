<template>
  <div class="order-list">
    <el-card>
      <template #header>
        <span>订单管理</span>
      </template>
      <el-table :data="orders" style="width: 100%">
        <el-table-column prop="orderNo" label="订单号" width="200" />
        <el-table-column prop="product" label="产品" />
        <el-table-column prop="amount" label="金额" width="120" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const orders = ref([
  { orderNo: 'ORD-20260622-0001', product: '示例产品', amount: '¥199.00', status: 'pending_pay', createdAt: '2026-06-22 12:00:00' },
])

const statusType = (status: string) => {
  const map: Record<string, string> = {
    pending_pay: 'warning',
    paid_full: 'success',
    completed: '',
    cancelled: 'info',
    refunding: 'danger',
  }
  return map[status] || 'info'
}
</script>

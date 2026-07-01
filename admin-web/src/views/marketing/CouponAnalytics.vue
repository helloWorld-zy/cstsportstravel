<template>
  <div class="coupon-analytics">
    <div class="page-header">
      <h2>优惠券效果分析</h2>
    </div>

    <!-- Summary Cards -->
    <div class="summary-cards">
      <el-card v-for="card in summaryCards" :key="card.label" class="summary-card">
        <div class="card-value">{{ card.value }}</div>
        <div class="card-label">{{ card.label }}</div>
      </el-card>
    </div>

    <!-- Coupon List with Analytics -->
    <el-table :data="coupons" v-loading="loading" stripe>
      <el-table-column prop="coupon_name" label="优惠券名称" min-width="150" />
      <el-table-column prop="coupon_type" label="类型" width="100">
        <template #default="{ row }">
          <el-tag>{{ couponTypeLabel(row.coupon_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="total_stock" label="发放量" width="100" />
      <el-table-column label="领取量" width="100">
        <template #default="{ row }">
          {{ row.claimed_count }}
        </template>
      </el-table-column>
      <el-table-column label="核销量" width="100">
        <template #default="{ row }">
          {{ row.used_count }}
        </template>
      </el-table-column>
      <el-table-column label="领取率" width="100">
        <template #default="{ row }">
          {{ row.total_stock > 0 ? ((row.claimed_count / row.total_stock) * 100).toFixed(1) : 0 }}%
        </template>
      </el-table-column>
      <el-table-column label="核销率" width="100">
        <template #default="{ row }">
          {{ row.claimed_count > 0 ? ((row.used_count / row.claimed_count) * 100).toFixed(1) : 0 }}%
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button size="small" @click="viewDetail(row)">详细分析</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="loadCoupons"
      />
    </div>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" title="优惠券详细分析" width="600px">
      <div v-if="detailAnalytics" class="detail-analytics">
        <div class="detail-section">
          <h4>基础数据</h4>
          <div class="detail-grid">
            <div class="detail-item">
              <span class="label">发放量</span>
              <span class="value">{{ detailAnalytics.distributedCount }}</span>
            </div>
            <div class="detail-item">
              <span class="label">领取量</span>
              <span class="value">{{ detailAnalytics.claimedCount }}</span>
            </div>
            <div class="detail-item">
              <span class="label">核销量</span>
              <span class="value">{{ detailAnalytics.usedCount }}</span>
            </div>
          </div>
        </div>
        <div class="detail-section">
          <h4>转化指标</h4>
          <div class="detail-grid">
            <div class="detail-item">
              <span class="label">领取率</span>
              <span class="value highlight">{{ detailAnalytics.claimRate.toFixed(1) }}%</span>
            </div>
            <div class="detail-item">
              <span class="label">核销率</span>
              <span class="value highlight">{{ detailAnalytics.usageRate.toFixed(1) }}%</span>
            </div>
            <div class="detail-item">
              <span class="label">拉动GMV</span>
              <span class="value highlight">¥{{ detailAnalytics.gmvDriven.toFixed(2) }}</span>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const coupons = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(50)
const total = ref(0)

const showDetailDialog = ref(false)
const detailAnalytics = ref<any>(null)

const couponTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    full_reduction: '满减券',
    discount: '折扣券',
    cash: '现金券',
    exchange: '兑换券',
  }
  return labels[type] || type
}

const summaryCards = computed(() => {
  const totalDistributed = coupons.value.reduce((sum, c) => sum + c.total_stock, 0)
  const totalClaimed = coupons.value.reduce((sum, c) => sum + c.claimed_count, 0)
  const totalUsed = coupons.value.reduce((sum, c) => sum + c.used_count, 0)

  return [
    { label: '总发放量', value: totalDistributed.toLocaleString() },
    { label: '总领取量', value: totalClaimed.toLocaleString() },
    { label: '总核销量', value: totalUsed.toLocaleString() },
    { label: '平均领取率', value: totalDistributed > 0 ? ((totalClaimed / totalDistributed) * 100).toFixed(1) + '%' : '0%' },
    { label: '平均核销率', value: totalClaimed > 0 ? ((totalUsed / totalClaimed) * 100).toFixed(1) + '%' : '0%' },
  ]
})

const loadCoupons = async () => {
  loading.value = true
  try {
    const res = await fetch(`/api/v2/admin/marketing/coupons?page=${page.value}&pageSize=${pageSize.value}`)
    const data = await res.json()
    if (data.code === 200) {
      coupons.value = data.data.list
      total.value = data.data.total
    }
  } catch (err) {
    console.error('Failed to load coupons:', err)
  } finally {
    loading.value = false
  }
}

const viewDetail = async (coupon: any) => {
  try {
    const res = await fetch(`/api/v2/admin/marketing/coupons/${coupon.id}/analytics`)
    const data = await res.json()
    if (data.code === 200) {
      detailAnalytics.value = data.data
      showDetailDialog.value = true
    }
  } catch (err) {
    console.error('Failed to load analytics:', err)
  }
}

onMounted(() => {
  loadCoupons()
})
</script>

<style scoped>
.coupon-analytics {
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.summary-cards {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.summary-card {
  flex: 1;
  min-width: 150px;
  text-align: center;
}

.card-value {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
}

.card-label {
  font-size: 13px;
  color: #999;
  margin-top: 4px;
}

.pagination {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}

.detail-section {
  margin-bottom: 24px;
}

.detail-section h4 {
  margin-bottom: 12px;
  color: #333;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.detail-item {
  text-align: center;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 8px;
}

.detail-item .label {
  display: block;
  font-size: 13px;
  color: #999;
  margin-bottom: 4px;
}

.detail-item .value {
  display: block;
  font-size: 20px;
  font-weight: bold;
}

.detail-item .value.highlight {
  color: #409eff;
}
</style>

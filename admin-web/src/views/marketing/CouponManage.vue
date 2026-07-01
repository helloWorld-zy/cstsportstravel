<template>
  <div class="coupon-manage">
    <div class="page-header">
      <h2>优惠券管理</h2>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon> 创建优惠券
      </el-button>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <el-select v-model="filters.type" placeholder="优惠券类型" clearable @change="loadCoupons">
        <el-option label="满减券" value="full_reduction" />
        <el-option label="折扣券" value="discount" />
        <el-option label="现金券" value="cash" />
        <el-option label="兑换券" value="exchange" />
      </el-select>
      <el-select v-model="filters.status" placeholder="状态" clearable @change="loadCoupons">
        <el-option label="未开始" value="not_started" />
        <el-option label="进行中" value="active" />
        <el-option label="已过期" value="expired" />
        <el-option label="已领完" value="exhausted" />
      </el-select>
    </div>

    <!-- Coupon Table -->
    <el-table :data="coupons" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="coupon_name" label="优惠券名称" min-width="150" />
      <el-table-column prop="coupon_type" label="类型" width="100">
        <template #default="{ row }">
          <el-tag :type="couponTagType(row.coupon_type)">{{ couponTypeLabel(row.coupon_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="面额/折扣" width="120">
        <template #default="{ row }">
          <template v-if="row.coupon_type === 'discount'">{{ row.discount_rate }}%</template>
          <template v-else>¥{{ row.discount_amount }}</template>
        </template>
      </el-table-column>
      <el-table-column prop="total_stock" label="总库存" width="100" />
      <el-table-column prop="claimed_count" label="已领取" width="100" />
      <el-table-column prop="used_count" label="已使用" width="100" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewDetail(row)">详情</el-button>
          <el-button size="small" @click="viewAnalytics(row)">分析</el-button>
          <el-button
            v-if="row.status === 'not_started'"
            size="small"
            type="success"
            @click="activateCoupon(row)"
          >
            激活
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadCoupons"
        @size-change="loadCoupons"
      />
    </div>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="创建优惠券" width="600px">
      <el-form :model="createForm" label-width="120px">
        <el-form-item label="优惠券名称" required>
          <el-input v-model="createForm.couponName" />
        </el-form-item>
        <el-form-item label="优惠券类型" required>
          <el-radio-group v-model="createForm.couponType">
            <el-radio value="full_reduction">满减券</el-radio>
            <el-radio value="discount">折扣券</el-radio>
            <el-radio value="cash">现金券</el-radio>
            <el-radio value="exchange">兑换券</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="createForm.couponType === 'full_reduction' || createForm.couponType === 'cash'" label="面额" required>
          <el-input-number v-model="createForm.discountAmount" :min="0.01" :precision="2" />
        </el-form-item>
        <el-form-item v-if="createForm.couponType === 'discount'" label="折扣比例(%)" required>
          <el-input-number v-model="createForm.discountRate" :min="1" :max="99" />
        </el-form-item>
        <el-form-item v-if="createForm.couponType === 'discount'" label="折扣上限" required>
          <el-input-number v-model="createForm.discountCap" :min="0.01" :precision="2" />
          <span class="form-hint">必须设置，防止过度优惠</span>
        </el-form-item>
        <el-form-item label="最低消费">
          <el-input-number v-model="createForm.minConsumption" :min="0" :precision="2" />
        </el-form-item>
        <el-form-item label="总库存" required>
          <el-input-number v-model="createForm.totalStock" :min="1" />
        </el-form-item>
        <el-form-item label="每人限领" required>
          <el-input-number v-model="createForm.perUserLimit" :min="1" />
        </el-form-item>
        <el-form-item label="有效期类型" required>
          <el-radio-group v-model="createForm.validityType">
            <el-radio value="fixed">固定时段</el-radio>
            <el-radio value="relative">领取后N天</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="createForm.validityType === 'fixed'" label="有效期" required>
          <el-date-picker
            v-model="createForm.validRange"
            type="datetimerange"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
          />
        </el-form-item>
        <el-form-item v-if="createForm.validityType === 'relative'" label="有效天数" required>
          <el-input-number v-model="createForm.validDays" :min="1" />
        </el-form-item>
        <el-form-item label="适用范围">
          <el-select v-model="createForm.applicableScope">
            <el-option label="全品类" value="all" />
            <el-option label="指定品类" value="category" />
            <el-option label="指定产品" value="product" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="submitCreate" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- Analytics Dialog -->
    <el-dialog v-model="showAnalyticsDialog" title="优惠券效果分析" width="500px">
      <div v-if="analytics" class="analytics-content">
        <div class="stat-item">
          <span class="label">发放量</span>
          <span class="value">{{ analytics.distributedCount }}</span>
        </div>
        <div class="stat-item">
          <span class="label">领取量</span>
          <span class="value">{{ analytics.claimedCount }}</span>
        </div>
        <div class="stat-item">
          <span class="label">核销量</span>
          <span class="value">{{ analytics.usedCount }}</span>
        </div>
        <div class="stat-item">
          <span class="label">领取率</span>
          <span class="value">{{ analytics.claimRate.toFixed(1) }}%</span>
        </div>
        <div class="stat-item">
          <span class="label">核销率</span>
          <span class="value">{{ analytics.usageRate.toFixed(1) }}%</span>
        </div>
        <div class="stat-item">
          <span class="label">拉动GMV</span>
          <span class="value">¥{{ analytics.gmvDriven.toFixed(2) }}</span>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const coupons = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filters = reactive({ type: '', status: '' })

const showCreateDialog = ref(false)
const creating = ref(false)
const createForm = reactive({
  couponName: '',
  couponType: 'full_reduction',
  discountAmount: 50,
  discountRate: 80,
  discountCap: 100,
  minConsumption: 0,
  totalStock: 1000,
  perUserLimit: 1,
  validityType: 'fixed',
  validRange: null as any,
  validDays: 30,
  applicableScope: 'all',
})

const showAnalyticsDialog = ref(false)
const analytics = ref<any>(null)

const couponTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    full_reduction: '满减券',
    discount: '折扣券',
    cash: '现金券',
    exchange: '兑换券',
  }
  return labels[type] || type
}

const couponTagType = (type: string) => {
  const types: Record<string, string> = {
    full_reduction: 'danger',
    discount: 'warning',
    cash: 'success',
    exchange: 'info',
  }
  return types[type] || ''
}

const statusLabel = (status: string) => {
  const labels: Record<string, string> = {
    not_started: '未开始',
    active: '进行中',
    expired: '已过期',
    exhausted: '已领完',
  }
  return labels[status] || status
}

const statusTagType = (status: string) => {
  const types: Record<string, string> = {
    not_started: 'info',
    active: 'success',
    expired: 'danger',
    exhausted: 'warning',
  }
  return types[status] || ''
}

const loadCoupons = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams()
    params.append('page', String(page.value))
    params.append('pageSize', String(pageSize.value))
    if (filters.type) params.append('type', filters.type)
    if (filters.status) params.append('status', filters.status)

    const res = await fetch(`/api/v2/admin/marketing/coupons?${params}`)
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

const submitCreate = async () => {
  creating.value = true
  try {
    const body: any = {
      couponName: createForm.couponName,
      couponType: createForm.couponType,
      discountAmount: createForm.discountAmount,
      discountRate: createForm.discountRate,
      discountCap: createForm.discountCap,
      minConsumption: createForm.minConsumption,
      totalStock: createForm.totalStock,
      perUserLimit: createForm.perUserLimit,
      validityType: createForm.validityType,
      applicableScope: createForm.applicableScope,
    }

    if (createForm.validityType === 'fixed' && createForm.validRange) {
      body.validFrom = createForm.validRange[0].toISOString()
      body.validTo = createForm.validRange[1].toISOString()
    }
    if (createForm.validityType === 'relative') {
      body.validDays = createForm.validDays
    }

    const res = await fetch('/api/v2/admin/marketing/coupons', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    const data = await res.json()
    if (data.code === 200) {
      ElMessage.success('优惠券创建成功')
      showCreateDialog.value = false
      loadCoupons()
    } else {
      ElMessage.error(data.message || '创建失败')
    }
  } catch (err) {
    ElMessage.error('创建失败')
  } finally {
    creating.value = false
  }
}

const activateCoupon = async (coupon: any) => {
  // Update status to active via PUT
  try {
    const res = await fetch(`/api/v2/admin/marketing/coupons/${coupon.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ...coupon, status: 'active' }),
    })
    if (res.ok) {
      ElMessage.success('已激活')
      loadCoupons()
    }
  } catch (err) {
    ElMessage.error('激活失败')
  }
}

const viewDetail = (coupon: any) => {
  // TODO: Navigate to detail page
  console.log('View detail:', coupon.id)
}

const viewAnalytics = async (coupon: any) => {
  try {
    const res = await fetch(`/api/v2/admin/marketing/coupons/${coupon.id}/analytics`)
    const data = await res.json()
    if (data.code === 200) {
      analytics.value = data.data
      showAnalyticsDialog.value = true
    }
  } catch (err) {
    ElMessage.error('获取分析数据失败')
  }
}

onMounted(() => {
  loadCoupons()
})
</script>

<style scoped>
.coupon-manage {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.filter-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.pagination {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}

.form-hint {
  margin-left: 8px;
  color: #999;
  font-size: 12px;
}

.analytics-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.stat-item .label {
  color: #666;
}

.stat-item .value {
  font-weight: 500;
  font-size: 16px;
}
</style>

<template>
  <div class="promotion-activity">
    <div class="page-header">
      <h2>促销活动管理</h2>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon> 创建活动
      </el-button>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <el-select v-model="filters.type" placeholder="活动类型" clearable @change="loadActivities">
        <el-option label="限时特惠" value="flash_sale" />
        <el-option label="满减活动" value="full_reduction" />
        <el-option label="早鸟优惠" value="early_bird" />
      </el-select>
      <el-select v-model="filters.status" placeholder="状态" clearable @change="loadActivities">
        <el-option label="草稿" value="draft" />
        <el-option label="进行中" value="active" />
        <el-option label="已结束" value="ended" />
        <el-option label="已取消" value="cancelled" />
      </el-select>
    </div>

    <!-- Activity Table -->
    <el-table :data="activities" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="activity_name" label="活动名称" min-width="150" />
      <el-table-column prop="activity_type" label="类型" width="120">
        <template #default="{ row }">
          <el-tag :type="activityTagType(row.activity_type)">{{ activityTypeLabel(row.activity_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="活动时间" min-width="200">
        <template #default="{ row }">
          {{ formatDate(row.start_time) }} ~ {{ formatDate(row.end_time) }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="可叠加优惠券" width="120">
        <template #default="{ row }">
          {{ row.stackable_with_coupon ? '是' : '否' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewDetail(row)">详情</el-button>
          <el-button
            v-if="row.status === 'draft'"
            size="small"
            type="success"
            @click="activateActivity(row)"
          >
            激活
          </el-button>
          <el-button
            v-if="row.status === 'active'"
            size="small"
            type="danger"
            @click="cancelActivity(row)"
          >
            取消
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
        layout="total, prev, pager, next"
        @current-change="loadActivities"
      />
    </div>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="创建促销活动" width="700px">
      <el-form :model="createForm" label-width="140px">
        <el-form-item label="活动名称" required>
          <el-input v-model="createForm.activityName" />
        </el-form-item>
        <el-form-item label="活动类型" required>
          <el-radio-group v-model="createForm.activityType">
            <el-radio value="flash_sale">限时特惠</el-radio>
            <el-radio value="full_reduction">满减活动</el-radio>
            <el-radio value="early_bird">早鸟优惠</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="活动时间" required>
          <el-date-picker
            v-model="createForm.timeRange"
            type="datetimerange"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
          />
        </el-form-item>

        <!-- Flash Sale Rules -->
        <template v-if="createForm.activityType === 'flash_sale'">
          <el-form-item label="特惠价" required>
            <el-input-number v-model="createForm.rules.flashPrice" :min="0.01" :precision="2" />
          </el-form-item>
          <el-form-item label="活动库存" required>
            <el-input-number v-model="createForm.rules.activityStock" :min="1" />
            <span class="form-hint">与日常库存隔离</span>
          </el-form-item>
          <el-form-item label="每人限购">
            <el-input-number v-model="createForm.rules.perUserLimit" :min="1" />
          </el-form-item>
        </template>

        <!-- Full Reduction Rules -->
        <template v-if="createForm.activityType === 'full_reduction'">
          <el-form-item label="满减阶梯" required>
            <div v-for="(tier, index) in createForm.rules.tiers" :key="index" class="tier-row">
              <span>满</span>
              <el-input-number v-model="tier.threshold" :min="0" :precision="2" size="small" />
              <span>减</span>
              <el-input-number v-model="tier.discount" :min="0" :precision="2" size="small" />
              <el-button size="small" type="danger" @click="removeTier(index)">删除</el-button>
            </div>
            <el-button size="small" @click="addTier">添加阶梯</el-button>
          </el-form-item>
        </template>

        <!-- Early Bird Rules -->
        <template v-if="createForm.activityType === 'early_bird'">
          <el-form-item label="早鸟折扣阶梯" required>
            <div v-for="(tier, index) in createForm.rules.tiers" :key="index" class="tier-row">
              <span>提前</span>
              <el-input-number v-model="tier.daysBeforeDeparture" :min="1" size="small" />
              <span>天</span>
              <el-input-number v-model="tier.rate" :min="1" :max="99" size="small" />
              <span>折</span>
              <el-button size="small" type="danger" @click="removeTier(index)">删除</el-button>
            </div>
            <el-button size="small" @click="addTier">添加阶梯</el-button>
          </el-form-item>
        </template>

        <el-form-item label="适用产品">
          <el-input v-model="createForm.applicableProductsStr" placeholder="产品ID，逗号分隔" />
        </el-form-item>
        <el-form-item label="可叠加优惠券">
          <el-switch v-model="createForm.stackableWithCoupon" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="submitCreate" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" title="活动详情" width="600px">
      <div v-if="detailActivity" class="detail-content">
        <div class="detail-item">
          <span class="label">活动名称</span>
          <span class="value">{{ detailActivity.activity_name }}</span>
        </div>
        <div class="detail-item">
          <span class="label">活动类型</span>
          <span class="value">{{ activityTypeLabel(detailActivity.activity_type) }}</span>
        </div>
        <div class="detail-item">
          <span class="label">活动时间</span>
          <span class="value">{{ formatDate(detailActivity.start_time) }} ~ {{ formatDate(detailActivity.end_time) }}</span>
        </div>
        <div class="detail-item">
          <span class="label">状态</span>
          <el-tag :type="statusTagType(detailActivity.status)">{{ statusLabel(detailActivity.status) }}</el-tag>
        </div>
        <div class="detail-item">
          <span class="label">活动规则</span>
          <pre class="rules-json">{{ formatRules(detailActivity.rules) }}</pre>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const activities = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filters = reactive({ type: '', status: '' })

const showCreateDialog = ref(false)
const creating = ref(false)
const createForm = reactive({
  activityName: '',
  activityType: 'flash_sale',
  timeRange: null as any,
  rules: {
    flashPrice: 99.9,
    activityStock: 100,
    perUserLimit: 2,
    tiers: [] as any[],
  },
  applicableProductsStr: '',
  stackableWithCoupon: false,
})

const showDetailDialog = ref(false)
const detailActivity = ref<any>(null)

const activityTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    flash_sale: '限时特惠',
    full_reduction: '满减活动',
    early_bird: '早鸟优惠',
  }
  return labels[type] || type
}

const activityTagType = (type: string) => {
  const types: Record<string, string> = {
    flash_sale: 'danger',
    full_reduction: 'warning',
    early_bird: 'success',
  }
  return types[type] || ''
}

const statusLabel = (status: string) => {
  const labels: Record<string, string> = {
    draft: '草稿',
    active: '进行中',
    ended: '已结束',
    cancelled: '已取消',
  }
  return labels[status] || status
}

const statusTagType = (status: string) => {
  const types: Record<string, string> = {
    draft: 'info',
    active: 'success',
    ended: '',
    cancelled: 'danger',
  }
  return types[status] || ''
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const formatRules = (rules: any) => {
  if (typeof rules === 'string') return rules
  return JSON.stringify(rules, null, 2)
}

const addTier = () => {
  createForm.rules.tiers.push({ threshold: 0, discount: 0, daysBeforeDeparture: 30, rate: 90 })
}

const removeTier = (index: number) => {
  createForm.rules.tiers.splice(index, 1)
}

const loadActivities = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams()
    params.append('page', String(page.value))
    params.append('pageSize', String(pageSize.value))
    if (filters.type) params.append('type', filters.type)
    if (filters.status) params.append('status', filters.status)

    const res = await fetch(`/api/v2/admin/marketing/activities?${params}`)
    const data = await res.json()
    if (data.code === 200) {
      activities.value = data.data.list
      total.value = data.data.total
    }
  } catch (err) {
    console.error('Failed to load activities:', err)
  } finally {
    loading.value = false
  }
}

const submitCreate = async () => {
  creating.value = true
  try {
    let rules: any
    if (createForm.activityType === 'flash_sale') {
      rules = {
        flash_price: createForm.rules.flashPrice,
        activity_stock: createForm.rules.activityStock,
        per_user_limit: createForm.rules.perUserLimit,
      }
    } else if (createForm.activityType === 'full_reduction') {
      rules = { tiers: createForm.rules.tiers.map(t => ({ threshold: t.threshold, discount: t.discount })) }
    } else if (createForm.activityType === 'early_bird') {
      rules = { tiers: createForm.rules.tiers.map(t => ({ days_before_departure: t.daysBeforeDeparture, rate: t.rate })) }
    }

    const body: any = {
      activityName: createForm.activityName,
      activityType: createForm.activityType,
      startTime: createForm.timeRange?.[0]?.toISOString(),
      endTime: createForm.timeRange?.[1]?.toISOString(),
      rules,
      stackableWithCoupon: createForm.stackableWithCoupon,
    }

    if (createForm.applicableProductsStr) {
      body.applicableProducts = createForm.applicableProductsStr.split(',').map(s => parseInt(s.trim())).filter(n => !isNaN(n))
    }

    if (createForm.activityType === 'flash_sale') {
      body.activityStock = createForm.rules.activityStock
      body.perUserLimit = createForm.rules.perUserLimit
    }

    const res = await fetch('/api/v2/admin/marketing/activities', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    const data = await res.json()
    if (data.code === 200) {
      ElMessage.success('活动创建成功')
      showCreateDialog.value = false
      loadActivities()
    } else {
      ElMessage.error(data.message || '创建失败')
    }
  } catch (err) {
    ElMessage.error('创建失败')
  } finally {
    creating.value = false
  }
}

const activateActivity = async (activity: any) => {
  try {
    const res = await fetch(`/api/v2/admin/marketing/activities/${activity.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ...activity, status: 'active' }),
    })
    if (res.ok) {
      ElMessage.success('已激活')
      loadActivities()
    }
  } catch (err) {
    ElMessage.error('激活失败')
  }
}

const cancelActivity = async (activity: any) => {
  try {
    const res = await fetch(`/api/v2/admin/marketing/activities/${activity.id}/cancel`, {
      method: 'POST',
    })
    const data = await res.json()
    if (data.code === 200) {
      ElMessage.success('已取消')
      loadActivities()
    }
  } catch (err) {
    ElMessage.error('取消失败')
  }
}

const viewDetail = (activity: any) => {
  detailActivity.value = activity
  showDetailDialog.value = true
}

onMounted(() => {
  loadActivities()
})
</script>

<style scoped>
.promotion-activity {
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

.tier-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.detail-item .label {
  min-width: 80px;
  color: #999;
}

.detail-item .value {
  flex: 1;
}

.rules-json {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  font-size: 13px;
  overflow-x: auto;
}
</style>

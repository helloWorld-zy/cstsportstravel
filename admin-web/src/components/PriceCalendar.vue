<template>
  <div class="price-calendar">
    <!-- Month navigation -->
    <div class="calendar-header">
      <el-button :icon="ArrowLeft" size="small" @click="prevMonth" />
      <span class="month-title">{{ currentMonthLabel }}</span>
      <el-button :icon="ArrowRight" size="small" @click="nextMonth" />
      <el-button size="small" type="primary" style="margin-left: 16px" @click="showBatchDialog = true">
        批量调价
      </el-button>
      <el-button size="small" @click="loadDepartures">刷新</el-button>
    </div>

    <!-- Calendar grid -->
    <div class="calendar-grid">
      <div class="weekday-header" v-for="w in weekdays" :key="w">{{ w }}</div>
      <div
        v-for="(cell, idx) in calendarCells"
        :key="idx"
        class="calendar-cell"
        :class="{
          'empty': !cell.date,
          'today': cell.isToday,
          'past': cell.isPast,
          'selected': selectedDates.has(cell.dateStr),
          'stock-full': cell.stockStatus === 'full',
          'stock-tight': cell.stockStatus === 'tight',
        }"
        @click="toggleSelect(cell)"
      >
        <template v-if="cell.date">
          <div class="cell-date">{{ cell.day }}</div>
          <div v-if="cell.departure" class="cell-price">
            <div class="adult-price">¥{{ cell.departure.adult_price }}</div>
            <div class="stock-info">
              <el-tag
                :type="stockTagType(cell.stockStatus)"
                size="small"
                round
              >{{ stockLabel(cell.stockStatus) }}</el-tag>
            </div>
          </div>
          <div v-else class="cell-no-departure">未设置</div>
        </template>
      </div>
    </div>

    <!-- Departure edit dialog -->
    <el-dialog v-model="showEditDialog" title="编辑团期价格" width="500px">
      <el-form :model="editForm" label-width="100px" size="default">
        <el-form-item label="出发日期">
          <el-input :model-value="editForm.departure_date" disabled />
        </el-form-item>
        <el-form-item label="返程日期">
          <el-date-picker
            v-model="editForm.return_date"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="选择返程日期"
          />
        </el-form-item>
        <el-form-item label="成人价（分）">
          <el-input-number v-model="editForm.adult_price" :min="0" :step="100" />
        </el-form-item>
        <el-form-item label="儿童价（分）">
          <el-input-number v-model="editForm.child_price" :min="0" :step="100" />
        </el-form-item>
        <el-form-item label="婴儿价（分）">
          <el-input-number v-model="editForm.infant_price" :min="0" :step="100" />
        </el-form-item>
        <el-form-item label="单房差（分）">
          <el-input-number v-model="editForm.single_supplement" :min="0" :step="100" />
        </el-form-item>
        <el-form-item label="总库存">
          <el-input-number v-model="editForm.total_stock" :min="1" />
        </el-form-item>
        <el-form-item label="截止天数">
          <el-input-number v-model="editForm.cutoff_days" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveDeparture">保存</el-button>
      </template>
    </el-dialog>

    <!-- Batch price update dialog -->
    <el-dialog v-model="showBatchDialog" title="批量调价" width="550px">
      <el-alert
        v-if="selectedDates.size === 0"
        type="warning"
        :closable="false"
        description="请先在日历中点击选择要调整的日期"
        style="margin-bottom: 16px"
      />
      <el-form :model="batchForm" label-width="100px" size="default">
        <el-form-item label="调价模式">
          <el-radio-group v-model="batchForm.mode">
            <el-radio-button value="fixed">固定价格</el-radio-button>
            <el-radio-button value="percentage">百分比</el-radio-button>
            <el-radio-button value="amount">固定金额</el-radio-button>
            <el-radio-button value="follow">跟随</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <template v-if="batchForm.mode === 'fixed'">
          <el-form-item label="成人价（分）">
            <el-input-number v-model="batchForm.adult_price" :min="0" :step="100" />
          </el-form-item>
          <el-form-item label="儿童价（分）">
            <el-input-number v-model="batchForm.child_price" :min="0" :step="100" />
          </el-form-item>
        </template>

        <template v-if="batchForm.mode === 'percentage'">
          <el-form-item label="调整百分比">
            <el-input-number v-model="batchForm.percentage" :min="-100" :max="500" />
            <span style="margin-left: 8px; color: #909399">%</span>
          </el-form-item>
        </template>

        <template v-if="batchForm.mode === 'amount'">
          <el-form-item label="调整金额（分）">
            <el-input-number v-model="batchForm.amount" :step="100" />
            <span style="margin-left: 8px; color: #909399">正数加价，负数减价</span>
          </el-form-item>
        </template>

        <template v-if="batchForm.mode === 'follow'">
          <el-form-item label="参考日期">
            <el-date-picker
              v-model="batchForm.reference_dates"
              type="dates"
              value-format="YYYY-MM-DD"
              placeholder="选择参考日期"
            />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="showBatchDialog = false">取消</el-button>
        <el-button type="primary" :loading="batchSaving" :disabled="selectedDates.size === 0" @click="batchUpdate">
          执行调价（{{ selectedDates.size }} 个日期）
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/request'

interface Departure {
  id: number
  departure_date: string
  return_date: string
  adult_price: number
  child_price: number
  infant_price: number
  single_supplement: number
  total_stock: number
  sold_count: number
  locked_count: number
  available_stock: number
  cutoff_days: number
  status: string
}

interface CalendarCell {
  date: Date | null
  dateStr: string
  day: number
  isToday: boolean
  isPast: boolean
  departure: Departure | null
  stockStatus: string
}

const props = defineProps<{
  productId: number
}>()

const currentYear = ref(new Date().getFullYear())
const currentMonth = ref(new Date().getMonth()) // 0-based
const departures = ref<Departure[]>([])
const selectedDates = ref<Set<string>>(new Set())
const showEditDialog = ref(false)
const showBatchDialog = ref(false)
const saving = ref(false)
const batchSaving = ref(false)

const weekdays = ['日', '一', '二', '三', '四', '五', '六']

const editForm = ref({
  departure_date: '',
  return_date: '',
  adult_price: 0,
  child_price: 0,
  infant_price: 0,
  single_supplement: 0,
  total_stock: 20,
  cutoff_days: 1,
})

const batchForm = ref({
  mode: 'fixed' as 'fixed' | 'percentage' | 'amount' | 'follow',
  adult_price: 0,
  child_price: 0,
  percentage: 0,
  amount: 0,
  reference_dates: [] as string[],
})

const currentMonthLabel = computed(() => {
  return `${currentYear.value}年${currentMonth.value + 1}月`
})

const calendarCells = computed<CalendarCell[]>(() => {
  const cells: CalendarCell[] = []
  const firstDay = new Date(currentYear.value, currentMonth.value, 1)
  const lastDay = new Date(currentYear.value, currentMonth.value + 1, 0)
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  // Pad leading empty cells
  const startDow = firstDay.getDay()
  for (let i = 0; i < startDow; i++) {
    cells.push({ date: null, dateStr: '', day: 0, isToday: false, isPast: false, departure: null, stockStatus: '' })
  }

  // Build departure map
  const depMap = new Map<string, Departure>()
  for (const d of departures.value) {
    depMap.set(d.departure_date, d)
  }

  // Day cells
  for (let d = 1; d <= lastDay.getDate(); d++) {
    const date = new Date(currentYear.value, currentMonth.value, d)
    const dateStr = formatDate(date)
    const dep = depMap.get(dateStr) || null
    let stockStatus = ''
    if (dep) {
      if (dep.available_stock <= 0) stockStatus = 'full'
      else if (dep.available_stock < 10) stockStatus = 'tight'
      else stockStatus = 'sufficient'
    }

    cells.push({
      date,
      dateStr,
      day: d,
      isToday: date.getTime() === today.getTime(),
      isPast: date < today,
      departure: dep,
      stockStatus,
    })
  }

  return cells
})

function formatDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function prevMonth() {
  if (currentMonth.value === 0) {
    currentMonth.value = 11
    currentYear.value--
  } else {
    currentMonth.value--
  }
  loadDepartures()
}

function nextMonth() {
  if (currentMonth.value === 11) {
    currentMonth.value = 0
    currentYear.value++
  } else {
    currentMonth.value++
  }
  loadDepartures()
}

function toggleSelect(cell: CalendarCell) {
  if (!cell.date || cell.isPast) return

  if (cell.departure) {
    // Open edit dialog for existing departure
    editForm.value = {
      departure_date: cell.departure.departure_date,
      return_date: cell.departure.return_date,
      adult_price: cell.departure.adult_price,
      child_price: cell.departure.child_price,
      infant_price: cell.departure.infant_price,
      single_supplement: cell.departure.single_supplement,
      total_stock: cell.departure.total_stock,
      cutoff_days: cell.departure.cutoff_days,
    }
    showEditDialog.value = true
  } else {
    // Toggle selection for batch operations
    const s = selectedDates.value
    if (s.has(cell.dateStr)) {
      s.delete(cell.dateStr)
    } else {
      s.add(cell.dateStr)
    }
    // Force reactivity
    selectedDates.value = new Set(s)
  }
}

async function loadDepartures() {
  const month = `${currentYear.value}-${String(currentMonth.value + 1).padStart(2, '0')}`
  try {
    const data = await adminApi.get<Departure[]>(
      `/admin/products/${props.productId}/departures`,
      { params: { month } },
    )
    departures.value = data || []
  } catch (e: any) {
    ElMessage.error(e.message || '加载团期失败')
  }
}

async function saveDeparture() {
  saving.value = true
  try {
    await adminApi.post(`/admin/products/${props.productId}/departures`, {
      departures: [{
        departure_date: editForm.value.departure_date,
        return_date: editForm.value.return_date || editForm.value.departure_date,
        adult_price: editForm.value.adult_price,
        child_price: editForm.value.child_price,
        infant_price: editForm.value.infant_price,
        single_supplement: editForm.value.single_supplement,
        total_stock: editForm.value.total_stock,
        cutoff_days: editForm.value.cutoff_days,
      }],
    })
    ElMessage.success('保存成功')
    showEditDialog.value = false
    loadDepartures()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function batchUpdate() {
  batchSaving.value = true
  try {
    await adminApi.put(`/admin/products/${props.productId}/departures/batch-price`, {
      mode: batchForm.value.mode,
      target_dates: Array.from(selectedDates.value),
      adult_price: batchForm.value.mode === 'fixed' ? batchForm.value.adult_price : undefined,
      child_price: batchForm.value.mode === 'fixed' ? batchForm.value.child_price : undefined,
      percentage: batchForm.value.mode === 'percentage' ? batchForm.value.percentage : undefined,
      amount: batchForm.value.mode === 'amount' ? batchForm.value.amount : undefined,
      reference_dates: batchForm.value.mode === 'follow' ? batchForm.value.reference_dates : undefined,
    })
    ElMessage.success('批量调价成功')
    showBatchDialog.value = false
    selectedDates.value = new Set()
    loadDepartures()
  } catch (e: any) {
    ElMessage.error(e.message || '批量调价失败')
  } finally {
    batchSaving.value = false
  }
}

function stockTagType(status: string) {
  if (status === 'full') return 'danger'
  if (status === 'tight') return 'warning'
  return 'success'
}

function stockLabel(status: string) {
  if (status === 'full') return '售罄'
  if (status === 'tight') return '紧张'
  return '充足'
}

watch(() => props.productId, () => {
  if (props.productId) loadDepartures()
})

onMounted(() => {
  if (props.productId) loadDepartures()
})
</script>

<style scoped>
.price-calendar {
  width: 100%;
}
.calendar-header {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
  gap: 12px;
}
.month-title {
  font-size: 16px;
  font-weight: 600;
  min-width: 120px;
  text-align: center;
}
.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 4px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 8px;
}
.weekday-header {
  text-align: center;
  font-weight: 600;
  color: #606266;
  padding: 8px 0;
  font-size: 13px;
}
.calendar-cell {
  min-height: 80px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 6px;
  cursor: pointer;
  transition: all 0.2s;
}
.calendar-cell:hover:not(.empty):not(.past) {
  border-color: #409eff;
  background: #ecf5ff;
}
.calendar-cell.empty {
  cursor: default;
  border-color: transparent;
}
.calendar-cell.past {
  background: #fafafa;
  color: #c0c4cc;
  cursor: default;
}
.calendar-cell.today {
  border-color: #409eff;
}
.calendar-cell.selected {
  background: #d9ecff;
  border-color: #409eff;
}
.calendar-cell.stock-full {
  background: #fef0f0;
}
.calendar-cell.stock-tight {
  background: #fdf6ec;
}
.cell-date {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 4px;
}
.cell-price {
  font-size: 12px;
}
.adult-price {
  color: #f56c6c;
  font-weight: 600;
  font-size: 14px;
}
.stock-info {
  margin-top: 2px;
}
.cell-no-departure {
  font-size: 11px;
  color: #c0c4cc;
  margin-top: 8px;
}
</style>

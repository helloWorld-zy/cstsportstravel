<template>
  <div class="departure-calendar">
    <div class="calendar-header">
      <button class="nav-btn" :disabled="!canGoPrev" @click="prevMonth">
        ‹
      </button>
      <span class="current-month">{{ currentMonthLabel }}</span>
      <button class="nav-btn" @click="nextMonth">›</button>
    </div>

    <div class="calendar-weekdays">
      <span v-for="day in weekdays" :key="day" class="weekday">{{ day }}</span>
    </div>

    <div class="calendar-grid">
      <div
        v-for="(cell, idx) in calendarCells"
        :key="idx"
        class="calendar-cell"
        :class="cellClass(cell)"
        @click="cell.clickable && emit('select', cell.departure)"
      >
        <span v-if="cell.day" class="day-num">{{ cell.day }}</span>
        <span v-if="cell.price" class="day-price">¥{{ cell.price }}</span>
        <span v-if="cell.status === 'sold_out'" class="day-status">售罄</span>
        <span v-else-if="cell.status === 'tight'" class="day-status tight">紧张</span>
      </div>
    </div>

    <div class="calendar-legend">
      <span class="legend-item"><span class="dot sufficient"></span> 充足</span>
      <span class="legend-item"><span class="dot tight"></span> 紧张</span>
      <span class="legend-item"><span class="dot sold-out"></span> 售罄</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { DepartureDate } from '~/composables/useProduct'

const props = defineProps<{
  departures: DepartureDate[]
  selectedDate?: string
}>()

const emit = defineEmits<{
  select: [departure: DepartureDate]
}>()

const weekdays = ['日', '一', '二', '三', '四', '五', '六']

const today = new Date()
const currentMonth = ref(new Date(today.getFullYear(), today.getMonth(), 1))

const currentMonthLabel = computed(() => {
  const y = currentMonth.value.getFullYear()
  const m = currentMonth.value.getMonth() + 1
  return `${y}年${m}月`
})

const canGoPrev = computed(() => {
  const prev = new Date(currentMonth.value)
  prev.setMonth(prev.getMonth() - 1)
  return prev >= new Date(today.getFullYear(), today.getMonth(), 1)
})

const prevMonth = () => {
  const d = new Date(currentMonth.value)
  d.setMonth(d.getMonth() - 1)
  currentMonth.value = d
}

const nextMonth = () => {
  const d = new Date(currentMonth.value)
  d.setMonth(d.getMonth() + 1)
  currentMonth.value = d
}

interface CalendarCell {
  day: number | null
  date: string | null
  price: number | null
  status: string | null
  clickable: boolean
  departure: DepartureDate | null
  isToday: boolean
  isPast: boolean
}

const calendarCells = computed<CalendarCell[]>(() => {
  const year = currentMonth.value.getFullYear()
  const month = currentMonth.value.getMonth()
  const firstDay = new Date(year, month, 1).getDay()
  const daysInMonth = new Date(year, month + 1, 0).getDate()

  // Build departure map by date string
  const depMap = new Map<string, DepartureDate>()
  for (const d of props.departures) {
    const key = d.departure_date.substring(0, 10) // YYYY-MM-DD
    depMap.set(key, d)
  }

  const cells: CalendarCell[] = []

  // Leading empty cells
  for (let i = 0; i < firstDay; i++) {
    cells.push({ day: null, date: null, price: null, status: null, clickable: false, departure: null, isToday: false, isPast: false })
  }

  // Day cells
  for (let day = 1; day <= daysInMonth; day++) {
    const date = new Date(year, month, day)
    const dateStr = `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    const isPast = date < new Date(today.getFullYear(), today.getMonth(), today.getDate())
    const isToday = date.toDateString() === today.toDateString()
    const dep = depMap.get(dateStr)

    cells.push({
      day,
      date: dateStr,
      price: dep ? dep.adult_price : null,
      status: dep ? dep.stock_status : null,
      clickable: !!dep && !isPast && dep.stock_status !== 'sold_out',
      departure: dep || null,
      isToday,
      isPast,
    })
  }

  return cells
})

function cellClass(cell: CalendarCell): string[] {
  const classes: string[] = []
  if (!cell.day) classes.push('empty')
  if (cell.isToday) classes.push('today')
  if (cell.isPast) classes.push('past')
  if (cell.status === 'sold_out') classes.push('sold-out')
  if (cell.status === 'tight') classes.push('tight')
  if (cell.clickable) classes.push('clickable')
  if (props.selectedDate && cell.date === props.selectedDate) classes.push('selected')
  return classes
}
</script>

<style scoped>
.departure-calendar {
  background: #fff;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 16px;
}

.calendar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.current-month {
  font-size: 16px;
  font-weight: 600;
}

.nav-btn {
  background: none;
  border: 1px solid #ddd;
  border-radius: 4px;
  width: 32px;
  height: 32px;
  cursor: pointer;
  font-size: 18px;
  color: #333;
}

.nav-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.calendar-weekdays {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  text-align: center;
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
}

.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 2px;
}

.calendar-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 6px 2px;
  min-height: 52px;
  border-radius: 4px;
}

.calendar-cell.empty {
  background: transparent;
}

.calendar-cell.today {
  background: #e3f2fd;
}

.calendar-cell.past {
  opacity: 0.4;
}

.calendar-cell.sold-out {
  background: #f5f5f5;
}

.calendar-cell.clickable {
  cursor: pointer;
}

.calendar-cell.clickable:hover {
  background: #fff3e0;
}

.calendar-cell.selected {
  background: #ff5722;
  color: #fff;
}

.calendar-cell.selected .day-price {
  color: #fff;
}

.day-num {
  font-size: 13px;
  font-weight: 500;
}

.day-price {
  font-size: 11px;
  color: #ff5722;
  font-weight: 500;
}

.day-status {
  font-size: 10px;
  color: #999;
}

.day-status.tight {
  color: #ff9800;
}

.calendar-legend {
  display: flex;
  gap: 16px;
  margin-top: 12px;
  justify-content: center;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #666;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot.sufficient {
  background: #4caf50;
}

.dot.tight {
  background: #ff9800;
}

.dot.sold-out {
  background: #ccc;
}
</style>

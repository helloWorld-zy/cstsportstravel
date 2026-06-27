<template>
  <div class="itinerary-editor">
    <div class="editor-header">
      <span>行程安排（共 {{ days.length }} 天）</span>
      <el-button size="small" @click="autoGenerate">根据天数自动生成</el-button>
    </div>

    <div v-if="days.length === 0" class="empty-tip">
      <el-empty description="暂无行程，请先设置产品天数后自动生成" />
    </div>

    <div v-for="(day, index) in days" :key="index" class="day-card">
      <div class="day-header">
        <el-tag type="primary" size="small">Day {{ day.day_no }}</el-tag>
        <el-input
          v-model="day.title"
          placeholder="行程标题，如：抵达丽江，自由活动"
          style="flex: 1; margin: 0 12px"
        />
        <el-button
          v-if="index > 0"
          size="small"
          :icon="Top"
          circle
          @click="moveUp(index)"
        />
        <el-button
          v-if="index < days.length - 1"
          size="small"
          :icon="Bottom"
          circle
          @click="moveDown(index)"
        />
        <el-button
          size="small"
          type="danger"
          :icon="Delete"
          circle
          @click="removeDay(index)"
        />
      </div>

      <div class="day-body">
        <el-form label-width="80px" size="small">
          <el-form-item label="行程概述">
            <el-input
              v-model="day.description"
              type="textarea"
              :rows="2"
              placeholder="当日行程详细描述"
            />
          </el-form-item>

          <el-form-item label="景点">
            <div class="spots-list">
              <div v-for="(spot, si) in day.spots" :key="si" class="spot-item">
                <el-input v-model="spot.name" placeholder="景点名称" style="width: 150px" />
                <el-input v-model="spot.duration" placeholder="游览时长" style="width: 100px" />
                <el-input v-model="spot.description" placeholder="景点描述" style="flex: 1" />
                <el-button size="small" type="danger" :icon="Delete" circle @click="day.spots.splice(si, 1)" />
              </div>
              <el-button size="small" @click="addSpot(day)">+ 添加景点</el-button>
            </div>
          </el-form-item>

          <el-row :gutter="16">
            <el-col :span="8">
              <el-form-item label="用餐">
                <el-checkbox v-model="day.meals.breakfast">早餐</el-checkbox>
                <el-checkbox v-model="day.meals.lunch">午餐</el-checkbox>
                <el-checkbox v-model="day.meals.dinner">晚餐</el-checkbox>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="住宿">
                <el-input v-model="day.hotel" placeholder="酒店名称/类型" />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="交通">
                <el-input v-model="day.transport" placeholder="交通方式" />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Top, Bottom, Delete } from '@element-plus/icons-vue'

interface Spot {
  name: string
  description: string
  duration: string
  image: string
}

interface Meals {
  breakfast: boolean
  lunch: boolean
  dinner: boolean
}

interface ItineraryDay {
  day_no: number
  title: string
  description: string
  meals: Meals
  hotel: string
  transport: string
  spots: Spot[]
  images: string[]
}

const props = defineProps<{
  modelValue: ItineraryDay[]
  totalDays: number
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ItineraryDay[]): void
}>()

const days = ref<ItineraryDay[]>(props.modelValue || [])

watch(() => props.modelValue, (val) => {
  if (val) days.value = val
}, { deep: true })

watch(days, (val) => {
  emit('update:modelValue', val)
}, { deep: true })

function autoGenerate() {
  const total = props.totalDays || 1
  const newDays: ItineraryDay[] = []
  for (let i = 1; i <= total; i++) {
    // Preserve existing data if available
    const existing = days.value.find(d => d.day_no === i)
    newDays.push(existing || {
      day_no: i,
      title: `第${i}天`,
      description: '',
      meals: { breakfast: false, lunch: false, dinner: false },
      hotel: '',
      transport: '',
      spots: [],
      images: [],
    })
  }
  days.value = newDays
}

function addSpot(day: ItineraryDay) {
  day.spots.push({ name: '', description: '', duration: '', image: '' })
}

function moveUp(index: number) {
  if (index <= 0) return
  const temp = days.value[index]
  days.value[index] = days.value[index - 1]
  days.value[index - 1] = temp
  // Re-number
  days.value.forEach((d, i) => { d.day_no = i + 1 })
}

function moveDown(index: number) {
  if (index >= days.value.length - 1) return
  const temp = days.value[index]
  days.value[index] = days.value[index + 1]
  days.value[index + 1] = temp
  days.value.forEach((d, i) => { d.day_no = i + 1 })
}

function removeDay(index: number) {
  days.value.splice(index, 1)
  days.value.forEach((d, i) => { d.day_no = i + 1 })
}
</script>

<style scoped>
.itinerary-editor {
  width: 100%;
}
.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.day-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  margin-bottom: 12px;
  overflow: hidden;
}
.day-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}
.day-body {
  padding: 16px;
}
.spots-list {
  width: 100%;
}
.spot-item {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}
.empty-tip {
  padding: 32px 0;
}
</style>

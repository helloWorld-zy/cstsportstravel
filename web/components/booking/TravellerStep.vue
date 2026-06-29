<template>
  <div class="traveller-step">
    <h3>填写出游人信息</h3>

    <div v-for="(traveller, index) in travellers" :key="index" class="traveller-form">
      <div class="form-header">
        <span>{{ travellerLabel(index) }}</span>
        <el-button size="small" @click="fillFromFrequent(index)">从常用出游人选择</el-button>
      </div>

      <el-form :model="traveller" label-width="80px">
        <el-form-item label="姓名" required>
          <el-input v-model="traveller.real_name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="身份证号" required>
          <el-input
            v-model="traveller.id_card_no"
            placeholder="18位身份证号"
            maxlength="18"
            @blur="validateIDCard(index)"
          />
          <span v-if="traveller.idCardError" class="error-text">{{ traveller.idCardError }}</span>
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="traveller.phone" placeholder="手机号码" maxlength="11" />
        </el-form-item>
        <el-form-item label="出生日期">
          <el-date-picker
            v-model="traveller.birth_date"
            type="date"
            placeholder="选择日期"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item label="性别">
          <el-radio-group v-model="traveller.gender">
            <el-radio value="male">男</el-radio>
            <el-radio value="female">女</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- Child must link to adult -->
        <el-form-item v-if="traveller.is_child || traveller.is_infant" label="关联成人" required>
          <el-select v-model="traveller.linked_adult_traveller_index" placeholder="选择关联的成人">
            <el-option
              v-for="(adult, aIdx) in adultIndices"
              :key="aIdx"
              :label="`成人${aIdx + 1} - ${travellers[aIdx]?.real_name || '未填写'}`"
              :value="aIdx"
            />
          </el-select>
        </el-form-item>
      </el-form>
    </div>

    <!-- Contact info -->
    <div class="section">
      <h4>联系人信息</h4>
      <el-form :model="contactForm" label-width="80px">
        <el-form-item label="联系人" required>
          <el-input v-model="contactForm.name" placeholder="联系人姓名" />
        </el-form-item>
        <el-form-item label="手机号" required>
          <el-input v-model="contactForm.phone" placeholder="联系人手机号" maxlength="11" />
        </el-form-item>
      </el-form>
    </div>

    <div class="actions">
      <el-button @click="emit('back')">上一步</el-button>
      <el-button type="primary" :disabled="!isValid" @click="handleNext">下一步</el-button>
    </div>

    <!-- CHK004: Frequent traveller selection dialog -->
    <el-dialog v-model="showFrequentDialog" title="选择常用出游人" width="500px">
      <div class="frequent-list">
        <div
          v-for="ft in frequentTravellers"
          :key="ft.id"
          class="frequent-item"
          @click="selectFrequentTraveller(ft)"
        >
          <div class="frequent-name">{{ ft.real_name }}</div>
          <div class="frequent-meta">{{ ft.id_card_no }} · {{ ft.phone }}</div>
        </div>
      </div>
      <div v-if="!frequentTravellers.length" style="text-align: center; color: #999; padding: 20px;">
        暂无常用出游人
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const props = defineProps<{
  adultCount: number
  childCount: number
  infantCount: number
}>()

const emit = defineEmits<{
  update: [data: any]
  next: []
  back: []
}>()

interface TravellerForm {
  real_name: string
  id_card_no: string
  phone: string
  birth_date: string
  gender: string
  is_child: boolean
  is_infant: boolean
  linked_adult_traveller_index: number | null
  idCardError: string
}

const travellers = ref<TravellerForm[]>([])
const contactForm = ref({ name: '', phone: '' })

onMounted(() => {
  // Initialize traveller forms
  const forms: TravellerForm[] = []
  for (let i = 0; i < props.adultCount; i++) {
    forms.push(createTravellerForm(false, false))
  }
  for (let i = 0; i < props.childCount; i++) {
    forms.push(createTravellerForm(true, false))
  }
  for (let i = 0; i < props.infantCount; i++) {
    forms.push(createTravellerForm(false, true))
  }
  travellers.value = forms
})

function createTravellerForm(isChild: boolean, isInfant: boolean): TravellerForm {
  return {
    real_name: '',
    id_card_no: '',
    phone: '',
    birth_date: '',
    gender: '',
    is_child: isChild,
    is_infant: isInfant,
    linked_adult_traveller_index: null,
    idCardError: '',
  }
}

function travellerLabel(index: number): string {
  if (index < props.adultCount) return `成人 ${index + 1}`
  if (index < props.adultCount + props.childCount) return `儿童 ${index - props.adultCount + 1}`
  return `婴儿 ${index - props.adultCount - props.childCount + 1}`
}

const adultIndices = computed(() => {
  return Array.from({ length: props.adultCount }, (_, i) => i)
})

function validateIDCard(index: number) {
  const t = travellers.value[index]
  if (!t.id_card_no) {
    t.idCardError = ''
    return
  }
  if (t.id_card_no.length !== 18) {
    t.idCardError = '身份证号应为18位'
    return
  }
  // ISO 7064:1983.MOD 11-2 checksum
  const weights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2]
  const checkCodes = ['1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2']
  let sum = 0
  for (let i = 0; i < 17; i++) {
    sum += parseInt(t.id_card_no[i]) * weights[i]
  }
  const expected = checkCodes[sum % 11]
  const actual = t.id_card_no[17].toUpperCase()
  if (actual !== expected) {
    t.idCardError = '身份证号码校验码不正确'
  } else {
    t.idCardError = ''
  }
}

// CHK004: Implement frequent traveller selection
const frequentTravellers = ref<any[]>([])
const showFrequentDialog = ref(false)
const fillingIndex = ref(0)

async function fillFromFrequent(index: number) {
  fillingIndex.value = index
  try {
    const { data } = await useFetch('/api/v1/users/me/travellers')
    const response = data.value as any
    if (response?.code === 0 && response.data?.length) {
      frequentTravellers.value = response.data
      showFrequentDialog.value = true
    } else {
      ElMessage.info('暂无常用出游人，请先在个人中心添加')
    }
  } catch {
    ElMessage.error('获取常用出游人失败')
  }
}

function selectFrequentTraveller(traveller: any) {
  const idx = fillingIndex.value
  const t = travellers.value[idx]
  if (!t) return

  t.real_name = traveller.real_name || ''
  t.id_card_no = traveller.id_card_no || ''
  t.phone = traveller.phone || ''
  t.birth_date = traveller.birth_date || ''
  t.gender = traveller.gender || ''

  // Auto-detect child/infant by birth date
  if (traveller.birth_date) {
    const birth = new Date(traveller.birth_date)
    const now = new Date()
    const age = now.getFullYear() - birth.getFullYear()
    if (age < 2) {
      t.is_infant = true
      t.is_child = false
    } else if (age < 12) {
      t.is_child = true
      t.is_infant = false
    }
  }

  validateIDCard(idx)
  showFrequentDialog.value = false
  ElMessage.success('已填充出游人信息')
}

const isValid = computed(() => {
  // Check all travellers have required fields
  for (const t of travellers.value) {
    if (!t.real_name || !t.id_card_no || t.idCardError) return false
    if ((t.is_child || t.is_infant) && t.linked_adult_traveller_index === null) return false
  }
  // Check contact info
  if (!contactForm.value.name || !contactForm.value.phone) return false
  return true
})

function handleNext() {
  emit('update', {
    travellers: travellers.value.map(t => ({
      real_name: t.real_name,
      id_card_no: t.id_card_no,
      phone: t.phone,
      birth_date: t.birth_date,
      gender: t.gender,
      is_child: t.is_child,
      is_infant: t.is_infant,
      linked_adult_traveller_index: t.linked_adult_traveller_index,
    })),
    contactName: contactForm.value.name,
    contactPhone: contactForm.value.phone,
  })
  emit('next')
}
</script>

<style scoped>
.traveller-step h3 {
  margin-bottom: 20px;
}

.traveller-form {
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 500;
}

.error-text {
  color: #ff4d4f;
  font-size: 12px;
}

.section {
  margin-top: 24px;
}

.section h4 {
  margin-bottom: 12px;
}

.actions {
  display: flex;
  justify-content: space-between;
  margin-top: 24px;
}

/* CHK004: Frequent traveller dialog styles */
.frequent-list {
  max-height: 400px;
  overflow-y: auto;
}

.frequent-item {
  padding: 12px;
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.frequent-item:hover {
  border-color: #409eff;
  background: #f5f7ff;
}

.frequent-name {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.frequent-meta {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}
</style>

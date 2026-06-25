<template>
  <div class="payment-countdown" :class="{ warning: isWarning, expired: isExpired }">
    <template v-if="isExpired">
      <span class="expired-text">支付已超时</span>
    </template>
    <template v-else>
      <span class="label">支付剩余时间：</span>
      <span class="time">{{ displayMinutes }}:{{ displaySeconds }}</span>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

const props = defineProps<{
  expireAt: string | Date
}>()

const emit = defineEmits<{
  expired: []
}>()

const remaining = ref(0) // seconds
let timer: ReturnType<typeof setInterval> | null = null

const isWarning = computed(() => remaining.value > 0 && remaining.value <= 300) // <=5min
const isExpired = computed(() => remaining.value <= 0)

const displayMinutes = computed(() => {
  const mins = Math.floor(remaining.value / 60)
  return String(mins).padStart(2, '0')
})

const displaySeconds = computed(() => {
  const secs = remaining.value % 60
  return String(secs).padStart(2, '0')
})

function updateRemaining() {
  const now = new Date()
  const expire = new Date(props.expireAt)
  const diff = Math.floor((expire.getTime() - now.getTime()) / 1000)
  remaining.value = Math.max(0, diff)

  if (remaining.value <= 0) {
    emit('expired')
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }
}

onMounted(() => {
  updateRemaining()
  timer = setInterval(updateRemaining, 1000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style scoped>
.payment-countdown {
  text-align: center;
  padding: 16px;
  font-size: 18px;
}

.label {
  color: #666;
}

.time {
  font-weight: bold;
  font-size: 24px;
  color: #333;
  font-variant-numeric: tabular-nums;
}

.warning .time {
  color: #faad14;
}

.expired-text {
  color: #ff4d4f;
  font-weight: bold;
}
</style>

<template>
  <div class="real-name-page">
    <h1>实名认证</h1>

    <div v-if="user?.real_name_status === 'verified'" class="verified-status">
      <el-result icon="success" title="已实名认证" sub-title="您已完成实名认证，可以预订境内游产品" />
    </div>

    <div v-else-if="user?.real_name_status === 'pending'" class="pending-status">
      <el-result icon="info" title="认证审核中" sub-title="您的实名认证正在审核中，请耐心等待" />
    </div>

    <div v-else class="form-container">
      <RealNameForm @success="handleSuccess" />
    </div>
  </div>
</template>

<script setup lang="ts">
const { user, fetchProfile, init } = useAuth()

onMounted(async () => {
  await init()
  if (!user.value) {
    navigateTo('/auth/login')
  }
})

async function handleSuccess(status: string) {
  await fetchProfile()
}
</script>

<style scoped>
.real-name-page {
  max-width: 600px;
  margin: 0 auto;
  padding: var(--space-lg);
}
.real-name-page h1 {
  margin-bottom: var(--space-lg);
}
.verified-status,
.pending-status {
  background: var(--color-bg-white);
  border-radius: var(--radius-lg);
  padding: var(--space-xl);
}
.form-container {
  background: var(--color-bg-white);
  border-radius: var(--radius-lg);
  padding: var(--space-xl);
}
</style>

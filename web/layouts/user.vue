<template>
  <NuxtLayout name="default">
    <div class="user-layout">
      <div class="user-layout-inner">
        <!-- Left Sidebar (Desktop Only) -->
        <aside class="user-sidebar">
          <div class="user-profile-summary">
            <el-avatar :size="56" :src="user?.avatar_url" />
            <div class="user-summary-details">
              <h3 class="user-name">{{ user?.nickname || '未登录' }}</h3>
              <span class="user-level">
                <span class="lv-crown">👑</span> Lv.{{ user?.member_level || 1 }} 会员
              </span>
            </div>
          </div>
          
          <nav class="user-nav-menu">
            <NuxtLink to="/user" class="user-nav-item" :class="{ active: $route.path === '/user' }">
              <span class="nav-icon">👤</span> 账户中心
            </NuxtLink>
            <NuxtLink to="/user/orders" class="user-nav-item" :class="{ active: $route.path.startsWith('/user/orders') || $route.path.includes('/user/order-') }">
              <span class="nav-icon">📋</span> 我的订单
            </NuxtLink>
            <NuxtLink to="/user/travellers" class="user-nav-item" :class="{ active: $route.path === '/user/travellers' }">
              <span class="nav-icon">👥</span> 常用出游人
            </NuxtLink>
            <NuxtLink to="/user/real-name" class="user-nav-item" :class="{ active: $route.path === '/user/real-name' }">
              <span class="nav-icon">🪪</span> 实名认证
            </NuxtLink>
            <div class="nav-divider"></div>
            <button @click="handleLogout" class="user-nav-item logout-btn">
              <span class="nav-icon">🚪</span> 退出登录
            </button>
          </nav>
        </aside>
        
        <!-- Right Main Panel -->
        <main class="user-main-content">
          <slot />
        </main>
      </div>
    </div>
  </NuxtLayout>
</template>

<script setup lang="ts">
const { user, logout, init } = useAuth()

onMounted(async () => {
  await init()
  if (!user.value) {
    navigateTo('/auth/login')
  }
})

function handleLogout() {
  logout()
}
</script>

<style scoped>
.user-layout {
  background-color: #f8fafc;
  min-height: calc(100vh - 70px);
  padding: 40px 24px;
}

.user-layout-inner {
  max-width: 1200px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 32px;
  align-items: start;
}

/* Sidebar Styles */
.user-sidebar {
  background: #fff;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  padding: 24px 16px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.01), 0 2px 4px -2px rgba(0, 0, 0, 0.01);
  position: sticky;
  top: 90px;
}

.user-profile-summary {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 20px;
  margin-bottom: 20px;
  border-bottom: 1px solid #f1f5f9;
}

.user-summary-details h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
}

.user-level {
  font-size: 11px;
  color: #f59e0b;
  font-weight: 700;
  background-color: #fffbeb;
  padding: 2px 8px;
  border-radius: 20px;
  display: inline-flex;
  align-items: center;
  gap: 2px;
  margin-top: 4px;
}

.lv-crown {
  font-size: 10px;
}

.user-nav-menu {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-radius: 8px;
  text-decoration: none;
  font-size: 14px;
  font-weight: 600;
  color: #475569;
  transition: all 0.2s ease;
  background: transparent;
  border: none;
  width: 100%;
  text-align: left;
  cursor: pointer;
}

.user-nav-item:hover {
  background-color: #f1f5f9;
  color: #0f172a;
}

.user-nav-item.active {
  background-color: #eff6ff;
  color: #2563eb;
}

.nav-divider {
  height: 1px;
  background-color: #f1f5f9;
  margin: 8px 0;
}

.logout-btn {
  color: #ef4444;
}

.logout-btn:hover {
  background-color: #fef2f2;
  color: #dc2626;
}

/* Right Content Panel */
.user-main-content {
  background: #fff;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  padding: 32px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.01), 0 2px 4px -2px rgba(0, 0, 0, 0.01);
  min-height: 550px;
}

/* Mobile Responsive */
@media (max-width: 968px) {
  .user-layout {
    padding: 16px 12px;
  }
  
  .user-layout-inner {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  
  .user-sidebar {
    position: static;
    padding: 16px;
  }
  
  .user-profile-summary {
    margin-bottom: 0;
    border-bottom: none;
    padding-bottom: 0;
  }
  
  .user-nav-menu {
    display: none; /* Hide sidebar nav on mobile, use index.vue as hub */
  }
  
  .user-main-content {
    padding: 20px 16px;
  }
}
</style>

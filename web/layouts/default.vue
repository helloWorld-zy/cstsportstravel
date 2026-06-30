<template>
  <div class="site-layout">
    <!-- Header -->
    <header class="site-header">
      <div class="header-container">
        <div class="logo-area" @click="navigateTo('/')">
          <span class="logo-icon">🏔️</span>
          <span class="logo-text">山海体育旅游</span>
        </div>

        <!-- Navigation Links -->
        <nav class="nav-links">
          <NuxtLink to="/" class="nav-item" active-class="active">首页</NuxtLink>
          <NuxtLink to="/products" class="nav-item" active-class="active">跟团游</NuxtLink>
          <NuxtLink to="/user" class="nav-item" active-class="active">个人中心</NuxtLink>
        </nav>

        <!-- User Profile Area -->
        <div class="user-area">
          <template v-if="isLoggedIn">
            <el-dropdown trigger="click" @command="handleUserCommand">
              <span class="user-profile-trigger">
                <el-avatar :size="32" :src="user?.avatar_url" class="avatar" />
                <span class="nickname hide-on-mobile">{{ user?.nickname || '我的' }}</span>
              </span>
              <template #dropdown>
                <el-dropdown-menu class="modern-dropdown">
                  <el-dropdown-item command="profile">
                    <span class="dropdown-icon">👤</span>个人中心
                  </el-dropdown-item>
                  <el-dropdown-item command="orders">
                    <span class="dropdown-icon">📋</span>我的订单
                  </el-dropdown-item>
                  <el-dropdown-item command="travellers">
                    <span class="dropdown-icon">👥</span>常用出游人
                  </el-dropdown-item>
                  <el-dropdown-item command="real-name">
                    <span class="dropdown-icon">🪪</span>实名认证
                  </el-dropdown-item>
                  <el-dropdown-item divided command="logout" class="logout-item">
                    <span class="dropdown-icon">🚪</span>退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
          <template v-else>
            <el-button type="primary" size="default" round @click="navigateTo('/auth/login')" class="login-btn">
              登录 / 注册
            </el-button>
          </template>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="main-content">
      <slot />
    </main>

    <!-- Footer -->
    <footer class="site-footer">
      <div class="footer-container">
        <div class="footer-info">
          <h3>山海体育旅游</h3>
          <p>高品质境内外体育活动、户外跟团、徒步探索一站式服务平台。</p>
          <div class="contact-number">
            <span class="phone-icon">📞</span>
            <span class="phone-num">400-888-9999</span>
          </div>
        </div>
        <div class="footer-section">
          <h4>热门目的地</h4>
          <ul>
            <li><NuxtLink to="/products?destination=云南">云南跟团游</NuxtLink></li>
            <li><NuxtLink to="/products?destination=海南">海南海岛游</NuxtLink></li>
            <li><NuxtLink to="/products?destination=北京">北京文化游</NuxtLink></li>
            <li><NuxtLink to="/products?destination=四川">四川九寨沟</NuxtLink></li>
          </ul>
        </div>
        <div class="footer-section">
          <h4>出行指南</h4>
          <ul>
            <li><a href="#">预订流程说明</a></li>
            <li><a href="#">退改规则须知</a></li>
            <li><a href="#">出游人要求</a></li>
            <li><a href="#">常见问题解答</a></li>
          </ul>
        </div>
      </div>
      <div class="footer-bottom">
        <p>&copy; 2026 山海体育旅游 CST Sports Travel. All Rights Reserved. 粤ICP备12345678号</p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { useAuth } from '~/composables/useAuth'

const { user, isLoggedIn, logout, init } = useAuth()

onMounted(async () => {
  await init()
})

function handleUserCommand(command: string) {
  if (command === 'logout') {
    logout()
  } else if (command === 'profile') {
    navigateTo('/user')
  } else if (command === 'orders') {
    navigateTo('/user/orders')
  } else if (command === 'travellers') {
    navigateTo('/user/travellers')
  } else if (command === 'real-name') {
    navigateTo('/user/real-name')
  }
}
</script>

<style>
/* Global site layout styling */
.site-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: #f8fafc;
}

.site-header {
  position: sticky;
  top: 0;
  z-index: 1000;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(226, 232, 240, 0.8);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px -2px rgba(0, 0, 0, 0.02);
}

.header-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
  height: 70px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.logo-icon {
  font-size: 24px;
}

.logo-text {
  font-size: 20px;
  font-weight: 800;
  background: linear-gradient(135deg, #1d4ed8 0%, #3b82f6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: 0.5px;
}

.nav-links {
  display: flex;
  gap: 32px;
}

.nav-item {
  font-size: 15px;
  font-weight: 500;
  color: #475569;
  text-decoration: none;
  padding: 6px 4px;
  border-bottom: 2px solid transparent;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.nav-item:hover {
  color: #2563eb;
}

.nav-item.active {
  color: #2563eb;
  font-weight: 600;
  border-bottom-color: #2563eb;
}

.user-area {
  display: flex;
  align-items: center;
}

.user-profile-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 12px;
  border-radius: 20px;
  transition: background 0.2s;
}

.user-profile-trigger:hover {
  background: #f1f5f9;
}

.user-profile-trigger .nickname {
  font-size: 14px;
  font-weight: 500;
  color: #334155;
}

.modern-dropdown .dropdown-icon {
  margin-right: 8px;
  font-size: 16px;
}

.modern-dropdown .logout-item {
  color: #ef4444;
}

.login-btn {
  background: linear-gradient(135deg, #1d4ed8 0%, #3b82f6 100%) !important;
  border: none !important;
  font-weight: 500 !important;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.2);
  transition: all 0.2s !important;
}

.login-btn:hover {
  opacity: 0.9;
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(255, 87, 34, 0.3);
}

.main-content {
  flex: 1;
  width: 100%;
}

.site-footer {
  background: #0f172a;
  color: #94a3b8;
  padding: 56px 24px 28px;
  border-top: 1px solid #1e293b;
}

.footer-container {
  max-width: 1200px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 2fr 1fr 1fr;
  gap: 48px;
  margin-bottom: 40px;
}

.footer-info h3 {
  font-size: 18px;
  color: #f1f5f9;
  margin-bottom: 16px;
  font-weight: 700;
}

.footer-info p {
  line-height: 1.6;
  margin-bottom: 24px;
  font-size: 14px;
}

.contact-number {
  display: flex;
  align-items: center;
  gap: 8px;
}

.phone-icon {
  font-size: 20px;
}

.phone-num {
  font-size: 24px;
  font-weight: 700;
  color: #fff;
}

.footer-section h4 {
  font-size: 16px;
  color: #f1f5f9;
  margin-bottom: 20px;
  font-weight: 600;
}

.footer-section ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.footer-section ul li {
  margin-bottom: 12px;
}

.footer-section ul li a {
  color: #94a3b8;
  text-decoration: none;
  font-size: 14px;
  transition: color 0.2s;
}

.footer-section ul li a:hover {
  color: #2563eb;
}

.footer-bottom {
  max-width: 1200px;
  margin: 0 auto;
  border-top: 1px solid #1e293b;
  padding-top: 24px;
  text-align: center;
  font-size: 13px;
  color: #64748b;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .header-container {
    padding: 0 16px;
  }
  
  .nav-links {
    display: none; /* Hide standard links on mobile */
  }
  
  .footer-container {
    grid-template-columns: 1fr;
    gap: 32px;
  }
  
  .hide-on-mobile {
    display: none !important;
  }
}
</style>

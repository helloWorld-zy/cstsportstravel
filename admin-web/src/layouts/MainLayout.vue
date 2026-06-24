<template>
  <el-container class="main-layout">
    <!-- Sidebar -->
    <el-aside :width="isCollapsed ? '64px' : '220px'" class="sidebar">
      <div class="logo">
        <span v-if="!isCollapsed">旅行管理后台</span>
        <span v-else>旅</span>
      </div>

      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapsed"
        :router="true"
        background-color="#001529"
        text-color="#ffffffa6"
        active-text-color="#ffffff"
        class="sidebar-menu"
      >
        <template v-for="item in menuItems" :key="item.path">
          <!-- Submenu with children -->
          <el-sub-menu v-if="item.children?.length" :index="item.path">
            <template #title>
              <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
              <span>{{ item.title }}</span>
            </template>
            <el-menu-item
              v-for="child in item.children"
              :key="child.path"
              :index="child.path"
            >
              {{ child.title }}
            </el-menu-item>
          </el-sub-menu>

          <!-- Single menu item -->
          <el-menu-item v-else :index="item.path">
            <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <!-- Main content area -->
    <el-container>
      <!-- Header -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="isCollapsed = !isCollapsed">
            <Fold v-if="!isCollapsed" />
            <Expand v-else />
          </el-icon>

          <!-- Breadcrumb -->
          <el-breadcrumb separator="/" class="breadcrumb">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
              {{ item.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="32" icon="User" />
              <span class="username">{{ username }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                <el-dropdown-item command="password">修改密码</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- Content -->
      <el-main class="content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { Fold, Expand, User } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const isCollapsed = ref(false)
const username = ref('管理员')

// Menu structure — in production, this is dynamically generated from RBAC permissions
const menuItems = ref([
  {
    title: '产品管理',
    path: '/products',
    icon: 'Goods',
    children: [
      { title: '产品列表', path: '/products' },
      { title: '产品审核', path: '/products/review' },
    ],
  },
  {
    title: '订单管理',
    path: '/orders',
    icon: 'Document',
    children: [
      { title: '订单列表', path: '/orders' },
      { title: '退款审核', path: '/orders/refunds' },
    ],
  },
  {
    title: '系统管理',
    path: '/system',
    icon: 'Setting',
    children: [
      { title: '用户管理', path: '/system/users' },
      { title: '角色管理', path: '/system/roles' },
    ],
  },
])

// Active menu item from current route
const activeMenu = computed(() => route.path)

// Breadcrumbs from route meta
const breadcrumbs = computed(() => {
  const crumbs: Array<{ path: string; title: string }> = []
  if (route.meta?.title) {
    crumbs.push({ path: route.path, title: route.meta.title as string })
  }
  return crumbs
})

function handleCommand(command: string) {
  switch (command) {
    case 'logout':
      localStorage.removeItem('admin_token')
      router.push('/login')
      break
    case 'profile':
      // TODO: navigate to profile
      break
    case 'password':
      // TODO: show change password dialog
      break
  }
}
</script>

<style scoped>
.main-layout {
  height: 100vh;
}

.sidebar {
  background-color: #001529;
  transition: width 0.3s;
  overflow: hidden;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid #ffffff1a;
}

.sidebar-menu {
  border-right: none;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  padding: 0 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: #333;
}

.breadcrumb {
  font-size: 14px;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.username {
  font-size: 14px;
  color: #333;
}

.content {
  background: #f5f5f5;
  padding: 20px;
  overflow-y: auto;
}
</style>

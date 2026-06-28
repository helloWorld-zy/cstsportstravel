<template>
  <div class="homepage-config">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>首页配置</span>
        </div>
      </template>

      <!-- Banner Management -->
      <div class="config-section">
        <div class="section-header">
          <h3>Banner 管理</h3>
          <el-button type="primary" size="small" @click="addBanner">添加 Banner</el-button>
        </div>
        <el-table :data="banners" style="width: 100%">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column label="图片" width="120">
            <template #default="{ row }">
              <el-image
                :src="row.image_url"
                :preview-src-list="[row.image_url]"
                style="width: 80px; height: 45px"
                fit="cover"
              />
            </template>
          </el-table-column>
          <el-table-column prop="title" label="标题" />
          <el-table-column prop="link" label="链接" />
          <el-table-column prop="sort_order" label="排序" width="80" />
          <el-table-column prop="status" label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'">
                {{ row.status === 'active' ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row, $index }">
              <el-button size="small" @click="editBanner(row, $index)">编辑</el-button>
              <el-button size="small" type="danger" @click="removeBanner($index)">删除</el-button>
              <el-button
                size="small"
                :type="row.status === 'active' ? 'warning' : 'success'"
                @click="toggleBannerStatus(row)"
              >
                {{ row.status === 'active' ? '禁用' : '启用' }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- Popular Destinations Config -->
      <div class="config-section">
        <div class="section-header">
          <h3>热门目的地配置</h3>
          <el-button type="primary" size="small" @click="addDestination">添加目的地</el-button>
        </div>
        <el-table :data="destinations" style="width: 100%">
          <el-table-column prop="name" label="目的地" />
          <el-table-column prop="product_count" label="产品数" width="100" />
          <el-table-column prop="min_price" label="最低价" width="100">
            <template #default="{ row }">¥{{ row.min_price }}</template>
          </el-table-column>
          <el-table-column prop="sort_order" label="排序" width="80" />
          <el-table-column label="操作" width="160">
            <template #default="{ row, $index }">
              <el-button size="small" @click="editDestination(row, $index)">编辑</el-button>
              <el-button size="small" type="danger" @click="removeDestination($index)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- Banner Edit Dialog -->
    <el-dialog v-model="bannerDialogVisible" :title="editingBannerIndex >= 0 ? '编辑 Banner' : '添加 Banner'" width="500px">
      <el-form :model="editingBanner" label-width="80px">
        <el-form-item label="标题" required>
          <el-input v-model="editingBanner.title" placeholder="Banner 标题" />
        </el-form-item>
        <el-form-item label="图片" required>
          <el-input v-model="editingBanner.image_url" placeholder="图片 URL" />
          <!-- TODO: image upload -->
        </el-form-item>
        <el-form-item label="链接">
          <el-input v-model="editingBanner.link" placeholder="点击跳转链接" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="editingBanner.sort_order" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="editingBanner.status" active-value="active" inactive-value="hidden" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bannerDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveBanner">保存</el-button>
      </template>
    </el-dialog>

    <!-- Destination Edit Dialog -->
    <el-dialog v-model="destDialogVisible" :title="editingDestIndex >= 0 ? '编辑目的地' : '添加目的地'" width="400px">
      <el-form :model="editingDest" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="editingDest.name" placeholder="目的地名称" />
        </el-form-item>
        <el-form-item label="图片">
          <el-input v-model="editingDest.image_url" placeholder="图片 URL" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="editingDest.sort_order" :min="0" :max="999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="destDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveDestination">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { listBanners, createBanner, updateBanner, deleteBanner, listDestinations } from '@/api/banner'
import type { Banner, Destination } from '@/api/banner'

// Banners
const banners = ref<Banner[]>([])
const bannersLoading = ref(false)

async function loadBanners() {
  bannersLoading.value = true
  try {
    const res = await listBanners()
    banners.value = res.data?.items || []
  } catch {
    ElMessage.error('加载 Banner 列表失败')
  } finally {
    bannersLoading.value = false
  }
}

const bannerDialogVisible = ref(false)
const editingBannerIndex = ref(-1)
const editingBanner = reactive<Banner>({
  id: 0,
  image_url: '',
  title: '',
  link: '',
  sort_order: 0,
  status: 'active',
})

function addBanner() {
  editingBannerIndex.value = -1
  Object.assign(editingBanner, { id: 0, image_url: '', title: '', link: '', sort_order: banners.value.length, status: 'active' })
  bannerDialogVisible.value = true
}

function editBanner(row: Banner, index: number) {
  editingBannerIndex.value = index
  Object.assign(editingBanner, row)
  bannerDialogVisible.value = true
}

async function saveBanner() {
  if (!editingBanner.title || !editingBanner.image_url) {
    ElMessage.warning('请填写标题和图片')
    return
  }
  try {
    if (editingBannerIndex.value >= 0) {
      // Update existing banner
      await updateBanner(editingBanner.id, {
        title: editingBanner.title,
        image_url: editingBanner.image_url,
        link: editingBanner.link,
        sort_order: editingBanner.sort_order,
        status: editingBanner.status,
      })
      banners.value[editingBannerIndex.value] = { ...editingBanner }
    } else {
      // Create new banner
      const res = await createBanner({
        title: editingBanner.title,
        image_url: editingBanner.image_url,
        link: editingBanner.link,
        sort_order: editingBanner.sort_order,
        status: editingBanner.status,
      })
      banners.value.push({ ...editingBanner, id: res.data?.id || Date.now() })
    }
    bannerDialogVisible.value = false
    ElMessage.success('保存成功')
  } catch {
    ElMessage.error('保存失败')
  }
}

async function removeBanner(index: number) {
  const banner = banners.value[index]
  try {
    if (banner.id) {
      await deleteBanner(banner.id)
    }
    banners.value.splice(index, 1)
    ElMessage.success('已删除')
  } catch {
    ElMessage.error('删除失败')
  }
}

async function toggleBannerStatus(row: Banner) {
  const newStatus = row.status === 'active' ? 'hidden' : 'active'
  try {
    await updateBanner(row.id, { status: newStatus })
    row.status = newStatus
  } catch {
    ElMessage.error('状态更新失败')
  }
}

// Destinations
const destinations = ref<Destination[]>([])
const destsLoading = ref(false)

async function loadDestinations() {
  destsLoading.value = true
  try {
    const res = await listDestinations()
    destinations.value = res.data?.items || []
  } catch {
    ElMessage.error('加载目的地列表失败')
  } finally {
    destsLoading.value = false
  }
}

const destDialogVisible = ref(false)
const editingDestIndex = ref(-1)
const editingDest = reactive<Destination>({
  id: 0,
  name: '',
  image_url: '',
  product_count: 0,
  min_price: 0,
  sort_order: 0,
})

function addDestination() {
  editingDestIndex.value = -1
  Object.assign(editingDest, { id: 0, name: '', image_url: '', product_count: 0, min_price: 0, sort_order: destinations.value.length })
  destDialogVisible.value = true
}

function editDestination(row: Destination, index: number) {
  editingDestIndex.value = index
  Object.assign(editingDest, row)
  destDialogVisible.value = true
}

function saveDestination() {
  if (!editingDest.name) {
    ElMessage.warning('请填写目的地名称')
    return
  }
  if (editingDestIndex.value >= 0) {
    destinations.value[editingDestIndex.value] = { ...editingDest }
  } else {
    destinations.value.push({ ...editingDest })
  }
  destDialogVisible.value = false
  ElMessage.success('保存成功')
}

function removeDestination(index: number) {
  destinations.value.splice(index, 1)
  ElMessage.success('已删除')
}

// Load data on mount
onMounted(() => {
  loadBanners()
  loadDestinations()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.config-section {
  margin-bottom: 32px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}
</style>

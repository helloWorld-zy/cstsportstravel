<template>
  <div class="cancellation-rule">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>退改规则配置</span>
          <div>
            <el-button @click="loadDefaultRules">加载默认模板</el-button>
            <el-button type="primary" @click="handleCreate">新建模板</el-button>
          </div>
        </div>
      </template>

      <!-- Template List -->
      <el-table :data="templates" style="width: 100%" v-loading="isLoading">
        <el-table-column prop="rule_name" label="规则名称" min-width="200" />
        <el-table-column label="距出发天数" width="200">
          <template #default="{ row }">
            {{ row.days_before_min }}天
            <template v-if="row.days_before_max"> ~ {{ row.days_before_max }}天</template>
            <template v-else>以上</template>
          </template>
        </el-table-column>
        <el-table-column label="退款比例" width="120">
          <template #default="{ row }">
            <span class="percentage">{{ row.refund_percentage }}%</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="说明" min-width="250" />
        <el-table-column prop="is_template" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_template ? 'primary' : 'info'" size="small">
              {{ row.is_template ? '模板' : '产品规则' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>

      <el-divider />

      <!-- Assign to Product -->
      <h4>将模板分配给产品</h4>
      <el-form :inline="true" size="default" style="margin-top: 12px">
        <el-form-item label="产品ID">
          <el-input-number v-model="assignProductId" :min="1" placeholder="产品ID" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="assigning" @click="assignToProduct">
            分配选中模板
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="showCreateDialog" :title="editingId ? '编辑规则' : '新建退改规则'" width="700px">
      <div class="rule-editor">
        <div v-for="(rule, index) in editRules" :key="index" class="rule-row">
          <el-form :inline="true" size="default">
            <el-form-item label="规则名称">
              <el-input v-model="rule.rule_name" style="width: 150px" />
            </el-form-item>
            <el-form-item label="最小天数">
              <el-input-number v-model="rule.days_before_min" :min="0" style="width: 100px" />
            </el-form-item>
            <el-form-item label="最大天数">
              <el-input-number v-model="rule.days_before_max" :min="0" style="width: 100px" />
            </el-form-item>
            <el-form-item label="退款比例%">
              <el-input-number v-model="rule.refund_percentage" :min="0" :max="100" style="width: 100px" />
            </el-form-item>
            <el-form-item>
              <el-button type="danger" :icon="Delete" circle size="small" @click="editRules.splice(index, 1)" />
            </el-form-item>
          </el-form>
          <el-input
            v-model="rule.description"
            placeholder="说明"
            size="small"
            style="margin-top: -8px; margin-bottom: 8px; width: 100%"
          />
        </div>
        <el-button @click="addRule" style="margin-top: 8px">+ 添加规则</el-button>
      </div>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveRules">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/request'

interface Rule {
  id?: number
  rule_name: string
  days_before_min: number
  days_before_max: number | null
  refund_percentage: number
  description: string
  is_template: boolean
}

const templates = ref<Rule[]>([])
const isLoading = ref(false)
const showCreateDialog = ref(false)
const saving = ref(false)
const assigning = ref(false)
const editingId = ref<number | null>(null)
const assignProductId = ref<number>(1)
const editRules = ref<Rule[]>([])

async function loadTemplates() {
  isLoading.value = true
  try {
    const data = await adminApi.get<Rule[]>('/admin/cancellation-rules')
    templates.value = data || []
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    isLoading.value = false
  }
}

async function loadDefaultRules() {
  try {
    const data = await adminApi.get<Rule[]>('/admin/cancellation-rules/defaults')
    if (data && data.length > 0) {
      editRules.value = data.map(r => ({
        rule_name: r.rule_name,
        days_before_min: r.days_before_min,
        days_before_max: r.days_before_max,
        refund_percentage: r.refund_percentage,
        description: r.description,
        is_template: true,
      }))
      editingId.value = null
      showCreateDialog.value = true
    }
  } catch {
    ElMessage.info('暂无默认规则')
  }
}

function handleCreate() {
  editRules.value = [
    { rule_name: '', days_before_min: 30, days_before_max: null, refund_percentage: 100, description: '', is_template: true },
  ]
  editingId.value = null
  showCreateDialog.value = true
}

function addRule() {
  editRules.value.push({
    rule_name: '',
    days_before_min: 0,
    days_before_max: null,
    refund_percentage: 0,
    description: '',
    is_template: true,
  })
}

async function saveRules() {
  if (editRules.value.length === 0) {
    ElMessage.warning('请添加至少一条规则')
    return
  }
  saving.value = true
  try {
    await adminApi.post('/admin/cancellation-rules', {
      rules: editRules.value.map(r => ({
        rule_name: r.rule_name,
        days_before_min: r.days_before_min,
        days_before_max: r.days_before_max,
        refund_percentage: r.refund_percentage,
        description: r.description,
      })),
    })
    ElMessage.success('保存成功')
    showCreateDialog.value = false
    loadTemplates()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function assignToProduct() {
  const templateIDs = templates.value.filter(t => t.is_template && t.id).map(t => t.id!)
  if (templateIDs.length === 0) {
    ElMessage.warning('没有可用的模板')
    return
  }
  assigning.value = true
  try {
    await adminApi.post(`/admin/cancellation-rules/assign?product_id=${assignProductId.value}`, {
      template_ids: templateIDs,
    })
    ElMessage.success('分配成功')
  } catch (e: any) {
    ElMessage.error(e.message || '分配失败')
  } finally {
    assigning.value = false
  }
}

onMounted(() => {
  loadTemplates()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.percentage {
  color: #e6a23c;
  font-weight: 600;
}
.rule-editor {
  max-height: 400px;
  overflow-y: auto;
}
.rule-row {
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 8px;
}
h4 {
  margin: 0;
  color: #303133;
}
</style>

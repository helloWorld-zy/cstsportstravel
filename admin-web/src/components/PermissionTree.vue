<template>
  <div class="permission-tree">
    <div class="tree-header">
      <el-checkbox
        v-model="selectAll"
        :indeterminate="isIndeterminate"
        @change="handleSelectAll"
      >
        全选
      </el-checkbox>
      <el-button size="small" text @click="handleSelectNone">清空</el-button>
    </div>

    <el-tree
      ref="treeRef"
      :data="data"
      :props="treeProps"
      show-checkbox
      node-key="id"
      :default-expand-all="expandAll"
      :default-checked-keys="modelValue"
      @check="handleCheck"
    >
      <template #default="{ data }">
        <div class="tree-node">
          <span class="node-label">{{ data.permission_name || data.menu_name }}</span>
          <el-tag
            v-if="data.permission_type"
            :type="typeTagMap[data.permission_type] || 'info'"
            size="small"
            class="node-tag"
          >
            {{ typeLabelMap[data.permission_type] || data.permission_type }}
          </el-tag>
          <span v-if="data.permission_code" class="node-code">{{ data.permission_code }}</span>
        </div>
      </template>
    </el-tree>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import type { ElTree } from 'element-plus'

export interface TreeNode {
  id: number
  permission_name?: string
  menu_name?: string
  permission_code?: string
  permission_type?: string
  children?: TreeNode[]
}

const props = withDefaults(defineProps<{
  /** Tree data (permission tree or menu tree). */
  data: TreeNode[]
  /** Currently selected IDs (v-model). */
  modelValue?: number[]
  /** Whether to expand all nodes by default. */
  expandAll?: boolean
}>(), {
  modelValue: () => [],
  expandAll: true,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: number[]): void
}>()

const treeRef = ref<InstanceType<typeof ElTree>>()
const selectAll = ref(false)
const isIndeterminate = ref(false)

const treeProps = {
  children: 'children',
  label: 'permission_name',
}

const typeTagMap: Record<string, string> = {
  menu: '',
  button: 'success',
  api: 'warning',
  data: 'danger',
}

const typeLabelMap: Record<string, string> = {
  menu: '菜单',
  button: '按钮',
  api: 'API',
  data: '数据',
}

// Flatten all leaf node IDs
function getAllLeafIds(nodes: TreeNode[]): number[] {
  const ids: number[] = []
  for (const node of nodes) {
    if (!node.children?.length) {
      ids.push(node.id)
    } else {
      ids.push(...getAllLeafIds(node.children))
    }
  }
  return ids
}

// Sync checked state when modelValue changes externally
watch(() => props.modelValue, async (newVal) => {
  await nextTick()
  if (treeRef.value) {
    treeRef.value.setCheckedKeys(newVal, false)
    updateSelectState()
  }
}, { immediate: true })

function handleCheck() {
  if (!treeRef.value) return
  const checkedKeys = treeRef.value.getCheckedKeys(false) as number[]
  emit('update:modelValue', checkedKeys)
  updateSelectState()
}

function handleSelectAll(checked: boolean) {
  if (!treeRef.value) return
  if (checked) {
    const allIds = getAllLeafIds(props.data)
    treeRef.value.setCheckedKeys(allIds, false)
    emit('update:modelValue', allIds)
  } else {
    treeRef.value.setCheckedKeys([], false)
    emit('update:modelValue', [])
  }
  isIndeterminate.value = false
}

function handleSelectNone() {
  if (!treeRef.value) return
  treeRef.value.setCheckedKeys([], false)
  emit('update:modelValue', [])
  selectAll.value = false
  isIndeterminate.value = false
}

function updateSelectState() {
  if (!treeRef.value) return
  const allIds = getAllLeafIds(props.data)
  const checkedCount = treeRef.value.getCheckedKeys(false).length
  selectAll.value = checkedCount === allIds.length && allIds.length > 0
  isIndeterminate.value = checkedCount > 0 && checkedCount < allIds.length
}
</script>

<style scoped>
.permission-tree {
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 12px;
}

.tree-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 8px;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.node-label {
  font-size: 14px;
}

.node-tag {
  margin-left: 4px;
}

.node-code {
  color: #909399;
  font-size: 12px;
  margin-left: auto;
}
</style>

<template>
  <div class="admin-orgs" v-loading="loading">
    <div class="page-header">
      <h3>机构管理</h3>
      <div class="header-actions">
        <el-button type="primary" @click="openAddChild('branch')">新增分支机构</el-button>
        <el-button @click="openAddChild('representative')">新增代表机构</el-button>
      </div>
    </div>

    <el-card>
      <div v-if="!loading && rootNode" class="root-card">
        <!-- Root -->
        <div class="root-header">
          <div class="root-info">
            <el-icon :size="22" color="#3b82f6"><OfficeBuilding /></el-icon>
            <span class="root-name">{{ rootNode.name }}</span>
            <el-tag size="small" type="info" effect="plain">上级机构</el-tag>
          </div>
          <div class="root-actions">
            <el-button text size="small" type="primary" @click="editRoot">
              <el-icon><Edit /></el-icon> 改名
            </el-button>
          </div>
        </div>

        <!-- Children -->
        <div class="children-list" v-if="children.length > 0">
          <div v-for="child in children" :key="child.id" class="child-item">
            <div class="child-info">
              <el-tag
                size="small"
                :type="child.type === 'branch' ? 'primary' : 'success'"
                effect="dark"
              >
                {{ child.type === 'branch' ? '分支机构' : '代表机构' }}
              </el-tag>
              <span class="child-name">{{ child.name }}</span>
            </div>
            <div class="child-actions">
              <el-button text size="small" type="warning" @click="editChild(child)">
                <el-icon><Edit /></el-icon> 编辑
              </el-button>
              <el-button text size="small" type="danger" @click="delChild(child)">
                <el-icon><Delete /></el-icon> 删除
              </el-button>
            </div>
          </div>
        </div>
        <div class="children-empty" v-else>
          <span class="empty-hint">暂无下属机构，请点击上方按钮新增</span>
        </div>
      </div>
      <el-empty v-else-if="!loading && !rootNode" description="尚未创建机构" />
    </el-card>

    <!-- Dialog: Edit Root -->
    <el-dialog v-model="showRootDialog" title="修改机构名称" width="400px">
      <el-form :model="rootForm" size="large">
        <el-form-item label="名称" required>
          <el-input v-model="rootForm.name" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRootDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveRoot">保存</el-button>
      </template>
    </el-dialog>

    <!-- Dialog: Add / Edit Child -->
    <el-dialog v-model="showChildDialog" :title="editingChild ? '编辑' : (childType === 'branch' ? '新增分支机构' : '新增代表机构')" width="420px">
      <el-form :model="childForm" size="large">
        <el-form-item label="类型" v-if="!editingChild">
          <el-tag :type="childType === 'branch' ? 'primary' : 'success'">
            {{ childType === 'branch' ? '分支机构' : '代表机构' }}
          </el-tag>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="childForm.name" :placeholder="childType === 'branch' ? '如：变压器分会' : '如：华北代表处'" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="childForm.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="childForm.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showChildDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveChild">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { orgApi } from '@/api/index'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit, Delete } from '@element-plus/icons-vue'

const tree = ref<any[]>([])
const loading = ref(true)
const saving = ref(false)

const rootNode = computed(() => tree.value[0] || null)
const children = computed(() => rootNode.value?.children || [])

// Root dialog
const showRootDialog = ref(false)
const rootForm = reactive({ name: '' })

// Child dialog
const showChildDialog = ref(false)
const editingChild = ref<any>(null)
const childType = ref<'branch' | 'representative'>('branch')
const childForm = reactive({ name: '', description: '', sort: 0 })

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const r = await orgApi.getTree()
    tree.value = r.data || []
  } catch {} finally { loading.value = false }
}

// ---- Root operations ----

function editRoot() {
  rootForm.name = rootNode.value?.name || ''
  showRootDialog.value = true
}

async function saveRoot() {
  if (!rootForm.name) { ElMessage.warning('请输入名称'); return }
  saving.value = true
  try {
    await adminApi.updateOrg(rootNode.value.id, { name: rootForm.name })
    ElMessage.success('名称已修改')
    showRootDialog.value = false
    fetchData()
  } catch {} finally { saving.value = false }
}

// ---- Child operations ----

function openAddChild(type: 'branch' | 'representative') {
  childType.value = type
  editingChild.value = null
  childForm.name = ''
  childForm.description = ''
  childForm.sort = 0
  showChildDialog.value = true
}

function editChild(data: any) {
  childType.value = data.type
  editingChild.value = data
  childForm.name = data.name
  childForm.description = data.description
  childForm.sort = data.sort || 0
  showChildDialog.value = true
}

async function saveChild() {
  if (!childForm.name) { ElMessage.warning('请输入名称'); return }
  saving.value = true
  try {
    if (editingChild.value) {
      await adminApi.updateOrg(editingChild.value.id, { name: childForm.name, description: childForm.description })
      ElMessage.success('修改成功')
    } else {
      await adminApi.createOrg({
        name: childForm.name,
        parent_id: rootNode.value.id,
        type: childType.value,
        description: childForm.description,
        sort: childForm.sort
      })
      ElMessage.success(childType.value === 'branch' ? '分支机构创建成功' : '代表机构创建成功')
    }
    showChildDialog.value = false
    editingChild.value = null
    fetchData()
  } catch {} finally { saving.value = false }
}

async function delChild(data: any) {
  const label = data.type === 'branch' ? '分支机构' : '代表机构'
  try {
    await ElMessageBox.confirm(`确认删除${label}「${data.name}」？`, '警告', { type: 'warning', confirmButtonText: '删除' })
    await adminApi.deleteOrg(data.id)
    ElMessage.success('已删除')
    fetchData()
  } catch {}
}
</script>

<style scoped lang="scss">
.admin-orgs { max-width: 800px; }
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  .header-actions { display: flex; gap: 12px; }
}

.root-card {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  overflow: hidden;
}

.root-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18px 24px;
  background: #f0f4f8;
  border-bottom: 1px solid #e5e7eb;

  .root-info {
    display: flex;
    align-items: center;
    gap: 10px;
    .root-name {
      font-size: 17px;
      font-weight: 700;
      color: #1f2937;
    }
  }
}

.children-list {
  padding: 4px 0;
}

.child-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 24px 14px 32px;
  border-bottom: 1px solid #f3f4f6;
  transition: background 0.15s;

  &:last-child { border-bottom: none; }
  &:hover { background: #f9fafb; }

  .child-info {
    display: flex;
    align-items: center;
    gap: 10px;
    .child-name {
      font-size: 14px;
      color: #374151;
    }
  }
  .child-actions {
    display: flex;
    gap: 4px;
    opacity: 0;
    transition: opacity 0.15s;
  }
  &:hover .child-actions { opacity: 1; }
}

.children-empty {
  padding: 24px;
  text-align: center;
  .empty-hint { color: #9ca3af; font-size: 14px; }
}
</style>

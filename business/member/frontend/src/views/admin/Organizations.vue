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
          <div class="root-levels" v-if="rootNode.levels?.length">
            <el-tag v-for="lvl in rootNode.levels" :key="lvl.id" size="small" effect="plain" round>{{ lvl.level?.name || lvl.name }}</el-tag>
          </div>
          <div class="root-actions">
            <el-button text size="small" type="primary" @click="editRoot">
              <el-icon><Edit /></el-icon> 编辑
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
              <div class="child-levels" v-if="child.levels?.length">
                <el-tag v-for="lvl in child.levels" :key="lvl.id" size="small" effect="plain" round>{{ lvl.level?.name || lvl.name }}</el-tag>
              </div>
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
    <el-dialog v-model="showRootDialog" title="编辑上级机构" width="420px">
      <el-form :model="rootForm" size="large">
        <el-form-item label="名称" required>
          <el-input v-model="rootForm.name" />
        </el-form-item>
        <el-form-item label="关联等级">
          <el-select v-model="rootForm.levelIds" multiple placeholder="选择关联的会员等级" style="width:100%">
            <el-option v-for="l in allLevels" :key="l.id" :label="l.name" :value="l.id" />
          </el-select>
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
        <el-form-item label="关联等级">
          <el-select v-model="childForm.levelIds" multiple placeholder="选择关联的会员等级" style="width:100%">
            <el-option v-for="l in allLevels" :key="l.id" :label="l.name" :value="l.id" />
          </el-select>
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
const allLevels = ref<any[]>([])
const loading = ref(true)
const saving = ref(false)

const rootNode = computed(() => tree.value[0] || null)
const children = computed(() => rootNode.value?.children || [])

// Root dialog
const showRootDialog = ref(false)
const rootForm = reactive({ name: '', levelIds: [] as number[] })

// Child dialog
const showChildDialog = ref(false)
const editingChild = ref<any>(null)
const childType = ref<'branch' | 'representative'>('branch')
const childForm = reactive({ name: '', description: '', sort: 0, levelIds: [] as number[] })

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const [treeRes, levelsRes] = await Promise.all([
      orgApi.getTree(),
      adminApi.getMemberLevels()
    ])
    tree.value = treeRes.data || []
    allLevels.value = levelsRes.data || []
  } catch {} finally { loading.value = false }
}

// ---- Root operations ----

function editRoot() {
  const node = rootNode.value
  rootForm.name = node?.name || ''
  rootForm.levelIds = node?.levels?.map((l: any) => l.level_id || l.id) || []
  showRootDialog.value = true
}

async function saveRoot() {
  if (!rootForm.name) { ElMessage.warning('请输入名称'); return }
  saving.value = true
  try {
    await adminApi.updateOrg(rootNode.value.id, { name: rootForm.name })
    // Save level associations
    if (rootNode.value.id && rootForm.levelIds.length > 0) {
      await adminApi.setOrgLevels(rootNode.value.id, rootForm.levelIds)
    } else if (rootNode.value.id) {
      await adminApi.setOrgLevels(rootNode.value.id, [])
    }
    ElMessage.success('已保存')
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
  childForm.levelIds = []
  showChildDialog.value = true
}

async function editChild(data: any) {
  childType.value = data.type
  editingChild.value = data
  childForm.name = data.name
  childForm.description = data.description
  childForm.sort = data.sort || 0
  // Load existing level associations
  childForm.levelIds = data.levels?.map((l: any) => l.level_id || l.id) || []
  // If levels haven't been fetched yet, fetch them
  if (!data.levels && data.id) {
    try {
      const r = await adminApi.getOrgLevels(data.id)
      childForm.levelIds = (r.data || []).map((l: any) => l.level_id)
    } catch {}
  }
  showChildDialog.value = true
}

async function saveChild() {
  if (!childForm.name) { ElMessage.warning('请输入名称'); return }
  saving.value = true
  try {
    let orgId: number
    if (editingChild.value) {
      await adminApi.updateOrg(editingChild.value.id, { name: childForm.name, description: childForm.description })
      orgId = editingChild.value.id
      ElMessage.success('修改成功')
    } else {
      const r = await adminApi.createOrg({
        name: childForm.name,
        parent_id: rootNode.value.id,
        type: childType.value,
        description: childForm.description,
        sort: childForm.sort
      })
      orgId = r.data?.id
      ElMessage.success(childType.value === 'branch' ? '分支机构创建成功' : '代表机构创建成功')
    }
    // Save level associations
    if (orgId && childForm.levelIds.length > 0) {
      await adminApi.setOrgLevels(orgId, childForm.levelIds)
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
.admin-orgs {
 width: 100%;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  h3 {
    font-size: 22px;
    font-weight: 600;
    color: #1a1a2e;
    margin: 0;
  }
  .header-actions { display: flex; gap: 12px; }
}

.el-card {
  border-radius: 10px;
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
  .root-levels {
    margin-left: 12px;
    display: flex;
    align-items: center;
    gap: 6px;
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

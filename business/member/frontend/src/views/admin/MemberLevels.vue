<template>
  <div class="admin-levels" v-loading="loading">
    <div class="page-header">
      <h3>会员等级管理</h3>
      <el-button type="primary" @click="openCreate">新增等级</el-button>
    </div>
    <el-card>
      <el-table :data="list" stripe row-class-name="level-row">
        <el-table-column type="index" label="排序" width="80">
          <template #default="{ $index }">{{ $index + 1 }}</template>
        </el-table-column>
        <el-table-column prop="name" label="等级名称" min-width="200">
          <template #default="{ row }">
            <span class="level-badge" :style="{ background: levelColor(row.level) }">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="250" />
        <el-table-column label="操作" width="200">
          <template #default="{ row, $index }">
            <el-button text size="small" type="primary" :disabled="$index === 0" @click="moveUp(row)">
              <el-icon><Top /></el-icon> 上移
            </el-button>
            <el-button text size="small" type="primary" :disabled="$index === list.length - 1" @click="moveDown(row)">
              <el-icon><Bottom /></el-icon> 下移
            </el-button>
            <el-button text size="small" type="warning" @click="editLevel(row)">编辑</el-button>
            <el-button text size="small" type="danger" @click="delLevel(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无会员等级" />
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="showDialog" :title="editingId ? '编辑等级' : '新增等级'" width="480px">
      <el-form :model="levelForm" label-width="80px" size="large">
        <el-form-item label="等级名称" required>
          <el-input v-model="levelForm.name" placeholder="如：会员单位、理事单位" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="levelForm.description" type="textarea" :rows="3" placeholder="对该等级的说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveLevel">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Top, Bottom } from '@element-plus/icons-vue'

const list = ref<any[]>([])
const loading = ref(true)
const showDialog = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)

const levelForm = reactive({ name: '', description: '' })

const levelColors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#b37feb']

function levelColor(level: number) {
  return levelColors[level % levelColors.length]
}

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const res = await adminApi.getMemberLevels()
    list.value = res.data || []
  } catch {} finally { loading.value = false }
}

function openCreate() {
  editingId.value = null
  levelForm.name = ''
  levelForm.description = ''
  showDialog.value = true
}

function editLevel(row: any) {
  editingId.value = row.id
  levelForm.name = row.name
  levelForm.description = row.description
  showDialog.value = true
}

async function saveLevel() {
  if (!levelForm.name) { ElMessage.warning('请输入等级名称'); return }
  saving.value = true
  try {
    if (editingId.value) {
      await adminApi.updateLevelDefinition(editingId.value, { name: levelForm.name, description: levelForm.description })
      ElMessage.success('更新成功')
    } else {
      await adminApi.createMemberLevel({ name: levelForm.name, description: levelForm.description })
      ElMessage.success('创建成功')
    }
    showDialog.value = false
    fetchData()
  } catch {} finally { saving.value = false }
}

async function delLevel(row: any) {
  try {
    await ElMessageBox.confirm(`确认删除等级「${row.name}」？`, '警告', { type: 'warning', confirmButtonText: '删除' })
    await adminApi.deleteMemberLevel(row.id)
    ElMessage.success('已删除')
    fetchData()
  } catch {}
}

async function moveUp(row: any) {
  try {
    await adminApi.moveLevelUp(row.id)
    ElMessage.success('上移成功')
    fetchData()
  } catch {}
}

async function moveDown(row: any) {
  try {
    await adminApi.moveLevelDown(row.id)
    ElMessage.success('下移成功')
    fetchData()
  } catch {}
}
</script>

<style scoped lang="scss">
.admin-levels {
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
}
.el-card {
  border-radius: 10px;
}
.level-badge {
  display: inline-block;
  padding: 4px 14px;
  border-radius: 20px;
  color: #fff;
  font-size: 13px;
  font-weight: 500;
}
</style>

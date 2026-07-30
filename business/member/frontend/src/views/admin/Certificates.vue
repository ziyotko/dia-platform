<template>
  <div class="admin-certs" v-loading="loading">
    <div class="page-header">
      <h3>证书样式管理</h3>
      <el-button type="primary" @click="openCreate" :disabled="availableLevels.length === 0">
        {{ availableLevels.length === 0 ? '所有等级已创建' : '新增样式' }}
      </el-button>
    </div>

    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="name" label="样式名称" min-width="180" />
        <el-table-column label="所属等级" width="150">
          <template #default="{ row }">{{ row.level?.name || '—' }}</template>
        </el-table-column>
        <el-table-column label="模板文件" width="180">
          <template #default="{ row }">
            <a v-if="row.template_file" :href="row.template_file" target="_blank" class="file-link">
              <el-icon><Document /></el-icon> 查看PDF
            </a>
            <span v-else class="no-file">未上传</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button text size="small" type="primary" @click="editTemplate(row)">编辑</el-button>
            <el-button text size="small" type="danger" @click="deleteTemplate(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无证书样式" />
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="showDialog" :title="editingId ? '编辑样式' : '新增样式'" width="600px">
      <el-form :model="form" label-width="100px" size="large">
        <el-form-item label="样式名称" required>
          <el-input v-model="form.name" placeholder="如：理事单位证书、会员单位证书" />
        </el-form-item>
        <el-form-item label="证书分类" required>
          <el-select v-model="form.levelId" placeholder="选择会员等级" style="width:100%" :disabled="!!editingId">
            <el-option
              v-for="l in editingId ? levels : availableLevels"
              :key="l.id" :label="l.name" :value="l.id"
            />
          </el-select>
          <div v-if="!editingId && availableLevels.length === 0" class="level-tip">所有等级已有证书样式</div>
        </el-form-item>
        <el-form-item label="PDF模板">
          <el-upload
            :show-file-list="false"
            :before-upload="uploadPdf"
            accept=".pdf"
          >
            <div class="upload-area">
              <template v-if="form.templateFile">
                <el-icon size="28" color="#409eff"><Document /></el-icon>
                <span>已上传模板</span>
                <el-button text size="small" type="primary" @click.stop="previewPdf">预览</el-button>
              </template>
              <template v-else>
                <el-icon size="28"><Plus /></el-icon>
                <span>上传PDF模板</span>
                <span class="upload-hint">A4横向（297×210mm）</span>
              </template>
            </div>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveTemplate">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { authApi } from '@/api/auth'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Plus } from '@element-plus/icons-vue'

const list = ref<any[]>([])
const levels = ref<any[]>([])
const loading = ref(true)
const showDialog = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)

const form = reactive({
  name: '',
  levelId: null as number | null,
  templateFile: ''
})

/** Levels that don't have a template yet — for create mode */
const availableLevels = computed(() => {
  const takenIds = new Set(list.value.map((t: any) => t.level_id))
  return levels.value.filter((l: any) => !takenIds.has(l.id))
})

function formatDate(d: string) { return d ? d.replace('T', ' ').slice(0, 16) : '' }

onMounted(async () => {
  try {
    const [tplRes, lvRes] = await Promise.all([
      adminApi.getCertTemplates(),
      adminApi.getMemberLevels()
    ])
    list.value = tplRes.data || []
    levels.value = lvRes.data || []
  } catch {} finally { loading.value = false }
})

function openCreate() {
  editingId.value = null
  form.name = ''
  form.levelId = null
  form.templateFile = ''
  showDialog.value = true
}

function editTemplate(row: any) {
  editingId.value = row.id
  form.name = row.name
  form.levelId = row.level_id
  form.templateFile = row.template_file || ''
  showDialog.value = true
}

async function uploadPdf(file: File) {
  const fd = new FormData()
  fd.append('file', file)
  fd.append('dir', 'templates')
  try {
    const res = await authApi.upload(fd)
    form.templateFile = res.data?.url || ''
    ElMessage.success('模板上传成功')
  } catch {} finally { return false }
}

function previewPdf() {
  if (form.templateFile) window.open(form.templateFile, '_blank')
}

async function saveTemplate() {
  if (!form.name) { ElMessage.warning('请输入样式名称'); return }
  if (!form.levelId) { ElMessage.warning('请选择证书分类'); return }
  saving.value = true
  try {
    const data = {
      name: form.name,
      level_id: form.levelId,
      template_file: form.templateFile
    }
    if (editingId.value) {
      await adminApi.updateCertTemplate(editingId.value, data)
      ElMessage.success('更新成功')
    } else {
      await adminApi.createCertTemplate(data)
      ElMessage.success('创建成功')
    }
    showDialog.value = false
    fetchData()
  } catch {} finally { saving.value = false }
}

async function fetchData() {
  loading.value = true
  try {
    const res = await adminApi.getCertTemplates()
    list.value = res.data || []
  } catch {} finally { loading.value = false }
}

async function deleteTemplate(row: any) {
  try {
    await ElMessageBox.confirm(`确认删除样式「${row.name}」？`, '警告', { type: 'warning', confirmButtonText: '删除' })
    await adminApi.deleteCertTemplate(row.id)
    ElMessage.success('已删除')
    fetchData()
  } catch {}
}
</script>

<style scoped lang="scss">
.admin-certs {
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
.file-link {
  display: inline-flex; align-items: center; gap: 4px; color: #409eff; text-decoration: none;
  &:hover { text-decoration: underline; }
}
.no-file { color: #999; font-size: 13px; }
.upload-area {
  display: flex; flex-direction: column; align-items: center; gap: 6px;
  padding: 24px; border: 2px dashed #dcdfe6; border-radius: 8px;
  cursor: pointer; transition: border-color .2s; color: #666;
  &:hover { border-color: #409eff; color: #409eff; }
  .upload-hint { font-size: 12px; color: #999; }
}
.level-tip { font-size: 12px; color: #999; margin-top: 4px; }
</style>

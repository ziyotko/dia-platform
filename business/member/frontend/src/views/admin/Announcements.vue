<template>
  <div class="admin-anns" v-loading="loading">
    <div class="page-header"><h3>公告管理</h3><el-button type="primary" @click="openCreate">发布公告</el-button></div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="title" label="标题" min-width="200" />
        <el-table-column label="类型" width="80"><template #default="{row}">{{ {notice:'通知',article:'文章',policy:'政策'}[row.type] || row.type }}</template></el-table-column>
        <el-table-column prop="is_pinned" label="置顶" width="70"><template #default="{row}"><el-tag size="small" :type="row.is_pinned?'danger':''">{{ row.is_pinned?'是':'否' }}</el-tag></template></el-table-column>
        <el-table-column prop="view_count" label="浏览" width="70" />
        <el-table-column prop="published_at" label="发布时间" width="160"><template #default="{row}">{{ row.published_at?.slice(0,16) || '未发布' }}</template></el-table-column>
        <el-table-column label="操作" width="180"><template #default="{row}">
          <el-button text size="small" type="primary" @click="editAnn(row)">编辑</el-button>
          <el-button text size="small" type="danger" @click="delAnn(row)">删除</el-button>
        </template></el-table-column>
      </el-table>
      <div class="pagination"><el-pagination background layout="prev, pager, next" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" /></div>
    </el-card>

    <el-dialog v-model="showDialog" :title="editingId?'编辑公告':'发布公告'" width="640px">
      <el-form :model="annForm" size="large" label-width="80px">
        <el-form-item label="标题"><el-input v-model="annForm.title" /></el-form-item>
        <el-form-item label="类型"><el-select v-model="annForm.type"><el-option label="通知" value="notice" /><el-option label="文章" value="article" /><el-option label="政策" value="policy" /></el-select></el-form-item>
        <el-form-item label="内容"><el-input v-model="annForm.content" type="textarea" :rows="6" /></el-form-item>
        <el-form-item label="置顶"><el-switch v-model="annForm.isPinned" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="showDialog=false">取消</el-button><el-button type="primary" @click="saveAnn">发布</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([]); const loading = ref(true); const showDialog = ref(false); const editingId = ref<number|null>(null)
const page = ref(1); const size = ref(10); const total = ref(0)
const annForm = reactive({ title: '', content: '', type: 'notice', isPinned: false })

onMounted(() => fetchData())
async function fetchData() {
  loading.value = true
  try { const r = await adminApi.getAnnouncements({ page: page.value, size: size.value }); list.value = r.data?.list || []; total.value = r.data?.total || 0 } catch {} finally { loading.value = false }
}
function openCreate() { editingId.value = null; Object.assign(annForm, { title: '', content: '', type: 'notice', isPinned: false }); showDialog.value = true }
function editAnn(row: any) { editingId.value = row.id; Object.assign(annForm, { title: row.title, content: row.content, type: row.type, isPinned: row.is_pinned }); showDialog.value = true }
async function saveAnn() {
  if (!annForm.title) { ElMessage.warning('请输入标题'); return }
  try {
    const data = { ...annForm, publish_now: true }
    if (editingId.value) { await adminApi.updateAnnouncement(editingId.value, data) }
    else { await adminApi.createAnnouncement(data) }
    ElMessage.success('保存成功'); showDialog.value = false; fetchData()
  } catch {}
}
async function delAnn(row: any) {
  try { await ElMessageBox.confirm('确认删除？','警告',{type:'warning'}); await adminApi.deleteAnnouncement(row.id); ElMessage.success('已删除'); fetchData() } catch {}
}
</script>

<style scoped lang="scss">
.admin-anns { max-width: 1000px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.pagination { display: flex; justify-content: center; margin-top: 24px; }
</style>

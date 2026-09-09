<template>
  <div class="admin-articles" v-loading="loading">
    <div class="page-header"><h3>文章管理</h3></div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column label="标题" min-width="180">
          <template #default="{row}">
            <el-link type="primary" :underline="false" @click="viewArticle(row)">{{ row.title }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="来源单位/个人" width="240" show-overflow-tooltip><template #default="{row}">{{ sourceLabel(row.member) }}</template></el-table-column>
        <el-table-column prop="member.username" label="作者" width="100" />
        <el-table-column prop="category.name" label="分类" width="100" />
        <el-table-column prop="status" label="状态" width="100"><template #default="{row}"><el-tag :type="row.status==='pending'?'warning':row.status==='published'?'success':'info'">{{ statusLabel(row.status) }}</el-tag></template></el-table-column>
        <el-table-column prop="created_at" label="时间" width="160"><template #default="{row}">{{ row.created_at?.slice(0,16).replace('T',' ') }}</template></el-table-column>
        <el-table-column label="操作" width="220" v-if="list.length"><template #default="{row}">
          <template v-if="row.status==='pending'">
            <el-button size="small" type="success" @click="review(row, true)">通过</el-button>
            <el-button size="small" type="danger" @click="review(row, false)">拒绝</el-button>
          </template>
          <span v-else style="color:#9ca3af"></span>
          <el-button size="small" type="danger" plain @click="deleteRow(row)">删除</el-button>
        </template></el-table-column>
      </el-table>
      <div class="pagination"><el-pagination background layout="prev, pager, next" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" /></div>
    </el-card>

    <!-- View Article Dialog -->
    <el-dialog v-model="showView" :title="viewed?.title || '文章内容'" width="800px" top="5vh">
      <div class="view-meta" v-if="viewed">
        <el-tag size="small" :type="viewed.status==='pending'?'warning':viewed.status==='published'?'success':'info'">{{ statusLabel(viewed.status) }}</el-tag>
        <span>来源：{{ sourceLabel(viewed.member) }}</span>
        <span>作者：{{ viewed.member?.username }}</span>
        <span v-if="viewed.category">分类：{{ viewed.category.name }}</span>
        <span>时间：{{ viewed.created_at?.slice(0,16).replace('T',' ') }}</span>
      </div>
      <div class="view-body">
        <img v-if="viewed?.cover_image" :src="viewed.cover_image" class="view-cover" />
        <div class="view-summary" v-if="viewed?.summary">{{ viewed.summary }}</div>
        <div class="view-content" v-html="viewed?.content || '<p style=&quot;color:#9ca3af&quot;>暂无内容</p>'"></div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([]); const loading = ref(true)
const page = ref(1); const size = ref(10); const total = ref(0)
const showView = ref(false); const viewed = ref<any>(null)

const sm: Record<string,string> = { draft:'草稿', pending:'待审核', published:'已发布', rejected:'已拒绝' }
function statusLabel(s: string) { return sm[s] || s }
function sourceLabel(member: any) {
  if (!member) return '-'
  if (member.member_type === 'unit') return member.company_name || member.username || '-'
  return member.name || member.username || '-'
}
function viewArticle(row: any) { viewed.value = row; showView.value = true }

onMounted(() => fetchData())
async function fetchData() {
  loading.value = true
  try { const r = await adminApi.getArticles({ page: page.value, size: size.value }); list.value = r.data?.list || []; total.value = r.data?.total || 0 } catch {} finally { loading.value = false }
}
async function review(row: any, approved: boolean) {
  if (approved) {
    try {
      await ElMessageBox.confirm(
        `确认通过文章「${row.title}」的审核？通过后将公开发布。`,
        '通过确认',
        { type: 'warning', confirmButtonText: '确认通过', cancelButtonText: '取消' }
      )
    } catch { return }
  }
  try {
    const comment = approved ? '' : (await ElMessageBox.prompt('拒绝理由', '拒绝')).value || ''
    await adminApi.reviewArticle(row.id, { approved, comment })
    ElMessage.success(approved ? '已发布' : '已拒绝'); fetchData()
  } catch {}
}
async function deleteRow(row: any) {
  try {
    await ElMessageBox.confirm(
      `确认删除文章「${row.title}」？删除后不可恢复。`,
      '删除确认',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消', confirmButtonClass: 'el-button--danger' }
    )
  } catch { return }
  try {
    await adminApi.deleteArticle(row.id)
    ElMessage.success('已删除')
    fetchData()
  } catch {}
}
</script>

<style scoped lang="scss">
.admin-articles {
 width: 100%;
}
.page-header {
  margin-bottom: 24px;
  h3 {
    font-size: 22px;
    font-weight: 600;
    color: #1d2739;
    margin: 0;
  }
}
.el-card {
  border-radius: 10px;
}
.pagination { display: flex; justify-content: center; padding: 20px 0; }
.view-meta { display: flex; flex-wrap: wrap; gap: 16px; align-items: center; margin-bottom: 16px; font-size: 13px; color: #6b7280; }
.view-body {
  max-height: 70vh;
  overflow-y: auto;
}
.view-cover {
  width: 100%;
  max-height: 300px;
  object-fit: cover;
  border-radius: 8px;
  margin-bottom: 16px;
}
.view-summary {
  color: #6b7280;
  font-size: 14px;
  line-height: 1.7;
  background: #f8fafc;
  border-left: 3px solid #002fa7;
  padding: 10px 14px;
  border-radius: 4px;
  margin-bottom: 16px;
}
.view-content {
  line-height: 1.9;
  font-size: 15px;
  color: #303133;
  :deep(img) { max-width: 100%; border-radius: 6px; }
  :deep(p) { margin: 0 0 12px; }
  :deep(blockquote) { border-left: 4px solid #dcdfe6; margin: 12px 0; padding: 8px 14px; color: #6b7280; background: #fafafa; }
}
</style>

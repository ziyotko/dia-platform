<template>
  <div class="admin-articles" v-loading="loading">
    <div class="page-header"><h3>文章管理</h3></div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="title" label="标题" min-width="180" />
        <el-table-column prop="member.username" label="作者" width="100" />
        <el-table-column prop="category.name" label="分类" width="100" />
        <el-table-column prop="status" label="状态" width="100"><template #default="{row}"><el-tag :type="row.status==='pending'?'warning':row.status==='published'?'success':'info'">{{ statusLabel(row.status) }}</el-tag></template></el-table-column>
        <el-table-column prop="created_at" label="时间" width="160"><template #default="{row}">{{ row.created_at?.slice(0,16) }}</template></el-table-column>
        <el-table-column label="操作" width="180" v-if="list.length"><template #default="{row}">
          <template v-if="row.status==='pending'">
            <el-button size="small" type="success" @click="review(row, true)">通过</el-button>
            <el-button size="small" type="danger" @click="review(row, false)">拒绝</el-button>
          </template>
          <span v-else style="color:#9ca3af">-</span>
        </template></el-table-column>
      </el-table>
      <div class="pagination"><el-pagination background layout="prev, pager, next" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" /></div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([]); const loading = ref(true)
const page = ref(1); const size = ref(10); const total = ref(0)

const sm: Record<string,string> = { draft:'草稿', pending:'待审核', published:'已发布', rejected:'已拒绝' }
function statusLabel(s: string) { return sm[s] || s }

onMounted(() => fetchData())
async function fetchData() {
  loading.value = true
  try { const r = await adminApi.getArticles({ page: page.value, size: size.value }); list.value = r.data?.list || []; total.value = r.data?.total || 0 } catch {} finally { loading.value = false }
}
async function review(row: any, approved: boolean) {
  try {
    const comment = approved ? '' : (await ElMessageBox.prompt('拒绝理由', '拒绝')).value || ''
    await adminApi.reviewArticle(row.id, { approved, comment })
    ElMessage.success(approved ? '已发布' : '已拒绝'); fetchData()
  } catch {}
}
</script>

<style scoped lang="scss">
.admin-articles { max-width: 1100px; }
.page-header { margin-bottom: 20px; }
.pagination { display: flex; justify-content: center; margin-top: 24px; }
</style>

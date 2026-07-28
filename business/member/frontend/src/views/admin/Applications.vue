<template>
  <div class="admin-apps" v-loading="loading">
    <div class="page-header"><h3>入会审核</h3></div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="member.company_name" label="申请单位" />
        <el-table-column prop="member.username" label="用户名" />
        <el-table-column prop="org.name" label="申请分会" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{row}"><el-tag :type="row.status==='pending_review'?'warning':row.status==='approved'?'success':'danger'">{{ statusLabel(row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="created_at" label="申请时间" width="170"><template #default="{row}">{{ row.created_at?.slice(0,16) }}</template></el-table-column>
        <el-table-column label="操作" width="200" v-if="list.length">
          <template #default="{row}">
            <template v-if="row.status === 'pending_review'">
              <el-button size="small" type="success" @click="review(row, true)">通过</el-button>
              <el-button size="small" type="danger" @click="review(row, false)">拒绝</el-button>
            </template>
            <span v-else style="color:#9ca3af">-</span>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length===0" />
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

const sm: Record<string,string> = { draft:'草稿', pending_review:'待审核', approved:'已通过', rejected:'已拒绝' }
function statusLabel(s: string) { return sm[s] || s }

onMounted(() => fetchData())
async function fetchData() {
  loading.value = true
  try { const r = await adminApi.getApplications({ page: page.value, size: size.value }); list.value = r.data?.list || []; total.value = r.data?.total || 0 } catch {} finally { loading.value = false }
}
async function review(row: any, approved: boolean) {
  try {
    const comment = approved ? '' : (await ElMessageBox.prompt('请输入拒绝理由', '拒绝申请')).value || ''
    await adminApi.reviewApplication(row.id, { approved, comment })
    ElMessage.success(approved ? '已通过' : '已拒绝')
    fetchData()
  } catch {}
}
</script>

<style scoped lang="scss">
.admin-apps { max-width: 1100px; }
.page-header { margin-bottom: 20px; }
.pagination { display: flex; justify-content: center; margin-top: 24px; }
</style>

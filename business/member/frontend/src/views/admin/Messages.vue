<template>
  <div class="admin-msgs" v-loading="loading">
    <div class="page-header"><h3>会员留言管理</h3></div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="member.username" label="会员" />
        <el-table-column prop="title" label="标题" />
        <el-table-column prop="content" label="内容" show-overflow-tooltip />
        <el-table-column prop="reply" label="回复"><template #default="{row}"><span :style="{color:row.reply?'#22c55e':'#9ca3af'}">{{ row.reply || '未回复' }}</span></template></el-table-column>
        <el-table-column prop="created_at" label="时间" width="160"><template #default="{row}">{{ row.created_at?.slice(0,16) }}</template></el-table-column>
        <el-table-column label="操作" width="100"><template #default="{row}"><el-button v-if="!row.reply" text size="small" type="primary" @click="replyMsg(row)">回复</el-button></template></el-table-column>
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

onMounted(() => fetchData())
async function fetchData() {
  loading.value = true
  try { const r = await adminApi.getMessages({ page: page.value, size: size.value }); list.value = r.data?.list || []; total.value = r.data?.total || 0 } catch {} finally { loading.value = false }
}
async function replyMsg(row: any) {
  try { const { value } = await ElMessageBox.prompt('回复内容', '回复留言'); if (value) { await adminApi.replyMessage(row.id, value); ElMessage.success('已回复'); fetchData() } } catch {}
}
</script>

<style scoped lang="scss">
.admin-msgs { max-width: 1100px; }
.page-header { margin-bottom: 20px; }
.pagination { display: flex; justify-content: center; margin-top: 24px; }
</style>

<template>
  <div class="admin-msgs" v-loading="loading">
    <div class="page-header"><h3>会员留言管理</h3></div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="member.username" label="会员" />
        <el-table-column prop="title" label="标题" />
        <el-table-column prop="content" label="内容" show-overflow-tooltip />
        <el-table-column prop="reply" label="回复"><template #default="{row}"><span :style="{color:row.reply?'#22c55e':'#9ca3af'}">{{ row.reply || '未回复' }}</span></template></el-table-column>
        <el-table-column prop="created_at" label="时间" width="160"><template #default="{row}">{{ row.created_at?.replace('T', ' ').slice(0,16) }}</template></el-table-column>
        <el-table-column label="操作" width="100"><template #default="{row}"><el-button v-if="!row.reply" text size="small" type="primary" @click="openReply(row)">回复</el-button></template></el-table-column>
      </el-table>

      <!-- 回复对话框 -->
      <el-dialog v-model="replyVisible" title="回复留言" width="560px">
        <div class="reply-origin" v-if="replyTarget">
          <div class="reply-label">原留言：</div>
          <div class="reply-content">{{ replyTarget.content }}</div>
        </div>
        <el-input
          v-model="replyText"
          type="textarea"
          :rows="6"
          placeholder="请输入回复内容..."
          maxlength="2000"
          show-word-limit
        />
        <template #footer>
          <el-button @click="replyVisible = false">取消</el-button>
          <el-button type="primary" :disabled="!replyText.trim()" @click="submitReply">确认回复</el-button>
        </template>
      </el-dialog>
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
const replyVisible = ref(false)
const replyTarget = ref<any>(null)
const replyText = ref('')

onMounted(() => fetchData())
async function fetchData() {
  loading.value = true
  try { const r = await adminApi.getMessages({ page: page.value, size: size.value }); list.value = r.data?.list || []; total.value = r.data?.total || 0 } catch {} finally { loading.value = false }
}
function openReply(row: any) {
  replyTarget.value = row
  replyText.value = ''
  replyVisible.value = true
}
async function submitReply() {
  if (!replyTarget.value || !replyText.value.trim()) return
  try {
    await adminApi.replyMessage(replyTarget.value.id, replyText.value.trim())
    ElMessage.success('已回复')
    replyVisible.value = false
    fetchData()
  } catch {}
}
</script>

<style scoped lang="scss">
.admin-msgs {
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px 0;
}
.page-header {
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
.pagination { display: flex; justify-content: center; padding: 20px 0; }

.reply-origin {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 16px;
}
.reply-label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 6px;
}
.reply-content {
  font-size: 14px;
  color: #303133;
  line-height: 1.6;
  white-space: pre-wrap;
}
</style>

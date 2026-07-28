<template>
  <div class="messages-page" v-loading="loading">
    <div class="page-header">
      <h3>会员留言</h3>
      <el-button type="primary" @click="showCreate = true">新建留言</el-button>
    </div>

    <el-card>
      <el-table :data="messages" stripe>
        <el-table-column prop="title" label="标题" min-width="180" />
        <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip />
        <el-table-column prop="reply" label="回复" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.reply" style="color:#22c55e">{{ row.reply }}</span>
            <span v-else style="color:#9ca3af">暂无回复</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'replied' ? 'success' : row.status === 'read' ? 'info' : 'warning'">
              {{ row.status === 'replied' ? '已回复' : row.status === 'read' ? '已读' : '未读' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && messages.length === 0" description="暂无留言" />
      <div class="pagination" v-if="total > size">
        <el-pagination background layout="prev, pager, next" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" />
      </div>
    </el-card>

    <el-dialog v-model="showCreate" title="新建留言" width="500px">
      <el-form :model="msgForm" size="large">
        <el-form-item label="标题" required><el-input v-model="msgForm.title" /></el-form-item>
        <el-form-item label="内容" required>
          <el-input v-model="msgForm.content" type="textarea" :rows="5" placeholder="请输入您的建议、诉求..." />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="createMsg">提交留言</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { messageApi } from '@/api/index'
import { ElMessage } from 'element-plus'

const messages = ref<any[]>([])
const loading = ref(true)
const showCreate = ref(false)
const page = ref(1)
const size = ref(10)
const total = ref(0)
const msgForm = reactive({ title: '', content: '' })

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const res = await messageApi.getMyMessages({ page: page.value, size: size.value })
    messages.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch {} finally { loading.value = false }
}

async function createMsg() {
  if (!msgForm.title || !msgForm.content) { ElMessage.warning('请填写完整'); return }
  try {
    await messageApi.createMessage(msgForm)
    ElMessage.success('留言成功')
    showCreate.value = false
    msgForm.title = ''; msgForm.content = ''
    fetchData()
  } catch {}
}

function formatDate(d: string) { return d ? d.slice(0, 16) : '' }
</script>

<style scoped lang="scss">
.messages-page { max-width: 1000px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.pagination { display: flex; justify-content: center; margin-top: 24px; }
</style>

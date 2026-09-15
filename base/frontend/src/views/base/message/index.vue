<template>
  <div class="message-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>消息管理</span>
          <div class="header-actions">
            <el-radio-group v-model="query.isRead" v-if="activeTab === 'inbox'" @change="handleFilterChange">
              <el-radio-button :label="-1">全部</el-radio-button>
              <el-radio-button :label="0">未读</el-radio-button>
              <el-radio-button :label="1">已读</el-radio-button>
            </el-radio-group>
            <el-button v-if="activeTab === 'inbox'" @click="handleMarkAll">全部标为已读</el-button>
            <el-button type="primary" @click="handleSend">发送消息</el-button>
          </div>
        </div>
      </template>
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="收件箱" name="inbox">
          <el-table :data="tableData" v-loading="loading" border>
            <el-table-column prop="title" label="标题" />
            <el-table-column prop="senderName" label="发送者" width="120" />
            <el-table-column prop="type" label="类型" width="100" />
            <el-table-column prop="priority" label="优先级" width="100">
              <template #default="{ row }">
                <el-tag :type="priorityMap[row.priority]?.type">{{ priorityMap[row.priority]?.label }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.isRead ? 'info' : 'danger'">{{ row.isRead ? '已读' : '未读' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="sendAt" label="发送时间" width="180" />
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="handleView(row)">查看</el-button>
                <el-button v-if="!row.isRead" link type="primary" @click="handleRead(row)">标为已读</el-button>
                <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="发件箱" name="sent">
          <el-table :data="tableData" v-loading="loading" border>
            <el-table-column prop="title" label="标题" />
            <el-table-column prop="receiverType" label="接收类型" width="120" />
            <el-table-column prop="type" label="类型" width="100" />
            <el-table-column prop="priority" label="优先级" width="100" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 2 ? 'warning' : 'success'">{{ row.status === 2 ? '草稿' : '已发送' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="sendAt" label="发送时间" width="180" />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.size"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <el-dialog v-model="sendDialogVisible" title="发送消息" width="600px">
      <el-form :model="sendForm" :rules="sendRules" ref="sendFormRef" label-width="100px">
        <el-form-item label="接收类型" prop="receiverType">
          <el-radio-group v-model="sendForm.receiverType">
            <el-radio-button label="user">指定用户</el-radio-button>
            <el-radio-button label="all">全员广播</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="接收用户" v-if="sendForm.receiverType === 'user'">
          <el-select v-model="sendForm.receiverIds" multiple placeholder="请选择" style="width: 100%">
            <el-option v-for="u in userOptions" :key="u.id" :label="u.username" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="发送渠道">
          <el-select v-model="sendForm.channel" style="width: 100%">
            <el-option label="站内信" value="in-app" />
            <el-option label="邮件（额外发送，需在系统设置中配置 SMTP）" value="email" />
          </el-select>
          <div class="form-tip">短信 / 企微渠道尚未接入发送器，故未开放</div>
        </el-form-item>
        <el-form-item label="消息类型">
          <el-select v-model="sendForm.type" style="width: 100%">
            <el-option label="通知" value="notice" />
            <el-option label="系统" value="system" />
            <el-option label="私信" value="private" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="sendForm.priority" style="width: 100%">
            <el-option label="低" value="low" />
            <el-option label="普通" value="normal" />
            <el-option label="高" value="high" />
            <el-option label="紧急" value="urgent" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="sendForm.title" />
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <el-input v-model="sendForm.content" type="textarea" :rows="5" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="sendDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSendSubmit">发送</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="viewDialogVisible" title="消息详情" width="500px">
      <h3>{{ currentMessage?.title }}</h3>
      <p class="message-meta">发送者：{{ currentMessage?.senderName }} | 时间：{{ currentMessage?.sendAt }}</p>
      <div class="message-content">{{ currentMessage?.content }}</div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getMessageList, sendMessage, markMessageRead, markAllMessageRead, deleteMessage } from '@/api/message'
import { getUserList } from '@/api/user'
import type { Message } from '@/api/message'

// 通知外层布局刷新未读红点
const notifyUnreadChanged = () => window.dispatchEvent(new Event('base:unread-changed'))

const loading = ref(false)
const tableData = ref<Message[]>([])
const total = ref(0)
const activeTab = ref('inbox')
const query = reactive<{ page: number; size: number; isRead: number }>({ page: 1, size: 10, isRead: -1 })
const sendDialogVisible = ref(false)
const viewDialogVisible = ref(false)
const currentMessage = ref<Message | null>(null)
const sendFormRef = ref<any>(null)
const userOptions = ref<any[]>([])
const sendForm = reactive({
  receiverType: 'user',
  receiverIds: [] as number[],
  channel: 'in-app',
  type: 'notice',
  priority: 'normal',
  title: '',
  content: ''
})

const priorityMap: Record<string, { label: string; type: any }> = {
  low: { label: '低', type: 'info' },
  normal: { label: '普通', type: '' },
  high: { label: '高', type: 'warning' },
  urgent: { label: '紧急', type: 'danger' }
}

const sendRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入内容', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const isInbox = activeTab.value === 'inbox'
    const res: any = await getMessageList({
      page: query.page,
      size: query.size,
      box: isInbox ? 'inbox' : 'sent',
      isRead: isInbox ? query.isRead : -1
    })
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  query.page = 1
  fetchData()
}

const handleFilterChange = () => {
  query.page = 1
  fetchData()
}

// 全部标为已读
const handleMarkAll = async () => {
  await ElMessageBox.confirm('确认将收件箱中全部未读消息标为已读？', '提示', { type: 'warning' })
  const res: any = await markAllMessageRead()
  ElMessage.success(`已标记 ${res.data?.count ?? 0} 条为已读`)
  fetchData()
  notifyUnreadChanged()
}

const handleSend = async () => {
  sendFormRef.value?.resetFields()
  sendForm.receiverIds = []
  sendForm.receiverType = 'user'
  sendForm.channel = 'in-app'
  sendForm.title = ''
  sendForm.content = ''
  sendDialogVisible.value = true
  const res: any = await getUserList({ page: 1, size: 1000 })
  userOptions.value = res.data.list || []
}

const handleSendSubmit = async () => {
  const valid = await sendFormRef.value?.validate().catch(() => false)
  if (!valid) return
  await sendMessage({ ...sendForm })
  ElMessage.success('发送成功')
  sendDialogVisible.value = false
  fetchData()
}

const handleView = (row: Message) => {
  currentMessage.value = row
  viewDialogVisible.value = true
  if (!row.isRead) {
    markMessageRead(row.id).then(() => {
      fetchData()
      notifyUnreadChanged()
    })
  }
}

const handleRead = async (row: Message) => {
  await markMessageRead(row.id)
  ElMessage.success('已标记为已读')
  fetchData()
  notifyUnreadChanged()
}

const handleDelete = async (row: Message) => {
  await ElMessageBox.confirm('确认删除该消息？', '提示', { type: 'warning' })
  await deleteMessage(row.id)
  ElMessage.success('删除成功')
  fetchData()
  if (!row.isRead) {
    notifyUnreadChanged()
  }
}

onMounted(fetchData)
</script>

<style scoped lang="scss">
.message-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .header-actions {
    display: flex;
    gap: 10px;
    align-items: center;
  }

  .form-tip {
    width: 100%;
    font-size: 12px;
    color: #94a3b8;
    line-height: 1.6;
  }
  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
  .message-meta {
    color: #909399;
    font-size: 13px;
    margin: 8px 0 16px;
  }
  .message-content {
    line-height: 1.8;
    color: #303133;
    white-space: pre-wrap;
  }
}
</style>

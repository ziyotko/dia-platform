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
            <el-button v-if="activeTab === 'inbox' && can('base:message:read-all')" @click="handleMarkAll">全部标为已读</el-button>
            <el-button v-if="can('base:message:create')" @click="handleNewDraft">新建草稿</el-button>
            <el-button v-if="can('base:message:send')" type="primary" @click="handleSend">发送消息</el-button>
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
                <el-button v-if="!row.isRead && can('base:message:read')" link type="primary" @click="handleRead(row)">标为已读</el-button>
                <el-button v-if="can('base:message:delete')" link type="danger" @click="handleDelete(row)">删除</el-button>
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
            <el-table-column label="操作" width="220" fixed="right">
              <template #default="{ row }">
                <template v-if="row.status === 2">
                  <el-button v-if="can('base:message:update')" link type="primary" @click="handleEditDraft(row)">编辑</el-button>
                  <el-button v-if="can('base:message:send-draft')" link type="success" @click="handleSendDraft(row)">发送</el-button>
                </template>
                <el-button v-if="can('base:message:delete')" link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog
      v-model="sendDialogVisible"
      :title="dialogTitle"
      width="620px"
      @closed="resetDialog"
    >
      <el-form :model="sendForm" :rules="sendRules" ref="sendFormRef" label-width="100px">
        <el-form-item v-if="dialogMode === 'send'" label="使用模板">
          <el-select
            v-model="templateCode"
            clearable
            placeholder="可不选，直接在下方填写标题与内容"
            style="width: 100%"
            @change="handleTemplateChange"
          >
            <el-option v-for="t in templates" :key="t.code" :label="`${t.name}（${t.code}）`" :value="t.code" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="templateVars.length" label="模板变量">
          <div class="template-vars">
            <div v-for="v in templateVars" :key="v.key" class="template-var">
              <span class="template-var-key">{{ v.key }}</span>
              <el-input v-model="v.value" :placeholder="v.label" @input="applyTemplate" />
            </div>
          </div>
        </el-form-item>
        <el-form-item label="接收类型" prop="receiverType">
          <el-radio-group v-model="sendForm.receiverType">
            <el-radio-button label="user">指定用户</el-radio-button>
            <el-radio-button label="all">全员广播</el-radio-button>
          </el-radio-group>
          <div v-if="dialogMode === 'draft'" class="form-tip">草稿支持单个接收人或全员广播</div>
        </el-form-item>
        <el-form-item label="接收用户" v-if="sendForm.receiverType === 'user'">
          <!-- 始终用 multiple：单选模式下 el-select 会把 v-model 写成数字，
               导致草稿的接收人丢失并退化成全员广播，同时编辑草稿也无法回显 -->
          <el-select
            v-model="sendForm.receiverIds"
            multiple
            :multiple-limit="dialogMode === 'draft' ? 1 : 0"
            collapse-tags
            collapse-tags-tooltip
            placeholder="请选择"
            style="width: 100%"
          >
            <el-option v-for="u in userOptions" :key="u.id" :label="u.username" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="dialogMode === 'send'" label="发送渠道">
          <el-select v-model="sendForm.channel" style="width: 100%">
            <el-option v-for="c in channelOptions" :key="c.value" :label="c.label" :value="c.value" />
          </el-select>
          <div class="form-tip">站外渠道需先在「系统设置 → 通知渠道」中配置；邮件/短信需指定接收人，企业微信为群推送</div>
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
        <el-button v-if="dialogMode === 'draft'" type="primary" :loading="submitting" @click="handleSubmitDialog">{{ editingDraftId ? '保存草稿' : '创建草稿' }}</el-button>
        <el-button v-else type="primary" :loading="submitting" @click="handleSubmitDialog">发送</el-button>
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
import { reactive, ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getMessageList,
  getMessageChannels,
  sendMessage,
  markMessageRead,
  markAllMessageRead,
  deleteMessage,
  createMessageDraft,
  updateMessageDraft,
  sendMessageDraft,
  getMessageTemplateList,
  renderMessageTemplate
} from '@/api/message'
import { getUserList } from '@/api/user'
import type { Message, MessageTemplate } from '@/api/message'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
// 按钮级权限：与后端 base:message:* 权限点对齐
const can = (code: string) => userStore.can(code)

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
// 弹窗模式：send 直接发送 / draft 存为草稿
const dialogMode = ref<'send' | 'draft'>('send')
const editingDraftId = ref<number | null>(null)
const submitting = ref(false)
const templates = ref<MessageTemplate[]>([])
const templateCode = ref('')
const templateVars = ref<{ key: string; label: string; value: string }[]>([])
// 发送渠道：站内信始终可用，站外渠道由后端按「是否已配置」下发
const channelOptions = ref<{ label: string; value: string }[]>([{ label: '站内信', value: 'in-app' }])
const channelLabels: Record<string, string> = {
  email: '邮件（额外发送，需在系统设置中配置 SMTP）',
  wechat: '企业微信（群机器人推送）',
  sms: '短信（按接收人手机号发送）'
}
const sendForm = reactive({
  receiverType: 'user',
  receiverIds: [] as number[],
  channel: 'in-app',
  type: 'notice',
  priority: 'normal',
  title: '',
  content: ''
})

const dialogTitle = computed(() => {
  if (dialogMode.value === 'draft') return editingDraftId.value ? '编辑草稿' : '新建草稿'
  return '发送消息'
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

// 打开弹窗（mode=send 发送 / draft 草稿）
const openDialog = async (mode: 'send' | 'draft', draft?: Message) => {
  dialogMode.value = mode
  editingDraftId.value = draft?.id ?? null
  sendForm.receiverType = draft && draft.receiverId === 0 ? 'all' : 'user'
  sendForm.receiverIds = draft && draft.receiverId ? [draft.receiverId] : []
  sendForm.channel = 'in-app'
  sendForm.type = draft?.type || 'notice'
  sendForm.priority = draft?.priority || 'normal'
  sendForm.title = draft?.title || ''
  sendForm.content = draft?.content || ''
  templateCode.value = ''
  templateVars.value = []
  sendDialogVisible.value = true

  // 接收人下拉：复用用户列表；模板仅发送模式需要
  try {
    const res: any = await getUserList({ page: 1, size: 1000 })
    userOptions.value = res.data.list || []
  } catch (error) {
    userOptions.value = []
  }
  if (mode === 'send') {
    if (templates.value.length === 0) {
      try {
        const res: any = await getMessageTemplateList({ page: 1, size: 200 })
        templates.value = res.data.list || []
      } catch (error) {
        // 无模板权限时降级为手写标题与内容
        templates.value = []
      }
    }
    try {
      const res: any = await getMessageChannels()
      const list: string[] = res.data?.channels || []
      channelOptions.value = [
        { label: '站内信', value: 'in-app' },
        ...list
          .filter((c) => c !== 'in-app')
          .map((c) => ({ label: channelLabels[c] || c, value: c }))
      ]
    } catch (error) {
      channelOptions.value = [{ label: '站内信', value: 'in-app' }]
    }
  }
}

const handleSend = () => openDialog('send')
const handleNewDraft = () => openDialog('draft')
const handleEditDraft = (row: Message) => openDialog('draft', row)

// 选择模板：填入标题/内容，并按 variables 定义生成变量输入框
const handleTemplateChange = () => {
  templateVars.value = []
  const tpl = templates.value.find((t) => t.code === templateCode.value)
  if (!tpl) return
  let parsed: Record<string, string> = {}
  if (tpl.variables) {
    try {
      parsed = JSON.parse(tpl.variables)
    } catch (error) {
      parsed = {}
    }
  }
  templateVars.value = Object.keys(parsed).map((key) => ({ key, label: parsed[key], value: '' }))
  applyTemplate()
}

// 用模板 + 变量渲染标题与内容（未填变量的占位符原样保留，便于人工补齐）
const applyTemplate = () => {
  const tpl = templates.value.find((t) => t.code === templateCode.value)
  if (!tpl) return
  const vars: Record<string, string> = {}
  templateVars.value.forEach((v) => (vars[v.key] = v.value))
  sendForm.title = renderMessageTemplate(tpl.subject || '', vars)
  sendForm.content = renderMessageTemplate(tpl.content || '', vars)
}

const resetDialog = () => {
  editingDraftId.value = null
  templateCode.value = ''
  templateVars.value = []
  submitting.value = false
}

const handleSubmitDialog = async () => {
  const valid = await sendFormRef.value?.validate().catch(() => false)
  if (!valid) return

  const isBroadcast = sendForm.receiverType === 'all'
  // 防御性归一：v-model 必须是数组（历史数据或异常状态下可能是单值）
  const picked = Array.isArray(sendForm.receiverIds) ? sendForm.receiverIds : []
  sendForm.receiverIds = picked
  const receiverIds = isBroadcast ? [] : picked
  if (!isBroadcast && receiverIds.length === 0) {
    ElMessage.warning('请选择接收用户')
    return
  }
  if (dialogMode.value === 'draft' && receiverIds.length > 1) {
    ElMessage.warning('草稿支持单个接收人，请只选择一个用户')
    return
  }

  submitting.value = true
  try {
    if (dialogMode.value === 'draft') {
      const payload = {
        receiverId: isBroadcast ? 0 : receiverIds[0],
        title: sendForm.title,
        content: sendForm.content,
        type: sendForm.type,
        priority: sendForm.priority
      }
      if (editingDraftId.value) {
        await updateMessageDraft(editingDraftId.value, payload)
        ElMessage.success('草稿已保存')
      } else {
        await createMessageDraft(payload)
        ElMessage.success('草稿已创建')
      }
      activeTab.value = 'sent'
    } else {
      await sendMessage({
        receiverType: isBroadcast ? 'all' : 'user',
        receiverIds,
        channel: sendForm.channel,
        type: sendForm.type,
        priority: sendForm.priority,
        title: sendForm.title,
        content: sendForm.content
      })
      ElMessage.success('发送成功')
    }
    sendDialogVisible.value = false
    fetchData()
  } finally {
    submitting.value = false
  }
}

// 发送已有草稿
const handleSendDraft = async (row: Message) => {
  await ElMessageBox.confirm('确认发送该草稿？', '提示', { type: 'warning' })
  await sendMessageDraft(row.id)
  ElMessage.success('发送成功')
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
  .template-vars {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .template-var {
    display: flex;
    align-items: center;
    gap: 8px;

    .template-var-key {
      flex: 0 0 auto;
      font-size: 12px;
      color: #64748b;
      background: #f1f5f9;
      border-radius: 4px;
      padding: 2px 6px;
    }
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

<template>
  <div class="page-container">
    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <span>会员专区</span>
          <div class="header-actions">
            <el-button v-if="activeTab === 'column' && isAdmin" type="primary" @click="handleAddColumn">
              <el-icon><Plus /></el-icon>新增会员栏目
            </el-button>
            <el-button v-if="activeTab === 'content'" type="primary" @click="handleAddContent">
              <el-icon><Plus /></el-icon>发布会员内容
            </el-button>
          </div>
        </div>
      </template>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="会员栏目" name="column">
          <el-form :model="columnQuery" inline>
            <el-form-item label="栏目名称">
              <el-input
                v-model="columnQuery.name"
                placeholder="请输入栏目名称"
                clearable
                @keyup.enter="handleColumnSearch"
              />
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="columnQuery.status" placeholder="全部状态" clearable style="width: 120px">
                <el-option label="启用" :value="1" />
                <el-option label="禁用" :value="0" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleColumnSearch">
                <el-icon><Search /></el-icon>查询
              </el-button>
              <el-button @click="resetColumnQuery">
                <el-icon><RefreshRight /></el-icon>重置
              </el-button>
            </el-form-item>
          </el-form>

          <el-table :data="columnData" v-loading="columnLoading" border stripe>
            <el-table-column type="index" width="60" align="center" />
            <el-table-column prop="name" label="栏目名称" min-width="150" />
            <el-table-column prop="code" label="栏目编码" min-width="140" />
            <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
            <el-table-column prop="sort" label="排序" width="90" align="center" />
            <el-table-column prop="status" label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-switch
                  v-model="row.status"
                  :active-value="1"
                  :inactive-value="0"
                  :disabled="!isAdmin"
                  @change="(val: number) => handleColumnStatusChange(row, val)"
                />
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="创建时间" width="170" />
            <el-table-column v-if="isAdmin" label="操作" width="180" align="center" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="handleEditColumn(row)">
                  <el-icon><Edit /></el-icon>编辑
                </el-button>
                <el-button link type="danger" @click="handleDeleteColumn(row)">
                  <el-icon><Delete /></el-icon>删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-pagination
              v-model:current-page="columnQuery.page"
              v-model:page-size="columnQuery.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="columnTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleColumnSizeChange"
              @current-change="fetchColumns"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane label="会员内容" name="content">
          <el-form :model="contentQuery" inline>
            <el-form-item label="内容标题">
              <el-input
                v-model="contentQuery.title"
                placeholder="请输入内容标题"
                clearable
                @keyup.enter="handleContentSearch"
              />
            </el-form-item>
            <el-form-item label="会员栏目">
              <el-select
                v-model="contentQuery.columnId"
                placeholder="全部栏目"
                clearable
                style="width: 160px"
              >
                <el-option
                  v-for="item in columnOptions"
                  :key="item.id"
                  :label="item.status === 1 ? item.name : `${item.name}（已禁用）`"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="内容类型">
              <el-select v-model="contentQuery.type" placeholder="全部类型" clearable style="width: 120px">
                <el-option label="新闻" :value="1" />
                <el-option label="数据" :value="2" />
                <el-option label="视频" :value="3" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="contentQuery.status" placeholder="全部状态" clearable style="width: 120px">
                <el-option label="草稿" :value="0" />
                <el-option label="已发布" :value="1" />
                <el-option label="已下线" :value="2" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleContentSearch">
                <el-icon><Search /></el-icon>查询
              </el-button>
              <el-button @click="resetContentQuery">
                <el-icon><RefreshRight /></el-icon>重置
              </el-button>
            </el-form-item>
          </el-form>

          <el-table :data="contentData" v-loading="contentLoading" border stripe>
            <el-table-column type="index" width="60" align="center" />
            <el-table-column prop="title" label="内容标题" min-width="200" show-overflow-tooltip />
            <el-table-column prop="memberColumnName" label="会员栏目" min-width="130" />
            <el-table-column prop="type" label="类型" width="90" align="center">
              <template #default="{ row }">
                <el-tag size="small" :type="typeTagType(row.type)">{{ typeName(row.type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="author" label="作者" width="110" />
            <el-table-column prop="publishTime" label="发布时间" width="170" />
            <el-table-column prop="status" label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag size="small" :type="statusTagType(row.status)">{{ statusName(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="isTop" label="置顶" width="80" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.isTop === 1" size="small" type="danger">置顶</el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="220" align="center" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="handleEditContent(row)">
                  <el-icon><Edit /></el-icon>编辑
                </el-button>
                <el-button link type="primary" @click="handleToggleContentStatus(row)">
                  {{ row.status === 1 ? '下线' : '发布' }}
                </el-button>
                <el-button link type="danger" @click="handleDeleteContent(row)">
                  <el-icon><Delete /></el-icon>删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-pagination
              v-model:current-page="contentQuery.page"
              v-model:page-size="contentQuery.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="contentTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleContentSizeChange"
              @current-change="fetchContents"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog
      v-model="columnDialogVisible"
      :title="columnDialogTitle"
      width="600px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form ref="columnFormRef" :model="columnForm" :rules="columnRules" label-width="90px">
        <el-form-item label="栏目名称" prop="name">
          <el-input v-model="columnForm.name" placeholder="请输入会员栏目名称" />
        </el-form-item>
        <el-form-item label="栏目编码" prop="code">
          <el-input v-model="columnForm.code" placeholder="请输入会员栏目编码" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="columnForm.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="columnForm.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="columnForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="columnDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="columnSubmitLoading" @click="handleSubmitColumn">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="contentDialogVisible"
      :title="contentDialogTitle"
      width="900px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form ref="contentFormRef" :model="contentForm" :rules="contentRules" label-width="90px">
        <el-form-item label="内容标题" prop="title">
          <el-input v-model="contentForm.title" placeholder="请输入内容标题" maxlength="200" show-word-limit />
        </el-form-item>
        <el-form-item label="会员栏目" prop="memberColumnId">
          <el-select v-model="contentForm.memberColumnId" placeholder="请选择会员栏目" style="width: 100%">
            <el-option
              v-for="item in columnOptions"
              :key="item.id"
              :label="item.status === 1 ? item.name : `${item.name}（已禁用）`"
              :value="item.id"
              :disabled="item.status === 0 && item.id !== contentForm.memberColumnId"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="内容类型" prop="type">
          <el-radio-group v-model="contentForm.type">
            <el-radio :value="1">新闻</el-radio>
            <el-radio :value="2">数据</el-radio>
            <el-radio :value="3">视频</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="内容摘要" prop="summary">
          <el-input
            v-model="contentForm.summary"
            type="textarea"
            :rows="3"
            placeholder="请输入内容摘要"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="封面">
          <el-upload
            :show-file-list="false"
            :http-request="handleCoverUpload"
            accept="image/*"
          >
            <el-button><el-icon><Upload /></el-icon>上传封面</el-button>
          </el-upload>
          <div v-if="contentForm.cover" class="cover-preview">
            <el-image :src="contentForm.cover" fit="cover" />
            <el-button link type="danger" @click="contentForm.cover = ''">移除</el-button>
          </div>
        </el-form-item>
        <el-form-item v-if="contentForm.type === 3" label="视频地址">
          <el-input v-model="contentForm.videoUrl" placeholder="请输入视频地址或点击右侧上传" />
          <el-upload
            class="inline-upload"
            :show-file-list="false"
            :http-request="handleVideoUpload"
            accept="video/mp4"
          >
            <el-button><el-icon><VideoCamera /></el-icon>上传视频</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item v-if="contentForm.type === 2" label="数据附件">
          <el-upload :show-file-list="false" :http-request="handleAttachmentUpload">
            <el-button><el-icon><Paperclip /></el-icon>上传附件</el-button>
          </el-upload>
          <span v-if="contentForm.attachmentName" class="attachment-name">
            {{ contentForm.attachmentName }}
            <el-button link type="danger" @click="clearAttachment">移除</el-button>
          </span>
        </el-form-item>
        <el-form-item v-if="contentForm.type !== 3" label="内容正文" prop="content" class="editor-form-item">
          <div class="editor-wrapper">
            <Toolbar
              style="border-bottom: 1px solid #e4e7ed"
              :editor="editorRef"
              :defaultConfig="toolbarConfig"
              mode="default"
            />
            <Editor
              v-model="contentForm.content"
              :defaultConfig="editorConfig"
              mode="default"
              @onCreated="handleEditorCreated"
            />
          </div>
        </el-form-item>
        <el-form-item label="来源" prop="source">
          <el-input v-model="contentForm.source" placeholder="请输入来源" maxlength="200" />
        </el-form-item>
        <el-form-item label="发布时间" prop="publishTime">
          <el-date-picker
            v-model="contentForm.publishTime"
            type="datetime"
            placeholder="请选择发布时间"
            value-format="YYYY-MM-DD HH:mm:ss"
          />
        </el-form-item>
        <el-form-item label="置顶">
          <el-switch v-model="contentForm.isTop" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="contentForm.status">
            <el-radio :value="0">草稿</el-radio>
            <el-radio :value="1">已发布</el-radio>
            <el-radio :value="2">已下线</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="contentDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="contentSubmitLoading" @click="handleSubmitContent">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, shallowRef } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Delete,
  Edit,
  Paperclip,
  Plus,
  RefreshRight,
  Search,
  Upload,
  VideoCamera
} from '@element-plus/icons-vue'
import '@wangeditor/editor/dist/css/style.css'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import type { IDomEditor, IEditorConfig, IToolbarConfig } from '@wangeditor/editor'

import { useUserStore } from '@/stores/user'
import { hasAdminRole } from '@/utils/permission'
import { uploadFile } from '@/api/upload'
import {
  createMemberColumn,
  createMemberContent,
  deleteMemberColumn,
  deleteMemberContent,
  getMemberColumns,
  getMemberContent,
  getMemberContents,
  updateMemberColumn,
  updateMemberColumnStatus,
  updateMemberContent,
  updateMemberContentStatus
} from '@/api/member-zone'

const userStore = useUserStore()
// 会员栏目的新增/编辑/删除/改状态均为管理员接口（写路由在 admin 组）
const isAdmin = computed(() => hasAdminRole(userStore.userInfo?.roleIds))

const activeTab = ref<'column' | 'content'>('column')

// ============================ 会员栏目 ============================
const columnLoading = ref(false)
const columnTotal = ref(0)
const columnData = ref<any[]>([])
const columnDialogVisible = ref(false)
const columnDialogTitle = ref('')
const columnSubmitLoading = ref(false)
const columnFormRef = ref()

const columnQuery = reactive({
  page: 1,
  pageSize: 10,
  name: '',
  status: undefined as number | undefined
})

const columnForm = reactive({
  id: undefined as number | undefined,
  name: '',
  code: '',
  description: '',
  sort: 0,
  status: 1
})

const columnRules = {
  name: [{ required: true, message: '请输入栏目名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入栏目编码', trigger: 'blur' }]
}

// 供「会员栏目」下拉使用（含禁用项，便于编辑历史内容时保留原栏目）
const columnOptions = ref<any[]>([])

const resetColumnForm = () => {
  Object.assign(columnForm, {
    id: undefined,
    name: '',
    code: '',
    description: '',
    sort: 0,
    status: 1
  })
}

const fetchColumns = async () => {
  columnLoading.value = true
  try {
    const params: any = { page: columnQuery.page, pageSize: columnQuery.pageSize }
    if (columnQuery.name) params.name = columnQuery.name
    if (columnQuery.status !== undefined) params.status = columnQuery.status
    const res: any = await getMemberColumns(params)
    columnData.value = res.data?.list || []
    columnTotal.value = res.data?.total || 0
  } catch {
    // 错误提示由 request 拦截器统一给出
  } finally {
    columnLoading.value = false
  }
}

const fetchColumnOptions = async () => {
  try {
    // 下拉需同时包含「已禁用」的栏目：编辑历史内容时其原栏目可能已被禁用，
    // 后端只允许「新建 / 更换栏目」时要求栏目启用，故这里取较大页容量一次性拉取。
    const res: any = await getMemberColumns({ page: 1, pageSize: 200 })
    columnOptions.value = res.data?.list || []
  } catch {
    columnOptions.value = []
  }
}

const handleColumnSearch = () => {
  columnQuery.page = 1
  fetchColumns()
}

const resetColumnQuery = () => {
  columnQuery.name = ''
  columnQuery.status = undefined
  columnQuery.page = 1
  fetchColumns()
}

const handleColumnSizeChange = () => {
  columnQuery.page = 1
  fetchColumns()
}

const handleAddColumn = () => {
  columnDialogTitle.value = '新增会员栏目'
  resetColumnForm()
  columnDialogVisible.value = true
}

const handleEditColumn = (row: any) => {
  columnDialogTitle.value = '编辑会员栏目'
  Object.assign(columnForm, {
    id: row.id,
    name: row.name,
    code: row.code,
    description: row.description,
    sort: row.sort,
    status: row.status
  })
  columnDialogVisible.value = true
}

const handleSubmitColumn = async () => {
  const valid = await columnFormRef.value?.validate().catch(() => false)
  if (!valid) return
  columnSubmitLoading.value = true
  try {
    if (columnForm.id) {
      await updateMemberColumn(columnForm.id, { ...columnForm })
      ElMessage.success('修改成功')
    } else {
      await createMemberColumn({ ...columnForm })
      ElMessage.success('新增成功')
    }
    columnDialogVisible.value = false
    fetchColumns()
    fetchColumnOptions()
  } finally {
    columnSubmitLoading.value = false
  }
}

const handleColumnStatusChange = async (row: any, val: number) => {
  try {
    await updateMemberColumnStatus(row.id, val)
    ElMessage.success(`会员栏目状态已${val === 1 ? '启用' : '禁用'}`)
    fetchColumnOptions()
  } catch {
    row.status = val === 1 ? 0 : 1
  }
}

const handleDeleteColumn = (row: any) => {
  ElMessageBox.confirm(`确定要删除会员栏目 「${row.name}」 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      await deleteMemberColumn(row.id)
      ElMessage.success('删除成功')
      fetchColumns()
      fetchColumnOptions()
    })
    .catch(() => {})
}

// ============================ 会员内容 ============================
const contentLoading = ref(false)
const contentTotal = ref(0)
const contentData = ref<any[]>([])
const contentDialogVisible = ref(false)
const contentDialogTitle = ref('')
const contentSubmitLoading = ref(false)
const contentFormRef = ref()

const contentQuery = reactive({
  page: 1,
  pageSize: 10,
  title: '',
  columnId: undefined as number | undefined,
  type: undefined as number | undefined,
  status: undefined as number | undefined
})

const contentForm = reactive({
  id: undefined as number | undefined,
  memberColumnId: undefined as number | undefined,
  title: '',
  type: 1,
  summary: '',
  content: '',
  cover: '',
  videoUrl: '',
  attachmentName: '',
  attachmentUrl: '',
  source: '',
  publishTime: '',
  status: 0,
  isTop: 0
})

const contentRules = {
  title: [{ required: true, message: '请输入内容标题', trigger: 'blur' }],
  memberColumnId: [{ required: true, message: '请选择会员栏目', trigger: 'change' }],
  type: [{ required: true, message: '请选择内容类型', trigger: 'change' }]
}

const typeName = (type: number) => {
  const map: Record<number, string> = { 1: '新闻', 2: '数据', 3: '视频' }
  return map[type] || '新闻'
}

const typeTagType = (type: number): 'primary' | 'success' | 'warning' => {
  const map: Record<number, 'primary' | 'success' | 'warning'> = {
    1: 'primary',
    2: 'success',
    3: 'warning'
  }
  return map[type] || 'primary'
}

const statusName = (status: number) => {
  const map: Record<number, string> = { 0: '草稿', 1: '已发布', 2: '已下线' }
  return map[status] || '草稿'
}

const statusTagType = (status: number): 'info' | 'success' | 'warning' => {
  const map: Record<number, 'info' | 'success' | 'warning'> = {
    0: 'info',
    1: 'success',
    2: 'warning'
  }
  return map[status] || 'info'
}

const resetContentForm = () => {
  Object.assign(contentForm, {
    id: undefined,
    memberColumnId: undefined,
    title: '',
    type: 1,
    summary: '',
    content: '',
    cover: '',
    videoUrl: '',
    attachmentName: '',
    attachmentUrl: '',
    source: '',
    publishTime: '',
    status: 0,
    isTop: 0
  })
}

const fetchContents = async () => {
  contentLoading.value = true
  try {
    const params: any = { page: contentQuery.page, pageSize: contentQuery.pageSize }
    if (contentQuery.title) params.title = contentQuery.title
    if (contentQuery.columnId !== undefined) params.columnId = contentQuery.columnId
    if (contentQuery.type !== undefined) params.type = contentQuery.type
    if (contentQuery.status !== undefined) params.status = contentQuery.status
    const res: any = await getMemberContents(params)
    contentData.value = res.data?.list || []
    contentTotal.value = res.data?.total || 0
  } catch {
    // 错误提示由 request 拦截器统一给出
  } finally {
    contentLoading.value = false
  }
}

const handleContentSearch = () => {
  contentQuery.page = 1
  fetchContents()
}

const resetContentQuery = () => {
  contentQuery.title = ''
  contentQuery.columnId = undefined
  contentQuery.type = undefined
  contentQuery.status = undefined
  contentQuery.page = 1
  fetchContents()
}

const handleContentSizeChange = () => {
  contentQuery.page = 1
  fetchContents()
}

const handleAddContent = () => {
  contentDialogTitle.value = '发布会员内容'
  resetContentForm()
  contentDialogVisible.value = true
}

const handleEditContent = async (row: any) => {
  contentDialogTitle.value = '编辑会员内容'
  resetContentForm()
  try {
    const res: any = await getMemberContent(row.id)
    const detail = res.data || {}
    Object.assign(contentForm, {
      id: detail.id,
      memberColumnId: detail.memberColumnId,
      title: detail.title,
      type: detail.type,
      summary: detail.summary,
      content: detail.content || '',
      cover: detail.cover,
      videoUrl: detail.videoUrl,
      attachmentName: detail.attachmentName,
      attachmentUrl: detail.attachmentUrl,
      source: detail.source,
      publishTime: detail.publishTime || '',
      status: detail.status,
      isTop: detail.isTop
    })
    contentDialogVisible.value = true
  } catch {
    // 错误提示由 request 拦截器统一给出
  }
}

const handleSubmitContent = async () => {
  const valid = await contentFormRef.value?.validate().catch(() => false)
  if (!valid) return
  contentSubmitLoading.value = true
  try {
    if (contentForm.id) {
      await updateMemberContent(contentForm.id, { ...contentForm })
      ElMessage.success('修改成功')
    } else {
      await createMemberContent({ ...contentForm })
      ElMessage.success('发布成功')
    }
    contentDialogVisible.value = false
    fetchContents()
  } finally {
    contentSubmitLoading.value = false
  }
}

const handleToggleContentStatus = (row: any) => {
  const nextStatus = row.status === 1 ? 2 : 1
  const actionText = nextStatus === 1 ? '发布' : '下线'
  ElMessageBox.confirm(`确定要${actionText}内容 「${row.title}」 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      await updateMemberContentStatus(row.id, nextStatus)
      ElMessage.success(`${actionText}成功`)
      fetchContents()
    })
    .catch(() => {})
}

const handleDeleteContent = (row: any) => {
  ElMessageBox.confirm(`确定要删除内容 「${row.title}」 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      await deleteMemberContent(row.id)
      ElMessage.success('删除成功')
      fetchContents()
    })
    .catch(() => {})
}

// ============================ 上传 ============================
const handleCoverUpload = async (options: any) => {
  const res: any = await uploadFile(options.file, 'article')
  contentForm.cover = res.data?.url || ''
  ElMessage.success('封面上传成功')
}

const handleVideoUpload = async (options: any) => {
  const res: any = await uploadFile(options.file, 'video')
  contentForm.videoUrl = res.data?.url || ''
  ElMessage.success('视频上传成功')
}

const handleAttachmentUpload = async (options: any) => {
  const res: any = await uploadFile(options.file, 'attachment')
  contentForm.attachmentUrl = res.data?.url || ''
  contentForm.attachmentName = res.data?.name || options.file?.name || ''
  ElMessage.success('附件上传成功')
}

const clearAttachment = () => {
  contentForm.attachmentUrl = ''
  contentForm.attachmentName = ''
}

// ============================ 富文本编辑器 ============================
const editorRef = shallowRef<IDomEditor>()
const toolbarConfig: Partial<IToolbarConfig> = {}
const IMAGE_MAX_SIZE = 10 * 1024 * 1024

const uploadImageFile = async (
  file: File,
  insertFn: (url: string, alt: string, href: string) => void
) => {
  if (file.size > IMAGE_MAX_SIZE) {
    ElMessage.error('图片大小不能超过 10MB')
    throw new Error('图片大小超出限制')
  }
  const res: any = await uploadFile(file, 'article')
  const url = res.data?.url || ''
  if (url) {
    insertFn(url, '', '')
  } else {
    ElMessage.error('图片上传失败')
  }
}

const editorConfig: Partial<IEditorConfig> = {
  placeholder: '请输入内容正文',
  MENU_CONF: {
    uploadImage: {
      maxFileSize: IMAGE_MAX_SIZE,
      maxNumberOfFiles: 1,
      customBrowseAndUpload(insertFn: (url: string, alt: string, href: string) => void) {
        const input = document.createElement('input')
        input.type = 'file'
        input.accept = 'image/*'
        input.style.display = 'none'
        input.onchange = () => {
          const file = input.files?.[0]
          if (!file) return
          uploadImageFile(file, insertFn).finally(() => {
            input.value = ''
            input.remove()
          })
        }
        document.body.appendChild(input)
        input.click()
      },
      async customUpload(file: File, insertFn: (url: string, alt: string, href: string) => void) {
        await uploadImageFile(file, insertFn)
      }
    }
  }
}

const handleEditorCreated = (editor: IDomEditor) => {
  editorRef.value = editor
}

onMounted(() => {
  fetchColumns()
  fetchColumnOptions()
  fetchContents()
})

onBeforeUnmount(() => {
  editorRef.value?.destroy()
})
</script>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.cover-preview {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 10px;
}

.cover-preview .el-image {
  width: 120px;
  height: 80px;
  border-radius: 8px;
  border: 1px solid var(--app-brand-soft);
}

.inline-upload {
  display: inline-block;
  margin-left: 10px;
}

.attachment-name {
  margin-left: 10px;
  color: var(--app-text-secondary);
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.editor-form-item :deep(.el-form-item__content) {
  display: block;
}

.editor-wrapper {
  width: 100%;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  overflow: hidden;
}

.editor-wrapper :deep(.w-e-text-container) {
  height: 320px !important;
}
</style>

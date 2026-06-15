<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="文章标题">
          <el-input v-model="queryForm.title" placeholder="请输入文章标题" clearable />
        </el-form-item>
        <el-form-item label="文章分类">
          <el-select v-model="queryForm.categoryId" placeholder="全部分类" clearable style="width: 140px">
            <el-option
              v-for="item in categoryList"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="发布状态">
          <el-select v-model="queryForm.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="已发布" :value="1" />
            <el-option label="草稿" :value="0" />
            <el-option label="已下线" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="queryForm.auditStatus" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="待审核" :value="0" />
            <el-option label="审核中" :value="1" />
            <el-option label="已审核" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>查询
          </el-button>
          <el-button @click="resetQuery">
            <el-icon><RefreshRight /></el-icon>重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <span>文章列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增文章
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="title" label="文章标题" min-width="180" show-overflow-tooltip />
        <el-table-column label="封面图" width="80" align="center">
          <template #default="{ row }">
            <el-image
              v-if="row.cover"
              :src="row.cover"
              :preview-src-list="[row.cover]"
              fit="cover"
              style="width: 60px; height: 60px; border-radius: 8px"
            />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="categoryName" label="所属分类" width="120" />
        <el-table-column label="标签" width="120">
          <template #default="{ row }">
            <el-tag
              v-for="tag in getRowTags(row)"
              :key="tag.id"
              size="small"
              :color="tag.color"
              style="margin-right: 4px; color: #fff; border: none;"
            >
              {{ tag.name }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="author" label="作者" width="100" />
        <el-table-column prop="source" label="来源" width="140" show-overflow-tooltip />
        <el-table-column prop="columnCount" label="发布栏目数量" width="120" align="center" />
        <el-table-column prop="status" label="发布状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : row.status === 0 ? 'info' : 'danger'">
              {{ row.status === 1 ? '已发布' : row.status === 0 ? '草稿' : '已下线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="auditStatus" label="审核状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.auditStatus === 2 ? 'success' : row.auditStatus === 1 ? 'warning' : 'info'"
              :class="{ 'audit-status-clickable': row.columnCount > 0 && (row.auditStatus === 0 || row.auditStatus === 1 || row.auditStatus === 2) }"
              @click="row.columnCount > 0 && (row.auditStatus === 0 || row.auditStatus === 1 || row.auditStatus === 2) && handleShowAuditFlow(row)"
            >
              {{ row.auditStatus === 2 ? '已审核' : row.auditStatus === 1 ? '审核中' : '待审核' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="isTop" label="置顶" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isTop" type="warning">置顶</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="isBold" label="加粗" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isBold" type="danger">加粗</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="defaultColor" label="颜色" width="80" align="center">
          <template #default="{ row }">
            <span v-if="row.defaultColor" :style="{ display: 'inline-block', width: '20px', height: '20px', backgroundColor: row.defaultColor, borderRadius: '4px', border: '1px solid #dcdfe6' }" />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="420" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handlePreview(row)">
              <el-icon><View /></el-icon>预览
            </el-button>
            <el-button v-if="isAuthor(row) && row.auditStatus !== 1 && row.auditStatus !== 2" link type="primary" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button v-if="isAuthor(row) && row.auditStatus !== 2" link type="success" @click="row.auditStatus === 1 ? ElMessage.warning('审核中的文章不能修改栏目') : handleSetColumns(row)">
              <el-icon><FolderOpened /></el-icon>栏目
            </el-button>
            <el-button v-if="isAuthor(row) && row.status === 0 && row.columnCount > 0 && row.auditStatus === 0" link type="warning" @click="handleAudit(row)">
              <el-icon><CircleCheck /></el-icon>提交审核
            </el-button>
            <el-button v-if="isAuthor(row) && row.status === 0 && row.columnCount > 0 && row.auditStatus === 2" link type="warning" @click="handleReAudit(row)">
              <el-icon><CircleCheck /></el-icon>重新提交审核
            </el-button>
            <el-button v-if="isAuthor(row) && row.auditStatus === 1" link type="warning" @click="handleWithdrawAudit(row)">
              <el-icon><CircleClose /></el-icon>撤回审核
            </el-button>
            <el-button v-if="row.status === 1" link type="danger" @click="handleOffShelf(row)">
              <el-icon><CircleClose /></el-icon>下线
            </el-button>
            <el-button v-if="isAuthor(row)" link type="danger" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="queryForm.page"
          v-model:page-size="queryForm.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      fullscreen
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="80px"
        class="article-form"
      >
        <div class="form-section">
          <div class="section-title">基本信息</div>
          <el-form-item label="文章标题" prop="title">
            <el-input v-model="form.title" placeholder="请输入文章标题" clearable />
          </el-form-item>
          <el-row :gutter="20">
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="所属分类" prop="categoryId">
                <el-select v-model="form.categoryId" placeholder="请选择分类" style="width: 100%">
                  <el-option
                    v-for="item in categoryList"
                    :key="item.id"
                    :label="item.name"
                    :value="item.id"
                  />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="文章标签" prop="tagIds">
                <el-select v-model="form.tagIds" multiple placeholder="请选择标签" style="width: 100%">
                  <el-option
                    v-for="item in tagList"
                    :key="item.id"
                    :label="item.name"
                    :value="item.id"
                  />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="发布状态">
                <el-tag :type="form.status === 1 ? 'success' : form.status === 0 ? 'info' : 'danger'">
                  {{ form.status === 1 ? '已发布' : form.status === 0 ? '草稿' : '已下线' }}
                </el-tag>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :xs="24" :sm="12">
              <el-form-item label="封面图" prop="cover">
                <div class="article-cover-uploader">
                  <el-upload
                    v-if="!form.cover"
                    class="cover-uploader"
                    action=""
                    :http-request="handleCoverUpload"
                    :show-file-list="false"
                    accept="image/*"
                  >
                    <el-icon class="uploader-icon"><Plus /></el-icon>
                    <div class="uploader-text">点击上传封面</div>
                    <div class="uploader-hint">建议尺寸 800×480</div>
                  </el-upload>
                  <div v-else class="cover-preview">
                    <div class="cover-image-wrapper">
                      <el-image
                        :src="form.cover"
                        fit="cover"
                        style="width: 100%; height: 100%"
                        :preview-src-list="[form.cover]"
                      />
                      <div class="cover-overlay" @click="handleRemoveCover">
                        <el-icon><Delete /></el-icon>
                        <span>删除封面</span>
                      </div>
                    </div>
                  </div>
                </div>
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12">
              <el-form-item label="文章来源" prop="source">
                <el-input v-model="form.source" placeholder="请输入文章来源" clearable />
              </el-form-item>
            </el-col>
          </el-row>
        </div>

        <el-divider />

        <div class="form-section">
          <div class="section-title">样式设置</div>
          <el-row :gutter="20">
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="是否置顶" prop="isTop">
                <el-radio-group v-model="form.isTop">
                  <el-radio :value="1">置顶</el-radio>
                  <el-radio :value="0">不置顶</el-radio>
                </el-radio-group>
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="是否加粗" prop="isBold">
                <el-radio-group v-model="form.isBold">
                  <el-radio :value="1">加粗</el-radio>
                  <el-radio :value="0">不加粗</el-radio>
                </el-radio-group>
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="标题颜色" prop="defaultColor">
                <el-color-picker v-model="form.defaultColor" show-alpha />
              </el-form-item>
            </el-col>
          </el-row>
        </div>

        <el-divider />

        <div class="form-section">
          <div class="section-title">内容编辑</div>
          <el-form-item label="文章摘要" prop="summary">
            <el-input
              v-model="form.summary"
              type="textarea"
              :rows="3"
              placeholder="请输入文章摘要，简要描述文章核心内容"
              maxlength="500"
              show-word-limit
            />
          </el-form-item>
          <el-form-item label="文章内容" prop="content" class="editor-form-item">
            <div class="editor-wrapper">
              <Toolbar
                style="border-bottom: 1px solid #e4e7ed"
                :editor="editorRef"
                :defaultConfig="toolbarConfig"
                mode="default"
              />
              <Editor
                v-model="form.content"
                :defaultConfig="editorConfig"
                mode="default"
                @onCreated="handleCreated"
                @onChange="handleEditorChange"
                @customPaste="handleCustomPaste"
              />
            </div>
          </el-form-item>
        </div>

        <el-divider />

        <div class="form-section">
          <div class="section-title">附件管理</div>
          <el-form-item label="文章附件" prop="attachments">
            <el-upload
              v-model:file-list="form.attachments"
              action="#"
              :http-request="handleAttachmentUpload"
              :before-upload="handleAttachmentBeforeUpload"
              :on-remove="handleAttachmentRemove"
              multiple
              :limit="10"
              class="attachment-uploader"
            >
              <el-button type="primary" plain>
                <el-icon><Plus /></el-icon>上传附件
              </el-button>
              <template #tip>
                <div class="attachment-tip">
                  支持 PDF、Word、Excel、PPT、TXT、ZIP、RAR、7Z、MP4、MP3 等常见格式，单个文件不超过 50MB，最多上传 10 个附件
                </div>
              </template>
            </el-upload>
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="previewVisible" title="文章预览" width="800px" destroy-on-close>
      <div class="preview-content">
        <h2>{{ previewData.title }}</h2>
        <div class="preview-meta">
          <span>作者：{{ previewData.author }}</span>
          <span>来源：{{ previewData.source }}</span>
          <span>分类：{{ previewData.categoryName }}</span>
          <span>时间：{{ previewData.createTime }}</span>
        </div>
        <div class="preview-summary" v-if="previewData.summary">
          <strong>摘要：</strong>{{ previewData.summary }}
        </div>
        <div class="preview-body" v-html="previewData.content" />
        <div class="preview-attachments" v-if="previewData.attachments && previewData.attachments.length > 0">
          <strong>附件：</strong>
          <div class="attachment-list">
            <a
              v-for="att in previewData.attachments"
              :key="att.id || att.url"
              :href="att.url"
              target="_blank"
              class="attachment-item"
            >
              <el-icon><Document /></el-icon>
              <span class="attachment-name">{{ att.name }}</span>
            </a>
          </div>
        </div>
      </div>
    </el-dialog>

    <el-dialog v-model="columnDialogVisible" :title="`栏目设置 - ${columnDialogTitle}`" width="500px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="选择栏目">
          <el-select v-model="selectedColumnIds" multiple placeholder="请选择栏目" style="width: 100%">
            <el-option-group v-for="page in pageList" :key="page.id" :label="page.name">
              <el-option
                v-for="col in getColumnsByPage(page.id)"
                :key="col.id"
                :label="col.name"
                :value="col.id"
              />
            </el-option-group>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="columnDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="columnSubmitLoading" @click="handleSubmitColumns">确定</el-button>
      </template>
    </el-dialog>

    <!-- 审核流程预览 -->
    <el-dialog v-model="auditFlowDialogVisible" title="栏目审核流程" width="640px" destroy-on-close>
      <div v-if="auditFlowArticleTitle" class="audit-flow-subtitle">
        文章：{{ auditFlowArticleTitle }}
      </div>
      <el-skeleton v-if="auditFlowLoading" :rows="6" animated />
      <el-empty v-else-if="auditFlowList.length === 0" description="暂无栏目或未绑定审核流程" />
      <div v-else class="audit-flow-list">
        <div
          v-for="(item, index) in auditFlowList"
          :key="index"
          class="audit-flow-card"
          :class="{ 'audit-flow-rejected': item.auditStatus === 2 }"
        >
          <div class="audit-flow-card-header">
            <div class="audit-flow-index">{{ index + 1 }}</div>
            <div class="audit-flow-column-name">{{ item.columnName }}</div>
            <el-tag v-if="item.auditStatus === 2" size="small" type="danger" effect="light">已驳回</el-tag>
            <el-tag v-else-if="item.auditStatus === 1" size="small" type="success" effect="light">已通过</el-tag>
            <el-tag v-else-if="item.workflow && item.auditStatus === 0" size="small" type="warning" effect="light">审核中</el-tag>
            <el-tag v-else-if="item.workflow" size="small" type="primary" effect="light">
              {{ item.workflow.name }}
            </el-tag>
            <el-tag v-else size="small" type="info" effect="light">未绑定流程</el-tag>
          </div>
          <div v-if="item.workflow" class="audit-flow-card-body">
            <div v-if="item.workflow.nodes && item.workflow.nodes.length > 0" class="audit-flow-steps">
              <el-steps
                :active="getStepActive(item)"
                align-center
              >
                <el-step
                  v-for="node in item.workflow.nodes"
                  :key="node.id"
                  :title="node.name"
                  :status="getNodeStatus(item, node.id)"
                >
                  <template #description>
                    <div v-if="getNodeHistory(item, node.id)" class="audit-step-desc" :class="getNodeHistory(item, node.id).action === 2 ? 'audit-step-reject' : ''">
                      <el-icon :color="getNodeHistory(item, node.id).action === 2 ? '#f56c6c' : '#67c23a'" size="12">
                        <CircleCheck v-if="getNodeHistory(item, node.id).action === 1" />
                        <CircleClose v-else />
                      </el-icon>
                      <span>{{ getNodeHistory(item, node.id).operatorName }} {{ formatAuditTime(getNodeHistory(item, node.id).createTime) }}</span>
                    </div>
                  </template>
                </el-step>
              </el-steps>
            </div>
            <el-empty v-else description="该流程未配置节点" :image-size="60" />
            <div v-if="item.auditStatus === 1 && item.approveUserName" class="audit-flow-result">
              <el-icon color="#67c23a"><CircleCheck /></el-icon>
              <span>已通过：{{ item.approverRemark }}</span>
            </div>
            <div v-if="item.auditStatus === 2 && item.rejectRemark" class="audit-flow-result audit-flow-reject-result">
              <el-icon color="#f56c6c"><CircleClose /></el-icon>
              <span>已驳回：{{ item.rejectRemark }}</span>
            </div>
            <div v-if="item.auditStatus === 0 && item.workflow && item.workflow.nodes && item.workflow.nodes.length > 0">
              <div v-if="item.currentApproverId == 0 || currentUserId == item.currentApproverId" class="audit-flow-actions">
                <el-button type="primary" size="small" @click="handleAdvanceAuditNode(item.columnId)">
                  <el-icon><CircleCheck /></el-icon>通过当前节点
                </el-button>
                <el-button type="danger" size="small" plain @click="handleRejectAuditNode(item.columnId)">
                  <el-icon><CircleClose /></el-icon>驳回
                </el-button>
              </div>
              <div v-else class="audit-flow-no-auth">
                <el-icon><Warning /></el-icon>
                <span>您不是当前节点审批人，无权操作</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="auditFlowDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, shallowRef, onBeforeUnmount, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  RefreshRight,
  Plus,
  Edit,
  Delete,
  View,
  CircleCheck,
  CircleClose,
  FolderOpened,
  Warning,
  Document
} from '@element-plus/icons-vue'
import '@wangeditor/editor/dist/css/style.css'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import type { IDomEditor, IEditorConfig, IToolbarConfig } from '@wangeditor/editor'

import { useUserStore } from '@/stores/user'
import request from '@/utils/request'
import {
  getArticles,
  createArticle,
  updateArticle,
  deleteArticle,
  auditArticle,
  restartArticleAudit,
  withdrawArticleAudit,
  updateArticleStatus,
  setArticleColumns,
  getArticleAuditProgress,
  advanceArticleAudit,
  rejectArticleAudit,
  getArticleAuditHistory
} from '@/api/article'
import { getWorkflowByID } from '@/api/workflow'
import { getAllCategories } from '@/api/category'
import { getAllTags } from '@/api/tag'
import { getPages } from '@/api/page'
import { getColumns } from '@/api/column'
import { uploadFile } from '@/api/upload'

const userStore = useUserStore()
const currentUserId = computed(() => userStore.userInfo?.id || 0)

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const total = ref(0)
const formRef = ref()

const columnDialogVisible = ref(false)
const columnDialogTitle = ref('')
const selectedColumnIds = ref<number[]>([])
const columnSubmitLoading = ref(false)
const currentArticleId = ref<number | undefined>(undefined)
const pageList = ref<any[]>([])
const columnList = ref<any[]>([])

const auditFlowDialogVisible = ref(false)
const auditFlowLoading = ref(false)
const auditFlowList = ref<any[]>([])
const auditFlowArticleTitle = ref('')

const queryForm = reactive({
  page: 1,
  pageSize: 10,
  title: '',
  categoryId: undefined as number | undefined,
  status: undefined as number | undefined,
  auditStatus: undefined as number | undefined
})

const form = reactive({
  id: undefined as number | undefined,
  title: '',
  categoryId: undefined as number | undefined,
  tagIds: [] as number[],
  summary: '',
  content: '',
  status: 0,
  auditStatus: 0,
  isTop: 0,
  isBold: 0,
  defaultColor: '',
  cover: '',
  source: '',
  attachments: [] as any[]
})

const formRules = {
  title: [{ required: true, message: '请输入文章标题', trigger: 'blur' }],
  categoryId: [{ required: true, message: '请选择所属分类', trigger: 'change' }]
}

const route = useRoute()
const router = useRouter()

const tableData = ref<any[]>([])
const categoryList = ref<any[]>([])
const tagList = ref<any[]>([])

const tagMap = computed(() => {
  const map: Record<number, any> = {}
  tagList.value.forEach(tag => {
    map[tag.id] = tag
  })
  return map
})

const getRowTags = (row: any) => {
  const ids = row.tagIds || []
  return ids.map((id: number) => tagMap.value[id]).filter(Boolean)
}

const getColumnsByPage = (pageId: number) => {
  return columnList.value.filter((col: any) => col.pageId === pageId)
}

const fetchPages = async () => {
  try {
    const res: any = await getPages()
    pageList.value = res.data || []
  } catch (error) {
    // ignore
  }
}

const fetchColumns = async () => {
  try {
    const res: any = await getColumns()
    columnList.value = res.data || []
  } catch (error) {
    // ignore
  }
}

// 编辑器
const editorRef = shallowRef<IDomEditor>()
const toolbarConfig: Partial<IToolbarConfig> = {}
const base64ToBlob = (base64: string): Blob => {
  const parts = base64.split(',')
  const mime = parts[0].match(/:(.*?);/)?.[1] || 'image/png'
  const bstr = atob(parts[1])
  let n = bstr.length
  const u8arr = new Uint8Array(n)
  while (n--) {
    u8arr[n] = bstr.charCodeAt(n)
  }
  return new Blob([u8arr], { type: mime })
}

const editorConfig: Partial<IEditorConfig> = {
  placeholder: '',
  MENU_CONF: {
    uploadImage: {
      async customUpload(file: File, insertFn: (url: string, alt: string, href: string) => void) {
        const formData = new FormData()
        formData.append('file', file)
        try {
          const res: any = await request.post('/upload', formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
          })
          const url = res.data?.url || ''
          if (url) {
            insertFn(url, '', '')
          } else {
            ElMessage.error('图片上传失败')
          }
        } catch {
          ElMessage.error('图片上传失败')
        }
      }
    }
  }
}

const handleCreated = (editor: IDomEditor) => {
  editorRef.value = editor
}

const handleEditorChange = () => {
  if (formRef.value) {
    formRef.value.validateField('content').catch(() => {})
  }
}

const skipRTFGroup = (rtf: string, start: number): number => {
  let i = start
  if (rtf[i] !== '{') return i
  let depth = 1
  i++
  while (i < rtf.length && depth > 0) {
    if (rtf[i] === '{') depth++
    else if (rtf[i] === '}') depth--
    i++
  }
  return i
}

const extractImagesFromRTF = (rtf: string): Blob[] => {
  const blobs: Blob[] = []
  let pos = 0
  while (true) {
    const pictIdx = rtf.indexOf('\\pict', pos)
    if (pictIdx === -1) break
    let i = pictIdx + 5 // 跳过 \pict

    // 跳过 \pict 后面所有参数控制字和嵌套组（如 {\*\picprop ...} \pngblip \picw123 等）
    while (i < rtf.length) {
      if (rtf[i] === '\\') {
        i++
        while (i < rtf.length && /[a-zA-Z*]/.test(rtf[i])) i++
        while (i < rtf.length && /[0-9-]/.test(rtf[i])) i++
        while (i < rtf.length && /\s/.test(rtf[i])) i++
      } else if (rtf[i] === '{') {
        i = skipRTFGroup(rtf, i)
        while (i < rtf.length && /\s/.test(rtf[i])) i++
      } else if (/\s/.test(rtf[i])) {
        i++
      } else if (rtf[i] === '}') {
        break
      } else {
        break
      }
    }

    // 读取十六进制数据
    let hexStr = ''
    while (i < rtf.length) {
      const ch = rtf[i]
      if (/[0-9a-fA-F]/.test(ch)) {
        hexStr += ch
      } else if (/\s/.test(ch)) {
        // skip whitespace
      } else if (ch === '}' || ch === '\\' || ch === '{') {
        break
      } else {
        if (hexStr.length > 0) break
      }
      i++
    }

    if (hexStr.length > 0 && hexStr.length % 2 === 0) {
      const bytes = new Uint8Array(hexStr.length / 2)
      for (let j = 0; j < hexStr.length; j += 2) {
        bytes[j / 2] = parseInt(hexStr.substring(j, j + 2), 16)
      }
      let type = 'image/png'
      if (bytes[0] === 0xFF && bytes[1] === 0xD8) type = 'image/jpeg'
      else if (bytes[0] === 0x47 && bytes[1] === 0x49) type = 'image/gif'
      else if (bytes[0] === 0x89 && bytes[1] === 0x50) type = 'image/png'
      else if (bytes[0] === 0x42 && bytes[1] === 0x4D) type = 'image/bmp'
      blobs.push(new Blob([bytes], { type }))
    }
    pos = pictIdx + 1
  }
  return blobs
}

const handleCustomPaste = (editor: IDomEditor, event: ClipboardEvent) => {
  const html = event.clipboardData?.getData('text/html')
  const rtf = event.clipboardData?.getData('text/rtf')
  const hasBase64 = html && html.includes('data:image')

  // 方案1：HTML 中包含 base64 图片，提取并上传
  if (hasBase64) {
    event.preventDefault()
    const div = document.createElement('div')
    div.innerHTML = html
    const imgs = div.querySelectorAll('img')
    Promise.all(
      Array.from(imgs).map(async (img) => {
        const src = img.getAttribute('src') || ''
        if (src.startsWith('data:image')) {
          try {
            const blob = base64ToBlob(src)
            const ext = blob.type.split('/')[1] || 'png'
            const file = new File([blob], `image.${ext}`, { type: blob.type })
            const formData = new FormData()
            formData.append('file', file)
            const res: any = await request.post('/upload', formData, {
              headers: { 'Content-Type': 'multipart/form-data' }
            })
            const url = res.data?.url || ''
            if (url) {
              img.setAttribute('src', url)
            }
          } catch {
            // 上传失败则保留原 base64
          }
        }
      })
    ).then(() => {
      editor.dangerouslyInsertHtml(div.innerHTML)
    })
    return false
  }

  // 方案1.5：HTML 有 img 但无 base64，且有 RTF，尝试从 RTF 提取图片
  if (html && rtf && html.includes('<img') && !hasBase64) {
    const pictIdx = rtf.indexOf('\\pict')
    console.log('[paste] rtf has \\pict:', pictIdx !== -1)
    if (pictIdx !== -1) {
      console.log('[paste] rtf around \\pict:', rtf.substring(Math.max(0, pictIdx - 100), pictIdx + 300))
    } else {
      console.log('[paste] rtf snippet:', rtf.substring(0, 800))
    }
    const blobs = extractImagesFromRTF(rtf)
    console.log('[paste] RTF images extracted:', blobs.length)
    if (blobs.length > 0) {
      event.preventDefault()
      Promise.all(
        blobs.map((blob) => {
          const ext = blob.type.split('/')[1] || 'png'
          const file = new File([blob], `image.${ext}`, { type: blob.type })
          const formData = new FormData()
          formData.append('file', file)
          return request.post('/upload', formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
          })
        })
      ).then((results: any[]) => {
        const urls = results.map((res) => res.data?.url || '').filter(Boolean)
        const div = document.createElement('div')
        div.innerHTML = html
        const imgs = div.querySelectorAll('img')
        imgs.forEach((img, index) => {
          if (urls[index]) {
            img.setAttribute('src', urls[index])
          }
        })
        // 移除仍然带有本地路径的图片，避免浏览器报 Not allowed to load local resource
        div.querySelectorAll('img').forEach((img) => {
          const src = img.getAttribute('src') || ''
          if (src.startsWith('file://')) {
            img.remove()
          }
        })
        editor.dangerouslyInsertHtml(div.innerHTML)
      })
      return false
    }
  }

  // 方案2：剪贴板中有独立的图片文件，直接上传
  const items = event.clipboardData?.items
  if (items) {
    const imageFiles: File[] = []
    for (let i = 0; i < items.length; i++) {
      if (items[i].type.startsWith('image/')) {
        const file = items[i].getAsFile()
        if (file) imageFiles.push(file)
      }
    }
    if (imageFiles.length > 0) {
      event.preventDefault()
      imageFiles.forEach((file) => {
        const formData = new FormData()
        formData.append('file', file)
        request.post('/upload', formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        }).then((res: any) => {
          const url = res.data?.url || ''
          if (url) {
            editor.insertNode({
              type: 'image',
              src: url,
              alt: '',
              href: '',
              children: [{ text: '' }]
            } as any)
          }
        })
      })
      return false
    }
  }

  return true
}

onBeforeUnmount(() => {
  const editor = editorRef.value
  if (editor == null) return
  editor.destroy()
})

const fetchData = async () => {
  loading.value = true
  try {
    const params: any = {
      page: queryForm.page,
      pageSize: queryForm.pageSize
    }
    if (queryForm.title) params.title = queryForm.title
    if (queryForm.categoryId !== undefined) params.categoryId = queryForm.categoryId
    if (queryForm.status !== undefined) params.status = queryForm.status
    if (queryForm.auditStatus !== undefined) params.auditStatus = queryForm.auditStatus
    const res: any = await getArticles(params)
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (error) {
    // request interceptor 已处理错误提示
  } finally {
    loading.value = false
  }
}

const fetchCategories = async () => {
  try {
    const res: any = await getAllCategories()
    categoryList.value = res.data || []
  } catch (error) {
    // ignore
  }
}

const fetchTags = async () => {
  try {
    const res: any = await getAllTags()
    tagList.value = res.data || []
  } catch (error) {
    // ignore
  }
}

const handleSearch = () => {
  queryForm.page = 1
  fetchData()
}

const resetQuery = () => {
  queryForm.title = ''
  queryForm.categoryId = undefined
  queryForm.status = undefined
  queryForm.auditStatus = undefined
  queryForm.page = 1
  fetchData()
}

const handleAdd = () => {
  dialogTitle.value = '新增文章'
  resetForm()
  dialogVisible.value = true
}

const isAuthor = (row: any) => {
  return String(currentUserId.value) === String(row.authorCode)
}

const handleEdit = async (row: any) => {
  if (row.auditStatus === 1) {
    ElMessage.warning('审核中的文章不能编辑')
    return
  }
  dialogTitle.value = '编辑文章'
  resetForm()
  Object.assign(form, {
    id: row.id,
    title: row.title,
    categoryId: row.categoryId,
    tagIds: row.tagIds || [],
    summary: row.summary || '',
    content: row.content || '',
    status: row.status,
    auditStatus: row.auditStatus ?? 0,
    isTop: row.isTop,
    isBold: row.isBold ?? 0,
    defaultColor: row.defaultColor || '',
    cover: row.cover || '',
    source: row.source || '',
    attachments: (row.attachments || []).map((att: any) => ({
      ...att,
      uid: att.uid || Date.now() + Math.random().toString(36).slice(2),
      status: 'success'
    }))
  })
  dialogVisible.value = true
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除文章 "${row.title}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await deleteArticle(row.id)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const currentAuditRow = ref<any>(null)

const getStepActive = (item: any) => {
  if (!item.workflow || !item.workflow.nodes || item.workflow.nodes.length === 0) return -1
  if (item.auditStatus === 1) return item.workflow.nodes.length
  if (item.auditStatus === 2) return item.workflow.nodes.length
  if (item.currentNodeId === 0) return 0
  const idx = item.workflow.nodes.findIndex((n: any) => n.id === item.currentNodeId)
  if (idx >= 0) return idx
  return 0
}

const getNodeHistory = (item: any, nodeId: number) => {
  if (!item.histories || item.histories.length === 0) return null
  return item.histories.find((h: any) => h.nodeId === nodeId) || null
}

const getNodeStatus = (item: any, nodeId: number) => {
  const history = getNodeHistory(item, nodeId)
  if (!history) return 'wait'
  if (history.action === 1) return 'success'
  if (history.action === 2) return 'error'
  return 'wait'
}

const formatAuditTime = (timeStr: string) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  if (isNaN(date.getTime())) return timeStr
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  const h = String(date.getHours()).padStart(2, '0')
  const min = String(date.getMinutes()).padStart(2, '0')
  const s = String(date.getSeconds()).padStart(2, '0')
  return `${y}-${m}-${d} ${h}:${min}:${s}`
}

const handleShowAuditFlow = async (row: any) => {
  auditFlowDialogVisible.value = true
  auditFlowLoading.value = true
  auditFlowList.value = []
  auditFlowArticleTitle.value = row.title || ''
  currentAuditRow.value = row
  try {
    if (columnList.value.length === 0) {
      await fetchColumns()
    }
    const columnIds = row.columnIds || []
    const columns = columnList.value.filter((col: any) => columnIds.includes(col.id))
    // 获取审核进度
    let progressRes: any = { data: [] }
    if (row.auditStatus === 1 || row.auditStatus === 2) {
      try {
        progressRes = await getArticleAuditProgress(row.id)
      } catch {
        progressRes = { data: [] }
      }
    }
    const progressList: any[] = progressRes.data || []
    const list: any[] = []
    for (const col of columns) {
      const item: any = {
        columnId: col.id,
        columnName: col.name,
        workflow: null,
        currentNodeId: 0,
        auditStatus: -1,
        currentApproverId: 0,
        approveUserName: '',
        approveTime: '',
        rejectRemark: '',
        histories: []
      }
      const progress = progressList.find((p: any) => p.columnId === col.id)
      if (progress) {
        item.currentNodeId = progress.currentNodeId
        item.auditStatus = progress.status
        item.approveUserName = progress.approveUserName || ''
        item.approveTime = progress.approveTime || ''
        item.rejectRemark = progress.rejectRemark || ''
      }
      if (col.workflowId) {
        try {
          const res: any = await getWorkflowByID(col.workflowId)
          item.workflow = res.data || null
          if (item.workflow && item.workflow.nodes && item.currentNodeId) {
            const node = item.workflow.nodes.find((n: any) => n.id === item.currentNodeId)
            if (node) {
              item.currentApproverId = node.approverId || 0
            }
          }
        } catch {
          item.workflow = null
        }
      }
      // 获取审核历史
      try {
        const historyRes: any = await getArticleAuditHistory(row.id, col.id)
        item.histories = historyRes.data || []
      } catch {
        item.histories = []
      }
      list.push(item)
    }
    auditFlowList.value = list
  } finally {
    auditFlowLoading.value = false
  }
}

const handleAdvanceAuditNode = async (columnId: number) => {
  if (!currentAuditRow.value) return
  try {
    const { value } = await ElMessageBox.prompt('请输入审核通过原因（可选）', '审核通过', {
      confirmButtonText: '确定通过',
      cancelButtonText: '取消',
      inputPattern: /^.{0,500}$/,
      inputErrorMessage: '原因最多500字'
    })
    await advanceArticleAudit(currentAuditRow.value.id, columnId, value || '')
    ElMessage.success('已通过当前节点')
    handleShowAuditFlow(currentAuditRow.value)
    fetchData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error?.message || '操作失败')
    }
  }
}

const handleRejectAuditNode = async (columnId: number) => {
  if (!currentAuditRow.value) return
  try {
    const { value } = await ElMessageBox.prompt('请输入驳回原因', '审核驳回', {
      confirmButtonText: '确定驳回',
      cancelButtonText: '取消',
      inputPattern: /\S+/,
      inputErrorMessage: '驳回原因不能为空'
    })
    await rejectArticleAudit(currentAuditRow.value.id, columnId, value || '')
    ElMessage.success('已驳回')
    handleShowAuditFlow(currentAuditRow.value)
    fetchData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error?.message || '操作失败')
    }
  }
}

const handleAudit = async (row: any) => {
  if (columnList.value.length === 0) {
    await fetchColumns()
  }
  const columns = columnList.value.filter((col: any) => (row.columnIds || []).includes(col.id))
  const noWorkflowColumns = columns.filter((col: any) => !col.workflowId)
  let message = `确定要提交文章 "${row.title}" 进行审核吗？`
  if (noWorkflowColumns.length > 0) {
    const names = noWorkflowColumns.map((col: any) => col.name).join('、')
    message += `\n\n以下栏目未配置审核流程，将直接通过：${names}`
  }
  ElMessageBox.confirm(message, '提交审核', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await auditArticle(row.id, 1)
    ElMessage.success('提交审核成功')
    fetchData()
  })
}

const handleReAudit = async (row: any) => {
  if (columnList.value.length === 0) {
    await fetchColumns()
  }
  const columns = columnList.value.filter((col: any) => (row.columnIds || []).includes(col.id))
  const noWorkflowColumns = columns.filter((col: any) => !col.workflowId)
  let message = `确定要重新提交文章 "${row.title}" 进行审核吗？此操作将清空之前的栏目审核记录。`
  if (noWorkflowColumns.length > 0) {
    const names = noWorkflowColumns.map((col: any) => col.name).join('、')
    message += `\n\n以下栏目未配置审核流程，将直接通过：${names}`
  }
  ElMessageBox.confirm(
    message,
    '重新提交审核',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    await restartArticleAudit(row.id)
    ElMessage.success('重新提交审核成功')
    fetchData()
  })
}

const handleWithdrawAudit = (row: any) => {
  ElMessageBox.confirm(`确定要撤回文章 "${row.title}" 的审核吗？`, '撤回审核', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await withdrawArticleAudit(row.id)
    ElMessage.success('撤回审核成功')
    fetchData()
  })
}

const handleOffShelf = (row: any) => {
  ElMessageBox.confirm(`确定要下线文章 "${row.title}" 吗？`, '下线确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await updateArticleStatus(row.id, 2)
    ElMessage.success('下线成功')
    fetchData()
  })
}

const handleSetColumns = async (row: any) => {
  currentArticleId.value = row.id
  columnDialogTitle.value = row.title
  selectedColumnIds.value = row.columnIds || []
  if (pageList.value.length === 0) {
    await fetchPages()
  }
  if (columnList.value.length === 0) {
    await fetchColumns()
  }
  columnDialogVisible.value = true
}

const handleSubmitColumns = async () => {
  if (!currentArticleId.value) return
  columnSubmitLoading.value = true
  try {
    await setArticleColumns(currentArticleId.value, selectedColumnIds.value)
    ElMessage.success('栏目设置成功')
    columnDialogVisible.value = false
    fetchData()
  } finally {
    columnSubmitLoading.value = false
  }
}

const previewVisible = ref(false)
const previewData = reactive({
  title: '',
  author: '',
  authorCode: '',
  source: '',
  categoryName: '',
  createTime: '',
  summary: '',
  content: '',
  attachments: [] as any[]
})

const handlePreview = (row: any) => {
  Object.assign(previewData, {
    title: row.title,
    author: row.author,
    source: row.source || '',
    categoryName: row.categoryName,
    createTime: row.createTime,
    summary: row.summary || '',
    content: row.content || '',
    attachments: row.attachments || []
  })
  previewVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    const data = {
      title: form.title,
      categoryId: form.categoryId as number,
      tagIds: form.tagIds,
      summary: form.summary,
      content: form.content,
      status: form.status,
      auditStatus: form.auditStatus,
      isTop: form.isTop,
      isBold: form.isBold,
      defaultColor: form.defaultColor,
      cover: form.cover,
      source: form.source,
      attachments: buildAttachmentPayload(form.attachments)
    }
    if (form.id) {
      await updateArticle(form.id, data)
      ElMessage.success('修改成功')
    } else {
      await createArticle(data)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally {
    submitLoading.value = false
  }
}

const handleCoverUpload = async (options: any) => {
  try {
    const res: any = await uploadFile(options.file, 'article')
    form.cover = res.data?.url || res.url || ''
    ElMessage.success('封面图上传成功')
  } catch (error: any) {
    ElMessage.error(error?.message || '封面图上传失败')
  }
}

const handleRemoveCover = () => {
  form.cover = ''
}

const handleAttachmentBeforeUpload = (file: File) => {
  const isDuplicate = form.attachments.some(
    (att: any) => att.name === file.name && att.size === file.size
  )
  if (isDuplicate) {
    ElMessage.warning(`文件 "${file.name}" 已存在，请勿重复上传`)
    return false
  }
  return true
}

const handleAttachmentUpload = async (options: any) => {
  const file = options.file
  const formData = new FormData()
  formData.append('file', file)
  formData.append('dir', 'attachment')
  try {
    const res: any = await request.post('/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    const url = res.data?.url || ''
    if (url) {
      options.onSuccess({ url, name: file.name, size: file.size, uid: file.uid })
      // 手动同步 url 到 form.attachments，确保提交时能正确读取
      const idx = form.attachments.findIndex((a: any) => a.uid === file.uid)
      if (idx >= 0) {
        form.attachments[idx].url = url
        form.attachments[idx].status = 'success'
      }
    } else {
      options.onError(new Error('上传失败'))
      ElMessage.error('附件上传失败')
    }
  } catch (error: any) {
    options.onError(error)
    ElMessage.error(error?.message || '附件上传失败')
  }
}

const handleAttachmentRemove = (file: any, fileList: any[]) => {
  form.attachments = fileList
}

const buildAttachmentPayload = (attachments: any[]) => {
  const seen = new Set<string>()
  return attachments
    .map((att: any) => ({
      name: att.name,
      url: att.url || att.response?.url || '',
      size: att.size || 0
    }))
    .filter((att: any) => {
      if (!att.url || seen.has(att.url)) return false
      seen.add(att.url)
      return true
    })
}

const resetForm = () => {
  form.id = undefined
  form.title = ''
  form.categoryId = undefined
  form.tagIds = []
  form.summary = ''
  form.content = ''
  form.status = 0
  form.auditStatus = 0
  form.isTop = 0
  form.isBold = 0
  form.defaultColor = ''
  form.cover = ''
  form.source = ''
  form.attachments = []
}

const handleSizeChange = (val: number) => {
  queryForm.pageSize = val
  fetchData()
}

const handleCurrentChange = (val: number) => {
  queryForm.page = val
  fetchData()
}

const checkAutoAudit = () => {
  const auditArticleId = route.query.auditArticleId
  if (auditArticleId) {
    const row = tableData.value.find((item: any) => String(item.id) === String(auditArticleId))
    if (row) {
      handleShowAuditFlow(row)
    }
    router.replace({ path: '/content/article', query: {} })
  }
}

onMounted(() => {
  fetchData().then(() => checkAutoAudit())
  fetchCategories()
  fetchTags()
})
</script>

<style scoped lang="scss">
.page-container {
  .search-card {
    margin-bottom: 20px;
    border-radius: 12px;
    border: 1px solid #e6f2ff;
  }

  .table-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 600;
      color: #2c3e50;
    }
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}

.article-form {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 8px 16px;

  .form-section {
    .section-title {
      font-size: 16px;
      font-weight: 600;
      color: #2c3e50;
      margin-bottom: 16px;
      padding-left: 10px;
      border-left: 4px solid #409eff;
    }
  }

  .el-divider {
    margin: 20px 0;
  }

  .editor-form-item {
    flex: 1;
    margin-bottom: 0;

    :deep(.el-form-item__content) {
      height: calc(100vh - 520px);
      display: block;
    }
  }
}

.editor-wrapper {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  overflow: hidden;
  height: 100%;
  display: flex;
  flex-direction: column;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.02);

  :deep(.w-e-text-container) {
    flex: 1;
    overflow-y: auto;
  }
}

.attachment-uploader {
  .attachment-tip {
    font-size: 13px;
    color: #909399;
    margin-top: 8px;
    line-height: 1.5;
  }

  :deep(.el-upload-list) {
    margin-top: 12px;

    .el-upload-list__item {
      border-radius: 6px;
      transition: all 0.2s;

      &:hover {
        background-color: #f5f7fa;
      }
    }
  }
}

.preview-content {
  h2 {
    margin: 0 0 12px 0;
    color: #2c3e50;
  }

  .preview-meta {
    color: #999;
    font-size: 14px;
    margin-bottom: 16px;

    span {
      margin-right: 16px;
    }
  }

  .preview-summary {
    background: #f5f7fa;
    padding: 12px;
    border-radius: 8px;
    margin-bottom: 16px;
    color: #666;
  }

  .preview-body {
    line-height: 1.8;
    color: #333;
  }

  .preview-attachments {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px dashed #e6f2ff;

    strong {
      display: block;
      margin-bottom: 8px;
      color: #2c3e50;
    }

    .attachment-list {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .attachment-item {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      color: #409eff;
      text-decoration: none;
      font-size: 14px;

      &:hover {
        text-decoration: underline;
      }

      .attachment-name {
        word-break: break-all;
      }
    }
  }
}

.audit-status-clickable {
  cursor: pointer;
}

.audit-flow-subtitle {
  color: #666;
  font-size: 14px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e6f2ff;
}

.audit-flow-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.audit-flow-card {
  background: #fff;
  border: 1px solid #e6f2ff;
  border-radius: 10px;
  overflow: hidden;
  transition: box-shadow 0.2s;

  &:hover {
    box-shadow: 0 4px 12px rgba(64, 158, 255, 0.1);
  }
}

.audit-flow-card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: linear-gradient(90deg, #f5faff 0%, #ffffff 100%);
  border-bottom: 1px solid #e6f2ff;
}

.audit-flow-index {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: #409eff;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 2px 6px rgba(64, 158, 255, 0.25);
}

.audit-flow-column-name {
  flex: 1;
  font-size: 15px;
  font-weight: 600;
  color: #2c3e50;
}

.audit-flow-card-body {
  padding: 16px;
}

.audit-flow-steps {
  :deep(.el-step__title) {
    font-size: 13px;
  }
}

.audit-flow-rejected {
  border-color: #fde2e2 !important;

  .audit-flow-card-header {
    background: linear-gradient(90deg, #fef5f5 0%, #ffffff 100%);
    border-bottom-color: #fde2e2;
  }
}

.audit-flow-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px dashed #e6f2ff;
}

.audit-flow-no-auth {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px dashed #e6f2ff;
  color: #909399;
  font-size: 13px;
}

.audit-flow-result {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px dashed #e6f2ff;
  color: #67c23a;
  font-size: 13px;
}

.audit-step-desc {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  margin-top: 4px;
  font-size: 12px;
  color: #67c23a;
}

.audit-step-desc.audit-step-reject {
  color: #f56c6c;
}

.audit-flow-reject-result {
  color: #f56c6c;
}

.article-cover-uploader {
  .cover-uploader {
    width: 240px;
    height: 140px;
    border: 2px dashed var(--el-border-color);
    border-radius: 10px;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    color: #8c939d;
    transition: all 0.3s;
    background: #fafbfc;

    &:hover {
      border-color: var(--el-color-primary);
      background: #f5faff;
      color: var(--el-color-primary);
    }

    .uploader-icon {
      font-size: 32px;
    }

    .uploader-text {
      font-size: 14px;
      font-weight: 500;
    }

    .uploader-hint {
      font-size: 12px;
      color: #c0c4cc;
    }
  }

  .cover-preview {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;

    .cover-image-wrapper {
      position: relative;
      width: 240px;
      height: 140px;
      border-radius: 10px;
      overflow: hidden;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);

      .cover-overlay {
        position: absolute;
        inset: 0;
        background: rgba(0, 0, 0, 0.5);
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 6px;
        color: #fff;
        font-size: 13px;
        opacity: 0;
        transition: opacity 0.3s;
        cursor: pointer;

        &:hover {
          opacity: 1;
        }
      }
    }
  }
}

:global(.el-image-viewer__wrapper) {
  z-index: 9999 !important;
  .el-image-viewer__canvas {
    width: 600px !important;
    height: 500px !important;
    left: 50% !important;
    top: 50% !important;
    transform: translate(-50%, -50%) !important;
  }
  .el-image-viewer__img {
    max-width: 600px !important;
    max-height: 500px !important;
    width: auto !important;
    height: auto !important;
    object-fit: contain !important;
  }
}
</style>

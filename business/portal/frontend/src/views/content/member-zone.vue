<template>
  <div class="page-container">
    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <span>会员专区</span>
          <div class="header-actions">
            <el-button v-if="isAdmin" @click="openColumnManager">
              <el-icon><Grid /></el-icon>
              会员栏目管理
            </el-button>
            <el-button type="primary" @click="handleAddContent">
              <el-icon><Plus /></el-icon>
              发布会员内容
            </el-button>
          </div>
        </div>
      </template>

      <!-- 主界面只呈现会员内容；会员栏目维护收进「会员栏目管理」弹窗（见下方 dialog） -->
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
          <el-select
            v-model="contentQuery.type"
            placeholder="全部类型"
            clearable
            style="width: 120px"
          >
            <el-option label="新闻" :value="1" />
            <el-option label="数据" :value="2" />
            <el-option label="视频" :value="3" />
            <el-option label="报刊" :value="4" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="contentQuery.status"
            placeholder="全部状态"
            clearable
            style="width: 120px"
          >
            <el-option label="草稿" :value="0" />
            <el-option label="已发布" :value="1" />
            <el-option label="已下线" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="发布时间">
          <el-date-picker
            v-model="contentQuery.publishRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item label="栏目状态">
          <el-select
            v-model="contentQuery.columnStatus"
            placeholder="全部状态"
            clearable
            style="width: 120px"
          >
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleContentSearch">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="resetContentQuery">
            <el-icon><RefreshRight /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="contentLoading" :data="contentData" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="title" label="内容标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="memberColumnName" label="会员栏目" min-width="130" />
        <el-table-column prop="type" label="类型" width="90" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTagType(row.type)">{{ typeName(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="关键信息" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ keyInfo(row) }}</template>
        </el-table-column>
        <el-table-column prop="author" label="作者" width="110" />
        <el-table-column prop="publishTime" label="发布时间" width="170" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="statusTagType(row.status)">
              {{ statusName(row.status) }}
            </el-tag>
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
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button link type="primary" @click="handleToggleContentStatus(row)">
              {{ row.status === 1 ? '下线' : '发布' }}
            </el-button>
            <el-button link type="danger" @click="handleDeleteContent(row)">
              <el-icon><Delete /></el-icon>
              删除
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
    </el-card>

    <!-- 会员栏目管理：栏目列表与增删改全部收在该弹窗内，主界面不再展示栏目列表 -->
    <!-- 宽度按表格全部列的最小宽度之和（序号60 + 名称150 + 描述200 + 排序90 + 状态100 + 创建时间170 + 操作180 = 950）外加班内边距留出余量，避免横向滚动条 -->
    <el-dialog
      v-model="columnManagerVisible"
      title="会员栏目管理"
      width="1080px"
      align-center
      :close-on-click-modal="false"
    >
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
          <el-select
            v-model="columnQuery.status"
            placeholder="全部状态"
            clearable
            style="width: 120px"
          >
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleColumnSearch">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="resetColumnQuery">
            <el-icon><RefreshRight /></el-icon>
            重置
          </el-button>
          <el-button type="primary" @click="handleAddColumn">
            <el-icon><Plus /></el-icon>
            新增会员栏目
          </el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="columnLoading" :data="columnData" border stripe max-height="420">
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="name" label="栏目名称" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="90" align="center" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="(val: number) => handleColumnStatusChange(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" />
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEditColumn(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button link type="danger" @click="handleDeleteColumn(row)">
              <el-icon><Delete /></el-icon>
              删除
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

      <template #footer>
        <el-button @click="columnManagerVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="columnFormVisible"
      :title="columnDialogTitle"
      width="600px"
      align-center
      destroy-on-close
      append-to-body
      :close-on-click-modal="false"
    >
      <el-form ref="columnFormRef" :model="columnForm" :rules="columnRules" label-width="90px">
        <el-form-item label="栏目名称" prop="name">
          <el-input v-model="columnForm.name" placeholder="请输入会员栏目名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="columnForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述"
          />
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
        <el-button @click="columnFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="columnSubmitLoading" @click="handleSubmitColumn">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="contentDialogVisible"
      :title="contentDialogTitle"
      width="900px"
      align-center
      destroy-on-close
      :close-on-click-modal="false"
      @closed="stopResignTimer"
    >
      <el-form ref="contentFormRef" :model="contentForm" :rules="contentRules" label-width="160px">
        <el-form-item label="内容标题" prop="title">
          <el-input
            v-model="contentForm.title"
            placeholder="请输入内容标题"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="会员栏目" prop="memberColumnId">
          <el-select
            v-model="contentForm.memberColumnId"
            placeholder="请选择会员栏目"
            style="width: 100%"
          >
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
            <el-radio :value="4">报刊</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 封面图：新闻 / 视频 / 报刊 使用 -->
        <el-form-item v-if="contentForm.type !== 2" label="封面图" prop="cover">
          <el-upload :show-file-list="false" :http-request="handleCoverUpload" accept="image/*">
            <el-button>
              <el-icon><Upload /></el-icon>
              上传封面图
            </el-button>
          </el-upload>
          <div v-if="contentForm.cover" class="cover-preview">
            <el-image :src="coverDisplay" fit="cover" />
            <el-button link type="danger" @click="clearCover">移除</el-button>
          </div>
        </el-form-item>

        <!-- 新闻：文章内容 + 文章附件 -->
        <template v-if="contentForm.type === 1">
          <el-form-item label="文章内容" prop="content" class="editor-form-item">
            <div class="editor-wrapper">
              <Toolbar
                style="border-bottom: 1px solid #e4e7ed"
                :editor="editorRef"
                :default-config="toolbarConfig"
                mode="default"
              />
              <Editor
                v-model="contentForm.content"
                :default-config="editorConfig"
                mode="default"
                @on-created="handleEditorCreated"
              />
            </div>
          </el-form-item>
          <el-form-item label="文章附件" prop="attachmentUrl">
            <el-upload :show-file-list="false" :http-request="handleAttachmentUpload">
              <el-button>
                <el-icon><Paperclip /></el-icon>
                上传文章附件
              </el-button>
            </el-upload>
            <span v-if="contentForm.attachmentName" class="attachment-name">
              <el-link
                type="primary"
                :href="attachmentDisplay"
                :disabled="!attachmentDisplay"
                target="_blank"
                rel="noopener"
              >
                {{ contentForm.attachmentName }}
              </el-link>
              <el-button link type="danger" @click="clearAttachment">移除</el-button>
            </span>
          </el-form-item>
        </template>

        <!-- 数据 -->
        <template v-if="contentForm.type === 2">
          <el-form-item label="数据年份" prop="dataYear">
            <el-date-picker
              v-model="contentForm.dataYear"
              type="year"
              placeholder="请选择数据年份"
              value-format="YYYY"
            />
          </el-form-item>
          <el-form-item label="单位名称" prop="unitName">
            <el-input v-model="contentForm.unitName" placeholder="请输入单位名称" maxlength="200" />
          </el-form-item>
          <el-form-item label="所属省份及直辖市" prop="province">
            <el-select
              v-model="contentForm.province"
              placeholder="请选择所属省份及直辖市"
              clearable
              filterable
              style="width: 100%"
              @change="handleProvinceChange"
            >
              <el-option v-for="item in PROVINCE_OPTIONS" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="所属地区" prop="region">
            <el-select
              v-model="contentForm.region"
              placeholder="选择省份后自动带出，可手工调整"
              clearable
              filterable
              style="width: 100%"
            >
              <el-option v-for="item in REGION_OPTIONS" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="是否一带" prop="isBelt">
            <el-radio-group v-model="contentForm.isBelt">
              <el-radio :value="1">是</el-radio>
              <el-radio :value="0">否</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="是否一轴" prop="isAxis">
            <el-radio-group v-model="contentForm.isAxis">
              <el-radio :value="1">是</el-radio>
              <el-radio :value="0">否</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="细分领域" prop="subField">
            <el-select
              v-model="contentForm.subField"
              placeholder="请选择细分领域"
              clearable
              filterable
              style="width: 100%"
            >
              <el-option
                v-for="item in SUB_FIELD_OPTIONS"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="主营业务收入（亿元）" prop="mainBusinessIncome">
            <el-input-number
              v-model="contentForm.mainBusinessIncome"
              :min="0"
              :precision="2"
              :step="100"
              controls-position="right"
            />
          </el-form-item>
        </template>

        <!-- 视频：完整视频 + 预览视频 -->
        <template v-if="contentForm.type === 3">
          <el-form-item label="完整视频" prop="fullVideoUrl">
            <div class="media-row">
              <el-input
                v-model="contentForm.fullVideoUrl"
                placeholder="上传后自动填入，也可直接填写视频地址"
              />
              <el-upload
                :show-file-list="false"
                :http-request="handleFullVideoUpload"
                accept="video/mp4"
              >
                <el-button>
                  <el-icon><VideoCamera /></el-icon>
                  上传完整视频
                </el-button>
              </el-upload>
            </div>
          </el-form-item>
          <el-form-item label="预览视频" prop="previewVideoUrl">
            <div class="media-row">
              <el-input
                v-model="contentForm.previewVideoUrl"
                placeholder="上传后自动填入，也可直接填写视频地址"
              />
              <el-upload
                :show-file-list="false"
                :http-request="handlePreviewVideoUpload"
                accept="video/mp4"
              >
                <el-button>
                  <el-icon><VideoCamera /></el-icon>
                  上传预览视频
                </el-button>
              </el-upload>
            </div>
          </el-form-item>
        </template>

        <!-- 报刊：期号 + 出版年月 + 摘要 + 报刊文件 -->
        <template v-if="contentForm.type === 4">
          <el-form-item label="期号" prop="issueNo">
            <el-input v-model="contentForm.issueNo" placeholder="请输入期号" maxlength="50" />
          </el-form-item>
          <el-form-item label="出版年月" prop="publishYearMonth">
            <el-date-picker
              v-model="contentForm.publishYearMonth"
              type="month"
              placeholder="请选择出版年月"
              value-format="YYYY-MM"
            />
          </el-form-item>
          <el-form-item label="摘要" prop="summary">
            <el-input
              v-model="contentForm.summary"
              type="textarea"
              :rows="3"
              placeholder="请输入摘要"
              maxlength="500"
              show-word-limit
            />
          </el-form-item>
          <el-form-item label="报刊文件" prop="paperFileUrl">
            <el-upload :show-file-list="false" :http-request="handlePaperFileUpload">
              <el-button>
                <el-icon><Paperclip /></el-icon>
                上传报刊文件
              </el-button>
            </el-upload>
            <span v-if="contentForm.paperFileName" class="attachment-name">
              <el-link
                type="primary"
                :href="paperDisplay"
                :disabled="!paperDisplay"
                target="_blank"
                rel="noopener"
              >
                {{ contentForm.paperFileName }}
              </el-link>
              <el-button link type="danger" @click="clearPaperFile">移除</el-button>
            </span>
          </el-form-item>
        </template>
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
        <el-button type="primary" :loading="contentSubmitLoading" @click="handleSubmitContent">
          确定
        </el-button>
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
  Grid,
  Paperclip,
  Plus,
  RefreshRight,
  Search,
  Upload,
  VideoCamera,
} from '@element-plus/icons-vue'
import '@wangeditor/editor/dist/css/style.css'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import type { IDomEditor, IEditorConfig, IToolbarConfig } from '@wangeditor/editor'

import { useUserStore } from '@/stores/user'
import { hasAdminRole } from '@/utils/permission'
import { PROVINCE_OPTIONS, REGION_OPTIONS, resolveRegionByProvince } from '@/utils/regions'
import { SUB_FIELD_OPTIONS } from '@/utils/member-zone-options'
import { uploadFile } from '@/api/upload'
import {
  createMemberColumn,
  createMemberContent,
  deleteMemberColumn,
  deleteMemberContent,
  getMemberColumns,
  getMemberContent,
  getMemberContents,
  signMemberFile,
  updateMemberColumn,
  updateMemberColumnStatus,
  updateMemberContent,
  updateMemberContentStatus,
} from '@/api/member-zone'

const userStore = useUserStore()
// 会员栏目的新增/编辑/删除/改状态均为管理员接口（写路由在 admin 组）
const isAdmin = computed(() => hasAdminRole(userStore.userInfo?.roleIds))

// ============================ 会员栏目 ============================
const columnManagerVisible = ref(false)
const columnLoading = ref(false)
const columnTotal = ref(0)
const columnData = ref<any[]>([])
const columnFormVisible = ref(false)
const columnDialogTitle = ref('')
const columnSubmitLoading = ref(false)
const columnFormRef = ref()

const columnQuery = reactive({
  page: 1,
  pageSize: 10,
  name: '',
  status: undefined as number | undefined,
})

const columnForm = reactive({
  id: undefined as number | undefined,
  name: '',
  description: '',
  sort: 0,
  status: 1,
})

const columnRules = {
  name: [{ required: true, message: '请输入栏目名称', trigger: 'blur' }],
}

// 供「会员栏目」下拉使用（含禁用项，便于编辑历史内容时保留原栏目）
const columnOptions = ref<any[]>([])

const resetColumnForm = () => {
  Object.assign(columnForm, {
    id: undefined,
    name: '',
    description: '',
    sort: 0,
    status: 1,
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

// 会员栏目维护收在弹窗里，打开时再拉列表（主界面不展示栏目列表）
const openColumnManager = () => {
  columnManagerVisible.value = true
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
  columnFormVisible.value = true
}

const handleEditColumn = (row: any) => {
  columnDialogTitle.value = '编辑会员栏目'
  Object.assign(columnForm, {
    id: row.id,
    name: row.name,
    description: row.description,
    sort: row.sort,
    status: row.status,
  })
  columnFormVisible.value = true
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
    columnFormVisible.value = false
    fetchColumns()
    fetchColumnOptions()
  } catch {
    // 错误提示由 request 拦截器统一给出，此处仅兜住 rejected promise
  } finally {
    columnSubmitLoading.value = false
  }
}

const handleColumnStatusChange = async (row: any, val: number) => {
  // 在途标记：连点会发出并发 PATCH，先失败的那次回滚会覆盖后一次的成功结果
  if (row.switching) return
  row.switching = true
  try {
    await updateMemberColumnStatus(row.id, val)
    ElMessage.success(`会员栏目状态已${val === 1 ? '启用' : '禁用'}`)
    fetchColumnOptions()
  } catch {
    row.status = val === 1 ? 0 : 1
  } finally {
    row.switching = false
  }
}

const handleDeleteColumn = (row: any) => {
  ElMessageBox.confirm(`确定要删除会员栏目 「${row.name}」 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
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
  status: undefined as number | undefined,
  // 所属会员栏目的启用状态：1 启用 / 0 禁用；undefined = 全部
  columnStatus: undefined as number | undefined,
  // 发布时间区间（YYYY-MM-DD）；空数组 = 不限
  publishRange: [] as string[],
})

const contentForm = reactive({
  id: undefined as number | undefined,
  memberColumnId: undefined as number | undefined,
  title: '',
  type: 1,
  source: '',
  publishTime: '',
  status: 0,
  isTop: 0,
  cover: '',
  // 新闻
  content: '',
  attachmentName: '',
  attachmentUrl: '',
  // 数据
  dataYear: '',
  unitName: '',
  province: '',
  region: '',
  isBelt: 0,
  isAxis: 0,
  subField: '',
  mainBusinessIncome: 0,
  // 视频
  fullVideoUrl: '',
  previewVideoUrl: '',
  // 报刊
  issueNo: '',
  publishYearMonth: '',
  summary: '',
  paperFileName: '',
  paperFileUrl: '',
})

/** 富文本正文转纯文本后判断是否真的填了内容（空编辑器会产出 <p><br></p>） */
const hasRichText = (html: string) =>
  (html || '')
    .replace(/<[^>]*>/g, '')
    .replace(/&nbsp;/g, ' ')
    .trim().length > 0

// 与后端 services.validateMemberContentByType 口径一致：
// 仅当状态为「已发布」时才校验各类型的业务关键字段（草稿/已下线允许先建后补）
const contentRules = computed(() => {
  const rules: Record<string, any> = {
    title: [{ required: true, message: '请输入内容标题', trigger: 'blur' }],
    memberColumnId: [{ required: true, message: '请选择会员栏目', trigger: 'change' }],
    type: [{ required: true, message: '请选择内容类型', trigger: 'change' }],
  }
  if (contentForm.status !== 1) return rules
  if (contentForm.type === 1) {
    rules.content = [
      {
        validator: (_rule: any, _value: any, callback: (error?: Error) => void) => {
          if (hasRichText(contentForm.content) || contentForm.attachmentUrl) callback()
          else callback(new Error('发布新闻前请填写文章内容或上传文章附件'))
        },
        trigger: 'blur',
      },
    ]
  } else if (contentForm.type === 2) {
    rules.dataYear = [{ required: true, message: '发布数据前请填写数据年份', trigger: 'change' }]
    rules.unitName = [{ required: true, message: '发布数据前请填写单位名称', trigger: 'blur' }]
  } else if (contentForm.type === 3) {
    rules.fullVideoUrl = [{ required: true, message: '发布视频前请上传完整视频', trigger: 'blur' }]
  } else if (contentForm.type === 4) {
    rules.issueNo = [{ required: true, message: '发布报刊前请填写期号', trigger: 'blur' }]
    rules.publishYearMonth = [
      { required: true, message: '发布报刊前请填写出版年月', trigger: 'change' },
    ]
    rules.paperFileUrl = [{ required: true, message: '发布报刊前请上传报刊文件', trigger: 'blur' }]
  }
  return rules
})

const typeName = (type: number) => {
  const map: Record<number, string> = { 1: '新闻', 2: '数据', 3: '视频', 4: '报刊' }
  return map[type] || '新闻'
}

const typeTagType = (type: number): 'primary' | 'success' | 'warning' | 'danger' => {
  const map: Record<number, 'primary' | 'success' | 'warning' | 'danger'> = {
    1: 'primary',
    2: 'success',
    3: 'warning',
    4: 'danger',
  }
  return map[type] || 'primary'
}

/** 列表「关键信息」列：按类型取最能区分该条内容的字段 */
const keyInfo = (row: any) => {
  if (row.type === 2) return [row.dataYear, row.unitName].filter(Boolean).join(' · ') || '-'
  if (row.type === 4) return [row.issueNo, row.publishYearMonth].filter(Boolean).join(' · ') || '-'
  return '-'
}

const statusName = (status: number) => {
  const map: Record<number, string> = { 0: '草稿', 1: '已发布', 2: '已下线' }
  return map[status] || '草稿'
}

const statusTagType = (status: number): 'info' | 'success' | 'warning' => {
  const map: Record<number, 'info' | 'success' | 'warning'> = {
    0: 'info',
    1: 'success',
    2: 'warning',
  }
  return map[status] || 'info'
}

const resetContentForm = () => {
  // 展示用的签名地址也要一并清空，避免上一份内容的预览残留
  coverDisplay.value = ''
  attachmentDisplay.value = ''
  paperDisplay.value = ''
  Object.assign(contentForm, {
    id: undefined,
    memberColumnId: undefined,
    title: '',
    type: 1,
    source: '',
    publishTime: '',
    status: 0,
    isTop: 0,
    cover: '',
    content: '',
    attachmentName: '',
    attachmentUrl: '',
    dataYear: '',
    unitName: '',
    province: '',
    region: '',
    isBelt: 0,
    isAxis: 0,
    subField: '',
    mainBusinessIncome: 0,
    fullVideoUrl: '',
    previewVideoUrl: '',
    issueNo: '',
    publishYearMonth: '',
    summary: '',
    paperFileName: '',
    paperFileUrl: '',
  })
}

// 选定「所属省份及直辖市」后自动带出「所属地区」（用户仍可手工调整为其它区域；清空省份则一并清空）
const handleProvinceChange = (value: string) => {
  contentForm.region = resolveRegionByProvince(value)
}

const fetchContents = async () => {
  contentLoading.value = true
  try {
    const params: any = { page: contentQuery.page, pageSize: contentQuery.pageSize }
    if (contentQuery.title) params.title = contentQuery.title
    if (contentQuery.columnId !== undefined) params.columnId = contentQuery.columnId
    if (contentQuery.type !== undefined) params.type = contentQuery.type
    if (contentQuery.status !== undefined) params.status = contentQuery.status
    if (contentQuery.columnStatus !== undefined) params.columnStatus = contentQuery.columnStatus
    if (contentQuery.publishRange && contentQuery.publishRange.length === 2) {
      params.publishStart = contentQuery.publishRange[0]
      params.publishEnd = contentQuery.publishRange[1]
    }
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
  contentQuery.columnStatus = undefined
  contentQuery.publishRange = []
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
      source: detail.source,
      publishTime: detail.publishTime || '',
      status: detail.status,
      isTop: detail.isTop,
      cover: detail.cover,
      // 私有图需签名地址才能在编辑器内显示，故载入时逐个签发并注入（提交时剔除，不落库）
      content: await injectSignedFileUrls(detail.content || ''),
      attachmentName: detail.attachmentName,
      attachmentUrl: detail.attachmentUrl,
      dataYear: detail.dataYear,
      unitName: detail.unitName,
      province: detail.province,
      region: detail.region,
      isBelt: detail.isBelt ?? 0,
      isAxis: detail.isAxis ?? 0,
      subField: detail.subField,
      mainBusinessIncome: detail.mainBusinessIncome ?? 0,
      fullVideoUrl: detail.fullVideoUrl,
      previewVideoUrl: detail.previewVideoUrl,
      issueNo: detail.issueNo,
      publishYearMonth: detail.publishYearMonth,
      summary: detail.summary,
      paperFileName: detail.paperFileName,
      paperFileUrl: detail.paperFileUrl,
    })
    // 表单里的私有文件也需要签名地址才能预览/打开（入库地址不带签名）
    coverDisplay.value = await ensureSignedFile(detail.cover)
    attachmentDisplay.value = await ensureSignedFile(detail.attachmentUrl)
    paperDisplay.value = await ensureSignedFile(detail.paperFileUrl)
    contentDialogVisible.value = true
    // 编辑期间定期重签正文内联图（签名 5 分钟过期，长文编辑会中途裂图）
    startResignTimer()
  } catch {
    // 错误提示由 request 拦截器统一给出
  }
}

const handleSubmitContent = async () => {
  const valid = await contentFormRef.value?.validate().catch(() => false)
  if (!valid) return
  contentSubmitLoading.value = true
  try {
    // 正文里的会员专区图片地址在提交前剔除签名参数，避免把签名写进数据库
    const payload = { ...contentForm, content: stripFileSignParams(contentForm.content) }
    if (contentForm.id) {
      await updateMemberContent(contentForm.id, payload)
      ElMessage.success('修改成功')
    } else {
      await createMemberContent(payload)
      ElMessage.success('发布成功')
    }
    contentDialogVisible.value = false
    fetchContents()
  } catch {
    // 错误提示由 request 拦截器统一给出，此处仅兜住 rejected promise
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
    type: 'warning',
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
    type: 'warning',
  })
    .then(async () => {
      await deleteMemberContent(row.id)
      ElMessage.success('删除成功')
      fetchContents()
    })
    .catch(() => {})
}

// ============================ 上传（私有文件 + 短时效签名 URL） ============================
// 会员专区的上传统一传 dir=member：后端落到私有目录 ./private_uploads，返回 { url, signedUrl }：
//   - url       形如 /business_portal/api/member-files/xxx，是**入库用的干净地址**（不带签名）；
//   - signedUrl 是当场签发的短时效地址（5 分钟），供立即预览——此刻内容还没保存，反查不到引用。
const MEMBER_UPLOAD_DIR = 'member'

/** 已签发的短时效地址缓存（按文件名）；签名 5 分钟过期，超过 4 分钟就重新签发 */
const signedFileCache = new Map<string, { url: string; at: number }>()
const SIGNED_TTL_MS = 4 * 60 * 1000

const memberFileNameOf = (url: string) => url.substring(url.lastIndexOf('/') + 1).split('?')[0]

/** 取（必要时重新签发）某私有文件的短时效访问地址；非私有地址原样返回。force=true 时忽略缓存强制重签 */
const ensureSignedFile = async (url?: string, force = false) => {
  if (!url) return ''
  if (!url.includes('/member-files/')) return url
  const name = memberFileNameOf(url)
  const cached = signedFileCache.get(name)
  if (!force && cached && Date.now() - cached.at < SIGNED_TTL_MS) return cached.url
  try {
    const res: any = await signMemberFile(name)
    const signed = res.data?.url || ''
    if (signed) signedFileCache.set(name, { url: signed, at: Date.now() })
    return signed
  } catch {
    // 「无权访问」/已过期等由 request 拦截器统一提示
    return ''
  }
}

// 表单里展示用的地址（私有文件必须先换成签名地址，不能直接用入库的干净地址打开）
const coverDisplay = ref('')
const attachmentDisplay = ref('')
const paperDisplay = ref('')

/** 上传后直接记下当场签发的地址，避免再走一次签发接口 */
const rememberSigned = (cleanUrl: string, signedUrl?: string) => {
  if (cleanUrl && signedUrl) {
    signedFileCache.set(memberFileNameOf(cleanUrl), { url: signedUrl, at: Date.now() })
  }
}

// 正文 HTML 里内联的会员专区图片同样需要签名地址才能显示：
//   - 载入（编辑）时逐个签发并注入；提交前剔除签名参数；
//   - **数据库里存的正文地址不带任何签名参数**（否则签名过期后整篇正文裂图）。
const MEMBER_FILE_URL_RE = /(\/member-files\/[A-Za-z0-9._-]+)(\?[^"'\s>]*)?/g
const stripFileSignParams = (html: string) =>
  (html || '').replace(MEMBER_FILE_URL_RE, (_all, url: string) => url)

const injectSignedFileUrls = async (html: string) => {
  if (!html || !html.includes('/member-files/')) return html || ''
  const names = Array.from(
    new Set(Array.from(html.matchAll(/\/member-files\/([A-Za-z0-9._-]+)/g)).map((m) => m[1]))
  )
  // 并发签发：原先逐张 await，一篇含 20~30 张内联图的正文要串行等 20~30 个往返才弹出弹窗
  const entries = await Promise.all(
    names.map(async (name): Promise<[string, string]> => [name, await ensureSignedFile(`/member-files/${name}`)])
  )
  const signed = new Map(entries.filter(([, url]) => !!url))
  return html.replace(
    MEMBER_FILE_URL_RE,
    (all, path: string) => signed.get(memberFileNameOf(path)) || all
  )
}

// 编辑弹窗打开期间定期重签正文内联图片：
// 前端签名缓存 4 分钟、后端签名 5 分钟，用户编辑较久时图片会突然裂图且无法自救。
// 只改 DOM 的 img.src（**不动编辑器的数据模型**）：提交时本就会 stripFileSignParams 剔除
// 签名参数，因此入库内容不受影响，也不会打断光标/输入。
const RESIGN_REFRESH_MS = 3 * 60 * 1000
let resignTimer: ReturnType<typeof setInterval> | null = null

const refreshInlineSignedImages = async () => {
  try {
    const imgs = Array.from(
      document.querySelectorAll<HTMLImageElement>(
        '.editor-wrapper .w-e-text-container img[src*="/member-files/"]'
      )
    )
    await Promise.all(
      imgs.map(async (img) => {
        const url = await ensureSignedFile(img.src, true)
        if (url) img.src = url
      })
    )
  } catch {
    // 刷新预览失败不影响编辑
  }
}

const stopResignTimer = () => {
  if (resignTimer) {
    clearInterval(resignTimer)
    resignTimer = null
  }
}

const startResignTimer = () => {
  stopResignTimer()
  resignTimer = setInterval(() => void refreshInlineSignedImages(), RESIGN_REFRESH_MS)
}

const handleCoverUpload = async (options: any) => {
  const res: any = await uploadFile(options.file, MEMBER_UPLOAD_DIR)
  contentForm.cover = res.data?.url || ''
  rememberSigned(contentForm.cover, res.data?.signedUrl)
  coverDisplay.value = res.data?.signedUrl || ''
  ElMessage.success('封面图上传成功')
}

const handleAttachmentUpload = async (options: any) => {
  const res: any = await uploadFile(options.file, MEMBER_UPLOAD_DIR)
  contentForm.attachmentUrl = res.data?.url || ''
  contentForm.attachmentName = res.data?.name || options.file?.name || ''
  rememberSigned(contentForm.attachmentUrl, res.data?.signedUrl)
  attachmentDisplay.value = res.data?.signedUrl || ''
  ElMessage.success('文章附件上传成功')
}

const handleFullVideoUpload = async (options: any) => {
  const res: any = await uploadFile(options.file, MEMBER_UPLOAD_DIR)
  contentForm.fullVideoUrl = res.data?.url || ''
  rememberSigned(contentForm.fullVideoUrl, res.data?.signedUrl)
  ElMessage.success('完整视频上传成功')
}

const handlePreviewVideoUpload = async (options: any) => {
  const res: any = await uploadFile(options.file, MEMBER_UPLOAD_DIR)
  contentForm.previewVideoUrl = res.data?.url || ''
  rememberSigned(contentForm.previewVideoUrl, res.data?.signedUrl)
  ElMessage.success('预览视频上传成功')
}

const handlePaperFileUpload = async (options: any) => {
  const res: any = await uploadFile(options.file, MEMBER_UPLOAD_DIR)
  contentForm.paperFileUrl = res.data?.url || ''
  contentForm.paperFileName = res.data?.name || options.file?.name || ''
  rememberSigned(contentForm.paperFileUrl, res.data?.signedUrl)
  paperDisplay.value = res.data?.signedUrl || ''
  ElMessage.success('报刊文件上传成功')
}

const clearCover = () => {
  contentForm.cover = ''
  coverDisplay.value = ''
}

const clearAttachment = () => {
  contentForm.attachmentUrl = ''
  contentForm.attachmentName = ''
  attachmentDisplay.value = ''
}

const clearPaperFile = () => {
  contentForm.paperFileUrl = ''
  contentForm.paperFileName = ''
  paperDisplay.value = ''
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
  const res: any = await uploadFile(file, MEMBER_UPLOAD_DIR)
  const url = res.data?.url || ''
  if (url) {
    // 编辑器内 <img> 必须用签名地址才能显示；提交前由 stripFileSignParams() 剔除，库里仍是干净地址
    insertFn(res.data?.signedUrl || url, '', '')
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
      },
    },
  },
}

const handleEditorCreated = (editor: IDomEditor) => {
  editorRef.value = editor
}

onMounted(() => {
  fetchColumnOptions()
  fetchContents()
})

onBeforeUnmount(() => {
  stopResignTimer()
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

.media-row {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
}

.media-row .el-input {
  flex: 1;
  min-width: 0;
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

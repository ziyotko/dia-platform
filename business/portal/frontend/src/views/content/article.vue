<template>
  <div class="page-container" v-loading="deleting" :element-loading-text="deletingText">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="文章标题">
          <el-input v-model="queryForm.title" placeholder="请输入文章标题" clearable />
        </el-form-item>
        <el-form-item label="发布栏目">
          <el-cascader
            v-model="queryForm.columnPath"
            :options="columnCascaderOptions"
            :props="{ expandTrigger: 'hover' }"
            placeholder="请选择页面/栏目/子栏目"
            clearable
            filterable
            style="width: 280px"
          />
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
        <el-form-item label="文章标签">
          <el-select v-model="queryForm.tagId" placeholder="全部标签" clearable style="width: 140px">
            <el-option
              v-for="item in tagList"
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
        <el-form-item label="文章类型">
          <el-select v-model="queryForm.type" placeholder="全部类型" clearable style="width: 120px">
            <el-option label="图文" :value="1" />
            <el-option label="视频" :value="2" />
            <el-option label="数据" :value="3" />
            <el-option label="报刊" :value="4" />
          </el-select>
        </el-form-item>
        <el-form-item label="作者">
          <el-input v-model="queryForm.author" placeholder="请输入作者" clearable />
        </el-form-item>
        <el-form-item label="来源">
          <el-input v-model="queryForm.source" placeholder="请输入来源" clearable />
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
          <div class="header-actions">
            <el-button type="primary" @click="handleAdd">
              <el-icon><Plus /></el-icon>新增文章
            </el-button>
            <el-button type="primary" @click="handleAddVideo">
              <el-icon><VideoCamera /></el-icon>新增视频
            </el-button>
            <el-button type="primary" @click="handleAddData">
              <el-icon><DataLine /></el-icon>新增数据
            </el-button>
            <el-button type="primary" @click="handleAddPaper">
              <el-icon><Files /></el-icon>新增报刊
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column label="文章标题" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="title-with-type">
              <el-tag v-if="row.type === 2" type="danger" effect="light" size="small" class="type-tag">
                <el-icon><VideoCamera /></el-icon>视频
              </el-tag>
              <el-tag v-else-if="row.type === 3" type="warning" effect="light" size="small" class="type-tag">
                <el-icon><DataLine /></el-icon>数据
              </el-tag>
              <el-tag v-else-if="row.type === 4" type="success" effect="light" size="small" class="type-tag">
                <el-icon><Files /></el-icon>报刊
              </el-tag>
              <el-tag v-else type="primary" effect="light" size="small" class="type-tag">
                <el-icon><Document /></el-icon>图文
              </el-tag>
              <el-link
                v-if="row.status === 1"
                type="primary"
                :underline="'never'"
                @click="handlePreview(row)"
              >
                {{ (row.type === 3 || row.type === 4) && row.summary ? `${row.summary}${row.title}` : row.title }}
              </el-link>
              <span v-else>{{ (row.type === 3 || row.type === 4) && row.summary ? `${row.summary}${row.title}` : row.title }}</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="author" label="作者" width="100" />
        <el-table-column prop="source" label="来源" width="140" show-overflow-tooltip />
        <el-table-column prop="publishTime" label="发布时间" width="120" align="center">
          <template #default="{ row }">
            {{ row.publishTime ? (row.type === 4 ? row.publishTime.slice(0, 7) : row.publishTime.slice(0, 10)) : '-' }}
          </template>
        </el-table-column>
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
      
        <el-table-column prop="createdAt" label="创建时间" width="170" />
        <el-table-column label="操作" width="300" align="center" fixed="right">
          <template #default="{ row }">
            <el-button v-if="isAuthor(row) && row.auditStatus !== 1 && (row.auditStatus !== 2 || row.status === 2)" link type="primary" @click="handleEdit(row)">
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
            <el-button v-if="(isAuthor(row) || isAdmin) && row.status === 1" link type="danger" @click="handleOffShelf(row)">
              <el-icon><CircleClose /></el-icon>下线
            </el-button>
            <el-button v-if="isAuthor(row) || isAdmin" link type="danger" @click="handleDelete(row)">
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
              <el-form-item label="所属分类" prop="categoryIds">
                <el-select v-model="form.categoryIds" multiple placeholder="请选择分类" style="width: 100%">
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
          </el-row>
          <el-row :gutter="20">
            <el-col :xs="24" :sm="12">
              <el-form-item label="文章来源" prop="source">
                <el-input v-model="form.source" placeholder="请输入文章来源" clearable />
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12">
              <el-form-item label="发布时间" prop="publishTime">
                <el-date-picker
                  v-model="form.publishTime"
                  type="datetime"
                  placeholder="请选择发布时间"
                  format="YYYY-MM-DD HH:mm"
                  value-format="YYYY-MM-DD HH:mm:00"
                  style="width: 100%"
                />
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
          <div class="section-title">跳转链接</div>
          <el-form-item label="URL地址" prop="url">
            <el-input
              v-model="form.url"
              placeholder="请输入跳转链接，如 https://example.com"
              clearable
              @blur="formRef?.validateField(['url', 'content']).catch(() => {})"
              @input="formRef?.validateField(['url', 'content']).catch(() => {})"
            />
            <div class="url-tip">填写 URL 后，文章内容和附件无需填写，保存后将直接跳转至该链接</div>
          </el-form-item>
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
              :on-change="handleAttachmentChange"
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

    <el-dialog v-model="videoDialogVisible" :title="videoDialogTitle" width="680px" destroy-on-close :close-on-click-modal="false">
      <el-form ref="videoFormRef" :model="videoForm" :rules="videoFormRules" label-width="90px">
        <el-form-item label="视频标题" prop="title">
          <el-input v-model="videoForm.title" placeholder="请输入视频标题" clearable />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12">
            <el-form-item label="所属分类" prop="categoryIds">
              <el-select v-model="videoForm.categoryIds" multiple placeholder="请选择分类" style="width: 100%">
                <el-option
                  v-for="item in categoryList"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="文章标签" prop="tagIds">
              <el-select v-model="videoForm.tagIds" multiple placeholder="请选择标签" style="width: 100%">
                <el-option
                  v-for="item in tagList"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12">
            <el-form-item label="发布时间" prop="publishTime">
              <el-date-picker
                v-model="videoForm.publishTime"
                type="datetime"
                placeholder="请选择发布时间"
                format="YYYY-MM-DD HH:mm"
                value-format="YYYY-MM-DD HH:mm:00"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="来源" prop="source">
              <el-input v-model="videoForm.source" placeholder="请输入来源" clearable />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="封面图" prop="cover">
          <div class="article-cover-uploader">
            <el-upload
              v-if="!videoForm.cover"
              class="cover-uploader"
              action=""
              :http-request="handleVideoCoverUpload"
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
                  :src="videoForm.cover"
                  fit="cover"
                  style="width: 100%; height: 100%"
                  :preview-src-list="[videoForm.cover]"
                />
                <div class="cover-overlay" @click="handleRemoveVideoCover">
                  <el-icon><Delete /></el-icon>
                  <span>删除封面</span>
                </div>
              </div>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="上传视频" prop="videoUrl">
          <el-upload
            action="#"
            :http-request="handleVideoUpload"
            :before-upload="handleVideoBeforeUpload"
            :on-remove="handleVideoRemove"
            :limit="1"
            accept="video/*"
            class="video-uploader"
          >
            <el-button type="primary" plain :disabled="!!videoForm.videoUrl">
              <el-icon><Plus /></el-icon>上传视频
            </el-button>
            <template #tip>
              <div class="video-tip">支持 MP4、MOV 等常见视频格式，单个文件不超过 800MB</div>
              <el-progress
                v-if="videoUploadProgress > 0 && !videoForm.videoUrl"
                :percentage="videoUploadProgress"
                :stroke-width="6"
                style="margin-top: 6px"
              />
            </template>
          </el-upload>
          <div v-if="videoForm.videoUrl" class="video-preview">
            <video :src="videoForm.videoUrl" controls style="max-width: 100%; max-height: 240px; border-radius: 8px" />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="videoDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="videoSubmitLoading" @click="handleSubmitVideo">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dataDialogVisible" :title="dataDialogTitle" width="680px" destroy-on-close :close-on-click-modal="false">
      <el-form ref="dataFormRef" :model="dataForm" :rules="dataFormRules" label-width="90px">
        <el-form-item label="数据标题" prop="title">
          <el-input v-model="dataForm.title" placeholder="请输入数据标题" clearable />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12">
            <el-form-item label="所属分类" prop="categoryIds">
              <el-select v-model="dataForm.categoryIds" multiple placeholder="请选择分类" style="width: 100%">
                <el-option
                  v-for="item in categoryList"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="文章标签" prop="tagIds">
              <el-select v-model="dataForm.tagIds" multiple placeholder="请选择标签" style="width: 100%">
                <el-option
                  v-for="item in tagList"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12">
            <el-form-item label="发布时间" prop="publishTime">
              <el-date-picker
                v-model="dataForm.publishTime"
                type="datetime"
                placeholder="请选择发布时间"
                format="YYYY-MM-DD HH:mm"
                value-format="YYYY-MM-DD HH:mm:00"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="来源" prop="source">
              <el-input v-model="dataForm.source" placeholder="请输入来源" clearable />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="选择年月" prop="yearMonth">
          <el-date-picker
            v-model="dataForm.yearMonth"
            type="month"
            placeholder="请选择年月"
            format="YYYY年MM月"
            value-format="YYYY-MM"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="数据内容" prop="content">
          <el-input
            v-model="dataForm.content"
            type="textarea"
            :rows="8"
            placeholder="请输入数据内容"
            maxlength="80000"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dataDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="dataSubmitLoading" @click="handleSubmitData">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="paperDialogVisible" :title="paperDialogTitle" width="680px" destroy-on-close :close-on-click-modal="false">
      <el-form ref="paperFormRef" :model="paperForm" :rules="paperFormRules" label-width="90px">
        <el-form-item label="报刊标题" prop="title">
          <el-input v-model="paperForm.title" placeholder="请输入报刊标题" clearable />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12">
            <el-form-item label="所属分类" prop="categoryIds">
              <el-select v-model="paperForm.categoryIds" multiple placeholder="请选择分类" style="width: 100%">
                <el-option
                  v-for="item in categoryList"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="文章标签" prop="tagIds">
              <el-select v-model="paperForm.tagIds" multiple placeholder="请选择标签" style="width: 100%">
                <el-option
                  v-for="item in tagList"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12">
            <el-form-item label="期号" prop="issueNo">
              <el-input v-model="paperForm.issueNo" placeholder="如 2026年第9期" clearable />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="出版年月" prop="yearMonth">
              <el-date-picker
                v-model="paperForm.yearMonth"
                type="month"
                placeholder="请选择出版年月"
                format="YYYY年MM月"
                value-format="YYYY-MM"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="来源" prop="source">
          <el-input v-model="paperForm.source" placeholder="请输入来源" clearable />
        </el-form-item>
        <el-form-item label="封面图" prop="cover">
          <div class="article-cover-uploader">
            <el-upload
              v-if="!paperForm.cover"
              class="cover-uploader"
              action=""
              :http-request="handlePaperCoverUpload"
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
                  :src="paperForm.cover"
                  fit="cover"
                  style="width: 100%; height: 100%"
                  :preview-src-list="[paperForm.cover]"
                />
                <div class="cover-overlay" @click="handleRemovePaperCover">
                  <el-icon><Delete /></el-icon>
                  <span>删除封面</span>
                </div>
              </div>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="报刊摘要" prop="abstract">
          <el-input
            v-model="paperForm.abstract"
            type="textarea"
            :rows="4"
            placeholder="请输入报刊摘要，简要描述本期报刊内容"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="报刊文件" prop="pdf">
          <el-upload
            action="#"
            :http-request="handlePaperAttachmentUpload"
            :before-upload="handlePaperAttachmentBeforeUpload"
            :on-remove="handlePaperAttachmentRemove"
            :limit="1"
            accept=".pdf,application/pdf"
            class="paper-uploader"
          >
            <el-button type="primary" plain :disabled="!!paperForm.pdf">
              <el-icon><Plus /></el-icon>上传 PDF 报刊文件
            </el-button>
            <template #tip>
              <div class="paper-tip">支持 PDF 格式报刊文件，单个文件不超过 50MB</div>
            </template>
          </el-upload>
          <div v-if="paperForm.pdf && paperForm.pdf.url" class="paper-file-preview">
            <el-icon><Document /></el-icon>
            <a :href="paperForm.pdf.url" target="_blank">{{ paperForm.pdf.name }}</a>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="paperDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="paperSubmitLoading" @click="handleSubmitPaper">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="previewVisible" title="文章预览" width="800px" destroy-on-close>
      <div class="preview-content">
        <h2>{{ previewData.title }}</h2>
        <div class="preview-meta">
          <span>作者：{{ previewData.author }}</span>
          <span>来源：{{ previewData.source }}</span>
          <span>时间：{{ previewData.createdAt }}</span>
        </div>
        <div class="preview-cover" v-if="previewData.cover">
          <el-image
            :src="previewData.cover"
            fit="cover"
            style="width: 100%; max-height: 360px"
            :preview-src-list="[previewData.cover]"
            class="preview-cover-image"
          />
        </div>
        <div class="preview-summary" v-if="previewData.summary">
          <strong>摘要：</strong>{{ previewData.summary }}
        </div>
        <div class="preview-url" v-if="previewData.url">
          <strong>跳转链接：</strong><a :href="previewData.url" target="_blank">{{ previewData.url }}</a>
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

    <el-dialog
      v-model="columnDialogVisible"
      :title="`栏目设置 - ${columnDialogTitle}`"
      width="980px"
      destroy-on-close
      class="column-setting-dialog"
      :close-on-click-modal="false"
    >
      <div class="column-setting-hint">
        <el-icon><InfoFilled /></el-icon>
        <span>
          当前文章类型：<el-tag size="small" type="primary" effect="light">{{ articleTypeName }}</el-tag>；
          <template v-if="currentArticleType === 2">仅展示「视频展示」类栏目。</template>
          <template v-else-if="currentArticleType === 3">仅展示「数据展示」类栏目。</template>
          <template v-else-if="currentArticleType === 4">仅展示「报刊展示」类栏目。</template>
          <template v-else>不展示「视频展示」「数据展示」和「报刊展示」类栏目。</template>
          请先选择页面，再勾选该页面下的栏目；支持跨页面多选。
        </span>
      </div>
      <el-row :gutter="16" class="column-setting-body">
        <el-col :span="8">
          <div class="column-setting-panel">
            <div class="column-setting-panel-title">
              <el-icon><Monitor /></el-icon>
              <span>选择页面</span>
              <span class="column-setting-count">({{ pageList.length }})</span>
            </div>
            <el-scrollbar height="420px" class="column-setting-scroll">
              <div v-if="pageList.length === 0" class="column-setting-empty">
                <el-empty description="暂无页面" :image-size="80" />
              </div>
              <div
                v-for="page in pageList"
                :key="page.id"
                :class="['page-item', { active: selectedColumnPageId === page.id }]"
                @click="selectColumnPage(page.id)"
              >
                <div class="page-item-name">{{ page.name }}</div>
                <div class="page-item-meta">{{ getColumnCountByPage(page.id) }} 个栏目</div>
                <el-icon v-if="selectedColumnPageId === page.id" class="page-item-check"><Check /></el-icon>
              </div>
            </el-scrollbar>
          </div>
        </el-col>
        <el-col :span="16">
          <div class="column-setting-panel">
            <div class="column-setting-panel-title">
              <el-icon><Collection /></el-icon>
              <span>选择栏目</span>
              <span class="column-setting-count">({{ selectedColumnPageColumns.length }})</span>
              <el-checkbox
                v-if="selectedColumnPageColumns.length > 0"
                v-model="selectedPageAllSelected"
                class="column-select-all"
                @change="toggleSelectAllPageColumns"
              >全选</el-checkbox>
            </div>
            <el-scrollbar height="420px" class="column-setting-scroll">
              <div v-if="!selectedColumnPageId" class="column-setting-empty">
                <el-empty description="请先选择左侧页面" :image-size="100" />
              </div>
              <div v-else-if="selectedColumnPageColumns.length === 0" class="column-setting-empty">
                <el-empty description="该页面下暂无符合当前文章类型的栏目" :image-size="100" />
              </div>
              <el-checkbox-group v-else v-model="selectedColumnIds" class="column-checkbox-group">
                <div
                  v-for="col in selectedColumnPageColumns"
                  :key="col.id"
                  :class="['column-card', { checked: selectedColumnIds.includes(col.id), 'is-child': col.parentId && col.parentId > 0 }]"
                  @click="toggleColumnSelection(col.id)"
                >
                  <el-checkbox :label="col.id" @click.stop>
                    <div class="column-card-info">
                      <span class="column-card-name">
                        <el-icon v-if="col.parentId && col.parentId > 0" class="child-column-icon"><ArrowRight /></el-icon>
                        {{ col.name }}
                      </span>
                      <div class="column-card-tags">
                        <el-tag v-if="col.parentId && col.parentId > 0" size="small" type="warning" effect="light" class="column-card-tag">子栏目</el-tag>
                        <el-tag v-if="col.workflowId" size="small" type="success" effect="light" class="column-card-tag">需审核</el-tag>
                        <el-tag v-else size="small" type="info" effect="light" class="column-card-tag">免审核</el-tag>
                      </div>
                    </div>
                    <div v-if="col.parentId && col.parentId > 0" class="column-card-parent">
                      上级：{{ getColumnParentName(col) }}
                    </div>
                  </el-checkbox>
                </div>
              </el-checkbox-group>
            </el-scrollbar>
          </div>
        </el-col>
      </el-row>
      <div class="column-setting-selected">
        <div class="column-setting-selected-title">
          <span>已选栏目</span>
          <span class="column-setting-count">{{ selectedColumnIds.length }} 个</span>
        </div>
        <div v-if="selectedColumnIds.length === 0" class="column-setting-selected-empty">暂未选择任何栏目</div>
        <div v-else class="column-setting-selected-list">
          <el-tag
            v-for="item in selectedColumnSummary"
            :key="item.id"
            closable
            type="primary"
            effect="light"
            class="column-selected-tag"
            @close="removeSelectedColumn(item.id)"
          >
            {{ item.pageName }} / {{ item.name }}
          </el-tag>
        </div>
      </div>
      <template #footer>
        <el-button @click="columnDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="columnSubmitLoading" @click="handleSubmitColumns">确定</el-button>
      </template>
    </el-dialog>

    <!-- 审核流程预览 -->
    <el-dialog
      v-model="auditFlowDialogVisible"
      title="栏目审核流程"
      width="640px"
      destroy-on-close
      :close-on-click-modal="false"
      :close-on-press-escape="!auditFlowSubmitting"
      :show-close="!auditFlowSubmitting"
    >
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
                      <span>{{ getNodeHistory(item, node.id).operatorName }} {{ formatAuditTime(getNodeHistory(item, node.id).createdAt) }}</span>
                    </div>
                  </template>
                </el-step>
              </el-steps>
            </div>
            <el-empty v-else description="该流程未配置节点" :image-size="60" />
            <div v-if="item.auditStatus === 0 && item.workflow && item.workflow.nodes && item.currentNodeId" class="audit-current-node">
              <el-icon><User /></el-icon>
              <span>当前节点审批人：{{ formatApprover(item) }}</span>
              <span v-if="item.currentApproverName" class="audit-debug-name">（调试：{{ item.currentApproverName }}）</span>
            </div>
            <div v-if="item.auditStatus === 1 && item.approveUserName" class="audit-flow-result">
              <el-icon color="#67c23a"><CircleCheck /></el-icon>
              <span>已通过：{{ item.approveRemark || '—' }}</span>
            </div>
            <div v-if="item.auditStatus === 2 && item.rejectRemark" class="audit-flow-result audit-flow-reject-result">
              <el-icon color="#f56c6c"><CircleClose /></el-icon>
              <span>已驳回：{{ item.rejectRemark }}</span>
            </div>
            <div v-if="item.auditStatus === 0 && item.workflow && item.workflow.nodes && item.workflow.nodes.length > 0">
              <div v-if="item.canApprove" class="audit-flow-actions">
                <el-button
                  type="primary"
                  size="small"
                  :disabled="auditFlowSubmitting"
                  :loading="isAuditActionLoading(item.columnId, 'advance')"
                  @click="handleAdvanceAuditNode(item.columnId)"
                >
                  <el-icon><CircleCheck /></el-icon>通过当前节点
                </el-button>
                <el-button
                  type="danger"
                  size="small"
                  plain
                  :disabled="auditFlowSubmitting"
                  :loading="isAuditActionLoading(item.columnId, 'reject')"
                  @click="handleRejectAuditNode(item.columnId)"
                >
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
        <el-button :disabled="auditFlowSubmitting" @click="auditFlowDialogVisible = false">关闭</el-button>
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
  CircleCheck,
  CircleClose,
  FolderOpened,
  Warning,
  Document,
  VideoCamera,
  DataLine,
  InfoFilled,
  Monitor,
  Collection,
  Check,
  ArrowRight,
  User,
  Files
} from '@element-plus/icons-vue'
import '@wangeditor/editor/dist/css/style.css'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import type { IDomEditor, IEditorConfig, IToolbarConfig } from '@wangeditor/editor'

import { useUserStore } from '@/stores/user'
import request from '@/utils/request'
import { hasAdminRole } from '@/utils/permission'
import {
  getArticles,
  getArticleByID,
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
import { getUserOptions } from '@/api/user'
import { getWorkflowRoleOptions } from '@/api/workflow-role'
import { getAllCategories } from '@/api/category'
import { getAllTags } from '@/api/tag'
import { getPages } from '@/api/page'
import { getColumns } from '@/api/column'
import { uploadFile } from '@/api/upload'

const userStore = useUserStore()
const currentUserId = computed(() => userStore.userInfo?.id || 0)
// 管理员判定（角色 1），与后端 models.HasAdminRoleIDs 保持一致
const isAdmin = computed(() => hasAdminRole(userStore.userInfo?.roleIds))

const loading = ref(false)
const deleting = ref(false)
const deletingText = ref('文章删除中...')
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const total = ref(0)
const formRef = ref()

const videoDialogVisible = ref(false)
const videoSubmitLoading = ref(false)
const videoFormRef = ref()
const videoForm = reactive({
  id: undefined as number | undefined,
  categoryIds: [] as number[],
  tagIds: [] as number[],
  title: '',
  source: '',
  publishTime: '',
  cover: '',
  videoUrl: '',
  videoName: '',
  videoSize: 0
})

const videoDialogTitle = ref('新增视频')
// 视频上传进度（0-100），仅在上传过程中展示
const videoUploadProgress = ref(0)

const dataDialogVisible = ref(false)
const dataSubmitLoading = ref(false)
const dataFormRef = ref()
const dataForm = reactive({
  id: undefined as number | undefined,
  categoryIds: [] as number[],
  tagIds: [] as number[],
  title: '',
  source: '',
  publishTime: '',
  yearMonth: '',
  content: ''
})

const dataDialogTitle = ref('新增数据')

const paperDialogVisible = ref(false)
const paperSubmitLoading = ref(false)
const paperFormRef = ref()
const paperForm = reactive({
  id: undefined as number | undefined,
  title: '',
  issueNo: '',
  yearMonth: '',
  source: '',
  cover: '',
  abstract: '',
  categoryIds: [] as number[],
  tagIds: [] as number[],
  pdf: null as any
})

const paperDialogTitle = ref('新增报刊')

const columnDialogVisible = ref(false)
const columnDialogTitle = ref('')
const selectedColumnIds = ref<number[]>([])
const selectedColumnPageId = ref<number | undefined>(undefined)
const columnSubmitLoading = ref(false)
const currentArticleId = ref<number | undefined>(undefined)
const currentArticleType = ref<number | undefined>(undefined)
const pageList = ref<any[]>([])
const columnList = ref<any[]>([])

const articleTypeName = computed(() => {
  const map: Record<number, string> = { 1: '图文', 2: '视频', 3: '数据', 4: '报刊' }
  return map[currentArticleType.value || 1] || '图文'
})

const selectedColumnPageColumns = computed(() => {
  if (!selectedColumnPageId.value) return []
  let cols = columnList.value.filter((col: any) => col.pageId === selectedColumnPageId.value)
  const type = currentArticleType.value
  if (type === 2) {
    cols = cols.filter((col: any) => col.displayType === 8)
  } else if (type === 3) {
    cols = cols.filter((col: any) => col.displayType === 7)
  } else if (type === 4) {
    cols = cols.filter((col: any) => col.displayType === 6)
  } else {
    cols = cols.filter((col: any) => col.displayType !== 6 && col.displayType !== 7 && col.displayType !== 8)
  }
  return cols.sort((a: any, b: any) => {
    const aIsRoot = !a.parentId || a.parentId === 0
    const bIsRoot = !b.parentId || b.parentId === 0
    if (aIsRoot && !bIsRoot) return -1
    if (!aIsRoot && bIsRoot) return 1
    return a.sort - b.sort
  })
})

const getColumnParentName = (column: any) => {
  if (!column.parentId) return ''
  const parent = columnList.value.find((col: any) => col.id === column.parentId)
  return parent?.name || ''
}

const selectedPageAllSelected = computed({
  get() {
    const cols = selectedColumnPageColumns.value
    return cols.length > 0 && cols.every((col: any) => selectedColumnIds.value.includes(col.id))
  },
  set(val: boolean) {
    const ids = selectedColumnPageColumns.value.map((col: any) => col.id)
    if (val) {
      selectedColumnIds.value = Array.from(new Set([...selectedColumnIds.value, ...ids]))
    } else {
      selectedColumnIds.value = selectedColumnIds.value.filter((id: number) => !ids.includes(id))
    }
  }
})

const selectedColumnSummary = computed(() => {
  return selectedColumnIds.value
    .map((id: number) => {
      const col = columnList.value.find((c: any) => c.id === id)
      if (!col) return null
      const page = pageList.value.find((p: any) => p.id === col.pageId)
      return { id, name: col.name, pageName: page?.name || '未知页面' }
    })
    .filter(Boolean) as { id: number; name: string; pageName: string }[]
})

const auditFlowDialogVisible = ref(false)
const auditFlowLoading = ref(false)
const auditFlowList = ref<any[]>([])
const auditFlowArticleTitle = ref('')
const auditFlowUserList = ref<any[]>([])
const auditFlowRoleList = ref<any[]>([])
// 审核节点操作进行中：用于禁用弹窗内其他操作，避免重复提交/并发提交
const auditFlowSubmitting = ref(false)
// 当前进行中的操作标识：`${columnId}:advance` 或 `${columnId}:reject`
const auditFlowSubmittingKey = ref('')

const isAuditActionLoading = (columnId: number, action: 'advance' | 'reject') =>
  auditFlowSubmittingKey.value === `${columnId}:${action}`

const queryForm = reactive({
  page: 1,
  pageSize: 10,
  title: '',
  categoryId: undefined as number | undefined,
  tagId: undefined as number | undefined,
  status: undefined as number | undefined,
  auditStatus: undefined as number | undefined,
  type: undefined as number | undefined,
  author: '',
  source: '',
  columnPath: [] as number[]
})

const form = reactive({
  id: undefined as number | undefined,
  title: '',
  type: 1,
  categoryIds: [] as number[],
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
  publishTime: '',
  url: '',
  attachments: [] as any[]
})

const isValidUrl = (url: string) => /^https?:\/\/.+/i.test(url)
const hasRealContent = (html: string) => {
  if (!html) return false
  const text = html.replace(/<[^>]+>/g, '').replace(/&nbsp;/g, '').trim()
  return text.length > 0
}

const formRules = {
  title: [{ required: true, message: '请输入文章标题', trigger: 'blur' }],
  publishTime: [{ required: true, message: '请选择发布时间', trigger: 'change' }],
  url: [
    {
      validator: (_rule: any, value: any, callback: any) => {
        const hasContent = hasRealContent(form.content)
        const hasAttachments = form.attachments && form.attachments.length > 0
        const hasUrl = value && value.trim().length > 0
        if (!hasUrl && !hasContent && !hasAttachments) {
          callback(new Error('URL、文章内容、文章附件至少填写一项'))
        } else if (hasUrl && !isValidUrl(value.trim())) {
          callback(new Error('请输入以 http:// 或 https:// 开头的URL地址'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  content: [
    {
      validator: (_rule: any, _value: any, callback: any) => {
        const hasUrl = form.url && form.url.trim().length > 0
        const hasAttachments = form.attachments && form.attachments.length > 0
        const hasContent = hasRealContent(form.content)
        if (!hasUrl && !hasContent && !hasAttachments) {
          callback(new Error('URL、文章内容、文章附件至少填写一项'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ]
}

const videoFormRules = {
  title: [{ required: true, message: '请输入视频标题', trigger: 'blur' }],
  publishTime: [{ required: true, message: '请选择发布时间', trigger: 'change' }],
  videoUrl: [{ required: true, message: '请上传视频', trigger: 'change' }]
}

const dataFormRules = {
  title: [{ required: true, message: '请输入数据标题', trigger: 'blur' }],
  publishTime: [{ required: true, message: '请选择发布时间', trigger: 'change' }],
  yearMonth: [{ required: true, message: '请选择年月', trigger: 'change' }],
  content: [{ required: true, message: '请输入数据内容', trigger: 'blur' }]
}

const paperFormRules = {
  title: [{ required: true, message: '请输入报刊标题', trigger: 'blur' }],
  yearMonth: [{ required: true, message: '请选择出版年月', trigger: 'change' }]
}

const route = useRoute()
const router = useRouter()

const tableData = ref<any[]>([])
const categoryList = ref<any[]>([])
const tagList = ref<any[]>([])

const getColumnsByPage = (pageId: number) => {
  return columnList.value.filter((col: any) => col.pageId === pageId)
}

const getColumnCountByPage = (pageId: number) => {
  const cols = getColumnsByPage(pageId)
  const type = currentArticleType.value
  if (type === 2) {
    return cols.filter((col: any) => col.displayType === 8).length
  } else if (type === 3) {
    return cols.filter((col: any) => col.displayType === 7).length
  } else if (type === 4) {
    return cols.filter((col: any) => col.displayType === 6).length
  } else {
    return cols.filter((col: any) => col.displayType !== 6 && col.displayType !== 7 && col.displayType !== 8).length
  }
}

const selectColumnPage = (pageId: number) => {
  selectedColumnPageId.value = pageId
}

const toggleColumnSelection = (columnId: number) => {
  const index = selectedColumnIds.value.indexOf(columnId)
  if (index > -1) {
    selectedColumnIds.value.splice(index, 1)
  } else {
    selectedColumnIds.value.push(columnId)
  }
}

const toggleSelectAllPageColumns = (val: any) => {
  const ids = selectedColumnPageColumns.value.map((col: any) => col.id)
  if (val) {
    selectedColumnIds.value = Array.from(new Set([...selectedColumnIds.value, ...ids]))
  } else {
    selectedColumnIds.value = selectedColumnIds.value.filter((id: number) => !ids.includes(id))
  }
}

const removeSelectedColumn = (columnId: number) => {
  selectedColumnIds.value = selectedColumnIds.value.filter((id: number) => id !== columnId)
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

// 栏目级联选项：页面 → 栏目 → 子栏目
const columnCascaderOptions = computed(() => {
  return pageList.value.map((page: any) => {
    const cols = columnList.value.filter((col: any) => col.pageId === page.id)
    const rootCols = cols.filter((col: any) => !col.parentId || col.parentId === 0)
    return {
      value: page.id,
      label: page.name,
      children: rootCols.map((rootCol: any) => {
        const children = cols.filter((col: any) => col.parentId === rootCol.id)
        return {
          value: rootCol.id,
          label: rootCol.name,
          children: children.length
            ? children.map((child: any) => ({ value: child.id, label: child.name }))
            : undefined
        }
      })
    }
  })
})

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

const IMAGE_MAX_SIZE = 10 * 1024 * 1024

const uploadImageFile = async (
  file: File,
  insertFn: (url: string, alt: string, href: string) => void
) => {
  if (file.size > IMAGE_MAX_SIZE) {
    ElMessage.error(`图片大小不能超过 10MB`)
    throw new Error('图片大小超出限制')
  }
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

const editorConfig: Partial<IEditorConfig> = {
  placeholder: '',
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

const handleCreated = (editor: IDomEditor) => {
  editorRef.value = editor
}

const handleEditorChange = () => {
  if (formRef.value) {
    formRef.value.validateField('content').catch(() => {})
  }
}

// 为文章内容中的外链图片添加 referrerpolicy="no-referrer"，绕过图床防盗链（如腾讯 gtimg.com 等）
const noReferrerContent = (html: string): string => {
  if (!html || !/<img\b/i.test(html)) return html || ''
  return html.replace(/<img\b(?![^>]*\breferrerpolicy=)[^>]*>/gi, (tag) =>
    tag.replace(/^<img\b/i, '<img referrerpolicy="no-referrer"')
  )
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
    const blobs = extractImagesFromRTF(rtf)
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
    if (queryForm.tagId !== undefined) params.tagId = queryForm.tagId
    if (queryForm.columnPath && queryForm.columnPath.length > 0) {
      params.columnId = queryForm.columnPath[queryForm.columnPath.length - 1]
    }
    if (queryForm.status !== undefined) params.status = queryForm.status
    if (queryForm.auditStatus !== undefined) params.auditStatus = queryForm.auditStatus
    if (queryForm.type !== undefined) params.type = queryForm.type
    if (queryForm.author) params.author = queryForm.author
    if (queryForm.source) params.source = queryForm.source
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
  queryForm.tagId = undefined
  queryForm.status = undefined
  queryForm.auditStatus = undefined
  queryForm.type = undefined
  queryForm.author = ''
  queryForm.source = ''
  queryForm.columnPath = []
  queryForm.page = 1
  fetchData()
}

const handleAdd = () => {
  dialogTitle.value = '新增文章'
  resetForm()
  dialogVisible.value = true
}

const handleAddPaper = () => {
  paperDialogTitle.value = '新增报刊'
  resetPaperForm()
  paperDialogVisible.value = true
}

const handleAddVideo = () => {
  videoDialogTitle.value = '新增视频'
  resetVideoForm()
  videoDialogVisible.value = true
}

const handleAddData = () => {
  dataDialogTitle.value = '新增数据'
  resetDataForm()
  dataDialogVisible.value = true
}

const isAuthor = (row: any) => {
  return String(currentUserId.value) === String(row.authorCode)
}

const handleEdit = async (row: any) => {
  if (row.auditStatus === 1) {
    ElMessage.warning('审核中的文章不能编辑')
    return
  }
  if (row.auditStatus === 2 && row.status !== 2) {
    ElMessage.warning('已审核通过的文章不能编辑')
    return
  }
  const res: any = await getArticleByID(row.id)
  const detail = res.data || row
  if (row.type === 2) {
    videoDialogTitle.value = '编辑视频'
    resetVideoForm()
    const att = detail.attachments?.[0] || {}
    Object.assign(videoForm, {
      id: detail.id,
      title: detail.title,
      categoryIds: detail.categoryIds || [],
      tagIds: detail.tagIds || [],
      source: detail.source || '',
      publishTime: detail.publishTime || '',
      cover: detail.cover || '',
      videoUrl: att.url || '',
      videoName: att.name || '',
      videoSize: att.size || 0
    })
    videoDialogVisible.value = true
    return
  }
  if (row.type === 3) {
    dataDialogTitle.value = '编辑数据'
    resetDataForm()
    Object.assign(dataForm, {
      id: detail.id,
      title: detail.title,
      categoryIds: detail.categoryIds || [],
      tagIds: detail.tagIds || [],
      source: detail.source || '',
      publishTime: detail.publishTime || '',
      yearMonth: detail.summary || '',
      content: detail.content || ''
    })
    dataDialogVisible.value = true
    return
  }
  if (row.type === 4) {
    paperDialogTitle.value = '编辑报刊'
    resetPaperForm()
    Object.assign(paperForm, {
      id: detail.id,
      title: detail.title,
      issueNo: detail.summary || '',
      yearMonth: detail.publishTime ? detail.publishTime.slice(0, 7) : '',
      source: detail.source || '',
      cover: detail.cover || '',
      abstract: detail.content || '',
      categoryIds: detail.categoryIds || [],
      tagIds: detail.tagIds || [],
      pdf: detail.attachments?.[0] || null
    })
    paperDialogVisible.value = true
    return
  }
  dialogTitle.value = '编辑文章'
  resetForm()
  Object.assign(form, {
    id: detail.id,
    title: detail.title,
    type: detail.type || 1,
    categoryIds: detail.categoryIds || [],
    tagIds: detail.tagIds || [],
    summary: detail.summary || '',
    content: detail.content || '',
    status: detail.status,
    auditStatus: detail.auditStatus ?? 0,
    isTop: detail.isTop,
    isBold: detail.isBold ?? 0,
    defaultColor: detail.defaultColor || '',
    cover: detail.cover || '',
    source: detail.source || '',
    publishTime: detail.publishTime || '',
    url: detail.url || '',
    attachments: (detail.attachments || []).map((att: any) => ({
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
    deletingText.value = '文章删除中...'
    deleting.value = true
    try {
      await deleteArticle(row.id)
      ElMessage.success('删除成功')
      fetchData()
    } finally {
      deleting.value = false
    }
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

const formatApprover = (item: any) => {
  const type = item.currentApproverType || 'user'
  const id = item.currentApproverId || 0
  switch (type) {
    case 'dept_head':
      return '部门负责人'
    case 'role': {
      const role = auditFlowRoleList.value.find((r: any) => r.id === id)
      return role ? `角色：${role.name}` : '角色'
    }
    case 'user':
    default: {
      const user = auditFlowUserList.value.find((u: any) => u.id === id)
      return user ? `指定成员：${user.username}` : '指定成员'
    }
  }
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
    // 加载用户/角色选项，用于显示审批人名称
    // 使用菜单豁免的轻量选项接口（仅 id/名称）：非管理员没有「用户管理/流程角色」菜单，
    // 若改回 getAllUsers/getAllWorkflowRoles 会因缺少授权而报「没有授权」。
    if (auditFlowUserList.value.length === 0) {
      try {
        const userRes: any = await getUserOptions()
        auditFlowUserList.value = userRes.data || []
      } catch {
        auditFlowUserList.value = []
      }
    }
    if (auditFlowRoleList.value.length === 0) {
      try {
        const roleRes: any = await getWorkflowRoleOptions()
        auditFlowRoleList.value = roleRes.data || []
      } catch {
        auditFlowRoleList.value = []
      }
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
        currentApproverType: 'user',
        currentApproverName: '',
        canApprove: false,
        approveUserName: '',
        approveTime: '',
        approveRemark: '',
        rejectRemark: '',
        histories: []
      }
      const progress = progressList.find((p: any) => p.columnId === col.id)
      if (progress) {
        item.currentNodeId = progress.currentNodeId
        item.auditStatus = progress.status
        item.approveUserName = progress.approveUserName || ''
        item.approveTime = progress.approveTime || ''
        item.approveRemark = progress.approveRemark || ''
        item.rejectRemark = progress.rejectRemark || ''
        item.canApprove = !!progress.canApprove
        item.currentApproverName = progress.currentApproverName || ''
      }
      if (col.workflowId) {
        try {
          const res: any = await getWorkflowByID(col.workflowId)
          item.workflow = res.data || null
          if (item.workflow && item.workflow.nodes && item.currentNodeId) {
            const node = item.workflow.nodes.find((n: any) => n.id === item.currentNodeId)
            if (node) {
              item.currentApproverId = node.approverId || 0
              item.currentApproverType = node.approverType || 'user'
              // 优先使用后端 progress 返回的具体审批人名称，workflow 节点本身不保存名称
              if (!item.currentApproverName) {
                item.currentApproverName = node.approverName || ''
              }
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
  // 处理中或已有其他审核操作进行时，忽略重复点击
  if (!currentAuditRow.value || auditFlowSubmitting.value) return
  let remark = ''
  try {
    const { value } = await ElMessageBox.prompt('请输入审核通过原因（可选）', '审核通过', {
      confirmButtonText: '确定通过',
      cancelButtonText: '取消',
      inputPattern: /^.{0,500}$/,
      inputErrorMessage: '原因最多500字'
    })
    remark = value || ''
  } catch {
    return // 用户取消
  }
  const row = currentAuditRow.value
  auditFlowSubmitting.value = true
  auditFlowSubmittingKey.value = `${columnId}:advance`
  try {
    await advanceArticleAudit(row.id, columnId, remark)
    ElMessage.success('已通过当前节点')
    await handleShowAuditFlow(row)
    await fetchData()
  } catch (error: any) {
    ElMessage.error(error?.message || '操作失败')
  } finally {
    auditFlowSubmitting.value = false
    auditFlowSubmittingKey.value = ''
  }
}

const handleRejectAuditNode = async (columnId: number) => {
  // 处理中或已有其他审核操作进行时，忽略重复点击
  if (!currentAuditRow.value || auditFlowSubmitting.value) return
  let remark = ''
  try {
    const { value } = await ElMessageBox.prompt('请输入驳回原因', '审核驳回', {
      confirmButtonText: '确定驳回',
      cancelButtonText: '取消',
      inputPattern: /\S+/,
      inputErrorMessage: '驳回原因不能为空'
    })
    remark = value || ''
  } catch {
    return // 用户取消
  }
  const row = currentAuditRow.value
  auditFlowSubmitting.value = true
  auditFlowSubmittingKey.value = `${columnId}:reject`
  try {
    await rejectArticleAudit(row.id, columnId, remark)
    ElMessage.success('已驳回')
    await handleShowAuditFlow(row)
    await fetchData()
  } catch (error: any) {
    ElMessage.error(error?.message || '操作失败')
  } finally {
    auditFlowSubmitting.value = false
    auditFlowSubmittingKey.value = ''
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
    deletingText.value = '提交审核中...'
    deleting.value = true
    try {
      await auditArticle(row.id, 1)
      ElMessage.success('提交审核成功')
      fetchData()
    } finally {
      deleting.value = false
    }
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
    deletingText.value = '重新提交审核中...'
    deleting.value = true
    try {
      await restartArticleAudit(row.id)
      ElMessage.success('重新提交审核成功')
      fetchData()
    } finally {
      deleting.value = false
    }
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
    deletingText.value = '文章下线中...'
    deleting.value = true
    try {
      await updateArticleStatus(row.id, 2)
      ElMessage.success('下线成功')
      fetchData()
    } finally {
      deleting.value = false
    }
  })
}

const handleSetColumns = async (row: any) => {
  currentArticleId.value = row.id
  currentArticleType.value = row.type
  columnDialogTitle.value = row.title
  selectedColumnIds.value = row.columnIds || []
  selectedColumnPageId.value = undefined
  if (pageList.value.length === 0) {
    await fetchPages()
  }
  if (columnList.value.length === 0) {
    await fetchColumns()
  }
  // 若文章已有栏目，默认选中第一个有效栏目所在页面
  if (selectedColumnIds.value.length > 0) {
    const firstCol = columnList.value.find((col: any) => col.id === selectedColumnIds.value[0])
    if (firstCol) {
      selectedColumnPageId.value = firstCol.pageId
    }
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
  createdAt: '',
  summary: '',
  content: '',
  url: '',
  cover: '',
  attachments: [] as any[]
})

const handlePreview = async (row: any) => {
  // 列表不再返回正文（大字段），预览时按需拉取详情
  let content = row.content || ''
  if (!content) {
    try {
      const res: any = await getArticleByID(row.id)
      content = res.data?.content || ''
    } catch {
      content = ''
    }
  }
  Object.assign(previewData, {
    title: row.title,
    author: row.author,
    source: row.source || '',
    createdAt: row.createdAt,
    summary: row.summary || '',
    content: noReferrerContent(content),
    url: row.url || '',
    cover: row.cover || '',
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
      type: form.type,
      categoryIds: form.categoryIds,
      tagIds: form.tagIds,
      summary: form.summary,
      content: noReferrerContent(form.content),
      // status/auditStatus 不由列表编辑接口变更：发布/下线只能走状态接口或审核流程（后端亦会忽略这两个字段）
      isTop: form.isTop,
      isBold: form.isBold,
      defaultColor: form.defaultColor,
      cover: form.cover,
      source: form.source,
      publishTime: form.publishTime,
      url: form.url,
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

const handleVideoCoverUpload = async (options: any) => {
  try {
    const res: any = await uploadFile(options.file, 'article')
    videoForm.cover = res.data?.url || res.url || ''
    ElMessage.success('封面图上传成功')
  } catch (error: any) {
    ElMessage.error(error?.message || '封面图上传失败')
  }
}

const handleRemoveVideoCover = () => {
  videoForm.cover = ''
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

const handleAttachmentRemove = (_file: any, fileList: any[]) => {
  form.attachments = fileList
}

const handleAttachmentChange = () => {
  formRef.value?.validateField(['url', 'content']).catch(() => {})
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
  form.type = 1
  form.categoryIds = []
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
  form.publishTime = ''
  form.url = ''
  form.attachments = []
}

const resetVideoForm = () => {
  videoForm.id = undefined
  videoForm.categoryIds = []
  videoForm.tagIds = []
  videoForm.title = ''
  videoForm.source = ''
  videoForm.publishTime = ''
  videoForm.cover = ''
  videoForm.videoUrl = ''
  videoForm.videoName = ''
  videoForm.videoSize = 0
  videoUploadProgress.value = 0
}

const resetDataForm = () => {
  dataForm.id = undefined
  dataForm.categoryIds = []
  dataForm.tagIds = []
  dataForm.title = ''
  dataForm.source = ''
  dataForm.publishTime = ''
  dataForm.yearMonth = ''
  dataForm.content = ''
}

const handleVideoBeforeUpload = (file: File) => {
  const maxSize = 800 * 1024 * 1024
  if (file.size > maxSize) {
    ElMessage.warning('视频文件大小不能超过 800MB')
    return false
  }
  return true
}

const handleVideoUpload = async (options: any) => {
  const file = options.file
  const formData = new FormData()
  formData.append('file', file)
  formData.append('dir', 'video')
  videoUploadProgress.value = 0
  try {
    const res: any = await request.post('/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      // 大文件上传不套用默认 30s 超时，避免 800MB 视频被提前中断
      timeout: 0,
      // 进度提示，便于观察大文件上传进度
      onUploadProgress: (evt: any) => {
        if (evt.total) {
          const pct = Math.round((evt.loaded / evt.total) * 100)
          videoUploadProgress.value = pct
        }
      }
    })
    const url = res.data?.url || ''
    if (url) {
      videoForm.videoUrl = url
      videoForm.videoName = file.name
      videoForm.videoSize = file.size
      options.onSuccess({ url, name: file.name, size: file.size })
      ElMessage.success('视频上传成功')
    } else {
      options.onError(new Error('上传失败'))
      ElMessage.error('视频上传失败')
    }
  } catch (error: any) {
    options.onError(error)
    ElMessage.error(error?.message || '视频上传失败')
  }
}

const handleVideoRemove = () => {
  videoForm.videoUrl = ''
  videoForm.videoName = ''
  videoForm.videoSize = 0
  videoUploadProgress.value = 0
}

const handleSubmitVideo = async () => {
  const valid = await videoFormRef.value?.validate().catch(() => false)
  if (!valid) return
  videoSubmitLoading.value = true
  try {
    const data = {
      title: videoForm.title,
      type: 2,
      categoryIds: videoForm.categoryIds,
      tagIds: videoForm.tagIds,
      summary: '',
      content: '',
      status: 0,
      auditStatus: 0,
      isTop: 0,
      isBold: 0,
      defaultColor: '',
      cover: videoForm.cover,
      source: videoForm.source,
      publishTime: videoForm.publishTime,
      url: '',
      attachments: [
        {
          name: videoForm.videoName || 'video',
          url: videoForm.videoUrl,
          size: videoForm.videoSize || 0
        }
      ]
    }
    if (videoForm.id) {
      await updateArticle(videoForm.id, data)
      ElMessage.success('编辑视频成功')
    } else {
      await createArticle(data)
      ElMessage.success('新增视频成功')
    }
    videoDialogVisible.value = false
    fetchData()
  } finally {
    videoSubmitLoading.value = false
  }
}

const handleSubmitData = async () => {
  const valid = await dataFormRef.value?.validate().catch(() => false)
  if (!valid) return
  dataSubmitLoading.value = true
  try {
    const data = {
      title: dataForm.title,
      type: 3,
      categoryIds: dataForm.categoryIds,
      tagIds: dataForm.tagIds,
      summary: dataForm.yearMonth,
      content: noReferrerContent(dataForm.content),
      status: 0,
      auditStatus: 0,
      isTop: 0,
      isBold: 0,
      defaultColor: '',
      cover: '',
      source: dataForm.source,
      publishTime: dataForm.publishTime,
      url: '',
      attachments: [] as any[]
    }
    if (dataForm.id) {
      await updateArticle(dataForm.id, data)
      ElMessage.success('编辑数据成功')
    } else {
      await createArticle(data)
      ElMessage.success('新增数据成功')
    }
    dataDialogVisible.value = false
    fetchData()
  } finally {
    dataSubmitLoading.value = false
  }
}

const resetPaperForm = () => {
  paperForm.id = undefined
  paperForm.title = ''
  paperForm.issueNo = ''
  paperForm.yearMonth = ''
  paperForm.source = ''
  paperForm.cover = ''
  paperForm.abstract = ''
  paperForm.categoryIds = []
  paperForm.tagIds = []
  paperForm.pdf = null
}

const handlePaperCoverUpload = async (options: any) => {
  try {
    const res: any = await uploadFile(options.file, 'article')
    paperForm.cover = res.data?.url || res.url || ''
    ElMessage.success('封面图上传成功')
  } catch (error: any) {
    ElMessage.error(error?.message || '封面图上传失败')
  }
}

const handleRemovePaperCover = () => {
  paperForm.cover = ''
}

const handlePaperAttachmentBeforeUpload = (file: File) => {
  const isPdf = /\.pdf$/i.test(file.name) || file.type === 'application/pdf'
  if (!isPdf) {
    ElMessage.warning('仅支持 PDF 格式的报刊文件')
    return false
  }
  if (file.size > 50 * 1024 * 1024) {
    ElMessage.warning('PDF 文件大小不能超过 50MB')
    return false
  }
  return true
}

const handlePaperAttachmentUpload = async (options: any) => {
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
      paperForm.pdf = { url, name: file.name, size: file.size, uid: file.uid }
      options.onSuccess({ url, name: file.name, size: file.size })
      ElMessage.success('报刊文件上传成功')
    } else {
      options.onError(new Error('上传失败'))
      ElMessage.error('报刊文件上传失败')
    }
  } catch (error: any) {
    options.onError(error)
    ElMessage.error(error?.message || '报刊文件上传失败')
  }
}

const handlePaperAttachmentRemove = () => {
  paperForm.pdf = null
}

const handleSubmitPaper = async () => {
  const valid = await paperFormRef.value?.validate().catch(() => false)
  if (!valid) return
  paperSubmitLoading.value = true
  try {
    const data = {
      title: paperForm.title,
      type: 4,
      categoryIds: paperForm.categoryIds,
      tagIds: paperForm.tagIds,
      summary: paperForm.issueNo,
      content: noReferrerContent(paperForm.abstract),
      status: 0,
      auditStatus: 0,
      isTop: 0,
      isBold: 0,
      defaultColor: '',
      cover: paperForm.cover,
      source: paperForm.source,
      publishTime: paperForm.yearMonth ? `${paperForm.yearMonth}-01 00:00:00` : '',
      url: '',
      attachments: paperForm.pdf && paperForm.pdf.url
        ? [{ name: paperForm.pdf.name, url: paperForm.pdf.url, size: paperForm.pdf.size || 0 }]
        : []
    }
    if (paperForm.id) {
      await updateArticle(paperForm.id, data)
      ElMessage.success('编辑报刊成功')
    } else {
      await createArticle(data)
      ElMessage.success('新增报刊成功')
    }
    paperDialogVisible.value = false
    fetchData()
  } finally {
    paperSubmitLoading.value = false
  }
}

const handleSizeChange = (val: number) => {
  queryForm.pageSize = val
  fetchData()
}

const handleCurrentChange = (val: number) => {
  queryForm.page = val
  fetchData()
}

const checkAutoAudit = async () => {
  const auditArticleId = route.query.auditArticleId
  if (!auditArticleId) return
  let row = tableData.value.find((item: any) => String(item.id) === String(auditArticleId))
  if (!row) {
    // 列表按归属过滤后，审核人看不到他人文章的表格行；直接按 ID 拉详情构造最小行，
    // 保证从「待审核」页/仪表盘跳转过来仍能打开审核流程弹窗（后端仅对作者/管理员/可审人放行）。
    try {
      const res: any = await getArticleByID(Number(auditArticleId))
      row = res.data
    } catch {
      ElMessage.warning('无法打开该文章的审核流程，可能没有权限或文章不存在')
    }
  }
  if (row) {
    await handleShowAuditFlow(row)
  }
  router.replace({ path: '/content/article', query: {} })
}

onMounted(() => {
  fetchData().then(() => checkAutoAudit())
  fetchCategories()
  fetchTags()
  fetchPages()
  fetchColumns()
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

      .header-actions {
        display: flex;
        gap: 12px;
      }
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
      border-left: 4px solid #002fa7;
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

.url-tip {
  font-size: 13px;
  color: #909399;
  margin-top: 8px;
  line-height: 1.5;
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

  .preview-cover {
    margin-bottom: 16px;
    border-radius: 10px;
    overflow: hidden;
    border: 1px solid #e6f2ff;

    .preview-cover-image {
      display: block;
      width: 100%;
    }
  }

  .preview-summary {
    background: #f5f7fa;
    padding: 12px;
    border-radius: 8px;
    margin-bottom: 16px;
    color: #666;
  }

  .preview-url {
    background: #f0f9ff;
    padding: 12px;
    border-radius: 8px;
    margin-bottom: 16px;
    color: #666;

    a {
      color: #002fa7;
      text-decoration: none;
      word-break: break-all;

      &:hover {
        text-decoration: underline;
      }
    }
  }

  .preview-body {
    line-height: 1.8;
    color: #333;
    word-break: break-word;

    :deep(img) {
      max-width: 100%;
      height: auto;
      display: block;
    }

    :deep(video) {
      max-width: 100%;
      height: auto;
    }

    :deep(table) {
      max-width: 100%;
    }
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
      color: #002fa7;
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
    box-shadow: 0 4px 12px rgba(0, 47, 167, 0.1);
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
  background: #002fa7;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 2px 6px rgba(0, 47, 167, 0.25);
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

.audit-current-node {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 12px;
  color: #606266;
  font-size: 13px;

  .audit-debug-name {
    color: #909399;
    font-size: 12px;
  }
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

.video-uploader {
  .video-tip {
    font-size: 13px;
    color: #909399;
    margin-top: 8px;
    line-height: 1.5;
  }
}

.video-preview {
  margin-top: 16px;
  padding: 12px;
  background: #fafbfc;
  border-radius: 8px;
  border: 1px solid #e6f2ff;
}

.paper-uploader {
  .paper-tip {
    font-size: 13px;
    color: #909399;
    margin-top: 8px;
    line-height: 1.5;
  }
}

.paper-file-preview {
  margin-top: 12px;
  padding: 10px 12px;
  background: #fafbfc;
  border-radius: 8px;
  border: 1px solid #e6f2ff;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #002fa7;
  font-size: 14px;

  a {
    color: #002fa7;
    text-decoration: none;
    word-break: break-all;
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

.column-setting-dialog {
  :deep(.el-dialog__body) {
    padding: 20px;
    padding-bottom: 12px;
  }

  .column-setting-hint {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 14px;
    background: #f0f9ff;
    border: 1px solid #ccd5ed;
    border-radius: 8px;
    color: #002fa7;
    font-size: 13px;
    margin-bottom: 16px;
  }

  .column-setting-body {
    margin-bottom: 16px;
  }

  .column-setting-panel {
    height: 468px;
    border: 1px solid #e6f2ff;
    border-radius: 12px;
    background: #fafbfc;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .column-setting-panel-title {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 12px 14px;
    font-size: 15px;
    font-weight: 600;
    color: #2c3e50;
    border-bottom: 1px solid #e6f2ff;
    background: #fff;

    .el-icon {
      color: #002fa7;
      font-size: 18px;
    }

    .column-setting-count {
      font-size: 13px;
      color: #909399;
      font-weight: 400;
      margin-left: 2px;
    }

    .column-select-all {
      margin-left: auto;
      font-weight: 400;
    }
  }

  .column-setting-scroll {
    flex: 1;
    padding: 10px;
  }

  .column-setting-empty {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .page-item {
    position: relative;
    padding: 12px 14px;
    margin-bottom: 8px;
    background: #fff;
    border: 1px solid #e6f2ff;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      border-color: #8097d3;
      box-shadow: 0 2px 8px rgba(0, 47, 167, 0.08);
    }

    &.active {
      border-color: #002fa7;
      background: #f0f9ff;
      box-shadow: 0 2px 8px rgba(0, 47, 167, 0.12);
    }

    .page-item-name {
      font-size: 14px;
      font-weight: 500;
      color: #2c3e50;
      margin-bottom: 4px;
      padding-right: 20px;
    }

    .page-item-meta {
      font-size: 12px;
      color: #909399;
    }

    .page-item-check {
      position: absolute;
      right: 10px;
      top: 50%;
      transform: translateY(-50%);
      color: #002fa7;
      font-size: 16px;
    }
  }

  .column-checkbox-group {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }

  .column-card {
    background: #fff;
    border: 1px solid #e6f2ff;
    border-radius: 8px;
    padding: 12px;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      border-color: #8097d3;
      box-shadow: 0 2px 8px rgba(0, 47, 167, 0.08);
    }

    &.checked {
      border-color: #002fa7;
      background: #f0f9ff;
      box-shadow: 0 2px 8px rgba(0, 47, 167, 0.12);
    }

    &.is-child {
      border-left: 3px solid #e6a23c;
      background: #fdfcf6;
    }

    :deep(.el-checkbox) {
      height: auto;
      align-items: flex-start;

      .el-checkbox__input {
        margin-top: 2px;
      }

      .el-checkbox__label {
        display: flex;
        flex-direction: column;
        gap: 6px;
        padding-left: 8px;
        white-space: normal;
        line-height: 1.4;
      }
    }

    .column-card-info {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .column-card-name {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 14px;
      color: #2c3e50;
      font-weight: 500;

      .child-column-icon {
        color: #e6a23c;
        font-size: 14px;
      }
    }

    .column-card-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
    }

    .column-card-tag {
      align-self: flex-start;
    }

    .column-card-parent {
      font-size: 12px;
      color: #909399;
      padding-left: 20px;
    }
  }

  .column-setting-selected {
    border: 1px solid #e6f2ff;
    border-radius: 12px;
    padding: 12px 14px;
    background: #fff;

    .column-setting-selected-title {
      display: flex;
      align-items: center;
      justify-content: space-between;
      font-size: 14px;
      font-weight: 600;
      color: #2c3e50;
      margin-bottom: 10px;

      .column-setting-count {
        font-size: 13px;
        color: #909399;
        font-weight: 400;
      }
    }

    .column-setting-selected-empty {
      font-size: 13px;
      color: #c0c4cc;
      padding: 8px 0;
    }

    .column-setting-selected-list {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }

    .column-selected-tag {
      font-size: 13px;
    }
  }
}

.title-with-type {
  display: inline-flex;
  align-items: center;
  gap: 6px;

  .type-tag {
    flex-shrink: 0;
  }

  .el-link,
  > span:last-child {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>

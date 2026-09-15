<template>
  <div class="template-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>消息模板</span>
          <el-button type="primary" @click="handleAdd">新增模板</el-button>
        </div>
      </template>
      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column prop="code" label="模板编码" />
        <el-table-column prop="name" label="模板名称" />
        <el-table-column prop="channel" label="渠道" width="120" />
        <el-table-column prop="subject" label="主题" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑模板' : '新增模板'" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="模板编码" prop="code">
          <el-input v-model="form.code" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="渠道" prop="channel">
          <el-select v-model="form.channel" style="width: 100%">
            <el-option label="站内信" value="in-app" />
            <el-option label="邮件" value="email" />
          </el-select>
          <div class="form-tip">短信 / 企微渠道尚未接入发送器，故未开放</div>
        </el-form-item>
        <el-form-item label="主题">
          <el-input v-model="form.subject" />
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <el-input v-model="form.content" type="textarea" :rows="5" />
        </el-form-item>
        <el-form-item label="变量说明">
          <el-input v-model="form.variables" type="textarea" :rows="3" placeholder='{"name":"用户名"}' />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getMessageTemplateList, createMessageTemplate, updateMessageTemplate, deleteMessageTemplate } from '@/api/message'
import type { MessageTemplate } from '@/api/message'

const loading = ref(false)
const tableData = ref<MessageTemplate[]>([])
const total = ref(0)
const query = reactive({ page: 1, size: 10, keyword: '' })
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<any>(null)
const form = reactive<MessageTemplate>({
  id: 0,
  code: '',
  name: '',
  channel: 'in-app',
  subject: '',
  content: '',
  variables: '',
  status: 1,
  description: ''
})

const rules = {
  code: [{ required: true, message: '请输入模板编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  channel: [{ required: true, message: '请选择渠道', trigger: 'change' }],
  content: [{ required: true, message: '请输入内容', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  const res: any = await getMessageTemplateList(query)
  tableData.value = res.data.list
  total.value = res.data.total
  loading.value = false
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: MessageTemplate) => {
  isEdit.value = true
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row: MessageTemplate) => {
  await ElMessageBox.confirm('确认删除该模板？', '提示', { type: 'warning' })
  await deleteMessageTemplate(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (isEdit.value) {
    await updateMessageTemplate(form.id, form)
  } else {
    await createMessageTemplate(form)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchData()
}

const resetForm = () => {
  form.id = 0
  form.code = ''
  form.name = ''
  form.channel = 'in-app'
  form.subject = ''
  form.content = ''
  form.variables = ''
  form.status = 1
  form.description = ''
}

onMounted(fetchData)
</script>

<style scoped lang="scss">
.template-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
  .form-tip {
    width: 100%;
    font-size: 12px;
    color: #94a3b8;
    line-height: 1.6;
  }
}
</style>

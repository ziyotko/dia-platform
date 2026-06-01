<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="评论人">
          <el-input v-model="queryForm.nickname" placeholder="请输入评论人" clearable />
        </el-form-item>
        <el-form-item label="所属文章">
          <el-input v-model="queryForm.articleTitle" placeholder="请输入文章标题" clearable />
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="queryForm.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="已通过" :value="1" />
            <el-option label="待审核" :value="0" />
            <el-option label="已拒绝" :value="2" />
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
          <span>评论列表</span>
          <el-button type="danger" @click="handleBatchDelete">
            <el-icon><Delete /></el-icon>批量删除
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="nickname" label="评论人" width="120" />
        <el-table-column prop="content" label="评论内容" min-width="250" show-overflow-tooltip />
        <el-table-column prop="articleTitle" label="所属文章" min-width="180" show-overflow-tooltip />
        <el-table-column prop="status" label="审核状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : row.status === 0 ? 'warning' : 'danger'">
              {{ row.status === 1 ? '已通过' : row.status === 0 ? '待审核' : '已拒绝' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="评论时间" width="170" />
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 0" link type="success" @click="handleApprove(row)">
              <el-icon><CircleCheck /></el-icon>通过
            </el-button>
            <el-button v-if="row.status === 0" link type="danger" @click="handleReject(row)">
              <el-icon><CircleClose /></el-icon>拒绝
            </el-button>
            <el-button link type="primary" @click="handleReply(row)">
              <el-icon><ChatDotRound /></el-icon>回复
            </el-button>
            <el-button link type="danger" @click="handleDelete(row)">
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
      v-model="replyVisible"
      title="回复评论"
      width="500px"
      destroy-on-close
    >
      <el-form :model="replyForm" label-width="80px">
        <el-form-item label="原评论">
          <el-input :model-value="currentRow?.content" type="textarea" :rows="2" disabled />
        </el-form-item>
        <el-form-item label="回复内容" prop="content">
          <el-input v-model="replyForm.content" type="textarea" :rows="4" placeholder="请输入回复内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="replyVisible = false">取消</el-button>
        <el-button type="primary" @click="handleReplySubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  RefreshRight,
  Delete,
  CircleCheck,
  CircleClose,
  ChatDotRound
} from '@element-plus/icons-vue'

const loading = ref(false)
const replyVisible = ref(false)
const total = ref(0)
const currentRow = ref<any>(null)
const selectedRows = ref<any[]>([])

const queryForm = reactive({
  page: 1,
  pageSize: 10,
  nickname: '',
  articleTitle: '',
  status: undefined as number | undefined
})

const replyForm = reactive({
  content: ''
})

const tableData = ref<any[]>([])

const fetchData = async () => {
  loading.value = true
  try {
    const mockData = [
      { id: 1, nickname: '小明', content: '写得真好，受益匪浅！', articleTitle: 'Vue3 组合式函数最佳实践', status: 1, createTime: '2026-05-21 10:00:00' },
      { id: 2, nickname: '小红', content: '请问这个在 Vue2 中可以用吗？', articleTitle: 'Vue3 组合式函数最佳实践', status: 0, createTime: '2026-05-21 11:30:00' },
      { id: 3, nickname: '张三', content: '期待更多这样的文章', articleTitle: '2026年前端技术发展趋势', status: 1, createTime: '2026-05-20 14:00:00' },
      { id: 4, nickname: '李四', content: '有点过时了，建议更新', articleTitle: 'Element Plus 深度定制指南', status: 2, createTime: '2026-05-19 09:00:00' },
      { id: 5, nickname: '王五', content: '可以分享下源码吗？', articleTitle: '公司新产品发布会回顾', status: 0, createTime: '2026-05-18 16:00:00' }
    ]
    tableData.value = mockData
    total.value = mockData.length
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  queryForm.page = 1
  fetchData()
}

const resetQuery = () => {
  queryForm.nickname = ''
  queryForm.articleTitle = ''
  queryForm.status = undefined
  queryForm.page = 1
  fetchData()
}

const handleSelectionChange = (val: any[]) => {
  selectedRows.value = val
}

const handleApprove = (row: any) => {
  ElMessage.success(`已通过评论: ${row.content.substring(0, 20)}...`)
  row.status = 1
}

const handleReject = (row: any) => {
  ElMessage.warning(`已拒绝评论: ${row.content.substring(0, 20)}...`)
  row.status = 2
}

const handleReply = (row: any) => {
  currentRow.value = row
  replyForm.content = ''
  replyVisible.value = true
}

const handleReplySubmit = () => {
  if (!replyForm.content.trim()) {
    ElMessage.warning('请输入回复内容')
    return
  }
  ElMessage.success('回复成功')
  replyVisible.value = false
}

const handleDelete = (_row: any) => {
  ElMessageBox.confirm(`确定要删除该评论吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleBatchDelete = () => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning('请选择要删除的评论')
    return
  }
  ElMessageBox.confirm(`确定要删除选中的 ${selectedRows.value.length} 条评论吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success('批量删除成功')
    selectedRows.value = []
    fetchData()
  })
}

const handleSizeChange = (val: number) => {
  queryForm.pageSize = val
  fetchData()
}

const handleCurrentChange = (val: number) => {
  queryForm.page = val
  fetchData()
}

onMounted(() => {
  fetchData()
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
</style>

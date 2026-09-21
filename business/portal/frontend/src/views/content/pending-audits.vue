<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form inline>
        <el-form-item>
          <el-button type="primary" @click="fetchData">
            <el-icon><RefreshRight /></el-icon>刷新
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <span>待处理</span>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="title" label="文章标题" min-width="300" show-overflow-tooltip>
          <template #default="{ row }">
            <span>文章《{{ row.title }}》待审核</span>
          </template>
        </el-table-column>
        <el-table-column prop="author" label="作者" width="120" />
        <el-table-column prop="createdAt" label="提交时间" width="170" />
        <el-table-column label="操作" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleAudit(row)">
              <el-icon><View /></el-icon>审核
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { RefreshRight, View } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getMyAuditArticles } from '@/api/dashboard'
import { canAccessPath } from '@/utils/permission'

const router = useRouter()
const loading = ref(false)
const total = ref(0)

const queryForm = reactive({
  page: 1,
  pageSize: 10
})

const tableData = ref<any[]>([])

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getMyAuditArticles({
      page: queryForm.page,
      pageSize: queryForm.pageSize
    })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('获取待审核文章失败')
  } finally {
    loading.value = false
  }
}

const handleAudit = (row: any) => {
  // 审核详情页属于「图文管理」菜单，未授权的用户跳过去只会落到 404，这里直接给出提示
  if (!canAccessPath('/content/article')) {
    ElMessage.warning('您没有「图文管理」的访问权限，无法处理审核')
    return
  }
  router.push({
    path: '/content/article',
    query: { auditArticleId: String(row.id) }
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

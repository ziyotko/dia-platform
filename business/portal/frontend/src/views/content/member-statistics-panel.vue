<template>
  <div class="statistics-panel">
    <div class="toolbar">
      <div class="dataset-select">
        <span>年度数据集</span>
        <el-select v-model="selectedId" placeholder="选择统计年度" style="width: 340px" @change="selectDataset">
          <el-option
            v-for="item in datasets"
            :key="item.id"
            :label="`${item.statisticsYear} 年 · ${item.title}`"
            :value="item.id"
          />
        </el-select>
        <el-button :loading="datasetLoading" @click="loadDatasets">刷新</el-button>
      </div>
      <div v-if="detail" class="status-actions">
        <el-tag :type="statusTagType(detail.dataset.status)">{{ statusText(detail.dataset.status) }}</el-tag>
        <el-button
          v-if="detail.dataset.status !== 1"
          type="primary"
          :loading="statusLoading"
          @click="changeStatus(1)"
        >发布</el-button>
        <el-button v-else type="warning" :loading="statusLoading" @click="changeStatus(2)">下线</el-button>
      </div>
    </div>

    <el-empty v-if="!datasetLoading && !detail" description="暂无已导入的年度统计数据" />

    <template v-if="detail">
      <div class="dataset-meta">
        <span>所属栏目：{{ detail.columnName }}</span>
        <span>对比年度：{{ detail.dataset.comparisonYear }}</span>
        <span>导入时间：{{ formatDate(detail.dataset.importedAt) }}</span>
        <span :title="detail.dataset.sourceSha256">来源文件：{{ detail.dataset.sourceFileName }}</span>
      </div>

      <div class="summary-grid">
        <div class="summary-card"><span>样本记录</span><strong>{{ detail.summary.recordCount }}</strong><small>条</small></div>
        <div class="summary-card"><span>去重企业</span><strong>{{ detail.summary.enterpriseCount }}</strong><small>家</small></div>
        <div class="summary-card"><span>覆盖省份</span><strong>{{ detail.summary.provinceCount }}</strong><small>个</small></div>
        <div class="summary-card"><span>{{ detail.dataset.statisticsYear }} 年主营业务收入</span><strong>{{ revenueYi(detail.summary.revenueYearWan) }}</strong><small>亿元</small></div>
        <div class="summary-card"><span>收入整体同比</span><strong>{{ growthPercent(detail.summary.totalYoyPercent) }}</strong><small>%</small></div>
      </div>

      <div class="quality-alerts">
        <el-alert
          v-if="detail.duplicateNames.length"
          type="warning"
          :closable="false"
          show-icon
        >
          <template #title>
            <span>发现 {{ detail.duplicateNames.length }} 组重复企业名称。记录均按 Excel 原始行保留；“去重企业”已按名称去重。</span>
          </template>
          <template #default>
            <el-button
              v-for="item in detail.duplicateNames"
              :key="item.enterpriseName"
              link
              type="warning"
              @click="showDuplicate(item.enterpriseName)"
            >{{ item.enterpriseName }}（Excel 第 {{ item.sourceRows.join('、') }} 行）</el-button>
          </template>
        </el-alert>
        <el-alert
          v-if="detail.summary.missingRdYear || detail.summary.missingRdPrior"
          type="info"
          :closable="false"
          show-icon
          :title="`研发投入缺失：${detail.dataset.statisticsYear} 年 ${detail.summary.missingRdYear} 条，${detail.dataset.comparisonYear} 年 ${detail.summary.missingRdPrior} 条。空白保留为 NULL，未当作 0。`"
        />
      </div>

      <el-form :model="query" inline class="filter-form">
        <el-form-item label="企业名称">
          <el-input v-model="query.enterpriseName" clearable placeholder="输入企业名称" @keyup.enter="searchRows" />
        </el-form-item>
        <el-form-item label="省份">
          <el-select v-model="query.province" clearable placeholder="全部省份" style="width: 130px">
            <el-option v-for="value in detail.filters.provinces" :key="value" :label="value" :value="value" />
          </el-select>
        </el-form-item>
        <el-form-item label="地区">
          <el-select v-model="query.region" clearable placeholder="全部地区" style="width: 130px">
            <el-option v-for="value in detail.filters.regions" :key="value" :label="value" :value="value" />
          </el-select>
        </el-form-item>
        <el-form-item label="细分领域">
          <el-select v-model="query.subField" clearable placeholder="全部领域" style="width: 150px">
            <el-option v-for="value in detail.filters.subFields" :key="value" :label="value" :value="value" />
          </el-select>
        </el-form-item>
        <el-form-item><el-checkbox v-model="query.missingRd">仅看研发缺失</el-checkbox></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchRows">查询</el-button>
          <el-button @click="resetRows">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="rowsLoading"
        :data="rows"
        border
        stripe
        height="430"
        :row-class-name="rowClassName"
        empty-text="没有符合条件的明细"
      >
        <el-table-column prop="sourceRow" label="Excel 行" width="90" fixed="left" />
        <el-table-column prop="enterpriseName" label="企业名称" min-width="250" fixed="left" show-overflow-tooltip />
        <el-table-column prop="province" label="省份" width="85" />
        <el-table-column prop="region" label="地区" width="85" />
        <el-table-column prop="subField" label="细分领域" width="105" />
        <el-table-column :label="`${detail.dataset.statisticsYear} 年收入（万元）`" width="175" align="right">
          <template #default="{ row }">{{ moneyWan(row.revenueYearWan) }}</template>
        </el-table-column>
        <el-table-column :label="`${detail.dataset.comparisonYear} 年收入（万元）`" width="175" align="right">
          <template #default="{ row }">{{ moneyWan(row.revenuePriorWan) }}</template>
        </el-table-column>
        <el-table-column label="收入同比" width="110" align="right">
          <template #default="{ row }">{{ ratioPercent(row.yoySource) }}</template>
        </el-table-column>
        <el-table-column :label="`${detail.dataset.statisticsYear} 年研发（万元）`" width="160" align="right">
          <template #default="{ row }">{{ moneyWan(row.rdYearWan) }}</template>
        </el-table-column>
        <el-table-column :label="`${detail.dataset.comparisonYear} 年研发（万元）`" width="160" align="right">
          <template #default="{ row }">{{ moneyWan(row.rdPriorWan) }}</template>
        </el-table-column>
        <el-table-column :label="`${detail.dataset.statisticsYear} 年研发占比`" width="145" align="right">
          <template #default="{ row }">{{ ratioPercent(row.rdShareYear) }}</template>
        </el-table-column>
        <el-table-column :label="`${detail.dataset.comparisonYear} 年研发占比`" width="145" align="right">
          <template #default="{ row }">{{ ratioPercent(row.rdSharePrior) }}</template>
        </el-table-column>
        <el-table-column label="一带" width="70" align="center">
          <template #default="{ row }">{{ row.isBelt ? '是' : '—' }}</template>
        </el-table-column>
        <el-table-column label="一轴" width="70" align="center">
          <template #default="{ row }">{{ row.isAxis ? '是' : '—' }}</template>
        </el-table-column>
        <el-table-column prop="source" label="数据来源" min-width="105" />
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.pageSize"
          :page-sizes="[20, 50, 100]"
          :total="rowTotal"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="searchRows"
          @current-change="loadRows"
        />
      </div>
      <p class="publish-note">发布后，会员门户“数据中心”将展示该年度统计汇总；下线后立即不再展示。</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getMemberStatisticsDataset,
  getMemberStatisticsDatasets,
  getMemberStatisticsRows,
  updateMemberStatisticsStatus,
  type MemberStatisticsDataset,
  type MemberStatisticsDetail,
  type MemberStatisticsRow,
} from '@/api/member-statistics'

const datasets = ref<MemberStatisticsDataset[]>([])
const selectedId = ref<number>()
const detail = ref<MemberStatisticsDetail | null>(null)
const rows = ref<MemberStatisticsRow[]>([])
const rowTotal = ref(0)
const datasetLoading = ref(false)
const rowsLoading = ref(false)
const statusLoading = ref(false)
const query = reactive({
  enterpriseName: '',
  province: '',
  region: '',
  subField: '',
  missingRd: false,
  page: 1,
  pageSize: 20,
})

const duplicateSourceRows = computed(() =>
  new Set(detail.value?.duplicateNames.flatMap((item) => item.sourceRows) || [])
)

const statusText = (status: number) => ({ 0: '草稿', 1: '已发布', 2: '已下线' })[status] || '未知'
const statusTagType = (status: number) => (status === 1 ? 'success' : status === 2 ? 'info' : 'warning')
const formatDate = (value: string) => (value ? new Date(value).toLocaleString('zh-CN') : '—')
const moneyWan = (value: string | null) =>
  value == null ? '—' : Number(value).toLocaleString('zh-CN', { maximumFractionDigits: 6 })
const revenueYi = (value: string) => (Number(value) / 10000).toLocaleString('zh-CN', { maximumFractionDigits: 1 })
const growthPercent = (value: string) => (value === '' ? '—' : Number(value).toFixed(1))
const ratioPercent = (value: string | null) =>
  value == null ? '—' : `${(Number(value) * 100).toFixed(2)}%`
const rowClassName = ({ row }: { row: MemberStatisticsRow }) =>
  duplicateSourceRows.value.has(row.sourceRow) ? 'duplicate-row' : ''

async function loadDatasets() {
  datasetLoading.value = true
  try {
    const res: any = await getMemberStatisticsDatasets()
    datasets.value = res.data?.list || []
    if (!datasets.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = datasets.value[0]?.id
    }
    await selectDataset()
  } catch {
    detail.value = null
  } finally {
    datasetLoading.value = false
  }
}

async function selectDataset() {
  detail.value = null
  rows.value = []
  if (!selectedId.value) return
  try {
    const res: any = await getMemberStatisticsDataset(selectedId.value)
    detail.value = res.data as MemberStatisticsDetail
    resetQuery()
    await loadRows()
  } catch {
    detail.value = null
  }
}

async function loadRows() {
  if (!selectedId.value) return
  rowsLoading.value = true
  try {
    const res: any = await getMemberStatisticsRows(selectedId.value, {
      enterpriseName: query.enterpriseName || undefined,
      province: query.province || undefined,
      region: query.region || undefined,
      subField: query.subField || undefined,
      missingRd: query.missingRd ? 1 : undefined,
      page: query.page,
      pageSize: query.pageSize,
    })
    rows.value = res.data?.list || []
    rowTotal.value = res.data?.total || 0
  } catch {
    rows.value = []
    rowTotal.value = 0
  } finally {
    rowsLoading.value = false
  }
}

function resetQuery() {
  query.enterpriseName = ''
  query.province = ''
  query.region = ''
  query.subField = ''
  query.missingRd = false
  query.page = 1
}

function searchRows() {
  query.page = 1
  void loadRows()
}

function resetRows() {
  resetQuery()
  void loadRows()
}

function showDuplicate(name: string) {
  resetQuery()
  query.enterpriseName = name
  void loadRows()
}

async function changeStatus(status: 1 | 2) {
  if (!detail.value) return
  const action = status === 1 ? '发布' : '下线'
  try {
    await ElMessageBox.confirm(`确定${action}「${detail.value.dataset.title}」吗？`, `${action}年度统计`, {
      confirmButtonText: action,
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    return
  }
  statusLoading.value = true
  try {
    await updateMemberStatisticsStatus(detail.value.dataset.id, status)
    ElMessage.success(`已${action}`)
    await loadDatasets()
  } catch {
    // 统一请求拦截器已提示错误。
  } finally {
    statusLoading.value = false
  }
}

onMounted(loadDatasets)
</script>

<style scoped>
.statistics-panel { min-height: 550px; }
.toolbar, .dataset-select, .status-actions, .dataset-meta { display: flex; align-items: center; gap: 12px; }
.toolbar { justify-content: space-between; margin-bottom: 16px; }
.dataset-meta { flex-wrap: wrap; color: #606266; font-size: 13px; margin-bottom: 18px; }
.dataset-meta span { margin-right: 12px; }
.summary-grid { display: grid; grid-template-columns: repeat(5, minmax(150px, 1fr)); gap: 12px; margin-bottom: 16px; }
.summary-card { padding: 14px 16px; background: #f6f8fa; border: 1px solid #e9edf1; border-radius: 8px; }
.summary-card span { display: block; color: #606266; font-size: 13px; margin-bottom: 8px; }
.summary-card strong { font-size: 25px; color: #1f2937; }
.summary-card small { margin-left: 5px; color: #606266; }
.quality-alerts { display: grid; gap: 8px; margin-bottom: 16px; }
.filter-form { margin-bottom: 4px; }
.pagination { display: flex; justify-content: flex-end; margin-top: 16px; }
.publish-note { color: #909399; font-size: 12px; margin-top: 14px; }
:deep(.duplicate-row) { --el-table-tr-bg-color: #fff7e7; }
@media (max-width: 1100px) { .summary-grid { grid-template-columns: repeat(3, 1fr); } }
</style>

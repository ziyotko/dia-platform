<template>
  <div class="admin-level-changes" v-loading="loading">
    <div class="page-header">
      <h3>会籍变更记录</h3>
      <div class="filters">
        <el-input v-model="keyword" placeholder="搜索用户名/公司名/等级" clearable style="width:220px" @clear="search" @keyup.enter="search" />
        <el-select v-model="filterType" placeholder="会员类型" clearable style="width:140px" @change="search">
          <el-option label="单位会员" value="unit" />
          <el-option label="个人会员" value="personal" />
        </el-select>
        <el-select v-model="filterYear" placeholder="变更年份" clearable style="width:130px" @change="search">
          <el-option v-for="y in years" :key="y" :label="`${y} 年`" :value="y" />
        </el-select>
        <el-button type="primary" @click="search">查询</el-button>
        <el-button type="success" :loading="exporting" @click="handleExport">导出</el-button>
      </div>
    </div>

    <el-card>
      <el-table :data="list" stripe style="width: 100%" class="level-changes-table" :row-class-name="tableRowClassName" @row-click="openDetail">
        <el-table-column prop="username" label="变更用户名" min-width="120" />
        <el-table-column prop="member_name" label="公司名称/姓名" min-width="180" />
        <el-table-column label="类型" min-width="90">
          <template #default="{ row }">{{ typeLabel(row.member_type) }}</template>
        </el-table-column>
        <el-table-column label="变更时间" min-width="150">
          <template #default="{ row }">{{ fmt(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="change_year" label="变更年份" min-width="90" />
        <el-table-column prop="org_name" label="入会机构" min-width="150">
          <template #default="{ row }">{{ row.org_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="原始会籍" min-width="150">
          <template #default="{ row }">
            <span v-if="row.old_level_id">{{ row.old_level_name }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="新的会籍" min-width="150">
          <template #default="{ row }">
            <el-tag v-if="row.new_level_id" type="warning">{{ row.new_level_name }}</el-tag>
            <el-tag v-else type="info" effect="plain">已失效</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="变更原因" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.reason || '-' }}</template>
        </el-table-column>
        <el-table-column prop="operator" label="变更人" min-width="110">
          <template #default="{ row }">{{ row.operator || '-' }}</template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无会籍变更记录" />
      <div class="pagination">
        <el-pagination background layout="prev, pager, next, total" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" />
      </div>
    </el-card>

    <!-- 会员详情弹窗 -->
    <MemberDetailDialog v-model="detailVisible" :detail="detail" :loading="detailLoading" :level-name="levelName" />

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import MemberDetailDialog from '@/components/MemberDetailDialog.vue'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([])
const loading = ref(true)
const page = ref(1); const size = ref(10); const total = ref(0)
const keyword = ref(''); const filterType = ref(''); const filterYear = ref<number | ''>('')
const years = ref<number[]>([])
const exporting = ref(false)
const levels = ref<any[]>([])

const detailVisible = ref(false)
const detail = ref<any>(null)
const detailLoading = ref(false)

function typeLabel(t: string) {
  return t === 'unit' ? '单位' : t === 'personal' ? '个人' : (t || '-')
}
function levelName(id: any) {
  if (id === null || id === undefined || id === '') return '-'
  const lvl = levels.value.find((l: any) => String(l.id) === String(id))
  return lvl ? lvl.name : String(id)
}
function fmt(d: string) { return d ? d.replace('T', ' ').slice(0, 16) : '-' }
function tableRowClassName() { return 'level-changes-row' }

onMounted(() => { fetchData(); fetchLevels(); fetchYears() })

async function fetchYears() {
  try {
    const res = await adminApi.getLevelChangeYears()
    years.value = res.data || []
  } catch {}
}

async function fetchLevels() {
  try {
    const res = await adminApi.getMemberLevels()
    levels.value = res.data || []
  } catch {}
}

async function openDetail(row: any) {
  detailLoading.value = true
  detailVisible.value = true
  try {
    const res = await adminApi.getMember(row.member_id)
    detail.value = res.data || row
  } catch {
    detail.value = row
  } finally { detailLoading.value = false }
}

async function fetchData() {
  loading.value = true
  try {
    const res = await adminApi.getMemberLevelChanges({
      page: page.value,
      size: size.value,
      keyword: keyword.value,
      member_type: filterType.value,
      year: filterYear.value || undefined
    })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch {} finally { loading.value = false }
}

async function handleExport() {
  exporting.value = true
  try {
    const data: any = await adminApi.exportMemberLevelChanges({
      keyword: keyword.value,
      member_type: filterType.value,
      year: filterYear.value || undefined
    })
    const blob = new Blob([data], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `member_level_changes_${new Date().getTime()}.csv`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch {
    ElMessage.error('导出失败')
  } finally { exporting.value = false }
}

function search() { page.value = 1; fetchData() }
</script>

<style scoped lang="scss">
.admin-level-changes {
  width: 100%;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 12px;
  h3 {
    font-size: 22px;
    font-weight: 600;
    color: #1d2739;
    margin: 0;
  }
}
.filters { display: flex; gap: 12px; }
.el-card { border-radius: 10px; }
.pagination { display: flex; justify-content: center; padding: 20px 0; }

.level-changes-table :deep(.level-changes-row) {
  cursor: pointer;
}
.level-changes-table :deep(.level-changes-row:hover td) {
  background: #f5f7fb !important;
}
</style>

<!-- 详情弹窗样式（dialog 使用 teleport，故使用非 scoped 样式并以类名限定作用域） -->
<style lang="scss">
</style>

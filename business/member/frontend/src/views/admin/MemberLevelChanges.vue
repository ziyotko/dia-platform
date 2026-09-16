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
    <el-dialog v-model="detailVisible" width="800px" class="member-detail-dialog" :show-close="false" :close-on-click-modal="true" append-to-body>
      <template #header>
        <div class="detail-head" v-if="detail">
          <el-avatar :size="60" :src="fileUrl(detail.avatar)" class="detail-avatar">{{ initial(detail) }}</el-avatar>
          <div class="detail-title">
            <div class="name-line">
              <span class="name">{{ displayName(detail) }}</span>
              <el-tag :type="statusTag(detail.status)" size="large">{{ statusLabel(detail.status) }}</el-tag>
            </div>
            <div class="sub">{{ detail.username }} · {{ fullTypeLabel(detail.member_type) }}<template v-if="detail.member_level"> · 等级 {{ levelName(detail.member_level) }}</template></div>
          </div>
        </div>
      </template>

      <div v-loading="detailLoading" class="detail-body">
        <template v-if="detail">
          <div class="section">
            <div class="section-title">基本资料</div>
            <div class="grid">
              <div class="item"><span class="label">会员类型</span><span class="value">{{ fullTypeLabel(detail.member_type) }}</span></div>
              <div class="item"><span class="label">入会机构</span><span class="value">{{ detail.org_name || '-' }}</span></div>
              <div class="item"><span class="label">会员等级</span><span class="value">{{ levelName(detail.member_level) }}</span></div>
              <div class="item"><span class="label">注册时间</span><span class="value">{{ fmt(detail.created_at) }}</span></div>
              <div class="item"><span class="label">最近更新</span><span class="value">{{ fmt(detail.updated_at) }}</span></div>
              <div class="item"><span class="label">手机号</span><span class="value">{{ detail.mobile || '-' }}</span></div>
              <div class="item"><span class="label">邮箱</span><span class="value">{{ detail.email || '-' }}</span></div>
            </div>
          </div>

          <div class="section" v-if="detail.member_type === 'unit'">
            <div class="section-title">单位信息</div>
            <div class="grid">
              <div class="item"><span class="label">公司名称</span><span class="value">{{ detail.company_name || '-' }}</span></div>
              <div class="item"><span class="label">统一社会信用代码</span><span class="value">{{ detail.credit_code || '-' }}</span></div>
              <div class="item"><span class="label">法定代表人</span><span class="value">{{ detail.legal_person || '-' }}</span></div>
              <div class="item"><span class="label">所属行业</span><span class="value">{{ detail.industry || '-' }}</span></div>
              <div class="item"><span class="label">联系人</span><span class="value">{{ detail.contact_person || '-' }}<template v-if="detail.contact_title">（{{ detail.contact_title }}）</template></span></div>
              <div class="item"><span class="label">联系电话</span><span class="value">{{ detail.contact_mobile || '-' }}</span></div>
              <div class="item"><span class="label">成立日期</span><span class="value">{{ detail.founded_date || '-' }}</span></div>
              <div class="item"><span class="label">注册资本</span><span class="value">{{ detail.registered_capital || '-' }}</span></div>
              <div class="item"><span class="label">员工人数</span><span class="value">{{ detail.employee_count != null ? detail.employee_count + ' 人' : '-' }}</span></div>
              <div class="item"><span class="label">邮编</span><span class="value">{{ detail.postal_code || '-' }}</span></div>
              <div class="item span2"><span class="label">地址</span><span class="value">{{ detail.address || '-' }}</span></div>
              <div class="item"><span class="label">网站</span><span class="value"><el-link v-if="detail.website" :href="detail.website" target="_blank" type="primary">{{ detail.website }}</el-link><template v-else>-</template></span></div>
              <div class="item"><span class="label">营业执照</span><span class="value"><el-link v-if="detail.cert_file" :href="fileUrl(detail.cert_file)" target="_blank" type="primary">查看 / 下载</el-link><template v-else>-</template></span></div>
              <div class="item span2"><span class="label">经营范围</span><span class="value">{{ detail.business_scope || '-' }}</span></div>
              <div class="item span2"><span class="label">公司简介</span><span class="value">{{ detail.description || '-' }}</span></div>
            </div>
          </div>

          <div class="section" v-if="detail.member_type === 'personal'">
            <div class="section-title">个人信息</div>
            <div class="grid">
              <div class="item"><span class="label">姓名</span><span class="value">{{ detail.name || '-' }}</span></div>
              <div class="item"><span class="label">身份证号</span><span class="value">{{ detail.id_card || '-' }}</span></div>
            </div>
          </div>
        </template>
      </div>
    </el-dialog>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
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

const statusMap: Record<string, { l: string; t: string }> = {
  registering: { l: '注册中', t: 'info' }, pending_review: { l: '待审核', t: 'warning' },
  pending_payment: { l: '待缴费', t: 'danger' },
  active: { l: '正式会员', t: 'success' }, rejected: { l: '已拒绝', t: 'danger' },
  expired: { l: '已过期', t: 'info' }
}
function statusLabel(s: string) { return statusMap[s]?.l || s }
function statusTag(s: string) { return statusMap[s]?.t || 'info' as any }
function typeLabel(t: string) {
  return t === 'unit' ? '单位' : t === 'personal' ? '个人' : (t || '-')
}
function fullTypeLabel(t: string) { return t === 'unit' ? '单位会员' : t === 'personal' ? '个人会员' : (t || '-') }
function displayName(m: any) {
  if (!m) return '-'
  return m.member_type === 'unit' ? (m.company_name || m.username || '-') : (m.name || m.username || '-')
}
function initial(m: any) { return ((displayName(m) || '?').trim()[0] || '?').toUpperCase() }
function levelName(id: any) {
  if (id === null || id === undefined || id === '') return '-'
  const lvl = levels.value.find((l: any) => String(l.id) === String(id))
  return lvl ? lvl.name : String(id)
}
function fmt(d: string) { return d ? d.replace('T', ' ').slice(0, 16) : '-' }
function fileUrl(path?: string) {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  // 兼容 `uploads/...`（相对）、`/uploads/...`（旧绝对）与带部署前缀的绝对路径，统一指向当前部署子路径
  const base = import.meta.env.BASE_URL || '/'
  const clean = path.replace(/^\.?\//, '')
  return clean.startsWith('uploads/') ? `${base}${clean}` : `/${clean}`
}
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
.member-detail-dialog {
  border-radius: 14px;
  overflow: hidden;

  .el-dialog__header {
    padding: 22px 24px 16px;
    background: linear-gradient(135deg, #f4f7ff 0%, #eef1ff 100%);
    border-bottom: 1px solid #eef0f3;
    margin-right: 0;
  }
  .el-dialog__body {
    padding: 24px;
  }

  .detail-head {
    display: flex;
    align-items: center;
    gap: 16px;
    padding-right: 40px;
  }
  .detail-avatar {
    background: linear-gradient(135deg, #002fa7, #002686);
    color: #fff;
    font-size: 24px;
    font-weight: 600;
    flex-shrink: 0;
  }
  .detail-title {
    .name-line {
      display: flex;
      align-items: center;
      gap: 12px;
      .name { font-size: 20px; font-weight: 600; color: #1d2739; }
    }
    .sub { margin-top: 6px; color: #909399; font-size: 13px; }
  }

  .detail-body {
    padding-top: 4px;
    .section { margin-bottom: 26px; &:last-child { margin-bottom: 0; } }
    .section-title {
      font-size: 15px;
      font-weight: 600;
      color: #1d2739;
      padding-left: 10px;
      border-left: 4px solid #002fa7;
      margin-bottom: 14px;
      line-height: 1.2;
    }
    .grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 14px 28px;
      background: #fafbfc;
      border: 1px solid #eef0f3;
      border-radius: 10px;
      padding: 18px 22px;
      .item {
        display: flex;
        flex-direction: column;
        gap: 4px;
        &.span2 { grid-column: 1 / -1; }
      }
      .label { font-size: 12px; color: #909399; }
      .value { font-size: 14px; color: #303133; line-height: 1.6; word-break: break-word; }
    }
  }
}
</style>

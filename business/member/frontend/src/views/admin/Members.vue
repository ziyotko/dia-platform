<template>
  <div class="admin-members" v-loading="loading">
    <div class="page-header">
      <h3>会籍管理</h3>
      <div class="filters">
        <el-input v-model="keyword" placeholder="搜索用户名/公司名" clearable style="width:200px" @clear="search" @keyup.enter="search" />
        <el-select v-model="filterStatus" placeholder="状态" clearable style="width:130px" @change="search">
          <el-option label="注册中" value="registering" /><el-option label="待审核" value="pending_review" />
          <el-option label="待缴费" value="pending_payment" /><el-option label="正式会员" value="active" />
          <el-option label="已拒绝" value="rejected" /><el-option label="已过期" value="expired" />
        </el-select>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>

    <el-card>
      <el-table :data="list" stripe style="width: 100%" class="members-table" :row-class-name="tableRowClassName" @row-click="openDetail">
        <el-table-column prop="id" label="ID" min-width="70" />
        <el-table-column prop="username" label="用户名" min-width="120"/>
        <el-table-column label="公司名称/姓名" min-width="200">
          <template #default="{row}">{{ row.member_type === 'unit' ? row.company_name : row.name }}</template>
        </el-table-column>
        <el-table-column prop="member_type" label="类型" min-width="100"><template #default="{row}">{{ row.member_type === 'unit' ? '单位' : '个人' }}</template></el-table-column>
        <el-table-column prop="status" label="状态" min-width="120">
          <template #default="{row}"><el-tag :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" min-width="180">
          <template #default="{row}">
            <el-button text size="small" type="primary" @click.stop="openDetail(row)">查看</el-button>
            <el-button text size="small" type="warning" :disabled="row.status !== 'active'" @click.stop="openLevelDialog(row)">变更等级</el-button>
            <el-button text size="small" type="danger" :disabled="row.status !== 'registering'" @click.stop="delMember(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination"><el-pagination background layout="prev, pager, next" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" /></div>
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
            <div class="sub">{{ detail.username }} · {{ typeLabel(detail.member_type) }}<template v-if="detail.member_level"> · 等级 {{ detail.member_level }}</template></div>
          </div>
        </div>
      </template>

      <div v-loading="detailLoading" class="detail-body">
        <template v-if="detail">
          <div class="section">
            <div class="section-title">基本资料</div>
            <div class="grid">
              <div class="item"><span class="label">会员类型</span><span class="value">{{ typeLabel(detail.member_type) }}</span></div>
              <div class="item"><span class="label">会员等级</span><span class="value">{{ detail.member_level || '-' }}</span></div>
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

    <!-- 变更会员等级弹窗 -->
    <el-dialog v-model="levelDialogVisible" title="变更会员等级" width="460px" class="member-level-dialog" :close-on-click-modal="false" append-to-body>
      <div v-loading="levelLoading">
        <p class="level-hint">仅可为该会员选择其已缴费加入的会籍等级。</p>
        <el-form label-width="90px">
          <el-form-item label="当前等级">
            <span>{{ levelTarget?.member_level || '-' }}</span>
          </el-form-item>
          <el-form-item label="新等级" required>
            <el-select v-model="selectedLevel" placeholder="请选择会员等级" style="width:100%">
              <el-option v-for="lvl in levelOptions" :key="lvl.id" :label="lvl.name" :value="lvl.name" />
            </el-select>
            <div v-if="!levelLoading && !levelOptions.length" class="level-empty">暂无已缴费的会籍等级可选</div>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="levelDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingLevel" :disabled="!levelOptions.length || !selectedLevel" @click="confirmChangeLevel">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([])
const loading = ref(true)
const page = ref(1); const size = ref(10); const total = ref(0)
const keyword = ref(''); const filterStatus = ref('')

const detailVisible = ref(false)
const detail = ref<any>(null)
const detailLoading = ref(false)

const levelDialogVisible = ref(false)
const levelTarget = ref<any>(null)
const levelOptions = ref<any[]>([])
const levelLoading = ref(false)
const selectedLevel = ref('')
const savingLevel = ref(false)

const statusMap: Record<string, { l: string; t: string }> = {
  registering: { l: '注册中', t: 'info' }, pending_review: { l: '待审核', t: 'warning' },
  pending_payment: { l: '待缴费', t: 'danger' },
  active: { l: '正式会员', t: 'success' }, rejected: { l: '已拒绝', t: 'danger' },
  expired: { l: '已过期', t: 'info' }
}
function statusLabel(s: string) { return statusMap[s]?.l || s }
function statusTag(s: string) { return statusMap[s]?.t || 'info' as any }
function typeLabel(t: string) { return t === 'unit' ? '单位会员' : t === 'personal' ? '个人会员' : (t || '-') }
function displayName(m: any) {
  if (!m) return '-'
  return m.member_type === 'unit' ? (m.company_name || m.username || '-') : (m.name || m.username || '-')
}
function initial(m: any) { return ((displayName(m) || '?').trim()[0] || '?').toUpperCase() }
function fmt(d: string) { return d ? d.replace('T', ' ').slice(0, 16) : '-' }
function fileUrl(path?: string) {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  return '/' + path.replace(/^\//, '')
}
function tableRowClassName() { return 'members-row' }

onMounted(() => fetchData())
async function fetchData() {
  loading.value = true
  try {
    const res = await adminApi.getMembers({ page: page.value, size: size.value, keyword: keyword.value, status: filterStatus.value })
    list.value = res.data?.list || []; total.value = res.data?.total || 0
  } catch {} finally { loading.value = false }
}
function search() { page.value = 1; fetchData() }

async function openDetail(row: any) {
  detailLoading.value = true
  detailVisible.value = true
  try {
    const res = await adminApi.getMember(row.id)
    detail.value = res.data || row
  } catch {
    detail.value = row
  } finally { detailLoading.value = false }
}

async function delMember(row: any) {
  if (row.status !== 'registering') {
    ElMessage.warning('仅“注册中”状态的会员才可删除')
    return
  }
  try {
    await ElMessageBox.confirm('确认删除该会员？', '警告', { type: 'warning' })
    await adminApi.deleteMember(row.id); ElMessage.success('已删除'); fetchData()
  } catch {}
}

async function openLevelDialog(row: any) {
  if (row.status !== 'active') {
    ElMessage.warning('仅正式会员可变更等级')
    return
  }
  levelTarget.value = row
  selectedLevel.value = row.member_level || ''
  levelOptions.value = []
  levelDialogVisible.value = true
  levelLoading.value = true
  try {
    const res = await adminApi.getMemberLevelOptions(row.id)
    levelOptions.value = res.data || []
  } catch {} finally { levelLoading.value = false }
}

async function confirmChangeLevel() {
  if (!selectedLevel.value) {
    ElMessage.warning('请选择会员等级')
    return
  }
  if (!levelTarget.value) return
  savingLevel.value = true
  try {
    await adminApi.updateMemberLevel(levelTarget.value.id, selectedLevel.value)
    ElMessage.success('等级变更成功')
    levelDialogVisible.value = false
    fetchData()
  } catch {} finally { savingLevel.value = false }
}
</script>

<style scoped lang="scss">
.admin-members {
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
    color: #1a1a2e;
    margin: 0;
  }
}
.filters { display: flex; gap: 12px; }
.el-card {
  border-radius: 10px;
}
.pagination { display: flex; justify-content: center; padding: 20px 0; }

.members-table :deep(.members-row) {
  cursor: pointer;
}
.members-table :deep(.members-row:hover td) {
  background: #f5f7fa !important;
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
    background: linear-gradient(135deg, #4f8cff, #6f6bff);
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
      .name { font-size: 20px; font-weight: 600; color: #1a1a2e; }
    }
    .sub { margin-top: 6px; color: #909399; font-size: 13px; }
  }

  .detail-body {
    padding-top: 4px;
    .section { margin-bottom: 26px; &:last-child { margin-bottom: 0; } }
    .section-title {
      font-size: 15px;
      font-weight: 600;
      color: #1a1a2e;
      padding-left: 10px;
      border-left: 4px solid #4f8cff;
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
.member-level-dialog {
  .level-hint { margin: 0 0 16px; color: #909399; font-size: 13px; line-height: 1.6; }
  .level-empty { margin-top: 4px; font-size: 12px; color: #f56c6c; }
}
</style>

<template>
  <div class="admin-members" v-loading="loading">
    <div class="page-header">
      <h3>会员管理</h3>
      <div class="filters">
        <el-input v-model="keyword" placeholder="搜索用户名/公司名" clearable style="width:200px" @clear="search" @keyup.enter="search" />
        <el-select v-model="filterStatus" placeholder="状态" clearable style="width:130px" @change="search">
          <el-option label="注册中" value="registering" /><el-option label="待审核" value="pending_review" />
          <el-option label="待缴费" value="pending_payment" /><el-option label="正式会员" value="active" />
          <el-option label="已拒绝" value="rejected" /><el-option label="已过期" value="expired" />
        </el-select>
        <el-select v-model="filterType" placeholder="会员类型" clearable style="width:130px" @change="search">
          <el-option label="单位会员" value="unit" /><el-option label="个人会员" value="personal" />
        </el-select>
        <el-button type="primary" @click="search">查询</el-button>
        <el-button type="success" @click="openCreate">新增会员</el-button>
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
        <el-table-column label="入会机构" min-width="150"><template #default="{row}">{{ row.org_name || '-' }}</template></el-table-column>
        <el-table-column prop="member_level" label="会员等级" min-width="120"><template #default="{row}">{{ levelName(row.member_level) }}</template></el-table-column>
        <el-table-column prop="status" label="状态" min-width="120">
          <template #default="{row}"><el-tag :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" min-width="420">
          <template #default="{row}">
            <el-button text size="small" type="primary" @click.stop="openOrgs(row)">所有会籍</el-button>
            <el-button text size="small" type="warning" :disabled="row.status !== 'active'" @click.stop="openLevelDialog(row)">变更等级</el-button>
            <el-button text size="small" type="primary" @click.stop="openHistory(row)">会籍历史</el-button>
            <el-button text size="small" type="info" @click.stop="resetPassword(row)">重置密码</el-button>
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
            <div class="sub">{{ detail.username }} · {{ typeLabel(detail.member_type) }}<template v-if="detail.member_level"> · 等级 {{ levelName(detail.member_level) }}</template></div>
          </div>
        </div>
      </template>

      <div v-loading="detailLoading" class="detail-body">
        <template v-if="detail">
          <div class="section">
            <div class="section-title">基本资料</div>
            <div class="grid">
              <div class="item"><span class="label">会员类型</span><span class="value">{{ typeLabel(detail.member_type) }}</span></div>
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

    <!-- 变更会员等级弹窗 -->
    <el-dialog v-model="levelDialogVisible" title="变更会员等级" width="460px" class="member-level-dialog" :close-on-click-modal="false" append-to-body>
      <div v-loading="levelLoading">
        <p class="level-hint">仅可为该会员选择其已缴费加入的机构所支持的会员等级。</p>
        <el-form label-width="90px">
          <el-form-item label="当前等级">
            <span>{{ levelName(levelTarget?.member_level) }}</span>
          </el-form-item>
          <el-form-item label="新等级" required>
            <el-select v-model="selectedLevel" placeholder="请选择会员等级" style="width:100%">
              <el-option v-for="lvl in levelOptions" :key="lvl.id" :label="lvl.name" :value="lvl.id" />
            </el-select>
            <div v-if="!levelLoading && !levelOptions.length" class="level-empty">该会员暂无已缴费加入机构可选的会员等级</div>
          </el-form-item>
          <el-form-item label="变更原因" required>
            <el-input v-model="levelReason" type="textarea" :rows="3" maxlength="500" show-word-limit placeholder="请输入变更原因" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="levelDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingLevel" :disabled="!levelOptions.length || !selectedLevel" @click="confirmChangeLevel">确定</el-button>
      </template>
    </el-dialog>

    <!-- 会籍变更历史弹窗 -->
    <el-dialog v-model="historyVisible" :title="`会籍变更历史 - ${historyTarget ? displayName(historyTarget) : ''}`" width="640px" class="member-history-dialog" append-to-body>
      <div ref="historyBodyRef" class="history-body" @scroll="onHistoryScroll">
        <div v-for="group in historyGroups" :key="group.year" class="history-year-group">
          <div class="history-year">{{ group.year }} 年</div>
          <el-timeline>
            <el-timeline-item v-for="item in group.items" :key="item.id" :timestamp="fmt(item.created_at)" placement="top">
              <div class="history-card">
                <div class="history-line">
                  <span class="from">{{ item.old_level_name || '无' }}</span>
                  <el-icon class="arrow"><Right /></el-icon>
                  <span class="to" :class="{ 'is-off': !item.new_level_id }">{{ item.new_level_name || '已失效' }}</span>
                </div>
                <div class="history-meta">
                  <span>入会机构：{{ item.org_name || '-' }}</span>
                  <span>变更人：{{ item.operator || '-' }}</span>
                </div>
                <div class="history-reason" v-if="item.reason">变更原因：{{ item.reason }}</div>
              </div>
            </el-timeline-item>
          </el-timeline>
        </div>
        <el-empty v-if="!historyGroups.length && !historyLoading" description="暂无会籍变更记录" />
        <div v-if="historyLoading" class="history-loading">加载中...</div>
        <div v-else-if="historyList.length > 0 && historyList.length >= historyTotal" class="history-end">已加载全部记录</div>
      </div>
    </el-dialog>

    <!-- 加入组织机构弹窗 -->
    <el-dialog v-model="orgsVisible" :title="`所有会籍 - ${orgsTarget ? displayName(orgsTarget) : ''}`" width="760px" class="member-orgs-dialog" append-to-body>
      <div v-loading="orgsLoading" class="orgs-body">
        <div class="orgs-section">
          <div class="orgs-section-title">
            <span>付费加入</span>
            <el-tag type="success" effect="plain" size="small">{{ orgsInfo.paid.length }}</el-tag>
          </div>
          <el-table v-if="orgsInfo.paid.length" :data="orgsInfo.paid" size="small" class="orgs-table">
            <el-table-column prop="org_name" label="机构名称" min-width="180">
              <template #default="{row}">{{ row.org_name || '-' }}</template>
            </el-table-column>
            <el-table-column label="会员等级" min-width="120">
              <template #default="{row}">{{ row.level_name || '-' }}</template>
            </el-table-column>
            <el-table-column prop="year" label="年度" width="80" />
            <el-table-column label="缴费金额" min-width="120">
              <template #default="{row}">{{ fmtMoney(row.paid_amount > 0 ? row.paid_amount : row.amount) }}</template>
            </el-table-column>
            <el-table-column label="缴费日期" min-width="150">
              <template #default="{row}">{{ fmt(row.paid_date) }}</template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="暂无付费加入记录" image-size="60" />
        </div>

        <div class="orgs-section">
          <div class="orgs-section-title">
            <span>主动加入</span>
            <el-tag type="info" effect="plain" size="small">{{ orgsInfo.voluntary.length }}</el-tag>
          </div>
          <el-table v-if="orgsInfo.voluntary.length" :data="orgsInfo.voluntary" size="small" class="orgs-table">
            <el-table-column prop="org_name" label="机构名称" min-width="200">
              <template #default="{row}">{{ row.org_name || '-' }}</template>
            </el-table-column>
            <el-table-column label="会员等级" min-width="140">
              <template #default="{row}">{{ row.level_name || '-' }}</template>
            </el-table-column>
            <el-table-column label="加入时间" min-width="150">
              <template #default="{row}">{{ fmt(row.joined_at) }}</template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="暂无主动加入记录" image-size="60" />
        </div>
      </div>
    </el-dialog>

    <!-- 新增会员弹窗 -->
    <el-dialog v-model="createVisible" title="新增会员" width="960px" class="member-create-dialog" :close-on-click-modal="false" append-to-body>
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="160px" size="large">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="用户名" prop="username"><el-input v-model="createForm.username" maxlength="32" placeholder="登录用户名" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="密码" prop="password"><el-input v-model="createForm.password" type="password" show-password maxlength="32" placeholder="留空则默认 Abcd@1234" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号" prop="mobile"><el-input v-model="createForm.mobile" maxlength="11" placeholder="请输入11位手机号" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱" prop="email"><el-input v-model="createForm.email" maxlength="64" placeholder="请输入邮箱" /></el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="会员类型">
              <el-radio-group v-model="createForm.member_type">
                <el-radio value="unit">单位会员</el-radio>
                <el-radio value="personal">个人会员</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>

        <template v-if="createForm.member_type === 'unit'">
          <el-divider content-position="left">单位信息</el-divider>
          <el-row :gutter="16">
            <el-col :span="12"><el-form-item label="公司名称" prop="company_name"><el-input v-model="createForm.company_name" maxlength="100" /></el-form-item></el-col>
            <el-col :span="12"><el-form-item label="法定代表人"><el-input v-model="createForm.legal_person" maxlength="32" /></el-form-item></el-col>
            <el-col :span="24"><el-form-item label="统一社会信用代码" prop="credit_code"><el-input v-model="createForm.credit_code" maxlength="18" placeholder="请输入18位统一社会信用代码" /></el-form-item></el-col>
            <el-col :span="12"><el-form-item label="联系人"><el-input v-model="createForm.contact_person" maxlength="32" /></el-form-item></el-col>
            <el-col :span="12"><el-form-item label="联系电话" prop="contact_mobile"><el-input v-model="createForm.contact_mobile" maxlength="20" placeholder="手机号或固定电话" /></el-form-item></el-col>
            <el-col :span="24"><el-form-item label="单位地址"><el-input v-model="createForm.address" maxlength="100" /></el-form-item></el-col>
          </el-row>
        </template>

        <template v-else>
          <el-divider content-position="left">个人信息</el-divider>
          <el-row :gutter="16">
            <el-col :span="12"><el-form-item label="姓名" prop="name"><el-input v-model="createForm.name" maxlength="32" placeholder="请输入姓名" /></el-form-item></el-col>
            <el-col :span="12"><el-form-item label="身份证号" prop="id_card"><el-input v-model="createForm.id_card" maxlength="18" placeholder="请输入18位身份证号" /></el-form-item></el-col>
          </el-row>
        </template>

        <el-divider content-position="left">加入的组织机构</el-divider>
        <el-form-item label="总会" required>
          <el-select v-model="selectedRootOrgId" placeholder="请选择总会" style="width:100%" @change="onRootOrgChange">
            <el-option v-for="org in rootOrgs" :key="org.id" :label="org.name" :value="org.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="会员等级" required>
          <el-select v-model="selectedLevelId" placeholder="请选择该总会默认会员等级" style="width:100%">
            <el-option v-for="lvl in rootLevelOptions" :key="lvl.level_id" :label="lvl.level?.name || lvl.name || ('等级' + lvl.level_id)" :value="lvl.level_id" />
          </el-select>
          <div v-if="selectedRootOrgId && !rootLevelOptions.length" class="root-level-tip">该总会尚未关联会员等级，请先在「组织机构」中配置</div>
        </el-form-item>
        <el-form-item label="其他机构">
          <el-select v-model="selectedOrgIds" multiple :disabled="!selectedRootOrgId" :placeholder="selectedRootOrgId ? '可多选该总会下的分支机构/代表机构，自动沿用总会等级' : '请先选择总会'" style="width:100%">
            <el-option v-for="org in childOrgs" :key="org.id" :label="org.name" :value="org.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="submitCreate">确定新增</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { orgApi } from '@/api/index'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Right } from '@element-plus/icons-vue'

const list = ref<any[]>([])
const levels = ref<any[]>([])
const loading = ref(true)
const page = ref(1); const size = ref(10); const total = ref(0)
const keyword = ref(''); const filterStatus = ref(''); const filterType = ref('')

const detailVisible = ref(false)
const detail = ref<any>(null)
const detailLoading = ref(false)

const levelDialogVisible = ref(false)
const levelTarget = ref<any>(null)
const levelOptions = ref<any[]>([])
const levelLoading = ref(false)
const selectedLevel = ref<number | null>(null)
const levelReason = ref('')
const savingLevel = ref(false)

const historyVisible = ref(false)
const historyTarget = ref<any>(null)
const historyList = ref<any[]>([])
const historyPage = ref(1)
const historySize = 10
const historyTotal = ref(0)
const historyLoading = ref(false)
const historyBodyRef = ref<HTMLElement | null>(null)

const orgsVisible = ref(false)
const orgsTarget = ref<any>(null)
const orgsLoading = ref(false)
const orgsInfo = ref<{ paid: any[]; voluntary: any[] }>({ paid: [], voluntary: [] })

const createVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref()
const orgTree = ref<any[]>([])
const selectedOrgIds = ref<number[]>([])
const selectedRootOrgId = ref<number | null>(null)
const selectedLevelId = ref<number | null>(null)

const rootOrgs = computed(() => orgTree.value.filter((n: any) => (n.parent_id ?? 0) === 0))
const selectedRootOrg = computed(() => rootOrgs.value.find((o: any) => o.id === selectedRootOrgId.value))
const childOrgs = computed(() => selectedRootOrg.value?.children || [])
const rootLevelOptions = computed(() => selectedRootOrg.value?.levels || [])

function onRootOrgChange() {
  selectedLevelId.value = null
  selectedOrgIds.value = []
}

const createForm = reactive<any>({
  username: '',
  password: '',
  mobile: '',
  email: '',
  member_type: 'unit',
  company_name: '',
  credit_code: '',
  legal_person: '',
  contact_person: '',
  contact_mobile: '',
  address: '',
  name: '',
  id_card: ''
})
// 防抖：避免输入过程中频繁请求后端查重
function debounce<A extends any[]>(fn: (...args: A) => void, delay = 400) {
  let timer: ReturnType<typeof setTimeout> | null = null
  return (...args: A) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }
}

// 查重字段 -> 提示名称
const dupFieldLabels: Record<string, string> = {
  username: '用户名',
  mobile: '手机号',
  email: '邮箱',
  contact_mobile: '联系电话',
  company_name: '公司名称',
  credit_code: '统一社会信用代码'
}

// 后端查重（admin）：值已存在则提示；按字段各自防抖，避免多字段共用定时器互相取消
const dupCheckers: Record<string, (value: string, cb: any) => void> = {}
function checkDup(field: string, value: string, cb: any) {
  if (!dupCheckers[field]) {
    dupCheckers[field] = debounce(async (v: string, c: any) => {
      if (!v) return c()
      try {
        const res = await adminApi.checkMemberExists(field, v)
        if (res.data?.exists) c(new Error(`该${dupFieldLabels[field]}已存在`))
        else c()
      } catch {
        c()
      }
    })
  }
  dupCheckers[field](value, cb)
}

// 用户名：必填 + 查重
function validateCreateUsername(_r: any, v: string, cb: any) {
  const val = (v || '').trim()
  if (!val) return cb(new Error('请输入用户名'))
  checkDup('username', val, cb)
}

// 公司名称：选填，填写时查重
function validateCreateCompanyName(_r: any, v: string, cb: any) {
  const val = (v || '').trim()
  if (!val) return cb()
  checkDup('company_name', val, cb)
}

// 手机号：必填 + 11位手机号格式 + 查重
function validateCreateMobile(_r: any, v: string, cb: any) {
  const val = (v || '').trim()
  if (!val) return cb(new Error('请输入手机号'))
  if (!/^1[3-9]\d{9}$/.test(val)) return cb(new Error('请输入合法手机号'))
  checkDup('mobile', val, cb)
}

// 邮箱：必填 + 邮箱格式 + 查重
function validateCreateEmail(_r: any, v: string, cb: any) {
  const val = (v || '').trim()
  if (!val) return cb(new Error('请输入邮箱'))
  if (val.length > 64) return cb(new Error('邮箱不能超过64位'))
  if (!/^[\w.+-]+@[\w-]+(\.[\w-]+)+$/.test(val)) return cb(new Error('请输入合法邮箱'))
  checkDup('email', val, cb)
}

// 联系电话：必填，手机号或固定电话（可带区号/连字符）+ 查重
function validateCreateContactMobile(_r: any, v: string, cb: any) {
  const val = (v || '').trim()
  if (!val) return cb(new Error('请输入联系电话'))
  if (!/^(1[3-9]\d{9}|0\d{2,3}-?\d{7,8})$/.test(val)) return cb(new Error('请输入合法联系电话（手机号或固定电话）'))
  checkDup('contact_mobile', val, cb)
}

// 统一社会信用代码：必填 + GB 32100-2015 校验码算法（与注册/资料页一致）+ 查重
function validateCreateCreditCode(_r: any, v: string, cb: any) {
  const val = (v || '').trim()
  if (!val) return cb(new Error('请输入统一社会信用代码'))
  const code = val.toUpperCase()
  const charSet = '0123456789ABCDEFGHJKLMNPQRTUWXY'
  const weights = [1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28]
  if (!/^[0-9A-HJ-NPQRTUWXY]{2}[0-9]{6}[0-9A-HJ-NPQRTUWXY]{10}$/.test(code)) {
    return cb(new Error('统一社会信用代码格式不正确'))
  }
  let sum = 0
  for (let i = 0; i < 17; i++) sum += charSet.indexOf(code[i]) * weights[i]
  if (charSet[(31 - (sum % 31)) % 31] !== code[17]) {
    return cb(new Error('统一社会信用代码校验不通过'))
  }
  checkDup('credit_code', code, cb)
}

// 身份证号：必填 + 18位/出生日期/校验码（与注册/资料页一致）
function validateCreateIdCard(_r: any, v: string, cb: any) {
  const id = (v || '').trim().toUpperCase()
  if (!id) return cb(new Error('请输入身份证号'))
  if (!/^\d{17}[\dX]$/.test(id)) return cb(new Error('身份证号应为18位，末位可为X'))
  const year = +id.slice(6, 10)
  const month = +id.slice(10, 12)
  const day = +id.slice(12, 14)
  const date = new Date(year, month - 1, day)
  if (date.getFullYear() !== year || date.getMonth() + 1 !== month || date.getDate() !== day) {
    return cb(new Error('身份证号出生日期不合法'))
  }
  const weights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2]
  const codes = '10X98765432'
  let sum = 0
  for (let i = 0; i < 17; i++) sum += +id[i] * weights[i]
  if (codes[sum % 11] !== id[17]) return cb(new Error('身份证号校验不通过'))
  cb()
}

const createRules = {
  username: [{ required: true, validator: validateCreateUsername, trigger: 'blur' }],
  password: [{ min: 6, message: '密码至少6位', trigger: 'blur' }],
  mobile: [{ required: true, validator: validateCreateMobile, trigger: 'blur' }],
  email: [{ required: true, validator: validateCreateEmail, trigger: 'blur' }],
  company_name: [{ validator: validateCreateCompanyName, trigger: 'blur' }],
  // 以下四项仅对应会员类型渲染，字段未渲染时不会参与校验
  credit_code: [{ required: true, validator: validateCreateCreditCode, trigger: 'blur' }],
  contact_mobile: [{ required: true, validator: validateCreateContactMobile, trigger: 'blur' }],
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  id_card: [{ required: true, validator: validateCreateIdCard, trigger: 'blur' }]
}

const statusMap: Record<string, { l: string; t: string }> = {
  registering: { l: '注册中', t: 'info' }, pending_review: { l: '待审核', t: 'warning' },
  pending_payment: { l: '待缴费', t: 'danger' },
  active: { l: '正式会员', t: 'success' }, rejected: { l: '已拒绝', t: 'danger' },
  expired: { l: '已过期', t: 'info' }
}
function statusLabel(s: string) { return statusMap[s]?.l || s }
function statusTag(s: string) { return statusMap[s]?.t || 'info' as any }
function typeLabel(t: string) { return t === 'unit' ? '单位会员' : t === 'personal' ? '个人会员' : (t || '-') }
function levelName(id: any) {
  if (id === null || id === undefined || id === '') return '-'
  const lvl = levels.value.find((l: any) => String(l.id) === String(id))
  return lvl ? lvl.name : String(id)
}
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

onMounted(() => { fetchData(); fetchLevels() })
async function fetchLevels() {
  try {
    const res = await adminApi.getMemberLevels()
    levels.value = res.data || []
  } catch {}
}
async function fetchData() {
  loading.value = true
  try {
    const res = await adminApi.getMembers({ page: page.value, size: size.value, keyword: keyword.value, status: filterStatus.value, member_type: filterType.value })
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

async function resetPassword(row: any) {
  try {
    await ElMessageBox.confirm(
      `将重置「${displayName(row)}」的登录密码为默认密码 Abcd@1234，重置后请提醒会员及时修改密码。确认继续？`,
      '重置密码',
      { type: 'warning', confirmButtonText: '确定重置', cancelButtonText: '取消' }
    )
    await adminApi.resetMemberPassword(row.id)
    ElMessage.success('密码已重置为 Abcd@1234，请提醒会员及时修改密码')
  } catch {}
}

function openCreate() {
  Object.assign(createForm, {
    username: '',
    password: '',
    mobile: '',
    email: '',
    member_type: 'unit',
    company_name: '',
    credit_code: '',
    legal_person: '',
    contact_person: '',
    contact_mobile: '',
    address: '',
    name: '',
    id_card: ''
  })
  selectedOrgIds.value = []
  selectedRootOrgId.value = null
  selectedLevelId.value = null
  createVisible.value = true
  loadOrgTree()
}

async function loadOrgTree() {
  try {
    const res = await orgApi.getTree()
    orgTree.value = res.data || []
  } catch {
    orgTree.value = []
  }
  // 默认选中第一个总会
  if (!selectedRootOrgId.value && orgTree.value.length) {
    selectedRootOrgId.value = orgTree.value[0].id
  }
}

async function submitCreate() {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return
  if (!selectedRootOrgId.value) {
    ElMessage.warning('请选择总会')
    return
  }
  if (!selectedLevelId.value) {
    ElMessage.warning('请选择总会默认会员等级')
    return
  }
  createLoading.value = true
  try {
    await adminApi.createMember({
      ...createForm,
      mobile: createForm.mobile.trim(),
      email: createForm.email.trim(),
      contact_mobile: (createForm.contact_mobile || '').trim(),
      credit_code: createForm.member_type === 'unit' ? (createForm.credit_code || '').trim().toUpperCase() : '',
      name: (createForm.name || '').trim(),
      id_card: createForm.member_type === 'personal' ? (createForm.id_card || '').trim().toUpperCase() : '',
      password: createForm.password || undefined,
      root_org_id: selectedRootOrgId.value,
      org_ids: selectedOrgIds.value,
      level_id: selectedLevelId.value
    })
    ElMessage.success('新增会员成功')
    createVisible.value = false
    fetchData()
  } catch {} finally { createLoading.value = false }
}

async function openLevelDialog(row: any) {
  if (row.status !== 'active') {
    ElMessage.warning('仅正式会员可变更等级')
    return
  }
  levelTarget.value = row
  selectedLevel.value = null
  levelReason.value = ''
  levelOptions.value = []
  levelDialogVisible.value = true
  levelLoading.value = true
  try {
    const res = await adminApi.getMemberLevelOptions(row.id)
    levelOptions.value = res.data || []
    const current = levelOptions.value.find((lvl: any) => String(lvl.id) === String(row.member_level))
    if (current) selectedLevel.value = current.id
  } catch {} finally { levelLoading.value = false }
}

async function confirmChangeLevel() {
  if (!selectedLevel.value) {
    ElMessage.warning('请选择会员等级')
    return
  }
  if (!levelTarget.value) return
  const current = levelOptions.value.find((lvl: any) => String(lvl.id) === String(levelTarget.value.member_level))
  if (selectedLevel.value === current?.id) {
    ElMessage.warning('新旧会员等级不能相同')
    return
  }
  if (!levelReason.value.trim()) {
    ElMessage.warning('请填写变更原因')
    return
  }
  savingLevel.value = true
  try {
    await adminApi.updateMemberLevel(levelTarget.value.id, selectedLevel.value, levelReason.value.trim())
    ElMessage.success('等级变更成功')
    levelDialogVisible.value = false
    fetchData()
  } catch {} finally { savingLevel.value = false }
}

async function openHistory(row: any) {
  historyTarget.value = row
  historyList.value = []
  historyPage.value = 1
  historyTotal.value = 0
  historyVisible.value = true
  await fetchHistory()
}

// 会籍历史按变更年份分组（接口按时间倒序返回，分组后年份自然从新到旧）
const historyGroups = computed(() => {
  const groups: { year: number; items: any[] }[] = []
  const map = new Map<number, { year: number; items: any[] }>()
  for (const item of historyList.value) {
    const year = Number(item.change_year) || 0
    let g = map.get(year)
    if (!g) { g = { year, items: [] }; map.set(year, g); groups.push(g) }
    g.items.push(item)
  }
  return groups
})

async function fetchHistory() {
  if (historyLoading.value) return
  if (historyList.value.length >= historyTotal.value && historyPage.value > 1) return
  historyLoading.value = true
  try {
    const res = await adminApi.getMemberLevelChangesByMember(historyTarget.value.id, {
      page: historyPage.value,
      size: historySize
    })
    const list = res.data?.list || []
    historyTotal.value = res.data?.total || 0
    historyList.value = historyList.value.concat(list)
    historyPage.value += 1
  } catch {} finally { historyLoading.value = false }
}

function onHistoryScroll() {
  const el = historyBodyRef.value
  if (!el) return
  if (el.scrollTop + el.clientHeight >= el.scrollHeight - 60) {
    fetchHistory()
  }
}

async function openOrgs(row: any) {
  orgsTarget.value = row
  orgsInfo.value = { paid: [], voluntary: [] }
  orgsVisible.value = true
  orgsLoading.value = true
  try {
    const res = await adminApi.getMemberOrgs(row.id)
    orgsInfo.value = res.data || { paid: [], voluntary: [] }
  } catch {} finally { orgsLoading.value = false }
}

function fmtMoney(n: any) {
  if (n === null || n === undefined || n === '') return '-'
  const num = Number(n)
  if (!isFinite(num)) return '-'
  return '¥' + num.toFixed(2)
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
    color: #1d2739;
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
  background: #f5f7fb !important;
}
.root-org-name {
  color: #303133;
  line-height: 32px;
}
.root-org-tip {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}
.root-level-tip {
  margin-top: 4px;
  font-size: 12px;
  color: #f56c6c;
}
</style>

<!-- 详情弹窗样式（dialog 使用 teleport，故使用非 scoped 样式并以类名限定作用域） -->
<style lang="scss">
/* 新增会员弹窗：标签不换行（dialog teleport 到 body，需非 scoped 样式） */
.member-create-dialog {
  .el-form-item__label {
    white-space: nowrap;
  }
}

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
.member-level-dialog {
  .level-hint { margin: 0 0 16px; color: #909399; font-size: 13px; line-height: 1.6; }
  .level-empty { margin-top: 4px; font-size: 12px; color: #f56c6c; }
}

.member-history-dialog {
  border-radius: 14px;
  overflow: hidden;

  .history-body {
    max-height: 460px;
    overflow-y: auto;
    padding: 4px 8px 4px 0;

    .history-year-group {
      .history-year {
        margin: 4px 0 12px;
        font-size: 13px;
        font-weight: 600;
        color: #303133;
      }
      & + .history-year-group { margin-top: 10px; }
    }

    .history-card {
      background: #fafbfc;
      border: 1px solid #eef0f3;
      border-radius: 10px;
      padding: 12px 16px;

      .history-line {
        display: flex;
        align-items: center;
        gap: 10px;
        font-size: 15px;
        font-weight: 600;
        .from { color: #909399; }
        .arrow { color: #c0c4cc; }
        .to {
          color: #e6a23c;
          &.is-off { color: #909399; }
        }
      }
      .history-meta {
        display: flex;
        flex-wrap: wrap;
        gap: 6px 18px;
        margin-top: 10px;
        font-size: 13px;
        color: #606266;
      }
      .history-reason {
        margin-top: 10px;
        padding: 8px 10px;
        background: #fff7e6;
        border-radius: 6px;
        font-size: 13px;
        color: #7a5b1a;
        line-height: 1.6;
        word-break: break-word;
      }
    }

    .history-loading, .history-end {
      text-align: center;
      color: #909399;
      font-size: 13px;
      padding: 14px 0;
    }
  }
}

.member-orgs-dialog {
  border-radius: 14px;
  overflow: hidden;

  .orgs-body {
    min-height: 200px;
  }
  .orgs-section {
    margin-bottom: 24px;
    &:last-child { margin-bottom: 0; }
  }
  .orgs-section-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 15px;
    font-weight: 600;
    color: #1d2739;
    padding-left: 10px;
    border-left: 4px solid #002fa7;
    margin-bottom: 12px;
    line-height: 1.2;
  }
  .orgs-table {
    width: 100%;
  }
}
</style>

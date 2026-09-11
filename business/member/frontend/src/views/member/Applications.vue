<template>
  <div class="applications-page" v-loading="loading">
    <div class="page-header">
      <h3>我的申请</h3>
      <el-button type="primary" @click="handleCreateApp" v-if="!hasPending">发起入会申请</el-button>
    </div>

    <!-- Application List -->
    <el-card v-if="applications.length">
      <el-table :data="applications" stripe>
        <el-table-column prop="id" label="编号" width="80" />
        <el-table-column prop="org.name" label="申请分会" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="review_comment" label="审核意见" />
        <el-table-column prop="created_at" label="申请时间" width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button v-if="row.status === 'pending_review'" text type="danger" @click="withdrawApp(row)">撤回</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-empty v-else description="暂无申请记录" />

    <!-- Create Application Dialog -->
    <el-dialog v-model="showCreate" title="发起入会申请" width="640px" :close-on-click-modal="false">
      <el-steps :active="createStep" align-center style="margin-bottom:28px">
        <el-step title="填写信息" />
        <el-step title="确认资料" />
        <el-step title="上传文件" />
      </el-steps>

      <!-- Step 1: Fill Info -->
      <el-form v-if="createStep === 0" :model="appForm" label-width="100px" size="large">
        <el-form-item label="申请入会" required>
          <el-select v-model="appForm.orgId" placeholder="选择总会或分会" style="width:100%">
            <el-option v-for="org in orgs" :key="org.id" :label="org.name" :value="org.id" />
          </el-select>
          <div class="charter-hint">
            <el-link type="primary" :icon="Download" @click="downloadCharter" :underline="false">
              入会章程
            </el-link>
          </div>
          <el-alert
            v-if="isBranchSelected"
            class="join-root-hint"
            type="info"
            show-icon
            :closable="false"
            title="加入任何一个分会默认都会加入总会"
          />
        </el-form-item>
        <template v-if="memberType === 'unit'">
        <el-form-item label="单位名称"><el-input v-model="appForm.companyName" /></el-form-item>
        <el-form-item label="信用代码"><el-input v-model="appForm.creditCode" /></el-form-item>
        <el-form-item label="联系人"><el-input v-model="appForm.contactPerson" /></el-form-item>
        <el-form-item label="单位地址"><el-input v-model="appForm.address" /></el-form-item>
        </template>
      </el-form>

      <!-- Step 2: Confirm & Download -->
      <div v-if="createStep === 1" class="confirm-section">
        <el-alert title="请确认填写信息是否正确" type="info" show-icon :closable="false" style="margin-bottom:20px" />
        <el-alert
          v-if="isBranchSelected"
          class="join-root-hint"
          type="success"
          show-icon
          :closable="false"
          title="加入任何一个分会默认都会加入总会"
          style="margin-bottom:20px"
        />
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="申请入会">{{ selectedOrgName }}</el-descriptions-item>
          <template v-if="memberType === 'unit'">
          <el-descriptions-item label="单位名称">{{ appForm.companyName || '未填写' }}</el-descriptions-item>
          <el-descriptions-item label="信用代码">{{ appForm.creditCode || '未填写' }}</el-descriptions-item>
          <el-descriptions-item label="联系人">{{ appForm.contactPerson || '未填写' }}</el-descriptions-item>
          <el-descriptions-item label="单位地址">{{ appForm.address || '未填写' }}</el-descriptions-item>
          </template>
        </el-descriptions>
        <div class="download-area">
          <p class="step-tip">确认信息无误后，请下载入会申请表</p>
          <el-button type="primary" :icon="Download" @click="downloadTemplate">下载会员申请表</el-button>
          <p class="step-desc">将下载好的申请表打印，在相应位置签字盖章</p>
        </div>
      </div>

      <!-- Step 3: Upload & Submit -->
      <div v-if="createStep === 2" class="upload-section">
        <el-alert
          title="请在下方上传签字盖章的申请表电子版（PDF扫描版优先，如照片请保证清晰度）"
          type="warning"
          show-icon
          :closable="false"
          style="margin-bottom:16px"
        />
        <div class="upload-area">
          <template v-if="!uploadedFile">
            <el-upload
              drag
              :action="uploadUrl"
              :headers="uploadHeaders"
              :on-success="handleUploadSuccess"
              :on-error="handleUploadError"
              accept=".pdf,.jpg,.jpeg,.png,.zip,.rar"
            >
              <el-icon class="el-icon--upload" :size="48"><UploadFilled /></el-icon>
              <div class="el-upload__text">拖拽文件到此处，或<em>点击选择文件</em></div>
              <template #tip>
                <div class="el-upload__tip">
                  支持 PDF、JPG、PNG 格式，多页扫描件可打包为 ZIP/RAR
                </div>
              </template>
            </el-upload>
          </template>
          <template v-else>
            <div class="uploaded-file">
              <el-icon :size="40" color="#22c55e"><CircleCheckFilled /></el-icon>
              <div class="file-info">
                <p class="file-name">{{ uploadedFile.name }}</p>
                <p class="file-url">{{ uploadedFile.url }}</p>
              </div>
              <el-button text type="primary" @click="removeFile">重新上传</el-button>
            </div>
          </template>
        </div>
        <p class="step-note">注：如多页扫描件可以放到压缩包里上传</p>
      </div>

      <template #footer>
        <el-button v-if="createStep > 0" @click="createStep--">上一步</el-button>
        <el-button v-if="createStep < 2" type="primary" @click="createStep++">下一步</el-button>
        <el-button v-if="createStep === 2" type="primary" :loading="submitting" @click="submitApp">提交申请</el-button>
        <el-button @click="closeDialog">取消</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { applicationApi, orgApi } from '@/api/index'
import { authApi } from '@/api/auth'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const applications = ref<any[]>([])
const orgs = ref<any[]>([])
const loading = ref(true)
const showCreate = ref(false)
const submitting = ref(false)
const createStep = ref(0)
const uploadedFile = ref<{ name: string; url: string } | null>(null)
const memberType = ref('unit')

const appForm = reactive({ orgId: null as number | null, companyName: '', creditCode: '', contactPerson: '', address: '' })

// 打开对话框时自动读取会员档案信息，预填单位信息
watch(showCreate, async (val) => {
  if (val) {
    try {
      const res = await authApi.getProfile()
      if (res.data) {
        memberType.value = res.data.member_type || 'unit'
        appForm.companyName = res.data.company_name || ''
        appForm.creditCode = res.data.credit_code || ''
        appForm.contactPerson = res.data.contact_person || ''
        appForm.address = res.data.address || ''
      }
    } catch {
      // 读取失败则留空让用户手动填写
    }
  }
})

const uploadUrl = computed(() => `${import.meta.env.VITE_API_BASE_URL || '/member/api'}/upload`)
const uploadHeaders = computed(() => ({ Authorization: `Bearer ${userStore.token}` }))

const hasPending = computed(() => applications.value.some((a: any) => !['rejected', 'draft'].includes(a.status)))
const selectedOrg = computed(() => orgs.value.find((o: any) => o.id === appForm.orgId) || null)
const selectedOrgName = computed(() => selectedOrg.value?.name || '')
// 总会为一级组织（parent_id 为 0），分会/代表机构为其下级组织
const isBranchSelected = computed(() => !!selectedOrg.value && !!selectedOrg.value.parent_id)

const statusMap: Record<string, { label: string; tag: string }> = {
  draft: { label: '草稿', tag: 'info' },
  pending_review: { label: '待审核', tag: 'warning' },
  approved: { label: '已通过', tag: 'success' },
  rejected: { label: '已拒绝', tag: 'danger' }
}

function statusLabel(s: string) { return statusMap[s]?.label || s }
function statusTag(s: string) { return statusMap[s]?.tag || 'info' as any }

onMounted(async () => {
  try {
    const [apps, orgData] = await Promise.all([applicationApi.getMyApplications(), orgApi.getTree()])
    applications.value = apps.data || []
    orgs.value = flattenOrgs(orgData.data || [])
  } catch {} finally { loading.value = false }
})

function flattenOrgs(nodes: any[]): any[] {
  let r: any[] = []
  for (const n of nodes) { r.push(n); if (n.children) r = r.concat(flattenOrgs(n.children)) }
  return r
}

async function handleCreateApp() {
  try {
    const res = await authApi.getProfile()
    const profile = res.data
    if (profile.member_type === 'unit' && !profile.cert_file) {
      ElMessage.warning('请先上传组织机构代码证，补全会员信息后再发起申请')
      router.push('/member/profile')
      return
    }
    if (profile.member_type !== 'unit' && (!profile.name || !profile.id_card)) {
      ElMessage.warning('请先补充完整的姓名和身份证号，补全会员信息后再发起申请')
      router.push('/member/profile')
      return
    }
    showCreate.value = true
  } catch {
    ElMessage.error('获取会员信息失败，请重试')
  }
}

async function withdrawApp(row: any) {
  try {
    await ElMessageBox.confirm('确定要撤回该入会申请吗？', '确认撤回', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await applicationApi.withdraw(row.id)
    ElMessage.success('申请已撤回')
    const res = await applicationApi.getMyApplications()
    applications.value = res.data || []
  } catch {}
}

function closeDialog() {
  showCreate.value = false
  createStep.value = 0
  uploadedFile.value = null
  Object.assign(appForm, { orgId: null, companyName: '', creditCode: '', contactPerson: '', address: '' })
}

function downloadTemplate() {
  const baseUrl = import.meta.env.VITE_API_BASE_URL || '/member/api'
  window.open(`${baseUrl}/application-template`, '_blank')
}

function downloadCharter() {
  const baseUrl = import.meta.env.VITE_API_BASE_URL || '/member/api'
  window.open(`${baseUrl}/charter`, '_blank')
}

function handleUploadSuccess(response: any) {
  if (response.code === 0 || response.code === 200) {
    uploadedFile.value = { name: response.data?.name || '申请表', url: response.data?.url || '' }
    ElMessage.success('文件上传成功')
  } else {
    ElMessage.error(response.message || '上传失败')
  }
}

function handleUploadError() {
  ElMessage.error('文件上传失败，请重试')
}

function removeFile() {
  uploadedFile.value = null
}

async function submitApp() {
  if (!appForm.orgId) { ElMessage.warning('请选择申请分会'); return }
  if (!uploadedFile.value) { ElMessage.warning('请上传签字盖章的申请表'); return }
  submitting.value = true
  try {
    await applicationApi.createApplication({
      org_id: appForm.orgId,
      company_name: appForm.companyName,
      credit_code: appForm.creditCode,
      contact_person: appForm.contactPerson,
      address: appForm.address,
      signed_file: uploadedFile.value.url
    })
    ElMessage.success('入会申请提交成功！')
    closeDialog()
    const res = await applicationApi.getMyApplications()
    applications.value = res.data || []
  } catch {} finally { submitting.value = false }
}

function formatDate(d: string) { return d ? d.slice(0, 16).replace('T', ' ') : '' }
</script>

<style scoped lang="scss">
.applications-page { width: 100%;}
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }

.charter-hint {
  margin-top: 6px;
  padding-left: 2px;
}

.join-root-hint {
  margin-top: 10px;
  width: 100%;
}

.confirm-section {
  .download-area {
    margin-top: 24px;
    padding: 24px;
    background: #f8fafc;
    border: 2px dashed #d1d5db;
    border-radius: 12px;
    text-align: center;
    .step-tip {
      color: #374151;
      font-size: 15px;
      margin-bottom: 16px;
    }
    .step-desc {
      color: #9ca3af;
      font-size: 13px;
      margin-top: 16px;
    }
  }
}

.upload-section {
  .upload-area {
    .uploaded-file {
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 20px;
      background: #f0fdf4;
      border: 1px solid #bbf7d0;
      border-radius: 12px;
      .file-info {
        flex: 1;
        .file-name { font-weight: 600; color: #16a34a; margin-bottom: 4px; }
        .file-url { font-size: 12px; color: #9ca3af; }
      }
    }
  }
  .step-note {
    color: #9ca3af;
    font-size: 13px;
    margin-top: 12px;
    text-align: center;
  }
}
</style>

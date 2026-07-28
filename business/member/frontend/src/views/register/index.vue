<template>
  <div class="register-page">
    <div class="register-card">
      <div class="card-header">
        <h2>入会申请</h2>
        <el-steps :active="step" align-center>
          <el-step title="账户注册" />
          <el-step title="填写资料" />
          <el-step title="确认提交" />
        </el-steps>
      </div>

      <!-- Step 1: Account -->
      <el-form v-if="step === 0" ref="form1Ref" :model="form1" :rules="rules1" size="large" label-width="100px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form1.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form1.password" type="password" placeholder="至少6位" show-password />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPwd">
          <el-input v-model="form1.confirmPwd" type="password" placeholder="再次输入密码" show-password />
        </el-form-item>
        <el-form-item label="手机号" prop="mobile">
          <el-input v-model="form1.mobile" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form1.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="会员类型">
          <el-radio-group v-model="form1.memberType">
            <el-radio value="unit">单位会员</el-radio>
            <el-radio value="personal">个人会员</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="验证码" prop="captchaCode">
          <div class="captcha-row">
            <el-input v-model="form1.captchaCode" placeholder="验证码" />
            <img :src="captchaImage" class="captcha-img" @click="loadCaptcha" />
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="nextStep">下一步</el-button>
        </el-form-item>
      </el-form>

      <!-- Step 2: Profile -->
      <el-form v-if="step === 1" ref="form2Ref" :model="form2" size="large" label-width="100px">
        <el-form-item label="单位名称" prop="companyName">
          <el-input v-model="form2.companyName" placeholder="请输入单位全称" />
        </el-form-item>
        <el-form-item label="信用代码">
          <el-input v-model="form2.creditCode" placeholder="统一社会信用代码" />
        </el-form-item>
        <el-form-item label="法定代表人">
          <el-input v-model="form2.legalPerson" placeholder="法定代表人姓名" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="form2.contactPerson" placeholder="联系人姓名" />
        </el-form-item>
        <el-form-item label="单位地址">
          <el-input v-model="form2.address" placeholder="单位详细地址" />
        </el-form-item>
        <el-form-item label="申请分会">
          <el-select v-model="form2.orgId" placeholder="选择要加入的分会" style="width:100%">
            <el-option v-for="org in orgs" :key="org.id" :label="org.name" :value="org.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button @click="step = 0">上一步</el-button>
          <el-button type="primary" @click="nextStep">下一步</el-button>
        </el-form-item>
      </el-form>

      <!-- Step 3: Confirm -->
      <div v-if="step === 2" class="confirm-step">
        <el-descriptions title="请确认入会信息" :column="1" border>
          <el-descriptions-item label="用户名">{{ form1.username }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ form1.mobile }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ form1.email }}</el-descriptions-item>
          <el-descriptions-item label="会员类型">{{ form1.memberType === 'unit' ? '单位会员' : '个人会员' }}</el-descriptions-item>
          <el-descriptions-item label="单位名称">{{ form2.companyName }}</el-descriptions-item>
          <el-descriptions-item label="联系人">{{ form2.contactPerson }}</el-descriptions-item>
          <el-descriptions-item label="申请分会">{{ selectedOrgName }}</el-descriptions-item>
        </el-descriptions>
        <div class="confirm-actions">
          <el-button @click="step = 1">上一步</el-button>
          <el-button type="primary" :loading="submitting" @click="submitApplication">提交入会申请</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '@/api/auth'
import { orgApi } from '@/api/index'
import { ElMessage } from 'element-plus'

const router = useRouter()
const step = ref(0)
const submitting = ref(false)
const captchaImage = ref('')
const captchaId = ref('')
const orgs = ref<any[]>([])
const form1Ref = ref()
const form2Ref = ref()

const form1 = reactive({
  username: '', password: '', confirmPwd: '', mobile: '', email: '',
  memberType: 'unit', captchaCode: ''
})

const validatePass = (_rule: any, value: string, callback: any) => {
  if (value !== form1.password) callback(new Error('两次密码不一致'))
  else callback()
}

const rules1 = {
  username: [{ required: true, message: '请输入用户名' }],
  password: [{ required: true, min: 6, message: '密码至少6位' }],
  confirmPwd: [{ required: true, validator: validatePass, trigger: 'blur' }],
  mobile: [{ required: true, message: '请输入手机号' }],
  email: [{ type: 'email', message: '邮箱格式不正确' }],
  captchaCode: [{ required: true, message: '请输入验证码' }]
}

const form2 = reactive({
  companyName: '', creditCode: '', legalPerson: '',
  contactPerson: '', address: '', orgId: null as number | null
})

const selectedOrgName = computed(() => {
  const org = orgs.value.find((o: any) => o.id === form2.orgId)
  return org?.name || ''
})

onMounted(async () => {
  loadCaptcha()
  try {
    const res = await orgApi.getTree()
    orgs.value = flattenOrgs(res.data || [])
  } catch {}
})

function flattenOrgs(nodes: any[]): any[] {
  let result: any[] = []
  for (const n of nodes) {
    result.push({ id: n.id, name: n.name })
    if (n.children) result = result.concat(flattenOrgs(n.children))
  }
  return result
}

async function loadCaptcha() {
  try {
    const res = await authApi.getCaptcha()
    captchaId.value = res.data.captcha_id
    captchaImage.value = res.data.captcha_image
  } catch {}
}

async function nextStep() {
  if (step.value === 0) {
    const valid = await form1Ref.value?.validate().catch(() => false)
    if (!valid) return
    // Register and get token directly
    try {
      const res = await authApi.register({
        username: form1.username,
        password: form1.password,
        mobile: form1.mobile,
        email: form1.email,
        member_type: form1.memberType,
        captcha_id: captchaId.value,
        captcha_code: form1.captchaCode
      })
      // Use the token returned from register, no need to call login again
      const { token, member } = res.data
      if (token) {
        localStorage.setItem('member-token', token)
      }
      ElMessage.success('注册成功')
      step.value = 1
    } catch {
      loadCaptcha()
    }
  } else if (step.value === 1) {
    step.value = 2
  }
}

async function submitApplication() {
  submitting.value = true
  try {
    // Use application API to submit
    const { applicationApi } = await import('@/api/index')
    await applicationApi.createApplication({
      org_id: form2.orgId,
      form_data: JSON.stringify(form2),
      company_name: form2.companyName,
      credit_code: form2.creditCode,
      legal_person: form2.legalPerson,
      contact_person: form2.contactPerson,
      address: form2.address,
      member_type: form1.memberType
    })
    ElMessage.success('入会申请提交成功！')
    router.push('/member/dashboard')
  } catch {} finally {
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
.register-page {
  min-height: calc(100vh - 200px);
  display: flex;
  justify-content: center;
  padding: 40px;
  background: linear-gradient(135deg, #f5f7fa, #e5e7eb);
}
.register-card {
  width: 680px;
  background: #fff;
  border-radius: 16px;
  padding: 48px 48px 36px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08);
  .card-header {
    h2 { text-align: center; margin-bottom: 32px; font-size: 24px; }
    margin-bottom: 36px;
  }
  .captcha-row {
    display: flex; gap: 12px;
    .captcha-img { width: 120px; height: 40px; border-radius: 8px; cursor: pointer; border: 1px solid #e5e7eb; }
  }
  .confirm-step {
    .confirm-actions { display: flex; justify-content: center; gap: 16px; margin-top: 32px; }
  }
}
</style>

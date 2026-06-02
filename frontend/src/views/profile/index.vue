<template>
  <div class="page-container">
    <el-row :gutter="20">
      <el-col :xs="24" :md="8">
        <el-card shadow="hover" class="profile-card" v-loading="profileLoading">
          <div class="profile-header">
            <el-avatar :size="100" :src="profile.avatar || defaultAvatar" />
            <h3>{{ profile.nickname || profile.username }}</h3>
            <p>{{ profile.username }}</p>
            <el-tag type="primary">{{ profile.roleName || '用户' }}</el-tag>
          </div>
          <div class="profile-stats">
            <div class="stat-item">
              <div class="stat-num">{{ profile.articleCount }}</div>
              <div class="stat-label">发布文章</div>
            </div>
            <div class="stat-item">
              <div class="stat-num">{{ profile.operationCount }}</div>
              <div class="stat-label">操作次数</div>
            </div>
            <div class="stat-item">
              <div class="stat-num">{{ profile.onlineDays }}</div>
              <div class="stat-label">在线天数</div>
            </div>
          </div>
          <div class="profile-info">
            <div class="info-item">
              <el-icon><User /></el-icon>
              <span>{{ profile.roleName || '用户' }}</span>
            </div>
            <div class="info-item">
              <el-icon><Message /></el-icon>
              <span>{{ profile.email || '-' }}</span>
            </div>
            <div class="info-item">
              <el-icon><Phone /></el-icon>
              <span>{{ profile.phone || '-' }}</span>
            </div>
            <div class="info-item">
              <el-icon><Location /></el-icon>
              <span>北京市朝阳区</span>
            </div>
            <div class="info-item">
              <el-icon><Clock /></el-icon>
              <span>注册于 {{ profile.createdAt }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="16">
        <el-card shadow="hover" class="edit-card">
          <template #header>
            <div class="card-header">基本信息</div>
          </template>
          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            label-width="100px"
            class="profile-form"
          >
            <el-form-item label="用户昵称" prop="nickname">
              <el-input v-model="form.nickname" placeholder="请输入昵称" />
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="form.email" placeholder="请输入邮箱" />
            </el-form-item>
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="form.phone" placeholder="请输入手机号" />
            </el-form-item>
            <el-form-item label="个人简介" prop="bio">
              <el-input v-model="form.bio" type="textarea" :rows="4" placeholder="请输入个人简介" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="submitLoading" @click="handleSubmit">保存修改</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="hover" class="edit-card" style="margin-top: 20px;">
          <template #header>
            <div class="card-header">修改密码</div>
          </template>
          <el-form
            ref="pwdFormRef"
            :model="pwdForm"
            :rules="pwdRules"
            label-width="100px"
          >
            <el-form-item label="原密码" prop="oldPassword">
              <el-input v-model="pwdForm.oldPassword" type="password" show-password clearable placeholder="请输入原密码" maxlength="20" />
            </el-form-item>
            <el-form-item label="新密码" prop="newPassword">
              <el-input v-model="pwdForm.newPassword" type="password" show-password clearable placeholder="请输入新密码" maxlength="20" />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirmPassword">
              <el-input v-model="pwdForm.confirmPassword" type="password" show-password clearable placeholder="请再次输入新密码" maxlength="20" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="pwdLoading" @click="handleChangePassword">确认修改</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { User, Message, Phone, Location, Clock } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { getUserInfo, updateProfile, changePassword } from '@/api/auth'
import type { ProfileUser } from '@/api/auth'

const userStore = useUserStore()
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const profileLoading = ref(false)
const profile = reactive<ProfileUser>({
  id: 0,
  username: '',
  nickname: '',
  email: '',
  phone: '',
  account: '',
  roleName: '',
  avatar: '',
  createdAt: '',
  onlineDays: 0,
  articleCount: 0,
  operationCount: 0,
  bio: ''
})

const formRef = ref()
const pwdFormRef = ref()
const submitLoading = ref(false)
const pwdLoading = ref(false)

const form = reactive({
  nickname: '',
  email: '',
  phone: '',
  bio: ''
})

const rules = {
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }]
}

const pwdForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const pwdRules = {
  oldPassword: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: Function) => {
        if (value !== pwdForm.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    await updateProfile({
      nickname: form.nickname,
      email: form.email,
      phone: form.phone,
      bio: form.bio
    })
    ElMessage.success('保存成功')
    await loadProfile()
  } catch {
    // request interceptor 已处理错误提示
  } finally {
    submitLoading.value = false
  }
}

const handleChangePassword = async () => {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return
  pwdLoading.value = true
  try {
    await changePassword({
      oldPassword: pwdForm.oldPassword,
      newPassword: pwdForm.newPassword
    })
    ElMessage.success('密码修改成功')
    pwdForm.oldPassword = ''
    pwdForm.newPassword = ''
    pwdForm.confirmPassword = ''
  } catch {
    // request interceptor 已处理错误提示
  } finally {
    pwdLoading.value = false
  }
}

const loadProfile = async () => {
  profileLoading.value = true
  try {
    const res: any = await getUserInfo()
    const data = res.data.user as ProfileUser
    Object.assign(profile, data)
    form.nickname = data.nickname || ''
    form.email = data.email || ''
    form.phone = data.phone || ''
    form.bio = data.bio || ''
    if (userStore.userInfo) {
      userStore.setUserInfo({
        ...userStore.userInfo,
        nickname: data.nickname || data.username
      })
    }
  } catch {
    ElMessage.error('获取用户信息失败')
  } finally {
    profileLoading.value = false
  }
}

onMounted(() => {
  loadProfile()
})
</script>

<style scoped lang="scss">
.page-container {
  .profile-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .profile-header {
      text-align: center;
      padding: 20px 0;
      border-bottom: 1px solid #f0f7ff;

      h3 {
        margin: 12px 0 4px;
        font-size: 18px;
        color: #2c3e50;
      }

      p {
        color: #909399;
        font-size: 13px;
        margin-bottom: 10px;
      }
    }

    .profile-stats {
      display: flex;
      justify-content: space-around;
      padding: 20px 0;
      border-bottom: 1px solid #f0f7ff;

      .stat-item {
        text-align: center;

        .stat-num {
          font-size: 20px;
          font-weight: 700;
          color: #409eff;
        }

        .stat-label {
          font-size: 12px;
          color: #909399;
          margin-top: 4px;
        }
      }
    }

    .profile-info {
      padding: 16px 20px;

      .info-item {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 10px 0;
        color: #606266;
        font-size: 14px;

        .el-icon {
          color: #409eff;
        }
      }
    }
  }

  .edit-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      font-weight: 600;
      color: #2c3e50;
    }
  }

  .profile-form {
    max-width: 500px;
  }
}
</style>
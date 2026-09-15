<template>
  <div class="dashboard-page">
    <!-- 欢迎横幅 -->
    <div class="welcome-banner">
      <div class="welcome-content">
        <h1 class="welcome-title">欢迎使用 Base 底座平台</h1>
        <p class="welcome-desc">统一的企业级管理后台，提供租户、应用、用户、角色、权限等基础能力。</p>
      </div>
      <div class="welcome-decoration">
        <el-icon :size="80" color="rgba(255,255,255,0.15)"><Management /></el-icon>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="24" :sm="12" :md="8" :lg="4" v-for="stat in statList" :key="stat.title">
        <el-card class="stat-card-wrapper" shadow="hover">
          <stat-card v-bind="stat" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 功能介绍 -->
    <el-card class="feature-card" shadow="never">
      <template #header>
        <div class="feature-header">
          <el-icon :size="20" color="#2563eb"><InfoFilled /></el-icon>
          <span>平台能力</span>
        </div>
      </template>
      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8" v-for="feature in features" :key="feature.title">
          <div class="feature-item">
            <div class="feature-icon" :style="{ backgroundColor: feature.bg, color: feature.color }">
              <el-icon :size="22"><component :is="feature.icon" /></el-icon>
            </div>
            <div class="feature-info">
              <div class="feature-title">{{ feature.title }}</div>
              <div class="feature-desc">{{ feature.desc }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- 我的应用（来自 /app-instances/my：当前租户已开通且启用的应用） -->
    <el-card v-if="myApps.length" class="feature-card app-card" shadow="never">
      <template #header>
        <div class="feature-header">
          <el-icon :size="20" color="#2563eb"><Grid /></el-icon>
          <span>我的应用</span>
        </div>
      </template>
      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8" v-for="app in myApps" :key="app.id">
          <div class="feature-item app-item" @click="openApp(app)">
            <div class="feature-icon" :style="{ backgroundColor: '#e0f2fe', color: '#409EFF' }">
              <el-icon :size="22"><Grid /></el-icon>
            </div>
            <div class="feature-info">
              <div class="feature-title">{{ app.name }}</div>
              <div class="feature-desc">{{ app.apiPrefix || app.frontendUrl || 'API 代理接入' }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import StatCard from './components/StatCard.vue'
import { getDashboardStats, type DashboardStats } from '@/api/dashboard'
import { Management, InfoFilled } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import type { App } from '@/api/app'

const userStore = useUserStore()
const appStore = useAppStore()
// 仅平台超级管理员（tenantId === 0）可看全平台统计
const isSuperAdmin = computed(() => userStore.userInfo?.tenantId === 0)

// 当前租户已开通的应用（普通用户也可见，未开通则为空、不展示卡片）
const myApps = computed(() => appStore.myApps)

const openApp = (app: App) => {
  if (app.frontendUrl) {
    window.open(app.frontendUrl, '_blank')
    return
  }
  ElMessage.info('该应用未配置前端入口，请通过左侧菜单访问')
}

const stats = reactive<DashboardStats>({
  tenantCount: 0,
  appCount: 0,
  userCount: 0,
  roleCount: 0,
  organizationCount: 0,
  messageCount: 0
})

// 统计口径：平台超管看全平台；租户用户只看本租户（租户数对租户无意义，仅超管展示）
const statList = computed(() => {
  const list = [
    { title: '租户数', value: stats.tenantCount, icon: 'OfficeBuilding', color: '#409EFF', superAdminOnly: true },
    { title: isSuperAdmin.value ? '应用数' : '已开通应用', value: stats.appCount, icon: 'Grid', color: '#67C23A', superAdminOnly: false },
    { title: '用户数', value: stats.userCount, icon: 'User', color: '#E6A23C', superAdminOnly: false },
    { title: '角色数', value: stats.roleCount, icon: 'UserFilled', color: '#F56C6C', superAdminOnly: false },
    { title: '机构数', value: stats.organizationCount, icon: 'OfficeBuilding', color: '#909399', superAdminOnly: false },
    { title: '消息数', value: stats.messageCount, icon: 'Message', color: '#8E44AD', superAdminOnly: false }
  ]
  return list.filter((item) => !item.superAdminOnly || isSuperAdmin.value)
})

const features = [
  { title: '多租户管理', desc: '支持租户隔离与平台级资源管理', icon: 'OfficeBuilding', color: '#409EFF', bg: '#e0f2fe' },
  { title: '应用接入', desc: 'IFrame、API 代理或微应用方式接入', icon: 'Grid', color: '#67C23A', bg: '#dcfce7' },
  { title: '权限控制', desc: '基于角色与菜单的细粒度权限控制', icon: 'Lock', color: '#F56C6C', bg: '#fee2e2' },
  { title: '组织架构', desc: '支持多级机构与人员管理', icon: 'Connection', color: '#E6A23C', bg: '#fef3c7' },
  { title: '消息中心', desc: '统一消息模板与通知能力', icon: 'Message', color: '#8E44AD', bg: '#f3e8ff' },
  { title: '系统配置', desc: '字典、参数与审计日志统一管理', icon: 'Setting', color: '#909399', bg: '#f3f4f6' }
]

const fetchStats = async () => {
  const res: any = await getDashboardStats()
  Object.assign(stats, res.data || {})
}

const fetchMyApps = async () => {
  try {
    await appStore.fetchMyApps()
  } catch {
    // 拉取失败不影响仪表盘其它内容
  }
}

onMounted(() => {
  fetchStats()
  fetchMyApps()
})
</script>

<style scoped lang="scss">
.dashboard-page {
  .welcome-banner {
    position: relative;
    border-radius: 16px;
    padding: 32px;
    margin-bottom: 20px;
    background: linear-gradient(135deg, #2563eb 0%, #4f46e5 100%);
    color: #fff;
    overflow: hidden;
    box-shadow: 0 12px 32px rgba(37, 99, 235, 0.3);

    .welcome-content {
      position: relative;
      z-index: 1;
    }

    .welcome-title {
      margin: 0;
      font-size: 26px;
      font-weight: 700;
      margin-bottom: 10px;
    }

    .welcome-desc {
      margin: 0;
      font-size: 15px;
      opacity: 0.9;
      max-width: 560px;
      line-height: 1.6;
    }

    .welcome-decoration {
      position: absolute;
      right: 32px;
      top: 50%;
      transform: translateY(-50%);
    }
  }

  .stat-row {
    margin-bottom: 20px;

    .el-col {
      margin-bottom: 16px;
    }
  }

  .stat-card-wrapper {
    border-radius: 14px;
    border: none;
    transition: transform 0.3s ease, box-shadow 0.3s ease;

    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 12px 28px rgba(0, 0, 0, 0.08);
    }

    :deep(.el-card__body) {
      padding: 20px;
    }
  }

  .feature-card {
    border-radius: 14px;
    border: none;

    :deep(.el-card__header) {
      padding: 18px 24px;
      border-bottom: 1px solid #f1f5f9;
    }

    :deep(.el-card__body) {
      padding: 24px;
    }
  }

  .feature-header {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 16px;
    font-weight: 600;
    color: #1e293b;
  }

  .feature-item {
    display: flex;
    align-items: flex-start;
    gap: 14px;
    padding: 16px;
    border-radius: 12px;
    transition: background 0.25s ease;

    &:hover {
      background: #f8fafc;
    }
  }

  .app-card {
    margin-top: 20px;
  }

  .app-item {
    cursor: pointer;
  }

  .feature-icon {
    width: 46px;
    height: 46px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .feature-info {
    min-width: 0;
  }

  .feature-title {
    font-size: 15px;
    font-weight: 600;
    color: #1e293b;
    margin-bottom: 4px;
  }

  .feature-desc {
    font-size: 13px;
    color: #64748b;
    line-height: 1.5;
  }
}

@media (max-width: 768px) {
  .dashboard-page {
    .welcome-banner {
      padding: 24px;

      .welcome-decoration {
        display: none;
      }
    }
  }
}
</style>

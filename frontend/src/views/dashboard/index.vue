<template>
  <div class="dashboard">
    <el-row :gutter="20" class="stat-cards">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(64, 158, 255, 0.1); color: #409eff;">
              <el-icon size="28"><UserFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">1,286</div>
              <div class="stat-label">注册用户</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(103, 194, 58, 0.1); color: #67c23a;">
              <el-icon size="28"><DocumentChecked /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">3,452</div>
              <div class="stat-label">内容文章</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(230, 162, 60, 0.1); color: #e6a23c;">
              <el-icon size="28"><View /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">89,234</div>
              <div class="stat-label">今日访问</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(245, 108, 108, 0.1); color: #f56c6c;">
              <el-icon size="28"><BellFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">12</div>
              <div class="stat-label">待处理消息</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="dashboard-main">
      <el-col :xs="24" :lg="16">
        <el-card class="chart-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>访问趋势</span>
              <el-radio-group v-model="chartPeriod" size="small">
                <el-radio-button label="week">本周</el-radio-button>
                <el-radio-button label="month">本月</el-radio-button>
                <el-radio-button label="year">全年</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="chart-placeholder">
            <div class="mock-chart">
              <div v-for="(item, index) in visitData" :key="index" class="bar-item">
                <div class="bar" :style="{ height: item.value + '%' }"></div>
                <div class="bar-label">{{ item.label }}</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="8">
        <el-card class="notice-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>系统公告</span>
              <el-link type="primary" :underline="false">更多</el-link>
            </div>
          </template>
          <div class="notice-list">
            <div v-for="(item, index) in notices" :key="index" class="notice-item">
              <el-tag :type="item.type" size="small">{{ item.tag }}</el-tag>
              <span class="notice-title">{{ item.title }}</span>
              <span class="notice-time">{{ item.time }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="dashboard-bottom">
      <el-col :xs="24" :lg="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>快捷入口</span>
            </div>
          </template>
          <div class="quick-links">
            <div v-for="(item, index) in quickLinks" :key="index" class="quick-item" @click="$router.push(item.path)">
              <div class="quick-icon" :style="{ background: item.bg, color: item.color }">
                <el-icon size="24"><component :is="item.icon" /></el-icon>
              </div>
              <div class="quick-name">{{ item.name }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>登录日志</span>
            </div>
          </template>
          <el-table :data="loginLogs" size="small" :show-header="false">
            <el-table-column prop="time" width="160" />
            <el-table-column prop="ip" width="140" />
            <el-table-column prop="device" />
            <el-table-column width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === '成功' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {
  UserFilled,
  DocumentChecked,
  View,
  BellFilled
} from '@element-plus/icons-vue'

const chartPeriod = ref('week')

const visitData = [
  { label: '周一', value: 45 },
  { label: '周二', value: 62 },
  { label: '周三', value: 55 },
  { label: '周四', value: 78 },
  { label: '周五', value: 68 },
  { label: '周六', value: 85 },
  { label: '周日', value: 72 }
]

const notices = [
  { tag: '通知', title: '系统将于今晚进行例行维护', time: '2小时前', type: 'info' as const },
  { tag: '公告', title: '新版内容编辑器已上线', time: '5小时前', type: 'success' as const },
  { tag: '预警', title: '检测到异常登录尝试', time: '1天前', type: 'warning' as const },
  { tag: '通知', title: '请及时更新账户密码', time: '2天前', type: 'info' as const },
  { tag: '公告', title: '新增数据导出功能', time: '3天前', type: 'success' as const }
]

const quickLinks = [
  { name: '用户管理', icon: 'User', path: '/users', bg: 'rgba(64, 158, 255, 0.1)', color: '#409eff' },
  { name: '内容发布', icon: 'EditPen', path: '/dashboard', bg: 'rgba(103, 194, 58, 0.1)', color: '#67c23a' },
  { name: '系统监控', icon: 'Monitor', path: '/dashboard', bg: 'rgba(230, 162, 60, 0.1)', color: '#e6a23c' },
  { name: '系统设置', icon: 'Setting', path: '/settings', bg: 'rgba(245, 108, 108, 0.1)', color: '#f56c6c' }
]

const loginLogs = [
  { time: '2026-06-01 09:30:22', ip: '192.168.1.100', device: 'Chrome / Windows', status: '成功' },
  { time: '2026-05-31 18:15:06', ip: '192.168.1.100', device: 'Chrome / Windows', status: '成功' },
  { time: '2026-05-31 08:42:33', ip: '192.168.1.105', device: 'Safari / macOS', status: '成功' },
  { time: '2026-05-30 14:20:11', ip: '10.0.0.55', device: 'Firefox / Linux', status: '失败' }
]
</script>

<style scoped lang="scss">
.dashboard {
  .stat-cards {
    margin-bottom: 20px;
  }

  .stat-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    :deep(.el-card__body) {
      padding: 20px;
    }
  }

  .stat-content {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .stat-icon {
    width: 56px;
    height: 56px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .stat-value {
    font-size: 24px;
    font-weight: 700;
    color: #2c3e50;
    line-height: 1.2;
  }

  .stat-label {
    font-size: 13px;
    color: #909399;
    margin-top: 4px;
  }

  .dashboard-main {
    margin-bottom: 20px;
  }

  .chart-card,
  .notice-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-weight: 600;
    color: #2c3e50;
  }

  .mock-chart {
    display: flex;
    align-items: flex-end;
    justify-content: space-around;
    height: 200px;
    padding: 20px 0;

    .bar-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 8px;
      flex: 1;
    }

    .bar {
      width: 32px;
      background: linear-gradient(180deg, #409eff 0%, #a0cfff 100%);
      border-radius: 6px 6px 0 0;
      transition: height 0.5s ease;
      min-height: 20px;
    }

    .bar-label {
      font-size: 12px;
      color: #909399;
    }
  }

  .notice-list {
    .notice-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px 0;
      border-bottom: 1px solid #f0f7ff;

      &:last-child {
        border-bottom: none;
      }

      .notice-title {
        flex: 1;
        font-size: 14px;
        color: #606266;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .notice-time {
        font-size: 12px;
        color: #c0c4cc;
      }
    }
  }

  .dashboard-bottom {
    .el-card {
      border-radius: 12px;
      border: 1px solid #e6f2ff;
    }
  }

  .quick-links {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;

    .quick-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 10px;
      padding: 20px;
      border-radius: 12px;
      cursor: pointer;
      transition: all 0.3s;

      &:hover {
        background: #f5f9ff;
        transform: translateY(-2px);
      }

      .quick-icon {
        width: 48px;
        height: 48px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
      }

      .quick-name {
        font-size: 13px;
        color: #606266;
      }
    }
  }
}
</style>
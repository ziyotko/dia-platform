<template>
  <div class="page-container">
    <el-card shadow="hover" class="settings-card">
      <template #header>
        <div class="card-header">静态化设置</div>
      </template>

      <el-form :model="form" label-width="180px" class="settings-form">
        <el-form-item label="静态化输出路径">
          <el-input
            v-model="form.staticPath"
            placeholder="请输入静态化输出路径，如 D:/static"
            clearable
          />
        </el-form-item>

        <el-form-item label="静态化程序访问地址">
          <el-input
            v-model="form.staticProgramAddr"
            placeholder="请输入静态化程序访问地址，如 127.0.0.1:8889"
            clearable
          />
        </el-form-item>

        <el-form-item label="静态化程序访问令牌名">
          <el-input
            v-model="form.staticProgramTokenName"
            placeholder="请输入静态化程序访问令牌名，如 CAAM_TOKEN"
            clearable
          />
        </el-form-item>

        <el-form-item label="首页整体变灰">
          <el-switch v-model="form.homeGray" active-text="开启" inactive-text="关闭" />
        </el-form-item>

        <el-alert
          type="info"
          :closable="false"
          show-icon
          title="「固定静态化时间」暂未启用：后端尚无定时调度器读取这些配置，为避免误以为已生效，相关设置项暂不可操作。"
          style="margin-bottom: 18px"
        />

        <el-form-item label="设置首页固定静态化时间">
          <div class="time-row">
            <el-switch v-model="form.homeStaticTimeEnabled" :disabled="!staticTimeSchedulerEnabled" />
            <el-time-picker
              v-model="form.homeStaticTime"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="请选择时间"
              :disabled="!staticTimeSchedulerEnabled"
              class="time-picker"
            />
          </div>
        </el-form-item>

        <el-form-item label="设置栏目页固定静态化时间">
          <div class="time-row">
            <el-switch v-model="form.columnStaticTimeEnabled" :disabled="!staticTimeSchedulerEnabled" />
            <el-time-picker
              v-model="form.columnStaticTime"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="请选择时间"
              :disabled="!staticTimeSchedulerEnabled"
              class="time-picker"
            />
          </div>
        </el-form-item>

        <el-form-item label="设置专题页固定静态化时间">
          <div class="time-row">
            <el-switch v-model="form.specialStaticTimeEnabled" :disabled="!staticTimeSchedulerEnabled" />
            <el-time-picker
              v-model="form.specialStaticTime"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="请选择时间"
              :disabled="!staticTimeSchedulerEnabled"
              class="time-picker"
            />
          </div>
        </el-form-item>

        <el-form-item label="设置详情页固定静态化时间">
          <div class="time-row">
            <el-switch v-model="form.detailStaticTimeEnabled" :disabled="!staticTimeSchedulerEnabled" />
            <el-time-picker
              v-model="form.detailStaticTime"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="请选择时间"
              :disabled="!staticTimeSchedulerEnabled"
              class="time-picker"
            />
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleSave">保存设置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSettings, updateSettings } from '@/api/settings'
import type { Settings } from '@/api/settings'

const loading = ref(false)

// 「固定静态化时间」调度器暂未实现（后端无任何定时任务读取这些配置）。
// 暂时置为 false，禁用相关设置项；待调度器实现后改回 true 即可恢复。
const staticTimeSchedulerEnabled = false

const form = reactive({
  staticPath: '',
  staticProgramAddr: '',
  staticProgramTokenName: '',
  homeGray: false,
  // 固定静态化时间（调度器未实现，仅作受控占位，始终禁用）
  homeStaticTimeEnabled: false,
  homeStaticTime: '',
  columnStaticTimeEnabled: false,
  columnStaticTime: '',
  specialStaticTimeEnabled: false,
  specialStaticTime: '',
  detailStaticTimeEnabled: false,
  detailStaticTime: ''
})

const loadSettings = async () => {
  try {
    const res: any = await getSettings()
    const data = res.data as Settings
    form.staticPath = data.staticPath || ''
    form.staticProgramAddr = data.staticProgramAddr || ''
    form.staticProgramTokenName = data.staticProgramTokenName || ''
    form.homeGray = data.homeGray ?? false
    // 固定静态化时间调度器暂未实现，不读取/展示这些死配置（保持默认关闭）
  } catch {
    ElMessage.error('获取静态化设置失败')
  }
}

const handleSave = async () => {
  loading.value = true
  try {
    await updateSettings({
      staticPath: form.staticPath,
      staticProgramAddr: form.staticProgramAddr,
      staticProgramTokenName: form.staticProgramTokenName,
      homeGray: form.homeGray
      // 固定静态化时间调度器暂未实现，不提交这些死配置
    })
    ElMessage.success('保存成功')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped lang="scss">
.page-container {
  .settings-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      font-weight: 600;
      color: #2c3e50;
    }
  }

  .settings-form {
    max-width: 600px;
    padding: 20px 0;
  }

  .time-row {
    display: flex;
    align-items: center;
    gap: 12px;

    .time-picker {
      width: 160px;
    }
  }
}
</style>

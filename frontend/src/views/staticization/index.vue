<template>
  <div class="page-container">
    <el-card shadow="hover" class="settings-card">
      <template #header>
        <div class="card-header">静态化设置</div>
      </template>

      <el-form :model="form" label-width="180px" class="settings-form">
        <el-form-item label="首页整体变灰">
          <el-switch v-model="form.homeGray" active-text="开启" inactive-text="关闭" />
        </el-form-item>

        <el-form-item label="首页自动静态化时间">
          <div class="time-row">
            <el-switch v-model="form.homeStaticTimeEnabled" />
            <el-time-picker
              v-model="form.homeStaticTime"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="请选择时间"
              :disabled="!form.homeStaticTimeEnabled"
              class="time-picker"
            />
          </div>
        </el-form-item>

        <el-form-item label="栏目页自动静态化时间">
          <div class="time-row">
            <el-switch v-model="form.columnStaticTimeEnabled" />
            <el-time-picker
              v-model="form.columnStaticTime"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="请选择时间"
              :disabled="!form.columnStaticTimeEnabled"
              class="time-picker"
            />
          </div>
        </el-form-item>

        <el-form-item label="专题页自动静态化时间">
          <div class="time-row">
            <el-switch v-model="form.specialStaticTimeEnabled" />
            <el-time-picker
              v-model="form.specialStaticTime"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="请选择时间"
              :disabled="!form.specialStaticTimeEnabled"
              class="time-picker"
            />
          </div>
        </el-form-item>

        <el-form-item label="详情页自动静态化时间">
          <div class="time-row">
            <el-switch v-model="form.detailStaticTimeEnabled" />
            <el-time-picker
              v-model="form.detailStaticTime"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="请选择时间"
              :disabled="!form.detailStaticTimeEnabled"
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

const form = reactive({
  homeGray: false,
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
    form.homeGray = data.homeGray ?? false
    form.homeStaticTimeEnabled = data.homeStaticTimeEnabled ?? false
    form.homeStaticTime = data.homeStaticTime || ''
    form.columnStaticTimeEnabled = data.columnStaticTimeEnabled ?? false
    form.columnStaticTime = data.columnStaticTime || ''
    form.specialStaticTimeEnabled = data.specialStaticTimeEnabled ?? false
    form.specialStaticTime = data.specialStaticTime || ''
    form.detailStaticTimeEnabled = data.detailStaticTimeEnabled ?? false
    form.detailStaticTime = data.detailStaticTime || ''
  } catch {
    ElMessage.error('获取静态化设置失败')
  }
}

const handleSave = async () => {
  loading.value = true
  try {
    await updateSettings({
      homeGray: form.homeGray,
      homeStaticTimeEnabled: form.homeStaticTimeEnabled,
      homeStaticTime: form.homeStaticTimeEnabled ? form.homeStaticTime : '',
      columnStaticTimeEnabled: form.columnStaticTimeEnabled,
      columnStaticTime: form.columnStaticTimeEnabled ? form.columnStaticTime : '',
      specialStaticTimeEnabled: form.specialStaticTimeEnabled,
      specialStaticTime: form.specialStaticTimeEnabled ? form.specialStaticTime : '',
      detailStaticTimeEnabled: form.detailStaticTimeEnabled,
      detailStaticTime: form.detailStaticTimeEnabled ? form.detailStaticTime : ''
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

<template>
  <div>
    <el-card>
      <h3>{{ isEdit ? '编辑会议' : '创建会议' }}</h3>
      <el-form :model="form" label-width="100px">
        <el-form-item label="会议名称"><el-input v-model="form.title" /></el-form-item>
        <el-row :gutter="16">
          <el-col :span="8"><el-form-item label="类型"><el-select v-model="form.type"><el-option label="线上" value="online" /><el-option label="线下" value="offline" /><el-option label="混合" value="hybrid" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="名额"><el-input-number v-model="form.capacity" :min="0" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="费用"><el-input-number v-model="form.fee" :min="0" :precision="2" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="时间"><el-date-picker v-model="timeRange" type="datetimerange" range-separator="至" style="width:100%" /></el-form-item>
        <el-form-item label="地点"><el-input v-model="form.location" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
        <el-divider>功能开关</el-divider>
        <el-form-item label="报名审核"><el-switch v-model="form.needApproval" /></el-form-item>
        <el-form-item label="候补机制"><el-switch v-model="form.allowWaitlist" /></el-form-item>
        <el-form-item label="会议收费"><el-switch v-model="form.isPaid" /></el-form-item>
        <el-form-item label="开启投票"><el-switch v-model="form.enableVote" /></el-form-item>
        <el-form-item label="开启问卷"><el-switch v-model="form.enableSurvey" /></el-form-item>
        <el-form-item label="开启学分"><el-switch v-model="form.enableCredit" /></el-form-item>
        <el-form-item label="开启直播"><el-switch v-model="form.enableLive" /></el-form-item>
        <el-form-item><el-button type="primary" @click="save">保存</el-button><el-button @click="$router.back()">返回</el-button></el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'

const route = useRoute()
const isEdit = ref(false)
const timeRange = ref<any[]>([])

const form = reactive<any>({
  title: '', type: 'offline', capacity: 0, fee: 0, location: '', description: '',
  needApproval: false, allowWaitlist: false, isPaid: false,
  enableVote: false, enableSurvey: false, enableCredit: false, enableLive: false
})

onMounted(async () => {
  const id = route.params.id
  if (id) {
    isEdit.value = true
    try {
      const res = await adminApi.getMeeting(Number(id))
      Object.assign(form, res.data)
      if (res.data.startTime && res.data.endTime) timeRange.value = [new Date(res.data.startTime), new Date(res.data.endTime)]
    } catch (e) {}
  }
})

async function save() {
  try {
    if (timeRange.value.length === 2) { form.startTime = timeRange.value[0]; form.endTime = timeRange.value[1] }
    if (isEdit.value) await adminApi.updateMeeting(Number(route.params.id), form)
    else await adminApi.createMeeting(form)
    ElMessage.success('保存成功')
  } catch (e: any) {}
}
</script>

<template>
  <div>
    <el-card>
      <h3>{{ isEdit ? '编辑投票' : '创建投票' }}</h3>
      <el-form :model="form" label-width="100px">
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="会议"><el-select v-model="form.meetingId"><el-option v-for="m in meetings" :key="m.id" :label="m.title" :value="m.id" /></el-select></el-form-item>
        <el-form-item label="类型"><el-select v-model="form.type"><el-option label="单选" value="single" /><el-option label="多选" value="multiple" /><el-option label="等额" value="equal" /><el-option label="差额" value="differential" /></el-select></el-form-item>
        <el-form-item label="时间"><el-date-picker v-model="timeRange" type="datetimerange" style="width:100%" /></el-form-item>
        <el-form-item label="匿名投票"><el-switch v-model="form.isAnonymous" /></el-form-item>
        <el-form-item><el-button type="primary" @click="save">保存</el-button></el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'

const route = useRoute(); const isEdit = ref(false)
const meetings = ref<any[]>([]); const timeRange = ref<any[]>([])

const form = reactive<any>({ title: '', meetingId: 0, type: 'single', isAnonymous: false, options: [] })

onMounted(async () => {
  try { const res = await adminApi.getMeetings({ pageSize: 1000 }); meetings.value = res.data?.list || [] } catch (e) {}
  const id = route.params.id
  if (id) {
    isEdit.value = true
    try { const res = await adminApi.getVote(Number(id)); Object.assign(form, res.data); if (form.startTime) timeRange.value = [new Date(form.startTime), new Date(form.endTime)] } catch (e) {}
  }
})

async function save() {
  try {
    if (timeRange.value.length === 2) { form.startTime = timeRange.value[0]; form.endTime = timeRange.value[1] }
    const data = { vote: { ...form, options: undefined }, options: form.options || [] }
    if (isEdit.value) await adminApi.updateVote(Number(route.params.id), form)
    else await adminApi.createVote(data)
    ElMessage.success('保存成功')
  } catch (e: any) {}
}
</script>

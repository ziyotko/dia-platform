<template>
  <div>
    <el-card>
      <h3>{{ isEdit ? '编辑问卷' : '创建问卷' }}</h3>
      <el-form :model="form" label-width="100px">
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="form.description" type="textarea" /></el-form-item>
        <el-form-item label="会议"><el-select v-model="form.meetingId"><el-option v-for="m in meetings" :key="m.id" :label="m.title" :value="m.id" /></el-select></el-form-item>
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

const route = useRoute(); const isEdit = ref(false); const meetings = ref<any[]>([])

const form = reactive<any>({ title: '', description: '', meetingId: 0, questions: [] })

onMounted(async () => {
  try { const res = await adminApi.getMeetings({ pageSize: 1000 }); meetings.value = res.data?.list || [] } catch (e) {}
  const id = route.params.id
  if (id) {
    isEdit.value = true
    try { const res = await adminApi.getSurvey(Number(id)); Object.assign(form, res.data) } catch (e) {}
  }
})

async function save() {
  try {
    const data = { survey: form, questions: form.questions || [] }
    if (isEdit.value) await adminApi.updateSurvey(Number(route.params.id), form)
    else await adminApi.createSurvey(data)
    ElMessage.success('保存成功')
  } catch (e: any) {}
}
</script>

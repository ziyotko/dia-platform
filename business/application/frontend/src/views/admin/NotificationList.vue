<template>
  <div class="page-card" style="max-width:700px">
    <h3 style="margin-bottom:16px">发送通知</h3>
    <el-form :model="form" label-width="90px">
      <el-form-item label="通知对象">
        <el-radio-group v-model="form.userId">
          <el-radio :label="0">全部申报人</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="通知标题"><el-input v-model="form.title" /></el-form-item>
      <el-form-item label="通知内容"><el-input v-model="form.content" type="textarea" :rows="5" /></el-form-item>
      <el-form-item>
        <el-button type="primary" @click="send" :loading="loading">发送通知</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'

const form = reactive({ userId: 0, title: '', content: '', type: 'system' })
const loading = ref(false)

async function send() {
  if (!form.title || !form.content) return ElMessage.warning('请填写标题和内容')
  loading.value = true
  try {
    await adminApi.sendNotification({ userId: form.userId, title: form.title, content: form.content, type: form.type })
    ElMessage.success('通知已发送')
    form.title = ''
    form.content = ''
  } finally {
    loading.value = false
  }
}
</script>

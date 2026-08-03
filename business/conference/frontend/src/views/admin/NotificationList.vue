<template>
  <div>
    <el-card>
      <el-button type="primary" @click="sendVisible=true" style="margin-bottom:16px">发送通知</el-button>
      <el-table :data="list" stripe>
        <el-table-column prop="title" label="标题" />
        <el-table-column label="目标" width="100"><template #default="{row}">{{ targetLabel(row.targetType) }}</template></el-table-column>
        <el-table-column label="发送时间" width="160"><template #default="{row}">{{ row.sentAt?.slice(0,16) }}</template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>

    <el-dialog v-model="sendVisible" title="发送通知" width="500px">
      <el-form :model="sendForm">
        <el-form-item label="标题"><el-input v-model="sendForm.title" /></el-form-item>
        <el-form-item label="内容"><el-input v-model="sendForm.content" type="textarea" :rows="4" /></el-form-item>
        <el-form-item label="目标"><el-select v-model="sendForm.targetType"><el-option label="全部会员" value="all" /><el-option label="指定会员" value="specific" /></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="sendVisible=false">取消</el-button><el-button type="primary" @click="doSend">发送</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([]); const page = ref(1); const size = 10; const total = ref(0)
const sendVisible = ref(false)
const sendForm = reactive({ title: '', content: '', targetType: 'all', targetIds: [] as number[] })

function targetLabel(t: string) { const m: any = { all:'全部会员', specific:'指定会员', registered:'报名人员', branch:'指定分会', level:'指定等级' }; return m[t] || t }

async function load() { try { const res = await adminApi.getSentNotifications({ page: page.value, pageSize: size }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {} }

async function doSend() { try { await adminApi.sendNotification(sendForm); ElMessage.success('发送成功'); sendVisible.value = false; load() } catch (e: any) {} }

load()
</script>

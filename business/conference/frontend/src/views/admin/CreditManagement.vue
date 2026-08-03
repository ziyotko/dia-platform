<template>
  <div>
    <el-card>
      <el-button type="primary" @click="adjustVisible=true" style="margin-bottom:16px">手动调整学分</el-button>
      <el-table :data="list" stripe>
        <el-table-column prop="user.realName" label="用户" />
        <el-table-column prop="meeting.title" label="会议" />
        <el-table-column prop="credits" label="学分" width="80" />
        <el-table-column label="来源" width="80"><template #default="{row}">{{ row.source==='auto_assign'?'自动':'手动' }}</template></el-table-column>
        <el-table-column prop="remark" label="备注" />
        <el-table-column label="时间" width="160"><template #default="{row}">{{ row.createdAt?.slice(0,16) }}</template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>

    <el-dialog v-model="adjustVisible" title="手动调整学分" width="400px">
      <el-form :model="adjustForm">
        <el-form-item label="用户ID"><el-input-number v-model="adjustForm.userId" /></el-form-item>
        <el-form-item label="会议ID"><el-input-number v-model="adjustForm.meetingId" /></el-form-item>
        <el-form-item label="学分"><el-input-number v-model="adjustForm.credits" :precision="1" /></el-form-item>
        <el-form-item label="原因"><el-input v-model="adjustForm.remark" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="adjustVisible=false">取消</el-button><el-button type="primary" @click="doAdjust">确认</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([]); const page = ref(1); const size = 10; const total = ref(0)
const adjustVisible = ref(false)
const adjustForm = reactive({ userId: 0, meetingId: 0, credits: 0, remark: '' })

async function load() { try { const res = await adminApi.getCreditRecords({ page: page.value, pageSize: size }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {} }

async function doAdjust() {
  try { await adminApi.manualAdjust(adjustForm); ElMessage.success('调整成功'); adjustVisible.value = false; load() } catch (e: any) {}
}

load()
</script>

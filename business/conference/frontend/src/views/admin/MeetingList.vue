<template>
  <div>
    <el-card>
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
        <el-input v-model="keyword" placeholder="搜索会议" style="width:240px" clearable @change="load" />
        <el-button type="primary" @click="$router.push('/admin/meetings/create')">创建会议</el-button>
      </div>
      <el-table :data="list" stripe>
        <el-table-column prop="title" label="会议名称" />
        <el-table-column label="类型" width="80"><template #default="{row}"><el-tag size="small">{{ row.type==='online'?'线上':row.type==='offline'?'线下':'混合' }}</el-tag></template></el-table-column>
        <el-table-column label="状态" width="80"><template #default="{row}"><el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag></template></el-table-column>
        <el-table-column label="时间" width="160"><template #default="{row}">{{ row.startTime?.slice(0,16) }}</template></el-table-column>
        <el-table-column label="操作" width="240"><template #default="{row}">
          <el-button size="small" @click="$router.push(`/admin/meetings/${row.id}`)">详情</el-button>
          <el-button size="small" @click="$router.push(`/admin/meetings/${row.id}/edit`)">编辑</el-button>
          <el-button size="small" type="danger" @click="del(row.id)">删除</el-button>
          <el-button size="small" type="warning" v-if="row.status==='open'" @click="close(row.id)">关闭</el-button>
        </template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([]); const page = ref(1); const size = 10; const total = ref(0); const keyword = ref('')

function statusType(s: string) { const m: any = { draft:'info', open:'success', closed:'warning', archived:'info' }; return m[s] || 'info' }
function statusLabel(s: string) { const m: any = { draft:'草稿', open:'进行中', closed:'已关闭', archived:'已归档' }; return m[s] || s }

async function load() {
  try { const res = await adminApi.getMeetings({ page: page.value, pageSize: size, keyword: keyword.value }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {}
}

async function del(id: number) {
  try { await ElMessageBox.confirm('确认删除？'); await adminApi.deleteMeeting(id); ElMessage.success('已删除'); load() } catch (e) {}
}

async function close(id: number) {
  try { await adminApi.closeMeeting(id); ElMessage.success('会议已关闭'); load() } catch (e: any) {}
}

load()
</script>

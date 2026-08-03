<template>
  <div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="username" label="操作用户" />
        <el-table-column label="类型" width="80"><template #default="{row}"><el-tag size="small">{{ row.userType==='admin'?'管理员':'会员' }}</el-tag></template></el-table-column>
        <el-table-column prop="action" label="操作" />
        <el-table-column prop="resource" label="资源" />
        <el-table-column prop="ip" label="IP" width="140" />
        <el-table-column label="时间" width="160"><template #default="{row}">{{ row.createdAt?.slice(0,16) }}</template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([]); const page = ref(1); const size = 10; const total = ref(0)

async function load() { try { const res = await adminApi.getAuditLogs({ page: page.value, pageSize: size }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {} }
load()
</script>

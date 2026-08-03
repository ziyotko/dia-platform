<template>
  <div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="meetingTitle" label="会议" />
        <el-table-column prop="meetingType" label="类型" width="80" />
        <el-table-column label="年份" width="80"><template #default="{row}">{{ row.meetingYear }}</template></el-table-column>
        <el-table-column label="归档时间" width="160"><template #default="{row}">{{ row.archivedAt?.slice(0,16) }}</template></el-table-column>
        <el-table-column label="操作" width="120"><template #default="{row}"><el-button size="small" @click="$router.push(`/admin/archives/${row.id}`)">查看</el-button></template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([]); const page = ref(1); const size = 10; const total = ref(0)

async function load() { try { const res = await adminApi.getArchives({ page: page.value, pageSize: size }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {} }
load()
</script>

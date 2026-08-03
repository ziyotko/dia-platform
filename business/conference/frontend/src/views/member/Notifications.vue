<template>
  <div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="title" label="标题" />
        <el-table-column label="内容" show-overflow-tooltip><template #default="{row}">{{ row.content }}</template></el-table-column>
        <el-table-column label="状态" width="80"><template #default="{row}"><el-tag :type="row.is_read?'info':'warning'" size="small">{{ row.is_read?'已读':'未读' }}</el-tag></template></el-table-column>
        <el-table-column label="时间" width="160"><template #default="{row}">{{ row.sent_at?.slice(0,16) || row.created_at?.slice(0,16) }}</template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { memberApi } from '@/api/member'

const list = ref<any[]>([])
const page = ref(1); const size = 10; const total = ref(0)

async function load() {
  try { const res = await memberApi.getNotifications({ page: page.value, pageSize: size }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {}
}

load()
</script>

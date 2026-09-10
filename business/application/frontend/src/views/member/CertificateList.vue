<template>
  <div class="page-card">
    <el-table :data="list" v-loading="loading">
      <el-table-column prop="certNo" label="证书编号" width="180" />
      <el-table-column prop="title" label="证书名称" min-width="200" />
      <el-table-column prop="holder" label="持有人" width="120" />
      <el-table-column label="所属项目" min-width="180">
        <template #default="{ row }">{{ row.application?.title || '-' }}</template>
      </el-table-column>
      <el-table-column label="颁发时间" width="170">
        <template #default="{ row }">{{ fmt(row.issuedAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.fileUrl" size="small" type="primary" @click="download(row)">下载证书</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { memberApi } from '@/api/member'
import { fileUrl, fmt } from '@/utils/constants'

const list = ref<any[]>([])
const loading = ref(false)

async function fetch() {
  loading.value = true
  try {
    const res = await memberApi.getMyCertificates()
    list.value = res.data
  } finally {
    loading.value = false
  }
}

function download(row: any) {
  window.open(fileUrl(row.fileUrl), '_blank')
}

onMounted(fetch)
</script>

<template>
  <div class="charter-page" v-loading="loading">
    <div class="charter-card" v-if="content || file || !loading">
      <h1>协会章程</h1>
      <div v-if="file" class="charter-file">
        <el-icon :size="16"><Document /></el-icon>
        <span class="file-name">{{ file.name }}</span>
        <span class="file-meta">{{ formatSize(file.size) }}</span>
        <el-button type="primary" plain size="small" @click="downloadPdf">下载 PDF</el-button>
      </div>
      <div v-if="content" class="rich-text-content" v-html="content"></div>
      <el-empty v-else-if="!file" description="章程暂未维护，请联系管理员" />
    </div>
    <div class="back">
      <el-button @click="$router.back()">返回</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { charterApi } from '@/api/index'

const content = ref('')
const file = ref<any>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await charterApi.getContent()
    content.value = res.data?.content || ''
    file.value = res.data?.file || null
  } catch {} finally { loading.value = false }
})

function downloadPdf() {
  const base = import.meta.env.VITE_API_BASE_URL || '/business_member/api'
  window.open(`${base}/charter`, '_blank')
}

function formatSize(n?: number) {
  if (!n) return ''
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(2)} MB`
}
</script>

<style scoped lang="scss">
.charter-page {
  max-width: 860px;
  margin: 40px auto;
  padding: 0 24px;

  .charter-card {
    background: #fff;
    border-radius: 12px;
    padding: 40px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);

    h1 {
      font-size: 24px;
      font-weight: 600;
      color: var(--app-text-primary);
      text-align: center;
      margin: 0 0 32px;
      padding-bottom: 20px;
      border-bottom: 1px solid var(--app-border);
    }
  }

  .charter-file {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: -12px 0 28px;
    padding: 12px 16px;
    background: #f7f9fc;
    border: 1px solid var(--app-border);
    border-radius: 8px;
    font-size: 13px;
    color: var(--app-text-secondary);

    .file-name {
      font-weight: 500;
      color: var(--app-text-primary);
      word-break: break-all;
    }
    .file-meta { color: var(--app-text-muted); }
  }

  .back { margin-top: 24px; }
}
</style>

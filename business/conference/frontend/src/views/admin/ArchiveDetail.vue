<template>
  <div>
    <el-card v-if="archive.id">
      <h3>{{ archive.meetingTitle }}</h3>
      <p>类型：{{ archive.meetingType }} | 年份：{{ archive.meetingYear }}</p>
      <h4 style="margin-top:16px">归档项目</h4>
      <el-table :data="archive.items" stripe>
        <el-table-column label="类型" width="120"><template #default="{row}">{{ itemLabel(row.itemType) }}</template></el-table-column>
        <el-table-column prop="fileName" label="文件名" />
        <el-table-column label="操作"><template #default="{row}"><el-button size="small" v-if="row.filePath" @click="openFile(row.filePath)">下载</el-button></template></el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { adminApi } from '@/api/admin'

const route = useRoute()
const archive = ref<any>({})

function itemLabel(t: string) { const m: any = { registration:'报名名单',signin:'签到表',finance:'缴费记录',vote:'投票结果',survey:'问卷结果',material:'会议资料',minutes:'会议纪要',live:'直播录播' }; return m[t] || t }
function openFile(path: string) { window.open(path) }

onMounted(async () => { try { const res = await adminApi.getArchive(Number(route.params.id)); archive.value = res.data } catch (e) {} })
</script>

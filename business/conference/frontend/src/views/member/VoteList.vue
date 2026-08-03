<template>
  <div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="title" label="投票标题" />
        <el-table-column label="类型" width="120"><template #default="{row}">{{ voteType(row.type) }}</template></el-table-column>
        <el-table-column label="状态" width="100"><template #default="{row}"><el-tag :type="row.status==='open'?'success':'info'">{{ row.status==='open'?'进行中':'已结束' }}</el-tag></template></el-table-column>
        <el-table-column label="时间" width="200"><template #default="{row}">{{ row.startTime?.slice(0,16) }} ~ {{ row.endTime?.slice(0,16) }}</template></el-table-column>
        <el-table-column label="操作" width="120"><template #default="{row}">
          <el-button size="small" type="primary" @click="openVote(row)" v-if="row.status==='open'">投票</el-button>
          <el-button size="small" @click="viewResult(row)">结果</el-button>
        </template></el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="voteVisible" title="投票" width="500px">
      <h3>{{ currentVote?.title }}</h3>
      <p style="color:#666;margin:8px 0">{{ currentVote?.isAnonymous ? '匿名投票' : '实名投票' }}</p>
      <el-radio-group v-model="singleVote" v-if="currentVote?.type==='single'||currentVote?.type==='equal'">
        <el-radio v-for="opt in currentVote?.options" :key="opt.id" :value="opt.id" style="display:block;margin:8px 0">{{ opt.label || opt.candidateName }}</el-radio>
      </el-radio-group>
      <el-checkbox-group v-model="multiVote" v-else>
        <el-checkbox v-for="opt in currentVote?.options" :key="opt.id" :value="opt.id" style="display:block;margin:8px 0">{{ opt.label || opt.candidateName }}</el-checkbox>
      </el-checkbox-group>
      <template #footer><el-button @click="voteVisible=false">取消</el-button><el-button type="primary" @click="submitVote">提交</el-button></template>
    </el-dialog>

    <el-dialog v-model="resultVisible" title="投票结果" width="500px">
      <div v-for="opt in resultOptions" :key="opt.id" style="margin:8px 0"><span>{{ opt.label || opt.candidateName }}</span> - <strong>{{ opt.voteCount }}票</strong></div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { memberApi } from '@/api/member'

const list = ref<any[]>([])
const voteVisible = ref(false); const currentVote = ref<any>(null); const singleVote = ref(0); const multiVote = ref<number[]>([])
const resultVisible = ref(false); const resultOptions = ref<any[]>([])

function voteType(t: string) { const m: any = { single:'单选', multiple:'多选', equal:'等额', differential:'差额' }; return m[t] || t }

async function load() { try { const res = await memberApi.getAvailableVotes(); list.value = res.data || [] } catch (e) {} }

function openVote(vote: any) { currentVote.value = vote; singleVote.value = 0; multiVote.value = []; voteVisible.value = true }

async function submitVote() {
  try {
    const data: any = {}
    if (currentVote.value?.type === 'single' || currentVote.value?.type === 'equal') data.optionId = singleVote.value
    else data.optionIds = multiVote.value
    await memberApi.castVote(currentVote.value.id, data)
    ElMessage.success('投票成功'); voteVisible.value = false; load()
  } catch (e: any) {}
}

function viewResult(vote: any) { resultOptions.value = vote.options || []; resultVisible.value = true }

load()
</script>

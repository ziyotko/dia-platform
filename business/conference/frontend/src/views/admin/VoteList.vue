<template>
  <div>
    <el-card>
      <div style="margin-bottom:16px"><el-button type="primary" @click="$router.push('/admin/votes/create')">创建投票</el-button></div>
      <el-table :data="list" stripe>
        <el-table-column prop="title" label="标题" />
        <el-table-column label="类型" width="80"><template #default="{row}">{{ vt(row.type) }}</template></el-table-column>
        <el-table-column label="状态" width="80"><template #default="{row}"><el-tag :type="row.status==='open'?'success':'info'" size="small">{{ row.status==='open'?'进行中':'已结束' }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="200"><template #default="{row}">
          <el-button size="small" @click="$router.push(`/admin/votes/${row.id}/edit`)">编辑</el-button>
          <el-button size="small" type="danger" @click="del(row.id)">删除</el-button>
          <el-button size="small" @click="results(row)">结果</el-button>
        </template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>

    <el-dialog v-model="resultVisible" title="投票结果" width="500px">
      <div v-for="opt in resultOptions" :key="opt.id" style="margin:8px 0">{{ opt.label || opt.candidateName }} - {{ opt.voteCount }}票</div>
      <p style="margin-top:12px">总投票数：{{ resultTotal }}</p>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([]); const page = ref(1); const size = 10; const total = ref(0)
const resultVisible = ref(false); const resultOptions = ref<any[]>([]); const resultTotal = ref(0)

function vt(t: string) { const m: any = { single:'单选', multiple:'多选', equal:'等额', differential:'差额' }; return m[t] || t }

async function load() { try { const res = await adminApi.getVotes({ page: page.value, pageSize: size }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {} }

async function del(id: number) { try { await ElMessageBox.confirm('确认删除？'); await adminApi.deleteVote(id); load() } catch (e) {} }

async function results(row: any) {
  try { const res = await adminApi.getVoteResults(row.id); resultOptions.value = res.data?.options || []; resultTotal.value = res.data?.total_votes || 0; resultVisible.value = true } catch (e) {}
}

load()
</script>

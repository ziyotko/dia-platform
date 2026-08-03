<template>
  <div>
    <el-card>
      <div style="margin-bottom:16px"><el-button type="primary" @click="$router.push('/admin/surveys/create')">创建问卷</el-button></div>
      <el-table :data="list" stripe>
        <el-table-column prop="title" label="标题" />
        <el-table-column label="状态" width="80"><template #default="{row}"><el-tag :type="row.status==='open'?'success':'info'" size="small">{{ row.status==='open'?'开放中':'已关闭' }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="240"><template #default="{row}">
          <el-button size="small" @click="$router.push(`/admin/surveys/${row.id}/edit`)">编辑</el-button>
          <el-button size="small" type="danger" @click="del(row.id)">删除</el-button>
          <el-button size="small" type="success" @click="$router.push(`/admin/surveys/${row.id}/results`)">结果</el-button>
        </template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([]); const page = ref(1); const size = 10; const total = ref(0)

async function load() { try { const res = await adminApi.getSurveys({ page: page.value, pageSize: size }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {} }

async function del(id: number) { try { await ElMessageBox.confirm('确认删除？'); await adminApi.deleteSurvey(id); load() } catch (e) {} }

load()
</script>

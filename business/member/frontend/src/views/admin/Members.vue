<template>
  <div class="admin-members" v-loading="loading">
    <div class="page-header">
      <h3>会员管理</h3>
      <div class="filters">
        <el-input v-model="keyword" placeholder="搜索用户名/公司名" clearable style="width:200px" @clear="search" @keyup.enter="search" />
        <el-select v-model="filterStatus" placeholder="状态" clearable style="width:130px" @change="search">
          <el-option label="注册中" value="registering" /><el-option label="待审核" value="pending_review" />
          <el-option label="正式会员" value="active" /><el-option label="已拒绝" value="rejected" />
        </el-select>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="company_name" label="公司名称" />
        <el-table-column prop="mobile" label="手机号" width="130" />
        <el-table-column prop="member_type" label="类型" width="90"><template #default="{row}">{{ row.member_type === 'unit' ? '单位' : '个人' }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{row}"><el-tag :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{row}">
            <el-button text size="small" type="primary" @click="changeStatus(row)">状态</el-button>
            <el-button text size="small" type="danger" @click="delMember(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination"><el-pagination background layout="prev, pager, next" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" /></div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([])
const loading = ref(true)
const page = ref(1); const size = ref(10); const total = ref(0)
const keyword = ref(''); const filterStatus = ref('')

const statusMap: Record<string, { l: string; t: string }> = {
  registering: { l: '注册中', t: 'info' }, pending_review: { l: '待审核', t: 'warning' },
  pending_cert: { l: '待证书', t: 'warning' }, pending_payment: { l: '待缴费', t: 'danger' },
  active: { l: '正式会员', t: 'success' }, rejected: { l: '已拒绝', t: 'danger' }
}
function statusLabel(s: string) { return statusMap[s]?.l || s }
function statusTag(s: string) { return statusMap[s]?.t || 'info' as any }

onMounted(() => fetchData())
async function fetchData() {
  loading.value = true
  try {
    const res = await adminApi.getMembers({ page: page.value, size: size.value, keyword: keyword.value, status: filterStatus.value })
    list.value = res.data?.list || []; total.value = res.data?.total || 0
  } catch {} finally { loading.value = false }
}
function search() { page.value = 1; fetchData() }

async function changeStatus(row: any) {
  const statuses = ['active', 'pending_review', 'pending_cert', 'pending_payment', 'rejected', 'registering', 'expired']
  try {
    const { value } = await ElMessageBox.prompt('输入新状态: ' + statuses.join(', '), '修改状态', { inputValue: row.status })
    if (value) { await adminApi.updateMemberStatus(row.id, value); ElMessage.success('已更新'); fetchData() }
  } catch {}
}
async function delMember(row: any) {
  try {
    await ElMessageBox.confirm('确认删除该会员？', '警告', { type: 'warning' })
    await adminApi.deleteMember(row.id); ElMessage.success('已删除'); fetchData()
  } catch {}
}
</script>

<style scoped lang="scss">
.admin-members {
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px 0;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 12px;
  h3 {
    font-size: 22px;
    font-weight: 600;
    color: #1a1a2e;
    margin: 0;
  }
}
.filters { display: flex; gap: 12px; }
.el-card {
  border-radius: 10px;
}
.pagination { display: flex; justify-content: center; padding: 20px 0; }
</style>

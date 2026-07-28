<template>
  <div class="orgs-page" v-loading="loading">
    <div class="page-header">
      <h3>参加的组织机构</h3>
      <el-button type="primary" @click="showJoin = true">加入分会</el-button>
    </div>

    <el-card>
      <el-table :data="myOrgs" stripe>
        <el-table-column prop="org.name" label="组织名称" />
        <el-table-column prop="created_at" label="加入时间" width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button text type="danger" @click="leaveOrg(row)">退出</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && myOrgs.length === 0" description="暂未加入任何分会" />
    </el-card>

    <el-dialog v-model="showJoin" title="选择要加入的分会" width="480px">
      <el-tree :data="orgTree" :props="{ label: 'name', children: 'children' }" node-key="id" @node-click="selectOrg" highlight-current />
      <div class="selected" v-if="selectedOrg">已选择：<strong>{{ selectedOrg.name }}</strong></div>
      <template #footer>
        <el-button @click="showJoin = false">取消</el-button>
        <el-button type="primary" :disabled="!selectedOrg" @click="joinOrg">确认加入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { orgApi } from '@/api/index'
import { ElMessage, ElMessageBox } from 'element-plus'

const myOrgs = ref<any[]>([])
const orgTree = ref<any[]>([])
const loading = ref(true)
const showJoin = ref(false)
const selectedOrg = ref<any>(null)

onMounted(async () => {
  try {
    const [myRes, treeRes] = await Promise.all([orgApi.getMyOrgs(), orgApi.getTree()])
    myOrgs.value = myRes.data || []
    orgTree.value = treeRes.data || []
  } catch {} finally { loading.value = false }
})

function selectOrg(node: any) { selectedOrg.value = node }

async function joinOrg() {
  if (!selectedOrg.value) return
  try {
    await orgApi.joinOrg(selectedOrg.value.id)
    ElMessage.success('加入成功')
    showJoin.value = false
    const res = await orgApi.getMyOrgs()
    myOrgs.value = res.data || []
  } catch {}
}

async function leaveOrg(row: any) {
  try {
    await ElMessageBox.confirm('确认退出该组织？', '提示', { type: 'warning' })
    await orgApi.leaveOrg(row.id)
    ElMessage.success('已退出')
    const res = await orgApi.getMyOrgs()
    myOrgs.value = res.data || []
  } catch {}
}

function formatDate(d: string) { return d ? d.slice(0, 16) : '' }
</script>

<style scoped lang="scss">
.orgs-page { max-width: 800px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.selected { margin-top: 16px; padding: 12px; background: #f0f9ff; border-radius: 8px; }
</style>

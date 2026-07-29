<template>
  <div class="orgs-page" v-loading="loading">
    <div class="page-header">
      <h3>加入的总会或分会信息</h3>
      <el-button type="primary" @click="showJoin = true">新的加入</el-button>
    </div>

    <!-- 卡片列表 -->
    <div v-if="combinedList.length" class="card-list">
      <div v-for="item in combinedList" :key="item.id" class="org-card">
        <div class="card-body">
          <div class="card-left">
            <div class="org-name">{{ item.org?.name || '未知机构' }}</div>
            <div class="card-meta">
              <el-tag type="warning" effect="plain" size="small">
                {{ memberLevel || '-' }}
              </el-tag>
              <span class="meta-time">{{ formatDate(item.created_at) }}</span>
            </div>
          </div>
          <div class="card-right">
            <el-button v-if="item._type === 'org'" text type="danger" size="small" @click="leaveOrg(item)">退出</el-button>
          </div>
        </div>
      </div>
    </div>
    <el-empty v-else-if="!loading" description="暂未加入任何组织机构" image-size="80" />

    <el-dialog v-model="showJoin" title="选择要加入的组织机构" width="480px">
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
import { ref, computed, onMounted } from 'vue'
import { orgApi, applicationApi, feeApi } from '@/api/index'
import { ElMessage, ElMessageBox } from 'element-plus'

const myOrgs = ref<any[]>([])
const memberLevel = ref('')
const orgTree = ref<any[]>([])
const approvedApps = ref<any[]>([])
const loading = ref(true)
const showJoin = ref(false)
const selectedOrg = ref<any>(null)

const combinedList = computed(() => {
  const apps = approvedApps.value.map((a: any) => ({ ...a, _type: 'application' }))
  const orgs = myOrgs.value.map((o: any) => ({ ...o, _type: 'org' }))
  return [...apps, ...orgs]
})

onMounted(async () => {
  try {
    const currentYear = new Date().getFullYear()
    const [myRes, treeRes, appRes, feeRes] = await Promise.all([
      orgApi.getMyOrgs(),
      orgApi.getTree(),
      applicationApi.getMyApplications(),
      feeApi.getMyFees({ year: currentYear })
    ])
    myOrgs.value = myRes.data || []
    orgTree.value = treeRes.data || []
    approvedApps.value = (appRes.data || []).filter((a: any) => a.status === 'approved')
    const fees: any[] = feeRes.data || []
    // 优先显示已缴费级别，没有则显示未缴费级别
    const paid = fees.find((f: any) => f.status === 'paid')
    const unpaid = fees.find((f: any) => f.status === 'unpaid')
    memberLevel.value = paid?.level_name || unpaid?.level_name || ''
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

function formatDate(d: string) { return d ? d.replace('T', ' ').slice(0, 16) : '' }
</script>

<style scoped lang="scss">
.orgs-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 20px 0;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  h3 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #303133;
    position: relative;
    padding-left: 14px;
    &::before {
      content: '';
      position: absolute;
      left: 0;
      top: 2px;
      bottom: 2px;
      width: 4px;
      background: #409eff;
      border-radius: 2px;
    }
  }
}
.card-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.org-card {
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 12px;
  padding: 20px 24px;
  transition: all 0.25s ease;
  &:hover {
    border-color: #c8d6e5;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
    transform: translateY(-2px);
  }
}
.card-body {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-left {
  flex: 1;
  min-width: 0;
}
.org-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 10px;
}
.card-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.meta-time {
  font-size: 13px;
  color: #909399;
}
.card-right {
  flex-shrink: 0;
  margin-left: 20px;
}
.selected {
  margin-top: 16px;
  padding: 12px 16px;
  background: #ecf5ff;
  border-radius: 8px;
  color: #409eff;
}
</style>

<template>
  <div class="orgs-page" v-loading="loading">
    <div class="page-header">
      <h3>加入的总会或分会信息</h3>
      <el-tooltip v-if="!hasPaidOrg" content="暂无缴费加入的组织，无法加入新组织" placement="top">
        <el-button type="primary" disabled>新的加入</el-button>
      </el-tooltip>
      <el-button v-else type="primary" @click="showJoin = true">新的加入</el-button>
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
              <el-tag
                :type="item._type === 'application' ? 'success' : 'info'"
                effect="plain"
                size="small"
              >
                {{ item._type === 'application' ? '缴费加入' : '主动加入' }}
              </el-tag>
              <span class="meta-time">{{ formatDate(item._type === 'org' ? item.joined_at : item.created_at) }}</span>
            </div>
          </div>
          <div class="card-right">
            <el-button v-if="item._type === 'org'" text type="danger" size="small" @click="leaveOrg(item)">退出</el-button>
          </div>
        </div>
      </div>
    </div>
    <el-empty v-else-if="!loading" description="暂未加入任何组织机构" image-size="80" />

    <el-dialog v-model="showJoin" title="选择要加入的组织机构" width="520px" class="join-dialog">
      <div class="dialog-body">
        <el-input
          v-model="searchQuery"
          placeholder="搜索组织机构..."
          clearable
          class="search-input"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>

        <el-tree
          ref="treeRef"
          :data="processedOrgTree"
          :props="{ label: 'name', children: 'children', disabled: 'disabled' }"
          node-key="id"
          :filter-node-method="filterNode"
          default-expand-all
          highlight-current
          class="org-tree"
          @node-click="selectOrg"
        >
          <template #default="{ data }">
            <span class="tree-node-label">
              <span>{{ data.name }}</span>
              <el-tag v-if="data.disabled" type="info" size="small" effect="plain">已加入</el-tag>
            </span>
          </template>
        </el-tree>

        <div class="selected" v-if="selectedOrg && !selectedOrg.disabled">
          <el-icon><Check /></el-icon>
          <span>已选择：<strong>{{ selectedOrg.name }}</strong></span>
        </div>

        <!-- 会员级别选择 -->
        <div v-if="selectedOrg && !selectedOrg.disabled && filteredLevels.length" class="level-section">
          <div class="level-section-title">选择会员级别</div>
          <el-radio-group v-model="selectedLevelId" class="level-radio-group">
            <el-radio
              v-for="item in filteredLevels"
              :key="item.level_id"
              :value="item.level_id"
              class="level-radio"
            >
              <div class="level-radio-content">
                <span class="level-name">{{ item.level?.name }}</span>
                <el-tag v-if="item.level?.level === currentLevelNum" size="small" type="warning" effect="plain">同级</el-tag>
                <el-tag v-else size="small" type="info" effect="plain">低级</el-tag>
              </div>
            </el-radio>
          </el-radio-group>
          <div v-if="!filteredLevels.length && availableLevels.length" class="level-empty-tip">
            该组织暂无适合您当前级别的选项
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="handleCancel">取消</el-button>
        <el-button type="primary" :disabled="!canJoin" @click="joinOrg">
          确认加入
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { orgApi, applicationApi, feeApi } from '@/api/index'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Check } from '@element-plus/icons-vue'

const myOrgs = ref<any[]>([])
const memberLevel = ref('')
const orgTree = ref<any[]>([])
const approvedApps = ref<any[]>([])
const loading = ref(true)
const showJoin = ref(false)
const selectedOrg = ref<any>(null)
const searchQuery = ref('')
const treeRef = ref<any>(null)

const combinedList = computed(() => {
  const apps = approvedApps.value.map((a: any) => ({ ...a, _type: 'application' }))
  const orgs = myOrgs.value.map((o: any) => ({ ...o, _type: 'org' }))
  return [...apps, ...orgs]
})

// 是否有通过缴费加入的组织
const hasPaidOrg = computed(() => approvedApps.value.length > 0)

// 已加入的组织 ID 集合（含申请批准的 + 直接加入的）
const joinedOrgIds = computed(() => {
  const ids = new Set<number>()
  myOrgs.value.forEach((o: any) => ids.add(o.org_id ?? o.org?.id))
  approvedApps.value.forEach((a: any) => ids.add(a.org_id ?? a.org?.id))
  return ids
})

// 计算树中每个节点的深度（用于层级过滤）
function computeDepthMap(nodes: any[], depth: number = 0, map: Map<number, number> = new Map()): Map<number, number> {
  for (const n of nodes) {
    map.set(n.id, depth)
    if (n.children) computeDepthMap(n.children, depth + 1, map)
  }
  return map
}

// 在组织树节点上标记 disabled
// 已加入的组织不可选；比已加入组织更高层级的不可选
const processedOrgTree = computed(() => {
  const depthMap = computeDepthMap(orgTree.value)

  // 已加入组织的深度集合
  const joinedDepths = new Set<number>()
  joinedOrgIds.value.forEach((id: number) => {
    const d = depthMap.get(id)
    if (d !== undefined) joinedDepths.add(d)
  })

  // 已加入组织的最小深度（最顶层），低于此层级的组织不可加入
  const minDepth = joinedDepths.size > 0 ? Math.min(...Array.from(joinedDepths)) : -1

  function markDisabled(nodes: any[]): any[] {
    return nodes.map(n => {
      const isJoined = joinedOrgIds.value.has(n.id)
      const nodeDepth = depthMap.get(n.id) ?? 0
      const disabled = isJoined || (minDepth >= 0 && nodeDepth < minDepth)
      return {
        ...n,
        disabled,
        children: n.children ? markDisabled(n.children) : n.children
      }
    })
  }
  return markDisabled(orgTree.value)
})

// 所有会员级别列表（从 API 获取）
const allLevels = ref<any[]>([])

// 当前会员级别的数值
const currentLevelNum = computed(() => {
  // 先从已加入组织的 level_id 推断
  for (const org of myOrgs.value) {
    if (org.level_id > 0) {
      const found = allLevels.value.find((l: any) => l.id === org.level_id)
      if (found) return found.level
    }
  }
  // 再尝试从 memberLevel 名称匹配
  if (memberLevel.value) {
    const found = allLevels.value.find((l: any) => l.name === memberLevel.value)
    if (found) return found.level
  }
  return -1 // 未找到则不过滤
})

// 所选组织的可用级别
const availableLevels = computed(() => {
  if (!selectedOrg.value || selectedOrg.value.disabled) return []
  return selectedOrg.value.levels || []
})

// 过滤后的级别（同级+低级，优先同级）
const filteredLevels = computed(() => {
  const levels = availableLevels.value
  if (currentLevelNum.value < 0) return levels
  // 同级别
  const same = levels.filter((l: any) => l.level?.level === currentLevelNum.value)
  // 低级别
  const lower = levels.filter((l: any) => l.level?.level < currentLevelNum.value)
  return [...same, ...lower]
})

// 默认选中第一个可用级别（优先同级）
const selectedLevelId = ref<number>(0)

// 选择组织时自动设置默认级别
watch(selectedOrg, (org) => {
  if (org && !org.disabled) {
    const levels = filteredLevels.value
    if (levels.length > 0) {
      selectedLevelId.value = levels[0].level_id
    } else {
      selectedLevelId.value = 0
    }
  } else {
    selectedLevelId.value = 0
  }
})

// 是否允许确认加入
const canJoin = computed(() => !!selectedOrg.value && !selectedOrg.value.disabled && selectedLevelId.value > 0)

onMounted(async () => {
  try {
    const currentYear = new Date().getFullYear()
    const [myRes, treeRes, appRes, feeRes, levelRes] = await Promise.all([
      orgApi.getMyOrgs(),
      orgApi.getTree(),
      applicationApi.getMyApplications(),
      feeApi.getMyFees({ year: currentYear }),
      orgApi.getMemberLevels()
    ])
    myOrgs.value = myRes.data || []
    orgTree.value = treeRes.data || []
    approvedApps.value = (appRes.data || []).filter((a: any) => a.status === 'approved')
    allLevels.value = levelRes.data || []
    const fees: any[] = feeRes.data || []
    // 优先显示已缴费级别，没有则显示未缴费级别
    const paid = fees.find((f: any) => f.status === 'paid')
    const unpaid = fees.find((f: any) => f.status === 'unpaid')
    memberLevel.value = paid?.level_name || unpaid?.level_name || ''
  } catch {} finally { loading.value = false }
})

// 树搜索过滤
watch(searchQuery, (val) => {
  treeRef.value?.filter(val)
})

function filterNode(value: string, data: any) {
  if (!value) return true
  return data.name.toLowerCase().includes(value.toLowerCase())
}

function selectOrg(node: any) {
  if (node.disabled) {
    ElMessage.info('您已加入该组织，无需重复加入')
    selectedOrg.value = null
    return
  }
  selectedOrg.value = node
}

function handleCancel() {
  showJoin.value = false
  searchQuery.value = ''
  selectedOrg.value = null
}

async function joinOrg() {
  if (!selectedOrg.value) return
  try {
    await orgApi.joinOrg(selectedOrg.value.id, selectedLevelId.value)
    ElMessage.success('加入成功')
    showJoin.value = false
    searchQuery.value = ''
    selectedOrg.value = null
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
 width: 100%;
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
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}

/* ── 加入对话框 ── */
.join-dialog {
  :deep(.el-dialog__body) {
    padding: 0;
  }
}

.dialog-body {
  padding: 20px 24px;
}

.search-input {
  margin-bottom: 16px;
}

.org-tree {
  max-height: 380px;
  overflow-y: auto;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 8px 0;

  :deep(.el-tree-node__content) {
    height: 40px;
    padding: 0 12px;
  }

  :deep(.el-tree-node.is-disabled > .el-tree-node__content) {
    cursor: not-allowed;
    opacity: 0.6;
  }
}

.tree-node-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

/* ── 级别选择 ── */
.level-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}
.level-section-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 10px;
}
.level-radio-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.level-radio {
  margin-right: 0;
}
.level-radio-content {
  display: flex;
  align-items: center;
  gap: 8px;
}
.level-name {
  font-size: 14px;
}
.level-empty-tip {
  font-size: 13px;
  color: #909399;
  padding: 8px 0;
}
</style>

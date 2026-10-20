<template>
  <div class="orgs-page" v-loading="loading">
    <div class="page-header">
      <h3>加入的组织机构信息</h3>
      <el-tooltip v-if="!hasMembership" content="暂无有效会籍，无法加入新组织" placement="top">
        <el-button type="primary" disabled>新的加入</el-button>
      </el-tooltip>
      <el-button v-else type="primary" @click="showJoin = true">新的加入</el-button>
    </div>

    <!-- 卡片列表 -->
    <div v-if="combinedList.length" class="card-list">
      <div v-for="item in combinedList" :key="item._key" class="org-card">
        <div class="card-body">
          <div class="card-left">
            <div class="org-name">{{ item.org?.name || '未知机构' }}</div>
            <div
              v-if="item._intro"
              class="org-intro"
              :class="{ expanded: introExpanded(introKey(item)) }"
              :ref="setIntroRef(introKey(item))"
            >{{ item._intro }}</div>
            <button
              v-if="item._intro && introOverflowed(introKey(item))"
              type="button"
              class="intro-toggle"
              @click="toggleIntro(introKey(item))"
            >{{ introExpanded(introKey(item)) ? '收起' : '展开' }}</button>
            <div class="card-meta">
              <el-tag type="warning" effect="plain" size="small">
                {{ itemLevelName(item) }}
              </el-tag>
              <el-tag
                :type="item._type === 'application' ? 'success' : 'info'"
                effect="plain"
                size="small"
              >
                {{ item._type === 'application' ? '入会机构' : '所属机构' }}
              </el-tag>
              <span class="meta-time">{{ formatDate(item._type === 'org' ? item.joined_at : item.created_at) }}</span>
            </div>
          </div>
          <div class="card-right">
            <el-button
              v-if="item._canLeave && item.org?.parent_id !== 0"
              text
              type="danger"
              size="small"
              @click="leaveOrg(item)"
            >退出</el-button>
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
            <span class="tree-node-label" :title="data.description || ''">
              <span>{{ data.name }}</span>
              <el-tag v-if="data.disabled" type="info" size="small" effect="plain">
                {{ data._disabledReason === 'hierarchy' ? '不可加入' : '已加入' }}
              </el-tag>
            </span>
          </template>
        </el-tree>

        <div class="selected" v-if="selectedOrg && !selectedOrg.disabled">
          <el-icon><Check /></el-icon>
          <span>已选择：<strong>{{ selectedOrg.name }}</strong></span>
        </div>

        <!-- 所选机构简介 -->
        <div v-if="selectedOrg && !selectedOrg.disabled && selectedOrg.description" class="org-intro-box">
          <div class="intro-title">机构简介</div>
          <p class="intro-text">{{ selectedOrg.description }}</p>
        </div>

        <!-- 会员级别选择（外层用 availableLevels 判断：无可选级别时也要显示下方提示，
             原先外层要求 filteredLevels.length，内层提示恒不可见） -->
        <div v-if="selectedOrg && !selectedOrg.disabled && availableLevels.length" class="level-section">
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
        <el-button type="primary" :disabled="!canJoin" :loading="joining" @click="joinOrg">
          确认加入
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { orgApi, applicationApi } from '@/api/index'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Check } from '@element-plus/icons-vue'
import { useIntroToggle } from '@/utils/introToggle'

const myOrgs = ref<any[]>([])
const memberLevel = ref('')
const orgTree = ref<any[]>([])
const approvedApps = ref<any[]>([])
const loading = ref(true)
const showJoin = ref(false)
const joining = ref(false)
const selectedOrg = ref<any>(null)
const searchQuery = ref('')
const treeRef = ref<any>(null)

// 机构简介映射：从机构树按 id 索引。
// 历史申请或会籍关系可能只有机构 ID，需回查机构树补简介。
const orgIntroMap = computed(() => {
  const map = new Map<number, string>()
  const walk = (nodes: any[]) => {
    for (const n of nodes) {
      if (n.description) map.set(n.id, n.description)
      if (n.children?.length) walk(n.children)
    }
  }
  walk(orgTree.value)
  return map
})

function orgIntroOf(item: any) {
  const id = Number(item.org_id ?? item.org?.id)
  return item.org?.description || orgIntroMap.value.get(id) || ''
}

// 简介「展开/收起」：key 需在列表刷新前后保持稳定（用记录的唯一 _key）
const { setIntroRef, introExpanded, introOverflowed, toggleIntro, measureIntro } = useIntroToggle()
function introKey(item: any) {
  return 'org-intro-' + item._key
}

const combinedList = computed(() => {
  // 会籍关系以 member_user_orgs 为主，已通过申请仅作为历史数据兜底。
  // 同一组织机构可能同时存在多条记录，按组织 ID 去重，
  // 优先级：member_user_orgs > 已批准入会申请（历史数据兜底）。
  const seen = new Set<number>()
  const list: any[] = []
  const pushUnique = (orgId: any, item: any) => {
    const id = Number(orgId)
    if (id > 0) {
      if (seen.has(id)) return
      seen.add(id)
    }
    list.push(item)
  }

  // 1. 中性会籍关系
  myOrgs.value.forEach((o: any) => {
    pushUnique(o.org_id ?? o.org?.id, {
      ...o,
      _type: 'org',
      _canLeave: true,
      _key: 'org-' + o.id
    })
  })

  // 2. 兜底：已批准的入会申请（历史数据可能尚无会籍关系）。
  approvedApps.value.forEach((a: any) => {
    const isRoot = (a.org?.parent_id ?? 0) === 0
    pushUnique(a.org_id ?? a.org?.id, {
      ...a,
      _type: isRoot ? 'application' : 'org',
      _key: 'app-' + a.id,
      joined_at: a.joined_at || a.created_at
    })
  })

  return list.map((it) => ({ ...it, _intro: orgIntroOf(it) }))
})

// 数据/文本变化后重新测量是否被截断（已展开的项不会被误判）
watch(combinedList, () => { void measureIntro() }, { immediate: true })

const hasMembership = computed(() => approvedApps.value.length > 0 || myOrgs.value.length > 0)

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
      const hierarchyBlocked = minDepth >= 0 && nodeDepth < minDepth
      const disabled = isJoined || hierarchyBlocked
      return {
        ...n,
        disabled,
        _disabledReason: isJoined ? 'joined' : (hierarchyBlocked ? 'hierarchy' : ''),
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
    const [myRes, treeRes, appRes, levelRes] = await Promise.all([
      orgApi.getMyOrgs(),
      orgApi.getTree(),
      applicationApi.getMyApplications(),
      orgApi.getMemberLevels()
    ])
    myOrgs.value = myRes.data || []
    orgTree.value = treeRes.data || []
    approvedApps.value = (appRes.data || []).filter((a: any) => a.status === 'approved')
    allLevels.value = levelRes.data || []
    const membershipWithLevel = myOrgs.value.find((o: any) => o.level_id)
    const currentLevel = allLevels.value.find((l: any) => l.id === membershipWithLevel?.level_id)
    memberLevel.value = currentLevel?.name || ''
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
    ElMessage.info(
      node._disabledReason === 'hierarchy'
        ? '只要加入任意分支机构或代表机构则默认加入更高层级的组织'
        : '您已加入该组织，无需重复加入'
    )
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
  // 防重复提交：双击会第二次拿到「已加入该组织」的错误提示
  if (joining.value) return
  joining.value = true
  try {
    await orgApi.joinOrg(selectedOrg.value.id, selectedLevelId.value)
    ElMessage.success('加入成功')
    showJoin.value = false
    searchQuery.value = ''
    selectedOrg.value = null
    const res = await orgApi.getMyOrgs()
    myOrgs.value = res.data || []
  } catch {} finally { joining.value = false }
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

// 每条加入记录的会员级别：优先记录自带的 level_name，其次按 level_id 反查，最后回退全局级别
function itemLevelName(item: any) {
  if (item.level_name) return item.level_name
  if (item.level_id) {
    const lvl = allLevels.value.find((l: any) => l.id === item.level_id)
    if (lvl) return lvl.name
  }
  return memberLevel.value || '-'
}
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
      background: #366d28;
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
.org-intro {
  font-size: 13px;
  line-height: 1.7;
  color: #606266;
  margin-bottom: 10px;
  /* 卡片内最多展示两行，保持列表紧凑；点击「展开」看全文 */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  &.expanded {
    display: block;
    -webkit-line-clamp: unset;
    overflow: visible;
  }
}
.intro-toggle {
  display: inline-block;
  margin: -6px 0 10px;
  padding: 0;
  border: none;
  background: none;
  color: #366d28;
  font-size: 13px;
  line-height: 1;
  cursor: pointer;
  &:hover { text-decoration: underline; }
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
  background: #ebf0ea;
  border-radius: 8px;
  color: #366d28;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}

/* 所选机构简介 */
.org-intro-box {
  margin-top: 12px;
  padding: 12px 16px;
  background: #f7f9fc;
  border: 1px solid #ebeef5;
  border-radius: 8px;

  .intro-title {
    font-size: 13px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 6px;
  }
  .intro-text {
    margin: 0;
    font-size: 13px;
    line-height: 1.7;
    color: #606266;
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 120px;
    overflow-y: auto;
  }
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

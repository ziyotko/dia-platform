<template>
  <div class="admin-fee-standards" v-loading="loading">
    <div class="page-header">
      <h3>会费标准管理</h3>
      <div class="header-tip">按会员等级设置年度缴费金额，机构关联等级后将按此标准生成会费</div>
    </div>

    <!-- No levels -->
    <el-empty v-if="!loading && levels.length === 0" description="暂无会员等级，请先创建会员等级" />

    <!-- Level cards -->
    <div class="level-cards" v-else>
      <el-card
        v-for="level in levels"
        :key="level.id"
        class="level-card"
        :body-style="{ padding: '0' }"
      >
        <div class="level-card-header">
          <div class="level-info">
            <span class="level-badge" :style="{ background: levelColor(level.level) }">{{ level.name }}</span>
            <span class="level-desc">{{ level.description }}</span>
          </div>
          <el-button size="small" type="primary" plain @click="addYear(level)">
            <el-icon><Plus /></el-icon> 新增年度
          </el-button>
        </div>

        <!-- Fee standard table for this level -->
        <el-table :data="feeMap[level.id] || []" stripe empty-text="暂未设置会费标准" style="width:100%">
          <el-table-column label="缴费年度" width="160">
            <template #default="{ row }">
              <el-input
                v-model="row._editing.year"
                type="number"
                min="2000"
                :max="maxYear"
                size="small"
                style="width:120px"
                :disabled="row._saving"
                placeholder="例如：2026"
              />
            </template>
          </el-table-column>
          <el-table-column label="会费金额（元/年）" min-width="220">
            <template #default="{ row }">
              <div class="amount-cell">
                <span class="amount-prefix">¥</span>
                <el-input-number
                  v-model="row._editing.amount"
                  :min="0"
                  :precision="2"
                  :step="100"
                  size="small"
                  style="width:160px"
                  :disabled="row._saving"
                  placeholder="输入金额"
                />
              </div>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="120">
            <template #default="{ row }">
              <el-tag v-if="row._changed" type="warning" size="small">已修改</el-tag>
              <el-tag v-else type="success" size="small" effect="plain">已保存</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button
                size="small"
                type="primary"
                :loading="row._saving"
                :disabled="!row._changed"
                @click="saveRow(level, row)"
              >
                保存
              </el-button>
              <el-button
                size="small"
                :disabled="row._saving"
                @click="resetRow(row)"
              >
                重置
              </el-button>
              <el-button
                size="small"
                type="danger"
                :loading="row._deleting"
                @click="deleteRow(level, row)"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

interface FeeItem {
  id?: number
  level_id: number
  year: number
  amount: number
  _editing: { year: number; amount: number }
  _changed: boolean
  _saving: boolean
  _deleting: boolean
}

const loading = ref(true)
const levels = ref<any[]>([])
const feeMap = reactive<Record<number, FeeItem[]>>({})

const maxYear = computed(() => new Date().getFullYear() + 10)

const levelColors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#b37feb']

function levelColor(level: number) {
  return levelColors[level % levelColors.length]
}

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const [levelRes, feeRes] = await Promise.all([
      adminApi.getMemberLevels(),
      adminApi.getFeeStandards()
    ])
    levels.value = levelRes.data || []

    // Build fee map
    const fees: FeeItem[] = (feeRes.data || []).map((f: any) => toFeeItem(f))
    for (const level of levels.value) {
      feeMap[level.id] = fees.filter((f) => f.level_id === level.id)
    }
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
}

function toFeeItem(f: any): FeeItem {
  return {
    id: f.id,
    level_id: f.level_id,
    year: f.year,
    amount: f.amount,
    _editing: { year: f.year, amount: f.amount },
    _changed: false,
    _saving: false,
    _deleting: false
  }
}

function addYear(level: any) {
  const existing = feeMap[level.id] || []
  // Suggest next year: current year, or last year in list + 1
  const maxExistingYear = existing.reduce((max, r) => Math.max(max, r.year), 0)
  const suggestedYear = maxExistingYear > 0 ? maxExistingYear + 1 : new Date().getFullYear()

  const newItem: FeeItem = {
    level_id: level.id,
    year: suggestedYear,
    amount: 0,
    _editing: { year: suggestedYear, amount: 0 },
    _changed: true,
    _saving: false,
    _deleting: false
  }
  if (!feeMap[level.id]) {
    feeMap[level.id] = []
  }
  feeMap[level.id].unshift(newItem)
}

async function saveRow(level: any, row: FeeItem) {
  const year = Number(row._editing.year)
  const amount = Number(row._editing.amount)
  if (!year || year < 2000) {
    ElMessage.warning('请输入有效的年度')
    return
  }
  if (amount < 0) {
    ElMessage.warning('金额不能为负数')
    return
  }

  row._saving = true
  try {
    const res = await adminApi.upsertFeeStandard({
      level_id: level.id,
      year,
      amount
    })
    // Update with server response
    const saved = res.data
    row.id = saved.id
    row.year = saved.year
    row.amount = saved.amount
    row._editing.year = saved.year
    row._editing.amount = saved.amount
    row._changed = false
    ElMessage.success('保存成功')
  } catch {
    // ignore
  } finally {
    row._saving = false
  }
}

function resetRow(row: FeeItem) {
  row._editing.year = row.year
  row._editing.amount = row.amount
  row._changed = false
}

async function deleteRow(level: any, row: FeeItem) {
  if (!row.id) {
    // Remove unsaved row
    const list = feeMap[level.id]
    if (list) {
      const idx = list.indexOf(row)
      if (idx >= 0) list.splice(idx, 1)
    }
    return
  }

  try {
    await ElMessageBox.confirm(`确认删除 ${row.year} 年度「${level.name}」的会费标准？`, '确认删除', {
      type: 'warning',
      confirmButtonText: '删除'
    })
  } catch {
    return
  }

  row._deleting = true
  try {
    await adminApi.deleteFeeStandard(row.id)
    const list = feeMap[level.id]
    if (list) {
      const idx = list.indexOf(row)
      if (idx >= 0) list.splice(idx, 1)
    }
    ElMessage.success('已删除')
  } catch {
    // ignore
  } finally {
    row._deleting = false
  }
}
</script>

<style scoped lang="scss">
.admin-fee-standards { max-width: 960px; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 20px;

  h3 { margin: 0; }

  .header-tip {
    font-size: 13px;
    color: #9ca3af;
  }
}

.level-cards {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.level-card {
  border-radius: 10px;
  overflow: hidden;

  .level-card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    background: #f8fafc;
    border-bottom: 1px solid #f0f0f0;

    .level-info {
      display: flex;
      align-items: center;
      gap: 12px;

      .level-badge {
        display: inline-block;
        padding: 4px 14px;
        border-radius: 20px;
        color: #fff;
        font-size: 14px;
        font-weight: 500;
      }

      .level-desc {
        font-size: 13px;
        color: #9ca3af;
      }
    }
  }
}

.amount-cell {
  display: flex;
  align-items: center;
  gap: 4px;

  .amount-prefix {
    font-size: 14px;
    color: #6b7280;
    font-weight: 600;
  }
}
</style>
